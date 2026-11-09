#!/usr/bin/env python3
"""Source-function candidate with document-structure-aware edit proposals."""

import argparse
import hashlib
import json
import re
from pathlib import Path

import candidate


FROZEN_CORE_SHA256 = "1291ffc4fec191f83d3811a4fa7b5698f1ccf2cd0aed244fa9c8a9bcc4777081"
LIST_MARKER = re.compile(rb"^[ \t]*(?:[0-9]+[.)]|[-*+])[ \t]*")


def plan_edit(source: bytes, diagnosis: dict, excluded: list[dict]) -> dict:
    result = dict(diagnosis)
    if result["edit_status"] != "proposed_unverified":
        return result
    edit = dict(result["edit"])
    start, end = edit["start_byte"], edit["end_byte"]
    line_start = source.rfind(b"\n", 0, start) + 1
    line_end = source.find(b"\n", start)
    if line_end < 0:
        line_end = len(source)
    if LIST_MARKER.match(source[line_start:line_end]):
        result.pop("edit")
        result["edit_status"] = "unresolved"
        result["edit_reason"] = "List structure may require renumbering or reflow."
        return result
    if end < len(source) and source[end:end + 1] == b" ":
        proposed_end = end + 1
        if not candidate.intersects_excluded(end, proposed_end, excluded):
            edit["end_byte"] = proposed_end
    if candidate.intersects_excluded(edit["start_byte"], edit["end_byte"], excluded):
        raise ValueError("proposed edit intersects excluded source bytes")
    result["edit"] = edit
    result["edit_reason"] = "Remove a complete redundant sentence and adjacent space when present."
    return result


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--sources-dir", type=Path, required=True)
    parser.add_argument("--reports-dir", type=Path, required=True)
    parser.add_argument("--inventory", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if args.output.exists():
        raise ValueError("candidate output already exists")
    core_path = Path(candidate.__file__)
    if hashlib.sha256(core_path.read_bytes()).hexdigest() != FROZEN_CORE_SHA256:
        raise ValueError("frozen candidate core changed")
    inventory_bytes = args.inventory.read_bytes()
    inventory = json.loads(inventory_bytes)
    sources = inventory["sources"]
    if len(sources) != 12 or any(s["number"] != i for i, s in enumerate(sources, 1)):
        raise ValueError("expected twelve ordered selected sources")
    outputs = []
    page_results = []
    for source_meta in sources:
        number = source_meta["number"]
        if source_meta["status"] != "valid":
            page_results.append({"page": number, "status": source_meta["status"],
                                 "diagnoses": 0})
            continue
        data = (args.sources_dir / source_meta["source_file"]).read_bytes()
        if hashlib.sha256(data).hexdigest() != source_meta["source_sha256"]:
            raise ValueError(f"page {number}: source hash mismatch")
        report_path = args.reports_dir / f"page-{number:03d}.json"
        report_bytes = report_path.read_bytes()
        report = json.loads(report_bytes)
        if report.get("status") != "complete" or report.get("errors"):
            page_results.append({"page": number, "status": "operational_error",
                                 "source_sha256": source_meta["source_sha256"],
                                 "report_sha256": hashlib.sha256(report_bytes).hexdigest(),
                                 "errors": report.get("errors", []),
                                 "eligible_sentences": 0, "diagnoses": 0})
            continue
        diagnoses, sentence_count = candidate.run_page(data, report, number)
        excluded = report["documents"][0]["excluded"]
        diagnoses = [plan_edit(data, finding, excluded) for finding in diagnoses]
        outputs.extend(diagnoses)
        page_results.append({"page": number, "status": "valid",
                             "source_sha256": source_meta["source_sha256"],
                             "report_sha256": hashlib.sha256(report_bytes).hexdigest(),
                             "eligible_sentences": sentence_count,
                             "diagnoses": len(diagnoses)})
    for index, finding in enumerate(outputs, 1):
        finding["id"] = f"transfer-v2-{index:03d}"
    result = {"format": "unswell-349-source-function-transfer-v2",
              "core_sha256": FROZEN_CORE_SHA256,
              "inventory_sha256": hashlib.sha256(inventory_bytes).hexdigest(),
              "pages": page_results, "diagnoses": outputs}
    args.output.write_text(json.dumps(result, indent=2, ensure_ascii=False) + "\n")
    print(f"processed {len(sources)} selected pages; emitted {len(outputs)} diagnoses")


if __name__ == "__main__":
    main()
