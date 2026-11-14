#!/usr/bin/env python3
"""Run a build of unswell over pages and store the page index and its signal counts.

usage: collect_signals.py <workspace> <unswell binary> <out jsonl>

Pages: both versions of every de-slop pair (strict message, no assistant trailer) from the main set, the reserve, and
the web-flow layer, and README files as they were before November 2022. A row holds the set, an id, the side, the
repository, and the "ai_sounding" object of the report. A finished page is not run again.
"""
import concurrent.futures, json, subprocess, sys, tempfile
from pathlib import Path

W, binary, out = sys.argv[1].rstrip("/") + "/", sys.argv[2], Path(sys.argv[3])
sys.path.insert(0, W + "assisted/scripts")
from strict import strict

done = set()
if out.exists():
    for line in out.read_text().splitlines():
        row = json.loads(line)
        done.add((row["set"], row["id"], row["side"]))


def pages():
    for name, path in (("deslop", "assisted/deslop-files.jsonl"), ("reserve", "assisted/deslop-reserve-files.jsonl"),
                       ("web", "assisted/manual-files.jsonl")):
        for line in open(W + path):
            doc = json.loads(line)
            if doc.get("skip") or doc["assistant_trailer"] or not strict(doc["message"]):
                continue
            yield name, doc["id"], "before", doc["repository"], doc["text"]
            yield name, doc["id"], "after", doc["repository"], doc["edited"]
    seen = set()
    for path in sorted(Path(W + "prellm").glob("readmes*.jsonl")):
        for line in open(path):
            doc = json.loads(line)
            if doc["repository"] in seen:
                continue
            seen.add(doc["repository"])
            yield "human", doc["repository"], "page", doc["repository"], doc["text"]


def run(page):
    name, identifier, side, repository, text = page
    with tempfile.TemporaryDirectory() as scratch:
        done_ = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--no-gate", "--allow-empty",
                                "--timeout", "2m", "--report", "json:-"], input=text.encode(), cwd=scratch, capture_output=True)
    row = {"set": name, "id": identifier, "side": side, "repository": repository}
    try:
        row["ai_sounding"] = json.loads(done_.stdout)["documents"][0].get("ai_sounding")
    except (json.JSONDecodeError, KeyError, IndexError):
        row["failed"] = True
    return row


todo = [page for page in pages() if (page[0], page[1], page[2]) not in done]
with out.open("a") as handle, concurrent.futures.ThreadPoolExecutor(8) as pool:
    for number, row in enumerate(pool.map(run, todo), 1):
        handle.write(json.dumps(row) + "\n")
        if number % 500 == 0:
            handle.flush()
            print(f"{number} of {len(todo)}", flush=True)
print(f"done: {len(todo)} pages")
