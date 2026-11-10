"""Verify the frozen source and source-only review before any model call."""

import hashlib
import json
from pathlib import Path
import sys

ROOT = Path(__file__).resolve().parent
FILES = (
    ".gitattributes", "PROTOCOL.md", "README.md", "sources.json", "input.json", "inventory.json",
    "reference.json", "verify_source.py", "test_verify_source.py",
    "sources/nushell.source", "sources/hyperfine.source",
    "sources/eza.source", "sources/cargo-binstall.source",
)


def need(condition, message):
    if not condition:
        raise ValueError(message)


def sha(data):
    return hashlib.sha256(data).hexdigest()


def read(root, name):
    return json.loads((root / name).read_text(encoding="utf-8"))


def verify(root=ROOT):
    freeze = read(root, "freeze.json")
    need(set(freeze) == set(FILES), "frozen file set changed")
    for name in FILES:
        need(sha((root / name).read_bytes()) == freeze[name], "frozen file changed: " + name)

    sources = read(root, "sources.json")
    inp = read(root, "input.json")
    inv = read(root, "inventory.json")
    ref = read(root, "reference.json")
    source_rows = sources["pages"]
    need(sources["format"] == "unswell-349-additive-admission-sources-v1", "source format")
    need(len(source_rows) == 4 and len({p["id"] for p in source_rows}) == 4, "source count")
    need(inp["version"] == 1 and inp["basis"] == "source-function-inventory", "input contract")
    need(inv["version"] == 1 and inv["basis"] == inp["basis"], "inventory contract")
    need(inv["unit_contract"] == "eligible-piece-v1", "unit contract")
    need(sha((root / "input.json").read_bytes()) == inv["input_sha256"], "input binding")
    need([p["id"] for p in source_rows] == [p["id"] for p in inp["pages"]]
         == [p["id"] for p in inv["pages"]], "source order")
    need(ref["format"] == "unswell-349-additive-admission-reference-v1", "reference format")
    need(ref["reviewer"].startswith("Codex assistant,") and ref["independent_human"] is False,
         "reviewer identity")
    need(len(ref["events"]) == 4 and len(ref["retain_controls"]) == 4, "review counts")

    source_bytes = {}
    units_by_page = {}
    for source_row, input_row, inventory_row, reference_row in zip(
            source_rows, inp["pages"], inv["pages"], ref["pages"]):
        page = source_row["id"]
        data = (root / source_row["local_source"]).read_bytes()
        source_bytes[page] = data
        blob = hashlib.sha1(b"blob " + str(len(data)).encode() + b"\0" + data).hexdigest()
        need(blob == source_row["git_blob_sha1"], "Git blob: " + page)
        need(sha(data) == source_row["sha256"] == input_row["sha256"]
             == inventory_row["sha256"] == reference_row["source_sha256"],
             "source SHA-256: " + page)
        need(len(data) == source_row["bytes"] == inventory_row["bytes"]
             == reference_row["bytes"], "source bytes: " + page)
        need(data.decode("utf-8") == input_row["text"] and input_row["format"] == "markdown"
             and inventory_row["format"] == "markdown", "source text: " + page)
        units = inventory_row["units"]
        need(len(units) == reference_row["eligible_units"], "unit count: " + page)
        units_by_page[page] = {u["ordinal"]: u for u in units}
        need(len(units_by_page[page]) == len(units), "duplicate unit: " + page)
        for number, unit in enumerate(units):
            span = unit["raw_span"]
            need(unit["ordinal"] == number, "unit ordinal: " + page)
            need(0 <= span["start"] < span["end"] <= len(data), "raw span: " + page)
            need(data[span["start"]:span["end"]].decode("utf-8") == unit["raw_slice"],
                 "raw slice: " + page)
            segments = unit["binding"]["segments"]
            need(segments and all(span["start"] <= s["start"] < s["end"] <= span["end"]
                                  for s in segments), "eligible segments: " + page)

    ids = set()
    for item in ref["events"] + ref["retain_controls"]:
        page = item["page"]
        need(item["id"] not in ids, "duplicate review ID")
        ids.add(item["id"])
        need(page in source_bytes, "review page")
        data = source_bytes[page]
        quote = item["quote"].encode("utf-8")
        span = item["span"]
        need(data.count(quote) == 1 and data[span["start"]:span["end"]] == quote,
             "review quote: " + item["id"])
        unit = units_by_page[page].get(item["unit_ordinal"])
        need(unit is not None and any(s["start"] <= span["start"]
                                      and span["end"] <= s["end"]
                                      for s in unit["binding"]["segments"]),
             "review eligible binding: " + item["id"])
        need(item["rationale"].strip(), "review rationale: " + item["id"])
    return {"pages": len(source_rows), "eligible_units": sum(len(p["units"]) for p in inv["pages"]),
            "events": len(ref["events"]), "retain_controls": len(ref["retain_controls"])}


if __name__ == "__main__":
    print(json.dumps(verify(Path(sys.argv[1]) if len(sys.argv) == 2 else ROOT), sort_keys=True))
