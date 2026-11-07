#!/usr/bin/env python3
"""Score frozen AITDNA document rows with existing Unswell origin features."""

from __future__ import annotations

import collections
import hashlib
import importlib.util
import json
import statistics
from pathlib import Path


HERE = Path(__file__).resolve().parent
OFFSETS = (0, 100, 200, 300)
TOTAL = 362
HEAD = "21d77f7d44833507750c0da309295c4147c1bda5"
FEATURE = Path(
    "/private/tmp/unswell-434-editorial-tree/research/origin/runs/"
    "2026-11-07-richness-screen/run.py"
)


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def load_json(path: Path):
    return json.loads(path.read_text())


def require(value, message: str) -> None:
    if not value:
        raise ValueError(message)


def auc(rows: list[dict], feature: str) -> float | None:
    positive = [r[feature] for r in rows if r["assisted"]]
    negative = [r[feature] for r in rows if not r["assisted"]]
    if not positive or not negative:
        return None
    wins = sum((p > n) + 0.5 * (p == n)
               for p in positive for n in negative)
    return wins / (len(positive) * len(negative))


def paired_directions(rows: list[dict], feature: str) -> dict:
    authors: dict[str, list[dict]] = collections.defaultdict(list)
    for row in rows:
        authors[row["author"]].append(row)
    directions = collections.Counter()
    for group in authors.values():
        assisted = [r[feature] for r in group if r["assisted"]]
        human = [r[feature] for r in group if not r["assisted"]]
        if not assisted or not human:
            continue
        delta = statistics.mean(assisted) - statistics.mean(human)
        directions["positive" if delta > 0 else "negative" if delta < 0 else "tied"] += 1
    return {key: directions[key] for key in ("positive", "negative", "tied")}


def main() -> None:
    freeze = load_json(HERE / "freeze.json")
    acquisition = load_json(HERE / "acquisition.json")
    require(freeze["dataset_head"] == HEAD, "Dataset HEAD changed in freeze")
    require(acquisition["head_before"] == HEAD and
            acquisition["head_after"] == HEAD, "Dataset HEAD changed during acquisition")
    require(freeze["script_sha256"] == sha(Path(__file__)), "Scorer changed")
    require(freeze["protocol_sha256"] == sha(HERE / "PROTOCOL.md"), "Protocol changed")
    require(freeze["feature_sha256"] == sha(FEATURE), "Feature source changed")

    spec = importlib.util.spec_from_file_location("unswell_richness_screen", FEATURE)
    require(spec is not None and spec.loader is not None, "Feature import unavailable")
    feature_module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(feature_module)

    all_rows = []
    features_schema = None
    for offset in OFFSETS:
        path = HERE / f"rows-{offset:03d}.json"
        require(acquisition["pages"][str(offset)] == sha(path), "Raw page changed")
        page = load_json(path)
        require(page["num_rows_total"] == TOTAL and page["partial"] is False,
                "Incomplete dataset page")
        require(page["num_rows_per_page"] == 100, "Page length differs")
        if features_schema is None:
            features_schema = page["features"]
        require(page["features"] == features_schema, "Feature schema differs")
        expected = min(100, TOTAL - offset)
        require(len(page["rows"]) == expected, "Row count differs")
        for i, item in enumerate(page["rows"]):
            require(item["row_idx"] == offset + i, "Row index differs")
            all_rows.append(item["row"])
    require(len(all_rows) == TOTAL, "Total row count differs")

    scored = []
    abstentions = []
    snippet_authors = collections.Counter()
    for index, row in enumerate(all_rows):
        metadata = row["metadata"]
        parts = row["data"]
        require(isinstance(parts, list) and parts, "Missing document text")
        require(isinstance(metadata["author"], str) and metadata["author"],
                "Missing participant identity")
        require(type(metadata["human_only"]) is bool, "Missing setting label")
        require(isinstance(metadata["task"], str) and metadata["task"],
                "Missing task")
        text_parts = []
        bot_words = 0
        total_words = 0
        for part in parts:
            require(isinstance(part["text"], str) and
                    isinstance(part["author"], str), "Malformed snippet")
            snippet_authors[part["author"]] += 1
            text_parts.append(part["text"])
            count = len(feature_module.WORD.findall(feature_module.prose(part["text"])))
            total_words += count
            if part["author"] == "Bot":
                bot_words += count
        text = "\n".join(text_parts)
        words = [m.group().casefold()
                 for m in feature_module.WORD.finditer(feature_module.prose(text))]
        require(len(words) == total_words, "Snippet word count mismatch")
        if len(words) < 100:
            abstentions.append({"row_idx": index, "reason": "under_100_prose_tokens"})
            continue
        values = feature_module.features(" ".join(words[:100]))
        require(values["prose_tokens"] == 100 and values["richness"] is not None
                and values["entropy"] is not None, "Feature unavailable")
        scored.append({
            "row_idx": index,
            "author": metadata["author"],
            "task": metadata["task"],
            "assisted": not metadata["human_only"],
            "prose_tokens": len(words),
            "bot_word_fraction": bot_words / total_words,
            "richness": values["richness"],
            "entropy": values["entropy"],
        })

    tasks = sorted({r["task"] for r in scored})
    def class_counts(group: list[dict]) -> dict:
        return {"human_only": sum(not r["assisted"] for r in group),
                "ai_assisted": sum(r["assisted"] for r in group)}

    task_results = {}
    for task in tasks:
        group = [r for r in scored if r["task"] == task]
        task_results[task] = {"counts": class_counts(group),
                              "richness_auc": auc(group, "richness"),
                              "entropy_auc": auc(group, "entropy")}
    pairs = {feature: paired_directions(scored, feature)
             for feature in ("richness", "entropy")}
    overall = {feature + "_auc": auc(scored, feature)
               for feature in ("richness", "entropy")}
    require(all(value is not None for value in overall.values()),
            "Missing one study class")
    screen = all(overall[f + "_auc"] >= 0.60 and
                 sum(result[f + "_auc"] is not None and result[f + "_auc"] >= 0.60
                     for result in task_results.values()) >= 2 and
                 pairs[f]["positive"] >= 0.60 * sum(pairs[f].values())
                 for f in ("richness", "entropy"))
    result = {
        "format": "unswell-aitdna-mixed-authorship-screen-v1",
        "dataset_head_observed": HEAD,
        "source_identity_limit": "Dataset-server pages are hash-frozen but lack commit binding",
        "raw_page_sha256": acquisition["pages"],
        "rows_total": TOTAL,
        "scored": len(scored),
        "abstentions": len(abstentions),
        "abstention_rows": abstentions,
        "authors": len({r["author"] for r in scored}),
        "snippet_author_counts": dict(sorted(snippet_authors.items())),
        "counts": class_counts(scored),
        "median_prose_tokens": {
            "human_only": statistics.median(r["prose_tokens"] for r in scored if not r["assisted"]),
            "ai_assisted": statistics.median(r["prose_tokens"] for r in scored if r["assisted"]),
        },
        "median_bot_word_fraction_assisted": statistics.median(
            r["bot_word_fraction"] for r in scored if r["assisted"]),
        "overall": overall,
        "by_task": task_results,
        "paired_author_directions": pairs,
        "feasibility_pass": screen,
        "interpretation": "Study-setting contrast, not machine-authorship accuracy or editorial value",
    }
    (HERE / "result.json").write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    print(json.dumps({"scored": result["scored"], "abstentions": result["abstentions"],
                      "counts": result["counts"], "overall": overall,
                      "paired": pairs, "screen": screen}, sort_keys=True))


if __name__ == "__main__":
    main()
