#!/usr/bin/env python3
"""Verify the frozen negative #349 contextual-unit location screen."""

from __future__ import annotations

import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
RESULT_SHA256 = "81e772f8611012b09119b1c25b1da78e0e01da7e0c4ca5ba8d0673327fb05142"
COHORTS = {"whole": (57, 28, 3, 2),
           "context": (46, 16, 3, 2),
           "confirmation": (20, 4, 1, 0)}
CATEGORIES = {"empty_framing": 2, "unjustified_intensifiers": 1,
              "vague_claims": 1}
POSITIVE_UNITS = {"empty_framing": 62, "wordiness": 36,
                  "vague_claims": 22}


def require(ok: bool, message: str) -> None:
    if not ok:
        raise ValueError(message)


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def verify(result: dict, freeze: dict) -> dict:
    require(freeze["format"] == "unswell-349-contextual-unit-freeze-v1" and
            freeze["status"] == "frozen_before_run", "freeze changed")
    require(result["format"] == "unswell-349-contextual-unit-generator-v1" and
            result["scope"] == "exposed_group_heldout_location_screen_only" and
            result["qualified"] is False and result["screen_passed"] is False and
            result["hosted_model_calls"] == 0, "scope or result changed")
    for key in ("main_commit", "main_tree", "archive_sha256", "exporter_sha256",
                "protocol_sha256", "script_sha256", "ledger_sha256",
                "frontier_sha256", "dependencies"):
        require(result[key] == freeze[key], "frozen identity changed: " + key)
    require((result["pages"], result["groups"], result["events"],
             result["units"], result["selected_units"],
             result["selected_with_modeled_target_overlap"],
             result["newly_covered_events"]) == (36, 13, 123, 7625, 108, 11, 4),
            "frozen aggregate changed")
    require(result["positive_units_by_category"] == POSITIVE_UNITS and
            result["newly_covered_categories"] == CATEGORIES,
            "category accounting changed")
    require(set(result["cohorts"]) == set(COHORTS), "cohort set changed")
    for name, expected in COHORTS.items():
        row = result["cohorts"][name]
        require(set(row) == {"events", "current_full", "location_covered",
                             "previously_nonfull_location_covered"},
                "cohort fields changed")
        require(all(type(value) is int for value in row.values()),
                "noninteger event count")
        actual = tuple(row[key] for key in
                       ("events", "current_full", "location_covered",
                        "previously_nonfull_location_covered"))
        require(actual == expected, "cohort result changed: " + name)
        require(0 <= row["previously_nonfull_location_covered"] <=
                row["location_covered"] <= row["events"] and
                row["current_full"] <= row["events"],
                "impossible event accounting")
    require(sum(row["previously_nonfull_location_covered"]
                for row in result["cohorts"].values()) ==
            result["newly_covered_events"] ==
            sum(result["newly_covered_categories"].values()),
            "gain totals disagree")
    return {"selected_units": 108, "newly_covered_events": 4,
            "screen_passed": False, "qualified": False}


def verify_public_files() -> dict:
    freeze = json.loads((HERE / "freeze.json").read_text())
    require(sha(HERE / "PROTOCOL.md") == freeze["protocol_sha256"] and
            sha(HERE / "run.py") == freeze["script_sha256"],
            "frozen protocol or method changed")
    require(sha(HERE / "result.json") == RESULT_SHA256,
            "saved result changed")
    return verify(json.loads((HERE / "result.json").read_text()), freeze)


if __name__ == "__main__":
    print(verify_public_files())
