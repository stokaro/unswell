#!/usr/bin/env python3
"""Fetch exactly the commits and Git blobs selected before page review."""

import argparse
import base64
import hashlib
import json
import subprocess
from pathlib import Path
from urllib.parse import quote


def git_blob_sha1(data: bytes) -> str:
    return hashlib.sha1(f"blob {len(data)}\0".encode("ascii") + data).hexdigest()


def api(endpoint: str) -> dict:
    completed = subprocess.run(["gh", "api", "--method", "GET", endpoint],
                               capture_output=True, text=True, timeout=90,
                               check=False)
    if completed.returncode:
        raise RuntimeError(f"GitHub API exited {completed.returncode}: "
                           f"{completed.stderr.strip()[:300]}")
    return json.loads(completed.stdout)


def fetch(item: dict, source_dir: Path, number: int) -> dict:
    result = {"number": number, **item}
    endpoint = (f"repos/{item['repository']}/contents/"
                f"{quote(item['path'], safe='/')}?ref={item['commit']}")
    try:
        content = api(endpoint)
        if content.get("type") != "file" or content.get("encoding") != "base64":
            raise ValueError("selected path is not a base64-encoded file")
        raw = base64.b64decode("".join(content["content"].split()), validate=True)
        raw.decode("utf-8", errors="strict")
        sha1 = git_blob_sha1(raw)
        result.update({"source_sha256": hashlib.sha256(raw).hexdigest(),
                       "source_bytes": len(raw), "blob_sha1": sha1,
                       "api_blob_sha1": content.get("sha")})
        if sha1 != item["search_blob_sha1"] or sha1 != content.get("sha"):
            result["status"] = "invalid_blob_mismatch"
            return result
        filename = f"page-{number:03d}.source"
        (source_dir / filename).write_bytes(raw)
        result.update({"status": "valid", "source_file": filename})
    except (KeyError, ValueError, UnicodeError, RuntimeError,
            subprocess.TimeoutExpired) as error:
        result.update({"status": "invalid_or_operational_error",
                       "error": str(error)[:500]})
    return result


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--selection", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    args = parser.parse_args()
    if args.output_dir.exists():
        raise SystemExit("source output directory already exists")
    selection_raw = args.selection.read_bytes()
    selection = json.loads(selection_raw)
    if (selection["format"] != "unswell-349-atomic-claim-selection-v1" or
            len(selection["selected"]) != 9):
        raise SystemExit("unexpected source selection")
    args.output_dir.mkdir(parents=True)
    results = []
    for number, item in enumerate(selection["selected"], 1):
        result = fetch(item, args.output_dir, number)
        results.append(result)
        print(f"{number}/9 {item['repository']}: {result['status']}", flush=True)
    inventory = {"format": "unswell-349-atomic-claim-sources-v1",
                 "selection_sha256": hashlib.sha256(selection_raw).hexdigest(),
                 "sources": results}
    (args.output_dir / "inventory.json").write_text(
        json.dumps(inventory, indent=2, sort_keys=True) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
