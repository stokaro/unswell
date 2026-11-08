"""Verify source identities and original UTF-8 spans in the source-role review."""

from __future__ import annotations

import hashlib
import gzip
import json
from pathlib import Path, PurePosixPath


ROOT = Path(__file__).resolve().parent
REVIEW = ROOT / "review.json"


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def checked_path(root: Path, name: str) -> Path:
    relative = PurePosixPath(name)
    require(
        not relative.is_absolute()
        and len(relative.parts) >= 3
        and relative.parts[0] == "sources"
        and ".." not in relative.parts,
        f"unsafe source path: {name}",
    )
    return root.joinpath(*relative.parts)


def verify(root: Path, review: dict) -> dict[str, int]:
    require(review["status"] == "exposed source-only development", "changed status")
    require("Codex assistant" in review["reviewer"], "missing actual reviewer")
    sources = review["sources"]
    require(set(sources) == {"caddy", "carbon", "animal-ai", "discourse"}, "changed sources")
    contents = {}
    for source_id, source in sources.items():
        require(len(source["revision"]) == 40, f"bad revision: {source_id}")
        stored = checked_path(root, source["path"]).read_bytes()
        content = gzip.decompress(stored) if source["path"].endswith(".gz") else stored
        license_content = checked_path(root, source["license_path"]).read_bytes()
        require(hashlib.sha256(content).hexdigest() == source["sha256"], f"source digest: {source_id}")
        require(
            hashlib.sha256(license_content).hexdigest() == source["license_sha256"],
            f"license digest: {source_id}",
        )
        git_blob = hashlib.sha1(f"blob {len(content)}\0".encode() + content).hexdigest()
        require(git_blob == source["blob"], f"Git blob: {source_id}")
        content.decode("utf-8")
        contents[source_id] = content

    counts = {"positive": 0, "control": 0, "unresolved": 0}
    seen = set()
    for case in review["cases"]:
        case_id = case["id"]
        require(case_id not in seen, f"duplicate case: {case_id}")
        seen.add(case_id)
        role = case["role"]
        require(role in counts, f"unknown role: {case_id}")
        require(case["source"] in contents, f"unknown source: {case_id}")
        require(case["rationale"] and case["safe_action"], f"missing review: {case_id}")
        spans = case["spans"]
        require(spans, f"empty spans: {case_id}")
        previous_end = -1
        content = contents[case["source"]]
        for start, end, quote in spans:
            require(
                isinstance(start, int)
                and isinstance(end, int)
                and previous_end <= start < end <= len(content),
                f"invalid span: {case_id}",
            )
            require(content[start:end] == quote.encode("utf-8"), f"changed quote: {case_id}")
            previous_end = end
        counts[role] += 1
    require(counts == {"positive": 3, "control": 7, "unresolved": 1}, "changed case counts")
    return counts


if __name__ == "__main__":
    result = verify(ROOT, json.loads(REVIEW.read_text(encoding="utf-8")))
    print(f"verified 4 exact sources and {sum(result.values())} source-role cases: {result}")
