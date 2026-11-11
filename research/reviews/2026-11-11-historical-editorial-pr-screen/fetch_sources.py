#!/usr/bin/env python3
"""Fetch and verify frozen pre-edit public documentation sources."""

import argparse
import base64
import csv
import hashlib
import json
from pathlib import Path
import subprocess
from urllib.parse import quote


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args()
    root = Path(__file__).resolve().parent
    with (root / "selection.tsv").open(newline="", encoding="utf-8") as stream:
        rows = list(csv.DictReader(stream, delimiter="\t"))
    args.out.mkdir(parents=True, exist_ok=True)
    manifest = []
    for row in rows:
        endpoint = (
            f"repos/{row['repo']}/contents/{quote(row['path'], safe='/')}"
            f"?ref={row['base_sha']}"
        )
        response = subprocess.run(
            ["gh", "api", endpoint],
            check=True,
            capture_output=True,
            text=True,
        )
        payload = json.loads(response.stdout)
        if payload.get("encoding") != "base64" or not payload.get("content"):
            raise ValueError(f"missing base64 content: {row['repo']}#{row['pr']}")
        source = base64.b64decode(payload["content"], validate=False)
        blob = hashlib.sha1(
            f"blob {len(source)}\0".encode("ascii") + source
        ).hexdigest()
        if blob != payload["sha"]:
            raise ValueError(f"Git blob mismatch: {row['repo']}#{row['pr']}")
        slug = row["repo"].replace("/", "__") + f"__{row['pr']}"
        dest = args.out / (slug + ".source")
        dest.write_bytes(source)
        manifest.append(
            {
                **row,
                "source_file": dest.name,
                "bytes": len(source),
                "sha256": hashlib.sha256(source).hexdigest(),
                "git_blob_sha1": blob,
            }
        )
        print(f"{row['repo']}#{row['pr']}: {len(source)} bytes, {blob}")
    (args.out / "manifest.json").write_text(
        json.dumps(manifest, indent=2, ensure_ascii=False) + "\n",
        encoding="utf-8",
    )


if __name__ == "__main__":
    main()
