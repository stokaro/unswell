#!/usr/bin/env python3
"""Count how often the blind panel picks the post, from the answers of the judges to one set of pairs.

usage: panel.py <label> <pairs jsonl> <judge answers jsonl>...

A post is a unit with its three pairs. The panel picks it when the votes for it outnumber the votes for the human
text; one judge alone picks it when that judge's votes do. Prints counts only.
"""
import collections, json, sys

label, pairs_path, answer_paths = sys.argv[1], sys.argv[2], sys.argv[3:]
pairs = {r["id"]: r for r in map(json.loads, open(pairs_path)) if r.get("comparison") == "set"}
units = sorted({p["unit"] for p in pairs.values()})
together = collections.defaultdict(collections.Counter)
alone = []
answered = []
for path in answer_paths:
    answers = {r["id"]: r for r in map(json.loads, open(path))}
    mine = collections.defaultdict(collections.Counter)
    for pair_id, pair in pairs.items():
        if pair_id in answers:
            mine[pair["unit"]][answers[pair_id]["picked"]] += 1
            together[pair["unit"]][answers[pair_id]["picked"]] += 1
    alone.append(sum(1 for votes in mine.values() if votes["model"] > votes["human"]))
    answered.append(sum(1 for pair_id in pairs if pair_id in answers))
picked = sum(1 for votes in together.values() if votes["model"] > votes["human"])
print(f"{label}: posts {len(units)} | panel picks the post {picked} ({100 * picked / len(units):.0f}%) | each judge alone {alone} | "
      f"pairs answered {answered} of {len(pairs)}")
