#!/usr/bin/env python3
"""Make a score-blind source-quality review queue for the new controls."""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path
import re


def flags(row: dict, raw: str) -> list[str]:
    found = []
    lower_path = row["path"].lower()
    if re.search(r"(^|[/_.-])(test|tests|fixture|fixtures|generated)([/_.-]|$)", lower_path):
        found.append("path_may_be_fixture")
    if re.search(r"(^|[/_.-])(example|examples|sample|samples)([/_.-]|$)", lower_path):
        found.append("path_is_example")
    alphabetic = [char for char in raw if char.isalpha()]
    if alphabetic and sum(ord(char) > 127 for char in alphabetic) / len(alphabetic) > 0.1:
        found.append("substantial_non_ascii_letters")
    words = re.findall(r"[A-Za-z]+", raw.lower())
    common = {"the", "and", "to", "of", "in", "is", "for", "with", "that",
              "this", "you", "can", "are", "from", "on", "use"}
    if len(words) >= 50 and sum(word in common for word in words) / len(words) < 0.025:
        found.append("low_english_function_word_rate")
    if re.search(r"auto.?generated|automatically generated|generated markup|do not edit|"
                 r"\{\{Fill in", raw[:2000], re.IGNORECASE):
        found.append("generated_or_placeholder_signature")
    lines = [line for line in raw.splitlines() if line.strip()]
    if lines and sum(line.lstrip().startswith("|") for line in lines) / len(lines) > 0.5:
        found.append("mostly_markdown_table")
    if row["cleaned_words"] < 75:
        found.append("short_cleaned_prose")
    return found


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--eligibility", type=Path, required=True)
    parser.add_argument("--sources-dir", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    rows = []
    eligibility = json.loads(args.eligibility.read_text())
    for row in eligibility["rows"]:
        if not row["eligible"]:
            continue
        raw = (args.sources_dir / f"{row['source_id']}.source").read_text()
        reasons = flags(row, raw)
        if reasons:
            rows.append({"source_id": row["source_id"], "repository": row["repository"],
                         "path": row["path"], "source_sha256": row["source_sha256"],
                         "cleaned_words": row["cleaned_words"], "flags": reasons})
    output = {"format": "unswell-origin-new-control-review-queue-v1",
              "eligibility_sha256": hashlib.sha256(args.eligibility.read_bytes()).hexdigest(),
              "rows": rows}
    args.output.write_text(json.dumps(output, sort_keys=True, indent=2) + "\n")
    from collections import Counter
    print(json.dumps({"eligible": sum(row["eligible"] for row in eligibility["rows"]),
                      "queued": len(rows),
                      "flag_counts": Counter(flag for row in rows for flag in row["flags"])},
                     sort_keys=True))


if __name__ == "__main__":
    main()
