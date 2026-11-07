#!/usr/bin/env python3
"""Verify the pinned external hope-framing screen without fetching sources."""

import argparse
import collections
import hashlib
import json
from pathlib import Path

HERE = Path(__file__).resolve().parent


def digest(data):
    return hashlib.sha256(data).hexdigest()


def read_json(path):
    return json.loads(path.read_text(encoding="utf-8"))


def source_bytes(selection, root):
    found = {}
    for row in selection["sources"]:
        path = root / f"{row['id']}.md"
        data = path.read_bytes()
        assert len(data) == row["bytes"], path
        assert digest(data) == row["source_sha256"], path
        blob = b"blob " + str(len(data)).encode("ascii") + b"\0" + data
        assert hashlib.sha1(blob).hexdigest() == row["blob_sha1"], path
        data.decode("utf-8")
        found[f"{row['id']}.md"] = data
    assert len(found) == 9
    return found


def verify_reference(reference, sources):
    counts = collections.Counter()
    for row in reference["decisions"]:
        data = sources[f"{row['source_id']}.md"]
        assert 0 <= row["start"] < row["end"] <= len(data), row["event_id"]
        target = data[row["start"] : row["end"]]
        assert digest(target) == row["target_sha256"], row["event_id"]
        assert len(target) == row["target_bytes"], row["event_id"]
        assert digest(data) == row["source_sha256"], row["event_id"]
        counts[row["decision"]] += 1
    assert len(reference["decisions"]) == 13
    assert counts == {"acceptable_unchanged": 11, "uncertain": 2}
    return counts


def normalized(finding):
    return json.dumps(
        {key: value for key, value in finding.items()
         if key not in {"id", "rule_version", "fingerprint"}},
        sort_keys=True,
    )


def verify_report(path, expected_hash, sources):
    raw = path.read_bytes()
    assert digest(raw) == expected_hash, path
    report = json.loads(raw)
    assert report["status"] == "complete" and not report["errors"], path
    names = set()
    for doc in report["documents"]:
        name = doc["name"]
        assert name in sources and name not in names, name
        data = sources[name]
        assert doc["source"].encode("utf-8") == data, name
        assert doc["source_hash"] == digest(data), name
        assert doc["bytes"] == len(data), name
        names.add(name)
    assert names == set(sources), path
    for finding in report["findings"]:
        primary = finding["primary"]
        data = sources[primary["path"]]
        span = primary["span"]
        assert 0 <= span["start"] < span["end"] <= len(data), finding["id"]
    return report


def compare(main, candidate, reference):
    old = collections.Counter(normalized(f) for f in main["findings"])
    new = collections.Counter(normalized(f) for f in candidate["findings"])
    removed = old - new
    added = new - old
    assert not removed, "previous finding content was lost or changed"
    assert sum(added.values()) == 1, "expected one added finding"
    extra = json.loads(next(iter(added)))
    assert extra["rule_id"] == "filler.evaluative-closure"
    assert extra["primary"]["path"] == "q2-r6.md"
    span = extra["primary"]["span"]
    expected = next(row for row in reference["decisions"] if row["event_id"] == "h06")
    assert expected["decision"] == "acceptable_unchanged"
    assert expected["start"] <= span["start"] < span["end"] <= expected["end"]
    assert extra["message"] == "The hoped-if-would construction buries the design goal."
    assert extra["evidence"]["suggestion"] == (
        "State the intended behavior directly; preserve the condition, actor and modification limit."
    )
    return {
        "main_findings": len(main["findings"]),
        "candidate_findings": len(candidate["findings"]),
        "added": 1,
        "removed_or_changed": 0,
        "new_finding": {
            "source": "q2-r6.md",
            "start": span["start"],
            "end": span["end"],
            "review_decision": expected["decision"],
            "advice_review": "inapplicable: the council request has no modification limit",
        },
    }


def verify(source_root, report_root):
    selection = read_json(HERE / "prospective-selection.json")
    reference = read_json(HERE / "prospective-reference.json")
    runs = read_json(HERE / "prospective-run.json")
    for key, filename in (
        ("protocol_sha256", "PROSPECTIVE.md"),
        ("search_sha256", "prospective-search.json"),
        ("selection_sha256", "prospective-selection.json"),
        ("reference_sha256", "prospective-reference.json"),
    ):
        assert digest((HERE / filename).read_bytes()) == runs[key], filename
    assert selection["candidate_commit"] == reference["candidate_commit"] == runs["candidate_commit"]
    sources = source_bytes(selection, source_root)
    counts = verify_reference(reference, sources)
    reports = {}
    for run in runs["runs"]:
        assert run["exit_code"] == 0
        key = (run["arm"], run["profile"])
        assert key not in reports
        path = report_root / f"{run['arm']}-{run['profile']}.json"
        reports[key] = verify_report(path, run["report_sha256"], sources)
        assert reports[key]["manifest"]["scoring_profile"] == f"{run['profile']}-v1"
    assert set(reports) == {
        ("main", "technical"), ("candidate", "technical"),
        ("main", "strict"), ("candidate", "strict"),
    }
    profiles = {
        profile: compare(reports["main", profile], reports["candidate", profile], reference)
        for profile in ("technical", "strict")
    }
    assert profiles["technical"]["new_finding"] == profiles["strict"]["new_finding"]
    return {
        "format": "unswell-hope-prospective-result-v1",
        "sources": len(sources),
        "reference_events": dict(counts),
        "profiles": profiles,
        "operator_transfer": "failed: one new warning targets an acceptable unchanged passage",
        "recall_evidence": "insufficient: no definite actionable event in this trigger-enriched set",
        "product_decision": "reject default promotion of candidate commit 35e3350",
        "scope": "This is a rule-specific source-local counterexample, not #349 cohort qualification.",
    }


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--sources", type=Path, required=True)
    parser.add_argument("--reports", type=Path, required=True)
    parser.add_argument("--update", action="store_true")
    args = parser.parse_args()
    result = verify(args.sources, args.reports)
    output = HERE / "prospective-result.json"
    if args.update:
        output.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    else:
        assert read_json(output) == result, "published result differs from verified artifacts"
    print(json.dumps({
        "sources": result["sources"],
        "reference_events": result["reference_events"],
        "technical": result["profiles"]["technical"],
        "strict": result["profiles"]["strict"],
        "product_decision": result["product_decision"],
    }))


if __name__ == "__main__":
    main()
