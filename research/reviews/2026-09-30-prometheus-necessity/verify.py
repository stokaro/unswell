"""Verify the frozen local judge's completed admission screen without ML libraries."""

import hashlib
import json
from pathlib import Path
import re
import tarfile

ROOT = Path(__file__).resolve().parent


def require(condition, reason):
    if not condition:
        raise ValueError(reason)


def read(path):
    return json.loads(path.read_text())


def sha(data):
    return hashlib.sha256(data).hexdigest()


def decision(rows, expectations):
    cases = expectations["cases"]
    require(len(rows) == len(cases) == 35 and len({r["id"] for r in rows}) == 35
            and {r["id"] for r in rows} == set(cases), "Changed control inventory")
    groups = {}
    for row in rows:
        expected = cases[row["id"]]
        require(type(expected["retain"]) is bool, "Invalid expected disposition")
        require(row["status"] == "complete" and row["finish_reason"] == "stop", "Unavailable or truncated control")
        match = re.fullmatch(r"(?s)(.+?)\[RESULT\]\s*([1-5])\s*", row["raw"])
        require(row["raw"].count("[RESULT]") == 1 and match is not None and len(match[1].strip()) >= 20,
                "Invalid terminal ordinal grade")
        score = int(match[2])
        require(type(row["score"]) is int and row["score"] == score and type(row["retained"]) is bool
                and row["retained"] == (score >= 4), "Recorded grade or decision changed")
        group = groups.setdefault(expected["group"], {"total": 0, "available": 0, "retained": 0, "correct": 0})
        group["total"] += 1
        group["available"] += 1
        group["retained"] += row["retained"]
        group["correct"] += row["retained"] == expected["retain"]
    require({name: group["total"] for name, group in groups.items()} == {
        "useful": 12, "unsafe": 12, "optional": 4, "identifier": 2, "claim_loss": 1, "context": 4}, "Changed control groups")
    checks = {"useful_at_least_10_of_12": groups["useful"]["retained"] >= 10,
        "unsafe_at_most_1_of_12": groups["unsafe"]["retained"] <= 1,
        "optional_at_most_1_of_4": groups["optional"]["retained"] <= 1,
        "both_identifier_cases_correct": groups["identifier"]["correct"] == 2,
        "grammar_does_not_erase_rhetoric": groups["claim_loss"]["correct"] == 1,
        "context_at_least_3_of_4": groups["context"]["correct"] >= 3}
    return {"groups": groups, "checks": checks, "passed": all(checks.values()),
        "interpretation": "Constructed task-admission screen only, not real-document recall or product accuracy."}


def verify(root=ROOT):
    manifest = read(root / "manifest.json")
    for name, identity in manifest["files"].items():
        require(Path(name).name == name, "Unsafe public artifact name")
        data = (root / name).read_bytes()
        require(len(data) == identity["bytes"] and sha(data) == identity["sha256"], "Changed public artifact: " + name)
    require(manifest["development_payload_published"] is False and manifest["frozen_editorial_annotations_published"] is False
            and manifest["constructed_expectations_published"] is True and manifest["weights_published"] is False,
            "Unexpected private data or checkpoint publication")
    archived = {}
    with tarfile.open(root / "runner-source.tar.gz", "r:gz") as archive:
        members = archive.getmembers()
        require(len(members) == len(manifest["runner_members"]) and {m.name for m in members} == set(manifest["runner_members"]),
                "Changed runner archive inventory")
        for member in members:
            require(member.isfile() and 0 < member.size <= 4 * 1024**2 and not Path(member.name).is_absolute()
                    and ".." not in Path(member.name).parts, "Unsafe archived source")
            data = archive.extractfile(member).read()
            identity = manifest["runner_members"][member.name]
            require(len(data) == identity["bytes"] and sha(data) == identity["sha256"], "Changed runner member")
            archived[member.name] = data
    freeze = read(root / "freeze.json")
    private_names = set()
    for name, digest in freeze["files"].items():
        if (root / name).is_file():
            require(sha((root / name).read_bytes()) == digest, "Changed frozen public input")
        elif name in archived:
            require(sha(archived[name]) == digest, "Changed frozen runner")
        else:
            private_names.add(name)
    require(private_names == {"development-requests.json"}, "Unexpected unavailable frozen artifact")
    protocol, model, acquired = (read(root / name) for name in ["protocol.json", "model-manifest.json", "acquisition-record.json"])
    require(acquired["model"] == model["model"] == protocol["model"] and
            acquired["revision"] == model["revision"] == protocol["revision"], "Changed checkpoint identity")
    weights = [item for item in model["files"] if item["kind"] == "weight_shard"]
    require(len(weights) == 8 and acquired["shards"] == [{k: item[k] for k in ["name", "bytes", "sha256"]} for item in weights],
            "Incomplete verified shard inventory")
    require(acquired["parameter_bytes"] == 14483464192 and acquired["tensors"] == 291 and
            acquired["precision"] == "BF16, unquantized" and acquired["temporary_shards_removed"] is True, "Changed checkpoint precision or shape inventory")
    for item in model["files"]:
        if item["kind"] == "metadata":
            require(sha(archived["model/" + item["name"]]) == item["sha256"], "Changed pinned model metadata")
    requests = read(root / "controls-requests.json")["requests"]
    rows = read(root / "predictions.json")["rows"]
    require(len(requests) == len(rows) == 35, "Changed request denominator")
    for row, request in zip(rows, requests, strict=True):
        require(row["id"] == request["id"] and row["request_sha256"] == sha(json.dumps(request, sort_keys=True).encode()), "Changed model request")
    result = decision(rows, read(root / "control-expectations.json"))
    record = read(root / "run-record.json")
    require(result == read(root / "preflight-summary.json") == record["preflight"], "Aggregate result differs from raw outputs")
    require(record["status"] == "complete" and record["phase"] == "terminal" and record["new_external_model_calls"] == 0,
            "Nonterminal or externally executed admission screen")
    require(record["product_qualified"] is False and record["prospective_pages_used"] is False, "Unsupported qualification or reserve use")
    require(record["mlx_peak_bytes"] <= protocol["limits"]["mlx_memory_bytes"] and
            record["rss_peak_bytes"] <= protocol["limits"]["rss_peak_bytes"], "Resource budget exceeded")
    require(result["passed"] is False and record["development"]["requests"] == 0, "Failed preflight allowed development inference")
    setup = read(root / "setup-records.json")
    for item in setup["failed_setups"]:
        failure = item["record"]
        require(item["responses"] == 0 and failure["status"] == "failed" and failure["phase"] == "acquisition", "A model run was mislabeled as setup")
        require(sha((json.dumps(failure, ensure_ascii=False, indent=2) + "\n").encode()) == item["record_sha256"], "Changed original setup failure")
    return {"responses": len(rows), "screen_passed": result["passed"], "groups": result["groups"], "development_calls": 0,
            "product_qualified": False, "private_payloads_published": False}


if __name__ == "__main__":
    print(json.dumps(verify(), indent=2))
