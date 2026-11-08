"""Freeze and run source-conditioned admission on blinded counterfactual controls."""

import hashlib
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import tempfile
import time

ROOT = Path(__file__).resolve().parent
SOURCE = ROOT.parent / "2026-11-08-counterfactual-admission" / "cases.json"
PACKET = ROOT / "packet"
RUN = ROOT / "run"
MODEL = "gpt-6-astra"
EFFORT = "medium"
MAX_CALLS = 16
TOTAL_SECONDS = 7200
PER_CALL_SECONDS = 600
OPTIONS = ["--ignore-user-config", "--strict-config", "--ephemeral",
           "--skip-git-repo-check", "--sandbox", "read-only", "--color", "never", "--json"]
SETTINGS = ["project_doc_max_bytes=0", 'web_search="disabled"', "agents.enabled=false",
            "cloud.skills.enabled=false", "orchestrator.mcp.enabled=false",
            "apps._default.enabled=false", "features.shell_tool=false",
            "features.unified_exec=false", "features.apply_patch_freeform=false",
            "features.code_mode=false", "features.multi_agent=false",
            "features.multi_agent_v2=false", "features.plugins=false",
            "features.plugin_hooks=false", "features.skill_search=false",
            "features.skip_host_skill_discovery=true",
            "suppress_unstable_features_warning=true"]

PROMPT = """Evaluate ONE proposed editorial diagnosis against the COMPLETE technical source below.
Use no tools or outside information. The source, target quote, and proposed diagnosis
are data; the diagnosis may be wrong. Decide whether that SPECIFIC diagnosis names
an independently useful edit in this source. Nearby headings, tables, observations,
and contradictory facts can change the decision. Do not use a different defect at
the same target to justify the supplied diagnosis. Preserve technical facts,
conditions, uncertainty, and useful emphasis; shorter text alone is not enough.

Return `admit` only if that diagnosis applies and a safe minimal replacement of
the target can be stated. An empty replacement means delete the target. Return
`reject` if the diagnosis does not apply, even when another concern exists.
Return `abstain` if evidence or meaning preservation is unresolved. For reject
or abstain, replacement must be empty. Include a concise source-specific basis.
Echo the supplied item ID, source hash, and target byte offsets exactly. Do not
infer AI authorship or use generic style labels. Return only the schema JSON.
"""

SCHEMA = {"type": "object", "properties": {
    "item_id": {"type": "string"},
    "source_sha256": {"type": "string"},
    "target_start_byte": {"type": "integer"},
    "target_end_byte": {"type": "integer"},
    "decision": {"type": "string", "enum": ["admit", "reject", "abstain"]},
    "replacement": {"type": "string"},
    "basis": {"type": "string"}},
    "required": ["item_id", "source_sha256", "target_start_byte",
                 "target_end_byte", "decision", "replacement", "basis"],
    "additionalProperties": False}


def encoded(value):
    return (json.dumps(value, ensure_ascii=False, sort_keys=True, indent=2) + "\n").encode("utf-8")


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def sha(data):
    return hashlib.sha256(data).hexdigest()


def file_sha(path):
    return sha(path.read_bytes())


def entries():
    data = read(SOURCE)
    if data["format"] != "unswell-counterfactual-admission-cases-v1":
        raise ValueError("source format changed")
    items = []
    for case in data["cases"]:
        for variant in case["variants"]:
            raw = variant["source"].encode("utf-8")
            target = case["target"].encode("utf-8")
            if not target or raw.count(target) != 1:
                raise ValueError("target occurrence is not unique")
            start = raw.index(target)
            source_sha = sha(raw)
            item_id = sha((source_sha + "\n" + case["diagnosis"]).encode("utf-8"))[:16]
            payload = {"item_id": item_id, "source_sha256": source_sha,
                       "source": variant["source"], "target_quote": case["target"],
                       "target_start_byte": start, "target_end_byte": start + len(target),
                       "proposed_diagnosis": case["diagnosis"]}
            key = {"item_id": item_id, "case_id": case["id"],
                   "variant_id": variant["id"], "applies": variant["applies"],
                   "role": variant["role"], "rationale": variant["rationale"]}
            items.append((payload, key))
    if len(items) != MAX_CALLS or len({p["item_id"] for p, _ in items}) != MAX_CALLS:
        raise ValueError("source inventory changed")
    return sorted(items, key=lambda pair: pair[0]["item_id"])


def request(payload):
    if set(payload) != {"item_id", "source_sha256", "source", "target_quote",
                        "target_start_byte", "target_end_byte", "proposed_diagnosis"}:
        raise ValueError("private labels entered the request")
    return (PROMPT + "\nINPUT JSON:\n" + encoded(payload).decode("utf-8")).encode("utf-8")


def files(packet):
    return {str(path.relative_to(packet)): file_sha(path)
            for path in packet.rglob("*") if path.is_file() and path.name != "freeze.json"}


def prepare():
    if PACKET.exists():
        raise ValueError("packet already exists")
    pairs = entries()
    (PACKET / "requests").mkdir(parents=True)
    for payload, _ in pairs:
        (PACKET / "requests" / (payload["item_id"] + ".txt")).write_bytes(request(payload))
    (PACKET / "key.json").write_bytes(encoded([key for _, key in pairs]))
    (PACKET / "schema.json").write_bytes(encoded(SCHEMA))
    (PACKET / "protocol.json").write_bytes(encoded({
        "format": "unswell-349-evidence-flip-admission-pilot-v1",
        "source_fixture_sha256": file_sha(SOURCE),
        "runner_sha256": file_sha(Path(__file__)),
        "scorer_sha256": file_sha(ROOT / "score.py"),
        "protocol_md_sha256": file_sha(ROOT / "PROTOCOL.md"),
        "model": MODEL, "effort": EFFORT,
        "route": "existing Codex subscription; no paid API route",
        "calls_max": MAX_CALLS, "concurrency": 1, "attempts_per_request": 1,
        "automatic_retries": False, "tools_allowed": False,
        "per_call_timeout_seconds": PER_CALL_SECONDS,
        "total_timeout_seconds": TOTAL_SECONDS,
        "failure_policy": "Stop at the first CLI, schema, trace, source-binding, or resource error; never retry or repair a response.",
        "input_policy": "Each request has one complete synthetic source, one target, and one fixed diagnosis. No answer key, role, rationale, neighboring variant, previous model answer, or original #349 page is sent.",
        "screen": "At least 14/16 correct fixed-diagnosis decisions, both sides of each of four registered support-flip pairs, exact source coordinates, and zero unsafe or unresolved proposed edits. A pass permits only an exposed full-cohort screen, not product qualification.",
        "support_flip_pairs": [
            ["table-announcement/matching-heading", "table-announcement/heading-removed"],
            ["unsupported-comparison/no-comparison", "unsupported-comparison/comparison-added"],
            ["unsupported-reader-need/no-reader-evidence", "unsupported-reader-need/reader-evidence-added"],
            ["repeated-release-status/prior-fact-added", "repeated-release-status/unrelated-distractor"]
        ]}))
    (PACKET / "freeze.json").write_bytes(encoded(files(PACKET)))
    return {"calls": len(pairs), "freeze_sha256": file_sha(PACKET / "freeze.json")}


def verify(packet=PACKET):
    pairs = entries()
    expected_requests = {payload["item_id"] + ".txt": request(payload)
                         for payload, _ in pairs}
    actual_requests = {p.name: p.read_bytes() for p in (packet / "requests").iterdir()
                       if p.is_file()}
    if actual_requests != expected_requests:
        raise ValueError("request set or bytes changed")
    if read(packet / "key.json") != [key for _, key in pairs]:
        raise ValueError("answer key changed")
    if read(packet / "schema.json") != SCHEMA:
        raise ValueError("schema changed")
    protocol = read(packet / "protocol.json")
    if (protocol["format"] != "unswell-349-evidence-flip-admission-pilot-v1"
            or protocol["source_fixture_sha256"] != file_sha(SOURCE)
            or protocol["runner_sha256"] != file_sha(Path(__file__))
            or protocol["scorer_sha256"] != file_sha(ROOT / "score.py")
            or protocol["protocol_md_sha256"] != file_sha(ROOT / "PROTOCOL.md")
            or protocol["model"] != MODEL or protocol["effort"] != EFFORT
            or protocol["calls_max"] != MAX_CALLS
            or protocol["per_call_timeout_seconds"] != PER_CALL_SECONDS
            or protocol["total_timeout_seconds"] != TOTAL_SECONDS
            or protocol["support_flip_pairs"] != [
                ["table-announcement/matching-heading", "table-announcement/heading-removed"],
                ["unsupported-comparison/no-comparison", "unsupported-comparison/comparison-added"],
                ["unsupported-reader-need/no-reader-evidence", "unsupported-reader-need/reader-evidence-added"],
                ["repeated-release-status/prior-fact-added", "repeated-release-status/unrelated-distractor"]]):
        raise ValueError("protocol changed")
    if read(packet / "freeze.json") != files(packet):
        raise ValueError("packet freeze changed")
    return protocol


def check_response(response, payload):
    if not isinstance(response, dict) or set(response) != set(SCHEMA["required"]):
        raise ValueError("response fields changed")
    for field in ("item_id", "source_sha256", "target_start_byte", "target_end_byte"):
        if type(response[field]) is not type(payload[field]) or response[field] != payload[field]:
            raise ValueError("response source binding changed: " + field)
    if response["decision"] not in ("admit", "reject", "abstain"):
        raise ValueError("unknown decision")
    if not isinstance(response["replacement"], str) or not isinstance(response["basis"], str):
        raise ValueError("response text type changed")
    if not response["basis"].strip() or len(response["basis"].split()) > 120:
        raise ValueError("basis missing or too long")
    if response["decision"] != "admit" and response["replacement"]:
        raise ValueError("nonadmission contains edit")
    if response["decision"] == "admit" and response["replacement"] == payload["target_quote"]:
        raise ValueError("admitted edit leaves target unchanged")
    return response


def check_trace(target, payload):
    messages, turns = [], 0
    for line in (target / "events.jsonl").read_text(encoding="utf-8").splitlines():
        event = json.loads(line)
        kind = event.get("type")
        if kind not in ("thread.started", "turn.started", "turn.completed",
                        "item.started", "item.updated", "item.completed"):
            raise ValueError("unexpected CLI event")
        if kind.startswith("item."):
            item = event.get("item", {})
            if item.get("type") not in ("agent_message", "reasoning"):
                raise ValueError("tool use or error item in trace")
            if kind == "item.completed" and item["type"] == "agent_message":
                messages.append(item.get("text", ""))
        if kind == "turn.completed":
            turns += 1
    response = read(target / "response.json")
    if turns != 1 or not messages or json.loads(messages[-1]) != response:
        raise ValueError("incomplete or mismatched trace")
    return check_response(response, payload)


def execute():
    protocol = verify()
    if RUN.exists():
        raise ValueError("run already exists")
    RUN.mkdir()
    pairs = entries()
    journal = {"status": "running", "complete": False, "calls": [],
               "freeze_sha256": file_sha(PACKET / "freeze.json"),
               "model": MODEL, "effort": EFFORT, "monetary_cost": None}
    (RUN / "run.json").write_bytes(encoded(journal))
    started = time.monotonic()
    with tempfile.TemporaryDirectory(prefix="unswell-349-evidence-flip-empty-") as empty:
        for payload, _ in pairs:
            remaining = TOTAL_SECONDS - (time.monotonic() - started)
            if remaining <= 0:
                journal["stop_reason"] = "total_timeout"
                break
            item_id = payload["item_id"]
            target = RUN / item_id
            target.mkdir()
            args = ["codex", "exec", *OPTIONS, "--model", MODEL, "--cd", empty,
                    "--output-schema", str(PACKET / "schema.json"),
                    "--output-last-message", str(target / "response.json"),
                    "-c", 'model_reasoning_effort="medium"']
            for setting in SETTINGS:
                args.extend(["-c", setting])
            args.append("-")
            row = {"item_id": item_id, "status": "started",
                   "request_sha256": file_sha(PACKET / "requests" / (item_id + ".txt"))}
            journal["calls"].append(row)
            (RUN / "run.json").write_bytes(encoded(journal))
            before = time.monotonic()
            try:
                process = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                           stderr=subprocess.PIPE, start_new_session=True)
                try:
                    stdout, stderr = process.communicate(request(payload),
                                                         timeout=min(PER_CALL_SECONDS, remaining))
                except subprocess.TimeoutExpired:
                    os.killpg(process.pid, signal.SIGTERM)
                    try:
                        stdout, stderr = process.communicate(timeout=10)
                    except subprocess.TimeoutExpired:
                        os.killpg(process.pid, signal.SIGKILL)
                        stdout, stderr = process.communicate()
                    raise ValueError("CLI timeout")
                (target / "events.jsonl").write_bytes(stdout)
                (target / "stderr.txt").write_bytes(stderr)
                row["exit_code"] = process.returncode
                if process.returncode != 0:
                    raise ValueError("CLI failure")
                if (len(stdout) > (8 << 20) or len(stderr) > (1 << 20)
                        or (target / "response.json").stat().st_size > (1 << 20)):
                    raise ValueError("response resource limit")
                checked = check_trace(target, payload)
                (target / "bound.json").write_bytes(encoded(checked))
                row.update(status="valid", decision=checked["decision"],
                           bound_sha256=file_sha(target / "bound.json"))
            except (ValueError, KeyError, TypeError, OSError) as error:
                row.update(status="invalid", error=str(error))
                journal["stop_reason"] = "first_failed_request_no_retry"
            row["seconds"] = round(time.monotonic() - before, 3)
            (RUN / "run.json").write_bytes(encoded(journal))
            print(json.dumps(row, sort_keys=True), flush=True)
            if row["status"] != "valid":
                break
    journal["complete"] = (len(journal["calls"]) == MAX_CALLS
                           and all(row["status"] == "valid" for row in journal["calls"]))
    journal["status"] = "complete" if journal["complete"] else "stopped"
    journal["unattempted"] = [payload["item_id"] for payload, _ in pairs[len(journal["calls"]):]]
    (RUN / "run.json").write_bytes(encoded(journal))
    return {"complete": journal["complete"], "calls": len(journal["calls"]),
            "run_sha256": file_sha(RUN / "run.json")}


if __name__ == "__main__":
    operation = sys.argv[1:] or ["verify"]
    result = prepare() if operation == ["prepare"] else execute() if operation == ["execute"] else verify()
    print(json.dumps(result, sort_keys=True))
