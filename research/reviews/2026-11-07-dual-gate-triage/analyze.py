#!/usr/bin/env python3
"""Recompute the exposed #349 recall and accepted-delivery bottlenecks."""

import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
REVIEW = HERE.parent / "2026-11-07-current-main-editorial-review"
COHORTS = ("whole", "context", "confirmation")
PROFILES = ("technical", "strict")


def require(condition, message):
    if not condition:
        raise ValueError(message)


def load_inputs():
    freeze = json.loads((REVIEW / "freeze.json").read_text())
    result = {}
    for name in ("events.json", "findings.json"):
        path = REVIEW / name
        expected = freeze["inputs"][str(path.relative_to(HERE.parents[2]))]
        require(hashlib.sha256(path.read_bytes()).hexdigest() == expected,
                "review input changed: " + name)
        result[name] = json.loads(path.read_text())
    result["summary.json"] = json.loads((REVIEW / "summary.json").read_text())
    result["freeze.json"] = freeze
    return result


def ceil_div(numerator, denominator):
    return -(-numerator // denominator)


def analyze(inputs):
    events = inputs["events.json"]
    findings = inputs["findings.json"]
    summary = inputs["summary.json"]
    require(len(events) == 123 and len(findings) == 176,
            "review must contain all 123 events and 176 strict findings")
    require(len({row["event"] for row in events}) == len(events), "duplicate event")
    require(len({row["id"] for row in findings}) == len(findings), "duplicate finding")
    require({row["path"].split("/", 1)[0] for row in events + findings} == set(COHORTS),
            "unknown or missing cohort")
    output = {
        "format": "unswell-dual-gate-triage-v1",
        "source_review_commit": inputs["freeze.json"]["runtime_commit"],
        "interpretation": "Arithmetic on exposed assistant judgments, not a new detector result.",
        "profiles": {},
    }
    for profile in PROFILES:
        cohorts = {}
        for cohort in COHORTS:
            reference = summary["profiles"][profile]["by_cohort"][cohort]
            selected_events = [row for row in events if row["event"].startswith(cohort + "/")]
            selected_findings = [row for row in findings
                                 if row["path"].startswith(cohort + "/") and profile in row["profiles"]]
            event_counts = Counter(row["status"] for row in selected_events)
            finding_counts = Counter(row["status"] for row in selected_findings)
            full = event_counts["full"]
            required = reference["events"]["required_full_80pct"]
            total = len(selected_findings)
            accepted = finding_counts["actionable"]
            require(len(selected_events) == reference["events"]["total"] and
                    full == reference["events"]["full"], "event summary mismatch")
            require(total == reference["findings"]["total"] and
                    accepted == reference["findings"]["accepted"], "finding summary mismatch")
            require(required == ceil_div(4 * len(selected_events), 5), "recall gate changed")
            require(reference["findings"]["required_accepted_85pct"] ==
                    ceil_div(17 * total, 20), "acceptance gate changed")
            gap = required - full
            require(gap > 0, "expected unresolved recall gap")
            nonfull = [row for row in selected_events if row["status"] != "full"]
            category_counts = Counter(row["category"] for row in nonfull)
            nonaccepted_by_rule = Counter(row["rule_id"] for row in selected_findings
                                          if row["status"] != "actionable")
            accepted_by_rule = Counter(row["rule_id"] for row in selected_findings
                                       if row["status"] == "actionable")
            perfect_additions_alone = max(0, ceil_div(17 * total - 20 * accepted, 3))
            minimum_nonaccepted_removals = max(
                0, total + gap - (20 * (accepted + gap)) // 17)
            cohorts[cohort] = {
                "events": len(selected_events),
                "full": full,
                "required_full": required,
                "new_full_needed": gap,
                "nonfull_by_category": dict(sorted(category_counts.items())),
                "minimum_non_empty_framing_new_full":
                    max(0, gap - category_counts["empty_framing"]),
                "delivered": total,
                "accepted": accepted,
                "nonaccepted": total - accepted,
                "nonaccepted_by_rule": dict(sorted(nonaccepted_by_rule.items())),
                "accepted_by_rule": dict(sorted(accepted_by_rule.items())),
                "perfect_new_findings_needed_without_removal": perfect_additions_alone,
                "minimum_nonaccepted_removals_if_exactly_recall_gap_perfect_findings_added":
                    minimum_nonaccepted_removals,
            }
        output["profiles"][profile] = cohorts
    return output


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    result = json.dumps(analyze(load_inputs()), indent=2) + "\n"
    target = HERE / "result.json"
    if args.check:
        require(target.read_text() == result, "triage result changed")
    else:
        target.write_text(result)
    print("Verified exposed dual-gate triage" if args.check else "Wrote exposed dual-gate triage")


if __name__ == "__main__":
    main()
