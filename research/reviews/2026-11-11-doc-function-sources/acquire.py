#!/usr/bin/env python3
"""Freeze topic searches and select documentation blobs without reading text."""

import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess
from urllib.parse import quote


QUERIES = {
    "database": "topic:database stars:500..5000 archived:false",
    "security": "topic:security stars:500..5000 archived:false",
    "observability": "topic:observability stars:500..5000 archived:false",
    "testing": "topic:testing stars:500..5000 archived:false",
}
REPOSITORY = re.compile(r"[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+\Z")
GUIDE = {"guide", "guides", "tutorial", "tutorials", "howto", "quickstart",
         "manual"}
CONCEPT = {"concept", "concepts", "explanation", "explanations", "reference",
           "references", "api", "architecture", "design", "internals",
           "implementation"}
EXCLUDED_BASENAMES = ("readme", "changelog", "contributing", "license")


def digest(data):
    return hashlib.sha256(data).hexdigest()


def canonical(value):
    return (json.dumps(value, sort_keys=True, indent=2, ensure_ascii=False) + "\n").encode()


def write_new(path, data):
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("xb") as output:
        output.write(data)


def get_json(endpoint, *, fields=()):
    command = ["gh", "api", "-X", "GET", endpoint]
    for key, value in fields:
        command += ["-f", f"{key}={value}"]
    result = subprocess.run(command, check=True, capture_output=True)
    return result.stdout, json.loads(result.stdout)


def file_functions(item):
    path = item.get("path", "")
    parts = path.lower().split("/")
    if (item.get("type") != "blob" or
            not isinstance(item.get("size"), int) or
            not 8192 <= item["size"] <= 32768 or
            not any(part in {"docs", "doc", "documentation"} for part in parts[:-1]) or
            not parts[-1].endswith((".md", ".mdx")) or
            parts[-1].startswith(EXCLUDED_BASENAMES)):
        return set()
    tokens = set(re.split(r"[^a-z0-9]+", path.lower()))
    normalized = "-" + re.sub(r"[^a-z0-9]+", "-", path.lower()) + "-"
    functions = set()
    if tokens & GUIDE or "-how-to-" in normalized or "-getting-started-" in normalized:
        functions.add("guide")
    if tokens & CONCEPT:
        functions.add("concept_reference")
    return functions


def best_files(tree):
    choices = {}
    for item in tree:
        for function in file_functions(item):
            prior = choices.get(function)
            if prior is None or (-item["size"], item["path"]) < (-prior["size"], prior["path"]):
                choices[function] = item
    return choices


def search(directory):
    if any((directory / f"{name}.json").exists() for name in QUERIES):
        raise SystemExit("search snapshot already exists")
    for name, query in QUERIES.items():
        raw, response = get_json("search/repositories", fields=(
            ("q", query), ("sort", "stars"), ("order", "desc"), ("per_page", "100"),
        ))
        if response.get("incomplete_results") or not isinstance(response.get("items"), list):
            raise SystemExit(f"incomplete search response: {name}")
        write_new(directory / f"{name}.json", raw)
        print(json.dumps({"topic": name, "items": len(response["items"]),
                          "sha256": digest(raw)}))


def select(directory, exclusions_path, output_path):
    exclusions_raw = exclusions_path.read_bytes()
    exclusions = json.loads(exclusions_raw)["repositories"]
    if exclusions != sorted(set(exclusions)) or not all(REPOSITORY.fullmatch(x) for x in exclusions):
        raise ValueError("exclusions must be sorted, unique repository identities")
    used = set(exclusions)
    selected = []
    considered = []
    snapshots = {}
    for topic, query in QUERIES.items():
        raw = (directory / f"{topic}.json").read_bytes()
        response = json.loads(raw)
        if response.get("incomplete_results") or not isinstance(response.get("items"), list) or len(response["items"]) > 100:
            raise ValueError(f"invalid search snapshot: {topic}")
        snapshots[topic] = {"query": query, "sha256": digest(raw),
                            "items": len(response["items"])}
        counts = {"guide": 0, "concept_reference": 0}
        for rank, repo_item in enumerate(response["items"], 1):
            if all(count == 4 for count in counts.values()):
                break
            repo = repo_item.get("full_name", "")
            entry = {"topic": topic, "rank": rank, "repo": repo}
            if not REPOSITORY.fullmatch(repo):
                entry["reason"] = "invalid_repository_identity"
            elif repo_item.get("private") or repo_item.get("fork") or repo_item.get("archived"):
                entry["reason"] = "private_fork_or_archived"
            elif repo.lower() in used:
                entry["reason"] = "prior_source_or_duplicate"
            elif not repo_item.get("default_branch"):
                entry["reason"] = "missing_default_branch"
            else:
                branch = repo_item["default_branch"]
                _, commit = get_json(f"repos/{repo}/commits/{quote(branch, safe='')}")
                commit_sha = commit["sha"]
                tree_sha = commit["commit"]["tree"]["sha"]
                _, tree = get_json(f"repos/{repo}/git/trees/{tree_sha}?recursive=1")
                if tree.get("truncated") or not isinstance(tree.get("tree"), list):
                    entry["reason"] = "missing_or_truncated_tree"
                else:
                    choices = best_files(tree["tree"])
                    available = [function for function in ("guide", "concept_reference")
                                 if function in choices and counts[function] < 4]
                    if not available:
                        entry["reason"] = "no_eligible_document_in_open_cell"
                    else:
                        function = min(available, key=lambda name: (counts[name],
                                                                    name != "guide"))
                        document = choices[function]
                        counts[function] += 1
                        used.add(repo.lower())
                        entry.update({"function": function, "selection_ordinal": counts[function],
                                      "default_branch": branch, "commit_sha": commit_sha,
                                      "tree_sha": tree_sha, "path": document["path"],
                                      "blob_sha1": document["sha"], "bytes": document["size"]})
                        selected.append(entry.copy())
            considered.append(entry)
    status = "complete" if len(selected) == 32 else "underpowered_selection"
    result = {"format": "unswell-349-doc-function-selection-v1",
              "main_commit": "8208fe6ca16bdc508e4adef654561d386b3da16b",
              "exclusions_sha256": digest(exclusions_raw),
              "search_snapshots": snapshots, "selected": selected,
              "considered": considered, "status": status}
    write_new(output_path, canonical(result))
    print(json.dumps({"status": status, "selected": len(selected),
                      "considered": len(considered), "sha256": digest(canonical(result))}))


def main():
    parser = argparse.ArgumentParser()
    sub = parser.add_subparsers(dest="command", required=True)
    search_args = sub.add_parser("search")
    search_args.add_argument("directory", type=Path)
    select_args = sub.add_parser("select")
    select_args.add_argument("directory", type=Path)
    select_args.add_argument("exclusions", type=Path)
    select_args.add_argument("output", type=Path)
    args = parser.parse_args()
    if args.command == "search":
        search(args.directory)
    else:
        select(args.directory, args.exclusions, args.output)


if __name__ == "__main__":
    main()
