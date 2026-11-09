#!/usr/bin/env python3
"""Verify saved WikiSQE transfer scores and their frozen-span ceiling offline."""

from __future__ import annotations

from collections import Counter, defaultdict
import gzip
import hashlib
import importlib.util
import json
import math
from pathlib import Path
import sys


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
REVIEWS = ROOT / "research/reviews"
sys.path.insert(0, str(REVIEWS / "2026-09-26-outcome-announcements"))
from replay import sources  # noqa: E402

label_spec = importlib.util.spec_from_file_location(
    "unswell_outcome_verify", REVIEWS / "2026-09-26-outcome-announcements/verify.py")
assert label_spec and label_spec.loader
label_module = importlib.util.module_from_spec(label_spec)
label_spec.loader.exec_module(label_module)
labels = label_module.labels

SOURCE_FILES = {
    "sentences": REVIEWS / "2026-11-07-ape-proposition-screen/scores.json.gz",
    "source_hashes": REVIEWS / "2026-11-07-ape-proposition-screen/source-hashes.json",
    "events": REVIEWS / "2026-11-07-current-main-editorial-review/events.json",
    "frontier": REVIEWS / "2026-11-08-dual-gate-feasibility/category-frontier.json",
}
DATASET_REVISION = "e6a3da39b44df3efa2c75706ccba1dc26dd17076"
DATASET_SHA256 = {
    "train": "25b2d9609cdd34d88dadcf8ab687bd05d7a6212766a9b010513ae5741d4a4bbe",
    "val": "dcfdd685a6082df6909f8fb1c03c81ac63c9a83376d262e1dc14584b2d1bc5aa",
    "test": "37d789c7ba762bb7b40b95091a3ddf29e2eb46b9702d757dac1dad44e515c72f",
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def overlapping(row: dict, target: dict) -> bool:
    return any(max(span["start"], target["start"]) <
               min(span["end"], target["end"]) for span in row["spans"])


def verify(artifact_dir: Path = HERE) -> dict:
    result = json.loads((artifact_dir / "result.json").read_text())
    require(result["format"] == "unswell-349-wikisqe-vague-screen-v1",
            "Unexpected result format")
    require(result["dataset_revision"] == DATASET_REVISION and
            result["dataset_sha256"] == DATASET_SHA256,
            "Dataset identity changed")
    require(result["protocol_sha256"] == digest(HERE / "PROTOCOL.md") and
            result["run_script_sha256"] == digest(HERE / "run.py"),
            "Frozen protocol or run script changed")
    for name, path in SOURCE_FILES.items():
        require(result["local_input_sha256"][name] == digest(path),
                f"Local input changed: {name}")
    require(result["scores_sha256"] == digest(artifact_dir / "scores.json.gz") and
            result["top_candidates_sha256"] == digest(artifact_dir / "top-candidates.json"),
            "Saved inference output changed")

    raw = json.loads(gzip.decompress(SOURCE_FILES["sentences"].read_bytes()))
    rows = raw["sentences"]
    scored = json.loads(gzip.decompress((artifact_dir / "scores.json.gz").read_bytes()))
    probabilities = scored["probabilities"]
    selected = scored["selected_indices"]
    require(len(rows) == len(probabilities) == 4139 and
            all(row["index"] == i for i, row in enumerate(rows)),
            "Source sentence or score count changed")
    require(all(isinstance(value, (int, float)) and math.isfinite(value) and
                0 <= value <= 1 for value in probabilities), "Invalid probability")
    require(len(selected) == 108 and len(set(selected)) == 108 and
            all(isinstance(i, int) and 0 <= i < len(rows) for i in selected),
            "Invalid selected indices")
    source_pages = sources()
    require(len(source_pages) == 36 and
            {row["page"] for row in rows} == set(source_pages),
            "Source page set changed")
    source_hashes = json.loads(SOURCE_FILES["source_hashes"].read_text())
    require(source_hashes == {page: hashlib.sha256(data).hexdigest()
                              for page, data in source_pages.items()},
            "Source bytes changed")
    by_page: dict[str, list[int]] = defaultdict(list)
    for i, row in enumerate(rows):
        source = source_pages[row["page"]]
        require(row["spans"] and all(0 <= span["start"] < span["end"] <= len(source)
                                     for span in row["spans"]),
                "Invalid original-byte span")
        by_page[row["page"]].append(i)
    expected_selected = []
    for page in sorted(source_pages):
        ranked = sorted(by_page[page], key=lambda i: (
            -probabilities[i], rows[i]["spans"][0]["start"], i))
        expected_selected.extend(ranked[:3])
    require(selected == expected_selected, "Candidate ranking changed")
    expected_top = [{"index": i, "page": rows[i]["page"],
                     "text": rows[i]["text"], "spans": rows[i]["spans"],
                     "score": probabilities[i]} for i in selected]
    require(json.loads((artifact_dir / "top-candidates.json").read_text()) ==
            expected_top, "Candidate text or span changed")

    events = json.loads(SOURCE_FILES["events"].read_text())
    frozen = labels()
    frontier = json.loads(SOURCE_FILES["frontier"].read_text())
    additions = set(frontier["reviewed_additions"])
    require(len(events) == len(frozen) == 123 and
            {event["event"] for event in events} == set(frozen) and
            len(additions) == 7, "Frozen event set changed")
    counts: dict[str, Counter] = defaultdict(Counter)
    possible = []
    for event in events:
        label = frozen[event["event"]]
        page = label["path"]
        for target in label["targets"]:
            require(source_pages[page][target["start"]:target["end"]].decode() ==
                    target["quote"], "Frozen target quote changed")
        matches = [i for i in selected if rows[i]["page"] == page and
                   any(overlapping(rows[i], target) for target in label["targets"])]
        complete = all(any(overlapping(rows[i], target) for i in matches)
                       for target in label["targets"])
        current_full = event["status"] == "full" or event["event"] in additions
        cohort = event["event"].split("/", 1)[0]
        counts[cohort]["events"] += 1
        counts[cohort]["baseline_full"] += current_full
        counts[cohort]["span_ceiling"] += complete
        counts[cohort]["new_span_ceiling"] += complete and not current_full
        if complete:
            possible.append({"event": event["event"], "category": label["category"],
                             "baseline_full": current_full,
                             "selected_indices": matches})
    new_possible = [item for item in possible if not item["baseline_full"]]
    require(result["cohorts"] == {name: dict(counter)
                                  for name, counter in sorted(counts.items())} and
            result["possible_events"] == possible and
            result["new_full_event_span_ceiling"] == len(new_possible) and
            result["new_possible_categories"] == dict(Counter(
                item["category"] for item in new_possible)),
            "Frozen-event assessment changed")
    require(result["admission_min_new_full_events"] == 8 and
            result["admission_possible_by_span"] == (len(new_possible) >= 8) and
            result["status"] == ("negative_early_stop" if len(new_possible) < 8
                                 else "requires_editorial_review") and
            result["editorial_review_completed"] is False,
            "Admission or review status changed")
    return result


if __name__ == "__main__":
    summary = verify()
    print("WikiSQE transfer screen verified:", summary["status"],
          summary["new_full_event_span_ceiling"])
