#!/usr/bin/env python3
"""Verify the sealed source review and apply the preregistered pair screen."""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent


def require(ok: bool, message: str) -> None:
    if not ok:
        raise ValueError(message)


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def load(path: Path) -> dict:
    return json.loads(path.read_text())


def evaluate(packet_dir: Path, study_dir: Path = HERE) -> dict:
    packet_path = packet_dir / "packet.json"
    packet = load(packet_path)
    key = load(study_dir / "key.json")
    review = load(study_dir / "review.json")
    seal = load(study_dir / "review-seal.json")
    manifest = load(packet_dir / "manifest.json")

    require(packet["format"] == "unswell-long-sentence-dependency-packet-v1",
            "Unexpected packet format")
    require(key["format"] == "unswell-long-sentence-dependency-key-v1",
            "Unexpected key format")
    require(review["format"] == "unswell-long-sentence-dependency-source-review-v1",
            "Unexpected review format")
    require(seal["format"] == "unswell-long-sentence-dependency-review-seal-v1",
            "Unexpected seal format")
    require(manifest["format"] == "unswell-long-sentence-dependency-manifest-v1",
            "Unexpected manifest format")
    require(packet["freeze_sha256"] == sha(study_dir / "freeze.json"),
            "Packet does not match freeze")
    packet_hash = sha(packet_path)
    require(packet_hash == manifest["packet_sha256"] == key["packet_sha256"]
            == review["packet_sha256"] == seal["packet_sha256"],
            "Packet hash mismatch")
    require(sha(study_dir / "review.json") == seal["review_sha256"],
            "Review changed after sealing")
    require(seal["key_opened_before_seal"] is False,
            "Review is not declared label-hidden")
    require(sha(packet_dir / "key.json") == manifest["key_sha256"]
            == sha(study_dir / "key.json"), "Key mismatch")

    cases = packet["cases"]
    grades = key["cases"]
    decisions = review["decisions"]
    require(len(cases) == len(grades) == 6 and len(decisions) == 12,
            "Expected exactly six pairs and twelve decisions")
    decision_map = {(row["pair"], row["member"]): row for row in decisions}
    require(len(decision_map) == 12, "Duplicate review member")
    outcomes = []
    witness_split = False
    witness_count = 0
    for case, grade_case in zip(cases, grades):
        pair = case["id"]
        require(pair == grade_case["id"], "Pair identity mismatch")
        raw = (packet_dir / case["source_file"]).read_bytes()
        require(hashlib.sha256(raw).hexdigest() == case["source_sha256"]
                == manifest["source_sha256"][case["source_file"]],
                "Original source hash mismatch")
        require(len(case["members"]) == len(grade_case["members"]) == 2,
                "Pair membership changed")
        by_grade = {}
        for member, grade in zip(case["members"], grade_case["members"]):
            letter = member["id"]
            require(letter == grade["id"], "A/B mapping changed")
            span = member["span"]
            require(raw[span["start"]:span["end"]].decode("utf-8")
                    == member["quote"], "Original UTF-8 span mismatch")
            row = decision_map[(pair, letter)]
            require(row["decision"] in ("split_needed", "retain", "uncertain"),
                    "Unknown decision")
            require(bool(row["source_evidence_and_rationale"]), "Missing source rationale")
            require(grade["old_grade"] in ("actionable", "uncertain"),
                    "Unexpected prior grade")
            require(grade["old_grade"] not in by_grade, "Duplicate prior grade")
            by_grade[grade["old_grade"]] = row["decision"]
            if grade["full_event"]:
                witness_count += 1
                require(grade["full_event"] == "whole/p026-d02"
                        and grade["finding_id"] == "5f3c42940d99ad29",
                        "Unexpected FULL-event witness")
                witness_split = row["decision"] == "split_needed"
        require(set(by_grade) == {"actionable", "uncertain"},
                "Pair must have actionable and uncertain members")
        separated = (by_grade["actionable"] == "split_needed"
                     and by_grade["uncertain"] == "retain")
        outcomes.append({"pair": pair, "actionable": by_grade["actionable"],
                         "uncertain": by_grade["uncertain"], "separated": separated})
    require(set(decision_map) == {(case["id"], member["id"])
                                 for case in cases for member in case["members"]},
            "Extra review member")
    require(witness_count == 1, "Expected one FULL-event witness")
    separated_count = sum(row["separated"] for row in outcomes)
    return {"format": "unswell-long-sentence-dependency-result-v1",
            "packet_sha256": packet_hash, "review_sha256": seal["review_sha256"],
            "pairs": outcomes, "separated_pairs": separated_count,
            "witness_split_needed": witness_split,
            "screen_passed": separated_count >= 5 and witness_split}


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--packet-dir", type=Path, required=True)
    args = parser.parse_args()
    print(json.dumps(evaluate(args.packet_dir), indent=2) + "\n", end="")


if __name__ == "__main__":
    main()
