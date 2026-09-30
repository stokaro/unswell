"""Adopt source-verified rule observations as an explicit child research record.

The original packet, quotations, labels, and targets remain unchanged. Retention
replays inventory preservation, not a new verdict about editorial usefulness.
"""

import argparse
import copy
import hashlib
import importlib.util
import json
from pathlib import Path
import time

ROOT = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("legacy_claim_replay", ROOT.parent / "2026-09-30-claim-accounting/replay.py")
LEGACY = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(LEGACY)


def sha(data):
    return hashlib.sha256(data).hexdigest()


def require(condition, message):
    if not condition:
        raise ValueError(message)


def read(path):
    return json.loads(path.read_text())


def index_report(report, observations):
    require(report["status"] == "complete", "Incomplete reconstructed report")
    require(observations["version"] == "unswell-rule-evidence-v1" and observations["editorial_qualified"] is False,
            "Unsupported observational contract or qualification claim")
    require(observations["tool_commit"] == report["manifest"]["tool_commit"] and
            observations["config_hash"] == report["manifest"]["config_hash"], "Report identity changed")
    sources = {doc["name"]: doc["source_hash"] for doc in report["documents"]}
    require(len(sources) == len(report["documents"]), "Duplicate source identity")
    findings = {finding["id"]: finding for finding in report["findings"]}
    rows = {row["id"]: row for row in observations["observations"]}
    require(len(findings) == len(report["findings"]) == len(rows) == len(observations["observations"])
            and set(findings) == set(rows), "Missing or duplicate original finding")
    for identity, row in rows.items():
        finding = findings[identity]
        require(row["source_sha256"] == sources[finding["primary"]["path"]], "Observation source changed")
        require(row["rule_id"] == finding["rule_id"] and row["rule_version"] == finding["rule_version"]
                and row["diagnostic"] == finding["message"], "Original rule or diagnostic changed")
        require(row["evidence"] == finding["evidence"] and row["suppressed"] == finding["suppressed"], "Rule evidence changed")
        for field in ["primary", "related"]:
            original = copy.deepcopy(finding[field])
            locations = [original] if field == "primary" else original
            for location in locations:
                location.pop("snippet", None)
            require(row[field] == original, "Original source location changed")
        require(row["editorial_verdict"] == "unreviewed" and row["reason"].strip() and row["action"].strip(),
                "Missing observations or invented editorial verdict")
        suggestion = finding["evidence"]["suggestion"]
        require(row["action"] == (suggestion or finding["message"]) and
                row["action_basis"] == ("evidence_suggestion" if suggestion else "engine_diagnostic"),
                "An action was invented instead of preserving engine guidance")
    return findings, rows


def bind_original(candidate, document, reference, finding, observation):
    require(candidate["origin"] == "rule" and candidate["id"] == "rule/" + finding["id"], "Original candidate identity changed")
    require(finding["primary"]["path"] == reference and observation["source_sha256"] == document["source_sha256"],
            "Original document identity changed")
    require(candidate["category"] == finding["rule_id"] and candidate["diagnostic"] == finding["message"], "Original diagnosis changed")
    raw, targets = document["source"].encode(), []
    require(sha(raw) == document["source_sha256"], "Original source bytes changed")
    for location in [finding["primary"]] + finding["related"]:
        require(location["path"] == reference, "Original finding crosses documents")
        for span in location.get("segments", [location["span"]]):
            start, end = span["start"], span["end"]
            require(0 <= start < end <= len(raw), "Invalid original source span")
            targets.append((span, raw[start:end].decode()))
    require(targets == [(target["span"], target["quote"]) for target in candidate["targets"]], "Original targets or quotes changed")
    child = copy.deepcopy(candidate)
    child["rationale"], child["suggestion"] = observation["reason"], observation["action"]
    return child


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ["packet", "report", "observations", "binary", "output"]:
        parser.add_argument("--" + name, type=Path, required=True)
    args = parser.parse_args()
    args.output.mkdir(parents=True, exist_ok=False)
    report, observations = read(args.report), read(args.observations)
    findings, indexed = index_report(report, observations)
    manifest_data = (args.packet / "manifest.json").read_bytes()
    manifest, freeze = json.loads(manifest_data), read(args.packet / "freeze.json")
    require(sha(manifest_data) == freeze["manifest.json"] and len(manifest["pages"]) == 36, "Frozen page inventory changed")
    require(report["manifest"]["tool_commit"] == "e146774df81ee8c106ece4abc4017d41c23a46e0", "Different original baseline commit")
    require({doc["name"]: doc["source_hash"] for doc in report["documents"]} ==
            {page["reference"]: page["source_sha256"] for page in manifest["pages"]}, "Reconstructed source inventory changed")
    started, rows, seen = time.monotonic(), [], set()
    lineage = "rule-observations-" + sha(args.observations.read_bytes())
    for page in manifest["pages"]:
        path = "requests/" + page["page"] + ".json"
        data = (args.packet / path).read_bytes()
        require(sha(data) == freeze[path] == page["request_sha256"], "Frozen request changed")
        request = json.loads(data)
        document = request["document"]
        prepared, error = LEGACY.invoke(args.binary, LEGACY.envelope(document, []))
        require(error is None, "Original source extraction failed: " + str(error))
        page_rows = []
        for candidate in request["candidates"]:
            if candidate["origin"] != "rule":
                continue
            identity = candidate["id"].removeprefix("rule/")
            require(identity in findings and identity not in seen, "Missing or repeated original rule candidate")
            seen.add(identity)
            child = bind_original(candidate, document, page["reference"], findings[identity], indexed[identity])
            row = LEGACY.replay_candidate(args.binary, document, child, prepared["units"], lineage)
            row.update(original_candidate=candidate["id"], original_targets_unchanged=True,
                       evidence_sha256=sha(json.dumps(indexed[identity], sort_keys=True).encode()),
                       new_editorial_judgment=False)
            page_rows.append(row)
        rows.extend(page_rows)
        (args.output / (page["page"] + ".json")).write_text(json.dumps(page_rows, indent=2) + "\n")
    require(len(rows) == len(seen) == len(findings) == 165, "Changed original rule denominator")
    counts, invalid = {}, {}
    for row in rows:
        counts[row["status"]] = counts.get(row["status"], 0) + 1
        if row["status"] == "invalid":
            invalid[row["reason"]] = invalid.get(row["reason"], 0) + 1
    summary = {"version": "unswell-rule-observation-replay-v1", "pages": 36, "original_rule_candidates": 165,
        "observations_preserved": len(rows), "counts": counts, "invalid_reasons": invalid,
        "unaccounted": 0, "source_targets_changed": 0, "support_roles_inferred": False,
        "independent_claim_splits_inferred": False, "semantic_duplicates_collapsed": 0,
        "new_editorial_judgments": 0, "frozen_inputs_modified": False, "original_gzip_bytes_recovered": False,
        "reference_event_denominators": {"whole": 57, "context": 46, "historical_exposed": 20, "total": 123},
        "full_event_metrics_recomputed": False, "editorial_qualified": False, "new_model_calls": 0,
        "packet_manifest_sha256": sha(manifest_data), "report_sha256": sha(args.report.read_bytes()),
        "observations_sha256": sha(args.observations.read_bytes()), "binary_sha256": sha(args.binary.read_bytes()),
        "seconds": time.monotonic() - started}
    (args.output / "summary.json").write_text(json.dumps(summary, indent=2) + "\n")
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()
