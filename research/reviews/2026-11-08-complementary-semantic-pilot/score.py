"""Score the complete prior and complementary finding streams."""

import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent
CONTROL = ROOT.parent / "2026-11-08-source-function-pilot"
PAGES = ("caddy", "restic", "gitea", "otelcol")


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def score():
    labels = read(ROOT / "labels.json")
    control = read(CONTROL / "review.json")
    review = read(ROOT / "review.json")
    events = {event["id"] for event in labels["events"]}
    if len(events) != 8 or (ROOT / "labels.json").read_bytes() != (CONTROL / "labels.json").read_bytes():
        raise ValueError("frozen event denominator changed")
    if len(control["findings"]) != 17:
        raise ValueError("control denominator changed")
    delivered = {(page, index) for page in PAGES for index, _ in enumerate(
        read(ROOT / "run" / page / "bound.json")["findings"])}
    reviewed = {(row["page"], row["index"]) for row in review["findings"]}
    if len(reviewed) != len(review["findings"]) or reviewed != delivered:
        raise ValueError("new finding review does not cover the complete stream")
    if any(row["disposition"] not in ("accepted", "uncertain", "rejected")
           for row in review["findings"]):
        raise ValueError("invalid finding disposition")
    grades = {row["event_id"]: row["grade"] for row in review["events"]}
    if len(grades) != len(review["events"]) or set(grades) != events or any(
        grade not in ("FULL", "PARTIAL", "MISSED") for grade in grades.values()
    ):
        raise ValueError("event review does not cover frozen labels")
    control_accepted = sum(row["disposition"] == "accepted" for row in control["findings"])
    new_accepted = sum(row["disposition"] == "accepted" for row in review["findings"])
    combined_delivered = len(control["findings"]) + len(review["findings"])
    combined_accepted = control_accepted + new_accepted
    full = sum(grade == "FULL" for grade in grades.values())
    complete = all(read(ROOT / "run" / page / "bound.json")["page_id"] == page
                   for page in PAGES) and read(ROOT / "run" / "run.json")["complete"]
    return {
        "events": len(events), "full": full,
        "partial": sum(grade == "PARTIAL" for grade in grades.values()),
        "missed": sum(grade == "MISSED" for grade in grades.values()),
        "control_delivered": len(control["findings"]),
        "control_accepted": control_accepted,
        "new_delivered": len(review["findings"]),
        "new_accepted": new_accepted,
        "combined_delivered": combined_delivered,
        "combined_accepted": combined_accepted,
        "combined_accepted_fraction": combined_accepted / combined_delivered,
        "valid_pages": 4 if complete else 0,
        "protected_counterexamples_flagged": review["protected_counterexamples_flagged"],
        "supplied_edits": 0, "suggestion_safety": "not applicable",
        "advance": complete and full >= 7 and
                   combined_accepted / combined_delivered >= 0.85 and
                   not review["protected_counterexamples_flagged"],
        "qualification": False,
    }


if __name__ == "__main__":
    print(json.dumps(score(), indent=2, sort_keys=True))
