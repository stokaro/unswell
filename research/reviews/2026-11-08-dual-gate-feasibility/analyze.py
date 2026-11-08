#!/usr/bin/env python3
"""Calculate a fixed, optimistic #349 recall and delivery scenario."""

from __future__ import annotations

import hashlib
import json
from pathlib import Path
import sys


ROOT = Path(__file__).resolve().parents[3]
SOURCE = ROOT / "research/reviews/2026-11-07-quality-claim-screen/result.json"
SOURCE_SHA256 = "e292562947b6b0ee8d534d62bba993d4b0e1643feaa1aafb556368ad5eaa2fd1"
COHORTS = ("whole", "context", "confirmation")
PROFILES = ("technical", "strict")


def analyze(source_bytes: bytes) -> dict:
    if hashlib.sha256(source_bytes).hexdigest() != SOURCE_SHA256:
        raise ValueError("quality-claim result identity changed")
    source = json.loads(source_bytes)
    if source.get("format") != "unswell-quality-claim-screen-v1":
        raise ValueError("unexpected quality-claim result format")

    rows = []
    for profile in PROFILES:
        cohorts = source["profiles"][profile]["cohorts"]
        for cohort in COHORTS:
            row = cohorts[cohort]
            full, needed = row["full"], row["required_full"]
            accepted, delivered = row["accepted"], row["delivered"]
            events = row["events"]
            if not (0 <= full <= needed <= events and 0 <= accepted <= delivered):
                raise ValueError("invalid baseline counts")
            if needed != (4 * events + 4) // 5:
                raise ValueError("FULL target differs from 80 percent")
            if row["required_accepted"] != (17 * delivered + 19) // 20:
                raise ValueError("accepted target differs from 85 percent")

            # One new, accepted finding per additional FULL event is a fixed
            # comparison scenario, not a claim that these findings exist.
            gap = needed - full
            projected_accepted = accepted + gap
            projected_delivered = delivered + gap
            removable = delivered - accepted
            suppress = max(0, projected_delivered - (20 * projected_accepted) // 17)
            if suppress > removable:
                raise ValueError("scenario cannot reach the acceptance target")
            rows.append({
                "profile": profile,
                "cohort": cohort,
                "full_now": full,
                "full_required": needed,
                "full_gap": gap,
                "accepted_now": accepted,
                "delivered_now": delivered,
                "accepted_after_gap_additions": projected_accepted,
                "delivered_after_gap_additions": projected_delivered,
                "nonaccepted_suppressions_needed": suppress,
                "nonaccepted_available": removable,
            })

    return {
        "format": "unswell-dual-gate-feasibility-v1",
        "source_sha256": SOURCE_SHA256,
        "scenario": "exactly_one_accepted_finding_per_missing_full_event_no_prior_finding_lost",
        "rows": rows,
    }


def render(result: dict) -> bytes:
    return (json.dumps(result, indent=2) + "\n").encode("utf-8")


def verify_result(saved: bytes, source_bytes: bytes) -> None:
    if saved != render(analyze(source_bytes)):
        raise ValueError("dual-gate result differs from pinned calculation")


def main() -> None:
    if len(sys.argv) > 2 or (len(sys.argv) == 2 and sys.argv[1] != "--check"):
        raise SystemExit("usage: analyze.py [--check]")
    result = analyze(SOURCE.read_bytes())
    output = Path(__file__).with_name("result.json")
    if len(sys.argv) == 2:
        verify_result(output.read_bytes(), SOURCE.read_bytes())
    else:
        output.write_bytes(render(result))


if __name__ == "__main__":
    main()
