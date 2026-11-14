#!/usr/bin/env python3
"""Export the page index of the pages behind its registered reading into the record.

usage: export_page_index.py <workspace> <signals jsonl> <unswell binary> <record data file>

Rows: README files as they were before November 2022 from the test bucket of repositories, both versions of the
English de-slop pairs of the reserve and of the web-flow layer, and of the test part of the main set. A row holds
the group, an id, the side, the repository, the commit and the path, whether the repository is outside the main
set, the prose words, the count of each signal, and the index that the product reported. The binary runs over every
exported page again and must report the same counts and the same index.
"""
import concurrent.futures, gzip, hashlib, json, math, subprocess, sys, tempfile
from pathlib import Path

W, signals_path, binary, out = sys.argv[1].rstrip("/") + "/", sys.argv[2], sys.argv[3], Path(sys.argv[4])
sys.path.insert(0, W + "assisted/scripts")
from english import share

NAMES = ["em-dash", "emoji-in-heading", "emoji-at-list-item", "arrow", "stock-word", "bold-label-in-list"]


def development(repository):
    return int(hashlib.sha256(repository.encode()).hexdigest(), 16) % 10 < 6


docs, texts = {}, {}
for name, path in (("main", "assisted/deslop-files.jsonl"), ("reserve", "assisted/deslop-reserve-files.jsonl"),
                   ("web", "assisted/manual-files.jsonl")):
    for line in open(W + path):
        doc = json.loads(line)
        if not doc.get("skip"):
            docs[(name, doc["id"])] = doc
# The main set as measured: pairs with a strict message and no assistant trailer, which are the ones the signals hold.
main_repositories = {json.loads(line)["repository"] for line in open(signals_path) if json.loads(line)["set"] == "deslop"}
human = {}
for path in sorted(Path(W + "prellm").glob("readmes*.jsonl")):
    for line in open(path):
        doc = json.loads(line)
        human.setdefault(doc["repository"], doc)

measured = {}
for line in open(signals_path):
    row = json.loads(line)
    if row.get("failed") or not row.get("ai_sounding") or row["ai_sounding"]["status"] != "available":
        continue
    group = "main" if row["set"] == "deslop" else row["set"]
    measured[(group, row["id"], row["side"])] = row

rows = []
for (group, identifier, side), row in sorted(measured.items()):
    index = row["ai_sounding"]
    entry = {"group": group, "id": identifier, "side": side, "repository": row["repository"],
             "words": index["prose_words"], "counts": {s["name"]: s["count"] for s in index["signals"]}, "index": index["index"]}
    if group == "human":
        if development(row["repository"]):
            continue
        source = human[row["repository"]]
        entry.update(id=hashlib.sha256(row["repository"].encode()).hexdigest()[:16], commit=source["commit"], path=source["path"],
                     outside_main=row["repository"] not in main_repositories)
        texts[(group, entry["id"], side)] = source["text"]
    else:
        doc = docs.get((group, identifier))
        other = "after" if side == "before" else "before"
        if doc is None or (group, identifier, other) not in measured or share(doc["text"])[0] < 0.10:
            continue
        if group == "main" and development(row["repository"]):
            continue
        entry.update(commit=doc["commit"], path=doc["path"], outside_main=row["repository"] not in main_repositories)
        texts[(group, identifier, side)] = doc["text"] if side == "before" else doc["edited"]
    rows.append(entry)


def check(entry):
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--no-gate", "--allow-empty",
                               "--timeout", "2m", "--report", "json:-"],
                              input=texts[(entry["group"], entry["id"], entry["side"])].encode(), cwd=scratch, capture_output=True)
    index = json.loads(done.stdout)["documents"][0]["ai_sounding"]
    same = (index["index"] == entry["index"] and index["prose_words"] == entry["words"]
            and {s["name"]: s["count"] for s in index["signals"]} == entry["counts"]
            and len(index.get("marks", [])) == min(sum(entry["counts"].values()), 200))
    return entry["id"] if not same else None


with concurrent.futures.ThreadPoolExecutor(6) as pool:
    different = [name for name in pool.map(check, rows) if name]
assert not different, different[:5]
with gzip.GzipFile(out, "wb", mtime=0) as handle:
    for entry in rows:
        handle.write((json.dumps(entry, sort_keys=True, ensure_ascii=False) + "\n").encode())
groups = {}
for entry in rows:
    groups[(entry["group"], entry["side"])] = groups.get((entry["group"], entry["side"]), 0) + 1
print("rows", len(rows), groups)
