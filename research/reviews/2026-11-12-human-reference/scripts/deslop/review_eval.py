#!/usr/bin/env python3
"""Measure a system on sampled de-slop pairs: do its flags fall on what the commit changed, and are there fewer after?

usage: review_eval.py <review directory> <system> [--unconfirmed]

The directory holds docs.jsonl (two documents per pair, before and after) and
runs/<system>/<document id>.json with byte spans. On the text before the
commit, a word is flagged when a span overlaps it and changed when the commit
removed or replaced it. Chance is the share of changed words in the same
documents, weighted by the flagged words of each. The interval is a bootstrap
over documents. Across pairs, findings per 1,000 words are compared.
"""
import json, math, random, re, sys
from pathlib import Path

root, system = Path(sys.argv[1]), sys.argv[2]
extra = "--unconfirmed" in sys.argv
WORD = re.compile(r"[A-Za-z][A-Za-z'’-]*")


def wilson(yes, total, z=1.96):
    if not total:
        return 0.0, 0.0
    share = yes / total
    centre = (share + z * z / (2 * total)) / (1 + z * z / total)
    half = z * math.sqrt(share * (1 - share) / total + z * z / (4 * total * total)) / (1 + z * z / total)
    return centre - half, centre + half


def overlaps(begin, end, spans):
    return any(a < end and begin < b for a, b in spans)


pairs, rows, missing = {}, [], 0
for line in open(root / "docs.jsonl"):
    doc = json.loads(line)
    target = root / "runs" / system / f"{doc['id']}.json"
    if not target.exists():
        missing += 1
        continue
    run = json.loads(target.read_text())
    spans = [tuple(s) for s in run["spans"]] + ([tuple(s) for s in run.get("unconfirmed", [])] if extra else [])
    text = doc["text"]
    words = [(len(text[:m.start()].encode()), len(text[:m.end()].encode())) for m in WORD.finditer(text)]
    pairs.setdefault(doc["pair"], {})[doc["side"]] = (len(spans), len(words))
    if doc["side"] != "before" or not words:
        continue
    reference = [tuple(r["span"]) for r in doc["reference"]]
    flagged = [overlaps(b, e, spans) for b, e in words]
    changed = [overlaps(b, e, reference) for b, e in words]
    rows.append((sum(flagged), sum(f and c for f, c in zip(flagged, changed)), sum(flagged) * sum(changed) / len(words), len(words), len(spans)))
print(f"{system}: {len(rows)} documents before a commit, {missing} documents without a run")
flagged, hits, chance, words, spans = (sum(column) for column in zip(*rows))
rng, draws = random.Random(20261113), []
for _ in range(2000):
    a = b = c = 0.0
    for _ in rows:
        row = rows[rng.randrange(len(rows))]
        a += row[0]; b += row[1]; c += row[2]
    if a:
        draws.append((b - c) / a)
draws.sort()
print(f"  documents with a flag {sum(1 for r in rows if r[4])}, findings {spans} ({1000 * spans / words:.2f} per 1,000 words), "
      f"words flagged {flagged} of {words} ({flagged / words:.1%})")
if flagged:
    print(f"  flagged words that the commit changed {hits / flagged:.1%}, chance {chance / flagged:.1%}, "
          f"difference {100 * (hits - chance) / flagged:+.1f} points ({100 * draws[int(0.025 * len(draws))]:+.1f} to {100 * draws[int(0.975 * len(draws)) - 1]:+.1f})")
lower = higher = same = before = after = 0
for sides in pairs.values():
    if len(sides) < 2:
        continue
    (b, wb), (a, wa) = sides["before"], sides["after"]
    if not a and not b:
        continue
    before += b; after += a
    left, right = a * wb, b * wa
    lower += left < right; higher += left > right; same += left == right
low, high = wilson(lower, lower + higher)
print(f"  pairs where it fires {lower + higher + same}: lower after {lower}, higher {higher}, same {same}; findings {before} -> {after}; "
      f"lower among pairs that differ {lower / max(1, lower + higher):.1%} ({low:.1%} to {high:.1%})")
