#!/usr/bin/env python3
"""Select twelve transfer-page identities before source content review."""

import argparse
import hashlib
import json
import re
from pathlib import Path


QUERIES = (
    '"next generation" filename:README.md size:4000..20000',
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


def select(search_dir: Path, previous_selection: Path) -> dict:
    previous = json.loads(previous_selection.read_bytes())
    if previous["format"] != "unswell-349-fresh-source-function-selection-v1":
        raise ValueError("unexpected prior selection")
    seen = {item["repo"].lower() for item in previous["selected"]}
    if len(seen) != 9:
        raise ValueError("prior nine-page packet changed")
    searches = []
    selected = []
    for index, query in enumerate(QUERIES, 1):
        raw = (search_dir / f"{index:02d}.json").read_bytes()
        response = json.loads(raw)
        items = response["items"]
        if (response["incomplete_results"] is not False or len(items) != 30
                or response["total_count"] < 30):
            raise ValueError(f"query {index}: incomplete search response")
        searches.append({"query": query, "sha256": hashlib.sha256(raw).hexdigest(),
                         "total_count": response["total_count"], "returned": len(items)})
        picked = 0
        for rank, item in enumerate(items):
            repo = item["repository"]["full_name"]
            lower = repo.lower()
            name = lower.rsplit("/", 1)[-1]
            if lower in seen or lower in EXCLUDED_FULL or name in EXCLUDED_NAMES:
                continue
            if (item["path"].rsplit("/", 1)[-1].lower() != "readme.md"
                    or not HEX40.fullmatch(item["sha"])):
                raise ValueError(f"query {index}: invalid selected metadata")
            selected.append({"query_index": index, "rank": rank, "repo": repo,
                             "path": item["path"], "search_blob_sha1": item["sha"]})
            seen.add(lower)
            picked += 1
            if picked == 4:
                break
        if picked != 4:
            raise ValueError(f"query {index}: fewer than four eligible repositories")
    if len(selected) != 12:
        raise ValueError("transfer inventory changed")
    return {"format": "unswell-349-source-function-transfer-v2-selection-v1",
            "scope": "cue_enriched_page_disjoint_transfer_not_prevalence_or_prospective",
            "previous_selection_sha256": hashlib.sha256(previous_selection.read_bytes()).hexdigest(),
            "searches": searches, "selected": selected}


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--search-dir", type=Path, required=True)
    parser.add_argument("--previous-selection", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if args.output.exists():
        raise ValueError("selection already exists")
    args.output.write_text(json.dumps(select(args.search_dir, args.previous_selection),
                                      indent=2, sort_keys=True) + "\n")
    print("selected twelve transfer identities before source content review")


if __name__ == "__main__":
    main()
