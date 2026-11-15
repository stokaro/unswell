#!/usr/bin/env python3
"""Recount the record of the revision check.

usage: verify.py [--write]

The record holds no text. Its data are the answers of the judges on what a writer who follows the product publishes,
the verdict of `unswell compare` on each revision with what it rests on, what the check did with damaged copies and
with pairs whose difference is known, and what the fitting half showed. This program recounts every number of
results.json and of RESULTS.md from those data and fails when one of them differs, when a row is malformed, or when
a data file is not the one that files.sha256 names. With --write it writes results.json, RESULTS.md, and
files.sha256 again.
"""
import gzip
import hashlib
import json
import math
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
DATA = ("published", "revisions", "damage", "known", "fitting")

GPT, QWEN35, QWEN27 = "gpt-6-astra", "private-host/Qwen3.6-35B-A3B", "private-host/Qwen3.8-27B"
REVISERS = (GPT, QWEN35)
JUDGES = (GPT, QWEN27, QWEN35)
# The judge of the check is of the other family than the reviser. The third model revised no page and judges in no check.
CHECK_JUDGE = {GPT: QWEN35, QWEN35: GPT}
THIRD = QWEN27
# The judge that picked every post of the set before any revision.
STRICT = GPT
# The judge of the check on damaged copies, as the bar names it. A second judge was added after the bars were written.
DAMAGE_JUDGE = QWEN35
HUMAN = ("blog", "styled")
SOUNDS = ("better", "no difference", "worse")
# A statement that is not kept: the revision says something else, does not say it, or the judge quoted words that are not there.
NOT_KEPT = ("changed", "missing", "unquoted")
NAMED = ("page", "revision", "same", "unreadable")
ELEVATED, HIGH = 40, 70
# The bars, as they were written on issue 583 before the run.
BARS = {"panel": 0.55, "strict_judge": 0.65, "verdicts": 0.95, "accepted_better": 0.70, "fewest_accepted": 10,
        "em_dash_revisions": 0.05, "statements_kept": 0.98, "damage_noticed": 0.80, "known_difference": 0.60}


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


def median(values):
    values = sorted(values)
    if not values:
        return None
    middle = len(values) // 2
    return values[middle] if len(values) % 2 else (values[middle - 1] + values[middle]) / 2


def wilson(k, n):
    """The 95% Wilson interval of k of n, as two shares rounded to three places."""
    if n == 0:
        return None
    p, z = k / n, 1.96
    center = (p + z * z / (2 * n)) / (1 + z * z / n)
    half = z * math.sqrt(p * (1 - p) / n + z * z / (4 * n * n)) / (1 + z * z / n)
    return [round(center - half, 3), round(center + half, 3)]


def count(values, names):
    """Count each name; a value outside the names is malformed."""
    counted = dict.fromkeys(names, 0)
    for value in values:
        if value not in counted:
            fail(f"unknown value: {value}")
        counted[value] += 1
    return counted


# ---------------------------------------------------------------------------------------------------------------
# A post against human text
# ---------------------------------------------------------------------------------------------------------------

POST_VOTES = {"post": 1, "human": -1, "same": 0, "unreadable": 0}


def panel(votes_by_unit, judges=JUDGES):
    """Sum the votes of each unit: for the post +1, for the human text -1. A tie counts against the rate."""
    post = human = 0
    for votes in votes_by_unit.values():
        total = sum(value for judge, value in votes if judge in judges)
        post += total > 0
        human += total < 0
    units = len(votes_by_unit)
    return {"units": units, "post": post, "human": human, "tie": units - post - human, "rate": share(post, units),
            "interval": wilson(post, units)}


def judged_units(rows, what):
    """Group answers on pairs of a post and a human text by post, and require one answer of each judge on every pair."""
    units, pairs = {}, {}
    for row in rows:
        if row.get("judge") not in JUDGES or row.get("vote") not in POST_VOTES:
            fail(f"malformed answer on {what}: {row.get('unit')}")
        pairs.setdefault((row["unit"], row.get("against")), []).append(row["judge"])
        units.setdefault(row["unit"], []).append((row["judge"], POST_VOTES[row["vote"]]))
    for pair, judges in pairs.items():
        if sorted(judges) != sorted(JUDGES):
            fail(f"pair without one answer of each judge on {what}: {pair}")
    return units


def against_human(rows, what):
    units = judged_units(rows, what)
    return {"pairs": len(rows) // len(JUDGES), "panel": panel(units), "judges": {judge: panel(units, (judge,)) for judge in JUDGES},
            "unreadable": sum(row["vote"] == "unreadable" for row in rows)}


def published_results(rows):
    """A post as it is published. With kept "every revision" the revision stands in whatever the check said."""
    result = {"every revision": {}}
    for row in rows:
        if row.get("reviser") not in ("none",) + REVISERS or row.get("human") not in HUMAN or row.get("kept") not in (None, "every revision"):
            fail(f"published answer without a known reviser or kind of human text: {row.get('unit')}")
    checked = [row for row in rows if "kept" not in row]
    for reviser in ("none",) + REVISERS:
        result[reviser] = {}
        for human in HUMAN:
            mine = [row for row in checked if row["reviser"] == reviser and row["human"] == human]
            if not mine:
                fail(f"no published answers for {reviser} against {human}")
            result[reviser][human] = against_human(mine, f"{reviser} against {human}")
    for reviser in REVISERS:
        mine = [row for row in rows if row.get("kept") == "every revision" and row["reviser"] == reviser]
        if mine:
            result["every revision"][reviser] = against_human(mine, f"every revision of {reviser}")
    return result


# ---------------------------------------------------------------------------------------------------------------
# The revisions and the check
# ---------------------------------------------------------------------------------------------------------------


def checked_revisions(rows):
    seen = set()
    for row in rows:
        key = (row.get("reviser"), row.get("set"), row.get("writer"), row.get("id"))
        if row.get("reviser") not in REVISERS or row.get("set") not in ("set-s", "guided"):
            fail(f"revision without a known reviser or set: {key}")
        if key in seen:
            fail(f"revision listed twice: {key}")
        seen.add(key)
        if row.get("outcome") not in ("revised", "no mark listed", "no revision"):
            fail(f"revision with an unknown outcome: {key}")
        if row["outcome"] != "revised":
            if set(row) - {"reviser", "set", "writer", "id", "outcome"}:
                fail(f"page that was not revised with more than its outcome: {key}")
            continue
        for field in ("index", "em_dashes"):
            pair = row.get(field)
            if not (isinstance(pair, list) and len(pair) == 2 and all(isinstance(v, int) and v >= 0 for v in pair)):
                fail(f"revision without its {field} before and after: {key}")
        if row.get("third") not in NAMED:
            fail(f"revision without the answer of the third model: {key}")
        if "check" in row:
            checked_verdict(row, key)
    return rows


def checked_verdict(row, key):
    check = row["check"]
    if check.get("judge") != CHECK_JUDGE[row["reviser"]]:
        fail(f"check whose judge is not of the other family: {key}")
    if check.get("verdict") == "no verdict":
        if set(check) - {"judge", "verdict"}:
            fail(f"check without a verdict that holds more: {key}")
        return
    if check.get("verdict") not in ("accepted", "rejected") or check.get("sound") not in SOUNDS:
        fail(f"check with an unknown verdict or sound: {key}")
    named = check.get("named")
    if not (isinstance(named, list) and len(named) == 2 and all(value in NAMED for value in named)):
        fail(f"check without what the judge named in each order: {key}")
    if check["sound"] != sound_of(named):
        fail(f"check whose sound does not follow from the two answers: {key}")
    statements, kept, others = check.get("statements"), check.get("kept"), check.get("not_kept")
    if not (isinstance(statements, int) and isinstance(kept, int) and 0 <= kept <= statements):
        fail(f"check without its statements: {key}")
    if not (isinstance(others, dict) and set(others) == set(NOT_KEPT) and kept + sum(others.values()) == statements):
        fail(f"check whose statements do not add up: {key}")
    lost = check.get("lost")
    if not (isinstance(lost, dict) and set(lost) == {"numbers", "code", "links"} and all(isinstance(v, int) and v >= 0 for v in lost.values())):
        fail(f"check without what was lost: {key}")
    if (check["verdict"] == "accepted") != accepted(check, row["em_dashes"]):
        fail(f"check whose verdict does not follow from its parts: {key}")


def sound_of(named):
    """The revision is better only when the judge names the page in both orders, and worse only when it names the revision."""
    if named == ["page", "page"]:
        return "better"
    if named == ["revision", "revision"]:
        return "worse"
    return "no difference"


def accepted(check, em_dashes):
    """The rule of the command: the sound is better, the statements are kept, nothing is lost, and no em dash is added."""
    kept_enough = check["statements"] > 0 and check["kept"] >= BARS["statements_kept"] * check["statements"]
    return check["sound"] == "better" and kept_enough and not any(check["lost"].values()) and em_dashes[1] <= em_dashes[0]


def third_votes(rows):
    counted = count((row["third"] for row in rows), NAMED)
    decisive = counted["page"] + counted["revision"]
    return {**counted, "decisive": decisive, "page_share": share(counted["page"], decisive)}


def index_figures(rows):
    """The index of a group of revised pages before and after."""
    return {"revised": len(rows), "index_lower": sum(row["index"][1] < row["index"][0] for row in rows),
            "median_index": [median(row["index"][0] for row in rows), median(row["index"][1] for row in rows)],
            "high": [sum(row["index"][0] >= HIGH for row in rows), sum(row["index"][1] >= HIGH for row in rows)]}


def reviser_results(rows):
    revised = [row for row in rows if row["outcome"] == "revised"]
    checked = [row for row in revised if "check" in row]
    verdicts = [row for row in checked if row["check"]["verdict"] != "no verdict"]
    statements = sum(row["check"]["statements"] for row in verdicts)
    kept = sum(row["check"]["kept"] for row in verdicts)
    adding = [row for row in revised if row["em_dashes"][1] > row["em_dashes"][0]]
    by_verdict = lambda name: [row for row in verdicts if row["check"]["verdict"] == name]
    return {
        "pages": len(rows), "revised": len(revised), "no_mark_listed": sum(row["outcome"] == "no mark listed" for row in rows),
        "no_revision": sum(row["outcome"] == "no revision" for row in rows),
        "index_lower": sum(row["index"][1] < row["index"][0] for row in revised),
        "median_index": [median(row["index"][0] for row in revised), median(row["index"][1] for row in revised)],
        "high": [sum(row["index"][0] >= HIGH for row in revised), sum(row["index"][1] >= HIGH for row in revised)],
        "add_em_dash": len(adding), "add_em_dash_share": share(len(adding), len(revised)),
        "em_dashes_added": sum(row["em_dashes"][1] - row["em_dashes"][0] for row in adding),
        "third": third_votes(revised),
        "set_s": index_figures([row for row in revised if row["set"] == "set-s"]),
        "check": {
            "judge": CHECK_JUDGE[rows[0]["reviser"]] if rows else None, "pairs": len(checked), "verdicts": len(verdicts),
            "verdict_share": share(len(verdicts), len(checked)), "accepted": len(by_verdict("accepted")), "rejected": len(by_verdict("rejected")),
            "accepted_set_s": sum(row["set"] == "set-s" for row in by_verdict("accepted")),
            "sound": count((row["check"]["sound"] for row in verdicts), SOUNDS),
            "second_text_both_times": sum(row["check"]["named"] == ["revision", "page"] for row in verdicts),
            "first_text_both_times": sum(row["check"]["named"] == ["page", "revision"] for row in verdicts),
            "statements": statements, "kept": kept, "kept_share": share(kept, statements),
            "not_kept": {kind: sum(row["check"]["not_kept"][kind] for row in verdicts) for kind in NOT_KEPT},
            "under_the_bar": sum(row["check"]["kept"] < BARS["statements_kept"] * row["check"]["statements"] for row in verdicts),
            "lost_something": sum(any(row["check"]["lost"].values()) for row in verdicts),
            "accepted_add_em_dash": sum(row["em_dashes"][1] > row["em_dashes"][0] for row in by_verdict("accepted")),
            "third_on_accepted": third_votes(by_verdict("accepted")), "third_on_rejected": third_votes(by_verdict("rejected")),
        },
    }


def revisions_results(rows):
    result = {reviser: reviser_results([row for row in rows if row["reviser"] == reviser]) for reviser in REVISERS}
    verdicts = [row for row in rows if row["outcome"] == "revised" and row.get("check", {}).get("verdict") in ("accepted", "rejected")]
    result["together"] = {"accepted": sum(row["check"]["verdict"] == "accepted" for row in verdicts),
                          "third_on_accepted": third_votes([row for row in verdicts if row["check"]["verdict"] == "accepted"]),
                          "third_on_rejected": third_votes([row for row in verdicts if row["check"]["verdict"] == "rejected"])}
    return result


# ---------------------------------------------------------------------------------------------------------------
# Damage on purpose, and a difference that is known
# ---------------------------------------------------------------------------------------------------------------


def noticed(rows, what):
    for row in rows:
        if row.get("kind") not in ("sentence", "negation") or row.get("verdict") not in ("accepted", "rejected", "no verdict"):
            fail(f"malformed damaged copy in {what}: {row.get('unit')}")
        if row["verdict"] != "no verdict" and not isinstance(row.get("not_kept"), int):
            fail(f"damaged copy without its statements in {what}: {row.get('unit')}")
    found = lambda group: sum(row["verdict"] != "no verdict" and row["not_kept"] > 0 for row in group)
    result = {"copies": len(rows), "no_verdict": sum(row["verdict"] == "no verdict" for row in rows), "noticed": found(rows),
              "noticed_share": share(found(rows), len(rows)), "accepted": sum(row["verdict"] == "accepted" for row in rows)}
    for kind in ("sentence", "negation"):
        group = [row for row in rows if row["kind"] == kind]
        result[kind] = {"copies": len(group), "noticed": found(group)}
    return result


def damage_results(rows):
    """Damaged copies of the measured half, for each judge that checked them. The bar reads the judge that it names."""
    seen = set()
    for row in rows:
        key = (row.get("judge"), row.get("unit"), row.get("kind"))
        if row.get("judge") not in JUDGES:
            fail(f"damaged copy without a known judge: {row.get('unit')}")
        if key in seen:
            fail(f"damaged copy listed twice: {key}")
        seen.add(key)
    judges = sorted({row["judge"] for row in rows}, key=JUDGES.index)
    if DAMAGE_JUDGE not in judges:
        fail("no damaged copy was checked by the judge of the bar")
    result = {}
    for judge in judges:
        mine = [row for row in rows if row["judge"] == judge]
        result[judge] = noticed(mine, f"the measured half, {judge}")
        answered = [row for row in mine if row["verdict"] != "no verdict"]
        result[judge]["sound"] = count((damaged_sound(row) for row in answered), SOUNDS)
        result[judge]["better"] = {kind: sum(row["sound"] == "better" for row in answered if row["kind"] == kind) for kind in ("sentence", "negation")}
    return result


def damaged_sound(row):
    """The sound of a damaged copy against its post. The two texts differ by one sentence."""
    sound = row.get("sound")
    if sound not in SOUNDS or sound != sound_of(row.get("named")):
        fail(f"damaged copy whose sound does not follow from the two answers: {row.get('unit')}")
    kept = row.get("statements", 0) - row["not_kept"]
    if row["verdict"] == "accepted" and (sound != "better" or kept < BARS["statements_kept"] * row.get("statements", 0)):
        fail(f"damaged copy whose verdict does not follow from the sound and the statements: {row.get('unit')}")
    return sound


def both_orders(rows, what):
    """Count what a judge named in the two orders of each pair."""
    patterns = dict.fromkeys(("page in both orders", "revision in both orders", "second text both times", "first text both times", "other"), 0)
    for row in rows:
        named = row.get("named")
        if not (isinstance(named, list) and len(named) == 2 and all(value in NAMED for value in named)):
            fail(f"pair without what the judge named in each order in {what}: {row.get('unit', row.get('id'))}")
        key = {("page", "page"): "page in both orders", ("revision", "revision"): "revision in both orders",
               ("revision", "page"): "second text both times", ("page", "revision"): "first text both times"}.get(tuple(named), "other")
        patterns[key] += 1
    return {"pairs": len(rows), **patterns}


def known_results(rows):
    for row in rows:
        if row.get("sound") not in SOUNDS + ("no verdict",):
            fail(f"pair with a known difference and an unknown sound: {row.get('id')}")
        if row["sound"] != "no verdict" and row["sound"] != sound_of(row.get("named")):
            fail(f"pair with a known difference whose sound does not follow from the two answers: {row.get('id')}")
    answered = [row for row in rows if row["sound"] != "no verdict"]
    better = sum(row["sound"] == "better" for row in answered)
    return {"pairs": len(rows), "no_verdict": len(rows) - len(answered), "better": better, "better_share": share(better, len(rows)),
            "orders": both_orders(answered, "the known difference")}


# ---------------------------------------------------------------------------------------------------------------
# The fitting half
# ---------------------------------------------------------------------------------------------------------------

SIDE_VOTES = {"page": 1, "revision": -1, "same": 0, "unreadable": 0}


def keyed(rows, *names):
    groups = {}
    for row in rows:
        groups.setdefault(tuple(row.get(name) for name in names), []).append(row)
    return groups


def page_against_revision(rows):
    """Answers on a page and its revision, one order for each pair. The panel sums the judges that answered."""
    judges = sorted({row["judge"] for row in rows}, key=JUDGES.index)
    totals = {}
    for row in rows:
        if row.get("vote") not in SIDE_VOTES:
            fail(f"malformed answer on a page and its revision: {row.get('unit')}")
        totals[row["unit"]] = totals.get(row["unit"], 0) + SIDE_VOTES[row["vote"]]
    page, revision = sum(total > 0 for total in totals.values()), sum(total < 0 for total in totals.values())
    return {"pairs": len(totals), "judges": {judge: count((row["vote"] for row in rows if row["judge"] == judge), NAMED) for judge in judges},
            "panel": {"page": page, "revision": revision, "tie": len(totals) - page - revision}}


def revision_figures(rows):
    adding = [row for row in rows if row["em_dashes"][1] > row["em_dashes"][0]]
    return {"revised": len(rows), "median_index": [median(row["index"][0] for row in rows), median(row["index"][1] for row in rows)],
            "high": [sum(row["index"][0] >= HIGH for row in rows), sum(row["index"][1] >= HIGH for row in rows)],
            "add_em_dash": len(adding), "em_dashes_added": sum(row["em_dashes"][1] - row["em_dashes"][0] for row in adding)}


def sections(rows):
    net = [sum(row["votes"]) for row in rows]
    votes = [vote for row in rows for vote in row["votes"]]
    return {"pairs": len(rows), "sections": len(votes), "for_the_revision": votes.count(1), "against": votes.count(-1), "unsure": votes.count(0),
            "better": sum(n > 0 for n in net), "worse": sum(n < 0 for n in net), "no_difference": sum(n == 0 for n in net)}


def fitting_results(rows):
    experiments = keyed(rows, "experiment")
    known = set(experiments) - {("against human text",), ("page against revision",), ("revision",), ("both orders",), ("sections",), ("damage",)}
    if known:
        fail(f"unknown experiment of the fitting half: {sorted(known)}")
    of = lambda name: experiments.get((name,), [])
    result = {"against_human": [], "page_against_revision": [], "revisions": [], "both_orders": [], "sections": []}
    for (variant, reviser, human), group in keyed(of("against human text"), "variant", "reviser", "human").items():
        result["against_human"].append({"variant": variant, "reviser": reviser, "human": human, **against_human(group, f"{variant}, fitting half")})
    for (variant, reviser), group in keyed(of("page against revision"), "variant", "reviser").items():
        result["page_against_revision"].append({"variant": variant, "reviser": reviser, **page_against_revision(group)})
    for (variant, reviser), group in keyed(of("revision"), "variant", "reviser").items():
        result["revisions"].append({"variant": variant, "reviser": reviser, **revision_figures(group)})
    for (pairs, wording, judge), group in keyed(of("both orders"), "pairs", "wording", "judge").items():
        result["both_orders"].append({"pairs_of": pairs, "wording": wording, "judge": judge, **both_orders(group, "the fitting half")})
    for (pairs, judge), group in keyed(of("sections"), "pairs", "judge").items():
        result["sections"].append({"pairs_of": pairs, "judge": judge, **sections(group)})
    result["damage"] = noticed(of("damage"), "the fitting half")
    return result


# ---------------------------------------------------------------------------------------------------------------
# The whole record
# ---------------------------------------------------------------------------------------------------------------


def results(root):
    revisions = checked_revisions(read(root, "revisions"))
    result = {"bars": BARS, "published": published_results(read(root, "published")), "revisions": revisions_results(revisions),
              "damage": damage_results(read(root, "damage")), "known": known_results(read(root, "known")),
              "fitting": fitting_results(read(root, "fitting"))}
    result["registered"] = registered(result)
    return result


def registered(result):
    """Read each registered bar. A bar that does not apply is None."""
    held = {}
    for reviser in REVISERS:
        for human in HUMAN:
            entry = result["published"][reviser][human]
            held[f"A1, panel, {reviser}, {human}"] = entry["panel"]["rate"] <= BARS["panel"]
            held[f"A2, {STRICT} alone, {reviser}, {human}"] = entry["judges"][STRICT]["rate"] <= BARS["strict_judge"]
    for reviser in REVISERS:
        mine = result["revisions"][reviser]
        check = mine["check"]
        held[f"B1, verdicts, {reviser}"] = (check["verdict_share"] or 0) >= BARS["verdicts"]
        held[f"C1, em dashes, {reviser}"] = (mine["add_em_dash_share"] or 0) <= BARS["em_dash_revisions"] and check["accepted_add_em_dash"] == 0
        held[f"C2, statements kept, {reviser}"] = (check["kept_share"] or 0) >= BARS["statements_kept"]
    together = result["revisions"]["together"]
    on_accepted, on_rejected = together["third_on_accepted"]["page_share"], together["third_on_rejected"]["page_share"]
    enough = together["accepted"] >= BARS["fewest_accepted"] and on_accepted is not None
    held["B2, accepted revisions are better"] = on_accepted >= BARS["accepted_better"] if enough else None
    held["B3, rejected revisions are less often better"] = on_rejected < on_accepted if enough and on_rejected is not None else None
    held["C3, damage is noticed"] = (result["damage"][DAMAGE_JUDGE]["noticed_share"] or 0) >= BARS["damage_noticed"]
    held["D, a known difference is found"] = (result["known"]["better_share"] or 0) >= BARS["known_difference"]
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
        computed = verify(ROOT, write="--write" in sys.argv)
    except Malformed as problem:
        print(f"verify: {problem}", file=sys.stderr)
        return 1
    held = computed["registered"]
    applied = [value for value in held.values() if value is not None]
    print(f"revision check: {sum(applied)} of {len(applied)} registered bars hold, {len(held) - len(applied)} do not apply")
    return 0


if __name__ == "__main__":
    sys.exit(main())
