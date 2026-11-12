#!/usr/bin/env python3
"""Choose the paragraph pairs for the second reader comparison, with trim-only revisions.

usage: study_pairs.py <human directory>

Reads effect/<reviewer>/<document>.json. A pair is one paragraph of the raw
document and the same paragraph after revision. It must differ, hold 15 to 130
words in both versions, and contain no table, code block, heading, or list
marker. Ten pairs per reviewer model and arm, at most one per document and
arm, ordered by the SHA-256 of the document identifier. The side on which the
revised version appears is fixed by a hash. Writes study/pairs.json for the
page and study/key.json, which stays out of the page.
"""
import hashlib, json, re, sys
from pathlib import Path

ROOT = Path(sys.argv[1])
PER_CELL = 10


def blocks(text):
    return re.split(r"\n\s*\n", text)


def plain(block):
    stripped = block.strip()
    if not stripped or stripped.startswith(("#", "|", "```", ">", "<")) or "```" in stripped or "\n|" in stripped:
        return False
    if re.match(r"^(\s*[-*+] |\s*\d+[.)] )", stripped):
        return False
    return 15 <= len(stripped.split()) <= 130


# Paragraphs shown in the first comparison are not shown again.
first_pairs = json.loads((ROOT / "study/pairs.json").read_text())
first_key = json.loads((ROOT / "study/key.json").read_text())
shown = {(pair["right"] if first_key[pair["id"]]["revised"] == "left" else pair["left"]) for pair in first_pairs}
pairs, key, counts = [], {}, {}
for reviewer in ("claude", "astra"):
    directory = ROOT / "effect-trim" / reviewer
    results = sorted((json.loads(p.read_text()) for p in directory.glob("*.json")),
                     key=lambda r: hashlib.sha256(r["id"].encode()).hexdigest()) if directory.exists() else []
    for arm in ("reviewed", "control"):
        chosen = 0
        for result in results:
            if arm not in result or chosen == PER_CELL:
                continue
            before, after = blocks(result["raw"]), blocks(result[arm])
            if len(before) != len(after):
                continue
            changed = [(b, a) for b, a in zip(before, after) if b != a and plain(b) and plain(a) and b.strip() not in shown]
            if not changed:
                continue
            raw, revised = min(changed, key=lambda pair: hashlib.sha256(pair[0].encode()).hexdigest())
            identifier = hashlib.sha256(f"trim|{reviewer}|{arm}|{result['id']}".encode()).hexdigest()[:10]
            revised_left = int(hashlib.sha256(("side|" + identifier).encode()).hexdigest(), 16) % 2 == 0
            pairs.append({"id": identifier, "left": (revised if revised_left else raw).strip(),
                          "right": (raw if revised_left else revised).strip()})
            key[identifier] = {"reviewer": reviewer, "arm": arm, "document": result["id"],
                               "revised": "left" if revised_left else "right"}
            chosen += 1
        counts[f"{reviewer}/{arm}"] = chosen
pairs.sort(key=lambda p: hashlib.sha256(("order|" + p["id"]).encode()).hexdigest())
(ROOT / "study2").mkdir(exist_ok=True)
(ROOT / "study2/pairs.json").write_text(json.dumps(pairs, ensure_ascii=False, indent=1) + "\n")
(ROOT / "study2/key.json").write_text(json.dumps(key, indent=1, sort_keys=True) + "\n")
print(counts, "pairs", len(pairs))
