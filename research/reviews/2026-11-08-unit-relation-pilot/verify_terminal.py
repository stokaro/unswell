"""Verify the first-error stop of the V1 unit-relation pilot."""

from collections import Counter
import json
from pathlib import Path
import sys

import packet
import run

ROOT = Path(__file__).resolve().parent
RUN = ROOT / "run"
CHUNK = "fastfetch-0-88"


def verify(write=False):
    protocol = run.verify()
    journal = packet.read(RUN / "run.json")
    if journal["status"] != "stopped_at_first_error" or journal["complete"] is not False or \
            journal["stop_reason"] != "invalid inventory unit 0" or \
            journal["freeze_sha256"] != packet.sha((ROOT / "packet" / "freeze.json").read_bytes()) or \
            len(journal["calls"]) != 1:
        raise ValueError("terminal journal changed")
    row = journal["calls"][0]
    target = RUN / "fastfetch" / "inventory" / CHUNK
    request = (ROOT / "packet" / "requests" / (CHUNK + ".txt")).read_bytes()
    response_raw = (target / "response.json").read_bytes()
    response = json.loads(response_raw)
    if row["page"] != "fastfetch" or row["stage"] != "inventory:" + CHUNK or \
            row["status"] != "failed" or row["error"] != journal["stop_reason"] or \
            row["request_sha256"] != packet.sha(request) or \
            row["response_sha256"] != packet.sha(response_raw) or \
            (target / "request.txt").read_bytes() != request or \
            (target / "stderr.txt").read_bytes():
        raise ValueError("attempt bytes or status changed")
    run.trace(target, response)
    bounds = packet.chunks("fastfetch")[0]
    try:
        packet.validate_inventory_chunk("fastfetch", bounds, response)
    except ValueError as exc:
        if str(exc) != journal["stop_reason"]:
            raise
    else:
        raise ValueError("invalid V1 response unexpectedly accepted")
    if response["units"][0]["function"] != "Title":
        raise ValueError("first response changed")
    relations = Counter(x["relation"] for x in response["units"])
    labels = packet.read(ROOT / "labels.json")
    target_ordinals = [e["overlapping_units"][0]["ordinal"]
                       for e in labels["events"] if e["page"] == dict(packet.PAGES)["fastfetch"]
                       and e["overlapping_units"][0]["ordinal"] in range(*bounds)]
    exposed = [ordinal for ordinal in target_ordinals
               if response["units"][ordinal]["relation"] in packet.CANDIDATES]
    result = {"format": "unswell-349-unit-relation-v1-terminal-v1",
              "status": "stopped_at_first_error", "valid_pages": 0,
              "attempted_calls": 1, "unexecuted_calls": protocol["calls_max"] - 1,
              "first_response_sha256": packet.sha(response_raw),
              "source_units_returned": len(response["units"]),
              "invalid_field": "units[0].function", "invalid_value": "Title",
              "first_chunk_relation_counts": dict(relations),
              "frozen_target_ordinals_in_first_chunk": target_ordinals,
              "target_ordinals_selected_by_relation": exposed,
              "editorial_quality_score": None,
              "reason": "The inventory prompt omitted the allowed function vocabulary; no verdict stage ran."}
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
