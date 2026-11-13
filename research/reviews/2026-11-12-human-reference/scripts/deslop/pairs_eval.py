#!/usr/bin/env python3
"""For each rule and mark: does it fire more on the text before a de-slop commit than on the text after?

usage: pairs_eval.py <scores jsonl> <files jsonl> <development|test> [--web] [--loose]

A pair is one file before and after a commit whose message states that it
removes AI-sounding wording. For a feature, pairs where it fires on neither
side are left out. The table gives the share of the remaining pairs in which
the count per 1,000 words is lower after the commit, the share in which it is
higher, and the total counts before and after. A feature that tracks what
people remove has a high first share and a count that falls.
"""
import collections, hashlib, json, math, sys

sys.path.insert(0, __file__.rsplit("/", 1)[0])
from strict import strict

scores, files, part = sys.argv[1], sys.argv[2], sys.argv[3]
web_only, loose = "--web" in sys.argv, "--loose" in sys.argv
messages = {}
for line in open(files):
    doc = json.loads(line)
    if not doc.get("skip"):
        messages[doc["id"]] = doc["message"]


def development(repository):
    return int(hashlib.sha256(repository.encode()).hexdigest(), 16) % 10 < 6


def wilson(yes, total, z=1.96):
    if not total:
        return 0.0, 0.0
    share = yes / total
    centre = (share + z * z / (2 * total)) / (1 + z * z / total)
    half = z * math.sqrt(share * (1 - share) / total + z * z / (4 * total * total)) / (1 + z * z / total)
    return centre - half, centre + half


pairs = []
for line in open(scores):
    row = json.loads(line)
    if row.get("skip") or row["assistant_trailer"] or development(row["repository"]) != (part == "development"):
        continue
    if web_only and not row["web_edit"]:
        continue
    if not loose and not strict(messages.get(row["id"], "")):
        continue
    if row["before"]["words"] < 50 or row["after"]["words"] < 20:
        continue
    pairs.append(row)
print(f"{part}{' web edits' if web_only else ''}{' any message' if loose else ''}: {len(pairs)} pairs from {len({p['repository'] for p in pairs})} repositories")
NEW = ("format.decorative-emoji", "format.em-dash-density")
for pair in pairs:
    for side in ("before", "after"):
        counts = pair[side]["counts"]
        rules = {name: count for name, count in counts.items() if not name.startswith(("mark.", "set."))}
        counts["set.defaults of alpha.9"] = sum(count for name, count in rules.items() if name not in NEW)
        counts["set.defaults with the two rules"] = sum(rules.values())
features = collections.Counter()
for pair in pairs:
    for side in ("before", "after"):
        for name, count in pair[side]["counts"].items():
            if count:
                features[name] += 1
table = []
for name in features:
    lower = higher = same = 0
    before_total = after_total = 0
    for pair in pairs:
        b, a = pair["before"]["counts"].get(name, 0), pair["after"]["counts"].get(name, 0)
        if not b and not a:
            continue
        before_total += b
        after_total += a
        rate_before = 1000 * b / pair["before"]["words"]
        rate_after = 1000 * a / pair["after"]["words"]
        if rate_after < rate_before * 0.999:
            lower += 1
        elif rate_after > rate_before * 1.001:
            higher += 1
        else:
            same += 1
    active = lower + higher + same
    if active >= 30:
        low, high = wilson(lower, lower + higher)
        table.append((lower / max(1, lower + higher), name, active, lower, higher, same, before_total, after_total, low, high))
print(f"{'feature':38} pairs  lower higher same   before   after  lower among changed (95%)")
for share, name, active, lower, higher, same, before_total, after_total, low, high in sorted(table, reverse=True):
    print(f"{name:38} {active:5}  {lower:5} {higher:6} {same:4}  {before_total:7} {after_total:7}  {share:5.1%} ({low:.1%} to {high:.1%})")
