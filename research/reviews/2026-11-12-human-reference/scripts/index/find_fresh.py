#!/usr/bin/env python3
"""Find de-slop commits that the earlier searches did not return.

usage: find_fresh.py <workspace> <out jsonl> [--pages N] [--skip N] [--known <commits jsonl>]...

--skip leaves out the first N queries, which an earlier run of this script has paged already, and --known names the
commits file of such a run: its commits and repositories are left out as well.

The queries of find_deslop.py and a few more that name no particular mark, each sorted by committer date, newest first,
and paged deeper than before. A commit that an earlier search returned is left out, and so is a repository that any
earlier set holds. Only repository, commit, message, author, and committer are kept.
"""
import argparse, json, subprocess, sys, time
from pathlib import Path

QUERIES = ['"AI slop" README', '"AI slop" docs', '"de-slop" docs', '"de-slop" README', 'deslop README', 'deslop docs',
           '"AI-isms"', '"sounds like AI"', '"AI-sounding"', '"AI sounding"', '"LLM-isms"', '"slop" README wording',
           '"AI-generated" README rewrite', '"AI fluff"', '"remove AI" README wording', '"less AI" README',
           '"ChatGPT-isms"', '"sounds AI"', '"AI tells"', '"AI writing" README', '"marketing fluff" README', '"humanize" README',
           # Not in the earlier searches.
           '"AI slop" documentation', 'deslop documentation', '"de-slop"', 'unslop', '"LLM slop"', '"AI smell"', '"AI voice" README',
           '"sounds like ChatGPT"', '"GPT-isms"', '"AI tone" README', '"AI wording"', 'deslopify', '"anti-slop" README', '"humanise" README']
parser = argparse.ArgumentParser()
parser.add_argument("workspace")
parser.add_argument("out")
parser.add_argument("--pages", type=int, default=10)
parser.add_argument("--skip", type=int, default=0)
parser.add_argument("--known", action="append", default=[])
options = parser.parse_args()
W, out = options.workspace.rstrip("/") + "/", Path(options.out)
known_commits, known_repositories = set(), set()
for name in ("deslop-commits.jsonl", "deslop-reserve-commits.jsonl", "manual-commits.jsonl"):
    for line in open(W + "assisted/" + name):
        row = json.loads(line)
        known_commits.add(row["commit"])
for name in ("deslop-files.jsonl", "deslop-reserve-files.jsonl", "manual-files.jsonl", "deslop-patches.jsonl", "deslop-reserve-patches.jsonl",
             "manual-patches.jsonl", "manual-kept.jsonl"):
    for line in open(W + "assisted/" + name):
        known_repositories.add(json.loads(line).get("repository"))
for name in options.known:
    for line in open(name):
        row = json.loads(line)
        known_commits.add(row["commit"])
        known_repositories.add(row["repository"])
seen = {json.loads(line)["commit"] for line in out.read_text().splitlines()} if out.exists() else set()


def search(query, page):
    for _ in range(6):
        done = subprocess.run(["gh", "api", "-X", "GET", "search/commits", "-f", f"q={query} merge:false", "-f", "sort=committer-date",
                               "-f", "order=desc", "-f", "per_page=100", "-f", f"page={page}"], capture_output=True, text=True)
        if done.returncode == 0:
            return json.loads(done.stdout).get("items", [])
        if "rate limit" in done.stderr or "secondary" in done.stderr:
            time.sleep(75)
            continue
        print("search failed:", done.stderr.strip()[:120], file=sys.stderr)
        return []
    return []


added = 0
with out.open("a") as handle:
    for query in QUERIES[options.skip:]:
        for page in range(1, options.pages + 1):
            items = search(query, page)
            for item in items:
                repository = item["repository"]
                if (item["sha"] in seen or item["sha"] in known_commits or repository.get("private")
                        or repository["full_name"] in known_repositories):
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
            time.sleep(4)
            if len(items) < 100:
                break
