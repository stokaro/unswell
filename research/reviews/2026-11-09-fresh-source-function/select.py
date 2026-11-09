#!/usr/bin/env python3
"""Select source metadata from frozen GitHub code-search responses only."""

import argparse
import hashlib
import json
import re
from pathlib import Path


QUERIES = (
    '"it is worth noting" filename:README.md size:4000..20000',
    '"powerful and flexible" filename:README.md size:4000..20000',
    '"the following table" filename:README.md size:4000..20000',
)
EXCLUDED_NAMES = {
    "bat", "buildkit", "containerd", "curl", "etcd", "fzf", "imgui",
    "marked", "prometheus", "ptah", "pydantic", "ripgrep", "tokio",
    "book", "loki", "terraform", "cobra", "caddy", "restic", "gitea",
    "opentelemetry-collector", "atuin", "uv", "zed",
}
EXCLUDED_FULL = {
    "aliyuncontainerservice/pouch", "huolalatech/page-spy-web",
    "lispcookbook/cl-cookbook", "oceanographyforeveryone/openctd",
    "werwolv/imhex", "databendlabs/databend",
    "dstndstn/astrometry.net", "go-nunu/nunu",
    "google-research/tuning_playbook", "highlight/highlight",
    "huggingface/blog", "ifzhang/fairmot",
    "janoodleftw/timy-messenger", "llmware-ai/llmware",
    "ossu/computer-science", "quoid/userscripts", "rails/rails",
    "reapp/reapp", "swiftlang/swift", "uber/queryparser",
}
HEX40 = re.compile(r"[0-9a-f]{40}\Z")


def require(ok, reason):
    if not ok:
        raise ValueError(reason)


def select(search_dir: Path) -> dict:
    seen = set()
    searches = []
    selected = []
    for query_index, query in enumerate(QUERIES, 1):
        path = search_dir / f"{query_index:02d}.json"
        raw = path.read_bytes()
        response = json.loads(raw)
        items = response["items"]
        require(response["incomplete_results"] is False and
                0 < len(items) <= 30 and response["total_count"] >= len(items),
                "incomplete or invalid search response")
        searches.append({"query": query, "sha256": hashlib.sha256(raw).hexdigest(),
                         "total_count": response["total_count"], "returned": len(items)})
        picked = 0
        for rank, item in enumerate(items):
            repository = item["repository"]["full_name"]
            name = repository.rsplit("/", 1)[-1].lower()
            lower = repository.lower()
            if lower in seen or lower in EXCLUDED_FULL or name in EXCLUDED_NAMES:
                continue
            require(item["path"].rsplit("/", 1)[-1].lower() == "readme.md" and
                    HEX40.fullmatch(item["sha"]), "invalid source metadata")
            selected.append({"query_index": query_index, "rank": rank,
                             "repo": repository, "path": item["path"],
                             "search_blob_sha1": item["sha"]})
            seen.add(lower)
            picked += 1
            if picked == 3:
                break
        require(picked == 3, "fewer than three eligible results in a stratum")
    require(len(selected) == 9 and len(seen) == 9, "selection inventory changed")
    return {"format": "unswell-349-fresh-source-function-selection-v1",
            "scope": "cue_enriched_not_prevalence_or_prospective",
            "searches": searches, "selected": selected}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--search-dir", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    require(not args.output.exists(), "selection already exists")
    args.output.write_text(json.dumps(select(args.search_dir), indent=2,
                                      sort_keys=True) + "\n")
    print("selected nine source identities before content review")


if __name__ == "__main__":
    main()
