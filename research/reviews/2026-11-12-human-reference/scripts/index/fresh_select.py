#!/usr/bin/env python3
"""Select the pages of the fresh fix run from newly fetched de-slop pairs.

usage: fresh_select.py <workspace> <files jsonl> <unswell binary> <out jsonl>

A page is the version before the commit of a pair whose commit message states the removal of AI-sounding wording and
carries no assistant trailer, in English, of 150 to 1,500 words, with a page index of 40 or more as the given binary
reports it, from a repository that no earlier set holds; one page per repository, the first by id.
"""
import json, re, subprocess, sys, tempfile
from pathlib import Path

W, files, binary, out = sys.argv[1].rstrip("/") + "/", sys.argv[2], sys.argv[3], Path(sys.argv[4])
sys.path.insert(0, W + "assisted/scripts")
from english import share
from strict import strict

WORD = re.compile(r"[A-Za-z][A-Za-z'’-]*")
known = set()
for name in ("deslop-files.jsonl", "deslop-reserve-files.jsonl", "manual-files.jsonl", "deslop-patches.jsonl", "deslop-reserve-patches.jsonl",
             "manual-patches.jsonl", "manual-kept.jsonl"):
    for line in open(W + "assisted/" + name):
        known.add(json.loads(line).get("repository"))


def index(text):
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--no-gate", "--allow-empty",
                               "--timeout", "2m", "--report", "json:-"], input=text.encode(), cwd=scratch, capture_output=True)
    try:
        value = json.loads(done.stdout)["documents"][0].get("ai_sounding")
    except (json.JSONDecodeError, KeyError, IndexError):
        return None
    return value["index"] if value and value["status"] == "available" else None


counts = {"rows": 0, "not a pair": 0, "known repository": 0, "message or trailer": 0, "not English": 0, "length": 0, "no index": 0,
          "index below 40": 0, "second page of a repository": 0}
picked, seen = [], set()
for doc in sorted((json.loads(line) for line in open(files)), key=lambda item: item["id"]):
    counts["rows"] += 1
    if doc.get("skip") or not doc.get("text") or not doc.get("edited"):
        counts["not a pair"] += 1
    elif doc["repository"] in known:
        counts["known repository"] += 1
    elif doc["assistant_trailer"] or not strict(doc["message"]):
        counts["message or trailer"] += 1
    elif share(doc["text"])[0] < 0.10:
        counts["not English"] += 1
    elif not 150 <= len(WORD.findall(doc["text"])) <= 1500:
        counts["length"] += 1
    elif doc["repository"] in seen:
        counts["second page of a repository"] += 1
    else:
        value = index(doc["text"])
        if value is None:
            counts["no index"] += 1
        elif value < 40:
            counts["index below 40"] += 1
        else:
            seen.add(doc["repository"])
            picked.append({"id": doc["id"], "text": doc["text"], "index": value, "repository": doc["repository"], "commit": doc["commit"],
                           "path": doc["path"], "web_edit": bool(doc.get("web_edit"))})
with out.open("w") as handle:
    for row in picked:
        handle.write(json.dumps(row, ensure_ascii=False) + "\n")
print(counts)
print("picked", len(picked), "of them at high", sum(row["index"] >= 70 for row in picked))
