#!/usr/bin/env python3
"""Partition complete CLI dogfood evidence for bounded MCP replay."""

import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path


MAX_BYTES = 256 << 20
TOP_LEVEL = {
    "schema_version", "status", "manifest", "documents", "findings",
    "assessments", "gate", "errors", "features", "prepared_features",
}
PROBE_OUTCOMES = [
    "policy_failure", "pass", "policy_failure", "error",
    "policy_failure", "policy_failure", "policy_failure", "error",
]


def require(condition, message):
    if not condition:
        raise ValueError(message)


def digest(path):
    value = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            value.update(block)
    return value.hexdigest()


def read_json(path):
    with path.open(encoding="utf-8") as stream:
        return json.load(stream)


def write_json(path, value):
    with path.open("w", encoding="utf-8", newline="\n") as stream:
        json.dump(value, stream, ensure_ascii=False, separators=(",", ":"))
        stream.write("\n")


def identity(document):
    return document["name"], document["source_hash"]


def validate_coverage(expected, batches):
    actual = [identity(document)
              for batch in batches for document in batch["result"]["documents"]]
    require(Counter(actual) == Counter(expected),
            "MCP shards omit, duplicate, or change a source identity")
    require(len({name for name, _ in actual}) == len(actual),
            "MCP shards repeat a source path")


def split(report_path, directory):
    report = read_json(report_path)
    require(set(report) == TOP_LEVEL, "dogfood report shape changed")
    require(report["status"] == "complete" and report["gate"] ==
            {"passed": True, "reasons": []} and not report["errors"] and
            report["manifest"]["complete"] and
            report["manifest"]["include_source"],
            "dogfood report is not a complete passing source scan")
    documents = report["documents"]
    require(len(documents) >= 2, "dogfood scan needs at least two documents")
    names = [document["name"] for document in documents]
    require(len(set(names)) == len(names), "duplicate document path")
    all_names = set(names)
    for row in report["findings"]:
        require(row["primary"]["path"] in all_names,
                "finding points outside the full scan")
    for row in report["assessments"]:
        require(row["path"] in all_names,
                "assessment points outside the full scan")
    for kind in ("features", "prepared_features"):
        require(report[kind] is not None and
                Counter(row["path"] for row in report[kind]["sources"]) ==
                Counter(names), f"{kind} source coverage changed")
    directory.mkdir(parents=True, exist_ok=True)
    parts = []
    for number in (0, 1):
        selected = {name for index, name in enumerate(names) if index % 2 == number}
        piece = dict(report)
        piece["documents"] = [row for row in documents if row["name"] in selected]
        piece["findings"] = [row for row in report["findings"]
                             if row["primary"]["path"] in selected]
        piece["assessments"] = [row for row in report["assessments"]
                                if row["path"] in selected]
        for kind in ("features", "prepared_features"):
            piece[kind] = dict(report[kind])
            piece[kind]["sources"] = [row for row in report[kind]["sources"]
                                       if row["path"] in selected]
        path = directory / f"expected-{number}.json"
        write_json(path, piece)
        require(path.stat().st_size <= MAX_BYTES,
                f"MCP expected shard {number} exceeds 256 MiB")
        parts.append({"number": number, "sha256": digest(path),
                      "documents": [identity(row) for row in piece["documents"]]})
    require(Counter(item for part in parts for item in map(tuple, part["documents"])) ==
            Counter(map(identity, documents)), "partition lost a document")
    plan = {"format": "unswell-dogfood-mcp-shards-v1",
            "full_sha256": digest(report_path),
            "tool_commit": report["manifest"]["tool_commit"],
            "config_hash": report["manifest"]["config_hash"],
            "parts": parts}
    write_json(directory / "plan.json", plan)
    print(f"MCP dogfood split: {len(documents)} documents into "
          f"{[len(part['documents']) for part in parts]} bounded shards")


def combine(report_path, directory, output_path):
    plan = read_json(directory / "plan.json")
    require(plan["format"] == "unswell-dogfood-mcp-shards-v1" and
            digest(report_path) == plan["full_sha256"],
            "full dogfood evidence changed after splitting")
    batches = []
    probes = None
    expected = []
    for part in plan["parts"]:
        number = part["number"]
        require(number in (0, 1) and
                digest(directory / f"expected-{number}.json") == part["sha256"],
                "expected shard changed after splitting")
        expected.extend(map(tuple, part["documents"]))
        result = read_json(directory / f"mcp-{number}.json")
        require(result["version"] == "unswell-mcp-selfcheck-v1" and
                result["complete"] is True and
                result["tool_commit"] == plan["tool_commit"] and
                result["config_hash"] == plan["config_hash"] and
                "failure" not in result and result["repository_batches"],
                f"MCP shard {number} is incomplete or has the wrong build")
        outcomes = [row["outcome"] for row in result["probes"]]
        require(outcomes == PROBE_OUTCOMES,
                f"MCP shard {number} has a changed negative probe")
        if probes is None:
            probes = result["probes"]
        batches.extend(result["repository_batches"])
    require([part["number"] for part in plan["parts"]] == [0, 1],
            "MCP shard sequence changed")
    validate_coverage(expected, batches)
    combined = {"version": "unswell-mcp-selfcheck-v1", "complete": True,
                "tool_commit": plan["tool_commit"],
                "config_hash": plan["config_hash"],
                "repository_batches": batches, "probes": probes}
    write_json(output_path, combined)
    print(f"MCP dogfood combined: {len(expected)} documents, "
          f"{len(batches)} checked batches, {len(probes)} probes")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)
    for name in ("split", "combine"):
        command = sub.add_parser(name)
        command.add_argument("report", type=Path)
        command.add_argument("directory", type=Path)
        if name == "combine":
            command.add_argument("output", type=Path)
    args = parser.parse_args()
    if args.command == "split":
        split(args.report, args.directory)
    else:
        combine(args.report, args.directory, args.output)


if __name__ == "__main__":
    main()
