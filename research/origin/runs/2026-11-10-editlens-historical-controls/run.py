#!/usr/bin/env python3
"""Score pinned historical technical controls with the fixed EditLens model."""

from __future__ import annotations

import argparse
import collections
import hashlib
import importlib.util
import json
import math
from pathlib import Path
import statistics

from tokenizers import Tokenizer


HERE = Path(__file__).resolve().parent
RUNS = HERE.parent
RICHNESS = RUNS / "2026-11-07-richness-screen" / "run.py"
ORIGIN = RUNS / "2026-11-09-editlens-transfer" / "run.py"
AUDIT = RUNS / "2026-11-10-editlens-technical-windows" / "audit.py"
FULL = RUNS / "2026-11-10-editlens-technical-fullwindows" / "run.py"
ACADEMIC = RUNS / "2026-11-10-editlens-five-window" / "result.json"
FRAME = HERE / "frame.json"
INVENTORY = HERE / "inventory.json"
BLOB_AUDIT = HERE / "blob_audit.json"
ELIGIBILITY = HERE / "eligibility.json"
MODEL_SHA256 = "ddd1173f2ef517ad499965e5029fae8099a8054a2bc76d8134e5889cc4ed1b3e"
THRESHOLD = 0.27447814408224075
WINDOW_WIDTH = 100
EXPECTED_DOCS = 462
EXPECTED_WINDOWS = 3735


def need(ok: bool, message: str) -> None:
    if not ok:
        raise ValueError(message)


def sha(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def load(path: Path, name: str):
    spec = importlib.util.spec_from_file_location(name, path)
    need(spec is not None and spec.loader is not None, f"Missing helper: {name}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def source_id(row: dict) -> str:
    identity = "\0".join((row["repository"], row["commit"], row["path"]))
    return hashlib.sha256(identity.encode()).hexdigest()


def blob_sha(raw: bytes) -> str:
    header = f"blob {len(raw)}\0".encode()
    return hashlib.sha1(header + raw).hexdigest()


def frame_identity(rows: list[dict]) -> str:
    identities = [(row["repository"], row["commit"], row["path"],
                   row["blob_sha"], row["source_sha256"], row["windows_100"])
                  for row in rows]
    canonical = json.dumps(identities, ensure_ascii=False,
                           separators=(",", ":")).encode()
    return hashlib.sha256(canonical).hexdigest()


def prepare(frame: dict, sources: Path, tokenizer: Tokenizer, richness,
            origin, audit) -> tuple[list[dict], list[dict], str]:
    rows = frame["selected"]
    need(len(rows) == EXPECTED_DOCS and len({row["source_id"] for row in rows}) == len(rows),
         "Control frame identity changed")
    need(frame_identity(rows) == frame["selection_sha256"],
         "Control frame selection changed")
    prepared = []
    windows = []
    selection = []
    for row in rows:
        need(source_id(row) == row["source_id"], "Source ID changed")
        raw = (sources / (row["source_id"] + ".source")).read_bytes()
        need(len(raw) == row["bytes"] and blob_sha(raw) == row["blob_sha"] and
             hashlib.sha256(raw).hexdigest() == row["source_sha256"],
             f"Source blob changed: {row['source_id']}")
        text = raw.decode("utf-8")
        need("\0" not in text, "NUL in source")
        cleaned, removed_header = origin.clean_text(richness.prose(text))
        words = len(origin.WORD.findall(cleaned))
        parts = audit.word_windows(cleaned, origin.WORD, WINDOW_WIDTH)
        need(parts and "".join(parts) == cleaned, "Incomplete cleaned prose")
        ids = [tokenizer.encode(part).ids for part in parts]
        lengths = [len(items) for items in ids]
        need(words == row["cleaned_words"] and len(parts) == row["windows_100"] and
             max(lengths) == row["max_tokens_100"] and
             all(0 < length <= 512 for length in lengths),
             f"Control window inventory changed: {row['source_id']}")
        item = dict(row)
        item.update(removed_header=removed_header, score=None, flagged=None)
        prepared.append(item)
        selection.append((row["source_id"], row["source_sha256"], lengths))
        for index, token_ids in enumerate(ids):
            windows.append({"source_id": row["source_id"], "index": index,
                            "tokens": len(token_ids), "ids": token_ids, "score": None})
    need(len(windows) == EXPECTED_WINDOWS, "Window count changed")
    canonical = json.dumps(selection, ensure_ascii=False,
                           separators=(",", ":")).encode()
    return prepared, windows, hashlib.sha256(canonical).hexdigest()


def verify_freeze(freeze: dict, model_dir: Path, selection: str) -> Path:
    for name, path in (("protocol_sha256", HERE / "PROTOCOL.md"),
                       ("runner_sha256", HERE / "run.py"),
                       ("verifier_sha256", HERE / "verify.py"),
                       ("fetch_sha256", HERE / "fetch.py"),
                       ("selection_verifier_sha256", HERE / "verify_selection.py"),
                       ("runner_tests_sha256", HERE / "test_run.py"),
                       ("verifier_tests_sha256", HERE / "test_verify.py"),
                       ("selection_tests_sha256", HERE / "test_verify_selection.py"),
                       ("frame_sha256", FRAME),
                       ("inventory_sha256", INVENTORY),
                       ("blob_audit_sha256", BLOB_AUDIT),
                       ("eligibility_sha256", ELIGIBILITY),
                       ("richness_runner_sha256", RICHNESS),
                       ("origin_runner_sha256", ORIGIN),
                       ("window_audit_sha256", AUDIT),
                       ("fullwindow_runner_sha256", FULL),
                       ("academic_result_sha256", ACADEMIC)):
        need(sha(path) == freeze[name], f"Frozen input changed: {name}")
    academic = json.loads(ACADEMIC.read_text())
    need(academic["threshold_strictly_greater"] == THRESHOLD,
         "Academic threshold changed")
    need(selection == freeze["selection_sha256"], "Window selection changed")
    for name, expected in freeze["model_file_sha256"].items():
        need(sha(model_dir / name) == expected, f"Modelkit file changed: {name}")
    model = model_dir / "onnx/model.onnx"
    need(sha(model) == MODEL_SHA256 == freeze["model_sha256"], "Model changed")
    return model


def summarize(docs: list[dict], windows: list[dict], selection: str) -> dict:
    grouped = collections.defaultdict(list)
    for window in windows:
        need(window["score"] is not None and math.isfinite(window["score"]),
             "Unscored or invalid window")
        grouped[window["source_id"]].append(window)
    for doc in docs:
        ordered = sorted(grouped[doc["source_id"]], key=lambda row: row["index"])
        need([row["index"] for row in ordered] == list(range(doc["windows_100"])),
             "Missing or repeated source window")
        doc["score"] = statistics.mean(window["score"] for window in ordered)
        doc["flagged"] = doc["score"] > THRESHOLD
    flags = [doc for doc in docs if doc["flagged"]]
    scores = sorted(doc["score"] for doc in docs)

    def quantile(fraction: float) -> float:
        rank = fraction * (len(scores) - 1)
        lower = math.floor(rank)
        weight = rank - lower
        return scores[lower] * (1 - weight) + scores[math.ceil(rank)] * weight

    by_language = {}
    for language in ("C#", "C++", "Java", "Python"):
        subset = [doc for doc in docs if doc["language"] == language]
        by_language[language] = {"documents": len(subset),
                                 "repositories": len({doc["repository"] for doc in subset}),
                                 "flagged": sum(doc["flagged"] for doc in subset)}
    by_repository = {}
    for repository in sorted({doc["repository"] for doc in docs}):
        subset = [doc for doc in docs if doc["repository"] == repository]
        by_repository[repository] = {"documents": len(subset),
                                     "flagged": sum(doc["flagged"] for doc in subset)}
    by_length = {}
    for label, predicate in (("1-4", lambda n: n <= 4),
                             ("5-9", lambda n: 5 <= n <= 9),
                             ("10+", lambda n: n >= 10)):
        subset = [doc for doc in docs if predicate(doc["windows_100"])]
        by_length[label] = {"documents": len(subset),
                            "flagged": sum(doc["flagged"] for doc in subset)}
    return {"format": "unswell-editlens-historical-controls-v1",
            "scope": "pre-2021 technical control flag rate; authorship not verified",
            "model_sha256": MODEL_SHA256,
            "frame_sha256": sha(FRAME),
            "selection_sha256": selection,
            "threshold_strictly_greater": THRESHOLD,
            "documents": len(docs), "repositories": len({doc["repository"] for doc in docs}),
            "windows": len(windows), "flagged": len(flags),
            "flag_rate": len(flags) / len(docs),
            "median_score": statistics.median(scores),
            "score_quantiles": {f"p{percent}": quantile(percent / 100)
                                for percent in (10, 25, 75, 90)},
            "by_language": by_language, "by_repository": by_repository,
            "by_length": by_length,
            "flagged_source_ids": [doc["source_id"] for doc in flags],
            "control_transfer_pass": len(docs) == EXPECTED_DOCS and
                                     len(windows) == EXPECTED_WINDOWS and
                                     len(flags) == 0}


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--model-dir", type=Path, required=True)
    parser.add_argument("--sources-dir", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path)
    parser.add_argument("--preflight", action="store_true")
    args = parser.parse_args()
    need(args.preflight or args.output_dir is not None, "Output directory required")
    frame = json.loads(FRAME.read_text())
    richness = load(RICHNESS, "historical_richness")
    origin = load(ORIGIN, "historical_origin")
    audit = load(AUDIT, "historical_audit")
    tokenizer = Tokenizer.from_file(str(args.model_dir / "tokenizer.json"))
    docs, windows, selection = prepare(frame, args.sources_dir, tokenizer,
                                       richness, origin, audit)
    if args.preflight:
        print(json.dumps({"documents": len(docs), "windows": len(windows),
                          "selection_sha256": selection}, sort_keys=True))
        return
    freeze = json.loads((HERE / "freeze.json").read_text())
    model = verify_freeze(freeze, args.model_dir, selection)
    origin.score(windows, model)
    result = summarize(docs, windows, selection)
    args.output_dir.mkdir(parents=True, exist_ok=True)
    for name, payload in (("result.json", result), ("documents.json", docs),
                          ("window_scores.json", windows)):
        if name == "window_scores.json":
            payload = [{key: value for key, value in row.items() if key != "ids"}
                       for row in windows]
        (args.output_dir / name).write_text(json.dumps(payload, sort_keys=True, indent=2) + "\n")
    print(json.dumps(result, sort_keys=True))


if __name__ == "__main__":
    main()
