#!/usr/bin/env python3
"""Recompute the exposed long-sentence precision tradeoff for issue #349."""

from __future__ import annotations

import argparse
import gzip
import hashlib
import json
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
COHORTS = ("whole", "context", "confirmation")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def read_inputs(freeze: dict, root: Path = ROOT) -> dict[str, object]:
    inputs = {}
    for name, expected in freeze["inputs"].items():
        path = root / name
        raw = path.read_bytes()
        require(hashlib.sha256(raw).hexdigest() == expected,
                f"Input changed: {name}")
        if name.endswith(".gz"):
            raw = gzip.decompress(raw)
        inputs[Path(name).name] = json.loads(raw)
    require(set(inputs) == {"events.json", "findings.json", "result.json",
                            "technical.json.gz"}, "Input file set changed")
    return inputs


def pairwise_auc(rows: list[dict]) -> float | None:
    accepted = [row["length"] for row in rows
                if row["status"] == "actionable"]
    others = [row["length"] for row in rows
              if row["status"] != "actionable"]
    if not accepted or not others:
        return None
    wins = sum((left > right) + 0.5 * (left == right)
               for left in accepted for right in others)
    return wins / (len(accepted) * len(others))


def extract_rows(inputs: dict[str, object]) -> tuple[dict[str, list[dict]], list[dict]]:
    decisions = inputs["findings.json"]
    events = inputs["events.json"]
    report = inputs["technical.json.gz"]
    require(len(decisions) == 176 and len(events) == 123,
            "Frozen review denominators changed")
    require(report["status"] == "complete" and not report["errors"] and
            report["gate"] == {"passed": True, "reasons": []},
            "Technical report incomplete")
    by_id = {row["id"]: row for row in decisions}
    require(len(by_id) == len(decisions), "Duplicate review decision")
    require(len({row["event"] for row in events}) == len(events),
            "Duplicate frozen event")
    groups = {cohort: [] for cohort in COHORTS}
    for finding in report["findings"]:
        if finding["rule_id"] != "syntax.long-sentence":
            continue
        decision = by_id.get(finding["id"])
        require(decision is not None and decision["rule_id"] == finding["rule_id"]
                and decision["path"] == finding["primary"]["path"]
                and decision["span"] == finding["primary"]["span"]
                and "technical" in decision["profiles"],
                "Long-sentence decision or source span changed")
        require(decision["status"] in
                ("actionable", "uncertain", "nonactionable"),
                "Unknown editorial decision")
        metrics = [item for item in finding["evidence"]["metrics"]
                   if item["name"] == "length" and
                   item["unit"] == "prose-words"]
        require(len(metrics) == 1 and
                isinstance(metrics[0]["value"], int) and
                metrics[0]["value"] > 35, "Missing rule length")
        cohort = decision["path"].split("/", 1)[0]
        require(cohort in groups, "Unknown cohort")
        groups[cohort].append({"id": finding["id"],
                               "length": metrics[0]["value"],
                               "status": decision["status"]})
    require({cohort: len(groups[cohort]) for cohort in COHORTS} ==
            {"whole": 84, "context": 9, "confirmation": 4},
            "Long-sentence set changed")
    return groups, events


def lost_full_events(events: list[dict], cohort: str,
                     removed_ids: set[str]) -> int:
    return sum(row["status"] == "full" and
               bool(row["credited_finding_ids"]) and
               set(row["credited_finding_ids"]) <= removed_ids
               for row in events if row["event"].startswith(cohort + "/"))


def analyze(inputs: dict[str, object]) -> dict:
    groups, events = extract_rows(inputs)
    baseline = inputs["result.json"]["profiles"]["technical"]["cohorts"]
    require({cohort: (baseline[cohort]["full"],
                      baseline[cohort]["accepted"],
                      baseline[cohort]["delivered"])
             for cohort in COHORTS} ==
            {"whole": (28, 82, 126), "context": (16, 21, 27),
             "confirmation": (4, 6, 9)},
            "Current technical baseline changed")
    output = {"format": "unswell-length-triage-audit-v1",
              "interpretation": "Post hoc exposed review; no policy is selected.",
              "cohorts": {}}
    for cohort in COHORTS:
        rows = groups[cohort]
        accepted = sum(row["status"] == "actionable" for row in rows)
        nonaccepted = len(rows) - accepted
        recorded = baseline[cohort]
        gap = recorded["required_full"] - recorded["full"]
        require(gap > 0, "Expected unresolved recall gap")
        thresholds = []
        for onset in range(35, max(row["length"] for row in rows) + 1):
            removed = [row for row in rows if row["length"] <= onset]
            lost_accepted = sum(row["status"] == "actionable"
                                for row in removed)
            lost_nonaccepted = len(removed) - lost_accepted
            lost_full = lost_full_events(events, cohort,
                                         {row["id"] for row in removed})
            perfect_new = gap + lost_full
            after_accepted = recorded["accepted"] - lost_accepted + perfect_new
            after_delivered = recorded["delivered"] - len(removed) + perfect_new
            thresholds.append({"onset": onset,
                               "removed_accepted": lost_accepted,
                               "removed_nonaccepted": lost_nonaccepted,
                               "lost_prior_full_events": lost_full,
                               "perfect_new_findings": perfect_new,
                               "optimistic_accepted_fraction":
                               after_accepted / after_delivered})
        output["cohorts"][cohort] = {
            "findings": len(rows),
            "accepted": accepted,
            "nonaccepted": nonaccepted,
            "length_auroc": pairwise_auc(rows),
            "full_now": recorded["full"],
            "full_required": recorded["required_full"],
            "delivered_now": recorded["delivered"],
            "accepted_now": recorded["accepted"],
            "thresholds": thresholds,
        }
    return output


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    freeze = json.loads((HERE / "freeze.json").read_text())
    require(freeze["format"] == "unswell-length-triage-inputs-v1",
            "Freeze format changed")
    result = json.dumps(analyze(read_inputs(freeze)), indent=2) + "\n"
    target = HERE / "result.json"
    if args.check:
        require(target.read_text() == result, "Audit result changed")
    else:
        target.write_text(result)
    print("Verified exposed length triage audit" if args.check else
          "Wrote exposed length triage audit")


if __name__ == "__main__":
    main()
