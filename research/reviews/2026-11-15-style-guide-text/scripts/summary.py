#!/usr/bin/env python3
"""Count how often the panel picks the model's text in each comparison.

usage: summary.py <pairs jsonl> <label>=<judge jsonl>...

A judge's label starts with its family: qwen27, qwen35, gpt. A unit is one pair, or one post of the style-guide
set with its three pairs. The panel picks the model's text when the votes for it outnumber the votes for the
human text. The rate is over all units.
"""
import collections, json, math, sys


def wilson(k, n):
    if n == 0:
        return (0.0, 0.0)
    p, z = k / n, 1.96
    center, half = (p + z * z / (2 * n)) / (1 + z * z / n), z * math.sqrt(p * (1 - p) / n + z * z / (4 * n * n)) / (1 + z * z / n)
    return (center - half, center + half)


pairs = {r["id"]: r for r in map(json.loads, open(sys.argv[1]))}
votes = collections.defaultdict(dict)
for spec in sys.argv[2:]:
    label, path = spec.split("=", 1)
    for r in map(json.loads, open(path)):
        if r["id"] in pairs:
            votes[r["id"]][label] = {"model": 1, "human": -1, "same": 0}[r["picked"]]
family_of = lambda label: "qwen" if label.startswith("qwen") else "gpt"
labels = sorted({label for by in votes.values() for label in by})


def line(name, rows, keep=lambda label, pair: True):
    """Sum the kept votes of each unit and report the share of units where the model's text wins."""
    units, missing = collections.defaultdict(int), 0
    for pair in rows:
        cast = {label: v for label, v in votes.get(pair["id"], {}).items() if keep(label, pair)}
        expected = [label for label in labels if keep(label, pair)]
        if len(cast) < len(expected):
            missing += 1
        units[pair["unit"]] += sum(cast.values())
    n = len(units)
    model, human = sum(v > 0 for v in units.values()), sum(v < 0 for v in units.values())
    low, high = wilson(model, model + human)
    decisive = f"{model}/{model + human} decisive = {model / (model + human):.1%} ({low:.1%}-{high:.1%})" if model + human else "no decisive unit"
    print(f"  {name:34s} model {model:3d}  human {human:3d}  tie {n - model - human:3d}  of {n:3d}  rate {model / n:.1%}   {decisive}"
          + (f"   [{missing} pairs lack a vote]" if missing else ""))


for comparison in ("plain", "cross", "guided", "set"):
    rows = [p for p in pairs.values() if p["comparison"] == comparison]
    if not rows:
        continue
    print(comparison, f"({len(rows)} pairs)")
    line("panel", rows)
    for label in labels:
        line("  judge " + label, rows, lambda l, p, label=label: l == label)
    families = sorted({p["family"] for p in rows if p["family"]})
    for family in families:
        line("writer " + family, [p for p in rows if p["family"] == family])
        line("  judges of the other family", [p for p in rows if p["family"] == family], lambda l, p: family_of(l) != p["family"])
    for guide in sorted({p["guide"] for p in rows if p["guide"]}):
        line("style guide " + guide, [p for p in rows if p["guide"] == guide])
