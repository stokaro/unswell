#!/usr/bin/env python3
"""Verify saved ARB transfer evidence without the public raw texts."""

from __future__ import annotations

import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]
FEATURE = ROOT / "research/origin/runs/2026-11-07-richness-screen/run.py"
RESULT_SHA = "7f942c09a44e93f3c83f8acdedf92cf1283ec64b8efe2e16e945fa858316156f"


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def rate(value) -> bool:
    return (value is None or
            type(value) in (int, float) and 0 <= value <= 1)


def verify(directory: Path = HERE) -> None:
    freeze = json.loads((directory / "freeze.json").read_text())
    result_path = directory / "result.json"
    require(sha(directory / "PROTOCOL.md") == freeze["protocol_sha256"],
            "Protocol hash mismatch")
    require(sha(directory / "run.py") == freeze["script_sha256"],
            "Script hash mismatch")
    require(sha(FEATURE) == freeze["feature_sha256"],
            "Feature hash mismatch")
    require(sha(result_path) == RESULT_SHA, "Result hash mismatch")
    require(freeze["dataset_commit"] ==
            "e03387f3d53efcaec9ca068c8e902e5f55611fe9",
            "Dataset commit mismatch")
    require(freeze["dataset_sha256"] ==
            "66e48240618b92ce876e2097d8df52f06661a70e31791b91a95b15688a099ab0",
            "Dataset SHA-256 mismatch")
    require(freeze["split_sha256"] ==
            "2dd13f606cc19ab4ec137d4ff836df7cac2fbbdcf4743d18c5ed28a2604c3048",
            "Split SHA-256 mismatch")

    result = json.loads(result_path.read_text())
    for key in ("dataset_commit", "dataset_sha256", "feature_sha256",
                "split_sha256"):
        require(result["input"][key] == freeze[key],
                f"Result {key} mismatch")
    require(result["input"]["rows"] == 23400 and
            result["input"]["source_groups"] == 1800 and
            result["input"]["calibration_sources"] == 900 and
            result["input"]["test_sources"] == 900,
            "Result denominator mismatch")
    for mode in ("full", "prefix100"):
        for feature in ("richness", "entropy"):
            block = result["measures"][mode][feature]
            require(block["calibration"]["total"] == 900 and
                    block["test_human"]["total"] == 900,
                    "Human denominator mismatch")
            require(rate(block["calibration"]["fpr"]) and
                    rate(block["test_human"]["fpr"]),
                    "Invalid FPR")
            for regime in ("free_llm", "h2l", "llm2l"):
                entry = block["regimes"][regime]
                require(entry["total_machine"] == 3600,
                        "Machine denominator mismatch")
                require(0 <= entry["overall"]["eligible_pairs"] <= 3600,
                        "Invalid pair coverage")
                for key in ("auroc", "paired_win_fraction", "tpr"):
                    require(rate(entry["overall"][key]),
                            f"Invalid {mode} {feature} {regime} {key}")
    primary = result["measures"]["full"]["richness"]
    h2l = primary["regimes"]["h2l"]
    recomputed = all((
        h2l["overall"]["eligible_pairs"] / h2l["total_machine"] >= 0.95,
        h2l["overall"]["auroc"] >= 0.65,
        h2l["overall"]["tpr"] >= 0.10,
        primary["test_human"]["fpr"] <= 0.02,
        sum(row["auroc"] is not None and row["auroc"] > 0.60
            for row in h2l["by_domain"].values()) >= 2,
        sum(row["auroc"] is not None and row["auroc"] > 0.60
            for row in h2l["by_generator"].values()) >= 3,
    ))
    require(not recomputed and not result["richness_h2l_feasibility_pass"],
            "Registered feasibility outcome changed")
    print("ARB frozen source, result, and negative feasibility verified")


if __name__ == "__main__":
    verify()
