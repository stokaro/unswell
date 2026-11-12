#!/usr/bin/env python3
"""Collect the LAMP preference study into one file: three versions per paragraph and each writer's ranking.

usage: prepare_preferences.py <LAMP Writing_Alignment directory> <out jsonl>

Every row is one triplet as one writer saw it: the three texts in the order
shown, the label of each (Human-edited, AI-generated, AI-edited), and the
writer's order of preference. Rows that share the paragraph identifier are the
same triplet ranked by different writers.
"""
import ast, collections, json, sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from common import write_jsonl

ROOT, OUT = Path(sys.argv[1]) / "preference_ranking", Path(sys.argv[2])
full = {row["id"] for row in json.loads((ROOT / "LLM_edited_full.json").read_text())}
oracle = {row["id"] for row in json.loads((ROOT / "LLM_edited_oracle.json").read_text())}
rows, stats = [], collections.Counter()
for directory in sorted(path for path in ROOT.iterdir() if path.is_dir()):
    batch = directory.name
    for ranking_file in sorted(directory.glob("*_rankings.json")):
        worker = ranking_file.name.removesuffix("_rankings.json")
        shown = json.loads((ROOT / batch / f"{worker}.json").read_text())
        ranked = {int(entry["instruction_id"]): entry
                  for entry in json.loads((ROOT / batch / f"{worker}_rankings.json").read_text())}
        for entry in shown:
            ranking = ranked.get(int(entry["instruction_id"]))
            if ranking is None:
                stats["unranked"] += 1
                continue
            mapping = entry["mapping"]
            if isinstance(mapping, str):
                mapping = ast.literal_eval(mapping)
            order = [mapping[ranking[key]] for key in ("rank_1", "rank_2", "rank_3")]
            if sorted(order) != ["AI-edited", "AI-generated", "Human-edited"]:
                stats["incomplete ranking"] += 1
                continue
            keys = sorted(mapping, key=lambda k: int("".join(ch for ch in str(k) if ch.isdigit()) or 0))
            texts = {mapping[key]: entry["responses"][index] for index, key in enumerate(keys)}
            setting = "full" if entry["id"] in full else "oracle" if entry["id"] in oracle else "unknown"
            rows.append({"triplet": f"{batch}-{entry['instruction_id']}", "id": entry["id"], "batch": batch, "writer": worker,
                         "setting": setting, "texts": texts, "order": order})
            stats["rows"] += 1
            stats["setting " + setting] += 1
write_jsonl(OUT, rows)
print(dict(stats))
triplets = collections.defaultdict(list)
for row in rows:
    triplets[row["triplet"]].append(row)
print("triplets", len(triplets), "writers per triplet", collections.Counter(len(v) for v in triplets.values()))
first = collections.Counter(row["order"][0] for row in rows)
last = collections.Counter(row["order"][2] for row in rows)
print("ranked first", {k: f"{v / len(rows):.1%}" for k, v in first.items()})
print("ranked last", {k: f"{v / len(rows):.1%}" for k, v in last.items()})
sample = rows[0]
print(sample["triplet"], sample["writer"], sample["order"], {k: v[:60] for k, v in sample["texts"].items()})
same = sum(len({json.dumps(r["texts"], sort_keys=True) for r in group}) == 1 for group in triplets.values())
print("triplets whose texts are identical across writers", same)
