"""Freeze and execute a paired, repository-disjoint editorial rubric screen."""

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
PACKET = ROOT / "packet"
RUN = ROOT / "run"
BASE_RUNNER = ROOT.parent / "2026-11-08-evidence-flip-pilot" / "run.py"
SOURCE_RUNNER = ROOT.parent / "2026-11-08-real-source-admission-transfer-v2" / "run.py"
OLD_SAMPLE = ROOT.parent / "2026-11-08-real-source-admission-transfer-v2" / "packet/sample.json"
MAX_CALLS = 24
TOTAL_SECONDS = 10800
PER_CALL_SECONDS = 600
ARMS = ("control", "rubric")


def module(path, name):
    spec = importlib.util.spec_from_file_location(name, path)
    result = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(result)
    return result


def base():
    return module(BASE_RUNNER, "admission_base")


def original_sources():
    return module(SOURCE_RUNNER, "real_source_runner").original_sources()


def encoded(value):
    return (json.dumps(value, ensure_ascii=False, sort_keys=True, indent=2) + "\n").encode("utf-8")


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def file_sha(path):
    return sha(path.read_bytes())


def case_entries():
    sample = read(PACKET / "sample.json")
    key = read(PACKET / "key.json")
    examples = read(PACKET / "examples.json")
    if sample["format"] != "unswell-349-rubric-transfer-sample-v1" or \
            len(sample["items"]) != 12 or len(key) != 12 or \
            len(examples) != 22 or sample["source_pages"] != 36:
        raise ValueError("sample inventory changed")
    if len({e["source_group"] for e in examples}) != 11 or \
            any(e["source_group"] == "ptah" for e in examples) or \
            {e["prior_grade"] for e in examples} != {"accepted", "uncertain"}:
        raise ValueError("rubric example inventory changed")
    old_pages = {item["page"] for item in read(OLD_SAMPLE)["items"]}
    sources = original_sources()
    cases = []
    for item, grade in zip(sample["items"], key):
        if (set(item) != {"case_id", "candidate_id", "category", "group", "page",
                          "source_sha256", "target_start_byte", "target_end_byte",
                          "target_quote", "proposed_diagnosis"}
                or set(grade) != {"case_id", "candidate_id", "old_status",
                                      "old_rationale"}):
            raise ValueError("sample or grade fields changed")
        if item["case_id"] != grade["case_id"] or \
                item["candidate_id"] != grade["candidate_id"] or \
                item["group"] != "ptah" or item["page"] in old_pages or \
                grade["old_status"] not in ("accepted", "uncertain") or \
                not grade["old_rationale"].strip():
            raise ValueError("sample identity or page isolation changed")
        path, raw = sources[item["source_sha256"]]
        start, end = item["target_start_byte"], item["target_end_byte"]
        if not 0 <= start < end <= len(raw) or not 20 <= end - start <= 300 or \
                raw[start:end] != item["target_quote"].encode("utf-8") or \
                len(raw) > 65536 or not item["proposed_diagnosis"].strip():
            raise ValueError("source binding or resource limit changed")
        cases.append((item, grade, path, raw.decode("utf-8")))
    if len({item["case_id"] for item, _, _, _ in cases}) != 12 or \
            len({item["page"] for item, _, _, _ in cases}) != 12 or \
            sum(grade["old_status"] == "accepted" for _, grade, _, _ in cases) != 6:
        raise ValueError("page-disjoint balanced case inventory changed")
    return cases


def entries():
    pairs = []
    for item, grade, path, source in case_entries():
        for arm in ARMS:
            item_id = sha((item["case_id"] + "\n" + arm).encode("utf-8"))[:16]
            payload = {"item_id": item_id, "source_sha256": item["source_sha256"],
                       "source": source, "target_quote": item["target_quote"],
                       "target_start_byte": item["target_start_byte"],
                       "target_end_byte": item["target_end_byte"],
                       "proposed_diagnosis": item["proposed_diagnosis"]}
            pairs.append((payload, item["case_id"], arm, grade, path))
    if len({payload["item_id"] for payload, _, _, _, _ in pairs}) != MAX_CALLS:
        raise ValueError("paired request identity changed")
    return pairs


def request(payload, arm):
    if arm == "control":
        return base().request(payload)
    if arm != "rubric":
        raise ValueError("unknown arm")
    examples = read(PACKET / "examples.json")
    calibration = """
The examples below are prior assistant editorial grades from OTHER repositories.
Treat them as calibration data, not as facts about the current source or a rule
that every short rewrite is needed. `uncertain` means the prior review did not
establish an independently useful edit; it does not prove the diagnosis false.
Apply the same conservative standard to the current complete page. Distinguish
a demonstrated defect from an optional rephrasing or the author's useful
emphasis. In particular, a shorter safe sentence is not sufficient to admit.
Check the exact supplied diagnosis, the source-specific cost of leaving the
target unchanged, and whether a minimal edit preserves all factual distinctions,
conditions, negation, modality, identifiers, and technical boundaries. If the
source or examples cannot settle necessity or meaning preservation, abstain.
Do not copy any example's wording into the current replacement.

PRIOR EDITORIAL EXAMPLES JSON:
"""
    return (base().PROMPT + calibration + encoded(examples).decode("utf-8") +
            "\nCURRENT INPUT JSON:\n" + encoded(payload).decode("utf-8")).encode("utf-8")


def files():
    return {str(path.relative_to(PACKET)): file_sha(path)
            for path in PACKET.rglob("*") if path.is_file() and path.name != "freeze.json"}


def protocol_record():
    model = base()
    sources = {digest: path for digest, (path, _) in sorted(original_sources().items())}
    return {
        "format": "unswell-349-rubric-transfer-protocol-v1",
        "model": model.MODEL, "effort": model.EFFORT,
        "route": "existing Codex subscription; no paid API route",
        "calls_max": MAX_CALLS, "concurrency": 1, "attempts_per_request": 1,
        "automatic_retries": False, "tools_allowed": False,
        "per_call_timeout_seconds": PER_CALL_SECONDS,
        "total_timeout_seconds": TOTAL_SECONDS,
        "base_runner_sha256": file_sha(BASE_RUNNER),
        "source_runner_sha256": file_sha(SOURCE_RUNNER),
        "old_sample_sha256": file_sha(OLD_SAMPLE),
        "sample_script_sha256": file_sha(ROOT / "sample.py"),
        "runner_sha256": file_sha(Path(__file__)),
        "scorer_sha256": file_sha(ROOT / "score.py"),
        "protocol_md_sha256": file_sha(ROOT / "PROTOCOL.md"),
        "original_source_manifest_sha256": sha(encoded(sources)),
        "sample_sha256": file_sha(PACKET / "sample.json"),
        "key_sha256": file_sha(PACKET / "key.json"),
        "examples_sha256": file_sha(PACKET / "examples.json"),
        "screen": "24 valid bound calls; rubric >=10/12 prior-grade agreements and >=3 more than control; >=5/6 accepted retained and >=5/6 uncertain withheld; all rubric admitted edits safe.",
        "failure_policy": "Stop at first CLI, timeout, schema, trace, source-binding, or resource error; no retry or repair.",
    }


def prepare():
    if (PACKET / "freeze.json").exists() or (PACKET / "requests").exists():
        raise ValueError("packet already frozen")
    pairs = entries()
    (PACKET / "requests").mkdir()
    for payload, _, arm, _, _ in pairs:
        (PACKET / "requests" / (payload["item_id"] + ".txt")).write_bytes(request(payload, arm))
    (PACKET / "schema.json").write_bytes(encoded(base().SCHEMA))
    (PACKET / "protocol.json").write_bytes(encoded(protocol_record()))
    (PACKET / "freeze.json").write_bytes(encoded(files()))
    return {"calls": len(pairs), "freeze_sha256": file_sha(PACKET / "freeze.json")}


def verify():
    pairs = entries()
    expected = {payload["item_id"] + ".txt": request(payload, arm)
                for payload, _, arm, _, _ in pairs}
    actual = {p.name: p.read_bytes() for p in (PACKET / "requests").iterdir()
              if p.is_file()}
    if actual != expected or read(PACKET / "schema.json") != base().SCHEMA or \
            read(PACKET / "protocol.json") != protocol_record() or \
            read(PACKET / "freeze.json") != files():
        raise ValueError("frozen packet or requests changed")
    return read(PACKET / "protocol.json")


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
    with tempfile.TemporaryDirectory(prefix="unswell-349-rubric-empty-") as empty:
        for payload, case_id, arm, _, _ in pairs:
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
            row = {"item_id": item_id, "case_id": case_id, "arm": arm,
                   "status": "started",
                   "request_sha256": file_sha(PACKET / "requests" / (item_id + ".txt"))}
            journal["calls"].append(row)
            (RUN / "run.json").write_bytes(encoded(journal))
            before = time.monotonic()
            try:
                process = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                           stderr=subprocess.PIPE, start_new_session=True)
                try:
                    stdout, stderr = process.communicate(request(payload, arm),
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
                if len(stdout) > (8 << 20) or len(stderr) > (1 << 20) or \
                        (target / "response.json").stat().st_size > (1 << 20):
                    raise ValueError("response resource limit")
                response = model.check_trace(target, payload)
                (target / "bound.json").write_bytes(encoded(response))
                row.update(status="valid", decision=response["decision"],
                           bound_sha256=file_sha(target / "bound.json"))
            except (ValueError, KeyError, TypeError, OSError) as error:
                row.update(status="invalid", error=str(error))
                journal["stop_reason"] = "first_failed_request_no_retry"
            row["seconds"] = round(time.monotonic() - before, 3)
            (RUN / "run.json").write_bytes(encoded(journal))
            print(json.dumps(row, sort_keys=True), flush=True)
            if row["status"] != "valid":
                break
    journal["complete"] = len(journal["calls"]) == MAX_CALLS and \
        all(row["status"] == "valid" for row in journal["calls"])
    journal["status"] = "complete" if journal["complete"] else "stopped"
    journal["unattempted"] = [payload["item_id"]
                              for payload, _, _, _, _ in pairs[len(journal["calls"]):]]
    (RUN / "run.json").write_bytes(encoded(journal))
    return {"complete": journal["complete"], "calls": len(journal["calls"]),
            "run_sha256": file_sha(RUN / "run.json")}


if __name__ == "__main__":
    operation = sys.argv[1:] or ["verify"]
    result = prepare() if operation == ["prepare"] else execute() if operation == ["execute"] else verify()
    print(json.dumps(result, sort_keys=True))
