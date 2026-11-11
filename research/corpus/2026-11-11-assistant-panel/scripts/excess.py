#!/usr/bin/env python3
"""Excess vocabulary of generated documents against their source documents.

usage: excess.py <workspace> <out json>

For each word and two-word phrase, compares its rate in the generated versions
with its rate in the source documents of the long-prose study. A candidate must
be over-represented in at least two of the three model families and occur in
generated versions of at least six repositories.
"""
import collections, json, math, re, sys
from pathlib import Path

W, OUT = Path(sys.argv[1]), sys.argv[2]
TOKEN = re.compile(r"[A-Za-z]+(?:[-'][A-Za-z]+)*")
STOP = set("the a an of to in and or for on with by is are be as at this that these those it its from which can will not "
           "you your we our if when then than but so into over under more most such also may must should".split())
counts = collections.defaultdict(collections.Counter)   # family -> gram -> count
totals = collections.Counter()
repos = collections.defaultdict(set)
for line in open(W / "docs.jsonl"):
    doc = json.loads(line)
    family = doc["family"] or "original"
    text = re.sub(r"```.*?```|`[^`\n]*`", " ", doc["text"], flags=re.S)
    tokens = [t.lower() for t in TOKEN.findall(text)]
    totals[family] += len(tokens)
    seen = set()
    for i, token in enumerate(tokens):
        if token not in STOP and len(token) > 2:
            counts[family][token] += 1
            seen.add(token)
        if i + 1 < len(tokens) and not (token in STOP and tokens[i + 1] in STOP):
            gram = token + " " + tokens[i + 1]
            counts[family][gram] += 1
            seen.add(gram)
    if family != "original":
        for gram in seen:
            repos[gram].add(doc["repository"])
known = {p.lower() for g in json.loads((W / "lexicon-candidates.json").read_text()).values() for p in g}
families = [f for f in counts if f != "original"]
rows = []
vocabulary = set().union(*(counts[f] for f in families))
for gram in vocabulary:
    if len(repos[gram]) < 6:
        continue
    original = (counts["original"][gram] + 0.5) / (totals["original"] + 1)
    ratios = []
    total = 0
    for family in families:
        rate = (counts[family][gram] + 0.5) / (totals[family] + 1)
        ratios.append(rate / original)
        total += counts[family][gram]
    over = sum(r >= 4 for r in ratios)
    if total >= 20 and over >= 2:
        rows.append({"phrase": gram, "generated": total, "original": counts["original"][gram],
                     "ratios": [round(r, 1) for r in ratios], "repositories": len(repos[gram]),
                     "known": gram in known})
rows.sort(key=lambda r: -min(sorted(r["ratios"])[1:]) * math.log(r["generated"]))
json.dump({"families": families, "totals": dict(totals), "candidates": rows}, open(W / OUT, "w"), indent=1)
print("words by family:", dict(totals))
print("candidates:", len(rows), "already in the hand-written list:", sum(r["known"] for r in rows))
for row in rows[:70]:
    print(f"  {row['phrase']:26s} generated={row['generated']:4d} original={row['original']:2d} ratios={row['ratios']} repos={row['repositories']}{' *' if row['known'] else ''}")
