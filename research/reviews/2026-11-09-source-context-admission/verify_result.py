#!/usr/bin/env python3
"""Verify the frozen source-context screen and its explicit failed gates."""

import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
CONTROL = HERE.parent / "2026-11-09-structured-admission"
INPUTS = {
    "evaluation": "820a25b21c5c3d13b9a3819f9560e2e4ab37ae519368131d294b6ea784fd7f51",
    "metadata": "c1df0e533b90d18458d3782b8464b0380297cf439b70a59fd3de52e8dcec6f06",
    "cores": "2eff16e368786905cf2d49be3eae7589d70597aa0f0627b329f43d02b8317927",
    "reviews": "d57beb6ee3213ab16c286dd6a0a0bd7ed5adf24f1c67cd40f71d50458ab85ba2",
    "originals": "4f0051badc4b16dfe088d88cae69bd0dda89af2e9f62b4bed317309c4db65082",
}
EVENTS = {"whole": 57, "context": 46, "confirmation": 20}
BASELINE = {
    "whole": (48, 1627, 2101, 5),
    "context": (39, 515, 548, 0),
    "confirmation": (17, 190, 198, 0),
}
FIXED = {
    "whole": (44, 1425, 1709, 5),
    "context": BASELINE["context"],
    "confirmation": BASELINE["confirmation"],
}
RESULT_SHA256 = "f8965a0a8739d3cc4900f9525a80b97fd2b2b51c0e6d78b907f1178de12eb37a"


def require(condition, message):
    if not condition:
        raise ValueError(message)


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def counts(record):
    return tuple(record[key] for key in
                 ("full", "accepted", "delivered", "unresolved_advice"))


def check_metrics(metrics):
    require(set(metrics) == set(EVENTS), "Cohorts changed")
    for cohort, record in metrics.items():
        require(set(record) == {"full", "events", "accepted", "delivered",
                                "fraction", "unresolved_advice"},
                "Metric fields changed")
        require(record["events"] == EVENTS[cohort], "Event denominator changed")
        require(all(type(record[key]) is int for key in
                    ("full", "events", "accepted", "delivered", "unresolved_advice")),
                "Noninteger count")
        require(0 <= record["full"] <= record["events"] and
                0 <= record["accepted"] <= record["delivered"] and
                0 <= record["unresolved_advice"] <= record["delivered"] and
                record["delivered"] > 0,
                "Invalid count")
        require(abs(record["fraction"] - record["accepted"] /
                    record["delivered"]) < 1e-12,
                "Fraction mismatch")


def verify(data, freeze, control):
    require(data["format"] == "unswell-349-source-context-admission-exploratory-v1",
            "Result format changed")
    require(data["scope"] == "exposed_offline_method_screen_only",
            "Scope inflated")
    require(data["input_sha256"] == INPUTS and control["input_sha256"] == INPUTS,
            "Input identities changed")
    for key in ("protocol_sha256", "script_sha256", "control_script_sha256",
                "control_result_sha256"):
        require(data[key] == freeze[key], "Frozen file identity changed")
    require((data["sklearn_version"], data["numpy_version"],
             data["scipy_version"]) == ("1.7.2", "2.5.4", "1.18.1"),
            "Dependency versions changed")
    folds = data["folds"]
    require(folds == control["folds"] and len(folds) == 12 and
            sum(fold["heldout"] for fold in folds.values()) == 2101 and
            all(fold["train"] + fold["heldout"] == 2101 for fold in folds.values()),
            "Repository folds changed")
    for name, expected in (("baseline", BASELINE), ("fixed_0_5", FIXED)):
        metrics = data[name]
        check_metrics(metrics)
        require({cohort: counts(record) for cohort, record in metrics.items()} ==
                expected, name + " changed")
    require(data["baseline"] == control["baseline"], "Control baseline changed")
    best = data["optimistic_best_at_85_percent"]
    require(set(best) == {"threshold", "whole"} and
            abs(best["threshold"] - 0.536277464708663) < 1e-12,
            "Threshold changed")
    whole = best["whole"]
    require(set(whole) == set(data["baseline"]["whole"]) and
            whole["events"] == 57 and counts(whole) == (43, 1323, 1556, 5) and
            abs(whole["fraction"] - whole["accepted"] / whole["delivered"]) < 1e-12,
            "Best threshold accounting changed")
    require(abs(data["heldout_auc"] - 0.7732022126613398) < 1e-12,
            "Held-out AUC changed")
    fixed = data["fixed_0_5"]["whole"]
    prior = control["fixed_0_5"]["whole"]
    require(fixed["full"] > prior["full"] and
            fixed["fraction"] < prior["fraction"] and
            fixed["unresolved_advice"] > prior["unresolved_advice"] and
            data["fixed_improves_both"] is False and
            whole["fraction"] >= 0.85 and whole["full"] < 46 and
            data["optimistic_first_numeric_pass"] is None and
            data["optimistic_numeric_feasible"] is False,
            "Failure boundary changed")
    require(data["qualified"] is False and
            "no page-disjoint natural transfer" in data["reason"],
            "False qualification claim")
    return {"whole_full_fixed": fixed["full"],
            "whole_full_at_85_percent": whole["full"], "qualified": False}


def verify_public_files():
    freeze = json.loads((HERE / "freeze.json").read_text())
    require(freeze["format"] == "unswell-349-source-context-admission-freeze-v1" and
            freeze["status"] == "frozen_before_exploratory_scan",
            "Freeze status changed")
    require(freeze["protocol_sha256"] == digest(HERE / "PROTOCOL.md") and
            freeze["script_sha256"] == digest(HERE / "scan.py") and
            freeze["control_script_sha256"] == digest(CONTROL / "scan-v2.py") and
            freeze["control_result_sha256"] == digest(CONTROL / "result-v2.json"),
            "Frozen source identity changed")
    require(digest(HERE / "result.json") == RESULT_SHA256,
            "Saved result identity changed")
    data = json.loads((HERE / "result.json").read_text())
    control = json.loads((CONTROL / "result-v2.json").read_text())
    return verify(data, freeze, control)


if __name__ == "__main__":
    print(verify_public_files())
