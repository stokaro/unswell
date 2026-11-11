#!/usr/bin/env python3
"""Fetch frozen development documents with exact Git blob verification."""

import argparse
import base64
import hashlib
import json
from pathlib import Path
import subprocess

from acquire import canonical, digest, write_new


def verified_blob(response, selected):
    if response.get("encoding") != "base64" or response.get("sha") != selected["blob_sha1"]:
        raise ValueError("Git blob response identity or encoding mismatch")
    data = base64.b64decode(response["content"], validate=False)
    if len(data) != selected["bytes"]:
        raise ValueError("document byte count differs from frozen tree")
    sha = hashlib.sha1(b"blob " + str(len(data)).encode() + b"\0" + data).hexdigest()
    if sha != selected["blob_sha1"]:
        raise ValueError("document Git blob SHA-1 differs from frozen tree")
    data.decode("utf-8", errors="strict")
    if b"\0" in data:
        raise ValueError("document contains NUL bytes")
    return data


def fetch(selection_path, source_directory, output_path):
    selection_raw = selection_path.read_bytes()
    selection = json.loads(selection_raw)
    if selection["status"] != "complete" or len(selection["selected"]) != 32:
        raise SystemExit("document selection is incomplete")
    entries = []
    for selected in selection["selected"]:
        if selected["selection_ordinal"] > 2:
            continue
        repo, sha = selected["repo"], selected["blob_sha1"]
        result = subprocess.run(["gh", "api", "-X", "GET",
                                 f"repos/{repo}/git/blobs/{sha}"],
                                check=True, capture_output=True)
        data = verified_blob(json.loads(result.stdout), selected)
        filename = (f"{selected['topic']}-{selected['function']}-"
                    f"{selected['selection_ordinal']:02d}.source")
        path = source_directory / filename
        if path.exists():
            if path.read_bytes() != data:
                raise ValueError(f"existing source differs: {path}")
        else:
            write_new(path, data)
        entries.append({"topic": selected["topic"], "function": selected["function"],
                        "selection_ordinal": selected["selection_ordinal"],
                        "repo": repo, "path": selected["path"],
                        "commit_sha": selected["commit_sha"], "blob_sha1": sha,
                        "bytes": len(data), "sha256": digest(data),
                        "local_file": filename})
        print(json.dumps({"source": filename, "repo": repo, "sha256": digest(data)}))
    manifest = {"format": "unswell-349-doc-function-development-sources-v1",
                "selection_sha256": digest(selection_raw), "entries": entries}
    write_new(output_path, canonical(manifest))
    print(json.dumps({"sources": len(entries), "manifest_sha256": digest(canonical(manifest))}))


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("selection", type=Path)
    parser.add_argument("source_directory", type=Path)
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    fetch(args.selection, args.source_directory, args.output)


if __name__ == "__main__":
    main()
