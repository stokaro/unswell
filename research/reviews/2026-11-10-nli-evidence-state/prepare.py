#!/usr/bin/env python3
"""Bind the original frozen support-flip cases to target-free NLI premises."""

import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
REPO = HERE.parents[2]
ORIGINAL = REPO / "research/reviews/2026-11-08-evidence-flip-pilot/packet"
EXPECTED = {
    ("table-announcement", "matching-heading"): "support",
    ("table-announcement", "unrelated-distractor"): "support",
    ("table-announcement", "heading-removed"): "no_evidence",
    ("table-announcement", "table-contradicted"): "contradiction",
    ("repeated-release-status", "matching-heading"): "support",
    ("repeated-release-status", "prior-fact-added"): "support",
    ("repeated-release-status", "unrelated-distractor"): "no_evidence",
    ("repeated-release-status", "prior-fact-contradicts"): "contradiction",
    ("unsupported-comparison", "comparison-added"): "support",
    ("unsupported-comparison", "no-comparison"): "no_evidence",
    ("unsupported-comparison", "unrelated-distractor"): "no_evidence",
    ("unsupported-comparison", "comparison-contradicts"): "contradiction",
    ("unsupported-reader-need", "reader-evidence-added"): "support",
    ("unsupported-reader-need", "no-reader-evidence"): "no_evidence",
    ("unsupported-reader-need", "unrelated-distractor"): "no_evidence",
    ("unsupported-reader-need", "reader-evidence-contradicts"): "contradiction",
}


def digest(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def build() -> tuple[dict, dict]:
    frozen = json.loads((ORIGINAL / "freeze.json").read_text())
    for name, expected in frozen.items():
        actual = digest((ORIGINAL / name).read_bytes())
        if actual != expected:
            raise ValueError(f"original frozen input changed: {name}")
    key_rows = json.loads((ORIGINAL / "key.json").read_text())
    if len(key_rows) != 16:
        raise ValueError("expected sixteen original cases")
    cases = []
    expected_rows = []
    seen = set()
    for row in sorted(key_rows, key=lambda item: item["item_id"]):
        case_id, variant_id = row["case_id"], row["variant_id"]
        pair = (case_id, variant_id)
        if pair not in EXPECTED or pair in seen:
            raise ValueError(f"unexpected or repeated case: {pair}")
        seen.add(pair)
        raw = (ORIGINAL / "requests" / (row["item_id"] + ".txt")).read_text()
        item = json.loads(raw.split("INPUT JSON:\n", 1)[1])
        source = item["source"].encode("utf-8")
        start, end = item["target_start_byte"], item["target_end_byte"]
        target = item["target_quote"].encode("utf-8")
        if item["item_id"] != row["item_id"] or not (0 <= start < end <= len(source)):
            raise ValueError("target identity or bounds changed")
        if source[start:end] != target or digest(source) != item["source_sha256"]:
            raise ValueError("target bytes or source digest changed")
        cases.append({"item_id": row["item_id"], "source_sha256": digest(source),
                      "target_start_byte": start, "target_end_byte": end,
                      "premise": (source[:start] + source[end:]).decode("utf-8"),
                      "hypothesis": target.decode("utf-8")})
        expected_rows.append({"item_id": row["item_id"], "case_id": case_id,
                              "variant_id": variant_id,
                              "expected_relation": EXPECTED[pair],
                              "editorial_applies": row["applies"]})
    if seen != set(EXPECTED):
        raise ValueError("missing evidence-state case")
    return ({"format": "unswell-349-nli-evidence-input-v1", "cases": cases},
            {"format": "unswell-349-nli-evidence-key-v1", "cases": expected_rows})


def main() -> None:
    packet, key = build()
    for name, payload in (("packet.json", packet), ("key.json", key)):
        path = HERE / name
        data = (json.dumps(payload, ensure_ascii=False, indent=2) + "\n").encode()
        if path.exists() and path.read_bytes() != data:
            raise ValueError(f"existing {name} differs")
        path.write_bytes(data)


if __name__ == "__main__":
    main()
