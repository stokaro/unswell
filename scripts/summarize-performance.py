#!/usr/bin/env python3
"""Summarize retained scan reports without treating missing analysis as zero."""

import hashlib
import json
from pathlib import Path
import sys
import tempfile


def identity(path):
    if not path.is_file():
        return None
    return {"bytes": path.stat().st_size, "sha256": hashlib.sha256(path.read_bytes()).hexdigest()}


def sample(root, name, legacy, observation):
    directory = root / legacy
    result = {
        "name": name,
        "seconds": observation[f"{legacy}_seconds"],
        "exit_code": observation[f"{legacy}_exit_code"],
        "maximum_resident_bytes": observation[f"{legacy}_maximum_resident_bytes"],
        "reports": {extension: identity(directory / f"result.{extension}")
                    for extension in ("json", "sarif", "html", "md", "txt")},
        "stderr": identity(directory / "stderr.txt"),
        "analysis": None,
    }
    source = directory / "result.json"
    if not source.is_file() or source.stat().st_size == 0:
        result["report_status"] = "missing"
        return result
    try:
        report = json.loads(source.read_text(encoding="utf-8"))
    except json.JSONDecodeError:
        result["report_status"] = "malformed"
        return result
    documents = report["documents"]
    excluded = {}
    for document in documents:
        for exclusion in document.get("excluded") or []:
            row = excluded.setdefault(exclusion["reason"], {"spans": 0, "span_bytes": 0})
            row["spans"] += 1
            row["span_bytes"] += exclusion["span"]["end"] - exclusion["span"]["start"]
    manifest = report["manifest"]
    document_manifest = [{key: document[key] for key in
                          ("name", "format", "source_hash", "bytes", "prose_words", "excluded")}
                         for document in documents]
    encoded = json.dumps(document_manifest, sort_keys=True, separators=(",", ":")).encode()
    result["report_status"] = "available"
    result["analysis"] = {
        "status": report["status"],
        "gate_passed": report["gate"]["passed"],
        "documents": len(documents),
        "prose_words": sum(document["prose_words"] for document in documents),
        "findings": len(report["findings"]),
        "errors": report["errors"],
        "exclusions_by_reason": excluded,
        "abstentions": report.get("abstentions", []),
        "document_manifest_sha256": hashlib.sha256(encoded).hexdigest(),
        "manifest": {key: value for key, value in manifest.items() if key != "rules"},
    }
    return result


def summarize(root):
    observation = json.loads((root / "observation.json").read_text(encoding="utf-8"))
    samples = [sample(root, "first", "cold", observation), sample(root, "repeat", "warm", observation)]
    obsolete = {"measured_prose_words", "documents", "findings", "errors", "outcome", "gate_evaluated"}
    result = {key: value for key, value in observation.items()
              if key not in obsolete and not key.startswith(("cold_", "warm_"))}
    result.update({
        "format": "unswell-scan-cost-observation-v2",
        "cache_control": "uncontrolled; fresh process per sample; corpus preparation and calibration may warm files",
        "policy": identity(root / "policy.yaml"),
        "samples": samples,
    })
    return result


def self_test():
    # Distinct outcomes catch accidental reuse of the first sample's report.
    with tempfile.TemporaryDirectory() as temporary:
        root = Path(temporary)
        observation = {f"{name}_{key}": value for name in ("cold", "warm")
                       for key, value in (("seconds", 1.5), ("exit_code", 2), ("maximum_resident_bytes", 100))}
        (root / "observation.json").write_text(json.dumps(observation))
        (root / "cold").mkdir()
        (root / "warm").mkdir()
        report = {"documents": [], "findings": [], "errors": [{"path": "bad.go", "message": "parse failed"}],
                  "status": "incomplete", "gate": {"passed": False}, "manifest": {"complete": False}}
        (root / "cold/result.json").write_text(json.dumps(report))
        result = summarize(root)
        assert result["samples"][0]["analysis"]["errors"] == report["errors"]
        assert result["samples"][1]["analysis"] is None
        assert result["samples"][1]["report_status"] == "missing"
        (root / "warm/result.json").write_text("{")
        assert summarize(root)["samples"][1]["report_status"] == "malformed"
        report.update({"status": "complete", "errors": [], "abstentions": [{"reason": "budget_exhausted"}]})
        (root / "warm/result.json").write_text(json.dumps(report))
        result = summarize(root)
        assert result["samples"][0]["analysis"]["status"] == "incomplete"
        assert result["samples"][1]["analysis"]["status"] == "complete"
        assert result["samples"][1]["analysis"]["abstentions"] == report["abstentions"]
        assert "cold_seconds" not in result
    print("Scan summaries preserve independent outcomes, missing reports, errors, and abstentions.")


if __name__ == "__main__":
    if len(sys.argv) == 2 and sys.argv[1] == "--self-test":
        self_test()
    elif len(sys.argv) == 2:
        print(json.dumps(summarize(Path(sys.argv[1])), indent=2, sort_keys=True))
    else:
        sys.exit("Usage: summarize-performance.py ARTIFACT_DIR | --self-test")
