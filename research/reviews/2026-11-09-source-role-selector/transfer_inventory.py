"""Inventory pinned, page-disjoint README lines without consulting the selector."""

from __future__ import annotations

import hashlib
import json
import re
import sys
from datetime import datetime, timezone
from pathlib import Path


SOURCES = (
    ("nats-io/nats-server", "59bf235aea63881477dd485af3579be2ffc32b0c"),
    ("denoland/deno", "e7b39ebd55c85aaf6a29290c1f5a99e2e21e09dd"),
    ("neovim/neovim", "c868e1ab3d11ac2a0cc3c284b16808f3de5adf31"),
    ("syncthing/syncthing", "b34734376128492fa4e1b01412193de033e20613"),
    ("astral-sh/ruff", "47daf1c20b23a76e415fb0b3f7ed2b6100ba7164"),
    ("duckdb/duckdb", "45c50f87c4c3d67baaf31baef7c584ea9119ad8b"),
    ("qdrant/qdrant", "222b2e44ab1c3f17e8246396ead6558e45041df7"),
    ("grafana/loki", "83fa0b0504db2708757d2934d448eca3978c0398"),
)
MARKER = re.compile(r"\b(?:I|we|our|let's|this README|this page|this guide|this document|in this|now|next)\b", re.I)
EXCLUDED_START = re.compile(r"(?:[#>\-*+|`<\[!]|\d+\.)")


def inventory(root: Path) -> dict:
    sources = []
    targets = []
    for repo, commit in SOURCES:
        path = root / repo / "README.md"
        data = path.read_bytes()
        if len(data) > 200 * 1024:
            sources.append({"repo": repo, "omitted": "README exceeds 200 KiB"})
            continue
        data.decode("utf-8")
        sha256 = hashlib.sha256(data).hexdigest()
        git_blob = hashlib.sha1(b"blob " + str(len(data)).encode() + b"\0" + data).hexdigest()
        sources.append({
            "repo": repo,
            "commit": commit,
            "path": "README.md",
            "url": f"https://github.com/{repo}/blob/{commit}/README.md",
            "sha256": sha256,
            "git_blob_sha1": git_blob,
            "bytes": len(data),
            "retrieved_at_utc": datetime.fromtimestamp(path.stat().st_mtime, timezone.utc).isoformat(),
        })
        eligible = []
        offset = 0
        for line in data.splitlines(keepends=True):
            body = line.rstrip(b"\r\n")
            start, end = offset, offset + len(body)
            offset += len(line)
            if not 30 <= len(body) <= 500:
                continue
            quote = body.decode("utf-8")
            stripped = quote.lstrip()
            if not stripped or EXCLUDED_START.match(stripped) or not MARKER.search(quote):
                continue
            eligible.append({"start": start, "end": end, "quote": quote})
        for candidate in dict((item["start"], item) for item in eligible[:10] + eligible[-10:]).values():
            targets.append({"repo": repo, "source_sha256": sha256, **candidate})
    return {"format": "unswell-349-transfer-inventory-v1", "sources": sources, "targets": targets}


if __name__ == "__main__":
    if len(sys.argv) != 2:
        raise SystemExit("usage: transfer_inventory.py SOURCE_ROOT")
    print(json.dumps(inventory(Path(sys.argv[1])), indent=2, ensure_ascii=False))
