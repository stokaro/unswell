#!/usr/bin/env python3
"""Recount the record of two habits of a page.

usage: verify.py [--write]

The record holds no text. Its data are, for each page of twelve sets, how many places each of two rules of
`unswell check` reported, and for each pair of a file before and after a commit that says it removes AI-sounding
wording, whether each rule reported the file before and after. This program recounts every number of results.json
and of RESULTS.md from those data and fails when one of them differs, when a row is malformed, or when a data file
is not the one that files.sha256 names. With --write it writes results.json, RESULTS.md, and files.sha256 again.
"""
import gzip
import hashlib
import json
import math
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
DATA = ("pages", "pairs")

# The two rules, by the name of their field in the data.
RULES = {"denial": "syntax.closing-denial-density", "register": "syntax.uniform-register"}
# The sets of pages, each with its parts and the number of pages in each part. A set without parts is read whole.
SETS = {
    "set-s": {"fitting": 40, "measured": 30},
    "human-blog": {"fitting": 74, "measured": 77},
    "human-styled": {"fitting": 228, "measured": 205},
    "human-docs": {"fitting": 60, "measured": 61},
    "model-guided-1": {"fitting": 74},
    "model-guided-2": {"fitting": 73},
    "model-plain-1": {"fitting": 74},
    "model-plain-2": {"fitting": 74},
    "fresh-1": {"measured": 60},
    "fresh-2": {"measured": 60},
    "public-today": {"whole": 3302},
    "panel": {"whole": 455},
}
HUMAN = ("human-blog", "human-styled", "human-docs")
FRESH = ("fresh-1", "fresh-2")
PAIR_PARTS = ("development", "test")
# The closing-denial rule reports a page when more than two of its paragraphs end on a denial.
DENIAL_LEAST = 3
# The bars, as they were written on issue 603 before the run on the measured parts.
BARS = {"human_limit": 0.05, "posts_denial": 0.35, "posts_register": 0.40, "posts_either": 0.70, "fresh_register": 0.70}


class Malformed(Exception):
    """A row of the data, or the record around it, is not what the record says it is."""


def fail(message):
    raise Malformed(message)


def read(root, name):
    path = root / "data" / (name + ".jsonl.gz")
    if not path.exists():
        fail(f"missing data file: {path.name}")
    with gzip.open(path, "rt", encoding="utf-8") as handle:
        return [json.loads(line) for line in handle if line.strip()]


def count(value, least=0):
    return isinstance(value, int) and not isinstance(value, bool) and value >= least


def check_page(row, seen):
    if set(row) != {"corpus", "part", "id", "words", "denial", "register"}:
        fail(f"a page row has other fields: {sorted(row)}")
    if row["corpus"] not in SETS or row["part"] not in SETS[row["corpus"]]:
        fail(f"unknown set or part: {row['corpus']} {row['part']}")
    key = (row["corpus"], row["id"])
    if not isinstance(row["id"], str) or key in seen:
        fail(f"a page is listed twice or has no id: {key}")
    seen.add(key)
    pattern = r"s-\d\d" if row["corpus"] == "set-s" else r"[0-9a-f]{12}"
    if not re.fullmatch(pattern, row["id"]):
        fail(f"the id of a page is not of its set: {key}")
    if not count(row["words"]):
        fail(f"words of {key}")
    if not count(row["denial"]) or 0 < row["denial"] < DENIAL_LEAST:
        fail(f"places of the denial rule on {key}: {row['denial']}")
    if row["register"] not in (0, 1):
        fail(f"the register rule reports a page once: {key}")


def check_pair(row, seen):
    if set(row) != {"part", "id", "denial", "register"}:
        fail(f"a pair row has other fields: {sorted(row)}")
    if row["part"] not in PAIR_PARTS or not isinstance(row["id"], str) or row["id"] in seen:
        fail(f"unknown part, or a pair listed twice: {row.get('id')}")
    seen.add(row["id"])
    for rule in RULES:
        sides = row[rule]
        if not isinstance(sides, list) or len(sides) != 2 or any(side not in (0, 1) for side in sides):
            fail(f"a pair has a before and an after for each rule: {row['id']}")


def wilson(hits, total):
    """The 95% Wilson interval of a share."""
    if total == 0:
        return [0.0, 0.0]
    z = 1.959963984540054
    share = hits / total
    center = share + z * z / (2 * total)
    spread = z * math.sqrt(share * (1 - share) / total + z * z / (4 * total * total))
    return [round((center - spread) / (1 + z * z / total), 4), round((center + spread) / (1 + z * z / total), 4)]


def share(hits, total):
    return {"count": hits, "of": total, "share": round(hits / total, 4) if total else 0.0, "interval": wilson(hits, total)}


def page_counts(pages):
    cells = {}
    for name, parts in SETS.items():
        for part, expected in parts.items():
            rows = [row for row in pages if row["corpus"] == name and row["part"] == part]
            if len(rows) != expected:
                fail(f"{name} {part} has {len(rows)} pages, and the record says {expected}")
            cell = {rule: share(sum(1 for row in rows if row[rule] > 0), len(rows)) for rule in RULES}
            cell["either"] = share(sum(1 for row in rows if any(row[rule] > 0 for rule in RULES)), len(rows))
            cell["pages"] = len(rows)
            cells[f"{name}/{part}"] = cell
    return cells


def together(pages, names, part):
    rows = [row for row in pages if row["corpus"] in names and row["part"] == part]
    cell = {rule: share(sum(1 for row in rows if row[rule] > 0), len(rows)) for rule in RULES}
    cell["pages"] = len(rows)
    return cell


def bars(cells, fresh):
    """Each bar of the issue with the number that it is read on and whether it holds."""
    rows = []
    for name in HUMAN:
        for rule in RULES:
            value = cells[f"{name}/measured"][rule]
            rows.append({"bar": f"quiet on {name}: {RULES[rule]}", "value": value, "limit": BARS["human_limit"], "kind": "at most",
                         "holds": value["count"] <= BARS["human_limit"] * value["of"]})
    posts = cells["set-s/measured"]
    for label, value, limit in (
            (f"posts: {RULES['denial']}", posts["denial"], BARS["posts_denial"]),
            (f"posts: {RULES['register']}", posts["register"], BARS["posts_register"]),
            ("posts: either rule", posts["either"], BARS["posts_either"]),
            (f"fresh model texts: {RULES['register']}", fresh["register"], BARS["fresh_register"])):
        rows.append({"bar": label, "value": value, "limit": limit, "kind": "at least", "holds": value["count"] >= limit * value["of"]})
    return rows


def pair_counts(pairs):
    result = {}
    for part in PAIR_PARTS:
        rows = [row for row in pairs if row["part"] == part]
        if not rows:
            fail(f"no pair of the {part} part")
        cell = {"pairs": len(rows)}
        for rule in RULES:
            cell[rule] = {"before": sum(row[rule][0] for row in rows), "after": sum(row[rule][1] for row in rows),
                          "lost": sum(1 for row in rows if row[rule] == [1, 0]), "gained": sum(1 for row in rows if row[rule] == [0, 1])}
        result[part] = cell
    return result


def results(root=ROOT):
    pages, pairs = read(root, "pages"), read(root, "pairs")
    seen = set()
    for row in pages:
        check_page(row, seen)
    seen = set()
    for row in pairs:
        check_pair(row, seen)
    cells = page_counts(pages)
    fresh = together(pages, FRESH, "measured")
    documentation = together(pages, ("human-docs",), "fitting"), together(pages, ("human-docs",), "measured")
    every = {rule: share(sum(cell[rule]["count"] for cell in documentation), sum(cell["pages"] for cell in documentation)) for rule in RULES}
    read_bars = bars(cells, fresh)
    return {"rules": RULES, "bars": read_bars, "bars_held": sum(1 for bar in read_bars if bar["holds"]), "bars_total": len(read_bars),
            "sets": cells, "fresh_together": fresh, "documentation_together": every, "pairs": pair_counts(pairs)}


def digests(root):
    return "".join(f"{hashlib.sha256((root / 'data' / (name + '.jsonl.gz')).read_bytes()).hexdigest()}  data/{name}.jsonl.gz\n" for name in DATA)


def encode(result):
    return json.dumps(result, indent=2, sort_keys=True) + "\n"


def main(root=ROOT, write="--write" in sys.argv[1:]):
    sys.path.insert(0, str(Path(__file__).resolve().parent))
    import render
    result = results(root)
    expected = {"results.json": encode(result), "RESULTS.md": render.render(result), "files.sha256": digests(root)}
    for name, text in expected.items():
        path = root / name
        if write:
            path.write_text(text)
        elif not path.exists() or path.read_text() != text:
            fail(f"{name} is not what the data give; run verify.py --write and read the difference")
    print(f"page habits: {result['bars_held']} of {result['bars_total']} registered bars hold")
    return result


if __name__ == "__main__":
    try:
        main()
    except Malformed as error:
        sys.exit(f"verify: {error}")
