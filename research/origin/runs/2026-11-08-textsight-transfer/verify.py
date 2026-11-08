#!/usr/bin/env python3
"""Recompute the frozen TextSight transfer result from pinned public data."""

from __future__ import annotations

import argparse
import json
from pathlib import Path

import score


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo", type=Path, required=True)
    parser.add_argument("--data", type=Path, required=True)
    args = parser.parse_args()
    expected = json.loads((Path(__file__).parent / "result.json").read_text())
    actual = score.score(args.repo, args.data)
    if actual != expected:
        raise SystemExit("frozen result differs from recomputation")
    print("verified", actual["freeze_sha256"])


if __name__ == "__main__":
    main()
