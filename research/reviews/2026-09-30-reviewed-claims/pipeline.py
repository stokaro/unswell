"""Replay audit, rewrite, and selection without losing original editorial claims.

This optional local consumer uses the existing Go source and accounting command.
Stage responses and caller-supplied edit/duplicate approvals are separate inputs.
Successful accounting never qualifies a detector or replaces frozen labels.
"""

import argparse
import importlib.util
import json
from pathlib import Path
import subprocess
import time

SPEC = importlib.util.spec_from_file_location("migration", Path(__file__).with_name("migrate.py"))
M = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(M)
VERSION = "unswell-reviewed-claim-pipeline-v1"
APPROVAL_VERSION = "unswell-reviewed-claim-approvals-v1"
STAGES = ("audit", "rewrite", "selection")


def read_hashed(path):
    data = path.read_bytes()
    M.require(len(data) <= 32 << 20, "Research input exceeds the 32 MiB limit")
    return json.loads(data.decode("utf-8"), object_pairs_hook=M.unique_pairs), M.sha(data)


def migration(packet, binary, review, output):
    output.mkdir(parents=True, exist_ok=False)
    summary = M.run(packet, binary, review, output / "migration")
    M.require(summary["complete"], "Every original parent needs an explicit migration decision before pipeline replay")
    manifest = M.read(packet / "manifest.json")
    records = [(page, M.read(output / "migration" / (page["page"] + ".json"))) for page in manifest["pages"]]
    return summary, records


def identity(summary, version):
    return {"version": version, "packet_manifest_sha256": summary["packet_manifest_sha256"],
            "review_sha256": summary["review_sha256"]}


def page_identity(page, record):
    return {"page": page["page"], "request_sha256": page["request_sha256"],
            "claims_sha256": M.digest(record["account"]["claims"])}


def prepare(packet, binary, review, output):
    summary, records = migration(packet, binary, review, output)
    stages, approvals = {**identity(summary, VERSION), "pages": []}, []
    for page, record in records:
        pending = [{"id": stage, "decisions": [{"claim_id": claim["id"], "status": "uncertain", "edit_id": "",
                    "reason": "No " + stage + " assessment has been supplied for this original claim."}
                   for claim in record["account"]["claims"]], "edits": [], "duplicates": []} for stage in STAGES]
        stages["pages"].append({**page_identity(page, record), "stages": pending})
        approvals.append({**page_identity(page, record), "approvals": []})
    M.store(output / "stages.json", stages)
    M.store(output / "approvals.json", {**identity(summary, APPROVAL_VERSION),
            "stages_sha256": M.sha((output / "stages.json").read_bytes()), "pages": approvals})
    result = {**identity(summary, VERSION), "prepared": True, "complete": False,
              "pages": summary["pages"], "original_candidates": summary["original_candidates"],
              "bound_child_claims": summary["bound_child_claims"], "required_stages": list(STAGES),
              "editorial_qualified": False, "full_event_metrics_recomputed": False, "new_model_calls": 0}
    M.store(output / "preparation.json", result)
    return result


def index_pages(value, summary, records, approvals=False):
    extra = ["stages_sha256"] if approvals else []
    M.fields(value, ["version", "packet_manifest_sha256", "review_sha256", "pages", *extra])
    for key, expected in identity(summary, APPROVAL_VERSION if approvals else VERSION).items():
        M.require(value[key] == expected, "Pipeline input names an unsupported version or different migration")
    M.require(isinstance(value["pages"], list) and len(value["pages"]) == len(records),
              "Every original page needs an explicit pipeline input")
    known, result = {p["page"]: page_identity(p, r) for p, r in records}, {}
    for page in value["pages"]:
        field = "approvals" if approvals else "stages"
        M.fields(page, ["page", "request_sha256", "claims_sha256", field])
        name = page["page"]
        M.require(name in known and name not in result and all(page[k] == v for k, v in known[name].items()),
                  "Unknown, repeated, or changed pipeline page or claim inventory")
        M.require(isinstance(page[field], list), "Missing pipeline stage or approval array")
        if not approvals:
            M.require(len(page[field]) == len(STAGES), "Audit, rewrite, and selection are all required")
            for stage, expected in zip(page[field], STAGES):
                M.fields(stage, ["id", "decisions", "edits", "duplicates"])
                M.require(stage["id"] == expected, "Pipeline stages must appear as audit, rewrite, selection")
        result[name] = page[field]
    return result


def invoke(binary, request):
    process = subprocess.run([str(binary)], input=json.dumps(request, ensure_ascii=False).encode(),
                             capture_output=True, timeout=15, check=False)
    error = process.stderr.decode("utf-8").strip()
    M.require(process.returncode in [0, 2] and process.stdout,
              "Existing Go accounting failed: " + error)
    result = json.loads(process.stdout.decode("utf-8"), object_pairs_hook=M.unique_pairs)
    if process.returncode == 2:
        M.require(result.get("complete") is False and error == "required claim accounting remains uncertain",
                  "Existing Go accounting failed without an explicit uncertain result: " + error)
    else:
        M.require(result.get("complete") is True and not error, "Go pipeline did not produce complete accounting")
    return result


def verify_account(result, claims, stages, approvals, source_hash):
    M.require(result["version"] == M.LEGACY.VERSION and result["source_sha256"] == source_hash and
              result["claims"] == claims and result["editorial_qualified"] is False,
              "Go accounting changed the original source, claims, or qualification")
    M.require(len(result["stages"]) == len(STAGES), "A required downstream stage disappeared")
    identities = {claim["id"] for claim in claims}
    M.require(len(identities) == len(claims), "Repeated original child identity")
    for account, stage in zip(result["stages"], stages):
        M.require(account["stage_id"] == stage["id"] and account["claims"] == claims and
                  account["source_sha256"] == source_hash and account["editorial_qualified"] is False,
                  "A downstream stage mutated its original inventory")
        for field in ["decisions", "edits", "duplicates"]:
            M.require(account[field] == stage[field], "Go accounting changed a supplied stage decision")
        expected_approvals = [a for a in approvals if a["stage_id"] == stage["id"]]
        M.require(account["approvals"] == expected_approvals and
                  len(account["decisions"]) == len(claims) and
                  {d["claim_id"] for d in account["decisions"]} == identities,
                  "An original claim or caller approval escaped accounting")
        complete = all(d["status"] != "uncertain" for d in account["decisions"])
        M.require(account["complete"] is complete, "An uncertain claim passed a downstream stage")
    M.require(result["complete"] is all(a["complete"] for a in result["stages"]),
              "An incomplete earlier stage was hidden by final selection")


def parent_accounts(parents, claims, account):
    decisions = {d["claim_id"]: d for d in account["decisions"]}
    displayed = set(account["displayed_claims"])
    result = []
    for parent in parents:
        children = [c for c in claims if c["origin"]["candidate"] == parent["original"]["id"]]
        M.require([c["origin"]["key"] for c in children] == parent["child_keys"],
                  "A parent's independently declared criticism disappeared")
        dispositions = [{**decisions[c["id"]], "displayed": c["id"] in displayed} for c in children]
        statuses = {d["status"] for d in dispositions}
        status = next(iter(statuses)) if len(statuses) == 1 else "mixed"
        if not children:
            M.require(parent["status"] == "rejected", "A nonmigrated parent lacks an explicit scope decision")
            status = "migration_rejected"
        result.append({**parent, "pipeline_status": status, "children": dispositions})
    return result


def replay(packet, binary, review, stages_path, approvals_path, output):
    M.require(stages_path.resolve() != approvals_path.resolve(), "Stage output cannot supply its own approvals")
    summary, records = migration(packet, binary, review, output)
    started = time.monotonic()
    stages_data, stages_hash = read_hashed(stages_path)
    approvals_data, approvals_hash = read_hashed(approvals_path)
    M.require(approvals_data.get("stages_sha256") == stages_hash,
              "Caller approvals are not bound to the exact supplied stage file")
    stages = index_pages(stages_data, summary, records)
    approvals = index_pages(approvals_data, summary, records, approvals=True)
    complete, stage_counts, parent_counts, displayed = True, {s: {} for s in STAGES}, {}, 0
    for page, record in records:
        name, claims = page["page"], record["account"]["claims"]
        request, request_hash = read_hashed(packet / "requests" / (name + ".json"))
        M.require(request_hash == page["request_sha256"], "Frozen candidate request changed during replay")
        envelope = M.LEGACY.envelope(request["document"], [{k: v for k, v in c.items() if k != "id"} for c in claims])
        envelope.update(stages=stages[name], approvals=approvals[name])
        result = invoke(binary, envelope)
        verify_account(result, claims, stages[name], approvals[name], record["source_sha256"])
        complete = complete and result["complete"]
        parents_by_stage = []
        for account in result["stages"]:
            counts = stage_counts[account["stage_id"]]
            for decision in account["decisions"]:
                counts[decision["status"]] = counts.get(decision["status"], 0) + 1
            parents_by_stage.append({"stage_id": account["stage_id"],
                                     "parents": parent_accounts(record["parents"], claims, account)})
        for parent in parents_by_stage[-1]["parents"]:
            status = parent["pipeline_status"]
            parent_counts[status] = parent_counts.get(status, 0) + 1
        displayed += len(result["stages"][-1]["displayed_claims"])
        M.store(output / (name + ".json"), {"source_sha256": record["source_sha256"],
                "migration_reviewer": record["reviewer"], "parents_by_stage": parents_by_stage, "account": result})
    M.require(all(sum(c.values()) == summary["bound_child_claims"] for c in stage_counts.values()) and
              sum(parent_counts.values()) == summary["original_candidates"], "Original claims escaped downstream accounting")
    result = {**identity(summary, VERSION), "pages": summary["pages"], "original_candidates": summary["original_candidates"],
              "bound_child_claims": summary["bound_child_claims"], "required_stages": list(STAGES), "complete": complete,
              "stage_dispositions": stage_counts, "selected_parent_dispositions": parent_counts,
              "selected_displayed_claims": displayed, "unaccounted": 0, "new_model_calls": 0,
              "editorial_qualified": False, "full_event_metrics_recomputed": False, "frozen_inputs_modified": False,
              "stages_sha256": stages_hash, "approvals_sha256": approvals_hash,
              "binary_sha256": summary["binary_sha256"], "seconds": time.monotonic() - started}
    M.store(output / "summary.json", result)
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("mode", choices=["prepare", "replay"])
    for name in ["packet", "binary", "review", "output"]:
        parser.add_argument("--" + name, type=Path, required=True)
    parser.add_argument("--stages", type=Path)
    parser.add_argument("--approvals", type=Path)
    args = parser.parse_args()
    if args.mode == "prepare":
        if args.stages or args.approvals:
            parser.error("prepare does not accept stage responses or approvals")
        result = prepare(args.packet, args.binary, args.review, args.output)
    else:
        if not args.stages or not args.approvals:
            parser.error("replay requires separate --stages and --approvals inputs")
        result = replay(args.packet, args.binary, args.review, args.stages, args.approvals, args.output)
    print(json.dumps(result, indent=2))
    return 0 if args.mode == "prepare" or result["complete"] else 2


if __name__ == "__main__":
    raise SystemExit(main())
