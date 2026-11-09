#!/usr/bin/env python3
"""Verify the public sentence-shape result without the private source archive."""

import collections
import hashlib
import json
from pathlib import Path

HERE = Path(__file__).resolve().parent
RESULT_SHA256 = "e8fafc2468037626c68774b2ab5e94f33f9bec5f5acd34593511dd6aa9248e8d"
COHORTS = {"whole": 57, "context": 46, "confirmation": 20}
CURRENT_FULL = {"whole": 28, "context": 16, "confirmation": 4}
PATTERNS = {"significance-tail", "contrastive-def", "contrastive-coda", "negation-frame", "appositive-tail"}


def sha256(data):
    return hashlib.sha256(data).hexdigest()


def verify_data(result, freeze):
    if result.get("format") != "unswell-sentence-shape-reachability-v1":
        raise ValueError("wrong result format")
    if result.get("status") != "exposed-exploratory-location-only":
        raise ValueError("exploratory result cannot be promoted")
    if result.get("source_repository_commit") != freeze["source_repository_commit"] or result.get("unswell_main_commit") != freeze["unswell_main_commit"]:
        raise ValueError("source or main identity changed")
    if result.get("input_sha256") != {key: freeze[key] for key in ("archive_sha256", "decisions_sha256", "scoped_result_sha256", "quality_result_sha256")}:
        raise ValueError("frozen input identity changed")
    if result.get("source_pages") != 36 or result.get("frozen_events") != COHORTS or result.get("current_full") != CURRENT_FULL:
        raise ValueError("source or event denominator changed")
    matches = result.get("matches")
    if not isinstance(matches, dict) or set(matches) != PATTERNS:
        raise ValueError("pattern inventory changed")
    eligible_total = outside_total = 0
    for name, row in matches.items():
        fields = ("raw", "excluded_protected", "eligible", "touching_any_frozen_target", "outside_frozen_targets", "events_touched", "previously_non_full_events_touched")
        if set(row) != set(fields) or any(type(row[field]) is not int or row[field] < 0 for field in fields):
            raise ValueError(f"invalid match counters: {name}")
        if row["raw"] != row["excluded_protected"] + row["eligible"]:
            raise ValueError(f"protected match accounting failed: {name}")
        if row["eligible"] != row["touching_any_frozen_target"] + row["outside_frozen_targets"]:
            raise ValueError(f"eligible match accounting failed: {name}")
        if row["events_touched"] > row["touching_any_frozen_target"] or row["previously_non_full_events_touched"] > row["events_touched"]:
            raise ValueError(f"event count exceeds matches: {name}")
        eligible_total += row["eligible"]
        outside_total += row["outside_frozen_targets"]
    if result.get("total_eligible_matches") != eligible_total or result.get("total_matches_outside_frozen_targets") != outside_total:
        raise ValueError("union match totals changed")
    ids = result.get("previously_non_full_event_ids")
    if not isinstance(ids, list) or ids != sorted(set(ids)) or any(not isinstance(item, str) or "/" not in item for item in ids):
        raise ValueError("previously non-FULL event IDs invalid")
    by_cohort = collections.Counter(item.split("/", 1)[0] for item in ids)
    if set(by_cohort) - set(COHORTS) or result.get("previously_non_full_by_cohort") != {cohort: by_cohort[cohort] for cohort in COHORTS}:
        raise ValueError("non-FULL cohort accounting changed")
    if result.get("previously_non_full_events_with_any_overlap") != len(ids):
        raise ValueError("non-FULL event total changed")
    if result.get("events_with_any_target_overlap", 0) < len(ids) or result["events_with_any_target_overlap"] > sum(COHORTS.values()):
        raise ValueError("event overlap total invalid")
    all_touched = result.get("every_target_span_touched_by_cohort")
    if not isinstance(all_touched, dict) or set(all_touched) != set(COHORTS):
        raise ValueError("all-target cohort inventory changed")
    for cohort, value in all_touched.items():
        if type(value) is not int or value < 0 or value > by_cohort[cohort]:
            raise ValueError("all-target count exceeds touched events")
    if result.get("previously_non_full_with_every_target_span_touched") != sum(all_touched.values()):
        raise ValueError("all-target total changed")


def verify_file():
    freeze = json.loads((HERE / "freeze.json").read_text())
    if freeze["format"] != "unswell-sentence-shape-freeze-v2":
        raise ValueError("wrong freeze format")
    for name, file in (("script", "screen.py"), ("protocol", "PROTOCOL.md")):
        if sha256((HERE / file).read_bytes()) != freeze[name + "_sha256"]:
            raise ValueError(f"frozen {name} changed")
    data = (HERE / "result.json").read_bytes()
    if sha256(data) != RESULT_SHA256:
        raise ValueError("result SHA-256 mismatch")
    verify_data(json.loads(data), freeze)
    return RESULT_SHA256


if __name__ == "__main__":
    print(verify_file())
