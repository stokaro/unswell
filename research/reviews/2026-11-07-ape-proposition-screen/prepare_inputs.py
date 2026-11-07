#!/usr/bin/env python3
"""Materialize the pinned #349 source pages outside the repository."""

import argparse
import hashlib
import importlib.util
import json
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
REPLAY = ROOT / "2026-09-26-outcome-announcements" / "replay.py"


def load_sources():
    spec = importlib.util.spec_from_file_location("frozen_replay", REPLAY)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module.sources()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    output = args.output.resolve()
    output.mkdir(parents=True, exist_ok=False)
    manifest = {}
    for name, data in sorted(load_sources().items()):
        target = output / name
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_bytes(data)
        manifest[name] = hashlib.sha256(data).hexdigest()
    if len(manifest) != 36:
        raise ValueError(f"Expected 36 pages, got {len(manifest)}")
    (output / "source-hashes.json").write_text(json.dumps(manifest, indent=2) + "\n")
    print(f"Prepared {len(manifest)} frozen sources at {output}")


if __name__ == "__main__":
    main()
