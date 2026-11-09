#!/usr/bin/env python3
"""Collect source-bound shape candidates on pinned, source-disjoint pages."""

import argparse
import hashlib
import importlib.util
import json
import re
import subprocess
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
LOCAL = ROOT / "artifacts/research/2026-11-09-shape-context-transfer"
SHAPES = HERE.parent / "2026-11-09-sentence-shape-screen/screen.py"
MAX_BYTES = 200_000
MAX_MATCHES = 200


def digest(raw):
    return hashlib.sha256(raw).hexdigest()


def require(condition, message):
    if not condition:
        raise ValueError(message)


def load_shapes(expected_hash):
    require(digest(SHAPES.read_bytes()) == expected_hash, "shape source changed")
    spec = importlib.util.spec_from_file_location("frozen_shapes", SHAPES)
    require(spec is not None and spec.loader is not None, "shape import unavailable")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return {key: re.compile(value, re.IGNORECASE)
            for key, value in module.PATTERNS.items()}


def byte_span(raw_slice, match, base):
    start = base + len(raw_slice[:match.start()].encode("utf-8"))
    end = base + len(raw_slice[:match.end()].encode("utf-8"))
    return start, end


def within_one_segment(start, end, segments):
    return any(a <= start < end <= b for a, b in segments)


def extract_candidates(source, units, excluded, shapes):
    raw = source.encode("utf-8")
    exclusions = [(x["span"]["start"], x["span"]["end"]) for x in excluded]
    found = {}
    for ordinal, unit in enumerate(units):
        require(unit["ordinal"] == ordinal, "unit ordinal changed")
        base, bound = unit["raw_span"]["start"], unit["raw_span"]["end"]
        raw_slice = unit["raw_slice"]
        require(0 <= base < bound <= len(raw), "unit outside source")
        require(raw[base:bound].decode("utf-8") == raw_slice, "raw unit mismatch")
        segments = [(x["start"], x["end"])
                    for x in unit["binding"]["segments"]]
        for a, b in segments:
            require(0 <= a < b <= len(raw) and base <= a < b <= bound,
                    "eligible segment outside raw unit")
            require(all(b <= left or a >= right for left, right in exclusions),
                    "eligible segment intersects protected source")
        for name, pattern in shapes.items():
            for match in pattern.finditer(raw_slice):
                start, end = byte_span(raw_slice, match, base)
                require(raw[start:end].decode("utf-8") == match.group(),
                        "candidate byte binding mismatch")
                if not within_one_segment(start, end, segments):
                    continue
                found[(name, start, end)] = {
                    "shape": name, "start": start, "end": end,
                    "quote": match.group(),
                }
    return [found[key] for key in sorted(found,
                                         key=lambda key: (key[1], key[2], key[0]))]


def context(raw, start, end):
    line_start = raw.rfind(b"\n", 0, start) + 1
    line_end = raw.find(b"\n", end)
    if line_end < 0:
        line_end = len(raw)
    previous_end = max(0, line_start - 1)
    previous_start = raw.rfind(b"\n", 0, previous_end) + 1
    next_start = min(len(raw), line_end + 1)
    next_end = raw.find(b"\n", next_start)
    if next_end < 0:
        next_end = len(raw)
    prefix = raw[:line_start].decode("utf-8")
    headings = re.findall(r"(?m)^#{1,6}\s+(.+)$", prefix)
    return {
        "heading": headings[-1] if headings else "",
        "previous_line": raw[previous_start:previous_end].decode("utf-8"),
        "line": raw[line_start:line_end].decode("utf-8"),
        "next_line": raw[next_start:next_end].decode("utf-8"),
    }


def export_source(exporter, source, source_hash, page):
    request = json.dumps({
        "version": 1, "basis": "source-function-inventory",
        "pages": [{"id": page, "format": "markdown", "sha256": source_hash,
                   "text": source}],
    }, sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode("utf-8")
    try:
        run = subprocess.run([str(exporter), "--source-inventory"], input=request,
                             capture_output=True, timeout=30, check=False)
    except subprocess.TimeoutExpired:
        return None, "timeout"
    if run.returncode:
        return None, "exporter_error"
    try:
        result = json.loads(run.stdout)
    except (UnicodeDecodeError, json.JSONDecodeError):
        return None, "invalid_exporter_json"
    require(result["version"] == 1 and result["basis"] == "source-function-inventory"
            and result["input_sha256"] == digest(request) and
            len(result["pages"]) == 1, "exporter response identity mismatch")
    page_result = result["pages"][0]
    require(page_result["id"] == page and page_result["sha256"] == source_hash
            and page_result["bytes"] == len(source.encode("utf-8")),
            "exported source identity mismatch")
    return page_result, None


def collect(exporter):
    freeze = json.loads((HERE / "freeze.json").read_text())
    for name, path in (("protocol", HERE / "PROTOCOL.md"),
                       ("collector", HERE / "collect.py"),
                       ("sources", HERE / "sources.json")):
        require(digest(path.read_bytes()) == freeze[name + "_sha256"],
                f"frozen {name} changed")
    require(digest(exporter.read_bytes()) == freeze["exporter_sha256"],
            "exporter binary changed")
    shapes = load_shapes(freeze["shapes_sha256"])
    manifest = json.loads((HERE / "sources.json").read_text())
    sources = manifest["sources"]
    require(len(sources) == 12 and [x["number"] for x in sources] ==
            list(range(1, 13)), "source count or order changed")
    rows = {"design": [], "transfer": []}
    status = []
    for item in sources:
        number = item["number"]
        partition = "design" if number <= 6 else "transfer"
        require(item["partition"] == partition, "source partition changed")
        page = f"page-{number:03d}.source"
        require(item["source_file"] == page, "source file name changed")
        raw = (LOCAL / "sources" / page).read_bytes()
        require(digest(raw) == item["source_sha256"] and len(raw) == item["source_bytes"],
                "source SHA-256 or size changed")
        source = raw.decode("utf-8", "strict")
        if len(raw) > MAX_BYTES:
            status.append({"page": number, "partition": partition,
                           "state": "oversize", "source_bytes": len(raw)})
            continue
        inventory, error = export_source(exporter, source,
                                         item["source_sha256"], page)
        if error:
            status.append({"page": number, "partition": partition,
                           "state": error, "source_bytes": len(raw)})
            continue
        matches = extract_candidates(source, inventory["units"],
                                     inventory["excluded"], shapes)
        for match in matches:
            match.update({"page": number, "repository": item["repository"],
                          "source_sha256": item["source_sha256"],
                          "context": context(raw, match["start"], match["end"])})
        rows[partition].extend(matches)
        status.append({"page": number, "partition": partition,
                       "state": "complete", "source_bytes": len(raw),
                       "eligible_units": len(inventory["units"]),
                       "candidate_matches": len(matches)})
    require(len(rows["design"]) + len(rows["transfer"]) <= MAX_MATCHES,
            "candidate overload; stop before review")
    LOCAL.mkdir(parents=True, exist_ok=True)
    for partition in ("design", "transfer"):
        path = LOCAL / f"{partition}-candidates.json"
        require(not path.exists(), "candidate output already exists")
        path.write_text(json.dumps(rows[partition], indent=2,
                                   ensure_ascii=False) + "\n")
    summary = {
        "format": "unswell-349-shape-context-candidates-v1",
        "source_count": len(sources),
        "main_commit": freeze["main_commit"],
        "input_sha256": {name: freeze[name + "_sha256"] for name in
                         ("protocol", "collector", "sources", "shapes", "exporter")},
        "status": status,
        "design_matches": len(rows["design"]),
        "transfer_matches": len(rows["transfer"]),
    }
    (HERE / "candidate-summary.json").write_text(
        json.dumps(summary, indent=2, sort_keys=True) + "\n")
    return summary


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--exporter", type=Path, required=True)
    args = parser.parse_args()
    summary = collect(args.exporter)
    print(json.dumps({"pages": summary["source_count"],
                      "design_matches": summary["design_matches"],
                      "transfer_matches": summary["transfer_matches"],
                      "unavailable_pages": sum(x["state"] != "complete"
                                               for x in summary["status"])},
                     sort_keys=True))


if __name__ == "__main__":
    main()
