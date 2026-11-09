#!/usr/bin/env python3
"""Verify the frozen, exposed structured-admission screen without private inputs."""

import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
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
    "whole": (37, 1183, 1358, 4),
    "context": BASELINE["context"],
    "confirmation": BASELINE["confirmation"],
}
GROUPS = {
    "bat", "containerd", "curl", "etcd", "fzf", "imgui", "marked",
    "prometheus", "ptah", "pydantic", "ripgrep", "tokio",
}


def require(condition, message):
    if not condition:
        raise ValueError(message)


def digest(name):
    return hashlib.sha256((HERE / name).read_bytes()).hexdigest()


def verify_freezes():
    for suffix in ("", "-v2"):
        freeze = json.loads((HERE / f"freeze{suffix}.json").read_text())
        require(freeze["status"] == "frozen_before_exploratory_scan",
                "Freeze status changed")
        require(freeze["scikit_learn"] == "1.7.2" and
                freeze["numpy"] == "2.5.4" and freeze["scipy"] == "1.18.1",
                "Dependency versions changed")
        require(freeze["protocol_sha256"] == digest(f"PROTOCOL{suffix}.md") and
                freeze["script_sha256"] == digest(f"scan{suffix}.py"),
                "Frozen file identity changed")
    return json.loads((HERE / "freeze-v2.json").read_text())


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


def counts(record):
    return tuple(record[key] for key in
                 ("full", "accepted", "delivered", "unresolved_advice"))


def verify(data, freeze):
    require(data["format"] == "unswell-349-structured-admission-exploratory-v2",
            "Result format changed")
    require(data["scope"] == "exposed_offline_method_screen_only",
            "Scope inflated")
    require(data["input_sha256"] == INPUTS, "Input identities changed")
    require(data["script_sha256"] == freeze["script_sha256"] and
            data["protocol_sha256"] == freeze["protocol_sha256"],
            "Result file identities changed")
    require(data["sklearn_version"] == freeze["scikit_learn"],
            "Model version changed")
    folds = data["folds"]
    require(set(folds) == GROUPS and
            all(set(fold) == {"train", "heldout"} and
                type(fold["train"]) is int and type(fold["heldout"]) is int and
                fold["heldout"] > 0 and fold["train"] + fold["heldout"] == 2101
                for fold in folds.values()) and
            sum(fold["heldout"] for fold in folds.values()) == 2101,
            "Repository folds changed")
    for name, expected in (("baseline", BASELINE), ("fixed_0_5", FIXED)):
        metrics = data[name]
        check_metrics(metrics)
        require({cohort: counts(record) for cohort, record in metrics.items()} ==
                expected, name + " changed")
    best = data["optimistic_best_at_85_percent"]
    require(set(best) == {"threshold", "whole"} and
            abs(best["threshold"] - 0.45661362684259155) < 1e-12,
            "Threshold changed")
    require(set(best["whole"]) == set(data["baseline"]["whole"]),
            "Best metric fields changed")
    whole = best["whole"]
    require(whole["events"] == 57 and counts(whole) == (45, 1321, 1554, 5) and
            abs(whole["fraction"] - whole["accepted"] / whole["delivered"]) < 1e-12,
            "Best threshold accounting changed")
    require(abs(data["heldout_auc"] - 0.7774695992468861) < 1e-12,
            "Held-out AUC changed")
    require(whole["accepted"] / whole["delivered"] >= 0.85 and
            whole["full"] < 46 and
            data["optimistic_numeric_feasible"] is False and
            data["optimistic_first_numeric_pass"] is None,
            "Failure boundary changed")
    require(data["qualified"] is False and
            "not new-page confirmation" in data["reason"] and
            "advice safety" in data["reason"],
            "False qualification claim")
    return {"whole_full": whole["full"], "qualified": False}


if __name__ == "__main__":
    print(verify(json.loads((HERE / "result-v2.json").read_text()),
                 verify_freezes()))
