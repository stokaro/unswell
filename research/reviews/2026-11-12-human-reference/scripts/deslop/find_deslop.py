#!/usr/bin/env python3
"""Find commits whose message says that they remove AI-sounding wording.

usage: find_deslop.py <out jsonl> [--pages N]

Each query is paged and sliced by nothing else; results are deduplicated by
commit. Only repository, commit, message, author, and committer are kept.
"""
import argparse, json, subprocess, sys, time
from pathlib import Path

QUERIES = ['"AI slop" README', '"AI slop" docs', '"de-slop" docs', '"de-slop" README', 'deslop README', 'deslop docs',
           '"AI-isms"', '"sounds like AI"', '"AI-sounding"', '"AI sounding"', '"LLM-isms"', '"slop" README wording',
           '"AI-generated" README rewrite', '"AI fluff"', '"remove AI" README wording', '"less AI" README',
           '"ChatGPT-isms"', '"sounds AI"', '"AI tells"', '"AI writing" README', '"marketing fluff" README', '"humanize" README']
parser = argparse.ArgumentParser()
parser.add_argument("out")
parser.add_argument("--pages", type=int, default=3)
options = parser.parse_args()
out = Path(options.out)
seen = {json.loads(line)["commit"] for line in out.read_text().splitlines()} if out.exists() else set()


def search(query, page):
    for _ in range(5):
        done = subprocess.run(["gh", "api", "-X", "GET", "search/commits", "-f", f"q={query} merge:false", "-f", "per_page=100",
                               "-f", f"page={page}"], capture_output=True, text=True)
        if done.returncode == 0:
            return json.loads(done.stdout).get("items", [])
        if "rate limit" in done.stderr:
            time.sleep(75)
            continue
        print("search failed:", done.stderr.strip()[:120], file=sys.stderr)
        return []
    return []


added = 0
with out.open("a") as handle:
    for query in QUERIES:
        for page in range(1, options.pages + 1):
            items = search(query, page)
            for item in items:
                repository = item["repository"]
                if item["sha"] in seen or repository.get("private"):
                    continue
                seen.add(item["sha"])
                added += 1
                commit = item["commit"]
                handle.write(json.dumps({"repository": repository["full_name"], "fork": repository.get("fork", False),
                                         "commit": item["sha"], "message": commit["message"][:2000],
                                         "author": commit["author"]["name"], "committer": commit["committer"]["name"],
                                         "date": commit["committer"]["date"], "query": query}, ensure_ascii=False) + "\n")
            handle.flush()
            print(f"{query!r} page {page}: {len(items)} commits, {added} new so far", flush=True)
            time.sleep(8)
            if len(items) < 100:
                break
