#!/usr/bin/env python3
"""Summarize saved before/after CLI reports without rerunning any rule."""

import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path


def load(path):
    data = path.read_bytes()
    result = json.loads(data)
    if result["status"] != "complete" or result["errors"]:
        raise ValueError(f"incomplete report: {path}")
    sources = {row["name"]: row for row in result["documents"]}
    if len(sources) != len(result["documents"]):
        raise ValueError(f"duplicate source: {path}")
    counts = Counter(row["rule_id"] for row in result["findings"])
    pages = {name: Counter() for name in sources}
    for row in result["findings"]:
        pages[row["primary"]["path"]][row["rule_id"]] += 1
    summary = {
        "report_sha256": hashlib.sha256(data).hexdigest(),
        "documents": len(sources),
        "prose_words": sum(row["prose_words"] for row in sources.values()),
        "findings": len(result["findings"]),
        "by_rule": dict(sorted(counts.items())),
        "pages": [{"path": name, "source_sha256": sources[name]["source_hash"],
                   "words": sources[name]["prose_words"],
                   "by_rule": dict(sorted(pages[name].items()))}
                  for name in sorted(sources)],
    }
    return summary


def compare(before, after):
    left, right = load(before), load(after)
    for field in ("documents", "prose_words"):
        if left[field] != right[field]:
            raise ValueError(f"changed {field}; this comparison assumes unchanged extraction")
    bindings = lambda rows: [(r["path"], r["source_sha256"], r["words"]) for r in rows]
    if bindings(left["pages"]) != bindings(right["pages"]):
        raise ValueError("source or extraction bindings changed")
    return {"before": left, "after": right}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("before", type=Path)
    parser.add_argument("after", type=Path)
    args = parser.parse_args()
    result = {"version": "unswell-contextual-comparison-v1",
              "interpretation": "Descriptive warning load; no quality labels or confirmatory inference.",
              **compare(args.before, args.after)}
    print(json.dumps(result, indent=2, sort_keys=True))


if __name__ == "__main__":
    main()
