#!/usr/bin/env python3
"""Build the v0 document set from the long-prose-v3 study archives.

Reads three archives of the Unswell repository and writes docs.jsonl: one record
per document with its text and the items an annotator labels. An item is a
sentence, a paragraph without separate sentences, or a fragment of six or more
words, with exact byte ranges from the engine's own extraction.
"""
import collections, hashlib, io, json, re, sys, tarfile
from pathlib import Path

REPO = Path(sys.argv[1])
OUT = Path(sys.argv[2])
STUDY = REPO / "research/generation/studies/long-prose-v3"
MIN_FRAGMENT_WORDS = 6
MODEL_NAMES = {"luna": "gpt-5.6-luna", "haiku": "claude-haiku-4-5-20251001",
               "qwen": "private-host/Qwen3.6-35B-A3B"}


def members(name, prefix):
    with tarfile.open(STUDY / name) as archive:
        for member in archive:
            if member.isfile() and member.name.startswith(prefix):
                yield member.name, archive.extractfile(member).read()


def sha(data):
    return hashlib.sha256(data).hexdigest()


units_by_sha = collections.defaultdict(list)
for archive, prefix in (("measurements.tar.gz", "analysis/candidates/"),
                        ("inputs.tar.gz", "candidates/")):
    for _, data in members(archive, prefix):
        for wrapper in json.loads(data)["units"]:
            unit = wrapper["unit"]
            units_by_sha[unit["source"]["sha256"]].append((wrapper, unit))


def items_for(text_bytes):
    rows = units_by_sha.get(sha(text_bytes), [])
    spans = []
    seen = set()
    for wrapper, unit in rows:
        segments = [(s["start"], s["end"]) for s in unit["source"]["segments"]]
        key = (unit["kind"], tuple(segments))
        if key in seen:
            continue
        seen.add(key)
        spans.append({"kind": unit["kind"], "role": unit.get("role", ""), "segments": segments,
                      "start": segments[0][0], "end": segments[-1][1],
                      "text": unit["text"], "words": wrapper["words"]})
    sentences = [s for s in spans if s["kind"] == "sentence"]
    chosen = list(sentences)
    for span in spans:
        if span["kind"] == "paragraph":
            inside = any(span["start"] <= s["start"] and s["end"] <= span["end"] for s in sentences)
            if not inside:
                chosen.append(span)
        elif span["kind"] == "fragment" and span["words"] >= MIN_FRAGMENT_WORDS:
            chosen.append(span)
    chosen.sort(key=lambda s: (s["start"], s["end"]))
    for number, span in enumerate(chosen, 1):
        span["id"] = f"i{number:03d}"
    return chosen, len(rows)


tasks = {}
for name, data in members("inputs.tar.gz", "tasks.json"):
    for task in json.loads(data)["tasks"]:
        tasks[task["id"]] = task

docs = []
for task in tasks.values():
    text = task["text"].encode()
    items, total = items_for(text)
    docs.append({"doc_id": "orig-" + task["id"], "origin": "original", "family": "", "model": "",
                 "operation": "", "prompt": "", "task_id": task["id"], "repository": task["repository"],
                 "path": task["path"], "partition": task["partition"], "sha256": sha(text),
                 "text": task["text"], "items": items, "units_found": total})
for family in ("luna", "haiku", "qwen"):
    for name, data in members("outputs.tar.gz", f"runs/{family}/records.json"):
        for record in json.loads(data)["records"]:
            if record["status"] != "complete":
                continue
            task = tasks[record["task_id"]]
            stored = record["text"] + "\n"  # the study stores each response with a final newline
            text = stored.encode()
            items, total = items_for(text)
            docs.append({"doc_id": f"{family}-{record['response_id']}", "origin": "generated",
                         "family": family, "model": MODEL_NAMES[family],
                         "operation": record["operation"], "prompt": record["prompt"],
                         "task_id": record["task_id"], "repository": task["repository"],
                         "path": task["path"], "partition": task["partition"], "sha256": sha(text),
                         "text": stored, "items": items, "units_found": total})

with open(OUT, "w") as handle:
    for doc in docs:
        handle.write(json.dumps(doc, ensure_ascii=False) + "\n")

no_units = [d["doc_id"] for d in docs if not d["units_found"]]
count = collections.Counter(d["origin"] for d in docs)
items = sum(len(d["items"]) for d in docs)
words = sum(len(d["text"].split()) for d in docs)
by_partition = collections.Counter(d["partition"] for d in docs)
repos = len({d["repository"] for d in docs})
print(f"documents={len(docs)} {dict(count)} repositories={repos} words={words} items={items}")
print(f"partitions={dict(by_partition)} documents without extracted units={len(no_units)}")
kinds = collections.Counter(i["kind"] for d in docs for i in d["items"])
print(f"item kinds={dict(kinds)}")
