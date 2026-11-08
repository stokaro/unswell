"""Score the paired #349 pilot only after complete-stream assistant review."""

import json
from pathlib import Path
import sys

import packet
import run

ROOT = Path(__file__).resolve().parent
ARMS = ("control", "treatment")
GRADES = ("FULL", "PARTIAL", "MISSED")
DISPOSITIONS = ("accepted", "uncertain", "rejected")


def require(condition, reason):
    if not condition:
        raise ValueError(reason)


def score():
    protocol = run.verify()
    journal = packet.read(ROOT / "run" / "run.json")
    calls = [(slug, stage) for slug, _ in packet.PAGES
             for stage in ("inventory", *protocol["arm_order"][slug])]
    require(journal["status"] == "complete" and journal["complete"] is True and
            len(journal["calls"]) == 9 and
            journal["freeze_sha256"] == packet.sha((run.PACKET / "freeze.json").read_bytes()) and
            [(row["page"], row["stage"]) for row in journal["calls"]] == calls and
            all(row["status"] == "valid" for row in journal["calls"]),
            "incomplete or reordered pilot cannot be scored")
    outputs = {}
    for row in journal["calls"]:
        slug, stage = row["page"], row["stage"]
        target = ROOT / "run" / slug / stage
        if stage == "inventory":
            request = (run.PACKET / "requests" / f"{slug}-inventory.txt").read_bytes()
        else:
            request = packet.review_request(slug, outputs[(slug, "inventory")][0], stage)
        raw = (target / "response.json").read_bytes()
        require(row["request_sha256"] == packet.sha(request) and
                row["response_sha256"] == packet.sha(raw) and
                (target / "request.txt").read_bytes() == request,
                "run request or response changed")
        response = json.loads(raw)
        run.trace(target, response)
        if stage == "inventory":
            packet.validate_inventory(slug, response)
            outputs[(slug, stage)] = (raw, response)
        else:
            inventory_raw = outputs[(slug, "inventory")][0]
            if stage == "control":
                bound = packet.validate_diagnosis(slug, response, inventory_raw)
            else:
                bound = packet.validate_treatment(slug, response, inventory_raw)
            require((target / "bound.json").read_bytes() == packet.encoded(bound),
                    "source-bound output changed")
            outputs[(slug, stage)] = (raw, bound)
    labels = packet.read(ROOT / "labels.json")
    events = {row["id"]: row for row in labels["events"]}
    controls = {row["id"]: row for row in labels["protected_counterexamples"]}
    require(len(events) == 4 and len(controls) == 5, "frozen denominators changed")
    review = packet.read(ROOT / "review.json")
    require(review["format"] == "unswell-349-disposition-complete-stream-review-v1" and
            "Codex assistant" in review["reviewer"] and
            review["full_stream_reviewed"] is True and
            review["treatment_dispositions_reviewed"] == 128 and
            review["edits_supplied"] == 0, "review identity or completeness changed")
    delivered = {}
    for slug, _ in packet.PAGES:
        for arm in ARMS:
            result = outputs[(slug, arm)][1]
            for index, finding in enumerate(result["findings"]):
                delivered[(arm, slug, index)] = finding
    findings = review["finding_reviews"]
    keys = [(row["arm"], row["page"], row["index"]) for row in findings]
    require(len(keys) == len(delivered) and len(set(keys)) == len(keys) and
            set(keys) == set(delivered), "complete delivered stream was not reviewed")
    require(all(row["disposition"] in DISPOSITIONS and
                isinstance(row["rationale"], str) and row["rationale"].strip()
                for row in findings), "finding review lacks disposition or rationale")
    event_reviews = review["event_reviews"]
    event_keys = [(row["arm"], row["event_id"]) for row in event_reviews]
    require(len(event_keys) == len(ARMS) * len(events) and len(set(event_keys)) == len(event_keys) and
            set(event_keys) == {(arm, key) for arm in ARMS for key in events},
            "frozen event review is incomplete")
    require(all(row["grade"] in GRADES and isinstance(row["rationale"], str) and
                row["rationale"].strip() and
                isinstance(row["finding_indices"], list) and
                all(type(index) is int for index in row["finding_indices"])
                for row in event_reviews),
            "event review lacks grade or rationale")
    slug_for_page = {page_id: slug for slug, page_id in packet.PAGES}
    for row in event_reviews:
        slug = slug_for_page[events[row["event_id"]]["page"]]
        indices = row["finding_indices"]
        require(len(set(indices)) == len(indices) and
                all((row["arm"], slug, index) in delivered for index in indices) and
                bool(indices) == (row["grade"] != "MISSED"),
                "event grade has no traceable delivered finding")
    protected = review["protected_reviews"]
    protected_keys = [(row["arm"], row["counterexample_id"]) for row in protected]
    require(len(protected_keys) == len(ARMS) * len(controls) and
            len(set(protected_keys)) == len(protected_keys) and
            set(protected_keys) == {(arm, key) for arm in ARMS for key in controls},
            "protected counterexample review is incomplete")
    require(all(type(row["flagged"]) is bool and isinstance(row["rationale"], str) and
                row["rationale"].strip() for row in protected),
            "protected review lacks flag or rationale")
    for row in protected:
        control = controls[row["counterexample_id"]]
        slug = slug_for_page[control["page"]]
        found = any(arm == row["arm"] and page == slug and
                    finding["unit_ordinal"] in control["overlapping_unit_ordinals"]
                    for (arm, page, _), finding in delivered.items())
        require(row["flagged"] == found, "protected flag disagrees with delivered findings")
    summary = {}
    for arm in ARMS:
        grades = {grade: sum(row["arm"] == arm and row["grade"] == grade
                             for row in event_reviews) for grade in GRADES}
        rows = [row for row in findings if row["arm"] == arm]
        accepted = sum(row["disposition"] == "accepted" for row in rows)
        uncertain = sum(row["disposition"] == "uncertain" for row in rows)
        rejected = sum(row["disposition"] == "rejected" for row in rows)
        summary[arm] = {"events": len(events), "full": grades["FULL"],
                        "partial": grades["PARTIAL"], "missed": grades["MISSED"],
                        "delivered": len(rows), "accepted": accepted,
                        "uncertain": uncertain, "rejected": rejected,
                        "accepted_fraction": accepted / len(rows) if rows else None,
                        "protected_flagged": sum(row["arm"] == arm and row["flagged"]
                                                 for row in protected)}
    treatment_decisions = [row for slug, _ in packet.PAGES for row in
                           outputs[(slug, "treatment")][1]["dispositions"]]
    require(len(treatment_decisions) == 128, "treatment disposition denominator changed")
    summary["treatment"]["unit_dispositions"] = len(treatment_decisions)
    summary["treatment"]["uncertain_units"] = sum(row["decision"] == "uncertain"
                                                  for row in treatment_decisions)
    treatment = summary["treatment"]
    control = summary["control"]
    advance = (treatment["full"] >= 3 and treatment["full"] - control["full"] >= 2 and
               treatment["accepted_fraction"] is not None and
               treatment["accepted_fraction"] >= 0.85 and
               treatment["protected_flagged"] == 0)
    return {"format": "unswell-349-source-bound-disposition-pilot-result-v1",
            "pages": 3, "eligible_units": 128, "invalid_pages": 0, "arms": summary,
            "supplied_edits": 0, "suggestion_safety": "not applicable",
            "advance": advance, "qualification": False,
            "qualification_note": "Directional three-page pilot only; #349 development and separate confirmation remain."}


if __name__ == "__main__":
    if sys.argv[1:] != ["score"]:
        raise SystemExit("usage: score.py score")
    print(json.dumps(score(), indent=2, sort_keys=True))
