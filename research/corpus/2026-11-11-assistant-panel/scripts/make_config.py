#!/usr/bin/env python3
"""Write an engine configuration that enables every rule and extends the phrase lists.

usage: make_config.py <workspace> <unswell binary> <out yaml> [qualified json]
Without a qualified file every candidate phrase is included (broad configuration).
"""
import json, subprocess, sys
from pathlib import Path

W, BINARY, OUT = Path(sys.argv[1]), sys.argv[2], sys.argv[3]
lexicon = json.loads((W / "lexicon-candidates.json").read_text())
keep = None
if len(sys.argv) > 4:
    keep = set(json.loads(Path(sys.argv[4]).read_text())["phrases"])
v0 = {r["phrase"]: r for r in json.loads((W / "lexicon-v0.json").read_text())["phrases"]}
IMPORTANCE = {"it is important to note", "it's important to note", "important to note", "it is worth noting",
              "it's worth noting", "worth noting", "worth mentioning", "it should be noted", "keep in mind that",
              "bear in mind", "needless to say", "it goes without saying"}
DIVE = {"dive into", "deep dive", "let's explore", "let's dive", "let's take a look"}
CONNECTIVES = {"additionally", "furthermore", "moreover", "consequently", "in addition", "therefore", "thus", "hence",
               "nevertheless", "nonetheless", "accordingly", "subsequently", "likewise", "similarly"}
rules = {"hype.vague-praise": [], "policy.banned-phrases": [], "filler.announced-importance": [],
         "scaffold.dive-in": [], "density.connective-overuse": []}
assignment = {}
for group, phrases in lexicon.items():
    for phrase in phrases:
        if keep is not None and phrase not in keep:
            continue
        stat = v0.get(phrase)
        if keep is None and stat and sum(stat["select"]) + sum(stat["check"]) >= 40:
            seen = [a + b for a, b in zip(stat["select"], stat["check"])]
            if (seen[0] + seen[1]) / sum(seen) < 0.3:
                continue  # frequent and almost never flagged: ordinary technical wording
        if phrase in IMPORTANCE:
            rule = "filler.announced-importance"
        elif phrase in DIVE:
            rule = "scaffold.dive-in"
        elif phrase in CONNECTIVES:
            rule = "density.connective-overuse"
        elif group in ("praise", "nouns"):
            rule = "hype.vague-praise"
        else:
            rule = "policy.banned-phrases"
        rules[rule].append(phrase)
        assignment[phrase] = {"group": group, "rule": rule}
ids = [line.split("\t")[0] for line in subprocess.run([BINARY, "rules", "list"], capture_output=True, text=True).stdout.splitlines() if line.strip()]
defaults = {}
for rule in rules:
    shown = json.loads(subprocess.run([BINARY, "rules", "show", rule], capture_output=True, text=True).stdout)
    defaults[rule] = shown["defaults"]["parameters"].get("phrases") or []
lines = ["version: 1", "extends: [builtin:technical-v1]", "language: en", "calibration:", "  model: none", "rules:"]
for rule in ids:
    lines.append(f"  {rule}:")
    lines.append("    enabled: true")
    if rule in rules:
        merged = list(dict.fromkeys(defaults[rule] + rules[rule]))
        lines.append("    parameters:")
        lines.append("      phrases: " + json.dumps(merged))
Path(OUT).write_text("\n".join(lines) + "\n")
Path(OUT).with_suffix(".assignment.json").write_text(json.dumps(assignment, indent=1))
print({rule: len(v) for rule, v in rules.items()}, "phrases total", len(assignment), "rules enabled", len(ids))
