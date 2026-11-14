#!/usr/bin/env python3
"""Keep one commit per repository from a file of found commits: a strict message, not in a fork.

usage: fresh_commits.py <workspace> <commits jsonl> <out jsonl>

The first commit of a repository by its hash is kept. A repository that the output already holds keeps its commit, so
the script can run again while the search still adds commits.
"""
import json, sys
from pathlib import Path

W, source, out = sys.argv[1].rstrip("/") + "/", sys.argv[2], Path(sys.argv[3])
sys.path.insert(0, W + "assisted/scripts")
from strict import strict

kept = {}
if out.exists():
    for line in out.read_text().splitlines():
        row = json.loads(line)
        kept[row["repository"]] = row
rows = [json.loads(line) for line in open(source)]
for row in sorted(rows, key=lambda item: (item["repository"], item["commit"])):
    if strict(row["message"]) and not row["fork"]:
        kept.setdefault(row["repository"], row)
with out.open("w") as handle:
    for repository in sorted(kept):
        handle.write(json.dumps(kept[repository], ensure_ascii=False) + "\n")
print("one strict commit per repository:", len(kept))
