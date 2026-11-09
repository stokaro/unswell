"""Score the frozen joint pilot without changing its source-only labels."""

import json
from pathlib import Path
import sys

import run


ROOT = Path(__file__).resolve().parent
DISPOSITIONS = {row["finding_id"]: row["decision"] for row in
                run.read(ROOT / "baseline-dispositions.json")["rows"]}
LABELS = [row for row in run.read(run.PREVIOUS / "transfer-v2-labels.json")["judgments"]
          if row["page"] in run.PAGES]
POSITIVES = {row["id"]: row for row in LABELS if row["decision"] == "positive"}
CONTROLS = [row for row in LABELS if row["decision"] == "retain"]


def overlap(first, second):
    return first["start"] < second["end"] and second["start"] < first["end"]


def compute(bound_by_page, review):
    if set(bound_by_page) != {run.page_name(number) for number in run.PAGES}:
        raise ValueError("incomplete bound page set")
    if not isinstance(review, dict) or set(review) != {
            "format", "reviewer", "bound_sha256", "new_dispositions", "event_grades"}:
        raise ValueError("review fields changed")
    if review["format"] != "unswell-349-joint-pilot-source-review-v1":
        raise ValueError("review format changed")
    if "Codex assistant" not in review["reviewer"] or \
            "not a human annotator" not in review["reviewer"]:
        raise ValueError("reviewer identity missing")

    bound_digests = {name: run.digest(run.encoded(bound)) for name, bound in bound_by_page.items()}
    if review["bound_sha256"] != bound_digests:
        raise ValueError("bound response hash changed")

    delivered = {}
    withheld = []
    baseline_counts = {"keep": 0, "withhold": 0, "uncertain": 0}
    lost_accepted = []
    for number in run.PAGES:
        name = run.page_name(number)
        bound = bound_by_page[name]
        source = run.source_path(number).read_bytes()
        if bound["page_id"] != name or bound["source_sha256"] != run.digest(source):
            raise ValueError("bound source identity changed")
        baseline = {row["id"]: row for row in run.baseline_page(number)["findings"]}
        if {row["id"] for row in bound["reviews"]} != set(baseline):
            raise ValueError("incomplete baseline review")
        for decision in bound["reviews"]:
            finding_id = decision["id"]
            action = decision["decision"]
            if action not in baseline_counts:
                raise ValueError("invalid baseline action")
            baseline_counts[action] += 1
            if action == "keep":
                delivered[finding_id] = {
                    "page": number, "span": baseline[finding_id]["span"],
                    "decision": DISPOSITIONS[finding_id], "origin": "baseline",
                }
            else:
                withheld.append(finding_id)
                if DISPOSITIONS[finding_id] == "accepted":
                    lost_accepted.append(finding_id)
        for finding in bound["new_findings"]:
            finding_id = finding["id"]
            if finding_id in delivered:
                raise ValueError("duplicate delivered ID")
            delivered[finding_id] = {
                "page": number,
                "span": {"start": finding["start_byte"], "end": finding["end_byte"]},
                "decision": None, "origin": "new",
            }

    rows = review["new_dispositions"]
    if not isinstance(rows, list):
        raise ValueError("new dispositions must be a list")
    expected_new = {key for key, value in delivered.items() if value["origin"] == "new"}
    if len(rows) != len(expected_new) or {row.get("id") for row in rows} != expected_new:
        raise ValueError("new disposition coverage changed")
    allowed = {"accepted", "optional", "rejected", "uncertain"}
    for row in rows:
        if set(row) != {"id", "decision", "reason"} or \
                row["decision"] not in allowed or \
                not isinstance(row["reason"], str) or not row["reason"].strip():
            raise ValueError("invalid new disposition")
        delivered[row["id"]]["decision"] = row["decision"]

    grades = review["event_grades"]
    if not isinstance(grades, list) or len(grades) != len(POSITIVES):
        raise ValueError("event grade coverage changed")
    if {row.get("id") for row in grades} != set(POSITIVES):
        raise ValueError("missing or duplicate event grade")
    full = partial = missed = 0
    for row in grades:
        if set(row) != {"id", "grade", "witnesses", "reason"} or \
                row["grade"] not in {"FULL", "PARTIAL", "MISSED"} or \
                not isinstance(row["reason"], str) or not row["reason"].strip():
            raise ValueError("invalid event grade")
        witnesses = row["witnesses"]
        if not isinstance(witnesses, list) or len(witnesses) != len(set(witnesses)):
            raise ValueError("invalid event witnesses")
        if (row["grade"] == "MISSED") != (len(witnesses) == 0):
            raise ValueError("event witness and grade disagree")
        event = POSITIVES[row["id"]]
        event_span = {"start": event["start_byte"], "end": event["end_byte"]}
        for witness in witnesses:
            if witness not in delivered or delivered[witness]["page"] != event["page"] or \
                    not overlap(delivered[witness]["span"], event_span):
                raise ValueError("event witness is not delivered at the source")
        if row["grade"] == "FULL":
            full += 1
        elif row["grade"] == "PARTIAL":
            partial += 1
        else:
            missed += 1

    control_hits = []
    for control in CONTROLS:
        control_span = {"start": control["start_byte"], "end": control["end_byte"]}
        for finding_id, item in delivered.items():
            if item["page"] == control["page"] and overlap(item["span"], control_span):
                control_hits.append({"control": control["id"], "finding": finding_id})

    accepted = sum(item["decision"] == "accepted" for item in delivered.values())
    total = len(delivered)
    return {
        "format": "unswell-349-joint-pilot-score-v1",
        "pages": len(bound_by_page),
        "baseline_decisions": baseline_counts,
        "withheld_accepted_baseline": sorted(lost_accepted),
        "withheld_total": len(withheld),
        "new_findings": len(expected_new),
        "full": full, "partial": partial, "missed": missed,
        "positive_events": len(POSITIVES),
        "accepted_delivered": accepted, "delivered": total,
        "accepted_fraction": accepted / total if total else None,
        "retain_control_overlaps": control_hits,
        "suggestion_safety": "not_measured_no_new_edits",
        "directional_pass": full >= 7 and total > 0 and accepted / total >= 0.85
        and not control_hits and not lost_accepted,
        "qualified": False,
    }


def score_run():
    run.verify()
    journal = run.read(run.RUN / "run.json")
    if not journal["complete"] or journal["status"] != "complete" or \
            len(journal["calls"]) != len(run.PAGES) or \
            journal["freeze_sha256"] != run.file_digest(run.PACKET / "freeze.json") or \
            journal["model"] != run.MODEL or journal["effort"] != run.EFFORT or \
            [row["page"] for row in journal["calls"]] != list(run.PAGES):
        raise ValueError("model run incomplete")
    bound = {}
    for row in journal["calls"]:
        name = run.page_name(row["page"])
        target = run.RUN / name
        path = target / "bound.json"
        response_path = target / "response.json"
        request_path = run.PACKET / "requests" / f"{name}.txt"
        if row["status"] != "complete" or \
                row["request_sha256"] != run.file_digest(request_path) or \
                row["response_sha256"] != run.file_digest(response_path) or \
                row["bound_sha256"] != run.file_digest(path):
            raise ValueError("bound response changed")
        response = run.read(response_path)
        run.verify_trace(target, response)
        bound[name] = run.read(path)
        if bound[name] != run.bind(row["page"], response):
            raise ValueError("bound response differs from original source")
    result = compute(bound, run.read(ROOT / "review.json"))
    (ROOT / "result.json").write_bytes(run.encoded(result))
    return result


if __name__ == "__main__":
    if sys.argv[1:] != ["score"]:
        raise SystemExit("usage: score.py score")
    print(json.dumps(score_run(), sort_keys=True))
