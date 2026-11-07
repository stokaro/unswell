#!/usr/bin/env python3
"""Independently verify the sealed current-engine control outputs."""

import gzip
import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent


def digest(data):
    return hashlib.sha256(data).hexdigest()


def require(condition, message):
    if not condition:
        raise ValueError(message)


def main():
    freeze = json.loads((HERE / "freeze-v2.json").read_text())
    controls_bytes = (HERE / "controls.json").read_bytes()
    require(digest(controls_bytes) == freeze["controls_sha256"], "control freeze mismatch")
    require(digest((HERE / "run-v2.py").read_bytes()) == freeze["script_sha256"],
            "runner freeze mismatch")
    controls = json.loads(controls_bytes)["cases"]
    result = json.loads((HERE / "result-v2.json").read_bytes())
    compressed = (HERE / "reports-v2.json.gz").read_bytes()
    raw_bytes = gzip.decompress(compressed)
    require(digest(compressed) == result["raw_gzip_sha256"], "compressed report hash mismatch")
    require(digest(raw_bytes) == result["raw_sha256"], "raw report hash mismatch")
    raw = json.loads(raw_bytes)
    require(result["base_commit"] == freeze["base_commit"], "base commit mismatch")
    require(result["controls_sha256"] == freeze["controls_sha256"],
            "result control hash mismatch")
    require(result["script_sha256"] == freeze["script_sha256"],
            "result runner hash mismatch")
    require(len(raw) == len(result["cases"]) == len(controls) == 18,
            "control coverage mismatch")
    finding_count = 0
    config_hashes = set()
    for case, row, output in zip(controls, result["cases"], raw, strict=True):
        require(case["id"] == row["id"] == output["id"], "case identity mismatch")
        require(case["family"] == row["family"] and case["role"] == row["role"],
                "case metadata mismatch")
        source = case["source"].encode("utf-8")
        require(row["source_sha256"] == digest(source), "case source hash mismatch")
        report = output["stdout"]
        require(output["exit_code"] == 0 and report["status"] == "complete" and
                report["manifest"]["complete"] and not report["errors"] and
                not report["manifest"]["skipped_rules"], "incomplete CLI output")
        require(len(report["documents"]) == 1 and
                report["documents"][0]["name"] == case["id"] + ".md" and
                report["documents"][0]["source_hash"] == digest(source),
                "document source identity mismatch")
        config_hashes.add(report["manifest"]["config_hash"])
        require(len(row["findings"]) == len(report["findings"]), "finding count mismatch")
        for compact, finding in zip(row["findings"], report["findings"], strict=True):
            span = finding["primary"]["span"]
            require(compact["rule_id"] == finding["rule_id"] and compact["span"] == span and
                    compact["message"] == finding["message"] and
                    compact["suggestion"] == finding["evidence"]["suggestion"] and
                    compact["suppressed"] == finding["suppressed"] and
                    compact["related"] == [related["span"] for related in finding["related"]],
                    "compact finding differs from raw report")
            require(0 <= span["start"] < span["end"] <= len(source),
                    "finding span outside original source")
            require(source[span["start"]:span["end"]].decode("utf-8") == compact["quote"],
                    "finding quote differs from original bytes")
            for related in compact["related"]:
                require(0 <= related["start"] < related["end"] <= len(source) and
                        bool(source[related["start"]:related["end"]].decode("utf-8")),
                        "related span outside original source")
            finding_count += 1
    require(len(config_hashes) == 1, "mixed configuration hashes")
    print(json.dumps({"cases": len(controls), "findings": finding_count,
                      "config_hash": next(iter(config_hashes)),
                      "result_sha256": digest((HERE / "result-v2.json").read_bytes())},
                     sort_keys=True))


if __name__ == "__main__":
    main()
