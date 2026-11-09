#!/usr/bin/env python3
"""Independently check the completed RACE transfer and its saved decision."""

from __future__ import annotations

import argparse
import hashlib
import importlib.util
import json
import math
from pathlib import Path


HERE = Path(__file__).resolve().parent


def load_module(path: Path, name: str):
    spec = importlib.util.spec_from_file_location(name, path)
    if spec is None or spec.loader is None:
        raise ValueError(f"Cannot load {name}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def digest(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            h.update(block)
    return h.hexdigest()


def compare_result(saved: dict, recomputed: dict) -> None:
    if saved != recomputed:
        raise ValueError("Saved result differs from recomputed result")


def brute_auc(positives: list[float], negatives: list[float]) -> float:
    if not positives or not negatives:
        raise ValueError("Missing positive or negative scores")
    wins = sum((positive > negative) + 0.5 * (positive == negative)
               for positive in positives for negative in negatives)
    return wins / (len(positives) * len(negatives))


def verify(input_path: Path, arb_path: Path, status_path: Path,
           prediction_path: Path, journal_path: Path, result_path: Path,
           summary_path: Path) -> dict:
    freeze = load_module(HERE / "verify_freeze.py", "race_freeze_result")
    freeze.verify(input_path, arb_path)
    scorer = load_module(HERE / "score.py", "race_score_result")
    saved = json.loads(result_path.read_text(encoding="utf-8"))
    recomputed = scorer.run(input_path, status_path, prediction_path)
    compare_result(saved, recomputed)
    journal = json.loads(journal_path.read_text(encoding="utf-8"))
    if journal.get("status") != "complete" or journal.get("exit_code") != 0 or \
            journal.get("predictions_sha256") != digest(prediction_path) or \
            journal.get("checkpoint_sha256") != \
            "ae9ba214ad47fd6834f82a951ad0b7a54f0903e171fa28b7f7608d0dea969b60":
        raise ValueError("Incomplete or changed inference journal")
    parse_report = json.loads(status_path.read_text(encoding="utf-8"))
    if journal["parsed_sha256"] != parse_report["parsed_sha256"]:
        raise ValueError("Inference used another parsed input")

    rows = [json.loads(line) for line in input_path.read_text(encoding="utf-8").splitlines()]
    predictions = {row["id"]: row["logits"] for row in
                   json.loads(prediction_path.read_text(encoding="utf-8"))}
    by_pair = {}
    for row in rows:
        logits = predictions[row["item_id"]]
        peak = max(logits)
        exp = [math.exp(value - peak) for value in logits]
        score = 1 - exp[0] / sum(exp)
        by_pair.setdefault(row["pair_id"], {})[row["regime"], row["model_name"]] = (
            score, row["source_dataset"])
    positives = []
    negatives = []
    domain_pairs = {domain: [] for domain in ("owt", "wp", "xsum")}
    for members in by_pair.values():
        human, domain = members["human", None]
        for generator in ("gemma2_9b", "llama32_3b", "mistral7b", "qwen25_7b"):
            machine, machine_domain = members["h2l", generator]
            if machine_domain != domain:
                raise ValueError("Source domain changed within a pair")
            positives.append(machine)
            negatives.append(human)
            domain_pairs[domain].append((machine, human))
    if len(positives) != 240 or len(by_pair) != 60:
        raise ValueError("Incomplete paired score stream")
    independent_auc = brute_auc(positives, negatives)
    if not math.isclose(independent_auc, saved["metrics"]["race"]["auroc"],
                        rel_tol=0, abs_tol=1e-12):
        raise ValueError("Independent AUROC mismatch")
    for domain, pairs in domain_pairs.items():
        domain_auc = brute_auc([x for x, _ in pairs], [y for _, y in pairs])
        if not math.isclose(domain_auc, saved["by_domain"][domain]["auroc"],
                            rel_tol=0, abs_tol=1e-12):
            raise ValueError(f"Independent {domain} AUROC mismatch")
    if saved["status"] != "retire":
        raise ValueError("Registered advance rule was not enforced")
    summary = json.loads(summary_path.read_text(encoding="utf-8"))
    expected_hashes = {
        "input_sha256": digest(input_path),
        "parse_status_sha256": digest(status_path),
        "predictions_sha256": digest(prediction_path),
        "infer_journal_sha256": digest(journal_path),
        "result_sha256": digest(result_path),
    }
    if any(summary.get(key) != value for key, value in expected_hashes.items()) or \
            summary.get("freeze_commit") != \
            "f6b14a806dd94ac76a3cfd8d68ac1ded6fae74e9" or \
            summary.get("checkpoint_inference_passes") != 1 or \
            summary.get("generative_model_service_calls") != 0 or \
            summary.get("status") != "negative_directional_result":
        raise ValueError("Run summary changed or incomplete")
    return {"status": saved["status"], "sources": len(by_pair),
            "pairs": len(positives), "independent_auroc": independent_auc,
            "result_sha256": digest(result_path)}


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("input", "arb", "status", "predictions", "journal", "result",
                 "summary"):
        parser.add_argument(f"--{name}", type=Path, required=True)
    args = parser.parse_args()
    print(json.dumps(verify(args.input, args.arb, args.status,
                            args.predictions, args.journal, args.result,
                            args.summary)))


if __name__ == "__main__":
    main()
