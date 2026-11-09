#!/usr/bin/env python3
"""Score frozen long-prose-v3 technical prefixes with pinned EditLens."""

from __future__ import annotations

import argparse
import collections
import hashlib
import importlib.util
import json
import statistics
from pathlib import Path

import numpy as np
from tokenizers import Tokenizer


HERE = Path(__file__).resolve().parent
FAMILIES = ("haiku", "luna", "qwen")
OPERATIONS = ("generate", "polish")


def sha(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def load_module(path: Path, name: str):
    spec = importlib.util.spec_from_file_location(name, path)
    require(spec is not None and spec.loader is not None, f"Cannot load {name}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def load_docs(richness) -> tuple[list[dict], str]:
    with richness.verified_archive("inputs") as inputs:
        selection = richness.read_json(inputs, "selection.json")
        task_data = richness.read_json(inputs, "tasks.json")
    require(len(selection) == len(task_data["tasks"]) == 38, "Source count mismatch")
    selected = {item["source_id"]: item for item in selection}
    tasks = {item["id"]: item for item in task_data["tasks"]}
    require(len(tasks) == 38 and len({item["repository"] for item in tasks.values()}) == 20,
            "Source identity mismatch")
    docs = []
    identities = []
    for task_id, task in sorted(tasks.items()):
        source = selected[task["source_id"]]
        require(richness.digest(task["text"].encode()) == source["source_sha256"],
                "Source text hash mismatch")
        require(task["repository"] == source["repository"], "Repository mismatch")
        docs.append({"kind": "source", "task_id": task_id,
                     "repository": task["repository"], "text": task["text"],
                     "text_sha256": source["source_sha256"]})
        identities.append(("source", task_id, source["source_sha256"]))
    seen = set()
    with richness.verified_archive("outputs") as outputs:
        members = sorted(member.name for member in outputs.getmembers()
                         if member.name.endswith("/result.json") and "/raw/" in member.name)
        require(len(members) == 456, "Response count mismatch")
        for name in members:
            parent = name.rsplit("/", 1)[0]
            request = richness.read_json(outputs, parent + "/input.json")["request"]
            response = richness.read_json(outputs, name)
            task_id = request["task_id"]
            require(task_id in tasks, "Unknown source task")
            family = name.split("/")[1]
            operation = request["operation"]
            prompt = request["prompt"]
            require(family in FAMILIES and operation in OPERATIONS and
                    prompt in ("neutral", "plain"), "Unknown response condition")
            identity = (task_id, family, operation, prompt)
            require(identity not in seen, "Duplicate response identity")
            seen.add(identity)
            text = response["text"]
            digest = richness.digest(text.encode())
            docs.append({"kind": "response", "task_id": task_id,
                         "repository": tasks[task_id]["repository"],
                         "family": family, "operation": operation,
                         "prompt": prompt, "text": text,
                         "text_sha256": digest})
            identities.append(("response", *identity, digest))
    require(len(docs) == 494 and len(seen) == 456, "Incomplete source/response coverage")
    canonical = json.dumps(identities, ensure_ascii=False, separators=(",", ":"))
    return docs, hashlib.sha256(canonical.encode()).hexdigest()


def prepare(docs: list[dict], tokenizer: Tokenizer, richness, origin) -> list[dict]:
    prepared = []
    for doc in docs:
        prose = richness.prose(doc["text"])
        cleaned, removed = origin.clean_text(prose)
        words = len(origin.WORD.findall(cleaned))
        raw_words = len(richness.WORD.findall(doc["text"]))
        prose_words = len(richness.WORD.findall(prose))
        require(raw_words >= prose_words, "Invalid code stripping")
        ids = tokenizer.encode(cleaned).ids if words >= 50 else []
        item = {key: value for key, value in doc.items() if key != "text"}
        item.update(words=words, raw_words=raw_words, prose_words=prose_words,
                    prose_fraction=prose_words / raw_words if raw_words else None,
                    token_count=len(ids), truncated=len(ids) > 512,
                    removed_header=removed, ids=ids[:512], score=None)
        prepared.append(item)
    return prepared


def stats(rows: list[dict]) -> dict:
    if not rows:
        return {"eligible_pairs": 0, "repositories": 0, "median_delta": None,
                "paired_win_fraction": None, "positive_repositories": 0,
                "negative_repositories": 0, "zero_repositories": 0}
    deltas = [row["delta"] for row in rows]
    by_repo = collections.defaultdict(list)
    for row in rows:
        by_repo[row["repository"]].append(row["delta"])
    means = [statistics.mean(values) for values in by_repo.values()]
    return {"eligible_pairs": len(rows), "repositories": len(by_repo),
            "median_delta": statistics.median(deltas),
            "paired_win_fraction": sum((delta > 0) + (delta == 0) / 2
                                       for delta in deltas) / len(deltas),
            "positive_repositories": sum(value > 0 for value in means),
            "negative_repositories": sum(value < 0 for value in means),
            "zero_repositories": sum(value == 0 for value in means)}


def summarize(prepared: list[dict]) -> tuple[dict, list[dict]]:
    source = {item["task_id"]: item for item in prepared if item["kind"] == "source"}
    response = [item for item in prepared if item["kind"] == "response"]
    require(len(source) == 38 and len(response) == 456, "Scored row count mismatch")
    paired = []
    for item in response:
        original = source[item["task_id"]]
        if item["score"] is None or original["score"] is None:
            continue
        paired.append({"task_id": item["task_id"], "repository": item["repository"],
                       "family": item["family"], "operation": item["operation"],
                       "prompt": item["prompt"], "delta": item["score"] - original["score"],
                       "response_score": item["score"], "source_score": original["score"],
                       "word_ratio": item["words"] / original["words"],
                       "source_truncated": original["truncated"],
                       "response_truncated": item["truncated"]})
    result = {
        "sources": len(source), "responses": len(response),
        "eligible_sources": sum(item["score"] is not None for item in source.values()),
        "eligible_responses": sum(item["score"] is not None for item in response),
        "truncated_sources": sum(item["truncated"] for item in source.values()),
        "truncated_responses": sum(item["truncated"] for item in response),
        "header_removed_sources": sum(item["removed_header"] for item in source.values()),
        "header_removed_responses": sum(item["removed_header"] for item in response),
        "median_prose_fraction_sources": statistics.median(item["prose_fraction"] for item in source.values()),
        "median_prose_fraction_responses": statistics.median(item["prose_fraction"] for item in response),
        "operations": {},
    }
    for operation in OPERATIONS:
        subset = [item for item in paired if item["operation"] == operation]
        untruncated = [item for item in subset if not item["source_truncated"]
                       and not item["response_truncated"]]
        result["operations"][operation] = {
            "overall": stats(subset),
            "families": {family: stats([item for item in subset if item["family"] == family])
                         for family in FAMILIES},
            "median_word_ratio": statistics.median(item["word_ratio"] for item in subset),
            "length_band_count": sum(0.75 <= item["word_ratio"] <= 1.33 for item in subset),
            "untruncated": stats(untruncated),
            "direction_pass": len(subset) >= 217 and
                              stats(subset)["positive_repositories"] >= 15 and
                              all(stats([item for item in subset if item["family"] == family])["median_delta"] > 0
                                  for family in FAMILIES),
        }
    return result, paired


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--richness-loader", type=Path, required=True)
    parser.add_argument("--origin-runner", type=Path, required=True)
    parser.add_argument("--model-dir", type=Path, required=True)
    parser.add_argument("--preflight", action="store_true")
    args = parser.parse_args()
    richness = load_module(args.richness_loader, "pinned_richness")
    origin = load_module(args.origin_runner, "pinned_origin")
    docs, identity_hash = load_docs(richness)
    if args.preflight:
        print(json.dumps({"documents": len(docs), "identities_sha256": identity_hash},
                         sort_keys=True))
        return
    freeze = json.loads((HERE / "freeze.json").read_text())
    for name, path in (("protocol_sha256", HERE / "PROTOCOL.md"),
                       ("runner_sha256", HERE / "run.py"),
                       ("richness_loader_sha256", args.richness_loader),
                       ("origin_runner_sha256", args.origin_runner)):
        require(sha(path) == freeze[name], f"{name} mismatch")
    require(identity_hash == freeze["identities_sha256"], "Document identities changed")
    model = args.model_dir / "onnx/model.onnx"
    require(sha(model) == freeze["model_sha256"], "Model changed")
    for filename, digest in freeze["model_file_sha256"].items():
        require(sha(args.model_dir / filename) == digest, f"{filename} changed")
    tokenizer = Tokenizer.from_file(str(args.model_dir / "tokenizer.json"))
    prepared = prepare(docs, tokenizer, richness, origin)
    origin.score(prepared, model)
    result, paired = summarize(prepared)
    result["identities_sha256"] = identity_hash
    (HERE / "result.json").write_text(json.dumps(result, sort_keys=True, indent=2) + "\n")
    rows = [{key: value for key, value in item.items() if key != "ids"}
            for item in prepared]
    (HERE / "scores.json").write_text(json.dumps(rows, sort_keys=True, indent=2) + "\n")
    (HERE / "pairs.json").write_text(json.dumps(paired, sort_keys=True, indent=2) + "\n")
    print(json.dumps({"generate": result["operations"]["generate"],
                      "polish": result["operations"]["polish"]}, sort_keys=True))


if __name__ == "__main__":
    main()
