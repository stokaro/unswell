#!/usr/bin/env python3
"""Audit frozen #349 event reachability through current source inventory."""

from __future__ import annotations

import argparse
from bisect import bisect_right
from collections import Counter
import hashlib
import json
import math
from pathlib import Path
import subprocess


HERE = Path(__file__).resolve().parent
ARCHIVE_NAME = "joint-component-sol-high-evaluation/originals.json"
ARCHIVE_SHA256 = "4f0051badc4b16dfe088d88cae69bd0dda89af2e9f62b4bed317309c4db65082"
COHORTS = {"whole": 57, "context": 46, "confirmation": 20}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def source_content_positions(raw: bytes, start: int, end: int) -> list[int]:
    positions = []
    offset = start
    for character in raw[start:end].decode("utf-8"):
        width = len(character.encode("utf-8"))
        if character.isalnum():
            positions.extend(range(offset, offset + width))
        offset += width
    require(offset == end and positions, "target lacks aligned content bytes")
    return positions


def merged_spans(spans: list[tuple[int, int]]) -> list[tuple[int, int]]:
    merged = []
    for start, end in sorted(spans):
        require(0 <= start < end, "invalid source segment")
        if merged and start <= merged[-1][1]:
            merged[-1] = (merged[-1][0], max(merged[-1][1], end))
        else:
            merged.append((start, end))
    return merged


def contains(spans: list[tuple[int, int]], starts: list[int], position: int) -> bool:
    index = bisect_right(starts, position) - 1
    return index >= 0 and position < spans[index][1]


def validate_archive(data: dict) -> tuple[dict, dict, dict]:
    originals = data["original"]
    paths = data["paths"]
    events = data["reference"]["events"]
    require(len(originals) == len(paths) == 36 and
            len(data["groups"]) == 13 and len(events) == 123,
            "frozen source/event inventory changed")
    require(set(originals) == set(paths), "source path inventory changed")
    by_path = {}
    for page, record in originals.items():
        doc = record["document"]
        raw = doc["source"].encode("utf-8")
        require(doc["page"] == page and digest_bytes(raw) == doc["source_sha256"],
                "original source identity changed")
        require(paths[page] not in by_path, "duplicate source path")
        by_path[paths[page]] = page
    require(Counter(key.split("/", 1)[0] for key in events) == COHORTS,
            "event denominators changed")
    for event_id, event in events.items():
        require(event["path"] in by_path and event_id.startswith(
            event["path"].split("/", 1)[0] + "/"), "event source mismatch")
        raw = originals[by_path[event["path"]]]["document"]["source"].encode("utf-8")
        require(event["targets"], "event has no targets")
        for target in event["targets"]:
            start, end = target["start"], target["end"]
            require(0 <= start < end <= len(raw) and
                    raw[start:end].decode("utf-8") == target["quote"],
                    "frozen target quote/span mismatch")
            source_content_positions(raw, start, end)
    return originals, by_path, events


def digest_bytes(raw: bytes) -> str:
    return hashlib.sha256(raw).hexdigest()


def export_page(exporter: Path, page: str, doc: dict) -> tuple[dict | None, str | None]:
    raw = doc["source"].encode("utf-8")
    request = json.dumps({"version": 1, "basis": "source-function-inventory",
                          "pages": [{"id": page + ".md", "format": doc["format"],
                                     "sha256": doc["source_sha256"],
                                     "text": doc["source"]}]},
                         sort_keys=True, separators=(",", ":"),
                         ensure_ascii=False).encode("utf-8")
    try:
        run = subprocess.run([str(exporter), "--source-inventory"],
                             input=request, capture_output=True, timeout=30,
                             check=False)
    except subprocess.TimeoutExpired:
        return None, "timeout"
    if run.returncode != 0:
        return None, "exporter_error"
    try:
        out = json.loads(run.stdout)
    except (UnicodeDecodeError, json.JSONDecodeError):
        return None, "invalid_json"
    require(out["version"] == 1 and out["basis"] == "source-function-inventory" and
            out["input_sha256"] == digest_bytes(request) and len(out["pages"]) == 1,
            "exporter input contract changed")
    result = out["pages"][0]
    require(result["id"] == page + ".md" and result["sha256"] == doc["source_sha256"]
            and result["bytes"] == len(raw), "exported page identity changed")
    excluded = []
    for item in result["excluded"]:
        span = item["span"]
        excluded.append((span["start"], span["end"]))
    eligible = []
    for ordinal, unit in enumerate(result["units"]):
        require(unit["ordinal"] == ordinal, "unit ordinal changed")
        bound = unit["raw_span"]
        start, end = bound["start"], bound["end"]
        require(0 <= start < end <= len(raw) and
                raw[start:end].decode("utf-8") == unit["raw_slice"],
                "raw source slice changed")
        for item in unit["binding"]["segments"]:
            start, end = item["start"], item["end"]
            require(0 <= start < end <= len(raw) and
                    all(end <= left or start >= right for left, right in excluded),
                    "eligible segment intersects exclusion")
            eligible.append((start, end))
    return {"eligible": merged_spans(eligible),
            "excluded": merged_spans(excluded),
            "units": len(result["units"]), "source_sha256": result["sha256"]}, None


def event_coverage(event: dict, raw: bytes, inventory: dict | None) -> dict:
    if inventory is None:
        return {"available": False, "targets": len(event["targets"])}
    eligible = inventory["eligible"]
    excluded = inventory["excluded"]
    starts = [start for start, _ in eligible]
    excluded_starts = [start for start, _ in excluded]
    overlaps = []
    full = []
    protected = False
    for target in event["targets"]:
        positions = source_content_positions(raw, target["start"], target["end"])
        count = sum(contains(eligible, starts, position) for position in positions)
        overlaps.append(count > 0)
        full.append(count == len(positions))
        protected |= any(contains(excluded, excluded_starts, position)
                         for position in positions)
    return {"available": True, "targets": len(overlaps),
            "any_target": any(overlaps), "all_targets_some": all(overlaps),
            "all_targets_full": all(full), "zero_target": not any(overlaps),
            "protected_target": protected}


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--archive-root", type=Path, required=True)
    parser.add_argument("--exporter", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--private-detail", type=Path, required=True)
    args = parser.parse_args()
    require(not args.output.exists() and not args.private_detail.exists(),
            "output already exists")
    freeze = json.loads((HERE / "freeze.json").read_text())
    require(freeze["status"] == "frozen_before_audit" and
            freeze["protocol_sha256"] == digest(HERE / "PROTOCOL.md") and
            freeze["script_sha256"] == digest(Path(__file__)) and
            freeze["archive_sha256"] == ARCHIVE_SHA256 and
            freeze["exporter_sha256"] == digest(args.exporter),
            "frozen input identity changed")
    archive = args.archive_root / ARCHIVE_NAME
    require(digest(archive) == ARCHIVE_SHA256, "archived source changed")
    originals, by_path, events = validate_archive(json.loads(archive.read_text()))
    inventories, errors = {}, {}
    for page in sorted(originals):
        inventory, error = export_page(args.exporter, page,
                                       originals[page]["document"])
        inventories[page] = inventory
        if error:
            errors[page] = error
    details = {}
    for event_id, event in events.items():
        page = by_path[event["path"]]
        raw = originals[page]["document"]["source"].encode("utf-8")
        details[event_id] = event_coverage(event, raw, inventories[page])
    summaries = {}
    for cohort, denominator in COHORTS.items():
        cohort_events = [value for key, value in details.items()
                         if key.startswith(cohort + "/")]
        require(len(cohort_events) == denominator, "cohort denominator changed")
        summaries[cohort] = {"events": denominator,
                             "required_full": math.ceil(denominator * 0.8),
                             "available": sum(item["available"] for item in cohort_events),
                             "any_target": sum(item.get("any_target", False)
                                               for item in cohort_events),
                             "all_targets_some": sum(item.get("all_targets_some", False)
                                                     for item in cohort_events),
                             "all_targets_full": sum(item.get("all_targets_full", False)
                                                     for item in cohort_events),
                             "zero_target": sum(item.get("zero_target", False)
                                                for item in cohort_events),
                             "protected_target": sum(item.get("protected_target", False)
                                                     for item in cohort_events)}
    public = {"format": "unswell-349-event-reachability-audit-v1",
              "scope": "exposed_offline_source_availability_only",
              "main_commit": freeze["main_commit"],
              "main_tree": freeze["main_tree"],
              "archive_sha256": ARCHIVE_SHA256,
              "exporter_sha256": freeze["exporter_sha256"],
              "protocol_sha256": freeze["protocol_sha256"],
              "script_sha256": freeze["script_sha256"],
              "pages": 36, "available_pages": 36 - len(errors),
              "page_error_categories": dict(Counter(errors.values())),
              "exported_units": sum(value["units"] for value in inventories.values()
                                    if value is not None),
              "cohorts": summaries, "qualified": False}
    private = {"format": "unswell-349-event-reachability-private-v1",
               "page_errors": errors, "events": details,
               "page_source_sha256": {page: value["source_sha256"]
                                      for page, value in inventories.items()
                                      if value is not None}}
    args.output.write_text(json.dumps(public, indent=2, sort_keys=True) + "\n")
    args.private_detail.write_text(json.dumps(private, indent=2, sort_keys=True) + "\n")
    print(json.dumps({"available_pages": public["available_pages"],
                      "exported_units": public["exported_units"],
                      "cohorts": summaries}, sort_keys=True))


if __name__ == "__main__":
    main()
