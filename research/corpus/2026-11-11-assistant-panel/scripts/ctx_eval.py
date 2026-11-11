#!/usr/bin/env python3
"""Measure review in context through the MCP server.

usage: ctx_eval.py <workspace> <out json>

For each calling assistant: how many flags the server bound, how the
assistant of the other family judged them, what a third assistant said on the
part it judged, how many flags lie outside the findings of the default rules,
and how many items marked by the other annotator the flags touch.
"""
import collections, json, sys
from pathlib import Path

W, OUT = Path(sys.argv[1]), Path(sys.argv[2])
CALLERS = {"claude": "astra", "astra": "claude"}
THIRD = "qwen"

docs = {}
for name in ("docs-real.jsonl", "docs-real2.jsonl"):
    folder = name.removeprefix("docs-").removesuffix(".jsonl")
    for line in open(W / name):
        doc = json.loads(line)
        docs[f"{folder}/{doc['doc_id']}.md"] = doc
rule_spans = collections.defaultdict(list)
for f in json.loads((W / "engine/report-all5-final.json").read_text())["findings"]:
    span = f["primary"]["span"]
    rule_spans[f["primary"]["path"]].append((span["start"], span["end"]))


def marked(doc_id, annotator):
    path = W / "ann" / annotator / f"{doc_id}.json"
    if not path.exists():
        return set()
    return {f["item"] for f in json.loads(path.read_text())["flags"] if f["label"] == "needs_revision"}


def overlaps(span, others):
    return any(span[0] < end and start < span[1] for start, end in others)


def verdicts(caller, reviewer):
    out, rules = {}, {}
    directory = W / "ctx-rev" / caller / reviewer
    for path in directory.glob("*.json") if directory.exists() else []:
        for g in json.loads(path.read_text())["findings"]:
            if g.get("justified"):
                out[g["id"]] = g["justified"]
                rules[g["id"]] = g["rule"]
    return out, rules


result = {}
for caller, reviewer in CALLERS.items():
    directory = W / "ctx" / caller
    if not directory.exists():
        continue
    stats = collections.Counter()
    for path in sorted(directory.glob("*.json")):
        record = json.loads(path.read_text())
        doc = docs[record["path"]]
        stats["pages"] += 1
        stats["items"] += record["items"]
        stats["words"] += len(doc["text"].split())
        stats["flags"] += record["flags"]
        stats["bound"] += len(record["accepted"])
        stats["rejected"] += len(record["rejected"])
        spans = [(f["span"]["start"], f["span"]["end"]) for f in record["accepted"]]
        stats["beyond_rules"] += sum(not overlaps(span, rule_spans[record["path"]]) for span in spans)
        other = marked(doc["doc_id"], reviewer)
        for item in doc["items"]:
            if item["id"] in other:
                segments = [tuple(s) for s in item["segments"]]
                stats["other_marked"] += 1
                stats["other_marked_touched"] += any(overlaps(span, segments) for span in spans)
                stats["other_marked_touched_by_rules"] += any(overlaps(span, segments) for span in rule_spans[record["path"]])
    first, rules = verdicts(caller, reviewer)
    third, _ = verdicts(caller, THIRD)
    review = collections.Counter()
    by_category = collections.defaultdict(collections.Counter)
    for identifier, verdict in first.items():
        for counter in [review] + [by_category[c] for c in rules[identifier].removeprefix("contextual.").split("+")]:
            counter["reviewed"] += 1
            counter["accept"] += verdict == "yes"
            counter["reject"] += verdict == "no"
    check = collections.Counter()
    for identifier, verdict in third.items():
        if identifier in first:
            check["reviewed"] += 1
            check["accept"] += verdict == "yes"
            check["same"] += verdict == first[identifier]
            check["both_accept"] += verdict == "yes" and first[identifier] == "yes"
    result[caller] = {"reviewer": reviewer, **stats, "review": dict(review), "third": dict(check),
                      "categories": {c: dict(v) for c, v in sorted(by_category.items())}}
    n = max(review["reviewed"], 1)
    print(f"{caller}: pages {stats['pages']}, flags {stats['flags']}, bound {stats['bound']}, rejected {stats['rejected']}, "
          f"per 1,000 words {1000 * stats['bound'] / max(stats['words'], 1):.2f}, beyond rules {stats['beyond_rules']}")
    print(f"   reviewed {review['reviewed']} by {reviewer}: accept {review['accept'] / n:.1%}, reject {review['reject'] / n:.1%}")
    if check["reviewed"]:
        print(f"   third assistant on {check['reviewed']}: accept {check['accept'] / check['reviewed']:.1%}, "
              f"same answer {check['same'] / check['reviewed']:.1%}, both accept {check['both_accept'] / check['reviewed']:.1%}")
    print(f"   items marked by the other annotator {stats['other_marked']}: touched by flags {stats['other_marked_touched']}, "
          f"by rules {stats['other_marked_touched_by_rules']}")
    for category, c in sorted(by_category.items(), key=lambda kv: -kv[1]["reviewed"]):
        print(f"      {category:26} reviewed {c['reviewed']:>4} accept {c['accept'] / c['reviewed']:.1%}")
OUT.write_text(json.dumps(result, indent=1, sort_keys=True) + "\n")
