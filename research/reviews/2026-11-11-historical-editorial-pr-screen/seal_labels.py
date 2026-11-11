#!/usr/bin/env python3
"""Resolve source-only quotes to exact original UTF-8 byte spans."""

import argparse
import csv
import hashlib
import json
from pathlib import Path


def span(source: bytes, quote: str, field: str, case: str) -> dict:
    target = quote.encode("utf-8")
    start = source.find(target)
    if start < 0 or source.find(target, start + 1) >= 0:
        raise ValueError(f"{case}: {field} must occur exactly once")
    return {"start": start, "end": start + len(target), "quote": quote}


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--sources", type=Path, required=True)
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args()
    root = Path(__file__).resolve().parent
    with (root / "selection.tsv").open(newline="", encoding="utf-8") as stream:
        rows = list(csv.DictReader(stream, delimiter="\t"))
    manifest = json.loads((args.sources / "manifest.json").read_text(encoding="utf-8"))
    sources = {f"{row['repo']}#{row['pr']}": row for row in manifest}
    if {(row["repo"], row["pr"]) for row in rows} != {
        (row["repo"], row["pr"]) for row in manifest
    }:
        raise ValueError("selection and source manifest differ")
    labels = json.loads((root / "source_only_seed.json").read_text(encoding="utf-8"))
    if len(labels) != len({row["pr"] for row in labels}):
        raise ValueError("duplicate label")
    output = []
    for label in labels:
        case = label["pr"]
        if case not in sources:
            raise ValueError(f"unknown case: {case}")
        row = sources[case]
        source = (args.sources / row["source_file"]).read_bytes()
        if hashlib.sha256(source).hexdigest() != row["sha256"]:
            raise ValueError(f"source SHA-256 mismatch: {case}")
        item = {"pr": case, "source_sha256": row["sha256"], "decision": label["decision"]}
        for field in ("quote", "retain_quote", "evidence_quote"):
            if field in label:
                item[field.replace("quote", "span")] = span(source, label[field], field, case)
        for field in ("diagnosis", "retain_reason", "reason"):
            if field in label:
                item[field] = label[field]
        output.append(item)
    exposed = {
        "freedomofpress/encryption-works#102",
        "trinodb/trino#14520",
    }
    if {row["pr"] for row in output} != set(sources) - exposed:
        raise ValueError("labels do not cover every primary source")
    args.out.write_text(json.dumps(output, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"Sealed {len(output)} source-only decisions")


if __name__ == "__main__":
    main()
