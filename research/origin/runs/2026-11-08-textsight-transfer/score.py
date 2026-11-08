#!/usr/bin/env python3
"""Replay frozen ARB origin scores on the pinned TextSight control set."""

from __future__ import annotations

import argparse
import collections
import hashlib
import importlib.util
import json
import math
from pathlib import Path


HERE = Path(__file__).resolve().parent
FILES = {
    "wikipedia": ("human_wikipedia_pre2022.jsonl", 155, "human"),
    "pmc": ("human_pmc_ccby_nonanglophone.jsonl", 183, "human"),
    "literature": ("human_literature_publicdomain.jsonl", 306, "human"),
    "ai": ("ai_generated_2026-09-29.jsonl", 56, "ai"),
}
REVISION = "423b1768f0827675295fd6dc3cd255ad2f411183"
PINNED = {
    "ensemble_run": "55b552ee93690beaaa800548aef6d9e94b7cdb6b90270d00d630d7f24cbc5737",
    "ensemble_result": "3805be64972e79ee6487b32b0788755610e9a107014f1cd0e8ba3310d24f5eb0",
    "richness_run": "ed17bd5b238823357a09146e18ea871abad8a1eb3ced1a6dc2b6352e1a338ad7",
    "lexicon": "37ab010f5b2410df5c419d0291a55e86ab829ee1b26df2be361bd49b445f9be0",
}
THRESHOLDS = {"ensemble": 1.3251545797561963, "richness": 0.9199999999999999}
GENERATORS = {"claude-opus-5": 12, "claude-haiku-4-5": 12, "gpt-4.1": 12,
              "gpt-oss-120b": 12, "qwen-3.8-27b": 8}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def load_module(path: Path, name: str):
    spec = importlib.util.spec_from_file_location(name, path)
    require(spec is not None and spec.loader is not None, "module unavailable")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def sources(repo: Path) -> dict[str, Path]:
    root = repo / "research/origin/runs"
    return {
        "ensemble_run": root / "2026-11-08-interpretable-ensemble/run.py",
        "ensemble_result": root / "2026-11-08-interpretable-ensemble/result.json",
        "richness_run": root / "2026-11-07-richness-screen/run.py",
        "lexicon": root / "2026-11-08-gen-edit-transfer/closed_class.txt",
    }


def identities(repo: Path, data: Path) -> dict:
    source_paths = sources(repo)
    for name, expected in PINNED.items():
        require(digest(source_paths[name]) == expected, "source drift: " + name)
    files = {}
    for group, (name, _, _) in FILES.items():
        path = data / name
        require(path.is_file(), "missing data: " + name)
        files[group] = {"name": name, "sha256": digest(path), "bytes": path.stat().st_size}
    return {"format": "unswell-textsight-transfer-freeze-v1", "dataset_revision": REVISION,
            "protocol_sha256": digest(HERE / "PROTOCOL.md"),
            "scorer_sha256": digest(Path(__file__)),
            "source_sha256": PINNED, "data": files}


def read_rows(data: Path) -> dict[str, list[dict]]:
    groups = {}
    ids = set()
    for group, (name, expected, label) in FILES.items():
        rows = []
        for line_number, line in enumerate((data / name).read_text().splitlines(), 1):
            row = json.loads(line)
            require(row.get("label") == label and isinstance(row.get("text"), str)
                    and row["text"].strip() and isinstance(row.get("id"), str)
                    and row["id"], f"invalid {group} row {line_number}")
            require(row["id"] not in ids, "duplicate dataset id")
            ids.add(row["id"])
            rows.append(row)
        require(len(rows) == expected, f"changed {group} count")
        groups[group] = rows
    counts = collections.Counter(row.get("source") for row in groups["ai"])
    require(counts == GENERATORS, "generator counts changed")
    return groups


def auc(positive: list[float], negative: list[float]) -> float | None:
    if not positive or not negative:
        return None
    wins = sum(p > n for p in positive for n in negative)
    ties = sum(p == n for p in positive for n in negative)
    return (wins + 0.5 * ties) / (len(positive) * len(negative))


def count(rows: list[dict], key: str, threshold: float) -> dict:
    values = [row["scores"][key] for row in rows if row["scores"] is not None]
    return {"total": len(rows), "eligible": len(values),
            "abstained": len(rows) - len(values),
            "flagged": sum(value > threshold for value in values)}


def score(repo: Path, data: Path) -> dict:
    freeze = json.loads((HERE / "freeze.json").read_text())
    require(freeze == identities(repo, data), "frozen identity changed")
    rows = read_rows(data)
    paths = sources(repo)
    ensemble = load_module(paths["ensemble_run"], "unswell_ensemble")
    richness = load_module(paths["richness_run"], "unswell_richness")
    prior = json.loads(paths["ensemble_result"].read_text())
    require({name: prior["thresholds"][name]["score"] for name in THRESHOLDS}
            == THRESHOLDS, "threshold drift")
    closed = set(paths["lexicon"].read_text().splitlines())
    for group in rows.values():
        for row in group:
            vector = ensemble.feature_vector(row["text"], richness, closed)
            row["scores"] = ensemble.scores(vector, prior["fit"])
            require(row["scores"] is None or all(math.isfinite(value)
                    for value in row["scores"].values()), "nonfinite score")
    primary = rows["wikipedia"] + rows["pmc"]
    result = {"format": "unswell-textsight-transfer-result-v1",
              "freeze_sha256": digest(HERE / "freeze.json"),
              "thresholds": THRESHOLDS, "scores": {}}
    for name, threshold in THRESHOLDS.items():
        groups = {group: count(items, name, threshold) for group, items in rows.items()}
        generators = {source: count([row for row in rows["ai"] if row["source"] == source],
                                    name, threshold) for source in sorted(GENERATORS)}
        humans = [row["scores"][name] for row in primary if row["scores"] is not None]
        machines = [row["scores"][name] for row in rows["ai"] if row["scores"] is not None]
        human_flags = groups["wikipedia"]["flagged"] + groups["pmc"]["flagged"]
        passed = (len(humans) >= math.ceil(0.8 * 338) and
                  len(machines) >= math.ceil(0.8 * 56) and
                  human_flags / len(humans) <= 0.01 and
                  all(groups[g]["flagged"] / groups[g]["eligible"] <= 0.02
                      for g in ("wikipedia", "pmc")) and
                  groups["ai"]["flagged"] / len(machines) >= 0.10)
        result["scores"][name] = {
            "groups": groups, "generators": generators,
            "primary_human_fpr": human_flags / len(humans),
            "ai_tpr": groups["ai"]["flagged"] / len(machines),
            "unpaired_auroc": auc(machines, humans),
            "advance_pass": passed,
        }
    return result


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("phase", choices=("freeze", "score"))
    parser.add_argument("--repo", type=Path, required=True)
    parser.add_argument("--data", type=Path, required=True)
    args = parser.parse_args()
    if args.phase == "freeze":
        require(not (HERE / "freeze.json").exists(), "freeze exists")
        read_rows(args.data)
        output = identities(args.repo, args.data)
        (HERE / "freeze.json").write_text(json.dumps(output, indent=2, sort_keys=True) + "\n")
        print("frozen", digest(HERE / "freeze.json"))
    else:
        require(not (HERE / "result.json").exists(), "result exists")
        output = score(args.repo, args.data)
        (HERE / "result.json").write_text(json.dumps(output, indent=2, sort_keys=True) + "\n")
        print(json.dumps({name: {"advance_pass": row["advance_pass"],
                                "ai_tpr": row["ai_tpr"],
                                "primary_human_fpr": row["primary_human_fpr"]}
                          for name, row in output["scores"].items()}, sort_keys=True))


if __name__ == "__main__":
    main()
