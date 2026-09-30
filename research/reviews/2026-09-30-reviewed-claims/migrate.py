"""Replay explicitly reviewed claim boundaries without replacing the old pool.

This local research consumer calls the existing Go source/accounting command.
It never infers a semantic split, a support role, a quality label, or an edit.
Unreviewed parents remain in the complete original inventory.
"""

import argparse
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import time

ROOT = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("legacy", ROOT.parent / "2026-09-30-claim-accounting/replay.py")
LEGACY = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(LEGACY)
VERSION = "unswell-reviewed-claim-migration-v1"


def require(condition, message):
    if not condition:
        raise ValueError(message)


def sha(data):
    return hashlib.sha256(data).hexdigest()


def digest(value):
    return sha(json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":")).encode())


def unique_pairs(pairs):
    result = {}
    for key, value in pairs:
        require(key not in result, "Duplicate JSON field: " + key)
        result[key] = value
    return result


def read(path):
    data = path.read_bytes()
    require(len(data) <= 32 << 20, "Research input exceeds the 32 MiB limit")
    return json.loads(data.decode("utf-8"), object_pairs_hook=unique_pairs)


def fields(value, expected):
    require(isinstance(value, dict) and set(value) == set(expected), "Missing or unknown declaration fields")


def text(value):
    require(isinstance(value, str) and value.strip() and len(value.encode()) <= 2400,
            "Missing or excessive review identity or explanation")


def references(document, parent, declared, units, target=True):
    require(isinstance(declared, list) and len(declared) <= 64, "Excessive source references")
    raw, result = document["source"].encode(), []
    for reference in declared:
        fields(reference, ["span", "quote"])
        fields(reference["span"], ["start", "end"])
        start, end = reference["span"]["start"], reference["span"]["end"]
        require(type(start) is int and type(end) is int and 0 <= start < end <= len(raw), "Invalid UTF-8 byte range")
        require(raw[start:end].decode() == reference["quote"], "Reviewed quotation differs from the original bytes")
        require(not target or any(t["span"]["start"] <= start < end <= t["span"]["end"] for t in parent["targets"]),
                "Reviewed reference escapes the original candidate locations")
        owners = [u for u in units if not u["excluded"] and u["span"]["start"] <= start < end <= u["span"]["end"]]
        require(len(owners) == 1, "Reviewed reference lacks one eligible existing engine block")
        result.append({"block": owners[0]["block"], **reference})
    return result


def declarations(document, parent, decision, units, run):
    fields(decision, ["parent", "parent_sha256", "status", "reason", "claims"])
    require(decision["parent"] == parent["id"] and decision["parent_sha256"] == digest(parent),
            "Migration review is not bound to the complete unchanged original candidate")
    text(decision["reason"])
    status, claims = decision["status"], decision["claims"]
    require(status in ["migrated", "rejected", "uncertain"] and isinstance(claims, list) and len(claims) <= 64,
            "Unsupported or excessive migration decision")
    require(bool(claims) == (status == "migrated"), "Only a migrated parent must declare its original criticisms")
    result, seen = [], set()
    for claim in claims:
        fields(claim, ["key", "category", "diagnostic", "reason", "suggestion", "targets", "support"])
        for field in ["key", "category", "diagnostic", "reason", "suggestion"]:
            text(claim[field])
        require(claim["key"] not in seen, "Repeated independently declared criticism")
        seen.add(claim["key"])
        targets = references(document, parent, claim["targets"], units)
        support = references(document, parent, claim["support"], units, target=False)
        require(targets, "An independently declared claim needs a revision target")
        result.append({"origin": {"run": run, "candidate": parent["id"], "key": claim["key"]},
                       **{k: claim[k] for k in ["category", "diagnostic", "reason", "suggestion"]},
                       "targets": targets, "support": support})
    return result


def index_review(review, manifest):
    fields(review, ["version", "packet_manifest_sha256", "reviewer", "pages"])
    require(review["version"] == VERSION and review["packet_manifest_sha256"] == digest_manifest(manifest),
            "Unsupported migration version or a different frozen pool")
    fields(review["reviewer"], ["id", "basis"])
    text(review["reviewer"]["id"])
    text(review["reviewer"]["basis"])
    require(isinstance(review["pages"], list) and len(review["pages"]) <= len(manifest["pages"]), "Excessive reviewed pages")
    pages, seen = {}, set()
    known = {p["page"]: p for p in manifest["pages"]}
    require(len(known) == len(manifest["pages"]), "Repeated page identity")
    for page in review["pages"]:
        fields(page, ["page", "request_sha256", "decisions"])
        require(page["page"] in known and page["page"] not in pages and
                page["request_sha256"] == known[page["page"]]["request_sha256"], "Unknown, repeated, or changed reviewed page")
        require(isinstance(page["decisions"], list) and len(page["decisions"]) <= known[page["page"]]["candidates"],
                "Excessive parent decisions")
        by_parent = {}
        for decision in page["decisions"]:
            identity = decision["parent"]
            require(identity not in seen, "Repeated original candidate review")
            seen.add(identity)
            by_parent[identity] = decision
        pages[page["page"]] = by_parent
    return pages


def digest_manifest(manifest):
    # The manifest is hashed as its exact frozen file, not as a JSON reserialization.
    return manifest["_file_sha256"]


def store(path, value):
    with os.fdopen(os.open(path, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600), "w") as handle:
        json.dump(value, handle, ensure_ascii=False, indent=2)
        handle.write("\n")


def run(packet, binary, review_path, output):
    manifest_data = (packet / "manifest.json").read_bytes()
    manifest, freeze, review = json.loads(manifest_data), read(packet / "freeze.json"), read(review_path)
    require(sha(manifest_data) == freeze["manifest.json"], "Frozen manifest changed")
    manifest["_file_sha256"] = sha(manifest_data)
    indexed = index_review(review, manifest)
    output.mkdir(parents=True, exist_ok=False)
    started, counts, identities, child_count = time.monotonic(), {}, set(), 0
    # Adding a separately reviewed page must not rename existing original claims.
    # Keep the exact review hash in the record, outside the immutable claim origin.
    run_id = "reviewed-migration-" + sha(manifest_data)
    for page in manifest["pages"]:
        relative = "requests/" + page["page"] + ".json"
        data = (packet / relative).read_bytes()
        require(sha(data) == freeze[relative] == page["request_sha256"], "Frozen candidate request changed")
        request = json.loads(data)
        document, parents = request["document"], request["candidates"]
        require(len(parents) == page["candidates"] and sha(document["source"].encode()) == page["source_sha256"] ==
                document["source_sha256"], "Original source or candidate inventory changed")
        prepared, error = LEGACY.invoke(binary, LEGACY.envelope(document, []))
        require(error is None, "Existing Go extraction failed: " + str(error))
        decisions = indexed.get(page["page"], {})
        require(set(decisions) <= {p["id"] for p in parents}, "Unknown original candidate review")
        records, specifications = [], []
        for parent in parents:
            require(parent["id"] not in identities, "Repeated original candidate identity")
            identities.add(parent["id"])
            decision = decisions.get(parent["id"])
            children = declarations(document, parent, decision, prepared["units"], run_id) if decision else []
            status = decision["status"] if decision else "unreviewed"
            counts[status] = counts.get(status, 0) + 1
            specifications.extend(children)
            records.append({"original": parent, "original_sha256": digest(parent), "migration": decision,
                            "status": status, "child_keys": [c["origin"]["key"] for c in children]})
        envelope = LEGACY.envelope(document, specifications)
        bound, error = LEGACY.invoke(binary, envelope)
        require(error is None, "Reviewed declaration failed existing Go binding: " + str(error))
        claims = bound["claims"]
        require(len(claims) == len(specifications), "A reviewed original criticism disappeared")
        for original, claim in zip(specifications, claims):
            require({k: v for k, v in claim.items() if k != "id"} == original, "Reviewed criticism or roles changed")
        envelope["stages"] = [{"id": "reviewed-inventory-retention", "decisions": [
            {"claim_id": c["id"], "status": "retained", "edit_id": "", "reason":
             "Retain this explicitly reviewed original criticism; this stage does not judge editorial acceptance."}
            for c in claims], "edits": [], "duplicates": []}]
        retained, error = LEGACY.invoke(binary, envelope)
        require(error is None and retained["complete"] and retained["editorial_qualified"] is False,
                "Reviewed retention failed or claimed editorial qualification: " + str(error))
        require(retained["claims"] == claims == retained["stages"][0]["claims"] and
                retained["stages"][0]["displayed_claims"] == [c["id"] for c in claims], "An original criticism lost its identity")
        child_count += len(claims)
        store(output / (page["page"] + ".json"), {"source_sha256": document["source_sha256"], "reviewer": review["reviewer"],
                                                 "parents": records, "account": retained})
    total = sum(p["candidates"] for p in manifest["pages"])
    require(len(identities) == sum(counts.values()) == total, "Original candidates escaped accounting")
    summary = {"version": VERSION, "pages": len(manifest["pages"]), "pages_with_migration_decisions": len(indexed),
               "original_candidates": total,
               "parent_dispositions": counts, "bound_child_claims": child_count,
               "complete": not counts.get("unreviewed", 0) and not counts.get("uncertain", 0),
               "unaccounted": 0, "semantic_duplicates_collapsed": 0, "new_model_calls": 0,
               "frozen_inputs_modified": False, "full_event_metrics_recomputed": False, "editorial_qualified": False,
               "source_roles_reviewed_by": "caller-supplied reviewer; not inferred by the harness",
               "packet_manifest_sha256": sha(manifest_data), "review_sha256": sha(review_path.read_bytes()),
               "binary_sha256": sha(binary.read_bytes()), "reviewer": review["reviewer"], "seconds": time.monotonic() - started}
    store(output / "summary.json", summary)
    return summary


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ["packet", "binary", "review", "output"]:
        parser.add_argument("--" + name, type=Path, required=True)
    args = parser.parse_args()
    result = run(args.packet, args.binary, args.review, args.output)
    print(json.dumps(result, indent=2))
    return 0 if result["complete"] else 2


if __name__ == "__main__":
    raise SystemExit(main())
