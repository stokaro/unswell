#!/usr/bin/env python3
"""Collect README files as they were before language models: the last version before November 2022.

usage: find_human.py <out jsonl> [--weeks N] [--per-week N]

Repositories are sampled by creation week between 2016 and 2021, with 5 to 300 stars, not forks. For each, the last
commit that touched README.md before 2022-11-01 gives the text. Kept: English prose of 150 words or more. A row holds
repository, commit, date, path, and text. Nothing of a repository is executed.
"""
import argparse, datetime, json, random, re, subprocess, sys, time, urllib.error, urllib.parse, urllib.request
from pathlib import Path

sys.path.insert(0, __file__.rsplit("/", 1)[0])
from english import share

parser = argparse.ArgumentParser()
parser.add_argument("out")
parser.add_argument("--weeks", type=int, default=60)
parser.add_argument("--per-week", type=int, default=30)
parser.add_argument("--seed", type=int, default=20261113)
options = parser.parse_args()
out = Path(options.out)
seen = {json.loads(line)["repository"] for line in out.read_text().splitlines()} if out.exists() else set()
TOKEN = subprocess.run(["gh", "auth", "token"], capture_output=True, text=True).stdout.strip()
WORD = re.compile(r"[A-Za-z][A-Za-z'’-]*")
rng = random.Random(options.seed)


def api(url):
    request = urllib.request.Request(url, headers={"Authorization": f"Bearer {TOKEN}", "Accept": "application/vnd.github+json",
                                                   "User-Agent": "research-client"})
    for attempt in range(4):
        try:
            with urllib.request.urlopen(request, timeout=60) as response:
                return json.load(response), int(response.headers.get("X-RateLimit-Remaining", 5000))
        except urllib.error.HTTPError as error:
            if error.code in (403, 429):
                time.sleep(70)
                continue
            return None, 5000
        except Exception:  # noqa: BLE001
            time.sleep(5)
    return None, 5000


def raw(repository, sha, path):
    url = f"https://raw.githubusercontent.com/{repository}/{sha}/{urllib.parse.quote(path)}"
    try:
        with urllib.request.urlopen(urllib.request.Request(url, headers={"User-Agent": "research-client"}), timeout=60) as response:
            data = response.read(200_001)
        return None if len(data) > 200_000 else data.decode("utf-8", "replace")
    except Exception:  # noqa: BLE001
        return None


start, end = datetime.date(2016, 1, 4), datetime.date(2021, 12, 20)
weeks = [start + datetime.timedelta(days=7 * n) for n in range((end - start).days // 7)]
kept = 0
with out.open("a") as handle:
    for week in rng.sample(weeks, options.weeks):
        query = f"created:{week}..{week + datetime.timedelta(days=6)} stars:5..300 fork:false"
        found, _ = api("https://api.github.com/search/repositories?per_page=100&sort=updated&q=" + urllib.parse.quote(query))
        time.sleep(7)
        repositories = [item["full_name"] for item in (found or {}).get("items", []) if not item.get("private")]
        rng.shuffle(repositories)
        taken = 0
        for repository in repositories:
            if taken >= options.per_week or repository in seen:
                continue
            seen.add(repository)
            commits, remaining = api(f"https://api.github.com/repos/{repository}/commits?path=README.md&until=2022-10-31T23:59:59Z&per_page=1")
            if remaining < 800:
                time.sleep(600)
            if not commits or not isinstance(commits, list):
                continue
            sha, date = commits[0]["sha"], commits[0]["commit"]["committer"]["date"]
            text = raw(repository, sha, "README.md")
            if not text or len(WORD.findall(text)) < 150 or share(text)[0] < 0.10:
                continue
            handle.write(json.dumps({"repository": repository, "commit": sha, "date": date, "path": "README.md", "text": text},
                                    ensure_ascii=False) + "\n")
            handle.flush()
            taken += 1
            kept += 1
        print(f"{week}: {len(repositories)} repositories found, {taken} pages kept, {kept} in all", flush=True)
