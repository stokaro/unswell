#!/usr/bin/env python3
"""Check the published, source-free acquisition packet offline."""

import json
from pathlib import Path

from acquire import digest


HERE = Path(__file__).resolve().parent


def checked_span(span, source_size):
    if not isinstance(span, dict):
        raise ValueError("missing span")
    start, end = span["start"], span["end"]
    if not isinstance(start, int) or not isinstance(end, int) or not 0 <= start < end <= source_size:
        raise ValueError("span outside frozen source")
    if end - start != len(span["quote"].encode("utf-8")):
        raise ValueError("span length differs from UTF-8 quote")


def verify(root=HERE):
    def read(name):
        raw = (root / name).read_bytes()
        return raw, json.loads(raw)

    exclusions_raw, _ = read("exclusions.json")
    selection_raw, selection = read("selection.json")
    sources_raw, sources = read("development-sources.json")
    spec_raw, spec = read("review-spec.json")
    _, review = read("source-only-review.json")
    if digest(exclusions_raw) != selection["exclusions_sha256"]:
        raise ValueError("exclusion file differs from frozen selection")
    if selection["status"] != "complete" or len(selection["selected"]) != 24:
        raise ValueError("selection is not complete")
    considered = {(item["stratum"], item["rank"]): item
                  for item in selection["considered"]}
    if len(considered) != len(selection["considered"]):
        raise ValueError("duplicate considered search rank")
    for name, snapshot in selection["search_snapshots"].items():
        raw, response = read(f"search/{name}.json")
        if digest(raw) != snapshot["sha256"] or response.get("incomplete_results"):
            raise ValueError(f"search snapshot changed: {name}")
        for (stratum, rank), item in considered.items():
            if stratum == name and item["repo"] != response["items"][rank - 1]["full_name"]:
                raise ValueError("considered repository differs from search rank")
    for frozen in selection["selected"]:
        if frozen != considered.get((frozen["stratum"], frozen["rank"])):
            raise ValueError("selected repository differs from considered metadata")
    if sources["selection_sha256"] != digest(selection_raw):
        raise ValueError("development sources refer to a different selection")
    selected = [x for x in selection["selected"] if x["selection_ordinal"] <= 4]
    if len(selected) != 12 or len(sources["entries"]) != 12:
        raise ValueError("development source accounting differs from 12 pages")
    source_by_name = {}
    for frozen, entry in zip(selected, sources["entries"]):
        name = f"{frozen['stratum']}-{frozen['selection_ordinal']:02d}.source"
        if (entry["local_file"] != name or entry["repo"] != frozen["repo"] or
                entry["commit_sha"] != frozen["commit_sha"] or
                entry["readme_blob_sha1"] != frozen["readme_blob_sha1"] or
                entry["bytes"] != frozen["readme_bytes"]):
            raise ValueError(f"development source metadata differs from selection: {name}")
        source_by_name[name] = entry
    if (review["spec_sha256"] != digest(spec_raw) or
            review["source_manifest_sha256"] != digest(sources_raw) or
            sorted(review["reviewed_pages"]) != sorted(source_by_name) or
            review["reviewed_pages"] != spec["reviewed_pages"] or
            len(review["labels"]) != len(spec["labels"])):
        raise ValueError("review provenance or page accounting mismatch")
    ids = set()
    positives = set()
    for declared, label in zip(spec["labels"], review["labels"]):
        name = declared["source"]
        source = source_by_name[name]
        if (label["id"] != declared["id"] or label["kind"] != declared["kind"] or
                label["source"] != name or label["repo"] != source["repo"] or
                label["source_sha256"] != source["sha256"] or
                label["span"]["quote"] != declared["quote"]):
            raise ValueError("review label differs from source-only specification")
        for key, value in declared.items():
            if key not in ("quote", "evidence_quote") and label.get(key) != value:
                raise ValueError(f"review field differs from specification: {key}")
        if label["id"] in ids:
            raise ValueError("duplicate review label ID")
        ids.add(label["id"])
        checked_span(label["span"], source["bytes"])
        if declared.get("evidence_quote"):
            checked_span(label["evidence_span"], source["bytes"])
            if label["evidence_span"]["quote"] != declared["evidence_quote"]:
                raise ValueError("review evidence differs from specification")
        elif label["evidence_span"] is not None:
            raise ValueError("unexpected evidence span")
        if label["kind"] == "positive":
            positives.add(label["id"])
    for label in review["labels"]:
        if label["kind"] == "control" and label["paired_positive"] not in positives:
            raise ValueError("control lacks a positive pair")
    result = {"pages": len(source_by_name), "positives": len(positives),
              "controls": len(review["labels"]) - len(positives)}
    if result != {"pages": 12, "positives": 9, "controls": 13}:
        raise ValueError("review counts changed")
    return result


if __name__ == "__main__":
    print(json.dumps(verify()))
