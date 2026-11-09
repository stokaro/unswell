#!/usr/bin/env python3
"""Run the frozen, source-matched EditLens origin transfer screen."""

from __future__ import annotations

import argparse
import collections
import hashlib
import importlib.util
import json
import math
import re
from pathlib import Path

import duckdb  # Required by the imported, pinned ARB loader.
import emoji
import numpy as np
import onnxruntime as ort
from tokenizers import Tokenizer


HERE = Path(__file__).resolve().parent
MODEL_REVISION = "f7cb4b06e5067ecdb66c566c7982a413f86f569f"
MODEL_HASH = "ddd1173f2ef517ad499965e5029fae8099a8054a2bc76d8134e5889cc4ed1b3e"
DATA_HASH = "66e48240618b92ce876e2097d8df52f06661a70e31791b91a95b15688a099ab0"
GENERATORS = ("gemma2_9b", "llama32_3b", "mistral7b", "qwen25_7b")
DOMAINS = ("owt", "wp", "xsum")
HEADER_STARTS = ("Sure", "Here", "Abstract", "Title", "I'm happy to help", "Certainly")
WORD = re.compile(r"\b\w+\b")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def sha(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def load_arb_module(path: Path):
    spec = importlib.util.spec_from_file_location("pinned_arb_loader", path)
    require(spec is not None and spec.loader is not None, "ARB loader unavailable")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def select_rows(humans: dict, groups: dict) -> tuple[list[dict], str]:
    selected = []
    identities = []
    for domain in DOMAINS:
        ids = [pair for pair, row in humans.items() if row["source_dataset"] == domain]
        require(len(ids) == 600, f"Expected 600 human sources in {domain}")
        ids.sort(key=lambda pair: (
            hashlib.sha256((pair + "unswell-arb-v1").encode()).hexdigest(), pair
        ))
        for pair in ids[:100]:
            row = humans[pair]
            selected.append({"role": "calibration", "regime": "human", "row": row})
            identities.append((domain, pair, "calibration", None, row["text_sha256"]))
        for pair in ids[300:400]:
            index = int(hashlib.sha256((pair + "unswell-editlens-v1").encode()).hexdigest()[:8], 16) % 4
            generator = GENERATORS[index]
            members = groups[pair]
            for regime, gen in (("human", None), ("h2l", generator), ("free_llm", generator)):
                row = members[(regime, gen)]
                selected.append({"role": "test", "regime": regime, "row": row})
                identities.append((domain, pair, "test", regime, generator, row["text_sha256"]))
    require(len(selected) == 1200, "Sample size mismatch")
    canonical = json.dumps(identities, separators=(",", ":"), ensure_ascii=False)
    return selected, hashlib.sha256(canonical.encode()).hexdigest()


def clean_text(text: str) -> tuple[str, bool]:
    text = emoji.demojize(text)
    if "</think>" in text:
        text = text.split("</think>")[1].strip()
    paragraphs = [paragraph for paragraph in text.split("\n") if paragraph.strip()]
    removed_header = False
    if paragraphs:
        first = re.sub(r"^[^a-zA-Z0-9]*", "", paragraphs[0])
        first = emoji.replace_emoji(first, "")
        if any(first.startswith(phrase) for phrase in HEADER_STARTS) and len(paragraphs) > 1:
            text = "\n".join(paragraphs[1:])
            removed_header = True
    return re.sub(r"\s+", " ", text.lower()).strip(), removed_header


def prepare(selected: list[dict], tokenizer: Tokenizer) -> list[dict]:
    prepared = []
    for item in selected:
        cleaned, removed = clean_text(item["row"]["text"])
        word_count = len(WORD.findall(cleaned))
        token_ids = tokenizer.encode(cleaned).ids if word_count >= 50 else []
        prepared.append({
            "role": item["role"],
            "regime": item["regime"],
            "domain": item["row"]["source_dataset"],
            "pair_id": item["row"]["pair_id"],
            "generator": item["row"]["generator_model"],
            "text_sha256": item["row"]["text_sha256"],
            "words": word_count,
            "token_count": len(token_ids),
            "truncated": len(token_ids) > 512,
            "removed_header": removed,
            "ids": token_ids[:512],
            "score": None,
        })
    return prepared


def score(prepared: list[dict], model_path: Path) -> None:
    options = ort.SessionOptions()
    options.intra_op_num_threads = 4
    options.inter_op_num_threads = 1
    session = ort.InferenceSession(str(model_path), sess_options=options,
                                   providers=["CPUExecutionProvider"])
    require(session.get_providers() == ["CPUExecutionProvider"], "Unexpected model provider")
    eligible = [item for item in prepared if item["ids"]]
    eligible.sort(key=lambda item: len(item["ids"]))
    for at in range(0, len(eligible), 8):
        batch = eligible[at:at + 8]
        width = max(len(item["ids"]) for item in batch)
        input_ids = np.full((len(batch), width), 1, dtype=np.int64)
        attention = np.zeros((len(batch), width), dtype=np.int64)
        for index, item in enumerate(batch):
            length = len(item["ids"])
            input_ids[index, :length] = item["ids"]
            attention[index, :length] = 1
        logits = session.run(["logits"], {"input_ids": input_ids,
                                           "attention_mask": attention})[0]
        require(logits.shape == (len(batch), 4) and np.isfinite(logits).all(),
                "Invalid model output")
        shifted = logits - logits.max(axis=1, keepdims=True)
        probabilities = np.exp(shifted) / np.exp(shifted).sum(axis=1, keepdims=True)
        values = (probabilities @ np.arange(4)) / 3
        for item, value in zip(batch, values, strict=True):
            item["score"] = float(value)
        if at % 80 == 0:
            print(f"scored {min(at + len(batch), len(eligible))}/{len(eligible)}", flush=True)


def auc(pairs: list[tuple[float, float]]) -> float | None:
    if len(pairs) < 20:
        return None
    values = [(machine, 1) for human, machine in pairs]
    values += [(human, 0) for human, machine in pairs]
    values.sort()
    negative_seen = 0
    wins = 0.0
    at = 0
    while at < len(values):
        end = at + 1
        while end < len(values) and values[end][0] == values[at][0]:
            end += 1
        positives = sum(label for _, label in values[at:end])
        negatives = end - at - positives
        wins += positives * negative_seen + positives * negatives / 2
        negative_seen += negatives
        at = end
    return wins / (len(pairs) ** 2)


def summarize(prepared: list[dict]) -> dict:
    by_role = collections.defaultdict(list)
    for item in prepared:
        by_role[(item["role"], item["regime"])].append(item)
    cal = [item["score"] for item in by_role[("calibration", "human")]
           if item["score"] is not None]
    require(len(cal) >= 285, "Insufficient calibration humans")
    threshold = sorted(cal, reverse=True)[2]
    humans = {item["pair_id"]: item for item in by_role[("test", "human")]}
    output = {
        "threshold": threshold,
        "calibration": {"raw": 300, "eligible": len(cal),
                        "flagged": sum(value > threshold for value in cal)},
        "test_human": {},
        "regimes": {},
        "counts": {},
    }
    for key, rows in by_role.items():
        output["counts"]["/".join(key)] = {
            "raw": len(rows),
            "eligible": sum(item["score"] is not None for item in rows),
            "truncated": sum(item["truncated"] for item in rows),
            "removed_header": sum(item["removed_header"] for item in rows),
            "median_words": float(np.median([item["words"] for item in rows])),
        }
    test_human = [item["score"] for item in humans.values() if item["score"] is not None]
    output["test_human"] = {
        "raw": 300, "eligible": len(test_human),
        "flagged": sum(value > threshold for value in test_human),
        "fpr": sum(value > threshold for value in test_human) / len(test_human),
    }
    for regime in ("h2l", "free_llm"):
        rows = by_role[("test", regime)]
        paired = [(humans[item["pair_id"]], item) for item in rows
                  if item["score"] is not None and
                  humans[item["pair_id"]]["score"] is not None]

        def metrics(group: list[tuple[dict, dict]]) -> dict:
            pairs = [(human["score"], machine["score"]) for human, machine in group]
            return {
                "eligible_pairs": len(pairs),
                "auc": auc(pairs),
                "paired_win_fraction": sum((machine > human) + (machine == human) / 2
                                           for human, machine in pairs) / len(pairs) if pairs else None,
                "flagged": sum(machine > threshold for _, machine in pairs),
                "tpr": sum(machine > threshold for _, machine in pairs) / len(pairs)
                       if pairs else None,
            }

        output["regimes"][regime] = {
            "overall": metrics(paired),
            "by_domain": {domain: metrics([(h, m) for h, m in paired if m["domain"] == domain])
                          for domain in DOMAINS},
            "by_generator": {generator: metrics([(h, m) for h, m in paired
                                                  if m["generator"] == generator])
                             for generator in GENERATORS},
        }
    h2l = output["regimes"]["h2l"]["overall"]
    free = output["regimes"]["free_llm"]["overall"]
    output["feasibility_pass"] = all((
        len(cal) >= 285,
        len(test_human) >= 285,
        h2l["eligible_pairs"] >= 285,
        free["eligible_pairs"] >= 285,
        output["test_human"]["fpr"] <= 0.02,
        h2l["auc"] is not None and h2l["auc"] >= 0.75,
        h2l["tpr"] is not None and h2l["tpr"] >= 0.20,
        free["auc"] is not None and free["auc"] >= 0.85,
        free["tpr"] is not None and free["tpr"] >= 0.50,
    ))
    return output


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--data", type=Path, required=True)
    parser.add_argument("--model-dir", type=Path, required=True)
    parser.add_argument("--arb-loader", type=Path, required=True)
    parser.add_argument("--preflight", action="store_true")
    args = parser.parse_args()
    require(sha(args.data) == DATA_HASH, "ARB dataset hash mismatch")
    arb = load_arb_module(args.arb_loader)
    humans, groups = arb.load_rows(args.data)
    selected, selection_hash = select_rows(humans, groups)
    if args.preflight:
        print(json.dumps({"selected": len(selected), "selection_sha256": selection_hash},
                         sort_keys=True))
        return
    freeze = json.loads((HERE / "freeze.json").read_text())
    require(sha(HERE / "PROTOCOL.md") == freeze["protocol_sha256"], "Protocol changed")
    require(sha(HERE / "run.py") == freeze["runner_sha256"], "Runner changed")
    require(sha(args.arb_loader) == freeze["arb_loader_sha256"], "ARB loader changed")
    require(selection_hash == freeze["selection_sha256"], "Selection changed")
    model_path = args.model_dir / "onnx/model.onnx"
    require(sha(model_path) == MODEL_HASH, "Model hash mismatch")
    for filename, digest in freeze["model_file_sha256"].items():
        require(sha(args.model_dir / filename) == digest, f"{filename} hash mismatch")
    tokenizer = Tokenizer.from_file(str(args.model_dir / "tokenizer.json"))
    prepared = prepare(selected, tokenizer)
    score(prepared, model_path)
    result = summarize(prepared)
    result.update({"selection_sha256": selection_hash,
                   "model_revision": MODEL_REVISION,
                   "model_sha256": MODEL_HASH,
                   "dataset_sha256": DATA_HASH})
    (HERE / "result.json").write_text(json.dumps(result, sort_keys=True, indent=2) + "\n")
    rows = [{key: value for key, value in item.items() if key != "ids"}
            for item in prepared]
    (HERE / "scores.json").write_text(json.dumps(rows, sort_keys=True, indent=2) + "\n")
    print(json.dumps({"feasibility_pass": result["feasibility_pass"],
                      "h2l": result["regimes"]["h2l"]["overall"],
                      "free_llm": result["regimes"]["free_llm"]["overall"],
                      "test_human": result["test_human"]}, sort_keys=True))


if __name__ == "__main__":
    main()
