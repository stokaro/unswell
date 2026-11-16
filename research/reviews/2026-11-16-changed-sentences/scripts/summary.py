#!/usr/bin/env python3
"""Summarize a run of the check with the sentence comparison.

usage: summary.py <checks jsonl>...

For damaged copies: on how many the check reports the damage, by the statements, by the sentences, and by either, and
how many it accepts. For ordinary revisions: the statements kept, the suspects and what the judge said of them, and
the revisions that a sentence alone rejects.
"""
import json, re, sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from words import content_words

words = lambda text: set(re.findall(r"[a-z0-9]+", text.lower()))
# With --words <audit jsonl>..., the rule of the final build is applied to the kept answers: a missing sentence is
# cleared only when the quoted words carry one of its own words.
own = {}
audits = sys.argv[sys.argv.index("--words") + 1:] if "--words" in sys.argv else []
for audit in audits:
    for row in map(json.loads, open(audit)):
        for reason, text, _, sentence_words in row["suspects"]:
            own[(row["set"], row["id"], text)] = set(sentence_words)
paths = sys.argv[1:sys.argv.index("--words")] if "--words" in sys.argv else sys.argv[1:]


def final_rule(row):
    changed = 0
    for sentence in row.get("sentences", []):
        key = (row["set"], row["id"], sentence[2])
        if sentence[0] == "missing" and sentence[1] == "kept" and key in own and not (own[key] & content_words(sentence[3])):
            sentence[1] = "unquoted"
            changed += 1
    return changed


for path in paths:
    rows = [json.loads(line) for line in open(path)]
    if own:
        print(f"   the rule of the final build turns {sum(final_rule(r) for r in rows if 'error' not in r)} cleared sentences into lost ones")
    done = [r for r in rows if "error" not in r]
    print(f"== {path}: {len(rows)} pairs, {len(rows) - len(done)} without a verdict")
    by = {}
    for r in done:
        entry = by.setdefault(r["set"], {"pairs": 0, "accepted": 0, "sound better": 0, "statements": 0, "kept": 0, "unquoted": 0, "under 98%": 0,
                                         "suspects": 0, "cleared": 0, "with a lost sentence": 0, "rejected by sentences alone": 0,
                                         "damage in suspects": 0, "damage reported by sentences": 0, "by statements": 0, "by either": 0})
        facts, sentences = r["facts"], r.get("sentences", [])
        not_kept = sum(f[0] != "kept" for f in facts)
        lost = [s for s in sentences if s[1] != "kept"]
        entry["pairs"] += 1
        entry["accepted"] += r["accepted"]
        entry["sound better"] += r["sound"] == "better"
        entry["statements"] += len(facts)
        entry["kept"] += len(facts) - not_kept
        entry["unquoted"] += sum(f[0] == "unquoted" for f in facts)
        entry["under 98%"] += (len(facts) - not_kept) < 0.98 * len(facts)
        entry["suspects"] += len(sentences)
        entry["cleared"] += len(sentences) - len(lost)
        entry["with a lost sentence"] += bool(lost)
        entry["rejected by sentences alone"] += bool(lost) and (len(facts) - not_kept) >= 0.98 * len(facts)
        if "damage" in r:
            target = words(r["damage"])
            shared = lambda s: len(target & words(s[2]))
            hit = lambda group: any(shared(s) >= 3 and shared(s) >= 0.8 * min(len(target), len(words(s[2]))) for s in group)
            entry["damage in suspects"] += hit(sentences)
            entry["damage reported by sentences"] += hit(lost)
            entry["by statements"] += not_kept > 0
            entry["by either"] += not_kept > 0 or hit(lost)
    for name, entry in sorted(by.items()):
        share = f"{100 * entry['kept'] / entry['statements']:.1f}%" if entry["statements"] else "n/a"
        print(" ", name, {k: v for k, v in entry.items() if v or k in ("accepted",)}, "kept share", share)
