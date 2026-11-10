#!/usr/bin/env python3
"""Independently verify the five-window source selection and paper metrics."""

from __future__ import annotations

import argparse
import collections
import hashlib
import json
import math
from pathlib import Path
import re


HERE = Path(__file__).resolve().parent
PRIOR = HERE.parent / "2026-11-10-editlens-academic-transfer"
DATA_SHA = "6e0eb8058578afce4aaf81c822b70762afd9f18534eb5f1e89aa976cd051b793"
OLD_SCORES_SHA = "6779dff96cbd320c940276cde5629f62ac7a9d7fcf9600961b3386629d91a299"
WORDS = re.compile(r"\b\w+\b")


def need(ok, message):
    if not ok:
        raise ValueError(message)


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def file_sha(path):
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1048576), b""):
            digest.update(block)
    return digest.hexdigest()


def rank(*parts):
    return sha("\0".join(parts).encode())


def source_selection(data):
    controls = collections.defaultdict(list)
    with data.open(encoding="utf-8") as stream:
        for line in stream:
            item = json.loads(line)
            if (item["split"] == "human_only" and item["label"] == "human"
                    and item["year"] <= 2018 and item["negative"] == "unpaired"
                    and item["chunk_ok"] is True
                    and item["kept_original"] is False
                    and 50 <= len(WORDS.findall(item["text"])) <= 300):
                controls[item["paper_id"]].append(item)
    need(len(controls) == 400, "Control-paper count changed")
    ordered = sorted(controls, key=lambda paper: (
        rank(paper, "unswell-editlens-calibration-v1"), paper))
    identities = []
    for index, paper in enumerate(ordered):
        rows = controls[paper]
        need(len({row["pair_id"] for row in rows}) == len(rows),
             "Duplicate source paragraph")
        need(len(rows) >= 5, "Incomplete paper")
        selected = sorted(rows, key=lambda row: (
            rank(paper, row["pair_id"], "unswell-editlens-control-v1"),
            row["pair_id"]))[:5]
        for item in selected:
            identities.append({"role": "calibration" if index < 300 else "holdout",
                               "regime": "human", "paper_id": paper,
                               "pair_id": item["pair_id"],
                               "text_sha256": sha(item["text"].encode())})
    digest = sha(json.dumps(identities, sort_keys=True, separators=(",", ":"),
                            ensure_ascii=False).encode())
    return identities, digest


def paper_means(rows, role, regime):
    grouped = collections.defaultdict(list)
    for item in rows:
        if item["role"] == role and item["regime"] == regime:
            grouped[item["paper_id"]].append(item)
    means = {}
    for paper, items in grouped.items():
        need(len(items) == 5 and len({item["pair_id"] for item in items}) == 5,
             "Missing or duplicate paper window")
        if all(item["score"] is not None and not item["truncated"] for item in items):
            means[paper] = sum(item["score"] for item in items) / 5
    return means


def auc(pairs):
    if len(pairs) < 20:
        return None
    ranked = sorted([(human, 0) for human, _ in pairs]
                    + [(ai, 1) for _, ai in pairs])
    negatives_before = 0
    wins = 0.0
    at = 0
    while at < len(ranked):
        end = at + 1
        while end < len(ranked) and ranked[end][0] == ranked[at][0]:
            end += 1
        positives = sum(label for _, label in ranked[at:end])
        negatives = end - at - positives
        wins += positives * negatives_before + 0.5 * positives * negatives
        negatives_before += negatives
        at = end
    return wins / len(pairs) ** 2


def metric(pairs, threshold):
    need(pairs, "Empty paper group")
    tp = sum(ai > threshold for _, ai in pairs)
    return {"papers": len(pairs), "auc": auc(pairs),
            "win_fraction": sum(float(ai > human) + 0.5 * float(ai == human)
                                for human, ai in pairs) / len(pairs),
            "true_positives": tp, "tpr": tp / len(pairs)}


def equal(actual, expected, path="result"):
    if isinstance(expected, dict):
        need(isinstance(actual, dict) and set(actual) == set(expected),
             f"{path}: keys differ")
        for key, item in expected.items():
            equal(actual[key], item, path + "." + key)
    elif isinstance(expected, float):
        need(isinstance(actual, (float, int)) and
             math.isclose(actual, expected, rel_tol=1e-12, abs_tol=1e-12),
             f"{path}: value differs")
    else:
        need(actual == expected, f"{path}: value differs")


def verify(data, scores_path, result_path):
    freeze = json.loads((HERE / "freeze.json").read_text())
    for path, key in ((data, "data_sha256"),
                      (HERE / "PROTOCOL.md", "protocol_sha256"),
                      (HERE / "run.py", "runner_sha256"),
                      (HERE / "verify.py", "verifier_sha256"),
                      (PRIOR / "run.py", "academic_runner_sha256"),
                      (PRIOR.parent / "2026-11-09-editlens-transfer" / "run.py",
                       "arb_runner_sha256")):
        need(file_sha(path) == freeze[key], f"Frozen input changed: {key}")
    need(file_sha(PRIOR / "scores.json") == OLD_SCORES_SHA,
         "Prior test scores changed")
    identities, selection_hash = source_selection(data)
    need(selection_hash == freeze["selection_sha256"], "Selection changed")
    controls = json.loads(scores_path.read_text())
    need(len(controls) == len(identities) == 2000, "Control count changed")
    for index, (row, identity) in enumerate(zip(controls, identities, strict=True)):
        for key in identity:
            need(row.get(key) == identity[key], f"Control identity changed: {index}:{key}")
        need("text" not in row and "ids" not in row, "Raw text leaked")
        need(row["truncated"] == (row["token_count"] > 512),
             "Truncation mismatch")
        need(row["score"] is None or
             (math.isfinite(row["score"]) and 0 <= row["score"] <= 1),
             "Invalid score")
    tests = [item for item in json.loads((PRIOR / "scores.json").read_text())
             if item["role"] == "test"]
    need(len(tests) == 580, "Prior test count changed")
    cal = paper_means(controls, "calibration", "human")
    holdout = paper_means(controls, "holdout", "human")
    human = paper_means(tests, "test", "human")
    ai = paper_means(tests, "test", "rewrite")
    ids = sorted(set(human) & set(ai))
    need(len(cal) >= 3 and ids, "No eligible paper comparison")
    threshold = sorted(cal.values(), reverse=True)[2]
    pairs = [(human[paper], ai[paper]) for paper in ids]
    by_generator = collections.defaultdict(list)
    for paper in ids:
        names = {row["generator"] for row in tests
                 if row["paper_id"] == paper and row["regime"] == "rewrite"}
        need(len(names) == 1, "Mixed generators")
        by_generator[names.pop()].append((human[paper], ai[paper]))
    overall = metric(pairs, threshold)
    holdout_fp = sum(value > threshold for value in holdout.values())
    paired_fp = sum(human[paper] > threshold for paper in ids)
    expected = {
        "format": "unswell-editlens-five-window-v1",
        "scope": "exposed_academic_paper_aggregation_only",
        "dataset_sha256": DATA_SHA,
        "model_sha256": freeze["model_sha256"],
        "model_revision": freeze["model_revision"],
        "prior_scores_sha256": OLD_SCORES_SHA,
        "prior_selection_sha256": freeze["prior_selection_sha256"],
        "control_selection_sha256": selection_hash,
        "selected": {"calibration_papers": 300, "holdout_papers": 100,
                     "control_windows": 2000, "test_papers": 58,
                     "test_pairs": 290},
        "control_windows_selected": len(controls),
        "control_windows_with_score": sum(row["score"] is not None for row in controls),
        "control_windows_below_50": sum(row["score"] is None for row in controls),
        "control_windows_truncated": sum(row["truncated"] for row in controls),
        "eligible_calibration_papers": len(cal),
        "eligible_holdout_papers": len(holdout),
        "eligible_test_human_papers": len(human),
        "eligible_test_ai_papers": len(ai),
        "eligible_matched_papers": len(ids),
        "threshold_strictly_greater": threshold,
        "calibration_false_positives": sum(value > threshold for value in cal.values()),
        "holdout_false_positives": holdout_fp,
        "holdout_fpr": holdout_fp / len(holdout) if holdout else None,
        "paired_human_false_positives": paired_fp,
        "paired_human_fpr": paired_fp / len(ids),
        "overall": overall,
        "by_generator": {name: metric(values, threshold)
                         for name, values in sorted(by_generator.items())},
    }
    expected["feasibility_pass"] = all((
        len(cal) >= 285, len(holdout) >= 95, len(ids) >= 55,
        expected["holdout_fpr"] is not None and expected["holdout_fpr"] <= 0.02,
        expected["paired_human_fpr"] <= 0.02,
        overall["auc"] is not None and overall["auc"] >= 0.75,
        overall["tpr"] >= 0.20, overall["win_fraction"] >= 0.65,
    ))
    equal(json.loads(result_path.read_text()), expected)
    return {"calibration": len(cal), "holdout": len(holdout),
            "matched": len(ids), "holdout_false_positives": holdout_fp,
            "paired_human_false_positives": paired_fp,
            "true_positives": overall["true_positives"],
            "feasibility_pass": expected["feasibility_pass"]}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--data", type=Path, required=True)
    parser.add_argument("--scores", type=Path, default=HERE / "scores.json")
    parser.add_argument("--result", type=Path, default=HERE / "result.json")
    args = parser.parse_args()
    print(json.dumps(verify(args.data, args.scores, args.result), sort_keys=True))


if __name__ == "__main__":
    main()
