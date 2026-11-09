#!/usr/bin/env python3
"""Fetch and pin the twelve preselected transfer README files."""

import argparse
import hashlib
import json
from pathlib import Path

from fetch_sources import fetch


FROZEN_FETCH_CORE_SHA256 = "1c9f84d98d430b9c3a84abff801732f223f472a2f5d767d8aa2ad972d82843a0"


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--selection", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    args = parser.parse_args()
    if args.output_dir.exists():
        raise ValueError("output directory already exists")
    core = Path(__file__).with_name("fetch_sources.py")
    if hashlib.sha256(core.read_bytes()).hexdigest() != FROZEN_FETCH_CORE_SHA256:
        raise ValueError("frozen source fetcher changed")
    selection_bytes = args.selection.read_bytes()
    selection = json.loads(selection_bytes)
    if (selection["format"] != "unswell-349-source-function-transfer-v2-selection-v1"
            or len(selection["selected"]) != 12):
        raise ValueError("unexpected twelve-page selection")
    args.output_dir.mkdir(parents=True)
    results = []
    for number, item in enumerate(selection["selected"], 1):
        result = fetch(item, args.output_dir, number)
        results.append(result)
        print(f"{number}/12 {item['repo']}: {result['status']}", flush=True)
    inventory = {
        "format": "unswell-349-source-function-transfer-v2-sources-v1",
        "selection_sha256": hashlib.sha256(selection_bytes).hexdigest(),
        "sources": results,
    }
    (args.output_dir / "inventory.json").write_text(
        json.dumps(inventory, indent=2, sort_keys=True) + "\n"
    )


if __name__ == "__main__":
    main()
