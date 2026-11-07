#!/usr/bin/env python3
"""Independently verify the saved span-neighbor ledger and aggregate screen."""

from __future__ import annotations

import collections
import gzip
import hashlib
import json
import math
from pathlib import Path
import statistics


HERE = Path(__file__).resolve().parent
TASKS = ("Argumentative Writing", "Creative Writing", "Explanatory Writing")
EXPECTED_TASK_COUNTS = {"Argumentative Writing": 121, "Creative Writing": 116,
                        "Explanatory Writing": 115}


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def require(value: bool, message: str) -> None:
    if not value:
        raise ValueError(message)


def auc(rows: list[dict], field: str) -> float:
    positive = [r[field] for r in rows if r["assisted"]]
    negative = [r[field] for r in rows if not r["assisted"]]
    require(bool(positive) and bool(negative), "Task lacks a study class")
    wins = sum((p > n) + 0.5 * (p == n) for p in positive for n in negative)
    return wins / (len(positive) * len(negative))


def paired(rows: list[dict], by_task: bool) -> dict:
    groups = collections.defaultdict(list)
    for row in rows:
        key = (row["author_sha256"], row["task"]) if by_task else row["author_sha256"]
        groups[key].append(row)
    counts = collections.Counter()
    for group in groups.values():
        assisted = [r["score"] for r in group if r["assisted"]]
        human = [r["score"] for r in group if not r["assisted"]]
        if assisted and human:
            delta = statistics.mean(assisted) - statistics.mean(human)
            counts["positive" if delta > 0 else "negative" if delta < 0 else "tied"] += 1
    return {key: counts[key] for key in ("positive", "negative", "tied")}


def close(actual: float, expected: float, message: str, tolerance: float = 1e-7) -> None:
    require(math.isfinite(actual) and abs(actual - expected) <= tolerance, message)


def verify(directory: Path = HERE, feature_source: Path | None = None) -> dict:
    freeze = json.loads((directory / "freeze.json").read_text())
    result = json.loads((directory / "result.json").read_text())
    score_bytes = gzip.decompress((directory / "scores.json.gz").read_bytes())
    rows = json.loads(score_bytes)
    require(freeze["protocol_sha256"] == digest(directory / "PROTOCOL.md"),
            "Frozen protocol changed")
    require(freeze["runner_sha256"] == digest(directory / "run.py"),
            "Frozen runner changed")
    feature_source = feature_source or directory.parent / "2026-11-07-richness-screen" / "run.py"
    require(freeze["feature_sha256"] == digest(feature_source),
        "Baseline feature source changed")
    require(result["input_freeze_sha256"] == digest(directory / "freeze.json"),
            "Input freeze changed")
    require(result["scores_sha256"] == hashlib.sha256(score_bytes).hexdigest(),
            "Saved scores changed")
    require(result["format"] == "unswell-span-example-screen-v1", "Wrong result format")
    require(len(rows) == 362 and [r["row_idx"] for r in rows] == list(range(362)),
            "Incomplete or reordered study rows")
    require([r.get("abstain") for r in rows].count("no_same_task_control") == 10,
            "Peer-review abstention inventory changed")
    scored = {r["row_idx"]: r for r in rows if "score" in r}
    require(len(scored) == 352 and not any("abstain" in r for r in scored.values()),
            "Scored row inventory changed")
    for task, count in EXPECTED_TASK_COUNTS.items():
        require(sum(r["task"] == task for r in scored.values()) == count,
                "Task count changed")
    for row in scored.values():
        require(row["task"] in TASKS and type(row["assisted"]) is bool and
                type(row["bot_majority"]) is bool and
                len(row["author_sha256"]) == 64 and
                len(row["span_scores"]) == 4 and len(row["neighbors"]) == 4,
                "Malformed scored row")
        close(row["score"], statistics.mean(row["span_scores"]),
              "Row score disagrees with spans", 1e-12)
        for ordinal, evidence in enumerate(row["neighbors"]):
            require(evidence["span"] == ordinal, "Span order changed")
            means = {}
            for label, assisted in (("assisted", True), ("human_only", False)):
                neighbors = evidence[label]
                require(len(neighbors) == 5 and
                        len({pair[0] for pair in neighbors}) == 5,
                        "Incomplete or repeated neighbors")
                prior = None
                for source, similarity in neighbors:
                    peer = scored.get(source)
                    require(peer is not None and peer["task"] == row["task"] and
                            peer["author_sha256"] != row["author_sha256"] and
                            peer["assisted"] is assisted,
                            "Neighbor violates task, class, or author isolation")
                    require(-1.000001 <= similarity <= 1.000001 and math.isfinite(similarity),
                            "Invalid cosine similarity")
                    if prior is not None:
                        require((-similarity, source) >= prior,
                                "Neighbor order changed")
                    prior = (-similarity, source)
                means[label] = statistics.mean(pair[1] for pair in neighbors)
            close(row["span_scores"][ordinal], means["assisted"] - means["human_only"],
                  "Span score disagrees with neighbors")

    scored_rows = list(scored.values())
    expected_auc = auc(scored_rows, "score")
    expected_richness = auc(scored_rows, "richness")
    close(result["overall"]["span_auc"], expected_auc, "Span AUROC changed", 1e-12)
    close(result["overall"]["richness_auc"], expected_richness,
          "Richness AUROC changed", 1e-12)
    close(result["overall"]["delta_auc"], expected_auc - expected_richness,
          "AUROC delta changed", 1e-12)
    for task in TASKS:
        subset = [r for r in scored_rows if r["task"] == task]
        recorded = result["by_task"][task]
        require(recorded["rows"] == len(subset), "Task coverage changed")
        close(recorded["span_auc"], auc(subset, "score"), "Task span AUROC changed", 1e-12)
        close(recorded["richness_auc"], auc(subset, "richness"),
              "Task richness AUROC changed", 1e-12)
    for name, subset in (
        ("bot_majority_span_auc", [r for r in scored_rows
                                   if not r["assisted"] or r["bot_majority"]]),
        ("user_minority_span_auc", [r for r in scored_rows
                                    if not r["assisted"] or not r["bot_majority"]]),
    ):
        close(result[name], auc(subset, "score"), f"{name} changed", 1e-12)
    author = paired(scored_rows, by_task=False)
    author_task = paired(scored_rows, by_task=True)
    require(result["paired_author"] == author and
            result["paired_author_task"] == author_task,
            "Paired author directions changed")
    require(result["coverage"] == {"total": 362, "scored": 352,
                                   "abstentions": {"no_same_task_control": 10}} and
            result["encoder_spans"] == 1408 and result["overlength_rows"] == 0,
            "Coverage accounting changed")
    non_tied = author_task["positive"] + author_task["negative"]
    screen = (expected_auc >= 0.70 and expected_auc - expected_richness >= 0.05 and
              all(result["by_task"][task]["span_auc"] >= 0.60 for task in TASKS) and
              non_tied > 0 and author_task["positive"] / non_tied >= 0.60)
    require(result["feasibility_pass"] is screen and not screen,
            "Frozen feasibility decision changed")
    return {"rows": len(rows), "scored": len(scored), "span_auc": expected_auc,
            "richness_auc": expected_richness, "feasibility_pass": screen}


if __name__ == "__main__":
    print(json.dumps(verify(), sort_keys=True))
