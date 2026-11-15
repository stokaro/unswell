#!/usr/bin/env python3
"""Build the pairs that the judges compare: a model's post against a human post.

usage: pairs.py <human jsonl> <written jsonl>... --set <set jsonl> --split test|dev --out <pairs jsonl>

Comparisons: "plain" and "guided" pair a written post with the human post it was outlined from; "cross" pairs a
plain post with a human post of another source; "set" pairs each post of the style-guide set with three human
posts. In every row "first" is the model's text and "second" the human text; the judge script hides the order.
"""
import hashlib, json, sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
import prepare as lib

args = sys.argv[1:]
take = lambda name: args[args.index(name) + 1]
split, out, styled = take("--split"), Path(take("--out")), take("--set")
files = [a for a in args[: args.index("--set")]]
order = lambda salt, key: hashlib.sha256((salt + key).encode()).hexdigest()
human = [r for r in map(json.loads, open(files[0])) if r["split"] == split]
written = {}
for name in files[1:]:
    for r in map(json.loads, open(name)):
        if not r.get("failed"):
            written[(r["id"], r["family"], r["variant"])] = r
families = sorted({key[1] for key in written})
opening = lambda text: lib.excerpt(text, 450)
pairs = []
for h in human:
    for family in families:
        for variant in ("plain", "guided"):
            w = written.get((h["id"], family, variant))
            if w:
                pairs.append({"id": f"{variant}:{family}:{h['id']}", "comparison": variant, "family": family,
                              "guide": w.get("guide"), "unit": h["id"] + ":" + family,
                              "first": opening(w["text"]), "second": opening(h["text"])})
# A plain post against a human post of another source: the writer and the partner come from hashes.
ring = sorted(human, key=lambda r: order("cross:", r["id"]))
for position, h in enumerate(ring):
    family = families[int(order("writer:", h["id"])[:8], 16) % len(families)] if families else None
    w = written.get((h["id"], family, "plain"))
    partner = next((ring[(position + step) % len(ring)] for step in range(1, len(ring)) if ring[(position + step) % len(ring)]["repo"] != h["repo"]), None)
    if w and partner:
        pairs.append({"id": f"cross:{family}:{h['id']}:{partner['id']}", "comparison": "cross", "family": family,
                      "guide": None, "unit": h["id"], "first": opening(w["text"]), "second": opening(partner["text"])})
for p in (r for r in map(json.loads, open(styled)) if r["split"] == split):
    for partner in sorted(human, key=lambda r: order("set:" + p["id"] + ":", r["id"]))[:3]:
        pairs.append({"id": f"set:{p['id']}:{partner['id']}", "comparison": "set", "family": None, "guide": None,
                      "unit": p["id"], "first": opening(p["text"]), "second": opening(partner["text"])})
out.write_text("".join(json.dumps(row, ensure_ascii=False) + "\n" for row in pairs))
counts = {}
for row in pairs:
    counts[row["comparison"]] = counts.get(row["comparison"], 0) + 1
print(counts)
