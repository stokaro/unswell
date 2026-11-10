#!/usr/bin/env python3
"""Run the unchanged local engine on every frozen source page once."""

import argparse
import hashlib
import json
import subprocess
from pathlib import Path

from verify_reference import verify as verify_reference


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--packet", type=Path, required=True)
    parser.add_argument("--binary", type=Path, required=True)
    parser.add_argument("--project-root", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    args = parser.parse_args()
    if args.output_dir.exists():
        raise SystemExit("baseline output directory already exists")
    reference = verify_reference(args.packet)
    source_inventory = json.loads((args.packet / "sources" / "inventory.json")
                                  .read_text(encoding="utf-8"))
    args.output_dir.mkdir(parents=True)
    results = []
    for record in source_inventory["sources"]:
        number = record["number"]
        source = args.packet / "sources" / record["source_file"]
        report = args.output_dir / f"page-{number:03d}.json"
        command = [str(args.binary.resolve()), "check", "--stdin",
                   "--filename", f"page-{number:03d}.md", "--format", "markdown",
                   "--no-gate", "--project-root", str(args.project_root.resolve()),
                   "--report", f"json:{report.resolve()}"]
        completed = subprocess.run(command, input=source.read_bytes(),
                                   capture_output=True, timeout=120, check=False)
        (args.output_dir / f"page-{number:03d}.stdout").write_bytes(completed.stdout)
        (args.output_dir / f"page-{number:03d}.stderr").write_bytes(completed.stderr)
        result = {"page": number, "exit_code": completed.returncode,
                  "source_sha256": record["source_sha256"]}
        if report.exists():
            data = json.loads(report.read_text(encoding="utf-8"))
            result.update({"report_sha256": digest(report), "status": data["status"],
                           "tool_commit": data["manifest"]["tool_commit"],
                           "documents": len(data["documents"]),
                           "findings": len(data["findings"]),
                           "assessments": len(data["assessments"])})
            if data["documents"]:
                result["report_source_sha256"] = data["documents"][0]["source_hash"]
        else:
            result["status"] = "missing_report"
        results.append(result)
        print(f"{number}/9: {result['status']}, {result.get('findings', 0)} findings",
              flush=True)
        if (completed.returncode or result["status"] != "complete" or
                result.get("documents") != 1 or
                result.get("report_source_sha256") != record["source_sha256"]):
            break
    summary = {
        "format": "unswell-349-atomic-claim-baseline-v1",
        "reference_sha256": digest(args.packet / "reference.json"),
        "binary_sha256": digest(args.binary),
        "results": results,
    }
    (args.output_dir / "summary.json").write_text(
        json.dumps(summary, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    if len(results) != 9 or any(result["status"] != "complete" for result in results):
        raise SystemExit("baseline did not complete all nine pages")


if __name__ == "__main__":
    main()
