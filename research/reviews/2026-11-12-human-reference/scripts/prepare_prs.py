#!/usr/bin/env python3
"""Add the harvested pull requests to docs.jsonl with a split and the engine's prose sentences.

usage: prepare_prs.py <human directory> <unswell binary>

Repositories are ordered by the SHA-256 of their name: the first 60 are
development, the next 240 at most are the test. The unit of measurement is a
prose sentence as the engine extracts it; a document without an edited prose
sentence is dropped, because its pull request changed only code, tables, or
markup.
"""
import collections, hashlib, json, subprocess, sys, tempfile
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from common import overlaps, read_jsonl, write_jsonl

ROOT, BINARY = Path(sys.argv[1]), sys.argv[2]


def units(doc):
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([BINARY, "check", "--stdin", "--filename", doc["id"] + ".md", "--format", "markdown", "--no-gate",
                               "--allow-empty", "--prepared-feature", "prose-words", "--prepared-kind", "sentence",
                               "--report", "json:-"], input=doc["text"].encode(), cwd=scratch, capture_output=True)
    if done.returncode != 0:
        return None
    sources = (json.loads(done.stdout).get("prepared_features") or {}).get("sources") or []
    found = set()
    for unit in sources[0]["units"] if sources else []:
        if unit["binding"]["kind"] == "sentence":
            segments = unit["binding"]["segments"]
            found.add((segments[0]["start"], segments[-1]["end"]))
    return [list(span) for span in sorted(found)]


kept = read_jsonl(ROOT / "raw/prs/kept.jsonl")
kept.sort(key=lambda d: hashlib.sha256(d["repo"].lower().encode()).hexdigest())
stats, rows = collections.Counter(), []
for doc in kept:
    spans = units(doc)
    if spans is None:
        stats["engine could not read the file"] += 1
        continue
    reference = [entry["span"] for entry in doc["reference"]]
    edited = sum(overlaps(span, reference) for span in spans)
    if not edited:
        stats["no edited prose sentence"] += 1
        continue
    doc["units"] = spans
    doc["split"] = "dev" if len(rows) < 60 else "test"
    rows.append(doc)
    stats["sentences"] += len(spans)
    stats["edited sentences"] += edited
    if len(rows) == 300:
        break
others = [d for d in read_jsonl(ROOT / "docs.jsonl") if d["set"] != "prs"]
write_jsonl(ROOT / "docs.jsonl", others + rows)
print(dict(stats), collections.Counter(d["split"] for d in rows), "kinds", collections.Counter(d["kind"] for d in rows).most_common())
