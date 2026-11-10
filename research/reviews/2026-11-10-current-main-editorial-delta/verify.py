#!/usr/bin/env python3
"""Verify the #349 current-main replay and source-reviewed editorial delta."""

import argparse
from collections import Counter, defaultdict
import gzip
import hashlib
import json
from pathlib import Path
import subprocess
import sys


HERE = Path(__file__).resolve().parent
REPO = HERE.parents[2]
REVIEWS = HERE.parent
PRIOR = REVIEWS / "2026-09-26-outcome-announcements"
BASE = REVIEWS / "2026-11-07-current-main-baseline"
OLD = REVIEWS / "2026-11-07-current-main-editorial-review"
sys.path.insert(0, str(PRIOR))
import replay  # noqa: E402
import verify as labels_verify  # noqa: E402


def require(condition, message):
    if not condition:
        raise ValueError(message)


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path):
    data = path.read_bytes()
    return json.loads(gzip.decompress(data) if path.suffix == ".gz" else data)


def identity(finding):
    primary = finding["primary"]
    span = primary["span"]
    return finding["rule_id"], primary["path"], span["start"], span["end"]


def normalized(finding):
    return {key: value for key, value in finding.items()
            if key not in {"id", "fingerprint", "rule_version"}}


def overlap(left, right):
    return max(left["start"], right["start"]) < min(left["end"], right["end"])


def locations(finding):
    return [finding["primary"], *finding["related"]]


def validate_report(report, sources, commit, profile):
    require(report["status"] == "complete" and report["manifest"]["complete"] and
            not report["errors"] and not report["manifest"]["skipped_rules"],
            "incomplete " + profile + " report")
    require(report["manifest"]["tool_commit"] == commit, "wrong runtime commit: " + profile)
    documents = report["documents"]
    require(len(documents) == len(sources) == 36 and
            {row["name"] for row in documents} == set(sources),
            "source coverage changed: " + profile)
    for row in documents:
        source = sources[row["name"]]
        require(row["source_hash"] == hashlib.sha256(source).hexdigest() and
                row["bytes"] == len(source), "source mismatch: " + row["name"])
    ids = set()
    keys = set()
    for finding in report["findings"]:
        fid, key = finding["id"], identity(finding)
        require(fid not in ids and key not in keys, "duplicate finding: " + fid)
        ids.add(fid)
        keys.add(key)
        for location in locations(finding):
            path, span = location["path"], location["span"]
            require(path in sources, "finding outside frozen sources: " + fid)
            source = sources[path]
            require(0 <= span["start"] < span["end"] <= len(source),
                    "finding span outside source: " + fid)
            source[span["start"]:span["end"]].decode("utf-8")


def hits(finding, label):
    return [
        any(location["path"] == label["path"] and
            overlap(location["span"], target)
            for location in locations(finding))
        for target in label["targets"]
    ]


def counts(rows):
    return dict(sorted(Counter(rows).items()))


def evaluate(old_reports, current_reports, old_findings, old_events, new_review,
             baseline, sources, labels, expected_commit):
    require(len(sources) == 36 and len(labels) == 123, "frozen reference count changed")
    for event, label in labels.items():
        require(label["path"] in sources, "label outside frozen sources: " + event)
        for target in label["targets"]:
            quote = sources[label["path"]][target["start"]:target["end"]].decode("utf-8")
            require(quote == target["quote"], "target quote drift: " + event)
    for profile in ("technical", "strict"):
        validate_report(old_reports[profile], sources, baseline["base_commit"],
                        "old-" + profile)
        validate_report(current_reports[profile], sources, expected_commit, profile)
    old_strict = {identity(f): f for f in old_reports["strict"]["findings"]}
    current_strict = {identity(f): f for f in current_reports["strict"]["findings"]}
    old_keys, current_keys = set(old_strict), set(current_strict)
    require(len(old_keys) == len(old_findings) == 176, "old review is incomplete")
    require(old_keys <= current_keys, "a previously reviewed finding was removed")
    added_keys = current_keys - old_keys
    require(len(added_keys) == len(new_review) == 7, "unexpected new finding count")
    old_technical = {identity(f): f for f in old_reports["technical"]["findings"]}
    current_technical = {identity(f): f for f in current_reports["technical"]["findings"]}
    require(set(old_technical) <= set(current_technical), "technical finding removed")
    require(set(old_technical) <= old_keys and
            set(current_technical) <= current_keys, "technical/strict mismatch")
    require(added_keys == set(current_technical) - set(old_technical),
            "new finding is missing from a profile")

    old_by_id = {}
    current_by_id = {}
    carried = {}
    for key in sorted(old_keys):
        before, after = old_strict[key], current_strict[key]
        require(normalized(before) == normalized(after),
                "previously reviewed finding changed: " + str(key))
        require((key in old_technical) == (key in current_technical),
                "finding profile changed: " + str(key))
        old_by_id[before["id"]] = before
        current_by_id[after["id"]] = after
        carried[before["id"]] = after["id"]
    for key in added_keys:
        current_by_id[current_strict[key]["id"]] = current_strict[key]

    reviewed = {}
    require(len({row["id"] for row in old_findings}) == len(old_findings),
            "duplicate prior review ID")
    for row in old_findings:
        before = old_by_id.get(row["id"])
        require(before is not None and
                (row["rule_id"], row["path"], row["span"]["start"],
                 row["span"]["end"]) == identity(before),
                "prior review identity changed: " + row["id"])
        require(row["status"] in {"actionable", "uncertain", "nonactionable"} and
                row["rationale"].strip(), "invalid prior finding decision")
        require(bool(row["repair_action"]) == (row["status"] == "actionable"),
                "prior repair status changed")
        reviewed[carried[row["id"]]] = {
            "status": row["status"], "path": row["path"],
            "advice_status": row["advice_status"],
        }
    new_by_id = {}
    for row in new_review:
        fid = row["id"]
        require(fid not in new_by_id and fid in current_by_id and
                identity(current_by_id[fid]) in added_keys,
                "new review does not match a new finding: " + fid)
        require(row["finding_status"] == "actionable" and
                row["event_status"] in {"full", "partial"} and
                row["finding_rationale"].strip() and row["repair_action"].strip() and
                row["event_rationale"].strip(), "incomplete new review: " + fid)
        finding = current_by_id[fid]
        require(finding["rule_id"] == "filler.unscoped-assurance" and
                finding["evidence"]["suggestion"], "unexpected new finding: " + fid)
        event, label = row["event"], labels.get(row["event"])
        require(label is not None and label["path"] == finding["primary"]["path"] and
                all(hits(finding, label)), "new finding misses its labeled event: " + fid)
        other_hits = {key for key, other in labels.items()
                      if key != event and any(hits(finding, other))}
        require(not other_hits, "new finding affects unreviewed events: " + fid)
        new_by_id[fid] = row
        reviewed[fid] = {
            "status": row["finding_status"],
            "path": finding["primary"]["path"],
            "advice_status": "conditional_guidance",
        }
    require(len(reviewed) == len(current_strict), "incomplete current finding review")
    require(len(old_events) == len(labels) == 123 and
            len({row["event"] for row in old_events}) == len(old_events),
            "incomplete prior event review")
    old_event_by_name = {row["event"]: row for row in old_events}
    new_by_event = {row["event"]: row for row in new_review}
    require(len(new_by_event) == 7, "duplicate new event decision")
    event_rows = []
    current_technical_ids = {f["id"] for f in current_reports["technical"]["findings"]}
    for event, label in sorted(labels.items()):
        prior = old_event_by_name[event]
        require(prior["path"] == label["path"] and
                prior["category"] == label["category"] and
                prior["status"] in {"full", "partial", "uncertain", "missed"},
                "prior event identity changed: " + event)
        if event in new_by_event:
            require(prior["status"] == "missed" and
                    not prior["credited_finding_ids"],
                    "new event was previously reviewed as covered: " + event)
            decision = new_by_event[event]
            status = decision["event_status"]
            credit = [decision["id"]] if status == "full" else []
        else:
            status = prior["status"]
            credit = [carried[fid] for fid in prior["credited_finding_ids"]]
        if status == "full":
            require(credit and all(fid in current_technical_ids and
                    reviewed[fid]["status"] == "actionable" for fid in credit),
                    "unsupported full-event witness: " + event)
            for target_index in range(len(label["targets"])):
                require(any(hits(current_by_id[fid], label)[target_index]
                            for fid in credit), "uncovered full target: " + event)
        else:
            require(not credit, "credit on nonfull event: " + event)
        event_rows.append({"event": event, "path": label["path"],
                           "category": label["category"], "status": status,
                           "credited_finding_ids": credit})

    profiles = {}
    for profile in ("technical", "strict"):
        findings = current_reports[profile]["findings"]
        profile_rows = [reviewed[f["id"]] for f in findings]
        by_cohort = {}
        for cohort in ("whole", "context", "confirmation"):
            cohort_findings = [row for row in profile_rows
                               if row["path"].startswith(cohort + "/")]
            cohort_events = [row for row in event_rows
                             if row["event"].startswith(cohort + "/")]
            threshold = baseline["profiles"][profile]["by_cohort"][cohort]
            require(len(cohort_events) == threshold["events"],
                    "event denominator changed: " + cohort)
            accepted = sum(row["status"] == "actionable" for row in cohort_findings)
            full = sum(row["status"] == "full" for row in cohort_events)
            by_cohort[cohort] = {
                "findings": {
                    "total": len(cohort_findings), "accepted": accepted,
                    "required_accepted_85pct": (85 * len(cohort_findings) + 99) // 100,
                    "dispositions": counts(row["status"] for row in cohort_findings),
                },
                "events": {
                    "total": len(cohort_events), "full": full,
                    "required_full_80pct": threshold["required_full"],
                    "dispositions": counts(row["status"] for row in cohort_events),
                },
            }
        profiles[profile] = {
            "findings": len(findings),
            "finding_dispositions": counts(row["status"] for row in profile_rows),
            "by_cohort": by_cohort,
        }
    gained = sorted(row["event"] for row in event_rows
                    if row["status"] == "full" and
                    old_event_by_name[row["event"]]["status"] != "full")
    return {
        "format": "unswell-current-main-editorial-delta-v1",
        "reviewer": "Codex assistant; source review under ADR 0041, not a human annotator",
        "runtime_commit": expected_commit,
        "previous_runtime_commit": baseline["base_commit"],
        "frozen_sources": len(sources), "frozen_events": len(labels),
        "carried_findings": len(carried),
        "carried_finding_id_changes": sum(before != after
                                          for before, after in carried.items()),
        "removed_findings": 0, "new_findings": len(new_review),
        "new_event_decisions": counts(row["event_status"] for row in new_review),
        "gained_full_events": gained,
        "profiles": profiles,
        "interpretation": "Only byte-identical normalized old findings inherit their prior source review. New findings received assistant source review. Full events count accepted diagnosis, not edit safety or model origin detection.",
    }


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--write", action="store_true")
    args = parser.parse_args()
    freeze = read(HERE / "freeze.json")
    for name, expected in freeze["inputs"].items():
        require(digest(REPO / name) == expected, "frozen input changed: " + name)
    run = read(HERE / "runs.json")
    for profile in ("technical", "strict"):
        record = run[profile]
        require(record["binary_sha256"] == freeze["binary_sha256"] and
                record["report_sha256"] == digest(HERE / (profile + ".json.gz")) and
                record["exit_code"] in (0, 1), "run identity mismatch: " + profile)
    subprocess.run([sys.executable, str(OLD / "verify.py")], check=True,
                   capture_output=True, text=True)
    sources, labels = replay.sources(), labels_verify.labels()
    old_reports = {p: read(BASE / (p + ".json.gz")) for p in ("technical", "strict")}
    current_reports = {p: read(HERE / (p + ".json.gz")) for p in ("technical", "strict")}
    summary = evaluate(old_reports, current_reports, read(OLD / "findings.json"),
                       read(OLD / "events.json"), read(HERE / "new-review.json"),
                       read(BASE / "result.json"), sources, labels,
                       freeze["runtime_commit"])
    encoded = (json.dumps(summary, ensure_ascii=False, sort_keys=True, indent=2) + "\n").encode()
    if args.write:
        (HERE / "summary.json").write_bytes(encoded)
    else:
        require((HERE / "summary.json").read_bytes() == encoded,
                "recorded summary differs from source review")
    print(json.dumps({"summary_sha256": hashlib.sha256(encoded).hexdigest(),
                      "new_event_decisions": summary["new_event_decisions"],
                      "by_cohort": summary["profiles"]["technical"]["by_cohort"]},
                     sort_keys=True))


if __name__ == "__main__":
    main()
