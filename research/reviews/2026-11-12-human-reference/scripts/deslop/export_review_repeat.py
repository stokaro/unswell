#!/usr/bin/env python3
"""Export per-document counts of the two answer protocols of the review, for the record.

usage: export_review_repeat.py <workspace> <system> <reviewer name> <out jsonl.gz>

Reads confirm-new-1 and confirm-new-2 (a verdict for every item; the first also
holds the texts after the commit) and confirm-old-1 and confirm-old-2 (only
the items that need revision). A row is one text before a commit under one
protocol: words, words the commit changed, and for each of the two runs the
findings, the flagged words, the flagged words that the commit changed, and
the flagged sentences; then the sentences flagged in both runs. Rows of the
verdict form also say what the commit did to the flagged sentences of the
first run: how many of them it changed or removed, against the same count
for all sentences of six or more words, and how many of the sentences flagged
on the text after the commit are new or changed, or stood unchanged and were
flagged before.
"""
import bisect, gzip, json, re, sys

W = sys.argv[1].rstrip("/") + "/"
sys.path.insert(0, W + "human/scripts")
from common import sentences

system, reviewer, out = sys.argv[2], sys.argv[3], sys.argv[4]
WORD = re.compile(r"[A-Za-z][A-Za-z'’-]*")


def overlaps(begin, end, spans):
    return any(a < end and begin < b for a, b in spans)


def run(name, doc):
    return [tuple(s) for s in json.load(open(f"{W}{name}/runs/{system}/{doc['id']}.json"))["spans"]]


def flagged_sentences(doc, spans):
    units = sorted(tuple(s) for s in sentences(doc["text"]))
    starts = [u[0] for u in units]
    result = set()
    for begin, _ in spans:
        index = bisect.bisect_right(starts, begin) - 1
        result.add(index if index >= 0 and begin < units[index][1] else ("outside", begin))
    return result


def flagged_texts(doc, spans):
    """The flagged sentences as text, with runs of white space folded."""
    raw = doc["text"].encode()
    units = sorted(tuple(s) for s in sentences(doc["text"]))
    starts = [u[0] for u in units]
    result = set()
    for begin, _ in spans:
        index = bisect.bisect_right(starts, begin) - 1
        if index >= 0:
            result.add(" ".join(raw[units[index][0]:units[index][1]].decode("utf-8", "replace").split()))
    return result


docs = [json.loads(line) for line in open(W + "confirm-new-1/docs.jsonl")]
after = {doc["pair"]: doc for doc in docs if doc["side"] == "after"}
rows = []
for doc in docs:
    if doc["side"] != "before":
        continue
    text = doc["text"]
    words = [(len(text[:m.start()].encode()), len(text[:m.end()].encode())) for m in WORD.finditer(text)]
    reference = [tuple(r["span"]) for r in doc["reference"]]
    changed = [overlaps(b, e, reference) for b, e in words]
    for protocol, names in (("verdicts", ("confirm-new-1", "confirm-new-2")), ("list", ("confirm-old-1", "confirm-old-2"))):
        runs = [run(name, doc) for name in names]
        marks = [[overlaps(b, e, spans) for b, e in words] for spans in runs]
        sets = [flagged_sentences(doc, spans) for spans in runs]
        row = {"document": doc["pair"], "reviewer": reviewer, "protocol": protocol, "words": len(words), "changed": sum(changed),
               "findings": [len(spans) for spans in runs], "flagged": [sum(m) for m in marks],
               "flagged_changed": [sum(f and c for f, c in zip(m, changed)) for m in marks],
               "sentences": [len(s) for s in sets], "both": len(sets[0] & sets[1]), "after_words": None, "after_findings": None,
               "edit": None}
        if protocol == "verdicts":
            later = after[doc["pair"]]
            later_spans = [tuple(s) for s in json.load(open(f"{W}confirm-new-1/runs/{system}/{later['id']}.json"))["spans"]]
            row["after_words"] = len(WORD.findall(later["text"]))
            row["after_findings"] = len(later_spans)
            first, second = flagged_texts(doc, runs[0]), flagged_texts(later, later_spans)
            old_text, new_text = " ".join(text.split()), " ".join(later["text"].split())
            raw = text.encode()
            every = {" ".join(raw[a:b].decode("utf-8", "replace").split()) for a, b in sentences(text)}
            every = {q for q in every if len(WORD.findall(q)) >= 6} | first
            row["edit"] = {"sentences": len(every), "sentences_changed": sum(q not in new_text for q in every),
                           "flagged_before": len(first), "changed_by_commit": sum(q not in new_text for q in first),
                           "flagged_after": len(second), "on_new_text": sum(q not in old_text for q in second),
                           "flagged_both_times": sum(q in old_text and q in first for q in second)}
        rows.append(row)
with gzip.GzipFile(out, "wb", mtime=0) as handle:
    for row in rows:
        handle.write((json.dumps(row, sort_keys=True) + "\n").encode())
print(len(rows), "rows")
