"""Select a fixed cue-enriched page challenge from frozen GitHub search pages."""

from __future__ import annotations

import hashlib
import json
import sys
from pathlib import Path


QUERIES = (
    '"we hope this guide" extension:md',
    '"we hope this" extension:md',
    '"let\'s explore" extension:md',
    '"this is the place to start" extension:md',
)
EXCLUDED_NAMES = {
    "ptah", "containerd", "prometheus", "tokio", "pydantic", "bat",
    "ripgrep", "imgui", "fzf", "curl", "marked", "etcd", "buildkit",
}
EXCLUDED_REPOS = {
    "caddyserver/caddy", "carbon-design-system/carbon",
    "Kinds-of-Intelligence-CFI/animal-ai", "discourse/discourse",
    "nats-io/nats-server", "denoland/deno", "neovim/neovim",
    "syncthing/syncthing", "astral-sh/ruff", "duckdb/duckdb",
    "qdrant/qdrant", "grafana/loki",
}


def select(search_dir: Path) -> dict:
    used_repos = set()
    used_paths = set()
    pages = []
    selection = []
    for index, query in enumerate(QUERIES, 1):
        raw = (search_dir / f"{index}.json").read_bytes()
        response = json.loads(raw)
        items = response["items"]
        if len(items) > 30:
            raise ValueError("search page exceeds fixed per_page=30")
        pages.append({"query": query, "sha256": hashlib.sha256(raw).hexdigest(),
                      "total_count": response["total_count"], "returned": len(items)})
        kept = 0
        for rank, item in enumerate(items):
            repo = item["repository"]["full_name"]
            path = item["path"]
            if (repo in used_repos or (repo, path) in used_paths
                    or repo in EXCLUDED_REPOS
                    or repo.rsplit("/", 1)[-1].lower() in EXCLUDED_NAMES):
                continue
            used_repos.add(repo)
            used_paths.add((repo, path))
            selection.append({"query_index": index, "rank": rank,
                              "repo": repo, "path": path,
                              "search_blob_sha1": item["sha"]})
            kept += 1
            if kept == 5:
                break
    return {"format": "unswell-349-cue-search-selection-v1",
            "search_pages": pages, "selected": selection}


if __name__ == "__main__":
    if len(sys.argv) != 2:
        raise SystemExit("usage: select_transfer_v2.py SEARCH_DIR")
    print(json.dumps(select(Path(sys.argv[1])), indent=2))
