#!/usr/bin/env python3
"""Verify the public aggregate of the frozen semantic admission screen."""

from __future__ import annotations

import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
CONTROL = HERE.parent / "2026-11-09-structured-admission"
RESULT_SHA256 = "16d4e93349315b63580b3d903d8cb1c2f4c7383962b3dc330591c6f95c5f4194"
ORACLE_SCRIPT_SHA256 = "dfc3e237bb3855ba519623ebb2ee5b2d58df6ae7a080f4777547ce10ab802a5e"
ORACLE_RESULT_SHA256 = "21e9e1abdb5934675e374d9a23799e05dff5d7dbbb75263f7c94511684e18df7"
EVENTS = {"whole": 57, "context": 46, "confirmation": 20}
BASELINE = {
    "whole": (48, 1627, 2101, 5),
    "context": (39, 515, 548, 0),
    "confirmation": (17, 190, 198, 0),
}
FIXED = {
    "whole": (39, 1164, 1386, 3),
    "context": BASELINE["context"],
    "confirmation": BASELINE["confirmation"],
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def counts(record: dict) -> tuple[int, int, int, int]:
    return tuple(record[key] for key in
                 ("full", "accepted", "delivered", "unresolved_advice"))


def check_metrics(metrics: dict) -> None:
    require(set(metrics) == set(EVENTS), "cohorts changed")
    for cohort, record in metrics.items():
        require(set(record) == {"full", "events", "accepted", "delivered",
                                "fraction", "unresolved_advice"},
                "metric fields changed")
        require(record["events"] == EVENTS[cohort], "event denominator changed")
        require(all(type(record[key]) is int for key in
                    ("full", "events", "accepted", "delivered", "unresolved_advice")),
                "noninteger count")
        require(0 <= record["full"] <= record["events"] and
                0 <= record["accepted"] <= record["delivered"] and
                0 <= record["unresolved_advice"] <= record["delivered"] and
                record["delivered"] > 0, "invalid count")
        require(abs(record["fraction"] - record["accepted"] /
                    record["delivered"]) < 1e-12, "fraction mismatch")


def verify(data: dict, freeze: dict, control: dict) -> dict:
    require(data["format"] == "unswell-349-semantic-admission-exploratory-v1",
            "result format changed")
    require(data["scope"] == "exposed_offline_method_screen_only",
            "scope inflated")
    require(data["input_sha256"] == control["input_sha256"],
            "input identities changed")
    for key in ("protocol_sha256", "script_sha256", "control_script_sha256",
                "control_result_sha256", "model_config_sha256",
                "model_weights_sha256", "tokenizer_sha256"):
        require(data[key] == freeze[key], "frozen identity changed: " + key)
    require((data["sklearn_version"], data["numpy_version"],
             data["torch_version"]) == ("1.8.0", "2.4.4", "2.10.0"),
            "dependency versions changed")
    folds = data["folds"]
    require(folds == control["folds"] and len(folds) == 12 and
            sum(fold["heldout"] for fold in folds.values()) == 2101 and
            all(fold["train"] + fold["heldout"] == 2101 for fold in folds.values()),
            "repository folds changed")
    for name, expected in (("baseline", BASELINE), ("fixed_0_5", FIXED)):
        metrics = data[name]
        check_metrics(metrics)
        require({cohort: counts(record) for cohort, record in metrics.items()} ==
                expected, name + " changed")
    require(data["baseline"] == control["baseline"], "control baseline changed")
    best = data["optimistic_best_at_85_percent"]
    require(set(best) == {"threshold", "whole"} and
            abs(best["threshold"] - 0.5740501591768294) < 1e-12,
            "optimistic threshold changed")
    check_metrics({"whole": best["whole"],
                   "context": data["fixed_0_5"]["context"],
                   "confirmation": data["fixed_0_5"]["confirmation"]})
    require(counts(best["whole"]) == (36, 1078, 1268, 3) and
            best["whole"]["fraction"] >= 0.85 and
            best["whole"]["full"] < 46, "optimistic boundary changed")
    require(abs(data["heldout_auc"] - 0.7092769950129539) < 1e-12,
            "held-out AUC changed")
    require(data["fixed_0_5"]["whole"]["fraction"] < 0.85 and
            data["fixed_0_5"]["whole"]["full"] < 46 and
            data["optimistic_first_numeric_pass"] is None and
            data["optimistic_numeric_feasible"] is False and
            data["qualified"] is False and
            "new-page transfer" in data["reason"],
            "false qualification or gate claim")
    return {"whole_full_fixed": 39, "whole_full_at_85_percent": 36,
            "qualified": False}


def verify_oracle(data: dict, control: dict) -> None:
    require(data["format"] == "unswell-349-accepted-only-oracle-v1" and
            data["scope"] == "label_only_upper_bound_not_a_selector" and
            data["qualified"] is False, "oracle scope changed")
    require(data["input_sha256"] == control["input_sha256"] and
            data["control_script_sha256"] == digest(CONTROL / "scan-v2.py") and
            data["script_sha256"] == ORACLE_SCRIPT_SHA256,
            "oracle input or script changed")
    check_metrics(data["baseline"])
    check_metrics(data["accepted_only"])
    require(data["baseline"] == control["baseline"], "oracle baseline changed")
    require({cohort: counts(record) for cohort, record in data["accepted_only"].items()} ==
            {"whole": (48, 1627, 1627, 3),
             "context": (39, 515, 515, 0),
             "confirmation": (17, 190, 190, 0)},
            "oracle upper bound changed")
    require(all(record["fraction"] == 1.0 for record in
                data["accepted_only"].values()), "oracle acceptance changed")


def verify_public_files() -> dict:
    freeze = json.loads((HERE / "freeze.json").read_text())
    require(freeze["format"] == "unswell-349-semantic-admission-freeze-v1" and
            freeze["status"] == "frozen_before_inference", "freeze status changed")
    require(freeze["protocol_sha256"] == digest(HERE / "PROTOCOL.md") and
            freeze["script_sha256"] == digest(HERE / "scan.py") and
            freeze["control_script_sha256"] == digest(CONTROL / "scan-v2.py") and
            freeze["control_result_sha256"] == digest(CONTROL / "result-v2.json"),
            "frozen source identity changed")
    require(digest(HERE / "result.json") == RESULT_SHA256,
            "saved result identity changed")
    require(digest(HERE / "oracle.py") == ORACLE_SCRIPT_SHA256 and
            digest(HERE / "oracle.json") == ORACLE_RESULT_SHA256,
            "oracle file identity changed")
    control = json.loads((CONTROL / "result-v2.json").read_text())
    verify_oracle(json.loads((HERE / "oracle.json").read_text()), control)
    return verify(json.loads((HERE / "result.json").read_text()), freeze, control)


if __name__ == "__main__":
    print(verify_public_files())
