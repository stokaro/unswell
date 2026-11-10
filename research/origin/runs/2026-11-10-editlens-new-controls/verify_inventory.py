#!/usr/bin/env python3
"""Check the new corpus inventory independently of its fetcher."""

from __future__ import annotations

import argparse
import csv
import hashlib
import json
from pathlib import Path
import re


CSV_SHA256 = "5da66ce7e15c41af2df8d6883e8d1ccef79b2157afd0b665f4157420df087710"
EXCLUSIONS_SHA256 = "f3dcfc6021254fdee28012b4ffc563f2b83a71c6d23c5cc3dcab6e3170220f79"
SEED = b"unswell-techdoc-new-controls-v1"
LANGUAGES = {"C#", "C++", "Java", "Python"}
SKIP_DIRS = {".github", "dependencies", "dependency", "vendor", "vendored",
             "third_party", "third-party", "node_modules", "submodules", ".git"}
SKIP_NAME = re.compile(
    r"^(changelog|changes|contributing|contributors|license|licence|copying|"
    r"code_of_conduct|security|support|release.?notes|notice|authors|todo|"
    r"readme\.template)([._-].*)?\.(md|rst)$", re.IGNORECASE)


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def check(ok: bool, reason: str) -> None:
    if not ok:
        raise ValueError(reason)


def expected(csv_path: Path, exclusion_path: Path) -> list[dict]:
    check(digest(csv_path) == CSV_SHA256, "DocMine CSV hash")
    check(digest(exclusion_path) == EXCLUSIONS_SHA256, "Exclusion hash")
    excluded = {name.lower() for name in json.loads(exclusion_path.read_text())["excluded_repositories"]}
    result = []
    for item in csv.DictReader(csv_path.open(newline="")):
        repo = f"{item['User_Name']}/{item['Repo_Name']}"
        if (repo.lower() in excluded or item["Language"] not in LANGUAGES
                or int(item["stars"]) < 100 or int(item["Textual_files"]) < 3):
            continue
        result.append({"repository": repo, "language": item["Language"],
                       "docmine_stars": int(item["stars"]),
                       "docmine_textual_files": int(item["Textual_files"]),
                       "rank": hashlib.sha256(repo.lower().encode() + b"\0" + SEED).hexdigest()})
    return sorted(result, key=lambda item: (item["rank"], item["repository"].lower()))


def verify(csv_path: Path, exclusion_path: Path, inventory_path: Path) -> dict:
    inventory = json.loads(inventory_path.read_text())
    check(inventory["format"] == "unswell-origin-new-historical-inventory-v1", "Format")
    check(inventory["seed"] == SEED.decode(), "Seed")
    check(inventory["cutoff"] == "2021-08-15T00:00:00Z", "Date cutoff")
    check(inventory["source_csv_sha256"] == CSV_SHA256, "Source identity")
    check(inventory["exclusions_sha256"] == EXCLUSIONS_SHA256, "Exclusions identity")
    rows = inventory["rows"]
    selected = expected(csv_path, exclusion_path)
    check(len(rows) == len(selected) == 465, "Candidate count")
    seen = set()
    for row, candidate in zip(rows, selected, strict=True):
        check(all(row[field] == value for field, value in candidate.items()),
              f"Candidate order or identity: {candidate['repository']}")
        check(row["repository"].lower() not in seen, "Duplicate repository")
        seen.add(row["repository"].lower())
        if "error" in row:
            check("documents" not in row, "Error with document inventory")
            continue
        check(re.fullmatch(r"[a-f0-9]{40}", row["commit"]) is not None,
              "Invalid commit SHA")
        check(row["committed_at"] < inventory["cutoff"], "Source after cutoff")
        paths = set()
        for doc in row["documents"]:
            path = doc["path"]
            parts = Path(path).parts
            check(path not in paths and parts and not Path(path).is_absolute(), "Duplicate or absolute path")
            paths.add(path)
            check(not any(part.lower() in SKIP_DIRS for part in parts[:-1]), "Vendored path")
            check(re.search(r"\.(md|rst)$", parts[-1], re.IGNORECASE) is not None
                  and SKIP_NAME.match(parts[-1]) is None, "Invalid document path")
            check(2000 <= doc["bytes"] <= 200000 and
                  re.fullmatch(r"[a-f0-9]{40}", doc["blob_sha"]) is not None,
                  "Invalid blob metadata")
        check(row["documents"] == sorted(row["documents"], key=lambda doc: doc["path"]),
              "Document ordering")
    return {"repositories": len(rows),
            "with_documents": sum(bool(row.get("documents")) for row in rows),
            "candidate_documents": sum(len(row.get("documents", [])) for row in rows),
            "errors": sum("error" in row for row in rows),
            "inventory_sha256": digest(inventory_path)}


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--csv", type=Path, required=True)
    parser.add_argument("--exclusions", type=Path, required=True)
    parser.add_argument("--inventory", type=Path, required=True)
    args = parser.parse_args()
    print(json.dumps(verify(args.csv, args.exclusions, args.inventory), sort_keys=True))
