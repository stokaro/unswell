#!/usr/bin/env python3
"""Verify replay artifacts and record an independently repeated measurement."""

import argparse
from pathlib import Path
import sys

sys.dont_write_bytecode = True
from compare import PAIRS, digest, read, require, write


def verify(initial, repeated, record):
    files = [Path("bin") / name for name in ("new", "old-a1c6d7a", "old-d3c9f23")]
    files += [Path(f"{subject}-sources.json") for subject in ("ptah", "responses", "unswell")]
    for subject, profile, old in PAIRS:
        root = Path(subject) if subject == "responses" else Path(subject) / profile
        files += [root / engine / "result.json" for engine in (old, "new")]
    checks = []
    for name in files:
        first, second = digest(initial / name), digest(repeated / name)
        require(first == second, f"Reproduction differs: {name}")
        checks.append({"path": name.as_posix(), "initial_sha256": first, "reproduced_sha256": second})
    for run in read(record / "comparison.json")["runs"]:
        archive = record / run["alignment"]["path"]
        require(digest(archive) == run["alignment"]["sha256"], "Alignment archive changed")
    write(record / "reproduction.json", {"version": "unswell-review-reproduction-v1",
                                         "environment": read(repeated / "environment.json"),
                                         "checks": checks, "all_identical": True})
    print(f"Verified {len(files)} identical binaries, source manifests, and reports.")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("initial", type=Path)
    parser.add_argument("repeated", type=Path)
    parser.add_argument("record", type=Path)
    options = parser.parse_args()
    verify(options.initial, options.repeated, options.record)
