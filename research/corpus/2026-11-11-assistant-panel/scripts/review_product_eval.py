#!/usr/bin/env python3
"""Coverage and agreement of the review command on generated documents.

usage: review_product_eval.py <workspace> <run dir> <reference annotator> [--rules <engine report>] [--out <json>]

Every bound flag of the first reviewer is "before"; the flags the confirming
model accepted are "after". Marked items come from the annotation of the
reference annotator, who should be neither the reviewer nor the confirming
model of the run. Items are matched by byte overlap, not by identifier.
"""
import argparse, collections, json
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument("workspace")
parser.add_argument("run")
parser.add_argument("reference", choices=["astra", "claude"])
parser.add_argument("--rules")
parser.add_argument("--out")
options = parser.parse_args()
W = Path(options.workspace)
docs = {}
for line in open(W / "docs.jsonl"):
    doc = json.loads(line)
    docs[doc["doc_id"]] = doc


def marked(doc_id, annotator):
    flags = json.loads((W / "ann" / annotator / f"{doc_id}.json").read_text())["flags"]
    return {f["item"] for f in flags if f["label"] == "needs_revision"}


def overlaps(span, others):
    return any(span[0] < end and start < span[1] for start, end in others)


rule_spans = collections.defaultdict(list)
if options.rules:
    for finding in json.loads(Path(options.rules).read_text())["findings"]:
        span = finding["primary"]["span"]
        rule_spans[Path(finding["primary"]["path"]).stem].append((span["start"], span["end"]))

stats = collections.Counter()
verdicts = collections.Counter()
by_family = collections.defaultdict(collections.Counter)
other = "claude" if options.reference == "astra" else "astra"
for path in sorted(Path(options.run).glob("*.json")):
    doc = docs[path.stem]
    result = json.loads(path.read_text())["documents"][0]
    after = [(f["span"]["start"], f["span"]["end"]) for f in result["findings"]]
    dropped = [(f["span"]["start"], f["span"]["end"]) for f in result["unconfirmed"]]
    before = after + dropped
    for finding in result["findings"] + result["unconfirmed"]:
        verdicts[(finding.get("confirmation") or {}).get("justified") or "none"] += 1
    reference, second = marked(doc["doc_id"], options.reference), marked(doc["doc_id"], other)
    segments = {item["id"]: [tuple(s) for s in item["segments"]] for item in doc["items"]}
    sets = {"reference": reference, "both": reference & second, "any": reference | second}
    counters = [stats, by_family[doc["family"] + "/" + doc["operation"]]]
    for counter in counters:
        counter["pages"] += 1
        counter["words"] += len(doc["text"].split())
        counter["items"] += len(doc["items"])
        counter["flags"] += result["flags"]
        counter["unbound"] += len(result["unbound"])
        counter["before"] += len(before)
        counter["after"] += len(after)
        for name, items in sets.items():
            counter[name + "_marked"] += len(items)
            spans = [segments[i] for i in items]
            counter[name + "_touched_before"] += sum(any(overlaps(s, parts) for s in before) for parts in spans)
            counter[name + "_touched_after"] += sum(any(overlaps(s, parts) for s in after) for parts in spans)
            counter[name + "_touched_rules"] += sum(any(overlaps(s, parts) for s in rule_spans[doc["doc_id"]]) for parts in spans)
            flat = [segment for parts in spans for segment in parts]
            counter["before_on_" + name] += sum(overlaps(s, flat) for s in before)
            counter["after_on_" + name] += sum(overlaps(s, flat) for s in after)
            counter["dropped_on_" + name] += sum(overlaps(s, flat) for s in dropped)
            counter["rules_on_" + name] += sum(overlaps(s, flat) for s in rule_spans[doc["doc_id"]])
        counter["rules"] += len(rule_spans[doc["doc_id"]])


def share(part, whole):
    return f"{part}/{whole} = {part / whole:.1%}" if whole else "0/0"


def show(title, c):
    print(f"{title}: pages {c['pages']}, words {c['words']}, flags {c['flags']}, unbound {c['unbound']}, "
          f"bound {c['before']} ({1000 * c['before'] / c['words']:.2f} per 1,000 words), "
          f"after confirmation {c['after']} ({1000 * c['after'] / c['words']:.2f})")
    for name in ("reference", "both", "any"):
        print(f"   {name:9} marked {c[name + '_marked']:>5}: touched before {share(c[name + '_touched_before'], c[name + '_marked'])}, "
              f"after {share(c[name + '_touched_after'], c[name + '_marked'])}, rules {share(c[name + '_touched_rules'], c[name + '_marked'])}")
    for name in ("reference", "any"):
        print(f"   flags on an item marked by {name:9}: before {share(c['before_on_' + name], c['before'])}, "
              f"after {share(c['after_on_' + name], c['after'])}, dropped {share(c['dropped_on_' + name], c['before'] - c['after'])}, "
              f"rules {share(c['rules_on_' + name], c['rules'])}")


show(f"{options.run} (reference {options.reference})", stats)
print("   verdicts:", dict(verdicts))
for family, counter in sorted(by_family.items()):
    print(f"   {family:15} pages {counter['pages']:>3} reference touched before "
          f"{share(counter['reference_touched_before'], counter['reference_marked'])}, after "
          f"{share(counter['reference_touched_after'], counter['reference_marked'])}")
if options.out:
    Path(options.out).write_text(json.dumps({"run": options.run, "reference": options.reference, **stats,
                                             "verdicts": dict(verdicts),
                                             "families": {k: dict(v) for k, v in sorted(by_family.items())}},
                                            indent=1, sort_keys=True) + "\n")
