#!/usr/bin/env python3
"""Verify frozen origin-ensemble evidence without loading raw corpus text."""

from __future__ import annotations

import hashlib
import json
import math
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]
RESULT_SHA = "3805be64972e79ee6487b32b0788755610e9a107014f1cd0e8ba3310d24f5eb0"
FREEZE_SHA = "0a24dd751f258ecd42012b23b45575fdbc4e0b38f552b843c700a360a8e7d074"
POSTHOC_SHA = "621e0f1cb87ed93f21fad47571997a20acfab3931171695ebb2fac5b87fa5d1a"
DEPENDENCIES = {
    "protocol": HERE / "PROTOCOL.md",
    "script": HERE / "run.py",
    "richness": ROOT / "research/origin/runs/2026-11-07-richness-screen/run.py",
    "lexicon": ROOT / "research/origin/runs/2026-11-08-gen-edit-transfer/closed_class.txt",
    "arb_loader": ROOT / "research/origin/runs/2026-11-08-arb-transfer/run.py",
}


def require(value: bool, reason: str) -> None:
    if not value:
        raise ValueError(reason)


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def bounded(value) -> bool:
    return type(value) in (float, int) and math.isfinite(value) and 0 <= value <= 1


def verify(directory: Path = HERE) -> None:
    freeze_path = directory / "freeze.json"
    result_path = directory / "result.json"
    require(sha(freeze_path) == FREEZE_SHA, "Freeze hash mismatch")
    require(sha(result_path) == RESULT_SHA, "Result hash mismatch")
    freeze = json.loads(freeze_path.read_text())
    result = json.loads(result_path.read_text())
    dependencies = dict(DEPENDENCIES)
    dependencies["protocol"] = directory / "PROTOCOL.md"
    dependencies["script"] = directory / "run.py"
    for name, path in dependencies.items():
        require(sha(path) == freeze["sha256"][name],
                f"{name} hash mismatch")
    require(freeze["sha256"]["arb"] ==
            "66e48240618b92ce876e2097d8df52f06661a70e31791b91a95b15688a099ab0",
            "ARB identity mismatch")
    require(freeze["sha256"]["academic"] ==
            "428de968a71d64ddaf9f0d5afe7e4e295b2adeace77f57fd29c6252ae082d29c",
            "Academic identity mismatch")
    require(freeze["split_sha256"] ==
            "63dc835b4586260da00a4df7a14df8af60c273df2f8bb7e6596a9f09ddc1683e",
            "Split identity mismatch")
    require(freeze["arb_roles"] ==
            {"train": 600, "calibration": 300, "test": 900},
            "ARB split denominator mismatch")
    require(freeze["academic_counts"] == {
        "human_only/human": 400, "test/ai": 143, "test/human": 89,
        "train/ai": 1125, "train/human": 627,
        "val/ai": 131, "val/human": 83,
    }, "Academic split denominator mismatch")
    require(result["inputs"] == freeze, "Result input identity mismatch")
    require(result["format"] == "unswell-interpretable-ensemble-v1" and
            result["feature_names"] == [
                "richness", "entropy", "closed_fraction",
                "mean_length_tenths", "repeated_bigram_coverage"],
            "Feature contract mismatch")
    require(570 <= result["training"]["source_groups"] <= 600 and
            result["training"]["source_groups"] +
            result["training"]["excluded_source_groups"] == 600,
            "Training source denominator mismatch")
    fit = result["fit"]
    for field in ("means", "deviations", "weights"):
        require(len(fit[field]) == 5 and
                all(type(value) in (int, float) and math.isfinite(value)
                    for value in fit[field]), f"Invalid fit {field}")
    require(all(value > 0 for value in fit["deviations"]),
            "Invalid standard deviation")
    for name in ("ensemble", "richness"):
        threshold = result["thresholds"][name]
        require(285 <= threshold["eligible"] <= 300 and
                0 <= threshold["flagged"] <=
                math.ceil(0.01 * threshold["eligible"]) - 1,
                "Invalid calibration")
        block = result["scores"][name]
        arb = block["arb"]
        academic = block["academic"]
        for human, total in ((arb["human"], 900),
                             (academic["test_human"], 89),
                             (academic["human_only"], 400)):
            require(human["total"] == total and
                    0 <= human["eligible"] <= total and
                    0 <= human["flagged"] <= human["eligible"] and
                    bounded(human["fpr"]) and
                    math.isclose(human["fpr"],
                                 human["flagged"] / human["eligible"],
                                 abs_tol=1e-12),
                    "Invalid human false-positive accounting")
        for regime in ("h2l", "free_llm", "llm2l"):
            check_pairs(arb["regimes"][regime], 3600)
        for domain in ("owt", "wp", "xsum"):
            check_pairs(arb["h2l_by_domain"][domain], 1200)
        for generator in ("gemma2_9b", "llama32_3b",
                          "mistral7b", "qwen25_7b"):
            check_pairs(arb["h2l_by_generator"][generator], 900)
        check_pairs(academic["ai_rewrites"], 143)
        require(sum(row["raw_pairs"] for row in
                    academic["by_generator"].values()) == 143,
                "Academic generator denominator mismatch")
        for row in academic["by_generator"].values():
            check_pairs(row, row["raw_pairs"])
    ensemble = result["scores"]["ensemble"]
    baseline = result["scores"]["richness"]
    a = ensemble["arb"]["regimes"]["h2l"]
    b = ensemble["academic"]["ai_rewrites"]
    advance = all((
        a["eligible_pairs"] / 3600 >= 0.95,
        a["auroc"] >= 0.70, a["tpr"] >= 0.10,
        ensemble["arb"]["human"]["fpr"] <= 0.02,
        a["auroc"] >= baseline["arb"]["regimes"]["h2l"]["auroc"],
        a["tpr"] >= baseline["arb"]["regimes"]["h2l"]["tpr"],
        b["eligible_pairs"] / 143 >= 0.95,
        b["auroc"] >= 0.70, b["tpr"] >= 0.10,
        ensemble["academic"]["test_human"]["fpr"] <= 0.02,
        ensemble["academic"]["human_only"]["fpr"] <= 0.02,
    ))
    require(result["advance_pass"] is advance and advance is False,
            "Frozen decision mismatch")
    posthoc_path = directory / "posthoc-result.json"
    require(sha(posthoc_path) == POSTHOC_SHA, "Post-result audit hash mismatch")
    posthoc = json.loads(posthoc_path.read_text())
    require(posthoc["format"] == "unswell-interpretable-ensemble-posthoc-v1" and
            posthoc["input_sha256"] == freeze["sha256"] and
            posthoc["main_result_sha256"] == RESULT_SHA and
            posthoc["script_sha256"] == sha(directory / "posthoc.py"),
            "Post-result audit identity mismatch")
    for group, expected_count in (("arb_test_human", 900),
                                  ("academic_test_human", 89),
                                  ("academic_human_only", 400)):
        row = posthoc["groups"][group]
        require(row["sources"] == expected_count,
                "Post-result source denominator mismatch")
        for index, feature in enumerate(result["feature_names"]):
            values = row["features"][feature]
            mean = values["mean"]
            contribution = values["mean_logit_contribution"]
            expected = (fit["weights"][index] *
                        (mean - fit["means"][index]) / fit["deviations"][index])
            require(math.isfinite(mean) and math.isclose(contribution, expected,
                                                         abs_tol=1e-12),
                    "Post-result contribution mismatch")


def check_pairs(row: dict, total: int) -> None:
    require(row["raw_pairs"] == total and
            0 <= row["eligible_pairs"] <= total and
            row["abstained_pairs"] == total - row["eligible_pairs"] and
            0 <= row["flagged"] <= row["eligible_pairs"] and
            (bounded(row["auroc"]) if row["eligible_pairs"] >= 20
             else row["auroc"] is None) and
            all(bounded(row[key]) for key in
                ("paired_win_fraction", "tpr")) and
            math.isclose(row["tpr"],
                         row["flagged"] / row["eligible_pairs"],
                         abs_tol=1e-12), "Invalid paired score accounting")


if __name__ == "__main__":
    verify()
    print("Interpretable ensemble evidence verified")
