#!/usr/bin/env python3
"""Export the answers of the judges on pairs before and after a de-slop edit by people into the record.

usage: export_judge_pairs.py <workspace> <pairs jsonl> <record data file> <comparison>:<judge name>=<judge jsonl>...

A comparison is "edit" (the pair as it is, judged from the pairs file) or "equalized" (the file of pairs made with
`unswell fix`: its ids end in "|bfix|afix" for both versions fixed and in "|bfix|after" for the fixed earlier version
against the person's version). A row of the record holds the pair, its repository, commit and path, the layer, and
per comparison each judge's answer: "before", "after", or "same" for the side named as reading more like text written
by a language model. No page text and no model output is exported.
"""
import gzip, json, sys
from pathlib import Path

W, pairs_path, out = sys.argv[1].rstrip("/") + "/", sys.argv[2], Path(sys.argv[3])
docs = {}
for layer, path in (("reserve", "assisted/deslop-reserve-files.jsonl"), ("web", "assisted/manual-files.jsonl")):
    for line in open(W + path):
        doc = json.loads(line)
        if not doc.get("skip"):
            docs.setdefault(doc["id"], (layer, doc))
pairs = {json.loads(line)["id"]: json.loads(line) for line in open(pairs_path)}
rows = {}
for identifier in pairs:
    layer, doc = docs[identifier]
    rows[identifier] = {"pair": identifier, "repository": doc["repository"], "commit": doc["commit"], "path": doc["path"],
                        "layer": layer, "comparisons": {}}
KINDS = {"bfix|afix": "edit_both_fixed", "bfix|after": "fixed_against_edit"}
for argument in sys.argv[4:]:
    spec, path = argument.split("=", 1)
    comparison, judge = spec.split(":", 1)
    for line in open(path):
        row = json.loads(line)
        if row.get("failed"):
            continue
        if comparison == "edit":
            identifier, name = row["id"], "edit"
            first = pairs[identifier]["first_is"]
        else:
            identifier, _, kind = row["id"].partition("|")
            name, first = KINDS[kind], "before"
        second = "after" if first == "before" else "before"
        answer = {"first": first, "second": second, "same": "same"}[row["picked"]]
        rows[identifier]["comparisons"].setdefault(name, {})[judge] = answer
with gzip.GzipFile(out, "wb", mtime=0) as handle:
    for identifier in sorted(rows):
        handle.write((json.dumps(rows[identifier], sort_keys=True) + "\n").encode())
print("pairs", len(rows))
