#!/usr/bin/env python3
"""Seal score-blind source decisions and repository split for new controls."""

from __future__ import annotations

import argparse
import collections
import hashlib
import json
from pathlib import Path


INVENTORY_SHA256 = "6051df6a4af3f0aebe8394c982996c501a465999236724b5df8e56795ea3d194"
BLOB_AUDIT_SHA256 = "91651bf12c14ed9fc81b252d3b73bed523d857dbe4edef53af31693aa8f2e863"
ELIGIBILITY_SHA256 = "ded775eff6e65f56bab8edf9b2429ffb9eb93b5bd33a5e3c20353ea1d1dbdbd5"
SOURCE_REVIEW_SHA256 = "7a82c235f381d80f24d9ac819f00b76b64808aa799a70ffd9f44c7fceebf16dd"
SEED = b"unswell-techdoc-new-split-v1"


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def arm(repository: str) -> str:
    bits = hashlib.sha256(repository.lower().encode() + b"\0" + SEED).digest()
    return "calibration" if bits[0] & 1 == 0 else "test"


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--directory", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    required = {"inventory.json": INVENTORY_SHA256,
                "blob_audit.json": BLOB_AUDIT_SHA256,
                "eligibility.json": ELIGIBILITY_SHA256,
                "source_review.json": SOURCE_REVIEW_SHA256}
    for name, expected in required.items():
        if sha(args.directory / name) != expected:
            raise ValueError(f"Source audit changed: {name}")
    eligibility = json.loads((args.directory / "eligibility.json").read_text())
    review = json.loads((args.directory / "source_review.json").read_text())
    queue = json.loads((args.directory / "review_queue.json").read_text())
    if review["review_queue_sha256"] != sha(args.directory / "review_queue.json"):
        raise ValueError("Source review queue changed")
    queued = {row["source_id"] for row in queue["rows"]}
    excluded = {row["source_id"] for row in review["excluded"]}
    retained = set(review["retained_queued_source_ids"])
    if (len(excluded) != len(review["excluded"]) or excluded & retained or
            excluded | retained != queued):
        raise ValueError("Incomplete source-review accounting")
    review_rows = {row["source_id"]: row for row in review["excluded"]}
    selected = []
    discarded = []
    all_counts = collections.Counter()
    accepted = collections.Counter()
    seen_raw = set()
    seen_cleaned = set()
    for row in eligibility["rows"]:
        split = arm(row["repository"])
        all_counts[split] += 1
        sid = row["source_id"]
        if sid in review_rows:
            source_review = review_rows[sid]
            if (not row["eligible"] or row["source_sha256"] != source_review["source_sha256"] or
                    row["repository"] != source_review["repository"] or
                    row["path"] != source_review["path"]):
                raise ValueError(f"Invalid review exclusion: {sid}")
        if row["eligible"] and sid not in excluded:
            if row["source_sha256"] in seen_raw or row["cleaned_sha256"] in seen_cleaned:
                raise ValueError(f"Duplicate source in selected frame: {sid}")
            seen_raw.add(row["source_sha256"])
            seen_cleaned.add(row["cleaned_sha256"])
            selected.append({key: row[key] for key in
                             ("source_id", "repository", "commit", "committed_at", "path",
                              "blob_sha", "bytes", "source_sha256", "cleaned_sha256",
                              "language", "cleaned_words", "windows_100", "max_tokens_100")}
                            | {"arm": split})
            accepted[split] += 1
        else:
            discarded.append({"source_id": sid, "arm": split,
                              "reason": review_rows[sid]["reason"] if sid in excluded else row["reason"]})
    if len(selected) + len(discarded) != 860:
        raise ValueError("Incomplete candidate accounting")
    if accepted["calibration"] < 300 or accepted["test"] < 300:
        raise ValueError("Insufficient independent controls; no inference allowed")
    if set(row["repository"] for row in selected if row["arm"] == "calibration") & \
            set(row["repository"] for row in selected if row["arm"] == "test"):
        raise ValueError("Repository leaked between arms")
    identities = [[row[key] for key in ("source_id", "source_sha256", "cleaned_sha256", "arm")]
                  for row in selected]
    identity_sha = hashlib.sha256(json.dumps(identities, separators=(",", ":")).encode()).hexdigest()
    frame = {"format": "unswell-origin-new-historical-frame-v1",
             "status": "source_preparation_complete; model protocol not frozen",
             "input_sha256": required,
             "review_queue_sha256": sha(args.directory / "review_queue.json"),
             "split_seed": SEED.decode(),
             "identity_sha256": identity_sha,
             "candidate_counts": dict(all_counts),
             "selected_counts": dict(accepted),
             "selected": selected,
             "discarded": discarded,
             "model_scores": 0}
    args.output.write_text(json.dumps(frame, sort_keys=True, indent=2) + "\n")
    print(json.dumps({"candidates": sum(all_counts.values()),
                      "selected": dict(accepted), "discarded": len(discarded),
                      "test_coverage": accepted["test"] / all_counts["test"],
                      "identity_sha256": identity_sha}, sort_keys=True))


if __name__ == "__main__":
    main()
