#!/usr/bin/env python3
"""Verify the exact selected Git blobs and original UTF-8 source bytes."""

import hashlib
import json
from pathlib import Path

from fetch_sources import git_blob_sha1
from verify_selection import verify as verify_selection


def verify(packet_dir: Path) -> dict:
    selection = verify_selection(packet_dir)
    selection_raw = (packet_dir / "selection.json").read_bytes()
    source_dir = packet_dir / "sources"
    inventory = json.loads((source_dir / "inventory.json").read_text(encoding="utf-8"))
    if inventory["format"] != "unswell-349-atomic-claim-sources-v1":
        raise ValueError("wrong source inventory format")
    if inventory["selection_sha256"] != hashlib.sha256(selection_raw).hexdigest():
        raise ValueError("source selection hash changed")
    if len(inventory["sources"]) != len(selection["selected"]):
        raise ValueError("source count changed")
    expected_names = {f"page-{number:03d}.source" for number in
                      range(1, len(selection["selected"]) + 1)}
    actual_names = {path.name for path in source_dir.glob("page-*.source")}
    if actual_names != expected_names:
        raise ValueError("source file set changed")
    for number, (expected, record) in enumerate(
            zip(selection["selected"], inventory["sources"]), 1):
        for key, value in expected.items():
            if record.get(key) != value:
                raise ValueError(f"source identity changed at page {number}")
        if record["number"] != number or record["status"] != "valid":
            raise ValueError(f"source page {number} is not valid")
        name = f"page-{number:03d}.source"
        if record["source_file"] != name:
            raise ValueError(f"source filename changed at page {number}")
        raw = (source_dir / name).read_bytes()
        raw.decode("utf-8", errors="strict")
        if (len(raw) != record["source_bytes"] or
                hashlib.sha256(raw).hexdigest() != record["source_sha256"] or
                git_blob_sha1(raw) != record["blob_sha1"] or
                record["blob_sha1"] != record["api_blob_sha1"] or
                record["blob_sha1"] != expected["search_blob_sha1"]):
            raise ValueError(f"source bytes changed at page {number}")
    return inventory


if __name__ == "__main__":
    result = verify(Path(__file__).resolve().parent)
    print(f"verified {len(result['sources'])} pinned UTF-8 source pages")
