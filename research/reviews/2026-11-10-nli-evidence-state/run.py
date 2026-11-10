#!/usr/bin/env python3
"""Run one frozen local ONNX NLI pass over target-free source premises."""

import argparse
import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
LABELS = ("contradiction", "support", "no_evidence")


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--model-dir", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if args.output.exists():
        raise ValueError("output already exists; frozen run has no retry")
    frozen = json.loads((HERE / "freeze.json").read_text())
    for name, expected in frozen["files"].items():
        if digest(HERE / name) != expected:
            raise ValueError(f"frozen research file changed: {name}")
    assets = {"config": args.model_dir / "config.json",
              "tokenizer": args.model_dir / "tokenizer.json",
              "model": args.model_dir / "onnx/model_qint8_arm64.onnx"}
    for name, path in assets.items():
        if digest(path) != frozen["assets"][name]:
            raise ValueError(f"frozen model asset changed: {name}")
    config = json.loads(assets["config"].read_text())
    if config["id2label"] != {"0": "contradiction", "1": "entailment", "2": "neutral"}:
        raise ValueError("NLI label order changed")

    import numpy as np
    import onnxruntime as ort
    import tokenizers

    if (np.__version__, ort.__version__, tokenizers.__version__) != ("2.5.3", "1.30.0", "0.23.2"):
        raise ValueError("frozen local inference versions changed")
    tokenizer = tokenizers.Tokenizer.from_file(str(assets["tokenizer"]))
    if tokenizer.truncation is not None:
        raise ValueError("tokenizer silently truncates inputs")
    options = ort.SessionOptions()
    options.intra_op_num_threads = 2
    model = ort.InferenceSession(str(assets["model"]), sess_options=options,
                                 providers=["CPUExecutionProvider"])
    if [entry.name for entry in model.get_inputs()] != ["input_ids", "attention_mask"]:
        raise ValueError("unexpected ONNX model input contract")
    packet = json.loads((HERE / "packet.json").read_text())
    cases = packet["cases"]
    if packet["format"] != "unswell-349-nli-evidence-input-v1" or len(cases) != 16:
        raise ValueError("frozen packet shape changed")
    rows = []
    seen = set()
    for case in cases:
        item_id = case["item_id"]
        if item_id in seen:
            raise ValueError("duplicate item ID")
        seen.add(item_id)
        encoded = tokenizer.encode(case["premise"], case["hypothesis"])
        if len(encoded.ids) > 512:
            raise ValueError(f"{item_id}: input exceeds 512 token IDs")
        inputs = {"input_ids": np.asarray([encoded.ids], dtype=np.int64),
                  "attention_mask": np.asarray([encoded.attention_mask], dtype=np.int64)}
        scores = model.run(None, inputs)[0]
        if scores.shape != (1, 3) or not np.isfinite(scores).all():
            raise ValueError(f"{item_id}: invalid NLI scores")
        index = int(np.argmax(scores[0]))
        rows.append({"item_id": item_id, "token_ids": len(encoded.ids),
                     "logits": [float(value) for value in scores[0]],
                     "relation": LABELS[index]})
    result = {"format": "unswell-349-nli-evidence-result-v1",
              "freeze_sha256": digest(HERE / "freeze.json"),
              "packet_sha256": frozen["files"]["packet.json"],
              "model_sha256": frozen["assets"]["model"],
              "rows": rows}
    args.output.write_text(json.dumps(result, indent=2, ensure_ascii=False) + "\n")
    print(f"scored {len(rows)} frozen evidence-state cases")


if __name__ == "__main__":
    main()
