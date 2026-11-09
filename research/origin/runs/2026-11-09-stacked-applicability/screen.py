#!/usr/bin/env python3
"""Check paper-style span-mask activation on frozen calibration sources only."""

from __future__ import annotations

import collections
import hashlib
import importlib.util
import json
import math
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]
ENSEMBLE = ROOT / "research/origin/runs/2026-11-08-interpretable-ensemble"
RICHNESS = ROOT / "research/origin/runs/2026-11-07-richness-screen/run.py"
LEXICON = ROOT / "research/origin/runs/2026-11-08-gen-edit-transfer/closed_class.txt"
ARB_LOADER = ROOT / "research/origin/runs/2026-11-08-arb-transfer/run.py"
ARB = Path("/private/tmp/unswell-arb-20261108/arb.parquet")
ACADEMIC = Path("/private/tmp/unswell-aiwritingeffect-20261108/docs.jsonl")
MODEL = ENSEMBLE / "result.json"
WINDOW = 100
MAX_WINDOWS = 4
CONFIDENCE = 0.01
CAP = 0.25
ADVANCE_FRACTION = 0.05
GENERATORS = ("gemma2_9b", "llama32_3b", "mistral7b", "qwen25_7b")


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


def mask_count(logits: list[float]) -> tuple[int, int]:
    cap = min(math.ceil(CAP * len(logits)), max(0, len(logits) - 1))
    cutoff = math.log(CONFIDENCE / (1 - CONFIDENCE))
    confident = sum(logit < cutoff for logit in logits)
    return confident, min(cap, confident)


def screen_text(text: str, richness, ensemble, closed: set[str], fit: dict) -> dict:
    tokens = [match.group().casefold()
              for match in richness.WORD.finditer(richness.prose(text))]
    windows = min(len(tokens) // WINDOW, MAX_WINDOWS)
    logits = []
    for index in range(windows):
        chunk = " ".join(tokens[index * WINDOW:(index + 1) * WINDOW])
        vector = ensemble.feature_vector(chunk, richness, closed)
        require(vector is not None, "Full window lost tokens")
        logits.append(ensemble.scores(vector, fit)["ensemble"])
    eligible, filtered = mask_count(logits)
    return {"eligible": windows > 0, "windows": windows,
            "confident_windows": eligible, "filtered": filtered}


def summarize(rows: list[dict]) -> dict:
    eligible = [row for row in rows if row["eligible"]]
    return {"total": len(rows), "eligible": len(eligible),
            "abstained": len(rows) - len(eligible),
            "filtering_documents": sum(row["filtered"] > 0 for row in eligible),
            "filtered_windows": sum(row["filtered"] for row in eligible),
            "confident_windows": sum(row["confident_windows"] for row in eligible),
            "total_windows": sum(row["windows"] for row in eligible)}


def advance_decision(arb: dict, academic: dict) -> bool:
    return all((cohort["eligible"] > 0 and
                cohort["filtering_documents"] / cohort["eligible"] >=
                ADVANCE_FRACTION)
               for cohort in (arb, academic))


def run() -> dict:
    freeze = json.loads((HERE / "freeze.json").read_text())
    paths = {"protocol": HERE / "PROTOCOL.md", "script": Path(__file__),
             "test": HERE / "test_screen.py",
             "ensemble_script": ENSEMBLE / "run.py", "model": MODEL,
             "richness": RICHNESS, "lexicon": LEXICON,
             "arb_loader": ARB_LOADER, "arb": ARB, "academic": ACADEMIC}
    for name, path in paths.items():
        require(sha(path) == freeze["sha256"][name], f"{name} hash changed")
    ensemble = load_module(ENSEMBLE / "run.py", "frozen_ensemble")
    richness = load_module(RICHNESS, "frozen_richness")
    arb_loader = load_module(ARB_LOADER, "frozen_arb_loader")
    fit = json.loads(MODEL.read_text())["fit"]
    closed = set(LEXICON.read_text().splitlines())
    humans, groups = arb_loader.load_rows(ARB)
    roles, split_digest = ensemble.split_arb(humans)
    require(split_digest == freeze["split_sha256"], "ARB source roles changed")
    arb_human = []
    arb_ai = []
    for pair in sorted(pair for pair in roles if roles[pair] == "calibration"):
        members = groups[pair]
        arb_human.append(screen_text(members[("human", None)]["text"],
                                     richness, ensemble, closed, fit))
        for generator in GENERATORS:
            arb_ai.append(screen_text(members[("h2l", generator)]["text"],
                                      richness, ensemble, closed, fit))
    require(len(arb_human) == 300 and len(arb_ai) == 1200,
            "ARB calibration denominator changed")
    academic_human = []
    academic_ai = []
    counts = collections.Counter()
    seen = set()
    for line in ACADEMIC.open(encoding="utf-8"):
        row = json.loads(line)
        if row["split"] != "val":
            continue
        key = (row["paper_id"], row["variant"])
        require(key not in seen, "Duplicate academic validation variant")
        seen.add(key)
        target = academic_human if row["variant"] == "human" else academic_ai
        target.append(screen_text(row["text"], richness, ensemble, closed, fit))
        counts["human" if row["variant"] == "human" else "ai"] += 1
    require(counts == {"human": 83, "ai": 131},
            "Academic validation denominator changed")
    cohorts = {"arb_calibration_human": summarize(arb_human),
               "arb_calibration_ai_rewrite": summarize(arb_ai),
               "academic_validation_human": summarize(academic_human),
               "academic_validation_ai_rewrite": summarize(academic_ai)}
    ai_arb = cohorts["arb_calibration_ai_rewrite"]
    ai_academic = cohorts["academic_validation_ai_rewrite"]
    advance = advance_decision(ai_arb, ai_academic)
    return {"format": "unswell-stacked-applicability-v1",
            "source_main": freeze["source_main"], "sha256": freeze["sha256"],
            "split_sha256": split_digest,
            "parameters": {"window_tokens": WINDOW, "max_windows": MAX_WINDOWS,
                           "confidence": CONFIDENCE, "max_filter_fraction": CAP,
                           "advance_fraction": ADVANCE_FRACTION},
            "cohorts": cohorts, "advance": advance}


if __name__ == "__main__":
    result = run()
    (HERE / "result.json").write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    print(json.dumps({"advance": result["advance"], "cohorts": result["cohorts"]},
                     indent=2, sort_keys=True))
