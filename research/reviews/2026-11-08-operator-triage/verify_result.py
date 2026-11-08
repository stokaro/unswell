#!/usr/bin/env python3
"""Check public accounting and failure claims for the #349 triage audit."""

import collections
import json
from pathlib import Path


EXPECTED_INPUTS = {
    "evaluation": "820a25b21c5c3d13b9a3819f9560e2e4ab37ae519368131d294b6ea784fd7f51",
    "metadata": "c1df0e533b90d18458d3782b8464b0380297cf439b70a59fd3de52e8dcec6f06",
    "cores": "2eff16e368786905cf2d49be3eae7589d70597aa0f0627b329f43d02b8317927",
    "reviews": "d57beb6ee3213ab16c286dd6a0a0bd7ed5adf24f1c67cd40f71d50458ab85ba2",
    "originals": "4f0051badc4b16dfe088d88cae69bd0dda89af2e9f62b4bed317309c4db65082",
}
EXPECTED_EVENTS = {"whole": 57, "context": 46, "confirmation": 20}
EXPECTED_BASELINE = {
    "whole": (48, 1627, 2101, 5),
    "context": (39, 515, 548, 0),
    "confirmation": (17, 190, 198, 0),
}


def require(condition, message):
    if not condition:
        raise ValueError(message)


def check_metrics(metrics):
    require(set(metrics) == set(EXPECTED_EVENTS), "Cohorts changed")
    for cohort, record in metrics.items():
        require(set(record) == {"full", "events", "accepted", "delivered", "fraction", "unresolved_advice"},
                f"Fields changed for {cohort}")
        require(record["events"] == EXPECTED_EVENTS[cohort], f"Event denominator changed for {cohort}")
        require(all(isinstance(record[key], int) and not isinstance(record[key], bool) for key in
                    ("full", "events", "accepted", "delivered", "unresolved_advice")),
                f"Noninteger count for {cohort}")
        require(0 <= record["full"] <= record["events"] and
                0 <= record["accepted"] <= record["delivered"] and
                0 <= record["unresolved_advice"] <= record["delivered"],
                f"Invalid counts for {cohort}")
        require(record["delivered"] > 0 and
                abs(record["fraction"] - record["accepted"] / record["delivered"]) < 1e-12,
                f"Fraction mismatch for {cohort}")


def verify(data):
    require(data["format"] == "unswell-349-operator-triage-audit-v1", "Format changed")
    require(data["scope"] == "exposed_development_only", "Scope inflated")
    require(data["input_sha256"] == EXPECTED_INPUTS, "Input identities changed")
    for key in ("baseline", "exact_span_category_operator_deduplication"):
        check_metrics(data[key])
    baseline = data["baseline"]
    require({c: (v["full"], v["accepted"], v["delivered"], v["unresolved_advice"])
             for c, v in baseline.items()} == EXPECTED_BASELINE, "Baseline changed")
    dedup = data["exact_span_category_operator_deduplication"]
    require(all(dedup[c]["delivered"] <= baseline[c]["delivered"] and
                dedup[c]["full"] <= baseline[c]["full"] for c in baseline),
            "Deduplication claim is inconsistent")

    retrospective = data["retrospective_four_group_filter"]
    require(retrospective["selection_used_all_exposed_grades"] is True,
            "Retrospective selection not disclosed")
    require(retrospective["exclusions"] == [
        ["wordiness", "action_support"],
        ["unsupported_evaluation", "evaluation"],
        ["vague_claims", "evaluation"],
        ["vague_claims", "indirect_relation"],
    ], "Retrospective exclusions changed")
    check_metrics(retrospective["metrics"])
    require(retrospective["metrics"]["whole"]["unresolved_advice"] == 5,
            "Unresolved advice was hidden")

    transfer = data["leave_one_repository_out"]
    require(transfer["folds"] == 13 and len(transfer["per_group"]) == 13,
            "Repository folds changed")
    check_metrics(transfer["heldout"])
    count = collections.Counter(tuple(pair) for policy in transfer["per_group"].values()
                                for pair in policy)
    frequency = {tuple((record["category"], record["operator"])): record["folds"]
                 for record in transfer["policy_frequency"]}
    require(dict(count) == frequency, "Policy frequency mismatch")
    whole = transfer["heldout"]["whole"]
    require(whole["full"] == 45 and whole["accepted"] == 1443 and
            whole["delivered"] == 1700 and whole["unresolved_advice"] == 5,
            "Held-out result changed")
    require(whole["full"] / whole["events"] < 0.80 and whole["fraction"] < 0.85,
            "Failed gates misreported")
    require(data["qualified"] is False and "cached" in data["reason"].lower(),
            "False qualification claim")
    return {"whole_full": whole["full"], "whole_accepted": whole["accepted"],
            "whole_delivered": whole["delivered"], "qualified": False}


if __name__ == "__main__":
    path = Path(__file__).with_name("result.json")
    print(verify(json.loads(path.read_text())))
