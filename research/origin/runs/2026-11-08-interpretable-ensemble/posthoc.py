#!/usr/bin/env python3
"""Explain the frozen origin-score transfer failure without fitting a new score."""

from __future__ import annotations

import argparse
import importlib.util
import json
import statistics
from pathlib import Path


HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("frozen_ensemble", HERE / "run.py")
assert SPEC is not None and SPEC.loader is not None
study = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(study)


def describe(vectors: list[tuple[float, ...]], model: dict) -> dict:
    study.require(vectors and all(vector is not None for vector in vectors),
                  "Source feature abstention in post-result audit")
    values = {}
    for index, name in enumerate(study.FEATURE_NAMES):
        mean = statistics.mean(vector[index] for vector in vectors)
        values[name] = {
            "mean": mean,
            "mean_logit_contribution": model["weights"][index] *
            (mean - model["means"][index]) / model["deviations"][index],
        }
    return {"sources": len(vectors), "features": values}


def audit(arb_path: Path, academic_path: Path) -> dict:
    freeze = json.loads((HERE / "freeze.json").read_text())
    study.require(study.identities(arb_path, academic_path) == freeze["sha256"],
                  "Frozen input identity changed")
    saved = json.loads((HERE / "result.json").read_text())
    study.require(saved["inputs"] == freeze, "Frozen score identity changed")
    arb = study.load_module(study.ARB_LOADER, "posthoc_arb")
    humans, _ = arb.load_rows(arb_path)
    roles, digest = study.split_arb(humans)
    study.require(digest == freeze["split_sha256"], "Source split changed")
    module = study.load_module(study.RICHNESS, "posthoc_richness")
    closed = set(study.LEXICON.read_text().splitlines())
    score = lambda text: study.feature_vector(text, module, closed)
    arb_test = [score(humans[pair]["text"]) for pair in roles
                if roles[pair] == "test"]
    academic_test, _, academic_other, _ = study.load_academic(academic_path, score)
    return {
        "format": "unswell-interpretable-ensemble-posthoc-v1",
        "interpretation": "Exposed post-result feature attribution; no refit or new threshold",
        "input_sha256": freeze["sha256"],
        "main_result_sha256": study.sha(HERE / "result.json"),
        "script_sha256": study.sha(Path(__file__)),
        "groups": {
            "arb_test_human": describe(arb_test, saved["fit"]),
            "academic_test_human": describe(list(academic_test.values()), saved["fit"]),
            "academic_human_only": describe(list(academic_other.values()), saved["fit"]),
        },
    }


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--arb", type=Path, required=True)
    parser.add_argument("--academic", type=Path, required=True)
    parser.add_argument("--output", type=Path, default=HERE / "posthoc-result.json")
    args = parser.parse_args()
    args.output.write_text(json.dumps(audit(args.arb, args.academic),
                                      indent=2, sort_keys=True) + "\n")
    print(f"Wrote {args.output}")


if __name__ == "__main__":
    main()
