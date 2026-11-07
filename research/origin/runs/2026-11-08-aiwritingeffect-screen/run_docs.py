#!/usr/bin/env python3
"""Run the frozen post-result document-level origin-feature sensitivity."""

from __future__ import annotations

import argparse
import collections
import hashlib
import importlib.util
import json
import statistics
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]
FEATURE = ROOT / "research/origin/runs/2026-11-07-richness-screen/run.py"
EXPECTED = {
    ("train", "human"): 627, ("train", "ai"): 1125,
    ("val", "human"): 83, ("val", "ai"): 131,
    ("test", "human"): 89, ("test", "ai"): 143,
    ("human_only", "human"): 400,
}
GENERATORS = (
    "claude-4-8-opus-genai", "claude-5-1-fable-genai",
    "claude-5-5-opus-genai", "gemini-3-6-flash-genai",
    "gemini-3-8-flash-genai", "gpt-5-6-sol-genai-responses",
    "gpt-6-astra-genai",
)
PROMPTS = ("arb", "imbd", "shield", "v2")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def sha(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def features_for(text: str, module) -> dict:
    values = module.features(text)
    words = [match.group().casefold()
             for match in module.WORD.finditer(module.prose(text))]
    require(values["prose_tokens"] == len(words), "Token count mismatch")
    prefix = module.features(" ".join(words[:100])) if len(words) >= 100 else None
    return {
        "full": values if values["richness"] is not None else None,
        "prefix100": prefix,
        "prose_tokens": len(words),
    }


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


def summarize(pairs: list[dict]) -> dict:
    result = {"eligible_pairs": len(pairs),
              "papers": len({row["paper_id"] for row in pairs})}
    for feature in ("richness", "entropy"):
        win_fraction = sum(
            (row["ai"][feature] > row["human"][feature])
            + 0.5 * (row["ai"][feature] == row["human"][feature])
            for row in pairs
        ) / len(pairs) if pairs else None
        result[feature] = {
            "paired_win_fraction": win_fraction,
            "pooled_auroc": auc(pairs, feature),
        }
    result["median_ai_to_human_prose_tokens"] = statistics.median(
        row["ai_prose_tokens"] / row["human_prose_tokens"] for row in pairs
    ) if pairs else None
    return result


def run(source: Path) -> dict:
    freeze = json.loads((HERE / "freeze-v2.json").read_text())
    for name, path in (
        ("protocol_sha256", HERE / "PROTOCOL-v2.md"),
        ("script_sha256", Path(__file__)),
        ("feature_sha256", FEATURE),
        ("dataset_sha256", source),
        ("primary_result_sha256", HERE / "result.json"),
    ):
        require(sha(path) == freeze[name], f"{name} mismatch")
    require(source.stat().st_size == freeze["dataset_bytes"], "Dataset size mismatch")
    spec = importlib.util.spec_from_file_location("unswell_origin_features", FEATURE)
    require(spec is not None and spec.loader is not None, "Feature source unavailable")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)

    counts = collections.Counter()
    papers = collections.defaultdict(set)
    variants = set()
    human = {}
    ai = []
    for number, line in enumerate(source.open(encoding="utf-8"), 1):
        row = json.loads(line)
        split = row["split"]
        is_human = row["variant"] == "human"
        label = "human" if is_human else "ai"
        require((split, label) in EXPECTED, f"Unknown split at row {number}")
        require(isinstance(row["paper_id"], str) and row["paper_id"],
                f"Missing paper ID at row {number}")
        require(isinstance(row["text"], str) and row["text"],
                f"Missing text at row {number}")
        key = (row["paper_id"], row["variant"])
        require(key not in variants, f"Duplicate variant at row {number}")
        variants.add(key)
        if is_human:
            require(row["generator"] is None and row["prompt"] is None,
                    f"Human generator at row {number}")
        else:
            require(row["generator"] in GENERATORS and row["prompt"] in PROMPTS,
                    f"Unknown AI cell at row {number}")
        counts[split, label] += 1
        papers[split].add(row["paper_id"])
        if split != "test":
            continue
        scored = features_for(row["text"], module)
        if is_human:
            human[row["paper_id"]] = scored
        else:
            ai.append({
                "paper_id": row["paper_id"],
                "generator": row["generator"],
                "prompt": row["prompt"],
                "score": scored,
            })
    require(counts == EXPECTED, f"Unexpected row counts: {dict(counts)}")
    for left, left_papers in papers.items():
        for right, right_papers in papers.items():
            if left < right:
                require(not left_papers & right_papers,
                        f"Paper split overlap: {left}, {right}")
    for row in ai:
        require(row["paper_id"] in human, "Missing matched human paper")
    result = {
        "input": {
            "dataset_commit": freeze["dataset_commit"],
            "dataset_sha256": freeze["dataset_sha256"],
            "primary_result_sha256": freeze["primary_result_sha256"],
            "rows": sum(counts.values()),
            "counts": {f"{split}/{label}": count
                       for (split, label), count in sorted(counts.items())},
        },
        "interpretation": "Post-result sensitivity on the same academic-paper sources",
        "measures": {},
    }
    for measure in ("full", "prefix100"):
        pairs = []
        ai_abstained = 0
        source_abstained = 0
        for row in ai:
            source_row = human[row["paper_id"]]
            ai_score = row["score"][measure]
            human_score = source_row[measure]
            if ai_score is None:
                ai_abstained += 1
            if human_score is None:
                source_abstained += 1
            if ai_score is None or human_score is None:
                continue
            pairs.append({
                "paper_id": row["paper_id"],
                "generator": row["generator"],
                "prompt": row["prompt"],
                "ai": ai_score,
                "human": human_score,
                "ai_prose_tokens": row["score"]["prose_tokens"],
                "human_prose_tokens": source_row["prose_tokens"],
            })
        result["measures"][measure] = {
            "coverage": {
                "ai_total": len(ai),
                "ai_abstained": ai_abstained,
                "source_abstained_for_ai_rows": source_abstained,
                "eligible_pairs": len(pairs),
            },
            "overall": summarize(pairs),
            "by_generator": {
                key: summarize([row for row in pairs if row["generator"] == key])
                for key in GENERATORS
            },
            "by_prompt": {
                key: summarize([row for row in pairs if row["prompt"] == key])
                for key in PROMPTS
            },
        }
    return result


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--output", type=Path, default=HERE / "result-v2.json")
    args = parser.parse_args()
    result = run(args.input)
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    print(f"Wrote {args.output}; full-document eligible: "
          f"{result['measures']['full']['coverage']['eligible_pairs']}")


if __name__ == "__main__":
    main()
