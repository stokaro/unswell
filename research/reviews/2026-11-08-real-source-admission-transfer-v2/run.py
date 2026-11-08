"""Freeze and run a real-source transfer screen for diagnosis admission."""

import hashlib
import importlib.util
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import tempfile
import time

ROOT = Path(__file__).resolve().parent
BASE_RUNNER = ROOT.parent / "2026-11-08-evidence-flip-pilot" / "run.py"
REPLAY = ROOT.parent / "2026-09-26-outcome-announcements" / "replay.py"
PACKET = ROOT / "packet"
RUN = ROOT / "run"
MAX_CALLS = 24
TOTAL_SECONDS = 10800
PER_CALL_SECONDS = 600


def module(path, name):
    spec = importlib.util.spec_from_file_location(name, path)
    result = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(result)
    return result


def base():
    return module(BASE_RUNNER, "evidence_flip_admission_runner")


def encoded(value):
    return (json.dumps(value, ensure_ascii=False, sort_keys=True, indent=2) + "\n").encode("utf-8")


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def sha(data):
    return hashlib.sha256(data).hexdigest()


def file_sha(path):
    return sha(path.read_bytes())


def original_sources():
    sources = module(REPLAY, "original_reference_replay").sources()
    if len(sources) != 36:
        raise ValueError("original source inventory changed")
    mapped = {}
    for path, raw in sources.items():
        digest = sha(raw)
        if digest in mapped:
            raise ValueError("duplicate original source hash")
        mapped[digest] = (path, raw)
    return mapped


def entries():
    sample = read(PACKET / "sample.json")
    keys = read(PACKET / "key.json")
    if (sample["format"] != "unswell-349-real-source-admission-transfer-sample-v1"
            or len(sample["items"]) != MAX_CALLS or len(keys) != MAX_CALLS):
        raise ValueError("sample inventory changed")
    if sample["source_pages"] != 36:
        raise ValueError("original page count changed")
    if [item["item_id"] for item in sample["items"]] != sorted(
            item["item_id"] for item in sample["items"]):
        raise ValueError("sample order changed")
    sources = original_sources()
    pairs = []
    for item, key in zip(sample["items"], keys):
        if (set(item) != {"item_id", "candidate_id", "category", "group", "page",
                          "proposed_diagnosis", "source_sha256", "target_start_byte",
                          "target_end_byte", "target_quote"}
                or set(key) != {"item_id", "candidate_id", "old_status"}):
            raise ValueError("sample or answer key fields changed")
        if (item["item_id"] != key["item_id"] or
                item["candidate_id"] != key["candidate_id"] or
                key["old_status"] not in ("accepted", "uncertain")):
            raise ValueError("answer key identity changed")
        source_path, raw = sources[item["source_sha256"]]
        if (not 0 <= item["target_start_byte"] < item["target_end_byte"] <= len(raw)
                or raw[item["target_start_byte"]:item["target_end_byte"]] !=
                item["target_quote"].encode("utf-8")
                or not item["proposed_diagnosis"].strip()
                or len(raw) > 65536):
            raise ValueError("source binding or resource limit changed")
        payload = {"item_id": item["item_id"], "source_sha256": item["source_sha256"],
                   "source": raw.decode("utf-8"), "target_quote": item["target_quote"],
                   "target_start_byte": item["target_start_byte"],
                   "target_end_byte": item["target_end_byte"],
                   "proposed_diagnosis": item["proposed_diagnosis"]}
        pairs.append((payload, key, source_path))
    if (len({payload["item_id"] for payload, _, _ in pairs}) != MAX_CALLS
            or len({payload["source_sha256"] for payload, _, _ in pairs}) != 16
            or len({item["group"] for item in sample["items"]}) != 12
            or sum(key["old_status"] == "accepted" for _, key, _ in pairs) != 12):
        raise ValueError("stratified sample changed")
    return pairs


def request(payload):
    return base().request(payload)


def files():
    return {str(path.relative_to(PACKET)): file_sha(path)
            for path in PACKET.rglob("*")
            if path.is_file() and path.name != "freeze.json"}


def prepare():
    if (PACKET / "freeze.json").exists() or (PACKET / "requests").exists():
        raise ValueError("packet already frozen")
    pairs = entries()
    (PACKET / "requests").mkdir()
    for payload, _, _ in pairs:
        (PACKET / "requests" / (payload["item_id"] + ".txt")).write_bytes(request(payload))
    model = base()
    (PACKET / "schema.json").write_bytes(encoded(model.SCHEMA))
    source_manifest = {digest: path for digest, (path, _) in sorted(original_sources().items())}
    (PACKET / "protocol.json").write_bytes(encoded({
        "format": "unswell-349-real-source-admission-transfer-v2",
        "model": model.MODEL, "effort": model.EFFORT,
        "route": "existing Codex subscription; no paid API route",
        "calls_max": MAX_CALLS, "concurrency": 1, "attempts_per_request": 1,
        "automatic_retries": False, "tools_allowed": False,
        "per_call_timeout_seconds": PER_CALL_SECONDS,
        "total_timeout_seconds": TOTAL_SECONDS,
        "base_runner_sha256": file_sha(BASE_RUNNER),
        "runner_sha256": file_sha(Path(__file__)),
        "scorer_sha256": file_sha(ROOT / "score.py"),
        "protocol_md_sha256": file_sha(ROOT / "PROTOCOL.md"),
        "original_source_manifest_sha256": sha(encoded(source_manifest)),
        "sample_sha256": file_sha(PACKET / "sample.json"),
        "answer_key_sha256": file_sha(PACKET / "key.json"),
        "screen": "At least 21/24 agreements, at least 10/12 retained accepted diagnoses, at least 10/12 withheld uncertain diagnoses, 24 exact bindings, and zero unsafe or unresolved admitted edits.",
        "failure_policy": "Stop at the first CLI, schema, trace, source-binding, timeout, or resource error; never retry or repair a response.",
    }))
    (PACKET / "freeze.json").write_bytes(encoded(files()))
    return {"calls": len(pairs), "freeze_sha256": file_sha(PACKET / "freeze.json")}


def verify():
    pairs = entries()
    expected = {payload["item_id"] + ".txt": request(payload)
                for payload, _, _ in pairs}
    actual = {path.name: path.read_bytes() for path in (PACKET / "requests").iterdir()
              if path.is_file()}
    if actual != expected:
        raise ValueError("request set or bytes changed")
    model = base()
    if read(PACKET / "schema.json") != model.SCHEMA:
        raise ValueError("response schema changed")
    protocol = read(PACKET / "protocol.json")
    source_manifest = {digest: path for digest, (path, _) in sorted(original_sources().items())}
    if (protocol["format"] != "unswell-349-real-source-admission-transfer-v2"
            or protocol["model"] != model.MODEL or protocol["effort"] != model.EFFORT
            or protocol["calls_max"] != MAX_CALLS
            or protocol["per_call_timeout_seconds"] != PER_CALL_SECONDS
            or protocol["total_timeout_seconds"] != TOTAL_SECONDS
            or protocol["base_runner_sha256"] != file_sha(BASE_RUNNER)
            or protocol["runner_sha256"] != file_sha(Path(__file__))
            or protocol["scorer_sha256"] != file_sha(ROOT / "score.py")
            or protocol["protocol_md_sha256"] != file_sha(ROOT / "PROTOCOL.md")
            or protocol["original_source_manifest_sha256"] != sha(encoded(source_manifest))
            or protocol["sample_sha256"] != file_sha(PACKET / "sample.json")
            or protocol["answer_key_sha256"] != file_sha(PACKET / "key.json")
            or read(PACKET / "freeze.json") != files()):
        raise ValueError("frozen packet changed")
    return protocol


def execute():
    protocol = verify()
    if RUN.exists():
        raise ValueError("run already exists")
    RUN.mkdir()
    pairs = entries()
    model = base()
    journal = {"status": "running", "complete": False, "calls": [],
               "freeze_sha256": file_sha(PACKET / "freeze.json"),
               "model": protocol["model"], "effort": protocol["effort"],
               "monetary_cost": None}
    (RUN / "run.json").write_bytes(encoded(journal))
    started = time.monotonic()
    with tempfile.TemporaryDirectory(prefix="unswell-349-real-transfer-empty-") as empty:
        for payload, _, _ in pairs:
            remaining = TOTAL_SECONDS - (time.monotonic() - started)
            if remaining <= 0:
                journal["stop_reason"] = "total_timeout"
                break
            item_id = payload["item_id"]
            target = RUN / item_id
            target.mkdir()
            args = ["codex", "exec", *model.OPTIONS, "--model", model.MODEL,
                    "--cd", empty, "--output-schema", str(PACKET / "schema.json"),
                    "--output-last-message", str(target / "response.json"),
                    "-c", 'model_reasoning_effort="medium"']
            for setting in model.SETTINGS:
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
                checked = model.check_trace(target, payload)
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
    journal["unattempted"] = [payload["item_id"] for payload, _, _ in pairs[len(journal["calls"]):]]
    (RUN / "run.json").write_bytes(encoded(journal))
    return {"complete": journal["complete"], "calls": len(journal["calls"]),
            "run_sha256": file_sha(RUN / "run.json")}


if __name__ == "__main__":
    operation = sys.argv[1:] or ["verify"]
    result = prepare() if operation == ["prepare"] else execute() if operation == ["execute"] else verify()
    print(json.dumps(result, sort_keys=True))
