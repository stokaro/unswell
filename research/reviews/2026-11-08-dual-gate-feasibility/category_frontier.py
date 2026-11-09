#!/usr/bin/env python3
"""Audit optimistic category coverage against the frozen #349 event gates."""

from __future__ import annotations

import collections
import hashlib
import json
import math
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
INPUTS = {
    "events": (ROOT / "research/reviews/2026-11-07-current-main-editorial-review/events.json",
               "c65f47f7b0a4566f5e42f94d80fe033624c9e54fd67c7bce58e58dc7d8c4b53e"),
    "scoped": (ROOT / "research/reviews/2026-11-07-scoped-degree/result.json",
               "7abb1346cf37b851bd6c978527a9e4649e07bc6323cfe783317497b8584d4158"),
    "quality": (ROOT / "research/reviews/2026-11-07-quality-claim-screen/result.json",
                "e292562947b6b0ee8d534d62bba993d4b0e1643feaa1aafb556368ad5eaa2fd1"),
    "dual_gate": (HERE / "result.json",
                  "2e5dd69d46944afb7bd915f7cc491696c394f89bd8d2b804698494273a59dc84"),
}
ADDED_FULL = {
    "context/c02-d03", "context/c05-d02", "context/c05-d05",
    "whole/p013-d01", "whole/p018-d02", "whole/p019-d02", "whole/p023-d01",
}
COHORTS = {"whole": 57, "context": 46, "confirmation": 20}
STRATEGIES = {
    "empty_framing": ("empty_framing",),
    "empty_framing_and_wordiness": ("empty_framing", "wordiness"),
    "empty_framing_and_vague_claims": ("empty_framing", "vague_claims"),
    "empty_framing_wordiness_and_vague_claims":
        ("empty_framing", "wordiness", "vague_claims"),
}
STATUSES = {"full", "partial", "missed", "uncertain"}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def load_verified(name: str) -> object:
    path, expected = INPUTS[name]
    source = path.read_bytes()
    require(hashlib.sha256(source).hexdigest() == expected,
            f"{name} source hash changed")
    return json.loads(source)


def analyze(events: list[dict], scoped: dict, quality: dict, dual_gate: dict) -> dict:
    require(len(events) == 123, "Expected all 123 frozen events")
    by_id = {event.get("event"): event for event in events}
    require(len(by_id) == len(events), "Duplicate event identity")
    require(ADDED_FULL <= by_id.keys(), "Missing reviewed addition")
    require(all(by_id[name]["status"] == "missed" for name in ADDED_FULL),
            "Reviewed addition was already credited")
    require(scoped["status"] == quality["status"] == "exposed-development",
            "Unexpected review evidence status")
    require(scoped["profiles"]["technical"]["candidate_findings"] == 157 and
            quality["profiles"]["technical"]["candidate_findings"] == 162,
            "Reviewed finding sequence changed")

    rows = {}
    for cohort, expected_count in COHORTS.items():
        cohort_events = [event for event in events
                         if event["event"].startswith(cohort + "/")]
        require(len(cohort_events) == expected_count,
                f"{cohort} denominator changed")
        categories: dict[str, collections.Counter] = collections.defaultdict(collections.Counter)
        for event in cohort_events:
            require(event.get("status") in STATUSES and event.get("category"),
                    "Invalid frozen event")
            status = "full" if event["event"] in ADDED_FULL else event["status"]
            categories[event["category"]][status] += 1
        full = sum(counts["full"] for counts in categories.values())
        required = math.ceil(0.8 * expected_count)
        for profile in ("technical", "strict"):
            reviewed = quality["profiles"][profile]["cohorts"][cohort]
            require((reviewed["events"], reviewed["full"],
                     reviewed["required_full"]) == (expected_count, full, required),
                    f"{cohort} reviewed count mismatch")
        missing = required - full
        require(missing > 0, f"{cohort} is already qualified")
        nonfull = {category: sum(counts.values()) - counts["full"]
                   for category, counts in categories.items()}
        strategies = {}
        for name, selected in STRATEGIES.items():
            available = sum(nonfull.get(category, 0) for category in selected)
            strategies[name] = {
                "categories": list(selected),
                "nonfull_available": available,
                "optimistic_full_ceiling": full + available,
                "recall_gate_reachable": available >= missing,
                "minimum_selected_recovery_fraction":
                    missing / available if available >= missing else None,
            }
        rows[cohort] = {
            "events": expected_count, "full_now": full,
            "full_required": required, "full_gap": missing,
            "categories": {category: {status: counts[status]
                                      for status in sorted(STATUSES)}
                           for category, counts in sorted(categories.items())},
            "strategies": strategies,
        }

    require(dual_gate["source_sha256"] == INPUTS["quality"][1],
            "Dual-gate source changed")
    triage = {}
    for profile in ("technical", "strict"):
        matching = [row for row in dual_gate["rows"]
                    if row["profile"] == profile and row["cohort"] == "whole"]
        require(len(matching) == 1 and
                matching[0]["full_gap"] == rows["whole"]["full_gap"],
                "Dual-gate whole-page row mismatch")
        triage[profile] = {
            "nonaccepted_suppressions_needed": matching[0]["nonaccepted_suppressions_needed"],
            "nonaccepted_available": matching[0]["nonaccepted_available"],
        }
    return {
        "format": "unswell-349-category-frontier-v1",
        "status": "exposed-development-upper-bound",
        "source_sha256": {name: digest for name, (_, digest) in INPUTS.items()},
        "reviewed_additions": sorted(ADDED_FULL),
        "cohorts": rows,
        "optimistic_whole_page_triage_after_perfect_gap_findings": triage,
    }


def main() -> None:
    result = analyze(*(load_verified(name)
                       for name in ("events", "scoped", "quality", "dual_gate")))
    saved = HERE / "category-frontier.json"
    expected = json.dumps(result, indent=2, sort_keys=True) + "\n"
    if not saved.exists() or saved.read_text(encoding="utf-8") != expected:
        raise ValueError("Category frontier result changed")
    print("category frontier verified")


if __name__ == "__main__":
    main()
