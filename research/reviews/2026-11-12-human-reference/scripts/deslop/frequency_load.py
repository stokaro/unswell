#!/usr/bin/env python3
"""Count the pages that format.em-dash-frequency flags on documentation corpora, and the em dashes it lists.

usage: frequency_load.py <corpus workspace> <unswell binary> <out json>

Corpora: pages of documentation pull requests that people wrote before
November 2022, pages of public repositories, pages that models generated, and
the texts before de-slop commits.
"""
import collections, concurrent.futures, glob, hashlib, json, re, subprocess, sys, tempfile
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from strict import strict

W, binary, out = Path(sys.argv[1]), sys.argv[2], Path(sys.argv[3])
POLICY = "version: 1\nextends: [builtin:custom]\nrules:\n  format.em-dash-frequency:\n    enabled: true\n"


def corpora():
    for line in open(W / "human" / "docs.jsonl"):
        row = json.loads(line)
        if row["set"] == "prs":
            yield "written by people before November 2022", row["text"]
    for path in sorted(glob.glob(str(W / "docs-real*.jsonl"))):
        for line in open(path):
            yield "pages of public repositories", json.loads(line)["text"]
    for line in open(W / "docs.jsonl"):
        row = json.loads(line)
        if row.get("origin") == "generated":
            yield "pages that models generated", row["text"]
    for line in open(W / "assisted" / "deslop-files.jsonl"):
        row = json.loads(line)
        if not row.get("skip") and not row["assistant_trailer"] and strict(row["message"]):
            yield "texts before a de-slop commit", row["text"]


def run(item):
    name, text = item
    with tempfile.TemporaryDirectory() as scratch:
        Path(scratch, "policy.yaml").write_text(POLICY)
        done = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--config", "policy.yaml",
                               "--no-gate", "--allow-empty", "--timeout", "2m", "--report", "json:-"],
                              input=text.encode(), cwd=scratch, capture_output=True)
    counts = collections.Counter(pages=1, words=len(re.findall(r"[A-Za-z][A-Za-z'’-]*", text)))
    if done.returncode != 0:
        counts["failed"] += 1
        return name, counts
    findings = json.loads(done.stdout).get("findings") or []
    counts["pages_with_finding"] = bool(findings)
    counts["findings"] = sum(1 + len(finding.get("related") or []) for finding in findings)
    return name, counts


totals = collections.defaultdict(collections.Counter)
with concurrent.futures.ThreadPoolExecutor(8) as pool:
    for name, counts in pool.map(run, list(corpora())):
        totals[name].update(counts)
out.write_text(json.dumps({name: dict(counts) for name, counts in totals.items()}, indent=1, sort_keys=True) + "\n")
for name, c in totals.items():
    print(f"{name}: pages {c['pages']}, flagged {c['pages_with_finding']} ({c['pages_with_finding'] / c['pages']:.1%}), "
          f"em dashes listed {c['findings']} ({1000 * c['findings'] / c['words']:.2f} per 1,000 words), failed {c['failed']}")
