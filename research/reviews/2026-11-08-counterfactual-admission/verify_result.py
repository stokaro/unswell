"""Independently verify the frozen CLI baseline and semantic claim review."""

import hashlib
import json
import tarfile
from pathlib import Path

from verify import DIRECTORY, verify_freeze

EXPECTED_BASELINE_COMMIT = "fe727da86cad38b60015fbf1f80dfbbad9884b01"


def sha256(data):
    return hashlib.sha256(data).hexdigest()


def compact_finding(finding):
    return {
        "id": finding["id"],
        "rule": finding["rule_id"],
        "message": finding["message"],
        "span": finding["primary"]["span"],
    }


def read_reports(summary):
    expected = {
        (row["case"], row["variant"]): f"{row['case']}-{row['variant']}.json"
        for row in summary["variants"]
    }
    archive = DIRECTORY / "baseline-reports.tar.gz"
    reports = {}
    with tarfile.open(archive, "r:gz") as bundle:
        members = bundle.getmembers()
        if {member.name for member in members} != set(expected.values()) or len(members) != 16:
            raise ValueError("raw report file set changed")
        if any(not member.isfile() or member.size > 1_000_000 for member in members):
            raise ValueError("invalid raw report archive member")
        for key, name in expected.items():
            source = bundle.extractfile(name)
            if source is None:
                raise ValueError("raw report missing")
            reports[key] = source.read()
    return reports


def validate_result(summary, review, report_bytes):
    receipts = {(r["case"], r["variant"]): r for r in verify_freeze()}
    if summary.get("format") != "unswell-counterfactual-admission-baseline-v1":
        raise ValueError("unsupported baseline format")
    if summary.get("tool_commit") != EXPECTED_BASELINE_COMMIT:
        raise ValueError("baseline tool commit changed")
    if summary.get("packet_sha256") != sha256((DIRECTORY / "cases.json").read_bytes()):
        raise ValueError("packet hash changed")
    variants = summary.get("variants")
    if not isinstance(variants, list) or len(variants) != 16:
        raise ValueError("baseline variant count changed")
    if review.get("format") != "unswell-counterfactual-admission-baseline-review-v1":
        raise ValueError("unsupported review format")
    if review.get("independent_human") is not False or not review.get("reviewer"):
        raise ValueError("review provenance changed")
    if review.get("applied_edit_safety") != "not_assessed":
        raise ValueError("review safety scope changed")
    seen = set()
    overlapping = {}
    all_findings = 0
    for row in variants:
        key = row.get("case"), row.get("variant")
        if key not in receipts or key in seen:
            raise ValueError("missing, foreign, or duplicate baseline variant")
        seen.add(key)
        receipt = receipts[key]
        if any(row.get(field) != receipt[reference] for field, reference in (
            ("source_sha256", "source_sha256"),
            ("target_start", "target_start"),
            ("target_end", "target_end"),
            ("expected_fixed_diagnosis_applies", "applies"),
        )):
            raise ValueError(f"{key}: source or label receipt changed")
        if row.get("status") != "complete" or row.get("cli_exit") not in (0, 1):
            raise ValueError(f"{key}: incomplete CLI execution")
        raw = report_bytes.get(key)
        if not isinstance(raw, bytes) or sha256(raw) != row.get("report_sha256"):
            raise ValueError(f"{key}: raw report missing or changed")
        report = json.loads(raw)
        if report.get("status") != "complete" or report.get("errors"):
            raise ValueError(f"{key}: incomplete raw report")
        if report["manifest"]["tool_commit"] != EXPECTED_BASELINE_COMMIT:
            raise ValueError(f"{key}: raw tool commit changed")
        if len(report["documents"]) != 1 or report["documents"][0]["source_hash"] != receipt["source_sha256"]:
            raise ValueError(f"{key}: raw source identity changed")
        if report["gate"]["passed"] != row.get("gate_passed"):
            raise ValueError(f"{key}: gate outcome changed")
        findings = [compact_finding(f) for f in report["findings"]]
        if findings != row.get("findings"):
            raise ValueError(f"{key}: compact findings changed")
        start, end = receipt["target_start"], receipt["target_end"]
        matches = [f for f in findings if f["span"]["start"] < end and f["span"]["end"] > start]
        if matches != row.get("target_overlapping_findings"):
            raise ValueError(f"{key}: target overlap changed")
        all_findings += len(findings)
        for finding in matches:
            review_key = key + (finding["id"],)
            if review_key in overlapping:
                raise ValueError("duplicate finding identity")
            overlapping[review_key] = finding
    if seen != set(receipts):
        raise ValueError("baseline omitted a frozen variant")
    decisions = review.get("findings")
    if not isinstance(decisions, list):
        raise ValueError("review findings missing")
    reviewed = set()
    matched = 0
    for decision in decisions:
        key = decision.get("case"), decision.get("variant"), decision.get("finding_id")
        if key not in overlapping or key in reviewed:
            raise ValueError("missing, foreign, or duplicate reviewed finding")
        reviewed.add(key)
        if type(decision.get("matches_fixed_diagnosis")) is not bool or not decision.get("rationale"):
            raise ValueError("review decision or rationale missing")
        matched += decision["matches_fixed_diagnosis"]
    if reviewed != set(overlapping):
        raise ValueError("review omitted a target-overlapping finding")
    return {
        "variants": len(variants),
        "fixed_diagnosis_positive_controls": sum(r["applies"] for r in receipts.values()),
        "all_findings": all_findings,
        "target_overlapping_findings": len(overlapping),
        "matched_fixed_diagnoses": matched,
        "operational_errors": 0,
    }


def verify_files():
    baseline_path = DIRECTORY / "baseline.json"
    review_path = DIRECTORY / "review.json"
    decision_path = DIRECTORY / "decision.json"
    decision = json.loads(decision_path.read_text())
    if decision.get("format") != "unswell-counterfactual-admission-decision-v1":
        raise ValueError("unsupported decision format")
    if sha256(baseline_path.read_bytes()) != decision.get("baseline_sha256"):
        raise ValueError("baseline digest changed")
    if sha256(review_path.read_bytes()) != decision.get("review_sha256"):
        raise ValueError("review digest changed")
    if sha256((DIRECTORY / "baseline-reports.tar.gz").read_bytes()) != decision.get("reports_archive_sha256"):
        raise ValueError("raw report archive digest changed")
    summary = json.loads(baseline_path.read_text())
    review = json.loads(review_path.read_text())
    reports = read_reports(summary)
    result = validate_result(summary, review, reports)
    if result != decision.get("counts"):
        raise ValueError("decision counts changed")
    if decision.get("advance") is not False:
        raise ValueError("baseline cannot advance an unrun candidate")
    return result


if __name__ == "__main__":
    print(verify_files())
