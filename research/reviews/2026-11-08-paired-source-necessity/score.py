"""Score frozen paired necessity decisions against prior assistant grades."""

import json
from pathlib import Path
import sys

import run

ROOT = Path(__file__).resolve().parent


def require(condition, message):
    if not condition:
        raise ValueError(message)


def baseline_correct(pairs):
    count = 0
    for payload, key in pairs:
        decisions = {}
        for slot in ("A", "B"):
            item_id = payload["candidates"][slot]["item_id"]
            old = run.read(run.PRIOR / "run" / item_id / "response.json")
            decisions[slot] = old["decision"] == "admit"
        count += all(decisions[slot] == (key[slot + "_old_status"] == "accepted")
                     for slot in ("A", "B"))
    return count


def score():
    protocol = run.verify()
    pairs = run.selected()
    require(baseline_correct(pairs) == protocol["baseline_correct_pairs"],
            "independent baseline changed")
    journal = run.read(run.RUN / "run.json")
    require(journal["freeze_sha256"] == run.file_sha(run.PACKET / "freeze.json")
            and journal["model"] == run.MODEL and journal["effort"] == run.EFFORT,
            "run identity changed")
    calls = journal["calls"]
    require([row["pair_id"] for row in calls] ==
            [payload["pair_id"] for payload, _ in pairs[:len(calls)]],
            "run order changed")
    require(len(calls) <= run.MAX_CALLS and
            journal["unattempted"] ==
            [payload["pair_id"] for payload, _ in pairs[len(calls):]],
            "run call limit or unattempted list changed")
    valid = []
    for (payload, key), row in zip(pairs, calls):
        pair_id = payload["pair_id"]
        require(row["request_sha256"] ==
                run.file_sha(run.PACKET / "requests" / (pair_id + ".txt")),
                "request identity changed")
        if row["status"] != "valid":
            require(row["status"] == "invalid" and row == calls[-1],
                    "invalid call was retried or followed")
            break
        target = run.RUN / pair_id
        checked = run.check_trace(target, payload)
        require(run.read(target / "bound.json") == checked and
                row["bound_sha256"] == run.file_sha(target / "bound.json") and
                row["A_decision"] == checked["A_decision"] and
                row["B_decision"] == checked["B_decision"],
                "bound response or journal changed")
        valid.append((key, checked))
    complete = len(valid) == run.MAX_CALLS
    require(journal["complete"] is complete and
            journal["status"] == ("complete" if complete else "stopped"),
            "run completion claim changed")
    retained = withheld = exact = 0
    for key, response in valid:
        matched = True
        for slot in ("A", "B"):
            old = key[slot + "_old_status"]
            revise = response[slot + "_decision"] == "revise"
            if old == "accepted":
                retained += revise
                matched &= revise
            else:
                withheld += not revise
                matched &= not revise
        exact += matched
    advance = (complete and exact >= 6 and retained >= 7 and withheld >= 6)
    return {"format": "unswell-349-paired-source-necessity-result-v1",
            "scope": "exposed_development_directional_pilot",
            "complete": complete, "valid_pairs": len(valid),
            "source_bound_targets": 2 * len(valid),
            "baseline_independent_correct_pairs": baseline_correct(pairs),
            "paired_correct_pairs": exact,
            "prior_accepted_retained": retained,
            "prior_uncertain_withheld": withheld,
            "advance": advance,
            "qualified": False,
            "reason": "Directional exposed pilot only; full #349 development, advice safety, and separate prospective confirmation remain."}


if __name__ == "__main__":
    result = score()
    target = ROOT / "result.json"
    if target.exists():
        raise SystemExit("result already exists")
    target.write_bytes(run.encoded(result))
    print(json.dumps(result, sort_keys=True))
