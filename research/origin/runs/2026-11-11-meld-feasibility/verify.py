#!/usr/bin/env python3
"""Independently recompute MELD feasibility results from saved scores."""

from __future__ import annotations

import argparse
from collections import Counter, defaultdict
import hashlib
import json
import math
from pathlib import Path
import statistics


HERE = Path(__file__).resolve().parent


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def require(ok: bool, message: str) -> None:
    if not ok:
        raise ValueError(message)


def load_rows(path: Path) -> list[dict]:
    return [json.loads(line) for line in path.read_text().splitlines()]


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--scores", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    freeze = json.loads((HERE / "freeze.json").read_text())
    manifest = json.loads((HERE / "input-manifest.json").read_text())
    require(sha(HERE / "input-manifest.json") == freeze["manifest_sha256"],
            "Manifest changed")
    complete = json.loads((args.scores / "complete.json").read_text())
    require(complete["freeze_sha256"] == sha(HERE / "freeze.json"), "Freeze changed")
    for name, expected in complete["score_sha256"].items():
        require(sha(args.scores / f"{name}.jsonl") == expected, f"Scores changed: {name}")
    require(sha(args.scores / "threshold.json") == complete["threshold_sha256"],
            "Threshold changed")
    groups = {name: load_rows(args.scores / f"{name}.jsonl")
              for name in ("calibration", "test", "matched_source", "response")}
    expected = {
        "calibration": [row for row in manifest["rows"] if row["group"] == "control"
                        and row["arm"] == "calibration" and row["eligible"]],
        "test": [row for row in manifest["rows"] if row["group"] == "control"
                 and row["arm"] == "test" and row["eligible"]],
        "matched_source": [row for row in manifest["rows"]
                           if row["group"] == "matched_source" and row["eligible"]],
        "response": [row for row in manifest["rows"]
                     if row["group"] == "response" and row["eligible"]],
    }
    for name, rows in groups.items():
        require(len(rows) == len(expected[name]), f"Missing {name} scores")
        for scored, source in zip(rows, expected[name], strict=True):
            require(all(scored.get(key) == value for key, value in source.items()),
                    f"Identity changed in {name}")
            require(math.isfinite(scored["score"]) and scored["n_tokens"] > 0,
                    f"Invalid score in {name}")
    threshold_record = json.loads((args.scores / "threshold.json").read_text())
    calibration_scores = sorted(row["score"] for row in groups["calibration"])
    allowance = math.floor(0.01 * len(calibration_scores))
    threshold = calibration_scores[-(allowance + 1)]
    require(threshold_record["threshold"] == threshold
            and threshold_record["calibration_count"] == len(calibration_scores)
            and threshold_record["allowed_calibration_flags"] == allowance
            and threshold_record["calibration_sha256"] == sha(args.scores / "calibration.jsonl")
            and threshold_record["freeze_sha256"] == sha(HERE / "freeze.json"),
            "Calibration threshold changed")
    flagged = lambda row: row["score"] > threshold
    test = groups["test"]
    response = groups["response"]
    generated = [row for row in response if row["operation"] == "generate"]
    polished = [row for row in response if row["operation"] == "polish"]
    source = {row["task_id"]: row for row in groups["matched_source"]}
    require(len(source) == 38 and len(generated) == len(polished) == 228,
            "Matched packet incomplete")
    by_repo = defaultdict(lambda: {"test": 0, "test_flags": 0,
                                   "generate": 0, "generate_flags": 0,
                                   "polish": 0, "polish_flags": 0})
    for row in test:
        item = by_repo[row["repository"]]
        item["test"] += 1
        item["test_flags"] += flagged(row)
    for row in response:
        item = by_repo[row["repository"]]
        op = row["operation"]
        item[op] += 1
        item[f"{op}_flags"] += flagged(row)
    positive_deltas = Counter()
    for row in response:
        positive_deltas[row["operation"]] += row["score"] > source[row["task_id"]]["score"]
    counts = {"test_controls": len(test),
              "test_control_flags": sum(map(flagged, test)),
              "generation": len(generated),
              "generation_flags": sum(map(flagged, generated)),
              "polish": len(polished),
              "polish_flags": sum(map(flagged, polished)),
              "generation_above_source": positive_deltas["generate"],
              "polish_above_source": positive_deltas["polish"],
              "truncated": {name: sum(row["truncated"] for row in rows)
                            for name, rows in groups.items()},
              "coverage": {"calibration": [len(groups["calibration"]), 404],
                           "test": [len(test), 357],
                           "generation": [len(generated), 228],
                           "polish": [len(polished), 228]}}
    gates = {"test_control_alerts_at_most_1pct": counts["test_control_flags"] / len(test) <= 0.01,
             "generation_detection_at_least_50pct": counts["generation_flags"] / len(generated) >= 0.50,
             "polish_detection_at_least_20pct": counts["polish_flags"] / len(polished) >= 0.20}
    result = {"format": "unswell-meld-feasibility-result-v1",
              "freeze_sha256": sha(HERE / "freeze.json"),
              "score_sha256": complete["score_sha256"],
              "threshold": threshold, "counts": counts, "gates": gates,
              "passed": all(gates.values()),
              "repository_counts": dict(sorted(by_repo.items())),
              "score_medians": {name: statistics.median(row["score"] for row in rows)
                                for name, rows in groups.items()}}
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True,
                                      allow_nan=False) + "\n")
    print(json.dumps({"counts": counts, "gates": gates,
                      "passed": result["passed"]}, sort_keys=True))


if __name__ == "__main__":
    main()
