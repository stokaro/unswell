#!/usr/bin/env python3
"""Parse the fixed RACE input with the pinned English RST model."""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path
import time


INPUT_SHA256 = "feaa0b6f8fabf2f33f2ce3f4d248a1dc55d1b1b2dc3f7266bcfcdeb2043a6bb4"
PARSER_FILES = {
    "best_weights.pt": "c4008a42605d6fc2657be72ba1de50977b164f88bbc5f53b62a922f43fff3c32",
    "config.json": "b1f844b6b39f08023bfbde97c7a3075404c360c2014c5f283b47e8f9db4bea74",
    "relation_table.txt": "850c1978c57ef3c59f45ad8a120f635b20953d4fac1f98acaba858a0f61d402c",
}
TOKENIZER_FILES = {
    "config.json": "ef0185e2aae6e06c5f105a285006952c340e20c7dbf43c86ec82601b13fc45e9",
    "model.safetensors": "5bde1d28afb363d0103324efeb5afc8b2b397fe5e04beabb9b1ef355255ade81",
    "tokenizer.json": "847bbeab6174d66a88898f729d52fa8d355255ade81bea101cf960dd404581df70e",
    "tokenizer_config.json": "994f46754c5bf4014f1aa92d34b1374319c3a6b3f702105cd5b742beaecd18ce",
    "merges.txt": "1ce1664773c50f3e0cc8842619a93edc4624525b728b188a9e0be33b7726adc5",
    "vocab.json": "9e7f63c2d15d666b52e21d250d2e513b87c9b713cfa6987a82ed89e5e6e50655",
}


def digest(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            h.update(block)
    return h.hexdigest()


def require_hash(path: Path, expected: str) -> None:
    if digest(path) != expected:
        raise ValueError(f"SHA-256 mismatch: {path.name}")


def edu_count(node: dict | None) -> int:
    if not node:
        return 0
    if node.get("relation") == "elementary":
        return 1
    return edu_count(node.get("left")) + edu_count(node.get("right"))


def parse(input_path: Path, output_dir: Path, parser_dir: Path,
          backbone_dir: Path) -> dict:
    require_hash(input_path, INPUT_SHA256)
    for name, sha in PARSER_FILES.items():
        require_hash(parser_dir / name, sha)
    for name, sha in TOKENIZER_FILES.items():
        require_hash(backbone_dir / name, sha)

    from isanlp_rst.parser import Parser
    from isanlp_rst.utils import tree_to_dict
    from transformers import AutoTokenizer

    rst = Parser(model_dir=str(parser_dir), hf_model_name=None,
                 hf_model_version="rstdt", cuda_device=-1)
    tokenizer = AutoTokenizer.from_pretrained(str(backbone_dir), local_files_only=True)
    output_dir.mkdir(parents=True, exist_ok=True)
    status = []
    seen = set()
    started = time.monotonic()
    with input_path.open(encoding="utf-8") as source, \
            (output_dir / "parsed.jsonl").open("w", encoding="utf-8") as parsed:
        for number, line in enumerate(source, 1):
            if time.monotonic() - started > 2 * 3600:
                raise TimeoutError("Two-hour parse limit reached")
            row = json.loads(line)
            item_id = row.get("item_id")
            article = row.get("article")
            if not isinstance(item_id, str) or item_id in seen or \
                    not isinstance(article, str) or not article:
                raise ValueError(f"Invalid input record {number}")
            seen.add(item_id)
            if hashlib.sha256(article.encode("utf-8")).hexdigest() != row.get("text_sha256"):
                raise ValueError(f"Raw text hash mismatch at record {number}")
            token_count = len(tokenizer(article, add_special_tokens=True,
                                        truncation=False)["input_ids"])
            entry = {"item_id": item_id, "token_count": token_count,
                     "truncated": token_count > 512}
            result = rst(article)
            trees = result.get("rst") if result else None
            tree = tree_to_dict(trees[0]) if trees else None
            count = edu_count(tree)
            if count < 1:
                entry.update({"status": "no_tree", "edu_count": 0})
            else:
                row["rst_structure"] = tree
                parsed.write(json.dumps(row, ensure_ascii=False,
                                        separators=(",", ":")) + "\n")
                entry.update({"status": "parsed", "edu_count": count})
            status.append(entry)
            if number % 25 == 0:
                print(f"Parsed {number} records", flush=True)
    if len(status) != 300:
        raise ValueError(f"Expected 300 input records, got {len(status)}")
    report = {"input_sha256": INPUT_SHA256,
              "parsed_sha256": digest(output_dir / "parsed.jsonl"),
              "status": status}
    (output_dir / "parse-status.json").write_text(
        json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    return report


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--parser", type=Path, required=True)
    parser.add_argument("--backbone", type=Path, required=True)
    args = parser.parse_args()
    report = parse(args.input, args.output, args.parser, args.backbone)
    print(json.dumps({"parsed": sum(s["status"] == "parsed" for s in report["status"]),
                      "input": len(report["status"]),
                      "parsed_sha256": report["parsed_sha256"]}))


if __name__ == "__main__":
    main()
