"""Execute the frozen, tool-free #349 two-pass source inventory pilot."""

import hashlib
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import tempfile
import time

import packet

ROOT = Path(__file__).resolve().parent
PACKET = ROOT / "packet"
RUN = ROOT / "run"
MODEL = "gpt-6-astra"
EFFORT = "medium"
CALLS_MAX = 6
PER_CALL_SECONDS = 600
TOTAL_SECONDS = 10800
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


def files():
    paths = [ROOT / name for name in ("sources.json", "input.json", "inventory.json",
                                     "labels.json", "packet.py", "run.py", "score.py")]
    paths += [ROOT / "sources" / slug / "README.md" for slug, _ in packet.PAGES]
    paths += [PACKET / name for name in ("protocol.json", "inventory-prompt.txt",
                                        "diagnosis-prompt.txt", "inventory-schema.json",
                                        "diagnosis-schema.json")]
    paths += [PACKET / "requests" / f"{slug}-inventory.txt" for slug, _ in packet.PAGES]
    return {str(path.relative_to(ROOT)): packet.sha(path.read_bytes()) for path in paths}


def freeze():
    if RUN.exists() or (PACKET / "freeze.json").exists():
        raise ValueError("run or frozen packet already exists")
    packet.prepare_requests()
    (PACKET / "freeze.json").write_bytes(packet.encoded(files()))
    return packet.sha((PACKET / "freeze.json").read_bytes())


def verify():
    packet.validate_sources_and_labels()
    expected = packet.read(PACKET / "freeze.json")
    if expected != files():
        raise ValueError("frozen packet hash mismatch")
    protocol = packet.read(PACKET / "protocol.json")
    if protocol["format"] != "unswell-349-exhaustive-inventory-pilot-v2" or \
            protocol["model"] != MODEL or protocol["effort"] != EFFORT or \
            protocol["calls_max"] != CALLS_MAX or protocol["per_call_timeout_seconds"] != PER_CALL_SECONDS or \
            protocol["total_timeout_seconds"] != TOTAL_SECONDS or protocol["attempts_per_request"] != 1 or \
            protocol["automatic_retries"] is not False or protocol["tools_allowed"] is not False or \
            protocol["pages"] != [page_id for _, page_id in packet.PAGES]:
        raise ValueError("pilot execution protocol changed")
    for slug, _ in packet.PAGES:
        if (PACKET / "requests" / f"{slug}-inventory.txt").read_bytes() != packet.inventory_request(slug):
            raise ValueError("inventory request bytes changed")
    return protocol


def trace(target, response):
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
    if turns != 1 or not messages or json.loads(messages[-1]) != response:
        raise ValueError("incomplete or mismatched CLI trace")


def call_model(target, request, schema, empty, timeout):
    target.mkdir()
    (target / "request.txt").write_bytes(request)
    args = ["codex", "exec", *OPTIONS, "--model", MODEL, "--cd", empty,
            "--output-schema", str(schema), "--output-last-message", str(target / "response.json"),
            "-c", 'model_reasoning_effort="medium"']
    for setting in SETTINGS:
        args.extend(["-c", setting])
    args.append("-")
    process = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                               stderr=subprocess.PIPE, start_new_session=True)
    try:
        stdout, stderr = process.communicate(request, timeout=timeout)
    except subprocess.TimeoutExpired:
        os.killpg(process.pid, signal.SIGTERM)
        try:
            stdout, stderr = process.communicate(timeout=10)
        except subprocess.TimeoutExpired:
            os.killpg(process.pid, signal.SIGKILL)
            stdout, stderr = process.communicate()
        (target / "events.jsonl").write_bytes(stdout)
        (target / "stderr.txt").write_bytes(stderr)
        raise ValueError("CLI timeout") from None
    (target / "events.jsonl").write_bytes(stdout)
    (target / "stderr.txt").write_bytes(stderr)
    if process.returncode != 0:
        raise ValueError(f"CLI exited {process.returncode}")
    response_path = target / "response.json"
    if len(stdout) > 8 << 20 or len(stderr) > 1 << 20 or \
            not response_path.exists() or response_path.stat().st_size > 1 << 20:
        raise ValueError("response resource limit or missing response")
    response_raw = response_path.read_bytes()
    response = json.loads(response_raw)
    trace(target, response)
    return response_raw, response


def execute():
    verify()
    if RUN.exists():
        raise ValueError("run already exists")
    RUN.mkdir()
    journal = {"status": "running", "complete": False, "calls": [],
               "freeze_sha256": packet.sha((PACKET / "freeze.json").read_bytes()),
               "model": MODEL, "effort": EFFORT, "monetary_cost": None}
    journal_path = RUN / "run.json"
    journal_path.write_bytes(packet.encoded(journal))
    started = time.monotonic()

    def stop(reason):
        journal["status"] = "stopped_at_first_error"
        journal["stop_reason"] = reason
        journal_path.write_bytes(packet.encoded(journal))

    with tempfile.TemporaryDirectory(prefix="unswell-349-inventory-empty-") as empty:
        for slug, _ in packet.PAGES:
            inventory_raw = None
            for stage in ("inventory", "diagnosis"):
                if len(journal["calls"]) >= CALLS_MAX:
                    stop("call budget exhausted")
                    raise ValueError("call budget exhausted")
                remaining = TOTAL_SECONDS - (time.monotonic() - started)
                if remaining <= 0:
                    stop("total timeout")
                    raise ValueError("total timeout")
                if stage == "inventory":
                    request = (PACKET / "requests" / f"{slug}-inventory.txt").read_bytes()
                else:
                    request = packet.diagnosis_request(slug, inventory_raw)
                target = RUN / slug / stage
                target.parent.mkdir(exist_ok=True)
                row = {"page": slug, "stage": stage, "status": "started",
                       "request_sha256": packet.sha(request)}
                journal["calls"].append(row)
                journal_path.write_bytes(packet.encoded(journal))
                try:
                    response_raw, response = call_model(
                        target, request, PACKET / f"{stage}-schema.json", empty,
                        min(PER_CALL_SECONDS, remaining))
                    if stage == "inventory":
                        packet.validate_inventory(slug, response)
                        inventory_raw = response_raw
                    else:
                        bound = packet.validate_diagnosis(slug, response, inventory_raw)
                        (target / "bound.json").write_bytes(packet.encoded(bound))
                except Exception as exc:
                    row["status"] = "failed"
                    row["error"] = str(exc)
                    stop(str(exc))
                    raise
                row["status"] = "valid"
                row["response_sha256"] = packet.sha(response_raw)
                journal_path.write_bytes(packet.encoded(journal))
    journal["status"] = "complete"
    journal["complete"] = True
    journal_path.write_bytes(packet.encoded(journal))
    return len(journal["calls"])


if __name__ == "__main__":
    if sys.argv[1:] == ["freeze"]:
        print(freeze())
    elif sys.argv[1:] == ["verify"]:
        print(verify()["format"])
    elif sys.argv[1:] == ["execute"]:
        print("valid calls", execute())
    else:
        raise SystemExit("usage: run.py freeze|verify|execute")
