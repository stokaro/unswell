#!/usr/bin/env python3
"""Build a source-only six-pair packet and a separate sealed-grade key."""

from __future__ import annotations

import argparse
import hashlib
import importlib.util
import json
import re
from pathlib import Path


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
FINDINGS = ROOT / "research/reviews/2026-11-07-current-main-editorial-review/findings.json"
EVENTS = ROOT / "research/reviews/2026-11-07-current-main-editorial-review/events.json"
REPLAY = ROOT / "research/reviews/2026-09-26-outcome-announcements/replay.py"
PAGES = (
    "whole/p026.md", "whole/p017.md", "whole/p001.md", "whole/p002.md",
    "context/c04.md", "confirmation/c01.mdx",
)
WITNESS = "5f3c42940d99ad29"
WORD = re.compile(r"[A-Za-z]+(?:'[A-Za-z]+)?")


def require(ok: bool, message: str) -> None:
    if not ok:
        raise ValueError(message)


def sha_bytes(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def sha(path: Path) -> str:
    return sha_bytes(path.read_bytes())


def encoded(value: object) -> bytes:
    return (json.dumps(value, ensure_ascii=False, indent=2, sort_keys=True) + "\n").encode()


def score(identifier: str) -> str:
    return sha_bytes(("long-sentence-dependency-v1:" + identifier).encode())


def source_module():
    spec = importlib.util.spec_from_file_location("unswell_frozen_sources", REPLAY)
    require(spec is not None and spec.loader is not None, "Source loader missing")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def source_context(raw: bytes, start: int, end: int) -> str:
    left = raw.rfind(b"\n", 0, max(0, start - 350))
    right = raw.find(b"\n", min(len(raw), end + 350))
    if left < 0:
        left = 0
    else:
        left += 1
    if right < 0:
        right = len(raw)
    return raw[left:right].decode("utf-8")


def select(findings: list[dict], sources: dict[str, bytes], events: list[dict]):
    eligible = [row for row in findings if row["rule_id"] == "syntax.long-sentence"]
    require(len(eligible) == 97, "Long-sentence inventory changed")
    require({status: sum(row["status"] == status for row in eligible)
             for status in ("actionable", "uncertain", "nonactionable")} ==
            {"actionable": 52, "uncertain": 44, "nonactionable": 1},
            "Long-sentence grades changed")
    witness_events = [row for row in events if row["event"] == "whole/p026-d02"]
    require(len(witness_events) == 1 and witness_events[0]["status"] == "full"
            and witness_events[0]["credited_finding_ids"] == [WITNESS],
            "Sole FULL witness changed")
    cases = []
    key = []
    for index, page in enumerate(PAGES, 1):
        raw = sources[page]
        accepted = [row for row in eligible if row["path"] == page
                    and row["status"] == "actionable"]
        uncertain = [row for row in eligible if row["path"] == page
                     and row["status"] == "uncertain"]
        require(accepted and uncertain, f"Missing pair members on {page}")
        chosen = (next((row for row in accepted if row["id"] == WITNESS), None)
                  if page == PAGES[0] else min(accepted, key=lambda row: score(row["id"])))
        require(chosen is not None, "Sole witness missing")

        def words(row: dict) -> int:
            span = row["span"]
            return len(WORD.findall(raw[span["start"]:span["end"]].decode("utf-8")))

        partner = min(uncertain, key=lambda row: (abs(words(row) - words(chosen)),
                                                   score(row["id"])))
        pair = [chosen, partner]
        if int(score(chosen["id"])[-1], 16) & 1:
            pair.reverse()
        members = []
        secret = []
        for letter, row in zip(("A", "B"), pair):
            span = row["span"]
            require(0 <= span["start"] < span["end"] <= len(raw),
                    "Finding span outside original source")
            quote = raw[span["start"]:span["end"]].decode("utf-8")
            members.append({"id": letter, "span": span, "quote": quote,
                            "surrounding_source": source_context(raw, span["start"], span["end"])})
            secret.append({"id": letter, "finding_id": row["id"],
                           "old_grade": row["status"],
                           "full_event": "whole/p026-d02" if row["id"] == WITNESS else None})
        cases.append({"id": f"pair-{index:02d}", "source_file": f"sources/page-{index:02d}.source",
                      "source_sha256": sha_bytes(raw), "members": members})
        key.append({"id": f"pair-{index:02d}", "members": secret})
    return cases, key


def build(output: Path) -> dict:
    freeze = json.loads((HERE / "freeze.json").read_text())
    require(freeze["format"] == "unswell-long-sentence-dependency-freeze-v1",
            "Freeze format changed")
    paths = {"protocol": HERE / "PROTOCOL.md", "preparer": HERE / "prepare.py",
             "findings": FINDINGS, "events": EVENTS, "source_loader": REPLAY,
             "whole_archive": ROOT / "research/reviews/2026-09-17-full-page-recall/inputs.tar.gz",
             "whole_manifest": ROOT / "research/reviews/2026-09-17-full-page-recall/manifest.json",
             "context_archive": ROOT / "research/reviews/2026-09-17-proposition-repetition/confirmation/inputs.tar.gz",
             "context_manifest": ROOT / "research/reviews/2026-09-17-proposition-repetition/confirmation/manifest.json",
             "confirmation_inputs": ROOT / "research/reviews/2026-09-26-outcome-announcements/confirmation-inputs.json"}
    for name, path in paths.items():
        require(sha(path) == freeze["sha256"][name], f"{name} changed")
    sources = source_module().sources()
    require(len(sources) == 36, "Frozen source inventory changed")
    cases, key = select(json.loads(FINDINGS.read_text()), sources,
                        json.loads(EVENTS.read_text()))
    require(not output.exists(), "Output directory already exists")
    (output / "sources").mkdir(parents=True)
    for case, page in zip(cases, PAGES):
        (output / case["source_file"]).write_bytes(sources[page])
    packet = {"format": "unswell-long-sentence-dependency-packet-v1",
              "freeze_sha256": sha(HERE / "freeze.json"), "cases": cases}
    key_record = {"format": "unswell-long-sentence-dependency-key-v1",
                  "packet_sha256": sha_bytes(encoded(packet)), "cases": key}
    (output / "packet.json").write_bytes(encoded(packet))
    (output / "key.json").write_bytes(encoded(key_record))
    manifest = {"format": "unswell-long-sentence-dependency-manifest-v1",
                "packet_sha256": sha(output / "packet.json"),
                "key_sha256": sha(output / "key.json"),
                "source_sha256": {case["source_file"]: case["source_sha256"] for case in cases}}
    (output / "manifest.json").write_bytes(encoded(manifest))
    return manifest


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    manifest = build(args.output)
    print("Packet frozen:", manifest["packet_sha256"])


if __name__ == "__main__":
    main()
