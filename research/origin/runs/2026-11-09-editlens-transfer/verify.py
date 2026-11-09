#!/usr/bin/env python3
"""Independently verify aggregate claims from stored EditLens row scores."""

from __future__ import annotations

import collections
import hashlib
import json
import statistics
from pathlib import Path


HERE = Path(__file__).resolve().parent


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def close(actual: float, expected: float) -> None:
    if abs(actual - expected) > 1e-12:
        raise ValueError(f"Aggregate mismatch: {actual} versus {expected}")


def auc_brute(pairs: list[tuple[float, float]]) -> float:
    human = [value for value, _ in pairs]
    machine = [value for _, value in pairs]
    wins = sum((m > h) + (m == h) / 2 for m in machine for h in human)
    return wins / (len(pairs) ** 2)


def check_metrics(group: list[tuple[dict, dict]], reported: dict,
                  threshold: float) -> None:
    pairs = [(human["score"], machine["score"]) for human, machine in group]
    if len(pairs) != reported["eligible_pairs"]:
        raise ValueError("Pair denominator mismatch")
    if len(pairs) >= 20:
        close(auc_brute(pairs), reported["auc"])
    elif reported["auc"] is not None:
        raise ValueError("AUROC reported for too few pairs")
    if not pairs:
        if reported["paired_win_fraction"] is not None or reported["tpr"] is not None:
            raise ValueError("Metric reported with no pairs")
        return
    close(sum((machine > human) + (machine == human) / 2
              for human, machine in pairs) / len(pairs), reported["paired_win_fraction"])
    flags = sum(machine > threshold for _, machine in pairs)
    if flags != reported["flagged"]:
        raise ValueError("Machine flag count mismatch")
    close(flags / len(pairs), reported["tpr"])


def verify(root: Path = HERE) -> None:
    freeze = json.loads((root / "freeze.json").read_text())
    if sha(root / "PROTOCOL.md") != freeze["protocol_sha256"]:
        raise ValueError("Protocol changed")
    if sha(root / "run.py") != freeze["runner_sha256"]:
        raise ValueError("Runner changed")
    rows = json.loads((root / "scores.json").read_text())
    result = json.loads((root / "result.json").read_text())
    if result["selection_sha256"] != freeze["selection_sha256"]:
        raise ValueError("Selection digest mismatch")
    if result["dataset_sha256"] != freeze["data_sha256"]:
        raise ValueError("Dataset digest mismatch")
    if result["model_revision"] != freeze["model_revision"]:
        raise ValueError("Model revision mismatch")
    if len(rows) != 1200:
        raise ValueError("Wrong row count")
    groups = collections.defaultdict(dict)
    counts = collections.Counter()
    for row in rows:
        key = (row["role"], row["regime"])
        counts[key] += 1
        if row["role"] not in ("calibration", "test"):
            raise ValueError("Unknown role")
        if row["domain"] not in ("owt", "wp", "xsum"):
            raise ValueError("Unknown domain")
        if row["score"] is not None and not (0 <= row["score"] <= 1):
            raise ValueError("Invalid score")
        if row["score"] is None and row["words"] >= 50:
            raise ValueError("Unexplained abstention")
        if row["score"] is not None and row["words"] < 50:
            raise ValueError("Ineligible score")
        slot = groups[(row["role"], row["pair_id"])]
        if row["regime"] in slot:
            raise ValueError("Duplicate regime")
        slot[row["regime"]] = row
    if counts != {("calibration", "human"): 300,
                  ("test", "human"): 300,
                  ("test", "h2l"): 300,
                  ("test", "free_llm"): 300}:
        raise ValueError("Regime count mismatch")
    for key in counts:
        subset = [row for row in rows if (row["role"], row["regime"]) == key]
        reported = result["counts"]["/".join(key)]
        actual = {
            "raw": len(subset),
            "eligible": sum(row["score"] is not None for row in subset),
            "truncated": sum(row["truncated"] for row in subset),
            "removed_header": sum(row["removed_header"] for row in subset),
            "median_words": float(statistics.median(row["words"] for row in subset)),
        }
        if actual != reported:
            raise ValueError("Eligibility or truncation count mismatch")
    for (role, _), slot in groups.items():
        expected = {"human"} if role == "calibration" else {"human", "h2l", "free_llm"}
        if set(slot) != expected:
            raise ValueError("Incomplete source group")
        if role == "test" and slot["h2l"]["generator"] != slot["free_llm"]["generator"]:
            raise ValueError("Generator mismatch")
        if any(item["domain"] != slot["human"]["domain"] for item in slot.values()):
            raise ValueError("Domain mismatch")
    calibration = [slot["human"]["score"] for (role, _), slot in groups.items()
                   if role == "calibration" and slot["human"]["score"] is not None]
    threshold = sorted(calibration, reverse=True)[2]
    close(threshold, result["threshold"])
    if len(calibration) != result["calibration"]["eligible"]:
        raise ValueError("Calibration denominator mismatch")
    if sum(value > threshold for value in calibration) != result["calibration"]["flagged"]:
        raise ValueError("Calibration flags mismatch")
    test_human = [slot["human"]["score"] for (role, _), slot in groups.items()
                  if role == "test" and slot["human"]["score"] is not None]
    if len(test_human) != result["test_human"]["eligible"]:
        raise ValueError("Test human denominator mismatch")
    if sum(value > threshold for value in test_human) != result["test_human"]["flagged"]:
        raise ValueError("Test human flags mismatch")
    close(sum(value > threshold for value in test_human) / len(test_human),
          result["test_human"]["fpr"])
    for regime in ("h2l", "free_llm"):
        matched = [(slot["human"], slot[regime])
                   for (role, _), slot in groups.items()
                   if role == "test" and slot["human"]["score"] is not None
                   and slot[regime]["score"] is not None]
        reported = result["regimes"][regime]["overall"]
        check_metrics(matched, reported, threshold)
        if set(result["regimes"][regime]["by_domain"]) != {"owt", "wp", "xsum"}:
            raise ValueError("Domain subgroup set mismatch")
        if set(result["regimes"][regime]["by_generator"]) != set((
                "gemma2_9b", "llama32_3b", "mistral7b", "qwen25_7b")):
            raise ValueError("Generator subgroup set mismatch")
        for domain, subgroup in result["regimes"][regime]["by_domain"].items():
            if domain not in ("owt", "wp", "xsum"):
                raise ValueError("Unknown domain subgroup")
            check_metrics([(human, machine) for human, machine in matched
                           if machine["domain"] == domain], subgroup, threshold)
        for generator, subgroup in result["regimes"][regime]["by_generator"].items():
            if generator not in ("gemma2_9b", "llama32_3b", "mistral7b", "qwen25_7b"):
                raise ValueError("Unknown generator subgroup")
            check_metrics([(human, machine) for human, machine in matched
                           if machine["generator"] == generator], subgroup, threshold)
    h2l = result["regimes"]["h2l"]["overall"]
    free = result["regimes"]["free_llm"]["overall"]
    passed = all((len(calibration) >= 285, len(test_human) >= 285,
                  h2l["eligible_pairs"] >= 285, free["eligible_pairs"] >= 285,
                  result["test_human"]["fpr"] <= 0.02,
                  h2l["auc"] >= 0.75, h2l["tpr"] >= 0.20,
                  free["auc"] >= 0.85, free["tpr"] >= 0.50))
    if passed != result["feasibility_pass"]:
        raise ValueError("Feasibility decision mismatch")
    print("EditLens score and aggregate verification passed")


if __name__ == "__main__":
    verify()
