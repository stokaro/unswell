#!/usr/bin/env python3
"""Read the readers' answers to the blind comparison against the key.

usage: study_eval.py <human directory> <out json>

An answer is "left", "right", or "same". It prefers the revision when it names
the side on which the revised paragraph was shown.
"""
import collections, json, math, sys
from pathlib import Path

ROOT, OUT = Path(sys.argv[1]), Path(sys.argv[2])
key = json.loads((ROOT / "study/key.json").read_text())


def wilson(yes, n, z=1.96):
    if not n:
        return [0.0, 0.0]
    p = yes / n
    centre = (p + z * z / (2 * n)) / (1 + z * z / n)
    half = z * math.sqrt(p * (1 - p) / n + z * z / (4 * n * n)) / (1 + z * z / n)
    return [centre - half, centre + half]


readers = {}
for path in sorted((ROOT / "study/answers/answers").glob("*.json")):
    document = json.loads(path.read_text())
    readers[path.stem] = (document.get("data") or document).get("choices") or {}
result = {"readers": len(readers), "arms": {}, "cells": {}, "per_reader": {}}
totals = collections.defaultdict(collections.Counter)
cells = collections.defaultdict(collections.Counter)
for reader, choices in readers.items():
    own = collections.defaultdict(collections.Counter)
    for pair, choice in choices.items():
        entry = key.get(pair)
        if entry is None or choice not in ("left", "right", "same"):
            continue
        outcome = "same" if choice == "same" else "revised" if choice == entry["revised"] else "raw"
        for counter in (totals[entry["arm"]], cells[f"{entry['reviewer']}/{entry['arm']}"], own[entry["arm"]]):
            counter[outcome] += 1
    result["per_reader"][reader[:6]] = {arm: dict(counter) for arm, counter in own.items()}


def summary(counter):
    decisive = counter["revised"] + counter["raw"]
    return {"answers": decisive + counter["same"], "revised": counter["revised"], "raw": counter["raw"], "same": counter["same"],
            "decisive": decisive, "share_revised": counter["revised"] / decisive if decisive else None,
            "interval": wilson(counter["revised"], decisive)}


for arm, counter in sorted(totals.items()):
    result["arms"][arm] = summary(counter)
for cell, counter in sorted(cells.items()):
    result["cells"][cell] = summary(counter)
for arm, entry in result["arms"].items():
    share = f"{entry['share_revised']:.1%}" if entry["decisive"] else "n/a"
    print(f"{arm:9} answers {entry['answers']:>2}: revised {entry['revised']}, original {entry['raw']}, no difference {entry['same']}; "
          f"of decisive {share} [{entry['interval'][0]:.1%}, {entry['interval'][1]:.1%}]")
for cell, entry in result["cells"].items():
    print(f"   {cell:16} revised {entry['revised']}, original {entry['raw']}, no difference {entry['same']}")
reviewed, control = result["arms"].get("reviewed"), result["arms"].get("control")
if reviewed and control and reviewed["decisive"] and control["decisive"]:
    lead = reviewed["share_revised"] - control["share_revised"]
    result["decision"] = {"enough_answers": reviewed["decisive"] >= 15,
                          "revised_preferred": reviewed["share_revised"] >= 0.70 and reviewed["interval"][0] > 0.50,
                          "lead_over_control": lead, "ahead_of_control": lead >= 0.15}
    print("decision:", result["decision"])
OUT.write_text(json.dumps(result, indent=1, sort_keys=True) + "\n")
