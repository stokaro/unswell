#!/usr/bin/env python3
"""Replay a frozen external-control baseline through the current CLI."""

import argparse
import gzip
import hashlib
import io
import json
import subprocess
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]


def sha256(data):
    return hashlib.sha256(data).hexdigest()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--binary", type=Path, required=True)
    args = parser.parse_args()
    frozen = json.loads((HERE / "freeze-v2.json").read_text())
    controls_bytes = (HERE / "controls.json").read_bytes()
    code_bytes = (HERE / "run-v2.py").read_bytes()
    if sha256(controls_bytes) != frozen["controls_sha256"]:
        raise SystemExit("controls changed after freeze")
    if sha256(code_bytes) != frozen["script_sha256"]:
        raise SystemExit("script changed after freeze")
    binary_hash = sha256(args.binary.read_bytes())
    cases = json.loads(controls_bytes)["cases"]
    complete = []
    raw = []
    for case in cases:
        name = case["id"] + ".md"
        command = [str(args.binary), "check", "--stdin", "--filename", name,
                   "--no-gate", "--report", "json:-"]
        run = subprocess.run(command, cwd=ROOT, input=case["source"].encode(),
                             stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                             timeout=120, check=False)
        if run.returncode != 0:
            raise SystemExit(f"{case['id']}: exit {run.returncode}: {run.stderr.decode()}")
        report = json.loads(run.stdout)
        if report["status"] != "complete" or report["errors"]:
            raise SystemExit(f"{case['id']}: incomplete report")
        if len(report["documents"]) != 1 or report["documents"][0]["name"] != name:
            raise SystemExit(f"{case['id']}: unexpected document identity")
        findings = []
        for finding in report["findings"]:
            span = finding["primary"]["span"]
            source = case["source"].encode()
            quote = source[span["start"]:span["end"]].decode("utf-8")
            findings.append({"rule_id": finding["rule_id"], "span": span,
                             "quote": quote, "message": finding.get("message"),
                             "suggestion": finding.get("evidence", {}).get("suggestion"),
                             "related": [v["span"] for v in finding.get("related", [])],
                             "suppressed": finding.get("suppressed", False)})
        complete.append({"id": case["id"], "family": case["family"],
                         "role": case["role"], "source_sha256": sha256(case["source"].encode()),
                         "findings": findings})
        raw.append({"id": case["id"], "stdout": report,
                    "stderr": run.stderr.decode(), "exit_code": run.returncode})
    raw_bytes = json.dumps(raw, ensure_ascii=False, sort_keys=True,
                           separators=(",", ":")).encode()
    buffer = io.BytesIO()
    with gzip.GzipFile(fileobj=buffer, mode="wb", filename="", mtime=0) as out:
        out.write(raw_bytes)
    (HERE / "reports-v2.json.gz").write_bytes(buffer.getvalue())
    result = {"format": "unswell-external-control-baseline-v1",
              "base_commit": frozen["base_commit"], "binary_sha256": binary_hash,
              "controls_sha256": frozen["controls_sha256"],
              "script_sha256": frozen["script_sha256"],
              "raw_sha256": sha256(raw_bytes), "raw_gzip_sha256": sha256(buffer.getvalue()),
              "cases": complete}
    result_bytes = (json.dumps(result, ensure_ascii=False, sort_keys=True,
                               indent=2) + "\n").encode()
    (HERE / "result-v2.json").write_bytes(result_bytes)
    print(sha256(result_bytes))


if __name__ == "__main__":
    main()
