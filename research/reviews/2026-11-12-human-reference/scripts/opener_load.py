#!/usr/bin/env python3
"""Count the findings of scaffold.chat-preamble on documentation pages outside the human-reference sets.

usage: opener_load.py <corpus workspace> <unswell binary> <out json> [--workers N]

Runs the default rules on the real pages and on the generated pages of the
assistant-panel corpus, one process per page on standard input, and writes for
each part the number of pages, the number of findings of the rule, and where
each finding lies: at the very start of the page or after leading headings.
"""
import argparse, concurrent.futures, glob, json, subprocess, sys, tempfile
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument("workspace")
parser.add_argument("binary")
parser.add_argument("out")
parser.add_argument("--workers", type=int, default=6)
options = parser.parse_args()
W = Path(options.workspace).resolve()


def pages():
    for path in sorted(glob.glob(str(W / "docs-real*.jsonl"))):
        for line in open(path):
            row = json.loads(line)
            yield "real", row["doc_id"], row["text"]
    for line in open(W / "docs.jsonl"):
        row = json.loads(line)
        if row.get("origin") == "generated":
            yield "generated", row["doc_id"], row["text"]


def run(page):
    part, name, text = page
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([options.binary, "check", "--stdin", "--filename", name + ".md", "--format", "markdown",
                               "--no-gate", "--allow-empty", "--report", "json:-"],
                              input=text.encode(), cwd=scratch, capture_output=True)
    if done.returncode != 0:
        return part, name, None
    report = json.loads(done.stdout)
    raw = text.encode()
    found = []
    for finding in report.get("findings") or []:
        if finding["rule_id"] == "scaffold.chat-preamble":
            span = finding["primary"]["span"]
            found.append({"start": span["start"], "phrase": raw[span["start"]:span["end"]].decode().lower(),
                          "after_heading": span["start"] > 0 and raw[:span["start"]].strip() != b""})
    return part, name, found


result = {}
with concurrent.futures.ThreadPoolExecutor(options.workers) as pool:
    for part, name, found in pool.map(run, list(pages())):
        entry = result.setdefault(part, {"pages": 0, "failed": 0, "findings": []})
        entry["pages"] += 1
        if found is None:
            entry["failed"] += 1
        else:
            entry["findings"] += [{"page": name, **item} for item in found]
Path(options.out).write_text(json.dumps(result, indent=1, sort_keys=True) + "\n")
for part, entry in result.items():
    print(part, "pages", entry["pages"], "failed", entry["failed"], "findings", len(entry["findings"]))
