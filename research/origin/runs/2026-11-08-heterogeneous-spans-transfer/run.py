#!/usr/bin/env python3
"""Run the frozen, model-free heterogeneous AI spans transfer screen."""

from __future__ import annotations

import argparse
import collections
import hashlib
import importlib.util
import json
import math
from pathlib import Path

import duckdb


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]
FEATURE = ROOT / "research/origin/runs/2026-11-07-richness-screen/run.py"
SPLITS = {"train": 7018, "validation": 594, "test": 1188}
DATASETS = ("beigebook", "gutenberg_selected", "jmlr_pre2015",
            "standardebooks")
FEATURES = ("richness", "entropy")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def sha(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def import_feature():
    spec = importlib.util.spec_from_file_location("frozen_origin_feature", FEATURE)
    require(spec is not None and spec.loader is not None,
            "Cannot import frozen feature")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def load_rows(paths: dict[str, Path]) -> dict[str, dict[str, dict]]:
    database = duckdb.connect()
    rows_by_split: dict[str, dict[str, dict]] = {}
    ids: set[str] = set()
    source_split: dict[str, str] = {}
    group_split: dict[str, str] = {}
    ai_spans = 0
    for split, expected in SPLITS.items():
        rows = database.execute(
            "SELECT id,split,kind,cohort,requested_model,source_id,group_id,"
            "source_dataset,source_author,source_title,source_reference,"
            "source_license,source_text,text,text_sha256,offset_unit,"
            "human_verified,end_exclusive,ai_fraction_chars,spans "
            "FROM read_parquet(?)", [str(paths[split])]
        ).fetchall()
        require(len(rows) == expected, f"{split}: row count changed")
        pairs: dict[str, dict] = collections.defaultdict(dict)
        for (row_id, stored_split, kind, cohort, model, source_id, group_id,
             dataset, author, title, reference, license_name, source_text,
             text, text_hash, offset_unit, verified, exclusive, ai_fraction,
             spans) in rows:
            require(isinstance(row_id, str) and row_id not in ids,
                    "Missing or duplicate document ID")
            ids.add(row_id)
            require(stored_split == split and kind in ("human_control", "mixed"),
                    "Split or kind mismatch")
            require(isinstance(source_id, str) and source_id and
                    isinstance(group_id, str) and group_id,
                    "Missing source or group identity")
            require(source_split.setdefault(source_id, split) == split and
                    group_split.setdefault(group_id, split) == split,
                    "Source or group crosses splits")
            require(isinstance(source_text, str) and source_text and
                    isinstance(text, str) and text,
                    "Missing text")
            require(hashlib.sha256(text.encode("utf-8")).hexdigest() ==
                    text_hash, "Text hash mismatch")
            require(offset_unit == "unicode_codepoint" and exclusive is True
                    and verified is False, "Offset or provenance changed")
            require(isinstance(spans, list) and spans,
                    "Missing character-span partition")
            position = 0
            ai_chars = 0
            for span in spans:
                start, end = span["start"], span["end"]
                source_start, source_end = (span["source_start"],
                                            span["source_end"])
                require(isinstance(start, int) and isinstance(end, int) and
                        start == position and end > start and end <= len(text),
                        "Invalid or noncontiguous text span")
                require(isinstance(source_start, int) and
                        isinstance(source_end, int) and
                        0 <= source_start < source_end <= len(source_text),
                        "Invalid original source span")
                if span["label"] == "human":
                    require(text[start:end] ==
                            source_text[source_start:source_end],
                            "Retained human span changed")
                else:
                    require(span["label"] == "ai", "Unknown span label")
                    ai_chars += end - start
                    ai_spans += 1
                position = end
            require(position == len(text), "Span partition does not cover text")
            require(isinstance(ai_fraction, float) and
                    abs(ai_fraction - ai_chars / len(text)) < 1e-12,
                    "AI character fraction mismatch")
            if kind == "human_control":
                require(text == source_text and ai_chars == 0 and
                        len(spans) == 1 and model is None,
                        "Invalid human control")
            else:
                require(ai_chars > 0 and isinstance(model, str) and model,
                        "Mixed document has no AI replacement")
            require(kind not in pairs[source_id], "Duplicate source/kind row")
            pairs[source_id][kind] = {
                "id": row_id,
                "source_id": source_id,
                "group_id": group_id,
                "source_dataset": dataset,
                "cohort": cohort,
                "source_meta": (author, title, reference, license_name),
                "source_text": source_text,
                "text": text,
                "requested_model": model,
                "spans": spans,
            }
        require(len(pairs) == expected // 2,
                f"{split}: paired-source count changed")
        for source_id, pair in pairs.items():
            require(set(pair) == {"human_control", "mixed"},
                    f"{source_id}: incomplete pair")
            human, mixed = pair["human_control"], pair["mixed"]
            for key in ("source_id", "group_id", "source_dataset", "cohort",
                        "source_meta", "source_text"):
                require(human[key] == mixed[key],
                        f"{source_id}: source metadata mismatch")
        rows_by_split[split] = pairs
    require(len(ids) == 8800 and ai_spans == 5283,
            "Dataset identity or AI-span count changed")
    require(set(row["mixed"]["source_dataset"]
                for row in rows_by_split["test"].values()) == set(DATASETS),
            "Test source domains changed")
    return rows_by_split


def fixed100(text: str, feature) -> dict | None:
    words = [match.group().casefold()
             for match in feature.WORD.finditer(feature.prose(text))]
    if len(words) < 100:
        return None
    score = feature.features(" ".join(words[:100]))
    require(score["prose_tokens"] == 100, "Fixed window changed length")
    return {name: score[name] for name in FEATURES}


def document(text: str, feature) -> dict | None:
    score = feature.features(text)
    if score["richness"] is None:
        return None
    return {name: score[name] for name in FEATURES}


def ai_blocks(row: dict) -> list[tuple[dict, str, str]]:
    blocks = []
    for span in row["spans"]:
        if span["label"] == "ai":
            original = row["source_text"][span["source_start"]:
                                          span["source_end"]]
            generated = row["text"][span["start"]:span["end"]]
            blocks.append((span, original, generated))
    return sorted(blocks, key=lambda item: (
        item[0]["source_start"], item[0]["source_end"],
        item[0]["replacement_id"] or ""))


def auc(pairs: list[dict], feature: str) -> float | None:
    if len(pairs) < 20:
        return None
    values = []
    for pair in pairs:
        values.extend(((pair["machine"][feature], 1),
                       (pair["human"][feature], 0)))
    values.sort()
    negatives_seen = 0
    wins = 0.0
    at = 0
    while at < len(values):
        end = at + 1
        while end < len(values) and values[end][0] == values[at][0]:
            end += 1
        positives = sum(label for _, label in values[at:end])
        negatives = end - at - positives
        wins += positives * negatives_seen + positives * negatives / 2
        negatives_seen += negatives
        at = end
    return wins / (len(pairs) * len(pairs))


def calibrate(values: list[tuple[str, dict]], name: str) -> dict:
    scores = sorted((score[name] for _, score in values), reverse=True)
    require(len(scores) >= 800, "Insufficient human calibration coverage")
    rank = math.ceil(0.01 * len(scores))
    threshold = scores[rank - 1]
    return {
        "eligible": len(scores),
        "source_groups": len({group for group, _ in values}),
        "rank": rank,
        "threshold": threshold,
        "fpr": sum(score > threshold for score in scores) / len(scores),
    }


def summarize(pairs: list[dict], name: str, threshold: float) -> dict:
    count = len(pairs)
    return {
        "eligible_pairs": count,
        "source_groups": len({row["group_id"] for row in pairs}),
        "auroc": auc(pairs, name),
        "paired_win_fraction": (
            sum((row["machine"][name] > row["human"][name]) +
                0.5 * (row["machine"][name] == row["human"][name])
                for row in pairs) / count if count else None),
        "tpr": (sum(row["machine"][name] > threshold
                    for row in pairs) / count if count else None),
        "human_fpr": (sum(row["human"][name] > threshold
                          for row in pairs) / count if count else None),
    }


def evaluate(pairs: list[dict], calibration: list[tuple[str, dict]],
             total: int, name: str) -> dict:
    reference = calibrate(calibration, name)
    threshold = reference["threshold"]
    return {
        "total_pairs": total,
        "calibration": reference,
        "overall": summarize(pairs, name, threshold),
        "by_source_dataset": {
            dataset: summarize([row for row in pairs
                                if row["source_dataset"] == dataset],
                               name, threshold)
            for dataset in DATASETS
        },
        "by_requested_model": {
            model: summarize([row for row in pairs
                              if row["requested_model"] == model],
                             name, threshold)
            for model in sorted({row["requested_model"] for row in pairs})
        },
    }


def run(directory: Path) -> dict:
    require(duckdb.__version__ == "1.4.2", "DuckDB version changed")
    freeze = json.loads((HERE / "freeze.json").read_text())
    for key, path in (("protocol_sha256", HERE / "PROTOCOL.md"),
                      ("runner_sha256", HERE / "run.py"),
                      ("feature_sha256", FEATURE)):
        require(sha(path) == freeze[key], f"Frozen source changed: {path}")
    paths = {split: directory / f"{split}.parquet" for split in SPLITS}
    for split, path in paths.items():
        require(path.stat().st_size == freeze["files"][split]["bytes"] and
                sha(path) == freeze["files"][split]["sha256"],
                f"Input changed: {split}")
    rows = load_rows(paths)
    feature = import_feature()

    train_doc = []
    train_span = []
    for pair in rows["train"].values():
        human, mixed = pair["human_control"], pair["mixed"]
        doc = document(human["text"], feature)
        if doc is not None:
            train_doc.append((human["group_id"], doc))
        for _, original, _ in ai_blocks(mixed):
            span = fixed100(original, feature)
            if span is not None:
                train_span.append((human["group_id"], span))
                break

    test_doc = []
    test_span = []
    span_total = 0
    for pair in rows["test"].values():
        human, mixed = pair["human_control"], pair["mixed"]
        original_doc = document(human["text"], feature)
        mixed_doc = document(mixed["text"], feature)
        metadata = {
            "group_id": human["group_id"],
            "source_dataset": human["source_dataset"],
            "requested_model": mixed["requested_model"],
        }
        if original_doc is not None and mixed_doc is not None:
            test_doc.append({**metadata, "human": original_doc,
                             "machine": mixed_doc})
        for _, original, generated in ai_blocks(mixed):
            span_total += 1
            human_span = fixed100(original, feature)
            machine_span = fixed100(generated, feature)
            if human_span is not None and machine_span is not None:
                test_span.append({**metadata, "human": human_span,
                                  "machine": machine_span})

    result = {
        "dataset_revision": freeze["dataset_revision"],
        "input_sha256": {split: freeze["files"][split]["sha256"]
                         for split in SPLITS},
        "split_pairs": {split: len(rows[split]) for split in SPLITS},
        "span": {},
        "document": {},
    }
    for name in FEATURES:
        result["span"][name] = evaluate(test_span, train_span, span_total,
                                         name)
        result["document"][name] = evaluate(test_doc, train_doc,
                                             len(rows["test"]), name)
    span = result["span"]["richness"]
    document_result = result["document"]["richness"]
    span_overall = span["overall"]
    document_overall = document_result["overall"]
    result["richness_span_feasibility_pass"] = all((
        span_overall["eligible_pairs"] / span_total >= 0.75,
        span_overall["auroc"] is not None and
        span_overall["auroc"] >= 0.65,
        span_overall["tpr"] is not None and span_overall["tpr"] >= 0.10,
        span_overall["human_fpr"] is not None and
        span_overall["human_fpr"] <= 0.02,
        sum(row["eligible_pairs"] >= 40 and row["auroc"] is not None and
            row["auroc"] > 0.60
            for row in span["by_source_dataset"].values()) >= 3,
    ))
    result["richness_document_feasibility_pass"] = all((
        document_overall["eligible_pairs"] / len(rows["test"]) >= 0.95,
        document_overall["auroc"] is not None and
        document_overall["auroc"] >= 0.65,
        document_overall["tpr"] is not None and
        document_overall["tpr"] >= 0.10,
        document_overall["human_fpr"] is not None and
        document_overall["human_fpr"] <= 0.02,
    ))
    return result


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--input-directory", type=Path, required=True)
    parser.add_argument("--output", type=Path, default=HERE / "result.json")
    args = parser.parse_args()
    result = run(args.input_directory)
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    print("Wrote", args.output, "span pass:",
          result["richness_span_feasibility_pass"], "document pass:",
          result["richness_document_feasibility_pass"])


if __name__ == "__main__":
    main()
