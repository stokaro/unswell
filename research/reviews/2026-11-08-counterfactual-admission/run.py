"""Run the unchanged CLI on all frozen counterfactual admission controls."""

import argparse
import hashlib
import json
import subprocess
from pathlib import Path

from verify import DIRECTORY, verify_freeze


def digest(data):
    return hashlib.sha256(data).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=Path, required=True)
    parser.add_argument("--repo-root", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    root = args.repo_root.resolve()
    binary = args.binary.resolve()
    if not binary.is_file():
        raise ValueError("binary does not exist")
    commit = subprocess.check_output(
        ["git", "rev-parse", "HEAD"], cwd=root, text=True).strip()
    packet = json.loads((DIRECTORY / "cases.json").read_text(encoding="utf-8"))
    receipts = {(r["case"], r["variant"]): r for r in verify_freeze()}
    results = []
    raw_dir = args.output.resolve().parent / "baseline-reports"
    raw_dir.mkdir(parents=True, exist_ok=True)
    for case in packet["cases"]:
        for variant in case["variants"]:
            case_id = case["id"]
            variant_id = variant["id"]
            receipt = receipts[(case_id, variant_id)]
            filename = f"{case_id}-{variant_id}.md"
            report_path = raw_dir / f"{case_id}-{variant_id}.json"
            process = subprocess.run(
                [str(binary), "check", "--stdin", "--filename", filename,
                 "--config", ".unswell.yaml", "--timeout", "2m",
                 "--report", f"json:{report_path}"],
                input=variant["source"].encode("utf-8"),
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                cwd=root,
                timeout=150,
                check=False,
            )
            if process.returncode not in (0, 1):
                raise ValueError(
                    f"{case_id}/{variant_id}: CLI failed with {process.returncode}: "
                    + process.stderr.decode("utf-8", errors="replace")[:500])
            raw = report_path.read_bytes()
            report = json.loads(raw)
            if report.get("status") != "complete" or report.get("errors"):
                raise ValueError(f"{case_id}/{variant_id}: incomplete report")
            if report["manifest"]["tool_commit"] != commit:
                raise ValueError(f"{case_id}/{variant_id}: tool commit mismatch")
            if len(report["documents"]) != 1:
                raise ValueError(f"{case_id}/{variant_id}: expected one document")
            document = report["documents"][0]
            if document["source_hash"] != receipt["source_sha256"]:
                raise ValueError(f"{case_id}/{variant_id}: source hash mismatch")
            start, end = receipt["target_start"], receipt["target_end"]
            findings = []
            overlapping = []
            for finding in report["findings"]:
                primary = finding["primary"]
                row = {
                    "id": finding["id"],
                    "rule": finding["rule_id"],
                    "message": finding["message"],
                    "span": primary["span"],
                }
                findings.append(row)
                span = primary["span"]
                if span["start"] < end and span["end"] > start:
                    overlapping.append(row)
            results.append({
                "case": case_id,
                "variant": variant_id,
                "role": variant["role"],
                "expected_fixed_diagnosis_applies": variant["applies"],
                "source_sha256": receipt["source_sha256"],
                "target_start": start,
                "target_end": end,
                "report_sha256": digest(raw),
                "cli_exit": process.returncode,
                "status": report["status"],
                "gate_passed": report["gate"]["passed"],
                "findings": findings,
                "target_overlapping_findings": overlapping,
            })
    summary = {
        "format": "unswell-counterfactual-admission-baseline-v1",
        "research_role": "descriptive unchanged CLI baseline; no fixed-diagnosis verdict inferred",
        "tool_commit": commit,
        "packet_sha256": digest((DIRECTORY / "cases.json").read_bytes()),
        "variants": results,
    }
    args.output.write_text(json.dumps(summary, indent=2, sort_keys=True) + "\n")
    print({
        "variants": len(results),
        "findings": sum(len(r["findings"]) for r in results),
        "overlapping_findings": sum(
            len(r["target_overlapping_findings"]) for r in results),
        "nonzero_exits": sum(r["cli_exit"] != 0 for r in results),
    })


if __name__ == "__main__":
    main()
