#!/usr/bin/env python3
"""Inventory fresh historical technical sources without model inference."""

from __future__ import annotations

import argparse
import concurrent.futures
import csv
import hashlib
import json
from pathlib import Path
import re
import subprocess
import threading
import time


CSV_SHA256 = "5da66ce7e15c41af2df8d6883e8d1ccef79b2157afd0b665f4157420df087710"
EXCLUSIONS_SHA256 = "f3dcfc6021254fdee28012b4ffc563f2b83a71c6d23c5cc3dcab6e3170220f79"
SEED = b"unswell-techdoc-new-controls-v1"
CUTOFF = "2021-08-15T00:00:00Z"
LANGUAGES = ("C#", "C++", "Java", "Python")
PATH_SKIP = {".github", "dependencies", "dependency", "vendor", "vendored",
             "third_party", "third-party", "node_modules", "submodules", ".git"}
NAME_SKIP = re.compile(
    r"^(?:changelog|changes|contributing|contributors|license|licence|"
    r"copying|code_of_conduct|security|support|release.?notes|notice|authors|"
    r"todo|readme\.template)(?:[._-].*)?\.(?:md|rst)$", re.IGNORECASE)


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def rank(repo: str) -> str:
    return hashlib.sha256(repo.lower().encode() + b"\0" + SEED).hexdigest()


def candidates(csv_path: Path, exclusions_path: Path) -> list[dict]:
    if sha(csv_path) != CSV_SHA256 or sha(exclusions_path) != EXCLUSIONS_SHA256:
        raise ValueError("Source frame or exclusion manifest changed")
    exclusions = {name.lower() for name in json.loads(exclusions_path.read_text())["excluded_repositories"]}
    rows = []
    for source in csv.DictReader(csv_path.open(newline="")):
        repo = source["User_Name"] + "/" + source["Repo_Name"]
        if (source["Language"] not in LANGUAGES or int(source["stars"]) < 100 or
                int(source["Textual_files"]) < 3 or repo.lower() in exclusions):
            continue
        rows.append({"repository": repo, "language": source["Language"],
                     "docmine_stars": int(source["stars"]),
                     "docmine_textual_files": int(source["Textual_files"]),
                     "rank": rank(repo)})
    rows.sort(key=lambda row: (row["rank"], row["repository"].lower()))
    if len(rows) != 465 or len({row["repository"].lower() for row in rows}) != 465:
        raise ValueError("Unexpected fresh repository candidate count")
    return rows


def first_party(path: str) -> bool:
    parts = Path(path).parts
    return (bool(parts) and not any(part.lower() in PATH_SKIP for part in parts[:-1])
            and bool(re.search(r"\.(?:md|rst)$", parts[-1], re.IGNORECASE))
            and not NAME_SKIP.match(parts[-1]))


def api(path: str):
    for attempt in range(2):
        process = subprocess.run(["gh", "api", path], capture_output=True,
                                 text=True, check=False, timeout=90)
        if process.returncode == 0:
            return json.loads(process.stdout)
        error = process.stderr.strip()[:300]
        if attempt == 0 and not ("HTTP 404" in error or "HTTP 403" in error):
            time.sleep(1)
            continue
        raise RuntimeError(error)


def inspect(row: dict) -> dict:
    repo = row["repository"]
    try:
        commits = api(f"repos/{repo}/commits?until={CUTOFF}&per_page=1")
        if not commits:
            return {**row, "error": "no_historical_commit"}
        commit = commits[0]
        commit_sha = commit["sha"]
        committed_at = commit["commit"]["committer"]["date"]
        if committed_at >= CUTOFF:
            return {**row, "error": "commit_after_cutoff"}
        tree = api(f"repos/{repo}/git/trees/{commit_sha}?recursive=1")
        if tree.get("truncated"):
            return {**row, "commit": commit_sha,
                    "committed_at": committed_at, "error": "truncated_tree"}
        docs = sorted(({"path": b["path"], "blob_sha": b["sha"], "bytes": b["size"]}
                       for b in tree["tree"] if b.get("type") == "blob"
                       and 2000 <= b.get("size", -1) <= 200000
                       and first_party(b["path"])), key=lambda b: b["path"])
        return {**row, "commit": commit_sha, "committed_at": committed_at,
                "documents": docs}
    except Exception as error:
        return {**row, "error": str(error)}


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--csv", type=Path, required=True)
    parser.add_argument("--exclusions", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--workers", type=int, default=8)
    args = parser.parse_args()
    candidates_list = candidates(args.csv, args.exclusions)
    completed = 0
    lock = threading.Lock()

    def work(row: dict) -> dict:
        nonlocal completed
        result = inspect(row)
        with lock:
            completed += 1
            if completed % 25 == 0 or completed == len(candidates_list):
                print(f"inspected {completed}/{len(candidates_list)} repositories", flush=True)
        return result

    with concurrent.futures.ThreadPoolExecutor(max_workers=args.workers) as pool:
        results = list(pool.map(work, candidates_list))
    inventory = {"format": "unswell-origin-new-historical-inventory-v1",
                 "seed": SEED.decode(), "cutoff": CUTOFF,
                 "source_csv_sha256": CSV_SHA256,
                 "exclusions_sha256": EXCLUSIONS_SHA256,
                 "rows": results}
    args.output.write_text(json.dumps(inventory, indent=2, sort_keys=True) + "\n")
    print(json.dumps({"repositories": len(results),
                      "with_documents": sum(bool(r.get("documents")) for r in results),
                      "candidate_documents": sum(len(r.get("documents", [])) for r in results),
                      "errors": sum(bool(r.get("error")) for r in results)}, sort_keys=True))


if __name__ == "__main__":
    main()
