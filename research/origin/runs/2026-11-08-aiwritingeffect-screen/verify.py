#!/usr/bin/env python3
"""Verify saved AIWritingEffect research artifacts without raw dataset text."""

from __future__ import annotations

import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]
FEATURE = ROOT / "research/origin/runs/2026-11-07-richness-screen/run.py"


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path: Path) -> dict:
    return json.loads(path.read_text())


def rate(value) -> bool:
    return value is None or isinstance(value, (int, float)) and 0 <= value <= 1


def verify(directory: Path = HERE) -> None:
    frozen = read(directory / "freeze.json")
    v2 = read(directory / "freeze-v2.json")
    for manifest, names in (
        (frozen, (("protocol_sha256", "PROTOCOL.md"),
                  ("script_sha256", "run.py"))),
        (v2, (("protocol_sha256", "PROTOCOL-v2.md"),
              ("script_sha256", "run_docs.py"),
              ("primary_result_sha256", "result.json"))),
    ):
        for key, name in names:
            require(sha(directory / name) == manifest[key], f"{name} hash mismatch")
        require(sha(FEATURE) == manifest["feature_sha256"],
                "Feature source hash mismatch")
        require(manifest["dataset_commit"] ==
                "6a45eddc021257b7c49de398c2fffa370c1d1bfa",
                "Dataset commit mismatch")
    require(frozen["dataset_sha256"] ==
            "6e0eb8058578afce4aaf81c822b70762afd9f18534eb5f1e89aa976cd051b793",
            "Paragraph input identity mismatch")
    require(v2["dataset_sha256"] ==
            "428de968a71d64ddaf9f0d5afe7e4e295b2adeace77f57fd29c6252ae082d29c",
            "Document input identity mismatch")

    primary = read(directory / "result.json")
    require(primary["input"]["dataset_sha256"] == frozen["dataset_sha256"],
            "Paragraph result input mismatch")
    require(primary["input"]["feature_sha256"] == frozen["feature_sha256"],
            "Paragraph feature mismatch")
    require(primary["input"]["rows"] == 130293, "Paragraph count mismatch")
    coverage = primary["test_coverage"]
    require(coverage["ai_total"] == 7015, "Test AI denominator mismatch")
    require(coverage["eligible_pairs"] <=
            coverage["ai_total"] -
            max(coverage["ai_abstained"],
                coverage["source_abstained_for_ai_rows"]),
            "Paragraph coverage inconsistent")
    require(abs(coverage["eligible_fraction"] -
                coverage["eligible_pairs"] / coverage["ai_total"]) < 1e-12,
            "Paragraph eligible fraction mismatch")
    richness = primary["features"]["richness"]
    require(richness["overall"]["eligible_pairs"] == coverage["eligible_pairs"],
            "Richness pair count mismatch")
    require(not primary["richness_feasibility_pass"],
            "Frozen paragraph feasibility outcome changed")
    for feature in ("richness", "entropy"):
        block = primary["features"][feature]
        for key in ("pooled_auroc", "paired_win_fraction",
                    "tpr_at_calibrated_threshold"):
            require(rate(block["overall"][key]), f"{feature} {key} invalid")
        for group in ("test_paired_human", "human_only"):
            require(rate(block[group]["fpr"]), f"{feature} {group} FPR invalid")

    secondary = read(directory / "result-v2.json")
    require(secondary["input"]["dataset_sha256"] == v2["dataset_sha256"],
            "Document result input mismatch")
    require(secondary["input"]["primary_result_sha256"] ==
            v2["primary_result_sha256"], "Document predecessor mismatch")
    require(secondary["input"]["rows"] == 2598, "Document count mismatch")
    for measure in ("full", "prefix100"):
        block = secondary["measures"][measure]
        coverage = block["coverage"]
        require(coverage["ai_total"] == 143, "Document AI denominator mismatch")
        require(coverage["eligible_pairs"] <= coverage["ai_total"],
                "Document coverage exceeds denominator")
        require(block["overall"]["eligible_pairs"] == coverage["eligible_pairs"],
                "Document pair count mismatch")
        for feature in ("richness", "entropy"):
            for key in ("pooled_auroc", "paired_win_fraction"):
                require(rate(block["overall"][feature][key]),
                        f"{measure} {feature} {key} invalid")
    print("AIWritingEffect saved artifacts and frozen source hashes verified")


if __name__ == "__main__":
    verify()
