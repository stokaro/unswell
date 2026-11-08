"""Prepare and execute a frozen, tool-free complementary semantic pilot."""

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
CONTROL = ROOT.parent / "2026-11-08-source-function-pilot"
PACKET = ROOT / "packet"
RUN = ROOT / "run"
PAGES = ("caddy", "restic", "gitea", "otelcol")
MODEL = "gpt-6-astra"
EFFORT = "medium"
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

PROMPT = """Review the COMPLETE public technical README below for independently useful
semantic and rhetorical corrections. Do not infer AI authorship. Use only this
source; no tools. This is a second, complementary pass: spend the finding budget
on defects that a grammar pass would miss. Do not report grammar, spelling,
punctuation, or purely optional word substitutions.

Check these distinct editorial functions across the entire page:
1. Claims of superiority, universality, user behavior, or performance whose
   scope or intensity outruns the evidence or conditions stated on the page.
2. Repeated or circular claims that add no property beyond a nearby heading,
   label, or sentence, while retaining any concrete qualification.
3. Rhetorical questions, introductions, and navigation leads that delay the
   reader's task without adding a reason, condition, or necessary context.
4. A statement whose implied support conflicts with another part of the page.

For each candidate, ask whether the source gives a concrete fact, measurable
property, necessary condition, or useful authorial emphasis that would be lost
by a minimal repair. Preserve it when so. An absent external citation alone is
not a defect. Preserve attributed claims, qualified aspirations, examples,
technical definitions, useful emphasis, literal identifiers, URLs, versions,
numbers, negation, quotations, code, and source syntax. Abstain if a safe,
source-specific correction cannot be justified from the page.

For every finding quote exact contiguous source text, give its original 1-based
line, a precise diagnosis, source evidence, source function, and what a minimal
repair must retain. Give zero findings if none are independently useful.
Maximum 20 findings. Return diagnosis only: no replacement text or editing.
Return only schema JSON. Echo page ID and source SHA-256 exactly.
"""

SCHEMA = {"type": "object", "properties": {
    "page_id": {"type": "string"}, "source_sha256": {"type": "string"},
    "findings": {"type": "array", "items": {"type": "object", "properties": {
        "line": {"type": "integer"}, "quote": {"type": "string"},
        "diagnosis": {"type": "string"}, "evidence": {"type": "string"},
        "source_function": {"type": "string"}, "deletion_test": {"type": "string"}},
        "required": ["line", "quote", "diagnosis", "evidence", "source_function",
                     "deletion_test"], "additionalProperties": False}}},
    "required": ["page_id", "source_sha256", "findings"],
    "additionalProperties": False}


def encoded(value):
    return (json.dumps(value, ensure_ascii=False, sort_keys=True, indent=2) + "\n").encode("utf-8")


def sha(data):
    return hashlib.sha256(data).hexdigest()


def file_sha(path):
    return sha(path.read_bytes())


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def payload(page):
    raw = source_path(page).read_bytes()
    source = raw.decode("utf-8")
    numbered = "\n".join(f"{index:04d}| {line}" for index, line in
                         enumerate(source.splitlines(), 1))
    return {"page_id": page, "source_sha256": sha(raw), "numbered_source": numbered}


def source_path(page):
    return CONTROL / "sources" / page / "README.md"


def request(page):
    return (PROMPT + "\nINPUT JSON:\n" + encoded(payload(page)).decode("utf-8")).encode("utf-8")


def files():
    paths = [ROOT / "PROTOCOL.md", ROOT / "labels.json", ROOT / "run.py",
             ROOT / "score.py", ROOT / "test_packet.py"]
    paths += [CONTROL / "sources" / page / "README.md" for page in PAGES]
    paths += [CONTROL / "sources.json", CONTROL / "labels.json",
              CONTROL / "result.json", CONTROL / "review.json"]
    paths += [CONTROL / "run" / page / "bound.json" for page in PAGES]
    paths += sorted((PACKET / "requests").glob("*.txt"))
    paths += [PACKET / "schema.json", PACKET / "protocol.json"]
    return {str(path.relative_to(ROOT.parent)): file_sha(path) for path in paths}


def prepare():
    if PACKET.exists():
        raise ValueError("packet already exists")
    labels = read(ROOT / "labels.json")
    if (ROOT / "labels.json").read_bytes() != (CONTROL / "labels.json").read_bytes():
        raise ValueError("control labels changed")
    if len(labels["events"]) != 8 or {e["page"] for e in labels["events"]} != set(PAGES):
        raise ValueError("source-only labels changed")
    (PACKET / "requests").mkdir(parents=True)
    for page in PAGES:
        raw = source_path(page).read_bytes()
        for event in [e for e in labels["events"] if e["page"] == page]:
            assert event["source_sha256"] == sha(raw)
            assert raw[event["start_byte"]:event["end_byte"]] == event["quote"].encode("utf-8")
        (PACKET / "requests" / f"{page}.txt").write_bytes(request(page))
    (PACKET / "schema.json").write_bytes(encoded(SCHEMA))
    (PACKET / "protocol.json").write_bytes(encoded({
        "format": "unswell-349-complementary-semantic-pilot-v1",
        "model": MODEL, "effort": EFFORT,
        "route": "existing Codex subscription; no paid API route",
        "calls_max": 4, "concurrency": 1, "attempts_per_request": 1,
        "automatic_retries": False, "tools_allowed": False,
        "per_call_timeout_seconds": PER_CALL_SECONDS,
        "total_timeout_seconds": TOTAL_SECONDS,
        "failure_policy": "Stop at first CLI, schema, trace, source-binding, or resource error; never retry or repair a response.",
        "input_policy": "One complete numbered public source per request; no labels, source-only rationale, prior answer, or grade enters the request.",
        "screen": "Combined prior control and new semantic pass: at least 7/8 source-only events, at least 85% accepted actual delivered findings counting all raw findings including duplicates, no invalid page, no protected counterexample flagged; directional only.",
        "pages": list(PAGES)}))
    (PACKET / "freeze.json").write_bytes(encoded(files()))
    return file_sha(PACKET / "freeze.json")


def verify():
    if read(PACKET / "freeze.json") != files():
        raise ValueError("packet hash mismatch")
    for page in PAGES:
        if (PACKET / "requests" / f"{page}.txt").read_bytes() != request(page):
            raise ValueError("request bytes changed")
    if read(PACKET / "schema.json") != SCHEMA:
        raise ValueError("schema changed")
    protocol = read(PACKET / "protocol.json")
    if protocol["format"] != "unswell-349-complementary-semantic-pilot-v1" or \
            protocol["model"] != MODEL or protocol["effort"] != EFFORT or \
            protocol["calls_max"] != 4 or protocol["pages"] != list(PAGES):
        raise ValueError("protocol changed")
    return protocol


def bind(page, response):
    inp = payload(page)
    if set(response) != {"page_id", "source_sha256", "findings"} or \
            response["page_id"] != page or \
            response["source_sha256"] != inp["source_sha256"]:
        raise ValueError("page identity changed")
    findings = response["findings"]
    if not isinstance(findings, list) or len(findings) > 20:
        raise ValueError("findings limit or type")
    raw = source_path(page).read_bytes()
    bound = []
    for row in findings:
        if not isinstance(row, dict) or set(row) != set(SCHEMA["properties"]["findings"]["items"]["required"]):
            raise ValueError("finding fields changed")
        line, quote = row["line"], row["quote"]
        if type(line) is not int or line < 1 or not isinstance(quote, str) or not quote.strip():
            raise ValueError("invalid line or quote")
        if any(not isinstance(row[k], str) or not row[k].strip() for k in
               ("diagnosis", "evidence", "source_function", "deletion_test")):
            raise ValueError("missing diagnosis evidence")
        q = quote.encode("utf-8")
        matches = []
        pos = 0
        while True:
            pos = raw.find(q, pos)
            if pos < 0:
                break
            if raw[:pos].count(b"\n") + 1 == line:
                matches.append(pos)
            pos += 1
        if len(matches) != 1:
            raise ValueError("quote not uniquely bound to original line")
        start = matches[0]
        bound.append(dict(row, start_byte=start, end_byte=start + len(q)))
    return {"page_id": page, "source_sha256": inp["source_sha256"], "findings": bound}


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
        raise ValueError("incomplete or mismatched trace")


def execute():
    verify()
    if RUN.exists():
        raise ValueError("run already exists")
    RUN.mkdir()
    journal = {"status": "running", "complete": False, "calls": [],
               "freeze_sha256": file_sha(PACKET / "freeze.json"),
               "model": MODEL, "effort": EFFORT, "monetary_cost": None}
    (RUN / "run.json").write_bytes(encoded(journal))
    started = time.monotonic()
    with tempfile.TemporaryDirectory(prefix="unswell-349-source-function-empty-") as empty:
        for page in PAGES:
            remaining = TOTAL_SECONDS - (time.monotonic() - started)
            if remaining <= 0:
                journal["stop_reason"] = "total_timeout"
                break
            target = RUN / page
            target.mkdir()
            args = ["codex", "exec", *OPTIONS, "--model", MODEL, "--cd", empty,
                    "--output-schema", str(PACKET / "schema.json"),
                    "--output-last-message", str(target / "response.json"),
                    "-c", 'model_reasoning_effort="medium"']
            for setting in SETTINGS:
                args.extend(["-c", setting])
            args.append("-")
            row = {"page": page, "status": "started",
                   "request_sha256": file_sha(PACKET / "requests" / f"{page}.txt")}
            journal["calls"].append(row)
            (RUN / "run.json").write_bytes(encoded(journal))
            before = time.monotonic()
            try:
                process = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                           stderr=subprocess.PIPE, start_new_session=True)
                try:
                    stdout, stderr = process.communicate((PACKET / "requests" / f"{page}.txt").read_bytes(),
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
                if len(stdout) > 8 << 20 or len(stderr) > 1 << 20 or \
                        (target / "response.json").stat().st_size > 1 << 20:
                    raise ValueError("response resource limit")
                response = read(target / "response.json")
                trace(target, response)
                result = bind(page, response)
                (target / "bound.json").write_bytes(encoded(result))
                row.update(status="valid", findings=len(result["findings"]),
                           bound_sha256=file_sha(target / "bound.json"))
            except (ValueError, KeyError, TypeError, OSError) as error:
                row.update(status="invalid", error=str(error))
                journal["stop_reason"] = "first_failed_request_no_retry"
            row["seconds"] = round(time.monotonic() - before, 3)
            (RUN / "run.json").write_bytes(encoded(journal))
            print(json.dumps(row, sort_keys=True), flush=True)
            if row["status"] != "valid":
                break
    journal["complete"] = len(journal["calls"]) == 4 and all(
        row["status"] == "valid" for row in journal["calls"])
    journal["status"] = "complete" if journal["complete"] else "stopped"
    journal["unattempted"] = [p for p in PAGES[len(journal["calls"]):]]
    (RUN / "run.json").write_bytes(encoded(journal))
    return {"complete": journal["complete"], "calls": len(journal["calls"]),
            "run_sha256": file_sha(RUN / "run.json")}


if __name__ == "__main__":
    command = sys.argv[1:] or ["verify"]
    result = prepare() if command == ["prepare"] else execute() if command == ["execute"] else verify()
    print(json.dumps(result, sort_keys=True))
