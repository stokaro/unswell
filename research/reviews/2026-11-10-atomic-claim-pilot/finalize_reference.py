#!/usr/bin/env python3
"""Bind source-only editorial judgments to the pinned original byte spans."""

import hashlib
import json
from collections import Counter
from pathlib import Path

from verify_sources import verify as verify_sources


DEMOTED = {
    (1, "I also want to emphasize that"):
        "The personal lead-in can be shortened, but this author's advice has an intentional first-person voice.",
    (3, "So if you want to be brave and use a pre-release packages, you can do so by adding"):
        "The informal prerelease aside may suit this project's voice; the feed instruction is still clear.",
    (8, "Ready for the next step?"):
        "The question can serve as a navigation cue before the configuration link.",
}


def bind(entry: dict, page: dict, raw: bytes, source_sha256: str) -> dict:
    quote = entry["quote"].encode("utf-8")
    if not quote or raw.count(quote) != 1:
        raise ValueError("reference quote is empty or not unique")
    start = raw.index(quote)
    end = start + len(quote)
    units = [unit["ordinal"] for unit in page["units"]
             if any(segment["start"] <= start and end <= segment["end"]
                    for segment in unit["binding"]["segments"])]
    if len(units) != 1:
        raise ValueError("reference quote is not in one eligible source unit")
    return {**entry, "byte_span": {"start": start, "end": end},
            "unit_ordinal": units[0], "source_sha256": source_sha256}


def build(packet: Path) -> dict:
    sources = verify_sources(packet)["sources"]
    input_raw = (packet / "input.json").read_bytes()
    exporter_input = json.loads(input_raw)
    inventory_raw = (packet / "inventory.json").read_bytes()
    inventory = json.loads(inventory_raw)
    draft_raw = (packet / "reference_draft.json").read_bytes()
    draft = json.loads(draft_raw)
    if (inventory["basis"] != "source-function-inventory" or
            inventory["unit_contract"] != "eligible-piece-v1" or
            inventory["input_sha256"] != hashlib.sha256(input_raw).hexdigest() or
            exporter_input["basis"] != "source-function-inventory" or
            len(exporter_input["pages"]) != 9 or len(inventory["pages"]) != 9):
        raise ValueError("wrong source inventory")
    if draft["format"] != "unswell-349-atomic-claim-source-review-draft-v1":
        raise ValueError("wrong source review draft")
    if draft["reviewer"] != "root Codex assistant":
        raise ValueError("source reviewer identity changed")
    pages = []
    for number, (source, page) in enumerate(zip(sources, inventory["pages"]), 1):
        raw = (packet / "sources" / source["source_file"]).read_bytes()
        input_page = exporter_input["pages"][number - 1]
        if (page["id"] != f"{source['repository']}/{source['path']}" or
                page["sha256"] != source["source_sha256"] or
                page["bytes"] != len(raw) or input_page["id"] != page["id"] or
                input_page["sha256"] != source["source_sha256"] or
                input_page["text"].encode("utf-8") != raw):
            raise ValueError(f"source inventory changed at page {number}")
        pages.append((page, raw, source["source_sha256"]))
    events = []
    controls = []
    for entry in draft["events"]:
        key = (entry["page"], entry["quote"])
        if key in DEMOTED:
            controls.append({"page": entry["page"], "quote": entry["quote"],
                             "disposition": "optional", "source_role": entry["source_role"],
                             "reason": DEMOTED[key]})
        else:
            events.append(entry)
    controls.extend(draft["controls"])
    if len(events) != 15 or len(controls) != 21:
        raise ValueError("source-only review count changed")
    bound_events = []
    bound_controls = []
    for index, entry in enumerate(events, 1):
        page, raw, source_sha256 = pages[entry["page"] - 1]
        bound_events.append({"id": f"E{index:03d}",
                             **bind(entry, page, raw, source_sha256)})
    for index, entry in enumerate(controls, 1):
        page, raw, source_sha256 = pages[entry["page"] - 1]
        bound_controls.append({"id": f"C{index:03d}",
                               **bind(entry, page, raw, source_sha256)})
    if len(set(x["page"] for x in bound_events)) < 4:
        raise ValueError("too few positive repositories")
    if set(x["page"] for x in bound_controls) != set(range(1, 10)):
        raise ValueError("controls do not cover every page")
    return {
        "format": "unswell-349-atomic-claim-source-reference-v1",
        "reviewed_on": "2026-11-10",
        "reviewer": "root Codex assistant",
        "role": "exposed_development_source_only_before_candidate",
        "source_inventory_sha256": hashlib.sha256(inventory_raw).hexdigest(),
        "draft_sha256": hashlib.sha256(draft_raw).hexdigest(),
        "demoted_before_freeze": len(DEMOTED),
        "counts": {"events": len(bound_events), "controls": len(bound_controls),
                   "event_pages": {str(page): count for page, count in
                                   sorted(Counter(x["page"] for x in bound_events).items())},
                   "control_pages": {str(page): count for page, count in
                                     sorted(Counter(x["page"] for x in bound_controls).items())}},
        "events": bound_events,
        "controls": bound_controls,
    }


if __name__ == "__main__":
    packet = Path(__file__).resolve().parent
    output = packet / "reference.json"
    if output.exists():
        raise SystemExit("source-only reference already exists")
    result = build(packet)
    output.write_text(json.dumps(result, indent=2, ensure_ascii=False,
                                 sort_keys=True) + "\n", encoding="utf-8")
    print(f"frozen {len(result['events'])} events and {len(result['controls'])} controls")
