#!/usr/bin/env python3
"""Audit source text and EditLens window eligibility without model scores."""

from __future__ import annotations

import argparse
import collections
import hashlib
import importlib.util
import json
from pathlib import Path

from tokenizers import Tokenizer


TOKENIZER_SHA256 = "2bb1a22cfbe25b8e5a232b7fc4d7fc5073923b45724a5f813b00811bb6620f66"
HELPER_SHA256 = {
    "richness": "ed17bd5b238823357a09146e18ea871abad8a1eb3ced1a6dc2b6352e1a338ad7",
    "origin": "c411968219a2d273781769d1cbbd67e91b8dfac3bb10cc096195d1f87d19f264",
    "audit": "198b5c06cd79700fc6bec6a177146538116391b0fdf02ded096d375558cf89ae",
}


def digest(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def module(path: Path, name: str):
    if digest(path.read_bytes()) != HELPER_SHA256[name]:
        raise ValueError(f"Changed preprocessing helper: {name}")
    spec = importlib.util.spec_from_file_location(f"new_origin_{name}", path)
    if spec is None or spec.loader is None:
        raise ValueError(f"Cannot load helper: {name}")
    loaded = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(loaded)
    return loaded


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--blob-audit", type=Path, required=True)
    parser.add_argument("--sources-dir", type=Path, required=True)
    parser.add_argument("--exclusions", type=Path, required=True)
    parser.add_argument("--runs-dir", type=Path, required=True)
    parser.add_argument("--tokenizer", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if digest(args.tokenizer.read_bytes()) != TOKENIZER_SHA256:
        raise ValueError("Tokenizer changed")
    richness = module(args.runs_dir / "2026-11-07-richness-screen/run.py", "richness")
    origin = module(args.runs_dir / "2026-11-09-editlens-transfer/run.py", "origin")
    audit = module(args.runs_dir / "2026-11-10-editlens-technical-windows/audit.py", "audit")
    tokenizer = Tokenizer.from_file(str(args.tokenizer))
    blobs = json.loads(args.blob_audit.read_text())
    excluded = json.loads(args.exclusions.read_text())
    old_raw = set(excluded["historical_raw_source_sha256"])
    old_cleaned = set(excluded["prior_source_cleaned_text_sha256"])
    seen_raw: set[str] = set()
    seen_cleaned: set[str] = set()
    rows = []
    for source in blobs["rows"]:
        row = dict(source)
        reason = row.get("error")
        if reason:
            row["eligible"] = False
            row["reason"] = "blob_fetch_or_content_error"
            rows.append(row)
            continue
        raw = (args.sources_dir / f"{row['source_id']}.source").read_bytes()
        raw_sha = digest(raw)
        blob_sha = hashlib.sha1(f"blob {len(raw)}\0".encode() + raw).hexdigest()
        if (raw_sha != row["source_sha256"] or blob_sha != row["blob_sha"] or
                len(raw) != row["bytes"]):
            raise ValueError(f"Source integrity failed: {row['source_id']}")
        text = raw.decode("utf-8")
        if "\0" in text:
            raise ValueError(f"NUL in audited source: {row['source_id']}")
        cleaned, removed_header = origin.clean_text(richness.prose(text))
        clean_sha = digest(cleaned.encode())
        row.update(cleaned_sha256=clean_sha, removed_header=removed_header,
                   cleaned_words=len(origin.WORD.findall(cleaned)))
        if raw_sha in old_raw:
            reason = "previously_scored_raw_source"
        elif raw_sha in seen_raw:
            reason = "duplicate_raw_source"
        elif clean_sha in old_cleaned:
            reason = "previously_scored_cleaned_source"
        elif clean_sha in seen_cleaned:
            reason = "duplicate_cleaned_source"
        seen_raw.add(raw_sha)
        seen_cleaned.add(clean_sha)
        windows = audit.word_windows(cleaned, origin.WORD, 100)
        if windows and "".join(windows) != cleaned:
            raise ValueError(f"Incomplete window coverage: {row['source_id']}")
        lengths = [len(tokenizer.encode(window).ids) for window in windows]
        row["windows_100"] = len(windows)
        row["max_tokens_100"] = max(lengths, default=0)
        if reason is None and not lengths:
            reason = "no_cleaned_prose"
        elif reason is None and not all(0 < length <= 512 for length in lengths):
            reason = "token_length_out_of_range"
        row["eligible"] = reason is None
        row["reason"] = reason
        rows.append(row)
    output = {"format": "unswell-origin-new-historical-eligibility-v1",
              "blob_audit_sha256": digest(args.blob_audit.read_bytes()),
              "exclusions_sha256": digest(args.exclusions.read_bytes()),
              "tokenizer_sha256": TOKENIZER_SHA256,
              "helper_sha256": HELPER_SHA256,
              "source_review_pending": True,
              "rows": rows}
    args.output.write_text(json.dumps(output, sort_keys=True, indent=2) + "\n")
    counts = collections.Counter(row["reason"] or "eligible" for row in rows)
    print(json.dumps({"candidates": len(rows), "reasons": counts}, sort_keys=True))


if __name__ == "__main__":
    main()
