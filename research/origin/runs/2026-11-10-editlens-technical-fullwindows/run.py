#!/usr/bin/env python3
"""Score complete technical-document windows with the frozen EditLens model."""

from __future__ import annotations

import argparse
import collections
import hashlib
import importlib.util
import json
import statistics
from pathlib import Path

from tokenizers import Tokenizer


HERE = Path(__file__).resolve().parent
RUNS = HERE.parent
RICHNESS = RUNS / "2026-11-07-richness-screen" / "run.py"
ORIGIN = RUNS / "2026-11-09-editlens-transfer" / "run.py"
TECHNICAL = RUNS / "2026-11-09-editlens-technical-prefix" / "run.py"
AUDIT = RUNS / "2026-11-10-editlens-technical-windows" / "audit.py"
INVENTORY = AUDIT.with_name("result.json")
STUDY = HERE.parents[2] / "generation" / "studies" / "long-prose-v3"
MODEL_SHA256 = "ddd1173f2ef517ad499965e5029fae8099a8054a2bc76d8134e5889cc4ed1b3e"
DOCUMENT_IDENTITY = "de1ae6afa3304c34bad98ed41d7588bef2e094b5b69fb732cad0e57dd7797d3b"
WIDTH = 100
MAX_TOKENS = 512
FAMILIES = ("haiku", "luna", "qwen")
OPERATIONS = ("generate", "polish")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def sha(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def load(path: Path, name: str):
    spec = importlib.util.spec_from_file_location(name, path)
    require(spec is not None and spec.loader is not None, f"Missing helper: {name}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def doc_key(row: dict) -> str:
    if row["kind"] == "source":
        return "source/" + row["task_id"]
    require(row["kind"] == "response", "Unknown document kind")
    return "/".join(("response", row["task_id"], row["family"],
                     row["operation"], row["prompt"]))


def prepare(docs: list[dict], inventory: dict, tokenizer: Tokenizer,
            richness, origin, audit) -> tuple[list[dict], list[dict], str]:
    expected = {doc_key(row): row for row in inventory["rows"]}
    require(len(expected) == 494, "Duplicate inventory document")
    prepared = []
    windows = []
    selection = []
    for doc in docs:
        key = doc_key(doc)
        require(key in expected, f"Missing inventory document: {key}")
        prior = expected[key]
        require(doc["text_sha256"] == prior["text_sha256"], "Source text changed")
        prose = richness.prose(doc["text"])
        cleaned, removed_header = origin.clean_text(prose)
        words = len(origin.WORD.findall(cleaned))
        parts = audit.word_windows(cleaned, origin.WORD, WIDTH)
        require(parts and "".join(parts) == cleaned, "Incomplete cleaned prose")
        token_ids = [tokenizer.encode(part).ids for part in parts]
        lengths = [len(ids) for ids in token_ids]
        require(all(0 < length <= MAX_TOKENS for length in lengths),
                f"Over-cap or empty window: {key}")
        require(words == prior["cleaned_words"] and
                removed_header == prior["removed_header"] and
                len(parts) == prior["full_coverage"][str(WIDTH)]["windows"] and
                prior["full_coverage"][str(WIDTH)]["complete"],
                f"Inventory changed: {key}")
        row = {name: value for name, value in doc.items() if name != "text"}
        row.update(key=key, cleaned_words=words, windows=len(parts),
                   removed_header=removed_header, score=None)
        prepared.append(row)
        selection.append((key, doc["text_sha256"], lengths))
        for index, ids in enumerate(token_ids):
            windows.append({"key": key, "index": index, "tokens": len(ids),
                            "ids": ids, "score": None})
    require(len(prepared) == 494 and len({row["key"] for row in prepared}) == 494,
            "Incomplete document preparation")
    require(set(expected) == {row["key"] for row in prepared},
            "Inventory contains different documents")
    require(len(windows) == 6484, "Window count changed")
    canonical = json.dumps(selection, ensure_ascii=False, separators=(",", ":"))
    return prepared, windows, hashlib.sha256(canonical.encode()).hexdigest()


def stats(rows: list[dict]) -> dict:
    if not rows:
        return {"pairs": 0, "repositories": 0, "median_delta": None,
                "paired_win_fraction": None, "positive_repositories": 0,
                "negative_repositories": 0, "zero_repositories": 0}
    grouped = collections.defaultdict(list)
    for row in rows:
        grouped[row["repository"]].append(row["delta"])
    means = [statistics.mean(values) for values in grouped.values()]
    return {"pairs": len(rows), "repositories": len(grouped),
            "median_delta": statistics.median(row["delta"] for row in rows),
            "paired_win_fraction": sum((row["delta"] > 0) +
                                       (row["delta"] == 0) / 2 for row in rows) / len(rows),
            "positive_repositories": sum(mean > 0 for mean in means),
            "negative_repositories": sum(mean < 0 for mean in means),
            "zero_repositories": sum(mean == 0 for mean in means)}


def summarize(docs: list[dict], windows: list[dict], identity: str,
              selection_hash: str) -> tuple[dict, list[dict]]:
    by_key = collections.defaultdict(list)
    for window in windows:
        require(window["score"] is not None, "Unscored window")
        by_key[window["key"]].append(window)
    for doc in docs:
        ordered = sorted(by_key[doc["key"]], key=lambda row: row["index"])
        require([row["index"] for row in ordered] == list(range(doc["windows"])),
                "Missing or repeated window")
        doc["score"] = statistics.mean(row["score"] for row in ordered)
    sources = {doc["task_id"]: doc for doc in docs if doc["kind"] == "source"}
    responses = [doc for doc in docs if doc["kind"] == "response"]
    require(len(sources) == 38 and len(responses) == 456, "Document count changed")
    pairs = []
    for doc in responses:
        source = sources[doc["task_id"]]
        ratio = doc["cleaned_words"] / source["cleaned_words"]
        pairs.append({"key": doc["key"], "source_key": source["key"],
                      "repository": doc["repository"], "family": doc["family"],
                      "operation": doc["operation"], "prompt": doc["prompt"],
                      "source_score": source["score"], "response_score": doc["score"],
                      "delta": doc["score"] - source["score"],
                      "word_ratio": ratio, "source_windows": source["windows"],
                      "response_windows": doc["windows"]})
    result = {"format": "unswell-editlens-technical-fullwindows-v1",
              "document_identity_sha256": identity,
              "selection_sha256": selection_hash,
              "model_sha256": MODEL_SHA256,
              "window_width_words": WIDTH,
              "sources": len(sources), "responses": len(responses),
              "source_windows": sum(doc["windows"] for doc in sources.values()),
              "response_windows": sum(doc["windows"] for doc in responses),
              "operations": {}}
    for operation in OPERATIONS:
        subset = [pair for pair in pairs if pair["operation"] == operation]
        family_stats = {family: stats([pair for pair in subset if pair["family"] == family])
                        for family in FAMILIES}
        overall = stats(subset)
        result["operations"][operation] = {
            "overall": overall,
            "families": family_stats,
            "matched_length_band": stats([pair for pair in subset
                                          if 0.75 <= pair["word_ratio"] <= 1.33]),
            "median_word_ratio": statistics.median(pair["word_ratio"] for pair in subset),
            "median_source_windows": statistics.median(pair["source_windows"] for pair in subset),
            "median_response_windows": statistics.median(pair["response_windows"] for pair in subset),
            "direction_pass": (overall["pairs"] >= 217 and
                               overall["positive_repositories"] >= 17 and
                               overall["paired_win_fraction"] >= 0.70 and
                               all(row["median_delta"] > 0 for row in family_stats.values())),
        }
    return result, pairs


def verify_freeze(freeze: dict, model_dir: Path, identity: str,
                  selection_hash: str) -> Path:
    for name, path in (("protocol_sha256", HERE / "PROTOCOL.md"),
                       ("runner_sha256", HERE / "run.py"),
                       ("verifier_sha256", HERE / "verify.py"),
                       ("runner_tests_sha256", HERE / "test_run.py"),
                       ("richness_runner_sha256", RICHNESS),
                       ("origin_runner_sha256", ORIGIN),
                       ("technical_runner_sha256", TECHNICAL),
                       ("audit_runner_sha256", AUDIT),
                       ("inventory_sha256", INVENTORY),
                       ("input_archive_sha256", STUDY / "inputs.tar.gz"),
                       ("output_archive_sha256", STUDY / "outputs.tar.gz")):
        require(sha(path) == freeze[name], f"{name} changed")
    require(identity == DOCUMENT_IDENTITY == freeze["document_identity_sha256"],
            "Document identity changed")
    require(selection_hash == freeze["selection_sha256"], "Window selection changed")
    for filename, expected in freeze["model_file_sha256"].items():
        require(sha(model_dir / filename) == expected, f"Modelkit file changed: {filename}")
    model = model_dir / "onnx/model.onnx"
    require(sha(model) == MODEL_SHA256 == freeze["model_sha256"], "Model changed")
    return model


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--model-dir", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path)
    parser.add_argument("--preflight", action="store_true")
    args = parser.parse_args()
    require(args.preflight or args.output_dir is not None, "Output directory required")
    richness = load(RICHNESS, "fullwindow_richness")
    origin = load(ORIGIN, "fullwindow_origin")
    technical = load(TECHNICAL, "fullwindow_technical")
    audit = load(AUDIT, "fullwindow_audit")
    inventory = json.loads(INVENTORY.read_text())
    docs, identity = technical.load_docs(richness)
    tokenizer = Tokenizer.from_file(str(args.model_dir / "tokenizer.json"))
    prepared, windows, selection_hash = prepare(docs, inventory, tokenizer,
                                                richness, origin, audit)
    if args.preflight:
        print(json.dumps({"documents": len(prepared), "windows": len(windows),
                          "document_identity_sha256": identity,
                          "selection_sha256": selection_hash}, sort_keys=True))
        return
    freeze = json.loads((HERE / "freeze.json").read_text())
    model = verify_freeze(freeze, args.model_dir, identity, selection_hash)
    origin.score(windows, model)
    result, pairs = summarize(prepared, windows, identity, selection_hash)
    output = args.output_dir
    output.mkdir(parents=True, exist_ok=True)
    for name, value in (("result.json", result), ("documents.json", prepared),
                        ("pairs.json", pairs), ("window_scores.json", windows)):
        if name == "window_scores.json":
            value = [{key: item for key, item in row.items() if key != "ids"}
                     for row in windows]
        (output / name).write_text(json.dumps(value, sort_keys=True, indent=2) + "\n")
    print(json.dumps(result, sort_keys=True))


if __name__ == "__main__":
    main()
