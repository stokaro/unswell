"""Score complete-source review of the frozen #349 repair-first pilot."""

import hashlib
import json
from pathlib import Path
import sys

import run

ROOT = Path(__file__).resolve().parent
PREVIOUS = ROOT.parent / "2026-11-10-additive-admission-pilot"


def need(condition, message):
    if not condition:
        raise ValueError(message)


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def overlap(left, right):
    return left["start"] < right["end"] and right["start"] < left["end"]


def evaluate(bound_pages, review, reference):
    edits = {item["edit_id"]: {**item, "page_id": page["page_id"]}
             for page in bound_pages for item in page["edits"]}
    need(len(edits) == sum(len(page["edits"]) for page in bound_pages),
         "duplicate edit ID")
    need(type(review) is dict and set(review) == {"edits", "events"},
         "review fields")
    edit_reviews = {}
    for item in review["edits"]:
        need(type(item) is dict and set(item) ==
             {"edit_id", "status", "safety", "rationale"}, "edit review fields")
        need(item["edit_id"] in edits and item["edit_id"] not in edit_reviews,
             "missing, unknown, or duplicate edit review")
        need(item["status"] in ("accepted", "uncertain", "rejected"),
             "edit status")
        need(item["safety"] in ("yes", "uncertain", "no"), "edit safety")
        need(type(item["rationale"]) is str and item["rationale"].strip(),
             "edit rationale")
        edit_reviews[item["edit_id"]] = item
    need(set(edit_reviews) == set(edits), "not every edit reviewed")
    events = {item["id"]: item for item in reference["events"]}
    need(len(events) == 4, "reference event inventory")
    event_reviews = {}
    for item in review["events"]:
        need(type(item) is dict and set(item) ==
             {"event_id", "status", "credited_edit_ids", "rationale"},
             "event review fields")
        event_id = item["event_id"]
        need(event_id in events and event_id not in event_reviews,
             "missing, unknown, or duplicate event review")
        need(item["status"] in ("FULL", "PARTIAL", "MISSED"),
             "event status")
        need(type(item["credited_edit_ids"]) is list and
             len(set(item["credited_edit_ids"])) == len(item["credited_edit_ids"]),
             "event edit IDs")
        need(type(item["rationale"]) is str and item["rationale"].strip(),
             "event rationale")
        target = events[event_id]
        for edit_id in item["credited_edit_ids"]:
            need(edit_id in edits and edits[edit_id]["page_id"] == target["page"],
                 "event edit page")
            need(overlap(edits[edit_id]["span"], target["span"]),
                 "event edit does not overlap frozen target")
        if item["status"] == "FULL":
            need(any(edit_reviews[edit_id]["status"] == "accepted" and
                     edit_reviews[edit_id]["safety"] == "yes"
                     for edit_id in item["credited_edit_ids"]),
                 "FULL lacks accepted safe edit")
        if item["status"] == "MISSED":
            need(not item["credited_edit_ids"], "MISSED has credited edit")
        event_reviews[event_id] = item
    need(set(event_reviews) == set(events), "not every event reviewed")
    controls = reference["retain_controls"]
    need(len(controls) == 4, "retain control inventory")
    hits = {control["id"]: [edit_id for edit_id, edit in edits.items()
                            if edit["page_id"] == control["page"] and
                            overlap(edit["span"], control["span"])]
            for control in controls}
    accepted = sum(item["status"] == "accepted" for item in edit_reviews.values())
    safe = sum(item["safety"] == "yes" for item in edit_reviews.values())
    accepted_safe = sum(item["status"] == "accepted" and item["safety"] == "yes"
                        for item in edit_reviews.values())
    full = sum(item["status"] == "FULL" for item in event_reviews.values())
    delivered = len(edits)
    return {
        "format": "unswell-349-repair-first-pilot-result-v1",
        "pages": len(bound_pages), "events": len(events),
        "delivered_edits": delivered,
        "full_events": full,
        "partial_events": sum(item["status"] == "PARTIAL"
                              for item in event_reviews.values()),
        "missed_events": sum(item["status"] == "MISSED"
                             for item in event_reviews.values()),
        "accepted_edits": accepted,
        "safe_edits": safe,
        "accepted_safe_edits": accepted_safe,
        "accepted_fraction": accepted / delivered if delivered else 0,
        "accepted_safe_fraction": accepted_safe / delivered if delivered else 0,
        "uncertain_edits": sum(item["status"] == "uncertain"
                               for item in edit_reviews.values()),
        "rejected_edits": sum(item["status"] == "rejected"
                              for item in edit_reviews.values()),
        "unsafe_edits": sum(item["safety"] == "no"
                            for item in edit_reviews.values()),
        "unresolved_safety": sum(item["safety"] == "uncertain"
                                 for item in edit_reviews.values()),
        "retain_control_hits": hits,
        "advance": (len(bound_pages) == 4 and full >= 3 and delivered > 0 and
                    accepted_safe / delivered >= 0.85 and
                    not any(hits.values())),
        "qualified": False,
    }


def assess():
    run.verify()
    journal = read(ROOT / "run" / "run.json")
    need(journal["status"] == "complete" and journal["complete"] and
         len(journal["calls"]) == 4 and
         all(item["status"] == "valid" for item in journal["calls"]),
         "run incomplete")
    reference = read(PREVIOUS / "reference.json")
    bound_pages = []
    for row in run.rows():
        page_id = row["page_id"]
        path = ROOT / "run" / page_id / "bound.json"
        bound = read(path)
        need(bound["page_id"] == page_id and
             bound["source_sha256"] == row["source_sha256"],
             "bound source changed")
        need(sha(path) == next(item["bound_sha256"] for item in
                               journal["calls"] if item["page_id"] == page_id),
             "bound hash changed")
        bound_pages.append(bound)
    review_path = ROOT / "review.json"
    result = evaluate(bound_pages, read(review_path), reference)
    result.update(packet_sha256=sha(ROOT / "packet" / "freeze.json"),
                  run_sha256=sha(ROOT / "run" / "run.json"),
                  review_sha256=sha(review_path),
                  reference_sha256=sha(PREVIOUS / "reference.json"))
    return result


if __name__ == "__main__":
    if len(sys.argv) != 2 or sys.argv[1] != "assess":
        raise SystemExit("usage: score.py assess")
    output = ROOT / "result.json"
    need(not output.exists(), "result already exists")
    result = assess()
    output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n",
                      encoding="utf-8")
    print(json.dumps(result, sort_keys=True))
