#!/usr/bin/env python3
"""Evidence for the decisions that were fixed before final test 4 was read.

usage: round4_evidence.py <workspace> <out json>

Reads the reviewed findings of the current defaults on every real page and
reports three things: long sentences with eight or more commas, pooled and on
the new pages; exact-sentence findings, pooled; and vague-praise findings on
the pages opened after its list was fixed.
"""
import collections, json, sys
from pathlib import Path

W, OUT = Path(sys.argv[1]), Path(sys.argv[2])
EARLIER = {"reports": ["report-broad-dev.json", "report-broad-dev2.json", "report-final-new.json",
                       "report-real3-r2.json", "report-real4-r3.json"],
           "rounds": ["rev", "rev-final", "rev-real3", "rev-real4"],
           "current": ["report-all5-final.json", "report-real4-final.json"]}
NEW = {"reports": ["report-real5-r4.json"], "rounds": ["rev-real5"], "current": ["report-real5-r4.json"]}
RULES = {"syntax.long-sentence", "repetition.exact-sentence", "hype.vague-praise"}


def key(f):
    span = f["primary"]["span"]
    return (f["primary"]["path"], f["rule_id"], span["start"], span["end"])


def reviewed(part):
    current = {}
    for name in part["current"]:
        for f in json.loads((W / "engine" / name).read_text())["findings"]:
            if f["rule_id"] in RULES and not f["primary"]["path"].startswith("v0/"):
                current[key(f)] = f
    key_of = {}
    for name in part["reports"]:
        for f in json.loads((W / "engine" / name).read_text())["findings"]:
            if f["rule_id"] in RULES:
                key_of[f["id"]] = key(f)
    votes = collections.defaultdict(dict)
    for folder in part["rounds"]:
        for reviewer in ("claude", "astra", "qwen"):
            directory = W / folder / reviewer
            for path in directory.glob("*.json") if directory.exists() else []:
                for g in json.loads(path.read_text())["findings"]:
                    if g.get("justified") and key_of.get(g["id"]) in current:
                        votes[key_of[g["id"]]][reviewer] = g["justified"]
    rows = []
    for finding_key, vote in votes.items():
        first, second = vote.get("claude"), vote.get("astra") or vote.get("qwen")
        if first and second:
            rows.append((current[finding_key], first == "yes" and second == "yes"))
    return rows


def commas(f, sources):
    path = f["primary"]["path"]
    source = sources.setdefault(path, (W / "engine" / path).read_bytes())
    span = f["primary"]["span"]
    return source[span["start"]:span["end"]].count(b",")


def tally(rows, predicate):
    chosen = [accepted for f, accepted in rows if predicate(f)]
    return {"reviewed": len(chosen), "both_accept": sum(chosen)}


sources = {}
earlier, new = reviewed(EARLIER), reviewed(NEW)
long_sentence = lambda f: f["rule_id"] == "syntax.long-sentence"
result = {
    "long_sentence_eight_commas": {
        "earlier": tally(earlier, lambda f: long_sentence(f) and commas(f, sources) >= 8),
        "new": tally(new, lambda f: long_sentence(f) and commas(f, sources) >= 8)},
    "long_sentence_fewer_commas": {
        "earlier": tally(earlier, lambda f: long_sentence(f) and commas(f, sources) < 8),
        "new": tally(new, lambda f: long_sentence(f) and commas(f, sources) < 8)},
    "exact_sentence": {"earlier": tally(earlier, lambda f: f["rule_id"] == "repetition.exact-sentence"),
                       "new": tally(new, lambda f: f["rule_id"] == "repetition.exact-sentence")},
    "vague_praise": {"new": tally(new, lambda f: f["rule_id"] == "hype.vague-praise")},
}
OUT.write_text(json.dumps(result, indent=1, sort_keys=True) + "\n")
for name, parts in result.items():
    for part, counts in parts.items():
        share = f"{counts['both_accept'] / counts['reviewed']:.1%}" if counts["reviewed"] else "-"
        print(f"{name:28} {part:8} reviewed {counts['reviewed']:>5} both accept {share}")
