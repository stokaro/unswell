#!/usr/bin/env python3
"""Verify the exposed quality-claim replay and its frozen editorial targets."""

import argparse
import gzip
import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
PRIOR = ROOT / "research/reviews/2026-11-07-scoped-degree"
WHOLE = ROOT / "research/reviews/2026-09-17-full-page-recall/annotations.json"
CONTEXT = ROOT / "research/reviews/2026-09-17-proposition-repetition/confirmation/annotations.json"
OLD_EVENTS = ROOT / "research/reviews/2026-11-07-current-main-editorial-review/events.json"
EXPECTED = {
    ("context/c05.md", 4497, 4516): "context/c05-d05",
    ("whole/p013.md", 2430, 2466): "whole/p013-d01",
    ("whole/p018.md", 9629, 9645): "whole/p018-d02",
    ("whole/p019.md", 598, 628): "whole/p019-d02",
    ("whole/p023.md", 1456, 1492): "whole/p023-d01",
}


def read(path):
    if path.suffix == ".gz":
        with gzip.open(path, "rt", encoding="utf-8") as stream:
            return json.load(stream)
    return json.loads(path.read_text(encoding="utf-8"))


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def verified_report(path, expected_hash):
    if digest(path) != expected_hash:
        raise ValueError(f"Report hash drift: {path}")
    report = read(path)
    if (report["status"] != "complete" or not report["manifest"]["complete"]
            or report["manifest"]["skipped_rules"] or report["errors"]
            or len(report["documents"]) != 36):
        raise ValueError(f"Incomplete 36-page report: {path}")
    for document in report["documents"]:
        if hashlib.sha256(document["source"].encode("utf-8")).hexdigest() != document["source_hash"]:
            raise ValueError(f"Source hash drift: {document['name']}")
    return report


def key(finding):
    primary = finding["primary"]
    return (finding["rule_id"], primary["path"],
            primary["span"]["start"], primary["span"]["end"])


def location(finding):
    primary = finding["primary"]
    return primary["path"], primary["span"]["start"], primary["span"]["end"]


def indexed(findings):
    output = {key(finding): finding for finding in findings}
    if len(output) != len(findings):
        raise ValueError("Duplicate finding location")
    return output


def stable_content(finding):
    return {name: value for name, value in finding.items()
            if name not in {"id", "rule_version", "fingerprint"}}


def labels():
    pages = read(WHOLE) + read(CONTEXT)["pages"]
    targets = {}
    for page in pages:
        cohort = "whole" if page["page"].startswith("p") else "context"
        path = f"{cohort}/{page['page']}.md"
        for defect in page["defects"]:
            targets[f"{cohort}/{defect['id']}"] = (path, defect)
    return targets


def validate_additions(additions, documents, frozen, old_events):
    if {location(finding) for finding in additions} != set(EXPECTED):
        raise ValueError("Added locations differ from reviewed screen")
    for finding in additions:
        path, start, end = location(finding)
        event = EXPECTED[(path, start, end)]
        target_path, defect = frozen[event]
        if target_path != path or defect["category"] != "vague_claims":
            raise ValueError(f"Wrong frozen category: {event}")
        if old_events[event]["status"] != "missed":
            raise ValueError(f"Previously credited event: {event}")
        if not any(start < target["end"] and target["start"] < end for target in defect["targets"]):
            raise ValueError(f"Added finding misses frozen target: {event}")
        source = documents[path]["source"].encode("utf-8")
        if source[start:end].decode("utf-8") != finding["primary"]["snippet"]:
            raise ValueError(f"Added finding source-span drift: {event}")
        if (finding["rule_id"] != "filler.unscoped-assurance" or finding["rule_version"] != "12"
                or finding["severity"] != "warning" or not finding["evidence"]["suggestion"]):
            raise ValueError(f"Unexpected new finding contract: {event}")


def analyze(folder=HERE):
    ledger = read(folder / "runs.json")
    prior_ledger = read(PRIOR / "runs.json")
    prior_result = read(PRIOR / "result.json")
    frozen = labels()
    old_events = {entry["event"]: entry for entry in read(OLD_EVENTS)}
    if len(old_events) != 123:
        raise ValueError("Frozen event count drift")
    result = {"format": "unswell-quality-claim-screen-v1", "status": "exposed-development",
              "reviewer": "Codex assistant under ADR 0041", "profiles": {}}
    for profile in ("technical", "strict"):
        old = verified_report(PRIOR / (profile + ".json.gz"),
                              prior_ledger[profile]["report_sha256"])
        path = folder / (profile + ".json.gz")
        new = verified_report(path, ledger[profile]["report_sha256"])
        old_documents = {doc["name"]: doc for doc in old["documents"]}
        new_documents = {doc["name"]: doc for doc in new["documents"]}
        if {name: doc["source_hash"] for name, doc in old_documents.items()} != {
                name: doc["source_hash"] for name, doc in new_documents.items()}:
            raise ValueError("Document source drift")
        before, after = indexed(old["findings"]), indexed(new["findings"])
        if before.keys() - after.keys():
            raise ValueError("Previously delivered finding removed")
        for key_ in before.keys() & after.keys():
            previous, current = before[key_], after[key_]
            if stable_content(previous) != stable_content(current):
                raise ValueError(f"Changed prior finding content: {key_}")
            if current["rule_id"] == "filler.unscoped-assurance":
                if previous["rule_version"] != "11" or current["rule_version"] != "12":
                    raise ValueError("Unexpected assurance rule version")
            elif previous != current:
                raise ValueError(f"Changed unrelated finding: {key_}")
        additions = [after[item] for item in after.keys() - before.keys()]
        validate_additions(additions, new_documents, frozen, old_events)
        cohorts = {}
        for cohort, old_counts in prior_result["profiles"][profile]["cohorts"].items():
            gained = sum(event.startswith(cohort + "/") for event in EXPECTED.values())
            full = old_counts["full"] + gained
            delivered = old_counts["delivered"] + gained
            accepted = old_counts["accepted"] + gained
            cohorts[cohort] = {"full": full, "events": old_counts["events"],
                               "required_full": old_counts["required_full"],
                               "accepted": accepted, "delivered": delivered,
                               "required_accepted": (17 * delivered + 19) // 20}
        result["profiles"][profile] = {
            "prior_findings": len(before), "candidate_findings": len(after),
            "added_locations": [list(loc) for loc in sorted(EXPECTED)],
            "removed_locations": 0, "new_accepted": len(additions),
            "new_nonaccepted": 0, "cohorts": cohorts,
            "report_sha256": digest(path),
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
    print("quality-claim screen verified")


if __name__ == "__main__":
    main()
