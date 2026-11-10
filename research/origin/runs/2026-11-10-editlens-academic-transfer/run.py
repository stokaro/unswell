#!/usr/bin/env python3
"""Frozen EditLens transfer screen on matched pre-2022 academic prose."""

from __future__ import annotations

import argparse
import collections
import hashlib
import importlib.util
import json
import re
from pathlib import Path


HERE = Path(__file__).resolve().parent
ARB_RUNNER = HERE.parent / "2026-11-09-editlens-transfer" / "run.py"
DATA_HASH = "6e0eb8058578afce4aaf81c822b70762afd9f18534eb5f1e89aa976cd051b793"
MODEL_HASH = "ddd1173f2ef517ad499965e5029fae8099a8054a2bc76d8134e5889cc4ed1b3e"
MODEL_REVISION = "f7cb4b06e5067ecdb66c566c7982a413f86f569f"
WORDS = re.compile(r"\b\w+\b")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def sha_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def sha_text(text: str) -> str:
    return hashlib.sha256(text.encode("utf-8")).hexdigest()


def rank(*parts: str) -> str:
    return sha_text("\0".join(parts))


def source_eligible(row: dict) -> bool:
    return (row.get("chunk_ok") is True
            and row.get("kept_original") is False
            and 50 <= len(WORDS.findall(row["text"])) <= 300)


def read_source(path: Path) -> tuple[dict, dict, collections.Counter]:
    controls: dict[str, list[dict]] = collections.defaultdict(list)
    pairs: dict[tuple[str, str], dict] = collections.defaultdict(dict)
    counts: collections.Counter = collections.Counter()
    with path.open(encoding="utf-8") as source:
        for line_number, line in enumerate(source, 1):
            row = json.loads(line)
            split = row["split"]
            label = row["label"]
            counts[(split, label)] += 1
            if split == "human_only" and label == "human" and row["year"] <= 2018:
                controls[row["paper_id"]].append(row)
            if split == "test" and row["year"] <= 2021:
                key = (row["paper_id"], row["pair_id"])
                require(label in ("human", "ai"), f"Unexpected test label on line {line_number}")
                if label == "human":
                    require(label not in pairs[key], f"Duplicate human row on line {line_number}")
                    pairs[key][label] = row
                else:
                    require(row.get("generator") and row.get("prompt"),
                            f"Missing AI variant on line {line_number}")
                    variant = (row["generator"], row["prompt"])
                    variants = pairs[key].setdefault("ai", {})
                    require(variant not in variants,
                            f"Duplicate AI variant on line {line_number}")
                    variants[variant] = row
    return controls, pairs, counts


def choose_rows(controls: dict, pairs: dict) -> tuple[list[dict], str, dict]:
    available_controls = {}
    for paper, rows in controls.items():
        eligible = [row for row in rows if row["negative"] == "unpaired"
                    and source_eligible(row)]
        if eligible:
            available_controls[paper] = min(eligible, key=lambda row: (
                rank(paper, row["pair_id"], "unswell-editlens-control-v1"),
                row["pair_id"]))
    require(len(available_controls) == 400, "Unexpected human-only paper count")
    control_papers = sorted(available_controls, key=lambda paper: (
        rank(paper, "unswell-editlens-calibration-v1"), paper))[:300]
    selected = [{"role": "calibration", "regime": "human", "row": available_controls[paper]}
                for paper in control_papers]

    eligible_pairs: dict[str, dict[tuple[str, str], list[tuple[str, dict, dict]]]] = (
        collections.defaultdict(lambda: collections.defaultdict(list)))
    for (paper, pair), members in pairs.items():
        human = members.get("human")
        if human is None or human.get("negative") != "paired" or not source_eligible(human):
            continue
        for variant, ai in members.get("ai", {}).items():
            if not source_eligible(ai):
                continue
            require(human["paper_id"] == ai["paper_id"] == paper, "Pair crosses papers")
            require(human["pair_id"] == ai["pair_id"] == pair, "Pair ID drift")
            require(human["year"] == ai["year"], "Pair year drift")
            eligible_pairs[paper][variant].append((pair, human, ai))
    require(len(eligible_pairs) == 58, "Unexpected pre-2022 test paper count")
    require(not set(control_papers).intersection(eligible_pairs), "Calibration/test paper overlap")
    selected_metadata = collections.Counter()
    for paper in sorted(eligible_pairs):
        variants = [variant for variant, rows in eligible_pairs[paper].items()
                    if len(rows) >= 5]
        require(variants, f"No eligible variant in paper: {paper}")
        generator, prompt = min(variants, key=lambda variant: (
            rank(paper, variant[0], variant[1], "unswell-editlens-variant-v1"),
            variant[0], variant[1]))
        chosen = sorted(eligible_pairs[paper][(generator, prompt)], key=lambda item: (
            rank(paper, item[0], "unswell-editlens-academic-v1"), item[0]))[:5]
        require(len(chosen) == 5, f"Fewer than five eligible pairs: {paper}")
        selected_metadata[(generator, prompt)] += 1
        for _, human, ai in chosen:
            selected.append({"role": "test", "regime": "human", "row": human})
            selected.append({"role": "test", "regime": "rewrite", "row": ai})
    require(len(selected) == 880, "Unexpected selection size")
    identities = [{
        "role": item["role"], "regime": item["regime"],
        "paper_id": item["row"]["paper_id"], "pair_id": item["row"]["pair_id"],
        "source_sha256": sha_text(item["row"]["text"]),
        "generator": item["row"].get("generator"),
        "prompt": item["row"].get("prompt"),
    } for item in selected]
    selection_hash = sha_text(json.dumps(identities, sort_keys=True,
                                         separators=(",", ":"), ensure_ascii=False))
    metadata = {"calibration_papers": len(control_papers),
                "test_papers": len(eligible_pairs),
                "test_pairs": 5 * len(eligible_pairs),
                "by_variant_papers": {f"{g}|{p}": count for (g, p), count
                                      in sorted(selected_metadata.items())}}
    return selected, selection_hash, metadata


def load_arb_runner():
    spec = importlib.util.spec_from_file_location("frozen_editlens_arb_runner", ARB_RUNNER)
    require(spec is not None and spec.loader is not None, "Frozen runner unavailable")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def metric(rows: list[tuple[dict, dict]], threshold: float, old) -> dict:
    if not rows:
        return {"pairs": 0, "papers": 0, "auc": None, "win_fraction": None,
                "true_positives": 0, "tpr": None}
    values = [(h["score"], a["score"]) for h, a in rows]
    wins = sum(float(a > h) + 0.5 * float(a == h) for h, a in values)
    true_positives = sum(a > threshold for _, a in values)
    return {"pairs": len(rows), "papers": len({h["paper_id"] for h, _ in rows}),
            "auc": old.auc(values), "win_fraction": wins / len(rows),
            "true_positives": true_positives, "tpr": true_positives / len(rows)}


def summarize(prepared: list[dict], selection_hash: str, metadata: dict, old) -> dict:
    calibration = [row for row in prepared if row["role"] == "calibration"
                   and row["score"] is not None and not row["truncated"]]
    require(len(calibration) >= 3, "Insufficient calibration")
    threshold = sorted((row["score"] for row in calibration), reverse=True)[2]
    test_humans = [row for row in prepared if row["role"] == "test"
                   and row["regime"] == "human"
                   and row["score"] is not None and not row["truncated"]]
    false_positives = sum(row["score"] > threshold for row in test_humans)
    pairs: dict[str, dict[str, dict]] = collections.defaultdict(dict)
    for row in prepared:
        if row["role"] == "test" and row["score"] is not None and not row["truncated"]:
            require(row["regime"] not in pairs[row["pair_id"]], "Duplicate scored pair member")
            pairs[row["pair_id"]][row["regime"]] = row
    matched = [(members["human"], members["rewrite"]) for members in pairs.values()
               if "human" in members and "rewrite" in members]
    overall = metric(matched, threshold, old)
    by_generator = {name: metric([(h, a) for h, a in matched if a["generator"] == name],
                                 threshold, old)
                    for name in sorted({a["generator"] for _, a in matched})}
    by_prompt = {name: metric([(h, a) for h, a in matched if a["prompt"] == name],
                              threshold, old)
                 for name in sorted({a["prompt"] for _, a in matched})}
    paper_means = []
    for paper in sorted({h["paper_id"] for h, _ in matched}):
        group = [(h, a) for h, a in matched if h["paper_id"] == paper]
        paper_means.append(sum(a["score"] - h["score"] for h, a in group) / len(group))
    output = {
        "format": "unswell-editlens-academic-transfer-v1",
        "selection_sha256": selection_hash,
        "dataset_sha256": DATA_HASH,
        "model_revision": MODEL_REVISION,
        "model_sha256": MODEL_HASH,
        "selected": metadata,
        "raw_rows": len(prepared),
        "cleaned_below_50": sum(row["score"] is None for row in prepared),
        "truncated": sum(row["truncated"] for row in prepared),
        "eligible_calibration_humans": len(calibration),
        "threshold_strictly_greater": threshold,
        "calibration_false_positives": sum(row["score"] > threshold for row in calibration),
        "eligible_test_humans": len(test_humans),
        "test_human_false_positives": false_positives,
        "test_human_fpr": false_positives / len(test_humans) if test_humans else None,
        "overall": overall,
        "by_generator": by_generator,
        "by_prompt": by_prompt,
        "paper_mean_positive": sum(value > 0 for value in paper_means),
        "paper_mean_count": len(paper_means),
    }
    output["feasibility_pass"] = all((
        len(calibration) >= 285,
        len(test_humans) >= 275,
        overall["pairs"] >= 275,
        output["test_human_fpr"] is not None and output["test_human_fpr"] <= 0.02,
        overall["auc"] is not None and overall["auc"] >= 0.75,
        overall["tpr"] is not None and overall["tpr"] >= 0.20,
        overall["win_fraction"] is not None and overall["win_fraction"] >= 0.65,
    ))
    return output


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--data", type=Path, required=True)
    parser.add_argument("--model-dir", type=Path)
    parser.add_argument("--preflight", action="store_true")
    args = parser.parse_args()
    require(sha_file(args.data) == DATA_HASH, "Dataset hash mismatch")
    controls, pairs, counts = read_source(args.data)
    selected, selection_hash, metadata = choose_rows(controls, pairs)
    if args.preflight:
        print(json.dumps({"selection_sha256": selection_hash, "selected": metadata,
                          "dataset_rows": {f"{s}|{l}": count for (s, l), count
                                           in sorted(counts.items())}}, sort_keys=True))
        return
    require(args.model_dir is not None, "Model directory required")
    freeze = json.loads((HERE / "freeze.json").read_text())
    for filename, key in (("PROTOCOL.md", "protocol_sha256"),
                          ("run.py", "runner_sha256")):
        require(sha_file(HERE / filename) == freeze[key], f"{filename} changed")
    require(sha_file(ARB_RUNNER) == freeze["arb_runner_sha256"], "Frozen inference changed")
    require(selection_hash == freeze["selection_sha256"], "Selection changed")
    model_path = args.model_dir / "onnx/model.onnx"
    require(sha_file(model_path) == MODEL_HASH, "Model hash mismatch")
    for name, digest in freeze["model_file_sha256"].items():
        require(sha_file(args.model_dir / name) == digest, f"Model file changed: {name}")
    old = load_arb_runner()
    tokenizer = old.Tokenizer.from_file(str(args.model_dir / "tokenizer.json"))
    mapped = []
    for item in selected:
        row = item["row"]
        mapped.append({"role": item["role"], "regime": item["regime"],
                       "row": {"text": row["text"], "text_sha256": sha_text(row["text"]),
                               "source_dataset": "academic_paper", "pair_id": row["pair_id"],
                               "generator_model": row.get("generator")}})
    prepared = old.prepare(mapped, tokenizer)
    for item, row in zip(prepared, selected, strict=True):
        item.update({"paper_id": row["row"]["paper_id"],
                     "year": row["row"]["year"], "venue": row["row"]["venue"],
                     "prompt": row["row"].get("prompt")})
    old.score(prepared, model_path)
    result = summarize(prepared, selection_hash, metadata, old)
    (HERE / "result.json").write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    scores = [{key: value for key, value in row.items() if key != "ids"}
              for row in prepared]
    (HERE / "scores.json").write_text(json.dumps(scores, indent=2, sort_keys=True) + "\n")
    print(json.dumps({"feasibility_pass": result["feasibility_pass"],
                      "overall": result["overall"],
                      "test_human_false_positives": result["test_human_false_positives"],
                      "eligible_test_humans": result["eligible_test_humans"]},
                     sort_keys=True))


if __name__ == "__main__":
    main()
