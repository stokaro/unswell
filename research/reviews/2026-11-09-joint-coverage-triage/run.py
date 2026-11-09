"""Freeze and run a bounded, tool-free joint coverage and triage pilot."""

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
REPO = ROOT.parent.parent.parent
PREVIOUS = ROOT.parent / "2026-11-09-fresh-source-function"
PACKET = ROOT / "packet"
RUN = ROOT / "run"
PAGES = (2, 6, 8, 10)
MODEL = "gpt-6-astra"
EFFORT = "medium"
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

PROMPT = """Review the COMPLETE public technical README below for useful,
source-specific editorial corrections. This is an editorial task, not an
AI-authorship judgment. Use only the supplied page and baseline findings; no
tools, browsing, or external knowledge. Return only schema JSON.

Perform two independent jobs over the entire page:

1. For EVERY baseline finding, decide keep, withhold, or uncertain. Keep only
   when its actual diagnosis identifies a real reader burden and a repair can
   retain the page's technical meaning and intended author voice. Withhold
   findings that misread a necessary condition, measured claim, table scope,
   list structure, technical name, or purposeful repetition. Mark uncertain
   when page evidence cannot settle the judgment. Give a source-specific
   reason for every decision. Do not assume an existing warning is correct.

2. Independently inspect every section for missed defects in empty framing,
   wordiness, vague claims, unjustified intensifiers, needless repetition,
   and needless complexity. Look for the function of each suspect statement:
   definition, evidence, qualification, instruction, transition, authorial
   emphasis, or promotion. A claim is not defective merely because it lacks
   an external citation. Preserve exact numbers, negation, version conditions,
   identifiers, code, URLs, tables, list numbering, and the sentences that
   scope tables or examples. Do not repeat a baseline finding as a new one.

For each new diagnosis, identify the concrete reader burden and the useful
content a later repair must retain. Quote one exact contiguous slice of the
original source and its original 1-based line number. The quote must occur
only once on that line. Do not propose a replacement or deletion. If the
needed correction is merely a style preference or cannot be justified from
the complete page, return no finding for it. At most 20 new diagnoses per
page. Zero is valid.

Return exactly one review for each baseline finding ID. Echo page ID and the
source SHA-256 exactly. Do not mention or infer any hidden review labels.
"""

REVIEW_FIELDS = ("id", "decision", "reason")
FINDING_FIELDS = ("line", "quote", "diagnosis", "source_function",
                  "reader_burden", "retain")
SCHEMA = {
    "type": "object",
    "properties": {
        "page_id": {"type": "string"},
        "source_sha256": {"type": "string"},
        "reviews": {"type": "array", "items": {
            "type": "object",
            "properties": {field: {"type": "string"} for field in REVIEW_FIELDS},
            "required": list(REVIEW_FIELDS),
            "additionalProperties": False,
        }},
        "new_findings": {"type": "array", "items": {
            "type": "object",
            "properties": {
                **{"line": {"type": "integer"}},
                **{field: {"type": "string"} for field in FINDING_FIELDS[1:]},
            },
            "required": list(FINDING_FIELDS),
            "additionalProperties": False,
        }},
    },
    "required": ["page_id", "source_sha256", "reviews", "new_findings"],
    "additionalProperties": False,
}


def encoded(value):
    return (json.dumps(value, ensure_ascii=False, sort_keys=True, indent=2)
            + "\n").encode("utf-8")


def digest(data):
    return hashlib.sha256(data).hexdigest()


def file_digest(path):
    return digest(path.read_bytes())


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def page_name(number):
    return f"page-{number:03}"


def source_path(number):
    return ROOT / "sources" / f"{page_name(number)}.source"


def report_path(number):
    return ROOT / "baseline-reports" / f"{page_name(number)}.json"


def baseline_page(number):
    pages = read(ROOT / "baseline.json")["pages"]
    return next(page for page in pages if page["page"] == number)


def payload(number):
    source = source_path(number).read_bytes()
    numbered = "\n".join(f"{line:04d}| {text}" for line, text in
                         enumerate(source.decode("utf-8").splitlines(), 1))
    baseline = baseline_page(number)
    findings = [{key: value for key, value in finding.items()
                 if key in {"id", "rule_id", "message", "span", "line", "raw_slice"}}
                for finding in baseline["findings"]]
    return {"page_id": page_name(number), "source_sha256": digest(source),
            "numbered_source": numbered, "baseline_findings": findings}


def request(number):
    return (PROMPT + "\nINPUT JSON:\n" + encoded(payload(number)).decode("utf-8")).encode("utf-8")


def packet_policy():
    return {
        "format": "unswell-349-joint-coverage-triage-pilot-v1",
        "route": "existing Codex subscription; no paid API route",
        "authorization_basis": "The user gave standing explicit authorization for necessary work and asked not to be repeatedly queried in this thread.",
        "model": MODEL, "effort": EFFORT,
        "pages": list(PAGES), "calls_max": 4, "concurrency": 1,
        "attempts_per_request": 1, "automatic_retries": False,
        "tools_allowed": False,
        "per_call_timeout_seconds": PER_CALL_SECONDS,
        "total_timeout_seconds": TOTAL_SECONDS,
        "failure_policy": "Stop at first CLI, schema, trace, source-binding, or resource error; no retry or repair.",
        "input_policy": "Complete numbered public source and all exact-main baseline findings; no labels, grades, prior candidate, or prospective material.",
        "screen": "Four valid pages, at least 7/8 FULL source-only events, at least 85% accepted combined delivery, no retain-control hit or withheld accepted baseline finding; directional only.",
        "advice_policy": "No replacement or deletion requested; suggestion safety not measured.",
    }


def validate_inputs():
    sources = read(ROOT / "sources.json")
    baseline = read(ROOT / "baseline.json")
    dispositions = read(ROOT / "baseline-dispositions.json")
    labels = read(PREVIOUS / "transfer-v2-labels.json")
    assert sources["selected_pages"] == list(PAGES)
    assert [row["number"] for row in sources["pages"]] == list(PAGES)
    assert baseline["main_commit"] == "dbfe252b505294db7826cc59fc8c6b797650aa74"
    assert [row["page"] for row in baseline["pages"]] == list(PAGES)
    positives = [row for row in labels["judgments"]
                 if row["page"] in PAGES and row["decision"] == "positive"]
    assert len(positives) == 8
    assert dispositions["source_audit_sha256"] == file_digest(
        PREVIOUS / "transfer-v2-audit.json")
    expected_ids = {finding["id"] for page in baseline["pages"]
                    for finding in page["findings"]}
    assert {row["finding_id"] for row in dispositions["rows"]} == expected_ids
    assert len(dispositions["rows"]) == len(expected_ids) == 11
    for source_row, baseline_row in zip(sources["pages"], baseline["pages"]):
        raw = source_path(source_row["number"]).read_bytes()
        report_file = report_path(source_row["number"])
        report = read(report_file)
        assert len(raw) == source_row["source_bytes"]
        assert digest(raw) == source_row["source_sha256"]
        assert digest(raw) == baseline_row["source_sha256"]
        assert file_digest(report_file) == baseline_row["report_sha256"]
        assert report["status"] == "complete" and not report["errors"]
        assert report["manifest"]["tool_commit"] == baseline["main_commit"]
        assert len(report["documents"]) == 1
        assert report["documents"][0]["source_hash"] == digest(raw)
        assert {item["id"] for item in report["findings"]} == {
            item["id"] for item in baseline_row["findings"]}
        ids = set()
        for finding in baseline_row["findings"]:
            assert finding["id"] not in ids
            ids.add(finding["id"])
            span = finding["span"]
            assert 0 <= span["start"] < span["end"] <= len(raw)
            assert raw[span["start"]:span["end"]] == finding["raw_slice"].encode("utf-8")
            assert raw[:span["start"]].count(b"\n") + 1 == finding["line"]


def frozen_files():
    paths = [ROOT / name for name in
             ("PROTOCOL.md", "run.py", "test_run.py", "score.py", "test_score.py",
              "sources.json", "baseline.json", "baseline-dispositions.json")]
    paths += [PREVIOUS / name for name in
              ("transfer-v2-labels.json", "transfer-v2-source-inventory.json",
               "transfer-v2-audit.json")]
    paths += [source_path(number) for number in PAGES]
    paths += [report_path(number) for number in PAGES]
    paths += [REPO / "Makefile"]
    paths += sorted((PACKET / "requests").glob("*.txt"))
    paths += [PACKET / "schema.json", PACKET / "protocol.json"]
    return {str(path.relative_to(REPO)): file_digest(path) for path in paths}


def prepare():
    if (PACKET / "freeze.json").exists():
        raise ValueError("packet already frozen")
    validate_inputs()
    (PACKET / "requests").mkdir(parents=True, exist_ok=True)
    if list((PACKET / "requests").iterdir()):
        raise ValueError("request directory is not empty")
    for number in PAGES:
        (PACKET / "requests" / f"{page_name(number)}.txt").write_bytes(request(number))
    (PACKET / "schema.json").write_bytes(encoded(SCHEMA))
    (PACKET / "protocol.json").write_bytes(encoded(packet_policy()))
    (PACKET / "freeze.json").write_bytes(encoded(frozen_files()))
    return file_digest(PACKET / "freeze.json")


def verify():
    validate_inputs()
    if read(PACKET / "freeze.json") != frozen_files():
        raise ValueError("packet hash mismatch")
    if read(PACKET / "schema.json") != SCHEMA:
        raise ValueError("schema changed")
    if read(PACKET / "protocol.json") != packet_policy():
        raise ValueError("protocol changed")
    actual = {path.name for path in (PACKET / "requests").iterdir()}
    wanted = {f"{page_name(number)}.txt" for number in PAGES}
    if actual != wanted:
        raise ValueError("request set changed")
    for number in PAGES:
        if (PACKET / "requests" / f"{page_name(number)}.txt").read_bytes() != request(number):
            raise ValueError("request bytes changed")
    return packet_policy()


def bind(number, response):
    expected = payload(number)
    if not isinstance(response, dict) or set(response) != set(SCHEMA["required"]):
        raise ValueError("response fields changed")
    if response["page_id"] != expected["page_id"] or \
            response["source_sha256"] != expected["source_sha256"]:
        raise ValueError("page identity changed")
    reviews = response["reviews"]
    if not isinstance(reviews, list) or len(reviews) != len(expected["baseline_findings"]):
        raise ValueError("baseline review coverage changed")
    wanted = {finding["id"] for finding in expected["baseline_findings"]}
    seen = set()
    for review in reviews:
        if not isinstance(review, dict) or set(review) != set(REVIEW_FIELDS):
            raise ValueError("review fields changed")
        if review["id"] not in wanted or review["id"] in seen:
            raise ValueError("missing, unknown, or duplicate baseline ID")
        seen.add(review["id"])
        if review["decision"] not in {"keep", "withhold", "uncertain"}:
            raise ValueError("invalid review decision")
        if not isinstance(review["reason"], str) or not review["reason"].strip():
            raise ValueError("empty review reason")
    if seen != wanted:
        raise ValueError("missing baseline ID")
    findings = response["new_findings"]
    if not isinstance(findings, list) or len(findings) > 20:
        raise ValueError("new finding limit or type")
    raw = source_path(number).read_bytes()
    excluded = baseline_page(number)["excluded_spans"]
    bound = []
    for finding in findings:
        if not isinstance(finding, dict) or set(finding) != set(FINDING_FIELDS):
            raise ValueError("new finding fields changed")
        line, quote = finding["line"], finding["quote"]
        if type(line) is not int or line < 1 or not isinstance(quote, str) or not quote.strip():
            raise ValueError("invalid line or quote")
        if any(not isinstance(finding[key], str) or not finding[key].strip()
               for key in FINDING_FIELDS[2:]):
            raise ValueError("empty diagnosis, function, burden, or retention")
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
        start, end = matches[0], matches[0] + len(q)
        if any(start < span["end"] and span["start"] < end for span in excluded):
            raise ValueError("new finding overlaps protected bytes")
        bound.append(dict(finding, id=f"{page_name(number)}/new-{len(bound) + 1:02}",
                          start_byte=start, end_byte=end))
    return {"page_id": expected["page_id"], "source_sha256": expected["source_sha256"],
            "reviews": reviews, "new_findings": bound}


def verify_trace(target, response):
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
        raise ValueError("run directory already exists")
    RUN.mkdir()
    journal = {"status": "running", "complete": False, "calls": [],
               "freeze_sha256": file_digest(PACKET / "freeze.json"),
               "model": MODEL, "effort": EFFORT, "monetary_cost": None}
    (RUN / "run.json").write_bytes(encoded(journal))
    started = time.monotonic()
    with tempfile.TemporaryDirectory(prefix="unswell-349-joint-empty-") as empty:
        for number in PAGES:
            remaining = TOTAL_SECONDS - (time.monotonic() - started)
            if remaining <= 0:
                journal["stop_reason"] = "total_timeout"
                break
            name = page_name(number)
            target = RUN / name
            target.mkdir()
            args = ["codex", "exec", *OPTIONS, "--model", MODEL, "--cd", empty,
                    "--output-schema", str(PACKET / "schema.json"),
                    "--output-last-message", str(target / "response.json"),
                    "-c", 'model_reasoning_effort="medium"']
            for setting in SETTINGS:
                args.extend(["-c", setting])
            args.append("-")
            row = {"page": number, "status": "started",
                   "request_sha256": file_digest(PACKET / "requests" / f"{name}.txt")}
            journal["calls"].append(row)
            (RUN / "run.json").write_bytes(encoded(journal))
            before = time.monotonic()
            try:
                process = subprocess.Popen(args, stdin=subprocess.PIPE,
                                           stdout=subprocess.PIPE,
                                           stderr=subprocess.PIPE,
                                           start_new_session=True)
                try:
                    stdout, stderr = process.communicate(
                        (PACKET / "requests" / f"{name}.txt").read_bytes(),
                        timeout=min(PER_CALL_SECONDS, remaining))
                except subprocess.TimeoutExpired:
                    os.killpg(process.pid, signal.SIGKILL)
                    stdout, stderr = process.communicate()
                    (target / "events.jsonl").write_bytes(stdout)
                    (target / "stderr.txt").write_bytes(stderr)
                    raise TimeoutError("model call timed out")
                (target / "events.jsonl").write_bytes(stdout)
                (target / "stderr.txt").write_bytes(stderr)
                if process.returncode != 0:
                    raise RuntimeError(f"Codex CLI returned {process.returncode}")
                response = read(target / "response.json")
                verify_trace(target, response)
                bound = bind(number, response)
                (target / "bound.json").write_bytes(encoded(bound))
                row["status"] = "complete"
                row["response_sha256"] = file_digest(target / "response.json")
                row["bound_sha256"] = file_digest(target / "bound.json")
                row["elapsed_seconds"] = round(time.monotonic() - before, 3)
            except (OSError, ValueError, RuntimeError, TypeError, KeyError) as exc:
                row["status"] = "error"
                row["error"] = str(exc)
                row["elapsed_seconds"] = round(time.monotonic() - before, 3)
                journal["stop_reason"] = "first_error"
                break
            finally:
                (RUN / "run.json").write_bytes(encoded(journal))
    journal["complete"] = len(journal["calls"]) == len(PAGES) and all(
        row["status"] == "complete" for row in journal["calls"])
    journal["status"] = "complete" if journal["complete"] else "stopped"
    (RUN / "run.json").write_bytes(encoded(journal))
    return journal


if __name__ == "__main__":
    action = sys.argv[1] if len(sys.argv) == 2 else ""
    if action == "prepare":
        print(prepare())
    elif action == "verify":
        print(json.dumps(verify(), sort_keys=True))
    elif action == "execute":
        result = execute()
        print(json.dumps({"status": result["status"], "calls": result["calls"]},
                         sort_keys=True))
        raise SystemExit(0 if result["complete"] else 1)
    else:
        raise SystemExit("usage: run.py prepare|verify|execute")
