#!/usr/bin/env python3
"""Evaluate a frozen, model-free feature ensemble on paired origin corpora."""

from __future__ import annotations

import argparse
import collections
import hashlib
import importlib.util
import json
import math
import statistics
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]
ARB_LOADER = ROOT / "research/origin/runs/2026-11-08-arb-transfer/run.py"
RICHNESS = ROOT / "research/origin/runs/2026-11-07-richness-screen/run.py"
LEXICON = ROOT / "research/origin/runs/2026-11-08-gen-edit-transfer/closed_class.txt"
DOMAINS = ("owt", "wp", "xsum")
GENERATORS = ("gemma2_9b", "llama32_3b", "mistral7b", "qwen25_7b")
REGIMES = ("h2l", "free_llm", "llm2l")
ACADEMIC_COUNTS = {
    ("train", "human"): 627, ("train", "ai"): 1125,
    ("val", "human"): 83, ("val", "ai"): 131,
    ("test", "human"): 89, ("test", "ai"): 143,
    ("human_only", "human"): 400,
}
DATASET_SHA = {
    "arb": "66e48240618b92ce876e2097d8df52f06661a70e31791b91a95b15688a099ab0",
    "academic": "428de968a71d64ddaf9f0d5afe7e4e295b2adeace77f57fd29c6252ae082d29c",
}
DATASET_BYTES = {"arb": 27439482, "academic": 75432568}
SALT = "unswell-ensemble-v1|"
FEATURE_NAMES = ("richness", "entropy", "closed_fraction", "mean_length_tenths", "repeated_bigram_coverage")
SCORE_NAMES = ("ensemble", "richness")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def sha(path: Path) -> str:
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


def identities(arb: Path, academic: Path) -> dict:
    paths = {
        "protocol": HERE / "PROTOCOL.md", "script": Path(__file__),
        "richness": RICHNESS, "lexicon": LEXICON,
        "arb_loader": ARB_LOADER, "arb": arb, "academic": academic,
    }
    for name in DATASET_SHA:
        require(paths[name].stat().st_size == DATASET_BYTES[name],
                f"{name} byte count changed")
        require(sha(paths[name]) == DATASET_SHA[name],
                f"{name} SHA-256 changed")
    return {name: sha(path) for name, path in paths.items()}


def split_arb(humans: dict) -> tuple[dict[str, str], str]:
    by_domain = collections.defaultdict(list)
    for pair, row in humans.items():
        by_domain[row["source_dataset"]].append(pair)
    roles = {}
    audit = []
    for domain in DOMAINS:
        pairs = by_domain[domain]
        require(len(pairs) == 600, f"Expected 600 human sources in {domain}")
        pairs.sort(key=lambda pair: (hashlib.sha256((SALT + pair).encode()).hexdigest(), pair))
        for index, pair in enumerate(pairs):
            role = "train" if index < 200 else "calibration" if index < 300 else "test"
            roles[pair] = role
            audit.append((domain, pair, role))
    require(len(roles) == 1800, "ARB role collision")
    digest = hashlib.sha256(json.dumps(audit, separators=(",", ":")).encode()).hexdigest()
    return roles, digest


def load_academic(path: Path, score_text=None) -> tuple[dict, list, dict, dict]:
    counts = collections.Counter()
    seen = set()
    paper_roles = {}
    humans = {}
    machines = []
    human_only = {}
    with path.open(encoding="utf-8") as stream:
        for number, line in enumerate(stream, 1):
            row = json.loads(line)
            paper = row.get("paper_id")
            role = row.get("split")
            variant = row.get("variant")
            label = "human" if variant == "human" else "ai"
            require(isinstance(paper, str) and paper and isinstance(row.get("text"), str)
                    and row["text"], f"Invalid academic row {number}")
            require((role, label) in ACADEMIC_COUNTS, f"Unknown academic role {number}")
            key = (paper, variant)
            require(key not in seen, f"Duplicate academic variant {number}")
            seen.add(key)
            require(paper_roles.setdefault(paper, role) == role,
                    f"Paper crosses split at row {number}")
            if label == "human":
                require(row.get("generator") is None and row.get("prompt") is None,
                        f"Human generator at row {number}")
            else:
                require(isinstance(row.get("generator"), str) and
                        isinstance(row.get("prompt"), str),
                        f"Machine metadata at row {number}")
            counts[role, label] += 1
            if score_text is None:
                continue
            if role == "test" and label == "human":
                humans[paper] = score_text(row["text"])
            elif role == "test":
                machines.append((paper, row["generator"], score_text(row["text"])))
            elif role == "human_only":
                human_only[paper] = score_text(row["text"])
    require(counts == ACADEMIC_COUNTS, f"Academic split counts changed: {dict(counts)}")
    if score_text is not None:
        require(all(paper in humans for paper, _, _ in machines),
                "Machine variant lacks test human source")
    return humans, machines, human_only, dict(counts)


def feature_vector(text: str, module, closed: set[str]) -> tuple[float, ...] | None:
    cleaned = module.prose(text)
    words = [match.group().casefold() for match in module.WORD.finditer(cleaned)]
    if len(words) < 100:
        return None
    words = words[:100]
    base = module.features(" ".join(words))
    require(base["prose_tokens"] == 100, "Tokenizer disagreement")
    bigrams = list(zip(words, words[1:]))
    frequencies = collections.Counter(bigrams)
    values = (
        base["richness"], base["entropy"],
        sum(word in closed for word in words) / 100,
        sum(map(len, words)) / 1000,
        sum(frequencies[pair] > 1 for pair in bigrams) / 99,
    )
    require(all(value is not None and math.isfinite(value) for value in values),
            "Invalid feature")
    return values


def fit(training: list[tuple[tuple[float, ...], int, float]]) -> dict:
    require(2850 <= len(training) <= 3000 and len(training) % 5 == 0,
            "Training source coverage changed")
    means = tuple(statistics.mean(row[0][i] for row in training) for i in range(5))
    deviations = tuple(statistics.pstdev(row[0][i] for row in training) or 1.0
                       for i in range(5))
    weights = [0.0] * 5
    bias = 0.0
    for _ in range(500):
        gradients = [0.0] * 5
        bias_gradient = 0.0
        for vector, label, mass in training:
            normalized = [(vector[i] - means[i]) / deviations[i] for i in range(5)]
            z = bias + sum(weights[i] * normalized[i] for i in range(5))
            probability = 1 / (1 + math.exp(-max(-40, min(40, z))))
            error = mass * (probability - label)
            bias_gradient += error
            for i in range(5):
                gradients[i] += error * normalized[i]
        for i in range(5):
            weights[i] -= 0.5 * (gradients[i] + 0.02 * weights[i])
        bias -= 0.5 * bias_gradient
    return {"means": means, "deviations": deviations,
            "weights": tuple(weights), "bias": bias}


def scores(vector: tuple[float, ...] | None, model: dict) -> dict | None:
    if vector is None:
        return None
    z = model["bias"] + sum(
        model["weights"][i] * (vector[i] - model["means"][i]) / model["deviations"][i]
        for i in range(5)
    )
    return {"ensemble": z, "richness": vector[0]}


def threshold(humans: list[dict], name: str) -> dict:
    values = sorted((row[name] for row in humans if row is not None), reverse=True)
    require(285 <= len(values) <= 300,
            "Calibration human eligibility changed")
    rank = math.ceil(0.01 * len(values))
    point = values[rank - 1]
    return {"score": point, "eligible": len(values),
            "flagged": sum(value > point for value in values)}


def auc(pairs: list[tuple[dict, dict]], name: str) -> float | None:
    if len(pairs) < 20:
        return None
    values = sorted([(machine[name], 1) for human, machine in pairs]
                    + [(human[name], 0) for human, machine in pairs])
    wins = 0.0
    negatives_seen = 0
    index = 0
    while index < len(values):
        end = index + 1
        while end < len(values) and values[end][0] == values[index][0]:
            end += 1
        positives = sum(label for _, label in values[index:end])
        negatives = end - index - positives
        wins += positives * negatives_seen + positives * negatives / 2
        negatives_seen += negatives
        index = end
    return wins / (len(pairs) * len(pairs))


def summarize(pairs: list[tuple[dict | None, dict | None]], name: str,
              point: float) -> dict:
    eligible = [(human, machine) for human, machine in pairs
                if human is not None and machine is not None]
    wins = sum((machine[name] > human[name]) +
               0.5 * (machine[name] == human[name]) for human, machine in eligible)
    return {"raw_pairs": len(pairs), "eligible_pairs": len(eligible),
            "abstained_pairs": len(pairs) - len(eligible),
            "auroc": auc(eligible, name),
            "paired_win_fraction": wins / len(eligible) if eligible else None,
            "flagged": sum(machine[name] > point for _, machine in eligible),
            "tpr": sum(machine[name] > point for _, machine in eligible) / len(eligible)
            if eligible else None}


def human_fpr(rows: list[dict | None], name: str, point: float) -> dict:
    eligible = [row for row in rows if row is not None]
    return {"total": len(rows), "eligible": len(eligible),
            "flagged": sum(row[name] > point for row in eligible),
            "fpr": sum(row[name] > point for row in eligible) / len(eligible)
            if eligible else None}


def run(arb_path: Path, academic_path: Path, preflight: bool) -> dict:
    digests = identities(arb_path, academic_path)
    arb = load_module(ARB_LOADER, "unswell_arb_loader")
    humans, groups = arb.load_rows(arb_path)
    roles, split_digest = split_arb(humans)
    _, _, _, academic_counts = load_academic(academic_path)
    preflight_result = {"sha256": digests, "split_sha256": split_digest,
                        "arb_roles": dict(collections.Counter(roles.values())),
                        "academic_counts": {f"{k[0]}/{k[1]}": v
                                            for k, v in sorted(academic_counts.items())}}
    if preflight:
        return preflight_result
    freeze = json.loads((HERE / "freeze.json").read_text())
    require(freeze == preflight_result, "Frozen inputs or source roles changed")
    module = load_module(RICHNESS, "unswell_origin_richness")
    closed = set(LEXICON.read_text().splitlines())
    require(len(closed) > 0, "Closed-class lexicon is empty")
    scored = {pair: {key: feature_vector(row["text"], module, closed)
                     for key, row in members.items()}
              for pair, members in groups.items()}
    training_vectors = []
    retained_training_sources = 0
    for pair in sorted(roles):
        if roles[pair] != "train":
            continue
        human = scored[pair][("human", None)]
        machines = [scored[pair][("h2l", generator)] for generator in GENERATORS]
        if human is None or any(machine is None for machine in machines):
            continue
        retained_training_sources += 1
        training_vectors.append((human, 0))
        training_vectors.extend((machine, 1) for machine in machines)
    require(retained_training_sources >= 570, "Too few complete training sources")
    training = [(vector, label,
                 (0.5 if label == 0 else 0.125) / retained_training_sources)
                for vector, label in training_vectors]
    model = fit(training)
    scored_arb = {pair: {key: scores(vector, model) for key, vector in members.items()}
                  for pair, members in scored.items()}
    calibration = [scored_arb[pair][("human", None)] for pair in roles
                   if roles[pair] == "calibration"]
    thresholds = {name: threshold(calibration, name) for name in SCORE_NAMES}
    academic_human, academic_machine, academic_other, _ = load_academic(
        academic_path, lambda text: scores(feature_vector(text, module, closed), model))
    result = {"format": "unswell-interpretable-ensemble-v1",
              "inputs": preflight_result, "feature_names": FEATURE_NAMES,
              "fit": model,
              "training": {"source_groups": retained_training_sources,
                           "excluded_source_groups": 600 - retained_training_sources},
              "thresholds": thresholds, "scores": {}}
    for name in SCORE_NAMES:
        point = thresholds[name]["score"]
        test_pairs = [pair for pair in roles if roles[pair] == "test"]
        test_humans = [scored_arb[pair][("human", None)] for pair in test_pairs]
        arb_result = {"human": human_fpr(test_humans, name, point),
                      "regimes": {}, "h2l_by_domain": {}, "h2l_by_generator": {}}
        for regime in REGIMES:
            pairs = [(scored_arb[pair][("human", None)],
                      scored_arb[pair][(regime, generator)])
                     for pair in test_pairs for generator in GENERATORS]
            arb_result["regimes"][regime] = summarize(pairs, name, point)
        for domain in DOMAINS:
            pairs = [(scored_arb[pair][("human", None)],
                      scored_arb[pair][("h2l", generator)])
                     for pair in test_pairs if humans[pair]["source_dataset"] == domain
                     for generator in GENERATORS]
            arb_result["h2l_by_domain"][domain] = summarize(pairs, name, point)
        for generator in GENERATORS:
            pairs = [(scored_arb[pair][("human", None)],
                      scored_arb[pair][("h2l", generator)])
                     for pair in test_pairs]
            arb_result["h2l_by_generator"][generator] = summarize(pairs, name, point)
        academic_pairs = [(academic_human[paper], machine)
                          for paper, generator, machine in academic_machine]
        academic_result = {
            "test_human": human_fpr(list(academic_human.values()), name, point),
            "human_only": human_fpr(list(academic_other.values()), name, point),
            "ai_rewrites": summarize(academic_pairs, name, point),
            "by_generator": {generator: summarize(
                [(academic_human[paper], machine) for paper, gen, machine
                 in academic_machine if gen == generator], name, point)
                for generator in sorted({gen for _, gen, _ in academic_machine})},
        }
        result["scores"][name] = {"arb": arb_result, "academic": academic_result}
    ensemble = result["scores"]["ensemble"]
    baseline = result["scores"]["richness"]
    a = ensemble["arb"]["regimes"]["h2l"]
    b = ensemble["academic"]["ai_rewrites"]
    result["advance_pass"] = all((
        a["eligible_pairs"] / a["raw_pairs"] >= 0.95,
        a["auroc"] >= 0.70, a["tpr"] >= 0.10,
        ensemble["arb"]["human"]["fpr"] <= 0.02,
        a["auroc"] >= baseline["arb"]["regimes"]["h2l"]["auroc"],
        a["tpr"] >= baseline["arb"]["regimes"]["h2l"]["tpr"],
        b["eligible_pairs"] / b["raw_pairs"] >= 0.95,
        b["auroc"] >= 0.70, b["tpr"] >= 0.10,
        ensemble["academic"]["test_human"]["fpr"] <= 0.02,
        ensemble["academic"]["human_only"]["fpr"] <= 0.02,
    ))
    return result


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--arb", type=Path, required=True)
    parser.add_argument("--academic", type=Path, required=True)
    parser.add_argument("--preflight", action="store_true")
    parser.add_argument("--output", type=Path, default=HERE / "result.json")
    args = parser.parse_args()
    result = run(args.arb, args.academic, args.preflight)
    body = json.dumps(result, indent=2, sort_keys=True) + "\n"
    if args.preflight:
        print(body, end="")
    else:
        args.output.write_text(body)
        print(f"Wrote {args.output}; advance_pass={result['advance_pass']}")


if __name__ == "__main__":
    main()
