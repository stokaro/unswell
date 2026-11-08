"""Apply a frozen source-only role rule to excluded archived proposal aliases."""

from __future__ import annotations

import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path

from selector import classify_spans


REPO = Path(__file__).resolve().parents[3]
POLICY = REPO / "research/reviews/2026-11-08-operator-triage/result.json"
POLICY_SHA256 = "22493865a6221276c08d945057829b9e455eccb50e0659d54f6df51c8d3232d8"
INPUTS = {
    "evaluation": (
        "full-separated-judgment-source-necessity-astra-medium-v1-evaluation/evaluation.json",
        "820a25b21c5c3d13b9a3819f9560e2e4ab37ae519368131d294b6ea784fd7f51",
    ),
    "metadata": (
        "recovered-admission-bank-v1/alias-metadata.json",
        "c1df0e533b90d18458d3782b8464b0380297cf439b70a59fd3de52e8dcec6f06",
    ),
    "cores": (
        "recovered-admission-bank-v1/cores.json",
        "2eff16e368786905cf2d49be3eae7589d70597aa0f0627b329f43d02b8317927",
    ),
    "originals": (
        "joint-component-sol-high-evaluation/originals.json",
        "4f0051badc4b16dfe088d88cae69bd0dda89af2e9f62b4bed317309c4db65082",
    ),
}


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def load_pinned(path: Path, digest: str) -> dict:
    raw = path.read_bytes()
    require(hashlib.sha256(raw).hexdigest() == digest, f"changed input: {path.name}")
    return json.loads(raw)


def page_groups(originals: dict) -> dict[str, str]:
    groups = {}
    for group, pages in originals["groups"].items():
        for page_number in pages:
            page = f"page-{page_number:03d}"
            require(page not in groups, f"duplicate group: {page}")
            groups[page] = group
    require(len(groups) == 36, "changed page inventory")
    return groups


def validated_source(originals: dict, page: str) -> bytes:
    document = originals["original"][page]["document"]
    source = document["source"].encode("utf-8")
    require(hashlib.sha256(source).hexdigest() == document["source_sha256"], f"source digest: {page}")
    return source


def target_spans(source: bytes, proposal: dict) -> list[tuple[int, int]]:
    spans = []
    for target in proposal["targets"]:
        start, end = target["span"]["start"], target["span"]["end"]
        require(0 <= start < end <= len(source), "target outside original source")
        require(source[start:end] == target["quote"].encode("utf-8"), "target quote mismatch")
        spans.append((start, end))
    require(spans, "proposal without targets")
    return spans


def select(development_root: Path) -> dict:
    inputs = {
        name: load_pinned(development_root / path, digest)
        for name, (path, digest) in INPUTS.items()
    }
    policy = load_pinned(POLICY, POLICY_SHA256)["leave_one_repository_out"]["per_group"]
    evaluation, metadata, cores, originals = (inputs[name] for name in ("evaluation", "metadata", "cores", "originals"))
    groups = page_groups(originals)
    original_aliases = set(evaluation["all_selected_actual_aliases"])
    require(len(original_aliases) == 2847, "changed original alias inventory")
    kept = set()
    added = set()
    decisions = {}
    sources = {}
    for alias in sorted(original_aliases):
        meta = metadata[alias]
        page = meta["page"]
        proposal = cores[meta["stream"]][meta["finding_id"]]
        key = [proposal.get("category", "?"), proposal.get("operator", "?")]
        if key not in policy[groups[page]]:
            kept.add(alias)
            continue
        if key != ["empty_framing", "authorial_narration"]:
            continue
        source = sources.setdefault(page, validated_source(originals, page))
        decision = classify_spans(source, target_spans(source, proposal))
        decisions[alias] = {"action": decision.action, "reason": decision.reason, "page": page}
        if decision.action == "admit":
            added.add(alias)
    require(len(kept) == 2311, "changed leave-one-repository-out baseline")
    return {
        "format": "unswell-349-source-role-selector-v1",
        "scope": "exposed_development_only",
        "input_sha256": {name: digest for name, (_, digest) in INPUTS.items()},
        "policy_sha256": POLICY_SHA256,
        "original_selected": len(original_aliases),
        "baseline_aliases": sorted(kept),
        "added_aliases": sorted(added),
        "selected_aliases": sorted(kept | added),
        "decisions": decisions,
        "decision_counts": dict(Counter(item["action"] for item in decisions.values())),
    }


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("development_root", type=Path)
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    result = select(args.development_root)
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps({
        "original_selected": result["original_selected"],
        "baseline_aliases": len(result["baseline_aliases"]),
        "added_aliases": len(result["added_aliases"]),
        "decision_counts": result["decision_counts"],
    }, sort_keys=True))


if __name__ == "__main__":
    main()
