#!/usr/bin/env python3
"""Compare two engine reports on reviewed real pages.

usage: compare_rounds.py <workspace> <old report> <new report> <out json>
           [--reviewed report,...] [--rounds folder,...]

Reviews are matched to findings by page, rule, and byte span, so a finding
keeps its verdict when a rule version changes and its span does not. The
defaults name the reports and review folders of the pages opened before round
three.
"""
import argparse, collections, json
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument("workspace")
parser.add_argument("old")
parser.add_argument("new")
parser.add_argument("out")
parser.add_argument("--reviewed", default="report-broad-dev.json,report-broad-dev2.json,report-final-new.json,report-real3-r2.json")
parser.add_argument("--rounds", default="rev,rev-final,rev-real3")
options = parser.parse_args()
W = Path(options.workspace)


def load(name):
    report = json.loads((W / "engine" / name).read_text())
    out = {}
    for f in report["findings"]:
        span = f["primary"]["span"]
        out[(f["primary"]["path"], f["rule_id"], span["start"], span["end"])] = f
    return out, len(report["documents"])


key_of = {}
for name in options.reviewed.split(","):
    for key, f in load(name)[0].items():
        key_of[f["id"]] = key
votes = collections.defaultdict(dict)
for folder in options.rounds.split(","):
    for reviewer in ("claude", "astra", "qwen"):
        directory = W / folder / reviewer
        if not directory.exists():
            continue
        for path in directory.glob("*.json"):
            for g in json.loads(path.read_text())["findings"]:
                if g.get("justified") and g["id"] in key_of and not key_of[g["id"]][0].startswith("v0/"):
                    votes[key_of[g["id"]]][reviewer] = g["justified"]
verdict = {}
for key, vote in votes.items():
    first, second = vote.get("claude"), vote.get("astra") or vote.get("qwen")
    if first and second:
        verdict[key] = (first, second)


def count(counter, pair):
    counter["reviewed"] += 1
    counter["both_accept"] += pair == ("yes", "yes")
    counter["any_accept"] += "yes" in pair
    counter["both_reject"] += pair == ("no", "no")


def table(label, findings):
    by_rule = collections.defaultdict(collections.Counter)
    by_severity = collections.defaultdict(collections.Counter)
    reported = collections.Counter(f["severity"] for f in findings.values())
    total = collections.Counter()
    for key, f in findings.items():
        if key in verdict:
            count(by_rule[(key[1], f["severity"])], verdict[key])
            count(by_severity[f["severity"]], verdict[key])
            count(total, verdict[key])
    print(f"{label}: {len(findings)} findings, reviewed {total['reviewed']}, "
          f"both accept {total['both_accept'] / max(total['reviewed'], 1):.1%}")
    for severity, c in sorted(by_severity.items()):
        print(f"  {severity:8} reviewed={c['reviewed']:>5} both accept={c['both_accept'] / c['reviewed']:6.1%}")
    for (rule, severity), c in sorted(by_rule.items(), key=lambda kv: -kv[1]["reviewed"]):
        if c["reviewed"] >= 5:
            print(f"  {rule:34} {severity:8} reviewed={c['reviewed']:>5} both accept={c['both_accept'] / c['reviewed']:6.1%}")
    return {"findings": len(findings), "reported": dict(reported), "all": dict(total),
            "severity": {severity: dict(c) for severity, c in by_severity.items()},
            "rules": [{"rule": rule, "severity": severity, **c} for (rule, severity), c in
                      sorted(by_rule.items(), key=lambda kv: (-kv[1]["reviewed"], kv[0]))]}


(old, documents), (new, _) = load(options.old), load(options.new)
result = {"documents": documents, "old": table("old", old), "new": table("new", new),
          "removed": table("removed", {k: f for k, f in old.items() if k not in new})}
added = collections.Counter(k[1] for k in new if k not in old)
print("added", dict(added))
Path(options.out).write_text(json.dumps(result, indent=1, sort_keys=True) + "\n")
