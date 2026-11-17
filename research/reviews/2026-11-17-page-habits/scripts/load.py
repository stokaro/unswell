#!/usr/bin/env python3
"""Run the two habit rules on each page of the corpora and keep, for each page, what each rule reported.

usage: load.py <workspace> <unswell binary> <out jsonl> <part: dev|test|none|all> [--workers N] [--rules id,id]

A row holds the corpus, the id and the split of the page, its prose words, and for each rule the number of places
that its finding lists (0 when the rule is quiet). No text is kept.
"""
import concurrent.futures, glob, json, re, subprocess, sys, tempfile
from pathlib import Path

W, binary, out, split = Path(sys.argv[1]), sys.argv[2], Path(sys.argv[3]), sys.argv[4]
option = lambda name, default: sys.argv[sys.argv.index(name) + 1] if name in sys.argv else default
workers = int(option("--workers", "8"))
RULES = option("--rules", "syntax.closing-denial-density,syntax.uniform-register").split(",")
POLICY = "version: 1\nextends: [builtin:custom]\nrules:\n" + "".join(f"  {rule}:\n    enabled: true\n" for rule in RULES)


def rows(path, corpus, keep=lambda row: True, part=lambda row: row.get("split"), prefix=""):
    for number, line in enumerate(open(path)):
        row = json.loads(line)
        if "id" in row:
            row["id"] = prefix + str(row["id"])
        if keep(row) and "text" in row:
            yield {"corpus": corpus, "id": str(row.get("id", f"{Path(str(path)).stem}-{number}")), "split": part(row), "text": row["text"]}


def corpora():
    G = W / "guided"
    yield from rows(G / "set.jsonl", "posts of set S")
    yield from rows(G / "human.jsonl", "blog posts of people")
    yield from rows(G / "styled-articles.jsonl", "articles of people under a style guide")
    yield from rows(W / "human" / "docs.jsonl", "documentation pages of people", lambda row: row.get("set") == "prs")
    for name, corpus in (("dev-gpt-guided.jsonl", "model texts, guided, first writer"), ("dev-qwen-guided.jsonl", "model texts, guided, second writer"),
                         ("dev-gpt-plain.jsonl", "model texts, plain, first writer"), ("dev-qwen-plain.jsonl", "model texts, plain, second writer")):
        yield from rows(G / name, corpus)
    # Texts that two models wrote under a style guide and that no fitting has used.
    for stem in ("fresh", "fresh2", "fresh3", "fresh5"):
        # The sets hold texts on the same outlines, so the name of the set is part of the id of a text.
        yield from rows(G / f"{stem}-gpt-guided.jsonl", "fresh model texts, guided, first writer", prefix=stem + ":")
        yield from rows(G / f"{stem}-qwen-guided.jsonl", "fresh model texts, guided, second writer", prefix=stem + ":")
    # Corpora without parts. They are read whole and carry the part "none".
    for path in sorted(glob.glob(str(W / "docs-real*.jsonl"))):
        yield from rows(path, "pages of public repositories today", part=lambda row: "none")
    yield from rows(W / "docs.jsonl", "pages of the panel corpus that models wrote or polished", lambda row: row.get("origin") == "generated",
                    part=lambda row: "none")


def run(page):
    with tempfile.TemporaryDirectory() as scratch:
        Path(scratch, "policy.yaml").write_text(POLICY)
        done = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--config", "policy.yaml",
                               "--no-gate", "--allow-empty", "--timeout", "2m", "--report", "json:-"],
                              input=page["text"].encode(), cwd=scratch, capture_output=True)
    row = {key: page[key] for key in ("corpus", "id", "split")}
    row["words"] = len(re.findall(r"[A-Za-z][A-Za-z'’-]*", page["text"]))
    try:
        report = json.loads(done.stdout)
    except ValueError:
        row["error"] = done.stderr.decode()[-200:]
        return row
    findings = report.get("findings") or []
    for rule in RULES:
        places = 0
        for finding in findings:
            if finding.get("rule_id") == rule:
                places += 1 + len(finding.get("related") or [])
        row[rule] = places
    return row


pages = [p for p in corpora() if split == "all" or p["split"] == split]
with out.open("w") as handle, concurrent.futures.ThreadPoolExecutor(workers) as pool:
    for row in pool.map(run, pages):
        handle.write(json.dumps(row) + "\n")
print("pages", len(pages))
