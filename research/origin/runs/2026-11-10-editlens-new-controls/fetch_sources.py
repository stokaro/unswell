#!/usr/bin/env python3
"""Fetch the prespecified historical Git blobs and record source integrity."""

from __future__ import annotations

import argparse
import base64
import concurrent.futures
import hashlib
import json
from pathlib import Path
import subprocess
import threading
import time


INVENTORY_SHA256 = "6051df6a4af3f0aebe8394c982996c501a465999236724b5df8e56795ea3d194"
SEED = b"unswell-techdoc-new-document-v1"


def source_id(row: dict) -> str:
    return hashlib.sha256("\0".join((row["repository"], row["commit"],
                                        row["path"])).encode()).hexdigest()


def doc_rank(row: dict) -> tuple[str, str]:
    key = row["repository"].lower().encode() + b"\0" + row["path"].encode() + b"\0" + SEED
    return hashlib.sha256(key).hexdigest(), row["path"]


def check_blob(row: dict, raw: bytes) -> dict:
    blob_sha = hashlib.sha1(f"blob {len(raw)}\0".encode() + raw).hexdigest()
    if len(raw) != row["bytes"] or blob_sha != row["blob_sha"]:
        raise ValueError("Git blob identity mismatch")
    text = raw.decode("utf-8")
    if "\0" in text:
        raise ValueError("NUL in source")
    return {**row, "source_id": source_id(row),
            "source_sha256": hashlib.sha256(raw).hexdigest()}


def fetch(row: dict, sources: Path) -> dict:
    sid = source_id(row)
    destination = sources / f"{sid}.source"
    try:
        if destination.exists():
            return check_blob(row, destination.read_bytes())
        path = f"repos/{row['repository']}/git/blobs/{row['blob_sha']}"
        for attempt in range(2):
            process = subprocess.run(["gh", "api", path], capture_output=True,
                                     text=True, check=False, timeout=90)
            if process.returncode == 0:
                break
            if attempt == 0 and "HTTP 404" not in process.stderr:
                time.sleep(1)
                continue
            raise RuntimeError(process.stderr.strip()[:300])
        payload = json.loads(process.stdout)
        if payload.get("encoding") != "base64":
            raise ValueError("Unexpected blob encoding")
        raw = base64.b64decode(payload["content"], validate=False)
        result = check_blob(row, raw)
        destination.write_bytes(raw)
        return result
    except Exception as error:
        return {**row, "source_id": sid, "error": str(error)}


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--inventory", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--sources-dir", type=Path, required=True)
    parser.add_argument("--workers", type=int, default=8)
    args = parser.parse_args()
    digest = hashlib.sha256(args.inventory.read_bytes()).hexdigest()
    if digest != INVENTORY_SHA256:
        raise ValueError("Inventory changed")
    inventory = json.loads(args.inventory.read_text())
    rows = []
    for repository in inventory["rows"]:
        if repository.get("error"):
            continue
        for doc in sorted((dict(item, repository=repository["repository"],
                                commit=repository["commit"], language=repository["language"],
                                committed_at=repository["committed_at"])
                           for item in repository["documents"]), key=doc_rank)[:8]:
            rows.append(doc)
    args.sources_dir.mkdir(parents=True, exist_ok=True)
    completed = 0
    lock = threading.Lock()

    def work(row: dict) -> dict:
        nonlocal completed
        result = fetch(row, args.sources_dir)
        with lock:
            completed += 1
            if completed % 50 == 0 or completed == len(rows):
                print(f"verified {completed}/{len(rows)} candidate blobs", flush=True)
        return result

    with concurrent.futures.ThreadPoolExecutor(max_workers=args.workers) as pool:
        results = list(pool.map(work, rows))
    output = {"format": "unswell-origin-new-historical-blobs-v1",
              "inventory_sha256": digest, "document_seed": SEED.decode(),
              "rows": results}
    args.output.write_text(json.dumps(output, sort_keys=True, indent=2) + "\n")
    print(json.dumps({"candidates": len(results),
                      "verified": sum("error" not in row for row in results),
                      "errors": sum("error" in row for row in results)}, sort_keys=True))


if __name__ == "__main__":
    main()
