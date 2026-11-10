#!/usr/bin/env python3
"""Validate complete assistant review of both frozen #349 pilot arms."""

import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path

from verify_reference import verify as verify_reference

ROOT = Path(__file__).resolve().parent
ARMS = ("freeform", "atomic")


def read(path):
    return json.loads(Path(path).read_text(encoding="utf-8"))


def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def overlap(left, right):
    return left["start"] < right["end"] and right["start"] < left["end"]


def collect(run_dir):
    reference = verify_reference(ROOT)
    journal = read(run_dir / "run.json")
    if journal["format"] != "unswell-349-atomic-paired-run-v1":
        raise ValueError("run format changed")
    if not journal["complete"] or journal["status"] != "complete" or len(journal["calls"]) != 18:
        raise ValueError("partial or invalid run cannot pass the paired pilot")
    found = {}
    for row in journal["calls"]:
        if row["status"] != "valid":
            raise ValueError("invalid model call")
        name = row["request"]
        bound_path = run_dir / name / "bound.json"
        if sha(bound_path) != row["bound_sha256"]:
            raise ValueError("bound model response changed")
        bound = read(bound_path)
        if bound["arm"] not in ARMS or name in found:
            raise ValueError("arm identity or duplicate changed")
        found[name] = bound
    if set(found) != {f"page-{p:03d}-{a}" for p in range(1, 10) for a in ARMS}:
        raise ValueError("paired coverage changed")
    return reference, journal, found


def template(run_dir):
    reference, journal, found = collect(run_dir)
    rows = []
    for arm in ARMS:
        findings = {}
        for page in range(1, 10):
            for diagnosis in found[f"page-{page:03d}-{arm}"]["diagnoses"]:
                finding_id = diagnosis["id"]
                if finding_id in findings:
                    raise ValueError("duplicate bound finding ID")
                findings[finding_id] = {"page": page, "span": diagnosis["byte_span"],
                                        "problem": diagnosis["problem"],
                                        "preserve": diagnosis["preserve"]}
        events = []
        for event in reference["events"]:
            candidates = [finding_id for finding_id, item in findings.items()
                          if item["page"] == event["page"] and
                          overlap(item["span"], event["byte_span"])]
            events.append({"id": event["id"], "candidate_finding_ids": candidates,
                           "status": None, "finding_ids": [], "basis": ""})
        controls = []
        for control in reference["controls"]:
            candidates = [finding_id for finding_id, item in findings.items()
                          if item["page"] == control["page"] and
                          overlap(item["span"], control["byte_span"])]
            controls.append({"id": control["id"], "disposition": control["disposition"],
                             "candidate_finding_ids": candidates,
                             "status": None if candidates else "not_hit", "basis": ""})
        delivery = [{"id": finding_id, "status": None, "basis": ""}
                    for finding_id in sorted(findings)]
        rows.append({"arm": arm, "events": events, "controls": controls,
                     "delivery": delivery})
    return {"format": "unswell-349-atomic-paired-review-v1",
            "reference_sha256": sha(ROOT / "reference.json"),
            "run_sha256": sha(run_dir / "run.json"),
            "reviewer": "root Codex assistant",
            "arms": rows}


def grade(run_dir, reviewed):
    expected = template(run_dir)
    if set(reviewed) != set(expected) or any(
            reviewed[key] != expected[key] for key in
            ("format", "reference_sha256", "run_sha256", "reviewer")):
        raise ValueError("review identity changed")
    if type(reviewed["arms"]) is not list or len(reviewed["arms"]) != 2:
        raise ValueError("review arm count changed")
    result = {}
    for expected_arm, arm in zip(expected["arms"], reviewed["arms"]):
        if arm["arm"] != expected_arm["arm"] or set(arm) != set(expected_arm):
            raise ValueError("review arm identity changed")
        summary = {"FULL": 0, "PARTIAL": 0, "MISSED": 0,
                   "accepted": 0, "uncertain": 0, "rejected": 0,
                   "control_hits": 0, "retain_control_hits": 0, "control_uncertain": 0}
        for kind in ("events", "controls", "delivery"):
            if len(arm[kind]) != len(expected_arm[kind]):
                raise ValueError("review denominator changed")
        for actual, frozen in zip(arm["events"], expected_arm["events"]):
            if (set(actual) != set(frozen) or actual["id"] != frozen["id"] or
                    actual["candidate_finding_ids"] != frozen["candidate_finding_ids"] or
                    actual["status"] not in ("FULL", "PARTIAL", "MISSED") or
                    not isinstance(actual["basis"], str) or not actual["basis"].strip()):
                raise ValueError("event review changed or incomplete")
            ids = actual["finding_ids"]
            if type(ids) is not list or len(set(ids)) != len(ids) or any(
                    finding_id not in frozen["candidate_finding_ids"] for finding_id in ids):
                raise ValueError("event finding binding changed")
            if (actual["status"] == "MISSED") != (len(ids) == 0):
                raise ValueError("event status and finding IDs disagree")
            summary[actual["status"]] += 1
        for actual, frozen in zip(arm["controls"], expected_arm["controls"]):
            if (set(actual) != set(frozen) or actual["id"] != frozen["id"] or
                    actual["candidate_finding_ids"] != frozen["candidate_finding_ids"] or
                    actual["disposition"] != frozen["disposition"] or
                    actual["status"] not in ("not_hit", "hit", "uncertain") or
                    not isinstance(actual["basis"], str)):
                raise ValueError("control review changed or incomplete")
            if not frozen["candidate_finding_ids"] and actual["status"] != "not_hit":
                raise ValueError("control cannot be hit without source overlap")
            if actual["status"] != "not_hit" and not actual["basis"].strip():
                raise ValueError("control basis missing")
            if actual["status"] == "hit":
                summary["control_hits"] += 1
                if actual["disposition"] == "retain":
                    summary["retain_control_hits"] += 1
            elif actual["status"] == "uncertain":
                summary["control_uncertain"] += 1
        for actual, frozen in zip(arm["delivery"], expected_arm["delivery"]):
            if (set(actual) != set(frozen) or actual["id"] != frozen["id"] or
                    actual["status"] not in ("accepted", "uncertain", "rejected") or
                    not isinstance(actual["basis"], str) or not actual["basis"].strip()):
                raise ValueError("delivery review changed or incomplete")
            summary[actual["status"]] += 1
        accepted_ids = {item["id"] for item in arm["delivery"]
                        if item["status"] == "accepted"}
        for event in arm["events"]:
            if event["status"] == "FULL" and not any(
                    finding_id in accepted_ids for finding_id in event["finding_ids"]):
                raise ValueError("FULL event lacks an accepted finding")
        summary["event_count"] = len(arm["events"])
        summary["control_count"] = len(arm["controls"])
        summary["delivered"] = len(arm["delivery"])
        summary["accepted_fraction"] = (summary["accepted"] / summary["delivered"]
                                        if summary["delivered"] else 0)
        result[arm["arm"]] = summary
    atomic = result["atomic"]
    freeform = result["freeform"]
    return {"format": "unswell-349-atomic-paired-result-v1",
            "reference_sha256": expected["reference_sha256"],
            "run_sha256": expected["run_sha256"],
            "review_sha256": None,
            "arms": result,
            "advance": (atomic["FULL"] >= freeform["FULL"] + 2 and
                        atomic["accepted_fraction"] >= 0.85 and
                        atomic["retain_control_hits"] == 0 and atomic["delivered"] > 0),
            "qualified_349": False,
            "suggestion_safety": "not_applicable_to_diagnosis_only_arms"}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("operation", choices=("template", "grade"))
    parser.add_argument("--run-dir", type=Path, required=True)
    parser.add_argument("--review", type=Path)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if args.operation == "template":
        value = template(args.run_dir)
    else:
        if args.review is None:
            raise SystemExit("--review is required for grade")
        value = grade(args.run_dir, read(args.review))
        value["review_sha256"] = sha(args.review)
    args.output.write_text(json.dumps(value, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps({"output": str(args.output), "format": value["format"]}))


if __name__ == "__main__":
    main()
