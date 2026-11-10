#!/usr/bin/env python3
"""Snapshot previously named review repositories before selecting new pages."""

import argparse
import json
import re
from pathlib import Path


REPOSITORY = re.compile(r"[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+\Z")
SOURCE_KEYS = frozenset(("repo", "repository", "full_name"))


def named_repositories(value):
    if isinstance(value, list):
        for item in value:
            yield from named_repositories(item)
    elif isinstance(value, dict):
        for key, item in value.items():
            if key in SOURCE_KEYS and isinstance(item, str) and REPOSITORY.fullmatch(item):
                yield item.lower()
            else:
                yield from named_repositories(item)


def snapshot(review_root: Path, packet_name: str, base_commit: str) -> dict:
    repositories = set()
    scanned = []
    for path in sorted(review_root.rglob("*.json")):
        if packet_name in path.parts or path.stat().st_size > 3_000_000:
            continue
        try:
            record = json.loads(path.read_text(encoding="utf-8"))
        except (UnicodeDecodeError, json.JSONDecodeError):
            continue
        scanned.append(path.relative_to(review_root).as_posix())
        repositories.update(named_repositories(record))
    return {
        "format": "unswell-349-prior-repository-snapshot-v1",
        "base_commit": base_commit,
        "method": "Recursive JSON values under repo, repository, or full_name; case-folded; files over 3 MB skipped",
        "scanned_json_files": len(scanned),
        "repositories": sorted(repositories),
    }


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--review-root", type=Path, required=True)
    parser.add_argument("--packet-name", required=True)
    parser.add_argument("--base-commit", required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if args.output.exists():
        raise SystemExit("exclusion snapshot already exists")
    result = snapshot(args.review_root, args.packet_name, args.base_commit)
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n",
                           encoding="utf-8")
    print(f"snapshotted {len(result['repositories'])} prior repository names")


if __name__ == "__main__":
    main()
