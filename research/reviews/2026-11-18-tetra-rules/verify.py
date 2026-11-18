#!/usr/bin/env python3
"""Recount the record of the rules on the sentences that the editors of TETRA changed.

usage: verify.py [--write]

The record holds no text. Its data are one row for each sentence of each editor's file: where the sentence stands,
whether an edit of that editor touches it, and what `unswell check` finds on it with the default rules and with
every rule on. This program recounts every number of results.json and of RESULTS.md from those rows. It fails when
a number differs, when a row is malformed, when the data file is not the one that files.sha256 names, or when the
README states a count that the data do not give. With --write it writes results.json, RESULTS.md, and files.sha256
again.
"""
import difflib
import gzip
import hashlib
import itertools
import json
import random
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
DATA = "sentences"
FIELDS = {"paper", "editor", "section", "sentence", "words", "same", "edited", "redundancy", "default", "every"}
SECTIONS = ("title", "abstract", "introduction")
EDITORS = ("A", "B", "C")
# The eligible paper groups of the import record, each with three editors.
PAPERS = 56
RELEASE = "v0.1.0-alpha.20"
# The bar, as issue 365 states it before the run: with the default rules, the flagged share of edited sentences is
# at least five points above that of unedited sentences.
BAR = 0.05
# Sentences are compared at equal length in these bands of words.
BANDS = ((1, 10), (11, 20), (21, 30), (31, 40), (41, None))
# The interval resamples whole papers, since three editors read each paper.
RESAMPLES = 2000
SEED = 365
# A rule is listed by itself when it touches at least this many sentences.
LISTED = 20


class Malformed(Exception):
    """A row of the data, or the record around it, is not what the record says it is."""


def fail(message):
    raise Malformed(message)


def read(root):
    path = root / "data" / (DATA + ".jsonl.gz")
    if not path.exists():
        fail(f"missing data file: {path.name}")
    with gzip.open(path, "rt", encoding="utf-8") as handle:
        rows = [json.loads(line) for line in handle if line.strip()]
    seen = set()
    for row in rows:
        check_row(row, seen)
    papers = {row["paper"] for row in rows}
    if len(papers) != PAPERS:
        fail(f"the data hold {len(papers)} papers, and the record says {PAPERS}")
    for paper in papers:
        if {row["editor"] for row in rows if row["paper"] == paper} != set(EDITORS):
            fail(f"paper {paper} lacks an editor")
    return rows


def check_row(row, seen):
    if set(row) != FIELDS:
        fail(f"a sentence row has other fields: {sorted(row)}")
    key = (row["paper"], row["editor"], row["section"], row["sentence"])
    if not isinstance(row["paper"], str) or not re.fullmatch(r"[PW]\d\d-\d{4}", row["paper"]):
        fail(f"the paper of a sentence is no paper of the corpus: {key}")
    if row["editor"] not in EDITORS or row["section"] not in SECTIONS:
        fail(f"unknown editor or section: {key}")
    if type(row["sentence"]) is not int or row["sentence"] < 0 or key in seen:
        fail(f"a sentence is listed twice or has no number: {key}")
    seen.add(key)
    if type(row["words"]) is not int or row["words"] < 1 or not re.fullmatch(r"[0-9a-f]{12}", str(row["same"])):
        fail(f"the length or the hash of a sentence: {key}")
    if any(type(row[name]) is not bool for name in ("edited", "redundancy", "default")):
        fail(f"a mark of {key} is no yes or no")
    if row["redundancy"] and not row["edited"]:
        fail(f"{key} has a redundancy edit and no edit")
    rules = row["every"]
    if not isinstance(rules, list) or rules != sorted(set(rules)) or not all(isinstance(rule, str) and rule for rule in rules):
        fail(f"the rules on {key} are no sorted list of rule names")


def share(hits, total):
    return {"count": hits, "of": total, "share": round(hits / total, 4) if total else 0.0}


def flagged_by(system):
    """The test of a sentence for one run: the default rules, every rule, or one rule of the run with every rule."""
    if system == "default":
        return lambda row: row["default"]
    if system == "every":
        return lambda row: bool(row["every"])
    return lambda row: system in row["every"]


def difference(rows, flagged, mark="edited"):
    """The flagged share of marked sentences less that of the sentences that no edit touches."""
    marked = [row for row in rows if row[mark]]
    plain = [row for row in rows if not row["edited"]]
    if not marked or not plain:
        return None
    return sum(map(flagged, marked)) / len(marked) - sum(map(flagged, plain)) / len(plain)


def rates(rows, flagged, mark="edited"):
    marked = [row for row in rows if row[mark]]
    plain = [row for row in rows if not row["edited"]]
    return {"edited": share(sum(map(flagged, marked)), len(marked)), "unedited": share(sum(map(flagged, plain)), len(plain)),
            "difference": round(difference(rows, flagged, mark), 4)}


def band_of(row):
    return next(index for index, (low, high) in enumerate(BANDS) if row["words"] >= low and (high is None or row["words"] <= high))


def by_length(rows, flagged):
    """The same comparison inside each band of length, and one difference weighted by the edited sentences of a band."""
    bands, weighted, weight = [], 0.0, 0
    for index, (low, high) in enumerate(BANDS):
        part = [row for row in rows if band_of(row) == index]
        cell = rates(part, flagged)
        cell["words"] = f"{low} to {high}" if high else f"{low} and more"
        bands.append(cell)
        weighted += cell["edited"]["of"] * difference(part, flagged)
        weight += cell["edited"]["of"]
    return {"bands": bands, "difference_at_equal_length": round(weighted / weight, 4)}


def interval(rows, flagged):
    """A 95% interval of the difference from resampled papers."""
    papers = sorted({row["paper"] for row in rows})
    of_paper = {paper: [row for row in rows if row["paper"] == paper] for paper in papers}
    generator, values = random.Random(SEED), []
    for _ in range(RESAMPLES):
        drawn = [of_paper[papers[int(generator.random() * len(papers))]] for _ in papers]
        values.append(difference([row for part in drawn for row in part], flagged))
    values.sort()
    return [round(values[int(0.025 * RESAMPLES)], 4), round(values[int(0.975 * RESAMPLES) - 1], 4)]


def agreement(rows):
    """How two editors of one paper agree on the sentences to edit, on the sentences that both files hold."""
    files = {}
    for row in rows:
        files.setdefault((row["paper"], row["section"]), {}).setdefault(row["editor"], []).append(row)
    both = one = neither = longest = 0
    for editors in files.values():
        for first, second in itertools.combinations(sorted(editors), 2):
            left = sorted(editors[first], key=lambda row: row["sentence"])
            right = sorted(editors[second], key=lambda row: row["sentence"])
            longest += max(len(left), len(right))
            matcher = difflib.SequenceMatcher(None, [row["same"] for row in left], [row["same"] for row in right], autojunk=False)
            for block in matcher.get_matching_blocks():
                for offset in range(block.size):
                    marks = left[block.a + offset]["edited"] + right[block.b + offset]["edited"]
                    both, one, neither = both + (marks == 2), one + (marks == 1), neither + (marks == 0)
    pairs = both + one + neither
    observed, rate = (both + neither) / pairs, (2 * both + one) / (2 * pairs)
    chance = rate * rate + (1 - rate) * (1 - rate)
    return {"pairs": pairs, "of": longest, "both": both, "one": one, "neither": neither, "agreement": round(observed, 4),
            "kappa": round((observed - chance) / (1 - chance), 4)}


def by_rule(rows):
    """Each rule of the run with every rule that touches enough sentences, most frequent first."""
    names = sorted({rule for row in rows for rule in row["every"]})
    listed = []
    for name in names:
        cell = rates(rows, flagged_by(name))
        if cell["edited"]["count"] + cell["unedited"]["count"] >= LISTED:
            listed.append({"rule": name, **cell})
    listed.sort(key=lambda cell: (-(cell["edited"]["count"] + cell["unedited"]["count"]), cell["rule"]))
    return {"rules_that_fire": len(names), "listed_from": LISTED, "listed": listed}


def results(root=ROOT):
    rows = read(root)
    default, every = flagged_by("default"), flagged_by("every")
    value = difference(rows, default)
    return {
        "release": RELEASE,
        "papers": PAPERS,
        "files": len({(row["paper"], row["editor"]) for row in rows}),
        "sentences": len(rows),
        "edited": share(sum(row["edited"] for row in rows), len(rows)),
        "redundancy_edited": sum(row["redundancy"] for row in rows),
        "sections": {name: {"sentences": sum(row["section"] == name for row in rows),
                            "edited": sum(row["edited"] for row in rows if row["section"] == name)} for name in SECTIONS},
        "default": {**rates(rows, default), "interval": interval(rows, default), "redundancy": rates(rows, default, "redundancy")},
        "every": {**rates(rows, every), "interval": interval(rows, every), "redundancy": rates(rows, every, "redundancy"),
                  "length": by_length(rows, every), "by_rule": by_rule(rows)},
        "median_words": {"edited": median([row["words"] for row in rows if row["edited"]]),
                         "unedited": median([row["words"] for row in rows if not row["edited"]])},
        "editors": agreement(rows),
        "bar": {"statement": "default rules: flagged share of edited sentences less that of unedited sentences", "limit": BAR,
                "value": round(value, 4), "holds": value >= BAR},
    }


def median(values):
    return sorted(values)[len(values) // 2]


def of(cell):
    return f"{cell['count']:,} of {cell['of']:,}"


def percent(cell):
    """The share of a cell from its counts, so that it is rounded once."""
    return f"{100 * cell['count'] / cell['of']:.1f}%"


def points(value):
    return f"{100 * value:+.1f}"


def comparison(label, cell):
    return (f"| {label} | {of(cell['edited'])} | {percent(cell['edited'])} | {of(cell['unedited'])} | "
            f"{percent(cell['unedited'])} | {points(cell['difference'])} |")


def render(result):
    head = ["| | Edited, flagged | Share | Unedited, flagged | Share | Difference, points |", "| --- | ---: | ---: | ---: | ---: | ---: |"]
    default, every = result["default"], result["every"]
    lines = ["# Results", "", "Written by `verify.py` from `data/sentences.jsonl.gz`. Do not edit by hand.", "",
             f"{result['papers']} papers, {result['files']} files of editors, {result['sentences']:,} sentences; an edit touches "
             f"{of(result['edited'])} ({percent(result['edited'])}).", "",
             "## The bar", "",
             f"With the default rules the difference is {points(default['difference'])} points; the bar asks for at least "
             f"{points(result['bar']['limit'])}. The bar {'holds' if result['bar']['holds'] else 'does not hold'}.", "",
             "## Edited and unedited sentences", "", *head,
             comparison("Default rules, every edit", default),
             comparison("Default rules, redundancy edits", default["redundancy"]),
             comparison("Every rule, every edit", every),
             comparison("Every rule, redundancy edits", every["redundancy"]), "",
             f"Intervals of the difference at 95%, from resampled papers: default rules {points(default['interval'][0])} to "
             f"{points(default['interval'][1])}; every rule {points(every['interval'][0])} to {points(every['interval'][1])}.", "",
             "## Every rule, at equal length", "",
             f"The median edited sentence has {result['median_words']['edited']} words, and the median unedited sentence has "
             f"{result['median_words']['unedited']}.", "", *head]
    lines += [comparison(f"{band['words']} words", band) for band in every["length"]["bands"]]
    lines += ["", f"Weighted by the edited sentences of each band, the difference is "
              f"{points(every['length']['difference_at_equal_length'])} points.", "",
              "## Every rule, rule by rule", "",
              f"{every['by_rule']['rules_that_fire']} rules fire; the table lists those that touch at least "
              f"{every['by_rule']['listed_from']} sentences.", "", *head]
    lines += [comparison(f"`{cell['rule']}`", cell) for cell in every["by_rule"]["listed"]]
    editors = result["editors"]
    lines += ["", "## Two editors of one paper", "",
              f"{editors['pairs']:,} pairs of sentences that both files hold: both editors edit {editors['both']:,}, one edits "
              f"{editors['one']:,}, neither edits {editors['neither']:,}. Agreement {100 * editors['agreement']:.1f}%, kappa "
              f"{editors['kappa']:.2f}.", ""]
    return "\n".join(lines)


def stated(result):
    """The counts that the README states, in the words that it must use."""
    default, every = result["default"], result["every"]
    return [f"{result['sentences']:,} sentences", of(result["edited"]), of(default["edited"]), of(default["unedited"]),
            of(every["edited"]), of(every["unedited"]), f"{points(default['difference'])} points", f"{points(every['difference'])} points",
            f"{points(every['length']['difference_at_equal_length'])} points", f"kappa {result['editors']['kappa']:.2f}"]


def digests(root):
    return f"{hashlib.sha256((root / 'data' / (DATA + '.jsonl.gz')).read_bytes()).hexdigest()}  data/{DATA}.jsonl.gz\n"


def encode(result):
    return json.dumps(result, indent=2, sort_keys=True) + "\n"


def main(root=ROOT, write="--write" in sys.argv[1:]):
    result = results(root)
    expected = {"results.json": encode(result), "RESULTS.md": render(result), "files.sha256": digests(root)}
    for name, text in expected.items():
        path = root / name
        if write:
            path.write_text(text)
        elif not path.exists() or path.read_text() != text:
            fail(f"{name} is not what the data give; run verify.py --write and read the difference")
    readme = (root / "README.md").read_text() if (root / "README.md").exists() else ""
    for words in stated(result):
        if words not in readme:
            fail(f"the README does not state '{words}', which the data give")
    print(f"tetra rules: default rules {points(result['bar']['value'])} points against a bar of {points(BAR)}")
    return result


if __name__ == "__main__":
    try:
        main()
    except Malformed as error:
        sys.exit(f"verify: {error}")
