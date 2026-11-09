#!/usr/bin/env python3
"""Run the pinned upstream RACE evaluator on the parsed transfer input."""

from __future__ import annotations

import argparse
from datetime import datetime, timezone
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import time


CHECKPOINT_SHA256 = "ae9ba214ad47fd6834f82a951ad0b7a54f0903e171fa28b7f7608d0dea969b60"
CONFIG_SHA256 = "c61f2dc1fdaac5654bd10f6c66ed3e5b39fe79b11122288ecb34e46bfa20981b"


def digest(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            h.update(block)
    return h.hexdigest()


def require_hash(path: Path, expected: str) -> None:
    if digest(path) != expected:
        raise ValueError(f"SHA-256 mismatch: {path.name}")


def run(race_dir: Path, checkpoint: Path, config: Path,
        parsed: Path, parse_status: Path, output_dir: Path) -> dict:
    require_hash(checkpoint, CHECKPOINT_SHA256)
    require_hash(config, CONFIG_SHA256)
    status = json.loads(parse_status.read_text(encoding="utf-8"))
    require_hash(parsed, status["parsed_sha256"])
    if len(status["status"]) != 300:
        raise ValueError("Incomplete parser status")
    if not (race_dir / "test.py").is_file():
        raise ValueError("Pinned RACE evaluator missing")
    output_dir.mkdir(parents=True, exist_ok=True)
    started = time.monotonic()
    journal = {"started_utc": datetime.now(timezone.utc).isoformat(),
               "checkpoint_sha256": CHECKPOINT_SHA256,
               "config_sha256": CONFIG_SHA256,
               "parsed_sha256": status["parsed_sha256"],
               "upstream_commit": "61ce528430a8bb84fd1bfb3285ee34827182ef59"}
    env = os.environ.copy()
    env.update({"HF_HUB_OFFLINE": "1", "TRANSFORMERS_OFFLINE": "1",
                "OMP_NUM_THREADS": "2", "TOKENIZERS_PARALLELISM": "false"})
    try:
        result = subprocess.run(
            [sys.executable, str(race_dir / "test.py"), "--config", str(config),
             "--checkpoint_path", str(checkpoint), "--num_workers", "0"],
            cwd=race_dir, env=env, capture_output=True, text=True,
            timeout=2 * 3600, check=False,
        )
        (output_dir / "infer-stdout.txt").write_text(result.stdout, encoding="utf-8")
        (output_dir / "infer-stderr.txt").write_text(result.stderr, encoding="utf-8")
        journal["exit_code"] = result.returncode
    except subprocess.TimeoutExpired as error:
        journal["status"] = "timeout"
        journal["error_type"] = type(error).__name__
    journal["elapsed_seconds"] = round(time.monotonic() - started, 3)
    predictions = checkpoint.parent / "test_predictions.json"
    if journal.get("exit_code") == 0 and predictions.is_file():
        journal["status"] = "complete"
        journal["predictions_sha256"] = digest(predictions)
    elif "status" not in journal:
        journal["status"] = "failed"
    (output_dir / "infer-journal.json").write_text(
        json.dumps(journal, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    if journal["status"] != "complete":
        raise RuntimeError(f"RACE inference {journal['status']}; no retry")
    return journal


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--race", type=Path, required=True)
    parser.add_argument("--checkpoint", type=Path, required=True)
    parser.add_argument("--config", type=Path, required=True)
    parser.add_argument("--parsed", type=Path, required=True)
    parser.add_argument("--status", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    print(json.dumps(run(args.race, args.checkpoint, args.config,
                         args.parsed, args.status, args.output)))


if __name__ == "__main__":
    main()
