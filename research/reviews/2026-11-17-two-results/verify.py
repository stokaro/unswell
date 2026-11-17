#!/usr/bin/env python3
"""Recount the record of the two results of the check of a revision.

usage: verify.py [--write]

The record holds no text. Its data are what `unswell compare` did, under two judges, with copies of pages that were
damaged in six ways or changed in three harmless ways, with revisions that were checked twice, and with pairs of
the fitting half. This program recounts every number of results.json and of RESULTS.md from those data and fails
when one of them differs, when a row is malformed, or when a data file is not the one that files.sha256 names. With
--write it writes results.json, RESULTS.md, and files.sha256 again.
"""
import gzip
import hashlib
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
DATA = ("copies", "revisions", "fitting", "same")

GPT, QWEN35 = "gpt-6-astra", "private-host/Qwen3.6-35B-A3B"
JUDGES = (GPT, QWEN35)
# The judge of a revision is of the other family than its reviser.
REVISER = {GPT: QWEN35, QWEN35: GPT}
SOUNDS = ("better", "no difference", "worse")
# The checks behind "nothing lost". A pair that fails one of them lost something.
CHECKS = ("statements", "sentences", "losses", "dashes")
# A damaged copy has one sentence removed, the word "not" removed from one or put into one, one word replaced by
# its opposite, or the last clause or the last words of one cut off. A harmless copy has two sentences joined, its
# negations written as contractions, or a phrase of padding left out.
DAMAGE, HARMLESS = ("sentence", "negation", "addnot", "swap", "clause", "tail"), ("join", "contract", "trim")
PAGES = ("set-s", "fresh")
# The copies of the measured part, for each judge.
COPIES = {"damaged": {"set-s": 87, "fresh": 132}, "harmless": 155}
# The revisions of the measured part, for each judge and run: 27 of posts and 30 of fresh texts.
REVISIONS = {"set-s": 27, "fresh": 30}
RUNS = (1, 2)
# The fitting half: 40 revisions for each judge and run, and 75 harmless copies for each judge and build.
FITTING = {"revision": 40, "harmless": 75}
# The run of the copies with this judge was made in three parts: twice the judge was no longer available, and the
# run went on with the same build when it was again. A copy belongs to the part that gave it its verdict, and the
# copies were judged in the order of the file. A judge that is not named here judged all its copies in one part.
PARTS = {GPT: (198, 55, 121)}
# Two controls after the run: the revisions of each reviser under the judge of its own family. The other judge had
# checked them, so the two judges can be read on the same revisions. Neither is a check of a revision.
SAME = (QWEN35, GPT)
# The bars, as they were written on issue 605 before the run on the measured part.
BARS = {"harmless_pass": 0.85, "damaged_pass": 2, "repeat": 0.85, "verdicts": 0.95}


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


def check_result(row, where, extra=()):
    """The part of a row that says what one check did. A check without a verdict has nothing else."""
    base = {"retried", "verdict"}
    if row.get("verdict") is False:
        if not base <= set(row) or {"nothing_lost", "failed", "sound", "accepted"} & set(row):
            fail(f"{where}: a check without a verdict has no result")
        return
    if not base | {"nothing_lost", "failed", "sound", "accepted"} | set(extra) <= set(row):
        fail(f"{where}: fields {sorted(row)}")
    if row["verdict"] is not True or not isinstance(row["retried"], bool):
        fail(f"{where}: verdict or retried")
    if row["sound"] not in SOUNDS or not isinstance(row["failed"], list) or any(check not in CHECKS for check in row["failed"]):
        fail(f"{where}: sound or failed checks")
    if row["nothing_lost"] is not (not row["failed"]):
        fail(f"{where}: nothing is lost exactly when no check failed")
    if row["accepted"] is not (row["nothing_lost"] and row["sound"] == "better"):
        fail(f"{where}: with all required, a pair is accepted exactly when nothing is lost and the sound is better")


def check_copy(row, seen):
    where = f"copy {row.get('judge')} {row.get('kind')} {row.get('page')}"
    if row.get("judge") not in JUDGES or row.get("pages") not in PAGES or row.get("kind") not in DAMAGE + HARMLESS:
        fail(f"{where}: unknown judge, pages, or kind")
    key = (row["judge"], row["kind"], row.get("page"))
    if not isinstance(row.get("page"), str) or key in seen:
        fail(f"{where}: listed twice or without a page")
    seen.add(key)
    if row.get("part") not in range(1, len(PARTS.get(row["judge"], (0,))) + 1):
        fail(f"{where}: a part that the run of this judge did not have")
    check_result(row, where)


def check_revision(row, seen):
    where = f"revision {row.get('judge')} {row.get('page')} run {row.get('run')}"
    if row.get("judge") not in JUDGES or row.get("reviser") != REVISER.get(row.get("judge")):
        fail(f"{where}: the judge of a revision is of the other family than its reviser")
    if row.get("pages") not in PAGES or row.get("run") not in RUNS:
        fail(f"{where}: unknown pages or run")
    key = (row["judge"], row.get("page"), row["run"])
    if not isinstance(row.get("page"), str) or key in seen:
        fail(f"{where}: listed twice or without a page")
    seen.add(key)
    check_result(row, where, extra=("statements", "kept") if row.get("verdict") else ())
    if row["verdict"] and not 0 <= row["kept"] <= row["statements"]:
        fail(f"{where}: statements kept")


def check_fitting(row, seen):
    where = f"fitting {row.get('what')} {row.get('judge')} {row.get('build')} {row.get('page')} run {row.get('run')}"
    if row.get("what") not in FITTING or row.get("judge") not in JUDGES or row.get("build") not in ("earlier", "record"):
        fail(f"{where}: unknown kind of pair, judge, or build")
    if row.get("run") not in RUNS or (row["what"] == "harmless" and row.get("kind") not in ("join", "contract")):
        fail(f"{where}: unknown run or kind")
    key = (row["what"], row["judge"], row["build"], row.get("kind"), row.get("page"), row["run"])
    if not isinstance(row.get("page"), str) or key in seen:
        fail(f"{where}: listed twice or without a page")
    seen.add(key)
    check_result(row, where)


def check_same(row, seen):
    where = f"control {row.get('judge')} {row.get('page')}"
    if row.get("judge") not in SAME or row.get("reviser") != row["judge"] or row.get("pages") not in PAGES:
        fail(f"{where}: a control is the revisions of one reviser under the judge of its own family")
    key = (row["judge"], row.get("page"))
    if not isinstance(row.get("page"), str) or key in seen:
        fail(f"{where}: listed twice or without a page")
    seen.add(key)
    check_result(row, where)


def share(hits, total):
    return {"count": hits, "of": total, "share": round(hits / total, 4) if total else 0.0}


def passed(rows):
    return sum(1 for row in rows if row["verdict"] and row["nothing_lost"])


def judged_all(judge, what, found, expected):
    """A judge has a row for every copy of a group, and no other."""
    if found != expected:
        fail(f"{judge}: {found} {what}, and the record says {expected}")


def part_counts(judge, mine):
    """What each part of the run of a judge found, so that a reader can see whether the parts differ."""
    sizes = PARTS.get(judge, (len(mine),))
    parts = []
    for number, size in enumerate(sizes, start=1):
        rows = [row for row in mine if row["part"] == number]
        judged_all(judge, f"copies in part {number}", len(rows), size)
        harmless, damaged = [row for row in rows if row["kind"] in HARMLESS], [row for row in rows if row["kind"] in DAMAGE]
        parts.append({"part": number, "copies": len(rows), "harmless": share(passed(harmless), len(harmless)),
                      "damaged": share(passed(damaged), len(damaged)), "verdicts": share(sum(row["verdict"] for row in rows), len(rows))})
    return parts


def copy_counts(copies):
    result = {}
    every = COPIES["harmless"] + sum(COPIES["damaged"].values())
    for judge in JUDGES:
        mine = [row for row in copies if row["judge"] == judge]
        harmless = [row for row in mine if row["kind"] in HARMLESS]
        judged_all(judge, "copies", len(mine), every)
        judged_all(judge, "harmless copies", len(harmless), COPIES["harmless"])
        cell = {"copies": every, "parts": part_counts(judge, mine), "harmless": share(passed(harmless), len(harmless)),
                "verdicts": share(sum(row["verdict"] for row in mine), len(mine)),
                "first_attempt": share(sum(1 for row in mine if row["verdict"] and not row["retried"]), len(mine)),
                "retried": sum(1 for row in mine if row["retried"]),
                "harmless_accepted_all": sum(1 for row in harmless if row["verdict"] and row["accepted"]),
                "by_kind": {}, "damaged": {}}
        for kind in DAMAGE + HARMLESS:
            rows = [row for row in mine if row["kind"] == kind]
            cell["by_kind"][kind] = {"copies": len(rows), "nothing_lost": passed(rows),
                                     "without_verdict": sum(1 for row in rows if not row["verdict"]),
                                     "failed": {check: sum(1 for row in rows if row["verdict"] and check in row["failed"]) for check in CHECKS}}
        for pages in PAGES:
            rows = [row for row in mine if row["kind"] in DAMAGE and row["pages"] == pages]
            judged_all(judge, f"damaged copies of {pages}", len(rows), COPIES["damaged"][pages])
            cell["damaged"][pages] = {"copies": len(rows), "nothing_lost": passed(rows),
                                      "accepted_all": sum(1 for row in rows if row["verdict"] and row["accepted"])}
        result[judge] = cell
    return result


def repeats(rows, value):
    """Of the pairs with a verdict in both runs, those with the same value in both."""
    by_page = {}
    for row in rows:
        by_page.setdefault(row["page"], {})[row["run"]] = row
    both = [pair for pair in by_page.values() if len(pair) == 2 and pair[1]["verdict"] and pair[2]["verdict"]]
    return share(sum(1 for pair in both if value(pair[1]) == value(pair[2])), len(both)), both


def revision_counts(revisions):
    result = {}
    for judge in JUDGES:
        mine = [row for row in revisions if row["judge"] == judge]
        for run in RUNS:
            for pages, expected in REVISIONS.items():
                found = sum(1 for row in mine if row["run"] == run and row["pages"] == pages)
                if found != expected:
                    fail(f"{judge} run {run}: {found} revisions of {pages}, and the record says {expected}")
        same, both = repeats(mine, lambda row: row["nothing_lost"])
        cell = {"repeat": same, "sound_repeat": repeats(mine, lambda row: row["sound"])[0],
                "accepted_all_both": sum(1 for pair in both if pair[1]["accepted"] and pair[2]["accepted"]),
                "nothing_lost_both": sum(1 for pair in both if pair[1]["nothing_lost"] and pair[2]["nothing_lost"]), "runs": {}}
        for run in RUNS:
            rows = [row for row in mine if row["run"] == run]
            judged = [row for row in rows if row["verdict"]]
            cell["runs"][str(run)] = {
                "verdicts": share(len(judged), len(rows)),
                "first_attempt": share(sum(1 for row in judged if not row["retried"]), len(rows)),
                "nothing_lost": passed(rows), "accepted_all": sum(1 for row in judged if row["accepted"]),
                "sound": {sound: sum(1 for row in judged if row["sound"] == sound) for sound in SOUNDS},
                "statements": share(sum(row["kept"] for row in judged), sum(row["statements"] for row in judged)),
                "by_pages": {pages: {"revisions": sum(1 for row in rows if row["pages"] == pages),
                                     "nothing_lost": passed([row for row in rows if row["pages"] == pages]),
                                     "accepted_all": sum(1 for row in judged if row["pages"] == pages and row["accepted"])} for pages in PAGES},
            }
        result[judge] = cell
    return result


def same_counts(same, revisions):
    """The revisions of each reviser under both judges: what each found, and how often they agree on a revision."""
    controls = []
    for reviser in SAME:
        other = REVISER[reviser]
        mine = [row for row in same if row["judge"] == reviser]
        first = {row["page"]: row for row in revisions if row["judge"] == other and row["run"] == 1}
        if len(mine) != sum(REVISIONS.values()) or {row["page"] for row in mine} != set(first):
            fail(f"the control of {reviser} has the revisions that the other judge checked, each once")
        both = [(row, first[row["page"]]) for row in mine if row["verdict"] and first[row["page"]]["verdict"]]
        by_pages = {pages: {"revisions": sum(1 for own, _ in both if own["pages"] == pages),
                            "own_family": sum(1 for own, _ in both if own["pages"] == pages and own["nothing_lost"]),
                            "other_family": sum(1 for own, theirs in both if own["pages"] == pages and theirs["nothing_lost"])}
                    for pages in PAGES}
        controls.append({"reviser": reviser, "judge_of_own_family": reviser, "judge_of_other_family": other, "revisions": len(both),
                         "without_verdict": len(mine) - sum(1 for row in mine if row["verdict"]),
                         "own_family": sum(1 for own, _ in both if own["nothing_lost"]),
                         "other_family": sum(1 for _, theirs in both if theirs["nothing_lost"]),
                         "both": sum(1 for own, theirs in both if own["nothing_lost"] and theirs["nothing_lost"]),
                         "agree": share(sum(1 for own, theirs in both if own["nothing_lost"] == theirs["nothing_lost"]), len(both)),
                         "by_pages": by_pages})
    return controls


def fitting_counts(fitting):
    result = {"revisions": {}, "harmless": []}
    for judge in JUDGES:
        mine = [row for row in fitting if row["what"] == "revision" and row["judge"] == judge]
        for run in RUNS:
            if sum(1 for row in mine if row["run"] == run) != FITTING["revision"]:
                fail(f"fitting, {judge} run {run}: not {FITTING['revision']} revisions")
        same, both = repeats(mine, lambda row: row["nothing_lost"])
        result["revisions"][judge] = {
            "repeat": same, "accepted_all_both": sum(1 for pair in both if pair[1]["accepted"] and pair[2]["accepted"]),
            "runs": {str(run): {"nothing_lost": passed([row for row in mine if row["run"] == run]),
                                "accepted_all": sum(1 for row in mine if row["run"] == run and row["verdict"] and row["accepted"])} for run in RUNS}}
    for judge, build in ((GPT, "earlier"), (QWEN35, "earlier"), (QWEN35, "record")):
        rows = [row for row in fitting if row["what"] == "harmless" and row["judge"] == judge and row["build"] == build]
        if len(rows) != FITTING["harmless"]:
            fail(f"fitting, {judge} with the {build} build: {len(rows)} harmless copies, and the record says {FITTING['harmless']}")
        judged = [row for row in rows if row["verdict"]]
        result["harmless"].append({
            "judge": judge, "build": build, "with_verdict": len(judged), "nothing_lost": passed(rows),
            "accepted_all": sum(1 for row in judged if row["accepted"]),
            "by_kind": {kind: {"copies": sum(1 for row in judged if row["kind"] == kind),
                               "nothing_lost": passed([row for row in rows if row["kind"] == kind])} for kind in ("join", "contract")}})
    return result


def bars(copies, revisions):
    """Each bar of the issue with the number that it is read on and whether it holds."""
    rows = []
    for judge in JUDGES:
        value = copies[judge]["harmless"]
        rows.append({"bar": "G1", "judge": judge, "read_on": "harmless and trimmed copies", "value": value, "kind": "at least",
                     "limit": BARS["harmless_pass"], "holds": value["count"] >= BARS["harmless_pass"] * value["of"]})
    for judge in JUDGES:
        for pages in PAGES:
            cell = copies[judge]["damaged"][pages]
            rows.append({"bar": "G2", "judge": judge, "read_on": f"damaged copies of {pages}", "value": share(cell["nothing_lost"], cell["copies"]),
                         "kind": "at most", "limit": BARS["damaged_pass"], "holds": cell["nothing_lost"] <= BARS["damaged_pass"]})
    for judge in JUDGES:
        value = revisions[judge]["repeat"]
        rows.append({"bar": "G3", "judge": judge, "read_on": "revisions with a verdict in both runs", "value": value, "kind": "at least",
                     "limit": BARS["repeat"], "holds": value["count"] >= BARS["repeat"] * value["of"]})
    for judge in JUDGES:
        runs = [("the copies", copies[judge]["verdicts"])]
        runs += [(f"the revisions, run {run}", revisions[judge]["runs"][str(run)]["verdicts"]) for run in RUNS]
        for name, value in runs:
            rows.append({"bar": "G4", "judge": judge, "read_on": name, "value": value, "kind": "at least", "limit": BARS["verdicts"],
                         "holds": value["count"] >= BARS["verdicts"] * value["of"]})
    return rows


def results(root=ROOT):
    copies, revisions, fitting, same = read(root, "copies"), read(root, "revisions"), read(root, "fitting"), read(root, "same")
    for rows, check in ((copies, check_copy), (revisions, check_revision), (fitting, check_fitting), (same, check_same)):
        seen = set()
        for row in rows:
            check(row, seen)
    counted_copies, counted_revisions = copy_counts(copies), revision_counts(revisions)
    read_bars = bars(counted_copies, counted_revisions)
    return {"judges": list(JUDGES), "bars": read_bars, "bars_total": len(read_bars), "bars_held": sum(1 for bar in read_bars if bar["holds"]),
            "copies": counted_copies, "revisions": counted_revisions, "fitting": fitting_counts(fitting),
            "controls": same_counts(same, revisions)}


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
    print(f"two results: {result['bars_held']} of {result['bars_total']} registered bars hold")
    return result


if __name__ == "__main__":
    try:
        main()
    except Malformed as error:
        sys.exit(f"verify: {error}")
