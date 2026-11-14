#!/usr/bin/env python3
"""Export the run of the released binary on new pages into the record.

usage: export_fix_fresh.py <unswell binary> <sample jsonl> <arms jsonl> <record data file> [--versions a,b] <judge name>=<judge jsonl>...

--versions names the versions of the run as the run calls them; the default is mechanical,unswell,write-good.

A row holds the page, its repository, commit and path, the layer ("web" for a commit made through GitHub's web flow,
"search" otherwise), the index of the original, for each fixed version its index and the share of tokens that the fix
changed (null for a failed fix), and the answers of the judges per comparison. An answer names the version that reads
more like text written by a language model: "first", "second", or "same". No page text and no model output is exported.
"""
import difflib, gzip, json, re, subprocess, sys, tempfile
from pathlib import Path

binary, sample_path, arms_path, out = sys.argv[1], sys.argv[2], sys.argv[3], Path(sys.argv[4])
rest = sys.argv[5:]
versions = ("mechanical", "unswell", "write-good")
if rest and rest[0] == "--versions":
    versions, rest = tuple(rest[1].split(",")), rest[2:]
judges = dict(argument.split("=", 1) for argument in rest)
# The names of the versions in the run and in the record.
NAMES = {"mechanical": "fix", "unswell": "report", "write-good": "write-good", "original": "original"}
TOKEN = re.compile(r"\w+|[^\w\s]")
sample = {json.loads(line)["id"]: json.loads(line) for line in open(sample_path)}


def index(text):
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--no-gate", "--allow-empty",
                               "--report", "json:-"], input=text.encode(), cwd=scratch, capture_output=True)
    value = json.loads(done.stdout)["documents"][0].get("ai_sounding")
    return value["index"] if value and value["status"] == "available" else None


answers = {}
for name, path in judges.items():
    for line in open(path):
        row = json.loads(line)
        page, first, second = row["id"].split("|")
        if row.get("failed") or not {first, second} <= {"original", *versions}:
            continue
        answers.setdefault(page, {}).setdefault(f"{NAMES[first]}|{NAMES[second]}", {})[name] = row["picked"]
rows = []
for line in open(arms_path):
    row = json.loads(line)
    page = sample[row["id"]]
    assert page["text"] == row["original"], row["id"]
    before = TOKEN.findall(row["original"])
    entry = {"page": row["id"], "repository": page["repository"], "commit": page["commit"], "path": page["path"],
             "layer": "web" if page["web_edit"] else "search", "index": index(row["original"]), "arms": {},
             "answers": answers.get(row["id"], {})}
    for name in versions:
        text = row["arms"].get(name)
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
