#!/usr/bin/env python3
"""Compute the label-only upper bound for the archived candidate bank."""

from __future__ import annotations

import argparse
import hashlib
import importlib.util
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
CONTROL = HERE.parent / "2026-11-09-structured-admission" / "scan-v2.py"
CONTROL_SHA256 = "20a770ce8a8e88dcf2b5ac17408bf4f3006c5179a61bbb7e510eae61c994c063"


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--development-root", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if args.output.exists() or digest(CONTROL) != CONTROL_SHA256:
        raise ValueError("output exists or control changed")
    spec = importlib.util.spec_from_file_location("frozen_structured_admission", CONTROL)
    if spec is None or spec.loader is None:
        raise ValueError("control import failed")
    control = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(control)
    data = control.read_inputs(args.development_root)
    rows, witnesses, _ = control.inventory(data)
    baseline = control.measures(set(rows), rows, witnesses)
    accepted = {alias for alias, row in rows.items() if row["accepted"]}
    oracle = control.measures(accepted, rows, witnesses)
    result = {
        "format": "unswell-349-accepted-only-oracle-v1",
        "scope": "label_only_upper_bound_not_a_selector",
        "input_sha256": {key: expected for key, (_, expected) in control.INPUTS.items()},
        "control_script_sha256": CONTROL_SHA256,
        "script_sha256": digest(Path(__file__)),
        "baseline": baseline,
        "accepted_only": oracle,
        "qualified": False,
    }
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    print(json.dumps(oracle, sort_keys=True))


if __name__ == "__main__":
    main()
