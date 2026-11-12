#!/usr/bin/env python3
"""How the LAMP writers ranked the three versions, and how far they agree pair by pair.

usage: preference_humans.py <preferences jsonl> <out json>
"""
import collections, itertools, json, sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from common import read_jsonl

rows = read_jsonl(sys.argv[1])
PAIRS = [("Human-edited", "AI-generated"), ("AI-edited", "AI-generated"), ("Human-edited", "AI-edited")]
triplets = collections.defaultdict(list)
for row in rows:
    triplets[row["triplet"]].append(row)
result = {"judgments": len(rows), "triplets": len(triplets), "pairs": {}}
for first, second in PAIRS:
    wins = sum(row["order"].index(first) < row["order"].index(second) for row in rows)
    agree = total = unanimous = majority_first = decided = 0
    for group in triplets.values():
        votes = [row["order"].index(first) < row["order"].index(second) for row in group]
        for a, b in itertools.combinations(votes, 2):
            total += 1
            agree += a == b
        if len(votes) == 3:
            decided += 1
            unanimous += len(set(votes)) == 1
            majority_first += sum(votes) >= 2
    result["pairs"][f"{first} over {second}"] = {
        "judgments_preferring_first": wins / len(rows), "writer_pairs_agreeing": agree / total,
        "triplets_with_three_writers": decided, "unanimous": unanimous / decided, "majority_prefers_first": majority_first / decided}
    print(f"{first} over {second}: {wins / len(rows):.1%} of judgments; two writers agree {agree / total:.1%}; "
          f"majority prefers the first in {majority_first / decided:.1%} of {decided} triplets; unanimous {unanimous / decided:.1%}")
Path(sys.argv[2]).write_text(json.dumps(result, indent=1, sort_keys=True) + "\n")
