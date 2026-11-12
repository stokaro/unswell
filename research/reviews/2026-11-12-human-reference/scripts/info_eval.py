#!/usr/bin/env python3
"""Validate the model panel against the reader on the first comparison, and read its answers on the second.

usage: info_eval.py <human directory> <out json>
"""
import collections, json, math, sys
from pathlib import Path

ROOT, OUT = Path(sys.argv[1]), Path(sys.argv[2])
JUDGES = ("claude", "astra", "qwen")


def wilson(yes, n, z=1.96):
    if not n:
        return [0.0, 0.0]
    p = yes / n
    centre = (p + z * z / (2 * n)) / (1 + z * z / n)
    half = z * math.sqrt(p * (1 - p) / n + z * z / (4 * n * n)) / (1 + z * z / n)
    return [centre - half, centre + half]


def outcome(choice, entry):
    return "same" if choice == "same" else "revised" if choice == entry["revised"] else "original"


def panel(study):
    """Return {pair: {"judges": {name: outcome}, "panel": outcome}} for the pairs every judge answered."""
    key = json.loads((ROOT / study / "key.json").read_text())
    answers = {}
    for pair, entry in key.items():
        votes = {}
        for judge in JUDGES:
            path = ROOT / study / "judges" / judge / f"{pair}.json"
            if path.exists():
                votes[judge] = outcome(json.loads(path.read_text())["better"], entry)
        if len(votes) < len(JUDGES):
            continue
        counts = collections.Counter(votes.values())
        top, n = counts.most_common(1)[0]
        answers[pair] = {"judges": votes, "panel": top if n >= 2 else "same", "arm": entry["arm"], "reviewer": entry["reviewer"]}
    return key, answers


result = {}
key, answers = panel("study")
reader = {}
for path in sorted((ROOT / "study/answers/answers").glob("*.json")):
    document = json.loads(path.read_text())
    for pair, choice in ((document.get("data") or document).get("choices") or {}).items():
        if pair in key:
            reader[pair] = outcome(choice, key[pair])
table = collections.Counter((reader[pair], answers[pair]["panel"]) for pair in answers if pair in reader)
decisive = [pair for pair in answers if reader.get(pair) in ("revised", "original")]
hits = sum(answers[pair]["panel"] == reader[pair] for pair in decisive)
per_judge = {judge: sum(answers[pair]["judges"][judge] == reader[pair] for pair in decisive) for judge in JUDGES}
result["validation"] = {"pairs": len(answers), "reader_decisive": len(decisive), "panel_same_side": hits, "admitted": hits >= 20 and len(decisive) == 28,
                        "per_judge": per_judge, "table": {f"{r}/{p}": n for (r, p), n in sorted(table.items())}}
print(f"validation: panel names the reader's side on {hits} of {len(decisive)} decisive answers; per judge {per_judge}")
print("   reader/panel:", dict(sorted(result["validation"]["table"].items())))
print("   admitted:", result["validation"]["admitted"])

if (ROOT / "study2/judges").exists():
    key2, answers2 = panel("study2")
    arms = collections.defaultdict(collections.Counter)
    for pair, entry in answers2.items():
        arms[entry["arm"]][entry["panel"]] += 1
    summary = {}
    for arm in ("reviewed", "control"):
        counter = arms[arm]
        n = counter["revised"] + counter["original"]
        summary[arm] = {"revised": counter["revised"], "original": counter["original"], "same": counter["same"], "decisive": n,
                        "interval": wilson(counter["revised"], n)}
        share = f"{counter['revised'] / n:.1%}" if n else "n/a"
        print(f"trim {arm:9}: trimmed preferred {counter['revised']}, original {counter['original']}, no difference {counter['same']}; "
              f"{share} of decisive [{summary[arm]['interval'][0]:.1%}, {summary[arm]['interval'][1]:.1%}]")
    def meets(entry):
        return entry["decisive"] >= 15 and entry["revised"] >= 0.70 * entry["decisive"] and entry["interval"][0] > 0.5
    reviewed, control = summary["reviewed"], summary["control"]
    lead = (reviewed["revised"] / reviewed["decisive"] if reviewed["decisive"] else 0) - (control["revised"] / control["decisive"] if control["decisive"] else 0)
    result["trim"] = {"pairs": len(answers2), "arms": summary, "reviewed_meets_rule": meets(reviewed), "control_meets_rule": meets(control),
                      "lead_over_control": lead, "ahead_of_control": lead >= 0.15,
                      "unanimous": sum(len(set(e["judges"].values())) == 1 for e in answers2.values())}
    print("   decisions:", {k: v for k, v in result["trim"].items() if k not in ("arms",)})
OUT.write_text(json.dumps(result, indent=1, sort_keys=True) + "\n")
