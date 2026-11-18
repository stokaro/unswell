#!/usr/bin/env python3
"""Compare two builds of syntax.closing-denial-density on the pages of one part.

usage: forms.py <workspace> <old binary> <new binary> <out jsonl> <part: dev|test|none> [--workers N]

A row holds the corpus, the id and the split of a page, and for each build the number of closing sentences that the
rule lists with its limits as they ship (0 when it is quiet) and the sentences that it counts with no limit. The
sentences are text of the pages: the file stays in the workspace.
"""
import concurrent.futures, glob, json, subprocess, sys, tempfile
from pathlib import Path

W, old, new, out, split = Path(sys.argv[1]), sys.argv[2], sys.argv[3], Path(sys.argv[4]), sys.argv[5]
option = lambda name, default: sys.argv[sys.argv.index(name) + 1] if name in sys.argv else default
workers = int(option("--workers", "8"))
RULE = "syntax.closing-denial-density"
SHIPPED = f"version: 1\nextends: [builtin:custom]\nrules:\n  {RULE}:\n    enabled: true\n"
EVERY = SHIPPED + "    parameters: {allowed_occurrences: 0, onset: 0}\n"


def rows(path, corpus, keep=lambda row: True, part=lambda row: row.get("split"), prefix=""):
    for number, line in enumerate(open(path)):
        row = json.loads(line)
        if "id" in row:
            row["id"] = prefix + str(row["id"])
        if keep(row) and "text" in row and part(row) == split:
            yield {"corpus": corpus, "id": str(row.get("id", f"{Path(str(path)).stem}-{number}")), "split": split, "text": row["text"]}


def corpora():
    G = W / "guided"
    yield from rows(G / "set.jsonl", "posts of set S")
    yield from rows(G / "human.jsonl", "blog posts of people")
    yield from rows(G / "styled-articles.jsonl", "articles of people under a style guide")
    yield from rows(W / "human" / "docs.jsonl", "documentation pages of people", lambda row: row.get("set") == "prs")
    for name, corpus in (("dev-gpt-guided.jsonl", "model texts, guided, first writer"), ("dev-qwen-guided.jsonl", "model texts, guided, second writer"),
                         ("dev-gpt-plain.jsonl", "model texts, plain, first writer"), ("dev-qwen-plain.jsonl", "model texts, plain, second writer")):
        yield from rows(G / name, corpus)
    for stem in ("fresh", "fresh2", "fresh3", "fresh5"):
        yield from rows(G / f"{stem}-gpt-guided.jsonl", "fresh model texts, guided, first writer", prefix=stem + ":")
        yield from rows(G / f"{stem}-qwen-guided.jsonl", "fresh model texts, guided, second writer", prefix=stem + ":")
    # Corpora without parts. They are read whole and carry the part "none".
    for path in sorted(glob.glob(str(W / "docs-real*.jsonl"))):
        yield from rows(path, "pages of public repositories today", part=lambda row: "none")
    yield from rows(W / "docs.jsonl", "pages of the panel corpus that models wrote or polished", lambda row: row.get("origin") == "generated",
                    part=lambda row: "none")


def places(binary, policy, text):
    """The sentences that the rule lists, or None when the run failed."""
    with tempfile.TemporaryDirectory() as scratch:
        Path(scratch, "policy.yaml").write_text(policy)
        done = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--config", "policy.yaml",
                               "--no-gate", "--allow-empty", "--timeout", "2m", "--report", "json:-"],
                              input=text.encode(), cwd=scratch, capture_output=True)
    try:
        report = json.loads(done.stdout)
    except ValueError:
        return None
    source, found = text.encode(), []
    for finding in report.get("findings") or []:
        if finding.get("rule_id") == RULE:
            for location in [finding["primary"]] + (finding.get("related") or []):
                span = location["span"]
                found.append(source[span["start"]:span["end"]].decode(errors="replace"))
    return found


def run(page):
    row = {key: page[key] for key in ("corpus", "id", "split")}
    for name, binary in (("old", old), ("new", new)):
        shipped, every = places(binary, SHIPPED, page["text"]), places(binary, EVERY, page["text"])
        if shipped is None or every is None:
            row["error"] = name
            return row
        row[name] = len(shipped)
        row[name + "_sentences"] = every
    return row


pages = list(corpora())
with out.open("w") as handle, concurrent.futures.ThreadPoolExecutor(workers) as pool:
    for row in pool.map(run, pages):
        handle.write(json.dumps(row) + "\n")
print("pages", len(pages))
