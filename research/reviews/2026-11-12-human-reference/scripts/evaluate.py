#!/usr/bin/env python3
"""Measure systems against the sentences that human editors changed.

usage: evaluate.py <human directory> <out json> [system ...]

A system's findings are read from runs/<system>/<document id>.json as
{"spans": [[start, end], ...]} in bytes. Without system names every directory
under runs/ is measured. The ceiling compares the LAMP writers with each other
in the same way.
"""
import collections, itertools, json, sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from common import overlaps, read_jsonl, sentences

ROOT, OUT = Path(sys.argv[1]), Path(sys.argv[2])
docs = read_jsonl(ROOT / "docs.jsonl")
names = sys.argv[3:] or sorted(p.name for p in (ROOT / "runs").iterdir() if p.is_dir()) if (ROOT / "runs").exists() else []


def count(flags, reference, units, counter):
    """Add sentence and span counts for one document to counter."""
    for unit in units:
        edited, flagged = overlaps(unit, reference), overlaps(unit, flags)
        counter["sentences"] += 1
        counter["edited"] += edited
        counter["flagged"] += flagged
        counter["edited_flagged"] += edited and flagged
        counter["unedited_flagged"] += flagged and not edited
    counter["findings"] += len(flags)
    counter["findings_on_edit"] += sum(overlaps(span, reference) for span in flags)
    counter["edits"] += len(reference)
    counter["edits_touched"] += sum(overlaps(span, flags) for span in reference)
    counter["documents"] += 1


def rates(c):
    unedited = c["sentences"] - c["edited"]
    tpr = c["edited_flagged"] / c["edited"] if c["edited"] else 0.0
    fpr = c["unedited_flagged"] / unedited if unedited else 0.0
    return {**dict(c), "flag_rate": c["flagged"] / c["sentences"] if c["sentences"] else 0.0,
            "base_rate": c["edited"] / c["sentences"] if c["sentences"] else 0.0, "tpr": tpr, "fpr": fpr, "j": tpr - fpr,
            "span_precision": c["findings_on_edit"] / c["findings"] if c["findings"] else 0.0,
            "span_recall": c["edits_touched"] / c["edits"] if c["edits"] else 0.0}


def show(label, r):
    print(f"{label:46} docs {r['documents']:>4} flagged {r['flag_rate']:6.1%} TPR {r['tpr']:6.1%} FPR {r['fpr']:6.1%} "
          f"J {100 * r['j']:+6.1f}  spans: precision {r['span_precision']:6.1%} recall {r['span_recall']:6.1%} "
          f"(findings {r['findings']})")


result = {"sets": {}, "ceiling": {}, "systems": {}}
for (name, split), group in itertools.groupby(sorted(docs, key=lambda d: (d["set"], d["split"])), key=lambda d: (d["set"], d["split"])):
    group = list(group)
    c = collections.Counter()
    for doc in group:
        count([], [e["span"] for e in doc["reference"]], doc.get("units") or sentences(doc["text"]), c)
    result["sets"][f"{name}/{split}"] = rates(c)
    print(f"{name}/{split}: documents {len(group)}, sentences {c['sentences']}, edited {c['edited'] / c['sentences']:.1%}, "
          f"edits {c['edits']}")

by_text = collections.defaultdict(dict)
for doc in docs:
    if doc["split"] == "ceiling":
        by_text[doc["text"]][doc["editor"]] = [e["span"] for e in doc["reference"]]
pooled = collections.Counter()
pairs = collections.defaultdict(collections.Counter)
for text, writers in by_text.items():
    units = sentences(text)
    for first, second in itertools.permutations(sorted(writers), 2):
        count(writers[first], writers[second], units, pairs[f"{first} against {second}"])
        count(writers[first], writers[second], units, pooled)
for label, c in sorted(pairs.items()):
    result["ceiling"][label] = rates(c)
    show("ceiling " + label, result["ceiling"][label])
if pooled:
    result["ceiling"]["pooled"] = rates(pooled)
    show("ceiling pooled", result["ceiling"]["pooled"])

for system in names:
    for (name, split), group in itertools.groupby(sorted(docs, key=lambda d: (d["set"], d["split"])), key=lambda d: (d["set"], d["split"])):
        if split == "ceiling":
            continue
        c, missing = collections.Counter(), 0
        for doc in group:
            path = ROOT / "runs" / system / f"{doc['id']}.json"
            if not path.exists():
                missing += 1
                continue
            flags = json.loads(path.read_text())["spans"]
            count(flags, [e["span"] for e in doc["reference"]], doc.get("units") or sentences(doc["text"]), c)
        if not c["documents"]:
            continue
        result["systems"].setdefault(system, {})[f"{name}/{split}"] = {**rates(c), "missing": missing}
        show(f"{system} {name}/{split}" + (f" (missing {missing})" if missing else ""), rates(c))
OUT.write_text(json.dumps(result, indent=1, sort_keys=True) + "\n")
