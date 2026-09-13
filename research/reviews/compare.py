#!/usr/bin/env python3
"""Compare pinned review reports without confusing identity and finding changes.

This offline research helper uses only the Python standard library. It is not a
runtime, training, or normal product-test dependency.
"""

import argparse
from collections import Counter, defaultdict
import gc
import gzip
import hashlib
import json
from pathlib import Path


PAIRS = (
    ("responses", "all", "old-a1c6d7a"),
    ("ptah", "default", "old-d3c9f23"),
    ("ptah", "all", "old-d3c9f23"),
    ("ptah", "comments", "old-d3c9f23"),
    ("unswell", "all", "old-a1c6d7a"),
)


def require(condition, message):
    if not condition:
        raise ValueError(message)


def digest(path):
    with path.open("rb") as stream:
        return hashlib.file_digest(stream, "sha256").hexdigest()


def read(path):
    def invalid(value):
        raise ValueError(f"Nonfinite number in {path}: {value}")
    return json.loads(path.read_text(encoding="utf-8"), parse_constant=invalid)


def encode(value):
    return json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":"), allow_nan=False)


def write(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2, allow_nan=False) + "\n", encoding="utf-8")


def compact(finding):
    if finding is None:
        return None
    return {**{key: finding[key] for key in ("rule_id", "rule_version", "fingerprint", "primary")},
            "related": finding.get("related") or [],
            "content_sha256": hashlib.sha256(encode(content(finding)).encode()).hexdigest(),
            "mapping_sha256": hashlib.sha256(encode(mapping(finding)).encode()).hexdigest()}


def anchor(finding):
    primary = finding["primary"]
    return finding["rule_id"], primary["path"], primary["span"]["start"], primary["span"]["end"]


def content(finding):
    """Compare observable diagnostics, excluding identities and source mapping.

    Rule version and fingerprint are tracked separately. Local block/sentence IDs
    can change after extraction; they are not new evidence. Metrics, suggestions,
    occurrence count, activation, and policy remain material differences.
    """
    evidence = finding["evidence"]
    return {
        **{key: finding.get(key) for key in
           ("severity", "gate", "group", "scope", "message", "suppressed", "baseline_state", "derived")},
        "evidence": {key: value for key, value in evidence.items() if key != "occurrences"},
        "occurrences": len(evidence.get("occurrences") or []),
        "related": len(finding.get("related") or []),
    }


def mapping(finding):
    return {
        "primary": finding["primary"],
        "related": finding.get("related") or [],
        "occurrences": finding["evidence"].get("occurrences") or [],
    }


def align(before, after):
    """Match exact anchors, unique equal excerpts, then mutual unique overlaps.

    Overlapping many-to-many leftovers remain unresolved. No greedy matching,
    fuzzy text threshold, or fingerprint equality decides a diagnostic change.
    """
    old = {i: f for i, f in enumerate(before)}
    new = {i: f for i, f in enumerate(after)}
    result = []

    def pair(i, j, method):
        left, right = old.pop(i), new.pop(j)
        result.append({"status": "retained" if content(left) == content(right) else "changed",
                       "match": method, "mapping_changed": mapping(left) != mapping(right),
                       "identity_changed": any(left.get(k) != right.get(k) for k in ("id", "fingerprint", "rule_version")),
                       "before": left, "after": right})

    def unique(key, method):
        left, right = defaultdict(list), defaultdict(list)
        for i, f in old.items():
            left[key(f)].append(i)
        for j, f in new.items():
            right[key(f)].append(j)
        for value in sorted(left.keys() & right.keys()):
            if len(left[value]) == len(right[value]) == 1:
                pair(left[value][0], right[value][0], method)

    unique(anchor, "exact-span")
    unique(lambda f: (*anchor(f)[:2], f["primary"].get("snippet", "")), "unique-excerpt")

    groups = defaultdict(list)
    for j, finding in new.items():
        groups[anchor(finding)[:2]].append(j)
    forward, reverse = defaultdict(list), defaultdict(list)
    for i, finding in old.items():
        _, _, start, end = anchor(finding)
        for j in groups[anchor(finding)[:2]]:
            _, _, new_start, new_end = anchor(new[j])
            if max(start, new_start) < min(end, new_end):
                forward[i].append(j)
                reverse[j].append(i)
    for i, candidates in sorted(forward.items()):
        if len(candidates) == 1 and len(reverse[candidates[0]]) == 1:
            pair(i, candidates[0], "unique-overlap")
    for i, finding in old.items():
        result.append({"status": "unresolved-before" if forward[i] else "removed", "before": finding, "after": None})
    for j, finding in new.items():
        result.append({"status": "unresolved-after" if reverse[j] else "added", "before": None, "after": finding})
    result.sort(key=lambda row: (anchor(row["after"] or row["before"]), row["status"]))
    for index, row in enumerate(result):
        row["index"] = index
    require(sum(r["before"] is not None for r in result) == len(before), "Lost old findings")
    require(sum(r["after"] is not None for r in result) == len(after), "Lost new findings")
    return result


def load_report(path, manifest):
    report = read(path)
    require(report["schema_version"] == "1.0.0-alpha.1", "Unsupported report schema")
    expected = {s["path"]: s for s in manifest["sources"]}
    require(len(expected) == len(manifest["sources"]), "Duplicate source paths")
    sources, exclusions = {}, Counter()
    for document in report["documents"]:
        name, data = document["name"], document["source"].encode("utf-8")
        require(name not in sources, f"Duplicate document {name}")
        require(name in expected, f"Unexpected source {name}")
        require(hashlib.sha256(data).hexdigest() == document["source_hash"] == expected[name]["sha256"], f"Source hash mismatch: {name}")
        require(len(data) == document["bytes"] == expected[name]["bytes"], f"Source size mismatch: {name}")
        sources[name] = data
        exclusions.update(item["reason"] for item in document["excluded"])
    for finding in report["findings"]:
        for location in [finding["primary"], *(finding.get("related") or [])]:
            data = sources[location["path"]]
            start, end = location["span"]["start"], location["span"]["end"]
            require(0 <= start < end <= len(data), "Source range out of bounds")
        p = finding["primary"]
        data, span = sources[p["path"]], p["span"]
        # Context is a review aid, never used for matching. The complete source
        # range remains in the archived finding even when the excerpt is long.
        start = data.rfind(b"\n", 0, span["start"]) + 1
        stop = data.find(b"\n", span["end"])
        if stop < 0:
            stop = len(data)
        context = data[start:min(stop, start + 2400)].decode("utf-8", errors="replace")
        finding["review_context"] = context
    invocation = read(path.parent / "invocation.json")
    summary = {
        "status": report["status"], "exit_code": invocation["exit_code"],
        "documents": len(report["documents"]), "prose_words": sum(d["prose_words"] for d in report["documents"]),
        "findings": len(report["findings"]), "errors": report["errors"],
        "abstentions": report.get("abstentions", []), "exclusions_by_reason": dict(sorted(exclusions.items())),
        "manifest": report["manifest"], "gate": report["gate"],
        "binary_sha256": invocation["binary_sha256"],
        "policy_sha256": report["manifest"]["config_sources"][0]["sha256"],
        "report_sha256": digest(path), "workers": 2, "timeout": "20m",
        "max_candidates": 100000 if path.parent.parent.name == "default" else 2000000,
    }
    # Keep rule identity and policy, but omit long example catalogs from metadata.
    summary["manifest"] = {**report["manifest"], "rules": [
        {key: value for key, value in rule.items() if key in ("id", "version", "contexts", "requires", "defaults", "parameters", "status")}
        for rule in report["manifest"]["rules"]]}
    findings = report["findings"]
    del report, sources
    gc.collect()
    return findings, summary


def compare(root, output):
    output.mkdir(parents=True, exist_ok=True)
    runs = []
    review = []
    for subject, profile, previous in PAIRS:
        source_manifest = read(root / f"{subject}-sources.json")
        directory = root / subject if subject == "responses" else root / subject / profile
        before, old_summary = load_report(directory / previous / "result.json", source_manifest)
        after, new_summary = load_report(directory / "new/result.json", source_manifest)
        require(old_summary["policy_sha256"] == new_summary["policy_sha256"], "Policy bytes differ")
        historical = read(Path(__file__).parent / f"2026-09-11-{subject}/summary.json")["runs"][profile]
        for key in ("status", "documents", "findings"):
            require(old_summary[key] == historical[key], f"Historical {subject}/{profile}/{key} differs")
        if "prose_words" in historical:
            require(old_summary["prose_words"] == sum(historical["prose_words"].values()), "Historical word count differs")
        rows = align(before, after)
        name = f"{subject}-{profile}"
        # Stable gzip header makes the archive byte reproducible.
        payload = "".join(encode({**row, "before": compact(row["before"]), "after": compact(row["after"])}) + "\n"
                          for row in rows).encode("utf-8")
        archive = output / f"{name}.jsonl.gz"
        archive.write_bytes(gzip.compress(payload, mtime=0))
        counts = defaultdict(Counter)
        for row in rows:
            counts[(row["after"] or row["before"])["rule_id"]][row["status"]] += 1
        summary = {"subject": subject, "profile": profile, "before": old_summary, "after": new_summary,
                   "counts_by_rule": {key: dict(sorted(value.items())) for key, value in sorted(counts.items())},
                   "alignment": {"path": archive.name, "sha256": digest(archive), "rows": len(rows)}}
        summary["historical_counts_reproduced"] = True
        if profile == "all":
            sample = read(Path(__file__).parent / f"2026-09-11-{subject}/sample.json")
            selected = set()
            for original in sample["rows"]:
                candidates = [row for row in rows if row["before"]
                              and row["before"]["fingerprint"] == original["fingerprint"]
                              and row["before"]["primary"]["path"] == original["path"]
                              and ("line" not in original or row["before"]["primary"]["start"]["line"] == original["line"])]
                require(candidates, f"Original sample missing: {subject}/{original['index']}")
                for row in candidates:
                    selected.add(row["index"])
                    review.append({"run": name, "alignment_index": row["index"], "selection": "original-sample",
                                   "original_location_candidates": len(candidates), "original": original, **row})
            groups = defaultdict(list)
            for row in rows:
                if row["index"] not in selected and row["status"] in ("added", "changed", "unresolved-before", "unresolved-after"):
                    groups[((row["after"] or row["before"])["rule_id"], row["status"])].append(row)
            for group, candidates in sorted(groups.items()):
                # Predefined reproducible sample, independent of the review text.
                candidates.sort(key=lambda row: hashlib.sha256(("217-v1:" + name + ":" + encode(anchor(row["after"] or row["before"]))).encode()).hexdigest())
                for row in candidates[:5]:
                    review.append({"run": name, "alignment_index": row["index"], "selection": "delta-sample-sha256-217-v1", **row})
            summary["original_sample_matches"] = len(sample["rows"])
        runs.append(summary)
        print(name, dict(Counter(row["status"] for row in rows)), flush=True)
        del before, after, rows, payload
        gc.collect()
    write(output / "comparison.json", {"version": "unswell-review-comparison-v1", "runs": runs})
    write(output / "review-input.json", {"version": "unswell-review-input-v1", "rows": review})
    tables = ["# Replayed findings", "", "Generated by `../compare.py` from pinned reports. Changed means the diagnostic",
              "remains matched but its observable evidence or policy changed; it includes metric schema changes.",
              "Counts are not precision estimates. Unresolved entries are kept on their original side.", ""]
    for run in runs:
        tables += [f"## {run['subject']}: {run['profile']}", "",
                   "| Rule | Retained | Changed | Removed | Added | Unresolved old/new |",
                   "| --- | ---: | ---: | ---: | ---: | ---: |"]
        for rule, counts in run["counts_by_rule"].items():
            values = " | ".join(str(counts.get(key, 0)) for key in ("retained", "changed", "removed", "added"))
            tables.append(f"| `{rule}` | {values} | {counts.get('unresolved-before', 0)}/{counts.get('unresolved-after', 0)} |")
        tables.append("")
    (output / "tables.md").write_text("\n".join(tables), encoding="utf-8")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("artifacts", type=Path)
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    compare(args.artifacts, args.output)
