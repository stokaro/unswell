#!/usr/bin/env python3
"""Score frozen historical controls with calibration before test inference."""

from __future__ import annotations

import argparse
import hashlib
import importlib.util
import json
import math
import os
from pathlib import Path
import statistics

from tokenizers import Tokenizer


HERE = Path(__file__).resolve().parent
FRAME = HERE / "frame.json"
MODEL_SHA256 = "ddd1173f2ef517ad499965e5029fae8099a8054a2bc76d8134e5889cc4ed1b3e"


def need(ok: bool, message: str) -> None:
    if not ok:
        raise ValueError(message)


def sha(path: Path) -> str:
    value = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            value.update(block)
    return value.hexdigest()


def module(path: Path, name: str):
    spec = importlib.util.spec_from_file_location(name, path)
    need(spec is not None and spec.loader is not None, f"Missing helper: {name}")
    loaded = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(loaded)
    return loaded


def check_freeze(freeze: dict, model_dir: Path, runs_dir: Path) -> Path:
    for name, expected in freeze["local_file_sha256"].items():
        need(sha(HERE / name) == expected, f"Frozen file changed: {name}")
    for name, expected in freeze["helper_file_sha256"].items():
        need(sha(runs_dir / name) == expected, f"Preprocessing helper changed: {name}")
    for name, expected in freeze["model_file_sha256"].items():
        need(sha(model_dir / name) == expected, f"Modelkit changed: {name}")
    model_path = model_dir / "onnx/model.onnx"
    need(sha(model_path) == freeze["model_sha256"] == MODEL_SHA256, "Model changed")
    return model_path


def prepare(frame: dict, sources: Path, tokenizer: Tokenizer, richness, origin,
            audit) -> tuple[list[dict], list[dict]]:
    docs = []
    windows = []
    for selected in frame["selected"]:
        sid = selected["source_id"]
        raw = (sources / f"{sid}.source").read_bytes()
        blob_sha = hashlib.sha1(f"blob {len(raw)}\0".encode() + raw).hexdigest()
        need(len(raw) == selected["bytes"] and blob_sha == selected["blob_sha"]
             and hashlib.sha256(raw).hexdigest() == selected["source_sha256"],
             f"Source changed: {sid}")
        text = raw.decode("utf-8")
        need("\0" not in text, f"NUL in source: {sid}")
        cleaned, _ = origin.clean_text(richness.prose(text))
        need(hashlib.sha256(cleaned.encode()).hexdigest() == selected["cleaned_sha256"],
             f"Cleaned source changed: {sid}")
        parts = audit.word_windows(cleaned, origin.WORD, 100)
        lengths = [len(tokenizer.encode(part).ids) for part in parts]
        need(parts and "".join(parts) == cleaned and
             len(origin.WORD.findall(cleaned)) == selected["cleaned_words"] and
             len(parts) == selected["windows_100"] and
             max(lengths) == selected["max_tokens_100"] and
             all(0 < length <= 512 for length in lengths),
             f"Window selection changed: {sid}")
        docs.append({**selected, "score": None, "flagged": None})
        for index, part in enumerate(parts):
            ids = tokenizer.encode(part).ids
            windows.append({"source_id": sid, "arm": selected["arm"],
                            "index": index, "tokens": len(ids), "ids": ids,
                            "score": None})
    need(len(docs) == 761, "Document count changed")
    return docs, windows


def fill_document_scores(docs: list[dict], windows: list[dict], arm: str) -> list[float]:
    grouped = {}
    for window in windows:
        if window["arm"] == arm:
            need(window["score"] is not None and math.isfinite(window["score"]),
                 "Missing window score")
            grouped.setdefault(window["source_id"], []).append(window)
    values = []
    for doc in docs:
        if doc["arm"] != arm:
            continue
        rows = sorted(grouped.get(doc["source_id"], []), key=lambda item: item["index"])
        need([row["index"] for row in rows] == list(range(doc["windows_100"])),
             f"Incomplete document score: {doc['source_id']}")
        doc["score"] = statistics.mean(row["score"] for row in rows)
        values.append(doc["score"])
    return values


def save_new(path: Path, payload) -> None:
    with path.open("x") as stream:
        json.dump(payload, stream, sort_keys=True, indent=2, allow_nan=False)
        stream.write("\n")
        stream.flush()
        os.fsync(stream.fileno())


def public_windows(rows: list[dict]) -> list[dict]:
    return [{key: value for key, value in row.items() if key != "ids"} for row in rows]


def wilson(successes: int, total: int) -> list[float]:
    z = 1.959963984540054
    rate = successes / total
    denominator = 1 + z * z / total
    center = (rate + z * z / (2 * total)) / denominator
    half = z * math.sqrt(rate * (1 - rate) / total + z * z / (4 * total * total)) / denominator
    return [center - half, center + half]


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--model-dir", type=Path, required=True)
    parser.add_argument("--sources-dir", type=Path, required=True)
    parser.add_argument("--runs-dir", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path)
    parser.add_argument("--preflight", action="store_true")
    args = parser.parse_args()
    need(args.preflight or args.output_dir is not None, "Output directory required")
    freeze = json.loads((HERE / "freeze.json").read_text())
    model = check_freeze(freeze, args.model_dir, args.runs_dir)
    frame = json.loads(FRAME.read_text())
    frame_verifier = module(HERE / "verify_frame.py", "new_frame_verifier")
    frame_verifier.verify(HERE, args.sources_dir, FRAME)
    tokenizer = Tokenizer.from_file(str(args.model_dir / "tokenizer.json"))
    richness = module(args.runs_dir / "2026-11-07-richness-screen/run.py", "new_richness")
    origin = module(args.runs_dir / "2026-11-09-editlens-transfer/run.py", "new_origin")
    audit = module(args.runs_dir / "2026-11-10-editlens-technical-windows/audit.py", "new_audit")
    docs, windows = prepare(frame, args.sources_dir, tokenizer, richness, origin, audit)
    if args.preflight:
        print(json.dumps({"documents": len(docs), "windows": len(windows),
                          "frame_sha256": sha(FRAME)}, sort_keys=True))
        return
    args.output_dir.mkdir(parents=True, exist_ok=False)
    calibration_windows = [row for row in windows if row["arm"] == "calibration"]
    origin.score(calibration_windows, model)
    cal_scores = fill_document_scores(docs, windows, "calibration")
    need(len(cal_scores) == 404, "Calibration document count")
    threshold = max(cal_scores)
    for doc in docs:
        if doc["arm"] == "calibration":
            doc["flagged"] = doc["score"] > threshold
    cal = {"format": "unswell-new-historical-calibration-v1",
           "freeze_sha256": sha(HERE / "freeze.json"),
           "frame_sha256": sha(FRAME), "threshold_strictly_greater": threshold,
           "documents": [row for row in docs if row["arm"] == "calibration"],
           "windows": public_windows(calibration_windows)}
    save_new(args.output_dir / "calibration.json", cal)
    print(f"calibration persisted: {len(cal_scores)} documents; threshold {threshold:.17g}", flush=True)
    test_windows = [row for row in windows if row["arm"] == "test"]
    origin.score(test_windows, model)
    test_scores = fill_document_scores(docs, windows, "test")
    need(len(test_scores) == 357, "Test document count")
    for doc in docs:
        if doc["arm"] == "test":
            doc["flagged"] = doc["score"] > threshold
    flags = sum(row["flagged"] for row in docs if row["arm"] == "test")
    coverage = len(test_scores) / frame["candidate_counts"]["test"]
    test_docs = [row for row in docs if row["arm"] == "test"]
    by_repository = {}
    for repository in sorted({row["repository"] for row in test_docs}):
        group = [row for row in test_docs if row["repository"] == repository]
        by_repository[repository] = {"documents": len(group),
                                     "flagged": sum(row["flagged"] for row in group)}
    flagged_repos = sum(group["flagged"] > 0 for group in by_repository.values())
    by_language = {}
    for language in ("C#", "C++", "Java", "Python"):
        group = [row for row in test_docs if row["language"] == language]
        by_language[language] = {"documents": len(group),
                                 "repositories": len({row["repository"] for row in group}),
                                 "flagged": sum(row["flagged"] for row in group)}
    by_length = {}
    for label, predicate in (("1-4", lambda count: count <= 4),
                             ("5-9", lambda count: 5 <= count <= 9),
                             ("10+", lambda count: count >= 10)):
        group = [row for row in test_docs if predicate(row["windows_100"])]
        by_length[label] = {"documents": len(group),
                            "flagged": sum(row["flagged"] for row in group)}
    result = {"format": "unswell-new-historical-control-result-v1",
              "scope": "historical-control flag rate; authorship not verified",
              "freeze_sha256": sha(HERE / "freeze.json"),
              "frame_sha256": sha(FRAME),
              "calibration_sha256": sha(args.output_dir / "calibration.json"),
              "threshold_strictly_greater": threshold,
              "calibration_documents": len(cal_scores),
              "test_documents": len(test_scores),
              "test_candidates": frame["candidate_counts"]["test"],
              "test_coverage": coverage,
              "test_flagged": flags,
              "test_flag_rate": flags / len(test_scores),
              "test_repositories": len(by_repository),
              "test_flagged_repositories": flagged_repos,
              "test_repo_flag_wilson_95": wilson(flagged_repos, len(by_repository)),
              "by_language": by_language,
              "by_length": by_length,
              "by_repository": by_repository,
              "control_gate_pass": coverage >= 0.70 and flags / len(test_scores) <= 0.01}
    save_new(args.output_dir / "documents.json", docs)
    save_new(args.output_dir / "window_scores.json", public_windows(windows))
    save_new(args.output_dir / "result.json", result)
    print(json.dumps(result, sort_keys=True), flush=True)


if __name__ == "__main__":
    main()
