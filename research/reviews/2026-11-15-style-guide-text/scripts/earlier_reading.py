#!/usr/bin/env python3
"""Recount the held-out reading of the earlier record of the page index with another binary.

usage: earlier_reading.py <workspace> <repository checkout> <unswell binary> <out jsonl>

The pages are the rows of the earlier record's page-index data: README files written before November 2022, and
both versions of the de-slop pairs. Each is checked again as it is. The held-out reading compares the pages
before a de-slop commit, from repositories outside the main set, with the README files.
"""
import concurrent.futures, gzip, json, math, subprocess, sys, tempfile
from pathlib import Path

W, repo, binary, out = sys.argv[1].rstrip("/") + "/", sys.argv[2], sys.argv[3], Path(sys.argv[4])
record = subprocess.run(["git", "-C", repo, "show", "origin/main:research/reviews/2026-11-12-human-reference/data/page-index.jsonl.gz"],
                        capture_output=True).stdout
rows = [json.loads(line) for line in gzip.decompress(record).decode().splitlines()]
docs = {}
for name, path in (("main", "assisted/deslop-files.jsonl"), ("reserve", "assisted/deslop-reserve-files.jsonl"), ("web", "assisted/manual-files.jsonl")):
    for line in open(W + path):
        doc = json.loads(line)
        if not doc.get("skip"):
            docs[(name, doc["id"])] = doc
human = {}
for path in sorted(Path(W + "prellm").glob("readmes*.jsonl")):
    for line in open(path):
        doc = json.loads(line)
        human.setdefault(doc["repository"], doc)


def text(row):
    if row["group"] == "human":
        return human[row["repository"]]["text"]
    doc = docs[(row["group"], row["id"])]
    return doc["text"] if row["side"] == "before" else doc["edited"]


def check(row):
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--no-gate", "--allow-empty",
                               "--timeout", "2m", "--report", "json:-"], input=text(row).encode(), cwd=scratch, capture_output=True)
    index = json.loads(done.stdout)["documents"][0]["ai_sounding"]
    return {"group": row["group"], "id": row["id"], "side": row["side"], "outside_main": row["outside_main"], "earlier": row["index"],
            "index": index["index"], "level": index.get("level"), "readings": index.get("readings")}


with concurrent.futures.ThreadPoolExecutor(6) as pool:
    done = list(pool.map(check, rows))
out.write_text("".join(json.dumps(r) + "\n" for r in done))
auc = lambda pos, neg: sum((p > q) + 0.5 * (p == q) for p in pos for q in neg) / (len(pos) * len(neg))
for field in ("earlier", "index"):
    people = [r[field] for r in done if r["group"] == "human"]
    before = [r[field] for r in done if r["group"] in ("reserve", "web") and r["side"] == "before" and r["outside_main"]]
    pairs = {}
    for r in done:
        if r["group"] in ("reserve", "web") and r["outside_main"]:
            pairs.setdefault((r["group"], r["id"]), {})[r["side"]] = r[field]
    lower = sum(p["after"] < p["before"] for p in pairs.values())
    higher = sum(p["after"] > p["before"] for p in pairs.values())
    print(f"{'earlier record' if field == 'earlier' else 'this binary':15s} held-out AUC {auc(before, people):.3f} on {len(before)} pages before and {len(people)} README files;"
          f" lower after the commit in {lower} of {lower + higher} pairs that differ ({lower / (lower + higher):.1%});"
          f" README files high {sum(v >= 70 for v in people)} of {len(people)} ({sum(v >= 70 for v in people) / len(people):.1%})")
