#!/usr/bin/env python3
"""Verify the frozen aggregate of current-engine event reachability."""

from __future__ import annotations

import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
RESULT_SHA256 = "f2a0a1e8ec388d4402cdfb2cffe610147318825af443fd25d7016eb81db88dad"
ARCHIVE_SHA256 = "4f0051badc4b16dfe088d88cae69bd0dda89af2e9f62b4bed317309c4db65082"
MAIN_COMMIT = "52b6612622627c198a80da4d78244b13c54d3f02"
MAIN_TREE = "b78ad9a22aba1c59eab04f7fad6dc524eaed899f"
EXPORTER_SHA256 = "37b6f5ebb922ef5cfca069f9c791a294c2cd8b6ea0d708366fa20b5120d158c4"
COUNTS = {
    "whole": (57, 46, 57, 57, 57, 47, 0, 9),
    "context": (46, 37, 46, 46, 46, 44, 0, 2),
    "confirmation": (20, 16, 20, 20, 20, 19, 0, 1),
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def metrics(record: dict) -> tuple[int, ...]:
    return tuple(record[key] for key in
                 ("events", "required_full", "available", "any_target",
                  "all_targets_some", "all_targets_full", "zero_target",
                  "protected_target"))


def verify(result: dict, freeze: dict) -> dict:
    require(freeze["format"] == "unswell-349-event-reachability-freeze-v1" and
            freeze["status"] == "frozen_before_audit", "freeze status changed")
    require(result["format"] == "unswell-349-event-reachability-audit-v1" and
            result["scope"] == "exposed_offline_source_availability_only" and
            result["qualified"] is False, "scope or qualification changed")
    require(freeze["main_commit"] == result["main_commit"] == MAIN_COMMIT and
            freeze["main_tree"] == result["main_tree"] == MAIN_TREE and
            freeze["archive_sha256"] == result["archive_sha256"] == ARCHIVE_SHA256
            and freeze["exporter_sha256"] == result["exporter_sha256"] ==
            EXPORTER_SHA256, "frozen source identity changed")
    require(result["protocol_sha256"] == freeze["protocol_sha256"] and
            result["script_sha256"] == freeze["script_sha256"],
            "frozen method identity changed")
    require(result["pages"] == result["available_pages"] == 36 and
            result["page_error_categories"] == {} and
            result["exported_units"] == 7625, "page availability changed")
    require(set(result["cohorts"]) == set(COUNTS), "cohorts changed")
    for cohort, expected in COUNTS.items():
        record = result["cohorts"][cohort]
        require(set(record) == {"events", "required_full", "available",
                                "any_target", "all_targets_some",
                                "all_targets_full", "zero_target",
                                "protected_target"}, "cohort fields changed")
        require(all(type(value) is int for value in record.values()),
                "noninteger event count")
        require(metrics(record) == expected, "event reachability changed: " + cohort)
        require(0 <= record["all_targets_full"] <= record["all_targets_some"] <=
                record["any_target"] <= record["available"] <= record["events"]
                and record["any_target"] + record["zero_target"] ==
                record["available"] and
                record["any_target"] >= record["required_full"],
                "impossible reachability accounting")
    return {"available_pages": 36, "zero_target_events": 0,
            "qualified": False}


def verify_public_files() -> dict:
    freeze = json.loads((HERE / "freeze.json").read_text())
    require(freeze["protocol_sha256"] == digest(HERE / "PROTOCOL.md") and
            freeze["script_sha256"] == digest(HERE / "audit.py"),
            "frozen code or protocol changed")
    require(digest(HERE / "result.json") == RESULT_SHA256,
            "saved result identity changed")
    return verify(json.loads((HERE / "result.json").read_text()), freeze)


if __name__ == "__main__":
    print(verify_public_files())
