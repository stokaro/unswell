"""Score frozen real-source transfer decisions after separate edit review."""

import hashlib
import json
from pathlib import Path
import sys

import run

ROOT = Path(__file__).resolve().parent


def require(condition, message):
    if not condition:
        raise ValueError(message)


def review_rows(admitted):
    data = run.read(ROOT / "review.json")
    require(data["format"] == "unswell-349-real-source-transfer-edit-review-v2",
            "review format changed")
    require(data["reviewer"] == "Codex assistant under ADR 0041"
            and data["reviewer_role"] == "assistant, not a human annotator",
            "reviewer identity changed")
    rows = data["edits"]
    require(isinstance(rows, list) and len(rows) == len(admitted),
            "incomplete edit review")
    by_id = {}
    for row in rows:
        require(set(row) == {"item_id", "source_sha256", "replacement_sha256",
                             "safety", "rationale"}, "review fields changed")
        require(row["item_id"] in admitted and row["item_id"] not in by_id,
                "review item identity changed")
        response = admitted[row["item_id"]]
        require(row["source_sha256"] == response["source_sha256"]
                and row["replacement_sha256"] == hashlib.sha256(
                    response["replacement"].encode("utf-8")).hexdigest(),
                "review source or edit binding changed")
        require(row["safety"] in ("safe", "unsafe", "unresolved")
                and isinstance(row["rationale"], str)
                and row["rationale"].strip(), "review decision missing")
        by_id[row["item_id"]] = row
    return by_id


def calculate():
    protocol = run.verify()
    journal = run.read(run.RUN / "run.json")
    pairs = run.entries()
    require(journal["freeze_sha256"] == run.file_sha(run.PACKET / "freeze.json")
            and journal["model"] == protocol["model"]
            and journal["effort"] == protocol["effort"], "run identity changed")
    calls = journal["calls"]
    require(len(calls) <= run.MAX_CALLS and
            [row["item_id"] for row in calls] ==
            [payload["item_id"] for payload, _, _ in pairs[:len(calls)]],
            "run order changed")
    decisions = []
    admitted = {}
    for index, (payload, key, source_path) in enumerate(pairs):
        call = calls[index] if index < len(calls) else None
        decision = None
        response = None
        if call is not None and call["status"] == "valid":
            require(call["request_sha256"] == run.file_sha(
                        run.PACKET / "requests" / (payload["item_id"] + ".txt")),
                    "request hash changed")
            target = run.RUN / payload["item_id"]
            response = run.base().check_trace(target, payload)
            require(run.read(target / "bound.json") == response
                    and call["bound_sha256"] == run.file_sha(target / "bound.json"),
                    "bound response changed")
            decision = response["decision"]
            if decision == "admit":
                admitted[payload["item_id"]] = response
        elif call is not None:
            require(call["status"] == "invalid", "unfinished call")
        correct = (decision == "admit" if key["old_status"] == "accepted"
                   else decision in ("reject", "abstain"))
        decisions.append({"item_id": payload["item_id"],
                          "candidate_id": key["candidate_id"],
                          "path": source_path,
                          "old_status": key["old_status"],
                          "decision": decision,
                          "agrees_with_archived_grade": correct,
                          "source_sha256": payload["source_sha256"],
                          "target_start_byte": payload["target_start_byte"],
                          "target_end_byte": payload["target_end_byte"],
                          "basis": response["basis"] if response else None,
                          "replacement": response["replacement"] if response else None})
    reviews = review_rows(admitted)
    for row in decisions:
        row["edit_safety"] = reviews[row["item_id"]]["safety"] if row["item_id"] in reviews else None
    accepted_retained = sum(row["old_status"] == "accepted" and row["decision"] == "admit"
                            for row in decisions)
    uncertain_withheld = sum(row["old_status"] == "uncertain" and
                             row["decision"] in ("reject", "abstain")
                             for row in decisions)
    agreement = accepted_retained + uncertain_withheld
    complete = (journal["status"] == "complete" and journal["complete"]
                and len(calls) == run.MAX_CALLS
                and all(row["status"] == "valid" for row in calls))
    safety_clear = all(row["safety"] == "safe" for row in reviews.values())
    result = {"format": "unswell-349-real-source-transfer-result-v2",
              "freeze_sha256": run.file_sha(run.PACKET / "freeze.json"),
              "run_sha256": run.file_sha(run.RUN / "run.json"),
              "review_sha256": run.file_sha(ROOT / "review.json"),
              "total": run.MAX_CALLS, "calls": len(calls), "complete": complete,
              "accepted_total": 12, "accepted_retained": accepted_retained,
              "uncertain_total": 12, "uncertain_withheld": uncertain_withheld,
              "agreement": agreement, "admitted": len(admitted),
              "abstained": sum(row["decision"] == "abstain" for row in decisions),
              "safety_clear": safety_clear,
              "advance": (complete and agreement >= 21 and accepted_retained >= 10
                          and uncertain_withheld >= 10 and safety_clear),
              "decisions": decisions}
    return result


if __name__ == "__main__":
    result = calculate()
    (ROOT / "result.json").write_bytes(run.encoded(result))
    print(json.dumps({key: result[key] for key in
                      ("complete", "agreement", "accepted_retained",
                       "uncertain_withheld", "admitted", "safety_clear", "advance")},
                     sort_keys=True))
