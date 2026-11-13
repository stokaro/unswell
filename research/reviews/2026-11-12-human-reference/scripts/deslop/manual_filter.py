#!/usr/bin/env python3
"""Keep web-flow commits that look typed by a person: a strict de-slop message of one short line that edits, not adds.

usage: manual_filter.py <found jsonl> <out jsonl>
"""
import json, re, sys

sys.path.insert(0, __file__.rsplit("/", 1)[0])
from strict import strict

NOT_AN_EDIT = re.compile(r"^(add|added|adding|create|created|delete|deleted|rename|renamed|upload|uploaded|initial|init)\b", re.I)
kept = total = 0
seen = set()
with open(sys.argv[2], "w") as out:
    for line in open(sys.argv[1]):
        row = json.loads(line)
        total += 1
        lines = [x for x in row["message"].strip().splitlines() if x.strip()]
        if row["fork"] or row["commit"] in seen or len(lines) != 1 or len(lines[0]) > 72:
            continue
        if not strict(row["message"]) or NOT_AN_EDIT.match(lines[0].strip()):
            continue
        seen.add(row["commit"])
        out.write(line)
        kept += 1
print(f"{kept} of {total} commits kept")
