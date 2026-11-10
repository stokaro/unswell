#!/usr/bin/env python3
"""Diagnose a frozen control threshold on previously exposed AI responses."""

from __future__ import annotations

import argparse
import hashlib
import json
import math
from pathlib import Path


OLD_DOC_SHA256 = "a466fb3ccc4bbf5a427bd850063bfb1c34a89af0ce1313dbc0b3713b85ddc657"
MODEL_SHA256 = "ddd1173f2ef517ad499965e5029fae8099a8054a2bc76d8134e5889cc4ed1b3e"


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--old-documents", type=Path, required=True)
    parser.add_argument("--old-freeze", type=Path, required=True)
    parser.add_argument("--old-threshold-freeze", type=Path, required=True)
    parser.add_argument("--new-documents", type=Path, required=True)
    parser.add_argument("--new-result", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if sha(args.old_documents) != OLD_DOC_SHA256:
        raise ValueError("Previously scored AI responses changed")
    old_freeze = json.loads(args.old_freeze.read_text())
    current_freeze = json.loads((Path(__file__).parent / "freeze.json").read_text())
    if old_freeze["model_sha256"] != MODEL_SHA256 or current_freeze["model_sha256"] != MODEL_SHA256:
        raise ValueError("Model identity mismatch")
    old_threshold = json.loads(args.old_threshold_freeze.read_text())["threshold_strictly_greater"]
    previous = json.loads(args.old_documents.read_text())
    current = json.loads(args.new_documents.read_text())
    result = json.loads(args.new_result.read_text())
    controls = sorted((row["score"] for row in current if row["arm"] == "test"), reverse=True)
    ai = {name: sorted((row["score"] for row in previous
                        if row["kind"] == "response" and row["operation"] == name), reverse=True)
          for name in ("generate", "polish")}
    if len(controls) != 357 or any(len(rows) != 228 for rows in ai.values()):
        raise ValueError("Exposed cohort changed")
    threshold = result["threshold_strictly_greater"]
    if threshold != max(row["score"] for row in current if row["arm"] == "calibration"):
        raise ValueError("Calibration threshold changed")
    allowed_flags = math.floor(0.01 * len(controls))
    best_control_threshold = controls[allowed_flags]
    needed_polish = math.ceil(0.20 * len(ai["polish"]))
    polish_threshold = math.nextafter(ai["polish"][needed_polish - 1], -math.inf)

    def counts(value: float) -> dict:
        return {"test_controls": sum(item > value for item in controls),
                "generate": sum(item > value for item in ai["generate"]),
                "polish": sum(item > value for item in ai["polish"])}

    output = {"format": "unswell-new-historical-exposed-replay-v1",
              "status": "exploratory after the frozen control decision; not a preregistered gate",
              "old_documents_sha256": OLD_DOC_SHA256,
              "new_documents_sha256": sha(args.new_documents),
              "new_result_sha256": sha(args.new_result),
              "old_threshold": old_threshold,
              "old_threshold_counts": counts(old_threshold),
              "frozen_new_threshold": threshold,
              "frozen_new_threshold_counts": counts(threshold),
              "best_threshold_at_observed_one_percent_control_budget": best_control_threshold,
              "best_budget_counts": counts(best_control_threshold),
              "minimum_polish_for_twenty_percent": needed_polish,
              "threshold_to_reach_polish_target": polish_threshold,
              "polish_target_counts": counts(polish_threshold),
              "denominators": {"test_controls": len(controls),
                               "generate": len(ai["generate"]),
                               "polish": len(ai["polish"])}}
    args.output.write_text(json.dumps(output, sort_keys=True, indent=2) + "\n")
    print(json.dumps({key: output[key] for key in
                      ("old_threshold_counts", "frozen_new_threshold_counts",
                       "best_budget_counts", "polish_target_counts")}, sort_keys=True))


if __name__ == "__main__":
    main()
