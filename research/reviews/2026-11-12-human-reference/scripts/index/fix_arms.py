#!/usr/bin/env python3
"""Build fixed versions of pages: by a model that follows a linter's report, and mechanically for unswell's marks.

usage: fix_arms.py <workspace> <unswell binary> <model kind> <sample jsonl> <linter results jsonl> <out jsonl> [--workers N]

A sample row holds "id" and "text" (the page before a de-slop commit). Arms:
  unswell     a model gets the page and the text report of `unswell check` (findings, the page index, and the places
              of its marks) and resolves what it points at
  google      the same with the findings of Vale's Google style
  microsoft   the same with the findings of Vale's Microsoft style
  write-good  the same with the findings of write-good
  mechanical  no model: em dashes become commas (a colon after a bold label), emoji leave headings and the start of
              list items, arrows become "->", bold labels at the start of list items lose their asterisks
The model is told to change nothing else and to answer with the complete page. Answers are cached by prompt.
"""
import concurrent.futures, json, re, subprocess, sys, tempfile
from pathlib import Path

W, binary, kind, sample_path, linters_path, out = sys.argv[1].rstrip("/") + "/", sys.argv[2], sys.argv[3], sys.argv[4], sys.argv[5], Path(sys.argv[6])
workers = int(sys.argv[sys.argv.index("--workers") + 1]) if "--workers" in sys.argv else 8
ASK, CACHE = W + "ask_model.py", str(out) + ".cache"
EMOJI = "[\U0001F300-\U0001FAFF☀-➿⭐✨⚡][️‍]*"
INSTRUCTION = ("Below is a Markdown page and the report of a prose linter about it. Revise the page so that what the report points at "
               "is resolved. Change nothing else: keep every other sentence, heading, list, table, link, and code block as it is, and "
               "keep all facts. Answer with the complete revised page and nothing else, with no code fence around it.\n\n")
linters = {}
for line in open(linters_path):
    row = json.loads(line)
    if "tool" in row and not row.get("failed") and row["file"].endswith("-before.md"):
        linters[(row["tool"], row["file"][:-len("-before.md")])] = row["spans"]


def unswell_report(text):
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--no-gate", "--allow-empty"],
                              input=text.encode(), cwd=scratch, capture_output=True)
    lines = [l for l in done.stdout.decode().splitlines()
             if not l.startswith(("Revision probability", "Configuration:", "ADVISORY", "PASS", "FAIL", "The index counts"))]
    return "\n".join(lines)


def linter_report(text, spans):
    starts = [0]
    for line in text.splitlines(keepends=True):
        starts.append(starts[-1] + len(line))
    import bisect
    lines = []
    for begin, end, check in sorted(spans):
        number = bisect.bisect_right(starts, begin)
        lines.append(f"line {number}: \"{text[begin:end]}\" [{check}]")
    return "\n".join(lines)


def ask(prompt):
    return subprocess.run(["python3", ASK, kind, CACHE], input=prompt.encode(), capture_output=True).stdout.decode()


def mechanical(text):
    result, fenced = [], False
    for line in text.split("\n"):
        if re.match(r"\s*(```|~~~)", line):
            fenced = not fenced
        if fenced:
            result.append(line)
            continue
        line = re.sub(r"^(\s*(?:[-*+]|\d+[.)])\s+(?:\[[ xX]\]\s+)?)\*\*([^*\n]+)\*\*\s*—\s*", r"\1\2: ", line)
        line = re.sub(r"^(\s*(?:[-*+]|\d+[.)])\s+(?:\[[ xX]\]\s+)?)\*\*([^*\n]+)\*\*", r"\1\2", line)
        line = re.sub(r"^(#{1,6}\s+)((?:" + EMOJI + r"\s*)+)", r"\1", line)
        if line.startswith("#"):
            line = re.sub(r"\s*" + EMOJI, "", line)
        line = re.sub(r"^(\s*(?:[-*+]|\d+[.)])\s+(?:\[[ xX]\]\s+)?)(?:" + EMOJI + r"\s*)+", r"\1", line)
        line = re.sub(r"\s+—\s+", ", ", line).replace("—", ", ")
        line = line.replace("→", "->").replace("⇒", "=>")
        result.append(line)
    return "\n".join(result)


def reports_of(row):
    identifier, text = row["id"], row["text"]
    return {"unswell": unswell_report(text),
            "google": linter_report(text, linters.get(("vale:Google", identifier), [])),
            "microsoft": linter_report(text, linters.get(("vale:Microsoft", identifier), [])),
            "write-good": linter_report(text, linters.get(("write-good", identifier), []))}


def fix(task):
    row, name, report = task
    text = row["text"]
    if not report.strip():
        return row["id"], name, text
    answer = ask(INSTRUCTION + "===== REPORT =====\n" + report + "\n\n===== PAGE =====\n" + text)
    return row["id"], name, (answer if len(answer.split()) > 0.5 * len(text.split()) else None)


done = {json.loads(line)["id"] for line in out.read_text().splitlines()} if out.exists() else set()
rows = [json.loads(line) for line in open(sample_path)]
rows = [row for row in rows if row["id"] not in done]
reports = {row["id"]: reports_of(row) for row in rows}
tasks = [(row, name, report) for row in rows for name, report in reports[row["id"]].items()]
arms = {row["id"]: {"mechanical": mechanical(row["text"])} for row in rows}
with concurrent.futures.ThreadPoolExecutor(workers) as pool:
    for identifier, name, text in pool.map(fix, tasks):
        arms[identifier][name] = text
with out.open("a") as handle:
    for row in rows:
        handle.write(json.dumps({"id": row["id"], "original": row["text"], "arms": arms[row["id"]], "reports": reports[row["id"]],
                                 "report_lines": {n: len(r.splitlines()) for n, r in reports[row["id"]].items()}}, ensure_ascii=False) + "\n")
print(f"done: {len(rows)} pages")
