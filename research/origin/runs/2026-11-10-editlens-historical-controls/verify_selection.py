#!/usr/bin/env python3
"""Verify the historical source frame against the frozen candidate audit."""

from __future__ import annotations

import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
SEED = b"unswell-techdoc-document-pilot-v1"
SOURCE_FIELDS = ("repository", "commit", "path", "blob_sha", "source_sha256",
                 "windows_100")
EXCLUSION_REASONS = {"ineligible_100_word_windows", "vendored_dependency",
                     "non_english", "generated_test_fixture", "table_without_prose",
                     "blob_or_utf8_error"}


def need(ok: bool, message: str) -> None:
    if not ok:
        raise ValueError(message)


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(name: str) -> dict:
    return json.loads((HERE / name).read_text())


def key(row: dict) -> tuple[str, str, str]:
    return row["repository"], row["commit"], row["path"]


def rank(repository: str, path: str) -> tuple[str, str]:
    value = hashlib.sha256(repository.lower().encode() + b"\0" +
                           path.encode() + b"\0" + SEED).hexdigest()
    return value, path


def verify() -> dict:
    inventory = read("inventory.json")
    blobs = read("blob_audit.json")
    eligibility = read("eligibility.json")
    frame = read("frame.json")
    need(blobs["inventory_sha256"] == sha(HERE / "inventory.json"),
         "Inventory digest changed")
    need(eligibility["blobs_sha256"] == sha(HERE / "blob_audit.json"),
         "Blob audit digest changed")
    need(frame["eligibility_sha256"] == sha(HERE / "eligibility.json"),
         "Eligibility digest changed")
    need(inventory["top_per_language"] == 50 and len(inventory["rows"]) == 196,
         "Repository frame changed")
    need(len(blobs["rows"]) == len(eligibility["rows"]) == 506,
         "Candidate count changed")
    expected = {}
    repository_dates = {}
    for repository in inventory["rows"]:
        if repository.get("error") or not repository.get("documents"):
            continue
        repository_dates[repository["repository"], repository["commit"]] = repository["committed_at"]
        ordered = sorted(repository["documents"],
                         key=lambda item: rank(repository["repository"], item["path"]))
        for item in ordered[:8]:
            identity = (repository["repository"], repository["commit"], item["path"])
            need(identity not in expected, "Duplicate inventory candidate")
            expected[identity] = {**item, "repository": repository["repository"],
                                  "commit": repository["commit"]}
    audited = {key(row): row for row in blobs["rows"]}
    eligible = {key(row): row for row in eligibility["rows"]}
    need(len(expected) == len(audited) == len(eligible) == 506 and
         set(expected) == set(audited) == set(eligible),
         "Candidate set changed")
    for identity, row in audited.items():
        for field, value in expected[identity].items():
            need(row[field] == value, f"Inventory candidate changed: {identity}")
        for field, value in row.items():
            need(eligible[identity][field] == value,
                 f"Eligibility candidate changed: {identity}")
        need(repository_dates[identity[0], identity[1]] < "2021-08-15T00:00:00Z",
             "Source too recent")
    selected = {key(row): row for row in frame["selected"]}
    excluded = {key(row): row for row in frame["excluded"]}
    need(len(selected) == 462 and len(excluded) == 44 and
         not (set(selected) & set(excluded)) and
         set(selected) | set(excluded) == set(eligible),
         "Incomplete or duplicate source decision")
    need(len({row["source_id"] for row in selected.values()}) == 462 and
         len({row["source_sha256"] for row in selected.values()}) == 462,
         "Duplicate selected source")
    for identity, row in selected.items():
        prior = eligible[identity]
        need(prior["complete_100"] is True and prior["windows_100"] >= 1 and
             prior["max_tokens_100"] <= 512, "Ineligible selected source")
        for field, value in row.items():
            need(prior[field] == value,
                 f"Selected source metadata changed: {identity}")
    for identity, row in excluded.items():
        need(row["reason"] in EXCLUSION_REASONS,
             f"Unknown exclusion reason: {identity}")
    identity = [[row[field] for field in SOURCE_FIELDS] for row in frame["selected"]]
    digest = hashlib.sha256(json.dumps(identity, ensure_ascii=False,
                                       separators=(",", ":")).encode()).hexdigest()
    need(digest == frame["selection_sha256"], "Selected identity changed")
    return {"repositories_inspected": len(inventory["rows"]),
            "candidates": len(eligible), "selected": len(selected),
            "excluded": len(excluded),
            "selected_repositories": len({row["repository"] for row in selected.values()}),
            "selected_windows": sum(row["windows_100"] for row in selected.values()),
            "selection_sha256": digest}


if __name__ == "__main__":
    print(json.dumps(verify(), sort_keys=True))
