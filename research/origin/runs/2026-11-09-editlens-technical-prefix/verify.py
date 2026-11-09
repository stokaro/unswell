#!/usr/bin/env python3
"""Verify technical-prefix aggregates from no-text score rows."""

from __future__ import annotations

import collections
import hashlib
import json
import statistics
from pathlib import Path


HERE = Path(__file__).resolve().parent


def verify(root: Path = HERE) -> None:
    freeze = json.loads((root / "freeze.json").read_text())
    for file, key in (("PROTOCOL.md", "protocol_sha256"),
                      ("run.py", "runner_sha256")):
        if hashlib.sha256((root / file).read_bytes()).hexdigest() != freeze[key]:
            raise ValueError(f"{file} changed")
    scores = json.loads((root / "scores.json").read_text())
    pairs = json.loads((root / "pairs.json").read_text())
    result = json.loads((root / "result.json").read_text())
    if result["identities_sha256"] != freeze["identities_sha256"]:
        raise ValueError("Source identity digest mismatch")
    sources = {row["task_id"]: row for row in scores if row["kind"] == "source"}
    responses = [row for row in scores if row["kind"] == "response"]
    if len(scores) != 494 or len(sources) != 38 or len(responses) != 456:
        raise ValueError("Document denominator mismatch")
    for actual, stored, name in (
        (len(sources), result["sources"], "sources"),
        (len(responses), result["responses"], "responses"),
        (sum(row["score"] is not None for row in sources.values()),
         result["eligible_sources"], "eligible sources"),
        (sum(row["score"] is not None for row in responses),
         result["eligible_responses"], "eligible responses"),
        (sum(row["truncated"] for row in sources.values()),
         result["truncated_sources"], "truncated sources"),
        (sum(row["truncated"] for row in responses),
         result["truncated_responses"], "truncated responses"),
        (sum(row["removed_header"] for row in sources.values()),
         result["header_removed_sources"], "removed source headers"),
        (sum(row["removed_header"] for row in responses),
         result["header_removed_responses"], "removed response headers"),
    ):
        if actual != stored:
            raise ValueError(f"Incorrect {name} count")
    for actual, stored, name in (
        (statistics.median(row["prose_fraction"] for row in sources.values()),
         result["median_prose_fraction_sources"], "source prose fraction"),
        (statistics.median(row["prose_fraction"] for row in responses),
         result["median_prose_fraction_responses"], "response prose fraction"),
    ):
        if abs(actual - stored) > 1e-12:
            raise ValueError(f"Incorrect {name}")
    if len({(row["task_id"], row["family"], row["operation"], row["prompt"])
            for row in responses}) != 456:
        raise ValueError("Duplicate response")
    for row in scores:
        if row["score"] is not None and not 0 <= row["score"] <= 1:
            raise ValueError("Score outside [0,1]")
        if (row["score"] is None) != (row["words"] < 50):
            raise ValueError("Eligibility mismatch")
    expected = {}
    for row in responses:
        source = sources[row["task_id"]]
        if source["repository"] != row["repository"]:
            raise ValueError("Repository mismatch")
        if row["score"] is None or source["score"] is None:
            continue
        key = (row["task_id"], row["family"], row["operation"], row["prompt"])
        expected[key] = (row["score"] - source["score"],
                         row["words"] / source["words"])
    if len(expected) != len(pairs):
        raise ValueError("Eligible-pair denominator mismatch")
    for row in pairs:
        key = (row["task_id"], row["family"], row["operation"], row["prompt"])
        delta, ratio = expected.pop(key)
        if abs(delta - row["delta"]) > 1e-12 or abs(ratio - row["word_ratio"]) > 1e-12:
            raise ValueError("Pair score or length mismatch")
    if expected:
        raise ValueError("Missing pairs")
    for operation in ("generate", "polish"):
        rows = [row for row in pairs if row["operation"] == operation]
        report = result["operations"][operation]
        if len(rows) != report["overall"]["eligible_pairs"]:
            raise ValueError("Operation denominator mismatch")
        median = statistics.median(row["delta"] for row in rows)
        if abs(median - report["overall"]["median_delta"]) > 1e-12:
            raise ValueError("Median delta mismatch")
        by_repo = collections.defaultdict(list)
        for row in rows:
            by_repo[row["repository"]].append(row["delta"])
        positive = sum(statistics.mean(values) > 0 for values in by_repo.values())
        if positive != report["overall"]["positive_repositories"]:
            raise ValueError("Repository direction mismatch")
        for family in ("haiku", "luna", "qwen"):
            family_rows = [row for row in rows if row["family"] == family]
            actual = statistics.median(row["delta"] for row in family_rows)
            stored = report["families"][family]["median_delta"]
            if abs(actual - stored) > 1e-12:
                raise ValueError("Family median mismatch")
        passed = (len(rows) >= 217 and positive >= 15 and
                  all(report["families"][family]["median_delta"] > 0
                      for family in ("haiku", "luna", "qwen")))
        if passed != report["direction_pass"]:
            raise ValueError("Direction gate mismatch")
    print("Technical-prefix score and aggregate verification passed")


if __name__ == "__main__":
    verify()
