#!/usr/bin/env python3
"""Run the review command on generated documents, one process per document.

usage: review_product.py <workspace> <binary> <out dir> <reviewer> <confirmer|none> [--pages N] [--workers N]

reviewer: claude or astra (through ask_model.py as a local program).
confirmer: none, claude, astra (local program) or qwen (the endpoint named by
ROUTER_URL and ROUTER_MODEL, with the key in ROUTER_KEY).
Each document is sent on standard input from an empty directory, so no project
configuration applies. A finished document is not run again.
"""
import argparse, concurrent.futures, hashlib, json, os, subprocess, tempfile, time
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument("workspace")
parser.add_argument("binary")
parser.add_argument("out_dir")
parser.add_argument("reviewer", choices=["claude", "astra"])
parser.add_argument("confirmer", choices=["none", "claude", "astra", "qwen"])
parser.add_argument("--pages", type=int, default=60)
parser.add_argument("--workers", type=int, default=3)
parser.add_argument("--reverse", action="store_true", help="start from the end of the sample")
options = parser.parse_args()
W = Path(options.workspace).resolve()
out = Path(options.out_dir).resolve()
out.mkdir(parents=True, exist_ok=True)
ask = str(W / "ask_model.py")
cache = str(W / "product-cache")


def program(prefix, kind):
    return [f"--{prefix}command", "python3", f"--{prefix}command-arg", ask, f"--{prefix}command-arg", kind,
            f"--{prefix}command-arg", cache]


model = program("", options.reviewer)
if options.confirmer == "qwen":
    model += ["--confirm-endpoint", os.environ["ROUTER_URL"], "--confirm-model", os.environ["ROUTER_MODEL"],
              "--confirm-api-key-env", "ROUTER_KEY", "--confirm-max-tokens", "24000"]
elif options.confirmer != "none":
    model += program("confirm-", options.confirmer)

documents = [json.loads(line) for line in open(W / "docs.jsonl")]
documents = [d for d in documents if d["origin"] == "generated"]
documents = sorted(documents, key=lambda d: hashlib.sha256(d["doc_id"].encode()).hexdigest())[:options.pages]
if options.reverse:
    documents.reverse()


def review(doc):
    target = out / f"{doc['doc_id']}.json"
    if target.exists():
        return "cached"
    started = time.time()
    with tempfile.TemporaryDirectory() as scratch:
        args = [options.binary, "review", "--stdin", "--filename", doc["doc_id"] + ".md", "--format", "markdown",
                "--quiet", "--report", "json:" + str(target), "--request-timeout", "15m", "--timeout", "50m", *model]
        done = subprocess.run(args, input=doc["text"].encode(), cwd=scratch, capture_output=True)
    if done.returncode != 0:
        if target.exists():
            target.rename(target.with_suffix(".failed"))
        return f"FAILED exit {done.returncode}: {done.stderr.decode('utf-8', 'replace')[-200:]}"
    summary = json.loads(target.read_text())["summary"]
    return f"ok findings {summary['findings']} unconfirmed {summary['unconfirmed']} unbound {summary['unbound']} {time.time() - started:.0f}s"


with concurrent.futures.ThreadPoolExecutor(options.workers) as pool:
    for doc, outcome in zip(documents, pool.map(review, documents)):
        print(doc["doc_id"], outcome, flush=True)
