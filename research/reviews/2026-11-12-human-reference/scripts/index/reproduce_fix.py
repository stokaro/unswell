#!/usr/bin/env python3
"""Check that a binary writes the same fixed pages as the ones the judges read.

usage: reproduce_fix.py <unswell binary> <arms jsonl>...

For every page of the given runs the binary runs `unswell fix --write` on a copy of the original, and the result is
compared byte for byte with the "mechanical" version of the run. A page that the run recorded as refused must be
refused again.
"""
import concurrent.futures, json, subprocess, sys, tempfile
from pathlib import Path

binary = sys.argv[1]


def fix(text):
    with tempfile.TemporaryDirectory() as scratch:
        page = Path(scratch) / "page.md"
        page.write_text(text)
        done = subprocess.run([binary, "fix", "page.md", "--write", "--quiet"], cwd=scratch, capture_output=True)
        return page.read_text() if done.returncode == 0 else None


for path in sys.argv[2:]:
    rows = [json.loads(line) for line in open(path)]
    with concurrent.futures.ThreadPoolExecutor(6) as pool:
        again = list(pool.map(fix, [row["original"] for row in rows]))
    same = sum(new == row["arms"]["mechanical"] for new, row in zip(again, rows))
    refused = sum(new is None for new in again)
    print(f"{path}: {len(rows)} pages, {same} reproduced byte for byte, {refused} refused")
