#!/usr/bin/env python3
"""Bind the frozen contextual packet's units to exact original UTF-8 slices."""

import argparse
import copy
import hashlib
import json
import tarfile
from pathlib import Path


ROOT = Path(__file__).resolve().parent
ARCHIVE = ROOT.parent / "2026-09-26-contextual-evidence" / "evidence.tar.gz"
ARCHIVE_SHA256 = "95e04b63062b0e06aed325126f2726b69e4d6816872fb1eb9b62018832c3502b"
MANIFEST = ROOT / "manifest.json"


def require(condition, message):
    if not condition:
        raise ValueError(message)


def sha256(data):
    return hashlib.sha256(data).hexdigest()


def encoded(value):
    return (json.dumps(value, ensure_ascii=False, sort_keys=True,
                       separators=(",", ":")) + "\n").encode("utf-8")


def manifest_bytes(value):
    return (json.dumps(value, ensure_ascii=False, sort_keys=True, indent=2) +
            "\n").encode("utf-8")


def read_member(archive, name):
    member = archive.getmember(name)
    require(member.isfile(), "Not a regular archive member: " + name)
    stream = archive.extractfile(member)
    require(stream is not None, "Cannot read archive member: " + name)
    return stream.read()


def augment(packet):
    result = copy.deepcopy(packet)
    source = result["source"].encode("utf-8")
    require(sha256(source) == result["source_sha256"], "Changed source hash")
    require(isinstance(result["units"], list) and result["units"], "Missing units")
    ids = set()
    for unit in result["units"]:
        require(set(unit) == {"id", "span"}, "Original unit fields changed")
        span = unit["span"]
        require(set(span) == {"start", "end"} and
                type(span["start"]) is int and type(span["end"]) is int,
                "Invalid unit span")
        start, end = span["start"], span["end"]
        require(0 <= start < end <= len(source), "Unit span outside source")
        require(isinstance(unit["id"], str) and unit["id"] not in ids,
                "Missing or duplicate unit ID")
        ids.add(unit["id"])
        raw_bytes = source[start:end]
        try:
            unit["raw"] = raw_bytes.decode("utf-8")
        except UnicodeDecodeError as error:
            raise ValueError("Unit span splits a UTF-8 character") from error
        unit["raw_sha256"] = sha256(raw_bytes)
    return result


def verify_augmented(packet, original):
    require(set(packet) == set(original), "Packet fields changed")
    for field in packet:
        if field != "units":
            require(packet[field] == original[field], "Original packet field changed: " + field)
    require(len(packet["units"]) == len(original["units"]), "Unit count changed")
    source = original["source"].encode("utf-8")
    for unit, prior in zip(packet["units"], original["units"]):
        require(set(unit) == {"id", "span", "raw", "raw_sha256"},
                "Augmented unit fields changed")
        require(unit["id"] == prior["id"] and unit["span"] == prior["span"],
                "Original unit identity changed")
        start, end = prior["span"]["start"], prior["span"]["end"]
        raw = source[start:end]
        require(unit["raw"].encode("utf-8") == raw and
                unit["raw_sha256"] == sha256(raw), "Raw unit slice changed")


def build(archive_path=ARCHIVE):
    require(sha256(archive_path.read_bytes()) == ARCHIVE_SHA256,
            "Frozen source archive changed")
    packets = {}
    pages = []
    with tarfile.open(archive_path, "r:gz") as archive:
        original_manifest = json.loads(read_member(archive, "packet/manifest.json"))
        original_freeze = json.loads(read_member(archive, "packet/freeze.json"))
        records = original_manifest["pages"]
        require(len(records) == 36, "Frozen page count changed")
        expected_names = {"packet/requests/" + row["page"] + ".json" for row in records}
        actual_names = {member.name for member in archive.getmembers()
                        if member.name.startswith("packet/requests/")}
        require(actual_names == expected_names and len(expected_names) == 36,
                "Frozen request file set changed")
        for row in records:
            name = "packet/requests/" + row["page"] + ".json"
            original_bytes = read_member(archive, name)
            original_sha = sha256(original_bytes)
            require(original_sha == row["request_sha256"] ==
                    original_freeze[name.removeprefix("packet/")],
                    "Original request hash changed: " + name)
            original = json.loads(original_bytes)
            require(original["page"] == row["page"] and
                    original["source_sha256"] == row["source_sha256"] and
                    len(original["source"].encode("utf-8")) == row["bytes"] and
                    len(original["units"]) == row["units"],
                    "Original request metadata changed: " + name)
            packet = augment(original)
            verify_augmented(packet, original)
            packed = encoded(packet)
            packets[row["page"] + ".json"] = packed
            pages.append({"page": row["page"], "reference": row["reference"],
                          "original_request_sha256": original_sha,
                          "source_sha256": row["source_sha256"],
                          "bytes": row["bytes"], "units": row["units"],
                          "packet_sha256": sha256(packed)})
    manifest = {"format": "unswell-contextual-unit-slices-v1",
                "archive_sha256": ARCHIVE_SHA256,
                "original_source_tree": original_manifest["source_tree"],
                "pages": pages}
    return manifest, packets


def verify_manifest(manifest, expected=MANIFEST):
    require(json.loads(expected.read_bytes()) == manifest,
            "Generated packet manifest changed")


def write_packets(packets, output):
    output.mkdir(parents=True, exist_ok=False)
    for name, data in packets.items():
        (output / name).write_bytes(data)
    require({p.name for p in output.iterdir()} == set(packets),
            "Generated request file set changed")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument("--check", action="store_true")
    group.add_argument("--write-manifest", action="store_true")
    group.add_argument("--build", type=Path, metavar="DIRECTORY")
    args = parser.parse_args()
    manifest, packets = build()
    if args.write_manifest:
        MANIFEST.write_bytes(manifest_bytes(manifest))
    else:
        verify_manifest(manifest)
        if args.build is not None:
            write_packets(packets, args.build)
    print(json.dumps({"pages": len(manifest["pages"]),
                      "units": sum(row["units"] for row in manifest["pages"]),
                      "archive_sha256": ARCHIVE_SHA256}, sort_keys=True))


if __name__ == "__main__":
    main()
