#!/usr/bin/env python3
"""Write builtin/lexicon_qualified.go from the selection made on the reviewed findings.

usage: gen_lexicon_go.py <selection decisions json> <selection assignment json> <out go file>
"""
import json, sys
from pathlib import Path

decisions = json.loads(Path(sys.argv[1]).read_text())
assignment = json.loads(Path(sys.argv[2]).read_text())
by_rule = {}
for phrase, row in assignment.items():
    if row["rule"] != "unassigned":
        by_rule.setdefault(row["rule"], []).append(phrase)
cluster = decisions.get("cluster", [])


def block(name, comment, items):
    lines = [f"// {line}" for line in comment]
    lines.append(f"func {name}() []string {{")
    lines.append("\treturn []string{")
    row = "\t\t"
    for item in sorted(set(items)):
        piece = json.dumps(item) + ", "
        if len(row) + len(piece) > 100:
            lines.append(row.rstrip())
            row = "\t\t"
        row += piece
    if row.strip():
        lines.append(row.rstrip())
    lines += ["\t}", "}", ""]
    return "\n".join(lines)


out = ['package builtin', '',
       '// This file holds phrase lists qualified on the assistant-panel corpus.',
       '// Two assistants from different model families reviewed each finding under',
       '// the statement of its rule. A phrase is listed when both accepted at least',
       '// 80% of at least five reviewed findings of its word family.',
       '// docs/research/rule-qualification.md records the measurement.', '']
out.append(block("qualifiedPraise", ["qualifiedPraise lists general praise that names no property or measurement."],
                 by_rule.get("hype.vague-praise", [])))
out.append(block("qualifiedClusterWords", ["qualifiedClusterWords lists evaluative modifiers that count toward a cluster.",
                                           "One alone is not reported by the cluster rule."], cluster))
out.append(block("qualifiedWordyPhrases", ["qualifiedWordyPhrases lists stock verbs and phrases that a plain word replaces."],
                 by_rule.get("filler.wordy-phrase", [])))
importance = by_rule.get("filler.announced-importance", [])
if importance:
    out.append(block("qualifiedImportanceFrames", ["qualifiedImportanceFrames lists sentence openings that announce importance."],
                     importance))
Path(sys.argv[3]).write_text("\n".join(out).rstrip("\n") + "\n")
import subprocess
subprocess.run(["gofmt", "-w", sys.argv[3]], check=True)
print({k: len(v) for k, v in by_rule.items()}, "cluster", len(cluster))
