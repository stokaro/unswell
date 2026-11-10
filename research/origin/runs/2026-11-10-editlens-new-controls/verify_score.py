#!/usr/bin/env python3
"""Recompute the frozen control decision from saved window-level scores."""

from __future__ import annotations

import argparse
import hashlib
import json
import math
from pathlib import Path
import statistics


HERE = Path(__file__).resolve().parent


def need(ok: bool, message: str) -> None:
    if not ok:
        raise ValueError(message)


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def wilson(successes: int, total: int) -> list[float]:
    z = 1.959963984540054
    proportion = successes / total
    scale = 1 + z**2 / total
    midpoint = (proportion + z**2 / (2 * total)) / scale
    radius = z * math.sqrt(proportion * (1 - proportion) / total +
                           z**2 / (4 * total**2)) / scale
    return [midpoint - radius, midpoint + radius]


def verify(output: Path) -> dict:
    frame = json.loads((HERE / "frame.json").read_text())
    freeze = json.loads((HERE / "freeze.json").read_text())
    calibration = json.loads((output / "calibration.json").read_text())
    docs = json.loads((output / "documents.json").read_text())
    windows = json.loads((output / "window_scores.json").read_text())
    result = json.loads((output / "result.json").read_text())
    need(calibration["freeze_sha256"] == result["freeze_sha256"] == sha(HERE / "freeze.json"),
         "Freeze identity changed")
    need(calibration["frame_sha256"] == result["frame_sha256"] == sha(HERE / "frame.json"),
         "Frame identity changed")
    need(result["calibration_sha256"] == sha(output / "calibration.json"),
         "Calibration changed")
    need(len(docs) == len(frame["selected"]) == 761, "Document count")
    grouped = {}
    for row in windows:
        need(0 < row["tokens"] <= 512 and row["arm"] in ("calibration", "test"),
             "Invalid window")
        need(isinstance(row["score"], (float, int)) and
             math.isfinite(row["score"]) and 0 <= row["score"] <= 1,
             "Invalid score index")
        grouped.setdefault(row["source_id"], []).append(row)
    for recorded, source in zip(docs, frame["selected"], strict=True):
        need(all(recorded[key] == value for key, value in source.items()),
             f"Source identity: {source['source_id']}")
        listed = sorted(grouped.pop(source["source_id"], []), key=lambda item: item["index"])
        need(len(listed) == source["windows_100"] and
             [row["index"] for row in listed] == list(range(len(listed))) and
             all(row["arm"] == source["arm"] for row in listed),
             "Window coverage")
        calculated = statistics.mean(row["score"] for row in listed)
        need(recorded["score"] == calculated, "Document mean")
    need(not grouped, "Unassigned windows")
    cal_docs = [doc for doc in docs if doc["arm"] == "calibration"]
    test_docs = [doc for doc in docs if doc["arm"] == "test"]
    need(len(cal_docs) == 404 and len(test_docs) == 357, "Arm counts")
    need(calibration["documents"] == cal_docs and
         calibration["windows"] == [row for row in windows if row["arm"] == "calibration"],
         "Calibration artifact changed")
    threshold = max(doc["score"] for doc in cal_docs)
    need(calibration["threshold_strictly_greater"] ==
         result["threshold_strictly_greater"] == threshold, "Threshold changed")
    for doc in docs:
        need(doc["flagged"] == (doc["score"] > threshold), "Flag changed")
    flags = sum(doc["flagged"] for doc in test_docs)
    coverage = len(test_docs) / frame["candidate_counts"]["test"]
    need(result["calibration_documents"] == len(cal_docs) and
         result["test_documents"] == len(test_docs) and
         result["test_candidates"] == frame["candidate_counts"]["test"] and
         result["test_coverage"] == coverage and result["test_flagged"] == flags and
         result["test_flag_rate"] == flags / len(test_docs),
         "Control summary changed")
    by_repository = {}
    for repo in sorted({doc["repository"] for doc in test_docs}):
        group = [doc for doc in test_docs if doc["repository"] == repo]
        by_repository[repo] = {"documents": len(group),
                               "flagged": sum(doc["flagged"] for doc in group)}
    flagged_repos = sum(group["flagged"] > 0 for group in by_repository.values())
    need(result["by_repository"] == by_repository and
         result["test_repositories"] == len(by_repository) and
         result["test_flagged_repositories"] == flagged_repos and
         result["test_repo_flag_wilson_95"] == wilson(flagged_repos, len(by_repository)),
         "Repository breakdown changed")
    by_language = {}
    for language in ("C#", "C++", "Java", "Python"):
        group = [doc for doc in test_docs if doc["language"] == language]
        by_language[language] = {"documents": len(group),
                                 "repositories": len({doc["repository"] for doc in group}),
                                 "flagged": sum(doc["flagged"] for doc in group)}
    need(result["by_language"] == by_language, "Language breakdown changed")
    by_length = {}
    for label, predicate in (("1-4", lambda count: count <= 4),
                             ("5-9", lambda count: 5 <= count <= 9),
                             ("10+", lambda count: count >= 10)):
        group = [doc for doc in test_docs if predicate(doc["windows_100"])]
        by_length[label] = {"documents": len(group),
                            "flagged": sum(doc["flagged"] for doc in group)}
    need(result["by_length"] == by_length, "Length breakdown changed")
    need(result["control_gate_pass"] == (coverage >= 0.70 and flags / len(test_docs) <= 0.01),
         "Control gate changed")
    return {"calibration": len(cal_docs), "test": len(test_docs),
            "test_flagged": flags, "test_coverage": coverage,
            "control_gate_pass": result["control_gate_pass"],
            "result_sha256": sha(output / "result.json")}


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", type=Path, required=True)
    args = parser.parse_args()
    print(json.dumps(verify(args.output_dir), sort_keys=True))
