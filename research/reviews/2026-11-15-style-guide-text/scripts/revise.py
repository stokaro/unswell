#!/usr/bin/env python3
"""Have a model revise each page by the report of `unswell check`, and read the index before and after.

usage: revise.py <workspace> <unswell binary> <model kind> <cache dir> <out jsonl> <label>=<jsonl>... [--split dev|test] [--limit N] [--workers N]

The model gets the page and the lines of the report about the page index: the index line, one line for each
listed mark, and the lines about what a command can replace. It is told to resolve what the report points at and
to change nothing else. A page whose report lists no mark is left out. A row holds the page, the revision, and
the index, the readings and the level of both.
"""
import concurrent.futures, json, subprocess, sys, tempfile
from pathlib import Path

W, binary, kind, cache, out = sys.argv[1].rstrip("/") + "/", sys.argv[2], sys.argv[3], sys.argv[4], Path(sys.argv[5])
option = lambda name, default: sys.argv[sys.argv.index(name) + 1] if name in sys.argv else default
split, limit, workers = option("--split", "dev"), int(option("--limit", "100000")), int(option("--workers", "6"))
INSTRUCTION = ("Below is a Markdown page and the report of a prose linter about it. Revise the page so that what the report points at "
               "is resolved. Change nothing else: keep every other sentence, heading, list, table, link, and code block as it is, and "
               "keep all facts. Answer with the complete revised page and nothing else, with no code fence around it.\n\n")


# For a pilot: other words for the remedy of a run, and another instruction, to try before the report changes.
PLAIN_DEFAULT = "join two of them, or add the condition or the reason that one leaves out"
OPENER_DEFAULT = "open one of them with its condition, its time, or its reason"
plain_action, opener_action = option("--plain-action", PLAIN_DEFAULT), option("--opener-action", OPENER_DEFAULT)
if "--instruction" in sys.argv:
    INSTRUCTION = Path(option("--instruction", "")).read_text().rstrip("\n") + "\n\n"
no_report = "--no-report" in sys.argv


def check(text):
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--no-gate", "--allow-empty",
                               "--timeout", "2m", "--report", "json:index.json", "--report", "text:-"], input=text.encode(), cwd=scratch, capture_output=True)
        index = json.loads((Path(scratch) / "index.json").read_text())["documents"][0].get("ai_sounding") or {}
    lines = done.stdout.decode().splitlines()
    start = next((i for i, l in enumerate(lines) if l.startswith("AI-sounding index:")), None)
    report = [l for l in lines[start:] if not l.startswith(("The index is the higher", "Revision probability"))] if start is not None else []
    return index, "\n".join(report).replace(PLAIN_DEFAULT, plain_action).replace(OPENER_DEFAULT, opener_action)


def one(job):
    label, row = job
    page = "# " + (row.get("title") or "Post") + "\n\n" + row["text"]
    before, report = check(page)
    marks = sum(1 for l in report.splitlines() if l.startswith("  page.md:"))
    if before.get("status") != "available" or marks == 0:
        return {"set": label, "id": row["id"], "skipped": True, "index": before.get("index"), "level": before.get("level")}
    prompt = INSTRUCTION + ("" if no_report else "===== REPORT =====\n" + report + "\n\n") + "===== PAGE =====\n" + page
    answer = subprocess.run(["python3", W + "ask_model.py", kind, cache], input=prompt.encode(), capture_output=True).stdout.decode()
    if len(answer.split()) < 0.5 * len(page.split()):
        return {"set": label, "id": row["id"], "failed": True}
    after, _ = check(answer)
    keep = lambda i: {"index": i.get("index"), "level": i.get("level"), "readings": i.get("readings"), "words": i.get("prose_words")}
    return {"set": label, "id": row["id"], "original": page, "revised": answer, "report": report, "before": keep(before), "after": keep(after)}


done = {(r["set"], r["id"]) for r in map(json.loads, out.read_text().splitlines())} if out.exists() else set()
jobs = []
for spec in [a for a in sys.argv[6:] if "=" in a]:
    label, path = spec.split("=", 1)
    rows = [r for r in map(json.loads, open(path)) if not r.get("failed") and r.get("split", split) == split]
    jobs += [(label, r) for r in rows[:limit] if (label, r["id"]) not in done]
with out.open("a") as handle, concurrent.futures.ThreadPoolExecutor(workers) as pool:
    for row in pool.map(one, jobs):
        handle.write(json.dumps(row, ensure_ascii=False) + "\n")
        handle.flush()
print("done", len(jobs))
