#!/usr/bin/env python3
"""Reconcile the #442 source review with the later #504 replay review."""

from collections import Counter
import gzip
import hashlib
import json
from pathlib import Path
import subprocess
import sys


HERE = Path(__file__).resolve().parent
REVIEWS = HERE.parent
SCOPED = REVIEWS / "2026-11-07-scoped-degree"
CLAIMS = REVIEWS / "2026-11-07-quality-claim-screen"


def require(condition, message):
    if not condition:
        raise ValueError(message)


def read(path):
    raw = path.read_bytes()
    return json.loads(gzip.decompress(raw) if path.suffix == ".gz" else raw)


def location(finding):
    primary = finding["primary"]
    span = primary["span"]
    return finding["rule_id"], primary["path"], span["start"], span["end"]


def diagnostic(finding):
    return {key: value for key, value in finding.items()
            if key not in {"id", "fingerprint", "rule_version"}}


def report_index(report, profile):
    require(report["status"] == "complete" and report["manifest"]["complete"] and
            not report["errors"] and not report["manifest"]["skipped_rules"],
            "incomplete " + profile + " report")
    documents = report["documents"]
    require(len(documents) == 36 and
            len({row["name"] for row in documents}) == 36,
            "source coverage changed: " + profile)
    sources = {row["name"]: (row["source_hash"], row["bytes"])
               for row in documents}
    findings = report["findings"]
    indexed = {location(row): row for row in findings}
    require(len(indexed) == len(findings), "duplicate finding location: " + profile)
    return sources, indexed


def cohort_counts(result, profile, second=False):
    if second:
        by_cohort = result["profiles"][profile]["by_cohort"]
        return {
            cohort: {
                "full": row["events"]["full"],
                "events": row["events"]["total"],
                "accepted": row["findings"]["accepted"],
                "delivered": row["findings"]["total"],
            }
            for cohort, row in by_cohort.items()
        }
    return {
        cohort: {
            "full": row["full"], "events": row["events"],
            "accepted": row["accepted"], "delivered": row["delivered"],
        }
        for cohort, row in result["profiles"][profile]["cohorts"].items()
    }


def evaluate(claim_reports, replay_reports, scoped_result, claim_result,
             replay_result, second_review):
    require(len(second_review) == 7 and
            len({row["id"] for row in second_review}) == 7 and
            len({row["event"] for row in second_review}) == 7,
            "second review is incomplete")
    require(replay_result["runtime_commit"] ==
            "3acba5eb4c13b95207caf238da17e9c0a72630e3",
            "wrong replay runtime")

    same = {}
    replay_by_id = {}
    for profile, expected in (("technical", 162), ("strict", 183)):
        left_sources, left = report_index(claim_reports[profile],
                                          "claim-" + profile)
        right_sources, right = report_index(replay_reports[profile],
                                            "replay-" + profile)
        require(len(left) == len(right) == expected,
                "finding denominator changed: " + profile)
        require(left_sources == right_sources,
                "source identity changed: " + profile)
        require(set(left) == set(right), "finding locations changed: " + profile)
        for key in left:
            require(diagnostic(left[key]) == diagnostic(right[key]),
                    "finding diagnosis changed: " + str(key))
        require(claim_result["profiles"][profile]["candidate_findings"] == expected
                and replay_result["profiles"][profile]["findings"] == expected,
                "result denominator differs from reports")
        same[profile] = {"sources": len(left_sources),
                         "findings": len(left), "added_since_claim_screen": 0}
        if profile == "strict":
            replay_by_id = {row["id"]: row for row in right.values()}

    previous_locations = {
        tuple(row) for profile in ("technical", "strict")
        for row in scoped_result["profiles"][profile]["added_locations"]
    }
    claim_locations = {
        tuple(row) for profile in ("technical", "strict")
        for row in claim_result["profiles"][profile]["added_locations"]
    }
    require(len(previous_locations) == 2 and len(claim_locations) == 5 and
            previous_locations.isdisjoint(claim_locations),
            "prior construction provenance changed")
    reviewed_locations = set()
    for row in second_review:
        finding = replay_by_id.get(row["id"])
        require(finding is not None and row["finding_status"] == "actionable",
                "second review finding identity changed")
        key = location(finding)
        require(row["event"].split("/")[0] == key[1].split("/")[0],
                "second review cohort changed")
        reviewed_locations.add((key[1], key[2], key[3]))
    require(reviewed_locations == previous_locations | claim_locations,
            "second review did not revisit the seven prior additions")

    established = {}
    second = {}
    for profile in ("technical", "strict"):
        established[profile] = cohort_counts(claim_result, profile)
        second[profile] = cohort_counts(replay_result, profile, second=True)
        for cohort in ("whole", "context", "confirmation"):
            old, new = established[profile][cohort], second[profile][cohort]
            require(old["events"] == new["events"] and
                    old["accepted"] == new["accepted"] and
                    old["delivered"] == new["delivered"],
                    "cohort delivery changed: " + profile + "/" + cohort)
        require(established[profile]["whole"]["full"] == 28 and
                second[profile]["whole"]["full"] == 25 and
                established[profile]["context"]["full"] ==
                second[profile]["context"]["full"] == 16 and
                established[profile]["confirmation"]["full"] ==
                second[profile]["confirmation"]["full"] == 4,
                "review disagreement count changed: " + profile)
    disputed = sorted(row["event"] for row in second_review
                      if row["event_status"] == "partial")
    require(disputed == ["whole/p013-d01", "whole/p018-d02", "whole/p023-d01"] and
            Counter(row["event_status"] for row in second_review) ==
            {"full": 4, "partial": 3},
            "review disagreement identities changed")

    return {
        "format": "unswell-349-review-reconciliation-v1",
        "finding_stream": same,
        "earlier_additions": {"scoped_degree": len(previous_locations),
                              "quality_claim": len(claim_locations)},
        "new_detector_findings_in_replay": 0,
        "established_442_review": established,
        "conservative_504_review": second,
        "disputed_events": disputed,
        "interpretation": (
            "PR #504 replayed the same source-bound findings as #442. "
            "The three-event difference is assistant editorial review disagreement, "
            "not a detector gain or regression. Neither review qualifies #349."
        ),
    }


def main():
    for command in (
        [sys.executable, "-B", str(SCOPED / "analyze.py"), "--check"],
        [sys.executable, "-B", str(CLAIMS / "analyze.py"), "--check"],
        [sys.executable, "-B", str(HERE / "verify.py")],
    ):
        subprocess.run(command, check=True, capture_output=True, text=True)
    summary = evaluate(
        {profile: read(CLAIMS / (profile + ".json.gz"))
         for profile in ("technical", "strict")},
        {profile: read(HERE / (profile + ".json.gz"))
         for profile in ("technical", "strict")},
        read(SCOPED / "result.json"),
        read(CLAIMS / "result.json"),
        read(HERE / "summary.json"),
        read(HERE / "new-review.json"),
    )
    encoded = (json.dumps(summary, ensure_ascii=False, sort_keys=True, indent=2) +
               "\n").encode()
    require((HERE / "reconciliation.json").read_bytes() == encoded,
            "recorded reconciliation differs")
    print(json.dumps({"reconciliation_sha256": hashlib.sha256(encoded).hexdigest(),
                      "new_detector_findings": summary["new_detector_findings_in_replay"],
                      "disputed_events": summary["disputed_events"]},
                     sort_keys=True))


if __name__ == "__main__":
    main()
