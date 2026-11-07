#!/usr/bin/env python3
"""Check the scoped-degree CLI against the 18 frozen constructed controls."""

import argparse
import hashlib
import json
from pathlib import Path
import subprocess


ROOT = Path(__file__).resolve().parents[3]
CONTROLS = ROOT / "research/methods/audits/2026-11-07-external-controls"


def sha256(data):
    return hashlib.sha256(data).hexdigest()


def screen(binary):
    frozen = json.loads((CONTROLS / "freeze-v2.json").read_text())
    controls_bytes = (CONTROLS / "controls.json").read_bytes()
    if sha256(controls_bytes) != frozen["controls_sha256"]:
        raise ValueError("Control input drift")
    cases = json.loads(controls_bytes)["cases"]
    baseline = {entry["id"]: entry for entry in
                json.loads((CONTROLS / "result-v2.json").read_text())["cases"]}
    if len(cases) != 18 or set(baseline) != {case["id"] for case in cases}:
        raise ValueError("Control inventory drift")
    changes = []
    for case in cases:
        if sha256(case["source"].encode()) != baseline[case["id"]]["source_sha256"]:
            raise ValueError("Control source drift")
        command = [str(binary), "check", "--stdin", "--filename", case["id"] + ".md",
                   "--no-gate", "--report", "json:-"]
        run = subprocess.run(command, input=case["source"].encode(), capture_output=True,
                             timeout=120, check=True, cwd=ROOT)
        report = json.loads(run.stdout)
        if (report["status"] != "complete" or report["errors"] or
                report["manifest"]["skipped_rules"] or len(report["documents"]) != 1):
            raise ValueError("Incomplete control report")
        found = [(finding["rule_id"], finding["primary"]["span"], finding["message"])
                 for finding in report["findings"]]
        before = [(finding["rule_id"], finding["span"], finding["message"])
                  for finding in baseline[case["id"]]["findings"]]
        if found != before:
            changes.append(case["id"])
    return {"cases": len(cases), "changed_cases": changes,
            "binary_sha256": sha256(binary.read_bytes()),
            "controls_sha256": sha256(controls_bytes)}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=Path, required=True)
    args = parser.parse_args()
    result = screen(args.binary.resolve(strict=True))
    print(json.dumps(result, indent=2, sort_keys=True))
    if result["changed_cases"]:
        raise SystemExit("Control findings changed")


if __name__ == "__main__":
    main()
