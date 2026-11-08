"""Independently replay the frozen real-source admission transfer result."""

from collections import Counter
import hashlib
import importlib.util
import json
from pathlib import Path
import sys

ROOT = Path(__file__).resolve().parent
V1 = ROOT.parent / "2026-11-08-real-source-admission-transfer"
BASE = ROOT.parent / "2026-11-08-evidence-flip-pilot" / "run.py"
REPLAY = ROOT.parent / "2026-09-26-outcome-announcements" / "replay.py"


def require(condition, reason):
    if not condition:
        raise ValueError(reason)


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def sha(data):
    return hashlib.sha256(data).hexdigest()


def module(path, name):
    spec = importlib.util.spec_from_file_location(name, path)
    result = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(result)
    return result


def verify(base=ROOT, v1=V1):
    packet = base / "packet"
    folder = base / "run"
    protocol = read(packet / "protocol.json")
    require(protocol["format"] == "unswell-349-real-source-admission-transfer-v2"
            and protocol["model"] == "gpt-6-astra"
            and protocol["effort"] == "medium"
            and protocol["calls_max"] == 24
            and protocol["attempts_per_request"] == 1
            and protocol["automatic_retries"] is False
            and protocol["tools_allowed"] is False
            and protocol["per_call_timeout_seconds"] == 600
            and protocol["total_timeout_seconds"] == 10800,
            "protocol execution contract changed")
    for field, path in (("base_runner_sha256", BASE),
                        ("runner_sha256", base / "run.py"),
                        ("scorer_sha256", base / "score.py"),
                        ("protocol_md_sha256", base / "PROTOCOL.md"),
                        ("sample_sha256", packet / "sample.json"),
                        ("answer_key_sha256", packet / "key.json")):
        require(protocol[field] == digest(path), "protocol hash changed: " + field)
    freeze = read(packet / "freeze.json")
    files = {str(path.relative_to(packet)): digest(path)
             for path in packet.rglob("*")
             if path.is_file() and path.name != "freeze.json"}
    require(files == freeze and len(files) == 28, "packet file set or freeze changed")
    sample = read(packet / "sample.json")
    key = read(packet / "key.json")
    require(sample["format"] == "unswell-349-real-source-admission-transfer-sample-v1"
            and len(sample["items"]) == len(key) == 24
            and len({row["item_id"] for row in sample["items"]}) == 24
            and len({row["group"] for row in sample["items"]}) == 12
            and Counter(row["old_status"] for row in key) ==
            {"accepted": 12, "uncertain": 12}, "sample inventory changed")
    require(digest(packet / "sample.json") == digest(v1 / "packet/sample.json")
            and digest(packet / "key.json") == digest(v1 / "packet/key.json"),
            "v1 and v2 inputs changed")
    old_requests = {path.name: digest(path) for path in
                    (v1 / "packet/requests").glob("*.txt")}
    requests = {path.name: digest(path) for path in
                (packet / "requests").glob("*.txt")}
    require(requests == old_requests and len(requests) == 24,
            "v1 and v2 request bytes changed")
    replay = module(REPLAY, "reference_replay_for_transfer_verifier")
    sources = replay.sources()
    require(len(sources) == 36, "reference source count changed")
    by_sha = {sha(raw): (path, raw) for path, raw in sources.items()}
    require(len(by_sha) == 36, "reference source hash collision")
    manifest = {digest_value: path for digest_value, (path, _) in sorted(by_sha.items())}
    manifest_bytes = (json.dumps(manifest, ensure_ascii=False, sort_keys=True,
                                 indent=2) + "\n").encode("utf-8")
    require(protocol["original_source_manifest_sha256"] == sha(manifest_bytes),
            "reference source manifest changed")
    base_runner = module(BASE, "synthetic_admission_for_transfer_verifier")
    journal = read(folder / "run.json")
    require(journal["status"] == "complete" and journal["complete"] is True
            and journal["freeze_sha256"] == digest(packet / "freeze.json")
            and journal["model"] == "gpt-6-astra"
            and journal["effort"] == "medium"
            and journal["unattempted"] == []
            and len(journal["calls"]) == 24, "run incomplete or changed")
    require([row["item_id"] for row in journal["calls"]] ==
            [row["item_id"] for row in sample["items"]], "call order changed")
    require({path.name for path in folder.iterdir()} ==
            {row["item_id"] for row in sample["items"]} | {"run.json"},
            "run file inventory changed")
    require(all(row["item_id"] == old["item_id"] and
                row["candidate_id"] == old["candidate_id"]
                for row, old in zip(sample["items"], key)),
            "sample and answer key identity changed")
    decisions = []
    responses = {}
    usage = {name: 0 for name in ("input_tokens", "cached_input_tokens",
                                  "output_tokens", "reasoning_output_tokens")}
    for item, old, call in zip(sample["items"], key, journal["calls"]):
        path, raw = by_sha[item["source_sha256"]]
        start, end = item["target_start_byte"], item["target_end_byte"]
        require(0 <= start < end <= len(raw)
                and raw[start:end] == item["target_quote"].encode("utf-8"),
                "target source binding changed")
        payload = {"item_id": item["item_id"],
                   "source_sha256": item["source_sha256"],
                   "source": raw.decode("utf-8"),
                   "target_quote": item["target_quote"],
                   "target_start_byte": start, "target_end_byte": end,
                   "proposed_diagnosis": item["proposed_diagnosis"]}
        request = packet / "requests" / (item["item_id"] + ".txt")
        require(request.read_bytes() == base_runner.request(payload),
                "request text or grade-blinding changed")
        target = folder / item["item_id"]
        require({p.name for p in target.iterdir()} ==
                {"events.jsonl", "stderr.txt", "response.json", "bound.json"},
                "raw result file set changed")
        require((target / "stderr.txt").stat().st_size == 0,
                "model stderr is nonempty")
        messages, turns = [], []
        for line in (target / "events.jsonl").read_text(encoding="utf-8").splitlines():
            event = json.loads(line)
            kind = event.get("type")
            require(kind in {"thread.started", "turn.started", "turn.completed",
                             "item.started", "item.updated", "item.completed"},
                    "unexpected trace event")
            if kind.startswith("item."):
                row = event.get("item", {})
                require(row.get("type") in {"agent_message", "reasoning"},
                        "tool or error event in trace")
                if kind == "item.completed" and row["type"] == "agent_message":
                    messages.append(row.get("text", ""))
            if kind == "turn.completed":
                turns.append(event)
        response = read(target / "response.json")
        require(len(turns) == 1 and messages and json.loads(messages[-1]) == response
                and read(target / "bound.json") == response,
                "response or trace changed")
        require(set(response) == {"item_id", "source_sha256", "target_start_byte",
                                  "target_end_byte", "decision", "replacement", "basis"},
                "response schema changed")
        for field in ("item_id", "source_sha256", "target_start_byte", "target_end_byte"):
            require(type(response[field]) is type(payload[field])
                    and response[field] == payload[field],
                    "response source binding changed")
        require(response["decision"] in {"admit", "reject", "abstain"}
                and isinstance(response["replacement"], str)
                and isinstance(response["basis"], str)
                and response["basis"].strip(), "response decision invalid")
        require(response["decision"] == "admit" or not response["replacement"],
                "nonadmission contains edit")
        require(response["decision"] != "admit" or
                response["replacement"] != item["target_quote"],
                "admitted edit is unchanged")
        require(call["item_id"] == item["item_id"]
                and call["status"] == "valid" and call["exit_code"] == 0
                and call["request_sha256"] == digest(request)
                and call["bound_sha256"] == digest(target / "bound.json")
                and call["decision"] == response["decision"]
                and 0 < call["seconds"] <= 600,
                "run journal call changed")
        for field in usage:
            value = turns[0].get("usage", {}).get(field)
            require(isinstance(value, int) and value >= 0,
                    "model token usage missing")
            usage[field] += value
        responses[item["item_id"]] = response
        correct = (response["decision"] == "admit" if old["old_status"] == "accepted"
                   else response["decision"] in ("reject", "abstain"))
        decisions.append({"item_id": item["item_id"],
                          "candidate_id": old["candidate_id"], "path": path,
                          "old_status": old["old_status"],
                          "decision": response["decision"],
                          "agrees_with_archived_grade": correct,
                          "source_sha256": item["source_sha256"],
                          "target_start_byte": start, "target_end_byte": end,
                          "basis": response["basis"],
                          "replacement": response["replacement"]})
    review = read(base / "review.json")
    admitted = {item_id: response for item_id, response in responses.items()
                if response["decision"] == "admit"}
    require(review["format"] == "unswell-349-real-source-transfer-edit-review-v2"
            and review["reviewer"] == "Codex assistant under ADR 0041"
            and review["reviewer_role"] == "assistant, not a human annotator"
            and len(review["edits"]) == len(admitted),
            "edit review inventory changed")
    reviewed = {}
    for row in review["edits"]:
        require(set(row) == {"item_id", "source_sha256", "replacement_sha256",
                             "safety", "rationale"}
                and row["item_id"] in admitted
                and row["item_id"] not in reviewed
                and row["source_sha256"] == admitted[row["item_id"]]["source_sha256"]
                and row["replacement_sha256"] == sha(
                    admitted[row["item_id"]]["replacement"].encode("utf-8"))
                and row["safety"] in {"safe", "unsafe", "unresolved"}
                and isinstance(row["rationale"], str) and row["rationale"].strip(),
                "edit review binding or decision changed")
        reviewed[row["item_id"]] = row
    require(set(reviewed) == set(admitted), "admitted edit was not reviewed")
    for row in decisions:
        row["edit_safety"] = reviewed[row["item_id"]]["safety"] if row["item_id"] in reviewed else None
    accepted = sum(row["old_status"] == "accepted" and row["decision"] == "admit"
                   for row in decisions)
    withheld = sum(row["old_status"] == "uncertain" and
                   row["decision"] in ("reject", "abstain") for row in decisions)
    agreement = accepted + withheld
    safe = all(row["safety"] == "safe" for row in reviewed.values())
    advance = agreement >= 21 and accepted >= 10 and withheld >= 10 and safe
    result = read(base / "result.json")
    require(result["format"] == "unswell-349-real-source-transfer-result-v2"
            and result["freeze_sha256"] == digest(packet / "freeze.json")
            and result["run_sha256"] == digest(folder / "run.json")
            and result["review_sha256"] == digest(base / "review.json")
            and result["total"] == result["calls"] == 24
            and result["complete"] is True
            and result["accepted_total"] == result["uncertain_total"] == 12
            and result["accepted_retained"] == accepted
            and result["uncertain_withheld"] == withheld
            and result["agreement"] == agreement
            and result["admitted"] == len(admitted)
            and result["abstained"] == sum(row["decision"] == "abstain" for row in decisions)
            and result["safety_clear"] == safe
            and result["advance"] == advance
            and result["decisions"] == decisions,
            "scored result changed")
    first = read(v1 / "run/run.json")
    require(first["status"] == "stopped" and first["complete"] is False
            and len(first["calls"]) == 1
            and first["calls"][0]["status"] == "invalid"
            and first["calls"][0]["exit_code"] == 1
            and first["stop_reason"] == "first_failed_request_no_retry"
            and len(first["unattempted"]) == 23
            and (v1 / "run" / first["calls"][0]["item_id"] / "events.jsonl").stat().st_size == 0
            and "failed to initialize in-process app-server client" in
            (v1 / "run" / first["calls"][0]["item_id"] / "stderr.txt").read_text(),
            "v1 terminal stop changed")
    first_result = read(v1 / "result.json")
    require(first_result["format"] == "unswell-349-real-source-transfer-result-v1"
            and first_result["freeze_sha256"] == digest(v1 / "packet/freeze.json")
            and first_result["run_sha256"] == digest(v1 / "run/run.json")
            and first_result["review_sha256"] == digest(v1 / "review.json")
            and first_result["total"] == 24 and first_result["calls"] == 1
            and first_result["complete"] is False
            and first_result["agreement"] == 0
            and first_result["admitted"] == 0
            and first_result["advance"] is False
            and read(v1 / "review.json")["edits"] == [],
            "v1 terminal result changed")
    return {"format": "unswell-349-real-source-transfer-verification-v1",
            "v1_freeze_sha256": digest(v1 / "packet/freeze.json"),
            "v2_freeze_sha256": digest(packet / "freeze.json"),
            "run_sha256": digest(folder / "run.json"),
            "review_sha256": digest(base / "review.json"),
            "result_sha256": digest(base / "result.json"),
            "calls": 24, "agreement": agreement,
            "accepted_retained": accepted, "uncertain_withheld": withheld,
            "admitted": len(admitted), "safety": dict(Counter(
                row["safety"] for row in reviewed.values())),
            "advance": advance, "usage": usage}


if __name__ == "__main__":
    base = Path(sys.argv[1]) if len(sys.argv) > 1 else ROOT
    v1 = Path(sys.argv[2]) if len(sys.argv) > 2 else V1
    print(json.dumps(verify(base, v1), sort_keys=True))
