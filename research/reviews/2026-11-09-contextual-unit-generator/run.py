#!/usr/bin/env python3
"""Run one frozen, group-held-out #349 source-unit location screen."""

from __future__ import annotations

import argparse
from collections import Counter, defaultdict
import hashlib
import json
import math
import os
from pathlib import Path
import subprocess
import sys
import warnings


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
ARCHIVE_NAME = "joint-component-sol-high-evaluation/originals.json"
CATEGORIES = ("empty_framing", "wordiness", "vague_claims")
COHORTS = {"whole": 57, "context": 46, "confirmation": 20}
LEDGER = ROOT / "research/reviews/2026-11-07-current-main-editorial-review/events.json"
FRONTIER = ROOT / "research/reviews/2026-11-08-dual-gate-feasibility/category-frontier.json"


def require(ok: bool, message: str) -> None:
    if not ok:
        raise ValueError(message)


def sha_bytes(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def sha(path: Path) -> str:
    return sha_bytes(path.read_bytes())


def content_positions(raw: bytes, start: int, end: int) -> set[int]:
    require(0 <= start < end <= len(raw), "invalid target span")
    positions: set[int] = set()
    offset = start
    for char in raw[start:end].decode("utf-8"):
        width = len(char.encode("utf-8"))
        if char.isalnum():
            positions.update(range(offset, offset + width))
        offset += width
    require(offset == end and positions, "target has no aligned content")
    return positions


def span_positions(segments: list[dict]) -> set[int]:
    return {pos for span in segments for pos in range(span["start"], span["end"])}


def load_inputs(archive: Path, freeze: dict) -> tuple[dict, dict, dict, set[str], dict]:
    require(sha(archive) == freeze["archive_sha256"], "archive identity changed")
    require(sha(LEDGER) == freeze["ledger_sha256"] and
            sha(FRONTIER) == freeze["frontier_sha256"], "review identity changed")
    data = json.loads(archive.read_text())
    originals, paths, events = data["original"], data["paths"], data["reference"]["events"]
    require(len(originals) == len(paths) == 36 and len(data["groups"]) == 13 and
            len(events) == 123 and set(originals) == set(paths),
            "frozen source/event inventory changed")
    require(Counter(k.split("/", 1)[0] for k in events) == COHORTS,
            "cohort denominators changed")
    by_path = {path: page for page, path in paths.items()}
    require(len(by_path) == 36, "duplicate source path")
    group_for_page = {}
    for group, numbers in data["groups"].items():
        for number in numbers:
            page = f"page-{number:03d}"
            require(page in originals and page not in group_for_page,
                    "group mapping changed")
            group_for_page[page] = group
    require(set(group_for_page) == set(originals), "incomplete group mapping")
    for page, record in originals.items():
        doc = record["document"]
        require(doc["page"] == page and
                sha_bytes(doc["source"].encode("utf-8")) == doc["source_sha256"],
                "source hash changed")
    for event_id, event in events.items():
        require(event["path"] in by_path and
                event_id.startswith(event["path"].split("/", 1)[0] + "/"),
                "event source changed")
        require(event["targets"], "event has no targets")
        raw = originals[by_path[event["path"]]]["document"]["source"].encode("utf-8")
        for target in event["targets"]:
            require(raw[target["start"]:target["end"]].decode("utf-8") ==
                    target["quote"], "target quote/span changed")
            content_positions(raw, target["start"], target["end"])
    ledger = json.loads(LEDGER.read_text())
    require(len(ledger) == 123 and {item["event"] for item in ledger} == set(events),
            "review ledger events changed")
    additions = json.loads(FRONTIER.read_text())["reviewed_additions"]
    require(len(additions) == 7 and len(set(additions)) == 7 and
            set(additions) <= set(events), "reviewed addition set changed")
    old_full = {item["event"] for item in ledger if item["status"] == "full"}
    require(not old_full.intersection(additions), "duplicate FULL credit")
    full_now = old_full.union(additions)
    require(Counter(key.split("/", 1)[0] for key in full_now) ==
            {"whole": 28, "context": 16, "confirmation": 4},
            "current FULL baseline changed")
    return originals, paths, events, full_now, group_for_page


def export_units(exporter: Path, page: str, doc: dict) -> list[dict]:
    raw = doc["source"].encode("utf-8")
    request = json.dumps({"version": 1, "basis": "source-function-inventory",
                          "pages": [{"id": page + ".md", "format": doc["format"],
                                     "sha256": doc["source_sha256"],
                                     "text": doc["source"]}]},
                         sort_keys=True, separators=(",", ":"),
                         ensure_ascii=False).encode("utf-8")
    run = subprocess.run([str(exporter), "--source-inventory"], input=request,
                         capture_output=True, timeout=30, check=False)
    require(run.returncode == 0, "source exporter failed: " + page)
    result = json.loads(run.stdout)
    require(result["version"] == 1 and result["basis"] == "source-function-inventory"
            and result["input_sha256"] == sha_bytes(request) and
            len(result["pages"]) == 1, "exporter request identity changed")
    output = result["pages"][0]
    require(output["id"] == page + ".md" and
            output["sha256"] == doc["source_sha256"] and
            output["bytes"] == len(raw), "exported page identity changed")
    excluded = span_positions([item["span"] for item in output["excluded"]])
    units = output["units"]
    for ordinal, unit in enumerate(units):
        require(unit["ordinal"] == ordinal, "unit ordinal changed")
        bound = unit["raw_span"]
        start, end = bound["start"], bound["end"]
        require(0 <= start < end <= len(raw) and
                raw[start:end].decode("utf-8") == unit["raw_slice"],
                "unit raw slice changed")
        segments = unit["binding"]["segments"]
        require(segments and all(0 <= item["start"] < item["end"] <= len(raw)
                                 for item in segments), "invalid eligible segments")
        require(not span_positions(segments).intersection(excluded),
                "eligible segment intersects protected source")
        require(unit["text"].strip(), "empty source unit")
    return units


def marked_context(units: list[dict], index: int) -> str:
    unit = units[index]
    block = unit["binding"]["block_id"]
    kind = unit["binding"]["block_kind"]
    prev = (units[index - 1]["text"] if index > 0 and
            units[index - 1]["binding"]["block_id"] == block and
            units[index - 1]["binding"]["block_kind"] == kind else "")
    next_text = (units[index + 1]["text"] if index + 1 < len(units) and
                 units[index + 1]["binding"]["block_id"] == block and
                 units[index + 1]["binding"]["block_kind"] == kind else "")
    return f"PREV {prev} TARGET {unit['text']} NEXT {next_text}"


def make_records(originals: dict, paths: dict, events: dict, groups: dict,
                 exporter: Path) -> tuple[list[dict], dict, dict]:
    events_by_page = defaultdict(list)
    page_for_path = {path: page for page, path in paths.items()}
    for event_id, event in events.items():
        events_by_page[page_for_path[event["path"]]].append((event_id, event))
    records: list[dict] = []
    targets_by_event = {}
    for page in sorted(originals):
        doc = originals[page]["document"]
        raw = doc["source"].encode("utf-8")
        units = export_units(exporter, page, doc)
        category_positions = {category: set() for category in CATEGORIES}
        for event_id, event in events_by_page[page]:
            target_positions = [content_positions(raw, item["start"], item["end"])
                                for item in event["targets"]]
            targets_by_event[event_id] = (page, target_positions)
            if event["category"] in CATEGORIES:
                category_positions[event["category"]].update(
                    pos for target in target_positions for pos in target)
        for index, unit in enumerate(units):
            segments = unit["binding"]["segments"]
            positions = span_positions(segments)
            records.append({"page": page, "group": groups[page],
                            "ordinal": index, "segments": segments,
                            "text_sha256": unit["binding"]["text_sha256"],
                            "text": marked_context(units, index),
                            "labels": {category: bool(positions.intersection(
                                category_positions[category]))
                                       for category in CATEGORIES}})
    require(len(records) >= 1000 and len(targets_by_event) == 123,
            "source inventory incomplete")
    return records, targets_by_event, events_by_page


def score_heldout(records: list[dict], versions: dict) -> dict[str, list[float]]:
    import numpy as np
    import scipy
    import sklearn
    from scipy.sparse import hstack
    from sklearn.exceptions import ConvergenceWarning
    from sklearn.feature_extraction.text import TfidfVectorizer
    from sklearn.linear_model import LogisticRegression

    require({"numpy": np.__version__, "scipy": scipy.__version__,
             "scikit_learn": sklearn.__version__} == versions,
            "local classifier dependencies changed")
    groups = sorted({item["group"] for item in records})
    require(len(groups) == 13, "repository group count changed")
    scores = {category: [math.nan] * len(records) for category in CATEGORIES}
    for group in groups:
        train = [i for i, item in enumerate(records) if item["group"] != group]
        test = [i for i, item in enumerate(records) if item["group"] == group]
        require(train and test, "empty held-out fold")
        train_text = [records[i]["text"] for i in train]
        test_text = [records[i]["text"] for i in test]
        word = TfidfVectorizer(ngram_range=(1, 2), min_df=2,
                               max_features=30000, sublinear_tf=True,
                               token_pattern=r"(?u)\b\w+\b")
        char = TfidfVectorizer(analyzer="char", ngram_range=(3, 5), min_df=3,
                               max_features=30000, sublinear_tf=True)
        train_x = hstack((word.fit_transform(train_text),
                          char.fit_transform(train_text)), format="csr")
        test_x = hstack((word.transform(test_text), char.transform(test_text)),
                        format="csr")
        for category in CATEGORIES:
            labels = np.array([int(records[i]["labels"][category]) for i in train])
            require(set(labels.tolist()) == {0, 1},
                    "held-out training fold lacks a class: " + group + "/" + category)
            model = LogisticRegression(solver="liblinear", C=0.5,
                                       class_weight="balanced", max_iter=500,
                                       random_state=0)
            with warnings.catch_warnings():
                warnings.simplefilter("error", ConvergenceWarning)
                model.fit(train_x, labels)
            predicted = model.decision_function(test_x)
            require(len(predicted) == len(test) and np.isfinite(predicted).all(),
                    "invalid held-out scores")
            for index, score in zip(test, predicted, strict=True):
                scores[category][index] = float(score)
    require(all(math.isfinite(value) for values in scores.values() for value in values),
            "missing held-out score")
    return scores


def select(records: list[dict], scores: dict[str, list[float]]) -> list[int]:
    by_page = defaultdict(list)
    for index, item in enumerate(records):
        by_page[item["page"]].append(index)
    require(len(by_page) == 36, "candidate pages changed")
    selected = []
    for page in sorted(by_page):
        used = set()
        for category in CATEGORIES:
            ranked = sorted(by_page[page],
                            key=lambda i: (-scores[category][i],
                                           records[i]["ordinal"]))
            chosen = next((i for i in ranked if i not in used), None)
            require(chosen is not None, "page lacks three distinct units: " + page)
            used.add(chosen)
            selected.append(chosen)
    require(len(selected) == 108 and len(set(selected)) == 108,
            "selected candidate count changed")
    return selected


def evaluate(records: list[dict], selected: list[int], targets: dict,
             events: dict, full_now: set[str], freeze: dict,
             scores: dict[str, list[float]]) -> tuple[dict, dict]:
    selected_positions = defaultdict(set)
    for index in selected:
        item = records[index]
        selected_positions[item["page"]].update(span_positions(item["segments"]))
    covered = {event_id for event_id, (page, target_sets) in targets.items()
               if all(target <= selected_positions[page] for target in target_sets)}
    gains = covered - full_now
    cohort = {}
    for name, total in COHORTS.items():
        cohort[name] = {"events": total,
                        "current_full": sum(key.startswith(name + "/")
                                            for key in full_now),
                        "location_covered": sum(key.startswith(name + "/")
                                                for key in covered),
                        "previously_nonfull_location_covered":
                            sum(key.startswith(name + "/") for key in gains)}
    category_gains = Counter(events[key]["category"] for key in gains)
    screen_passed = (len(gains) >= 8 and
                     all(cohort[name]["previously_nonfull_location_covered"] >= 1
                         for name in COHORTS) and len(category_gains) >= 2)
    public = {"format": "unswell-349-contextual-unit-generator-v1",
              "scope": "exposed_group_heldout_location_screen_only",
              "main_commit": freeze["main_commit"],
              "main_tree": freeze["main_tree"],
              "archive_sha256": freeze["archive_sha256"],
              "exporter_sha256": freeze["exporter_sha256"],
              "protocol_sha256": freeze["protocol_sha256"],
              "script_sha256": freeze["script_sha256"],
              "ledger_sha256": freeze["ledger_sha256"],
              "frontier_sha256": freeze["frontier_sha256"],
              "dependencies": freeze["dependencies"],
              "pages": 36, "groups": 13, "events": 123,
              "units": len(records), "selected_units": len(selected),
              "positive_units_by_category":
                  {name: sum(item["labels"][name] for item in records)
                   for name in CATEGORIES},
              "selected_with_modeled_target_overlap":
                  sum(any(records[i]["labels"].values()) for i in selected),
              "cohorts": cohort,
              "newly_covered_categories": dict(sorted(category_gains.items())),
              "newly_covered_events": len(gains),
              "screen_passed": screen_passed,
              "qualified": False, "hosted_model_calls": 0}
    private = {"format": "unswell-349-contextual-unit-generator-private-v1",
               "selected": [{"page": records[i]["page"],
                             "ordinal": records[i]["ordinal"],
                             "text_sha256": records[i]["text_sha256"],
                             "segments": records[i]["segments"],
                             "scores": {name: scores[name][i]
                                        for name in CATEGORIES}}
                            for i in selected],
               "covered_events": sorted(covered), "gained_events": sorted(gains)}
    return public, private


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--archive-root", type=Path, required=True)
    parser.add_argument("--exporter", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--private-detail", type=Path, required=True)
    args = parser.parse_args()
    require(not args.output.exists() and not args.private_detail.exists(),
            "output already exists")
    require(os.environ.get("OMP_NUM_THREADS") == "1" and
            os.environ.get("OPENBLAS_NUM_THREADS") == "1" and
            os.environ.get("MKL_NUM_THREADS") == "1",
            "BLAS thread limits must be one")
    freeze = json.loads((HERE / "freeze.json").read_text())
    require(freeze["format"] == "unswell-349-contextual-unit-freeze-v1" and
            freeze["status"] == "frozen_before_run" and
            freeze["protocol_sha256"] == sha(HERE / "PROTOCOL.md") and
            freeze["script_sha256"] == sha(Path(__file__)) and
            freeze["exporter_sha256"] == sha(args.exporter) and
            freeze["python_version"] == sys.version.split()[0],
            "frozen method or exporter identity changed")
    commit = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=ROOT,
                                     text=True).strip()
    tree = subprocess.check_output(["git", "rev-parse", "HEAD^{tree}"],
                                   cwd=ROOT, text=True).strip()
    require(commit == freeze["main_commit"] and tree == freeze["main_tree"],
            "current main source tree changed")
    archive = args.archive_root / ARCHIVE_NAME
    originals, paths, events, full_now, groups = load_inputs(archive, freeze)
    records, targets, _ = make_records(originals, paths, events, groups, args.exporter)
    scores = score_heldout(records, freeze["dependencies"])
    selected = select(records, scores)
    public, private = evaluate(records, selected, targets, events, full_now,
                               freeze, scores)
    args.output.write_text(json.dumps(public, indent=2, sort_keys=True) + "\n")
    args.private_detail.write_text(json.dumps(private, indent=2, sort_keys=True) + "\n")
    print(json.dumps({"units": public["units"],
                      "selected_units": public["selected_units"],
                      "newly_covered_events": public["newly_covered_events"],
                      "screen_passed": public["screen_passed"]}, sort_keys=True))


if __name__ == "__main__":
    main()
