#!/usr/bin/env python3
"""Compare how often two runs of a reviewer flag the same sentences, for two answer protocols.

usage: repeat_eval.py <workspace> <system> <new run 1> <new run 2> <old run 1> <old run 2>

Each run directory holds docs.jsonl and runs/<system>/<document id>.json. Only
the texts before a commit that all four runs hold are used. The repeat rate is
twice the sentences flagged in both runs over the sum of the sentences flagged
in each. The interval of the difference is a bootstrap over documents.
"""
import bisect, json, os, random, sys

W = sys.argv[1].rstrip("/") + "/"
sys.path.insert(0, W + "human/scripts")
from common import sentences

system, names = sys.argv[2], sys.argv[3:7]


def flagged(name, doc):
    target = f"{W}{name}/runs/{system}/{doc['id']}.json"
    if not os.path.exists(target):
        return None
    spans = sorted(tuple(s) for s in sentences(doc["text"]))
    starts = [s[0] for s in spans]
    result = set()
    for begin, _ in json.load(open(target))["spans"]:
        index = bisect.bisect_right(starts, begin) - 1
        result.add(index if index >= 0 and begin < spans[index][1] else ("outside", begin))
    return result


rows = []
for line in open(W + names[0] + "/docs.jsonl"):
    doc = json.loads(line)
    if doc["side"] != "before":
        continue
    sets = [flagged(name, doc) for name in names]
    if any(s is None for s in sets):
        continue
    rows.append((len(sets[0] & sets[1]), len(sets[0]) + len(sets[1]), len(sets[2] & sets[3]), len(sets[2]) + len(sets[3]),
                 len(sets[0]), len(sets[1]), len(sets[2]), len(sets[3])))


def rates(sample):
    a, b, c, d = (sum(row[i] for row in sample) for i in range(4))
    return 2 * a / b if b else 0.0, 2 * c / d if d else 0.0


new, old = rates(rows)
rng, draws = random.Random(20261113), []
for _ in range(2000):
    sample = [rows[rng.randrange(len(rows))] for _ in rows]
    x, y = rates(sample)
    draws.append(x - y)
draws.sort()
totals = [sum(row[i] for row in rows) for i in range(8)]
print(f"{system}: {len(rows)} documents")
print(f"  a verdict for every item: sentences flagged {totals[4]} and {totals[5]}, in both {totals[0]}, repeat rate {new:.1%}")
print(f"  only what needs revision: sentences flagged {totals[6]} and {totals[7]}, in both {totals[2]}, repeat rate {old:.1%}")
print(f"  difference {100 * (new - old):+.1f} points ({100 * draws[49]:+.1f} to {100 * draws[1949]:+.1f})")
