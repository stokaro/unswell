#!/usr/bin/env python3
"""Run one frozen CPU NLI pass on explicit source-bound propositions."""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path
import urllib.request

HERE = Path(__file__).resolve().parent


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def require(ok: bool, reason: str) -> None:
    if not ok:
        raise ValueError(reason)


def asset(path: Path, url: str, expected: str, size: int) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    if not path.exists():
        partial = path.with_suffix(path.suffix + ".partial")
        require(not partial.exists(), "prior partial download requires manual investigation")
        with urllib.request.urlopen(url, timeout=120) as response, partial.open("wb") as output:
            while chunk := response.read(1024 * 1024):
                output.write(chunk)
        partial.rename(path)
    require(path.stat().st_size == size, f"asset size mismatch: {path.name}")
    require(digest(path) == expected, f"asset SHA-256 mismatch: {path.name}")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    require(not args.output.exists(), "the frozen run has no retry")
    freeze = json.loads((HERE / "freeze.json").read_text())
    require(freeze["format"] == "unswell-349-explicit-proposition-nli-freeze-v1", "freeze format")
    require(digest(HERE / "packet.json") == freeze["packet_sha256"], "packet changed")
    require(digest(HERE / "run.py") == freeze["runner_sha256"], "runner changed")
    require(digest(HERE / "score.py") == freeze["scorer_sha256"], "scorer changed")
    packet = json.loads((HERE / "packet.json").read_text())
    require(packet["format"] == "unswell-349-explicit-proposition-nli-v1" and
            len(packet["cases"]) == 16, "packet shape")

    model_dir = HERE / "model"
    base = ("https://huggingface.co/" + freeze["model_repository"] +
            "/resolve/" + freeze["model_revision"] + "/")
    paths = {}
    for name, info in freeze["assets"].items():
        path = model_dir / name
        asset(path, base + name, info["sha256"], info["size"])
        paths[name] = path
    config = json.loads(paths["config.json"].read_text())
    require(config["id2label"] == {"0": "entailment", "1": "neutral", "2": "contradiction"},
            "model label order changed")

    import numpy as np
    import onnxruntime as ort
    import tokenizers

    require((np.__version__, ort.__version__, tokenizers.__version__) ==
            ("2.5.3", "1.30.0", "0.23.2"), "inference dependency version changed")
    tokenizer = tokenizers.Tokenizer.from_file(str(paths["tokenizer.json"]))
    tokenizer.no_truncation()
    require(tokenizer.truncation is None, "tokenizer truncation still enabled")
    options = ort.SessionOptions()
    options.intra_op_num_threads = 2
    model = ort.InferenceSession(str(paths["onnx/model_quantized.onnx"]),
                                 sess_options=options, providers=["CPUExecutionProvider"])
    input_names = [entry.name for entry in model.get_inputs()]
    require("input_ids" in input_names and "attention_mask" in input_names and
            set(input_names) <= {"input_ids", "attention_mask", "token_type_ids"},
            "unexpected ONNX input contract")
    labels = [config["id2label"][str(index)] for index in range(3)]
    rows = []
    seen = set()
    for case in packet["cases"]:
        item_id = case["id"]
        require(item_id not in seen, "duplicate case ID")
        seen.add(item_id)
        encoded = tokenizer.encode(case["premise"], case["hypothesis"])
        require(len(encoded.ids) <= 512, f"{item_id}: too many token IDs")
        arrays = {"input_ids": encoded.ids,
                  "attention_mask": encoded.attention_mask,
                  "token_type_ids": encoded.type_ids}
        inputs = {name: np.asarray([arrays[name]], dtype=np.int64) for name in input_names}
        scores = model.run(None, inputs)[0]
        require(scores.shape == (1, 3) and bool(np.isfinite(scores).all()),
                f"{item_id}: invalid logits")
        index = int(np.argmax(scores[0]))
        rows.append({"id": item_id, "token_ids": len(encoded.ids),
                     "logits": [float(value) for value in scores[0]],
                     "relation": labels[index]})
    result = {"format": "unswell-349-explicit-proposition-nli-result-v1",
              "freeze_sha256": digest(HERE / "freeze.json"),
              "packet_sha256": freeze["packet_sha256"],
              "model_sha256": freeze["assets"]["onnx/model_quantized.onnx"]["sha256"],
              "input_names": input_names, "rows": rows}
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    print(json.dumps({"cases": len(rows), "result_sha256": digest(args.output)}))


if __name__ == "__main__":
    main()
