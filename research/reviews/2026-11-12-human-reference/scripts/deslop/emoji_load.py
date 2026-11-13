#!/usr/bin/env python3
"""Count the findings of format.decorative-emoji on documentation corpora, by kind of mark and by place.

usage: emoji_load.py <corpus workspace> <unswell binary> <out json>

Corpora: pages of documentation pull requests that people wrote before
November 2022, pages of public repositories, pages that models generated, and
the texts before de-slop commits. A status mark is a check mark, a cross, or a
ballot box.
"""
import collections, concurrent.futures, glob, hashlib, json, re, subprocess, sys, tempfile
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from strict import strict

W, binary, out = Path(sys.argv[1]), sys.argv[2], Path(sys.argv[3])
STATUS = set("✅✔✓☑❌✗✘❎☐⬜⬛")
POLICY = "version: 1\nextends: [builtin:custom]\nrules:\n  format.decorative-emoji:\n    enabled: true\n"


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
    raw = text.encode()
    findings = json.loads(done.stdout).get("findings") or []
    counts["findings"] = len(findings)
    counts["pages_with_finding"] = bool(findings)
    for finding in findings:
        for location in [finding["primary"], *finding.get("related", [])]:
            span = location["span"]
            mark = raw[span["start"]:span["end"]].decode("utf-8", "replace")
            line = raw[raw.rfind(b"\n", 0, span["start"]) + 1:span["start"]].lstrip()
            place = "heading" if line.startswith(b"#") else "list"
            kind = "status" if mark[:1] in STATUS else "other"
            counts[f"{place}_{kind}"] += 1
        status_only = all(raw[l["span"]["start"]:l["span"]["end"]].decode("utf-8", "replace")[:1] in STATUS
                          for l in [finding["primary"], *finding.get("related", [])])
        counts["findings_status_only"] += status_only
    return name, counts


totals = collections.defaultdict(collections.Counter)
with concurrent.futures.ThreadPoolExecutor(8) as pool:
    for name, counts in pool.map(run, list(corpora())):
        totals[name].update(counts)
out.write_text(json.dumps({name: dict(counts) for name, counts in totals.items()}, indent=1, sort_keys=True) + "\n")
for name, c in totals.items():
    print(f"{name}: pages {c['pages']}, with a finding {c['pages_with_finding']} ({c['pages_with_finding'] / c['pages']:.1%}), findings {c['findings']} "
          f"({1000 * c['findings'] / c['words']:.2f} per 1,000 words), status marks only {c['findings_status_only']}; "
          f"marks: heading other {c['heading_other']}, heading status {c['heading_status']}, list other {c['list_other']}, list status {c['list_status']}")
