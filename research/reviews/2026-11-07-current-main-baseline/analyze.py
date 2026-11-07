#!/usr/bin/env python3
"""Verify and compare a fresh current-main replay with frozen #349 references."""

import argparse
from collections import Counter, defaultdict
import gzip
import hashlib
import json
from pathlib import Path
import sys


HERE = Path(__file__).resolve().parent
REPO = HERE.parents[2]
PRIOR = HERE.parent / "2026-09-26-outcome-announcements"
sys.path.insert(0, str(PRIOR))
import replay  # noqa: E402
import verify  # noqa: E402


def digest(data):
    return hashlib.sha256(data).hexdigest()


def require(condition, message):
    if not condition:
        raise ValueError(message)


def read(path):
    data = path.read_bytes()
    return json.loads(gzip.decompress(data) if path.suffix == ".gz" else data)


def location_key(finding):
    primary = finding["primary"]
    span = primary["span"]
    return (finding["rule_id"], primary["path"], span["start"], span["end"])


def overlap(left, right):
    return max(left["start"], right["start"]) < min(left["end"], right["end"])


def verify_report(report, sources, commit):
    require(report["status"] == "complete" and report["manifest"]["complete"] and
            not report["errors"] and not report["manifest"]["skipped_rules"],
            "incomplete current report")
    require(report["manifest"]["tool_commit"] == commit, "wrong runtime commit")
    documents = report["documents"]
    require(len(documents) == len(sources) == 36 and
            {doc["name"] for doc in documents} == set(sources), "source coverage changed")
    for doc in documents:
        require(doc["source_hash"] == digest(sources[doc["name"]]),
                "source hash mismatch: " + doc["name"])
    spans = defaultdict(list)
    for finding in report["findings"]:
        for loc in [finding["primary"], *finding["related"]]:
            name, span = loc["path"], loc["span"]
            require(name in sources, "finding outside frozen sources")
            source = sources[name]
            require(0 <= span["start"] < span["end"] <= len(source),
                    "finding span outside source")
            source[span["start"]:span["end"]].decode("utf-8")
            spans[name].append((finding["id"], span))
    return spans


def summarize(profile, current, old, events, sources, old_review, commit):
    spans = verify_report(current, sources, commit)
    old_counter = Counter(location_key(f) for f in old["findings"])
    new_counter = Counter(location_key(f) for f in current["findings"])
    retained = old_counter & new_counter
    removed = old_counter - new_counter
    added = new_counter - old_counter
    old_by_id = {f["id"]: location_key(f) for f in old["findings"]}
    old_full = {key: row["after"]["findings"]
                for key, row in old_review["profiles"][profile]["events"].items()
                if row["after"]["status"] == "full"}
    require(set(old_full) <= set(events), "old full witness outside frozen events")
    retained_witness_locations = sum(all(new_counter[old_by_id[fid]] > 0 for fid in ids)
                                     for ids in old_full.values())
    rows = []
    for key, event in sorted(events.items()):
        matches = []
        for target in event["targets"]:
            matches.append(sorted({fid for fid, span in spans[event["path"]]
                                   if overlap(target, span)}))
        rows.append({"event": key, "category": event["category"],
                     "target_count": len(matches),
                     "targets_with_overlap": sum(bool(match) for match in matches),
                     "possible_full_by_span": all(matches),
                     "overlap_finding_ids_by_target": matches})
    by_cohort = {}
    for cohort, required, denominator in (("whole", 46, 57), ("context", 37, 46),
                                           ("confirmation", 16, 20)):
        selected = [row for row in rows if row["event"].startswith(cohort + "/")]
        require(len(selected) == denominator, "frozen cohort denominator changed")
        by_cohort[cohort] = {"events": denominator, "required_full": required,
                             "possible_full_by_span": sum(row["possible_full_by_span"]
                                                          for row in selected)}
    return {"status": current["status"], "documents": len(current["documents"]),
            "findings": len(current["findings"]), "errors": len(current["errors"]),
            "skipped_rules": len(current["manifest"]["skipped_rules"]),
            "config_hash": current["manifest"]["config_hash"],
            "old_findings": len(old["findings"]),
            "retained_location_count": sum(retained.values()),
            "removed_location_count": sum(removed.values()),
            "added_location_count": sum(added.values()),
            "removed_by_rule": dict(sorted(Counter({key[0]: sum(v for k, v in removed.items()
                                                      if k[0] == key[0]) for key in removed}).items())),
            "added_by_rule": dict(sorted(Counter({key[0]: sum(v for k, v in added.items()
                                                    if k[0] == key[0]) for key in added}).items())),
            "old_full_events": len(old_full),
            "old_full_witness_locations_retained": retained_witness_locations,
            "by_cohort": by_cohort, "events": rows}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--run", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    freeze = read(HERE / "freeze.json")
    require(digest((HERE / "analyze.py").read_bytes()) == freeze["analysis_sha256"],
            "analysis code changed after freeze")
    require(digest((HERE / "PROTOCOL.md").read_bytes()) == freeze["protocol_sha256"],
            "protocol changed after freeze")
    for name, expected in freeze["inputs"].items():
        require(digest((REPO / name).read_bytes()) == expected, "frozen input changed: " + name)
    sources = replay.sources()
    events = verify.labels()
    require(len(events) == 123 and len(sources) == 36, "frozen reference count changed")
    run = read(args.run / "runs.json")
    prior_review = read(PRIOR / "review.json")
    profiles = {}
    for profile in ("technical", "strict"):
        record = run[profile]
        compressed = args.run / (profile + ".json.gz")
        require(digest(compressed.read_bytes()) == record["report_sha256"] and
                record["binary_sha256"] == freeze["binary_sha256"] and
                record["exit_code"] in (0, 1), "run identity mismatch")
        current = read(compressed)
        old = read(PRIOR / ("after-" + profile + ".json.gz"))
        profiles[profile] = {"report_sha256": record["report_sha256"],
                             **summarize(profile, current, old, events, sources,
                                         prior_review, freeze["base_commit"])}
    result = {"format": "unswell-current-main-overlap-ceiling-v1",
              "base_commit": freeze["base_commit"],
              "binary_sha256": freeze["binary_sha256"],
              "label_count": len(events), "source_count": len(sources),
              "profiles": profiles,
              "interpretation": "Span overlap is a generous necessary-condition ceiling, not accepted-event recall or finding precision."}
    encoded = (json.dumps(result, ensure_ascii=False, indent=2, sort_keys=True) + "\n").encode()
    if args.check:
        require(args.output.read_bytes() == encoded, "recorded analysis differs from replay")
    else:
        args.output.write_bytes(encoded)
    print(json.dumps({"result_sha256": digest(encoded),
                      "technical": profiles["technical"]["by_cohort"],
                      "strict": profiles["strict"]["by_cohort"]}, sort_keys=True))


if __name__ == "__main__":
    main()
