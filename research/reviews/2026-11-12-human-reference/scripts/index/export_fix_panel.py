#!/usr/bin/env python3
"""Export the fix experiment into the record.

usage: export_fix_panel.py <workspace> <unswell binary> <arms jsonl> <record data file> [first:]<judge name>=<judge jsonl>...

A judge file with the prefix "first:" comes from the registered run, when the fix command still wrote a comma between
two full clauses. Only its comparisons of that version are taken, under the name "fix-comma"; the text of that version
is the "comma_fix" of a row of the arms file.

A row holds the page, its repository, commit and path, the layer it comes from, the index of the original, for each
fixed version its index and the share of tokens that the fix changed (null for a failed fix), and the answers of the
judges per comparison. An answer names the version that reads more like text written by a language model: "first",
"second", or "same". No page text and no model output is exported.
"""
import difflib, gzip, json, re, subprocess, sys, tempfile
from pathlib import Path

W, binary, arms_path, out = sys.argv[1].rstrip("/") + "/", sys.argv[2], sys.argv[3], Path(sys.argv[4])
judges = [(argument.split("=", 1)[0].removeprefix("first:"), argument.split("=", 1)[1], argument.startswith("first:"))
          for argument in sys.argv[5:]]
# The names of the versions in the run and in the record: the run calls the product's own fix "mechanical" and the
# model that follows unswell's report "unswell".
NAMES = {"mechanical": "fix", "unswell": "report", "google": "google", "microsoft": "microsoft", "write-good": "write-good",
         "original": "original", "fix-comma": "fix-comma"}
ARMS = ("mechanical", "unswell", "google", "microsoft", "write-good")
TOKEN = re.compile(r"\w+|[^\w\s]")
docs = {}
for layer, path in (("reserve", "assisted/deslop-reserve-files.jsonl"), ("web", "assisted/manual-files.jsonl")):
    for line in open(W + path):
        doc = json.loads(line)
        if not doc.get("skip"):
            docs.setdefault(doc["id"], (layer, doc))


def index(text):
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--no-gate", "--allow-empty",
                               "--report", "json:-"], input=text.encode(), cwd=scratch, capture_output=True)
    value = json.loads(done.stdout)["documents"][0].get("ai_sounding")
    return value["index"] if value and value["status"] == "available" else None


answers = {}
for name, path, first_run in judges:
    names = {**NAMES, "mechanical": "fix-comma"} if first_run else NAMES
    for line in open(path):
        row = json.loads(line)
        page, first, second = row["id"].split("|")
        if row.get("failed") or (first_run and "mechanical" not in (first, second)):
            continue
        answers.setdefault(page, {}).setdefault(f"{names[first]}|{names[second]}", {})[name] = row["picked"]
rows = []
for line in open(arms_path):
    row = json.loads(line)
    layer, doc = docs[row["id"]]
    # The run read its pages as text, which turns a CRLF line end into LF.
    assert doc["text"].replace("\r\n", "\n") == row["original"], row["id"]
    before = TOKEN.findall(row["original"])
    entry = {"page": row["id"], "repository": doc["repository"], "commit": doc["commit"], "path": doc["path"], "layer": layer,
             "index": index(row["original"]), "arms": {}, "answers": answers.get(row["id"], {})}
    versions = {**row["arms"], "fix-comma": row["comma_fix"]}
    for name in ("fix-comma", *ARMS):
        text = versions.get(name)
        value = index(text) if text is not None else None
        # A fix that leaves fewer than 150 prose words has no index and counts as failed.
        if value is None:
            entry["arms"][NAMES[name]] = None
            entry["answers"] = {key: value_ for key, value_ in entry["answers"].items() if NAMES[name] not in key.split("|")}
            continue
        changed = 1 - difflib.SequenceMatcher(None, before, TOKEN.findall(text), autojunk=False).ratio()
        entry["arms"][NAMES[name]] = {"index": value, "changed": round(changed, 4)}
    rows.append(entry)
with gzip.GzipFile(out, "wb", mtime=0) as handle:
    for entry in sorted(rows, key=lambda item: item["page"]):
        handle.write((json.dumps(entry, sort_keys=True) + "\n").encode())
print("pages", len(rows))
