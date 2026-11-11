#!/usr/bin/env python3
"""Score a frozen MELD packet in calibration, test, and response order."""

from __future__ import annotations

import argparse
import hashlib
import importlib.util
import json
import math
import os
from pathlib import Path
import sys
import time

import torch


HERE = Path(__file__).resolve().parent
FREEZE = HERE / "freeze.json"


def sha(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def require(ok: bool, message: str) -> None:
    if not ok:
        raise ValueError(message)


def write_new(path: Path, value: object) -> None:
    with path.open("x") as stream:
        json.dump(value, stream, sort_keys=True, indent=2, allow_nan=False)
        stream.write("\n")
        stream.flush()
        os.fsync(stream.fileno())


def load_upstream(path: Path):
    spec = importlib.util.spec_from_file_location("pinned_meld", path)
    require(spec is not None and spec.loader is not None, "Pinned scorer missing")
    loaded = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(loaded)
    return loaded


def verify(freeze: dict, packet: Path, model: Path) -> list[dict]:
    require(freeze["format"] == "unswell-meld-freeze-v1", "Unknown freeze")
    require(sha(packet) == freeze["packet_sha256"], "Packet changed")
    require(sha(HERE / "input-manifest.json") == freeze["manifest_sha256"],
            "Manifest changed")
    for name, expected in freeze["code_sha256"].items():
        require(sha(HERE / name) == expected, f"Runner input changed: {name}")
    for name, expected in freeze["model_sha256"].items():
        require(sha(model / name) == expected, f"Model input changed: {name}")
    require(torch.__version__ == freeze["environment"]["torch"], "Torch changed")
    import transformers
    import safetensors
    require(transformers.__version__ == freeze["environment"]["transformers"],
            "Transformers changed")
    require(safetensors.__version__ == freeze["environment"]["safetensors"],
            "Safetensors changed")
    require(torch.backends.mps.is_available(), "Apple MPS unavailable")
    records = [json.loads(line) for line in packet.read_text().splitlines()]
    manifest = json.loads((HERE / "input-manifest.json").read_text())
    require(len(records) == len(manifest["rows"]) == 1255, "Packet count changed")
    for row, public in zip(records, manifest["rows"], strict=True):
        require({key: value for key, value in row.items() if key != "text"} == public,
                "Packet metadata changed")
        require(hashlib.sha256(row["text"].encode()).hexdigest() == row["prose_sha256"],
                "Packet text changed")
    counts = {"calibration": 0, "test": 0, "matched_source": 0,
              "response": 0, "generate": 0, "polish": 0}
    for row in records:
        if row["eligible"]:
            if row["group"] == "control":
                counts[row["arm"]] += 1
            elif row["group"] == "matched_source":
                counts["matched_source"] += 1
            elif row["group"] == "response":
                counts["response"] += 1
                counts[row["operation"]] += 1
    require(counts == freeze["eligible_counts"], "Eligible counts changed")
    return records


def score_group(scorer, records: list[dict], output: Path, started: float,
                max_seconds: int) -> list[dict]:
    values = []
    with output.open("x") as stream:
        for index, row in enumerate(records):
            require(time.monotonic() - started < max_seconds, "Time budget exhausted")
            result = scorer.score(row["text"], batch_size=1)
            value = result["score"]
            require(math.isfinite(value), "Nonfinite model score")
            item = {key: value for key, value in row.items() if key != "text"}
            item.update(score=value, n_tokens=result["n_tokens"],
                        truncated=result["truncated"], family=result["family"],
                        task=result["task"])
            values.append(item)
            stream.write(json.dumps(item, ensure_ascii=False, sort_keys=True,
                                    allow_nan=False) + "\n")
            stream.flush()
            os.fsync(stream.fileno())
            if (index + 1) % 25 == 0 or index + 1 == len(records):
                print(f"{output.name}: {index + 1}/{len(records)}", flush=True)
    return values


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--packet", type=Path, required=True)
    parser.add_argument("--model", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    args.output.mkdir(parents=True, exist_ok=False)
    started = time.monotonic()
    phase = "preflight"
    try:
        freeze = json.loads(FREEZE.read_text())
        records = verify(freeze, args.packet, args.model)
        os.environ["HF_HUB_OFFLINE"] = "1"
        os.environ["TRANSFORMERS_OFFLINE"] = "1"
        torch.set_num_threads(4)
        upstream = load_upstream(args.model / "meld.py")
        scorer = upstream.Scorer(str(args.model), device="mps", canonical=True)
        groups = {
            "calibration": [row for row in records if row["eligible"]
                            and row["group"] == "control" and row["arm"] == "calibration"],
            "test": [row for row in records if row["eligible"]
                     and row["group"] == "control" and row["arm"] == "test"],
            "matched_source": [row for row in records if row["eligible"]
                               and row["group"] == "matched_source"],
            "response": [row for row in records if row["eligible"]
                         and row["group"] == "response"],
        }
        for name in groups:
            require(len(groups[name]) == freeze["eligible_counts"].get(name, len(groups[name])),
                    f"Group count changed: {name}")
        phase = "calibration"
        calibration = score_group(scorer, groups["calibration"],
                                  args.output / "calibration.jsonl", started,
                                  freeze["budget"]["max_seconds"])
        scores = sorted(row["score"] for row in calibration)
        allowed = math.floor(0.01 * len(scores))
        threshold = scores[-(allowed + 1)]
        write_new(args.output / "threshold.json", {
            "format": "unswell-meld-threshold-v1",
            "calibration_count": len(scores), "allowed_calibration_flags": allowed,
            "threshold": threshold, "comparison": "strictly_greater",
            "calibration_sha256": sha(args.output / "calibration.jsonl"),
            "freeze_sha256": sha(FREEZE),
        })
        for phase in ("test", "matched_source", "response"):
            score_group(scorer, groups[phase], args.output / f"{phase}.jsonl",
                        started, freeze["budget"]["max_seconds"])
        write_new(args.output / "complete.json", {
            "format": "unswell-meld-complete-v1", "freeze_sha256": sha(FREEZE),
            "elapsed_seconds": round(time.monotonic() - started, 3),
            "score_sha256": {name: sha(args.output / f"{name}.jsonl")
                             for name in groups},
            "threshold_sha256": sha(args.output / "threshold.json"),
        })
    except Exception as error:
        write_new(args.output / "error.json", {
            "phase": phase, "error_type": type(error).__name__, "error": str(error),
            "elapsed_seconds": round(time.monotonic() - started, 3),
        })
        raise


if __name__ == "__main__":
    sys.exit(main())
