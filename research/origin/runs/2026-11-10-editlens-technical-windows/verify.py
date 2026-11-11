#!/usr/bin/env python3
"""Independently check source identities and aggregate window accounting."""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
STUDY = HERE.parents[2] / "generation" / "studies" / "long-prose-v3"
RUNS = HERE.parent
EXPECTED = {
    "input_archive_sha256": "af2a668913d3525cbd7cd72e2f2ac5d5a2c7189b76452b6f17b025d17d07942e",
    "output_archive_sha256": "d0a6fd2c647e9e8112f16f1d9ef13fcefe781707ec050385d248ec8eb8c0d70f",
    "tokenizer_sha256": "2bb1a22cfbe25b8e5a232b7fc4d7fc5073923b45724a5f813b00811bb6620f66",
    "richness_runner_sha256": "ed17bd5b238823357a09146e18ea871abad8a1eb3ced1a6dc2b6352e1a338ad7",
    "editlens_runner_sha256": "c411968219a2d273781769d1cbbd67e91b8dfac3bb10cc096195d1f87d19f264",
    "technical_runner_sha256": "fdc2a617ffca6eddbf4275f94f7893ebc5120628cca0b32fa62896e57ba1a4fc",
}
FILES = {
    "input_archive_sha256": STUDY / "inputs.tar.gz",
    "output_archive_sha256": STUDY / "outputs.tar.gz",
    "richness_runner_sha256": RUNS / "2026-11-07-richness-screen" / "run.py",
    "editlens_runner_sha256": RUNS / "2026-11-09-editlens-transfer" / "run.py",
    "technical_runner_sha256": RUNS / "2026-11-09-editlens-technical-prefix" / "run.py",
}
WIDTHS = ("75", "100", "150")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def check(data: dict) -> dict:
    require(data.get("format") == "unswell-editlens-technical-window-inventory-v1", "Wrong format")
    rows = data.get("rows")
    require(isinstance(rows, list) and len(rows) == 494, "Wrong row count")
    sources = {}
    responses = []
    identities = []
    seen_responses = set()
    for row in rows:
        require("text" not in row and "ids" not in row, "Raw text or token IDs leaked")
        require(isinstance(row.get("text_sha256"), str) and len(row["text_sha256"]) == 64,
                "Missing text hash")
        require(isinstance(row.get("cleaned_words"), int) and row["cleaned_words"] >= 50,
                "Ineligible document")
        require(isinstance(row.get("isolated_paragraph_windows"), int) and
                row["isolated_paragraph_windows"] >= 0, "Invalid paragraph count")
        for width in WIDTHS:
            window = row.get("full_coverage", {}).get(width, {})
            require(isinstance(window.get("windows"), int) and window["windows"] > 0,
                    "Invalid window count")
            require(isinstance(window.get("over_cap"), int) and
                    0 <= window["over_cap"] <= window["windows"], "Invalid over-cap count")
            require(isinstance(window.get("max_tokens"), int) and window["max_tokens"] > 0,
                    "Invalid token maximum")
            require(window.get("complete") is (window["over_cap"] == 0 and
                    window["max_tokens"] <= 512), "False complete window coverage")
        if row.get("kind") == "source":
            require(row["task_id"] not in sources, "Duplicate source task")
            sources[row["task_id"]] = row
            identities.append(("source", row["task_id"], row["text_sha256"]))
        elif row.get("kind") == "response":
            key = (row["task_id"], row.get("family"), row.get("operation"), row.get("prompt"))
            require(key not in seen_responses, "Duplicate response identity")
            require(key[1] in ("haiku", "luna", "qwen") and
                    key[2] in ("generate", "polish") and
                    key[3] in ("neutral", "plain"), "Invalid response condition")
            seen_responses.add(key)
            responses.append(row)
            identities.append(("response", *key, row["text_sha256"]))
        else:
            raise ValueError("Invalid document kind")
    require(len(sources) == 38 and len(responses) == 456, "Source/response count changed")
    require(all(sum(row["task_id"] == task for row in responses) == 12
                for task in sources), "Incomplete source-response pairing")
    require(data.get("sources") == 38 and data.get("responses") == 456,
            "Wrong top-level counts")
    identity_hash = hashlib.sha256(json.dumps(identities, ensure_ascii=False,
                                           separators=(",", ":")).encode()).hexdigest()
    require(data.get("inputs", {}).get("document_identity_sha256") == identity_hash,
            "Document identities changed")
    paragraph = {
        "sources_at_least_three": sum(row["isolated_paragraph_windows"] >= 3 for row in sources.values()),
        "sources_at_least_five": sum(row["isolated_paragraph_windows"] >= 5 for row in sources.values()),
    }
    for operation in ("generate", "polish"):
        subset = [row for row in responses if row["operation"] == operation]
        require(len(subset) == 228, "Operation count changed")
        for count in (3, 5):
            paragraph[f"{operation}_paired_at_least_{'three' if count == 3 else 'five'}"] = sum(
                row["isolated_paragraph_windows"] >= count and
                sources[row["task_id"]]["isolated_paragraph_windows"] >= count
                for row in subset
            )
    require(data.get("isolated_paragraphs") == paragraph, "Paragraph aggregate changed")
    full = {}
    for width in WIDTHS:
        summary = {
            "source_complete": sum(row["full_coverage"][width]["complete"] for row in sources.values()),
            "response_complete": sum(row["full_coverage"][width]["complete"] for row in responses),
            "source_windows": sum(row["full_coverage"][width]["windows"] for row in sources.values()),
            "response_windows": sum(row["full_coverage"][width]["windows"] for row in responses),
            "response_over_cap_windows": sum(row["full_coverage"][width]["over_cap"] for row in responses),
        }
        for operation in ("generate", "polish"):
            summary[operation + "_paired_complete"] = sum(
                row["full_coverage"][width]["complete"] and
                sources[row["task_id"]]["full_coverage"][width]["complete"]
                for row in responses if row["operation"] == operation
            )
        full[width] = summary
    require(data.get("full_coverage") == full, "Full-coverage aggregate changed")
    return {"sources": len(sources), "responses": len(responses),
            "paragraph_five_sources": paragraph["sources_at_least_five"],
            "complete_100_word_pairs": full["100"]["generate_paired_complete"] +
                                       full["100"]["polish_paired_complete"]}


def verify(result: Path, tokenizer: Path) -> dict:
    data = json.loads(result.read_text())
    summary = check(data)
    files = dict(FILES, tokenizer_sha256=tokenizer)
    for name, expected in EXPECTED.items():
        require(data["inputs"].get(name) == expected, f"Unpinned input: {name}")
        require(digest(files[name]) == expected, f"Input hash changed: {name}")
    return summary


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--result", type=Path, default=HERE / "result.json")
    parser.add_argument("--tokenizer", type=Path, required=True)
    args = parser.parse_args()
    print(json.dumps(verify(args.result, args.tokenizer), sort_keys=True))


if __name__ == "__main__":
    main()
