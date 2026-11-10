"""Freeze and execute the #349 additive generation and admission pilot."""

import hashlib
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import tempfile
import time

import verify_source

ROOT = Path(__file__).resolve().parent
PACKET = ROOT / "packet"
RUN = ROOT / "run"
MODEL = "gpt-6-astra"
EFFORT = "medium"
MAX_GENERATION = 4
MAX_ADMISSION = 24
TOTAL_SECONDS = 10800
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

GENERATION_PROMPT = """Review one COMPLETE technical README for necessary editorial diagnoses.
Use only its source and eligible unit inventory. Return at most six new, specific
findings in empty_framing, wordiness, vague_claims, or needless_repetition.
A finding must identify a concrete passage whose revision would improve this
source; a phrase sounding promotional or AI-like alone is not sufficient.
Preserve project voice, technical facts, numbers, conditions, useful emphasis,
links, tables, code, and quoted material. Do not infer who wrote the page.

For each finding, select ONE eligible unit ordinal and quote an exact, unique
substring of its raw_slice wholly inside ONE eligible segment. Include a
source-specific diagnosis and explain why the surrounding heading, list,
example, or other evidence does not make the passage useful as written.
Do not suggest replacement text. Do not report grammar-only or factual errors.
Zero findings is valid. Return only schema JSON. No tools or outside sources.
"""

ADMISSION_PROMPT = """Judge ONE proposed editorial diagnosis against the COMPLETE source.
The supplied diagnosis may be wrong. Decide whether that SPECIFIC diagnosis
names an independently useful revision at the bound target, considering nearby
headings, lists, examples, and contradictory evidence. A different defect at
that location cannot justify admission. Preserve project voice, facts, numbers,
conditions, links, code, and useful emphasis. Shorter text alone is not enough.
Choose admit only when the diagnosis applies; choose reject when it does not;
choose abstain when source evidence is unresolved. Do not suggest replacement
text or infer AI authorship. Echo candidate ID, page ID, and source hash exactly.
Return only schema JSON. No tools or outside sources.
"""

GEN_SCHEMA = {"type": "object", "properties": {
    "page_id": {"type": "string"}, "source_sha256": {"type": "string"},
    "findings": {"type": "array", "items": {"type": "object", "properties": {
        "unit_ordinal": {"type": "integer"}, "quote": {"type": "string"},
        "category": {"type": "string", "enum": ["empty_framing", "wordiness",
                                              "vague_claims", "needless_repetition"]},
        "diagnosis": {"type": "string"}, "basis": {"type": "string"}},
        "required": ["unit_ordinal", "quote", "category", "diagnosis", "basis"],
        "additionalProperties": False}}},
    "required": ["page_id", "source_sha256", "findings"], "additionalProperties": False}

ADMIT_SCHEMA = {"type": "object", "properties": {
    "candidate_id": {"type": "string"}, "page_id": {"type": "string"},
    "source_sha256": {"type": "string"},
    "decision": {"type": "string", "enum": ["admit", "reject", "abstain"]},
    "basis": {"type": "string"}},
    "required": ["candidate_id", "page_id", "source_sha256", "decision", "basis"],
    "additionalProperties": False}


def encoded(value):
    return (json.dumps(value, ensure_ascii=False, sort_keys=True, indent=2) + "\n").encode("utf-8")


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def sha(data):
    return hashlib.sha256(data).hexdigest()


def file_sha(path):
    return sha(path.read_bytes())


def need(condition, message):
    if not condition:
        raise ValueError(message)


def source_rows():
    verify_source.verify()
    source = read(ROOT / "sources.json")
    inventory = read(ROOT / "inventory.json")
    by_id = {p["id"]: p for p in inventory["pages"]}
    rows = []
    for item in source["pages"]:
        page_id = item["id"]
        units = []
        for u in by_id[page_id]["units"]:
            units.append({"ordinal": u["ordinal"], "raw_span": u["raw_span"],
                          "raw_slice": u["raw_slice"],
                          "eligible_segments": u["binding"]["segments"]})
        rows.append({"page_id": page_id, "source_sha256": item["sha256"],
                     "source": (ROOT / item["local_source"]).read_text(encoding="utf-8"),
                     "eligible_units": units})
    need(len(rows) == MAX_GENERATION, "page count changed")
    return rows


def generation_request(row):
    return (GENERATION_PROMPT + "\nINPUT JSON:\n" + encoded(row).decode()).encode("utf-8")


def admission_payload(candidate, source_row):
    return {"candidate_id": candidate["candidate_id"], "page_id": candidate["page_id"],
            "source_sha256": candidate["source_sha256"],
            "target_start_byte": candidate["span"]["start"],
            "target_end_byte": candidate["span"]["end"],
            "target_quote": candidate["quote"], "category": candidate["category"],
            "proposed_diagnosis": candidate["diagnosis"], "complete_source": source_row["source"]}


def admission_request(candidate, source_row):
    return (ADMISSION_PROMPT + "\nINPUT JSON:\n" +
            encoded(admission_payload(candidate, source_row)).decode()).encode("utf-8")


def protocol():
    return {"format": "unswell-349-additive-admission-packet-v1",
            "source_freeze_sha256": file_sha(ROOT / "freeze.json"),
            "source_preparation_commit": "fdaef4f", "runner_sha256": file_sha(Path(__file__)),
            "scorer_sha256": file_sha(ROOT / "score.py"),
            "protocol_sha256": file_sha(ROOT / "PROTOCOL.md"),
            "model": MODEL, "effort": EFFORT, "route": "existing Codex subscription",
            "generation_calls_max": MAX_GENERATION, "admission_calls_max": MAX_ADMISSION,
            "concurrency": 1, "attempts_per_request": 1, "automatic_retries": False,
            "per_call_timeout_seconds": PER_CALL_SECONDS,
            "total_timeout_seconds": TOTAL_SECONDS,
            "failure_policy": "Stop at first CLI, trace, schema, binding, source, resource, or timeout error; never retry or repair a response.",
            "dynamic_admission_requests": "Frozen renderer materializes each exact request from a bound generated candidate; save request bytes and SHA-256 before that call.",
            "screen": "Four valid pages; at least 3/4 FULL frozen events; at least 85% accepted admitted diagnoses; zero retain-control hits; zero protected targets; no suggested edits."
            }


def packet_files(packet):
    return {str(p.relative_to(packet)): file_sha(p) for p in packet.rglob("*")
            if p.is_file() and p.name != "freeze.json"}


def prepare():
    need(not PACKET.exists(), "packet already exists")
    rows = source_rows()
    (PACKET / "requests").mkdir(parents=True)
    for row in rows:
        (PACKET / "requests" / (row["page_id"] + ".txt")).write_bytes(
            generation_request(row))
    (PACKET / "generation-schema.json").write_bytes(encoded(GEN_SCHEMA))
    (PACKET / "admission-schema.json").write_bytes(encoded(ADMIT_SCHEMA))
    (PACKET / "protocol.json").write_bytes(encoded(protocol()))
    (PACKET / "freeze.json").write_bytes(encoded(packet_files(PACKET)))
    return {"generation_requests": len(rows), "freeze_sha256": file_sha(PACKET / "freeze.json")}


def verify():
    rows = source_rows()
    need(PACKET.is_dir(), "packet missing")
    need(read(PACKET / "generation-schema.json") == GEN_SCHEMA, "generation schema changed")
    need(read(PACKET / "admission-schema.json") == ADMIT_SCHEMA, "admission schema changed")
    need(read(PACKET / "protocol.json") == protocol(), "protocol changed")
    need(read(PACKET / "freeze.json") == packet_files(PACKET), "packet freeze changed")
    expected = {row["page_id"] + ".txt": generation_request(row) for row in rows}
    actual = {p.name: p.read_bytes() for p in (PACKET / "requests").iterdir() if p.is_file()}
    need(actual == expected, "generation requests changed")
    return {"pages": len(rows), "freeze_sha256": file_sha(PACKET / "freeze.json")}


def bound_generation(response, row):
    need(type(response) is dict and set(response) == {"page_id", "source_sha256", "findings"},
         "generation response fields")
    need(response["page_id"] == row["page_id"] and
         response["source_sha256"] == row["source_sha256"], "generation source binding")
    findings = response["findings"]
    need(type(findings) is list and len(findings) <= 6, "generation finding limit")
    units = {u["ordinal"]: u for u in row["eligible_units"]}
    bound = []
    for index, item in enumerate(findings):
        need(type(item) is dict and set(item) == {"unit_ordinal", "quote", "category",
                                                  "diagnosis", "basis"}, "finding fields")
        need(type(item["unit_ordinal"]) is int and item["unit_ordinal"] in units,
             "unit ordinal")
        need(item["category"] in GEN_SCHEMA["properties"]["findings"]["items"]
             ["properties"]["category"]["enum"], "category")
        for key in ("quote", "diagnosis", "basis"):
            need(type(item[key]) is str and item[key].strip(), "empty finding field: " + key)
        need(len(item["quote"]) <= 400 and len(item["diagnosis"]) <= 700 and
             len(item["basis"]) <= 700, "finding text limit")
        unit = units[item["unit_ordinal"]]
        raw = unit["raw_slice"].encode("utf-8")
        quote = item["quote"].encode("utf-8")
        need(len(quote) >= 12 and raw.count(quote) == 1, "quote not unique in selected unit")
        start = unit["raw_span"]["start"] + raw.index(quote)
        end = start + len(quote)
        need(any(s["start"] <= start and end <= s["end"]
                 for s in unit["eligible_segments"]), "quote crosses protected span")
        need(row["source"].encode("utf-8")[start:end] == quote,
             "quote source mismatch")
        candidate_id = sha((row["page_id"] + "\0" + str(index) + "\0" +
                            item["diagnosis"]).encode("utf-8"))[:16]
        bound.append({**item, "candidate_id": candidate_id,
                      "page_id": row["page_id"], "source_sha256": row["source_sha256"],
                      "span": {"start": start, "end": end}})
    return {"page_id": row["page_id"], "source_sha256": row["source_sha256"],
            "findings": bound}


def bound_admission(response, candidate):
    need(type(response) is dict and set(response) ==
         {"candidate_id", "page_id", "source_sha256", "decision", "basis"},
         "admission response fields")
    for key in ("candidate_id", "page_id", "source_sha256"):
        need(response[key] == candidate[key], "admission binding: " + key)
    need(response["decision"] in ("admit", "reject", "abstain"), "admission decision")
    need(type(response["basis"]) is str and 0 < len(response["basis"].strip()) <= 700,
         "admission basis")
    return response


def checked_trace(path, response):
    messages, turns, usage = [], 0, None
    for line in (path / "events.jsonl").read_text(encoding="utf-8").splitlines():
        event = json.loads(line)
        kind = event.get("type")
        need(kind in ("thread.started", "turn.started", "turn.completed",
                      "item.started", "item.updated", "item.completed"), "unexpected CLI event")
        if kind.startswith("item."):
            item = event.get("item", {})
            need(item.get("type") in ("agent_message", "reasoning"), "tool use or error item")
            if kind == "item.completed" and item["type"] == "agent_message":
                messages.append(item.get("text", ""))
        if kind == "turn.completed":
            turns += 1
            usage = event.get("usage")
    need(turns == 1 and messages and json.loads(messages[-1]) == response,
         "incomplete or mismatched CLI trace")
    return usage


def call(kind, item_id, request, schema, empty, remaining, validate):
    target = RUN / kind / item_id
    target.mkdir(parents=True)
    (target / "request.txt").write_bytes(request)
    args = ["codex", "exec", *OPTIONS, "--model", MODEL, "--cd", empty,
            "--output-schema", str(PACKET / schema),
            "--output-last-message", str(target / "response.json"), "-c",
            'model_reasoning_effort="medium"']
    for setting in SETTINGS:
        args.extend(["-c", setting])
    args.append("-")
    before = time.monotonic()
    row = {"stage": kind, "item_id": item_id, "status": "started",
           "request_sha256": sha(request)}
    try:
        process = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                   stderr=subprocess.PIPE, start_new_session=True)
        try:
            stdout, stderr = process.communicate(request,
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
        need(process.returncode == 0, "CLI failure")
        need(len(stdout) <= 8 << 20 and len(stderr) <= 1 << 20 and
             (target / "response.json").stat().st_size <= 1 << 20,
             "response resource limit")
        response = read(target / "response.json")
        row["usage"] = checked_trace(target, response)
        bound = validate(response)
        (target / "bound.json").write_bytes(encoded(bound))
        row.update(status="valid", bound_sha256=file_sha(target / "bound.json"))
    except (ValueError, KeyError, TypeError, OSError, json.JSONDecodeError) as error:
        row.update(status="invalid", error=str(error))
    row["seconds"] = round(time.monotonic() - before, 3)
    return row


def execute():
    verify()
    need(not RUN.exists(), "run already exists")
    rows = source_rows()
    RUN.mkdir()
    (RUN / "generation").mkdir()
    (RUN / "admission").mkdir()
    journal = {"status": "running", "complete": False, "calls": [],
               "packet_sha256": file_sha(PACKET / "freeze.json"),
               "model": MODEL, "effort": EFFORT, "monetary_cost": None}
    (RUN / "run.json").write_bytes(encoded(journal))
    started = time.monotonic()
    bound_pages = []
    with tempfile.TemporaryDirectory(prefix="unswell-349-additive-empty-") as empty:
        for source in rows:
            remaining = TOTAL_SECONDS - (time.monotonic() - started)
            if remaining <= 0:
                journal["stop_reason"] = "total_timeout"
                break
            page_id = source["page_id"]
            request = generation_request(source)
            result = call("generation", page_id, request, "generation-schema.json", empty,
                          remaining, lambda response: bound_generation(response, source))
            journal["calls"].append(result)
            (RUN / "run.json").write_bytes(encoded(journal))
            print(json.dumps(result, sort_keys=True), flush=True)
            if result["status"] != "valid":
                journal["stop_reason"] = "first_failed_request_no_retry"
                break
            bound_pages.append(read(RUN / "generation" / page_id / "bound.json"))
        if len(bound_pages) == MAX_GENERATION:
            by_page = {row["page_id"]: row for row in rows}
            candidates = [f for page in bound_pages for f in page["findings"]]
            need(len(candidates) <= MAX_ADMISSION, "admission budget exceeded")
            for candidate in candidates:
                remaining = TOTAL_SECONDS - (time.monotonic() - started)
                if remaining <= 0:
                    journal["stop_reason"] = "total_timeout"
                    break
                item_id = candidate["candidate_id"]
                request = admission_request(candidate, by_page[candidate["page_id"]])
                result = call("admission", item_id, request, "admission-schema.json",
                              empty, remaining,
                              lambda response: bound_admission(response, candidate))
                journal["calls"].append(result)
                (RUN / "run.json").write_bytes(encoded(journal))
                print(json.dumps(result, sort_keys=True), flush=True)
                if result["status"] != "valid":
                    journal["stop_reason"] = "first_failed_request_no_retry"
                    break
    expected = MAX_GENERATION + sum(len(page["findings"]) for page in bound_pages)
    journal["complete"] = (len(bound_pages) == MAX_GENERATION
                           and len(journal["calls"]) == expected
                           and all(row["status"] == "valid" for row in journal["calls"]))
    journal["status"] = "complete" if journal["complete"] else "stopped"
    journal["unattempted_generation"] = [row["page_id"] for row in rows[len(bound_pages):]]
    if not journal["complete"]:
        all_candidates = [f["candidate_id"] for page in bound_pages for f in page["findings"]]
        attempted = {r["item_id"] for r in journal["calls"] if r["stage"] == "admission"}
        journal["unattempted_admission"] = [item for item in all_candidates if item not in attempted]
    (RUN / "run.json").write_bytes(encoded(journal))
    return {"complete": journal["complete"], "calls": len(journal["calls"]),
            "run_sha256": file_sha(RUN / "run.json")}


if __name__ == "__main__":
    operation = sys.argv[1:] or ["verify"]
    result = prepare() if operation == ["prepare"] else execute() if operation == ["execute"] else verify()
    print(json.dumps(result, sort_keys=True))
