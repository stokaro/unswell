#!/usr/bin/env python3
"""Replay the exposed AITDNA screen from a pinned local parquet file.

Research-only dependency: duckdb==1.4.2. Ordinary Unswell builds do not use it.
"""

from __future__ import annotations

import argparse
import collections
import hashlib
import importlib.util
import json
import statistics
from pathlib import Path

import duckdb


HERE = Path(__file__).resolve().parent
FEATURE = HERE.parent / "2026-11-07-richness-screen" / "run.py"


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path: Path):
    return json.loads(path.read_text())


def require(value, message: str) -> None:
    if not value:
        raise ValueError(message)


def auc(group: list[dict], feature: str) -> float | None:
    positive = [r[feature] for r in group if r["assisted"]]
    negative = [r[feature] for r in group if not r["assisted"]]
    if not positive or not negative:
        return None
    wins = sum((p > n) + 0.5 * (p == n)
               for p in positive for n in negative)
    return wins / (len(positive) * len(negative))


def pairs(group: list[dict], feature: str, by_task: bool) -> tuple[dict, list[float]]:
    grouped = collections.defaultdict(list)
    for row in group:
        key = (row["author"], row["task"]) if by_task else row["author"]
        grouped[key].append(row)
    deltas = []
    for rows in grouped.values():
        positive = [r[feature] for r in rows if r["assisted"]]
        negative = [r[feature] for r in rows if not r["assisted"]]
        if positive and negative:
            deltas.append(statistics.mean(positive) - statistics.mean(negative))
    directions = {"positive": sum(d > 0 for d in deltas),
                  "negative": sum(d < 0 for d in deltas),
                  "tied": sum(d == 0 for d in deltas)}
    return directions, deltas


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--parquet", type=Path, required=True)
    args = parser.parse_args()

    freeze = read(HERE / "freeze.json")
    freeze_v2 = read(HERE / "freeze-v2.json")
    acquisition = read(HERE / "acquisition.json")
    provenance = read(HERE / "provenance.json")
    require(digest(args.parquet) == provenance["parquet_sha256"] ==
            freeze_v2["parquet_sha256"], "Parquet hash differs")
    for name, expected in (("PROTOCOL.md", freeze["protocol_sha256"]),
                           ("PROTOCOL-v2.md", freeze_v2["protocol_sha256"]),
                           ("run.py", freeze["script_sha256"]),
                           ("sensitivity.py", freeze_v2["script_sha256"])):
        require(digest(HERE / name) == expected, f"Frozen file changed: {name}")
    require(digest(FEATURE) == freeze["feature_sha256"] ==
            freeze_v2["feature_sha256"], "Feature source changed")
    require(digest(HERE / "result.json") == provenance["result_v1_sha256"] ==
            freeze_v2["v1_result_sha256"], "V1 result changed")
    require(duckdb.__version__ == freeze_v2["duckdb_version"],
            "Parquet reader version differs")

    spec = importlib.util.spec_from_file_location("unswell_richness_screen", FEATURE)
    require(spec is not None and spec.loader is not None, "Feature import unavailable")
    feature_module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(feature_module)
    raw = duckdb.connect().execute(
        "SELECT data, metadata FROM read_parquet(?)", [str(args.parquet)]
    ).fetchall()
    require(len(raw) == 362, "Document count differs")

    scored = []
    snippet_authors = collections.Counter()
    for index, (data, metadata) in enumerate(raw):
        require(len(data) == 1, "Document notion changed")
        part = data[0]
        require(part["author"] in ("Bot", "User"), "Unexpected document label")
        snippet_authors[part["author"]] += 1
        words = [m.group().casefold() for m in
                 feature_module.WORD.finditer(feature_module.prose(part["text"]))]
        require(len(words) >= 100, "New abstention")
        values = feature_module.features(" ".join(words[:100]))
        require(values["prose_tokens"] == 100, "Feature token count differs")
        scored.append({"row_idx": index, "author": metadata["author"],
                       "task": metadata["task"],
                       "assisted": not metadata["human_only"],
                       "snippet_author": part["author"],
                       "prose_tokens": len(words),
                       "bot_word_fraction": float(part["author"] == "Bot"),
                       "richness": values["richness"],
                       "entropy": values["entropy"]})

    def counts(group: list[dict]) -> dict:
        return {"human_only": sum(not r["assisted"] for r in group),
                "ai_assisted": sum(r["assisted"] for r in group)}

    tasks = sorted({r["task"] for r in scored})
    task_results = {}
    for task in tasks:
        group = [r for r in scored if r["task"] == task]
        task_results[task] = {"counts": counts(group),
                              "richness_auc": auc(group, "richness"),
                              "entropy_auc": auc(group, "entropy")}
    overall = {feature + "_auc": auc(scored, feature)
               for feature in ("richness", "entropy")}
    directions = {feature: pairs(scored, feature, False)[0]
                  for feature in ("richness", "entropy")}
    screen = all(overall[f + "_auc"] >= 0.60 and
                 sum(v[f + "_auc"] is not None and v[f + "_auc"] >= 0.60
                     for v in task_results.values()) >= 2 and
                 directions[f]["positive"] >= 0.60 * sum(directions[f].values())
                 for f in ("richness", "entropy"))
    result = {"format": "unswell-aitdna-mixed-authorship-screen-v1",
              "dataset_head_observed": freeze["dataset_head"],
              "source_identity_limit": "Dataset-server pages are hash-frozen but lack commit binding",
              "raw_page_sha256": acquisition["pages"],
              "rows_total": 362, "scored": 362, "abstentions": 0,
              "abstention_rows": [],
              "authors": len({r["author"] for r in scored}),
              "snippet_author_counts": dict(sorted(snippet_authors.items())),
              "counts": counts(scored),
              "median_prose_tokens": {
                  "human_only": statistics.median(r["prose_tokens"] for r in scored
                                                  if not r["assisted"]),
                  "ai_assisted": statistics.median(r["prose_tokens"] for r in scored
                                                   if r["assisted"]),
              },
              "median_bot_word_fraction_assisted": statistics.median(
                  r["bot_word_fraction"] for r in scored if r["assisted"]),
              "overall": overall, "by_task": task_results,
              "paired_author_directions": directions,
              "feasibility_pass": screen,
              "interpretation": "Study-setting contrast, not machine-authorship accuracy or editorial value"}
    require(result == read(HERE / "result.json"), "V1 aggregate differs")

    matched = {}
    split = {}
    for feature in ("richness", "entropy"):
        _, deltas = pairs(scored, feature, True)
        matched[feature] = {"pairs": len(deltas),
                            "positive": sum(d > 0 for d in deltas),
                            "negative": sum(d < 0 for d in deltas),
                            "tied": sum(d == 0 for d in deltas),
                            "median_delta": statistics.median(deltas)}
        negatives = [r[feature] for r in scored if not r["assisted"]]
        split[feature] = {}
        for label in ("Bot", "User"):
            positives = [r[feature] for r in scored
                         if r["assisted"] and r["snippet_author"] == label]
            wins = sum((p > n) + 0.5 * (p == n)
                       for p in positives for n in negatives)
            split[feature][label] = {
                "assisted_count": len(positives),
                "human_only_count": len(negatives),
                "directional_auc": wins / (len(positives) * len(negatives))}
    result_v2 = {"format": "unswell-aitdna-mixed-authorship-sensitivity-v2",
                 "post_result": True, "rows": 362,
                 "human_only_snippet_authors": dict(sorted(collections.Counter(
                     r["snippet_author"] for r in scored if not r["assisted"]).items())),
                 "assisted_snippet_authors": dict(sorted(collections.Counter(
                     r["snippet_author"] for r in scored if r["assisted"]).items())),
                 "matched_author_task": matched,
                 "assisted_by_snippet_author": split,
                 "interpretation": "Exposed sensitivity only, with no new pass gate"}
    require(result_v2 == read(HERE / "result-v2.json"), "V2 aggregate differs")
    print("AITDNA v1 and v2 aggregates match the pinned parquet and feature source")


if __name__ == "__main__":
    main()
