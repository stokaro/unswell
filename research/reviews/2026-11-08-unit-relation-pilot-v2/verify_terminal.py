"""Replay the V2 first-error stop without scoring its invalid verdict."""

from collections import Counter
import json
from pathlib import Path
import sys

import packet
import run

ROOT = Path(__file__).resolve().parent
RUN = ROOT / "run"
VERDICT_FIELDS = {"ordinal", "decision", "start", "end", "quote",
                  "diagnosis", "evidence", "source_function", "deletion_test", "reason"}


def replay(row, target, request):
    if row["request_sha256"] != packet.sha(request) or \
            (target / "request.txt").read_bytes() != request or \
            (target / "stderr.txt").read_bytes():
        raise ValueError("request or stderr changed")
    raw = (target / "response.json").read_bytes()
    if row["response_sha256"] != packet.sha(raw):
        raise ValueError("response hash changed")
    response = json.loads(raw)
    run.trace(target, response)
    return response


def verify(write=False):
    protocol = run.verify()
    journal = packet.read(RUN / "run.json")
    calls = journal["calls"]
    if journal["status"] != "stopped_at_first_error" or journal["complete"] is not False or \
            journal["stop_reason"] != "invalid verdict fields" or \
            journal["model"] != protocol["model"] or journal["effort"] != protocol["effort"] or \
            journal["freeze_sha256"] != packet.sha((ROOT / "packet" / "freeze.json").read_bytes()) or \
            journal["inherited_from_v1"] != protocol["v1_terminal_sha256"] or \
            len(calls) != 3 or \
            [(row["page"], row["stage"], row["mode"], row["status"]) for row in calls] != [
                ("fastfetch", "inventory:fastfetch-0-88", "inherited", "valid"),
                ("fastfetch", "inventory:fastfetch-88-176", "new", "valid"),
                ("fastfetch", "verdict", "new", "failed")
            ] or calls[2]["error"] != journal["stop_reason"]:
        raise ValueError("terminal journal changed")
    if any((RUN / slug).exists() for slug in ("mise", "mkcert", "zellij")):
        raise ValueError("model calls continued after first error")

    responses = []
    for row, bounds in zip(calls[:2], packet.chunks("fastfetch")):
        cid = packet.chunk_id("fastfetch", bounds)
        target = RUN / "fastfetch" / "inventory" / cid
        request = (ROOT / "packet" / "requests" / (cid + ".txt")).read_bytes()
        response = replay(row, target, request)
        packet.validate_inventory_chunk("fastfetch", bounds, response)
        if row["mode"] == "inherited":
            for name in ("request.txt", "response.json", "events.jsonl", "stderr.txt"):
                if (target / name).read_bytes() != (ROOT / "inherited" / name).read_bytes():
                    raise ValueError("inherited V1 response changed")
        responses.append(response)
    merged_raw = packet.encoded(packet.merge_inventory("fastfetch", responses))
    if (RUN / "fastfetch" / "merged.json").read_bytes() != merged_raw:
        raise ValueError("merged inventory changed")

    target = RUN / "fastfetch" / "verdict"
    verdict = replay(calls[2], target, packet.verdict_request("fastfetch", merged_raw))
    if (target / "bound.json").exists():
        raise ValueError("invalid verdict was bound")
    try:
        packet.validate_verdict("fastfetch", verdict, merged_raw)
    except ValueError as exc:
        if str(exc) != journal["stop_reason"]:
            raise
    else:
        raise ValueError("invalid response unexpectedly accepted")
    expected = packet.candidates(json.loads(merged_raw))
    rows = verdict["verdicts"]
    if len(rows) != len(expected) or \
            [row["ordinal"] for row in rows] != [row["ordinal"] for row in expected] or \
            any(set(row) != VERDICT_FIELDS for row in rows):
        raise ValueError("candidate or response fields changed")
    counts = Counter(row["decision"] for row in rows)
    if counts != {"abstain": 7, "diagnosis": 5}:
        raise ValueError("raw verdict decisions changed")
    labels = packet.read(ROOT / "labels.json")
    target_ordinals = [event["overlapping_units"][0]["ordinal"]
                       for event in labels["events"]
                       if event["page"] == dict(packet.PAGES)["fastfetch"]]
    decisions = {row["ordinal"]: row["decision"] for row in rows}
    if target_ordinals != [70, 71, 81] or \
            [decisions[ordinal] for ordinal in target_ordinals] != [
                "abstain", "abstain", "diagnosis"]:
        raise ValueError("frozen-target raw decisions changed")
    result = {
        "format": "unswell-349-unit-relation-v2-terminal-v1",
        "status": "stopped_at_first_error",
        "valid_pages": 0,
        "attempted_new_calls": 2,
        "unexecuted_new_calls": protocol["calls_max"] - 2,
        "inherited_valid_responses": 1,
        "new_valid_inventory_responses": 1,
        "invalid_verdict_response_sha256": packet.sha((target / "response.json").read_bytes()),
        "invalid_field": "verdicts[*].decision",
        "invalid_value": "diagnosis",
        "expected_value": "diagnose",
        "fastfetch_candidate_count": len(expected),
        "raw_fastfetch_decision_counts": dict(counts),
        "frozen_fastfetch_target_ordinals": target_ordinals,
        "raw_fastfetch_target_decisions": [decisions[ordinal] for ordinal in target_ordinals],
        "official_editorial_quality_score": None,
        "registered_advance_possible_with_this_verdict": False,
        "reason": "The prompt and JSON schema did not constrain the exact decision token. "
                  "The raw answer abstained on two of three Fastfetch targets; even if "
                  "later pages recovered both remaining events, the 4/5 advance threshold could not be met."
    }
    path = ROOT / "terminal.json"
    if write:
        path.write_bytes(packet.encoded(result))
    elif path.read_bytes() != packet.encoded(result):
        raise ValueError("terminal result changed")
    return result


if __name__ == "__main__":
    if sys.argv[1:] not in (["--write"], ["--check"]):
        raise SystemExit("usage: verify_terminal.py --write|--check")
    print(verify(write=sys.argv[1] == "--write")["status"])
