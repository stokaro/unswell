#!/usr/bin/env python3
"""Resolve source-only review quotes to original UTF-8 byte spans."""

import argparse
import json
from pathlib import Path

from acquire import canonical, digest, write_new


def locate(data, quote):
    needle = quote.encode("utf-8")
    start = data.find(needle)
    if start < 0 or data.find(needle, start + 1) >= 0:
        raise ValueError(f"source quote missing or repeated: {quote[:80]!r}")
    return {"start": start, "end": start + len(needle), "quote": quote}


def materialize(spec_path, source_manifest_path, source_directory, output_path):
    spec_raw = spec_path.read_bytes()
    spec = json.loads(spec_raw)
    source_manifest_raw = source_manifest_path.read_bytes()
    sources = {entry["local_file"]: entry for entry in
               json.loads(source_manifest_raw)["entries"]}
    reviewed = spec["reviewed_pages"]
    if sorted(reviewed) != sorted(sources):
        raise ValueError("review must account for all 12 development pages")
    source_data = {}
    for name, entry in sources.items():
        data = (source_directory / name).read_bytes()
        if digest(data) != entry["sha256"]:
            raise ValueError(f"source SHA-256 mismatch: {name}")
        source_data[name] = data
    labels = []
    ids = set()
    for item in spec["labels"]:
        if item["id"] in ids:
            raise ValueError(f"duplicate label ID: {item['id']}")
        ids.add(item["id"])
        name = item["source"]
        data = source_data[name]
        label = {key: value for key, value in item.items()
                 if key not in ("quote", "evidence_quote")}
        label["span"] = locate(data, item["quote"])
        evidence = item.get("evidence_quote")
        label["evidence_span"] = locate(data, evidence) if evidence else None
        label["repo"] = sources[name]["repo"]
        label["source_sha256"] = sources[name]["sha256"]
        labels.append(label)
    positives = {item["id"] for item in labels if item["kind"] == "positive"}
    for label in labels:
        if label["kind"] == "control" and label["paired_positive"] not in positives:
            raise ValueError(f"unpaired control: {label['id']}")
    result = {"format": "unswell-349-source-only-review-v1",
              "reviewer": "root Codex assistant; maintainer-accepted assistant review under ADR 0041",
              "spec_sha256": digest(spec_raw),
              "source_manifest_sha256": digest(source_manifest_raw),
              "reviewed_pages": reviewed, "labels": labels}
    write_new(output_path, canonical(result))
    print(json.dumps({"pages": len(reviewed), "positives": len(positives),
                      "controls": len(labels) - len(positives),
                      "sha256": digest(canonical(result))}))


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("spec", type=Path)
    parser.add_argument("source_manifest", type=Path)
    parser.add_argument("source_directory", type=Path)
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    materialize(args.spec, args.source_manifest, args.source_directory, args.output)


if __name__ == "__main__":
    main()
