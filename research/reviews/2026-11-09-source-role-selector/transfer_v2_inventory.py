"""Inventory exact cue lines from signed, commit-pinned public source files."""

from __future__ import annotations

import hashlib
import json
import re
import sys
from datetime import datetime, timezone
from pathlib import Path


PHRASES = (
    "we hope this guide",
    "we hope this",
    "let's explore",
    "this is the place to start",
)
EXCLUDED_START = re.compile(r"[#>\-*+|`<\[!]")


def inventory(source_root: Path, selection_file: Path) -> dict:
    selected = json.loads(selection_file.read_bytes())
    sources = []
    targets = []
    for item in selected["sources"]:
        repo, file_rel, commit = item["repo"], item["path"], item["commit"]
        source_file = source_root / repo / file_rel
        row = {key: item[key] for key in ("repo", "path", "query_index", "commit", "url")}
        try:
            data = source_file.read_bytes()
        except FileNotFoundError:
            sources.append({**row, "omitted": "file missing"})
            continue
        if len(data) > 200 * 1024:
            sources.append({**row, "omitted": "file exceeds 200 KiB", "bytes": len(data)})
            continue
        try:
            data.decode("utf-8")
        except UnicodeDecodeError:
            sources.append({**row, "omitted": "file is not UTF-8"})
            continue
        blob_sha1 = hashlib.sha1(b"blob " + str(len(data)).encode() + b"\0" + data).hexdigest()
        if blob_sha1 != item["search_blob_sha1"]:
            raise ValueError(f"source blob differs from search result: {repo}/{file_rel}")
        source_sha256 = hashlib.sha256(data).hexdigest()
        sources.append({**row, "sha256": source_sha256, "git_blob_sha1": blob_sha1,
                        "bytes": len(data), "retrieved_at_utc": datetime.fromtimestamp(
                            source_file.stat().st_mtime, timezone.utc).isoformat()})
        phrase = PHRASES[item["query_index"] - 1]
        offset = 0
        kept = 0
        for line in data.split(b"\n"):
            start, end = offset, offset + len(line)
            offset = end + 1
            if len(line) > 500:
                continue
            quote = line.decode("utf-8")
            stripped = quote.lstrip()
            if not stripped or EXCLUDED_START.match(stripped):
                continue
            if re.search(re.escape(phrase), quote, re.I) is None:
                continue
            targets.append({"repo": repo, "path": file_rel,
                            "query_index": item["query_index"],
                            "source_sha256": source_sha256, "start": start,
                            "end": end, "quote": quote})
            kept += 1
            if kept == 3:
                break
    return {"format": "unswell-349-cue-transfer-inventory-v1",
            "selection_sha256": hashlib.sha256(selection_file.read_bytes()).hexdigest(),
            "sources": sources, "targets": targets}


if __name__ == "__main__":
    if len(sys.argv) != 3:
        raise SystemExit("usage: transfer_v2_inventory.py SOURCE_ROOT SELECTION_JSON")
    print(json.dumps(inventory(Path(sys.argv[1]), Path(sys.argv[2])), indent=2,
                     ensure_ascii=False))
