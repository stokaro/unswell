#!/usr/bin/env python3
"""Freeze and run an exploratory five-window EditLens paper-level screen."""

from __future__ import annotations

import argparse
import collections
import hashlib
import importlib.util
import json
import math
from pathlib import Path


HERE = Path(__file__).resolve().parent
PRIOR = HERE.parent / "2026-11-10-editlens-academic-transfer"
ARB = HERE.parent / "2026-11-09-editlens-transfer" / "run.py"
DATA_HASH = "6e0eb8058578afce4aaf81c822b70762afd9f18534eb5f1e89aa976cd051b793"
MODEL_HASH = "ddd1173f2ef517ad499965e5029fae8099a8054a2bc76d8134e5889cc4ed1b3e"
OLD_SCORES_HASH = "6779dff96cbd320c940276cde5629f62ac7a9d7fcf9600961b3386629d91a299"
OLD_SELECTION_HASH = "9a31161e982dfcbbd96757a19a69f41e474f562459d200ce5e69cfb402876ab4"


def need(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def file_hash(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1048576), b""):
            digest.update(block)
    return digest.hexdigest()


def sha_text(value: str) -> str:
    return hashlib.sha256(value.encode("utf-8")).hexdigest()


def load(path: Path, name: str):
    spec = importlib.util.spec_from_file_location(name, path)
    need(spec is not None and spec.loader is not None, f"Missing helper: {name}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def select(data: Path, academic) -> tuple[list[dict], str, dict]:
    controls, pairs, _ = academic.read_source(data)
    need(len(controls) == 400, "Human-only paper count changed")
    paper_ids = sorted(controls, key=lambda paper: (
        academic.rank(paper, "unswell-editlens-calibration-v1"), paper))
    need(len(paper_ids) == 400, "Missing paper")
    selected = []
    for position, paper in enumerate(paper_ids):
        eligible = [row for row in controls[paper]
                    if row["negative"] == "unpaired" and academic.source_eligible(row)]
        need(len({row["pair_id"] for row in eligible}) == len(eligible),
             f"Duplicate control paragraph in {paper}")
        need(len(eligible) >= 5, f"Fewer than five control paragraphs in {paper}")
        chosen = sorted(eligible, key=lambda row: (
            academic.rank(paper, row["pair_id"], "unswell-editlens-control-v1"),
            row["pair_id"]))[:5]
        for row in chosen:
            selected.append({"role": "calibration" if position < 300 else "holdout",
                             "regime": "human", "row": row})
    old_selected, old_selection, old_metadata = academic.choose_rows(controls, pairs)
    need(old_selection == OLD_SELECTION_HASH, "Prior test selection changed")
    old_calibration = {item["row"]["paper_id"]: item["row"]["pair_id"]
                       for item in old_selected if item["role"] == "calibration"}
    need(len(old_calibration) == 300, "Prior calibration count changed")
    for item in selected:
        if item["role"] == "calibration" and item["row"]["paper_id"] not in old_calibration:
            raise ValueError("Calibration paper drift")
    first = {item["row"]["paper_id"]: item["row"]["pair_id"]
             for index, item in enumerate(selected) if index % 5 == 0 and item["role"] == "calibration"}
    need(first == old_calibration, "First control paragraph changed")
    test_papers = {item["row"]["paper_id"] for item in old_selected
                   if item["role"] == "test"}
    need(len(test_papers) == 58 and not set(paper_ids).intersection(test_papers),
         "Calibration or holdout overlaps test")
    identities = [{"role": item["role"], "regime": "human",
                   "paper_id": item["row"]["paper_id"],
                   "pair_id": item["row"]["pair_id"],
                   "text_sha256": sha_text(item["row"]["text"])}
                  for item in selected]
    selection_hash = sha_text(json.dumps(identities, sort_keys=True,
                                        separators=(",", ":"), ensure_ascii=False))
    metadata = {"calibration_papers": 300, "holdout_papers": 100,
                "control_windows": len(selected),
                "test_papers": old_metadata["test_papers"],
                "test_pairs": old_metadata["test_pairs"]}
    return selected, selection_hash, metadata


def check_prior(data: Path) -> list[dict]:
    need(file_hash(PRIOR / "scores.json") == OLD_SCORES_HASH,
         "Prior scores changed")
    prior_verifier = load(PRIOR / "verify.py", "five_window_prior_verify")
    checked = prior_verifier.verify(data, PRIOR / "scores.json",
                                    PRIOR / "result.json")
    need(checked == {"rows": 880, "papers": 58, "pairs": 290,
                     "false_positives": 2, "true_positives": 21,
                     "feasibility_pass": False}, "Prior result changed")
    scores = json.loads((PRIOR / "scores.json").read_text())
    test = [row for row in scores if row["role"] == "test"]
    need(len(test) == 580, "Prior test count changed")
    return test


def complete(groups: dict[str, list[dict]]) -> dict[str, float]:
    means = {}
    for paper, rows in groups.items():
        need(len(rows) == 5 and len({row["pair_id"] for row in rows}) == 5,
             f"Missing or duplicate paper windows: {paper}")
        if all(row["score"] is not None and not row["truncated"] for row in rows):
            means[paper] = sum(row["score"] for row in rows) / 5
    return means


def grouped(rows: list[dict], role: str, regime: str) -> dict[str, list[dict]]:
    groups = collections.defaultdict(list)
    for row in rows:
        if row["role"] == role and row["regime"] == regime:
            groups[row["paper_id"]].append(row)
    return groups


def summarize(controls: list[dict], tests: list[dict], selection_hash: str,
              metadata: dict, arb) -> dict:
    calibration = complete(grouped(controls, "calibration", "human"))
    holdout = complete(grouped(controls, "holdout", "human"))
    test_human = complete(grouped(tests, "test", "human"))
    test_ai = complete(grouped(tests, "test", "rewrite"))
    matched_ids = sorted(set(test_human) & set(test_ai))
    need(len(calibration) >= 3, "Insufficient calibration papers")
    threshold = sorted(calibration.values(), reverse=True)[2]
    pairs = [(test_human[paper], test_ai[paper]) for paper in matched_ids]
    need(pairs, "No matched test papers")
    test_by_paper = grouped(tests, "test", "rewrite")
    by_generator: dict[str, list[tuple[float, float]]] = collections.defaultdict(list)
    for paper in matched_ids:
        generators = {row["generator"] for row in test_by_paper[paper]}
        need(len(generators) == 1, "Mixed generators within paper")
        by_generator[generators.pop()].append((test_human[paper], test_ai[paper]))
    metric = lambda values: {
        "papers": len(values), "auc": arb.auc(values),
        "win_fraction": sum(float(ai > human) + 0.5 * float(ai == human)
                            for human, ai in values) / len(values),
        "true_positives": sum(ai > threshold for _, ai in values),
        "tpr": sum(ai > threshold for _, ai in values) / len(values),
    }
    overall = metric(pairs)
    control_fp = sum(value > threshold for value in holdout.values())
    paired_fp = sum(test_human[paper] > threshold for paper in matched_ids)
    result = {
        "format": "unswell-editlens-five-window-v1",
        "scope": "exposed_academic_paper_aggregation_only",
        "dataset_sha256": DATA_HASH, "model_sha256": MODEL_HASH,
        "model_revision": "f7cb4b06e5067ecdb66c566c7982a413f86f569f",
        "prior_scores_sha256": OLD_SCORES_HASH,
        "prior_selection_sha256": OLD_SELECTION_HASH,
        "control_selection_sha256": selection_hash,
        "selected": metadata,
        "control_windows_selected": len(controls),
        "control_windows_with_score": sum(row["score"] is not None for row in controls),
        "control_windows_below_50": sum(row["score"] is None for row in controls),
        "control_windows_truncated": sum(row["truncated"] for row in controls),
        "eligible_calibration_papers": len(calibration),
        "eligible_holdout_papers": len(holdout),
        "eligible_test_human_papers": len(test_human),
        "eligible_test_ai_papers": len(test_ai),
        "eligible_matched_papers": len(matched_ids),
        "threshold_strictly_greater": threshold,
        "calibration_false_positives": sum(value > threshold for value in calibration.values()),
        "holdout_false_positives": control_fp,
        "holdout_fpr": control_fp / len(holdout) if holdout else None,
        "paired_human_false_positives": paired_fp,
        "paired_human_fpr": paired_fp / len(matched_ids),
        "overall": overall,
        "by_generator": {name: metric(values) for name, values in sorted(by_generator.items())},
    }
    result["feasibility_pass"] = all((
        len(calibration) >= 285, len(holdout) >= 95, len(matched_ids) >= 55,
        result["holdout_fpr"] is not None and result["holdout_fpr"] <= 0.02,
        result["paired_human_fpr"] <= 0.02,
        overall["auc"] is not None and overall["auc"] >= 0.75,
        overall["tpr"] >= 0.20, overall["win_fraction"] >= 0.65,
    ))
    return result


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--data", type=Path, required=True)
    parser.add_argument("--model-dir", type=Path)
    parser.add_argument("--preflight", action="store_true")
    args = parser.parse_args()
    need(file_hash(args.data) == DATA_HASH, "Dataset changed")
    academic = load(PRIOR / "run.py", "five_window_academic")
    selected, selection_hash, metadata = select(args.data, academic)
    if args.preflight:
        print(json.dumps({"selection_sha256": selection_hash,
                          "selected": metadata}, sort_keys=True))
        return
    need(args.model_dir is not None, "Model directory required")
    freeze = json.loads((HERE / "freeze.json").read_text())
    need(freeze["data_sha256"] == DATA_HASH and
         freeze["model_sha256"] == MODEL_HASH and
         freeze["prior_scores_sha256"] == OLD_SCORES_HASH and
         freeze["prior_selection_sha256"] == OLD_SELECTION_HASH,
         "Frozen identity changed")
    for path, expected in ((HERE / "PROTOCOL.md", freeze["protocol_sha256"]),
                           (HERE / "run.py", freeze["runner_sha256"]),
                           (PRIOR / "run.py", freeze["academic_runner_sha256"]),
                           (ARB, freeze["arb_runner_sha256"])):
        need(file_hash(path) == expected, f"Frozen file changed: {path}")
    need(selection_hash == freeze["selection_sha256"], "Selection changed")
    need(file_hash(args.model_dir / "onnx/model.onnx") == MODEL_HASH,
         "Model changed")
    for name, expected in freeze["model_file_sha256"].items():
        need(file_hash(args.model_dir / name) == expected, f"Model file changed: {name}")
    tests = check_prior(args.data)
    arb = load(ARB, "five_window_arb")
    tokenizer = arb.Tokenizer.from_file(str(args.model_dir / "tokenizer.json"))
    mapped = [{"role": item["role"], "regime": "human",
               "row": {"text": item["row"]["text"],
                       "text_sha256": sha_text(item["row"]["text"]),
                       "source_dataset": "academic_paper",
                       "pair_id": item["row"]["pair_id"],
                       "generator_model": None}} for item in selected]
    controls = arb.prepare(mapped, tokenizer)
    for output, item in zip(controls, selected, strict=True):
        output["paper_id"] = item["row"]["paper_id"]
    arb.score(controls, args.model_dir / "onnx/model.onnx")
    need(all(row["score"] is None or
             (math.isfinite(row["score"]) and 0 <= row["score"] <= 1)
             for row in controls),
         "Invalid scores")
    result = summarize(controls, tests, selection_hash, metadata, arb)
    (HERE / "result.json").write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    without_text = [{key: value for key, value in row.items() if key != "ids"}
                    for row in controls]
    (HERE / "scores.json").write_text(json.dumps(without_text, indent=2,
                                               sort_keys=True) + "\n")
    print(json.dumps({"pass": result["feasibility_pass"],
                      "threshold": result["threshold_strictly_greater"],
                      "holdout_fpr": result["holdout_fpr"],
                      "paired_fpr": result["paired_human_fpr"],
                      "overall": result["overall"]}, sort_keys=True))


if __name__ == "__main__":
    main()
