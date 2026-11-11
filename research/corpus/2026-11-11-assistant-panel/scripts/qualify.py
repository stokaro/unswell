#!/usr/bin/env python3
"""Qualify candidate phrases against the annotated documents.

usage: qualify.py <workspace> <out json> <docs file>...

Half of the repositories (by a hash of the name) select phrases; the other half
checks them. A phrase qualifies when it occurs in at least MIN items of the
selection half and at least FLOOR of those items were flagged by an annotator.
"""
import collections, hashlib, json, re, sys
from pathlib import Path

W, OUT = Path(sys.argv[1]), sys.argv[2]
TOKEN = re.compile(r"[A-Za-z]+(?:[-'][A-Za-z]+)*")
MIN, FLOOR = 5, 0.7
lexicon = json.loads((W / "lexicon-candidates.json").read_text())
phrases = {p.lower(): group for group, items in lexicon.items() for p in items}
maxlen = max(len(TOKEN.findall(p)) for p in phrases)
keys = {" ".join(TOKEN.findall(p)).lower(): p for p in phrases}


def check_half(repo):
    return int(hashlib.sha256(("q" + repo).encode()).hexdigest(), 16) % 2 == 0


stats = {"select": collections.defaultdict(lambda: [0, 0, 0]), "check": collections.defaultdict(lambda: [0, 0, 0])}
totals = {"select": [0, 0, 0], "check": [0, 0, 0]}
examples = collections.defaultdict(list)
for name in sys.argv[3:]:
    for line in open(W / name):
        doc = json.loads(line)
        paths = [W / "ann" / n / f"{doc['doc_id']}.json" for n in ("astra", "claude")]
        if not all(p.exists() for p in paths):
            continue
        astra, claude = [{f["item"]: f for f in json.loads(p.read_text())["flags"]} for p in paths]
        half = "check" if check_half(doc["repository"]) else "select"
        for item in doc["items"]:
            a, c = astra.get(item["id"]), claude.get(item["id"])
            pa = bool(a and a["label"] == "needs_revision")
            pc = bool(c and c["label"] == "needs_revision")
            slot = 0 if pa and pc else 1 if pa or pc else 2
            totals[half][slot] += 1
            tokens = [t.lower() for t in TOKEN.findall(" ".join(item["text"].split()))]
            found = set()
            for n in range(1, maxlen + 1):
                for i in range(len(tokens) - n + 1):
                    gram = " ".join(tokens[i:i + n])
                    if gram in keys:
                        found.add(gram)
            for gram in found:
                stats[half][gram][slot] += 1
                if slot == 2 and len(examples[gram]) < 3:
                    examples[gram].append(" ".join(item["text"].split())[:140])
rows = []
for gram, original in keys.items():
    s = stats["select"].get(gram, [0, 0, 0])
    c = stats["check"].get(gram, [0, 0, 0])
    n = sum(s)
    rows.append({"phrase": original, "group": phrases[original], "select": s, "check": c,
                 "lenient": round((s[0] + s[1]) / n, 3) if n else None, "strict": round(s[0] / n, 3) if n else None,
                 "qualified": bool(n >= MIN and (s[0] + s[1]) / n >= FLOOR), "unflagged_examples": examples[gram]})
qualified = [r for r in rows if r["qualified"]]
json.dump({"min_items": MIN, "floor": FLOOR, "totals": totals, "phrases": rows}, open(W / OUT, "w"), indent=1)
print(f"candidates={len(rows)} seen in selection half={sum(1 for r in rows if sum(r['select']))} qualified={len(qualified)}")
b = sum(r["check"][0] for r in qualified); o = sum(r["check"][1] for r in qualified); n = sum(r["check"][2] for r in qualified)
print(f"check half, qualified phrases: occurrences both/one/none={b}/{o}/{n} "
      f"lenient={100 * (b + o) / max(1, b + o + n):.1f}% strict={100 * b / max(1, b + o + n):.1f}% "
      f"(consensus items in check half={totals['check'][0]})")
print("qualified:", ", ".join(f"{r['phrase']}({sum(r['select'])},{r['lenient']:.2f})" for r in sorted(qualified, key=lambda r: -sum(r['select']))))
rejected = [r for r in rows if sum(r["select"]) >= MIN and not r["qualified"]]
print("\nrejected for low precision:", ", ".join(f"{r['phrase']}({sum(r['select'])},{r['lenient']:.2f})" for r in sorted(rejected, key=lambda r: -sum(r['select']))[:60]))
