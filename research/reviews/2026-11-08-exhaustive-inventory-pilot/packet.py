"""Build and bind the frozen #349 inventory pilot without parsing Markdown."""

import hashlib
import json
from pathlib import Path
import sys

ROOT = Path(__file__).resolve().parent
PACKET = ROOT / "packet"
PAGES = (
    ("uv", "astral-sh/uv/README.md"),
    ("zed", "zed-industries/zed/README.md"),
    ("atuin", "atuinsh/atuin/README.md"),
)
FUNCTIONS = {"definition", "capability", "evidence", "qualification",
             "instruction", "transition", "promotion", "other"}
FINDING_FIELDS = {"unit_ordinal", "line", "quote", "diagnosis", "evidence",
                  "source_function", "deletion_test"}


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def encoded(value):
    return (json.dumps(value, ensure_ascii=False, sort_keys=True, indent=2) + "\n").encode("utf-8")


def page(slug):
    page_id = dict(PAGES)[slug]
    record = next(p for p in read(ROOT / "inventory.json")["pages"] if p["id"] == page_id)
    raw = (ROOT / "sources" / slug / "README.md").read_bytes()
    if sha(raw) != record["sha256"]:
        raise ValueError(f"source hash changed: {page_id}")
    return page_id, raw, record


def numbered(raw):
    source = raw.decode("utf-8")
    return "\n".join(f"{index:04d}| {line}" for index, line in enumerate(source.splitlines(), 1))


def source_payload(slug):
    page_id, raw, record = page(slug)
    return {"page_id": page_id, "source_sha256": sha(raw),
            "numbered_source": numbered(raw),
            "excluded": record["excluded"],
            "units": [{"ordinal": u["ordinal"], "kind": u["binding"]["kind"],
                       "text": u["text"], "raw_span": u["raw_span"],
                       "raw_slice": u["raw_slice"], "segments": u["binding"]["segments"]}
                      for u in record["units"]]}


def inventory_request(slug):
    prompt = (PACKET / "inventory-prompt.txt").read_text(encoding="utf-8")
    return (prompt + "\nINPUT JSON:\n" + encoded(source_payload(slug)).decode("utf-8")).encode("utf-8")


def diagnosis_request(slug, inventory_raw):
    inventory = json.loads(inventory_raw)
    validate_inventory(slug, inventory)
    prompt = (PACKET / "diagnosis-prompt.txt").read_text(encoding="utf-8")
    payload = source_payload(slug)
    payload["inventory_sha256"] = sha(inventory_raw)
    payload["inventory_records"] = inventory["units"]
    return (prompt + "\nINPUT JSON:\n" + encoded(payload).decode("utf-8")).encode("utf-8")


def validate_sources_and_labels():
    manifest = read(ROOT / "sources.json")
    labels = read(ROOT / "labels.json")
    inventory = read(ROOT / "inventory.json")
    if labels["source_manifest_sha256"] != sha((ROOT / "sources.json").read_bytes()) or \
            labels["inventory_sha256"] != sha((ROOT / "inventory.json").read_bytes()):
        raise ValueError("source-only labels changed")
    if len(labels["events"]) != 4 or len(manifest["sources"]) != len(PAGES) or \
            len(inventory["pages"]) != len(PAGES):
        raise ValueError("pilot denominators changed")
    for slug, page_id in PAGES:
        current_id, raw, record = page(slug)
        source = next(s for s in manifest["sources"] if s["repository"] + "/README.md" == page_id)
        blob = hashlib.sha1(b"blob " + str(len(raw)).encode("ascii") + b"\0" + raw).hexdigest()
        if current_id != page_id or source["source_sha256"] != sha(raw) or \
                source["git_blob_sha1"] != blob or source["bytes"] != len(raw) or \
                source["units"] != len(record["units"]):
            raise ValueError(f"source manifest changed: {page_id}")
        for event in [e for e in labels["events"] if e["page"] == page_id]:
            span = event["span"]
            if raw[span["start"]:span["end"]] != event["quote"].encode("utf-8"):
                raise ValueError(f"event source binding changed: {event['id']}")
            actual = [u["ordinal"] for u in record["units"] if any(
                segment["start"] < span["end"] and segment["end"] > span["start"]
                for segment in u["binding"]["segments"])]
            if actual != event["overlapping_unit_ordinals"] or not actual:
                raise ValueError(f"event eligible binding changed: {event['id']}")
        for counterexample in [c for c in labels["counterexamples"] if c["page"] == page_id]:
            if raw.count(counterexample["quote"].encode("utf-8")) != 1:
                raise ValueError("counterexample source binding changed")


def validate_inventory(slug, response):
    page_id, raw, record = page(slug)
    if type(response) is not dict or set(response) != {"page_id", "source_sha256", "units"} or \
            response["page_id"] != page_id or response["source_sha256"] != sha(raw):
        raise ValueError("inventory page identity changed")
    rows = response["units"]
    if type(rows) is not list or len(rows) != len(record["units"]):
        raise ValueError("inventory omitted or added unit IDs")
    for ordinal, row in enumerate(rows):
        if type(row) is not dict or set(row) != {"ordinal", "function", "proposition", "support"} or \
                type(row["ordinal"]) is not int or row["ordinal"] != ordinal or \
                type(row["function"]) is not str or row["function"] not in FUNCTIONS or \
                type(row["proposition"]) is not str or \
                type(row["support"]) is not str or not row["support"].strip() or \
                len(row["proposition"]) > 500 or len(row["support"]) > 500:
            raise ValueError(f"invalid inventory unit {ordinal}")
    return rows


def validate_diagnosis(slug, response, inventory_raw):
    page_id, raw, record = page(slug)
    if type(response) is not dict or set(response) != {"page_id", "source_sha256", "inventory_sha256", "findings"} or \
            response["page_id"] != page_id or response["source_sha256"] != sha(raw) or \
            response["inventory_sha256"] != sha(inventory_raw):
        raise ValueError("diagnosis page identity changed")
    rows = response["findings"]
    if type(rows) is not list or len(rows) > 20:
        raise ValueError("diagnosis finding limit or type")
    bound = []
    for row in rows:
        if type(row) is not dict or set(row) != FINDING_FIELDS or \
                type(row["unit_ordinal"]) is not int or type(row["line"]) is not int or row["line"] < 1 or \
                row["unit_ordinal"] < 0 or row["unit_ordinal"] >= len(record["units"]):
            raise ValueError("invalid diagnosis fields")
        if any(type(row[key]) is not str or not row[key].strip() or len(row[key]) > 2000
               for key in FINDING_FIELDS - {"unit_ordinal", "line"}):
            raise ValueError("empty or oversized diagnosis text")
        quote = row["quote"].encode("utf-8")
        matches = []
        pos = 0
        while True:
            pos = raw.find(quote, pos)
            if pos < 0:
                break
            if raw[:pos].count(b"\n") + 1 == row["line"]:
                matches.append(pos)
            pos += 1
        if len(matches) != 1:
            raise ValueError("diagnosis quote not uniquely bound to original line")
        start, end = matches[0], matches[0] + len(quote)
        unit = record["units"][row["unit_ordinal"]]
        if not any(s["start"] < end and s["end"] > start for s in unit["binding"]["segments"]):
            raise ValueError("diagnosis quote misses eligible unit")
        bound.append(dict(row, start_byte=start, end_byte=end))
    return {"page_id": page_id, "source_sha256": sha(raw), "findings": bound}


def prepare_requests():
    validate_sources_and_labels()
    protocol = read(PACKET / "protocol.json")
    if protocol["pages"] != [page_id for _, page_id in PAGES] or \
            protocol["model"] != "gpt-6-astra" or protocol["effort"] != "medium" or \
            protocol["calls_max"] != 6:
        raise ValueError("pilot protocol changed")
    for slug, _ in PAGES:
        target = PACKET / "requests" / f"{slug}-inventory.txt"
        expected = inventory_request(slug)
        if target.exists() and target.read_bytes() != expected:
            raise ValueError(f"inventory request changed: {slug}")
        target.write_bytes(expected)


if __name__ == "__main__":
    if sys.argv[1:] != ["prepare"]:
        raise SystemExit("usage: packet.py prepare")
    prepare_requests()
    print("prepared", len(PAGES), "inventory requests; no model calls")
