#!/usr/bin/env python3
"""Verify the bounded hope-framing development replay against frozen inputs."""

import argparse
import gzip
import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
BASE = ROOT / "research/reviews/2026-11-07-quality-claim-screen"
WHOLE = ROOT / "research/reviews/2026-09-17-full-page-recall/annotations.json"
CONTEXT = ROOT / "research/reviews/2026-09-17-proposition-repetition/confirmation/annotations.json"
CURRENT_EVENTS = ROOT / "research/reviews/2026-11-07-current-main-editorial-review/events.json"
EXPECTED = {
    ("whole/p016.md", 141, 222): "whole/p016-d01",
    ("whole/p017.md", 15250, 15329): "whole/p017-d03",
    ("context/c04.md", 959, 1126): "context/c04-d03",
}


def read(path):
    if path.suffix == ".gz":
        with gzip.open(path, "rt", encoding="utf-8") as stream:
            return json.load(stream)
    return json.loads(path.read_text(encoding="utf-8"))


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def report(path, expected_hash):
    if digest(path) != expected_hash:
        raise ValueError(f"Report hash drift: {path}")
    data = read(path)
    if (data["status"] != "complete" or not data["manifest"]["complete"]
            or data["manifest"]["skipped_rules"] or data["errors"]
            or len(data["documents"]) != 36):
        raise ValueError(f"Incomplete report: {path}")
    for doc in data["documents"]:
        if hashlib.sha256(doc["source"].encode("utf-8")).hexdigest() != doc["source_hash"]:
            raise ValueError(f"Source hash drift: {doc['name']}")
    return data


def location(finding):
    primary = finding["primary"]
    return primary["path"], primary["span"]["start"], primary["span"]["end"]


def key(finding):
    return (finding["rule_id"],) + location(finding)


def indexed(findings):
    output = {key(finding): finding for finding in findings}
    if len(output) != len(findings):
        raise ValueError("Duplicate finding location")
    return output


def stable_content(finding):
    return {name: value for name, value in finding.items()
            if name not in {"id", "rule_version", "fingerprint"}}


def frozen_targets():
    targets = {}
    for cohort, path in (("whole", WHOLE), ("context", CONTEXT)):
        data = read(path)
        pages = data if isinstance(data, list) else data["pages"]
        for page in pages:
            for defect in page["defects"]:
                targets[f"{cohort}/{defect['id']}"] = (page["page"], defect)
    return targets


def validate_additions(additions, documents, decisions, events, targets):
    if {location(finding) for finding in additions} != set(EXPECTED):
        raise ValueError("Added locations differ from frozen hope screen")
    if set(decisions) != set(EXPECTED.values()):
        raise ValueError("Review decision set drift")
    for finding in additions:
        where = location(finding)
        event = EXPECTED[where]
        decision = decisions[event]
        page_id, defect = targets[event]
        cohort = event.split("/", 1)[0]
        if (where[0] != f"{cohort}/{page_id}.md"
                or events[event]["status"] == "full"
                or decision["event_status"] != "full"
                or decision["status"] != "accepted"
                or decision["advice_status"] != "conditional_guidance; no applied edit"
                or not decision["rationale"] or not decision["meaning_preservation"]):
            raise ValueError(f"Invalid editorial decision: {event}")
        start, end = where[1:]
        if not any(start < target["end"] and target["start"] < end
                   for target in defect["targets"]):
            raise ValueError(f"Target span mismatch: {event}")
        source = documents[where[0]]["source"].encode("utf-8")
        if source[start:end].decode("utf-8") != finding["primary"]["snippet"]:
            raise ValueError(f"Original-byte span mismatch: {event}")
        if (finding["rule_id"] != "filler.evaluative-closure"
                or finding["rule_version"] != "13"
                or finding["severity"] != "warning"
                or not finding["evidence"]["suggestion"]):
            raise ValueError(f"Unexpected new finding contract: {event}")


def analyze(folder=HERE):
    runs = read(folder / "runs.json")
    base_runs = read(BASE / "runs.json")
    base_result = read(BASE / "result.json")
    review = read(folder / "review.json")
    decisions = {row["event"]: row for row in review["decisions"]}
    if len(decisions) != len(review["decisions"]):
        raise ValueError("Duplicate review decision")
    if review["reviewer"] != "Codex assistant under ADR 0041":
        raise ValueError("Reviewer identity drift")
    events = {row["event"]: row for row in read(CURRENT_EVENTS)}
    if len(events) != 123:
        raise ValueError("Frozen event count drift")
    targets = frozen_targets()
    result = {"format": "unswell-hope-framing-screen-v1",
              "status": "exposed-development", "reviewer": review["reviewer"],
              "event_denominators": {"whole": 57, "context": 46, "confirmation": 20},
              "profiles": {}}
    for profile in ("technical", "strict"):
        old = report(BASE / (profile + ".json.gz"), base_runs[profile]["report_sha256"])
        path = folder / (profile + ".json.gz")
        new = report(path, runs[profile]["report_sha256"])
        old_docs = {doc["name"]: doc for doc in old["documents"]}
        docs = {doc["name"]: doc for doc in new["documents"]}
        if ({name: doc["source_hash"] for name, doc in old_docs.items()} !=
                {name: doc["source_hash"] for name, doc in docs.items()}):
            raise ValueError("Document identity drift")
        before, after = indexed(old["findings"]), indexed(new["findings"])
        if before.keys() - after.keys():
            raise ValueError("Previous finding removed")
        for key_ in before.keys() & after.keys():
            previous, current = before[key_], after[key_]
            if stable_content(previous) != stable_content(current):
                raise ValueError(f"Previous finding content changed: {key_}")
            if current["rule_id"] == "filler.evaluative-closure":
                if previous["rule_version"] != "12" or current["rule_version"] != "13":
                    raise ValueError("Unexpected evaluative-closure version")
            elif previous != current:
                raise ValueError(f"Unrelated finding changed: {key_}")
        additions = [after[item] for item in after.keys() - before.keys()]
        validate_additions(additions, docs, decisions, events, targets)
        cohorts = {}
        for cohort, old_counts in base_result["profiles"][profile]["cohorts"].items():
            gained = sum(event.startswith(cohort + "/") for event in EXPECTED.values())
            full = old_counts["full"] + gained
            delivered = old_counts["delivered"] + gained
            accepted = old_counts["accepted"] + gained
            cohorts[cohort] = {"full": full, "events": old_counts["events"],
                               "required_full": old_counts["required_full"],
                               "accepted": accepted, "delivered": delivered,
                               "required_accepted": (17 * delivered + 19) // 20}
        result["profiles"][profile] = {
            "before_findings": len(before), "after_findings": len(after),
            "added_locations": [list(loc) for loc in sorted(EXPECTED)],
            "removed_locations": 0, "changed_content": 0,
            "new_accepted": len(additions), "new_nonaccepted": 0,
            "new_full_events": sorted(EXPECTED.values()),
            "lost_full_events": [], "cohorts": cohorts,
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
    print("hope-framing screen verified")


if __name__ == "__main__":
    main()
