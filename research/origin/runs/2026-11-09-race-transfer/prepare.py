#!/usr/bin/env python3
"""Select source-matched ARB documents for the frozen RACE feasibility screen."""

from __future__ import annotations

import argparse
import hashlib
import importlib.util
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]
ARB_LOADER = ROOT / "research/origin/runs/2026-11-08-arb-transfer/run.py"
ENSEMBLE = ROOT / "research/origin/runs/2026-11-08-interpretable-ensemble/run.py"
ARB_SHA256 = "66e48240618b92ce876e2097d8df52f06661a70e31791b91a95b15688a099ab0"
ARB_BYTES = 27_439_482
DOMAINS = ("owt", "wp", "xsum")
GENERATORS = ("gemma2_9b", "llama32_3b", "mistral7b", "qwen25_7b")
SALT = "unswell-race-v1|"


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def sha_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def verify_arb(path: Path) -> None:
    require(path.stat().st_size == ARB_BYTES, "ARB byte count changed")
    require(sha_file(path) == ARB_SHA256, "ARB SHA-256 changed")


def load_module(path: Path, name: str):
    spec = importlib.util.spec_from_file_location(name, path)
    require(spec is not None and spec.loader is not None, f"Cannot load {name}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def select_pairs(humans: dict, roles: dict[str, str]) -> list[str]:
    selected = []
    for domain in DOMAINS:
        ids = [pair for pair, row in humans.items()
               if row["source_dataset"] == domain and roles[pair] == "test"]
        require(len(ids) == 300, f"Expected 300 test sources in {domain}")
        ids.sort(key=lambda pair: (
            hashlib.sha256((SALT + pair).encode()).hexdigest(), pair
        ))
        selected.extend(ids[:20])
    require(len(selected) == 60 and len(set(selected)) == 60,
            "RACE source selection changed")
    return selected


def records(selected: list[str], groups: dict) -> list[dict]:
    output = []
    for pair in selected:
        members = groups[pair]
        human = members["human", None]
        variants = [("human", None, human)]
        variants += [("h2l", model, members["h2l", model])
                     for model in GENERATORS]
        for regime, model, row in variants:
            output.append({
                "item_id": f"{pair}:{regime}:{model or 'source'}",
                "group_id": f"{row['source_dataset']}-{pair}",
                "source_dataset": row["source_dataset"],
                "pair_id": pair,
                "regime": regime,
                "model_name": model,
                "label": ("human_written" if regime == "human"
                          else "human_ai_polished"),
                "article": row["text"],
                "text_sha256": hashlib.sha256(row["text"].encode("utf-8")).hexdigest(),
                "arb_text_sha256": row["text_sha256"],
            })
    require(len(output) == 300, "RACE document count changed")
    return output


def prepare(arb: Path, output_dir: Path) -> dict:
    verify_arb(arb)
    loader = load_module(ARB_LOADER, "unswell_race_arb_loader")
    ensemble = load_module(ENSEMBLE, "unswell_race_ensemble")
    humans, groups = loader.load_rows(arb)
    roles, split_sha = ensemble.split_arb(humans)
    require(split_sha == "63dc835b4586260da00a4df7a14df8af60c273df2f8bb7e6596a9f09ddc1683e",
            "ARB source role split changed")
    selected = select_pairs(humans, roles)
    rows = records(selected, groups)
    output_dir.mkdir(parents=True, exist_ok=True)
    input_path = output_dir / "input.jsonl"
    with input_path.open("w", encoding="utf-8") as stream:
        for row in rows:
            stream.write(json.dumps(row, ensure_ascii=False, separators=(",", ":")) + "\n")
    selection = {
        "arb_sha256": ARB_SHA256,
        "source_split_sha256": split_sha,
        "source_ids": selected,
        "document_hashes": {row["item_id"]: row["text_sha256"] for row in rows},
        "input_sha256": sha_file(input_path),
        "count": len(rows),
    }
    (output_dir / "selection.json").write_text(
        json.dumps(selection, indent=2, sort_keys=True) + "\n", encoding="utf-8"
    )
    return selection


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--arb", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    selection = prepare(args.arb, args.output)
    print(json.dumps({"count": selection["count"],
                      "input_sha256": selection["input_sha256"]}))


if __name__ == "__main__":
    main()
