#!/usr/bin/env python3
"""Fetch the Markdown changes of de-slop commits: the lines removed and the lines put in their place.

usage: fetch_deslop.py <commits jsonl> <out jsonl> [--limit N]

One API call per commit. A commit is kept when it changes at most twelve
files, at least one of them Markdown, and the Markdown patches remove prose.
For every Markdown file the row keeps the hunks: removed lines and added
lines, as GitHub reports them. Nothing of a repository is executed.
"""
import argparse, json, re, subprocess, sys, time
from pathlib import Path

ASSISTANT = re.compile(r"co-authored-by:[^\n]*(claude|copilot|cursor|aider|codex|gemini|devin|openhands|jules|windsurf|cline|sweep)"
                       r"|generated with \[?claude code|\U0001F916", re.I)
parser = argparse.ArgumentParser()
parser.add_argument("commits")
parser.add_argument("out")
parser.add_argument("--limit", type=int, default=0)
options = parser.parse_args()
out = Path(options.out)
done = {json.loads(line)["commit"] for line in out.read_text().splitlines()} if out.exists() else set()
entries = [json.loads(line) for line in Path(options.commits).read_text().splitlines()]
entries = [e for e in entries if e["commit"] not in done and not e["fork"]]
if options.limit:
    entries = entries[:options.limit]


def hunks(patch):
    result, removed, added = [], [], []
    for line in patch.splitlines():
        if line.startswith("@@"):
            if removed or added:
                result.append({"removed": removed, "added": added})
            removed, added = [], []
        elif line.startswith("-"):
            removed.append(line[1:])
        elif line.startswith("+"):
            added.append(line[1:])
        elif removed or added:
            result.append({"removed": removed, "added": added})
            removed, added = [], []
    if removed or added:
        result.append({"removed": removed, "added": added})
    return result


import concurrent.futures, threading, urllib.error, urllib.request

TOKEN = subprocess.run(["gh", "auth", "token"], capture_output=True, text=True).stdout.strip()
lock, state = threading.Lock(), {"remaining": 5000, "kept": 0, "count": 0}
FLOOR = 700  # Leave this much of the hourly budget to other work on the same account.


def fetch(entry):
    if state["remaining"] < FLOOR:
        return None
    request = urllib.request.Request(f"https://api.github.com/repos/{entry['repository']}/commits/{entry['commit']}",
                                     headers={"Authorization": f"Bearer {TOKEN}", "Accept": "application/vnd.github+json",
                                              "User-Agent": "research-client"})
    for attempt in range(3):
        try:
            with urllib.request.urlopen(request, timeout=60) as response:
                state["remaining"] = int(response.headers.get("X-RateLimit-Remaining", state["remaining"]))
                data = json.load(response)
            break
        except urllib.error.HTTPError as error:
            if error.code in (403, 429):
                time.sleep(60)
                continue
            return {"commit": entry["commit"], "repository": entry["repository"], "files": [], "error": error.code}
        except Exception:  # noqa: BLE001
            time.sleep(5)
    else:
        return None
    files = data.get("files", [])
    markdown = [f for f in files if f["filename"].lower().endswith((".md", ".mdx")) and f.get("patch")]
    return {"commit": entry["commit"], "repository": entry["repository"], "message": entry["message"],
            "author": entry["author"], "committer": entry["committer"], "date": entry["date"], "query": entry["query"],
            "assistant_trailer": bool(ASSISTANT.search(entry["message"])), "web_edit": entry["committer"] == "GitHub",
            "changed_files": len(files), "parent": (data.get("parents") or [{}])[0].get("sha"),
            "files": [{"path": f["filename"], "status": f["status"], "hunks": hunks(f["patch"])} for f in markdown]
            if len(files) <= 12 else []}


with out.open("a") as handle, concurrent.futures.ThreadPoolExecutor(4) as pool:
    for row in pool.map(fetch, entries):
        if row is None:
            continue
        with lock:
            handle.write(json.dumps(row, ensure_ascii=False) + "\n")
            handle.flush()
            state["count"] += 1
            state["kept"] += bool(row.get("files"))
            if state["count"] % 200 == 0:
                print(f"{state['count']} commits fetched, {state['kept']} with Markdown changes, budget {state['remaining']}", flush=True)
print(f"done: {state['count']} commits fetched, {state['kept']} with Markdown changes, budget {state['remaining']}")
