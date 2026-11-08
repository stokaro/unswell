#!/usr/bin/env python3
"""Exploratory leave-one-repository-out audit of coarse admission filters."""

import argparse
import collections
import hashlib
import json
import math
from pathlib import Path

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("development_root", type=Path)
parser.add_argument("output", type=Path)
args = parser.parse_args()
ROOT = args.development_root
BANK = ROOT / "recovered-admission-bank-v1"


def require(condition, message):
    if not condition:
        raise ValueError(message)


INPUTS = {
    "evaluation": (ROOT / "full-separated-judgment-source-necessity-astra-medium-v1-evaluation/evaluation.json", "820a25b21c5c3d13b9a3819f9560e2e4ab37ae519368131d294b6ea784fd7f51"),
    "metadata": (BANK / "alias-metadata.json", "c1df0e533b90d18458d3782b8464b0380297cf439b70a59fd3de52e8dcec6f06"),
    "cores": (BANK / "cores.json", "2eff16e368786905cf2d49be3eae7589d70597aa0f0627b329f43d02b8317927"),
    "reviews": (BANK / "reviews.json", "d57beb6ee3213ab16c286dd6a0a0bd7ed5adf24f1c67cd40f71d50458ab85ba2"),
    "originals": (ROOT / "joint-component-sol-high-evaluation/originals.json", "4f0051badc4b16dfe088d88cae69bd0dda89af2e9f62b4bed317309c4db65082"),
}
for name, (path, expected_hash) in INPUTS.items():
    require(hashlib.sha256(path.read_bytes()).hexdigest() == expected_hash,
            f"Input hash mismatch: {name}")
evaluation = json.loads(INPUTS["evaluation"][0].read_text())
metadata = json.loads(INPUTS["metadata"][0].read_text())
cores = json.loads(INPUTS["cores"][0].read_text())
reviews = json.loads(INPUTS["reviews"][0].read_text())
originals = json.loads(INPUTS["originals"][0].read_text())
selected = set(evaluation["all_selected_actual_aliases"])

page_group = {}
for group, numbers in originals["groups"].items():
    for number in numbers:
        page = f"page-{number:03d}"
        require(page not in page_group, f"Duplicate page group: {page}")
        page_group[page] = group
require(len(page_group) == 36, "Page group count changed")

cohorts = {}
groups = {}
accepted = {}
unsafe = {}
for alias in selected:
    m = metadata[alias]
    item = cores[m["stream"]][m["finding_id"]]
    review = reviews[m["stream"]]["judgments"][m["finding_id"]]
    cohorts[alias] = originals["paths"][m["page"]].split("/")[0]
    groups[alias] = (item.get("category", "?"), item.get("operator", "?"))
    accepted[alias] = review["status"] == "accepted"
    unsafe[alias] = review.get("suggestion_safety", {}).get("status") == "uncertain"

events = {c: {} for c in ("whole", "context", "confirmation")}
event_group = {}
for event_id, item in evaluation["complete_witnesses"].items():
    cohort = event_id.split("/")[0]
    events[cohort][event_id] = item
    path = originals["reference"]["events"][event_id]["path"]
    pages = [page for page, actual in originals["paths"].items() if actual == path]
    require(len(pages) == 1, f"Event page binding changed: {event_id}")
    event_group[event_id] = page_group[pages[0]]


def measures(kept, event_ids):
    result = {}
    for cohort in events:
        aliases = {a for a in kept if cohorts[a] == cohort}
        ids = [e for e in event_ids if e in events[cohort]]
        full = sum(
            any(set(witness) <= aliases for witness in events[cohort][e].get("full_groups", []))
            for e in ids
        )
        n = len(aliases)
        useful = sum(accepted[a] for a in aliases)
        result[cohort] = {
            "full": full,
            "events": len(ids),
            "accepted": useful,
            "delivered": n,
            "fraction": useful / n if n else 0.0,
            "unresolved_advice": sum(unsafe[a] for a in aliases),
        }
    return result


all_events = set(event_group)
baseline = measures(selected, all_events)
require(baseline == {
    "whole": {"full": 48, "events": 57, "accepted": 1627, "delivered": 2101, "fraction": 1627 / 2101, "unresolved_advice": 5},
    "context": {"full": 39, "events": 46, "accepted": 515, "delivered": 548, "fraction": 515 / 548, "unresolved_advice": 0},
    "confirmation": {"full": 17, "events": 20, "accepted": 190, "delivered": 198, "fraction": 190 / 198, "unresolved_advice": 0},
}, "Baseline does not match the archived result")

duplicate_groups = collections.defaultdict(list)
for alias in selected:
    m = metadata[alias]
    item = cores[m["stream"]][m["finding_id"]]
    spans = tuple((target["span"]["start"], target["span"]["end"]) for target in item.get("targets", []))
    duplicate_groups[(m["page"], spans, groups[alias])].append(alias)
dedup_priority = {stream: index for index, stream in enumerate(("original", "source_surface", "source_only", "public_contrast", "medium", "high"))}
deduplicated = {min(aliases, key=lambda alias: (dedup_priority[metadata[alias]["stream"]], alias)) for aliases in duplicate_groups.values()}
dedup_result = measures(deduplicated, all_events)

fixed_exclusions = [
    ("wordiness", "action_support"),
    ("unsupported_evaluation", "evaluation"),
    ("vague_claims", "evaluation"),
    ("vague_claims", "indirect_relation"),
]
retrospective = measures({alias for alias in selected if groups[alias] not in fixed_exclusions}, all_events)

holdout = set()
policies = {}
for excluded in sorted(set(page_group.values())):
    train = {a for a in selected if page_group[metadata[a]["page"]] != excluded}
    test = selected - train
    train_events = {e for e in all_events if event_group[e] != excluded}
    policy = []
    current = train.copy()
    for _ in range(8):
        base = measures(current, train_events)
        best = None
        for key in sorted(set(groups.values())):
            drop = {a for a in current if groups[a] == key}
            if sum(cohorts[a] == "whole" for a in drop) < 10:
                continue
            candidate = current - drop
            outcome = measures(candidate, train_events)
            if any(
                outcome[c]["full"] < math.ceil(0.8 * outcome[c]["events"])
                for c in outcome
            ):
                continue
            if any(outcome[c]["fraction"] < 0.85 for c in ("context", "confirmation")):
                continue
            gain = outcome["whole"]["fraction"] - base["whole"]["fraction"]
            if gain <= 0:
                continue
            if best is None or (gain, key) > (best[0], best[1]):
                best = (gain, key, candidate)
        if best is None:
            break
        current = best[2]
        policy.append(best[1])
        if measures(current, train_events)["whole"]["fraction"] >= 0.85:
            break
    policies[excluded] = policy
    holdout |= {a for a in test if groups[a] not in policy}

heldout = measures(holdout, all_events)
result = {
    "format": "unswell-349-operator-triage-audit-v1",
    "scope": "exposed_development_only",
    "input_sha256": {name: digest for name, (_, digest) in INPUTS.items()},
    "baseline": baseline,
    "exact_span_category_operator_deduplication": dedup_result,
    "retrospective_four_group_filter": {
        "exclusions": fixed_exclusions,
        "metrics": retrospective,
        "selection_used_all_exposed_grades": True,
    },
    "leave_one_repository_out": {
        "folds": len(policies),
        "heldout": heldout,
        "policy_frequency": [
            {"category": key[0], "operator": key[1], "folds": count}
            for key, count in sorted(collections.Counter(key for policy in policies.values() for key in policy).items())
        ],
        "per_group": {group: policy for group, policy in sorted(policies.items())},
    },
    "qualified": False,
    "reason": "Held-out whole-page recall and accepted fraction miss the unchanged targets; five supplied edits remain unresolved. The source bank is exposed and cached.",
}
args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
print("whole held-out:", heldout["whole"])
