#!/usr/bin/env python3
"""Recount the record of the fourth form of a closing denial.

usage: verify.py [--write]

The record holds no text. Its data are, for each page of twelve sets, what `syntax.closing-denial-density` reports
with three forms of a denial and with four: the places that the rule lists with its limits as they ship, and the
closing sentences that it takes for a denial with no limit on the page. This program recounts every number of
results.json and of RESULTS.md from those data and fails when one of them differs, when a row is malformed, or
when the data file is not the one that files.sha256 names. With --write it writes results.json, RESULTS.md, and
files.sha256 again.
"""
import gzip
import hashlib
import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
DATA = ("pages",)

# The two builds of the rule, by the name of their field in the data.
BUILDS = {"three": "three forms", "four": "four forms"}
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
# The rule reports a page when more than two of its paragraphs end on a denial.
LEAST = 3
# The bars, as they were written on issue 607 before the run on the measured parts.
BARS = {"human_limit": 0.05, "human_added": 3, "posts": 0.35, "gain": 0.05}


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


def count(value):
    return isinstance(value, int) and not isinstance(value, bool) and value >= 0


def check_page(row, seen):
    if set(row) != {"corpus", "part", "id", "three", "four", "counted_three", "counted_four", "kept"}:
        fail(f"a page row has other fields: {sorted(row)}")
    if row["corpus"] not in SETS or row["part"] not in SETS[row["corpus"]]:
        fail(f"unknown set or part: {row['corpus']} {row['part']}")
    key = (row["corpus"], row["id"])
    if not isinstance(row["id"], str) or key in seen:
        fail(f"a page is listed twice or has no id: {key}")
    seen.add(key)
    if not re.fullmatch(r"s-\d\d" if row["corpus"] == "set-s" else r"[0-9a-f]{12}", row["id"]):
        fail(f"the id of a page is not of its set: {key}")
    for build in BUILDS:
        places, counted = row[build], row["counted_" + build]
        if not count(places) or not count(counted):
            fail(f"counts of {key} with {BUILDS[build]}")
        # A page that is reported lists every closing denial that the rule counts on it, and at least three.
        if places not in (0, counted) or 0 < places < LEAST:
            fail(f"places on {key} with {BUILDS[build]}: {places} listed, {counted} counted")
    if not count(row["kept"]) or row["kept"] > min(row["counted_three"], row["counted_four"]):
        fail(f"closing denials of {key} that both builds count: {row['kept']}")


def share(hits, total):
    return {"count": hits, "of": total, "share": round(hits / total, 4) if total else 0.0}


def cell_of(rows):
    """What both builds report on the pages of one set and part."""
    return {
        "pages": len(rows),
        "three": share(sum(1 for row in rows if row["three"] > 0), len(rows)),
        "four": share(sum(1 for row in rows if row["four"] > 0), len(rows)),
        "added": sum(1 for row in rows if row["four"] > 0 and row["three"] == 0),
        "lost": sum(1 for row in rows if row["three"] > 0 and row["four"] == 0),
        "counted_three": sum(row["counted_three"] for row in rows),
        "counted_four": sum(row["counted_four"] for row in rows),
        "not_kept": sum(row["counted_three"] - row["kept"] for row in rows),
    }


def page_counts(pages):
    cells = {}
    for name, parts in SETS.items():
        for part, expected in parts.items():
            rows = [row for row in pages if row["corpus"] == name and row["part"] == part]
            if len(rows) != expected:
                fail(f"{name} {part} has {len(rows)} pages, and the record says {expected}")
            cells[f"{name}/{part}"] = cell_of(rows)
    return cells


def bars(cells, people, measured):
    """Each bar of the issue with the number that it is read on and whether it holds."""
    rows = []
    for name in HUMAN:
        value = cells[f"{name}/measured"]["four"]
        rows.append({"bar": f"H. quiet on {name}", "value": value, "limit": BARS["human_limit"], "kind": "at most, share",
                     "holds": value["count"] <= BARS["human_limit"] * value["of"]})
    rows.append({"bar": "H4. pages of people that only the fourth form reports", "value": share(people["added"], people["pages"]),
                 "limit": BARS["human_added"], "kind": "at most, pages", "holds": people["added"] <= BARS["human_added"]})
    posts = cells["set-s/measured"]
    rows.append({"bar": "P. the posts", "value": posts["four"], "limit": BARS["posts"], "kind": "at least, share",
                 "holds": posts["four"]["count"] >= BARS["posts"] * posts["four"]["of"]})
    more = posts["counted_four"] - posts["counted_three"]
    rows.append({"bar": "G. closing denials that the fourth form adds on the posts", "value": share(more, posts["counted_three"]),
                 "limit": BARS["gain"], "kind": "at least, share", "holds": more >= BARS["gain"] * posts["counted_three"]})
    rows.append({"bar": "K. pages that lose their report", "value": share(measured["lost"], measured["pages"]), "limit": 0,
                 "kind": "at most, pages", "holds": measured["lost"] == 0})
    return rows


def results(root=ROOT):
    pages = read(root, "pages")
    seen = set()
    for row in pages:
        check_page(row, seen)
    cells = page_counts(pages)
    people = cell_of([row for row in pages if row["corpus"] in HUMAN and row["part"] == "measured"])
    measured = cell_of([row for row in pages if row["part"] == "measured"])
    fresh = cell_of([row for row in pages if row["corpus"] in FRESH])
    read_bars = bars(cells, people, measured)
    return {"rule": "syntax.closing-denial-density", "builds": BUILDS, "bars": read_bars, "bars_held": sum(1 for bar in read_bars if bar["holds"]),
            "bars_total": len(read_bars), "sets": cells, "people_measured": people, "measured": measured, "fresh_together": fresh,
            "not_kept": sum(cell["not_kept"] for cell in cells.values())}


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
    print(f"late denial: {result['bars_held']} of {result['bars_total']} registered bars hold")
    return result


if __name__ == "__main__":
    try:
        main()
    except Malformed as error:
        sys.exit(f"verify: {error}")
