#!/usr/bin/env python3
"""Build complete-page input for the current source-inventory exporter."""

import json
from pathlib import Path

from verify_sources import verify


def main():
    packet = Path(__file__).resolve().parent
    output = packet / "input.json"
    if output.exists():
        raise SystemExit("source inventory input already exists")
    inventory = verify(packet)
    pages = []
    for record in inventory["sources"]:
        pages.append({"id": f"{record['repository']}/{record['path']}",
                      "format": "markdown", "sha256": record["source_sha256"],
                      "text": (packet / "sources" / record["source_file"])
                      .read_text(encoding="utf-8")})
    output.write_text(json.dumps({"version": 1, "basis": "source-function-inventory",
                                  "pages": pages}, ensure_ascii=False,
                                 separators=(",", ":")) + "\n", encoding="utf-8")
    print(f"built source-inventory input for {len(pages)} complete pages")


if __name__ == "__main__":
    main()
