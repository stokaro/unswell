#!/usr/bin/env python3
"""Stream and verify the pinned GEN edit corpus before feature scoring."""

from __future__ import annotations

import argparse
import collections
import gzip
import hashlib
import json
import sys
from pathlib import Path

HUMAN_SHA256 = "8ee87cb209bc72dfea449068db29325f3544f83f39c34a850c5542809d19f63f"
EXPECTED_BYTES = 1_521_045_885
DOMAINS = ("arxiv", "reddit", "story_generation", "wikihow", "wikipedia")
SALT = "unswell-gen-edit-v1"


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def norm(text: str) -> str:
    return " ".join(text.split())


def file_sha(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def human_groups(path: Path, expected_sha: str = HUMAN_SHA256) -> tuple[dict, dict, str]:
    require(file_sha(path) == expected_sha, "Human input SHA-256 changed")
    rows = {}
    by_domain = collections.defaultdict(list)
    with path.open("rb") as stream:
        for line in stream:
            row = json.loads(line)
            key = (row["source"], row["prompt_id"])
            require(row["source"] in DOMAINS and isinstance(row["prompt_id"], int),
                    "Invalid human group")
            require(key not in rows and row["label"] == 0 and row["text_type"] == "human" and
                    isinstance(row["text"], str) and row["text"], "Invalid human row")
            rows[key] = norm(row["text"])
            by_domain[row["source"]].append(row["prompt_id"])
    require(len(rows) == 5000 and all(len(by_domain[d]) == 1000 for d in DOMAINS),
            "Human source groups changed")
    assignment = {}
    for domain in DOMAINS:
        ids = sorted(by_domain[domain], key=lambda ident:
                     (hashlib.sha256(f"{SALT}|{domain}|{ident}".encode()).hexdigest(), ident))
        for index, ident in enumerate(ids):
            assignment[(domain, ident)] = "calibration" if index < 500 else "test"
    payload = [[domain, ident, assignment[(domain, ident)]] for domain, ident in sorted(rows)]
    split_sha = hashlib.sha256(json.dumps(payload, separators=(",", ":")).encode()).hexdigest()
    return rows, assignment, split_sha


def acquire(input_stream, human_path: Path, selected_path: Path,
            manifest_path: Path, expected_bytes: int = EXPECTED_BYTES,
            expected_human_sha: str = HUMAN_SHA256) -> dict:
    humans, roles, split_sha = human_groups(human_path, expected_human_sha)
    digest = hashlib.sha256()
    bytes_read = 0
    rows = 0
    selected = 0
    counts = collections.Counter()
    selected_counts = collections.Counter()
    edit_keys = set()
    duplicates = 0
    source_mismatches = collections.Counter()
    source_groups = set()
    selected_groups = set()
    selected_raw_sha = hashlib.sha256()
    selected_path.parent.mkdir(parents=True, exist_ok=True)
    with selected_path.open("wb") as raw_output:
        with gzip.GzipFile(filename="", mode="wb", fileobj=raw_output, mtime=0) as output:
            for line in input_stream:
                digest.update(line)
                bytes_read += len(line)
                require(line.endswith(b"\n"), "Edited JSONL row lacks a newline")
                row = json.loads(line)
                rows += 1
                domain, ident = row["source"], row["prompt_id"]
                key = (domain, ident)
                require(domain in DOMAINS and key in humans and
                        isinstance(row.get("seed_id"), int) and
                        DOMAINS.index(domain) * 200 <= row["seed_id"] <
                        (DOMAINS.index(domain) + 1) * 200,
                        f"Invalid edited source group at row {rows}: {domain}, {ident}, "
                        f"seed {row.get('seed_id')}")
                require(row["label"] == 2 and row["text_type"] == "ai_edited" and
                        isinstance(row["text"], str) and row["text"] and
                        isinstance(row["seed_text"], str) and row["seed_text"],
                        "Invalid edited text or label")
                require(row["target_ratio"] in (30, 50, 70) and
                        isinstance(row["edit_ratio"], (int, float)) and
                        0 <= row["edit_ratio"] <= 100 and
                        isinstance(row["edit_category"], str) and row["edit_category"] and
                        isinstance(row["edit_prompt"], str) and row["edit_prompt"] and
                        isinstance(row["model"], str) and row["model"],
                        "Invalid edited condition")
                source_groups.add(key)
                counts[(domain, row["model"], row["edit_category"], row["target_ratio"])] += 1
                if norm(row["seed_text"]) != humans[key]:
                    source_mismatches[domain] += 1
                    continue
                if row["target_ratio"] != 30:
                    continue
                variant = (key, row["model"], row["edit_prompt"], row["target_ratio"])
                if variant in edit_keys:
                    duplicates += 1
                edit_keys.add(variant)
                output.write(line)
                selected_raw_sha.update(line)
                selected += 1
                selected_groups.add(key)
                selected_counts[(domain, row["model"], row["edit_category"], roles[key])] += 1
    require(bytes_read == expected_bytes, f"Edited input size changed: {bytes_read}")
    require(rows > 0 and selected > 0, "No edited rows selected")
    manifest = {
        "format": "unswell-gen-edit-acquisition-v1",
        "input_bytes": bytes_read,
        "input_sha256": digest.hexdigest(),
        "input_rows": rows,
        "human_sha256": expected_human_sha,
        "human_groups": len(humans),
        "split_sha256": split_sha,
        "source_groups": len(source_groups),
        "selected_rows": selected,
        "selected_groups": len(selected_groups),
        "selected_raw_sha256": selected_raw_sha.hexdigest(),
        "selected_gzip_sha256": file_sha(selected_path),
        "selected_gzip_bytes": selected_path.stat().st_size,
        "duplicate_selected_condition_keys": duplicates,
        "source_mismatches_by_domain": dict(sorted(source_mismatches.items())),
        "input_counts": [[*key, count] for key, count in sorted(counts.items())],
        "selected_counts": [[*key, count] for key, count in sorted(selected_counts.items())],
    }
    manifest_path.write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n")
    return manifest


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--human", type=Path, required=True)
    parser.add_argument("--selected", type=Path, required=True)
    parser.add_argument("--manifest", type=Path, required=True)
    args = parser.parse_args()
    result = acquire(sys.stdin.buffer, args.human, args.selected, args.manifest)
    print(json.dumps({key: result[key] for key in (
        "input_bytes", "input_sha256", "input_rows", "source_groups",
        "selected_rows", "selected_groups", "duplicate_selected_condition_keys",
        "source_mismatches_by_domain")}, sort_keys=True))


if __name__ == "__main__":
    main()
