"""Score the frozen paired rubric screen after separate edit review."""

import hashlib
import json
from pathlib import Path

import run


ROOT = Path(__file__).resolve().parent


def require(condition, reason):
    if not condition:
        raise ValueError(reason)


def review_rows(admitted):
    data = run.read(ROOT / "review.json")
    require(data["format"] == "unswell-349-rubric-transfer-edit-review-v1" and
            data["reviewer"] == "Codex assistant under ADR 0041" and
            data["reviewer_role"] == "assistant, not a human annotator",
            "edit reviewer identity changed")
    rows = data["edits"]
    require(isinstance(rows, list) and len(rows) == len(admitted),
            "admitted edit review incomplete")
    found = {}
    for row in rows:
        require(set(row) == {"item_id", "source_sha256", "replacement_sha256",
                             "safety", "rationale"}, "edit review fields changed")
        require(row["item_id"] in admitted and row["item_id"] not in found,
                "edit review identity changed")
        response = admitted[row["item_id"]]
        require(row["source_sha256"] == response["source_sha256"] and
                row["replacement_sha256"] == hashlib.sha256(
                    response["replacement"].encode("utf-8")).hexdigest(),
                "edit review source or replacement binding changed")
        require(row["safety"] in ("safe", "unsafe", "unresolved") and
                isinstance(row["rationale"], str) and row["rationale"].strip(),
                "edit review missing safety rationale")
        found[row["item_id"]] = row
    return found


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
            [payload["item_id"] for payload, _, _, _, _ in pairs[:len(calls)]],
            "run call order changed")
    rows, admitted = [], {}
    for index, (payload, case_id, arm, grade, path) in enumerate(pairs):
        call = calls[index] if index < len(calls) else None
        response = None
        decision = None
        if call is not None:
            require(call["case_id"] == case_id and call["arm"] == arm and
                    call["request_sha256"] == run.file_sha(
                        run.PACKET / "requests" / (payload["item_id"] + ".txt")),
                    "call identity or request changed")
            if call["status"] == "valid":
                target = run.RUN / payload["item_id"]
                response = run.base().check_trace(target, payload)
                require(run.read(target / "bound.json") == response and
                        call["bound_sha256"] == run.file_sha(target / "bound.json"),
                        "bound response changed")
                decision = response["decision"]
                if decision == "admit":
                    admitted[payload["item_id"]] = response
            else:
                require(call["status"] == "invalid", "unfinished call in scored run")
        agrees = (decision == "admit" if grade["old_status"] == "accepted"
                  else decision in ("reject", "abstain"))
        rows.append({"item_id": payload["item_id"], "case_id": case_id,
                     "arm": arm, "candidate_id": grade["candidate_id"],
                     "path": path, "old_status": grade["old_status"],
                     "decision": decision, "agrees_with_archived_grade": agrees,
                     "source_sha256": payload["source_sha256"],
                     "target_start_byte": payload["target_start_byte"],
                     "target_end_byte": payload["target_end_byte"],
                     "basis": response["basis"] if response else None,
                     "replacement": response["replacement"] if response else None})
    reviews = review_rows(admitted)
    for row in rows:
        row["edit_safety"] = (reviews[row["item_id"]]["safety"]
                              if row["item_id"] in reviews else None)
    complete = (journal["status"] == "complete" and journal["complete"]
                and len(calls) == run.MAX_CALLS
                and all(row["status"] == "valid" for row in calls))
    arms = {}
    for arm in run.ARMS:
        subset = [row for row in rows if row["arm"] == arm]
        edited = [row for row in subset if row["decision"] == "admit"]
        arms[arm] = {
            "agreement": sum(row["agrees_with_archived_grade"] for row in subset),
            "accepted_retained": sum(row["old_status"] == "accepted" and
                                     row["decision"] == "admit" for row in subset),
            "uncertain_withheld": sum(row["old_status"] == "uncertain" and
                                      row["decision"] in ("reject", "abstain")
                                      for row in subset),
            "admitted": len(edited),
            "abstained": sum(row["decision"] == "abstain" for row in subset),
            "edit_safety": {s: sum(row["edit_safety"] == s for row in edited)
                            for s in ("safe", "unsafe", "unresolved")},
        }
    by_case = {}
    for row in rows:
        by_case.setdefault(row["case_id"], {})[row["arm"]] = row["decision"]
    require(len(by_case) == 12 and all(set(pair) == set(run.ARMS)
                                       for pair in by_case.values()),
            "paired results incomplete")
    delta = arms["rubric"]["agreement"] - arms["control"]["agreement"]
    safety_clear = (arms["rubric"]["edit_safety"]["unsafe"] == 0 and
                    arms["rubric"]["edit_safety"]["unresolved"] == 0)
    return {
        "format": "unswell-349-rubric-transfer-result-v1",
        "freeze_sha256": run.file_sha(run.PACKET / "freeze.json"),
        "run_sha256": run.file_sha(run.RUN / "run.json"),
        "review_sha256": run.file_sha(ROOT / "review.json"),
        "total_calls": run.MAX_CALLS, "calls": len(calls), "complete": complete,
        "cases": 12, "accepted_total": 6, "uncertain_total": 6,
        "arms": arms, "paired_decisions": by_case, "agreement_delta": delta,
        "safety_clear": safety_clear,
        "advance": (complete and arms["rubric"]["agreement"] >= 10 and delta >= 3
                    and arms["rubric"]["accepted_retained"] >= 5
                    and arms["rubric"]["uncertain_withheld"] >= 5
                    and safety_clear),
        "decisions": rows,
    }


if __name__ == "__main__":
    result = calculate()
    (ROOT / "result.json").write_bytes(run.encoded(result))
    print(json.dumps({key: value for key, value in result.items()
                      if key not in ("decisions", "paired_decisions")}, sort_keys=True))
