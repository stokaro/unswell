#!/usr/bin/env python3
"""Audit complete EditLens window availability on the saved technical corpus."""

from __future__ import annotations

import argparse
import hashlib
import importlib.util
import json
from pathlib import Path
import re

from tokenizers import Tokenizer


HERE = Path(__file__).resolve().parent
ORIGIN = HERE.parent
RICHNESS = ORIGIN / "2026-11-07-richness-screen" / "run.py"
EDITLENS = ORIGIN / "2026-11-09-editlens-transfer" / "run.py"
TECHNICAL = ORIGIN / "2026-11-09-editlens-technical-prefix" / "run.py"
TOKENIZER_SHA256 = "2bb1a22cfbe25b8e5a232b7fc4d7fc5073923b45724a5f813b00811bb6620f66"
WIDTHS = (75, 100, 150)
MIN_WORDS = 50
MAX_TOKENS = 512


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def load(path: Path, name: str):
    spec = importlib.util.spec_from_file_location(name, path)
    if spec is None or spec.loader is None:
        raise ValueError(f"Missing helper: {path}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def word_windows(cleaned: str, word_pattern: re.Pattern[str], width: int) -> list[str]:
    """Cover all cleaned words with disjoint windows, folding a short tail backward."""
    words = list(word_pattern.finditer(cleaned))
    if len(words) < MIN_WORDS:
        return []
    ranges = []
    for start in range(0, len(words), width):
        end = min(start + width, len(words))
        if end - start < MIN_WORDS and ranges:
            previous_start, _ = ranges[-1]
            ranges[-1] = (previous_start, end)
        else:
            ranges.append((start, end))
    if sum(end - start for start, end in ranges) != len(words):
        raise ValueError("Window coverage changed")
    if any(start >= end or (index and start != ranges[index - 1][1])
           for index, (start, end) in enumerate(ranges)):
        raise ValueError("Window boundaries changed")
    windows = [cleaned[0 if index == 0 else words[start].start():
                       len(cleaned) if index == len(ranges) - 1 else words[end].start()]
               for index, (start, end) in enumerate(ranges)]
    if "".join(windows) != cleaned:
        raise ValueError("Cleaned text coverage changed")
    return windows


def paragraph_count(text: str, richness, origin, tokenizer: Tokenizer) -> int:
    count = 0
    for paragraph in re.split(r"\n\s*\n", richness.prose(text)):
        cleaned, _ = origin.clean_text(paragraph)
        words = len(origin.WORD.findall(cleaned))
        if MIN_WORDS <= words <= 300 and len(tokenizer.encode(cleaned).ids) <= MAX_TOKENS:
            count += 1
    return count


def inspect(doc: dict, richness, origin, tokenizer: Tokenizer) -> dict:
    cleaned, removed_header = origin.clean_text(richness.prose(doc["text"]))
    word_count = len(origin.WORD.findall(cleaned))
    row = {key: value for key, value in doc.items() if key != "text"}
    row.update(cleaned_words=word_count, removed_header=removed_header,
               isolated_paragraph_windows=paragraph_count(doc["text"], richness, origin, tokenizer))
    rows = {}
    for width in WIDTHS:
        windows = word_windows(cleaned, origin.WORD, width)
        token_counts = [len(tokenizer.encode(window).ids) for window in windows]
        rows[str(width)] = {
            "windows": len(windows),
            "over_cap": sum(count > MAX_TOKENS for count in token_counts),
            "max_tokens": max(token_counts, default=0),
            "complete": bool(windows) and all(count <= MAX_TOKENS for count in token_counts),
        }
    row["full_coverage"] = rows
    return row


def summarize(rows: list[dict]) -> dict:
    sources = {row["task_id"]: row for row in rows if row["kind"] == "source"}
    responses = [row for row in rows if row["kind"] == "response"]
    if len(sources) != 38 or len(responses) != 456:
        raise ValueError("Corpus identity changed")
    result = {
        "format": "unswell-editlens-technical-window-inventory-v1",
        "sources": 38,
        "responses": 456,
        "isolated_paragraphs": {
            "sources_at_least_three": sum(row["isolated_paragraph_windows"] >= 3 for row in sources.values()),
            "sources_at_least_five": sum(row["isolated_paragraph_windows"] >= 5 for row in sources.values()),
        },
        "full_coverage": {},
    }
    for width in WIDTHS:
        label = str(width)
        summary = {
            "source_complete": sum(row["full_coverage"][label]["complete"] for row in sources.values()),
            "response_complete": sum(row["full_coverage"][label]["complete"] for row in responses),
            "source_windows": sum(row["full_coverage"][label]["windows"] for row in sources.values()),
            "response_windows": sum(row["full_coverage"][label]["windows"] for row in responses),
            "response_over_cap_windows": sum(row["full_coverage"][label]["over_cap"] for row in responses),
        }
        for operation in ("generate", "polish"):
            subset = [row for row in responses if row["operation"] == operation]
            summary[operation + "_paired_complete"] = sum(
                row["full_coverage"][label]["complete"] and
                sources[row["task_id"]]["full_coverage"][label]["complete"]
                for row in subset
            )
        result["full_coverage"][label] = summary
    for operation in ("generate", "polish"):
        subset = [row for row in responses if row["operation"] == operation]
        result["isolated_paragraphs"][operation + "_paired_at_least_three"] = sum(
            row["isolated_paragraph_windows"] >= 3 and
            sources[row["task_id"]]["isolated_paragraph_windows"] >= 3
            for row in subset
        )
        result["isolated_paragraphs"][operation + "_paired_at_least_five"] = sum(
            row["isolated_paragraph_windows"] >= 5 and
            sources[row["task_id"]]["isolated_paragraph_windows"] >= 5
            for row in subset
        )
    return result


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--tokenizer", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if digest(args.tokenizer) != TOKENIZER_SHA256:
        raise ValueError("Tokenizer identity changed")
    richness = load(RICHNESS, "technical_window_richness")
    origin = load(EDITLENS, "technical_window_origin")
    technical = load(TECHNICAL, "technical_window_technical")
    tokenizer = Tokenizer.from_file(str(args.tokenizer))
    docs, identity = technical.load_docs(richness)
    rows = [inspect(doc, richness, origin, tokenizer) for doc in docs]
    result = summarize(rows)
    result["inputs"] = {
        "document_identity_sha256": identity,
        "input_archive_sha256": richness.digest((richness.STUDY / "inputs.tar.gz").read_bytes()),
        "output_archive_sha256": richness.digest((richness.STUDY / "outputs.tar.gz").read_bytes()),
        "tokenizer_sha256": TOKENIZER_SHA256,
        "richness_runner_sha256": digest(RICHNESS),
        "editlens_runner_sha256": digest(EDITLENS),
        "technical_runner_sha256": digest(TECHNICAL),
    }
    result["rows"] = rows
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    print(json.dumps({key: value for key, value in result.items() if key != "rows"}, sort_keys=True))


if __name__ == "__main__":
    main()
