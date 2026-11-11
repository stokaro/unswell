#!/usr/bin/env python3
"""Verify the exposed #349 instruction-operand paired replay."""

import gzip
import hashlib
import importlib.util
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
REVIEWS = HERE.parent
INPUTS = {
    "whole_archive": REVIEWS / "2026-09-17-full-page-recall/inputs.tar.gz",
    "context_archive": REVIEWS / "2026-09-17-proposition-repetition/confirmation/inputs.tar.gz",
    "confirmation_sources": REVIEWS / "2026-09-26-outcome-announcements/confirmation-inputs.json",
    "replay_script": REVIEWS / "2026-09-26-outcome-announcements/replay.py",
    "archived_technical": REVIEWS / "2026-11-10-current-main-editorial-delta/technical.json.gz",
    "archived_strict": REVIEWS / "2026-11-10-current-main-editorial-delta/strict.json.gz",
    "prior_finding_review": REVIEWS / "2026-11-07-current-main-editorial-review/findings.json",
    "prior_event_review": REVIEWS / "2026-11-07-current-main-editorial-review/events.json",
    "prior_second_review": REVIEWS / "2026-11-10-current-main-editorial-delta/new-review.json",
    "prior_second_summary": REVIEWS / "2026-11-10-current-main-editorial-delta/summary.json",
}
REMOVED = {
    ("filler.instruction-scaffolding", "whole/p015.md", 8356, 8459),
    ("filler.instruction-scaffolding", "whole/p015.md", 22265, 22401),
}


def require(ok, message):
    if not ok:
        raise ValueError(message)


def digest(data):
    return hashlib.sha256(data).hexdigest()


def read(path):
    data = path.read_bytes()
    return json.loads(gzip.decompress(data) if path.suffix == ".gz" else data)


def location(finding):
    span = finding["primary"]["span"]
    return (finding["rule_id"], finding["primary"]["path"],
            span["start"], span["end"])


def diagnostic(finding):
    return {key: value for key, value in finding.items()
            if key not in {"id", "fingerprint", "rule_version"}}


def index(report, commit, sources):
    require(report["status"] == "complete" and report["manifest"]["complete"] and
            report["manifest"]["tool_commit"] == commit and
            not report["manifest"]["skipped_rules"] and not report["errors"] and
            report["gate"]["passed"], "incomplete or stale report")
    documents = {row["name"]: (row["source_hash"], row["bytes"])
                 for row in report["documents"]}
    require(len(documents) == len(report["documents"]) == 36 and
            documents == sources, "source coverage or identity changed")
    findings = {location(row): row for row in report["findings"]}
    require(len(findings) == len(report["findings"]), "duplicate finding location")
    return findings


def evaluate(archived, before, after, review, events, additions, prior, sources, record):
    result = {}
    reviewed = {(row["rule_id"], row["path"], row["span"]["start"],
                 row["span"]["end"]): row for row in review}
    require(REMOVED <= reviewed.keys() and
            all(reviewed[key]["status"] == "nonactionable" for key in REMOVED),
            "removed finding lacks a nonactionable review")
    old_ids = {reviewed[key]["id"] for key in REMOVED}
    require(not any(old_ids.intersection(row["credited_finding_ids"]) for row in events),
            "removed finding credited to an event")
    require(not any(row["id"] in old_ids for row in additions),
            "removed finding was in the later accepted review")
    for profile, expected, whole_total in (("technical", 162, 126), ("strict", 183, 145)):
        old = index(archived[profile], "3acba5eb4c13b95207caf238da17e9c0a72630e3", sources)
        left = index(before[profile], record["base_commit"], sources)
        right = index(after[profile], record["candidate_build_commit"], sources)
        require(len(old) == len(left) == expected and set(old) == set(left),
                "current-main baseline differs from archived review: " + profile)
        require(all(diagnostic(old[key]) == diagnostic(left[key]) for key in old),
                "baseline diagnosis differs from archived review: " + profile)
        require(set(left) - set(right) == REMOVED and not set(right) - set(left),
                "unexpected removed or added findings: " + profile)
        require(all(diagnostic(left[key]) == diagnostic(right[key]) for key in right),
                "retained diagnosis changed: " + profile)
        require(all(right[key]["rule_version"] == "18" for key in right
                    if key[0] == "filler.instruction-scaffolding"),
                "instruction rule version did not advance")
        cohort = prior["profiles"][profile]["by_cohort"]
        require(cohort["whole"]["findings"]["total"] == whole_total
                and cohort["whole"]["findings"]["accepted"] == 82,
                "prior complete-stream review denominator changed")
        result[profile] = {
            "before": len(left), "after": len(right), "removed_nonactionable": 2,
            "whole_accepted": 82,
            "whole_delivered": cohort["whole"]["findings"]["total"] - 2,
            "full_events": {name: row["events"]["full"] for name, row in cohort.items()},
        }
    return result


def expected_sources():
    spec = importlib.util.spec_from_file_location("unswell_replay", INPUTS["replay_script"])
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return {name: (digest(data), len(data)) for name, data in module.sources().items()}


def load_inputs():
    record = read(HERE / "record.json")
    require(record["format"] == "unswell-instruction-operand-guard-record-v1" and
            record["base_commit"] == "7f86b91ba5da7a6ce4ba98697ce2ce73ce0dde31" and
            record["candidate_build_commit"] == "local-operand-guard-v4", "record identity changed")
    require(set(record["inputs_sha256"]) == set(INPUTS), "input inventory changed")
    for name, path in INPUTS.items():
        require(digest(path.read_bytes()) == record["inputs_sha256"][name],
                "input hash changed: " + name)
    reports = {}
    require(set(record["reports_sha256"]) == {
        f"{phase}-{profile}.json.gz" for phase in ("before", "after")
        for profile in ("technical", "strict")}, "report inventory changed")
    for phase in ("before", "after"):
        reports[phase] = {}
        for profile in ("technical", "strict"):
            name = f"{phase}-{profile}.json.gz"
            path = HERE / name
            require(digest(path.read_bytes()) == record["reports_sha256"][name],
                    "report hash changed: " + name)
            reports[phase][profile] = read(path)
    return record, reports


def main():
    record, reports = load_inputs()
    archived = {profile: read(INPUTS[f"archived_{profile}"])
                for profile in ("technical", "strict")}
    result = evaluate(archived, reports["before"], reports["after"],
                      read(INPUTS["prior_finding_review"]),
                      read(INPUTS["prior_event_review"]),
                      read(INPUTS["prior_second_review"]),
                      read(INPUTS["prior_second_summary"]),
                      expected_sources(), record)
    print(json.dumps(result, indent=2, sort_keys=True))


if __name__ == "__main__":
    main()
