#!/usr/bin/env python3
"""Bind source-only editorial labels to immutable original UTF-8 byte spans."""

import argparse
import json
from pathlib import Path

from acquire import canonical, digest, write_new


def locate(data, quote):
    needle = quote.encode("utf-8")
    start = data.find(needle)
    if not needle or start < 0 or data.find(needle, start + 1) >= 0:
        raise ValueError(f"quote missing or repeated: {quote[:80]!r}")
    return {"start": start, "end": start + len(needle), "quote": quote}


def materialize(spec_path, manifest_path, source_dir):
    spec_raw = spec_path.read_bytes()
    spec = json.loads(spec_raw)
    manifest_raw = manifest_path.read_bytes()
    entries = json.loads(manifest_raw)["entries"]
    sources = {entry["local_file"]: entry for entry in entries}
    if len(entries) != 16 or len(sources) != 16:
        raise ValueError("expected 16 distinct development sources")
    dispositions = spec["source_dispositions"]
    if set(dispositions) != set(sources):
        raise ValueError("source dispositions must cover every development source")
    if any(value not in ("reviewed_english", "out_of_domain_chinese")
           for value in dispositions.values()):
        raise ValueError("invalid source disposition")
    source_bytes = {}
    for name, entry in sources.items():
        data = (source_dir / name).read_bytes()
        if digest(data) != entry["sha256"] or len(data) != entry["bytes"]:
            raise ValueError(f"source identity mismatch: {name}")
        data.decode("utf-8")
        source_bytes[name] = data

    labels = []
    ids = set()
    for item in spec["labels"]:
        name = item["source"]
        if item["id"] in ids:
            raise ValueError(f"duplicate label ID: {item['id']}")
        ids.add(item["id"])
        if dispositions[name] != "reviewed_english":
            raise ValueError(f"label on out-of-domain source: {name}")
        if item["kind"] not in ("positive", "control"):
            raise ValueError(f"invalid label kind: {item['kind']}")
        if item["kind"] == "positive" and not item.get("evidence_quote"):
            raise ValueError(f"positive has no source evidence: {item['id']}")
        label = {key: value for key, value in item.items()
                 if key not in ("quote", "evidence_quote")}
        label["span"] = locate(source_bytes[name], item["quote"])
        evidence = item.get("evidence_quote")
        label["evidence_span"] = (locate(source_bytes[name], evidence)
                                  if evidence else None)
        label["repo"] = sources[name]["repo"]
        label["function"] = sources[name]["function"]
        label["topic"] = sources[name]["topic"]
        label["source_sha256"] = sources[name]["sha256"]
        labels.append(label)

    considered = []
    for item in spec.get("considered_not_positive", []):
        if item["id"] in ids:
            raise ValueError(f"duplicate considered ID: {item['id']}")
        ids.add(item["id"])
        name = item["source"]
        if dispositions[name] != "reviewed_english":
            raise ValueError(f"considered item on out-of-domain source: {name}")
        considered.append({"id": item["id"], "source": name,
                           "reason": item["reason"],
                           "span": locate(source_bytes[name], item["quote"]),
                           "source_sha256": sources[name]["sha256"]})

    positives = [item for item in labels if item["kind"] == "positive"]
    controls = [item for item in labels if item["kind"] == "control"]
    distinct_positive_repos = len({item["repo"] for item in positives})
    retain_controls = sum(item["disposition"] == "retain" for item in controls)
    counts = {"selected_development": 16,
              "reviewed_english": sum(x == "reviewed_english" for x in dispositions.values()),
              "out_of_domain": sum(x == "out_of_domain_chinese" for x in dispositions.values()),
              "positives": len(positives),
              "positive_repositories": distinct_positive_repos,
              "controls": len(controls),
              "retain_controls": retain_controls}
    gate = (counts["positives"] >= 16 and distinct_positive_repos >= 8
            and counts["controls"] >= 32 and retain_controls >= 8)
    return {"format": "unswell-349-doc-function-source-review-v1",
            "reviewer": spec["reviewer"],
            "spec_sha256": digest(spec_raw),
            "development_sources_sha256": digest(manifest_raw),
            "source_dispositions": dispositions,
            "counts": counts,
            "acquisition_gate_passed": gate,
            "considered_not_positive": considered,
            "labels": labels}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("spec", type=Path)
    parser.add_argument("manifest", type=Path)
    parser.add_argument("source_dir", type=Path)
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    result = materialize(args.spec, args.manifest, args.source_dir)
    data = canonical(result)
    write_new(args.output, data)
    print(json.dumps({**result["counts"],
                      "acquisition_gate_passed": result["acquisition_gate_passed"],
                      "sha256": digest(data)}))


if __name__ == "__main__":
    main()
