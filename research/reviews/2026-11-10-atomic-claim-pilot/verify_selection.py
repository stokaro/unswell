#!/usr/bin/env python3
"""Reject changes to the source identities frozen before page review."""

import json
from pathlib import Path

from select_sources import select


def verify(packet_dir: Path) -> dict:
    record = json.loads((packet_dir / "selection.json").read_text(encoding="utf-8"))
    expected = select(packet_dir / "search",
                      packet_dir / "prior-source-repositories.json")
    if record != expected:
        raise ValueError("frozen source selection changed")
    exclusions = json.loads((packet_dir / "prior-source-repositories.json")
                            .read_text(encoding="utf-8"))
    if exclusions["base_commit"] != "9c4884ab44ef39b1f5dde62a63c5476ada02e7f1":
        raise ValueError("wrong exclusion base commit")
    return record


if __name__ == "__main__":
    packet = Path(__file__).resolve().parent
    result = verify(packet)
    print(f"verified {len(result['selected'])} frozen source identities")
