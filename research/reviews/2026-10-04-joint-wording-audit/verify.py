"""Verify public identities and aggregate accounting, not private semantic grades."""
import hashlib
import json
import math
from pathlib import Path

ROOT = Path(__file__).resolve().parent


def require(condition, message):
    if not condition:
        raise ValueError(message)


def count(value):
    require(type(value) is int and value >= 0, "Invalid count")
    return value


def check_counts(result, execution, manifest, protocol):
    expected = {"whole": (26, 57), "context": (6, 46), "confirmation": (4, 20)}
    require(set(result["cohorts"]) == set(expected), "Changed cohort inventory")
    require([result[k] for k in ["original_pages", "original_units", "original_labels", "original_events"]]
            == [36, 4549, 804, 123], "Changed original scope")
    require(result["registered_prior_criticisms"] == 3844, "Changed frozen prior count")
    totals = {k: 0 for k in ["full", "delivered", "accepted", "uncertain", "rejected"]}
    available, passed = 0, True
    for name, (pages, events) in expected.items():
        row = result["cohorts"][name]
        require(count(row["pages"]) == pages and count(row["events"]) == events,
                "Changed cohort denominator")
        require(sum(count(row[k]) for k in ["full", "partial", "missed"]) == events,
                "Incomplete event dispositions")
        require(sum(count(row[k]) for k in ["accepted", "uncertain", "rejected"]) == count(row["delivered"]),
                "Incomplete delivered-finding dispositions")
        require(count(row["available_pages"]) <= pages, "Invented available page")
        require(math.isclose(row["full_event_recall"], row["full"] / events), "Changed recall")
        ratio = row["accepted"] / row["delivered"] if row["delivered"] else 0
        require(math.isclose(row["accepted_fraction"], ratio), "Changed accepted fraction")
        require(count(row["safe_or_verification_only"]) <= row["delivered"], "Invented safety disposition")
        available += row["available_pages"]
        passed = passed and row["available_pages"] == pages and row["full_event_recall"] >= .8 and ratio >= .85
        passed = passed and row["safe_or_verification_only"] == row["delivered"]
        for key in totals:
            totals[key] += row[key]
    require(result["development_screen_passed"] is bool(passed), "False development acceptance")
    require(result["product_qualified"] is False, "Unconfirmed product qualification")
    require(result["full_events"] == totals["full"], "Changed full-event total")
    for key in ["delivered", "accepted", "uncertain", "rejected"]:
        require(sum(count(row[key]) for row in result["operators"].values()) == totals[key],
                "Changed operator accounting")
    for row in result["operators"].values():
        require(row["delivered"] == row["accepted"] + row["uncertain"] + row["rejected"],
                "Incomplete operator dispositions")
    comparison = result["event_comparison"]
    require(comparison["previous_full"] == comparison["retained_full"] + comparison["lost_full"] and
            comparison["current_full"] == comparison["retained_full"] + comparison["gained_full"] == totals["full"],
            "Changed event gain/loss accounting")
    rows = manifest["pages"]
    require(len(rows) == 36 and {row["page"] for row in rows} == set(protocol["queue"]), "Lost source identity")
    require(len({row["path"] for row in rows}) == 36 and sum(count(row["units"]) for row in rows) == 4549,
            "Changed source/unit inventory")
    require(sum(count(row["builtin_diagnostics"]) for row in rows) == 143, "Changed ordinary input stream")
    require(execution["calls"] == result["calls"] == 36 and execution["semantic_unavailable_pages"] == 36 - available,
            "Changed call or unavailable-page count")
    require(execution["seconds"] <= execution["budget_limit_seconds"] + 15, "Unreported execution overrun")
    require(execution["original_strict_trace_statuses"] == result["original_strict_trace_execution"]["statuses"] == {"invalid": 36},
            "Promoted original strict execution")
    require(result["original_strict_trace_execution"]["available_pages"] == 0 and
            result["original_strict_trace_execution"]["passed"] is False and
            result["original_strict_trace_execution"]["replaced_or_rewritten"] is False,
            "Rewritten strict result")
    require(all(execution[k] == 0 for k in ["new_calls_for_compatibility", "retries", "diagnosis_repairs", "paid_api_calls"]),
            "Unreported generation or repair")
    require(execution["private_grades_sent"] is False and execution["prospective_sources_read"] is False and
            result["independent_human_annotation"] is False, "Changed privacy or reviewer scope")


def verify(folder=ROOT):
    identities = json.loads((folder / "public-artifacts.json").read_text())
    for name, expected in identities.items():
        require(hashlib.sha256((folder / name).read_bytes()).hexdigest() == expected, "Changed public artifact: " + name)
    read = lambda name: json.loads((folder / name).read_text())
    check_counts(read("evaluation.json"), read("execution.json"), read("manifest.json"), read("protocol.json"))
    return {"public_identities_and_counts_verified": True, "private_semantic_grades_independently_verified": False}


if __name__ == "__main__":
    print(json.dumps(verify()))
