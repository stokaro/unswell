#!/usr/bin/env python3
"""Reviewed acceptance inside rules: length bands and table rows.

usage: round3_evidence.py <workspace> <out json>

Reads the reviewed long-sentence findings on real pages and splits them by
sentence length and by whether the source line is a table row.
"""
import collections, json, sys
from pathlib import Path

W, OUT = Path(sys.argv[1]), Path(sys.argv[2])
REPORTS = ["report-broad-dev.json", "report-broad-dev2.json", "report-final-new.json", "report-real3-r2.json"]
ROUNDS = ["rev", "rev-final", "rev-real3"]
BANDS = [(36, 39), (40, 44), (45, 49), (50, 54), (55, 59), (60, 69), (70, 10**6)]

findings = {}
for name in REPORTS:
    for f in json.loads((W / "engine" / name).read_text())["findings"]:
        if f["rule_id"] == "syntax.long-sentence" and not f["primary"]["path"].startswith("v0/"):
            findings[f["id"]] = f
votes = collections.defaultdict(dict)
for folder in ROUNDS:
    for reviewer in ("claude", "astra", "qwen"):
        directory = W / folder / reviewer
        if not directory.exists():
            continue
        for path in directory.glob("*.json"):
            for g in json.loads(path.read_text())["findings"]:
                if g["id"] in findings and g.get("justified"):
                    votes[g["id"]][reviewer] = g["justified"]
sources = {}
bands = collections.defaultdict(collections.Counter)
rows = collections.defaultdict(collections.Counter)
for identifier, vote in votes.items():
    first, second = vote.get("claude"), vote.get("astra") or vote.get("qwen")
    if not first or not second:
        continue
    f = findings[identifier]
    accepted = first == "yes" and second == "yes"
    length = int({m["name"]: m["value"] for m in f["evidence"]["metrics"]}["length"])
    for low, high in BANDS:
        if low <= length <= high:
            bands[low]["reviewed"] += 1
            bands[low]["both_accept"] += accepted
    path = f["primary"]["path"]
    source = sources.setdefault(path, (W / "engine" / path).read_bytes())
    start = f["primary"]["span"]["start"]
    line_start = source.rfind(b"\n", 0, start) + 1
    line_end = source.find(b"\n", start)
    line = source[line_start:line_end if line_end >= 0 else len(source)]
    kind = "table row" if line.lstrip().startswith(b"|") or line.count(b"|") >= 2 else "other"
    rows[kind]["reviewed"] += 1
    rows[kind]["both_accept"] += accepted
result = {"bands": [{"low": low, "high": None if high > 1000 else high, **bands[low]} for low, high in BANDS],
          "rows": {kind: dict(c) for kind, c in rows.items()}}
OUT.write_text(json.dumps(result, indent=1, sort_keys=True) + "\n")
print(json.dumps(result))
