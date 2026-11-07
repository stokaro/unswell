#!/usr/bin/env python3
"""Recompute frozen AITDNA features for post-result matched sensitivity checks."""

from __future__ import annotations

import collections
import hashlib
import importlib.util
import json
import statistics
from pathlib import Path

import duckdb


HERE = Path(__file__).resolve().parent
FEATURE = Path(
    "/private/tmp/unswell-434-editorial-tree/research/origin/runs/"
    "2026-11-07-richness-screen/run.py"
)


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def auc(positives: list[float], negatives: list[float]) -> float | None:
    if not positives or not negatives:
        return None
    wins = sum((p > n) + 0.5 * (p == n)
               for p in positives for n in negatives)
    return wins / (len(positives) * len(negatives))


def main() -> None:
    freeze = json.loads((HERE / "freeze-v2.json").read_text())
    assert freeze["script_sha256"] == sha(Path(__file__))
    assert freeze["protocol_sha256"] == sha(HERE / "PROTOCOL-v2.md")
    assert freeze["feature_sha256"] == sha(FEATURE)
    assert freeze["parquet_sha256"] == sha(HERE / "document.parquet")
    assert freeze["v1_result_sha256"] == sha(HERE / "result.json")
    assert duckdb.__version__ == "1.4.2"

    spec = importlib.util.spec_from_file_location("unswell_richness_screen", FEATURE)
    assert spec is not None and spec.loader is not None
    feature_module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(feature_module)
    rows = duckdb.connect().execute(
        "SELECT data, metadata FROM read_parquet(?)",
        [str(HERE / "document.parquet")],
    ).fetchall()
    assert len(rows) == 362
    scored = []
    for index, (data, metadata) in enumerate(rows):
        assert len(data) == 1
        snippet = data[0]
        assert snippet["author"] in ("Bot", "User")
        words = [m.group().casefold() for m in
                 feature_module.WORD.finditer(feature_module.prose(snippet["text"]))]
        assert len(words) >= 100
        values = feature_module.features(" ".join(words[:100]))
        assert values["prose_tokens"] == 100
        scored.append({"row_idx": index, "author": metadata["author"],
                       "task": metadata["task"],
                       "assisted": not metadata["human_only"],
                       "snippet_author": snippet["author"],
                       "richness": values["richness"],
                       "entropy": values["entropy"]})

    pairs = collections.defaultdict(list)
    for row in scored:
        pairs[(row["author"], row["task"])].append(row)
    matched = {}
    split = {}
    for feature in ("richness", "entropy"):
        deltas = []
        for group in pairs.values():
            human = [r[feature] for r in group if not r["assisted"]]
            assisted = [r[feature] for r in group if r["assisted"]]
            if human and assisted:
                deltas.append(statistics.mean(assisted) - statistics.mean(human))
        matched[feature] = {"pairs": len(deltas),
                            "positive": sum(d > 0 for d in deltas),
                            "negative": sum(d < 0 for d in deltas),
                            "tied": sum(d == 0 for d in deltas),
                            "median_delta": statistics.median(deltas)}
        negatives = [r[feature] for r in scored if not r["assisted"]]
        split[feature] = {}
        for label in ("Bot", "User"):
            positives = [r[feature] for r in scored
                         if r["assisted"] and r["snippet_author"] == label]
            split[feature][label] = {"assisted_count": len(positives),
                                     "human_only_count": len(negatives),
                                     "directional_auc": auc(positives, negatives)}
    result = {"format": "unswell-aitdna-mixed-authorship-sensitivity-v2",
              "post_result": True, "rows": len(scored),
              "human_only_snippet_authors": dict(sorted(collections.Counter(
                  r["snippet_author"] for r in scored if not r["assisted"]).items())),
              "assisted_snippet_authors": dict(sorted(collections.Counter(
                  r["snippet_author"] for r in scored if r["assisted"]).items())),
              "matched_author_task": matched,
              "assisted_by_snippet_author": split,
              "interpretation": "Exposed sensitivity only, with no new pass gate"}
    (HERE / "result-v2.json").write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    print(json.dumps(result, sort_keys=True))


if __name__ == "__main__":
    main()
