#!/usr/bin/env python3
"""Audit the exposed scoped-degree replay against the frozen current baseline."""

import argparse
import gzip
import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
BASE = ROOT / "research/reviews/2026-11-07-current-main-baseline"
REVIEW = ROOT / "research/reviews/2026-11-07-current-main-editorial-review"
CONTEXT = ROOT / "research/reviews/2026-09-17-proposition-repetition/confirmation"
ADDITIONS = {
    ("context/c02.md", 1002, 1021): "context/c02-d03",
    ("context/c05.md", 301, 319): "context/c05-d02",
}


def read_json(path):
    if path.suffix == ".gz":
        with gzip.open(path, "rt", encoding="utf-8") as stream:
            return json.load(stream)
    return json.loads(path.read_text(encoding="utf-8"))


def sha256(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def key(finding):
    primary = finding["primary"]
    return (finding["rule_id"], primary["path"], primary["span"]["start"], primary["span"]["end"])


def location(finding):
    primary = finding["primary"]
    return primary["path"], primary["span"]["start"], primary["span"]["end"]


def indexed(findings):
    result = {key(finding): finding for finding in findings}
    if len(result) != len(findings):
        raise ValueError("Duplicate finding location")
    return result


def without_identity(finding):
    return {name: value for name, value in finding.items()
            if name not in {"id", "rule_version", "fingerprint"}}


def verified_report(path, expected_hash):
    if sha256(path) != expected_hash:
        raise ValueError(f"Report hash drift: {path}")
    report = read_json(path)
    if (report["status"] != "complete" or not report["manifest"]["complete"]
            or report["manifest"]["skipped_rules"] or report["errors"]):
        raise ValueError(f"Incomplete report: {path}")
    if len(report["documents"]) != 36:
        raise ValueError(f"Document count drift: {path}")
    for document in report["documents"]:
        source = document["source"].encode("utf-8")
        if hashlib.sha256(source).hexdigest() != document["source_hash"]:
            raise ValueError(f"Source hash drift: {document['name']}")
    return report


def review_map():
    decisions = read_json(REVIEW / "findings.json")
    result = {(entry["rule_id"], entry["path"], entry["span"]["start"],
               entry["span"]["end"]): entry for entry in decisions}
    if len(result) != 176:
        raise ValueError("Frozen review count drift")
    return result


def target_map():
    pages = read_json(CONTEXT / "annotations.json")["pages"]
    result = {}
    for page in pages:
        path = "context/" + page["page"] + ".md"
        for defect in page["defects"]:
            result["context/" + defect["id"]] = (path, defect)
    return result


def analyze(folder=HERE):
    ledger = read_json(folder / "runs.json")
    old_decisions = review_map()
    context_labels = target_map()
    event_decisions = read_json(REVIEW / "events.json")
    if len(event_decisions) != 123:
        raise ValueError("Frozen event count drift")
    result = {"format": "unswell-scoped-degree-audit-v1", "status": "exposed-development",
              "reviewer": "Codex assistant under ADR 0041", "profiles": {}}
    for profile in ("technical", "strict"):
        old = verified_report(BASE / (profile + ".json.gz"),
                              read_json(BASE / "runs.json")[profile]["report_sha256"])
        new_path = folder / (profile + ".json.gz")
        new = verified_report(new_path, ledger[profile]["report_sha256"])
        old_documents = {entry["name"]: entry["source_hash"] for entry in old["documents"]}
        new_documents = {entry["name"]: entry["source_hash"] for entry in new["documents"]}
        if old_documents != new_documents:
            raise ValueError("Document source drift")
        before, after = indexed(old["findings"]), indexed(new["findings"])
        if before.keys() - after.keys():
            raise ValueError("Removed baseline finding")
        for match in before.keys() & after.keys():
            previous, current = before[match], after[match]
            if without_identity(previous) != without_identity(current):
                raise ValueError(f"Changed finding content: {match}")
            if previous["rule_id"] == "filler.unscoped-assurance":
                if previous["rule_version"] != "10" or current["rule_version"] != "11":
                    raise ValueError("Unexpected assurance rule version")
            elif previous != current:
                raise ValueError(f"Changed unrelated finding: {match}")
            if match not in old_decisions:
                raise ValueError(f"Missing frozen editorial decision: {match}")
        additions = [after[match] for match in after.keys() - before.keys()]
        if {location(entry) for entry in additions} != set(ADDITIONS):
            raise ValueError("Unexpected added locations")
        for finding in additions:
            loc = location(finding)
            if finding["rule_id"] != "filler.unscoped-assurance" or finding["rule_version"] != "11":
                raise ValueError("Unexpected added rule")
            event = ADDITIONS[loc]
            path, defect = context_labels[event]
            if path != loc[0] or defect["category"] != "unjustified_intensifiers":
                raise ValueError("Added finding does not match frozen category")
            if not all(loc[1] < target["end"] and target["start"] < loc[2]
                       for target in defect["targets"]):
                raise ValueError("Added finding misses frozen target")
            source = next(doc["source"].encode("utf-8") for doc in new["documents"]
                          if doc["name"] == loc[0])
            if source[loc[1]:loc[2]].decode("utf-8") != finding["primary"]["snippet"]:
                raise ValueError("Added finding source span drift")
            touched = {event for event, (path, defect) in context_labels.items()
                       if path == loc[0] and any(loc[1] < target["end"] and target["start"] < loc[2]
                                                  for target in defect["targets"])}
            if touched != {event}:
                raise ValueError("Added finding touches another frozen event")
        reviewed = [old_decisions[match]["status"] for match in before]
        accepted = sum(status == "actionable" for status in reviewed) + len(additions)
        nonaccepted = len(after) - accepted
        events = {entry["event"]: entry for entry in event_decisions}
        for event in ADDITIONS.values():
            if events[event]["status"] != "missed":
                raise ValueError("Previously credited event")
        cohort = {}
        for name, prefix in (("whole", "whole/"), ("context", "context/"),
                             ("confirmation", "confirmation/")):
            old_events = [entry for entry in event_decisions if entry["path"].startswith(prefix)]
            gained = sum(event.startswith(prefix) for event in ADDITIONS.values())
            full = sum(entry["status"] == "full" for entry in old_events) + gained
            findings = [entry for entry in after.values() if entry["primary"]["path"].startswith(prefix)]
            accepted_here = sum(old_decisions[match]["status"] == "actionable"
                                for match in before if match[1].startswith(prefix)) + gained
            cohort[name] = {"full": full, "events": len(old_events),
                            "required_full": (4 * len(old_events) + 4) // 5,
                            "accepted": accepted_here, "delivered": len(findings),
                            "required_accepted": (17 * len(findings) + 19) // 20}
        result["profiles"][profile] = {
            "baseline_findings": len(before), "candidate_findings": len(after),
            "added_locations": [list(loc) for loc in sorted(ADDITIONS)],
            "removed_locations": 0, "accepted": accepted,
            "nonaccepted": nonaccepted, "cohorts": cohort,
            "report_sha256": sha256(new_path),
        }
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    output = json.dumps(analyze(), indent=2, sort_keys=True) + "\n"
    path = HERE / "result.json"
    if args.check:
        if path.read_text(encoding="utf-8") != output:
            raise ValueError("Result drift")
    else:
        path.write_text(output, encoding="utf-8")
    print("scoped-degree audit verified")


if __name__ == "__main__":
    main()
