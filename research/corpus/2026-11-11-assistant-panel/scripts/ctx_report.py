#!/usr/bin/env python3
"""Turn the flags of one calling assistant into a report that reviewers can judge.

usage: ctx_report.py <workspace> <caller>

Writes ctx-workspace/report-<caller>.json in the shape of an engine report and
ctx-workspace/rule-statements.json with one statement per category set. The
reviewers then judge each flag the way they judge a rule finding.
"""
import hashlib, json, os, sys
from pathlib import Path

W, CALLER = Path(sys.argv[1]), sys.argv[2]
CATEGORIES = {
    "empty_framing": ("The opening only announces the statement and adds no information.",
                      "Keep a necessary warning or attribution."),
    "needless_repetition": ("The same claim is stated again with no new distinction.",
                            "Similar words can describe different conditions, objects, or values."),
    "wordiness": ("A longer phrase where a shorter one states the same thing.",
                  "A named stage or artifact may need the noun."),
    "unjustified_intensifiers": ("Praise or emphasis that adds no usable information.",
                                 "Keep calibrated uncertainty and supported comparisons."),
    "vague_claims": ("A claim with no scope or measurement.",
                     "Judge wording and available support, not the truth of an unseen benchmark."),
    "formulaic_transitions": ("A transition that expresses no relationship.",
                              "Keep those that carry a contrast or dependency."),
    "needless_complexity": ("Nesting that conveys no additional condition.", "Keep named stages and exact terms."),
}
out = W / "ctx-workspace"
out.mkdir(exist_ok=True)
if not (out / "engine").exists():
    os.symlink(W / "engine", out / "engine")
statements_path = out / "rule-statements.json"
statements = json.loads(statements_path.read_text()) if statements_path.exists() else {}
findings, unknown = [], 0
for path in sorted((W / "ctx" / CALLER).glob("*.json")):
    record = json.loads(path.read_text())
    for flag in record["accepted"]:
        categories = sorted(set(flag["categories"]))
        if not categories or any(c not in CATEGORIES for c in categories):
            unknown += 1
            continue
        rule = "contextual." + "+".join(categories)
        statements[rule] = {"summary": " Or: ".join(CATEGORIES[c][0] for c in categories),
                            "description": " ".join(CATEGORIES[c][1] for c in categories)}
        span = flag["span"]
        key = f"{record['path']}|{span['start']}|{span['end']}|{rule}"
        findings.append({"id": hashlib.sha256(key.encode()).hexdigest()[:16], "rule_id": rule, "severity": "note",
                         "primary": {"path": record["path"], "span": {"start": span["start"], "end": span["end"]}}})
(out / f"report-{CALLER}.json").write_text(json.dumps({"findings": findings}))
statements_path.write_text(json.dumps(statements, indent=1, sort_keys=True))
print(f"{CALLER}: {len(findings)} flags in {len({f['primary']['path'] for f in findings})} pages; unknown categories {unknown}")
