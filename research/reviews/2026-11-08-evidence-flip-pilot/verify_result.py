"""Independently verify the frozen packet, raw traces, grades, and decision."""

import hashlib
import json
from pathlib import Path
import sys

ROOT = Path(__file__).resolve().parent
SOURCE = ROOT.parent / "2026-11-08-counterfactual-admission" / "cases.json"


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def require(ok, reason):
    if not ok:
        raise ValueError(reason)


def verify(base=ROOT):
    fixture = read(SOURCE)
    packet = base / "packet"
    run = base / "run"
    protocol = read(packet / "protocol.json")
    freeze = read(packet / "freeze.json")
    require(protocol["source_fixture_sha256"] == digest(SOURCE), "source fixture changed")
    for field, filename in (("runner_sha256", "run.py"),
                            ("scorer_sha256", "score.py"),
                            ("protocol_md_sha256", "PROTOCOL.md")):
        require(protocol[field] == digest(ROOT / filename), filename + " changed")
    actual_files = {str(path.relative_to(packet)): digest(path)
                    for path in packet.rglob("*") if path.is_file() and path.name != "freeze.json"}
    require(actual_files == freeze and len(freeze) == 19, "packet freeze changed")
    journal = read(run / "run.json")
    require(journal["status"] == "complete" and journal["complete"]
            and journal["freeze_sha256"] == digest(packet / "freeze.json")
            and journal["model"] == "gpt-6-astra" and journal["effort"] == "medium"
            and not journal["unattempted"] and len(journal["calls"]) == 16,
            "run identity or completeness changed")
    cases = {}
    for family in fixture["cases"]:
        for variant in family["variants"]:
            name = family["id"] + "/" + variant["id"]
            raw = variant["source"].encode("utf-8")
            target = family["target"].encode("utf-8")
            require(raw.count(target) == 1, "source target is ambiguous")
            cases[name] = {"source": variant["source"], "source_sha256": hashlib.sha256(raw).hexdigest(),
                           "target_quote": family["target"], "target_start_byte": raw.index(target),
                           "target_end_byte": raw.index(target) + len(target),
                           "proposed_diagnosis": family["diagnosis"],
                           "applies": variant["applies"], "role": variant["role"],
                           "rationale": variant["rationale"]}
    key = read(packet / "key.json")
    require(len(key) == 16 and len({row["item_id"] for row in key}) == 16,
            "answer key inventory changed")
    expected_ids = []
    decisions = []
    usage = {"input_tokens": 0, "output_tokens": 0, "reasoning_output_tokens": 0}
    for key_row in key:
        name = key_row["case_id"] + "/" + key_row["variant_id"]
        case = cases[name]
        require({field: key_row[field] for field in ("applies", "role", "rationale")}
                == {field: case[field] for field in ("applies", "role", "rationale")},
                "answer key label changed")
        item_id = hashlib.sha256((case["source_sha256"] + "\n" + case["proposed_diagnosis"]).encode("utf-8")).hexdigest()[:16]
        require(item_id == key_row["item_id"], "answer key ID changed")
        expected_ids.append(item_id)
        request_path = packet / "requests" / (item_id + ".txt")
        request_text = request_path.read_text(encoding="utf-8")
        require(request_text.count("INPUT JSON:\n") == 1, "request format changed")
        sent = json.loads(request_text.split("INPUT JSON:\n", 1)[1])
        expected_sent = {field: case[field] for field in
                         ("source", "source_sha256", "target_quote", "target_start_byte",
                          "target_end_byte", "proposed_diagnosis")}
        expected_sent["item_id"] = item_id
        require(sent == expected_sent, "request source or private-label boundary changed")
        target_dir = run / item_id
        require({path.name for path in target_dir.iterdir()} ==
                {"events.jsonl", "stderr.txt", "response.json", "bound.json"},
                "raw output file set changed")
        messages, turns = [], []
        for line in (target_dir / "events.jsonl").read_text(encoding="utf-8").splitlines():
            event = json.loads(line)
            kind = event.get("type")
            require(kind in {"thread.started", "turn.started", "turn.completed",
                             "item.started", "item.updated", "item.completed"},
                    "unexpected trace event")
            if kind.startswith("item."):
                item = event.get("item", {})
                require(item.get("type") in {"agent_message", "reasoning"},
                        "tool or error item in trace")
                if kind == "item.completed" and item["type"] == "agent_message":
                    messages.append(item["text"])
            if kind == "turn.completed":
                turns.append(event)
        response = read(target_dir / "response.json")
        require(len(turns) == 1 and messages and json.loads(messages[-1]) == response
                and response == read(target_dir / "bound.json"),
                "response or trace changed")
        require(set(response) == {"item_id", "source_sha256", "target_start_byte",
                                  "target_end_byte", "decision", "replacement", "basis"},
                "response schema changed")
        for field in ("item_id", "source_sha256", "target_start_byte", "target_end_byte"):
            require(type(response[field]) is type(expected_sent[field])
                    and response[field] == expected_sent[field],
                    "response source binding changed")
        require(response["decision"] in {"admit", "reject", "abstain"}
                and isinstance(response["replacement"], str)
                and isinstance(response["basis"], str) and response["basis"].strip(),
                "response decision or text invalid")
        require(response["decision"] == "admit" or not response["replacement"],
                "nonadmission has advice")
        require(response["decision"] != "admit" or response["replacement"] != case["target_quote"],
                "admitted advice unchanged")
        call = journal["calls"][len(decisions)]
        require(call["item_id"] == item_id and call["status"] == "valid"
                and call["exit_code"] == 0
                and call["request_sha256"] == digest(request_path)
                and call["bound_sha256"] == digest(target_dir / "bound.json")
                and call["decision"] == response["decision"],
                "journal call changed")
        for field in usage:
            value = turns[0].get("usage", {}).get(field)
            require(isinstance(value, int) and value >= 0, "usage missing")
            usage[field] += value
        correct = (response["decision"] == "admit") == case["applies"] and response["decision"] != "abstain"
        decisions.append({"item_id": item_id, "case_id": key_row["case_id"],
                          "variant_id": key_row["variant_id"],
                          "expected_applies": case["applies"], "decision": response["decision"],
                          "correct": correct, "replacement": response["replacement"],
                          "basis": response["basis"]})
    require(expected_ids == sorted(expected_ids), "call order changed")
    require({path.name for path in run.iterdir()} == set(expected_ids) | {"run.json"},
            "run directory inventory changed")
    review = read(base / "review.json")
    admitted = {row["item_id"] for row in decisions if row["decision"] == "admit"}
    edits = review["edits"]
    require(review["format"] == "unswell-349-evidence-flip-safety-review-v1"
            and review["independent_human"] is False
            and len(edits) == len(admitted)
            and {row["item_id"] for row in edits} == admitted
            and all(row["safety"] in {"safe", "unsafe", "unresolved"}
                    and row["rationale"].strip() for row in edits),
            "safety review inventory changed")
    index = {row["case_id"] + "/" + row["variant_id"]: row for row in decisions}
    pairs = [{"pair": pair, "both_correct": all(index[name]["correct"] for name in pair)}
             for pair in protocol["support_flip_pairs"]]
    result = read(base / "result.json")
    require(result["format"] == "unswell-349-evidence-flip-pilot-result-v1"
            and result["source_fixture_sha256"] == digest(SOURCE)
            and result["freeze_sha256"] == digest(packet / "freeze.json")
            and result["run_sha256"] == digest(run / "run.json")
            and result["safety_review_sha256"] == digest(base / "review.json")
            and result["complete"] is True and result["calls"] == 16
            and result["correct"] == sum(row["correct"] for row in decisions)
            and result["admitted"] == len(admitted)
            and result["abstained"] == sum(row["decision"] == "abstain" for row in decisions)
            and result["pair_results"] == pairs and result["decisions"] == decisions,
            "result content changed")
    safety_clear = all(row["safety"] == "safe" for row in edits)
    advance = result["correct"] >= 14 and all(row["both_correct"] for row in pairs) and safety_clear
    require(result["safety_clear"] == safety_clear and result["advance"] == advance,
            "result gate changed")
    return {"calls": len(decisions), "correct": result["correct"],
            "admitted": len(admitted), "support_flip_pairs_correct": sum(row["both_correct"] for row in pairs),
            "safety_clear": safety_clear, "advance": advance, "usage": usage,
            "freeze_sha256": digest(packet / "freeze.json"),
            "run_sha256": digest(run / "run.json"),
            "result_sha256": digest(base / "result.json")}


if __name__ == "__main__":
    base = Path(sys.argv[1]) if len(sys.argv) == 2 else ROOT
    print(json.dumps(verify(base), sort_keys=True))
