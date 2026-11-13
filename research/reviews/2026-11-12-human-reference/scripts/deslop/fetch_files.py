#!/usr/bin/env python3
"""Fetch the Markdown files of de-slop commits before and after the commit, and locate what was changed.

usage: fetch_files.py <patches jsonl> <out jsonl> [--per-commit N]

The files come from raw.githubusercontent.com by commit, so no API budget is
spent. For each commit the Markdown files with the most removed lines are
taken. A row holds the text before, the text after, and the byte spans of the
words in the text before that the commit removed or replaced, found by a line
diff and then a word diff inside changed blocks.
"""
import argparse, concurrent.futures, difflib, hashlib, json, re, sys, time, urllib.error, urllib.parse, urllib.request
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument("patches")
parser.add_argument("out")
parser.add_argument("--per-commit", type=int, default=2)
options = parser.parse_args()
out = Path(options.out)
done = {json.loads(line)["id"] for line in out.read_text().splitlines()} if out.exists() else set()
MAX_BYTES = 200_000


def raw(repository, sha, path):
    url = f"https://raw.githubusercontent.com/{repository}/{sha}/{urllib.parse.quote(path)}"
    for attempt in range(3):
        try:
            with urllib.request.urlopen(urllib.request.Request(url, headers={"User-Agent": "research-client"}), timeout=60) as response:
                data = response.read(MAX_BYTES + 1)
            return None if len(data) > MAX_BYTES else data.decode("utf-8", "replace")
        except urllib.error.HTTPError as error:
            if error.code == 404:
                return None
            time.sleep(5 * (attempt + 1))
        except Exception:  # noqa: BLE001
            time.sleep(5 * (attempt + 1))
    return None


def reference(before, after):
    """Byte spans of words in before that are gone or replaced in after."""
    old, new = before.splitlines(keepends=True), after.splitlines(keepends=True)
    offsets, position = [], 0
    for line in old:
        offsets.append(position)
        position += len(line.encode())
    spans = []
    for tag, a0, a1, b0, b1 in difflib.SequenceMatcher(a=old, b=new, autojunk=False).get_opcodes():
        if tag not in ("replace", "delete"):
            continue
        chunk, other = "".join(old[a0:a1]), "".join(new[b0:b1]) if tag == "replace" else ""
        tokens = [(m.group(), m.start(), m.end()) for m in re.finditer(r"\S+", chunk)]
        matcher = difflib.SequenceMatcher(a=[t[0] for t in tokens], b=re.findall(r"\S+", other), autojunk=False)
        for kind, t0, t1, _, _ in matcher.get_opcodes():
            if kind in ("replace", "delete") and t1 > t0:
                begin = offsets[a0] + len(chunk[:tokens[t0][1]].encode())
                end = offsets[a0] + len(chunk[:tokens[t1 - 1][2]].encode())
                spans.append({"span": [begin, end], "category": kind})
    return spans


def work(job):
    row, entry = job
    identifier = hashlib.sha256(f"{row['repository']}|{row['commit']}|{entry['path']}".encode()).hexdigest()[:16]
    if identifier in done:
        return None
    before = raw(row["repository"], row["parent"], entry["path"]) if row.get("parent") else None
    after = raw(row["repository"], row["commit"], entry["path"])
    if not before or after is None or not before.strip():
        return {"id": identifier, "skip": True}
    spans = reference(before, after)
    if not spans:
        return {"id": identifier, "skip": True}
    return {"id": identifier, "repository": row["repository"], "commit": row["commit"], "parent": row["parent"], "path": entry["path"],
            "message": row["message"].strip().splitlines()[0][:160], "assistant_trailer": row["assistant_trailer"],
            "web_edit": row["web_edit"] and not re.search(r"\(#\d+\)\s*$", row["message"].strip().splitlines()[0]),
            "query": row["query"], "date": row["date"], "text": before, "edited": after, "reference": spans}


jobs = []
for line in open(options.patches):
    row = json.loads(line)
    files = [f for f in row.get("files") or [] if f["status"] == "modified"]
    files.sort(key=lambda f: -sum(len(h["removed"]) for h in f["hunks"]))
    jobs += [(row, entry) for entry in files[:options.per_commit]]
kept = 0
with out.open("a") as handle, concurrent.futures.ThreadPoolExecutor(8) as pool:
    for number, result in enumerate(pool.map(work, jobs), 1):
        if result is None:
            continue
        handle.write(json.dumps(result, ensure_ascii=False) + "\n")
        kept += not result.get("skip")
        if number % 500 == 0:
            handle.flush()
            print(f"{number} of {len(jobs)} files, {kept} kept", flush=True)
print(f"done: {len(jobs)} files, {kept} kept")
