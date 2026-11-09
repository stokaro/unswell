#!/usr/bin/env python3
"""Score a frozen RACE transfer without exposing per-document predictions."""

from __future__ import annotations

import argparse
import collections
import hashlib
import importlib.util
import json
import math
from pathlib import Path
import random


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]
RICHNESS = ROOT / "research/origin/runs/2026-11-07-richness-screen/run.py"
INPUT_SHA256 = "feaa0b6f8fabf2f33f2ce3f4d248a1dc55d1b1b2dc3f7266bcfcdeb2043a6bb4"
DOMAINS = ("owt", "wp", "xsum")
GENERATORS = ("gemma2_9b", "llama32_3b", "mistral7b", "qwen25_7b")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def digest(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            h.update(block)
    return h.hexdigest()


def load_module(path: Path, name: str):
    spec = importlib.util.spec_from_file_location(name, path)
    require(spec is not None and spec.loader is not None, f"Cannot load {name}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def score_logits(logits: list) -> tuple[float, float]:
    require(isinstance(logits, list) and len(logits) == 4 and
            all(isinstance(v, (int, float)) and math.isfinite(v) for v in logits),
            "Invalid four-class logits")
    peak = max(logits)
    weights = [math.exp(value - peak) for value in logits]
    total = sum(weights)
    return 1 - weights[0] / total, weights[1] / total


def auc(pairs: list[dict], key: str) -> float | None:
    if not pairs:
        return None
    values = []
    for row in pairs:
        values.extend(((row["machine"][key], 1), (row["human"][key], 0)))
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
    n = len(pairs)
    return wins / (n * n)


def paired_win(pairs: list[dict], key: str) -> float | None:
    if not pairs:
        return None
    wins = sum((row["machine"][key] > row["human"][key]) +
               0.5 * (row["machine"][key] == row["human"][key])
               for row in pairs)
    return wins / len(pairs)


def bootstrap(pairs: list[dict], key: str) -> dict:
    sources = sorted({row["pair_id"] for row in pairs})
    by_source = collections.defaultdict(list)
    for row in pairs:
        by_source[row["pair_id"]].append(row)
    rng = random.Random(3492026)
    aucs = []
    wins = []
    for _ in range(2000):
        chosen = [rng.choice(sources) for _ in sources]
        sample = [row for source in chosen for row in by_source[source]]
        aucs.append(auc(sample, key))
        wins.append(paired_win(sample, key))
    def interval(values: list[float]) -> list[float]:
        values.sort()
        return [values[49], values[1949]]
    return {"auroc_95pct": interval(aucs), "paired_win_95pct": interval(wins)}


def load_inputs(input_path: Path, status_path: Path, prediction_path: Path):
    require(digest(input_path) == INPUT_SHA256, "Input SHA-256 changed")
    rows = [json.loads(line) for line in input_path.read_text(encoding="utf-8").splitlines()]
    require(len(rows) == 300, "Input count changed")
    by_id = {row["item_id"]: row for row in rows}
    require(len(by_id) == 300, "Duplicate input ID")
    for row in rows:
        require(hashlib.sha256(row["article"].encode("utf-8")).hexdigest() ==
                row["text_sha256"], "Raw text hash changed")
    parse_report = json.loads(status_path.read_text(encoding="utf-8"))
    require(parse_report["input_sha256"] == INPUT_SHA256, "Parser used another input")
    status = parse_report["status"]
    require(len(status) == 300 and len({item["item_id"] for item in status}) == 300 and
            {item["item_id"] for item in status} == set(by_id),
            "Incomplete parse status")
    predictions = json.loads(prediction_path.read_text(encoding="utf-8"))
    require(isinstance(predictions, list), "Invalid predictions")
    scores = {}
    for item in predictions:
        item_id = item.get("id")
        require(item_id in by_id and item_id not in scores,
                "Unknown or duplicate prediction ID")
        scores[item_id] = score_logits(item.get("logits"))
    parsed_ids = {item["item_id"] for item in status if item["status"] == "parsed"}
    require(set(scores) == parsed_ids, "Predictions do not match parsed records")
    return rows, status, scores


def run(input_path: Path, status_path: Path, prediction_path: Path) -> dict:
    rows, status, scores = load_inputs(input_path, status_path, prediction_path)
    richness = load_module(RICHNESS, "unswell_race_richness")
    by_pair = collections.defaultdict(dict)
    for row in rows:
        item_id = row["item_id"]
        if item_id not in scores:
            continue
        features = richness.features(row["article"])
        words = [match.group().casefold() for match in richness.WORD.finditer(
            richness.prose(row["article"]))]
        prefix = richness.features(" ".join(words[:100])) if len(words) >= 100 else None
        by_pair[row["pair_id"]][row["regime"], row["model_name"]] = {
            "race": scores[item_id][0],
            "polished": scores[item_id][1],
            "richness_full": features["richness"],
            "richness_prefix100": prefix["richness"] if prefix else None,
            "domain": row["source_dataset"],
        }
    pairs = []
    for pair_id, members in by_pair.items():
        human = members.get(("human", None))
        if human is None:
            continue
        for generator in GENERATORS:
            machine = members.get(("h2l", generator))
            if machine is not None:
                pairs.append({"pair_id": pair_id, "generator": generator,
                              "domain": human["domain"], "human": human,
                              "machine": machine})
    human_count = sum(("human", None) in members for members in by_pair.values())
    coverage = {"human": human_count, "h2l": sum(
        item["item_id"] in scores and item["regime"] == "h2l" for item in rows),
        "pairs": len(pairs), "parse_status": dict(collections.Counter(
            item["status"] for item in status)),
        "truncated": sum(item["truncated"] for item in status)}
    metrics = {}
    for key in ("race", "polished", "richness_full", "richness_prefix100"):
        eligible = [row for row in pairs if row["human"][key] is not None and
                    row["machine"][key] is not None]
        metrics[key] = {"pairs": len(eligible), "sources": len({
            row["pair_id"] for row in eligible}),
            "auroc": auc(eligible, key), "paired_win": paired_win(eligible, key)}
        if key == "race" and eligible:
            metrics[key]["source_bootstrap"] = bootstrap(eligible, key)
    by_domain = {domain: {"pairs": len(group), "auroc": auc(group, "race"),
                          "paired_win": paired_win(group, "race")}
                 for domain in DOMAINS
                 for group in [[row for row in pairs if row["domain"] == domain]]}
    by_generator = {generator: {"pairs": len(group), "auroc": auc(group, "race"),
                                "paired_win": paired_win(group, "race")}
                    for generator in GENERATORS
                    for group in [[row for row in pairs if row["generator"] == generator]]}
    advance = (coverage["human"] >= 54 and coverage["h2l"] >= 216 and
               metrics["race"]["auroc"] is not None and
               metrics["race"]["auroc"] >= 0.75 and
               metrics["race"]["paired_win"] >= 0.70 and
               all(group["paired_win"] is not None and group["paired_win"] >= 0.60
                   for group in by_domain.values()))
    return {"status": "advance" if advance else "retire", "coverage": coverage,
            "metrics": metrics, "by_domain": by_domain,
            "by_generator": by_generator,
            "sha256": {"input": digest(input_path), "parse_status": digest(status_path),
                       "predictions": digest(prediction_path)}}


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--status", type=Path, required=True)
    parser.add_argument("--predictions", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    result = run(args.input, args.status, args.predictions)
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n",
                           encoding="utf-8")
    print(json.dumps({"status": result["status"], "coverage": result["coverage"],
                      "race": result["metrics"]["race"]}))


if __name__ == "__main__":
    main()
