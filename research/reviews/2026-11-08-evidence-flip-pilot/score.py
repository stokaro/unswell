"""Replay the frozen source-bound evidence-flip pilot and score its controls."""

import json
from pathlib import Path
import sys

import run

ROOT = Path(__file__).resolve().parent
REVIEW = ROOT / "review.json"
RESULT = ROOT / "result.json"


def evaluate(review_path=REVIEW):
    protocol = run.verify()
    pairs = run.entries()
    journal = run.read(run.RUN / "run.json")
    if (journal["freeze_sha256"] != run.file_sha(run.PACKET / "freeze.json")
            or journal["model"] != protocol["model"] or journal["effort"] != protocol["effort"]):
        raise ValueError("run identity changed")
    if len(journal["calls"]) > run.MAX_CALLS:
        raise ValueError("call budget exceeded")
    if any(row["status"] != "valid" for row in journal["calls"][:-1]):
        raise ValueError("execution continued after a failed request")
    if [row["item_id"] for row in journal["calls"]] != [
            payload["item_id"] for payload, _ in pairs[:len(journal["calls"])]]:
        raise ValueError("call order changed")
    decisions = []
    for (payload, key), call in zip(pairs, journal["calls"]):
        if call["request_sha256"] != run.file_sha(run.PACKET / "requests" / (payload["item_id"] + ".txt")):
            raise ValueError("request hash changed")
        if call["status"] != "valid":
            break
        target = run.RUN / payload["item_id"]
        response = run.check_trace(target, payload)
        if run.file_sha(target / "bound.json") != call["bound_sha256"]:
            raise ValueError("bound hash changed")
        if run.read(target / "bound.json") != response:
            raise ValueError("bound response changed")
        decisions.append({"item_id": payload["item_id"],
                          "case_id": key["case_id"], "variant_id": key["variant_id"],
                          "expected_applies": key["applies"],
                          "decision": response["decision"],
                          "correct": (response["decision"] == "admit") == key["applies"]
                          and response["decision"] != "abstain",
                          "replacement": response["replacement"],
                          "basis": response["basis"]})
    complete = (journal["status"] == "complete" and journal["complete"]
                and len(journal["calls"]) == run.MAX_CALLS
                and len(decisions) == run.MAX_CALLS
                and not journal["unattempted"])
    if journal["status"] == "complete" and not complete:
        raise ValueError("run completeness changed")
    by_name = {row["case_id"] + "/" + row["variant_id"]: row for row in decisions}
    pair_results = []
    for pair in protocol["support_flip_pairs"]:
        pair_results.append({"pair": pair,
                             "both_correct": all(by_name.get(name, {}).get("correct", False)
                                                 for name in pair)})
    admitted = {row["item_id"] for row in decisions if row["decision"] == "admit"}
    reviews = {}
    if review_path.is_file():
        review = run.read(review_path)
        if review.get("format") != "unswell-349-evidence-flip-safety-review-v1":
            raise ValueError("safety review format changed")
        rows = review.get("edits")
        if not isinstance(rows, list) or {row.get("item_id") for row in rows} != admitted:
            raise ValueError("safety review inventory changed")
        for row in rows:
            if (set(row) != {"item_id", "safety", "rationale"}
                    or row["safety"] not in ("safe", "unsafe", "unresolved")
                    or not isinstance(row["rationale"], str) or not row["rationale"].strip()
                    or row["item_id"] in reviews):
                raise ValueError("safety review row invalid")
            reviews[row["item_id"]] = row
    correct = sum(row["correct"] for row in decisions)
    all_pairs = len(pair_results) == 4 and all(row["both_correct"] for row in pair_results)
    safety_clear = len(reviews) == len(admitted) and all(
        row["safety"] == "safe" for row in reviews.values())
    result = {"format": "unswell-349-evidence-flip-pilot-result-v1",
              "source_fixture_sha256": protocol["source_fixture_sha256"],
              "freeze_sha256": run.file_sha(run.PACKET / "freeze.json"),
              "run_sha256": run.file_sha(run.RUN / "run.json"),
              "safety_review_sha256": run.file_sha(review_path) if review_path.is_file() else None,
              "complete": complete, "calls": len(journal["calls"]),
              "correct": correct, "total": run.MAX_CALLS,
              "admitted": len(admitted),
              "abstained": sum(row["decision"] == "abstain" for row in decisions),
              "pair_results": pair_results,
              "safety_clear": safety_clear,
              "advance": complete and correct >= 14 and all_pairs and safety_clear,
              "decisions": decisions}
    return result


if __name__ == "__main__":
    result = evaluate()
    if sys.argv[1:] == ["write"]:
        if RESULT.exists():
            raise ValueError("result already exists")
        RESULT.write_bytes(run.encoded(result))
    elif sys.argv[1:] != ["check"]:
        raise ValueError("use check or write")
    print(json.dumps({key: result[key] for key in
                      ("complete", "calls", "correct", "total", "admitted",
                       "abstained", "safety_clear", "advance")}, sort_keys=True))
