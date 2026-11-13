#!/usr/bin/env python3
"""Run every rule on the text before and after each de-slop commit and count findings per rule.

usage: score_pairs.py <files jsonl> <unswell binary> <config yaml> <out jsonl> [--workers N]

One process per text on standard input. A row holds, for both versions, the
number of words and the findings per rule, and counts of typographic marks.
"""
import argparse, collections, concurrent.futures, json, re, subprocess, sys, tempfile
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument("files")
parser.add_argument("binary")
parser.add_argument("config")
parser.add_argument("out")
parser.add_argument("--workers", type=int, default=8)
parser.add_argument("--part", choices=("development", "test"))
options = parser.parse_args()
out = Path(options.out)
done = {json.loads(line)["id"] for line in out.read_text().splitlines()} if out.exists() else set()
CONFIG = Path(options.config).read_text()
EMOJI = "[\U0001F300-\U0001FAFF☀-➿⭐⬆⬇✅❌✨⚡]"
MARKS = {"mark.emoji": re.compile(EMOJI), "mark.em-dash": re.compile("—"), "mark.bold": re.compile(r"\*\*[^*\n]+\*\*"),
         "mark.arrow": re.compile("→|⇒|(?<![-<=])[-=]>"), "mark.exclamation": re.compile(r"(?<=\w)!(?=\s|$)"),
         "mark.heading-emoji": re.compile(r"(?m)^#{1,6}[^\n]*?" + EMOJI)}


def prose(text):
    text = re.sub(r"(?ms)^(```|~~~).*?^\1[^\n]*$", " ", text)
    return re.sub(r"`[^`\n]*`", " ", text)


def measure(text):
    with tempfile.TemporaryDirectory() as scratch:
        Path(scratch, "policy.yaml").write_text(CONFIG)
        result = subprocess.run([options.binary, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--config", "policy.yaml",
                                 "--no-gate", "--allow-empty", "--timeout", "2m", "--report", "json:-"],
                                input=text.encode(), cwd=scratch, capture_output=True)
    if result.returncode != 0:
        return None
    counts = collections.Counter()
    for finding in json.loads(result.stdout).get("findings") or []:
        # The document-level em-dash rule reports once; its measure is the em dashes it lists.
        counts[finding["rule_id"]] += 1 + len(finding.get("related") or []) if finding["rule_id"] == "format.em-dash-frequency" else 1
    plain = prose(text)
    for name, pattern in MARKS.items():
        counts[name] = len(pattern.findall(plain))
    return {"words": len(re.findall(r"[A-Za-z][A-Za-z'’-]*", plain)), "counts": dict(counts)}


def work(doc):
    before, after = measure(doc["text"]), measure(doc["edited"])
    if before is None or after is None:
        return {"id": doc["id"], "skip": True}
    return {"id": doc["id"], "repository": doc["repository"], "web_edit": doc["web_edit"], "assistant_trailer": doc["assistant_trailer"],
            "query": doc["query"], "before": before, "after": after}


import hashlib


def in_part(doc):
    development = int(hashlib.sha256(doc["repository"].encode()).hexdigest(), 16) % 10 < 6
    return options.part is None or development == (options.part == "development")


docs = [d for d in map(json.loads, open(options.files)) if not d.get("skip") and d["id"] not in done and in_part(d)]
with out.open("a") as handle, concurrent.futures.ThreadPoolExecutor(options.workers) as pool:
    for number, row in enumerate(pool.map(work, docs), 1):
        handle.write(json.dumps(row) + "\n")
        if number % 200 == 0:
            handle.flush()
            print(f"{number} of {len(docs)}", flush=True)
print(f"done: {len(docs)} documents")
