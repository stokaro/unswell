#!/usr/bin/env python3
"""Revalidate the frozen paired run and reviewed result from raw artifacts."""

import argparse
import json
from pathlib import Path

from model_pilot import (ARMS, MAX_CALLS, MAX_RESPONSE_BYTES, MAX_TRACE_BYTES,
                         PER_CALL_SECONDS, TOTAL_SECONDS, check_response,
                         check_trace, encoded, file_sha, order, pages, read,
                         verify as verify_packet)
from score_model_pilot import grade

ROOT = Path(__file__).resolve().parent
RUN = ROOT / "model-run-v1"


def require(condition, message):
    if not condition:
        raise ValueError(message)


def verify(run_dir=RUN):
    packet = verify_packet()
    source_pages = pages()
    journal = read(run_dir / "run.json")
    require(journal["format"] == "unswell-349-atomic-paired-run-v1", "run format changed")
    require(journal["packet_freeze_sha256"] == file_sha(ROOT / "model-packet" / "freeze.json"),
            "packet identity changed")
    require(journal["model"] == packet["model"] and journal["effort"] == packet["effort"],
            "model identity changed")
    require(journal["complete"] is True and journal["status"] == "complete" and
            journal["unattempted"] == [] and len(journal["calls"]) == MAX_CALLS,
            "incomplete run")
    require("stop_reason" not in journal, "completed run has a stop reason")
    require({path.name for path in run_dir.iterdir() if path.is_dir()} ==
            {f"page-{number:03d}-{arm}" for number, arm in order()},
            "run directories changed")
    seconds = 0.0
    diagnoses = {arm: 0 for arm in ARMS}
    usage = {"input_tokens": 0, "output_tokens": 0}
    for row, (number, arm) in zip(journal["calls"], order()):
        name = f"page-{number:03d}-{arm}"
        target = run_dir / name
        request_path = ROOT / "model-packet" / "requests" / (name + ".txt")
        require(row["request"] == name and row["request_sha256"] == file_sha(request_path),
                f"request identity changed: {name}")
        require(row["status"] == "valid" and row["exit_code"] == 0,
                f"call failed: {name}")
        require(type(row["seconds"]) in (int, float) and
                0 < row["seconds"] <= PER_CALL_SECONDS,
                f"call duration invalid: {name}")
        seconds += row["seconds"]
        trace = target / "events.jsonl"
        stderr = target / "stderr.txt"
        response_path = target / "response.json"
        require(trace.stat().st_size <= MAX_TRACE_BYTES and
                stderr.stat().st_size <= MAX_TRACE_BYTES and
                response_path.stat().st_size <= MAX_RESPONSE_BYTES,
                f"response resource limit: {name}")
        response = read(response_path)
        require(check_trace(trace, response) == row["usage"],
                f"trace usage changed: {name}")
        bound = check_response(response, source_pages[number - 1], arm)
        require((target / "bound.json").read_bytes() == encoded(bound) and
                file_sha(target / "bound.json") == row["bound_sha256"],
                f"source binding changed: {name}")
        require(row["diagnoses"] == len(bound["diagnoses"]),
                f"diagnosis count changed: {name}")
        diagnoses[arm] += row["diagnoses"]
        for key in usage:
            require(type(row["usage"][key]) is int and row["usage"][key] >= 0,
                    f"usage changed: {name}")
            usage[key] += row["usage"][key]
    require(seconds <= TOTAL_SECONDS, "total run duration exceeded limit")
    review_path = run_dir / "review.json"
    result = grade(run_dir, read(review_path))
    result["review_sha256"] = file_sha(review_path)
    require(read(run_dir / "result.json") == result, "reviewed result changed")
    require(result["advance"] is False and result["qualified_349"] is False,
            "result status changed")
    return {"calls": len(journal["calls"]), "seconds": round(seconds, 3),
            "diagnoses": diagnoses, "usage": usage,
            "packet_sha256": journal["packet_freeze_sha256"],
            "run_sha256": file_sha(run_dir / "run.json"),
            "review_sha256": file_sha(review_path),
            "result_sha256": file_sha(run_dir / "result.json"),
            "advance": result["advance"], "qualified_349": result["qualified_349"]}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--run-dir", type=Path, default=RUN)
    args = parser.parse_args()
    print(json.dumps(verify(args.run_dir), sort_keys=True))


if __name__ == "__main__":
    main()
