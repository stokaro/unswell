#!/usr/bin/env python3
"""Run one frozen offline semantic admission screen for issue 349."""

from __future__ import annotations

import argparse
import hashlib
import importlib.util
import json
import math
import os
from pathlib import Path
import re

from ape import APE
import numpy as np
import sklearn
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import roc_auc_score
from sklearn.preprocessing import OneHotEncoder
import torch


HERE = Path(__file__).resolve().parent
CONTROL = HERE.parent / "2026-11-09-structured-admission" / "scan-v2.py"
CONTROL_SHA256 = "20a770ce8a8e88dcf2b5ac17408bf4f3006c5179a61bbb7e510eae61c994c063"
CONTROL_RESULT_SHA256 = "df663463350602d3592fc930caa89fa0589a073d3e3e00f4b9983f61cff38765"
MODEL_CONFIG_SHA256 = "327aec9064e20feb69b71ef975d2a80c3beee41aa42c499c00f26a1df644b97a"
MODEL_WEIGHTS_SHA256 = "ad1992f37a1327f9e6cd73dcbf7652660bf83d47694258e1b06bda8bba6155eb"
TOKENIZER_SHA256 = "d241a60d5e8f04cc1b2b3e9ef7a4921b27bf526d9f6050ab90f9267a1f9e5c66"


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def load_control():
    require(digest(CONTROL) == CONTROL_SHA256, "control script changed")
    require(digest(CONTROL.with_name("result-v2.json")) == CONTROL_RESULT_SHA256,
            "control result changed")
    spec = importlib.util.spec_from_file_location("frozen_structured_admission", CONTROL)
    require(spec is not None and spec.loader is not None, "control import failed")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def source_lines(raw: bytes, start: int, end: int) -> str:
    line_start = raw.rfind(b"\n", 0, start) + 1
    line_end = raw.find(b"\n", end)
    if line_end < 0:
        line_end = len(raw)
    previous_end = max(0, line_start - 1)
    previous_start = raw.rfind(b"\n", 0, previous_end) + 1
    next_start = min(len(raw), line_end + 1)
    next_end = raw.find(b"\n", next_start)
    if next_end < 0:
        next_end = len(raw)
    headings = re.findall(r"(?m)^#{1,6}\s+(.+)$",
                          raw[max(0, start - 2000):start].decode("utf-8", "replace"))
    parts = [raw[line_start:line_end].decode("utf-8", "replace")[:500]]
    if headings:
        parts.append("Heading: " + headings[-1][:120])
    parts.append("Before: " + raw[previous_start:previous_end].decode("utf-8", "replace")[:180])
    parts.append("After: " + raw[next_start:next_end].decode("utf-8", "replace")[:180])
    return "\n".join(parts)


def texts(data: dict, aliases: list[str]) -> tuple[list[str], list[str], list[str], list[list[str]]]:
    sources = {}
    for page, record in data["originals"]["original"].items():
        document = record["document"]
        require(document["page"] == page, "original page mismatch")
        raw = document["source"].encode("utf-8")
        require(hashlib.sha256(raw).hexdigest() == document["source_sha256"],
                "original source hash mismatch")
        sources[page] = raw
    require(len(sources) == 36, "source inventory changed")
    candidates, contexts, retentions, categories = [], [], [], []
    for alias in aliases:
        meta = data["metadata"][alias]
        item = data["cores"][meta["stream"]][meta["finding_id"]]
        raw = sources[meta["page"]]
        targets = sorted(item["targets"], key=lambda target:
                         (target["span"]["start"], target["span"]["end"]))
        require(targets, "candidate has no targets")
        source_parts, quotes = [], []
        for target in targets:
            start, end = target["span"]["start"], target["span"]["end"]
            require(0 <= start < end <= len(raw), "target outside original source")
            require(raw[start:end].decode("utf-8") == target["quote"],
                    "target quote differs from source bytes")
            if len(source_parts) < 2:
                source_parts.append(source_lines(raw, start, end))
                quotes.append(target["quote"])
        candidates.append("\n".join(["Target: " + " | ".join(quotes),
                                     "Diagnosis: " + item["diagnostic"],
                                     "Revision need: " + item.get("why_revision_needed", ""),
                                     "Wording burden: " + item.get("wording_burden", "")]))
        contexts.append("\n".join(source_parts))
        retentions.append("\n".join(["Retention: " + item.get("retention_case", ""),
                                      "Meaning: " + item.get("meaning_preservation", "")]))
        categories.append([item.get("category", "unknown"),
                           item.get("operator", "unknown"), meta["stream"]])
    return candidates, contexts, retentions, categories


def embed(model: APE, values: list[str]) -> np.ndarray:
    result = []
    with torch.inference_mode():
        for start in range(0, len(values), 32):
            encoded = model.tokenizer(values[start:start + 32], padding=True,
                                      truncation=True, max_length=128,
                                      return_tensors="pt")
            states = model.encoder(**encoded).last_hidden_state
            mask = encoded["attention_mask"].unsqueeze(-1)
            pooled = (states * mask).sum(dim=1) / mask.sum(dim=1)
            result.append(torch.nn.functional.normalize(pooled, dim=1).cpu().numpy())
    return np.vstack(result).astype(np.float64)


def heldout_scores(rows: dict, aliases: list[str], vectors: np.ndarray,
                   categories: list[list[str]]) -> tuple[dict, dict]:
    positions = {alias: index for index, alias in enumerate(aliases)}
    groups = sorted({row["group"] for row in rows.values() if row["cohort"] == "whole"})
    require(len(groups) == 12, "whole repository groups changed")
    scores, folds = {}, {}
    for excluded in groups:
        train = sorted(a for a in aliases if rows[a]["cohort"] == "whole" and
                       rows[a]["group"] != excluded)
        test = sorted(a for a in aliases if rows[a]["cohort"] == "whole" and
                      rows[a]["group"] == excluded)
        train_positions = [positions[a] for a in train]
        test_positions = [positions[a] for a in test]
        onehot = OneHotEncoder(handle_unknown="ignore", sparse_output=False)
        train_cats = onehot.fit_transform([categories[i] for i in train_positions])
        test_cats = onehot.transform([categories[i] for i in test_positions])
        train_matrix = np.hstack((vectors[train_positions], train_cats))
        test_matrix = np.hstack((vectors[test_positions], test_cats))
        classifier = LogisticRegression(C=0.01, class_weight="balanced",
                                        solver="liblinear", max_iter=1000,
                                        random_state=0)
        classifier.fit(train_matrix, [rows[a]["accepted"] for a in train])
        probabilities = classifier.predict_proba(test_matrix)[:, 1]
        scores.update(zip(test, map(float, probabilities), strict=True))
        folds[excluded] = {"train": len(train), "heldout": len(test)}
    require(len(scores) == 2101, "incomplete whole-page scores")
    return scores, folds


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--development-root", type=Path, required=True)
    parser.add_argument("--model", type=Path, required=True)
    parser.add_argument("--backbone", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    require(not args.output.exists(), "result already exists")
    require(os.environ.get("HF_HUB_OFFLINE") == "1" and
            os.environ.get("TRANSFORMERS_OFFLINE") == "1", "offline mode required")
    freeze = json.loads((HERE / "freeze.json").read_text())
    require(freeze["status"] == "frozen_before_inference" and
            freeze["protocol_sha256"] == digest(HERE / "PROTOCOL.md") and
            freeze["script_sha256"] == digest(Path(__file__)), "freeze mismatch")
    require(digest(args.model / "config.json") == MODEL_CONFIG_SHA256 and
            digest(args.model / "model.safetensors") == MODEL_WEIGHTS_SHA256 and
            digest(args.backbone / "tokenizer.json") == TOKENIZER_SHA256,
            "model asset mismatch")
    config = json.loads((args.model / "config.json").read_text())
    require(config["encoder_name"] == str(args.backbone.resolve()),
            "model backbone mismatch")
    control = load_control()
    data = control.read_inputs(args.development_root)
    rows, witnesses, _ = control.inventory(data)
    baseline = control.measures(set(rows), rows, witnesses)
    require([(baseline[c]["full"], baseline[c]["accepted"], baseline[c]["delivered"])
             for c in control.COHORTS] ==
            [(48, 1627, 2101), (39, 515, 548), (17, 190, 198)],
            "candidate bank baseline changed")
    aliases = sorted(rows)
    candidates, contexts, retentions, categories = texts(data, aliases)
    torch.set_num_threads(4)
    model = APE.from_pretrained(str(args.model)).to("cpu").eval()
    candidate_vectors = embed(model, candidates)
    context_vectors = embed(model, contexts)
    retention_vectors = embed(model, retentions)
    vectors = np.hstack((candidate_vectors, context_vectors, retention_vectors,
                         np.abs(candidate_vectors - context_vectors))) * math.sqrt(384)
    cosine = np.sum(candidate_vectors * context_vectors, axis=1, keepdims=True)
    vectors = np.hstack((vectors, cosine))
    scores, folds = heldout_scores(rows, aliases, vectors, categories)
    whole = {a: row for a, row in rows.items() if row["cohort"] == "whole"}
    outcome = control.screen(rows, witnesses, scores)
    result = {
        "format": "unswell-349-semantic-admission-exploratory-v1",
        "scope": "exposed_offline_method_screen_only",
        "input_sha256": {key: expected for key, (_, expected) in control.INPUTS.items()},
        "control_script_sha256": CONTROL_SHA256,
        "control_result_sha256": CONTROL_RESULT_SHA256,
        "model_config_sha256": MODEL_CONFIG_SHA256,
        "model_weights_sha256": MODEL_WEIGHTS_SHA256,
        "tokenizer_sha256": TOKENIZER_SHA256,
        "protocol_sha256": digest(HERE / "PROTOCOL.md"),
        "script_sha256": digest(Path(__file__)),
        "sklearn_version": sklearn.__version__,
        "numpy_version": np.__version__,
        "torch_version": torch.__version__,
        "folds": folds,
        "baseline": baseline,
        "heldout_auc": roc_auc_score([whole[a]["accepted"] for a in sorted(whole)],
                                      [scores[a] for a in sorted(whole)]),
        **outcome,
        "qualified": False,
        "reason": "Exposed candidate bank; new-page transfer, standalone generation, and advice safety are untested.",
    }
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    print(json.dumps({key: result[key] for key in
                      ("heldout_auc", "fixed_0_5", "optimistic_best_at_85_percent",
                       "optimistic_numeric_feasible")}, sort_keys=True))


if __name__ == "__main__":
    main()
