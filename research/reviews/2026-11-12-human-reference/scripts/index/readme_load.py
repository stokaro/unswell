#!/usr/bin/env python3
"""Count on how many README files written before November 2022 each default rule reports.

usage: readme_load.py <workspace> <unswell binary> <record page-index data> <out jsonl>

The pages are the ones of group "human" in the record's page-index data. A row holds the page id, its words, and
the rules that report on it with the default configuration.
"""
import concurrent.futures, gzip, hashlib, json, subprocess, sys, tempfile
from pathlib import Path

W, binary, record, out = sys.argv[1].rstrip("/") + "/", sys.argv[2], sys.argv[3], Path(sys.argv[4])
wanted = {row["repository"]: row for row in map(json.loads, gzip.open(record, "rt")) if row["group"] == "human"}
texts = {}
for path in sorted(Path(W + "prellm").glob("readmes*.jsonl")):
    for line in open(path):
        doc = json.loads(line)
        if doc["repository"] in wanted:
            texts.setdefault(doc["repository"], doc["text"])
assert len(texts) == len(wanted), (len(texts), len(wanted))


def run(repository):
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--no-gate", "--allow-empty",
                               "--timeout", "2m", "--report", "json:-"], input=texts[repository].encode(), cwd=scratch, capture_output=True)
    result = json.loads(done.stdout)
    rules = sorted({finding["rule_id"] for finding in result["findings"] if not finding.get("suppressed")})
    return {"page": wanted[repository]["id"], "words": result["documents"][0]["prose_words"], "rules": rules}


with concurrent.futures.ThreadPoolExecutor(6) as pool, out.open("w") as handle:
    for row in pool.map(run, sorted(wanted)):
        handle.write(json.dumps(row, sort_keys=True) + "\n")
counts = {}
rows = [json.loads(line) for line in out.open()]
for row in rows:
    for name in row["rules"]:
        counts[name] = counts.get(name, 0) + 1
print(len(rows), "pages; with any finding", sum(bool(row["rules"]) for row in rows))
for name, count in sorted(counts.items(), key=lambda item: -item[1]):
    print(f"{name}: {count} ({100 * count / len(rows):.1f}%)")
