#!/usr/bin/env python3
"""Score the frozen GEN source-disjoint, model-free origin-feature screen."""

from __future__ import annotations

import argparse
import collections
import gzip
import hashlib
import importlib.util
import json
import math
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]
FEATURE_PATH = ROOT / "research/origin/runs/2026-11-07-richness-screen/run.py"
ACQUIRE_PATH = HERE / "acquire.py"
PROTOCOL_PATH = HERE / "PROTOCOL.md"
LEXICON_PATH = HERE / "closed_class.txt"
FEATURE_NAMES = ("richness", "entropy", "closed_fraction", "reverse_entropy")
VIEWS = ("prefix100", "full")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def sha(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def load_module(path: Path, name: str):
    spec = importlib.util.spec_from_file_location(name, path)
    require(spec is not None and spec.loader is not None, f"Cannot load {name}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def load_inputs(human: Path, generated: Path, selected: Path, acquisition: Path) -> tuple[dict, dict]:
    freeze = json.loads((HERE / "freeze.json").read_text())
    paths = {"human": human, "generated": generated, "selected": selected,
             "acquisition": acquisition, "feature_source": FEATURE_PATH,
             "acquire_source": ACQUIRE_PATH, "protocol": PROTOCOL_PATH,
             "lexicon": LEXICON_PATH, "runner": HERE / "run.py"}
    for name, path in paths.items():
        require(sha(path) == freeze["sha256"][name], f"{name} SHA-256 changed")
    manifest = json.loads(acquisition.read_text())
    require(manifest["input_sha256"] == freeze["edited_input_sha256"] and
            manifest["input_rows"] == freeze["edited_input_rows"] and
            manifest["selected_rows"] == freeze["selected_rows"] and
            manifest["split_sha256"] == freeze["split_sha256"] and
            manifest["selected_gzip_sha256"] == freeze["sha256"]["selected"],
            "Acquisition identity changed")
    return freeze, manifest


def score(text: str, feature_module, closed: set[str]) -> dict:
    words = [match.group().casefold()
             for match in feature_module.WORD.finditer(feature_module.prose(text))]
    full = feature_module.features(text)
    require(full["prose_tokens"] == len(words), "Feature token count mismatch")
    result = {"prose_tokens": len(words), "prefix100": None, "full": None}
    if len(words) >= 50:
        result["full"] = {"richness": full["richness"],
                          "entropy": full["entropy"],
                          "closed_fraction": sum(word in closed for word in words) / len(words),
                          "reverse_entropy": -full["entropy"]}
    if len(words) >= 100:
        prefix = words[:100]
        fixed = feature_module.features(" ".join(prefix))
        result["prefix100"] = {"richness": fixed["richness"],
                               "entropy": fixed["entropy"],
                               "closed_fraction": sum(word in closed for word in prefix) / 100,
                               "reverse_entropy": -fixed["entropy"]}
    return result


def threshold(values: list[float], false_positive_fraction: float = 0.01) -> float:
    require(values and 0 <= false_positive_fraction < 1, "Invalid calibration")
    descending = sorted(values, reverse=True)
    allowed = math.floor(len(values) * false_positive_fraction)
    require(allowed < len(values), "Invalid calibration size")
    cutoff = math.nextafter(descending[allowed], math.inf)
    require(sum(value >= cutoff for value in values) <= allowed,
            "Calibration exceeded false-positive budget")
    return cutoff


def auc(positives: list[float], negatives: list[float]) -> float | None:
    if len(positives) < 20 or len(negatives) < 20:
        return None
    values = sorted([(x, 1) for x in positives] + [(x, 0) for x in negatives])
    won = 0.0
    below = 0
    i = 0
    while i < len(values):
        j = i + 1
        while j < len(values) and values[j][0] == values[i][0]:
            j += 1
        group = values[i:j]
        n_negative = sum(label == 0 for _, label in group)
        n_positive = len(group) - n_negative
        won += n_positive * (below + n_negative / 2)
        below += n_negative
        i = j
    return won / (len(positives) * len(negatives))


def quantile(values: list[float], fraction: float) -> float | None:
    if not values:
        return None
    ordered = sorted(values)
    return ordered[round((len(ordered) - 1) * fraction)]


def summarize(rows: list[dict], humans: dict, cutoff: float, view: str,
              feature: str) -> dict:
    eligible = []
    for row in rows:
        machine = row["scores"][view]
        control = humans[row["group"]][view]
        if machine is not None and control is not None:
            eligible.append((row, machine[feature], control[feature]))
    positives = [m for _, m, _ in eligible]
    controls = [h for _, _, h in eligible]
    wins = sum((m > h) + 0.5 * (m == h) for _, m, h in eligible)
    result = {"rows": len(rows), "eligible": len(eligible),
              "groups": len({row["group"] for row, _, _ in eligible}),
              "coverage": len(eligible) / len(rows) if rows else None,
              "auc": auc(positives, controls),
              "paired_wins": wins / len(eligible) if eligible else None,
              "tpr": sum(value >= cutoff for value in positives) / len(positives)
              if positives else None}
    ratios = [row["edit_ratio"] for row, _, _ in eligible if row["edit_ratio"] is not None]
    if ratios:
        result["measured_edit_ratio"] = {"p10": quantile(ratios, 0.1),
                                          "median": quantile(ratios, 0.5),
                                          "p90": quantile(ratios, 0.9)}
    return result


def group_summaries(rows: list[dict], humans: dict, cutoff: float,
                    view: str, feature: str, attribute: str) -> dict:
    groups = collections.defaultdict(list)
    for row in rows:
        groups[row[attribute]].append(row)
    return {key: summarize(value, humans, cutoff, view, feature)
            for key, value in sorted(groups.items())}


def read_humans(path: Path, acquire_module, feature_module, closed: set[str]) -> tuple[dict, dict]:
    normalized, roles, split_sha = acquire_module.human_groups(path)
    freeze = json.loads((HERE / "freeze.json").read_text())
    require(split_sha == freeze["split_sha256"], "Source split changed")
    humans = {group: score(text, feature_module, closed)
              for group, text in normalized.items()}
    return humans, roles


def read_generated(path: Path, roles: dict, feature_module, closed: set[str]) -> tuple[list[dict], dict]:
    by_key = {}
    counts = collections.Counter()
    raw_rows = 0
    with path.open("rb") as stream:
        for line in stream:
            x = json.loads(line)
            raw_rows += 1
            group = (x["source"], x["prompt_id"])
            require(group in roles and x["label"] == 1 and
                    x["text_type"] == "ai_generated" and
                    isinstance(x["text"], str) and x["text"] and
                    isinstance(x["model"], str) and x["model"],
                    "Invalid generated row")
            key = (group, x["model"])
            text_digest = hashlib.sha256(x["text"].encode()).hexdigest()
            if key not in by_key or text_digest < by_key[key][0]:
                by_key[key] = (text_digest, x["text"])
            counts[key] += 1
    require(raw_rows == 39391, "Generated row count changed")
    selected = []
    for (group, model), (_, text) in sorted(by_key.items()):
        if roles[group] != "test":
            continue
        selected.append({"group": group, "domain": group[0], "model": model,
                         "category": None, "edit_ratio": None,
                         "scores": score(text, feature_module, closed)})
    return selected, {"rows": raw_rows, "unique_source_model": len(by_key),
                      "duplicate_source_model": sum(n > 1 for n in counts.values()),
                      "missing_source_model": 5000 * 8 - len(by_key)}


def read_edited(path: Path, roles: dict, normalized: dict,
                feature_module, closed: set[str], expected_rows: int) -> list[dict]:
    selected = []
    rows = 0
    with gzip.open(path, "rt") as stream:
        for line in stream:
            x = json.loads(line)
            rows += 1
            group = (x["source"], x["prompt_id"])
            require(group in roles and x["label"] == 2 and
                    x["text_type"] == "ai_edited" and x["target_ratio"] == 30 and
                    " ".join(x["seed_text"].split()) == normalized[group],
                    "Edited source lineage changed")
            if roles[group] != "test":
                continue
            selected.append({"group": group, "domain": group[0], "model": x["model"],
                             "category": x["edit_category"], "edit_ratio": x["edit_ratio"],
                             "scores": score(x["text"], feature_module, closed)})
    require(rows == expected_rows, "Selected edit count changed")
    return selected


def evaluate(humans: dict, roles: dict, generated: list[dict],
             edited: list[dict], manifest: dict) -> dict:
    tasks = {"generated": generated, "edited": edited}
    result = {"format": "unswell-gen-edit-transfer-v1",
              "status": "development_only", "source_split_sha256": manifest["split_sha256"],
              "acquisition": {key: manifest[key] for key in (
                  "input_rows", "source_groups", "selected_rows", "selected_groups",
                  "source_mismatches_by_domain", "duplicate_selected_condition_keys")},
              "tasks": {}}
    for task, rows in tasks.items():
        features = ("richness", "entropy") if task == "generated" else (
            "closed_fraction", "reverse_entropy", "richness")
        result["tasks"][task] = {}
        for view in VIEWS:
            human_calibration = [values[view] for group, values in humans.items()
                                 if roles[group] == "calibration" and values[view] is not None]
            human_test = [values[view] for group, values in humans.items()
                          if roles[group] == "test" and values[view] is not None]
            for feature in features:
                cutoff = threshold([values[feature] for values in human_calibration])
                test_fpr = sum(values[feature] >= cutoff for values in human_test) / len(human_test)
                metrics = summarize(rows, humans, cutoff, view, feature)
                metrics.update({"calibration_humans": len(human_calibration),
                                "test_humans": len(human_test),
                                "threshold": cutoff, "test_human_fpr": test_fpr,
                                "by_domain": group_summaries(rows, humans, cutoff, view, feature, "domain"),
                                "by_model": group_summaries(rows, humans, cutoff, view, feature, "model")})
                if task == "edited":
                    metrics["by_category"] = group_summaries(
                        rows, humans, cutoff, view, feature, "category")
                if task == "edited" and view == "prefix100" and feature in (
                        "closed_fraction", "reverse_entropy"):
                    metrics["feasibility_pass"] = (
                        metrics["coverage"] is not None and metrics["coverage"] >= 0.80 and
                        metrics["auc"] is not None and metrics["auc"] >= 0.65 and
                        metrics["tpr"] is not None and metrics["tpr"] >= 0.10 and
                        test_fpr <= 0.02 and
                        sum(item["paired_wins"] is not None and
                            item["paired_wins"] >= 0.5 for item in
                            metrics["by_domain"].values()) >= 4)
                result["tasks"][task][f"{view}/{feature}"] = metrics
    return result


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--human", type=Path, required=True)
    parser.add_argument("--generated", type=Path, required=True)
    parser.add_argument("--selected", type=Path, required=True)
    parser.add_argument("--acquisition", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    _, manifest = load_inputs(args.human, args.generated, args.selected, args.acquisition)
    acquire_module = load_module(ACQUIRE_PATH, "gen_acquire")
    feature_module = load_module(FEATURE_PATH, "unswell_origin_features")
    lines = LEXICON_PATH.read_text().splitlines()
    require(lines == sorted(set(lines)) and len(lines) >= 100,
            "Closed-class lexicon changed")
    closed = set(lines)
    humans, roles = read_humans(args.human, acquire_module, feature_module, closed)
    normalized, _, _ = acquire_module.human_groups(args.human)
    generated, generation_identity = read_generated(
        args.generated, roles, feature_module, closed)
    edited = read_edited(args.selected, roles, normalized, feature_module,
                         closed, manifest["selected_rows"])
    result = evaluate(humans, roles, generated, edited, manifest)
    result["generation_identity"] = generation_identity
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    primary = result["tasks"]["edited"]["prefix100/closed_fraction"]
    print(json.dumps({"edited_test_rows": len(edited), "generated_test_rows": len(generated),
                      "edited_primary": {key: primary[key] for key in (
                          "eligible", "coverage", "auc", "paired_wins", "tpr",
                          "test_human_fpr", "feasibility_pass")}}, sort_keys=True))


if __name__ == "__main__":
    main()
