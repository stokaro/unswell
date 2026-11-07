#!/usr/bin/env python3
"""Apply the separately frozen equal-window sensitivity screen."""

from __future__ import annotations

import collections
import hashlib
import json
import math
import statistics

from run import HERE, WORD, digest, prose, read_json, verified_archive


def windows(text: str) -> list[tuple[float, float]]:
    words = [m.group().casefold() for m in WORD.finditer(prose(text))]
    result = []
    for i in range(0, len(words) - 50 + 1, 25):
        counts = collections.Counter(words[i:i + 50])
        richness = len(counts) / 50
        entropy = -sum((n / 50) * math.log2(n / 50) for n in counts.values()) / math.log2(50)
        result.append((richness, entropy))
    return result


def selected(values: list[tuple[float, float]], n: int, method: str) -> list[tuple[float, float]]:
    assert n > 0 and n <= len(values)
    if method == "prefix":
        return values[:n]
    assert method == "even"
    if n == 1:
        return values[:1]
    indices = [round(i * (len(values) - 1) / (n - 1)) for i in range(n)]
    assert len(set(indices)) == n
    return [values[i] for i in indices]


def summarize(rows: list[dict], feature: str) -> dict:
    values = [r[feature] for r in rows]
    by_repo: dict[str, list[float]] = collections.defaultdict(list)
    for row in rows:
        by_repo[row["repository"]].append(row[feature])
    means = [statistics.mean(v) for _, v in sorted(by_repo.items())]
    return {"count": len(values), "repositories": len(means),
            "median_delta": statistics.median(values) if values else None,
            "positive_repositories": sum(x > 0 for x in means),
            "negative_repositories": sum(x < 0 for x in means),
            "zero_repositories": sum(x == 0 for x in means)}


def main() -> None:
    with verified_archive("inputs") as inputs:
        tasks_data = read_json(inputs, "tasks.json")["tasks"]
    tasks = {t["id"]: t for t in tasks_data}
    assert len(tasks) == 38 and len({t["repository"] for t in tasks.values()}) == 20
    source_windows = {tid: windows(t["text"]) for tid, t in tasks.items()}
    rows: dict[str, list[dict]] = {method: [] for method in ("prefix", "even")}
    abstentions = 0
    with verified_archive("outputs") as outputs:
        names = sorted(m.name for m in outputs.getmembers()
                       if m.name.endswith("/result.json") and "/raw/" in m.name)
        assert len(names) == 456
        for name in names:
            parent = name.rsplit("/", 1)[0]
            request = read_json(outputs, parent + "/input.json")["request"]
            response = read_json(outputs, name)
            task = tasks[request["task_id"]]
            source = source_windows[task["id"]]
            target = windows(response["text"])
            n = min(len(source), len(target))
            if n == 0:
                abstentions += 1
                continue
            for method in rows:
                left, right = selected(source, n, method), selected(target, n, method)
                rows[method].append({
                    "repository": task["repository"],
                    "operation": request["operation"],
                    "family": name.split("/")[1],
                    "richness": statistics.mean(v[0] for v in right) - statistics.mean(v[0] for v in left),
                    "entropy": statistics.mean(v[1] for v in right) - statistics.mean(v[1] for v in left),
                })
    assert all(len(v) + abstentions == 456 for v in rows.values())
    result = {"format": "unswell-richness-equal-window-v1", "responses": 456,
              "abstentions": abstentions, "conditions": {}}
    for method in ("prefix", "even"):
        for operation in ("generate", "polish"):
            subset = [r for r in rows[method] if r["operation"] == operation]
            result["conditions"][f"{operation}/{method}"] = {
                "richness": summarize(subset, "richness"),
                "entropy": summarize(subset, "entropy"),
                "families": {
                    f: {feature: summarize([r for r in subset if r["family"] == f], feature)
                        for feature in ("richness", "entropy")}
                    for f in ("haiku", "luna", "qwen")
                },
            }
    result["robust_directional_pass"] = all(
        result["conditions"][f"generate/{method}"][feature]["positive_repositories"] >= 15
        and all(result["conditions"][f"generate/{method}"]["families"][family][feature]["median_delta"] > 0
                for family in ("haiku", "luna", "qwen"))
        for method in ("prefix", "even") for feature in ("richness", "entropy")
    )
    body = json.dumps(result, indent=2, sort_keys=True) + "\n"
    (HERE / "equal-length-result.json").write_text(body)
    print(hashlib.sha256(body.encode()).hexdigest())


if __name__ == "__main__":
    main()
