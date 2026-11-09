#!/usr/bin/env python3
"""Verify the source-disjoint shape screen and its explicit stop decision."""

import argparse
import collections
import hashlib
import json
import re
from pathlib import Path

import collect

HERE = Path(__file__).resolve().parent
ID = re.compile(r"page-(\d{3})/([a-z-]+)/([0-9]+)-([0-9]+)\Z")


def require(condition, message):
    if not condition:
        raise ValueError(message)


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(name):
    return json.loads((HERE / name).read_text())


def check_accounting(freeze, sources, summary, identities, review, result):
    require(freeze["status"] == "frozen_before_candidate_collection",
            "freeze status changed")
    require(sources["format"] == "unswell-349-shape-context-sources-v1" and
            len(sources["sources"]) == 12, "source inventory changed")
    source_by_number = {}
    for ordinal, item in enumerate(sources["sources"], 1):
        partition = "design" if ordinal <= 6 else "transfer"
        require(item["number"] == ordinal and item["partition"] == partition and
                item["source_file"] == f"page-{ordinal:03d}.source" and
                len(item["source_sha256"]) == 64 and item["source_bytes"] > 0,
                "source identity or partition changed")
        source_by_number[ordinal] = item
    require(summary["format"] == "unswell-349-shape-context-candidates-v1" and
            summary["source_count"] == 12 and len(summary["status"]) == 12,
            "candidate summary inventory changed")
    require(summary["main_commit"] == freeze["main_commit"] and
            summary["input_sha256"] == {key: freeze[key + "_sha256"] for key in
                                       ("protocol", "collector", "sources",
                                        "shapes", "exporter")},
            "candidate summary identity changed")
    for ordinal, item in enumerate(summary["status"], 1):
        require(item["page"] == ordinal and item["partition"] ==
                source_by_number[ordinal]["partition"] and
                item["source_bytes"] == source_by_number[ordinal]["source_bytes"],
                "page status identity changed")
        require(item["state"] in ("complete", "oversize", "timeout",
                                  "exporter_error", "invalid_exporter_json"),
                "unknown page status")
        if item["state"] == "complete":
            require(item["eligible_units"] >= 0 and
                    item["candidate_matches"] >= 0, "negative page count")
    require(identities["format"] ==
            "unswell-349-shape-context-candidate-identities-v1" and
            len(identities["design"]) == summary["design_matches"] and
            len(identities["transfer"]) == summary["transfer_matches"],
            "candidate identity count changed")
    known_shapes = set(collect.load_shapes(freeze["shapes_sha256"]))
    for partition in ("design", "transfer"):
        rows = identities[partition]
        require(len({x["id"] for x in rows}) == len(rows),
                "duplicate candidate identity")
        by_page = collections.Counter()
        for row in rows:
            match = ID.fullmatch(row["id"])
            require(match is not None, "candidate identity invalid")
            number, shape, start, end = int(match[1]), match[2], int(match[3]), int(match[4])
            require(number in source_by_number and
                    source_by_number[number]["partition"] == partition and
                    row["repository"] == source_by_number[number]["repository"] and
                    row["source_sha256"] == source_by_number[number]["source_sha256"] and
                    shape in known_shapes and
                    0 <= start < end <= source_by_number[number]["source_bytes"],
                    "candidate source binding invalid")
            by_page[number] += 1
        for item in summary["status"]:
            if item["partition"] == partition:
                require(by_page[item["page"]] == item.get("candidate_matches", 0),
                        "candidate per-page count changed")
    require(review["format"] == "unswell-349-shape-context-design-review-v1" and
            review["reviewer"].startswith("root Codex assistant") and
            review["supplied_edits"] == 0, "review basis changed")
    require([row["id"] for row in review["rows"]] ==
            [row["id"] for row in identities["design"]],
            "review omits or reorders a design candidate")
    dispositions = collections.Counter()
    warranted_repositories = set()
    for row, identity in zip(review["rows"], identities["design"], strict=True):
        require(row["disposition"] in ("warranted", "retain", "uncertain")
                and len(row["reason"].strip()) >= 20,
                "review disposition or reason invalid")
        dispositions[row["disposition"]] += 1
        if row["disposition"] == "warranted":
            warranted_repositories.add(identity["repository"])
    prereq = (dispositions["warranted"] >= 4 and
              len(warranted_repositories) >= 2 and dispositions["retain"] >= 4)
    require(not prereq, "design prerequisite passed; stop decision invalid")
    require(result["format"] == "unswell-349-shape-context-design-stop-v1" and
            result["status"] == "terminal_negative_design_screen" and
            result["main_commit"] == freeze["main_commit"] and
            result["source_pages"] == 12 and
            result["complete_pages"] ==
            sum(x["state"] == "complete" for x in summary["status"]) and
            result["unavailable_pages"] ==
            sum(x["state"] != "complete" for x in summary["status"]) and
            result["design_matches"] == len(identities["design"]) and
            result["transfer_matches_sealed_unreviewed"] ==
            len(identities["transfer"]) and
            result["design_dispositions"] == dict(dispositions) and
            result["warranted_repositories"] == len(warranted_repositories) and
            result["design_prerequisite_met"] is False and
            result["selector_defined"] is False and
            result["transfer_reviewed"] is False and
            result["model_calls"] == 0 and result["supplied_edits"] == 0 and
            result["editorial_full_event_credit"] == 0 and
            result["qualified"] is False,
            "result accounting or stop claim changed")
    return dict(dispositions)


def check_local_replay(freeze, sources, summary, identities, exporter, source_dir):
    require(sha(exporter) == freeze["exporter_sha256"],
            "exporter binary changed")
    shapes = collect.load_shapes(freeze["shapes_sha256"])
    rebuilt = {"design": [], "transfer": []}
    for item, expected in zip(sources["sources"], summary["status"], strict=True):
        raw = (source_dir / item["source_file"]).read_bytes()
        require(sha(source_dir / item["source_file"]) == item["source_sha256"]
                and len(raw) == item["source_bytes"] and
                hashlib.sha1(f"blob {len(raw)}\0".encode() + raw).hexdigest() ==
                item["git_blob_sha1"],
                "local source identity changed")
        if len(raw) > collect.MAX_BYTES:
            require(expected["state"] == "oversize", "page status changed")
            continue
        source = raw.decode("utf-8", "strict")
        page = item["source_file"]
        inventory, error = collect.export_source(exporter, source,
                                                  item["source_sha256"], page)
        if error:
            require(error == expected["state"], "exporter status changed")
            continue
        require(expected["state"] == "complete" and
                len(inventory["units"]) == expected["eligible_units"],
                "exported page accounting changed")
        matches = collect.extract_candidates(source, inventory["units"],
                                             inventory["excluded"], shapes)
        require(len(matches) == expected["candidate_matches"],
                "page candidate count changed")
        for match in matches:
            rebuilt[item["partition"]].append({
                "id": f"page-{item['number']:03d}/{match['shape']}/{match['start']}-{match['end']}",
                "repository": item["repository"],
                "source_sha256": item["source_sha256"],
            })
    require(rebuilt["design"] == identities["design"] and
            rebuilt["transfer"] == identities["transfer"],
            "local candidate replay differs from public identities")
    for partition in ("design", "transfer"):
        packet = collect.LOCAL / f"{partition}-candidates.json"
        require(sha(packet) == identities["input_sha256"][f"{partition}_packet"],
                "local candidate packet changed")


def verify(exporter=None, source_dir=None):
    freeze, sources, summary, identities, review, result = (
        read(name) for name in ("freeze.json", "sources.json",
                               "candidate-summary.json", "candidate-ids.json",
                               "design-review.json", "result.json")
    )
    for key in ("protocol", "acquire", "collector", "test", "sources", "shapes"):
        path = {"protocol": HERE / "PROTOCOL.md",
                "acquire": HERE / "acquire.py",
                "collector": HERE / "collect.py",
                "test": HERE / "test_collect.py",
                "sources": HERE / "sources.json",
                "shapes": collect.SHAPES}[key]
        require(sha(path) == freeze[key + "_sha256"], f"frozen {key} changed")
    require(identities["input_sha256"]["freeze"] == sha(HERE / "freeze.json") and
            identities["input_sha256"]["candidate_summary"] ==
            sha(HERE / "candidate-summary.json"),
            "candidate identity inputs changed")
    require(result["input_sha256"] == {name: sha(HERE / name) for name in
            ("PROTOCOL.md", "acquire.py", "collect.py", "test_collect.py",
             "sources.json", "freeze.json", "candidate-summary.json",
             "candidate-ids.json", "design-review.json")},
            "result input hash changed")
    counts = check_accounting(freeze, sources, summary, identities, review, result)
    if exporter is not None:
        require(source_dir is not None, "source directory required for replay")
        check_local_replay(freeze, sources, summary, identities, exporter,
                           source_dir)
    return counts


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--exporter", type=Path)
    parser.add_argument("--sources", type=Path)
    args = parser.parse_args()
    print(json.dumps({"verified": True, "design": verify(args.exporter,
                                                        args.sources)}, sort_keys=True))


if __name__ == "__main__":
    main()
