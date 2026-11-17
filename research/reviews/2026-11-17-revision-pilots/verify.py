#!/usr/bin/env python3
"""Recount the record of three pilots of revising a post.

usage: verify.py [--write]

The record holds no text. Its data are what two rules of a page reported on each post of the fitting half before
and after a revision, which text each of three blind judges named in each pair of a post and a human post, and what
one judge named in each round of a revision by its own remarks. This program recounts every number of results.json
and of RESULTS.md from those data and fails when one of them differs, when a row is malformed, or when a data file
is not the one that files.sha256 names. With --write it writes results.json, RESULTS.md, and files.sha256 again.
"""
import gzip
import hashlib
import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
DATA = ("revisions", "answers", "rounds", "subjects")

GPT, QWEN27, QWEN35 = "gpt-6-astra", "private-host/Qwen3.8-27B", "private-host/Qwen3.6-35B-A3B"
JUDGES = (GPT, QWEN27, QWEN35)
# The post as it is, and two ways of revising it. One model revised in both ways.
AS_IT_IS = "as-it-is"
ARMS = ("closing-sentences", "manner")
# The judges that were asked about each way. The second way was dropped after the first judge.
ASKED = {"closing-sentences": JUDGES, "manner": (QWEN27,)}
NAMED = ("post", "human", "same")
POSTS, PAIRS = 40, 3
# The pilot in rounds: the post as it is and four revisions, two pairs each.
ROUNDS, ROUND_PAIRS = 5, 2
# Texts that two models wrote under a style guide from the outline of a human post, each against that post. Two
# of the judges compared them when the fitting half was first read.
SUBJECT_JUDGES, WRITERS, SUBJECTS = (GPT, QWEN27), ("first", "second"), 147


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


def page(row):
    if not isinstance(row.get("page"), str) or not re.fullmatch(r"s-\d\d", row["page"]):
        fail(f"a post has the number of the earlier records and no name: {row.get('page')}")
    return row["page"]


def check_revision(row, seen):
    key = (row.get("arm"), page(row))
    if row["arm"] not in ARMS or key in seen:
        fail(f"unknown way of revising, or a post listed twice: {key}")
    seen.add(key)
    if row.get("revised") is False:
        if set(row) != {"arm", "page", "revised"}:
            fail(f"a post without a revision has nothing else: {key}")
        return
    if row.get("revised") is not True or set(row) != {"arm", "page", "revised", "before", "after"}:
        fail(f"fields of a revision: {key}")
    for side in ("before", "after"):
        found = row[side]
        if set(found) != {"index", "closing_sentences", "register"} or found["register"] not in (0, 1):
            fail(f"what the rules reported {side}: {key}")
        if not isinstance(found["index"], int) or not 0 <= found["index"] <= 100:
            fail(f"index {side}: {key}")
        if not isinstance(found["closing_sentences"], int) or found["closing_sentences"] in (1, 2) or found["closing_sentences"] < 0:
            fail(f"closing sentences {side}: the rule lists none or more than two: {key}")


def check_answer(row, seen):
    key = (row.get("arm"), row.get("judge"), page(row), row.get("pair"))
    if row["arm"] not in ARMS + (AS_IT_IS,) or row["judge"] not in JUDGES or row["pair"] not in range(1, PAIRS + 1) or key in seen:
        fail(f"unknown text, judge, or pair, or a pair listed twice: {key}")
    if row["arm"] in ASKED and row["judge"] not in ASKED[row["arm"]]:
        fail(f"a judge that was not asked about this way of revising: {key}")
    seen.add(key)
    if row.get("named") not in NAMED or set(row) != {"arm", "judge", "page", "pair", "named"}:
        fail(f"what a judge named: {key}")


def check_round(row, seen):
    key = (page(row), row.get("round"), row.get("pair"))
    if row["round"] not in range(ROUNDS) or row["pair"] not in range(1, ROUND_PAIRS + 1) or key in seen:
        fail(f"unknown round or pair, or a judgment listed twice: {key}")
    seen.add(key)
    if row.get("named") not in NAMED or set(row) != {"page", "round", "pair", "named"}:
        fail(f"what the judge named in a round: {key}")


def check_subject(row, seen):
    key = (row.get("judge"), row.get("id"))
    if row["judge"] not in SUBJECT_JUDGES or row.get("writer") not in WRITERS or key in seen:
        fail(f"unknown judge or writer, or a pair listed twice: {key}")
    seen.add(key)
    if not isinstance(row["id"], str) or not re.fullmatch(r"[0-9a-f]{12}", row["id"]) or row.get("named") not in ("model", "human", "same"):
        fail(f"a pair of a subject: {key}")
    if set(row) != {"judge", "writer", "id", "named"}:
        fail(f"fields of a pair of a subject: {sorted(row)}")


def subject_counts(subjects):
    result = {}
    for judge in SUBJECT_JUDGES:
        rows = [row for row in subjects if row["judge"] == judge]
        if len(rows) != SUBJECTS:
            fail(f"{judge}: {len(rows)} pairs of a subject, and the record says {SUBJECTS}")
        result[judge] = {"pairs": len(rows), "named_model": sum(1 for row in rows if row["named"] == "model")}
    return result


def median(values):
    ordered = sorted(values)
    return ordered[len(ordered) // 2] if ordered else 0


def picks(answers, arm, judges, pages):
    """How many of the pages the judges together pick, and each of them alone.

    A post is picked when the votes for it outnumber the votes for the human posts.
    """
    def picked(rows):
        return sum(1 for name in pages
                   if sum(1 for row in rows if row["page"] == name and row["named"] == "post")
                   > sum(1 for row in rows if row["page"] == name and row["named"] == "human"))

    mine = [row for row in answers if row["arm"] == arm and row["judge"] in judges and row["page"] in pages]
    for judge in judges:
        found = sum(1 for row in mine if row["judge"] == judge)
        if found != PAIRS * len(pages):
            fail(f"{arm}, {judge}: {found} answers for {len(pages)} posts with {PAIRS} pairs each")
    return {"together": picked(mine), "alone": {judge: picked([row for row in mine if row["judge"] == judge]) for judge in judges},
            "pairs_named_post": {judge: sum(1 for row in mine if row["judge"] == judge and row["named"] == "post") for judge in judges}}


def arm_counts(revisions, answers, arm):
    rows = [row for row in revisions if row["arm"] == arm]
    if len(rows) != POSTS:
        fail(f"{arm}: {len(rows)} posts, and the fitting half has {POSTS}")
    done = [row for row in rows if row["revised"]]
    pages = sorted(row["page"] for row in done)
    judges = ASKED[arm]
    side = lambda name: {
        "posts_with_closing_denials": sum(1 for row in done if row[name]["closing_sentences"]),
        "closing_sentences": sum(row[name]["closing_sentences"] for row in done),
        "posts_in_one_register": sum(row[name]["register"] for row in done),
        "median_index": median([row[name]["index"] for row in done]),
    }
    return {"posts": len(rows), "revised": len(done), "judges": list(judges), "before": side("before"), "after": side("after"),
            "picked_as_it_is": picks(answers, AS_IT_IS, judges, pages), "picked_revised": picks(answers, arm, judges, pages)}


def round_counts(rounds):
    pages = sorted({row["page"] for row in rounds})
    for name in pages:
        if sum(1 for row in rounds if row["page"] == name) != ROUNDS * ROUND_PAIRS:
            fail(f"rounds: {name} does not have {ROUNDS} rounds of {ROUND_PAIRS} pairs")
    last = [row for row in rounds if row["round"] == ROUNDS - 1]
    return {"posts": len(pages), "judgments": len(rounds), "named_post": sum(1 for row in rounds if row["named"] == "post"),
            "posts_not_named_after_the_last_round": sum(1 for name in pages if all(row["named"] != "post" for row in last if row["page"] == name))}


def results(root=ROOT):
    revisions, answers, rounds, subjects = read(root, "revisions"), read(root, "answers"), read(root, "rounds"), read(root, "subjects")
    for rows, check in ((revisions, check_revision), (answers, check_answer), (rounds, check_round), (subjects, check_subject)):
        seen = set()
        for row in rows:
            check(row, seen)
    every = sorted({row["page"] for row in answers if row["arm"] == AS_IT_IS})
    if len(every) != POSTS:
        fail(f"{len(every)} posts were judged as they are, and the fitting half has {POSTS}")
    return {"judges": list(JUDGES), "as_it_is": picks(answers, AS_IT_IS, JUDGES, every),
            "arms": {arm: arm_counts(revisions, answers, arm) for arm in ARMS}, "rounds": round_counts(rounds),
            "subjects": subject_counts(subjects)}


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
    first = result["arms"]["closing-sentences"]
    print(f"revision pilots: the panel picks {first['picked_revised']['together']} of {first['revised']} revised posts, "
          f"{first['picked_as_it_is']['together']} as they were")
    return result


if __name__ == "__main__":
    try:
        main()
    except Malformed as error:
        sys.exit(f"verify: {error}")
