#!/usr/bin/env python3
"""Reject changed source-only labels, spans, inventories, or source bytes."""

import json
from pathlib import Path

from finalize_reference import build


def verify(packet: Path) -> dict:
    frozen = json.loads((packet / "reference.json").read_text(encoding="utf-8"))
    if frozen != build(packet):
        raise ValueError("frozen source-only reference changed")
    return frozen


if __name__ == "__main__":
    result = verify(Path(__file__).resolve().parent)
    print(f"verified {len(result['events'])} events and {len(result['controls'])} controls")
