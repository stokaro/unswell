#!/usr/bin/env python3
"""Score every source-local sentence pair with a pinned offline APE model."""

import argparse
from collections import defaultdict
import gzip
import hashlib
import json
import os
from pathlib import Path
import signal
import time

from ape import APE, APEScore
import torch


def sha256(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def first_byte(row):
    return row["spans"][0]["start"]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--sentences", type=Path, required=True)
    parser.add_argument("--model", type=Path, required=True)
    parser.add_argument("--backbone", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    args.output.mkdir(parents=True, exist_ok=False)
    signal.alarm(3 * 60 * 60)
    torch.set_num_threads(4)
    started = time.monotonic()

    rows = json.loads(args.sentences.read_text())
    pages = {row["page"] for row in rows}
    if len(pages) != 36:
        raise ValueError(f"Expected 36 pages, got {len(pages)}")
    if os.environ.get("HF_HUB_OFFLINE") != "1" or os.environ.get("TRANSFORMERS_OFFLINE") != "1":
        raise ValueError("Offline environment is required")
    model_config = json.loads((args.model / "config.json").read_text())
    if model_config["encoder_name"] != str(args.backbone.resolve()):
        raise ValueError("Checkpoint is not bound to the pinned local backbone")

    model = APE.from_pretrained(str(args.model)).to("cpu")
    model.eval()
    texts = [row["text"] for row in rows]
    lengths = [len(ids) for ids in model.tokenizer(texts, truncation=False)["input_ids"]]
    eligible = [i for i, count in enumerate(lengths) if count <= 128]
    embeddings = {}
    for index, value in zip(eligible, model.encode([texts[i] for i in eligible],
                                                  batch_size=32, activity_threshold=0.6,
                                                  show_progress_bar=True)):
        embeddings[index] = value

    groups = defaultdict(list)
    for i, row in enumerate(rows):
        groups[(row["page"], row["block"], row["piece"])].append(i)
    scores = [None] * len(rows)
    pair_results = []
    for indices in groups.values():
        indices.sort(key=lambda i: first_byte(rows[i]))
        for ordinal, i in enumerate(indices):
            if i not in embeddings:
                continue
            peers = indices[max(0, ordinal - 2):ordinal] + indices[ordinal + 1:ordinal + 3]
            for j in peers:
                if j not in embeddings:
                    continue
                value = APEScore(embeddings[i], embeddings[j]).score("h_min_to_p")
                pair_results.append({"candidate": i, "premise": j, "score": value})
                if scores[i] is None or value > scores[i]:
                    scores[i] = value

    ranked = defaultdict(list)
    for i, score in enumerate(scores):
        if score is not None:
            ranked[rows[i]["page"]].append(i)
    selected = []
    for page in sorted(pages):
        order = sorted(ranked[page], key=lambda i: (-scores[i], first_byte(rows[i])))
        selected.extend(order[:3])

    records = [{"index": i, **row, "token_count": lengths[i], "score": scores[i]}
               for i, row in enumerate(rows)]
    raw = {"sentences": records, "pairs": pair_results, "selected_indices": selected}
    encoded = json.dumps(raw, separators=(",", ":"), ensure_ascii=False).encode()
    (args.output / "scores.json.gz").write_bytes(gzip.compress(encoded, mtime=0))
    (args.output / "top-candidates.json").write_text(json.dumps(
        [records[i] for i in selected], indent=2, ensure_ascii=False) + "\n")
    manifest = {
        "source_tree": "6e724618005623eddbd325653c0e04e7c38b317d",
        "ape_code_commit": "a6723d56e70a5fc92577858c9c2a2f17a6cbfe2a",
        "ape_model_revision": "e6dc09deb8c9e18323f17bbfd35b9025d89d8044",
        "backbone_revision": "5c38ec7c405ec4b44b94cc5a9bb96e735b38267a",
        "sentences_sha256": sha256(args.sentences),
        "model_config_sha256": sha256(args.model / "config.json"),
        "model_weights_sha256": sha256(args.model / "model.safetensors"),
        "backbone_asset_sha256": {name: sha256(args.backbone / name) for name in (
            "config.json", "tokenizer.json", "tokenizer_config.json", "vocab.txt",
            "special_tokens_map.json")},
        "torch": torch.__version__,
        "device": "cpu",
        "batch_size": 32,
        "threads": 4,
        "sentences": len(rows),
        "eligible_sentences": len(eligible),
        "skipped_length": len(rows) - len(eligible),
        "scored_pairs": len(pair_results),
        "selected_candidates": len(selected),
        "elapsed_seconds": time.monotonic() - started,
        "scores_sha256": sha256(args.output / "scores.json.gz"),
        "top_candidates_sha256": sha256(args.output / "top-candidates.json"),
    }
    (args.output / "run.json").write_text(json.dumps(manifest, indent=2) + "\n")
    print(json.dumps({key: manifest[key] for key in (
        "sentences", "eligible_sentences", "skipped_length", "scored_pairs",
        "selected_candidates", "elapsed_seconds")}, indent=2))


if __name__ == "__main__":
    main()
