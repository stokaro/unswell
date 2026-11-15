#!/usr/bin/env python3
"""Have a model revise each page by a report that also says what the sentences of the page are like.

usage: revise.py <workspace> <unswell binary> <model kind> <cache dir> <out jsonl> <label>=<jsonl>... [--split dev|test]
                      [--workers N] [--lines <file>] [--instruction <file>] [--keep-report]

A pilot for a line that the report does not print yet. The model gets the page and a report. The report is the
lines of `unswell check` about the page index, unless --keep-report is absent and --lines replaces them, followed by
the lines of the file given with --lines. In that file {n}, {plain}, {plain_share}, {long}, {long_share}, {mean},
and {wanted_long} stand for the sentences of the page in running paragraphs: how many there are, how many are plain
statements of five to twenty words, how many have more than twenty words, their mean length, and how many long
sentences a third of them would be. A row holds the page, the revision, and the index and the sentences of both.
"""
import concurrent.futures, json, subprocess, sys, tempfile
from pathlib import Path

HERE = Path(__file__).parent
exported = {"__file__": str(HERE.parents[1] / "2026-11-15-style-guide-text" / "scripts" / "control_export.py"), "__name__": "control_export"}
exec(compile((HERE.parents[1] / "2026-11-15-style-guide-text" / "scripts" / "control_export.py").read_text().rsplit("\nmain()\n", 1)[0], "control_export.py", "exec"), exported)
rebuild = exported["rebuild"]

W, binary, kind, cache, out = sys.argv[1].rstrip("/") + "/", sys.argv[2], sys.argv[3], sys.argv[4], Path(sys.argv[5])
option = lambda name, default: sys.argv[sys.argv.index(name) + 1] if name in sys.argv else default
split, workers = option("--split", "dev"), int(option("--workers", "6"))
INSTRUCTION = ("Below is a Markdown page and the report of a prose linter about it. Revise the page so that what the report points at "
               "is resolved. Change nothing else: keep every heading, list, table, link, and code block as it is, and keep all facts. "
               "Answer with the complete revised page and nothing else, with no code fence around it.\n\n")
if "--instruction" in sys.argv:
    INSTRUCTION = Path(option("--instruction", "")).read_text().rstrip("\n") + "\n\n"
lines_template = Path(option("--lines", "")).read_text().rstrip("\n") if "--lines" in sys.argv else ""
keep_report = "--keep-report" in sys.argv
# As in the record: a page whose report lists no mark is not given to the reviser.
skip_unlisted = "--skip-unlisted" in sys.argv
limit = int(option("--limit", "100000"))


def check(text):
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--no-gate", "--allow-empty",
                               "--timeout", "2m", "--report", "json:result.json", "--report", "text:-"], input=text.encode(), cwd=scratch,
                              capture_output=True)
        result = json.loads((Path(scratch) / "result.json").read_text())
    index = result["documents"][0].get("ai_sounding") or {}
    lines = done.stdout.decode().splitlines()
    start = next((i for i, l in enumerate(lines) if l.startswith("AI-sounding index:")), None)
    report = [l for l in lines[start:] if not l.startswith(("The index is the higher", "Revision probability"))] if start is not None else []
    sentences = rebuild(text, result)
    n = len(sentences)
    figures = {"n": n, "plain": sum(s[2] for s in sentences), "long": sum(s[1] > 20 for s in sentences),
               "mean": round(sum(s[1] for s in sentences) / n, 1) if n else 0}
    return index, "\n".join(report), figures


def filled(figures):
    n = max(1, figures["n"])
    return lines_template.format(n=figures["n"], plain=figures["plain"], plain_share=round(100 * figures["plain"] / n), long=figures["long"],
                                 long_share=round(100 * figures["long"] / n), mean=figures["mean"], wanted_long=round(n / 3))


def one(job):
    label, row = job
    page = "# " + (row.get("title") or "Post") + "\n\n" + row["text"]
    before, report, figures = check(page)
    listed = sum(1 for line in report.splitlines() if line.startswith("  page.md:"))
    if before.get("status") != "available" or (figures["n"] < 10 and not skip_unlisted) or (skip_unlisted and listed == 0):
        return {"set": label, "id": row["id"], "skipped": True, "index": before.get("index"), "level": before.get("level")}
    shown = "\n".join(part for part in ((report if keep_report or not lines_template else ""), filled(figures) if lines_template else "") if part)
    prompt = INSTRUCTION + "===== REPORT =====\n" + shown + "\n\n===== PAGE =====\n" + page
    answer = subprocess.run(["python3", W + "ask_model.py", kind, cache], input=prompt.encode(), capture_output=True).stdout.decode()
    if len(answer.split()) < 0.5 * len(page.split()):
        return {"set": label, "id": row["id"], "failed": True}
    after, _, figures_after = check(answer)
    keep = lambda i: {"index": i.get("index"), "level": i.get("level"), "readings": i.get("readings"), "words": i.get("prose_words")}
    return {"set": label, "id": row["id"], "original": page, "revised": answer, "report": shown, "before": keep(before), "after": keep(after),
            "sentences_before": figures, "sentences_after": figures_after}


done = {(r["set"], r["id"]) for r in map(json.loads, out.read_text().splitlines())} if out.exists() else set()
jobs = []
for spec in [a for a in sys.argv[6:] if "=" in a]:
    label, path = spec.split("=", 1)
    rows = [r for r in map(json.loads, open(path)) if not r.get("failed") and r.get("split", split) == split]
    rows.sort(key=lambda r: __import__("hashlib").sha256(("pick:" + r["id"]).encode()).hexdigest())
    jobs += [(label, r) for r in rows[:limit] if (label, r["id"]) not in done]
with out.open("a") as handle, concurrent.futures.ThreadPoolExecutor(workers) as pool:
    for row in pool.map(one, jobs):
        handle.write(json.dumps(row, ensure_ascii=False) + "\n")
        handle.flush()
print("done", len(jobs))
