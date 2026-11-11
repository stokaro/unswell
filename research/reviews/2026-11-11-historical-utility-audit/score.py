#!/usr/bin/env python3
"""Recompute the exposed utility audit without changing frozen event labels."""

from __future__ import annotations

from collections import Counter
import hashlib
import json
from pathlib import Path
import sys


ROOT = Path(__file__).resolve().parents[3]
SCREEN = ROOT / "research/reviews/2026-11-11-historical-editorial-pr-screen"
HERE = Path(__file__).resolve().parent
INPUTS = {
    "ledger": (SCREEN / "baseline-ledger.json", "3789ba9320783465dd76675e38d4d5dca84d559947015f248b5a1d1ac0fa9dcc"),
    "labels": (SCREEN / "source_only_labels.json", "4e3db32155e01ba354f8b0f44e0b3e1b57029804bc523c0610410837d7654a61"),
    "sources": (SCREEN / "sources.json", "714e9b35489fb14e6df166acdd07cd15f9d5adc997be01fb232f9010b37cf287"),
}
STATUSES = {"accepted", "optional", "rejected", "uncertain"}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def sha(raw: bytes) -> str:
    return hashlib.sha256(raw).hexdigest()


def load_inputs() -> tuple[dict, list, list, dict, str]:
    loaded = {}
    for name, (path, digest) in INPUTS.items():
        raw = path.read_bytes()
        require(sha(raw) == digest, f"changed frozen {name}")
        loaded[name] = json.loads(raw)
    raw = (HERE / "review.json").read_bytes()
    return loaded["ledger"], loaded["labels"], loaded["sources"], json.loads(raw), sha(raw)


def key(pr: str, start: int) -> tuple[str, int]:
    return pr, start


def overlaps(left: dict, right: dict) -> bool:
    return left["start"] < right["end"] and right["start"] < left["end"]


def score(ledger: dict, labels: list, sources: list, review: dict, review_sha: str) -> dict:
    require(review["format"] == "unswell-349-historical-utility-review-v1", "review format")
    require("Codex assistant" in review["reviewer"] and "not a human" in review["reviewer"], "reviewer identity")
    require(ledger["profile"] == "technical" and len(ledger["rows"]) == 7, "eligible page inventory")
    source_by_pr = {item["repo"] + "#" + item["pr"]: item for item in sources}
    label_by_pr = {item["pr"]: item for item in labels}
    require(len(source_by_pr) == 13 and len(label_by_pr) == 11, "source or label inventory")
    findings = {}
    for row in ledger["rows"]:
        pr = row["pr"]
        require(pr in source_by_pr and pr in label_by_pr, f"unlabeled page: {pr}")
        require(row["source_sha256"] == source_by_pr[pr]["sha256"], f"source mismatch: {pr}")
        require(row["status"] == "complete" and row["error_count"] == 0 and row["exit_code"] == 0,
                f"incomplete report: {pr}")
        require(row["finding_count"] == len(row["findings"]), f"finding count: {pr}")
        for finding in row["findings"]:
            span = finding["span"]
            require(0 <= span["start"] < span["end"] <= source_by_pr[pr]["bytes"], "invalid span")
            identity = key(pr, span["start"])
            require(identity not in findings, f"duplicate finding: {identity}")
            findings[identity] = finding
    require(len(findings) == 12, "changed finding denominator")

    dispositions = {}
    for item in review["findings"]:
        identity = key(item["pr"], item["start"])
        require(identity in findings and identity not in dispositions, f"unknown or duplicate review: {identity}")
        require(item["status"] in STATUSES and item["reason"].strip(), f"invalid disposition: {identity}")
        dispositions[identity] = item
    require(set(dispositions) == set(findings), "every delivered finding needs a review")

    positives = {pr: item for pr, item in label_by_pr.items()
                 if pr in {row["pr"] for row in ledger["rows"]} and item["decision"] == "necessary"}
    require(len(positives) == 4, "changed eligible positive denominator")
    overlap_pages = set()
    for pr, label in positives.items():
        require(label["source_sha256"] == source_by_pr[pr]["sha256"], f"label source mismatch: {pr}")
        if any(overlaps(finding["span"], label["span"])
               for (finding_pr, _), finding in findings.items() if finding_pr == pr):
            overlap_pages.add(pr)

    matches = [key(item["pr"], item["start"]) for item in review["full_event_matches"]]
    require(len(matches) == len(set(matches)), "duplicate event match")
    for identity in matches:
        pr = identity[0]
        require(identity in findings and pr in positives, f"unknown event match: {identity}")
        require(overlaps(findings[identity]["span"], positives[pr]["span"]), f"unbound event match: {identity}")
        require(dispositions[identity]["status"] == "accepted", f"nonaccepted event match: {identity}")
    statuses = Counter(item["status"] for item in dispositions.values())
    accepted_pages = {pr for (pr, _), item in dispositions.items() if item["status"] == "accepted"}
    return {
        "format": "unswell-349-historical-utility-score-v1",
        "scope": "exposed_historical_source_only",
        "input_sha256": {name: digest for name, (_, digest) in INPUTS.items()},
        "review_sha256": review_sha,
        "complete_pages": len(ledger["rows"]),
        "operational_errors": 0,
        "delivered_findings": len(findings),
        "dispositions": {status: statuses[status] for status in sorted(STATUSES)},
        "accepted_fraction": statuses["accepted"] / len(findings),
        "pages_with_accepted_diagnosis": len(accepted_pages),
        "frozen_positive_pages": len(positives),
        "positive_pages_with_span_overlap": len(overlap_pages),
        "positive_pages_with_full_event_match": len({pr for pr, _ in matches}),
        "suggestion_safety": "not_assessed; no replacement edits supplied",
        "product_qualified": False,
    }


def main() -> None:
    ledger, labels, sources, review, review_sha = load_inputs()
    result = score(ledger, labels, sources, review, review_sha)
    rendered = json.dumps(result, indent=2, sort_keys=True) + "\n"
    if sys.argv[1:] == ["--check"]:
        require(rendered == (HERE / "result.json").read_text(encoding="utf-8"), "changed scored result")
        print("historical utility audit verified")
    elif len(sys.argv) == 1:
        print(rendered, end="")
    else:
        raise SystemExit("usage: score.py [--check]")


if __name__ == "__main__":
    main()
