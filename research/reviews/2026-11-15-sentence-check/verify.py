#!/usr/bin/env python3
"""Recount the record of the sentence comparison.

usage: verify.py [--write]

The record holds no text. Its data are what `unswell compare` found without a model on damaged and on harmless
copies of pages, what the check did with the damaged copies under two judges, what it did with revisions, and how
the rule was fitted. This program recounts every number of results.json and of RESULTS.md from those data and fails
when one of them differs, when a row is malformed, or when a data file is not the one that files.sha256 names. With
--write it writes results.json, RESULTS.md, and files.sha256 again.
"""
import gzip
import hashlib
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
DATA = ("copies", "damage", "revisions", "variants")

GPT, QWEN35 = "gpt-6-astra", "private-host/Qwen3.6-35B-A3B"
JUDGES = (GPT, QWEN35)
# The judge of a revision is of the other family than its reviser.
CHECK_JUDGE = {GPT: QWEN35, QWEN35: GPT}
SOUNDS = ("better", "no difference", "worse")
NAMED = ("page", "revision", "same")
NOT_KEPT = ("changed", "missing", "unquoted")
REASONS = ("missing", "negation")
DAMAGE, HARMLESS = ("sentence", "negation"), ("join", "contract")
# The copies of the fitting half were made twice, with two salts. The measured copies are of the posts of the
# style-guide set and of texts that no earlier run had read.
COPY_PARTS = ("fitting, first copies", "fitting, second copies", "measured")
PAGES = ("set-s", "fresh")
# Revisions: of the fitting half, of the measured half as the earlier record checked them, and of the fresh texts.
REVISION_PARTS = ("fitting", "read", "fresh")
MEASURED_REVISIONS = ("read", "fresh")
# The bars, as they were written on issue 589 before the run.
BARS = {"damage_reported": 0.90, "damaged_accepted": 1, "harmless_suspected": 0.05, "statements_kept": 0.98, "verdicts": 0.95}


class Malformed(Exception):
    """A row of the data, or the record around it, is not what the record says it is."""


def fail(message):
    raise Malformed(message)


def read(root, name):
    path = root / "data" / (name + ".jsonl.gz")
    if not path.exists():
        fail(f"missing data file: {path.name}")
    with gzip.open(path, "rt", encoding="utf-8") as handle:
        return [json.loads(line) for line in handle]


def share(k, n):
    return round(k / n, 3) if n else None


def whole(row, *names):
    """The named fields of a row have to be whole numbers that are not negative."""
    for name in names:
        if not isinstance(row.get(name), int) or isinstance(row.get(name), bool) or row[name] < 0:
            fail(f"{name} is not a count: {row.get('unit')}")


def once(rows, what, *names):
    seen = set()
    for row in rows:
        key = tuple(row.get(name) for name in names)
        if key in seen:
            fail(f"{what} listed twice: {key}")
        seen.add(key)


def sound_of(named):
    if not (isinstance(named, list) and len(named) == 2 and all(value in NAMED for value in named)):
        fail(f"malformed answers to the two orders: {named}")
    return {("page", "page"): "better", ("revision", "revision"): "worse"}.get(tuple(named), "no difference")


# ---------------------------------------------------------------------------------------------------------------
# The comparison without a model
# ---------------------------------------------------------------------------------------------------------------


def copies_results(rows):
    """Copies of pages: on a damaged one, whether the changed sentence is a suspect; on a harmless one, whether any is."""
    once(rows, "copy", "part", "pages", "unit", "kind")
    result = {}
    for row in rows:
        if row.get("part") not in COPY_PARTS or row.get("pages") not in PAGES or row.get("kind") not in DAMAGE + HARMLESS:
            fail(f"copy of an unknown part, kind of pages, or kind: {row.get('unit')}")
        whole(row, "compared", "suspects")
        if row["kind"] in DAMAGE and not isinstance(row.get("named"), bool):
            fail(f"damaged copy that does not say whether the sentence is a suspect: {row.get('unit')}")
        if row["kind"] in DAMAGE and row["named"] and row["suspects"] == 0:
            fail(f"damaged copy whose sentence is a suspect among none: {row.get('unit')}")
        entry = result.setdefault(row["part"], {}).setdefault(row["pages"], {}).setdefault(row["kind"], {"copies": 0, "named": 0, "with_a_suspect": 0})
        entry["copies"] += 1
        entry["named"] += bool(row.get("named"))
        entry["with_a_suspect"] += row["suspects"] > 0
    for part in result.values():
        for pages in part.values():
            damaged = [pages[kind] for kind in DAMAGE if kind in pages]
            harmless = [pages[kind] for kind in HARMLESS if kind in pages]
            pages["damaged"] = {"copies": sum(e["copies"] for e in damaged), "named": sum(e["named"] for e in damaged)}
            pages["damaged"]["named_share"] = share(pages["damaged"]["named"], pages["damaged"]["copies"])
            pages["harmless"] = {"copies": sum(e["copies"] for e in harmless), "with_a_suspect": sum(e["with_a_suspect"] for e in harmless)}
            pages["harmless"]["suspected_share"] = share(pages["harmless"]["with_a_suspect"], pages["harmless"]["copies"])
    return result


def variants_results(rows):
    """The rule as it grew on the fitting half: each variant on the first damaged copies and on ordinary revisions."""
    once(rows, "variant row", "variant", "kind", "reviser", "unit")
    order, result = [], {}
    for row in rows:
        if row.get("kind") not in DAMAGE + ("revision",) or not isinstance(row.get("variant"), str):
            fail(f"malformed variant row: {row.get('unit')}")
        whole(row, "suspects")
        if row["variant"] not in result:
            order.append(row["variant"])
            result[row["variant"]] = {"damaged": {kind: {"copies": 0, "named": 0} for kind in DAMAGE}, "revisions": {}}
        entry = result[row["variant"]]
        if row["kind"] == "revision":
            if row.get("reviser") not in JUDGES:
                fail(f"variant row of an unknown reviser: {row.get('unit')}")
            mine = entry["revisions"].setdefault(row["reviser"], {"revisions": 0, "with_a_suspect": 0, "suspects": 0})
            mine["revisions"] += 1
            mine["with_a_suspect"] += row["suspects"] > 0
            mine["suspects"] += row["suspects"]
        else:
            if not isinstance(row.get("named"), bool):
                fail(f"variant row that does not say whether the sentence is a suspect: {row.get('unit')}")
            entry["damaged"][row["kind"]]["copies"] += 1
            entry["damaged"][row["kind"]]["named"] += row["named"]
    return [{"variant": name, **result[name]} for name in order]


# ---------------------------------------------------------------------------------------------------------------
# The check with a judge
# ---------------------------------------------------------------------------------------------------------------


def checked(row, what):
    """Read the parts of one verdict and see that they agree. Return whether the row has a verdict."""
    if row.get("judge") not in JUDGES or row.get("verdict") not in ("accepted", "rejected", "no verdict"):
        fail(f"{what} with an unknown judge or verdict: {row.get('unit')}")
    if row["verdict"] == "no verdict":
        return False
    if row.get("sound") not in SOUNDS or row["sound"] != sound_of(row.get("named")):
        fail(f"{what} whose sound does not follow from the two answers: {row.get('unit')}")
    whole(row, "statements", "kept", "suspects", "lost")
    if row["kept"] > row["statements"] or row["lost"] > row["suspects"]:
        fail(f"{what} whose counts do not add up: {row.get('unit')}")
    passes = row["sound"] == "better" and row["kept"] >= BARS["statements_kept"] * row["statements"] and row["statements"] > 0 and row["lost"] == 0
    if row["verdict"] == "accepted" and not passes:
        fail(f"{what} whose verdict does not follow from its parts: {row.get('unit')}")
    return True


def damage_results(rows):
    """Damaged copies under a judge: on how many the check reports the damaged sentence, and how many it accepts."""
    once(rows, "damaged copy", "part", "pages", "unit", "kind", "judge")
    result = {}
    for row in rows:
        if row.get("part") not in COPY_PARTS or row.get("pages") not in PAGES or row.get("kind") not in DAMAGE:
            fail(f"damaged copy of an unknown part, kind of pages, or kind: {row.get('unit')}")
        answered = checked(row, "damaged copy")
        entry = result.setdefault(row["part"], {}).setdefault(row["pages"], {}).setdefault(row["judge"], {
            "copies": 0, "verdicts": 0, "accepted": 0, "reported": 0, "by_statements": 0, "by_either": 0, "sound": dict.fromkeys(SOUNDS, 0),
            **{kind: {"copies": 0, "reported": 0} for kind in DAMAGE}})
        entry["copies"] += 1
        entry[row["kind"]]["copies"] += 1
        if not answered:
            continue
        if not isinstance(row.get("reported"), bool) or (row["reported"] and row["lost"] == 0):
            fail(f"damaged copy whose report does not follow from its sentences: {row.get('unit')}")
        statements = row["kept"] < row["statements"]
        entry["verdicts"] += 1
        entry["accepted"] += row["verdict"] == "accepted"
        entry["reported"] += row["reported"]
        entry[row["kind"]]["reported"] += row["reported"]
        entry["by_statements"] += statements
        entry["by_either"] += row["reported"] or statements
        entry["sound"][row["sound"]] += 1
    for part in result.values():
        for pages in part.values():
            for entry in pages.values():
                entry["verdict_share"] = share(entry["verdicts"], entry["copies"])
                entry["reported_share"] = share(entry["reported"], entry["copies"])
    return result


def revisions_results(rows):
    """Revisions under the judge of the other family: the statements, the suspects, and what a sentence alone rejects."""
    once(rows, "revision", "part", "pages", "unit", "reviser")
    result = {}
    for row in rows:
        if row.get("part") not in REVISION_PARTS or row.get("reviser") not in JUDGES:
            fail(f"revision of an unknown part or reviser: {row.get('unit')}")
        answered = checked(row, "revision")
        if row["judge"] != CHECK_JUDGE[row["reviser"]]:
            fail(f"revision whose judge is not of the other family: {row.get('unit')}")
        entry = result.setdefault(row["part"], {}).setdefault(row["reviser"], {
            "judge": row["judge"], "pairs": 0, "verdicts": 0, "accepted": 0, "sound": dict.fromkeys(SOUNDS, 0), "statements": 0, "kept": 0,
            "not_kept": dict.fromkeys(NOT_KEPT, 0), "under_the_bar": 0, "suspects": 0, "cleared": 0, "with_a_suspect": 0,
            "with_a_lost_sentence": 0, "lost": dict.fromkeys(REASONS, 0), "rejected_by_sentences_alone": 0})
        entry["pairs"] += 1
        if not answered:
            continue
        not_kept, lost = row.get("not_kept"), row.get("lost_by")
        if not (isinstance(not_kept, dict) and set(not_kept) == set(NOT_KEPT) and sum(not_kept.values()) == row["statements"] - row["kept"]):
            fail(f"revision whose statements do not add up: {row.get('unit')}")
        if not (isinstance(lost, dict) and set(lost) == set(REASONS) and sum(lost.values()) == row["lost"]):
            fail(f"revision whose lost sentences do not add up: {row.get('unit')}")
        enough = row["kept"] >= BARS["statements_kept"] * row["statements"]
        entry["verdicts"] += 1
        entry["accepted"] += row["verdict"] == "accepted"
        entry["sound"][row["sound"]] += 1
        entry["statements"] += row["statements"]
        entry["kept"] += row["kept"]
        entry["under_the_bar"] += not enough
        entry["suspects"] += row["suspects"]
        entry["cleared"] += row["suspects"] - row["lost"]
        entry["with_a_suspect"] += row["suspects"] > 0
        entry["with_a_lost_sentence"] += row["lost"] > 0
        entry["rejected_by_sentences_alone"] += row["lost"] > 0 and enough
        for kind in NOT_KEPT:
            entry["not_kept"][kind] += not_kept[kind]
        for reason in REASONS:
            entry["lost"][reason] += lost[reason]
    for part in result.values():
        for entry in part.values():
            entry["verdict_share"] = share(entry["verdicts"], entry["pairs"])
            entry["kept_share"] = share(entry["kept"], entry["statements"])
    return result


# ---------------------------------------------------------------------------------------------------------------
# The whole record
# ---------------------------------------------------------------------------------------------------------------


def results(root):
    result = {"bars": BARS, "copies": copies_results(read(root, "copies")), "variants": variants_results(read(root, "variants")),
              "damage": damage_results(read(root, "damage")), "revisions": revisions_results(read(root, "revisions"))}
    result["registered"] = registered(result)
    return result


def registered(result):
    """Read each registered bar on the measured part. A cell without data is malformed: every bar has to be read."""
    held = {}
    copies, damage, revisions = result["copies"].get("measured", {}), result["damage"].get("measured", {}), result["revisions"]
    for pages in PAGES:
        if pages not in copies or pages not in damage or set(damage[pages]) != set(JUDGES):
            fail(f"the measured copies of {pages} are not complete")
        held[f"H, harmless copies, {pages}"] = copies[pages]["harmless"]["suspected_share"] <= BARS["harmless_suspected"]
        for judge in JUDGES:
            entry = damage[pages][judge]
            held[f"D1, damage reported, {pages}, {judge}"] = entry["reported_share"] >= BARS["damage_reported"]
            held[f"D2, damaged copies accepted, {pages}, {judge}"] = entry["accepted"] <= BARS["damaged_accepted"]
            held[f"V, verdicts on damaged copies, {pages}, {judge}"] = entry["verdict_share"] >= BARS["verdicts"]
    for part in MEASURED_REVISIONS:
        if set(revisions.get(part, {})) != set(JUDGES):
            fail(f"the revisions of the part {part} are not complete")
        for reviser in JUDGES:
            entry = revisions[part][reviser]
            held[f"R, statements kept, {part}, {reviser}"] = (entry["kept_share"] or 0) >= BARS["statements_kept"]
            held[f"V, verdicts on revisions, {part}, {reviser}"] = entry["verdict_share"] >= BARS["verdicts"]
    return held


def digests(root):
    return "".join(f"{hashlib.sha256((root / 'data' / (name + '.jsonl.gz')).read_bytes()).hexdigest()}  data/{name}.jsonl.gz\n" for name in DATA)


def verify(root, write=False):
    """Recount the record. Return the results, or raise Malformed."""
    import render  # The page is rendered beside this file, so that a test can render from changed results.

    computed = results(root)
    text = json.dumps(computed, indent=2, sort_keys=True) + "\n"
    page = render.page(computed)
    if write:
        (root / "results.json").write_text(text)
        (root / "RESULTS.md").write_text(page)
        (root / "files.sha256").write_text(digests(root))
        return computed
    if (root / "files.sha256").read_text() != digests(root):
        fail("a data file is not the one that files.sha256 names")
    if (root / "results.json").read_text() != text:
        fail("results.json differs from the recount")
    if (root / "RESULTS.md").read_text() != page:
        fail("RESULTS.md differs from the recount")
    return computed


def main():
    sys.path.insert(0, str(ROOT))
    try:
        computed = verify(ROOT, write="--write" in sys.argv[1:])
    except Malformed as problem:
        print(f"sentence check: {problem}", file=sys.stderr)
        return 1
    held = computed["registered"]
    print(f"sentence check: {sum(held.values())} of {len(held)} registered bars hold")
    return 0


if __name__ == "__main__":
    sys.exit(main())
