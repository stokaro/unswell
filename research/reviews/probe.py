#!/usr/bin/env python3
"""Record compact known limitations without treating them as expected quality."""

import argparse
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile

sys.dont_write_bytecode = True
from compare import digest, read, require, write


def probe(binary, fixtures):
    with tempfile.TemporaryDirectory(prefix="unswell-review-probe-") as temporary:
        root = Path(temporary)
        sources = []
        for source in sorted(fixtures.glob("*.txt")):
            name = source.name.removesuffix(".txt")
            shutil.copyfile(source, root / name)
            sources.append({"path": name, "sha256": digest(source), "bytes": source.stat().st_size})
        shutil.copyfile(fixtures / "policy.yaml", root / "policy.yaml")
        process = subprocess.run([str(binary), "check", *[s["path"] for s in sources], "--config", "policy.yaml",
                                  "--no-gate", "--include-source", "--report", "json:result.json"],
                                 cwd=root, env={**os.environ, "GOMAXPROCS": "2"}, timeout=30, capture_output=True, check=False)
        require(process.returncode == 0, process.stderr.decode(errors="replace"))
        report = read(root / "result.json")
        require(report["status"] == "complete" and not report["errors"], "Incomplete probe")
        expected = {source["path"]: source["sha256"] for source in sources}
        require({d["name"]: d["source_hash"] for d in report["documents"]} == expected, "Probe sources differ")
        return {"binary_sha256": digest(binary), "sources": sources,
                "policy_sha256": digest(fixtures / "policy.yaml"), "result": report}


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--old", type=Path, required=True)
    parser.add_argument("--new", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    options = parser.parse_args()
    directory = Path(__file__).resolve().parent / "2026-09-13-replay/probes"
    write(options.output, {"version": "unswell-review-probes-v1",
                           "purpose": "Observed limitations and eligibility changes, not desired quality labels.",
                           "before": probe(options.old.resolve(), directory), "after": probe(options.new.resolve(), directory)})
