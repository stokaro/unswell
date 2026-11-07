#!/usr/bin/env python3
"""Run a frozen, offline span-neighbor screen on AITDNA study conditions."""

from __future__ import annotations

import argparse
import collections
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import statistics
import time

import numpy as np
import pyarrow.parquet as parquet
import torch

from ape import APE


HERE = Path(__file__).resolve().parent
FEATURE = HERE.parent / "2026-11-07-richness-screen" / "run.py"
ASSETS = ("config.json", "special_tokens_map.json", "tokenizer.json",
          "tokenizer_config.json", "vocab.txt")
TASKS = ("Argumentative Writing", "Creative Writing", "Explanatory Writing")
DATASET_ROWS = 362
SPAN_WORDS = 25
SPAN_COUNT = 4
NEIGHBORS = 5


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def require(value: bool, message: str) -> None:
    if not value:
        raise ValueError(message)


def load_feature_module():
    spec = importlib.util.spec_from_file_location("unswell_richness_screen", FEATURE)
    require(spec is not None and spec.loader is not None, "Feature source unavailable")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def verify_inputs(args: argparse.Namespace) -> dict:
    freeze = json.loads((HERE / "freeze.json").read_text())
    require(freeze["protocol_sha256"] == digest(HERE / "PROTOCOL.md"),
            "Protocol changed after freeze")
    require(freeze["runner_sha256"] == digest(Path(__file__)),
            "Runner changed after freeze")
    require(freeze["feature_sha256"] == digest(FEATURE),
            "Prose feature source changed")
    require(freeze["dataset_sha256"] == digest(args.parquet), "Dataset changed")
    require(freeze["ape_code_commit"] == "a6723d56e70a5fc92577858c9c2a2f17a6cbfe2a",
            "APE code pin changed")
    require(freeze["model_weights_sha256"] == digest(args.model / "model.safetensors"),
            "APE checkpoint changed")
    require(freeze["model_config_sha256"] == digest(args.model / "config.json"),
            "APE configuration changed")
    for name in ASSETS:
        require(freeze["backbone_sha256"][name] == digest(args.backbone / name),
                f"Backbone asset changed: {name}")
    config = json.loads((args.model / "config.json").read_text())
    require(Path(config["encoder_name"]).resolve() == args.backbone.resolve(),
            "APE checkpoint not bound to pinned local backbone")
    require(os.environ.get("HF_HUB_OFFLINE") == "1" and
            os.environ.get("TRANSFORMERS_OFFLINE") == "1", "Offline mode required")
    return freeze


def documents(path: Path, feature) -> list[dict]:
    rows = parquet.read_table(path).to_pylist()
    require(len(rows) == DATASET_ROWS, "AITDNA row count changed")
    documents_ = []
    for index, row in enumerate(rows):
        metadata = row["metadata"]
        parts = row["data"]
        require(isinstance(parts, list) and parts and
                isinstance(metadata["author"], str) and metadata["author"] and
                type(metadata["human_only"]) is bool and
                isinstance(metadata["task"], str), "Malformed AITDNA row")
        raw = "\n".join(part["text"] for part in parts)
        words = [match.group().casefold() for match in
                 feature.WORD.finditer(feature.prose(raw))]
        require(len(words) >= SPAN_WORDS * SPAN_COUNT, "Short AITDNA document")
        prefix = words[:SPAN_WORDS * SPAN_COUNT]
        richness = feature.features(" ".join(prefix))["richness"]
        bot_words = sum(len(feature.WORD.findall(feature.prose(part["text"])))
                        for part in parts if part["author"] == "Bot")
        total_words = sum(len(feature.WORD.findall(feature.prose(part["text"])))
                          for part in parts)
        require(total_words > 0, "Empty AITDNA document")
        documents_.append({
            "row_idx": index, "author": metadata["author"],
            "task": metadata["task"], "assisted": not metadata["human_only"],
            "bot_majority": bot_words * 2 >= total_words,
            "richness": richness,
            "spans": [" ".join(prefix[i:i + SPAN_WORDS])
                      for i in range(0, len(prefix), SPAN_WORDS)],
        })
    require(sum(row["task"] == "Peer Review" for row in documents_) == 10,
            "Peer-review abstention scope changed")
    require({row["task"] for row in documents_} == set(TASKS) | {"Peer Review"},
            "Task inventory changed")
    return documents_


def encode_spans(rows: list[dict], model_path: Path) -> tuple[np.ndarray, list[tuple[int, int]], set[int]]:
    torch.set_num_threads(4)
    model = APE.from_pretrained(str(model_path)).to("cpu").eval()
    spans = [(row["row_idx"], ordinal, text)
             for row in rows if row["task"] in TASKS
             for ordinal, text in enumerate(row["spans"])]
    texts = [text for _, _, text in spans]
    tokenizer = model.tokenizer
    encoded = tokenizer(texts, truncation=False, padding=False)
    too_long = {row for (row, _, _), ids in zip(spans, encoded["input_ids"])
                if len(ids) > 128}
    eligible = [(row, ordinal, text) for row, ordinal, text in spans
                if row not in too_long]
    vectors = []
    with torch.inference_mode():
        for start in range(0, len(eligible), 32):
            batch = [entry[2] for entry in eligible[start:start + 32]]
            tokens = tokenizer(batch, padding=True, truncation=False,
                               return_tensors="pt")
            require(tokens["input_ids"].shape[1] <= 128, "Hidden truncation")
            hidden = model.encoder(**tokens).last_hidden_state
            mask = tokens["attention_mask"].unsqueeze(-1)
            pooled = (hidden * mask).sum(dim=1) / mask.sum(dim=1)
            pooled = torch.nn.functional.normalize(pooled, p=2, dim=1)
            vectors.append(pooled.cpu().numpy().astype(np.float32))
    require(len(eligible) == (len(rows) - 10 - len(too_long)) * SPAN_COUNT,
            "Incomplete span embedding")
    return np.concatenate(vectors), [(row, ordinal) for row, ordinal, _ in eligible], too_long


def neighbors(vectors: np.ndarray, spans: list[tuple[int, int]],
              rows: list[dict]) -> list[dict]:
    similarity = vectors @ vectors.T
    by_task: dict[str, list[int]] = collections.defaultdict(list)
    for position, (row, _) in enumerate(spans):
        by_task[rows[row]["task"]].append(position)
    output = []
    for row in rows:
        if row["task"] == "Peer Review":
            output.append({"row_idx": row["row_idx"], "abstain": "no_same_task_control"})
            continue
        positions = [i for i, (source, _) in enumerate(spans)
                     if source == row["row_idx"]]
        if not positions:
            output.append({"row_idx": row["row_idx"], "abstain": "overlength_span"})
            continue
        require(len(positions) == SPAN_COUNT, "Partial row embedding")
        span_scores = []
        span_evidence = []
        reason = None
        for position in positions:
            by_class: dict[bool, dict[int, float]] = {False: {}, True: {}}
            for candidate in by_task[row["task"]]:
                source = spans[candidate][0]
                peer = rows[source]
                if peer["author"] == row["author"]:
                    continue
                bucket = by_class[peer["assisted"]]
                score = float(similarity[position, candidate])
                bucket[source] = max(score, bucket.get(source, -1.0))
            if any(len(bucket) < NEIGHBORS for bucket in by_class.values()):
                reason = "insufficient_other_author_neighbors"
                break
            best = {label: sorted(bucket.items(), key=lambda item: (-item[1], item[0]))[:NEIGHBORS]
                    for label, bucket in by_class.items()}
            margin = statistics.mean(score for _, score in best[True]) - \
                statistics.mean(score for _, score in best[False])
            span_scores.append(margin)
            span_evidence.append({"span": spans[position][1],
                                  "assisted": [[source, round(score, 8)] for source, score in best[True]],
                                  "human_only": [[source, round(score, 8)] for source, score in best[False]]})
        if reason:
            output.append({"row_idx": row["row_idx"], "abstain": reason})
        else:
            output.append({
                "row_idx": row["row_idx"], "task": row["task"],
                "assisted": row["assisted"], "bot_majority": row["bot_majority"],
                "author_sha256": hashlib.sha256(row["author"].encode()).hexdigest(),
                "score": statistics.mean(span_scores), "span_scores": span_scores,
                "richness": row["richness"], "neighbors": span_evidence,
            })
    return output


def auc(rows: list[dict], field: str) -> float | None:
    positive = [row[field] for row in rows if row["assisted"]]
    negative = [row[field] for row in rows if not row["assisted"]]
    if not positive or not negative:
        return None
    wins = sum((p > n) + 0.5 * (p == n) for p in positive for n in negative)
    return wins / (len(positive) * len(negative))


def paired(rows: list[dict], by_task: bool) -> dict[str, int]:
    groups: dict[tuple[str, ...], list[dict]] = collections.defaultdict(list)
    for row in rows:
        key = (row["author_sha256"], row["task"]) if by_task else (row["author_sha256"],)
        groups[key].append(row)
    counts = collections.Counter()
    for group in groups.values():
        positive = [r["score"] for r in group if r["assisted"]]
        negative = [r["score"] for r in group if not r["assisted"]]
        if positive and negative:
            delta = statistics.mean(positive) - statistics.mean(negative)
            counts["positive" if delta > 0 else "negative" if delta < 0 else "tied"] += 1
    return {key: counts[key] for key in ("positive", "negative", "tied")}


def summarize(rows: list[dict]) -> dict:
    scored = [row for row in rows if "score" in row]
    tasks = {}
    for task in TASKS:
        group = [row for row in scored if row["task"] == task]
        tasks[task] = {"rows": len(group), "span_auc": auc(group, "score"),
                       "richness_auc": auc(group, "richness")}
    both = paired(scored, by_task=True)
    positive_count = both["positive"] + both["negative"]
    span_auc = auc(scored, "score")
    richness_auc = auc(scored, "richness")
    require(span_auc is not None and richness_auc is not None, "Missing study class")
    bot = [r for r in scored if not r["assisted"] or r["bot_majority"]]
    user = [r for r in scored if not r["assisted"] or not r["bot_majority"]]
    return {
        "coverage": {"total": len(rows), "scored": len(scored),
                     "abstentions": dict(sorted(collections.Counter(
                         row["abstain"] for row in rows if "abstain" in row).items()))},
        "overall": {"span_auc": span_auc, "richness_auc": richness_auc,
                    "delta_auc": span_auc - richness_auc},
        "by_task": tasks,
        "bot_majority_span_auc": auc(bot, "score"),
        "user_minority_span_auc": auc(user, "score"),
        "paired_author": paired(scored, by_task=False),
        "paired_author_task": both,
        "feasibility_pass": span_auc >= 0.70 and span_auc - richness_auc >= 0.05 and
                            all(v["span_auc"] is not None and v["span_auc"] >= 0.60
                                for v in tasks.values()) and
                            positive_count > 0 and both["positive"] / positive_count >= 0.60,
    }


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--parquet", type=Path, required=True)
    parser.add_argument("--model", type=Path, required=True)
    parser.add_argument("--backbone", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    started = time.monotonic()
    freeze = verify_inputs(args)
    feature = load_feature_module()
    rows = documents(args.parquet, feature)
    vectors, spans, too_long = encode_spans(rows, args.model)
    scored = neighbors(vectors, spans, rows)
    require({row["row_idx"] for row in scored} == set(range(DATASET_ROWS)),
            "Incomplete row accounting")
    summary = summarize(scored)
    require(summary["coverage"]["scored"] + sum(summary["coverage"]["abstentions"].values())
            == DATASET_ROWS, "Coverage mismatch")
    args.output.mkdir(parents=True, exist_ok=False)
    score_bytes = (json.dumps(scored, indent=2, sort_keys=True) + "\n").encode()
    (args.output / "scores.json").write_bytes(score_bytes)
    result = {"format": "unswell-span-example-screen-v1",
              "input_freeze_sha256": digest(HERE / "freeze.json"),
              "scores_sha256": hashlib.sha256(score_bytes).hexdigest(),
              "encoder_spans": len(spans), "overlength_rows": len(too_long),
              "elapsed_seconds": time.monotonic() - started,
              **summary,
              "interpretation": "AITDNA study-setting contrast, not machine-authorship accuracy or editorial value"}
    (args.output / "result.json").write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    print(json.dumps({key: result[key] for key in
                      ("coverage", "overall", "paired_author_task", "feasibility_pass")},
                     sort_keys=True))


if __name__ == "__main__":
    main()
