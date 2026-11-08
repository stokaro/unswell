#!/usr/bin/env python3
"""Verify the frozen mixed-authorship screen's sources and aggregate result."""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]
FEATURE = ROOT / "research/origin/runs/2026-11-07-richness-screen/run.py"
RESULT_SHA256 = "819e6295f291586cd1befd088b7755c7a70cdf95212c467907e650c1a50b5d1f"
SPLITS = ("train", "validation", "test")
DATASETS = ("beigebook", "gutenberg_selected", "jmlr_pre2015",
            "standardebooks")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def sha(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def verify(result_path: Path, input_directory: Path | None = None) -> None:
    freeze = json.loads((HERE / "freeze.json").read_text())
    require(freeze["dataset_revision"] ==
            "c8a59b129c34cfd9c67f739ae049abfff26fc2d3" and
            freeze["configuration"] == "all_layout_neutral",
            "Dataset identity changed")
    for key, path in (("protocol_sha256", HERE / "PROTOCOL.md"),
                      ("runner_sha256", HERE / "run.py"),
                      ("feature_sha256", FEATURE)):
        require(sha(path) == freeze[key], f"Frozen source changed: {path}")
    if input_directory is not None:
        for split in SPLITS:
            path = input_directory / f"{split}.parquet"
            require(path.stat().st_size == freeze["files"][split]["bytes"] and
                    sha(path) == freeze["files"][split]["sha256"],
                    f"Pinned input changed: {split}")
    require(sha(result_path) == RESULT_SHA256, "Result hash changed")
    result = json.loads(result_path.read_text())
    require(result["dataset_revision"] == freeze["dataset_revision"] and
            result["input_sha256"] == {
                split: freeze["files"][split]["sha256"] for split in SPLITS
            } and result["split_pairs"] == {
                "train": 3509, "validation": 297, "test": 594
            }, "Result source identity changed")
    span = result["span"]["richness"]
    document = result["document"]["richness"]
    require(span["total_pairs"] == 713 and
            span["overall"]["eligible_pairs"] == 566 and
            document["total_pairs"] == 594 and
            document["overall"]["eligible_pairs"] == 594,
            "Result denominator changed")
    require(set(span["by_source_dataset"]) == set(DATASETS) and
            sum(row["eligible_pairs"] for row in
                span["by_source_dataset"].values()) == 566,
            "Span domain denominator changed")
    span_gate = all((
        span["overall"]["eligible_pairs"] / span["total_pairs"] >= 0.75,
        span["overall"]["auroc"] >= 0.65,
        span["overall"]["tpr"] >= 0.10,
        span["overall"]["human_fpr"] <= 0.02,
        sum(row["eligible_pairs"] >= 40 and row["auroc"] is not None and
            row["auroc"] > 0.60
            for row in span["by_source_dataset"].values()) >= 3,
    ))
    document_gate = all((
        document["overall"]["eligible_pairs"] / document["total_pairs"]
        >= 0.95,
        document["overall"]["auroc"] >= 0.65,
        document["overall"]["tpr"] >= 0.10,
        document["overall"]["human_fpr"] <= 0.02,
    ))
    require(span_gate is False and document_gate is False and
            result["richness_span_feasibility_pass"] == span_gate and
            result["richness_document_feasibility_pass"] == document_gate,
            "Negative feasibility decision changed")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--result", type=Path, default=HERE / "result.json")
    parser.add_argument("--input-directory", type=Path)
    args = parser.parse_args()
    verify(args.result, args.input_directory)
    print("Heterogeneous span sources, result, and negative gates verified")


if __name__ == "__main__":
    main()
