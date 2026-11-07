#!/usr/bin/env python3
"""Replay a frozen, model-free lexical feature screen on saved documents."""

from __future__ import annotations

import collections
import hashlib
import json
import math
import re
import statistics
import tarfile
from pathlib import Path


HERE = Path(__file__).resolve().parent
STUDY = HERE.parents[2] / "generation" / "studies" / "long-prose-v3"
WORD = re.compile(r"[A-Za-z]+(?:'[A-Za-z]+)?")
URL = re.compile(r"https?://\S+")
INLINE_CODE = re.compile(r"`[^`\n]*`")
HTML_TAG = re.compile(r"<[^>\n]*>")
WINDOW = 50
STRIDE = 25


def digest(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def verified_archive(name: str) -> tarfile.TarFile:
    wrapper = json.loads((STUDY / f"{name}.json").read_text())
    path = STUDY / wrapper["file"]
    data = path.read_bytes()
    assert len(data) == wrapper["bytes"], name
    assert digest(data) == wrapper["sha256"], name
    return tarfile.open(path, "r:gz")


def read_json(archive: tarfile.TarFile, name: str):
    member = archive.extractfile(name)
    assert member is not None, name
    return json.load(member)


def prose(text: str) -> str:
    lines: list[str] = []
    fence: str | None = None
    for line in text.splitlines():
        stripped = line.lstrip()
        marker = stripped[:3]
        if fence is not None:
            if marker == fence:
                fence = None
            continue
        if marker in ("```", "~~~"):
            fence = marker
            continue
        lines.append(line)
    cleaned = "\n".join(lines)
    cleaned = INLINE_CODE.sub(" ", cleaned)
    cleaned = URL.sub(" ", cleaned)
    return HTML_TAG.sub(" ", cleaned)


def features(text: str) -> dict:
    cleaned = prose(text)
    words = [m.group().casefold() for m in WORD.finditer(cleaned)]
    raw_count = len(WORD.findall(text))
    assert raw_count >= len(words), (raw_count, len(words))
    values = {"prose_tokens": len(words), "raw_tokens": raw_count,
              "prose_fraction": len(words) / raw_count if raw_count else None}
    if len(words) < WINDOW:
        values.update(richness=None, entropy=None)
        return values
    windows = [words[i:i + WINDOW]
               for i in range(0, len(words) - WINDOW + 1, STRIDE)]
    richness = []
    entropy = []
    for window in windows:
        counts = collections.Counter(window)
        richness.append(len(counts) / WINDOW)
        entropy.append(-sum((n / WINDOW) * math.log2(n / WINDOW)
                            for n in counts.values()) / math.log2(WINDOW))
    values.update(richness=statistics.mean(richness),
                  entropy=statistics.mean(entropy))
    assert WORD.findall(cleaned) == WORD.findall(re.sub(r"\s+", " ", cleaned))
    return values


def summary(rows: list[dict], feature: str) -> dict:
    deltas = [r[feature] for r in rows]
    by_repo: dict[str, list[float]] = collections.defaultdict(list)
    for row in rows:
        by_repo[row["repository"]].append(row[feature])
    repo_deltas = [statistics.mean(v) for _, v in sorted(by_repo.items())]
    return {
        "count": len(deltas),
        "repositories": len(repo_deltas),
        "median_delta": statistics.median(deltas) if deltas else None,
        "median_repository_mean_delta": statistics.median(repo_deltas) if repo_deltas else None,
        "positive_repositories": sum(v > 0 for v in repo_deltas),
        "negative_repositories": sum(v < 0 for v in repo_deltas),
        "zero_repositories": sum(v == 0 for v in repo_deltas),
    }


def main() -> None:
    with verified_archive("inputs") as inputs:
        selection = read_json(inputs, "selection.json")
        task_data = read_json(inputs, "tasks.json")
    assert len(selection) == 38 and len(task_data["tasks"]) == 38
    selected = {row["source_id"]: row for row in selection}
    tasks = {task["id"]: task for task in task_data["tasks"]}
    assert len(tasks) == 38 and len({t["repository"] for t in tasks.values()}) == 20
    sources = {}
    for task in tasks.values():
        row = selected[task["source_id"]]
        assert digest(task["text"].encode("utf-8")) == row["source_sha256"]
        assert task["repository"] == row["repository"]
        sources[task["id"]] = features(task["text"])

    rows = []
    identities = set()
    with verified_archive("outputs") as outputs:
        members = sorted(m.name for m in outputs.getmembers()
                         if m.name.endswith("/result.json") and "/raw/" in m.name)
        assert len(members) == 456, len(members)
        for result_name in members:
            parent = result_name.rsplit("/", 1)[0]
            request = read_json(outputs, parent + "/input.json")["request"]
            response = read_json(outputs, result_name)
            task = tasks[request["task_id"]]
            family = result_name.split("/")[1]
            operation = request["operation"]
            prompt = request["prompt"]
            assert family in {"haiku", "luna", "qwen"}
            assert operation in {"generate", "polish"}
            assert prompt in {"neutral", "plain"}
            identity = (task["id"], family, operation, prompt)
            assert identity not in identities, identity
            identities.add(identity)
            response_features = features(response["text"])
            source_features = sources[task["id"]]
            ratio = response_features["prose_tokens"] / source_features["prose_tokens"]
            rows.append({
                "repository": task["repository"], "family": family,
                "operation": operation, "prompt": prompt,
                "length_ratio": ratio,
                "richness": response_features["richness"] - source_features["richness"]
                if response_features["richness"] is not None and source_features["richness"] is not None else None,
                "entropy": response_features["entropy"] - source_features["entropy"]
                if response_features["entropy"] is not None and source_features["entropy"] is not None else None,
                "prose_fraction": response_features["prose_fraction"] - source_features["prose_fraction"],
            })
    assert len(identities) == 456
    assert all(r["richness"] is not None and r["entropy"] is not None for r in rows)
    result = {"format": "unswell-richness-screen-v1", "sources": len(tasks),
              "repositories": 20, "responses": len(rows), "abstentions": 0,
              "conditions": {}}
    for operation in ("generate", "polish"):
        for scope in ("all", "length_matched"):
            subset = [r for r in rows if r["operation"] == operation
                      and (scope == "all" or 0.75 <= r["length_ratio"] <= 1.33)]
            assert subset
            key = f"{operation}/{scope}"
            result["conditions"][key] = {
                "count": len(subset),
                "length_ratio_median": statistics.median(r["length_ratio"] for r in subset),
                "richness": summary(subset, "richness"),
                "entropy": summary(subset, "entropy"),
                "prose_fraction_median_delta": statistics.median(r["prose_fraction"] for r in subset),
                "families": {f: {"richness": summary([r for r in subset if r["family"] == f], "richness"),
                                 "entropy": summary([r for r in subset if r["family"] == f], "entropy")}
                             for f in ("haiku", "luna", "qwen")},
            }
    generated = result["conditions"]["generate/all"]
    result["feasibility_pass"] = (generated["richness"]["positive_repositories"] >= 15
                                  and generated["entropy"]["positive_repositories"] >= 15
                                  and all(generated["families"][f][feature]["median_delta"] > 0
                                          for f in ("haiku", "luna", "qwen")
                                          for feature in ("richness", "entropy")))
    body = json.dumps(result, indent=2, sort_keys=True) + "\n"
    (HERE / "result.json").write_text(body)
    print(digest(body.encode("utf-8")))


if __name__ == "__main__":
    main()
