#!/usr/bin/env python3
"""Choose phrases and rule defaults from the reviewed development findings.

usage: select_release.py <workspace> <unswell binary> <out yaml>

A phrase is kept when its shrunk both-reviewers-yes rate is at least 0.80. The
shrinkage pulls a phrase with few reviewed findings toward its group's rate.
A default-on rule is switched off when both reviewers reject most of its
findings. A default-off rule is switched on when both accept at least 85%.
"""
import collections, json, subprocess, sys
from pathlib import Path

W, BINARY, OUT = Path(sys.argv[1]), sys.argv[2], sys.argv[3]
PRIOR, KEEP, MIN_RULE = 4.0, 0.80, 15
reviews = {}
for name in ("claude", "astra"):
    for path in (W / "rev" / name).glob("*.json"):
        for finding in json.loads(path.read_text())["findings"]:
            reviews.setdefault(finding["id"], {"rule": finding["rule"], "group": finding["group"],
                                                "phrase": finding["phrase"]})[name] = finding.get("justified")
both = [v for v in reviews.values() if v.get("claude") and v.get("astra")]
yes = lambda v: v["claude"] == "yes" and v["astra"] == "yes"
no = lambda v: v["claude"] == "no" and v["astra"] == "no"
group_rate = {}
for group in {v["group"] for v in both if v["group"]}:
    rows = [v for v in both if v["group"] == group]
    group_rate[group] = sum(map(yes, rows)) / len(rows)
phrases = collections.defaultdict(list)
for v in both:
    if v["phrase"]:
        phrases[(v["group"], v["phrase"])].append(v)
def family(phrase):
    """Pool inflections of one word: leverage, leverages, leveraging; seamless, seamlessly."""
    if " " in phrase:
        return phrase
    stem = phrase
    for suffix in ("ingly", "ly", "ing", "ed", "es", "s", "e"):
        if stem.endswith(suffix) and len(stem) - len(suffix) >= 5:
            stem = stem[:-len(suffix)]
            break
    return stem


lexicon = json.loads((W / "lexicon-candidates.json").read_text())
members = collections.defaultdict(list)
for group, items in lexicon.items():
    for phrase in items:
        members[(group, family(phrase.lower()))].append(phrase.lower())
pooled = collections.defaultdict(list)
for (group, phrase), rows in phrases.items():
    pooled[(group, family(phrase))] += rows
kept, dropped = [], []
MIN_POOLED, KEEP = 5, 0.80
for (group, stem), rows in pooled.items():
    rate = sum(map(yes, rows)) / len(rows)
    good = len(rows) >= MIN_POOLED and rate >= KEEP
    for phrase in members[(group, stem)]:
        own = phrases.get((group, phrase), [])
        own_rate = sum(map(yes, own)) / len(own) if own else None
        ok = good and (own_rate is None or own_rate >= 0.6 or len(own) < 3)
        (kept if ok else dropped).append((group, phrase, len(own), sum(map(yes, own)), round(rate, 3)))
rules = collections.defaultdict(list)
for v in both:
    if not v["group"]:
        rules[v["rule"]].append(v)
catalog = {}
for line in subprocess.run([BINARY, "rules", "list"], capture_output=True, text=True).stdout.splitlines():
    rule = line.split("\t")[0]
    shown = json.loads(subprocess.run([BINARY, "rules", "show", rule], capture_output=True, text=True).stdout)
    catalog[rule] = shown["defaults"]
decisions = {}
for rule, rows in sorted(rules.items()):
    if len(rows) < MIN_RULE:
        continue
    accept, reject = sum(map(yes, rows)) / len(rows), sum(map(no, rows)) / len(rows)
    enabled = catalog[rule]["enabled"]
    if enabled and reject >= 0.6:
        decisions[rule] = ("off", len(rows), accept, reject)
    elif not enabled and accept >= 0.85:
        decisions[rule] = ("on", len(rows), accept, reject)
cluster = []
for (group, phrase), rows in phrases.items():
    if group != "praise" or len(rows) < 4 or " " in phrase:
        continue
    any_yes = sum(1 for v in rows if v["claude"] == "yes" or v["astra"] == "yes") / len(rows)
    if any_yes >= 0.85 and sum(map(no, rows)) / len(rows) <= 0.15:
        cluster.append(phrase)
assignment = {}
by_rule = collections.defaultdict(list)
by_rule["hype.modifier-cluster"] = sorted(cluster)
IMPORTANCE = ("important to note", "worth noting", "worth mentioning", "should be noted", "keep in mind", "bear in mind", "needless to say", "goes without saying")
for group, phrase, n, y, rate in sorted(kept):
    if group in ("praise", "nouns"):
        target = "hype.vague-praise"
    elif group == "verbs":
        target = "filler.wordy-phrase"
    elif any(marker in phrase for marker in IMPORTANCE):
        target = "filler.announced-importance"
    elif phrase.startswith(("is crucial", "is essential", "is vital", "is key to", "plays a", "a crucial role", "a key role", "a vital role")):
        target = "hype.vague-praise"  # an evaluation in place of the property
    elif phrase.startswith(("this document", "this section", "this guide", "provides an overview", "an overview of", "the following")):
        target = "unassigned"  # document self-description belongs to the grammar rule, not to a phrase list
    else:
        target = "filler.wordy-phrase"
    if target != "unassigned":
        by_rule[target].append(phrase)
    assignment[phrase] = {"group": group, "rule": target, "reviewed": n, "both_yes": y, "shrunk_rate": rate}
lines = ["version: 1", "extends: [builtin:technical-v1]", "language: en", "calibration:", "  model: none", "rules:"]
touched = set(by_rule) | set(decisions)
for rule in sorted(touched):
    lines.append(f"  {rule}:")
    state = decisions.get(rule, (None,))[0]
    if rule in by_rule:
        lines.append("    enabled: true")
        merged = list(dict.fromkeys((catalog[rule]["parameters"].get("phrases") or []) + by_rule[rule]))
        lines.append("    parameters:")
        lines.append("      phrases: " + json.dumps(merged))
    elif state:
        lines.append(f"    enabled: {'true' if state == 'on' else 'false'}")
Path(OUT).write_text("\n".join(lines) + "\n")
Path(OUT).with_suffix(".assignment.json").write_text(json.dumps(assignment, indent=1))
Path(OUT).with_suffix(".decisions.json").write_text(json.dumps({"rules": decisions, "kept": kept, "dropped": dropped, "group_rate": group_rate, "cluster": sorted(cluster)}, indent=1))
print(f"reviewed by both={len(both)} group rates={ {k: round(v, 2) for k, v in group_rate.items()} }")
print(f"phrases kept={len(kept)} dropped={len(dropped)} cluster words={len(cluster)}")
for target, items in by_rule.items():
    print(f"  {target}: {', '.join(sorted(items))}")
print("rule decisions:", {r: (d[0], d[1], round(d[2], 2), round(d[3], 2)) for r, d in decisions.items()})
