#!/usr/bin/env python3
"""Verify the frozen GEN screen and saved result without raw corpus text."""

from __future__ import annotations

import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def verify(directory: Path = HERE) -> None:
    freeze = json.loads((directory / "freeze.json").read_text())
    decision = json.loads((directory / "decision.json").read_text())
    result = json.loads((directory / "result.json").read_text())
    require(freeze["format"] == "unswell-gen-edit-freeze-v1" and
            freeze["dataset_revision"] == "a0f143c1ae9c35b684898a5f9c0ab6efb49be486",
            "Dataset freeze identity changed")
    require(decision["format"] == "unswell-gen-edit-decision-v1" and
            decision["preregistered_commit"] ==
            "ad5a06e662c493c65c205415f6c8347f9d1c2aef",
            "Preregistration identity changed")
    paths = {
        "protocol": directory / "PROTOCOL.md",
        "runner": directory / "run.py",
        "acquire_source": directory / "acquire.py",
        "lexicon": directory / "closed_class.txt",
        "feature_source": ROOT / "research/origin/runs/2026-11-07-richness-screen/run.py",
    }
    for name, path in paths.items():
        require(sha(path) == freeze["sha256"][name], f"{name} source changed")
    require(sha(directory / "result.json") == decision["result_sha256"],
            "Saved result changed")
    require(result["format"] == "unswell-gen-edit-transfer-v1" and
            result["source_split_sha256"] == freeze["split_sha256"],
            "Result input identity changed")
    acquisition = result["acquisition"]
    require(acquisition["input_rows"] == freeze["edited_input_rows"] == 273418 and
            acquisition["selected_rows"] == freeze["selected_rows"] == 90768 and
            acquisition["source_mismatches_by_domain"] ==
            {"story_generation": 1116},
            "Edited acquisition count changed")
    require(result["generation_identity"]["rows"] == 39391 and
            result["generation_identity"]["duplicate_source_model"] == 38 and
            result["generation_identity"]["missing_source_model"] == 647,
            "Generation identity changed")
    edited = result["tasks"]["edited"]
    for name, decision_key in (("closed_fraction", "edited_closed_fraction_pass"),
                               ("reverse_entropy", "edited_reverse_entropy_pass")):
        metrics = edited[f"prefix100/{name}"]
        require(metrics["feasibility_pass"] is False and
                decision[decision_key] is False and
                metrics["rows"] == 46407 and metrics["eligible"] == 45639 and
                metrics["groups"] == 498,
                f"{name} decision or denominator changed")
    generated = result["tasks"]["generated"]["prefix100/richness"]
    require(generated["rows"] == 19684 and generated["eligible"] == 17785 and
            generated["groups"] == 2497 and
            decision["generation_status"] ==
            "research_signal_requires_external_transfer" and
            decision["editorial_issue_349_status"] == "unchanged",
            "Generation interpretation changed")
    print("GEN edit-transfer freeze, saved result, and decisions verified")


if __name__ == "__main__":
    verify()
