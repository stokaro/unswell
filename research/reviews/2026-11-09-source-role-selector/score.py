"""Score one fixed source-role selection against exposed archived judgments."""

from __future__ import annotations

import argparse
from collections import Counter
import json
import math
from pathlib import Path

from admit import INPUTS, load_pinned, require, select


REVIEWS = (
    "recovered-admission-bank-v1/reviews.json",
    "d57beb6ee3213ab16c286dd6a0a0bd7ed5adf24f1c67cd40f71d50458ab85ba2",
)
EXPECTED_BASELINE = {
    "whole": {"full": 45, "events": 57, "accepted": 1443, "delivered": 1700, "unresolved_archived_edits": 5},
    "context": {"full": 38, "events": 46, "accepted": 389, "delivered": 413, "unresolved_archived_edits": 0},
    "confirmation": {"full": 17, "events": 20, "accepted": 190, "delivered": 198, "unresolved_archived_edits": 0},
}


def measures(aliases: set[str], evaluation: dict, metadata: dict, originals: dict, reviews: dict) -> dict:
    by_cohort = {cohort: set() for cohort in EXPECTED_BASELINE}
    for alias in aliases:
        page = metadata[alias]["page"]
        cohort = originals["paths"][page].split("/")[0]
        by_cohort[cohort].add(alias)
    result = {}
    for cohort, current in by_cohort.items():
        events = {
            event_id: item
            for event_id, item in evaluation["complete_witnesses"].items()
            if event_id.startswith(cohort + "/")
        }
        full = sum(
            any(set(witness) <= current for witness in item.get("full_groups", []))
            for item in events.values()
        )
        statuses = Counter()
        for alias in current:
            meta = metadata[alias]
            review = reviews[meta["stream"]]["judgments"][meta["finding_id"]]
            statuses[review["status"]] += 1
            if review.get("suggestion_safety", {}).get("status") == "uncertain":
                statuses["unresolved_archived_edits"] += 1
        accepted = statuses["accepted"]
        delivered = len(current)
        result[cohort] = {
            "full": full,
            "events": len(events),
            "accepted": accepted,
            "delivered": delivered,
            "fraction": accepted / delivered if delivered else 0.0,
            "unresolved_archived_edits": statuses["unresolved_archived_edits"],
        }
    return result


def score(development_root: Path) -> dict:
    selection = select(development_root)
    evaluation = load_pinned(development_root / INPUTS["evaluation"][0], INPUTS["evaluation"][1])
    metadata = load_pinned(development_root / INPUTS["metadata"][0], INPUTS["metadata"][1])
    originals = load_pinned(development_root / INPUTS["originals"][0], INPUTS["originals"][1])
    reviews = load_pinned(development_root / REVIEWS[0], REVIEWS[1])
    baseline = measures(set(selection["baseline_aliases"]), evaluation, metadata, originals, reviews)
    for cohort, expected in EXPECTED_BASELINE.items():
        require(
            all(baseline[cohort][key] == value for key, value in expected.items()),
            f"changed baseline: {cohort}",
        )
    candidate = measures(set(selection["selected_aliases"]), evaluation, metadata, originals, reviews)
    numeric_pass = all(
        row["full"] >= math.ceil(0.8 * row["events"]) and row["fraction"] >= 0.85
        for row in candidate.values()
    )
    return {
        "format": "unswell-349-source-role-score-v1",
        "scope": "exposed_cached_bank_only",
        "selection_input_sha256": selection["input_sha256"],
        "policy_sha256": selection["policy_sha256"],
        "reviews_sha256": REVIEWS[1],
        "decision_counts": selection["decision_counts"],
        "baseline": baseline,
        "candidate": candidate,
        "numeric_pass_on_archived_diagnoses": numeric_pass,
        "product_qualified": False,
        "reason_not_qualified": "Cached proposals are not standalone generation; diagnosis-only delivery and advice safety need fresh review and page-disjoint confirmation.",
    }


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("development_root", type=Path)
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    result = score(args.development_root)
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps({"candidate": result["candidate"], "numeric_pass": result["numeric_pass_on_archived_diagnoses"]}, sort_keys=True))


if __name__ == "__main__":
    main()
