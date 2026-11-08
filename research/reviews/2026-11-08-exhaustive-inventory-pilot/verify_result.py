"""Verify the frozen #349 inventory packet, complete run, review, and result."""

import hashlib
import json
from pathlib import Path
import sys

import packet
import run
import score

EXPECTED_SHA256 = {
    "packet/freeze.json": "7624796d13927bb06a197a206862a9661b4e3ed83b4f870715a19d58b71c7fb8",
    "run/run.json": "7be9bc00f1e7f07639f37ed5a24d8a5a5ca9719583546310310050ed392f0bd6",
    "review.json": "c524686433008a17a980d225300ce874ab9325412386e71dd346fab2c907ae39",
    "result.json": "d82a952854c1d3cfae36abe1b83755219cfdd675122549a6b4ea07f84e7d04e6",
    "prior-v1/freeze.json": "750474ef632d504485e921d0f35f7432db961fc01ed732146d1a2a82345d6956",
    "prior-v1/run.json": "84892e171cf6f80cf7a372aa30018ac4724b17319a90e839d58b7144728e04f8",
    "prior-v1/stderr.txt": "2d8b902fb51c4fb0108c58a40b037f3ed52641f00cdae67901d7cbda9b4b8473",
}


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def require(condition, reason):
    if not condition:
        raise ValueError(reason)


def verify(root=Path(__file__).resolve().parent):
    root = Path(root)
    old = (packet.ROOT, packet.PACKET, run.ROOT, run.PACKET, run.RUN, score.ROOT, score.RUN)
    try:
        packet.ROOT = root
        packet.PACKET = root / "packet"
        run.ROOT = root
        run.PACKET = root / "packet"
        run.RUN = root / "run"
        score.ROOT = root
        score.RUN = root / "run"
        return _verify(root)
    finally:
        (packet.ROOT, packet.PACKET, run.ROOT, run.PACKET, run.RUN, score.ROOT, score.RUN) = old


def _verify(root):
    for name, expected in EXPECTED_SHA256.items():
        require(sha((root / name).read_bytes()) == expected, "pinned artifact changed: " + name)
    protocol = run.verify()
    require(protocol["concurrency"] == 1 and protocol["tools_allowed"] is False and
            protocol["automatic_retries"] is False and protocol["attempts_per_request"] == 1 and
            protocol["calls_max"] == 6, "execution policy changed")
    prior = read(root / "prior-v1" / "run.json")
    require(prior["status"] == "stopped_at_first_error" and prior["complete"] is False and
            len(prior["calls"]) == 1 and prior["calls"][0]["status"] == "failed" and
            prior["freeze_sha256"] == sha((root / "prior-v1" / "freeze.json").read_bytes()) and
            (root / "prior-v1" / "events.jsonl").stat().st_size == 0 and
            "Operation not permitted" in (root / "prior-v1" / "stderr.txt").read_text(encoding="utf-8"),
            "first stopped attempt changed")
    manifest = read(root / "sources.json")
    require(manifest["format"] == "unswell-349-semantic-inventory-source-manifest-v1" and
            len(manifest["sources"]) == 3, "source manifest changed")
    packet.validate_sources_and_labels()
    journal = read(root / "run" / "run.json")
    calls = [(slug, stage) for slug, _ in packet.PAGES for stage in ("inventory", "diagnosis")]
    require(journal["status"] == "complete" and journal["complete"] is True and
            journal["freeze_sha256"] == sha((root / "packet" / "freeze.json").read_bytes()) and
            journal["model"] == protocol["model"] and journal["effort"] == protocol["effort"] and
            [(c["page"], c["stage"]) for c in journal["calls"]] == calls and
            all(c["status"] == "valid" for c in journal["calls"]),
            "delivered call sequence changed")
    for index, (slug, stage) in enumerate(calls):
        target = root / "run" / slug / stage
        row = journal["calls"][index]
        request_raw = (target / "request.txt").read_bytes()
        response_raw = (target / "response.json").read_bytes()
        response = json.loads(response_raw)
        require(row["request_sha256"] == sha(request_raw) and
                row["response_sha256"] == sha(response_raw), "call content hash changed")
        run.trace(target, response)
        if stage == "inventory":
            require(request_raw == packet.inventory_request(slug), "inventory request changed")
            packet.validate_inventory(slug, response)
        else:
            inventory_raw = (root / "run" / slug / "inventory" / "response.json").read_bytes()
            require(request_raw == packet.diagnosis_request(slug, inventory_raw),
                    "diagnosis request changed")
            expected_bound = packet.validate_diagnosis(slug, response, inventory_raw)
            require(read(target / "bound.json") == expected_bound, "source binding changed")
    review = read(root / "review.json")
    require(review["format"] == "unswell-349-exhaustive-inventory-complete-stream-review-v1" and
            "Codex assistant" in review["reviewer"] and
            len(review["findings"]) == 3 and
            review["protected_counterexamples_flagged"] is False and
            len(review["newly_discovered_unfrozen"]) == 1,
            "complete-stream review changed")
    result = read(root / "result.json")
    require(result == score.score() and result["events"] == 4 and result["full_recalled"] == 0 and
            result["partial"] == 2 and result["missed"] == 2 and
            result["delivered"] == 3 and result["accepted"] == 3 and
            result["advance"] is False and result["qualification"] is False and
            result["supplied_edits"] == 0,
            "result arithmetic or conclusion changed")
    return {"packet_sha256": sha((root / "packet" / "freeze.json").read_bytes()),
            "run_sha256": sha((root / "run" / "run.json").read_bytes()),
            "result_sha256": sha((root / "result.json").read_bytes()),
            "events": 4, "full_recalled": 0, "partial": 2,
            "delivered": 3, "accepted": 3, "advance": False}


if __name__ == "__main__":
    print(json.dumps(verify(Path(sys.argv[1]) if len(sys.argv) > 1 else
                            Path(__file__).resolve().parent), sort_keys=True))
