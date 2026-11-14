#!/usr/bin/env python3
"""Turn fixed versions of pages into blind comparisons for the panel.

usage: fix_pairs.py <arms jsonl> <pairs jsonl> <unswell binary>

For every page: the original against each fixed version, and the version fixed by unswell's report and the version
that `unswell fix` leaves against each version fixed by a third-party linter's report. A comparison of two identical texts is left out; the summary counts it
as "same".
"""
import json, subprocess, sys, tempfile
from pathlib import Path

ARMS = ("unswell", "google", "microsoft", "write-good", "mechanical")
LINTERS = ("google", "microsoft", "write-good")
# The mechanical version is what the product's own command writes: `unswell fix --write` on a copy of the page.
PRODUCT = sys.argv[3]


def product_fix(text):
    with tempfile.TemporaryDirectory() as scratch:
        page = Path(scratch) / "page.md"
        page.write_text(text)
        done = subprocess.run([PRODUCT, "fix", "page.md", "--write", "--quiet"], cwd=scratch, capture_output=True)
        if done.returncode != 0:
            raise SystemExit("fix failed: " + done.stderr.decode()[:300])
        return page.read_text()


rows = [json.loads(line) for line in open(sys.argv[1])]
for row in rows:
    if "python_mechanical" not in row:
        row["python_mechanical"] = row["arms"]["mechanical"]
        row["arms"]["mechanical"] = product_fix(row["original"])
with open(sys.argv[1], "w") as arms:
    for row in rows:
        arms.write(json.dumps(row, ensure_ascii=False) + "\n")
count = 0
with open(sys.argv[2], "w") as out:
    for row in rows:
        versions = {"original": row["original"], **row["arms"]}
        for first, second in [("original", arm) for arm in ARMS] + [(ours, linter) for ours in ("unswell", "mechanical") for linter in LINTERS]:
            a, b = versions.get(first), versions.get(second)
            if a is None or b is None or a.strip() == b.strip():
                continue
            out.write(json.dumps({"id": f"{row['id']}|{first}|{second}", "first": a, "second": b, "first_is": first}, ensure_ascii=False) + "\n")
            count += 1
print("comparisons", count)
