"""Validate the frozen synthetic counterfactual admission controls."""

import hashlib
import json
from pathlib import Path

DIRECTORY = Path(__file__).resolve().parent
EXPECTED_ROLES = {
    "table-announcement": {
        "matching-heading": (True, "diagnosis-supported"),
        "unrelated-distractor": (True, "distractor"),
        "heading-removed": (False, "support-removed"),
        "table-contradicted": (False, "different-problem"),
    },
    "unsupported-comparison": {
        "no-comparison": (True, "diagnosis-supported"),
        "unrelated-distractor": (True, "distractor"),
        "comparison-added": (False, "support-added"),
        "comparison-contradicts": (False, "different-problem"),
    },
    "unsupported-reader-need": {
        "no-reader-evidence": (True, "diagnosis-supported"),
        "unrelated-distractor": (True, "distractor"),
        "reader-evidence-added": (False, "support-added"),
        "reader-evidence-contradicts": (False, "different-problem"),
    },
    "repeated-release-status": {
        "matching-heading": (True, "diagnosis-supported"),
        "unrelated-distractor": (False, "distractor"),
        "prior-fact-added": (True, "support-added"),
        "prior-fact-contradicts": (False, "different-problem"),
    },
}


def sha256(data):
    return hashlib.sha256(data).hexdigest()


def validate(packet):
    if packet.get("format") != "unswell-counterfactual-admission-cases-v1":
        raise ValueError("unsupported packet format")
    if packet.get("original_349_events_used") is not False:
        raise ValueError("original event labels must not enter this packet")
    if packet.get("independent_human") is not False or not packet.get("reviewer"):
        raise ValueError("review provenance is missing")
    cases = packet.get("cases")
    if not isinstance(cases, list) or len(cases) != 4:
        raise ValueError("expected four fixed quartets")
    if {case.get("id") for case in cases} != set(EXPECTED_ROLES):
        raise ValueError("case identities changed")
    receipt = []
    seen_sources = set()
    for case in cases:
        case_id = case["id"]
        target = case.get("target")
        diagnosis = case.get("diagnosis")
        if not isinstance(target, str) or not target or not isinstance(diagnosis, str) or not diagnosis:
            raise ValueError(f"{case_id}: missing target or fixed diagnosis")
        target_bytes = target.encode("utf-8")
        variants = case.get("variants")
        expected = EXPECTED_ROLES[case_id]
        if not isinstance(variants, list) or {v.get("id") for v in variants} != set(expected):
            raise ValueError(f"{case_id}: variant identities changed")
        for variant in variants:
            variant_id = variant["id"]
            want, role = expected[variant_id]
            if type(variant.get("applies")) is not bool or variant["applies"] != want:
                raise ValueError(f"{case_id}/{variant_id}: expected decision changed")
            if variant.get("role") != role or not variant.get("rationale"):
                raise ValueError(f"{case_id}/{variant_id}: role or rationale changed")
            source = variant.get("source")
            if not isinstance(source, str) or not source.endswith("\n") or "\r" in source:
                raise ValueError(f"{case_id}/{variant_id}: source is not normalized UTF-8 text")
            raw = source.encode("utf-8")
            if raw.count(target_bytes) != 1:
                raise ValueError(f"{case_id}/{variant_id}: target must occur exactly once")
            source_hash = sha256(raw)
            if source_hash in seen_sources:
                raise ValueError(f"{case_id}/{variant_id}: duplicate source")
            seen_sources.add(source_hash)
            start = raw.index(target_bytes)
            receipt.append({
                "case": case_id,
                "variant": variant_id,
                "source_sha256": source_hash,
                "target_start": start,
                "target_end": start + len(target_bytes),
                "applies": want,
            })
    if len(receipt) != 16 or sum(row["applies"] for row in receipt) != 8:
        raise ValueError("expected 16 variants and eight positive decisions")
    return sorted(receipt, key=lambda row: (row["case"], row["variant"]))


def verify_freeze():
    packet_path = DIRECTORY / "cases.json"
    protocol_path = DIRECTORY / "PROTOCOL.md"
    freeze_path = DIRECTORY / "freeze.json"
    packet = json.loads(packet_path.read_text(encoding="utf-8"))
    receipt = validate(packet)
    freeze = json.loads(freeze_path.read_text(encoding="utf-8"))
    if freeze.get("format") != "unswell-counterfactual-admission-freeze-v1":
        raise ValueError("unsupported freeze format")
    for name, path in {
        "cases": packet_path,
        "protocol": protocol_path,
        "verifier": Path(__file__).resolve(),
        "runner": DIRECTORY / "run.py",
    }.items():
        if sha256(path.read_bytes()) != freeze["sha256"][name]:
            raise ValueError(f"{name} changed after freeze")
    if receipt != freeze["receipt"]:
        raise ValueError("source or target receipts changed after freeze")
    return receipt


if __name__ == "__main__":
    rows = verify_freeze()
    print({"variants": len(rows), "positive": sum(r["applies"] for r in rows)})
