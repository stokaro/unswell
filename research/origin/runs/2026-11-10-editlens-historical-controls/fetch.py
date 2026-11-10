#!/usr/bin/env python3
"""Fetch exact historical Git blobs named by the frozen control frame."""

from __future__ import annotations

import argparse
import base64
import concurrent.futures
import hashlib
import json
from pathlib import Path
import subprocess


HERE = Path(__file__).resolve().parent


def check(row: dict, raw: bytes) -> None:
    blob = hashlib.sha1(f"blob {len(raw)}\0".encode() + raw).hexdigest()
    if (len(raw) != row["bytes"] or blob != row["blob_sha"] or
            hashlib.sha256(raw).hexdigest() != row["source_sha256"]):
        raise ValueError(f"Source hash changed: {row['source_id']}")
    text = raw.decode("utf-8")
    if "\0" in text:
        raise ValueError(f"NUL in source: {row['source_id']}")


def fetch(item: tuple[dict, Path]) -> tuple[str, bool]:
    row, output = item
    path = output / (row["source_id"] + ".source")
    if path.exists():
        check(row, path.read_bytes())
        return row["source_id"], False
    command = ["gh", "api", f"repos/{row['repository']}/git/blobs/{row['blob_sha']}"]
    process = subprocess.run(command, capture_output=True, text=True, check=False)
    if process.returncode:
        raise RuntimeError(process.stderr.strip())
    payload = json.loads(process.stdout)
    if payload.get("encoding") != "base64":
        raise ValueError(f"Unexpected Git blob encoding: {row['source_id']}")
    raw = base64.b64decode(payload["content"])
    check(row, raw)
    path.write_bytes(raw)
    return row["source_id"], True


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--sources-dir", type=Path, required=True)
    args = parser.parse_args()
    frame = json.loads((HERE / "frame.json").read_text())
    rows = frame["selected"]
    args.sources_dir.mkdir(parents=True, exist_ok=True)
    with concurrent.futures.ThreadPoolExecutor(max_workers=8) as executor:
        outcomes = list(executor.map(fetch, ((row, args.sources_dir) for row in rows)))
    if len({source_id for source_id, _ in outcomes}) != len(rows):
        raise ValueError("Duplicate source identity")
    print(json.dumps({"verified": len(outcomes),
                      "fetched": sum(fetched for _, fetched in outcomes)}, sort_keys=True))


if __name__ == "__main__":
    main()
