#!/usr/bin/env python3
"""Run the frozen, offline ARB source-matched origin-feature transfer screen."""

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
DOMAINS = ("owt", "wp", "xsum")
GENERATORS = ("gemma2_9b", "llama32_3b", "mistral7b", "qwen25_7b")
REGIMES = ("free_llm", "h2l", "llm2l")
MODES = ("full", "prefix100")
FEATURES = ("richness", "entropy")
SALT = "unswell-arb-v1"
EXPECTED = {"human": 1800, "free_llm": 7200, "h2l": 7200, "llm2l": 7200}
PROVENANCE = {
    "human": ("human", "human", False),
    "free_llm": ("llm", "llm", False),
    "h2l": ("human", "llm", True),
    "llm2l": ("llm", "llm", True),
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def sha(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def text_sha(text: str) -> str:
    return hashlib.sha256(" ".join(text.split()).encode("utf-8")).hexdigest()


def split_pairs(humans: dict[str, dict]) -> tuple[dict[str, str], str]:
    by_domain = collections.defaultdict(list)
    for pair_id, row in humans.items():
        by_domain[row["source_dataset"]].append(pair_id)
    assignments = {}
    ordered = []
    for domain in DOMAINS:
        ids = by_domain[domain]
        require(len(ids) == 600, f"Expected 600 human sources in {domain}")
        ids.sort(key=lambda pair: (
            hashlib.sha256((pair + SALT).encode()).hexdigest(), pair
        ))
        for index, pair in enumerate(ids):
            role = "calibration" if index < 300 else "test"
            assignments[pair] = role
            ordered.append((domain, pair, role))
    canonical = json.dumps(ordered, separators=(",", ":"), ensure_ascii=False)
    digest = hashlib.sha256(canonical.encode()).hexdigest()
    return assignments, digest


def load_rows(path: Path) -> tuple[dict, dict]:
    database = duckdb.connect()
    rows = database.execute(
        "SELECT id,pair_id,source_dataset,sample_index,generator_model,"
        "regime,text,source_text,content_origin,surface_origin,"
        "is_llm_mediated,source_sha256,text_sha256 FROM read_parquet(?)", [str(path)]
    ).fetchall()
    require(len(rows) == 23400, "ARB row count mismatch")
    ids = set()
    groups = collections.defaultdict(dict)
    humans = {}
    counts = collections.Counter()
    for (row_id, pair_id, domain, source_index, generator, regime, text,
         source_text, content_origin, surface_origin, mediated,
         source_hash, text_hash) in rows:
        require(isinstance(row_id, str) and row_id not in ids,
                "Missing or duplicate row ID")
        ids.add(row_id)
        require(isinstance(pair_id, str) and pair_id, "Missing pair ID")
        require(domain in DOMAINS and isinstance(source_index, int),
                "Invalid source identity")
        require(regime in EXPECTED, "Unknown regime")
        require((content_origin, surface_origin, mediated) == PROVENANCE[regime],
                "Invalid provenance metadata")
        require(isinstance(text, str) and text and
                isinstance(source_text, str) and source_text,
                "Missing text")
        require(text_sha(text) == text_hash and
                text_sha(source_text) == source_hash,
                "Stored text or source hash mismatch")
        require(generator is None if regime == "human" else generator in GENERATORS,
                "Invalid generator")
        key = (regime, generator)
        require(key not in groups[pair_id], "Duplicate quartet member")
        record = {
            "pair_id": pair_id,
            "source_dataset": domain,
            "sample_index": source_index,
            "generator_model": generator,
            "regime": regime,
            "text": text,
            "source_text": source_text,
            "source_sha256": source_hash,
            "text_sha256": text_hash,
        }
        groups[pair_id][key] = record
        if regime == "human":
            humans[pair_id] = record
        counts[regime] += 1
    require(dict(counts) == EXPECTED and len(groups) == 1800,
            "Regime or pair count mismatch")
    for pair_id, members in groups.items():
        require(len(members) == 13, "Incomplete quartet")
        human = members.get(("human", None))
        require(human is not None and human["text_sha256"] ==
                human["source_sha256"], "Missing human source")
        for generator in GENERATORS:
            free = members.get(("free_llm", generator))
            h2l = members.get(("h2l", generator))
            llm2l = members.get(("llm2l", generator))
            require(free is not None and h2l is not None and llm2l is not None,
                    "Missing generator regime")
            for item in (free, h2l, llm2l):
                require((item["source_dataset"], item["sample_index"]) ==
                        (human["source_dataset"], human["sample_index"]),
                        "Pair source identity mismatch")
            require(free["source_sha256"] == human["text_sha256"] and
                    h2l["source_sha256"] == human["text_sha256"] and
                    llm2l["source_sha256"] == free["text_sha256"],
                    "Regime source lineage mismatch")
    return humans, groups


def score(text: str, feature_module) -> dict:
    full = feature_module.features(text)
    words = [match.group().casefold()
             for match in feature_module.WORD.finditer(feature_module.prose(text))]
    require(full["prose_tokens"] == len(words), "Feature token count mismatch")
    prefix = (feature_module.features(" ".join(words[:100]))
              if len(words) >= 100 else None)
    return {
        "prose_tokens": len(words),
        "full": {key: full[key] for key in FEATURES}
                if full["richness"] is not None else None,
        "prefix100": {key: prefix[key] for key in FEATURES}
                     if prefix is not None else None,
    }


def auc(pairs: list[dict], feature: str) -> float | None:
    if len(pairs) < 20:
        return None
    values = []
    for pair in pairs:
        values.extend(((pair["machine"][feature], 1),
                       (pair["control"][feature], 0)))
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


def calibrate(scores: list[dict | None], feature: str) -> dict:
    values = sorted((row[feature] for row in scores if row is not None),
                    reverse=True)
    require(len(values) >= 800, "Insufficient human calibration coverage")
    rank = math.ceil(0.01 * len(values))
    threshold = values[rank - 1]
    return {
        "total": len(scores),
        "eligible": len(values),
        "rank": rank,
        "threshold": threshold,
        "fpr": sum(value > threshold for value in values) / len(values),
    }


def human_fpr(scores: list[dict | None], feature: str, threshold: float) -> dict:
    values = [row[feature] for row in scores if row is not None]
    return {
        "total": len(scores),
        "eligible": len(values),
        "fpr": sum(value > threshold for value in values) / len(values)
               if values else None,
    }


def summarize(pairs: list[dict], feature: str, threshold: float) -> dict:
    n = len(pairs)
    wins = sum(
        (row["machine"][feature] > row["control"][feature])
        + 0.5 * (row["machine"][feature] == row["control"][feature])
        for row in pairs
    )
    return {
        "eligible_pairs": n,
        "sources": len({row["pair_id"] for row in pairs}),
        "auroc": auc(pairs, feature),
        "paired_win_fraction": wins / n if n else None,
        "tpr": sum(row["machine"][feature] > threshold for row in pairs) / n
               if n else None,
    }


def run(source: Path) -> dict:
    require(duckdb.__version__ == "1.4.2", "DuckDB version changed")
    freeze = json.loads((HERE / "freeze.json").read_text())
    for key, path in (
        ("dataset_sha256", source),
        ("protocol_sha256", HERE / "PROTOCOL.md"),
        ("script_sha256", Path(__file__)),
        ("feature_sha256", FEATURE),
    ):
        require(sha(path) == freeze[key], f"{key} mismatch")
    require(source.stat().st_size == freeze["dataset_bytes"],
            "Dataset size mismatch")
    humans, groups = load_rows(source)
    assignments, split_hash = split_pairs(humans)
    require(split_hash == freeze["split_sha256"], "Source split changed")
    spec = importlib.util.spec_from_file_location("unswell_origin_features", FEATURE)
    require(spec is not None and spec.loader is not None,
            "Feature source unavailable")
    feature_module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(feature_module)

    scores = {}
    for pair_id, members in groups.items():
        scores[pair_id, "human", None] = score(
            members["human", None]["text"], feature_module
        )
        if assignments[pair_id] != "test":
            continue
        for regime in REGIMES:
            for generator in GENERATORS:
                scores[pair_id, regime, generator] = score(
                    members[regime, generator]["text"], feature_module
                )

    result = {
        "input": {
            "dataset_commit": freeze["dataset_commit"],
            "dataset_sha256": freeze["dataset_sha256"],
            "feature_sha256": freeze["feature_sha256"],
            "split_sha256": split_hash,
            "rows": 23400,
            "source_groups": 1800,
            "calibration_sources": 900,
            "test_sources": 900,
        },
        "interpretation": "Matched benchmark regimes, not proof of authorship or editorial quality",
        "measures": {},
    }
    for mode in MODES:
        calibration_human = [
            scores[pair_id, "human", None][mode]
            for pair_id, role in assignments.items() if role == "calibration"
        ]
        test_human = [
            scores[pair_id, "human", None][mode]
            for pair_id, role in assignments.items() if role == "test"
        ]
        mode_result = {}
        for feature in FEATURES:
            calibration = calibrate(calibration_human, feature)
            threshold = calibration["threshold"]
            test_fpr = human_fpr(test_human, feature, threshold)
            regimes = {}
            for regime in REGIMES:
                pairs = []
                machine_abstained = 0
                control_abstained = 0
                for pair_id, role in assignments.items():
                    if role != "test":
                        continue
                    human_score = scores[pair_id, "human", None][mode]
                    for generator in GENERATORS:
                        machine_score = scores[pair_id, regime, generator][mode]
                        if machine_score is None:
                            machine_abstained += 1
                        if human_score is None:
                            control_abstained += 1
                        if machine_score is None or human_score is None:
                            continue
                        pairs.append({
                            "pair_id": pair_id,
                            "domain": humans[pair_id]["source_dataset"],
                            "generator": generator,
                            "machine": machine_score,
                            "control": human_score,
                        })
                regime_result = {
                    "total_machine": 3600,
                    "machine_abstained": machine_abstained,
                    "control_abstained_for_machine_rows": control_abstained,
                    "overall": summarize(pairs, feature, threshold),
                }
                if regime == "h2l":
                    regime_result["by_domain"] = {
                        domain: summarize(
                            [row for row in pairs if row["domain"] == domain],
                            feature, threshold
                        ) for domain in DOMAINS
                    }
                    regime_result["by_generator"] = {
                        generator: summarize(
                            [row for row in pairs if row["generator"] == generator],
                            feature, threshold
                        ) for generator in GENERATORS
                    }
                regimes[regime] = regime_result
            rewrite_pairs = []
            for pair_id, role in assignments.items():
                if role != "test":
                    continue
                for generator in GENERATORS:
                    rewritten = scores[pair_id, "llm2l", generator][mode]
                    generated = scores[pair_id, "free_llm", generator][mode]
                    if rewritten is not None and generated is not None:
                        rewrite_pairs.append({
                            "pair_id": pair_id,
                            "machine": rewritten,
                            "control": generated,
                        })
            mode_result[feature] = {
                "calibration": calibration,
                "test_human": test_fpr,
                "regimes": regimes,
                "llm2l_vs_free_llm": summarize(
                    rewrite_pairs, feature, threshold
                ),
            }
        result["measures"][mode] = mode_result
    full = result["measures"]["full"]["richness"]
    h2l = full["regimes"]["h2l"]
    result["richness_h2l_feasibility_pass"] = all((
        h2l["overall"]["eligible_pairs"] / h2l["total_machine"] >= 0.95,
        h2l["overall"]["auroc"] is not None
        and h2l["overall"]["auroc"] >= 0.65,
        h2l["overall"]["tpr"] is not None
        and h2l["overall"]["tpr"] >= 0.10,
        full["test_human"]["fpr"] is not None
        and full["test_human"]["fpr"] <= 0.02,
        sum(row["auroc"] is not None and row["auroc"] > 0.60
            for row in h2l["by_domain"].values()) >= 2,
        sum(row["auroc"] is not None and row["auroc"] > 0.60
            for row in h2l["by_generator"].values()) >= 3,
    ))
    return result


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--output", type=Path, default=HERE / "result.json")
    args = parser.parse_args()
    result = run(args.input)
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    print(f"Wrote {args.output}; H2L richness feasibility: "
          f"{'passed' if result['richness_h2l_feasibility_pass'] else 'failed'}")


if __name__ == "__main__":
    main()
