#!/usr/bin/env python3
"""Check that a binary reproduces the reports and the indexes behind the revisions of the record.

usage: reproduce.py <unswell binary> <revisions jsonl>...

For each revised page the binary has to print the report that the reviser was given, and to give the page and
its revision the index, the readings, and the level that the measurement recorded.
"""
import concurrent.futures, json, subprocess, sys, tempfile
from pathlib import Path

binary = sys.argv[1]


def check(text):
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--no-gate", "--allow-empty",
                               "--timeout", "2m", "--report", "json:index.json", "--report", "text:-"], input=text.encode(), cwd=scratch, capture_output=True)
        index = json.loads((Path(scratch) / "index.json").read_text())["documents"][0].get("ai_sounding") or {}
    lines = done.stdout.decode().splitlines()
    start = next((i for i, l in enumerate(lines) if l.startswith("AI-sounding index:")), None)
    report = [l for l in lines[start:] if not l.startswith(("The index is the higher", "Revision probability"))] if start is not None else []
    return {"index": index.get("index"), "level": index.get("level"), "readings": index.get("readings"), "words": index.get("prose_words")}, "\n".join(report)


def one(row):
    before, report = check(row["original"])
    after, _ = check(row["revised"])
    return report == row["report"], before == row["before"], after == row["after"]


for path in sys.argv[2:]:
    rows = [r for r in map(json.loads, open(path)) if "revised" in r]
    with concurrent.futures.ThreadPoolExecutor(6) as pool:
        outcomes = list(pool.map(one, rows))
    print(f"{Path(path).name}: {len(rows)} revised pages; the same report on {sum(o[0] for o in outcomes)},"
          f" the same index of the page on {sum(o[1] for o in outcomes)}, the same index of the revision on {sum(o[2] for o in outcomes)}")
