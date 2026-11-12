#!/usr/bin/env python3
"""Compare judge models with the LAMP writers, with Beemo experts, and read their pair judgments.

usage: judge_eval.py <human directory> <out json>
"""
import collections, itertools, json, math, sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from common import read_jsonl

ROOT, OUT = Path(sys.argv[1]), Path(sys.argv[2])
PAIRS = [("Human-edited", "AI-generated"), ("AI-edited", "AI-generated"), ("Human-edited", "AI-edited")]
ADMIT = {"Human-edited over AI-generated": 0.70, "AI-edited over AI-generated": 0.65}
BEEMO = 0.65


def wilson(yes, n, z=1.96):
    if not n:
        return (0.0, 0.0)
    p = yes / n
    centre = (p + z * z / (2 * n)) / (1 + z * z / n)
    half = z * math.sqrt(p * (1 - p) / n + z * z / (4 * n * n)) / (1 + z * z / n)
    return (centre - half, centre + half)


human = collections.defaultdict(list)
for row in read_jsonl(ROOT / "preferences.jsonl"):
    human[row["triplet"]].append(row["order"])
result = {"judges": {}, "writers": {}}
for first, second in PAIRS:
    agree = total = 0
    for orders in human.values():
        votes = [o.index(first) < o.index(second) for o in orders]
        for a, b in itertools.combinations(votes, 2):
            total += 1
            agree += a == b
    result["writers"][f"{first} over {second}"] = agree / total
    print(f"writers: {first} over {second}: one writer agrees with another {agree / total:.1%}")

for directory in sorted((ROOT / "judgments" / "triplets").iterdir()) if (ROOT / "judgments" / "triplets").exists() else []:
    judge = directory.name
    orders = {p.stem: json.loads(p.read_text())["order"] for p in directory.glob("*.json")}
    entry = {"triplets": len(orders), "pairs": {}}
    line = [f"{judge:7} triplets {len(orders):>3}"]
    for first, second in PAIRS:
        with_majority = with_writer = writer_total = prefers = 0
        for triplet, order in orders.items():
            choice = order.index(first) < order.index(second)
            votes = [o.index(first) < o.index(second) for o in human[triplet]]
            prefers += choice
            with_majority += choice == (sum(votes) * 2 > len(votes))
            with_writer += sum(choice == v for v in votes)
            writer_total += len(votes)
        name = f"{first} over {second}"
        entry["pairs"][name] = {"prefers_first": prefers / len(orders), "agrees_with_majority": with_majority / len(orders),
                                "agrees_with_a_writer": with_writer / writer_total}
        line.append(f"{name}: prefers {prefers / len(orders):.1%}, with majority {with_majority / len(orders):.1%}, "
                    f"with a writer {with_writer / writer_total:.1%}")
    result["judges"][judge] = entry
    print(" | ".join(line))

for directory in sorted((ROOT / "judgments" / "beemo").iterdir()) if (ROOT / "judgments" / "beemo").exists() else []:
    answers = [json.loads(p.read_text()) for p in directory.glob("*.json")]
    edited = sum(a["better"] == "edited" for a in answers)
    first = sum(a["position"] == "A" for a in answers)
    low, high = wilson(edited, len(answers))
    result["judges"].setdefault(directory.name, {})["beemo"] = {"pairs": len(answers), "prefers_edited": edited / len(answers),
                                                                "interval": [low, high], "chooses_first_position": first / len(answers)}
    print(f"{directory.name:7} beemo {len(answers):>3}: prefers the expert-edited answer {edited / len(answers):.1%} "
          f"[{low:.1%}, {high:.1%}]; chooses position A {first / len(answers):.1%}")

for judge, entry in result["judges"].items():
    pairs = entry.get("pairs", {})
    checks = [pairs.get(name, {}).get("agrees_with_majority", 0) >= floor for name, floor in ADMIT.items()]
    checks.append(entry.get("beemo", {}).get("prefers_edited", 0) >= BEEMO)
    entry["admitted"] = all(checks)
    print(f"{judge}: admitted {entry['admitted']} (writer-edited or raw, model-edited or raw, Beemo: {checks})")

result["effect"] = {}
for directory in sorted((ROOT / "judgments").glob("pairs-effect-*")):
    name = directory.name.removeprefix("pairs-effect-")
    for judge_dir in sorted(directory.iterdir()):
        answers = [json.loads(p.read_text()) for p in judge_dir.glob("*.json")]
        if not answers:
            continue
        counts = collections.Counter(a["better"] for a in answers)
        labels = sorted(counts)
        second = name.split("-")[-1]
        low, high = wilson(counts[second], len(answers))
        result["effect"].setdefault(name, {})[judge_dir.name] = {"pairs": len(answers), "prefers": dict(counts),
                                                                  "share_second": counts[second] / len(answers), "interval": [low, high]}
        print(f"effect {name:28} judge {judge_dir.name:7} pairs {len(answers):>3}: prefers {second} "
              f"{counts[second] / len(answers):.1%} [{low:.1%}, {high:.1%}] {dict(counts)}")
OUT.write_text(json.dumps(result, indent=1, sort_keys=True) + "\n")
