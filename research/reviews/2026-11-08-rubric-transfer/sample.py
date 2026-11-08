"""Select page-disjoint Ptah cases and non-Ptah editorial examples."""

import hashlib
import importlib.util
import json
from pathlib import Path
import sys


ROOT = Path(__file__).resolve().parent
TRANSFER = ROOT.parent / "2026-11-08-real-source-admission-transfer-v2"
SEED = "unswell-349-rubric-transfer-2026-11-08-v1"
ROWS_SHA256 = "628ccecd3f830fb37802e4479f8ea143a01b1addbf3f5549d714818503aa54aa"
GRADES_SHA256 = "a4194e85c3b04e26353d1ac740faa0efd838094f896ce3306406de6b8c7774db"


def encoded(value):
    return (json.dumps(value, ensure_ascii=False, sort_keys=True, indent=2) + "\n").encode("utf-8")


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def module(path, name):
    spec = importlib.util.spec_from_file_location(name, path)
    result = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(result)
    return result


def diagnosis(row):
    lines = row["hypothesis"].splitlines()
    value = next((line.removeprefix("Diagnosis: ") for line in lines
                  if line.startswith("Diagnosis: ")), None)
    if not value:
        raise ValueError("candidate diagnosis missing")
    return value


def target(row, sources):
    if len(row["targets"]) != 1 or row["source_sha256"] not in sources:
        raise ValueError("candidate has no unique original source")
    span = row["targets"][0]
    raw = sources[row["source_sha256"]][1]
    start, end = span["start"], span["end"]
    if not 0 <= start < end <= len(raw) or not 20 <= end - start <= 300:
        raise ValueError("candidate target outside eligible bounds")
    quote = raw[start:end].decode("utf-8")
    return start, end, quote


def rank(row):
    return sha((SEED + "\n" + row["id"]).encode("utf-8")), row["id"]


def prepare(rows_file, grades_file):
    packet = ROOT / "packet"
    if any((packet / name).exists() for name in ("sample.json", "key.json", "examples.json")):
        raise ValueError("sample already prepared")
    if sha(rows_file.read_bytes()) != ROWS_SHA256 or \
            sha(grades_file.read_bytes()) != GRADES_SHA256:
        raise ValueError("archived candidate or grade bank changed")
    rows = {row["id"]: row for line in rows_file.read_text().splitlines()
            for row in [json.loads(line)]}
    grades = json.loads(grades_file.read_text())["findings"]
    sources = module(TRANSFER / "run.py", "transfer_run").original_sources()
    old_sample = json.loads((TRANSFER / "packet/sample.json").read_text())
    old_pages = {item["page"] for item in old_sample["items"]}
    old_ids = {item["candidate_id"] for item in old_sample["items"]}
    candidates = []
    for row in rows.values():
        if row["group"] != "ptah" or row["page"] in old_pages or \
                row["id"] in old_ids or row["status"] not in ("accepted", "uncertain") or \
                len(row["targets"]) != 1 or row["source_sha256"] not in sources:
            continue
        try:
            target(row, sources)
        except (UnicodeDecodeError, ValueError):
            continue
        candidates.append(row)
    chosen, used_pages = [], set()
    for status in ("accepted", "uncertain"):
        eligible = sorted((r for r in candidates if r["status"] == status), key=rank)
        for row in eligible:
            if row["page"] not in used_pages:
                chosen.append(row)
                used_pages.add(row["page"])
            if sum(item["status"] == status for item in chosen) == 6:
                break
    if len(chosen) != 12 or len(used_pages) != 12:
        raise ValueError("cannot select 12 page-disjoint balanced Ptah cases")
    sample, key = [], []
    for row in sorted(chosen, key=rank):
        start, end, quote = target(row, sources)
        case_id = sha((SEED + "\ncase\n" + row["id"]).encode("utf-8"))[:16]
        sample.append({"case_id": case_id, "candidate_id": row["id"],
                       "category": row["category"], "group": row["group"],
                       "page": row["page"], "source_sha256": row["source_sha256"],
                       "target_start_byte": start, "target_end_byte": end,
                       "target_quote": quote, "proposed_diagnosis": diagnosis(row)})
        grade = grades[row["id"]]
        if grade["status"] != row["status"] or not grade["rationale"].strip():
            raise ValueError("archived grade changed")
        key.append({"case_id": case_id, "candidate_id": row["id"],
                    "old_status": grade["status"], "old_rationale": grade["rationale"]})
    by_id = {item["candidate_id"]: item for item in old_sample["items"]}
    examples = []
    for candidate_id in sorted(by_id):
        old = by_id[candidate_id]
        if old["group"] == "ptah":
            continue
        row = rows[candidate_id]
        start, end, quote = target(row, sources)
        raw = sources[row["source_sha256"]][1]
        before = raw[max(0, start - 240):start].decode("utf-8", errors="ignore")
        after = raw[end:end + 240].decode("utf-8", errors="ignore")
        grade = grades[candidate_id]
        if grade["status"] != row["status"]:
            raise ValueError("example grade changed")
        examples.append({"source_group": row["group"], "before": before,
                         "target": quote, "after": after,
                         "diagnosis": diagnosis(row), "prior_grade": grade["status"],
                         "prior_reason": grade["rationale"]})
    if len(examples) != 22 or len({r["source_group"] for r in examples}) != 11 or \
            {r["prior_grade"] for r in examples} != {"accepted", "uncertain"}:
        raise ValueError("example inventory changed")
    (packet / "sample.json").write_bytes(encoded({"format": "unswell-349-rubric-transfer-sample-v1",
                                                  "seed": SEED, "items": sample,
                                                  "source_pages": len(sources)}))
    (packet / "key.json").write_bytes(encoded(key))
    (packet / "examples.json").write_bytes(encoded(examples))
    return {"cases": len(sample), "pages": len(used_pages), "examples": len(examples),
            "status": {s: sum(k["old_status"] == s for k in key)
                       for s in ("accepted", "uncertain")}}


if __name__ == "__main__":
    if len(sys.argv) != 3:
        raise SystemExit("usage: sample.py <rows.jsonl> <expanded-grades.json>")
    print(json.dumps(prepare(Path(sys.argv[1]), Path(sys.argv[2])), sort_keys=True))
