#!/usr/bin/env python3
"""Per-rule and per-phrase precision from the two assistants' reviews of findings.

usage: rule_table.py <workspace> [out json]
A finding counts as justified when both reviewers say yes, as rejected when both say no.
"""
import collections, json, sys
from pathlib import Path
W = Path(sys.argv[1])
reviews = {}
for name in ("claude", "astra"):
    for path in (W / "rev" / name).glob("*.json"):
        record = json.loads(path.read_text())
        for finding in record["findings"]:
            reviews.setdefault(finding["id"], {"rule": finding["rule"], "group": finding["group"],
                                                "phrase": finding["phrase"], "path": record["path"]})[name] = finding.get("justified")
both = {k: v for k, v in reviews.items() if v.get("claude") and v.get("astra")}
print(f"findings reviewed by both: {len(both)} (claude only so far: {sum(1 for v in reviews.values() if v.get('claude') and not v.get('astra'))})")
agree = sum(v["claude"] == v["astra"] for v in both.values()) / max(1, len(both))
print(f"reviewer agreement on yes/no/uncertain: {100 * agree:.1f}%")


def table(key, minimum):
    rows = collections.defaultdict(collections.Counter)
    for v in both.values():
        k = key(v)
        if k is None:
            continue
        c, a = v["claude"], v["astra"]
        rows[k]["n"] += 1
        rows[k]["both_yes"] += c == "yes" and a == "yes"
        rows[k]["one_yes"] += (c == "yes") != (a == "yes")
        rows[k]["both_no"] += c == "no" and a == "no"
    out = []
    for k, c in rows.items():
        if c["n"] >= minimum:
            out.append((k, c["n"], c["both_yes"] / c["n"], (c["both_yes"] + c["one_yes"]) / c["n"], c["both_no"] / c["n"]))
    return sorted(out, key=lambda r: (-r[2], -r[1]))


print(f"\n{'rule':42s} {'n':>5s} {'both yes':>9s} {'any yes':>8s} {'both no':>8s}")
rules = table(lambda v: v["rule"] if not v["group"] else f"{v['rule']} [{v['group']}]", 5)
for k, n, strict, lenient, no in rules:
    print(f"{k:42s} {n:5d} {100 * strict:8.1f}% {100 * lenient:7.1f}% {100 * no:7.1f}%")
phrases = table(lambda v: v["phrase"] or None, 4)
print(f"\nphrases with at least 4 reviewed findings: {len(phrases)}")
good = [p for p in phrases if p[2] >= 0.85]
print("both-yes >= 85%:", ", ".join(f"{k}({n},{100 * s:.0f}%)" for k, n, s, _, _ in good))
mid = [p for p in phrases if 0.6 <= p[2] < 0.85]
print("60-85%:", ", ".join(f"{k}({n},{100 * s:.0f}%)" for k, n, s, _, _ in mid))
bad = [p for p in phrases if p[2] < 0.6]
print("below 60%:", ", ".join(f"{k}({n},{100 * s:.0f}%)" for k, n, s, _, _ in bad))
if len(sys.argv) > 2:
    json.dump({"rules": rules, "phrases": phrases}, open(sys.argv[2], "w"), indent=1)
