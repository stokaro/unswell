"""Freeze source-bound #349 relations and mandatory unit verdict requests."""

import hashlib
import json
from pathlib import Path
import sys

ROOT = Path(__file__).resolve().parent
PACKET = ROOT / "packet"
PAGES = (
    ("fastfetch", "fastfetch-cli/fastfetch/README.md"),
    ("mise", "jdx/mise/README.md"),
    ("mkcert", "FiloSottile/mkcert/README.md"),
    ("zellij", "zellij-org/zellij/README.md"),
)
FUNCTIONS = {"definition", "capability", "evidence", "qualification",
             "instruction", "transition", "promotion", "other"}
RELATIONS = {"aligned", "narrower", "contradicted", "unestablished", "not_claim"}
CANDIDATES = {"narrower", "contradicted", "unestablished"}
VERDICT_FIELDS = {"ordinal", "decision", "start", "end", "quote", "diagnosis",
                  "evidence", "source_function", "deletion_test", "reason"}


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def encoded(value):
    return (json.dumps(value, ensure_ascii=False, sort_keys=True, indent=2) + "\n").encode("utf-8")


def page(slug):
    page_id = dict(PAGES)[slug]
    manifest = next(p for p in read(ROOT / "sources.json")["pages"] if p["id"] == page_id)
    record = next(p for p in read(ROOT / "inventory.json")["pages"] if p["id"] == page_id)
    raw = (ROOT / "sources" / slug / "README.md").read_bytes()
    blob = hashlib.sha1(b"blob " + str(len(raw)).encode("ascii") + b"\0" + raw).hexdigest()
    if sha(raw) != record["sha256"] or sha(raw) != manifest["sha256"] or \
            blob != manifest["git_blob"] or len(raw) != manifest["bytes"] or \
            manifest["repository"] + "/README.md" != page_id:
        raise ValueError(f"source identity changed: {page_id}")
    return page_id, raw, record


def numbered(raw):
    return "\n".join(f"{index:04d}| {line}" for index, line in
                     enumerate(raw.decode("utf-8").splitlines(), 1))


def chunks(slug):
    return read(PACKET / "protocol.json")["inventory_chunks"][slug]


def chunk_id(slug, bounds):
    return f"{slug}-{bounds[0]}-{bounds[1]}"


def source_payload(slug):
    page_id, raw, record = page(slug)
    return {"page_id": page_id, "source_sha256": sha(raw),
            "numbered_source": numbered(raw), "excluded": record["excluded"]}


def source_unit(unit):
    return {"ordinal": unit["ordinal"], "kind": unit["binding"]["kind"],
            "text": unit["text"], "raw_span": unit["raw_span"],
            "raw_slice": unit["raw_slice"], "segments": unit["binding"]["segments"]}


def inventory_request(slug, bounds):
    prompt = (PACKET / "inventory-prompt.txt").read_text(encoding="utf-8")
    _, _, record = page(slug)
    payload = source_payload(slug)
    payload["chunk_id"] = chunk_id(slug, bounds)
    payload["units"] = [source_unit(u) for u in record["units"][bounds[0]:bounds[1]]]
    return (prompt + "\nINPUT JSON:\n" + encoded(payload).decode("utf-8")).encode("utf-8")


def validate_sources_and_labels():
    manifest = read(ROOT / "sources.json")
    labels = read(ROOT / "labels.json")
    inventory = read(ROOT / "inventory.json")
    source_input = read(ROOT / "input.json")
    if manifest["format"] != "unswell-349-unit-relation-source-manifest-v1" or \
            labels["format"] != "unswell-349-unit-relation-source-only-review-v1" or \
            labels["source_manifest_sha256"] != sha((ROOT / "sources.json").read_bytes()) or \
            labels["inventory_sha256"] != sha((ROOT / "inventory.json").read_bytes()) or \
            len(labels["events"]) != 5 or len(manifest["pages"]) != len(PAGES) or \
            len(inventory["pages"]) != len(PAGES) or len(source_input["pages"]) != len(PAGES):
        raise ValueError("source-only pilot denominators or hashes changed")
    if [p["id"] for p in manifest["pages"]] != [pid for _, pid in PAGES] or \
            [p["id"] for p in inventory["pages"]] != [pid for _, pid in PAGES] or \
            [p["id"] for p in source_input["pages"]] != [pid for _, pid in PAGES]:
        raise ValueError("page order changed")
    if inventory["input_sha256"] != sha((ROOT / "input.json").read_bytes()):
        raise ValueError("source inventory input changed")
    for slug, page_id in PAGES:
        _, raw, record = page(slug)
        source = next(x for x in source_input["pages"] if x["id"] == page_id)
        if source["text"].encode("utf-8") != raw or source["sha256"] != sha(raw):
            raise ValueError(f"source input changed: {page_id}")
        for event in (x for x in labels["events"] if x["page"] == page_id):
            span = event["span"]
            if raw[span["start"]:span["end"]] != event["quote"].encode("utf-8") or \
                    raw[:span["start"]].count(b"\n") + 1 != event["line"]:
                raise ValueError(f"event source binding changed: {event['id']}")
            actual = []
            for unit in record["units"]:
                eligible = sum(max(0, min(span["end"], seg["end"]) - max(span["start"], seg["start"]))
                               for seg in unit["binding"]["segments"])
                if eligible:
                    actual.append({"ordinal": unit["ordinal"], "eligible_bytes": eligible})
            if actual != event["overlapping_units"] or \
                    sum(x["eligible_bytes"] for x in actual) != span["end"] - span["start"]:
                raise ValueError(f"event eligible binding changed: {event['id']}")
        for example in (x for x in labels["counterexamples"] if x["page"] == page_id):
            if raw.count(example["quote"].encode("utf-8")) != 1:
                raise ValueError("counterexample source binding changed")


def validate_inventory_chunk(slug, bounds, response):
    page_id, raw, record = page(slug)
    expected = list(range(bounds[0], bounds[1]))
    if type(response) is not dict or set(response) != {"page_id", "source_sha256", "chunk_id", "units"} or \
            response["page_id"] != page_id or response["source_sha256"] != sha(raw) or \
            response["chunk_id"] != chunk_id(slug, bounds) or \
            type(response["units"]) is not list or len(response["units"]) != len(expected):
        raise ValueError("inventory page or chunk identity changed")
    for ordinal, row in zip(expected, response["units"]):
        if type(row) is not dict or set(row) != {"ordinal", "function", "proposition", "support", "relation"} or \
                type(row["ordinal"]) is not int or row["ordinal"] != ordinal or \
                type(row["function"]) is not str or row["function"] not in FUNCTIONS or \
                type(row["relation"]) is not str or row["relation"] not in RELATIONS or \
                any(type(row[key]) is not str or len(row[key]) > 500
                    for key in ("proposition", "support")) or \
                not row["support"].strip():
            raise ValueError(f"invalid inventory unit {ordinal}")
    if bounds[1] > len(record["units"]):
        raise ValueError("inventory chunk exceeds source units")
    return response["units"]


def merge_inventory(slug, responses):
    page_id, raw, record = page(slug)
    if len(responses) != len(chunks(slug)):
        raise ValueError("missing inventory chunk")
    rows = []
    for bounds, response in zip(chunks(slug), responses):
        rows.extend(validate_inventory_chunk(slug, bounds, response))
    if len(rows) != len(record["units"]):
        raise ValueError("merged inventory omits source units")
    return {"page_id": page_id, "source_sha256": sha(raw), "units": rows}


def candidates(merged):
    selected = [row for row in merged["units"] if row["relation"] in CANDIDATES]
    limit = read(PACKET / "protocol.json")["max_candidates_per_page"]
    if len(selected) > limit:
        raise ValueError(f"candidate budget exceeded: {len(selected)} > {limit}")
    return selected


def verdict_request(slug, merged_raw):
    merged = json.loads(merged_raw)
    page_id, raw, record = page(slug)
    if merged != merge_inventory(slug, [
            {"page_id": page_id, "source_sha256": sha(raw),
             "chunk_id": chunk_id(slug, bounds), "units": merged["units"][bounds[0]:bounds[1]]}
            for bounds in chunks(slug)]):
        raise ValueError("invalid merged inventory")
    prompt = (PACKET / "verdict-prompt.txt").read_text(encoding="utf-8")
    payload = source_payload(slug)
    payload["inventory_sha256"] = sha(merged_raw)
    payload["candidates"] = [dict(row, source_unit=source_unit(record["units"][row["ordinal"]]))
                             for row in candidates(merged)]
    return (prompt + "\nINPUT JSON:\n" + encoded(payload).decode("utf-8")).encode("utf-8")


def validate_verdict(slug, response, merged_raw):
    page_id, raw, record = page(slug)
    merged = json.loads(merged_raw)
    expected = candidates(merged)
    if type(response) is not dict or set(response) != {"page_id", "source_sha256", "inventory_sha256", "verdicts"} or \
            response["page_id"] != page_id or response["source_sha256"] != sha(raw) or \
            response["inventory_sha256"] != sha(merged_raw) or \
            type(response["verdicts"]) is not list or len(response["verdicts"]) != len(expected):
        raise ValueError("verdict page identity or candidate count changed")
    bound = []
    for candidate, row in zip(expected, response["verdicts"]):
        if type(row) is not dict or set(row) != VERDICT_FIELDS or \
                type(row["ordinal"]) is not int or row["ordinal"] != candidate["ordinal"] or \
                type(row["decision"]) is not str or row["decision"] not in {"diagnose", "abstain"} or \
                type(row["start"]) is not int or type(row["end"]) is not int or \
                any(type(row[key]) is not str or len(row[key]) > 2000
                    for key in VERDICT_FIELDS - {"ordinal", "decision", "start", "end"}):
            raise ValueError("invalid verdict fields")
        if row["decision"] == "abstain":
            if row["start"] != 0 or row["end"] != 0 or \
                    any(row[key] for key in ("quote", "diagnosis", "evidence", "source_function", "deletion_test")) or \
                    not row["reason"].strip():
                raise ValueError("invalid abstention")
            continue
        if row["reason"] or any(not row[key].strip() for key in
                                 ("quote", "diagnosis", "evidence", "source_function", "deletion_test")) or \
                row["start"] < 0 or row["start"] >= row["end"] or row["end"] > len(raw) or \
                raw[row["start"]:row["end"]] != row["quote"].encode("utf-8"):
            raise ValueError("unbound diagnosis")
        unit = record["units"][row["ordinal"]]
        if not any(seg["start"] <= row["start"] and row["end"] <= seg["end"]
                   for seg in unit["binding"]["segments"]):
            raise ValueError("diagnosis touches protected or other source material")
        bound.append(dict(row, line=raw[:row["start"]].count(b"\n") + 1))
    return {"page_id": page_id, "source_sha256": sha(raw),
            "candidates": len(expected), "abstentions": len(expected) - len(bound), "findings": bound}


def prepare_requests():
    validate_sources_and_labels()
    (PACKET / "requests").mkdir(exist_ok=True)
    protocol = read(PACKET / "protocol.json")
    if protocol["format"] != "unswell-349-unit-relation-pilot-v1" or \
            protocol["model"] != "gpt-6-astra" or protocol["effort"] != "medium" or \
            protocol["calls_max"] != 9 or protocol["frozen_events"] != 5 or \
            protocol["pages"] != [page_id for _, page_id in PAGES] or \
            set(protocol["candidate_relations"]) != CANDIDATES or \
            len(protocol["candidate_relations"]) != len(CANDIDATES):
        raise ValueError("pilot protocol changed")
    for slug, _ in PAGES:
        _, _, record = page(slug)
        ranges = chunks(slug)
        if ranges[0][0] != 0 or ranges[-1][1] != len(record["units"]) or \
                any(a[1] != b[0] for a, b in zip(ranges, ranges[1:])) or \
                any(a >= b for a, b in ranges):
            raise ValueError("inventory chunks omit source units")
        for bounds in ranges:
            target = PACKET / "requests" / (chunk_id(slug, bounds) + ".txt")
            expected = inventory_request(slug, bounds)
            if target.exists() and target.read_bytes() != expected:
                raise ValueError(f"inventory request changed: {slug}")
            target.write_bytes(expected)


if __name__ == "__main__":
    if sys.argv[1:] != ["prepare"]:
        raise SystemExit("usage: packet.py prepare")
    prepare_requests()
    print("prepared", sum(len(chunks(slug)) for slug, _ in PAGES), "inventory requests; no model calls")
