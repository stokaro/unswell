"""Verify the frozen #349 source-function packet, run, review, and result."""

import hashlib
import json
from pathlib import Path
import sys

import run
import score

PAGES = ("caddy", "restic", "gitea", "otelcol")
EXPECTED_SHA256 = {
    "sources.json": "2c1391ab588c6bb9df60ea9d216aea822e2ee3bc9bec0c0bc7293b685b0228e3",
    "packet/freeze.json": "0cec9103172cf084dc277fb4110a92ae2764c5931823599eb21026fd97174a8f",
    "run/run.json": "fa23ee46d207c9ac949797eebe71615626b86c56cf3f0cb4ff8ec02206a3da6b",
    "review.json": "cdc62695abfa0d3f38693e04709be6c6fb3a7d1ffdf598834057eb503a9e471f",
    "result.json": "474e54259847b6abbd73c6740085c8766011175e90270b0d01ea43d55f6b3417",
    "baseline_clean/caddy.json": "31c3220521f4f75b72715c208355e66e4a9ce141ca618086b45235abe4da904d",
    "baseline_clean/restic.json": "e8303cebd8c86205c60c82a84330c49fbaa73713345d3fec43afbeb471267061",
    "baseline_clean/gitea.json": "543f8e5d788dad34dcd13a720674eafb95d87540ee39d17d475766567ecb1111",
    "baseline_clean/otelcol.json": "5cfff2656bda600c49072c181087c358264356cdefd5151fb9a84479b1ad75ed",
}


def sha(data):
    return hashlib.sha256(data).hexdigest()


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def require(condition, reason):
    if not condition:
        raise ValueError(reason)


def verify(root=Path(__file__).resolve().parent):
    root = Path(root)
    for name, expected in EXPECTED_SHA256.items():
        require(sha((root / name).read_bytes()) == expected,
                "pinned artifact changed: " + name)
    run.ROOT = root
    run.PACKET = root / "packet"
    score.ROOT = root
    protocol = run.verify()
    require(protocol["tools_allowed"] is False and
            protocol["automatic_retries"] is False and
            protocol["attempts_per_request"] == 1 and
            protocol["concurrency"] == 1, "execution policy changed")
    sources = read(root / "sources.json")
    require(sources["format"] == "unswell-349-source-function-public-sources-v1" and
            [x["page"] for x in sources["sources"]] == list(PAGES),
            "source provenance changed")
    for item in sources["sources"]:
        page = item["page"]
        raw = (root / "sources" / page / "README.md").read_bytes()
        git_blob = hashlib.sha1(b"blob " + str(len(raw)).encode() + b"\0" + raw).hexdigest()
        require(item["sha256"] == sha(raw) and item["git_blob_sha1"] == git_blob and
                item["path"] == "README.md" and len(item["commit"]) == 40,
                "source identity changed")
    journal = read(root / "run" / "run.json")
    require(journal["complete"] is True and journal["status"] == "complete" and
            journal["model"] == protocol["model"] and
            journal["effort"] == protocol["effort"] and
            journal["freeze_sha256"] == sha((root / "packet" / "freeze.json").read_bytes()) and
            [x["page"] for x in journal["calls"]] == list(PAGES) and
            journal["unattempted"] == [], "run journal changed")
    total_findings = 0
    for row in journal["calls"]:
        page = row["page"]
        target = root / "run" / page
        require(row["status"] == "valid" and row["exit_code"] == 0 and
                row["request_sha256"] == sha((root / "packet" / "requests" / f"{page}.txt").read_bytes()) and
                row["bound_sha256"] == sha((target / "bound.json").read_bytes()),
                "call record changed")
        response = read(target / "response.json")
        run.trace(target, response)
        require(read(target / "bound.json") == run.bind(page, response),
                "bound findings changed")
        require(row["findings"] == len(response["findings"]),
                "finding count changed")
        total_findings += row["findings"]
    require(total_findings == 17, "delivered denominator changed")
    review = read(root / "review.json")
    require(review["format"] == "unswell-349-source-function-complete-stream-review-v1" and
            "Codex assistant" in review["reviewer"] and
            len(review["findings"]) == 17 and
            review["protected_counterexamples_flagged"] is False,
            "full-stream review changed")
    labels = read(root / "labels.json")
    ids = {x["id"] for x in labels["events"]}
    partitions = [set(review[x]) for x in
                  ("recalled_event_ids", "partial_event_ids", "missed_event_ids")]
    require(len(ids) == 8 and set.union(*partitions) == ids and
            sum(len(x) for x in partitions) == 8,
            "event partition changed")
    result = read(root / "result.json")
    require(result == score.score() and result["recalled"] == 3 and
            result["accepted"] == 16 and result["delivered"] == 17 and
            result["advance"] is False and result["supplied_edits"] == 0,
            "result arithmetic or conclusion changed")
    baseline = root / "baseline_clean"
    require({p.stem for p in baseline.glob("*.json")} == set(PAGES),
            "baseline inventory changed")
    expected_counts = {"caddy": 4, "restic": 0, "gitea": 2, "otelcol": 2}
    for item in sources["sources"]:
        page = item["page"]
        report = read(baseline / f"{page}.json")
        require(report["status"] == "complete" and report["errors"] == [] and
                len(report["findings"]) == expected_counts[page] and
                len(report["documents"]) == 1 and
                report["documents"][0]["source_hash"] == item["sha256"],
                "CLI baseline changed")
    return {"packet_sha256": sha((root / "packet" / "freeze.json").read_bytes()),
            "run_sha256": sha((root / "run" / "run.json").read_bytes()),
            "result_sha256": sha((root / "result.json").read_bytes()),
            "events": len(ids), "recalled": result["recalled"],
            "delivered": result["delivered"], "accepted": result["accepted"],
            "advance": result["advance"]}


if __name__ == "__main__":
    print(json.dumps(verify(Path(sys.argv[1]) if len(sys.argv) > 1 else
                            Path(__file__).resolve().parent), sort_keys=True))
