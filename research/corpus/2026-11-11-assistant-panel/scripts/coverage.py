#!/usr/bin/env python3
"""Coverage of the annotators' labels by an engine report.

usage: coverage.py <workspace> <report json> <folder=docs file>...
Reports how many items flagged by both annotators (and by at least one) are touched by a finding,
overall and without the long-sentence rule.
"""
import collections, json, sys
from pathlib import Path
W = Path(sys.argv[1])
report = json.loads(Path(sys.argv[2]).read_text())
items = {}
for pair in sys.argv[3:]:
    folder, name = pair.split("=")
    for line in open(W / name):
        doc = json.loads(line)
        paths = [W / "ann" / n / f"{doc['doc_id']}.json" for n in ("astra", "claude")]
        if not all(p.exists() for p in paths):
            continue
        astra, claude = [{f["item"]: f for f in json.loads(p.read_text())["flags"]} for p in paths]
        rows = []
        for item in doc["items"]:
            a, c = astra.get(item["id"]), claude.get(item["id"])
            pa = bool(a and a["label"] == "needs_revision")
            pc = bool(c and c["label"] == "needs_revision")
            rows.append([item["start"], item["end"], "both" if pa and pc else "one" if pa or pc else "none", set()])
        items[f"{folder}/{doc['doc_id']}.md"] = rows
for finding in report["findings"]:
    rows = items.get(finding["primary"]["path"])
    if rows is None:
        continue
    spans = finding["primary"].get("segments") or [finding["primary"]["span"]]
    for row in rows:
        if any(s["start"] < row[1] and row[0] < s["end"] for s in spans):
            row[3].add(finding["rule_id"])
flat = [row for rows in items.values() for row in rows]
words = sum(d["prose_words"] for d in report["documents"])
print(f"documents={len(items)} items={len(flat)} findings={len(report['findings'])} per 1000 words={1000 * len(report['findings']) / words:.2f}")
for label in ("both", "one"):
    group = [row for row in flat if row[2] == label]
    hit = sum(1 for row in group if row[3])
    hit_other = sum(1 for row in group if row[3] - {"syntax.long-sentence"})
    print(f"items flagged by {label:4s}: {len(group):5d}; touched by a finding {hit} ({100 * hit / max(1, len(group)):.1f}%); "
          f"without the long-sentence rule {hit_other} ({100 * hit_other / max(1, len(group)):.1f}%)")
touched = [row for row in flat if row[3] - {"syntax.long-sentence"}]
by = collections.Counter(row[2] for row in touched)
print(f"items touched by other rules: {len(touched)}; flagged by both {by['both']}, by one {by['one']}, by none {by['none']}")
