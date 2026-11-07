#!/usr/bin/env python3
"""Check the frozen APE run and its maximum possible event-span coverage."""

import argparse
from collections import defaultdict
import gzip
import hashlib
import json
from pathlib import Path
import sys


HERE = Path(__file__).resolve().parent
RESEARCH = HERE.parent
sys.path.insert(0, str(RESEARCH / "2026-09-26-outcome-announcements"))
from replay import sources  # noqa: E402
from verify import labels  # noqa: E402


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def require(condition, message):
    if not condition:
        raise ValueError(message)


def overlaps(row, target):
    return any(max(span["start"], target["start"]) < min(span["end"], target["end"])
               for span in row["spans"])


def assess(run_dir):
    assets = json.loads((HERE / "assets.json").read_text())
    run = json.loads((run_dir / "run.json").read_text())
    for name, key in (("PROTOCOL.md", "PROTOCOL.md_sha256"),
                      ("prepare_inputs.py", "prepare_inputs.py_sha256"),
                      ("export/main.go.txt", "export_main.go_sha256"),
                      ("score.py", "score.py_sha256")):
        require(digest(HERE / name) == assets[key],
                f"Frozen input changed: {name}")
    require(run["source_tree"] == assets["unswell_commit"], "Source tree changed")
    require(run["ape_code_commit"] == assets["ape_code_commit"], "APE code changed")
    require(run["ape_model_revision"] == assets["ape_checkpoint_revision"], "APE model changed")
    require(run["backbone_revision"] == assets["backbone_revision"], "APE backbone changed")
    require(run["model_config_sha256"] == assets["checkpoint_local_config_sha256"], "Model config changed")
    require(run["model_weights_sha256"] == assets["checkpoint_weights_sha256"], "Model weights changed")
    require(run["backbone_asset_sha256"] == assets["backbone_files_sha256"], "Backbone files changed")
    for filename, key in (("scores.json.gz", "scores_sha256"),
                          ("top-candidates.json", "top_candidates_sha256")):
        require(digest(run_dir / filename) == run[key],
                f"Run artifact changed: {filename}")

    raw = json.loads(gzip.decompress((run_dir / "scores.json.gz").read_bytes()))
    rows = raw["sentences"]
    selected = raw["selected_indices"]
    require(len(rows) == run["sentences"] == 4139, "Sentence count changed")
    require(len(selected) == run["selected_candidates"] == 108, "Candidate count changed")
    require(len(raw["pairs"]) == run["scored_pairs"] == 2192, "Pair count changed")
    require(len(set(selected)) == len(selected), "Duplicate selected candidate")
    expected_sources = sources()
    require(len(expected_sources) == 36, "Source cohort changed")
    require({row["page"] for row in rows} == set(expected_sources), "Page set changed")
    require(digest(HERE / "source-hashes.json") == assets["source_hash_manifest_sha256"],
            "Source manifest changed")
    source_hashes = json.loads((HERE / "source-hashes.json").read_text())
    require(source_hashes == {name: hashlib.sha256(data).hexdigest()
                              for name, data in expected_sources.items()}, "Source bytes changed")

    best = [None] * len(rows)
    for pair in raw["pairs"]:
        i, j, value = pair["candidate"], pair["premise"], pair["score"]
        require(0 <= i < len(rows) and 0 <= j < len(rows) and i != j, "Invalid pair")
        left, right = rows[i], rows[j]
        require((left["page"], left["block"], left["piece"]) ==
                (right["page"], right["block"], right["piece"]), "Cross-boundary pair")
        best[i] = value if best[i] is None else max(best[i], value)
    by_page = defaultdict(list)
    for i, row in enumerate(rows):
        require(row["index"] == i and row["score"] == best[i], "Score or index changed")
        source = expected_sources[row["page"]]
        require(row["spans"] and all(0 <= span["start"] < span["end"] <= len(source)
                                      for span in row["spans"]), "Invalid source span")
        if best[i] is not None:
            by_page[row["page"]].append(i)
    expected_selected = []
    for page in sorted(expected_sources):
        ranked = sorted(by_page[page], key=lambda i: (-best[i], rows[i]["spans"][0]["start"]))
        expected_selected.extend(ranked[:3])
    require(selected == expected_selected, "Candidate ranking changed")
    require(json.loads((run_dir / "top-candidates.json").read_text()) ==
            [rows[i] for i in selected], "Selected candidate artifact changed")

    all_labels = labels()
    events = json.loads((RESEARCH / "2026-11-07-current-main-editorial-review" /
                         "events.json").read_text())
    require(len(events) == len(all_labels) == 123, "Event count changed")
    require({event["event"] for event in events} == set(all_labels), "Event IDs changed")
    for label in all_labels.values():
        source = expected_sources[label["path"]]
        for target in label["targets"]:
            require(source[target["start"]:target["end"]].decode() == target["quote"],
                    "Frozen target changed")

    selected_rows = [rows[i] for i in selected]
    cohorts = {name: {"events": 0, "baseline_full": 0, "span_ceiling": 0,
                      "new_span_ceiling": 0} for name in ("whole", "context", "confirmation")}
    possible = []
    overlapping_candidates = set()
    for event in events:
        event_id = event["event"]
        label = all_labels[event_id]
        cohort = event_id.split("/", 1)[0]
        counts = cohorts[cohort]
        counts["events"] += 1
        counts["baseline_full"] += event["status"] == "full"
        matches = [i for i, row in zip(selected, selected_rows)
                   if row["page"] == label["path"] and
                   any(overlaps(row, target) for target in label["targets"])]
        overlapping_candidates.update(matches)
        if all(any(overlaps(row, target) for row in selected_rows
                   if row["page"] == label["path"]) for target in label["targets"]):
            counts["span_ceiling"] += 1
            is_new = event["status"] != "full"
            counts["new_span_ceiling"] += is_new
            possible.append({"event": event_id, "category": label["category"],
                             "baseline_status": event["status"],
                             "selected_indices": matches})
    new_possible = [item for item in possible if item["baseline_status"] != "full"]
    return {"status": "negative_early_stop", "source_tree": run["source_tree"],
            "baseline_events_sha256": digest(RESEARCH / "2026-11-07-current-main-editorial-review" /
                                             "events.json"),
            "source_hash_manifest_sha256": digest(HERE / "source-hashes.json"),
            "documents": len(expected_sources), "frozen_events": len(events),
            "sentences": len(rows), "scored_pairs": len(raw["pairs"]),
            "selected_candidates": len(selected),
            "candidates_overlapping_any_event": len(overlapping_candidates),
            "cohorts": cohorts, "new_full_event_span_ceiling": len(new_possible),
            "admission_min_new_full_events": 8,
            "admission_possible_by_span": len(new_possible) >= 8,
            "possible_events": possible,
            "editorial_review_completed": False,
            "credit_note": "Span overlap is only an upper bound, never diagnosis credit.",
            "stop_note": "The necessary eight-new-FULL gate failed before manual review."}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("run_dir", type=Path)
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    result = assess(args.run_dir)
    args.output.write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps({key: result[key] for key in (
        "status", "documents", "frozen_events", "selected_candidates",
        "new_full_event_span_ceiling", "admission_possible_by_span")}, indent=2))


if __name__ == "__main__":
    main()
