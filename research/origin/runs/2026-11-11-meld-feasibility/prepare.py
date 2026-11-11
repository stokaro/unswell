#!/usr/bin/env python3
"""Prepare the fixed historical and matched-response MELD inputs offline."""

from __future__ import annotations

import argparse
import hashlib
import importlib.util
import json
from pathlib import Path
import re


HERE = Path(__file__).resolve().parent
DOCUMENTS = HERE.parent / "2026-11-10-editlens-new-controls" / "documents.json"
RICHNESS = HERE.parent / "2026-11-07-richness-screen" / "run.py"
WORD = re.compile(r"\b\w+\b")


def digest(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def richness_module():
    spec = importlib.util.spec_from_file_location("meld_richness", RICHNESS)
    if spec is None or spec.loader is None:
        raise ValueError("Prose extraction helper unavailable")
    loaded = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(loaded)
    return loaded


def make_record(identity: dict, raw: str, richness) -> dict:
    prose = richness.prose(raw)
    words = len(WORD.findall(prose))
    return {**identity, "raw_sha256": digest(raw.encode()),
            "prose_sha256": digest(prose.encode()), "words": words,
            "eligible": words >= 100, "text": prose}


def controls(sources: Path, richness) -> list[dict]:
    selected = json.loads(DOCUMENTS.read_text())
    if len(selected) != 761:
        raise ValueError("Historical frame changed")
    rows = []
    for row in selected:
        raw = (sources / f"{row['source_id']}.source").read_bytes()
        blob = hashlib.sha1(f"blob {len(raw)}\0".encode() + raw).hexdigest()
        if (len(raw) != row["bytes"] or blob != row["blob_sha"]
                or digest(raw) != row["source_sha256"]):
            raise ValueError(f"Historical source changed: {row['source_id']}")
        text = raw.decode("utf-8")
        if "\0" in text:
            raise ValueError(f"NUL in historical source: {row['source_id']}")
        rows.append(make_record({"group": "control", "arm": row["arm"],
                                 "source_id": row["source_id"],
                                 "repository": row["repository"],
                                 "commit": row["commit"], "path": row["path"]},
                                text, richness))
    return rows


def matched(richness) -> list[dict]:
    with richness.verified_archive("inputs") as inputs:
        selection = richness.read_json(inputs, "selection.json")
        task_data = richness.read_json(inputs, "tasks.json")
    if len(selection) != 38 or len(task_data["tasks"]) != 38:
        raise ValueError("Matched source frame changed")
    chosen = {row["source_id"]: row for row in selection}
    tasks = {row["id"]: row for row in task_data["tasks"]}
    if len(tasks) != 38 or len({row["repository"] for row in tasks.values()}) != 20:
        raise ValueError("Matched source identities changed")
    rows = []
    for task_id, task in sorted(tasks.items()):
        source = chosen[task["source_id"]]
        if digest(task["text"].encode()) != source["source_sha256"]:
            raise ValueError(f"Matched source changed: {task_id}")
        rows.append(make_record({"group": "matched_source", "task_id": task_id,
                                 "repository": task["repository"]}, task["text"], richness))
    seen = set()
    with richness.verified_archive("outputs") as outputs:
        members = sorted(member.name for member in outputs.getmembers()
                         if member.name.endswith("/result.json") and "/raw/" in member.name)
        if len(members) != 456:
            raise ValueError("Matched response frame changed")
        for name in members:
            parent = name.rsplit("/", 1)[0]
            request = richness.read_json(outputs, parent + "/input.json")["request"]
            response = richness.read_json(outputs, name)
            task_id = request["task_id"]
            family = name.split("/")[1]
            operation = request["operation"]
            prompt = request["prompt"]
            identity = (task_id, family, operation, prompt)
            if (task_id not in tasks or family not in {"haiku", "luna", "qwen"}
                    or operation not in {"generate", "polish"}
                    or prompt not in {"neutral", "plain"} or identity in seen):
                raise ValueError(f"Invalid matched response: {name}")
            seen.add(identity)
            rows.append(make_record({"group": "response", "task_id": task_id,
                                     "repository": tasks[task_id]["repository"],
                                     "family": family, "operation": operation,
                                     "prompt": prompt}, response["text"], richness))
    if len(rows) != 494 or len(seen) != 456:
        raise ValueError("Matched packet incomplete")
    return rows


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--sources", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--manifest", type=Path, required=True)
    args = parser.parse_args()
    richness = richness_module()
    rows = controls(args.sources, richness) + matched(richness)
    with args.output.open("x") as stream:
        for row in rows:
            stream.write(json.dumps(row, ensure_ascii=False, sort_keys=True) + "\n")
    public_rows = [{key: value for key, value in row.items() if key != "text"}
                   for row in rows]
    manifest = {"format": "unswell-meld-input-manifest-v1",
                "input_sha256": digest(args.output.read_bytes()),
                "rows": public_rows}
    args.manifest.write_text(json.dumps(manifest, ensure_ascii=False,
                                         sort_keys=True, indent=2) + "\n")
    counts = {group: sum(row["group"] == group for row in rows)
              for group in ("control", "matched_source", "response")}
    eligible = {group: sum(row["group"] == group and row["eligible"] for row in rows)
                for group in counts}
    print(json.dumps({"counts": counts, "eligible": eligible,
                      "input_sha256": manifest["input_sha256"]}, sort_keys=True))


if __name__ == "__main__":
    main()
