"""Check the archived full-source baseline without changing editorial grades."""

from __future__ import annotations

import gzip
import hashlib
import json
from pathlib import Path

from verify import ROOT, checked_path, require, verify


def check_baseline(root: Path, review: dict, baseline: dict) -> None:
    verify(root, review)
    require(baseline["profile"] == "technical-v1", "changed profile")
    require(baseline["source_role_full"] == 0, "changed FULL accounting")
    require(baseline["source_role_positive_total"] == 3, "changed denominator")
    require(set(baseline["reports"]) == set(review["sources"]), "missing report")
    reports = {}
    for source_id, declared in baseline["reports"].items():
        source = review["sources"][source_id]
        raw_archive = (root / declared["path"]).read_bytes()
        require(hashlib.sha256(raw_archive).hexdigest() == declared["sha256"], f"report digest: {source_id}")
        report = json.loads(gzip.decompress(raw_archive))
        require(report["status"] == "complete" and not report["errors"], f"incomplete report: {source_id}")
        require(report["manifest"]["scoring_profile"] == baseline["profile"], f"report profile: {source_id}")
        require(report["manifest"]["include_source"] is True, f"missing source: {source_id}")
        require(report["manifest"]["no_gate"] is True, f"changed gate: {source_id}")
        require(len(report["documents"]) == 1, f"document count: {source_id}")
        document = report["documents"][0]
        require(document["source_hash"] == source["sha256"], f"document digest: {source_id}")
        original = checked_path(root, source["path"]).read_bytes()
        if source["path"].endswith(".gz"):
            original = gzip.decompress(original)
        require(document["source"].encode("utf-8") == original, f"document bytes: {source_id}")
        require(len(report["findings"]) == declared["findings"], f"finding count: {source_id}")
        require(len(report["assessments"]) == declared["assessments"], f"assessment count: {source_id}")
        reports[source_id] = report

    cases = {case["id"]: case for case in review["cases"]}
    for case_id, selected in baseline["reviewed_findings"].items():
        case = cases[case_id]
        findings = reports[case["source"]]["findings"]
        match = [finding for finding in findings if finding["id"] == selected["finding_id"]]
        require(len(match) == 1, f"missing reviewed finding: {case_id}")
        finding = match[0]
        require(finding["rule_id"] == selected["rule_id"], f"changed rule: {case_id}")
        begin, end = finding["primary"]["span"]["start"], finding["primary"]["span"]["end"]
        require(any(begin < stop and start < end for start, stop, _ in case["spans"]), f"nonoverlapping finding: {case_id}")


if __name__ == "__main__":
    review = json.loads((ROOT / "review.json").read_text(encoding="utf-8"))
    baseline = json.loads((ROOT / "baseline.json").read_text(encoding="utf-8"))
    check_baseline(ROOT, review, baseline)
    print("verified four complete source-bound baseline reports")
