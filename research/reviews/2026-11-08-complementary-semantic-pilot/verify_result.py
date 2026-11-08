"""Verify the frozen packet, tool-free traces, review, and result."""

import json
from pathlib import Path
import subprocess
import sys

import run
import score

ROOT = Path(__file__).resolve().parent
PAGES = ("caddy", "restic", "gitea", "otelcol")
EXPECTED_SHA256 = {
    "packet/freeze.json": "018d92918fbc6e2513e096aea880cf633dcf8a2c2020a1b6876b0108696fdd93",
    "run/run.json": "1d1b8387d10767a401ecf7bac8e4e744849e23578af72b7b7260f7594404ef98",
    "review.json": "2455f388bb1b89712c62381d1f70ed32a60ba66ed9bc02ec98fd9d0a6ce605f5",
    "result.json": "0d916718afcaf41b8a27eecd8325747f1177ce18401e217bf403042dd1439b15",
}


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def require(condition, message):
    if not condition:
        raise ValueError(message)


def verify():
    for name, expected in EXPECTED_SHA256.items():
        require(run.file_sha(ROOT / name) == expected, "pinned artifact changed: " + name)
    protocol = run.verify()
    require(protocol["calls_max"] == 4 and protocol["concurrency"] == 1 and
            protocol["attempts_per_request"] == 1 and
            protocol["automatic_retries"] is False and
            protocol["tools_allowed"] is False, "execution policy changed")
    control = ROOT.parent / "2026-11-08-source-function-pilot"
    checked = subprocess.run([sys.executable, "-B", str(control / "verify_result.py")],
                             cwd=control, capture_output=True, text=True, check=False)
    require(checked.returncode == 0, "prior control verifier failed")
    journal = read(ROOT / "run" / "run.json")
    require(journal["complete"] is True and journal["status"] == "complete" and
            journal["model"] == protocol["model"] and
            journal["effort"] == protocol["effort"] and
            journal["freeze_sha256"] == run.file_sha(ROOT / "packet" / "freeze.json") and
            [row["page"] for row in journal["calls"]] == list(PAGES) and
            journal["unattempted"] == [], "run journal changed")
    total = 0
    for row in journal["calls"]:
        page = row["page"]
        target = ROOT / "run" / page
        require(row["status"] == "valid" and row["exit_code"] == 0 and
                row["request_sha256"] == run.file_sha(ROOT / "packet" / "requests" / f"{page}.txt") and
                row["bound_sha256"] == run.file_sha(target / "bound.json"),
                "call record changed")
        response = read(target / "response.json")
        run.trace(target, response)
        require(read(target / "bound.json") == run.bind(page, response),
                "source binding changed")
        require(row["findings"] == len(response["findings"]), "finding count changed")
        total += row["findings"]
    require(total == 16, "new delivered denominator changed")
    review = read(ROOT / "review.json")
    require(review["format"] == "unswell-349-complementary-semantic-complete-stream-review-v1" and
            "Codex assistant" in review["reviewer"] and len(review["findings"]) == total and
            review["protected_counterexamples_flagged"] is False,
            "complete-stream review changed")
    result = read(ROOT / "result.json")
    require(result == score.score() and result["events"] == 8 and
            result["full"] == 5 and result["new_accepted"] == 12 and
            result["combined_accepted"] == 28 and result["combined_delivered"] == 33 and
            result["advance"] is False and result["qualification"] is False,
            "result arithmetic or conclusion changed")
    return {"packet_sha256": EXPECTED_SHA256["packet/freeze.json"],
            "run_sha256": EXPECTED_SHA256["run/run.json"],
            "events": result["events"], "full": result["full"],
            "accepted": result["combined_accepted"],
            "delivered": result["combined_delivered"],
            "advance": result["advance"]}


if __name__ == "__main__":
    print(json.dumps(verify(), sort_keys=True))
