#!/usr/bin/env python3
"""Summarize the fix experiment: the index after each fix, and what the panel says.

usage: fix_summary.py <unswell binary> <arms jsonl> <judge jsonl>... [--json <out>]

For each arm: pages fixed, how much of the page changed, the index before and after. For each comparison and judge:
which version was named the more AI-sounding one. A page-level verdict adds the judges: a judge naming the first
version counts for the second one and the reverse; "same" and a missing answer count for neither.
"""
import difflib, json, math, re, subprocess, sys, tempfile
from pathlib import Path

arguments = sys.argv[1:]
target = None
if "--json" in arguments:
    target = arguments[arguments.index("--json") + 1]
    del arguments[arguments.index("--json"):arguments.index("--json") + 2]
binary, arms_path, judge_paths = arguments[0], arguments[1], arguments[2:]
ARMS = ("unswell", "google", "microsoft", "write-good", "mechanical")
LINTERS = ("google", "microsoft", "write-good")
TOKEN = re.compile(r"\w+|[^\w\s]")


def wilson(yes, total, z=1.96):
    if not total:
        return (0.0, 1.0)
    p = yes / total
    centre, spread = p + z * z / (2 * total), z * math.sqrt(p * (1 - p) / total + z * z / (4 * total * total))
    return ((centre - spread) / (1 + z * z / total), (centre + spread) / (1 + z * z / total))


def index(text):
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--no-gate", "--allow-empty",
                               "--report", "json:-"], input=text.encode(), cwd=scratch, capture_output=True)
    value = json.loads(done.stdout)["documents"][0].get("ai_sounding")
    return value["index"] if value and value["status"] == "available" else None


rows = [json.loads(line) for line in open(arms_path)]
result = {"pages": len(rows), "arms": {}, "comparisons": {}}
pages = []
for row in rows:
    page = {"id": row["id"], "original": index(row["original"]), "arms": {}}
    before = TOKEN.findall(row["original"])
    for name in ARMS:
        text = row["arms"].get(name)
        if text is None:
            page["arms"][name] = None
            continue
        after = TOKEN.findall(text)
        page["arms"][name] = {"index": index(text), "changed": round(1 - difflib.SequenceMatcher(None, before, after, autojunk=False).ratio(), 4),
                              "length": round(len(after) / len(before), 4), "identical": text.strip() == row["original"].strip()}
    pages.append(page)
for name in ARMS:
    fixed = [page for page in pages if page["arms"][name] and page["arms"][name]["index"] is not None and page["original"] is not None]
    entry = {"failed": sum(page["arms"][name] is None for page in pages), "measured": len(fixed),
             "identical": sum(page["arms"][name]["identical"] for page in fixed),
             "lower": sum(page["arms"][name]["index"] < page["original"] for page in fixed),
             "higher": sum(page["arms"][name]["index"] > page["original"] for page in fixed),
             "median_before": sorted(page["original"] for page in fixed)[len(fixed) // 2] if fixed else None,
             "median_after": sorted(page["arms"][name]["index"] for page in fixed)[len(fixed) // 2] if fixed else None,
             "high_before": sum(page["original"] >= 70 for page in fixed), "high_after": sum(page["arms"][name]["index"] >= 70 for page in fixed),
             "median_changed": sorted(page["arms"][name]["changed"] for page in fixed)[len(fixed) // 2] if fixed else None}
    result["arms"][name] = entry
    print(f"{name:11} fixed {entry['measured']:3} failed {entry['failed']} identical {entry['identical']:2} | index lower {entry['lower']:3} higher {entry['higher']:2} "
          f"| median {entry['median_before']} -> {entry['median_after']} | high {entry['high_before']} -> {entry['high_after']} | median changed {entry['median_changed']}")

answers = {}
for path in judge_paths:
    judge = Path(path).stem
    for line in open(path):
        row = json.loads(line)
        if not row.get("failed"):
            answers.setdefault(row["id"], {})[judge] = row["picked"]
judges = [Path(path).stem for path in judge_paths]
for first, second in [("original", arm) for arm in ARMS] + [(ours, linter) for ours in ("unswell", "mechanical") for linter in LINTERS]:
    entry = {"judges": {}, "pages": {"second_less": 0, "first_less": 0, "tie": 0}}
    for judge in judges:
        tally = {"first": 0, "second": 0, "same": 0, "missing": 0}
        for row in rows:
            tally[answers.get(f"{row['id']}|{first}|{second}", {}).get(judge, "missing")] += 1
        entry["judges"][judge] = tally
    for row in rows:
        votes = answers.get(f"{row['id']}|{first}|{second}", {})
        score = sum((picked == "first") - (picked == "second") for picked in votes.values())
        entry["pages"]["second_less" if score > 0 else "first_less" if score < 0 else "tie"] += 1
    decisive = entry["pages"]["second_less"] + entry["pages"]["first_less"]
    low, high = wilson(entry["pages"]["second_less"], decisive)
    entry["pages"]["interval"] = [low, high]
    result["comparisons"][f"{first}|{second}"] = entry
    per_judge = "; ".join(f"{judge}: {t['first']}/{t['second']}/{t['same']}/{t['missing']}" for judge, t in entry["judges"].items())
    share = f"{100 * entry['pages']['second_less'] / decisive:.1f}% ({100 * low:.1f}-{100 * high:.1f})" if decisive else "n/a"
    print(f"{first} vs {second}: pages where {second} is the less AI-sounding {entry['pages']['second_less']}, {first} {entry['pages']['first_less']}, "
          f"tie {entry['pages']['tie']}: {share} | named more AI-sounding first/second/same/missing: {per_judge}")
result["page_rows"] = pages
if target:
    Path(target).write_text(json.dumps(result, indent=1))
