"""Independently replay the frozen paired rubric pilot and negative result."""

from collections import Counter
import hashlib
import importlib.util
import json
from pathlib import Path
import sys


ROOT = Path(__file__).resolve().parent
BASE = ROOT.parent / "2026-11-08-evidence-flip-pilot" / "run.py"
REPLAY = ROOT.parent / "2026-09-26-outcome-announcements" / "replay.py"
OLD_SAMPLE = ROOT.parent / "2026-11-08-real-source-admission-transfer-v2" / "packet/sample.json"
SOURCE_RUNNER = ROOT.parent / "2026-11-08-real-source-admission-transfer-v2" / "run.py"
EXPECTED_FREEZE_SHA256 = "4aff664bf6879b3b8ac830007efc30296ea26fd6c00f55f9f70453119297ec34"


def require(condition, reason):
    if not condition:
        raise ValueError(reason)


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def digest(path):
    return sha(path.read_bytes())


def encoded(value):
    return (json.dumps(value, ensure_ascii=False, sort_keys=True, indent=2) + "\n").encode("utf-8")


def module(path, name):
    spec = importlib.util.spec_from_file_location(name, path)
    result = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(result)
    return result


def verify(base=ROOT):
    packet = base / "packet"
    folder = base / "run"
    protocol = read(packet / "protocol.json")
    require(protocol["format"] == "unswell-349-rubric-transfer-protocol-v1"
            and protocol["model"] == "gpt-6-astra"
            and protocol["effort"] == "medium"
            and protocol["calls_max"] == 24
            and protocol["concurrency"] == protocol["attempts_per_request"] == 1
            and protocol["automatic_retries"] is False
            and protocol["tools_allowed"] is False
            and protocol["per_call_timeout_seconds"] == 600
            and protocol["total_timeout_seconds"] == 10800,
            "execution contract changed")
    for field, path in (("base_runner_sha256", BASE),
                        ("source_runner_sha256", SOURCE_RUNNER),
                        ("old_sample_sha256", OLD_SAMPLE),
                        ("sample_script_sha256", base / "sample.py"),
                        ("runner_sha256", base / "run.py"),
                        ("scorer_sha256", base / "score.py"),
                        ("protocol_md_sha256", base / "PROTOCOL.md"),
                        ("sample_sha256", packet / "sample.json"),
                        ("key_sha256", packet / "key.json"),
                        ("examples_sha256", packet / "examples.json")):
        require(protocol[field] == digest(path), "protocol hash changed: " + field)
    freeze = read(packet / "freeze.json")
    require(digest(packet / "freeze.json") == EXPECTED_FREEZE_SHA256,
            "prefrozen packet digest changed")
    files = {str(path.relative_to(packet)): digest(path)
             for path in packet.rglob("*") if path.is_file() and path.name != "freeze.json"}
    require(files == freeze and len(files) == 29,
            "frozen packet file set changed")
    sample = read(packet / "sample.json")
    key = read(packet / "key.json")
    examples = read(packet / "examples.json")
    require(sample["format"] == "unswell-349-rubric-transfer-sample-v1"
            and sample["source_pages"] == 36
            and len(sample["items"]) == len(key) == 12
            and len({item["case_id"] for item in sample["items"]}) == 12
            and len({item["page"] for item in sample["items"]}) == 12
            and Counter(row["old_status"] for row in key) ==
            {"accepted": 6, "uncertain": 6},
            "balanced page-disjoint sample changed")
    old_pages = {item["page"] for item in read(OLD_SAMPLE)["items"]}
    require(all(item["group"] == "ptah" and item["page"] not in old_pages
                for item in sample["items"]), "Ptah test pages overlap earlier pilot")
    require(len(examples) == 22 and
            Counter(row["prior_grade"] for row in examples) ==
            {"accepted": 11, "uncertain": 11} and
            len({row["source_group"] for row in examples}) == 11 and
            all(row["source_group"] != "ptah" for row in examples),
            "non-Ptah rubric examples changed")
    replay = module(REPLAY, "rubric_verifier_original_sources")
    sources = replay.sources()
    require(len(sources) == 36, "original source inventory changed")
    by_sha = {sha(raw): (path, raw) for path, raw in sources.items()}
    require(len(by_sha) == 36, "original source hash collision")
    manifest = {hash_value: path for hash_value, (path, _) in sorted(by_sha.items())}
    require(protocol["original_source_manifest_sha256"] == sha(encoded(manifest)),
            "original source manifest changed")
    base_runner = module(BASE, "rubric_verifier_base_prompt")
    require(read(packet / "schema.json") == base_runner.SCHEMA,
            "response schema changed")
    journal = read(folder / "run.json")
    require(journal["status"] == "complete" and journal["complete"] is True
            and journal["freeze_sha256"] == EXPECTED_FREEZE_SHA256
            and journal["model"] == "gpt-6-astra"
            and journal["effort"] == "medium"
            and journal["unattempted"] == []
            and len(journal["calls"]) == 24,
            "run incomplete or changed")
    pairs = []
    for item, old in zip(sample["items"], key):
        require(item["case_id"] == old["case_id"] and
                item["candidate_id"] == old["candidate_id"],
                "case and answer-key identity changed")
        for arm in ("control", "rubric"):
            item_id = sha((item["case_id"] + "\n" + arm).encode("utf-8"))[:16]
            pairs.append((item, old, arm, item_id))
    require(len({item_id for _, _, _, item_id in pairs}) == 24 and
            [call["item_id"] for call in journal["calls"]] ==
            [item_id for _, _, _, item_id in pairs],
            "paired call order changed")
    require({path.name for path in folder.iterdir()} ==
            {item_id for _, _, _, item_id in pairs} | {"run.json"},
            "run result file set changed")
    responses, decisions = {}, []
    usage = {name: 0 for name in ("input_tokens", "cached_input_tokens",
                                  "output_tokens", "reasoning_output_tokens")}
    for (item, old, arm, item_id), call in zip(pairs, journal["calls"]):
        path, raw = by_sha[item["source_sha256"]]
        start, end = item["target_start_byte"], item["target_end_byte"]
        require(0 <= start < end <= len(raw) and 20 <= end - start <= 300 and
                raw[start:end] == item["target_quote"].encode("utf-8"),
                "original UTF-8 target binding changed")
        payload = {"item_id": item_id, "source_sha256": item["source_sha256"],
                   "source": raw.decode("utf-8"), "target_quote": item["target_quote"],
                   "target_start_byte": start, "target_end_byte": end,
                   "proposed_diagnosis": item["proposed_diagnosis"]}
        request = packet / "requests" / (item_id + ".txt")
        raw_request = request.read_bytes()
        marker = b"\nINPUT JSON:\n" if arm == "control" else b"\nCURRENT INPUT JSON:\n"
        require(raw_request.count(marker) == 1 and
                json.loads(raw_request.split(marker, 1)[1]) == payload,
                "current request source or target changed")
        require(old["candidate_id"].encode() not in raw_request and
                old["old_rationale"].encode() not in raw_request,
                "current hidden grade entered request")
        if arm == "control":
            require(raw_request == base_runner.request(payload),
                    "control prompt changed")
        else:
            require(raw_request.startswith(base_runner.PROMPT.encode("utf-8")) and
                    encoded(examples) in raw_request and
                    b"PRIOR EDITORIAL EXAMPLES JSON:" in raw_request,
                    "rubric examples or prompt changed")
        target = folder / item_id
        require({p.name for p in target.iterdir()} ==
                {"events.jsonl", "stderr.txt", "response.json", "bound.json"}
                and (target / "stderr.txt").stat().st_size == 0,
                "raw call file set or stderr changed")
        messages, turns = [], []
        for line in (target / "events.jsonl").read_text(encoding="utf-8").splitlines():
            event = json.loads(line)
            kind = event.get("type")
            require(kind in {"thread.started", "turn.started", "turn.completed",
                             "item.started", "item.updated", "item.completed"},
                    "unexpected model trace event")
            if kind.startswith("item."):
                entry = event.get("item", {})
                require(entry.get("type") in {"agent_message", "reasoning"},
                        "tool or error item in model trace")
                if kind == "item.completed" and entry["type"] == "agent_message":
                    messages.append(entry.get("text", ""))
            if kind == "turn.completed":
                turns.append(event)
        response = read(target / "response.json")
        require(len(turns) == 1 and messages and json.loads(messages[-1]) == response
                and read(target / "bound.json") == response,
                "trace, response, or bound file changed")
        require(set(response) == set(base_runner.SCHEMA["required"]),
                "response field set changed")
        for field in ("item_id", "source_sha256", "target_start_byte", "target_end_byte"):
            require(type(response[field]) is type(payload[field]) and
                    response[field] == payload[field],
                    "model response source binding changed")
        require(response["decision"] in {"admit", "reject", "abstain"} and
                isinstance(response["replacement"], str) and
                isinstance(response["basis"], str) and response["basis"].strip() and
                (response["decision"] == "admit" or not response["replacement"]) and
                (response["decision"] != "admit" or
                 response["replacement"] != item["target_quote"]),
                "model response decision or replacement invalid")
        require(call["item_id"] == item_id and call["case_id"] == item["case_id"]
                and call["arm"] == arm and call["status"] == "valid"
                and call["exit_code"] == 0
                and call["request_sha256"] == digest(request)
                and call["bound_sha256"] == digest(target / "bound.json")
                and call["decision"] == response["decision"]
                and 0 < call["seconds"] <= 600,
                "call journal entry changed")
        for field in usage:
            value = turns[0].get("usage", {}).get(field)
            require(isinstance(value, int) and value >= 0,
                    "model token usage missing")
            usage[field] += value
        responses[item_id] = response
        agrees = (response["decision"] == "admit" if old["old_status"] == "accepted"
                  else response["decision"] in ("reject", "abstain"))
        decisions.append({"item_id": item_id, "case_id": item["case_id"],
                          "arm": arm, "candidate_id": old["candidate_id"],
                          "path": path, "old_status": old["old_status"],
                          "decision": response["decision"],
                          "agrees_with_archived_grade": agrees,
                          "source_sha256": item["source_sha256"],
                          "target_start_byte": start, "target_end_byte": end,
                          "basis": response["basis"],
                          "replacement": response["replacement"]})
    review = read(base / "review.json")
    admitted = {item_id: response for item_id, response in responses.items()
                if response["decision"] == "admit"}
    require(review["format"] == "unswell-349-rubric-transfer-edit-review-v1"
            and review["reviewer"] == "Codex assistant under ADR 0041"
            and review["reviewer_role"] == "assistant, not a human annotator"
            and len(review["edits"]) == len(admitted),
            "admitted edit review incomplete")
    reviewed = {}
    for row in review["edits"]:
        require(set(row) == {"item_id", "source_sha256", "replacement_sha256",
                             "safety", "rationale"} and
                row["item_id"] in admitted and row["item_id"] not in reviewed and
                row["source_sha256"] == admitted[row["item_id"]]["source_sha256"] and
                row["replacement_sha256"] == sha(
                    admitted[row["item_id"]]["replacement"].encode("utf-8")) and
                row["safety"] in {"safe", "unsafe", "unresolved"} and
                isinstance(row["rationale"], str) and row["rationale"].strip(),
                "edit review identity or safety changed")
        reviewed[row["item_id"]] = row
    require(set(reviewed) == set(admitted), "admitted edit not reviewed")
    for row in decisions:
        row["edit_safety"] = reviewed[row["item_id"]]["safety"] if row["item_id"] in reviewed else None
    arms = {}
    paired = {}
    for arm in ("control", "rubric"):
        subset = [row for row in decisions if row["arm"] == arm]
        edited = [row for row in subset if row["decision"] == "admit"]
        arms[arm] = {"agreement": sum(row["agrees_with_archived_grade"] for row in subset),
                     "accepted_retained": sum(row["old_status"] == "accepted" and
                                              row["decision"] == "admit" for row in subset),
                     "uncertain_withheld": sum(row["old_status"] == "uncertain" and
                                               row["decision"] in ("reject", "abstain")
                                               for row in subset),
                     "admitted": len(edited),
                     "abstained": sum(row["decision"] == "abstain" for row in subset),
                     "edit_safety": {s: sum(row["edit_safety"] == s for row in edited)
                                     for s in ("safe", "unsafe", "unresolved")}}
    for row in decisions:
        paired.setdefault(row["case_id"], {})[row["arm"]] = row["decision"]
    delta = arms["rubric"]["agreement"] - arms["control"]["agreement"]
    safe = (arms["rubric"]["edit_safety"]["unsafe"] == 0 and
            arms["rubric"]["edit_safety"]["unresolved"] == 0)
    advance = (arms["rubric"]["agreement"] >= 10 and delta >= 3 and
               arms["rubric"]["accepted_retained"] >= 5 and
               arms["rubric"]["uncertain_withheld"] >= 5 and safe)
    result = read(base / "result.json")
    require(result["format"] == "unswell-349-rubric-transfer-result-v1" and
            result["freeze_sha256"] == EXPECTED_FREEZE_SHA256 and
            result["run_sha256"] == digest(folder / "run.json") and
            result["review_sha256"] == digest(base / "review.json") and
            result["total_calls"] == result["calls"] == 24 and
            result["complete"] is True and result["cases"] == 12 and
            result["accepted_total"] == result["uncertain_total"] == 6 and
            result["arms"] == arms and result["paired_decisions"] == paired and
            result["agreement_delta"] == delta and result["safety_clear"] == safe and
            result["advance"] == advance and result["decisions"] == decisions,
            "scored result changed")
    return {"format": "unswell-349-rubric-transfer-verification-v1",
            "freeze_sha256": EXPECTED_FREEZE_SHA256,
            "run_sha256": digest(folder / "run.json"),
            "review_sha256": digest(base / "review.json"),
            "result_sha256": digest(base / "result.json"),
            "calls": 24, "arms": arms, "agreement_delta": delta,
            "advance": advance, "usage": usage}


if __name__ == "__main__":
    base = Path(sys.argv[1]) if len(sys.argv) > 1 else ROOT
    report = verify(base)
    print(json.dumps(report, indent=2, sort_keys=True))
