#!/usr/bin/env python3
"""Fetch the fixed historical controls and verify their exact Git blobs."""

from __future__ import annotations

import argparse
from concurrent.futures import ThreadPoolExecutor
import hashlib
import json
from pathlib import Path
import urllib.error
import urllib.parse
import urllib.request


HERE = Path(__file__).resolve().parent
DOCUMENTS = HERE.parent / "2026-11-10-editlens-new-controls" / "documents.json"


def check(row: dict, raw: bytes) -> None:
    expected_blob = hashlib.sha1(f"blob {len(raw)}\0".encode() + raw).hexdigest()
    if (len(raw) != row["bytes"] or expected_blob != row["blob_sha"]
            or hashlib.sha256(raw).hexdigest() != row["source_sha256"]):
        raise ValueError(f"Source identity mismatch: {row['source_id']}")
    text = raw.decode("utf-8")
    if "\0" in text:
        raise ValueError(f"NUL in source: {row['source_id']}")


def fetch(row: dict, destination: Path) -> None:
    target = destination / f"{row['source_id']}.source"
    if target.exists():
        check(row, target.read_bytes())
        return
    url = ("https://raw.githubusercontent.com/" + row["repository"] + "/"
           + row["commit"] + "/" + urllib.parse.quote(row["path"], safe="/"))
    request = urllib.request.Request(url, headers={"User-Agent": "unswell-research/1"})
    try:
        with urllib.request.urlopen(request, timeout=40) as response:
            raw = response.read(200_001)
    except urllib.error.HTTPError as error:
        raise RuntimeError(f"Source fetch failed for {row['source_id']}: HTTP {error.code}") from error
    check(row, raw)
    target.write_bytes(raw)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--destination", type=Path, required=True)
    parser.add_argument("--workers", type=int, default=12)
    args = parser.parse_args()
    rows = json.loads(DOCUMENTS.read_text())
    if len(rows) != 761 or len({row["source_id"] for row in rows}) != 761:
        raise ValueError("Unexpected historical control frame")
    args.destination.mkdir(parents=True, exist_ok=True)
    with ThreadPoolExecutor(max_workers=args.workers) as pool:
        list(pool.map(lambda row: fetch(row, args.destination), rows))
    print(json.dumps({"verified": len(rows),
                      "bytes": sum((args.destination / f"{row['source_id']}.source").stat().st_size
                                   for row in rows)}, sort_keys=True))


if __name__ == "__main__":
    main()
