#!/usr/bin/env python3
"""Independently verify source selection and recorded EditLens accounting."""

import argparse
import collections
import hashlib
import json
import math
import re
from pathlib import Path

HERE = Path(__file__).resolve().parent
WORDS = re.compile(r"\b\w+\b")


def need(ok, message):
    if not ok:
        raise ValueError(message)


def sha(data):
    return hashlib.sha256(data).hexdigest()


def file_sha(path):
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1048576), b""):
            digest.update(block)
    return digest.hexdigest()


def rank(*parts):
    return sha("\0".join(parts).encode())


def eligible(row):
    return row["chunk_ok"] is True and row["kept_original"] is False and 50 <= len(WORDS.findall(row["text"])) <= 300


def selection(data):
    controls = collections.defaultdict(list)
    test = collections.defaultdict(lambda: {"human": {}, "ai": collections.defaultdict(dict)})
    with data.open(encoding="utf-8") as stream:
        for line in stream:
            row = json.loads(line)
            if row["split"] == "human_only" and row["label"] == "human" and row["year"] <= 2018 and eligible(row) and row["negative"] == "unpaired":
                controls[row["paper_id"]].append(row)
            elif row["split"] == "test" and row["year"] <= 2021 and eligible(row):
                paper, pair = row["paper_id"], row["pair_id"]
                if row["label"] == "human" and row["negative"] == "paired":
                    need(pair not in test[paper]["human"], "Duplicate human row")
                    test[paper]["human"][pair] = row
                elif row["label"] == "ai":
                    variant = row["generator"], row["prompt"]
                    need(pair not in test[paper]["ai"][variant], "Duplicate AI variant")
                    test[paper]["ai"][variant][pair] = row
    need(len(controls) == 400, "Control-paper count changed")
    controls = {paper: min(rows, key=lambda row: (rank(paper, row["pair_id"], "unswell-editlens-control-v1"), row["pair_id"])) for paper, rows in controls.items()}
    chosen_controls = sorted(controls, key=lambda paper: (rank(paper, "unswell-editlens-calibration-v1"), paper))[:300]
    selected = [("calibration", "human", controls[paper]) for paper in chosen_controls]
    variant_counts = collections.Counter()
    paper_count = 0
    for paper in sorted(test):
        options = []
        for variant, ai_rows in test[paper]["ai"].items():
            paired_ids = set(ai_rows) & set(test[paper]["human"])
            if len(paired_ids) >= 5:
                options.append((variant, paired_ids))
        if not options:
            continue
        need(paper not in chosen_controls, "Split leakage")
        variant, paired_ids = min(options, key=lambda item: (rank(paper, item[0][0], item[0][1], "unswell-editlens-variant-v1"), item[0]))
        variant_counts["|".join(variant)] += 1
        paper_count += 1
        for pair in sorted(paired_ids, key=lambda pair: (rank(paper, pair, "unswell-editlens-academic-v1"), pair))[:5]:
            selected.append(("test", "human", test[paper]["human"][pair]))
            selected.append(("test", "rewrite", test[paper]["ai"][variant][pair]))
    need(paper_count == 58 and len(selected) == 880, "Selected denominator changed")
    identities = [{"role": role, "regime": regime, "paper_id": row["paper_id"], "pair_id": row["pair_id"], "source_sha256": sha(row["text"].encode()), "generator": row["generator"], "prompt": row["prompt"]} for role, regime, row in selected]
    digest = sha(json.dumps(identities, sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode())
    metadata = {"calibration_papers": 300, "test_papers": 58, "test_pairs": 290, "by_variant_papers": dict(sorted(variant_counts.items()))}
    return identities, digest, metadata


def auc(pairs):
    if len(pairs) < 20:
        return None
    ranked = sorted([(h["score"], 0) for h, _ in pairs] + [(a["score"], 1) for _, a in pairs])
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
    return wins / (len(pairs) ** 2)


def metric(pairs, threshold):
    if not pairs:
        return {"pairs": 0, "papers": 0, "auc": None, "win_fraction": None, "true_positives": 0, "tpr": None}
    wins = sum(float(a["score"] > h["score"]) + 0.5 * float(a["score"] == h["score"]) for h, a in pairs)
    tp = sum(a["score"] > threshold for _, a in pairs)
    return {"pairs": len(pairs), "papers": len({h["paper_id"] for h, _ in pairs}), "auc": auc(pairs), "win_fraction": wins / len(pairs), "true_positives": tp, "tpr": tp / len(pairs)}


def compare(actual, expected, path="result"):
    if isinstance(expected, dict):
        need(isinstance(actual, dict) and set(actual) == set(expected), path + ": keys differ")
        for key, value in expected.items():
            compare(actual[key], value, path + "." + key)
    elif isinstance(expected, float):
        need(isinstance(actual, (int, float)) and math.isclose(actual, expected, rel_tol=1e-12, abs_tol=1e-12), path + ": value differs")
    else:
        need(actual == expected, path + ": value differs")


def verify(data, scores_path, result_path):
    freeze = json.loads((HERE / "freeze.json").read_text())
    need(file_sha(data) == freeze["data_sha256"], "Dataset changed")
    need(file_sha(HERE / "PROTOCOL.md") == freeze["protocol_sha256"], "Protocol changed")
    need(file_sha(HERE / "run.py") == freeze["runner_sha256"], "Runner changed")
    need(file_sha(HERE.parent / "2026-11-09-editlens-transfer" / "run.py") == freeze["arb_runner_sha256"], "Inference helper changed")
    identities, digest, metadata = selection(data)
    need(digest == freeze["selection_sha256"], "Selection changed")
    scores = json.loads(scores_path.read_text())
    need(len(scores) == len(identities), "Score count changed")
    for index, (row, identity) in enumerate(zip(scores, identities, strict=True)):
        for score_key, identity_key in (("role", "role"), ("regime", "regime"), ("paper_id", "paper_id"), ("pair_id", "pair_id"), ("text_sha256", "source_sha256"), ("generator", "generator"), ("prompt", "prompt")):
            need(row.get(score_key) == identity[identity_key], f"Identity mismatch: {index}:{score_key}")
        need("text" not in row and "ids" not in row, "Raw text leaked")
        need(row["truncated"] == (row["token_count"] > 512), "Truncation mismatch")
        need(row["score"] is None or (math.isfinite(row["score"]) and 0 <= row["score"] <= 1), "Invalid score")
    usable = [row for row in scores if row["score"] is not None and not row["truncated"]]
    calibration = [row for row in usable if row["role"] == "calibration"]
    threshold = sorted((row["score"] for row in calibration), reverse=True)[2]
    humans = [row for row in usable if row["role"] == "test" and row["regime"] == "human"]
    by_pair = collections.defaultdict(dict)
    for row in usable:
        if row["role"] == "test":
            need(row["regime"] not in by_pair[row["pair_id"]], "Duplicate scored member")
            by_pair[row["pair_id"]][row["regime"]] = row
    pairs = [(items["human"], items["rewrite"]) for items in by_pair.values() if {"human", "rewrite"} <= items.keys()]
    overall = metric(pairs, threshold)
    fp = sum(row["score"] > threshold for row in humans)
    papers = {h["paper_id"] for h, _ in pairs}
    expected = {
        "format": "unswell-editlens-academic-transfer-v1",
        "selection_sha256": digest, "dataset_sha256": freeze["data_sha256"],
        "model_revision": freeze["model_revision"], "model_sha256": freeze["model_sha256"],
        "selected": metadata, "raw_rows": len(scores),
        "cleaned_below_50": sum(row["score"] is None for row in scores),
        "truncated": sum(row["truncated"] for row in scores),
        "eligible_calibration_humans": len(calibration),
        "threshold_strictly_greater": threshold,
        "calibration_false_positives": sum(row["score"] > threshold for row in calibration),
        "eligible_test_humans": len(humans), "test_human_false_positives": fp,
        "test_human_fpr": fp / len(humans) if humans else None,
        "overall": overall,
        "by_generator": {name: metric([(h, a) for h, a in pairs if a["generator"] == name], threshold) for name in sorted({a["generator"] for _, a in pairs})},
        "by_prompt": {name: metric([(h, a) for h, a in pairs if a["prompt"] == name], threshold) for name in sorted({a["prompt"] for _, a in pairs})},
        "paper_mean_positive": sum(sum(a["score"] - h["score"] for h, a in pairs if h["paper_id"] == paper) > 0 for paper in papers),
        "paper_mean_count": len(papers),
    }
    expected["feasibility_pass"] = all((len(calibration) >= 285, len(humans) >= 275, overall["pairs"] >= 275, expected["test_human_fpr"] <= 0.02, overall["auc"] >= 0.75, overall["tpr"] >= 0.20, overall["win_fraction"] >= 0.65))
    compare(json.loads(result_path.read_text()), expected)
    return {"rows": len(scores), "papers": len(papers), "pairs": len(pairs), "false_positives": fp, "true_positives": overall["true_positives"], "feasibility_pass": expected["feasibility_pass"]}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--data", type=Path, required=True)
    parser.add_argument("--scores", type=Path, default=HERE / "scores.json")
    parser.add_argument("--result", type=Path, default=HERE / "result.json")
    args = parser.parse_args()
    print(json.dumps(verify(args.data, args.scores, args.result), sort_keys=True))


if __name__ == "__main__":
    main()
