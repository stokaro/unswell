#!/usr/bin/env python3
"""Verify and score the frozen explicit-proposition NLI output."""

from __future__ import annotations

from collections import Counter
import hashlib
import json
import math
from pathlib import Path
import sys

HERE = Path(__file__).resolve().parent
LABELS = ("entailment", "neutral", "contradiction")


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def score(packet: dict, result: dict, freeze: dict, freeze_sha256: str) -> dict:
    if packet["format"] != "unswell-349-explicit-proposition-nli-v1" or len(packet["cases"]) != 16:
        raise ValueError("packet shape changed")
    if result["format"] != "unswell-349-explicit-proposition-nli-result-v1":
        raise ValueError("result format changed")
    if (result["freeze_sha256"] != freeze_sha256 or
            result["packet_sha256"] != freeze["packet_sha256"] or
            result["model_sha256"] != freeze["assets"]["onnx/model_quantized.onnx"]["sha256"]):
        raise ValueError("result identity changed")
    rows = result["rows"]
    if len(rows) != len(packet["cases"]):
        raise ValueError("missing or extra cases")
    counts = Counter()
    mistakes = []
    for case, row in zip(packet["cases"], rows):
        if row["id"] != case["id"] or not 1 <= row["token_ids"] <= 512:
            raise ValueError("case identity or token count changed")
        logits = row["logits"]
        if len(logits) != 3 or not all(isinstance(value, (int, float)) and
                                       math.isfinite(value) for value in logits):
            raise ValueError("invalid logits")
        predicted = LABELS[max(range(3), key=lambda index: logits[index])]
        if row["relation"] != predicted:
            raise ValueError("relation disagrees with logits")
        expected = case["expected_relation"]
        if expected not in LABELS:
            raise ValueError("unknown expected relation")
        counts[expected] += 1
        counts["correct"] += predicted == expected
        if predicted != expected:
            mistakes.append({"id": case["id"], "expected": expected, "predicted": predicted})
        if expected == "neutral" and predicted == "entailment":
            counts["neutral_false_entailment"] += 1
        if expected == "contradiction" and predicted == "contradiction":
            counts["correct_contradiction"] += 1
    if [counts[key] for key in LABELS] != [6, 6, 4]:
        raise ValueError("frozen class balance changed")
    passed = (counts["correct"] >= 14 and counts["correct_contradiction"] == 4
              and counts["neutral_false_entailment"] <= 1)
    return {"format": "unswell-349-explicit-proposition-nli-score-v1",
            "cases": 16, "correct": counts["correct"],
            "correct_contradiction": counts["correct_contradiction"],
            "neutral_false_entailment": counts["neutral_false_entailment"],
            "mistakes": mistakes, "directional_pass": passed}


def main() -> None:
    packet = json.loads((HERE / "packet.json").read_text())
    freeze = json.loads((HERE / "freeze.json").read_text())
    if (digest(HERE / "packet.json") != freeze["packet_sha256"] or
            digest(HERE / "run.py") != freeze["runner_sha256"] or
            digest(HERE / "score.py") != freeze["scorer_sha256"]):
        raise ValueError("frozen code or packet changed")
    path = Path(sys.argv[1]) if len(sys.argv) > 1 else HERE / "result.json"
    result = json.loads(path.read_text())
    summary = score(packet, result, freeze, digest(HERE / "freeze.json"))
    if len(sys.argv) > 2:
        Path(sys.argv[2]).write_text(json.dumps(summary, indent=2, sort_keys=True) + "\n")
    print(json.dumps(summary, sort_keys=True))


if __name__ == "__main__":
    main()
