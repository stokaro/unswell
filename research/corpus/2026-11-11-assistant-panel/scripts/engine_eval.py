#!/usr/bin/env python3
"""Compare the engine's findings with the annotators' labels, per rule."""
import collections, json, sys
from pathlib import Path
W = Path(sys.argv[1])
items = {}
for name, folder in (("docs.jsonl", "v0"), ("docs-real.jsonl", "real")):
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
            rows.append((item["start"], item["end"], "both" if pa and pc else "one" if pa or pc else "none"))
        items[f"{folder}/{doc['doc_id']}.md"] = rows
for profile in sys.argv[2:]:
    report = json.loads((W / "engine" / f"report-{profile}.json").read_text())
    per_rule = collections.defaultdict(collections.Counter)
    hit = set()
    for finding in report["findings"]:
        path = finding["primary"]["path"]
        if path not in items:
            continue
        spans = finding["primary"].get("segments") or [finding["primary"]["span"]]
        touched = [(k, row) for k, row in enumerate(items[path])
                   if any(s["start"] < row[1] and row[0] < s["end"] for s in spans)]
        if not touched:
            per_rule[finding["rule_id"]]["outside items"] += 1
            continue
        best = "both" if any(r[2] == "both" for _, r in touched) else "one" if any(r[2] == "one" for _, r in touched) else "none"
        per_rule[finding["rule_id"]][best] += 1
        for k, row in touched:
            if row[2] == "both":
                hit.add((path, k))
    total = sum(row[2] == "both" for rows in items.values() for row in rows)
    print(f"\nprofile={profile}: findings={len(report['findings'])} documents={len(items)} "
          f"consensus-positive items={total}, touched by any finding={len(hit)} ({100 * len(hit) / total:.1f}%)")
    print(f"{'rule':44s} {'total':>6s} {'both':>6s} {'one':>6s} {'none':>6s} {'other':>6s}  both+one share")
    for rule, c in sorted(per_rule.items(), key=lambda kv: -sum(kv[1].values())):
        n = sum(c.values())
        print(f"{rule:44s} {n:6d} {c['both']:6d} {c['one']:6d} {c['none']:6d} {c['outside items']:6d}  {100 * (c['both'] + c['one']) / n:5.1f}%")
