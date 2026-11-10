#!/usr/bin/env python3
"""Bind current-engine findings to the frozen source-only event and control spans."""

import argparse
import hashlib
import json
from pathlib import Path

from verify_reference import verify as verify_reference


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path: Path):
    return json.loads(path.read_text(encoding="utf-8"))


def overlap(left, right):
    return left["start"] < right["end"] and right["start"] < left["end"]


def compact(report):
    return [{"id": finding["id"], "rule_id": finding["rule_id"],
             "message": finding["message"],
             "primary": finding["primary"]["span"],
             "related": [item["span"] for item in finding.get("related", [])]}
            for finding in report["findings"]]


def audit(packet: Path, baseline: Path):
    reference = verify_reference(packet)
    summary = read(baseline / "summary.json")
    if (summary["format"] != "unswell-349-atomic-claim-baseline-v1" or
            summary["reference_sha256"] != sha(packet / "reference.json") or
            len(summary["results"]) != 9):
        raise ValueError("baseline identity or denominator changed")
    findings = []
    report_hashes = {}
    for page, result in enumerate(summary["results"], 1):
        report_path = baseline / f"page-{page:03d}.json"
        report = read(report_path)
        if (result["page"] != page or result["exit_code"] != 0 or
                result["status"] != "complete" or report["status"] != "complete" or
                result["report_sha256"] != sha(report_path) or
                result["report_source_sha256"] != result["source_sha256"] or
                len(report["documents"]) != 1 or
                report["documents"][0]["source_hash"] != result["source_sha256"] or
                len(report["findings"]) != result["findings"] or
                len(report["assessments"]) != result["assessments"] or
                report["manifest"]["tool_commit"] != result["tool_commit"]):
            raise ValueError(f"baseline report changed on page {page}")
        report_hashes[str(page)] = result["report_sha256"]
        for finding in compact(report):
            finding["page"] = page
            findings.append(finding)
    if len(findings) != 68 or sum(row["assessments"] for row in summary["results"]) != 1513:
        raise ValueError("baseline count changed")
    hits = {}
    for label in reference["events"] + reference["controls"]:
        matching = []
        for finding in findings:
            if finding["page"] != label["page"]:
                continue
            if overlap(label["byte_span"], finding["primary"]):
                matching.append({"finding_id": finding["id"], "rule_id": finding["rule_id"],
                                 "kind": "primary", "message": finding["message"]})
            elif any(overlap(label["byte_span"], span) for span in finding["related"]):
                matching.append({"finding_id": finding["id"], "rule_id": finding["rule_id"],
                                 "kind": "related", "message": finding["message"]})
        if matching:
            hits[label["id"]] = matching
    return {"format": "unswell-349-atomic-claim-baseline-audit-v1",
            "reference_sha256": sha(packet / "reference.json"),
            "baseline_summary_sha256": sha(baseline / "summary.json"),
            "binary_sha256": summary["binary_sha256"],
            "report_sha256": report_hashes,
            "finding_count": len(findings),
            "assessment_count": 1513,
            "event_primary_or_related_overlap": sum(event["id"] in hits for event in reference["events"]),
            "event_count": len(reference["events"]),
            "control_primary_or_related_overlap": sum(control["id"] in hits for control in reference["controls"]),
            "control_count": len(reference["controls"]),
            "hits": hits,
            "findings": findings}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--packet", type=Path, required=True)
    parser.add_argument("--baseline", type=Path, required=True)
    args = parser.parse_args()
    result = audit(args.packet, args.baseline)
    output = args.packet / "baseline-audit.json"
    output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    (args.packet / "baseline-summary.json").write_bytes((args.baseline / "summary.json").read_bytes())
    print(f"{result['finding_count']} findings; {result['event_primary_or_related_overlap']}/{result['event_count']} event spans overlap")


if __name__ == "__main__":
    main()
