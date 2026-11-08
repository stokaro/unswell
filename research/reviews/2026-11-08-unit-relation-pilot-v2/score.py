"""Score a complete source-only assistant review of the frozen pilot stream."""

from collections import Counter
import json
from pathlib import Path
import sys

import packet

ROOT = Path(__file__).resolve().parent
RUN = ROOT / "run"


def key(row):
    return (row["page"], row["ordinal"], row["start"], row["end"], row["quote"])


def score():
    protocol = packet.read(ROOT / "packet" / "protocol.json")
    labels = packet.read(ROOT / "labels.json")
    review = packet.read(ROOT / "review.json")
    journal = packet.read(RUN / "run.json")
    if journal["status"] != "complete" or journal["complete"] is not True or \
            len(journal["calls"]) != protocol["total_responses"] or \
            sum(x.get("mode") == "new" for x in journal["calls"]) != protocol["calls_max"] or \
            sum(x.get("mode") == "inherited" for x in journal["calls"]) != protocol["reused_calls"] or \
            journal["freeze_sha256"] != packet.sha((ROOT / "packet" / "freeze.json").read_bytes()):
        raise ValueError("pilot run incomplete or unbound")
    if review["format"] != "unswell-349-unit-relation-assistant-review-v1" or \
            review["reviewer"] != labels["reviewer"] or \
            review["basis"] != "complete delivered stream on all four pinned pages" or \
            type(review["findings"]) is not list:
        raise ValueError("missing complete assistant review")
    delivered, candidate_coverage, candidate_counts, abstentions = [], {}, {}, {}
    for slug, page_id in packet.PAGES:
        merged_raw = (RUN / slug / "merged.json").read_bytes()
        merged = json.loads(merged_raw)
        selected = packet.candidates(merged)
        candidate_counts[slug] = len(selected)
        bound = packet.read(RUN / slug / "verdict" / "bound.json")
        if bound["page_id"] != page_id or bound["source_sha256"] != merged["source_sha256"] or \
                bound["candidates"] != len(selected) or \
                bound["abstentions"] != len(selected) - len(bound["findings"]):
            raise ValueError("bound verdict counts changed")
        abstentions[slug] = bound["abstentions"]
        for finding in bound["findings"]:
            delivered.append(dict(finding, page=page_id))
        selected_ids = {x["ordinal"] for x in selected}
        for event in (e for e in labels["events"] if e["page"] == page_id):
            candidate_coverage[event["id"]] = any(
                x["ordinal"] in selected_ids for x in event["overlapping_units"])
    if len(review["findings"]) != len(delivered):
        raise ValueError("review does not cover every delivered finding")
    review_by_key = {}
    event_ids = {e["id"] for e in labels["events"]}
    for row in review["findings"]:
        if set(row) != {"page", "ordinal", "start", "end", "quote", "status", "rationale",
                        "covered_events", "partial_events"} or \
                row["status"] not in {"accepted", "uncertain", "nonactionable"} or \
                not row["rationale"].strip() or type(row["covered_events"]) is not list or \
                type(row["partial_events"]) is not list:
            raise ValueError("invalid finding review")
        if set(row["covered_events"] + row["partial_events"]) - event_ids or \
                len(set(row["covered_events"] + row["partial_events"])) != \
                len(row["covered_events"] + row["partial_events"]):
            raise ValueError("invalid event attribution")
        if row["status"] != "accepted" and (row["covered_events"] or row["partial_events"]):
            raise ValueError("nonaccepted finding cannot witness event")
        identity = key(row)
        if identity in review_by_key:
            raise ValueError("duplicate review row")
        review_by_key[identity] = row
    if set(review_by_key) != {key(x) for x in delivered}:
        raise ValueError("review finding identities changed")
    for row in review["findings"]:
        for event_id in row["covered_events"] + row["partial_events"]:
            event = next(e for e in labels["events"] if e["id"] == event_id)
            target = event["span"]
            if row["page"] != event["page"] or \
                    row["start"] >= target["end"] or row["end"] <= target["start"]:
                raise ValueError("event witness misses frozen source target")
    full = {event for row in review["findings"] for event in row["covered_events"]}
    partial = {event for row in review["findings"] for event in row["partial_events"]} - full
    grades = Counter(row["status"] for row in review["findings"])
    accepted = grades["accepted"]
    delivered_count = len(delivered)
    fraction = accepted / delivered_count if delivered_count else None
    source_controls = []
    for example in labels["counterexamples"]:
        slug = next(s for s, p in packet.PAGES if p == example["page"])
        _, raw, _ = packet.page(slug)
        start = raw.index(example["quote"].encode("utf-8"))
        end = start + len(example["quote"].encode("utf-8"))
        for row in delivered:
            if row["page"] == example["page"] and row["start"] < end and row["end"] > start:
                source_controls.append({"page": example["page"], "quote": example["quote"]})
    advance = protocol["advance"]
    result = {"format": "unswell-349-unit-relation-result-v2",
              "role": protocol["role"], "model": protocol["model"], "effort": protocol["effort"],
              "new_model_calls": protocol["calls_max"], "inherited_model_responses": protocol["reused_calls"],
              "pages": len(packet.PAGES), "valid_pages": len(packet.PAGES),
              "frozen_events": len(labels["events"]), "full_events": len(full),
              "partial_events": len(partial), "missed_events": len(labels["events"]) - len(full) - len(partial),
              "event_status": {e["id"]: "full" if e["id"] in full else "partial" if e["id"] in partial else "missed"
                               for e in labels["events"]},
              "candidate_stage_event_coverage": sum(candidate_coverage.values()),
              "candidate_stage_by_event": candidate_coverage,
              "candidate_counts": candidate_counts, "abstentions": abstentions,
              "delivered": delivered_count, "accepted": accepted, "uncertain": grades["uncertain"],
              "nonactionable": grades["nonactionable"],
              "accepted_fraction": fraction,
              "protected_counterexamples_flagged": len(source_controls),
              "protected_counterexample_details": source_controls,
              "supplied_edits": 0, "suggestion_safety": "not applicable; no replacement edits requested or supplied",
              "registered_advance": len(full) >= advance["minimum_full_events"] and
                                    fraction is not None and fraction >= advance["minimum_accepted_delivery_fraction"] and
                                    len(source_controls) <= advance["maximum_protected_counterexamples_flagged"] and
                                    len(packet.PAGES) == advance["required_valid_pages"]}
    return result


if __name__ == "__main__":
    if sys.argv[1:] != ["--write"]:
        raise SystemExit("usage: score.py --write")
    result = score()
    (ROOT / "result.json").write_bytes(packet.encoded(result))
    print(json.dumps(result, indent=2, sort_keys=True))
