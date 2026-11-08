#!/usr/bin/env python3
"""Run the frozen, offline paired technical-prose origin-feature screen."""

from __future__ import annotations

import argparse
import collections
import hashlib
import importlib.util
import json
import math
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]
FEATURE = ROOT / "research/origin/runs/2026-11-07-richness-screen/run.py"
EXPECTED = {
    ("train", "human"): 31020,
    ("train", "ai"): 56203,
    ("val", "human"): 4139,
    ("val", "ai"): 6603,
    ("test", "human"): 4300,
    ("test", "ai"): 7015,
    ("human_only", "human"): 21013,
}
GENERATORS = (
    "claude-4-8-opus-genai", "claude-5-1-fable-genai",
    "claude-5-5-opus-genai", "gemini-3-6-flash-genai",
    "gemini-3-8-flash-genai", "gpt-5-6-sol-genai-responses",
    "gpt-6-astra-genai",
)
PROMPTS = ("arb", "imbd", "shield", "v2")
FEATURES = ("richness", "entropy")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def sha(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def load_feature_module():
    spec = importlib.util.spec_from_file_location("unswell_origin_features", FEATURE)
    require(spec is not None and spec.loader is not None, "Feature source unavailable")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def score(text: str, feature_module) -> dict | None:
    words = [m.group().casefold()
             for m in feature_module.WORD.finditer(feature_module.prose(text))]
    if len(words) < 100:
        return None
    values = feature_module.features(" ".join(words[:100]))
    require(values["prose_tokens"] == 100, "Feature token count changed")
    require(all(values[f] is not None for f in FEATURES), "Feature unavailable")
    return {f: values[f] for f in FEATURES}


def auc(pairs: list[dict], feature: str) -> float | None:
    if len(pairs) < 20:
        return None
    values = []
    for pair in pairs:
        values.extend(((pair["ai"][feature], 1), (pair["human"][feature], 0)))
    values.sort()
    wins = 0.0
    negatives_seen = 0
    at = 0
    while at < len(values):
        end = at + 1
        while end < len(values) and values[end][0] == values[at][0]:
            end += 1
        positives = sum(label for _, label in values[at:end])
        negatives = end - at - positives
        wins += positives * negatives_seen + positives * negatives / 2
        negatives_seen += negatives
        at = end
    n = len(pairs)
    return wins / (n * n)


def paired_summary(pairs: list[dict], feature: str, threshold: float) -> dict:
    count = len(pairs)
    wins = sum(
        (pair["ai"][feature] > pair["human"][feature])
        + 0.5 * (pair["ai"][feature] == pair["human"][feature])
        for pair in pairs
    )
    return {
        "eligible_pairs": count,
        "papers": len({pair["paper_id"] for pair in pairs}),
        "paired_win_fraction": wins / count if count else None,
        "pooled_auroc": auc(pairs, feature),
        "tpr_at_calibrated_threshold":
            sum(pair["ai"][feature] > threshold for pair in pairs) / count
            if count else None,
    }


def calibrated_threshold(rows: list[dict], feature: str) -> dict:
    scores = sorted((r[feature] for r in rows if r is not None), reverse=True)
    require(len(scores) >= 100, "Insufficient calibration rows")
    rank = math.ceil(0.01 * len(scores))
    threshold = scores[rank - 1]
    return {
        "eligible": len(scores),
        "rank": rank,
        "threshold": threshold,
        "calibration_fpr": sum(value > threshold for value in scores) / len(scores),
    }


def negative_summary(rows: list[dict | None], feature: str, threshold: float) -> dict:
    eligible = [row for row in rows if row is not None]
    return {
        "total": len(rows),
        "eligible": len(eligible),
        "abstained": len(rows) - len(eligible),
        "fpr": sum(row[feature] > threshold for row in eligible) / len(eligible)
               if eligible else None,
    }


def run(source: Path) -> dict:
    freeze = json.loads((HERE / "freeze.json").read_text())
    for name, path in (
        ("protocol_sha256", HERE / "PROTOCOL.md"),
        ("script_sha256", Path(__file__)),
        ("feature_sha256", FEATURE),
        ("dataset_sha256", source),
    ):
        require(sha(path) == freeze[name], f"{name} mismatch")
    require(source.stat().st_size == freeze["dataset_bytes"], "Dataset size mismatch")
    feature_module = load_feature_module()

    counts = collections.Counter()
    papers = collections.defaultdict(set)
    human_ids = collections.defaultdict(set)
    variants = set()
    val_human = []
    test_human = {}
    test_ai = []
    human_only = []
    ai_ids = collections.defaultdict(set)
    for number, line in enumerate(source.open(encoding="utf-8"), 1):
        row = json.loads(line)
        split, label = row["split"], row["label"]
        require((split, label) in EXPECTED, f"Unknown label or split at row {number}")
        require(isinstance(row["text"], str) and row["text"], f"Missing text at row {number}")
        require(isinstance(row["paper_id"], str) and row["paper_id"],
                f"Missing paper ID at row {number}")
        require(isinstance(row["pair_id"], str) and row["pair_id"],
                f"Missing pair ID at row {number}")
        require(type(row["chunk_ok"]) is bool and type(row["kept_original"]) is bool,
                f"Missing quality flags at row {number}")
        counts[split, label] += 1
        papers[split].add(row["paper_id"])
        if label == "human":
            require(row["generator"] is None and row["prompt"] is None,
                    f"Human row has generator at row {number}")
            require(row["pair_id"] not in human_ids[split],
                    f"Duplicate human pair ID at row {number}")
            human_ids[split].add(row["pair_id"])
            if split == "val":
                val_human.append(score(row["text"], feature_module))
            elif split == "test":
                test_human[row["pair_id"]] = {
                    "paper_id": row["paper_id"],
                    "score": score(row["text"], feature_module),
                }
            elif split == "human_only":
                human_only.append(score(row["text"], feature_module))
        else:
            require(row["generator"] in GENERATORS and row["prompt"] in PROMPTS,
                    f"Unknown AI cell at row {number}")
            require(not row["kept_original"], f"Human text labeled AI at row {number}")
            ai_ids[split].add(row["pair_id"])
            key = (row["pair_id"], row["generator"], row["prompt"])
            require(key not in variants, f"Duplicate AI variant at row {number}")
            variants.add(key)
            if split == "test":
                test_ai.append({
                    "pair_id": row["pair_id"],
                    "paper_id": row["paper_id"],
                    "generator": row["generator"],
                    "prompt": row["prompt"],
                    "chunk_ok": row["chunk_ok"],
                    "score": score(row["text"], feature_module),
                })
    require(counts == EXPECTED, f"Unexpected row counts: {dict(counts)}")
    for split, ids in ai_ids.items():
        require(ids <= human_ids[split], f"AI source missing in {split}")
    for left, left_papers in papers.items():
        for right, right_papers in papers.items():
            if left < right:
                require(not left_papers & right_papers,
                        f"Paper split overlap: {left}, {right}")
    pairs = []
    ai_abstained = 0
    source_abstained = 0
    for row in test_ai:
        source_row = test_human.get(row["pair_id"])
        require(source_row is not None, f"AI source missing: {row['pair_id']}")
        require(source_row["paper_id"] == row["paper_id"], "AI source paper mismatch")
        if row["score"] is None:
            ai_abstained += 1
        if source_row["score"] is None:
            source_abstained += 1
        if row["score"] is None or source_row["score"] is None:
            continue
        pairs.append({**row, "ai": row["score"], "human": source_row["score"]})

    result = {
        "input": {
            "dataset_commit": freeze["dataset_commit"],
            "dataset_sha256": freeze["dataset_sha256"],
            "feature_sha256": freeze["feature_sha256"],
            "rows": sum(counts.values()),
            "counts": {f"{split}/{label}": count
                       for (split, label), count in sorted(counts.items())},
            "paper_counts": {split: len(group)
                             for split, group in sorted(papers.items())},
        },
        "test_coverage": {
            "ai_total": len(test_ai),
            "ai_abstained": ai_abstained,
            "source_abstained_for_ai_rows": source_abstained,
            "eligible_pairs": len(pairs),
            "eligible_fraction": len(pairs) / len(test_ai),
        },
        "features": {},
        "interpretation": "AI rewriting of paired academic prose; not editorial quality or authorship proof",
    }
    for feature in FEATURES:
        calibration = calibrated_threshold(val_human, feature)
        threshold = calibration["threshold"]
        human_test = negative_summary(
            [value["score"] for value in test_human.values()], feature, threshold
        )
        unpaired = negative_summary(human_only, feature, threshold)
        generators = {
            key: paired_summary([row for row in pairs if row["generator"] == key],
                                feature, threshold)
            for key in GENERATORS
        }
        prompts = {
            key: paired_summary([row for row in pairs if row["prompt"] == key],
                                feature, threshold)
            for key in PROMPTS
        }
        overall = paired_summary(pairs, feature, threshold)
        clean = paired_summary(
            [row for row in pairs if row["chunk_ok"]], feature, threshold
        )
        result["features"][feature] = {
            "calibration": calibration,
            "test_paired_human": human_test,
            "human_only": unpaired,
            "overall": overall,
            "chunk_ok_true": clean,
            "by_generator": generators,
            "by_prompt": prompts,
        }
    richness = result["features"]["richness"]
    result["richness_feasibility_pass"] = all((
        result["test_coverage"]["eligible_fraction"] >= 0.8,
        richness["overall"]["pooled_auroc"] is not None
        and richness["overall"]["pooled_auroc"] >= 0.65,
        richness["overall"]["tpr_at_calibrated_threshold"] is not None
        and richness["overall"]["tpr_at_calibrated_threshold"] >= 0.10,
        richness["test_paired_human"]["fpr"] is not None
        and richness["test_paired_human"]["fpr"] <= 0.02,
        richness["human_only"]["fpr"] is not None
        and richness["human_only"]["fpr"] <= 0.02,
        sum(value["pooled_auroc"] is not None and value["pooled_auroc"] > 0.60
            for value in richness["by_generator"].values()) >= 5,
        sum(value["pooled_auroc"] is not None and value["pooled_auroc"] > 0.60
            for value in richness["by_prompt"].values()) >= 3,
    ))
    return result


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--output", type=Path, default=HERE / "result.json")
    args = parser.parse_args()
    result = run(args.input)
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    print(f"Wrote {args.output}; richness feasibility: "
          f"{'passed' if result['richness_feasibility_pass'] else 'failed'}")


if __name__ == "__main__":
    main()
