#!/usr/bin/env python3
"""Audit every v2 diagnosis and unchanged-CLI finding against sealed labels."""

import argparse
import hashlib
import json
from pathlib import Path


# Decisions were made after the single v2 run, without changing sealed labels
# or candidate rules. 'partial' is not an accepted complete diagnosis.
CANDIDATE = [
    ("accepted", "Correctly removes an unsupported generation label."),
    ("optional", "A promotional heading can be shortened, but this was not a sealed clear event."),
    ("rejected", "Rich Model Support is followed by a concrete provider and format list."),
    ("optional", "Rich interface design is subjective marketing language with nearby detail."),
    ("rejected", "Advanced Markdown syntax is scoped by listed renderer features."),
    ("rejected", "Rich File Parsing is followed by supported input and storage formats."),
    ("rejected", "Rich search engines is scoped by a named engine list."),
    ("optional", "Rich dashboard is promotional but followed by management features."),
    ("rejected", "Advanced cache settings is scoped by specific settings."),
    ("optional", "Rich interface design in the project rationale is a voice choice."),
    ("partial", "Next-generation is promotional, but the sealed event is the entire repeated paragraph."),
    ("rejected", "Next generation names the actual cMeta successor."),
    ("rejected", "Next generation names the actual cMeta successor again."),
    ("optional", "Prophecy's informal introduction has a concrete example and was sealed optional."),
    ("optional", "Oban's broad introduction has detailed features and was sealed optional."),
    ("accepted", "Correctly marks the unsupported generation label in Singularity's opening."),
    ("optional", "Informal Breakpoint recommendation is accompanied by technical detail."),
    ("rejected", "A feature checklist is a distinct document function from later technical detail."),
    ("partial", "Advanced is unsupported, but leaving 'most' does not resolve the ranking event."),
    ("rejected", "Advanced Facial Animation is supported by the adjacent 33-expression and 400-combination claims."),
    ("rejected", "The A800 benchmark table cell records a technical failure; deleting it corrupts the comparison."),
    ("rejected", "The A800 table summary states hardware, resolution, duration, and compared implementations."),
]

BASELINE = {
    1: [],
    2: ["accepted"] * 9,
    3: [],
    4: ["accepted"] * 3,
    5: ["accepted"] * 5,
    6: ["optional"],
    7: ["accepted", "accepted", "optional"],
    8: [],
    9: ["rejected", "accepted"],
    10: ["rejected"],
    11: ["accepted"],
    12: ["rejected"],
}

BASELINE_REASONS = {
    (6, 1): "The quality phrase is a source-only optional control with nearby feature detail.",
    (7, 3): "The 229-byte sentence has a technical distinction; splitting is a voice choice.",
    (9, 1): "The conditional paper-update instruction must retain its if-condition.",
    (10, 1): "The purported long sentence crosses an introduction and a lettered list item.",
    (12, 1): "Module appears in distinct table cells, not as an accidental prose duplicate.",
}


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--review-dir", type=Path, required=True)
    parser.add_argument("--sources-dir", type=Path, required=True)
    parser.add_argument("--reports-dir", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if args.output.exists():
        raise ValueError("audit output already exists")
    labels_path = args.review_dir / "transfer-v2-labels.json"
    diagnoses_path = args.review_dir / "transfer-v2-candidate-output.json"
    labels = json.loads(labels_path.read_bytes())["judgments"]
    candidate = json.loads(diagnoses_path.read_bytes())
    diagnoses = candidate["diagnoses"]
    if len(diagnoses) != len(CANDIDATE):
        raise ValueError("unexpected diagnosis count")
    pages = {n: (args.sources_dir / f"page-{n:03d}.source").read_bytes() for n in range(1, 13)}
    out_candidate = []
    for index, (diagnosis, (decision, reason)) in enumerate(zip(diagnoses, CANDIDATE), 1):
        if diagnosis["id"] != f"transfer-v2-{index:03d}":
            raise ValueError("diagnosis order changed")
        page = diagnosis["page"]
        raw = pages[page]
        start, end = diagnosis["start_byte"], diagnosis["end_byte"]
        if raw[start:end].decode("utf-8") != diagnosis["quote"]:
            raise ValueError("diagnosis is not source-bound")
        overlapping = [label["id"] for label in labels if label["page"] == page
                       and start < label["end_byte"] and end > label["start_byte"]]
        unsafe = index in (21, 22)
        if diagnosis["edit_status"] == "proposed_unverified":
            edit = diagnosis["edit"]
            if edit["operation"] != "delete" or not (0 <= edit["start_byte"] < edit["end_byte"] <= len(raw)):
                raise ValueError("unexpected edit")
        out_candidate.append({"id": diagnosis["id"], "page": page, "decision": decision,
                              "reason": reason, "overlapping_sealed_labels": overlapping,
                              "unsafe_supplied_edit": unsafe})
    out_baseline = []
    for page in range(1, 13):
        report = json.loads((args.reports_dir / f"page-{page:03d}.json").read_bytes())
        if page == 3:
            if report["status"] != "incomplete" or not report["errors"]:
                raise ValueError("expected recorded parser failure")
            continue
        if report["status"] != "complete" or report["errors"]:
            raise ValueError(f"unexpected incomplete baseline: {page}")
        findings = report["findings"]
        if len(findings) != len(BASELINE[page]):
            raise ValueError(f"baseline finding count changed: {page}")
        for index, (finding, decision) in enumerate(zip(findings, BASELINE[page]), 1):
            span = finding["primary"]["span"]
            if not (0 <= span["start"] < span["end"] <= len(pages[page])):
                raise ValueError("invalid baseline source span")
            reason = BASELINE_REASONS.get((page, index),
                                          "The indicated span has a concrete readability edit that preserves its technical conditions.")
            out_baseline.append({"page": page, "index": index, "finding_id": finding["id"],
                                 "rule_id": finding["rule_id"], "decision": decision,
                                 "reason": reason})
    output = {"format": "unswell-349-transfer-v2-audit-v1",
              "reviewer": "root Codex assistant; not an independent human annotator",
              "labels_sha256": hashlib.sha256(labels_path.read_bytes()).hexdigest(),
              "candidate_sha256": hashlib.sha256(diagnoses_path.read_bytes()).hexdigest(),
              "candidate": out_candidate, "baseline": out_baseline}
    args.output.write_text(json.dumps(output, indent=2) + "\n")
    print(f"audited {len(out_candidate)} candidate diagnoses and {len(out_baseline)} baseline findings")


if __name__ == "__main__":
    main()
