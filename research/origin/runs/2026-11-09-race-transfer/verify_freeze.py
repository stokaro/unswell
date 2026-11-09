#!/usr/bin/env python3
"""Verify the RACE transfer packet before any selected-document inference."""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]


def digest(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            h.update(block)
    return h.hexdigest()


def verify_file(path: Path, expected: str) -> None:
    if digest(path) != expected:
        raise ValueError(f"Frozen file changed: {path}")


def verify(input_path: Path, arb_path: Path) -> dict:
    freeze = json.loads((HERE / "freeze.json").read_text(encoding="utf-8"))
    for relative, expected in freeze["files"].items():
        verify_file(ROOT / relative, expected)
    verify_file(input_path, freeze["input_sha256"])
    verify_file(arb_path, freeze["arb_sha256"])
    selection = json.loads((HERE / "selection.json").read_text(encoding="utf-8"))
    if len(selection["source_ids"]) != 60 or \
            len(set(selection["source_ids"])) != 60 or \
            len(selection["document_hashes"]) != 300 or \
            selection["input_sha256"] != freeze["input_sha256"]:
        raise ValueError("Frozen source selection changed")
    rows = [json.loads(line) for line in input_path.read_text(encoding="utf-8").splitlines()]
    if len(rows) != 300 or len({row["item_id"] for row in rows}) != 300:
        raise ValueError("Frozen input count or identity changed")
    for row in rows:
        actual = hashlib.sha256(row["article"].encode("utf-8")).hexdigest()
        if selection["document_hashes"].get(row["item_id"]) != actual:
            raise ValueError("Frozen raw text changed")
    return {"files": len(freeze["files"]), "sources": 60, "documents": 300,
            "input_sha256": freeze["input_sha256"]}


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--input", required=True, type=Path)
    parser.add_argument("--arb", required=True, type=Path)
    args = parser.parse_args()
    print(json.dumps(verify(args.input, args.arb)))


if __name__ == "__main__":
    main()
