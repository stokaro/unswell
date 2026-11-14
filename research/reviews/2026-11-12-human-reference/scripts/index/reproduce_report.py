#!/usr/bin/env python3
"""Check that a binary prints the same report as the one the fixer model was given.

usage: reproduce_report.py <unswell binary> <arms jsonl>...

For every page of the given runs the binary runs `unswell check` on the original as fix_arms.py does, and the report
is compared with the "unswell" report that the run stored for the page.
"""
import concurrent.futures, json, subprocess, sys, tempfile

binary = sys.argv[1]
DROPPED = ("Revision probability", "Configuration:", "ADVISORY", "PASS", "FAIL", "The index counts")


def report(text):
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--no-gate", "--allow-empty"],
                              input=text.encode(), cwd=scratch, capture_output=True)
    return "\n".join(line for line in done.stdout.decode().splitlines() if not line.startswith(DROPPED))


for path in sys.argv[2:]:
    rows = [json.loads(line) for line in open(path)]
    with concurrent.futures.ThreadPoolExecutor(6) as pool:
        again = list(pool.map(report, [row["original"] for row in rows]))
    same = sum(new == row["reports"]["unswell"] for new, row in zip(again, rows))
    print(f"{path}: {len(rows)} pages, {same} reports reproduced")
