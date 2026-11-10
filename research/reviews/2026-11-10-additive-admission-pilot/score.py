"""Check raw traces and score the frozen #349 additive pilot review."""

import json
from pathlib import Path
import sys

import run

ROOT = Path(__file__).resolve().parent
REVIEW = ROOT / "review.json"
RESULT = ROOT / "result.json"


def need(condition, message):
    if not condition:
        raise ValueError(message)


def overlap(a, b):
    return a["start"] < b["end"] and b["start"] < a["end"]


def collect():
    run.verify()
    journal = run.read(run.RUN / "run.json")
    need(journal["packet_sha256"] == run.file_sha(run.PACKET / "freeze.json"),
         "run packet binding")
    sources = run.source_rows()
    by_page = {row["page_id"]: row for row in sources}
    complete = journal["complete"] and journal["status"] == "complete"
    if not complete:
        need(journal["status"] == "stopped" and journal.get("stop_reason"),
             "unfinished run state")
        return journal, [], [], False
    pages, candidates, decisions, expected_calls = [], [], [], []
    for row in sources:
        page_id = row["page_id"]
        target = run.RUN / "generation" / page_id
        request = run.generation_request(row)
        need((target / "request.txt").read_bytes() == request,
             "generation request changed")
        response = run.read(target / "response.json")
        run.checked_trace(target, response)
        bound = run.bound_generation(response, row)
        need(run.read(target / "bound.json") == bound, "generation bound changed")
        pages.append(bound)
        expected_calls.append(("generation", page_id, run.sha(request),
                               run.file_sha(target / "bound.json")))
        candidates.extend(bound["findings"])
    need(len(candidates) <= run.MAX_ADMISSION, "admission count")
    for candidate in candidates:
        item_id = candidate["candidate_id"]
        target = run.RUN / "admission" / item_id
        request = run.admission_request(candidate, by_page[candidate["page_id"]])
        need((target / "request.txt").read_bytes() == request,
             "admission request changed")
        response = run.read(target / "response.json")
        run.checked_trace(target, response)
        bound = run.bound_admission(response, candidate)
        need(run.read(target / "bound.json") == bound, "admission bound changed")
        expected_calls.append(("admission", item_id, run.sha(request),
                               run.file_sha(target / "bound.json")))
        decisions.append(bound)
    need(len(journal["calls"]) == len(expected_calls), "journal call count")
    for entry, (stage, item_id, request_sha, bound_sha) in zip(
            journal["calls"], expected_calls):
        need(entry["stage"] == stage and entry["item_id"] == item_id and
             entry["status"] == "valid" and entry["request_sha256"] == request_sha
             and entry["bound_sha256"] == bound_sha and entry["exit_code"] == 0,
             "journal call changed")
    need(len(pages) == run.MAX_GENERATION and len(decisions) == len(candidates),
         "run coverage")
    return journal, candidates, decisions, True


def score(review_path=REVIEW):
    journal, candidates, decisions, complete = collect()
    reference = run.read(ROOT / "reference.json")
    events = {e["id"]: e for e in reference["events"]}
    retains = {e["id"]: e for e in reference["retain_controls"]}
    if not complete:
        return {"format": "unswell-349-additive-admission-result-v1",
                "status": "stopped", "qualified": False,
                "pilot_passed": False, "pages_required": run.MAX_GENERATION,
                "events_required": len(events), "calls_attempted": len(journal["calls"]),
                "stop_reason": journal["stop_reason"],
                "run_sha256": run.file_sha(run.RUN / "run.json"),
                "review_sha256": None}
    by_id = {c["candidate_id"]: c for c in candidates}
    need(len(by_id) == len(candidates), "candidate ID collision")
    admitted = {d["candidate_id"] for d in decisions if d["decision"] == "admit"}
    review = run.read(review_path)
    need(review["format"] == "unswell-349-additive-admission-review-v1" and
         review["reviewer"].startswith("Codex assistant,") and
         review["independent_human"] is False, "review identity")
    rows = review["findings"]
    need(len(rows) == len(admitted) and
         {row["candidate_id"] for row in rows} == admitted,
         "complete admitted review")
    credited, hit, accepted, uncertain, rejected = set(), set(), 0, 0, 0
    for row in rows:
        need(set(row) == {"candidate_id", "status", "rationale",
                          "credited_event_ids", "retain_hit_ids"}, "review row fields")
        need(row["status"] in ("accepted", "uncertain", "rejected") and
             row["rationale"].strip(), "review disposition")
        candidate = by_id[row["candidate_id"]]
        if row["status"] == "accepted":
            accepted += 1
        elif row["status"] == "uncertain":
            uncertain += 1
        else:
            rejected += 1
        need(len(set(row["credited_event_ids"])) == len(row["credited_event_ids"])
             and len(set(row["retain_hit_ids"])) == len(row["retain_hit_ids"]),
             "duplicate review credit")
        for event_id in row["credited_event_ids"]:
            event = events.get(event_id)
            need(row["status"] == "accepted" and event is not None and
                 event["page"] == candidate["page_id"] and
                 event["category"] == candidate["category"] and
                 overlap(event["span"], candidate["span"]),
                 "unsupported FULL credit")
            credited.add(event_id)
        for control_id in row["retain_hit_ids"]:
            control = retains.get(control_id)
            need(control is not None and control["page"] == candidate["page_id"]
                 and overlap(control["span"], candidate["span"]),
                 "unsupported retain-control hit")
            hit.add(control_id)
    counts = {x: sum(d["decision"] == x for d in decisions)
              for x in ("admit", "reject", "abstain")}
    need(counts["admit"] == accepted + uncertain + rejected, "review denominator")
    fraction = accepted / counts["admit"] if counts["admit"] else 0.0
    passed = (len(credited) >= 3 and fraction >= 0.85 and not hit)
    return {"format": "unswell-349-additive-admission-result-v1",
            "status": "complete", "qualified": False, "pilot_passed": passed,
            "pages_valid": run.MAX_GENERATION, "pages_required": run.MAX_GENERATION,
            "events_full": len(credited), "events_required": len(events),
            "credited_event_ids": sorted(credited),
            "retain_control_hits": sorted(hit),
            "generated": len(candidates), "admission": counts,
            "accepted": accepted, "uncertain": uncertain, "rejected": rejected,
            "accepted_fraction": fraction, "suggested_edits": 0,
            "edit_safety": "not_applicable",
            "run_sha256": run.file_sha(run.RUN / "run.json"),
            "review_sha256": run.file_sha(review_path)}


def write_result():
    need(not RESULT.exists(), "result already exists")
    data = score()
    RESULT.write_bytes(run.encoded(data))
    return data


def verify_result():
    need(RESULT.is_file(), "result missing")
    expected = score()
    need(run.read(RESULT) == expected, "result changed")
    return expected


if __name__ == "__main__":
    operation = sys.argv[1:] or ["verify"]
    result = write_result() if operation == ["score"] else verify_result()
    print(json.dumps(result, sort_keys=True))
