#!/usr/bin/env python3
"""Find de-slop commits that were made in GitHub's web editor, where a person most likely typed the change.

usage: find_manual.py <out jsonl> [--pages N]

The queries are those of find_deslop.py and a few informal ones, restricted to
commits whose committer is GitHub's web flow (noreply@github.com) and that are
not merges. A subject that ends in a pull request number is a merge made in the
web interface, not an edit, and is left out. Only repository, commit, message,
author, and committer are kept.
"""
import argparse, json, re, subprocess, sys, time
from pathlib import Path

QUERIES = ['"AI slop"', 'deslop', '"de-slop"', '"AI-isms"', '"sounds like AI"', '"AI-sounding"', '"AI sounding"', '"LLM-isms"',
           '"AI fluff"', '"less AI"', '"ChatGPT-isms"', '"sounds AI"', '"AI tells"', '"AI writing"', '"sounds like ChatGPT"',
           '"AI-ish"', '"too AI"', '"AI wording"', '"AI language"', '"AI tone"', '"AI generated" rewrite', '"AI-generated" wording',
           '"by hand" AI README', '"manually" "AI" rewrite README', '"my own words"', '"humanize"', '"slop" README',
           '"robotic" README', '"LLM slop"', '"GPT-isms"', '"AI text"', '"AI style"', '"sound less like AI"', '"less robotic"']
parser = argparse.ArgumentParser()
parser.add_argument("out")
parser.add_argument("--pages", type=int, default=10)
options = parser.parse_args()
out = Path(options.out)
seen = {json.loads(line)["commit"] for line in out.read_text().splitlines()} if out.exists() else set()
MERGED = re.compile(r"\(#\d+\)\s*$")


def search(query, page):
    for _ in range(5):
        done = subprocess.run(["gh", "api", "-X", "GET", "search/commits", "-f",
                               f"q={query} committer-email:noreply@github.com merge:false", "-f", "per_page=100", "-f", f"page={page}"],
                              capture_output=True, text=True)
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
                repository, commit = item["repository"], item["commit"]
                subject = commit["message"].strip().splitlines()[0] if commit["message"].strip() else ""
                if item["sha"] in seen or repository.get("private") or MERGED.search(subject):
                    continue
                seen.add(item["sha"])
                added += 1
                handle.write(json.dumps({"repository": repository["full_name"], "fork": repository.get("fork", False),
                                         "commit": item["sha"], "message": commit["message"][:2000],
                                         "author": commit["author"]["name"], "committer": commit["committer"]["name"],
                                         "date": commit["committer"]["date"], "query": query}, ensure_ascii=False) + "\n")
            handle.flush()
            print(f"{query!r} page {page}: {len(items)} commits, {added} new so far", flush=True)
            time.sleep(8)
            if len(items) < 100:
                break
