#!/usr/bin/env python3
"""Verify the frozen technical transfer and saved decision without source text."""

from __future__ import annotations

import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]


def require(ok: bool, message: str) -> None:
    if not ok:
        raise ValueError(message)


def sha(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def verify(directory: Path = HERE) -> None:
    freeze = json.loads((directory / "freeze.json").read_text())
    decision = json.loads((directory / "decision.json").read_text())
    result = json.loads((directory / "result.json").read_text())
    require(freeze["format"] == "unswell-gen-technical-transfer-freeze-v1" and
            freeze["dataset_commit"] == "6a45eddc021257b7c49de398c2fffa370c1d1bfa",
            "Freeze identity changed")
    require(decision["format"] == "unswell-gen-technical-transfer-decision-v1" and
            decision["preregistered_commit"] ==
            "5ccf936223afd62ad714b801489719fd1cf953e7",
            "Preregistration identity changed")
    paths = {
        "protocol": directory / "PROTOCOL.md",
        "runner": directory / "run.py",
        "feature": ROOT / "research/origin/runs/2026-11-07-richness-screen/run.py",
        "gen_result": ROOT / "research/origin/runs/2026-11-08-gen-edit-transfer/result.json",
        "inputs_wrapper": ROOT / "research/generation/studies/long-prose-v3/inputs.json",
        "inputs_archive": ROOT / "research/generation/studies/long-prose-v3/inputs.tar.gz",
        "outputs_wrapper": ROOT / "research/generation/studies/long-prose-v3/outputs.json",
        "outputs_archive": ROOT / "research/generation/studies/long-prose-v3/outputs.tar.gz",
    }
    for name, path in paths.items():
        require(sha(path) == freeze["sha256"][name], f"{name} changed")
    require(sha(directory / "result.json") == decision["result_sha256"],
            "Saved result changed")
    require(result["format"] == "unswell-gen-technical-transfer-v1" and
            result["freeze_sha256"] == sha(directory / "freeze.json") and
            result["thresholds"] == freeze["thresholds"],
            "Result identity or thresholds changed")
    technical = result["technical"]
    academic = result["academic"]
    require(technical["sources"]["total"] == 38 and
            technical["operations"]["generate"]["overall"]["total"] == 228 and
            technical["operations"]["polish"]["overall"]["total"] == 228 and
            academic["rows"] == 2598 and
            academic["test_rewrites"]["overall"]["total"] == 143,
            "Transfer denominators changed")
    require(result["decision"] == decision["decision"] and
            not decision["decision"]["advance_screen"] and
            decision["editorial_issue_349_status"] == "unchanged",
            "Decision changed")
    print("GEN technical transfer freeze, result, and decision verified")


if __name__ == "__main__":
    verify()
