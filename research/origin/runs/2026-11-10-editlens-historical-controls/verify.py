#!/usr/bin/env python3
"""Independently verify historical control scores and fixed-threshold flags."""

from __future__ import annotations

import argparse
import collections
import hashlib
import json
import math
from pathlib import Path
import statistics


HERE = Path(__file__).resolve().parent
RUNS = HERE.parent
FRAME = HERE / "frame.json"
THRESHOLD = 0.27447814408224075
LANGUAGES = ("C#", "C++", "Java", "Python")


def need(ok: bool, message: str) -> None:
    if not ok:
        raise ValueError(message)


def sha(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def read(path: Path):
    return json.loads(path.read_text())


def equal(left, right, label: str = "result") -> None:
    if isinstance(left, bool) or isinstance(right, bool):
        need(left is right, f"{label} changed")
    elif isinstance(left, (float, int)) and isinstance(right, (float, int)):
        need(math.isfinite(left) and math.isfinite(right) and
             math.isclose(left, right, abs_tol=1e-12, rel_tol=1e-12),
             f"{label} changed")
    elif isinstance(left, dict) and isinstance(right, dict):
        need(left.keys() == right.keys(), f"{label} keys changed")
        for key in left:
            equal(left[key], right[key], f"{label}.{key}")
    elif isinstance(left, list) and isinstance(right, list):
        need(len(left) == len(right), f"{label} length changed")
        for index, (a, b) in enumerate(zip(left, right, strict=True)):
            equal(a, b, f"{label}[{index}]")
    else:
        need(left == right, f"{label} changed")


def verify_payload(frame: dict, freeze: dict, docs: list[dict],
                   windows: list[dict], result: dict) -> dict:
    expected = {row["source_id"]: row for row in frame["selected"]}
    need(len(expected) == 462 and len(docs) == 462 and len(windows) == 3735,
         "Control coverage changed")
    actual = {}
    for doc in docs:
        source_id = doc["source_id"]
        prior = expected.get(source_id)
        need(prior is not None and source_id not in actual,
             "Unknown or duplicate control")
        need(set(doc) == set(prior) | {"removed_header", "score", "flagged"},
             "Control record shape changed")
        equal({key: doc[key] for key in prior}, prior, "source metadata")
        need(isinstance(doc["removed_header"], bool) and
             isinstance(doc["score"], (float, int)) and
             math.isfinite(doc["score"]) and 0 <= doc["score"] <= 1 and
             isinstance(doc["flagged"], bool), "Invalid document score")
        actual[source_id] = doc
    need(set(actual) == set(expected), "Missing source control")
    grouped = collections.defaultdict(list)
    seen = set()
    for row in windows:
        need(set(row) == {"source_id", "index", "tokens", "score"},
             "Window record shape changed")
        need(row["source_id"] in actual and isinstance(row["index"], int) and
             isinstance(row["tokens"], int) and 0 < row["tokens"] <= 512 and
             isinstance(row["score"], (float, int)) and
             math.isfinite(row["score"]) and 0 <= row["score"] <= 1,
             "Invalid window")
        identity = (row["source_id"], row["index"])
        need(identity not in seen, "Duplicate window")
        seen.add(identity)
        grouped[row["source_id"]].append(row)
    for source_id, doc in actual.items():
        ordered = sorted(grouped[source_id], key=lambda row: row["index"])
        need([row["index"] for row in ordered] == list(range(doc["windows_100"])),
             "Missing source window")
        mean = sum(row["score"] for row in ordered) / len(ordered)
        equal(doc["score"], mean, f"document mean {source_id}")
        need(doc["flagged"] is (mean > THRESHOLD),
             f"Flag decision changed: {source_id}")
    flagged = [doc for doc in docs if doc["flagged"]]
    scores = sorted(doc["score"] for doc in docs)

    def quantile(fraction: float) -> float:
        rank = fraction * (len(scores) - 1)
        before = int(rank)
        return scores[before] + (scores[min(before + 1, len(scores) - 1)] -
                                 scores[before]) * (rank - before)

    by_language = {}
    for language in LANGUAGES:
        subset = [doc for doc in docs if doc["language"] == language]
        by_language[language] = {"documents": len(subset),
                                 "repositories": len({doc["repository"] for doc in subset}),
                                 "flagged": sum(doc["flagged"] for doc in subset)}
    by_repository = {}
    for repository in sorted({doc["repository"] for doc in docs}):
        subset = [doc for doc in docs if doc["repository"] == repository]
        by_repository[repository] = {"documents": len(subset),
                                     "flagged": sum(doc["flagged"] for doc in subset)}
    by_length = {}
    for label, predicate in (("1-4", lambda n: n <= 4),
                             ("5-9", lambda n: 5 <= n <= 9),
                             ("10+", lambda n: n >= 10)):
        subset = [doc for doc in docs if predicate(doc["windows_100"])]
        by_length[label] = {"documents": len(subset),
                            "flagged": sum(doc["flagged"] for doc in subset)}
    expected_result = {
        "format": "unswell-editlens-historical-controls-v1",
        "scope": "pre-2021 technical control flag rate; authorship not verified",
        "model_sha256": freeze["model_sha256"],
        "frame_sha256": freeze["frame_sha256"],
        "selection_sha256": freeze["selection_sha256"],
        "threshold_strictly_greater": THRESHOLD,
        "documents": len(docs),
        "repositories": len({doc["repository"] for doc in docs}),
        "windows": len(windows), "flagged": len(flagged),
        "flag_rate": len(flagged) / len(docs),
        "median_score": statistics.median(scores),
        "score_quantiles": {f"p{percent}": quantile(percent / 100)
                            for percent in (10, 25, 75, 90)},
        "by_language": by_language, "by_repository": by_repository,
        "by_length": by_length,
        "flagged_source_ids": [doc["source_id"] for doc in flagged],
        "control_transfer_pass": len(flagged) == 0,
    }
    equal(result, expected_result)
    return {"documents": len(docs), "repositories": expected_result["repositories"],
            "windows": len(windows), "flagged": len(flagged),
            "control_transfer_pass": expected_result["control_transfer_pass"]}


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", type=Path, default=HERE)
    args = parser.parse_args()
    freeze = read(HERE / "freeze.json")
    for name, path in (("protocol_sha256", HERE / "PROTOCOL.md"),
                       ("runner_sha256", HERE / "run.py"),
                       ("verifier_sha256", HERE / "verify.py"),
                       ("fetch_sha256", HERE / "fetch.py"),
                       ("selection_verifier_sha256", HERE / "verify_selection.py"),
                       ("runner_tests_sha256", HERE / "test_run.py"),
                       ("verifier_tests_sha256", HERE / "test_verify.py"),
                       ("selection_tests_sha256", HERE / "test_verify_selection.py"),
                       ("frame_sha256", FRAME),
                       ("inventory_sha256", HERE / "inventory.json"),
                       ("blob_audit_sha256", HERE / "blob_audit.json"),
                       ("eligibility_sha256", HERE / "eligibility.json"),
                       ("richness_runner_sha256", RUNS / "2026-11-07-richness-screen/run.py"),
                       ("origin_runner_sha256", RUNS / "2026-11-09-editlens-transfer/run.py"),
                       ("window_audit_sha256", RUNS / "2026-11-10-editlens-technical-windows/audit.py"),
                       ("fullwindow_runner_sha256", RUNS / "2026-11-10-editlens-technical-fullwindows/run.py"),
                       ("academic_result_sha256", RUNS / "2026-11-10-editlens-five-window/result.json")):
        need(sha(path) == freeze[name], f"Frozen input changed: {name}")
    decision = read(args.output_dir / "decision.json")
    for name in ("result.json", "documents.json", "window_scores.json"):
        need(sha(args.output_dir / name) == decision["files_sha256"][name],
             f"Result file changed: {name}")
    summary = verify_payload(read(FRAME), freeze,
                             read(args.output_dir / "documents.json"),
                             read(args.output_dir / "window_scores.json"),
                             read(args.output_dir / "result.json"))
    print(json.dumps(summary, sort_keys=True))


if __name__ == "__main__":
    main()
