#!/usr/bin/env python3
"""Measure frozen GEN thresholds on saved technical text and academic rewrites."""

from __future__ import annotations

import argparse
import collections
import hashlib
import importlib.util
import json
import statistics
import tarfile
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]
FEATURE = ROOT / "research/origin/runs/2026-11-07-richness-screen/run.py"
GEN = ROOT / "research/origin/runs/2026-11-08-gen-edit-transfer/result.json"
STUDY = ROOT / "research/generation/studies/long-prose-v3"
EXPECTED = {
    ("train", "human"): 627, ("train", "ai"): 1125,
    ("val", "human"): 83, ("val", "ai"): 131,
    ("test", "human"): 89, ("test", "ai"): 143,
    ("human_only", "human"): 400,
}
GENERATORS = (
    "claude-4-8-opus-genai", "claude-5-1-fable-genai",
    "claude-5-5-opus-genai", "gemini-3-6-flash-genai",
    "gemini-3-8-flash-genai", "gpt-5-6-sol-genai-responses",
    "gpt-6-astra-genai",
)
PROMPTS = ("arb", "imbd", "shield", "v2")
FAMILIES = ("haiku", "luna", "qwen")


def require(ok: bool, message: str) -> None:
    if not ok:
        raise ValueError(message)


def sha(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def load_features():
    spec = importlib.util.spec_from_file_location("unswell_technical_features", FEATURE)
    require(spec is not None and spec.loader is not None, "Feature source unavailable")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def prefix100(text: str, feature) -> dict | None:
    words = [match.group().casefold()
             for match in feature.WORD.finditer(feature.prose(text))]
    if len(words) < 100:
        return None
    values = feature.features(" ".join(words[:100]))
    require(values["prose_tokens"] == 100, "Prefix token count changed")
    return {name: values[name] for name in ("richness", "entropy")}


def flags(score: dict | None, thresholds: dict) -> dict | None:
    if score is None:
        return None
    return {name: score[name] >= threshold
            for name, threshold in thresholds.items()}


def summarize(rows: list[dict], thresholds: dict) -> dict:
    eligible = [row for row in rows if row["score"] is not None]
    result = {"total": len(rows), "eligible": len(eligible),
              "abstained": len(rows) - len(eligible)}
    for name in thresholds:
        result[name] = {"flagged": sum(flags(row["score"], thresholds)[name]
                                       for row in eligible),
                        "fraction": (sum(flags(row["score"], thresholds)[name]
                                         for row in eligible) / len(eligible)
                                     if eligible else None)}
    return result


def paired(rows: list[dict], thresholds: dict) -> dict:
    eligible = [row for row in rows if row["score"] is not None
                and row["source_score"] is not None]
    result = summarize(rows, thresholds)
    result["eligible_pairs"] = len(eligible)
    result["source_abstentions"] = sum(row["source_score"] is None for row in rows)
    for name in thresholds:
        result[name]["paired_wins"] = (sum(
            (row["score"][name] > row["source_score"][name])
            + 0.5 * (row["score"][name] == row["source_score"][name])
            for row in eligible) / len(eligible) if eligible else None)
    return result


def archive(name: str, freeze: dict) -> tarfile.TarFile:
    wrapper_path = STUDY / f"{name}.json"
    require(sha(wrapper_path) == freeze["sha256"][name + "_wrapper"],
            f"{name} wrapper changed")
    wrapper = json.loads(wrapper_path.read_text())
    path = STUDY / wrapper["file"]
    require(path.stat().st_size == wrapper["bytes"], f"{name} archive size changed")
    require(sha(path) == wrapper["sha256"] == freeze["sha256"][name + "_archive"],
            f"{name} archive changed")
    return tarfile.open(path, "r:gz")


def member_json(archive_file: tarfile.TarFile, name: str):
    member = archive_file.extractfile(name)
    require(member is not None, f"Missing {name}")
    return json.load(member)


def technical(feature, thresholds: dict, freeze: dict) -> dict:
    with archive("inputs", freeze) as inputs:
        selection = member_json(inputs, "selection.json")
        task_data = member_json(inputs, "tasks.json")["tasks"]
    require(len(selection) == len(task_data) == 38, "Technical source count changed")
    selected = {row["source_id"]: row for row in selection}
    tasks = {task["id"]: task for task in task_data}
    require(len(selected) == len(tasks) == 38, "Duplicate technical source")
    require(len({task["repository"] for task in tasks.values()}) == 20,
            "Repository count changed")
    sources = {}
    source_rows = []
    for task in tasks.values():
        row = selected[task["source_id"]]
        require(hashlib.sha256(task["text"].encode()).hexdigest() == row["source_sha256"],
                "Source text changed")
        require(task["repository"] == row["repository"], "Repository mismatch")
        score = prefix100(task["text"], feature)
        sources[task["id"]] = score
        source_rows.append({"repository": task["repository"], "score": score})
    rows = []
    identities = set()
    with archive("outputs", freeze) as outputs:
        names = sorted(m.name for m in outputs.getmembers()
                       if m.name.endswith("/result.json") and "/raw/" in m.name)
        require(len(names) == 456, "Technical response count changed")
        for name in names:
            parent = name.rsplit("/", 1)[0]
            request = member_json(outputs, parent + "/input.json")["request"]
            response = member_json(outputs, name)
            require(request["task_id"] in tasks, "Unknown task")
            family = name.split("/")[1]
            operation = request["operation"]
            prompt = request["prompt"]
            require(family in FAMILIES and operation in ("generate", "polish")
                    and prompt in ("neutral", "plain"), "Unexpected cell")
            identity = (request["task_id"], family, operation, prompt)
            require(identity not in identities, "Duplicate response")
            identities.add(identity)
            require(isinstance(response["text"], str), "Missing response text")
            rows.append({"task_id": request["task_id"],
                         "repository": tasks[request["task_id"]]["repository"],
                         "family": family, "operation": operation,
                         "prompt": prompt,
                         "source_score": sources[request["task_id"]],
                         "score": prefix100(response["text"], feature)})
    require(len(identities) == 456, "Missing response cell")
    result = {"sources": summarize(source_rows, thresholds), "operations": {}}
    for operation in ("generate", "polish"):
        subset = [row for row in rows if row["operation"] == operation]
        require(len(subset) == 228, "Operation count changed")
        result["operations"][operation] = {
            "overall": paired(subset, thresholds),
            "by_family": {key: paired([row for row in subset if row["family"] == key],
                                     thresholds) for key in FAMILIES},
            "by_prompt": {key: paired([row for row in subset if row["prompt"] == key],
                                     thresholds) for key in ("neutral", "plain")},
            "by_repository": {key: paired([row for row in subset if row["repository"] == key],
                                         thresholds)
                              for key in sorted({row["repository"] for row in subset})},
        }
    generated = [row for row in rows if row["operation"] == "generate"
                 and row["score"] is not None and row["source_score"] is not None]
    deltas = collections.defaultdict(list)
    for row in generated:
        deltas[row["repository"]].append(row["score"]["richness"]
                                        - row["source_score"]["richness"])
    result["generate_positive_repository_means"] = sum(
        statistics.mean(values) > 0 for values in deltas.values())
    result["generate_eligible_repositories"] = len(deltas)
    return result


def academic(path: Path, feature, thresholds: dict) -> dict:
    counts = collections.Counter()
    papers = collections.defaultdict(set)
    variants = set()
    originals = collections.defaultdict(dict)
    rewrites = []
    for number, line in enumerate(path.open(encoding="utf-8"), 1):
        row = json.loads(line)
        split = row["split"]
        human = row["variant"] == "human"
        label = "human" if human else "ai"
        require((split, label) in EXPECTED, f"Unexpected split at row {number}")
        require(isinstance(row["paper_id"], str) and row["paper_id"],
                f"Missing paper ID at row {number}")
        require(isinstance(row["text"], str) and row["text"],
                f"Missing text at row {number}")
        key = (row["paper_id"], row["variant"])
        require(key not in variants, f"Duplicate variant at row {number}")
        variants.add(key)
        if human:
            require(row["generator"] is None and row["prompt"] is None,
                    f"Human metadata at row {number}")
        else:
            require(row["generator"] in GENERATORS and row["prompt"] in PROMPTS,
                    f"AI metadata at row {number}")
        counts[split, label] += 1
        papers[split].add(row["paper_id"])
        if human and split in ("test", "val", "human_only"):
            originals[split][row["paper_id"]] = {"score": prefix100(row["text"], feature)}
        elif not human and split == "test":
            rewrites.append({"paper_id": row["paper_id"], "generator": row["generator"],
                             "prompt": row["prompt"],
                             "score": prefix100(row["text"], feature)})
    require(counts == EXPECTED, f"Academic row counts changed: {dict(counts)}")
    splits = sorted(papers)
    for i, left in enumerate(splits):
        for right in splits[i + 1:]:
            require(not papers[left] & papers[right], "Paper split overlap")
    for row in rewrites:
        require(row["paper_id"] in originals["test"], "Missing matched test paper")
        row["source_score"] = originals["test"][row["paper_id"]]["score"]
    return {
        "rows": sum(counts.values()),
        "counts": {f"{split}/{label}": count for (split, label), count in sorted(counts.items())},
        "human": {split: summarize(list(originals[split].values()), thresholds)
                  for split in ("test", "val", "human_only")},
        "test_rewrites": {
            "overall": paired(rewrites, thresholds),
            "by_generator": {key: paired([row for row in rewrites if row["generator"] == key],
                                        thresholds) for key in GENERATORS},
            "by_prompt": {key: paired([row for row in rewrites if row["prompt"] == key],
                                     thresholds) for key in PROMPTS},
        },
    }


def decide(technical_result: dict, academic_result: dict) -> dict:
    generation = technical_result["operations"]["generate"]
    overall = generation["overall"]
    source = technical_result["sources"]
    families = generation["by_family"]
    human_only = academic_result["human"]["human_only"]
    gates = {
        "technical_pair_coverage_at_least_80_percent": overall["eligible_pairs"] >= 0.8 * 228,
        "technical_generate_flag_at_least_10_percent": (
            overall["richness"]["flagged"] >= 0.1 * overall["eligible"]),
        "technical_original_flags_at_most_one": source["richness"]["flagged"] <= 1
        and source["eligible"] == 38,
        "each_family_flag_at_least_5_percent": all(
            item["richness"]["flagged"] >= 0.05 * item["eligible"]
            and item["eligible"] > 0 for item in families.values()),
        "positive_repository_means_at_least_15": (
            technical_result["generate_positive_repository_means"] >= 15),
        "academic_human_only_flag_at_most_2_percent": (
            human_only["eligible"] > 0 and human_only["richness"]["flagged"]
            <= 0.02 * human_only["eligible"]),
    }
    return {"gates": gates, "advance_screen": all(gates.values())}


def run(dataset: Path) -> dict:
    freeze = json.loads((HERE / "freeze.json").read_text())
    require(freeze["format"] == "unswell-gen-technical-transfer-freeze-v1",
            "Freeze format changed")
    paths = {"protocol": HERE / "PROTOCOL.md", "runner": Path(__file__),
             "feature": FEATURE, "gen_result": GEN, "dataset": dataset}
    for name, path in paths.items():
        require(sha(path) == freeze["sha256"][name], f"{name} changed")
    require(dataset.stat().st_size == freeze["dataset_bytes"], "Dataset size changed")
    gen = json.loads(GEN.read_text())
    thresholds = {name: gen["tasks"]["generated"][f"prefix100/{name}"]["threshold"]
                  for name in ("richness", "entropy")}
    require(thresholds == freeze["thresholds"], "GEN threshold changed")
    feature = load_features()
    technical_result = technical(feature, thresholds, freeze)
    academic_result = academic(dataset, feature, thresholds)
    return {"format": "unswell-gen-technical-transfer-v1",
            "freeze_sha256": sha(HERE / "freeze.json"),
            "thresholds": thresholds,
            "technical": technical_result, "academic": academic_result,
            "decision": decide(technical_result, academic_result)}


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--dataset", type=Path, required=True)
    parser.add_argument("--output", type=Path, default=HERE / "result.json")
    args = parser.parse_args()
    result = run(args.dataset)
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    print(f"Wrote {args.output}; advance screen: {result['decision']['advance_screen']}")


if __name__ == "__main__":
    main()
