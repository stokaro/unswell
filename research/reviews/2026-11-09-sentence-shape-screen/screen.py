#!/usr/bin/env python3
"""Audit source-eligible sentence-shape locations against frozen #349 events."""

import argparse
import collections
import hashlib
import json
import re
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
PATTERNS = {
    "significance-tail": r",\s+so\s+(?:a|an|the|it|that|this|there|nothing|no|tests?|you|we|they)\b",
    "contrastive-def": r"\b(?:rather than|instead of)\b",
    "contrastive-coda": r",\s+not\s+(?!only|just|merely|because|that|to|if|when|matter|longer|yet|so|such|more|less|fewer|doubt)[a-z'\"`][^.!?;:]*[.!?]",
    "negation-frame": r"\b(?:is not|are not|does not|nothing|neither)\b",
    "appositive-tail": r",\s+which\s+(?:is|means|makes|gives|lets|keeps|does)\b",
}
COHORTS = {"whole": 57, "context": 46, "confirmation": 20}
CURRENT_FULL = {"whole": 28, "context": 16, "confirmation": 4}


def sha256(data):
    return hashlib.sha256(data).hexdigest()


def read_json(path, expected_hash):
    data = path.read_bytes()
    if sha256(data) != expected_hash:
        raise ValueError(f"input SHA-256 mismatch: {path.name}")
    return json.loads(data)


def overlap(a, b, c, d):
    return a < d and b > c


def audit(archive_path):
    freeze = json.loads((HERE / "freeze.json").read_text())
    for name, path in (("script", HERE / "screen.py"), ("protocol", HERE / "PROTOCOL.md")):
        if sha256(path.read_bytes()) != freeze[name + "_sha256"]:
            raise ValueError(f"frozen {name} SHA-256 mismatch")
    archive = read_json(archive_path, freeze["archive_sha256"])
    decisions = read_json(ROOT / "research/reviews/2026-11-07-current-main-editorial-review/events.json", freeze["decisions_sha256"])
    scoped = read_json(ROOT / "research/reviews/2026-11-07-scoped-degree/result.json", freeze["scoped_result_sha256"])
    quality = read_json(ROOT / "research/reviews/2026-11-07-quality-claim-screen/result.json", freeze["quality_result_sha256"])
    originals = archive["original"]
    events = archive["reference"]["events"]
    paths = archive["paths"]
    if len(originals) != 36 or len(events) != 123 or len(decisions) != 123 or len(set(paths.values())) != 36:
        raise ValueError("source or event denominator changed")
    page_by_path = {path: page for page, path in paths.items()}
    if set(page_by_path.values()) != set(originals):
        raise ValueError("source path inventory changed")
    decision_by_id = {d["event"]: d["status"] for d in decisions}
    if len(decision_by_id) != 123 or set(decision_by_id) != set(events):
        raise ValueError("editorial decision inventory changed")
    if quality["profiles"]["technical"]["cohorts"].keys() != COHORTS.keys():
        raise ValueError("quality cohort inventory changed")

    events_by_page = collections.defaultdict(list)
    event_counts = collections.Counter()
    for event_id, event in events.items():
        cohort = event_id.split("/", 1)[0]
        if cohort not in COHORTS or event["path"] not in page_by_path:
            raise ValueError("unknown event cohort or source")
        event_counts[cohort] += 1
        page = page_by_path[event["path"]]
        source = originals[page]["document"]["source"].encode("utf-8")
        for target in event["targets"]:
            start, end = target["start"], target["end"]
            if not 0 <= start < end <= len(source) or source[start:end].decode("utf-8") != target["quote"]:
                raise ValueError(f"event target quote mismatch: {event_id}")
        events_by_page[page].append((event_id, event))
    if dict(event_counts) != COHORTS:
        raise ValueError("event cohort denominator changed")

    current_full = {event_id for event_id, status in decision_by_id.items() if status == "full"}
    scoped_added = scoped["profiles"]["technical"]["added_locations"]
    quality_added = quality["profiles"]["technical"]["added_locations"]
    if len(scoped_added) != 2 or len(quality_added) != 5:
        raise ValueError("reviewed increment changed")
    for path, start, end in scoped_added + quality_added:
        page = page_by_path[path]
        credited = [event_id for event_id, event in events_by_page[page]
                    if any(overlap(start, end, t["start"], t["end"]) for t in event["targets"])]
        if len(credited) != 1:
            raise ValueError("quality claim location has ambiguous event credit")
        current_full.add(credited[0])
    full_counts = collections.Counter(event_id.split("/", 1)[0] for event_id in current_full)
    if dict(full_counts) != CURRENT_FULL:
        raise ValueError("current FULL baseline changed")
    for cohort, totals in quality["profiles"]["technical"]["cohorts"].items():
        if totals["events"] != COHORTS[cohort] or totals["full"] != CURRENT_FULL[cohort]:
            raise ValueError("quality result cohort accounting changed")

    compiled = {name: re.compile(expression, re.IGNORECASE) for name, expression in PATTERNS.items()}
    raw_hits = collections.Counter()
    protected_hits = collections.Counter()
    eligible_hits = collections.Counter()
    target_hits = collections.Counter()
    non_target_hits = collections.Counter()
    event_hits = collections.defaultdict(set)
    target_touches = collections.defaultdict(set)
    event_hits_by_pattern = collections.defaultdict(set)
    for page, original in originals.items():
        document = original["document"]
        source = document["source"].encode("utf-8")
        if sha256(source) != document["source_sha256"]:
            raise ValueError(f"source SHA-256 mismatch: {page}")
        protected = [(p["span"]["start"], p["span"]["end"]) for p in document["protected"]]
        for unit in original["raw_units"]:
            start, end = unit["span"]["start"], unit["span"]["end"]
            raw = unit["raw"]
            if not 0 <= start <= end <= len(source) or source[start:end].decode("utf-8") != raw:
                raise ValueError(f"raw unit byte slice mismatch: {page}/{unit['id']}")
            for name, expression in compiled.items():
                for match in expression.finditer(raw):
                    raw_hits[name] += 1
                    match_start = start + len(raw[:match.start()].encode("utf-8"))
                    match_end = start + len(raw[:match.end()].encode("utf-8"))
                    if any(overlap(match_start, match_end, a, b) for a, b in protected):
                        protected_hits[name] += 1
                        continue
                    eligible_hits[name] += 1
                    touched = False
                    for event_id, event in events_by_page[page]:
                        for index, target in enumerate(event["targets"]):
                            if overlap(match_start, match_end, target["start"], target["end"]):
                                touched = True
                                event_hits[event_id].add(name)
                                event_hits_by_pattern[name].add(event_id)
                                target_touches[event_id].add(index)
                    if touched:
                        target_hits[name] += 1
                    else:
                        non_target_hits[name] += 1
    if sum(raw_hits.values()) != sum(protected_hits.values()) + sum(eligible_hits.values()):
        raise ValueError("match accounting failed")
    all_touched = {event_id for event_id, names in event_hits.items()
                   if len(target_touches[event_id]) == len(events[event_id]["targets"])}
    non_full = set(events) - current_full
    cohort_new = collections.Counter(eid.split("/", 1)[0] for eid in event_hits if eid in non_full)
    cohort_all_touched = collections.Counter(eid.split("/", 1)[0] for eid in all_touched if eid in non_full)
    return {
        "format": "unswell-sentence-shape-reachability-v1",
        "status": "exposed-exploratory-location-only",
        "source_repository_commit": freeze["source_repository_commit"],
        "unswell_main_commit": freeze["unswell_main_commit"],
        "input_sha256": {key: freeze[key] for key in ("archive_sha256", "decisions_sha256", "scoped_result_sha256", "quality_result_sha256")},
        "source_pages": len(originals),
        "frozen_events": dict(event_counts),
        "current_full": dict(full_counts),
        "matches": {
            name: {
                "raw": raw_hits[name], "excluded_protected": protected_hits[name],
                "eligible": eligible_hits[name], "touching_any_frozen_target": target_hits[name],
                "outside_frozen_targets": non_target_hits[name],
                "events_touched": len(event_hits_by_pattern[name]),
                "previously_non_full_events_touched": len(event_hits_by_pattern[name] & non_full),
            } for name in PATTERNS
        },
        "total_eligible_matches": sum(eligible_hits.values()),
        "total_matches_outside_frozen_targets": sum(non_target_hits.values()),
        "events_with_any_target_overlap": len(event_hits),
        "previously_non_full_events_with_any_overlap": len(set(event_hits) & non_full),
        "previously_non_full_by_cohort": {cohort: cohort_new[cohort] for cohort in COHORTS},
        "previously_non_full_with_every_target_span_touched": len(all_touched & non_full),
        "every_target_span_touched_by_cohort": {cohort: cohort_all_touched[cohort] for cohort in COHORTS},
        "previously_non_full_event_ids": sorted(set(event_hits) & non_full),
    }


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--archive", required=True, type=Path)
    parser.add_argument("--out", type=Path)
    args = parser.parse_args()
    result = audit(args.archive)
    output = json.dumps(result, sort_keys=True, indent=2) + "\n"
    if args.out:
        args.out.write_text(output)
    else:
        print(output, end="")


if __name__ == "__main__":
    main()
