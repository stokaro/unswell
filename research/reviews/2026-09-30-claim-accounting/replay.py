"""Replay an existing private candidate pool through the Go accounting command.

No model calls, new labels, inferred claim splits, or automatic semantic deduplication.
The packet and detailed output stay local. Only the aggregate record is publishable.
"""

import argparse
import hashlib
import json
from pathlib import Path
import subprocess
import time

VERSION = "unswell-editorial-claims-v1"


def sha(data):
    return hashlib.sha256(data).hexdigest()


def require(condition, message):
    if not condition:
        raise ValueError(message)


def invoke(binary, value):
    data = json.dumps(value, ensure_ascii=False).encode()
    process = subprocess.run([str(binary)], input=data, capture_output=True, timeout=15, check=False)
    if process.returncode:
        return None, process.stderr.decode().strip()
    return json.loads(process.stdout), None


def envelope(document, claims):
    return {"version": VERSION, "source": {
        "path": document["page"], "format": document["format"],
        "text": document["source"], "sha256": document["source_sha256"],
    }, "claims": claims, "stages": [], "approvals": [], "include_raw": True}


def project(candidate, units, run):
    targets = []
    for original in candidate["targets"]:
        span = original["span"]
        owners = [unit for unit in units if not unit["excluded"] and
                  unit["span"]["start"] <= span["start"] < span["end"] <= unit["span"]["end"]]
        require(len(owners) == 1, "legacy range lacks one current eligible engine block")
        targets.append({"block": owners[0]["block"], "span": span, "quote": original["quote"]})
    return {"origin": {"run": run, "candidate": candidate["id"], "key": "unsplit-original"},
            "category": candidate["category"], "diagnostic": candidate["diagnostic"],
            "reason": candidate["rationale"], "suggestion": candidate["suggestion"],
            "targets": targets, "support": []}


def replay_candidate(binary, document, candidate, units, run):
    try:
        specification = project(candidate, units, run)
    except ValueError as error:
        return {"candidate": candidate["id"], "status": "invalid", "reason": str(error)}
    request = envelope(document, [specification])
    prepared, error = invoke(binary, request)
    if error:
        return {"candidate": candidate["id"], "status": "invalid", "reason": error}
    require(len(prepared["claims"]) == 1 and not prepared["complete"], "invalid preparation inventory")
    claim = prepared["claims"][0]
    require(claim["targets"] == specification["targets"] and claim["support"] == [], "source roles changed")
    request["stages"] = [{"id": "legacy-retention-replay", "decisions": [{
        "claim_id": claim["id"], "status": "retained", "edit_id": "",
        "reason": "Preserve the original candidate; this is not a new editorial judgment.",
    }], "edits": [], "duplicates": []}]
    result, error = invoke(binary, request)
    require(error is None, "prepared candidate failed unchanged retention replay: " + str(error))
    require(result["complete"] and result["editorial_qualified"] is False, "accounting was misrepresented as qualification")
    account = result["stages"][0]
    require(result["claims"] == account["claims"] == prepared["claims"], "original claim mutated")
    require(account["displayed_claims"] == [claim["id"]] and len(account["decisions"]) == 1, "original claim disappeared")
    return {"candidate": candidate["id"], "status": "retained", "claim": claim["id"]}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--packet", type=Path, required=True)
    parser.add_argument("--binary", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    args.output.mkdir(parents=True, exist_ok=False)
    manifest_data = (args.packet / "manifest.json").read_bytes()
    manifest = json.loads(manifest_data)
    require(len(manifest["pages"]) == 36, "changed exposed page denominator")
    run = "legacy-pool-" + sha(manifest_data)
    started = time.monotonic()
    rows, hashes, original_ids = [], {}, set()
    for page in manifest["pages"]:
        path = args.packet / "requests" / (page["page"] + ".json")
        data = path.read_bytes()
        require(sha(data) == page["request_sha256"], "frozen request changed")
        hashes[path.name] = sha(data)
        request = json.loads(data)
        document, candidates = request["document"], request["candidates"]
        require(sha(document["source"].encode()) == page["source_sha256"] == document["source_sha256"], "source changed")
        require(len(candidates) == page["candidates"], "candidate denominator changed")
        prepared, error = invoke(args.binary, envelope(document, []))
        require(error is None, "original source extraction failed: " + str(error))
        page_rows = []
        for candidate in candidates:
            require(candidate["id"] not in original_ids, "duplicate original candidate identity")
            original_ids.add(candidate["id"])
            page_rows.append(replay_candidate(args.binary, document, candidate, prepared["units"], run))
        rows.extend(page_rows)
        (args.output / (page["page"] + ".json")).write_text(json.dumps(page_rows, indent=2) + "\n")
    require(len(rows) == len(original_ids) == 804, "changed full candidate denominator")
    retained = sum(row["status"] == "retained" for row in rows)
    invalid = sum(row["status"] == "invalid" for row in rows)
    require(retained + invalid == 804, "candidate escaped accounting")
    reasons = {}
    for row in rows:
        if row["status"] == "invalid":
            reasons[row["reason"]] = reasons.get(row["reason"], 0) + 1
    result = {"version": "unswell-claim-accounting-replay-v1", "pages": 36, "original_candidates": 804,
              "retained": retained, "invalid": invalid, "unaccounted": 0, "duplicates_collapsed": 0,
              "reference_event_denominators": {"whole_page": 57, "contextual": 46, "historical_exposed": 20, "total": 123},
              "full_event_metrics_recomputed": False, "new_editorial_labels": 0, "new_model_calls": 0,
              "independent_claim_splits_inferred": False, "legacy_support_roles_reviewed": False,
              "editorial_qualified": False, "seconds": time.monotonic() - started,
              "packet_manifest_sha256": sha(manifest_data), "binary_sha256": sha(args.binary.read_bytes()),
              "request_hashes": hashes, "invalid_reasons": reasons,
              "scope": "Mechanical source binding and exact inventory retention only. Legacy target roles remain declared, unsplit, and unqualified."}
    (args.output / "summary.json").write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps({key: result[key] for key in ["pages", "original_candidates", "retained", "invalid", "unaccounted", "seconds"]}, indent=2))


if __name__ == "__main__":
    main()
