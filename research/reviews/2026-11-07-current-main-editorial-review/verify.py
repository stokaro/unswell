#!/usr/bin/env python3
"""Verify the exposed current-engine editorial review and cohort accounting."""

import argparse
from collections import Counter
import gzip
import hashlib
import json
from pathlib import Path
import sys


HERE = Path(__file__).resolve().parent
REPO = HERE.parents[2]
BASE = HERE.parent / "2026-11-07-current-main-baseline"
PRIOR = HERE.parent / "2026-09-26-outcome-announcements"
sys.path.insert(0, str(PRIOR))
import replay  # noqa: E402
import verify as prior_verify  # noqa: E402


def require(condition, message):
    if not condition:
        raise ValueError(message)


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path):
    with gzip.open(path, "rt", encoding="utf-8") if path.suffix == ".gz" else path.open(encoding="utf-8") as source:
        return json.load(source)


def sort_key(finding):
    primary = finding["primary"]
    span = primary["span"]
    return (primary["path"], span["start"], span["end"], finding["rule_id"])


def count_status(rows):
    return dict(sorted(Counter(row["status"] for row in rows).items()))


def evaluate(finding_rows, event_rows, reports, baseline, labels, sources):
    require(len(sources) == 36 and len(labels) == 123, "frozen source or label count changed")
    for event, label in labels.items():
        require(label["path"] in sources, "event outside frozen source: " + event)
        for target in label["targets"]:
            span = target
            text = sources[label["path"]][span["start"]:span["end"]].decode("utf-8")
            require(text == target["quote"], "target quote drift: " + event)

    expected_commit = baseline["base_commit"]
    for profile, report in reports.items():
        require(report["status"] == "complete" and report["manifest"]["complete"] and
                not report["errors"] and not report["manifest"]["skipped_rules"],
                "incomplete " + profile + " report")
        require(report["manifest"]["tool_commit"] == expected_commit,
                "wrong runtime commit: " + profile)
        require(len(report["documents"]) == 36, "document coverage changed: " + profile)
        require(len({finding["id"] for finding in report["findings"]}) ==
                len(report["findings"]), "duplicate finding ID: " + profile)

    technical = reports["technical"]["findings"]
    strict = reports["strict"]["findings"]
    technical_ids = {finding["id"] for finding in technical}
    strict_ids = {finding["id"] for finding in strict}
    require(technical_ids <= strict_ids, "technical finding absent from strict")
    ordered = sorted(strict, key=sort_key)
    require(len(finding_rows) == len(ordered) == 176, "incomplete finding review")
    reviewed = {}
    valid_status = {"actionable", "uncertain", "nonactionable"}
    for row, finding in zip(finding_rows, ordered):
        fid = finding["id"]
        require(row["id"] == fid and row["rule_id"] == finding["rule_id"] and
                row["path"] == finding["primary"]["path"] and
                row["span"] == finding["primary"]["span"],
                "finding identity drift: " + fid)
        profiles = ["technical", "strict"] if fid in technical_ids else ["strict"]
        require(row["profiles"] == profiles, "finding profile drift: " + fid)
        require(row["status"] in valid_status and row["rationale"].strip(),
                "missing finding decision: " + fid)
        require((isinstance(row["repair_action"], str) and row["repair_action"].strip())
                if row["status"] == "actionable" else row["repair_action"] is None,
                "repair proposal inconsistent: " + fid)
        has_advice = bool(finding["evidence"].get("suggestion"))
        expected_advice = ("not_supplied" if not has_advice else
                           "conditional_guidance" if row["status"] == "actionable" else
                           "not_applicable" if row["status"] == "nonactionable" else
                           "unresolved")
        require(row["advice_status"] == expected_advice,
                "advice disposition inconsistent: " + fid)
        reviewed[fid] = row

    baseline_rows = baseline["profiles"]["technical"]["events"]
    require(len(event_rows) == len(baseline_rows) == 123, "incomplete event review")
    event_status = {"full", "partial", "uncertain", "missed"}
    for row, reference in zip(event_rows, baseline_rows):
        event = reference["event"]
        label = labels[event]
        require(row["event"] == event and row["path"] == label["path"] and
                row["category"] == label["category"], "event identity drift: " + event)
        require(row["status"] in event_status and row["rationale"].strip(),
                "missing event decision: " + event)
        credited = row["credited_finding_ids"]
        require(len(credited) == len(set(credited)), "duplicate credited finding: " + event)
        if row["status"] == "full":
            require(credited and reference["possible_full_by_span"],
                    "unsupported full event: " + event)
            require(all(fid in technical_ids and reviewed[fid]["status"] == "actionable"
                        for fid in credited), "unaccepted full-event witness: " + event)
            for candidates in reference["overlap_finding_ids_by_target"]:
                require(set(credited) & set(candidates),
                        "uncovered full-event target: " + event)
        else:
            require(not credited, "credit on nonfull event: " + event)
            if not reference["targets_with_overlap"]:
                require(row["status"] == "missed", "no-span event credited: " + event)

    profiles = {}
    prior_review = read(PRIOR / "review.json")
    current_full = {row["event"] for row in event_rows if row["status"] == "full"}
    for profile, report in reports.items():
        rows = [reviewed[finding["id"]] for finding in report["findings"]]
        cohort_findings = {}
        cohort_events = {}
        for cohort in ("whole", "context", "confirmation"):
            selected = [row for row in rows if row["path"].startswith(cohort + "/")]
            event_selected = [row for row in event_rows if row["event"].startswith(cohort + "/")]
            reference = baseline["profiles"][profile]["by_cohort"][cohort]
            require(len(event_selected) == reference["events"],
                    "event denominator changed: " + cohort)
            full = sum(row["status"] == "full" for row in event_selected)
            accepted = sum(row["status"] == "actionable" for row in selected)
            cohort_findings[cohort] = {
                "total": len(selected), "accepted": accepted,
                "required_accepted_85pct": (85 * len(selected) + 99) // 100,
                "dispositions": count_status(selected),
            }
            cohort_events[cohort] = {
                "total": len(event_selected), "full": full,
                "required_full_80pct": reference["required_full"],
                "dispositions": count_status(event_selected),
            }
        old_full = {event for event, row in prior_review["profiles"][profile]["events"].items()
                    if row["after"]["status"] == "full"}
        require(old_full <= set(labels), "prior full event outside frozen labels")
        profiles[profile] = {
            "findings": len(rows), "finding_dispositions": count_status(rows),
            "advice_dispositions": dict(sorted(Counter(row["advice_status"] for row in rows).items())),
            "historical_event_comparison": {
                "prior_full": len(old_full), "current_full": len(current_full),
                "retained": sorted(current_full & old_full),
                "gained": sorted(current_full - old_full),
                "lost": sorted(old_full - current_full),
            },
            "by_cohort": {cohort: {"findings": cohort_findings[cohort],
                                   "events": cohort_events[cohort]}
                          for cohort in ("whole", "context", "confirmation")},
        }
    return {
        "format": "unswell-current-main-editorial-review-v1",
        "reviewer": "Codex assistant; one exposed source review under ADR 0041, not a human annotator",
        "runtime_commit": expected_commit,
        "review_base_commit": "b98cb51e4155213a208bdc8b1a044805c7c1875d",
        "operator_effect": {"admitted": False, "before_after_same_stream": True,
                            "gained_full_events": [], "lost_full_events": []},
        "profiles": profiles,
        "interpretation": "Accepted-delivery counts are this assistant's source-context judgments; uncertain and nonactionable findings remain in the denominator. Advice is guidance only, not a verified edit.",
    }


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--write", action="store_true")
    args = parser.parse_args()
    freeze = read(HERE / "freeze.json")
    for name, expected in freeze["inputs"].items():
        require(digest(REPO / name) == expected, "frozen input changed: " + name)
    require(digest(HERE / "verify.py") == freeze["verifier_sha256"],
            "verifier changed after freeze")
    baseline = read(BASE / "result.json")
    reports = {profile: read(BASE / (profile + ".json.gz"))
               for profile in ("technical", "strict")}
    summary = evaluate(read(HERE / "findings.json"), read(HERE / "events.json"),
                       reports, baseline, prior_verify.labels(), replay.sources())
    encoded = (json.dumps(summary, ensure_ascii=False, sort_keys=True, indent=2) + "\n").encode()
    if args.write:
        (HERE / "summary.json").write_bytes(encoded)
    else:
        require((HERE / "summary.json").read_bytes() == encoded,
                "recorded review summary differs")
    print(json.dumps({"summary_sha256": hashlib.sha256(encoded).hexdigest(),
                      "technical": summary["profiles"]["technical"]["by_cohort"],
                      "strict": summary["profiles"]["strict"]["by_cohort"]},
                     sort_keys=True))


if __name__ == "__main__":
    main()
