#!/usr/bin/env python3
"""Select pinned source identities without reading their content."""

import argparse
import hashlib
import json
import re
from pathlib import Path
from urllib.parse import parse_qs, urlparse


QUERIES = (
    '"it is important to" filename:README.md size:4000..20000',
    '"as a result" filename:README.md size:4000..20000',
    '"we believe" filename:README.md size:4000..20000',
)
HEX40 = re.compile(r"[0-9a-f]{40}\Z")


def require(condition, message):
    if not condition:
        raise ValueError(message)


def select(search_dir: Path, excluded_path: Path) -> dict:
    excluded_raw = excluded_path.read_bytes()
    excluded_record = json.loads(excluded_raw)
    require(excluded_record["format"] == "unswell-349-prior-repository-snapshot-v1",
            "wrong exclusion snapshot format")
    excluded = set(excluded_record["repositories"])
    require(len(excluded) == len(excluded_record["repositories"]),
            "duplicate excluded repository")
    seen = set()
    searches = []
    selected = []
    for query_index, query in enumerate(QUERIES, 1):
        raw = (search_dir / f"{query_index:02d}.json").read_bytes()
        response = json.loads(raw)
        items = response["items"]
        require(response["incomplete_results"] is False and
                len(items) == 30 and response["total_count"] >= 30,
                "incomplete or truncated search response")
        searches.append({"query": query, "sha256": hashlib.sha256(raw).hexdigest(),
                         "total_count": response["total_count"], "returned": len(items)})
        picked = 0
        for rank, item in enumerate(items):
            repository = item["repository"]["full_name"]
            normalized = repository.lower()
            if normalized in excluded or normalized in seen:
                continue
            path = item["path"]
            ref = parse_qs(urlparse(item["url"]).query).get("ref", [""])[0]
            require(path.rsplit("/", 1)[-1].lower() == "readme.md" and
                    HEX40.fullmatch(item["sha"]) and HEX40.fullmatch(ref),
                    "invalid source identity")
            require(f"/blob/{ref}/" in item["html_url"],
                    "search page URL does not match pinned commit")
            selected.append({"query_index": query_index, "rank": rank,
                             "repository": repository, "path": path,
                             "commit": ref, "search_blob_sha1": item["sha"],
                             "source_url": item["html_url"]})
            seen.add(normalized)
            picked += 1
            if picked == 3:
                break
        require(picked == 3, "fewer than three eligible results in a query")
    require(len(selected) == 9 and len(seen) == 9, "selection count changed")
    return {
        "format": "unswell-349-atomic-claim-selection-v1",
        "scope": "cue_enriched_exposed_development_not_prevalence_or_confirmation",
        "exclusions_sha256": hashlib.sha256(excluded_raw).hexdigest(),
        "searches": searches,
        "selected": selected,
    }


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--search-dir", type=Path, required=True)
    parser.add_argument("--excluded", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if args.output.exists():
        raise SystemExit("selection already exists")
    result = select(args.search_dir, args.excluded)
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n",
                           encoding="utf-8")
    print("selected nine source identities before content review")


if __name__ == "__main__":
    main()
