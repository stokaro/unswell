#!/usr/bin/env python3
"""Replay the technical CLI on eligible primary Markdown source snapshots."""

import argparse
import hashlib
import json
from pathlib import Path
import subprocess


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=Path, required=True)
    parser.add_argument("--sources", type=Path, required=True)
    parser.add_argument("--reports", type=Path, required=True)
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args()
    root = Path(__file__).resolve().parent
    manifest = json.loads((root / "sources.json").read_text(encoding="utf-8"))
    labels = json.loads((root / "source_only_labels.json").read_text(encoding="utf-8"))
    primary = {row["pr"] for row in labels}
    args.reports.mkdir(parents=True, exist_ok=True)
    rows = []
    for source_row in manifest:
        case = f"{source_row['repo']}#{source_row['pr']}"
        if case not in primary or not source_row["path"].endswith(".md"):
            continue
        source = (args.sources / source_row["source_file"]).read_bytes()
        if hashlib.sha256(source).hexdigest() != source_row["sha256"]:
            raise ValueError(f"source hash changed: {case}")
        report_file = args.reports / (source_row["source_file"] + ".json")
        command = [
            str(args.binary.resolve(strict=True)),
            "check", "--stdin", "--filename", source_row["path"],
            "--format", "markdown", "--profile", "technical", "--no-gate",
            "--report", f"json:{report_file}",
        ]
        process = subprocess.run(
            command, input=source, capture_output=True, timeout=90, check=False
        )
        report = json.loads(report_file.read_text(encoding="utf-8")) if report_file.exists() else None
        row = {
            "pr": case,
            "source_sha256": source_row["sha256"],
            "exit_code": process.returncode,
            "status": report.get("status") if report else None,
            "finding_count": len(report.get("findings", [])) if report else None,
            "error_count": len(report.get("errors", [])) if report else None,
            "report_sha256": hashlib.sha256(report_file.read_bytes()).hexdigest() if report else None,
            "findings": [
                {
                    "rule_id": finding["rule_id"],
                    "message": finding["message"],
                    "span": finding["primary"]["span"],
                    "line": finding["primary"]["start"]["line"],
                    "excerpt": source[
                        finding["primary"]["span"]["start"]:
                        finding["primary"]["span"]["end"]
                    ].decode("utf-8"),
                }
                for finding in report.get("findings", [])
            ] if report else [],
        }
        rows.append(row)
        print(case, row["status"], row["finding_count"], flush=True)
    tree = subprocess.check_output(["git", "rev-parse", "HEAD^{tree}"], text=True).strip()
    ledger = {
        "source_tree": tree,
        "binary_sha256": hashlib.sha256(args.binary.read_bytes()).hexdigest(),
        "profile": "technical",
        "rows": rows,
    }
    args.out.write_text(json.dumps(ledger, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    if any(row["exit_code"] != 0 or row["status"] != "complete" for row in rows):
        raise SystemExit("incomplete baseline replay")


if __name__ == "__main__":
    main()
