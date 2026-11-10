#!/usr/bin/env python3
"""Independently verify source, exclusion, and split accounting for the frame."""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path


def need(ok: bool, message: str) -> None:
    if not ok:
        raise ValueError(message)


def digest(raw: bytes) -> str:
    return hashlib.sha256(raw).hexdigest()


def sha(path: Path) -> str:
    return digest(path.read_bytes())


def source_id(row: dict) -> str:
    return digest((row["repository"] + "\0" + row["commit"] + "\0" + row["path"]).encode())


def arm(repository: str) -> str:
    value = hashlib.sha256(repository.lower().encode() + b"\0unswell-techdoc-new-split-v1").digest()
    return "calibration" if value[0] % 2 == 0 else "test"


def document_rank(row: dict) -> tuple[str, str]:
    key = (row["repository"].lower() + "\0" + row["path"] +
           "\0unswell-techdoc-new-document-v1").encode()
    return digest(key), row["path"]


def verify(directory: Path, sources: Path, frame_path: Path) -> dict:
    inventory = json.loads((directory / "inventory.json").read_text())
    blobs = json.loads((directory / "blob_audit.json").read_text())
    eligibility = json.loads((directory / "eligibility.json").read_text())
    queue = json.loads((directory / "review_queue.json").read_text())
    review = json.loads((directory / "source_review.json").read_text())
    exclusions = json.loads((directory / "origin-new-corpus-exclusions-2026-11-10.json").read_text())
    frame = json.loads(frame_path.read_text())
    need(frame["format"] == "unswell-origin-new-historical-frame-v1", "Frame format")
    need(frame["model_scores"] == 0, "Frame is not source-only")
    for name, expected in frame["input_sha256"].items():
        need(sha(directory / name) == expected, f"Input changed: {name}")
    need(frame["review_queue_sha256"] == sha(directory / "review_queue.json") ==
         review["review_queue_sha256"], "Review queue changed")
    need(len(inventory["rows"]) == 465, "Repository candidate count")
    excluded_repos = {name.lower() for name in exclusions["excluded_repositories"]}
    old_raw = set(exclusions["historical_raw_source_sha256"])
    old_cleaned = set(exclusions["prior_source_cleaned_text_sha256"])
    expected_blobs = []
    for repo in inventory["rows"]:
        need(repo["repository"].lower() not in excluded_repos, "Previously scored repository")
        if "error" in repo:
            continue
        need(repo["committed_at"] < "2021-08-15T00:00:00Z", "Late historical commit")
        rows = [{**doc, "repository": repo["repository"], "commit": repo["commit"],
                 "language": repo["language"], "committed_at": repo["committed_at"]}
                for doc in repo["documents"]]
        expected_blobs.extend(sorted(rows, key=document_rank)[:8])
    need(len(expected_blobs) == len(blobs["rows"]) == len(eligibility["rows"]) == 860,
         "Candidate blob count")
    for expected, blob, eligible in zip(expected_blobs, blobs["rows"], eligibility["rows"], strict=True):
        need(all(blob[key] == value for key, value in expected.items()),
             f"Blob selection changed: {expected['repository']} {expected['path']}")
        sid = source_id(expected)
        need(blob["source_id"] == eligible["source_id"] == sid, "Source ID changed")
        need(all(eligible[key] == value for key, value in blob.items()),
             f"Eligibility source changed: {sid}")
        if "error" in blob:
            need(not eligible["eligible"] and eligible["reason"] == "blob_fetch_or_content_error",
                 "Failed blob admitted")
            continue
        raw = (sources / (sid + ".source")).read_bytes()
        git_blob = hashlib.sha1(f"blob {len(raw)}\0".encode() + raw).hexdigest()
        need(len(raw) == blob["bytes"] and git_blob == blob["blob_sha"] and
             digest(raw) == blob["source_sha256"], f"Blob integrity: {sid}")
        need(eligible["eligible"] or eligible["reason"] is not None,
             "Unexplained ineligibility")
        if eligible["eligible"]:
            need(eligible["source_sha256"] not in old_raw and
                 eligible["cleaned_sha256"] not in old_cleaned and
                 eligible["windows_100"] >= 1 and
                 eligible["cleaned_words"] >= 50 and
                 0 < eligible["max_tokens_100"] <= 512,
                 f"Invalid eligible document: {sid}")
    queued = {row["source_id"] for row in queue["rows"]}
    review_excluded = {row["source_id"]: row for row in review["excluded"]}
    retained = set(review["retained_queued_source_ids"])
    need(len(review_excluded) == len(review["excluded"]) and
         set(review_excluded) | retained == queued and not set(review_excluded) & retained,
         "Source review accounting")
    need(all(row["source_id"] in queued for row in review["excluded"]),
         "Unqueued source exclusion")
    expected_selected = []
    expected_discarded = []
    counts = {"calibration": 0, "test": 0}
    candidates = {"calibration": 0, "test": 0}
    for row in eligibility["rows"]:
        split = arm(row["repository"])
        candidates[split] += 1
        sid = row["source_id"]
        if row["eligible"] and sid not in review_excluded:
            expected_selected.append({key: row[key] for key in
                                      ("source_id", "repository", "commit", "committed_at",
                                       "path", "blob_sha", "bytes", "source_sha256",
                                       "cleaned_sha256", "language", "cleaned_words",
                                       "windows_100", "max_tokens_100")} | {"arm": split})
            counts[split] += 1
        else:
            reason = review_excluded[sid]["reason"] if sid in review_excluded else row["reason"]
            expected_discarded.append({"source_id": sid, "arm": split, "reason": reason})
    need(frame["selected"] == expected_selected and frame["discarded"] == expected_discarded,
         "Frame decisions changed")
    need(frame["selected_counts"] == counts and frame["candidate_counts"] == candidates,
         "Frame counts changed")
    need(counts["calibration"] >= 300 and counts["test"] >= 300, "Insufficient arm")
    need(not ({row["repository"].lower() for row in expected_selected if row["arm"] == "calibration"}
              & {row["repository"].lower() for row in expected_selected if row["arm"] == "test"}),
         "Repository split leakage")
    need(len({row["source_sha256"] for row in expected_selected}) == len(expected_selected) and
         len({row["cleaned_sha256"] for row in expected_selected}) == len(expected_selected),
         "Duplicate source text")
    identities = [[row[key] for key in ("source_id", "source_sha256", "cleaned_sha256", "arm")]
                  for row in expected_selected]
    identity_sha = digest(json.dumps(identities, separators=(",", ":")).encode())
    need(frame["identity_sha256"] == identity_sha, "Selection identity changed")
    return {"candidates": len(expected_blobs), "selected": counts,
            "excluded": len(expected_discarded),
            "source_review_exclusions": len(review_excluded),
            "test_coverage": counts["test"] / candidates["test"],
            "identity_sha256": identity_sha,
            "frame_sha256": sha(frame_path)}


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--directory", type=Path, required=True)
    parser.add_argument("--sources-dir", type=Path, required=True)
    parser.add_argument("--frame", type=Path, required=True)
    args = parser.parse_args()
    print(json.dumps(verify(args.directory, args.sources_dir, args.frame), sort_keys=True))
