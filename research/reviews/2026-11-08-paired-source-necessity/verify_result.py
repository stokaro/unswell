"""Independently verify the paired pilot packet, raw trace set, and score."""

import json
from pathlib import Path
import sys

import run
import score

ROOT = Path(__file__).resolve().parent
FILES_PER_CALL = {"events.jsonl", "stderr.txt", "response.json", "bound.json"}


def require(condition, message):
    if not condition:
        raise ValueError(message)


def trace_hashes():
    pairs = run.selected()
    expected_names = {payload["pair_id"] for payload, _ in pairs} | {"run.json"}
    require({path.name for path in run.RUN.iterdir()} == expected_names,
            "run file set changed")
    hashes = {"run.json": run.file_sha(run.RUN / "run.json")}
    for payload, _ in pairs:
        pair_id = payload["pair_id"]
        target = run.RUN / pair_id
        require(target.is_dir() and {path.name for path in target.iterdir()} == FILES_PER_CALL,
                "call file set changed: " + pair_id)
        for name in sorted(FILES_PER_CALL):
            hashes[pair_id + "/" + name] = run.file_sha(target / name)
    return hashes


def evidence():
    run.verify()
    result = score.score()
    require(run.read(ROOT / "result.json") == result, "scored result changed")
    journal = run.read(run.RUN / "run.json")
    require(journal["complete"] is True and len(journal["calls"]) == run.MAX_CALLS,
            "incomplete run")
    require(all(row["status"] == "valid" and row["exit_code"] == 0
                for row in journal["calls"]), "failed call in complete run")
    require(journal.get("monetary_cost") is None, "invented monetary cost")
    require(result["valid_pairs"] == 8 and result["source_bound_targets"] == 16
            and result["baseline_independent_correct_pairs"] == 2
            and result["paired_correct_pairs"] == 4
            and result["prior_accepted_retained"] == 5
            and result["prior_uncertain_withheld"] == 7
            and result["advance"] is False and result["qualified"] is False,
            "negative result claim changed")
    return {"format": "unswell-349-paired-source-necessity-verification-v1",
            "packet_freeze_sha256": run.file_sha(run.PACKET / "freeze.json"),
            "result_sha256": run.file_sha(ROOT / "result.json"),
            "run_files_sha256": trace_hashes()}


def freeze():
    target = ROOT / "verification.json"
    require(not target.exists(), "verification already exists")
    record = evidence()
    target.write_bytes(run.encoded(record))
    return record


def verify():
    expected = evidence()
    require(run.read(ROOT / "verification.json") == expected,
            "verification record changed")
    return expected


if __name__ == "__main__":
    operation = sys.argv[1:] or ["verify"]
    if operation == ["freeze"]:
        result = freeze()
    elif operation == ["verify"]:
        result = verify()
    else:
        raise SystemExit("usage: verify_result.py freeze|verify")
    print(json.dumps({"format": result["format"],
                      "run_files": len(result["run_files_sha256"]),
                      "result_sha256": result["result_sha256"]}, sort_keys=True))
