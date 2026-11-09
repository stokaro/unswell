#!/usr/bin/env python3
"""Pin selected GitHub README bytes to a default-branch commit."""

import argparse
import base64
import hashlib
import json
import sys
from pathlib import Path

# The frozen selector is named select.py. Keep it from shadowing Python's
# standard-library select module when subprocess imports it below.
_script_dir = Path(__file__).resolve().parent
sys.path = [entry for entry in sys.path
            if Path(entry or ".").resolve() != _script_dir]

import subprocess
from urllib.parse import quote


def api(endpoint: str) -> dict:
    completed = subprocess.run(
        ["gh", "api", "--method", "GET", endpoint],
        capture_output=True,
        text=True,
        timeout=90,
        check=False,
    )
    if completed.returncode:
        raise RuntimeError(
            f"gh api exited {completed.returncode}: "
            f"{completed.stderr.strip()[:300]}"
        )
    return json.loads(completed.stdout)


def git_blob_sha1(data: bytes) -> str:
    header = f"blob {len(data)}\0".encode("ascii")
    return hashlib.sha1(header + data).hexdigest()


def fetch(item: dict, output_dir: Path, number: int) -> dict:
    result = {"number": number, **item}
    repo = item["repo"]
    try:
        repo_info = api(f"repos/{repo}")
        branch = repo_info["default_branch"]
        commit = api(f"repos/{repo}/commits/{quote(branch, safe='')}")
        commit_sha = commit["sha"]
        result.update({"default_branch": branch, "commit_sha": commit_sha})
        path = quote(item["path"], safe="/")
        content = api(f"repos/{repo}/contents/{path}?ref={commit_sha}")
        if content.get("type") != "file" or content.get("encoding") != "base64":
            raise ValueError("selected path is not a base64-encoded file")
        data = base64.b64decode("".join(content["content"].split()), validate=True)
        data.decode("utf-8", errors="strict")
        blob_sha1 = git_blob_sha1(data)
        result.update({
            "default_branch_blob_sha1": blob_sha1,
            "source_sha256": hashlib.sha256(data).hexdigest(),
            "source_bytes": len(data),
        })
        if blob_sha1 != item["search_blob_sha1"]:
            result["status"] = "invalid_search_blob_mismatch"
            return result
        name = f"page-{number:03d}.source"
        (output_dir / name).write_bytes(data)
        result.update({"status": "valid", "source_file": name})
    except (KeyError, ValueError, UnicodeError, RuntimeError,
            subprocess.TimeoutExpired) as error:
        result.update({"status": "invalid_or_operational_error",
                       "error": str(error)[:500]})
    return result


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--selection", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    args = parser.parse_args()
    if args.output_dir.exists():
        raise ValueError("output directory already exists")
    selection = json.loads(args.selection.read_text())
    if selection["format"] != "unswell-349-fresh-source-function-selection-v1":
        raise ValueError("unexpected selection format")
    if len(selection["selected"]) != 9:
        raise ValueError("expected nine selected sources")
    args.output_dir.mkdir(parents=True)
    results = []
    for number, item in enumerate(selection["selected"], 1):
        result = fetch(item, args.output_dir, number)
        results.append(result)
        print(f"{number}/9 {item['repo']}: {result['status']}", flush=True)
    inventory = {
        "format": "unswell-349-fresh-source-function-sources-v1",
        "selection_sha256": hashlib.sha256(args.selection.read_bytes()).hexdigest(),
        "sources": results,
    }
    (args.output_dir / "inventory.json").write_text(
        json.dumps(inventory, indent=2, sort_keys=True) + "\n"
    )


if __name__ == "__main__":
    main()
