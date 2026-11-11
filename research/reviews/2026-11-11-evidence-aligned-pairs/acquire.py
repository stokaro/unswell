#!/usr/bin/env python3
"""Freeze public repository search and select README blobs without reading text."""

import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess
from urllib.parse import quote


QUERIES = {
    "go": "topic:developer-tools language:Go stars:500..5000 archived:false",
    "rust": "topic:developer-tools language:Rust stars:500..5000 archived:false",
    "python": "topic:developer-tools language:Python stars:500..5000 archived:false",
}
REPOSITORY = re.compile(r"[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+\Z")


def digest(data):
    return hashlib.sha256(data).hexdigest()


def get_json(endpoint, *, fields=()):
    command = ["gh", "api", "-X", "GET", endpoint]
    for key, value in fields:
        command += ["-f", f"{key}={value}"]
    result = subprocess.run(command, check=True, capture_output=True)
    return result.stdout, json.loads(result.stdout)


def write_new(path, data):
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("xb") as output:
        output.write(data)


def canonical(value):
    return (json.dumps(value, sort_keys=True, indent=2, ensure_ascii=False) + "\n").encode()


def search(directory):
    if any((directory / f"{name}.json").exists() for name in QUERIES):
        raise SystemExit("search snapshot already exists")
    for name, query in QUERIES.items():
        raw, response = get_json("search/repositories", fields=(
            ("q", query), ("sort", "updated"), ("order", "desc"), ("per_page", "100"),
        ))
        if response.get("incomplete_results") or not isinstance(response.get("items"), list):
            raise SystemExit(f"incomplete search response: {name}")
        write_new(directory / f"{name}.json", raw)
        print(json.dumps({"stratum": name, "items": len(response["items"]), "sha256": digest(raw)}))


def selection(directory, excluded_path, output_path):
    excluded_raw = excluded_path.read_bytes()
    excluded = json.loads(excluded_raw)
    repositories = excluded["repositories"]
    if repositories != sorted(set(repositories)) or not all(REPOSITORY.fullmatch(x) for x in repositories):
        raise SystemExit("exclusion repositories must be sorted, unique owner/name identities")
    used = set(repositories)
    selected = []
    considered = []
    snapshots = {}
    for name, query in QUERIES.items():
        raw = (directory / f"{name}.json").read_bytes()
        response = json.loads(raw)
        if response.get("incomplete_results") or len(response.get("items", [])) > 100:
            raise SystemExit(f"invalid search response: {name}")
        snapshots[name] = {"query": query, "sha256": digest(raw), "items": len(response["items"])}
        count = 0
        for rank, item in enumerate(response["items"], 1):
            if count == 8:
                break
            repo = item.get("full_name", "")
            entry = {"stratum": name, "rank": rank, "repo": repo}
            if not REPOSITORY.fullmatch(repo):
                entry["reason"] = "invalid_repository_identity"
            elif item.get("private") or item.get("fork") or item.get("archived"):
                entry["reason"] = "private_fork_or_archived"
            elif repo.lower() in used:
                entry["reason"] = "prior_source_or_duplicate"
            else:
                branch = item.get("default_branch", "")
                if not branch:
                    entry["reason"] = "missing_default_branch"
                else:
                    _, commit = get_json(f"repos/{repo}/commits/{quote(branch, safe='')}")
                    commit_sha = commit["sha"]
                    tree_sha = commit["commit"]["tree"]["sha"]
                    _, tree = get_json(f"repos/{repo}/git/trees/{tree_sha}")
                    readme = next((x for x in tree["tree"] if x["path"] == "README.md" and
                                   x["type"] == "blob"), None)
                    if readme is None:
                        entry["reason"] = "missing_root_readme"
                    elif not 4096 <= readme["size"] <= 30720:
                        entry["reason"] = "readme_size_outside_cap"
                        entry["size"] = readme["size"]
                    else:
                        entry.update({"default_branch": branch, "commit_sha": commit_sha,
                                      "tree_sha": tree_sha, "readme_blob_sha1": readme["sha"],
                                      "readme_bytes": readme["size"], "selection_ordinal": count + 1})
                        count += 1
                        used.add(repo.lower())
                        selected.append(entry.copy())
            considered.append(entry)
    result = {"format": "unswell-349-evidence-aligned-selection-v1",
              "main_commit": "c03796fec42c4eb7107b69d9170ee06d09523136",
              "exclusions_sha256": digest(excluded_raw), "search_snapshots": snapshots,
              "selected": selected, "considered": considered,
              "status": "complete" if len(selected) == 24 else "underpowered_selection"}
    write_new(output_path, canonical(result))
    print(json.dumps({"status": result["status"], "selected": len(selected),
                      "considered": len(considered), "sha256": digest(canonical(result))}))


def main():
    parser = argparse.ArgumentParser()
    sub = parser.add_subparsers(dest="command", required=True)
    search_args = sub.add_parser("search")
    search_args.add_argument("directory", type=Path)
    select_args = sub.add_parser("select")
    select_args.add_argument("directory", type=Path)
    select_args.add_argument("excluded", type=Path)
    select_args.add_argument("output", type=Path)
    args = parser.parse_args()
    if args.command == "search":
        search(args.directory)
    else:
        selection(args.directory, args.excluded, args.output)


if __name__ == "__main__":
    main()
