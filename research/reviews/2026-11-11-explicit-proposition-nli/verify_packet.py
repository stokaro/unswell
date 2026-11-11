#!/usr/bin/env python3
"""Verify frozen source bindings for the explicit-proposition packet."""

from __future__ import annotations

from collections import Counter
import argparse
import hashlib
import json
from pathlib import Path

HERE = Path(__file__).resolve().parent
SOURCE_RESEARCH = HERE.parent / "2026-11-11-clause-role-transfer"


def digest(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def verify(packet: dict, selection: dict, inventory: dict, sources: Path,
           inventory_sha256: str) -> None:
    if packet["format"] != "unswell-349-explicit-proposition-nli-v1":
        raise ValueError("packet format")
    if packet["source_inventory_sha256"] != inventory_sha256:
        raise ValueError("source inventory identity")
    selected = {row["number"]: row for row in selection["selected"]}
    source_rows = {row["number"]: row for row in inventory["records"]}
    if len(selected) != 20 or len(source_rows) != 20 or len(packet["cases"]) != 16:
        raise ValueError("selection or case count")
    seen = set()
    labels = Counter()
    for case in packet["cases"]:
        if case["id"] in seen:
            raise ValueError("duplicate ID")
        seen.add(case["id"])
        number = case["source_number"]
        item, source = selected[number], source_rows[number]
        if source["status"] != "fetched" or any(
            case[key] != item[key] for key in ("repo", "path", "commit", "url")
        ) or case["source_sha256"] != source["source_sha256"]:
            raise ValueError("source identity")
        data = (sources / source["source_file"]).read_bytes()
        if digest(data) != source["source_sha256"]:
            raise ValueError("source bytes changed")
        blob = hashlib.sha1(b"blob " + str(len(data)).encode() + b"\0" + data).hexdigest()
        if blob != source["source_blob_sha1"] or blob != item["search_blob_sha1"]:
            raise ValueError("source Git blob changed")
        spans = case["evidence_spans"]
        if not spans or case["premise"] != "\n".join(span["quote"] for span in spans):
            raise ValueError("premise is not the bound excerpts")
        for span in spans:
            if not 0 <= span["start"] < span["end"] <= len(data):
                raise ValueError("evidence span out of range")
            if data[span["start"]:span["end"]] != span["quote"].encode():
                raise ValueError("evidence quote changed")
        if not case["hypothesis"].strip():
            raise ValueError("empty hypothesis")
        labels[case["expected_relation"]] += 1
    if labels != Counter({"entailment": 6, "neutral": 6, "contradiction": 4}):
        raise ValueError("class balance changed")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--sources", type=Path, required=True)
    args = parser.parse_args()
    packet = json.loads((HERE / "packet.json").read_text())
    selection = json.loads((SOURCE_RESEARCH / "selection.json").read_text())
    path = SOURCE_RESEARCH / "source-inventory.json"
    inventory = json.loads(path.read_text())
    verify(packet, selection, inventory, args.sources, digest(path.read_bytes()))
    print("verified 16 source-bound explicit proposition pairs")


if __name__ == "__main__":
    main()
