"""Verify public accounting and identities; semantic grades remain private."""
import hashlib
import json
import math
from pathlib import Path

ROOT = Path(__file__).resolve().parent
SCOPE = {"pages": 36, "units": 4549, "judgments": 804, "events": 123}
COHORTS = {"whole": (26, 57), "context": (6, 46), "confirmation": (4, 20)}


def require(condition, message):
    if not condition:
        raise ValueError(message)


def count(value):
    require(type(value) is int and value >= 0, "Invalid count")
    return value


def check_cohorts(rows):
    require(set(rows) == set(COHORTS), "Changed cohort inventory")
    passed = []
    for name, (pages, events) in COHORTS.items():
        row = rows[name]
        require(count(row["pages"]) == pages and count(row["events"]) == events, "Changed original cohort scope")
        available = count(row["available_pages"])
        require(available <= pages, "Invented availability")
        delivered, accepted = count(row["delivered"]), count(row["accepted"])
        require(accepted <= delivered, "More accepted than delivered")
        require(math.isclose(row["full_event_recall"], count(row["full"]) / events), "Changed recall")
        ratio = accepted / delivered if delivered else 0
        require(math.isclose(row["accepted_fraction"], ratio), "Changed useful fraction")
        unresolved = count(row["unresolved_supplied_advice"])
        qualifies = available == pages and bool(delivered) and 100 * row["full"] >= 80 * events
        qualifies = qualifies and 100 * accepted >= 85 * delivered and unresolved == 0
        require(row["passed"] is bool(qualifies), "False cohort acceptance")
        passed.append(qualifies)
    return all(passed)


def check(result, execution, manifest, protocol, comparison, ablation):
    require(result["original_scope"] == protocol["original_scope"] == comparison["original_scope"] == SCOPE,
            "Changed original denominator")
    require(result["actual_input_aliases"] == protocol["actual_input_aliases"] == 3588, "Changed criticism inventory")
    require(protocol["numeric_targets"] == {"full_event_recall": 0.8, "accepted_fraction": 0.85, "unresolved_supplied_advice": 0}, "Changed numeric targets")
    passed = check_cohorts(result["cohorts"])
    require(result["development_screen_passed"] is bool(passed), "False overall acceptance")
    require(result["full_events"] == sum(row["full"] for row in result["cohorts"].values()), "Changed full-event total")
    require(sum(row["delivered"] for row in result["cohorts"].values()) <= 3588, "Invented delivered criticism")
    for row in result["cohorts"].values():
        require(sum(count(row[key]) for key in ("full", "partial", "missed")) == row["events"], "Incomplete event accounting")
        require(sum(count(row[key]) for key in ("accepted", "uncertain", "rejected")) == row["delivered"], "Incomplete grade accounting")
        require(count(row["safe_or_verification_only_advice"]) + row["unresolved_supplied_advice"] == count(row["nonempty_supplied_advice"])
                <= row["delivered"], "Incomplete actual-advice accounting")
    for name in ("product_qualified", "objective_achieved", "standalone_generation_qualified", "prospective_confirmation_measured", "current_main_integration_qualified"):
        require(result[name] is False, "Cached admission promoted to product qualification")
    require(result["same_assistant_development"] is True and result["historical_confirmation_already_exposed"] is True
            and result["cached_proposal_admission_only"] is True, "Changed review or exposure scope")
    pages = manifest["pages"]
    require(len(pages) == len(protocol["queue"]) == len(set(protocol["queue"])) == 36, "Changed page queue")
    require({row["page"] for row in pages} == set(protocol["queue"]) and len({row["path"] for row in pages}) == 36, "Changed source identity")
    require(sum(count(row["original_units"]) for row in pages) == 4549, "Changed unit inventory")
    require(sum(count(row["actual_input_aliases"]) for row in pages) == 3588, "Lost original aliases")
    require(count(execution["calls"]) == execution["new_full_calls"] + execution["reused_pilot_calls"] == 36
            and execution["new_full_calls"] == 32 and execution["reused_pilot_calls"] == 4 and execution["available_pages"] == 36,
            "Changed actual execution")
    require(execution["elapsed_seconds_including_pilot_and_preparation"] <= execution["total_budget_seconds"] == 7200,
            "Restarted or exceeded the total time budget")
    require(execution["model"] == protocol["model"] == "gpt-6-astra" and execution["effort"] == protocol["effort"] == "medium", "Changed model route")
    require(all(count(execution[key]) == 0 for key in ("paid_api_calls", "retries", "tools_used", "response_or_diagnosis_repairs")), "Unreported inference or repair")
    require(execution["private_benchmark_grades_sent"] is False and execution["private_reference_events_sent"] is False
            and execution["prospective_source_bodies_read"] is False, "Changed inference privacy")
    require(set(comparison["cohorts"]) == set(COHORTS) and comparison["changes_are_composite"] is True
            and comparison["isolated_causal_effect_established"] is False, "False causal comparison")
    for name, row in comparison["cohorts"].items():
        current = result["cohorts"][name]
        require(row["previous_full"] == row["retained_full"] + row["lost_full"]
                and row["current_full"] == row["retained_full"] + row["gained_full"] == current["full"], "Changed gain/loss accounting")
        require(row["current_delivered"] == current["delivered"] and row["current_accepted"] == current["accepted"], "Changed comparison outcome")
    require(set(ablation["components"]) == {"high", "medium", "original", "public_contrast", "source_only", "source_surface"}, "Omitted a predeclared component")
    require(ablation["requests_co_judged_all_six_streams"] is True and ablation["isolated_component_inference_measured"] is False
            and ablation["per_cohort_component_routing_used"] is False and ablation["new_model_calls"] == 0
            and ablation["product_qualified"] is False, "Component analysis promoted to standalone inference")
    for component in ablation["components"].values():
        check_cohorts(component["cohorts"])
        require(component["actual_admitted_aliases"] == sum(row["delivered"] for row in component["cohorts"].values()), "Changed component accounting")
    require(sum(component["actual_input_aliases"] for component in ablation["components"].values()) == 3588, "Lost a component's original inputs")
    return {"public_identities_and_counts_verified": True, "private_semantic_grades_independently_verified": False}


def verify(folder=ROOT):
    identities = json.loads((folder / "public-artifacts.json").read_text())
    for name, digest in identities.items():
        require(hashlib.sha256((folder / name).read_bytes()).hexdigest() == digest, "Changed public artifact: " + name)
    read = lambda name: json.loads((folder / name).read_text())
    return check(*(read(name) for name in ("evaluation.json", "execution.json", "manifest.json", "protocol.json", "comparison.json", "component-ablation.json")))


if __name__ == "__main__":
    print(json.dumps(verify()))
