#!/usr/bin/env python3
"""Verify the public, source-free documentation acquisition packet offline."""

import json
from pathlib import Path

from acquire import digest


HERE = Path(__file__).resolve().parent


def checked_span(span, size):
    if not isinstance(span, dict):
        raise ValueError("missing source span")
    start, end = span["start"], span["end"]
    if not isinstance(start, int) or not isinstance(end, int) or not 0 <= start < end <= size:
        raise ValueError("source span outside frozen source")
    if end - start != len(span["quote"].encode("utf-8")):
        raise ValueError("source span differs from original UTF-8 quote")


def verify(root=HERE):
    def read(name):
        raw = (root / name).read_bytes()
        return raw, json.loads(raw)

    exclusions_raw, _ = read("exclusions.json")
    selection_raw, selection = read("selection.json")
    sources_raw, sources = read("development-sources.json")
    spec_raw, spec = read("review-spec.json")
    _, review = read("source-only-review.json")
    if selection["exclusions_sha256"] != digest(exclusions_raw):
        raise ValueError("exclusions differ from frozen selection")
    if selection["status"] != "complete" or len(selection["selected"]) != 32:
        raise ValueError("selection is incomplete")
    considered = {(x["topic"], x["rank"]): x for x in selection["considered"]}
    if len(considered) != len(selection["considered"]):
        raise ValueError("duplicate considered search rank")
    for topic, snapshot in selection["search_snapshots"].items():
        raw, response = read(f"search/{topic}.json")
        if digest(raw) != snapshot["sha256"] or response.get("incomplete_results"):
            raise ValueError(f"search snapshot changed: {topic}")
        if len(response["items"]) != snapshot["items"]:
            raise ValueError(f"search result count changed: {topic}")
        for (record_topic, rank), item in considered.items():
            if record_topic == topic and item["repo"] != response["items"][rank - 1]["full_name"]:
                raise ValueError("considered repository differs from search rank")
    for selected in selection["selected"]:
        if selected != considered.get((selected["topic"], selected["rank"])):
            raise ValueError("selected document differs from considered metadata")
    if sources["selection_sha256"] != digest(selection_raw):
        raise ValueError("development sources refer to another selection")
    selected_development = [x for x in selection["selected"] if x["selection_ordinal"] <= 2]
    if len(selected_development) != 16 or len(sources["entries"]) != 16:
        raise ValueError("development source count changed")
    by_name = {}
    for frozen, entry in zip(selected_development, sources["entries"]):
        name = f"{frozen['topic']}-{frozen['function']}-{frozen['selection_ordinal']:02d}.source"
        if (entry["local_file"] != name or entry["repo"] != frozen["repo"] or
                entry["path"] != frozen["path"] or
                entry["commit_sha"] != frozen["commit_sha"] or
                entry["blob_sha1"] != frozen["blob_sha1"] or
                entry["bytes"] != frozen["bytes"] or
                entry["topic"] != frozen["topic"] or
                entry["function"] != frozen["function"]):
            raise ValueError(f"development source differs from selection: {name}")
        by_name[name] = entry
    if (review["spec_sha256"] != digest(spec_raw) or
            review["development_sources_sha256"] != digest(sources_raw) or
            review["source_dispositions"] != spec["source_dispositions"] or
            set(review["source_dispositions"]) != set(by_name) or
            len(review["labels"]) != len(spec["labels"])):
        raise ValueError("source-only review provenance differs")
    for declared, label in zip(spec["labels"], review["labels"]):
        name = declared["source"]
        source = by_name[name]
        if (review["source_dispositions"][name] != "reviewed_english" or
                label["source"] != name or label["repo"] != source["repo"] or
                label["function"] != source["function"] or
                label["topic"] != source["topic"] or
                label["source_sha256"] != source["sha256"] or
                label["span"]["quote"] != declared["quote"]):
            raise ValueError("review label differs from frozen source")
        for key, value in declared.items():
            if key not in ("quote", "evidence_quote") and label.get(key) != value:
                raise ValueError(f"review field differs: {key}")
        checked_span(label["span"], source["bytes"])
        if declared.get("evidence_quote"):
            checked_span(label["evidence_span"], source["bytes"])
            if label["evidence_span"]["quote"] != declared["evidence_quote"]:
                raise ValueError("review evidence differs")
        elif label["evidence_span"] is not None:
            raise ValueError("unexpected evidence span")
    if len(review["considered_not_positive"]) != len(spec["considered_not_positive"]):
        raise ValueError("considered candidate count differs")
    for declared, item in zip(spec["considered_not_positive"], review["considered_not_positive"]):
        source = by_name[declared["source"]]
        if (item["id"] != declared["id"] or item["reason"] != declared["reason"] or
                item["source"] != declared["source"] or
                item["source_sha256"] != source["sha256"] or
                item["span"]["quote"] != declared["quote"]):
            raise ValueError("considered candidate differs from source-only review")
        checked_span(item["span"], source["bytes"])
    positives = [x for x in review["labels"] if x["kind"] == "positive"]
    controls = [x for x in review["labels"] if x["kind"] == "control"]
    counts = {"selected_development": 16,
              "reviewed_english": 15,
              "out_of_domain": 1,
              "positives": len(positives),
              "positive_repositories": len({x["repo"] for x in positives}),
              "controls": len(controls),
              "retain_controls": sum(x["disposition"] == "retain" for x in controls)}
    if review["counts"] != counts or counts != {"selected_development": 16,
                                                "reviewed_english": 15,
                                                "out_of_domain": 1,
                                                "positives": 8,
                                                "positive_repositories": 6,
                                                "controls": 32,
                                                "retain_controls": 32}:
        raise ValueError("review counts changed")
    if review["acquisition_gate_passed"]:
        raise ValueError("underpowered packet must not pass acquisition gate")
    return counts


if __name__ == "__main__":
    print(json.dumps(verify()))
