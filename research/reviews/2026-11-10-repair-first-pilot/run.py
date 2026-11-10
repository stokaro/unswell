"""Freeze and run the bounded #349 repair-first research pilot."""

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
PREVIOUS = ROOT.parent / "2026-11-10-additive-admission-pilot"
PACKET = ROOT / "packet"
RUN = ROOT / "run"
MODEL = "gpt-6-astra"
EFFORT = "medium"
PAGES = 4
PER_CALL_SECONDS = 600
TOTAL_SECONDS = 7200
OPTIONS = ["--ignore-user-config", "--strict-config", "--ephemeral",
           "--skip-git-repo-check", "--sandbox", "read-only", "--color",
           "never", "--json"]
SETTINGS = ["project_doc_max_bytes=0", 'web_search="disabled"',
            "agents.enabled=false", "cloud.skills.enabled=false",
            "orchestrator.mcp.enabled=false", "apps._default.enabled=false",
            "features.shell_tool=false", "features.unified_exec=false",
            "features.apply_patch_freeform=false", "features.code_mode=false",
            "features.multi_agent=false", "features.multi_agent_v2=false",
            "features.plugins=false", "features.plugin_hooks=false",
            "features.skill_search=false",
            "features.skip_host_skill_discovery=true",
            "suppress_unstable_features_warning=true"]

PROMPT = """Review the COMPLETE public technical README for specific editorial
revisions. Use only the supplied source and eligible unit inventory. Do not
infer AI authorship or use outside knowledge. Return only schema JSON.

Work from the edit backward: propose a finding only when you can give a
minimal, concrete replacement that fixes a real reader burden in this page.
For each finding, select ONE eligible unit ordinal and quote an exact unique
substring of its raw_slice wholly inside ONE eligible segment. The replacement
must be no longer than the quote. Empty replacement means delete that quote.
State the source-specific diagnosis, the reason the change helps the reader,
and the meaning the replacement retains. Do not propose a generic style
preference or an edit that only makes a sentence shorter.

Inspect the entire page, including introductions, list lead-ins, transitions,
and conclusions. Consider empty framing, wordiness, vague claims, needless
repetition, and unjustified intensifiers, but do not search for trigger words.
Preserve the author's useful emphasis and voice; negation, conditions,
uncertainty, attribution, numbers, units, versions, identifiers, links, code,
table values and scope, and meaningful distinctions. Do not edit a protected
span or quoted example. Up to six edits per page; zero is valid. No tools.
"""

CATEGORIES = ("empty_framing", "wordiness", "vague_claims",
              "needless_repetition", "unjustified_intensifiers")
FIELDS = ("unit_ordinal", "quote", "replacement", "category",
          "diagnosis", "reader_benefit", "retained_meaning")
SCHEMA = {"type": "object", "properties": {
    "page_id": {"type": "string"}, "source_sha256": {"type": "string"},
    "edits": {"type": "array", "items": {"type": "object",
        "properties": {"unit_ordinal": {"type": "integer"},
                       "quote": {"type": "string"},
                       "replacement": {"type": "string"},
                       "category": {"type": "string", "enum": list(CATEGORIES)},
                       "diagnosis": {"type": "string"},
                       "reader_benefit": {"type": "string"},
                       "retained_meaning": {"type": "string"}},
        "required": list(FIELDS), "additionalProperties": False}}},
    "required": ["page_id", "source_sha256", "edits"],
    "additionalProperties": False}


def need(condition, message):
    if not condition:
        raise ValueError(message)


def encode(value):
    return (json.dumps(value, ensure_ascii=False, sort_keys=True, indent=2)
            + "\n").encode("utf-8")


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def sha(data):
    return hashlib.sha256(data).hexdigest()


def file_sha(path):
    return sha(path.read_bytes())


def rows():
    sys.path.insert(0, str(PREVIOUS))
    spec = importlib.util.spec_from_file_location("additive_source_runner",
                                                  PREVIOUS / "run.py")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    result = module.source_rows()
    need(len(result) == PAGES, "source inventory changed")
    return result


def request(row):
    return (PROMPT + "\nINPUT JSON:\n" + encode(row).decode("utf-8")).encode("utf-8")


def policy():
    return {
        "format": "unswell-349-repair-first-pilot-v1",
        "model": MODEL, "effort": EFFORT,
        "route": "existing Codex subscription, no paid API route",
        "authorization_basis": "User granted standing explicit authorization for necessary work and asked not to be repeatedly queried.",
        "calls_max": PAGES, "concurrency": 1, "attempts_per_request": 1,
        "automatic_retries": False,
        "per_call_timeout_seconds": PER_CALL_SECONDS,
        "total_timeout_seconds": TOTAL_SECONDS,
        "failure_policy": "Stop at first CLI, trace, schema, binding, source, resource, or timeout error; never retry or repair.",
        "input_policy": "Four complete prior frozen public pages and exact eligible units; no labels, prior model output, grades, or prospective pages.",
        "screen": "All four pages valid, 3/4 FULL events, 85% accepted and safe delivered edits, zero retain-control hits.",
    }


def frozen_files():
    own = [ROOT / name for name in ("PROTOCOL.md", "run.py", "score.py",
                                   "test_run.py", "test_score.py")]
    prior = [PREVIOUS / name for name in ("PROTOCOL.md", "run.py",
                                         "verify_source.py", "freeze.json",
                                         "sources.json", "inventory.json",
                                         "reference.json")]
    prior += [PREVIOUS / page["local_source"]
              for page in read(PREVIOUS / "sources.json")["pages"]]
    packet = [p for p in PACKET.rglob("*")
              if p.is_file() and p.name != "freeze.json"]
    all_files = own + prior + packet
    return {str(path.relative_to(ROOT.parent.parent.parent)): file_sha(path)
            for path in all_files}


def prepare():
    need(not PACKET.exists(), "packet already exists")
    source = rows()
    (PACKET / "requests").mkdir(parents=True)
    for row in source:
        (PACKET / "requests" / (row["page_id"] + ".txt")).write_bytes(request(row))
    (PACKET / "schema.json").write_bytes(encode(SCHEMA))
    (PACKET / "policy.json").write_bytes(encode(policy()))
    (PACKET / "freeze.json").write_bytes(encode(frozen_files()))
    return {"pages": len(source), "freeze_sha256": file_sha(PACKET / "freeze.json")}


def verify():
    source = rows()
    need(PACKET.is_dir(), "packet missing")
    need(read(PACKET / "schema.json") == SCHEMA, "schema changed")
    need(read(PACKET / "policy.json") == policy(), "policy changed")
    need(read(PACKET / "freeze.json") == frozen_files(), "packet freeze changed")
    expected = {row["page_id"] + ".txt": request(row) for row in source}
    actual = {path.name: path.read_bytes()
              for path in (PACKET / "requests").iterdir() if path.is_file()}
    need(actual == expected, "request set or bytes changed")
    return {"pages": len(source), "freeze_sha256": file_sha(PACKET / "freeze.json")}


def bind(response, row):
    need(type(response) is dict and set(response) ==
         {"page_id", "source_sha256", "edits"}, "response fields")
    need(response["page_id"] == row["page_id"] and
         response["source_sha256"] == row["source_sha256"], "source identity")
    edits = response["edits"]
    need(type(edits) is list and len(edits) <= 6, "edit limit or type")
    units = {unit["ordinal"]: unit for unit in row["eligible_units"]}
    raw_source = row["source"].encode("utf-8")
    bound = []
    for index, item in enumerate(edits):
        need(type(item) is dict and set(item) == set(FIELDS), "edit fields")
        ordinal = item["unit_ordinal"]
        need(type(ordinal) is int and ordinal in units, "unit ordinal")
        need(item["category"] in CATEGORIES, "category")
        for key in FIELDS[1:]:
            need(type(item[key]) is str, "field type: " + key)
        for key in ("quote", "diagnosis", "reader_benefit", "retained_meaning"):
            need(item[key].strip(), "empty field: " + key)
        quote = item["quote"].encode("utf-8")
        replacement = item["replacement"].encode("utf-8")
        need(12 <= len(quote) <= 400 and len(replacement) <= len(quote),
             "quote or replacement length")
        need(all(len(item[key]) <= 700 for key in
                 ("diagnosis", "reader_benefit", "retained_meaning")),
             "explanation length")
        unit = units[ordinal]
        raw_unit = unit["raw_slice"].encode("utf-8")
        need(raw_unit.count(quote) == 1, "quote not unique in selected unit")
        start = unit["raw_span"]["start"] + raw_unit.index(quote)
        end = start + len(quote)
        need(raw_source[start:end] == quote, "original UTF-8 mismatch")
        need(any(segment["start"] <= start and end <= segment["end"]
                 for segment in unit["eligible_segments"]),
             "quote crosses protected span")
        edit_id = f'{row["page_id"]}/edit-{index + 1:02d}'
        bound.append({**item, "edit_id": edit_id,
                      "span": {"start": start, "end": end}})
    return {"page_id": row["page_id"],
            "source_sha256": row["source_sha256"], "edits": bound}


def trace_usage(path, response):
    turns, messages, usage = 0, [], None
    for line in (path / "events.jsonl").read_text(encoding="utf-8").splitlines():
        event = json.loads(line)
        kind = event.get("type")
        need(kind in ("thread.started", "turn.started", "turn.completed",
                      "item.started", "item.updated", "item.completed"),
             "unexpected CLI event")
        if kind.startswith("item."):
            item = event.get("item", {})
            need(item.get("type") in ("agent_message", "reasoning"),
                 "tool use or error in trace")
            if kind == "item.completed" and item["type"] == "agent_message":
                messages.append(item.get("text", ""))
        if kind == "turn.completed":
            turns += 1
            usage = event.get("usage")
    need(turns == 1 and messages and json.loads(messages[-1]) == response,
         "trace mismatch")
    return usage


def execute():
    verify()
    need(not RUN.exists(), "run already exists")
    RUN.mkdir()
    journal = {"status": "running", "complete": False, "calls": [],
               "packet_sha256": file_sha(PACKET / "freeze.json"),
               "model": MODEL, "effort": EFFORT, "monetary_cost": None}
    (RUN / "run.json").write_bytes(encode(journal))
    started = time.monotonic()
    with tempfile.TemporaryDirectory(prefix="unswell-349-repair-empty-") as empty:
        for row in rows():
            remaining = TOTAL_SECONDS - (time.monotonic() - started)
            if remaining <= 0:
                journal["stop_reason"] = "total_timeout"
                break
            page_id = row["page_id"]
            target = RUN / page_id
            target.mkdir()
            req = (PACKET / "requests" / (page_id + ".txt")).read_bytes()
            (target / "request.txt").write_bytes(req)
            args = ["codex", "exec", *OPTIONS, "--model", MODEL, "--cd", empty,
                    "--output-schema", str(PACKET / "schema.json"),
                    "--output-last-message", str(target / "response.json"),
                    "-c", 'model_reasoning_effort="medium"']
            for setting in SETTINGS:
                args.extend(["-c", setting])
            args.append("-")
            call = {"page_id": page_id, "status": "started",
                    "request_sha256": sha(req)}
            journal["calls"].append(call)
            (RUN / "run.json").write_bytes(encode(journal))
            before = time.monotonic()
            try:
                process = subprocess.Popen(args, stdin=subprocess.PIPE,
                                           stdout=subprocess.PIPE,
                                           stderr=subprocess.PIPE,
                                           start_new_session=True)
                try:
                    stdout, stderr = process.communicate(
                        req, timeout=min(PER_CALL_SECONDS, remaining))
                except subprocess.TimeoutExpired:
                    os.killpg(process.pid, signal.SIGKILL)
                    stdout, stderr = process.communicate()
                    (target / "events.jsonl").write_bytes(stdout)
                    (target / "stderr.txt").write_bytes(stderr)
                    raise TimeoutError("model call timed out")
                (target / "events.jsonl").write_bytes(stdout)
                (target / "stderr.txt").write_bytes(stderr)
                need(process.returncode == 0, "Codex CLI failure")
                need(len(stdout) <= 8 << 20 and len(stderr) <= 1 << 20 and
                     (target / "response.json").stat().st_size <= 1 << 20,
                     "response resource limit")
                response = read(target / "response.json")
                call["usage"] = trace_usage(target, response)
                bound = bind(response, row)
                (target / "bound.json").write_bytes(encode(bound))
                call["status"] = "valid"
                call["response_sha256"] = file_sha(target / "response.json")
                call["bound_sha256"] = file_sha(target / "bound.json")
            except (OSError, ValueError, KeyError, TypeError, TimeoutError,
                    json.JSONDecodeError) as error:
                call["status"] = "invalid"
                call["error"] = str(error)
                journal["stop_reason"] = "first_error"
            call["seconds"] = round(time.monotonic() - before, 3)
            (RUN / "run.json").write_bytes(encode(journal))
            print(json.dumps(call, sort_keys=True), flush=True)
            if call["status"] != "valid":
                break
    journal["complete"] = len(journal["calls"]) == PAGES and all(
        call["status"] == "valid" for call in journal["calls"])
    journal["status"] = "complete" if journal["complete"] else "stopped"
    (RUN / "run.json").write_bytes(encode(journal))
    return journal


if __name__ == "__main__":
    command = sys.argv[1] if len(sys.argv) == 2 else ""
    if command == "prepare":
        print(json.dumps(prepare(), sort_keys=True))
    elif command == "verify":
        print(json.dumps(verify(), sort_keys=True))
    elif command == "execute":
        result = execute()
        print(json.dumps({"status": result["status"],
                          "calls": len(result["calls"])}, sort_keys=True))
        raise SystemExit(0 if result["complete"] else 1)
    else:
        raise SystemExit("usage: run.py prepare|verify|execute")
