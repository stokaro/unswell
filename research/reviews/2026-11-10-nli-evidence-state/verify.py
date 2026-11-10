#!/usr/bin/env python3
"""Independently score frozen NLI logits and the editorial admission mapping."""

import hashlib
import json
import math
from pathlib import Path


HERE = Path(__file__).resolve().parent
LABELS = ("contradiction", "support", "no_evidence")
SUPPORT_FLICKS = (
    ("table-announcement", "matching-heading", "heading-removed"),
    ("unsupported-comparison", "no-comparison", "comparison-added"),
    ("unsupported-reader-need", "no-reader-evidence", "reader-evidence-added"),
    ("repeated-release-status", "prior-fact-added", "unrelated-distractor"),
)


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def score(result: dict, freeze: dict, packet: dict, key: dict) -> dict:
    if result["format"] != "unswell-349-nli-evidence-result-v1":
        raise ValueError("result format changed")
    if result["freeze_sha256"] != digest(HERE / "freeze.json"):
        raise ValueError("freeze identity changed")
    if result["packet_sha256"] != freeze["files"]["packet.json"]:
        raise ValueError("packet identity changed")
    if result["model_sha256"] != freeze["assets"]["model"]:
        raise ValueError("model identity changed")
    cases = packet["cases"]
    labels = key["cases"]
    rows = result["rows"]
    if not (len(cases) == len(labels) == len(rows) == 16):
        raise ValueError("incomplete sixteen-case result")
    if [row["item_id"] for row in rows] != [case["item_id"] for case in cases]:
        raise ValueError("inference order or item identity changed")
    if [row["item_id"] for row in labels] != [case["item_id"] for case in cases]:
        raise ValueError("key identity changed")
    relation_correct = 0
    decision_correct = 0
    contradiction_false_admissions = 0
    confusion = {label: {prediction: 0 for prediction in LABELS} for label in LABELS}
    decisions = {}
    for row, label in zip(rows, labels):
        logits = row["logits"]
        if (not isinstance(logits, list) or len(logits) != 3 or
                not all(isinstance(value, (int, float)) and math.isfinite(value)
                        for value in logits)):
            raise ValueError("invalid or missing scores")
        if not isinstance(row["token_ids"], int) or not 1 <= row["token_ids"] <= 512:
            raise ValueError("invalid token count")
        predicted = LABELS[max(range(3), key=lambda index: logits[index])]
        if row["relation"] != predicted:
            raise ValueError("reported relation disagrees with logits")
        expected = label["expected_relation"]
        if expected not in LABELS:
            raise ValueError("unknown expected relation")
        confusion[expected][predicted] += 1
        relation_correct += predicted == expected
        if label["case_id"] in ("table-announcement", "repeated-release-status"):
            admit = predicted == "support"
        elif label["case_id"] in ("unsupported-comparison", "unsupported-reader-need"):
            admit = predicted == "no_evidence"
        else:
            raise ValueError("unknown diagnosis family")
        decision_correct += admit == label["editorial_applies"]
        if expected == "contradiction" and admit:
            contradiction_false_admissions += 1
        decisions[(label["case_id"], label["variant_id"])] = admit
    flip_correct = []
    for family, admit_variant, reject_variant in SUPPORT_FLICKS:
        flip_correct.append(decisions[(family, admit_variant)] and
                            not decisions[(family, reject_variant)])
    return {"relation_correct": relation_correct, "editorial_decision_correct": decision_correct,
            "confusion": confusion, "support_flip_pairs_correct": sum(flip_correct),
            "contradiction_false_admissions": contradiction_false_admissions,
            "directional_gate_passed": (relation_correct >= 14 and decision_correct >= 14 and
                                        all(flip_correct) and contradiction_false_admissions == 0)}


def main() -> None:
    frozen = json.loads((HERE / "freeze.json").read_text())
    for name, expected in frozen["files"].items():
        if digest(HERE / name) != expected:
            raise ValueError(f"frozen input changed: {name}")
    packet = json.loads((HERE / "packet.json").read_text())
    key = json.loads((HERE / "key.json").read_text())
    result = json.loads((HERE / "result.json").read_text())
    observed = score(result, frozen, packet, key)
    summary_path = HERE / "summary.json"
    if summary_path.exists() and json.loads(summary_path.read_text()) != observed:
        raise ValueError("published summary differs from independent score")
    print(json.dumps(observed, indent=2))


if __name__ == "__main__":
    main()
