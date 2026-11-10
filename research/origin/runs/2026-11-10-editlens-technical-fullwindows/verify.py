#!/usr/bin/env python3
"""Independently recalculate full-window document and pair accounting."""

from __future__ import annotations

import argparse
import collections
import hashlib
import json
import math
import statistics
from pathlib import Path


HERE = Path(__file__).resolve().parent
RUNS = HERE.parent
STUDY = HERE.parents[2] / "generation" / "studies" / "long-prose-v3"
INVENTORY = RUNS / "2026-11-10-editlens-technical-windows" / "result.json"
FAMILIES = ("haiku", "luna", "qwen")
OPERATIONS = ("generate", "polish")


def need(ok: bool, message: str) -> None:
    if not ok:
        raise ValueError(message)


def sha(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def load(path: Path):
    return json.loads(path.read_text())


def equal(left, right, path: str = "result") -> None:
    if isinstance(left, bool) or isinstance(right, bool):
        need(left is right, f"{path} changed")
    elif isinstance(left, (float, int)) and isinstance(right, (float, int)):
        need(math.isfinite(left) and math.isfinite(right) and
             math.isclose(left, right, rel_tol=1e-12, abs_tol=1e-12),
             f"{path} changed")
    elif isinstance(left, dict) and isinstance(right, dict):
        need(left.keys() == right.keys(), f"{path} keys changed")
        for key in left:
            equal(left[key], right[key], f"{path}.{key}")
    elif isinstance(left, list) and isinstance(right, list):
        need(len(left) == len(right), f"{path} length changed")
        for index, (a, b) in enumerate(zip(left, right, strict=True)):
            equal(a, b, f"{path}[{index}]")
    else:
        need(left == right, f"{path} changed")


def key(row: dict) -> str:
    if row["kind"] == "source":
        return "source/" + row["task_id"]
    need(row["kind"] == "response", "Unknown document kind")
    return "/".join(("response", row["task_id"], row["family"],
                     row["operation"], row["prompt"]))


def stats(rows: list[dict]) -> dict:
    if not rows:
        return {"pairs": 0, "repositories": 0, "median_delta": None,
                "paired_win_fraction": None, "positive_repositories": 0,
                "negative_repositories": 0, "zero_repositories": 0}
    grouped = collections.defaultdict(list)
    for row in rows:
        grouped[row["repository"]].append(row["delta"])
    means = [sum(values) / len(values) for values in grouped.values()]
    return {"pairs": len(rows), "repositories": len(grouped),
            "median_delta": statistics.median([row["delta"] for row in rows]),
            "paired_win_fraction": sum(1 if row["delta"] > 0 else
                                       0.5 if row["delta"] == 0 else 0
                                       for row in rows) / len(rows),
            "positive_repositories": sum(value > 0 for value in means),
            "negative_repositories": sum(value < 0 for value in means),
            "zero_repositories": sum(value == 0 for value in means)}


def verify_payload(inventory: dict, freeze: dict, documents: list[dict],
                   windows: list[dict], pairs: list[dict], result: dict) -> dict:
    need(len(documents) == 494 and len(windows) == 6484 and len(pairs) == 456,
         "Output coverage changed")
    expected_docs = {key(row): row for row in inventory["rows"]}
    need(len(expected_docs) == 494, "Inventory identity changed")
    by_key = {}
    sources = {}
    responses = []
    for row in documents:
        doc_key = key(row)
        need(row["key"] == doc_key and doc_key not in by_key,
             "Document identity changed")
        need(set(row).isdisjoint({"text", "ids", "token_ids"}),
             "Raw document content in result")
        prior = expected_docs.get(doc_key)
        need(prior is not None and row["text_sha256"] == prior["text_sha256"] and
             row["repository"] == prior["repository"] and
             row["cleaned_words"] == prior["cleaned_words"] and
             row["removed_header"] == prior["removed_header"] and
             row["windows"] == prior["full_coverage"]["100"]["windows"],
             "Document differs from inventory")
        need(0 <= row["score"] <= 1 and math.isfinite(row["score"]),
             "Invalid document score")
        by_key[doc_key] = row
        if row["kind"] == "source":
            need(row["task_id"] not in sources, "Duplicate source task")
            sources[row["task_id"]] = row
        else:
            responses.append(row)
    need(set(by_key) == set(expected_docs) and len(sources) == 38 and
         len(responses) == 456, "Incomplete documents")
    grouped = collections.defaultdict(list)
    seen = set()
    for row in windows:
        need(set(row) == {"key", "index", "tokens", "score"},
             "Window record shape changed")
        need(row["key"] in by_key and isinstance(row["index"], int) and
             isinstance(row["tokens"], int) and 0 < row["tokens"] <= 512 and
             isinstance(row["score"], (float, int)) and
             math.isfinite(row["score"]) and 0 <= row["score"] <= 1,
             "Invalid window")
        identity = (row["key"], row["index"])
        need(identity not in seen, "Duplicate window")
        seen.add(identity)
        grouped[row["key"]].append(row)
    for doc_key, doc in by_key.items():
        selected = sorted(grouped[doc_key], key=lambda row: row["index"])
        need([row["index"] for row in selected] == list(range(doc["windows"])),
             "Missing window")
        equal(doc["score"], sum(row["score"] for row in selected) / len(selected),
              f"document mean {doc_key}")
    reconstructed = []
    for response in responses:
        source = sources.get(response["task_id"])
        need(source is not None and source["repository"] == response["repository"],
             "Broken source-response pairing")
        reconstructed.append({
            "key": response["key"], "source_key": source["key"],
            "repository": response["repository"], "family": response["family"],
            "operation": response["operation"], "prompt": response["prompt"],
            "source_score": source["score"], "response_score": response["score"],
            "delta": response["score"] - source["score"],
            "word_ratio": response["cleaned_words"] / source["cleaned_words"],
            "source_windows": source["windows"],
            "response_windows": response["windows"]})
    equal(pairs, reconstructed, "pairs")
    expected = {"format": "unswell-editlens-technical-fullwindows-v1",
                "document_identity_sha256": freeze["document_identity_sha256"],
                "selection_sha256": freeze["selection_sha256"],
                "model_sha256": freeze["model_sha256"],
                "window_width_words": 100, "sources": 38, "responses": 456,
                "source_windows": sum(row["windows"] for row in sources.values()),
                "response_windows": sum(row["windows"] for row in responses),
                "operations": {}}
    for operation in OPERATIONS:
        subset = [row for row in reconstructed if row["operation"] == operation]
        family_stats = {family: stats([row for row in subset if row["family"] == family])
                        for family in FAMILIES}
        overall = stats(subset)
        expected["operations"][operation] = {
            "overall": overall, "families": family_stats,
            "matched_length_band": stats([row for row in subset
                                          if 0.75 <= row["word_ratio"] <= 1.33]),
            "median_word_ratio": statistics.median(row["word_ratio"] for row in subset),
            "median_source_windows": statistics.median(row["source_windows"] for row in subset),
            "median_response_windows": statistics.median(row["response_windows"] for row in subset),
            "direction_pass": (overall["pairs"] >= 217 and
                               overall["positive_repositories"] >= 17 and
                               overall["paired_win_fraction"] >= 0.70 and
                               all(row["median_delta"] > 0 for row in family_stats.values())),
        }
    equal(result, expected)
    return {"sources": len(sources), "responses": len(responses),
            "windows": len(windows),
            "generate_pass": result["operations"]["generate"]["direction_pass"],
            "polish_pass": result["operations"]["polish"]["direction_pass"]}


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", type=Path, default=HERE)
    args = parser.parse_args()
    freeze = load(HERE / "freeze.json")
    inventory = load(INVENTORY)
    need(sha(INVENTORY) == freeze["inventory_sha256"], "Inventory hash changed")
    for name, path in (("protocol_sha256", HERE / "PROTOCOL.md"),
                       ("runner_sha256", HERE / "run.py"),
                       ("verifier_sha256", HERE / "verify.py"),
                       ("runner_tests_sha256", HERE / "test_run.py"),
                       ("richness_runner_sha256", RUNS / "2026-11-07-richness-screen/run.py"),
                       ("origin_runner_sha256", RUNS / "2026-11-09-editlens-transfer/run.py"),
                       ("technical_runner_sha256", RUNS / "2026-11-09-editlens-technical-prefix/run.py"),
                       ("audit_runner_sha256", RUNS / "2026-11-10-editlens-technical-windows/audit.py"),
                       ("input_archive_sha256", STUDY / "inputs.tar.gz"),
                       ("output_archive_sha256", STUDY / "outputs.tar.gz")):
        need(sha(path) == freeze[name], f"{name} changed")
    decision = load(args.output_dir / "decision.json")
    for name in ("result.json", "documents.json", "pairs.json", "window_scores.json"):
        need(sha(args.output_dir / name) == decision["files_sha256"][name],
             f"Result file changed: {name}")
    summary = verify_payload(inventory, freeze,
                             load(args.output_dir / "documents.json"),
                             load(args.output_dir / "window_scores.json"),
                             load(args.output_dir / "pairs.json"),
                             load(args.output_dir / "result.json"))
    print(json.dumps(summary, sort_keys=True))


if __name__ == "__main__":
    main()
