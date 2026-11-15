#!/usr/bin/env python3
"""Recount the record of text that a model wrote under a style guide.

usage: verify.py [--write]

The record holds no text. Its data are what the build counted on each page, the answers of the judges on the
measured half, the index of each revised page before and after, the pages of the earlier record of the index
as the build reads them, and the pages of a control: text that people wrote under a style guide. This program recounts every number of results.json and of RESULTS.md from those data and
fails when one of them differs, when a row is malformed, or when a data file is not the one that files.sha256
names. With --write it writes results.json, RESULTS.md, and files.sha256 again.
"""
import gzip
import hashlib
import json
import math
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
DATA = ("pages", "judgments", "revisions", "earlier", "control")

# The two readings of the index, as the product computes them. A Go test compares the product with these values
# through results.json.
MARKS = {"em-dash": 0.94, "emoji-in-heading": 0.69, "emoji-at-list-item": 0.71, "arrow": 0.81, "stock-word": 0.62,
         "bold-label-in-list": 0.44}
MARKS_BIAS = -1.23
RHYTHM = {"plain-sentence-run": 3.2, "subject-opener-run": 0.35}
RHYTHM_BIAS = -12.3
ELEVATED, HIGH = 40, 70

JUDGES = ("gpt-6-astra", "private-host/Qwen3.8-27B", "private-host/Qwen3.6-35B-A3B")
WRITERS = ("gpt-6-astra", "private-host/Qwen3.6-35B-A3B")
SETS = ("human-blog", "readme", "set-s", "guided", "plain")
HALVES = ("fitting", "measured")
# The bars, as they were written on issue 577 before each run.
BARS = {"admission": 0.70, "tells": 0.65, "separation": 0.75, "quiet": 0.05, "earlier": 0.80, "index_lower": 0.90,
        "control_quiet": 0.05, "control_separation": 0.75}

# The control. A page is matched when at least this share of its prose words stands in paragraphs: the tenth
# percentile of the posts of the style-guide set.
KINDS = ("articles", "documentation")
MATCH = 0.732
FEWEST_SENTENCES = 10
# What the sentences of the posts of the style-guide set are like, in the measured half. These figures are counted on
# the posts. The record holds nothing of a post but its readings, so they are stated here and not recounted.
SET_S_SENTENCES = {"posts": 30, "matched": 27, "words_per_sentence": 12.5, "over_20_words": 0.079, "plain": 0.681,
                   "plain_lower_quartile": 0.618, "in_runs": 0.583, "by_chance": 0.52, "runs_to_chance": 1.07,
                   "paragraph_sentences_per_thousand_words": 71.8, "runs_per_thousand_words": 42.3,
                   "auc_words_per_sentence": 0.933, "auc_plain": 0.9}


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


def index_of(score):
    return round(100 / (1 + math.exp(-score)))


def level(index):
    return "high" if index >= HIGH else "elevated" if index >= ELEVATED else "low"


def auc(positive, negative):
    """The share of pairs in which the first set has the higher index; a tie counts for half."""
    if not positive or not negative:
        return None
    wins = sum((p > q) + 0.5 * (p == q) for p in positive for q in negative)
    return round(wins / (len(positive) * len(negative)), 3)


def wilson(k, n):
    """The 95% Wilson interval of k of n, as two shares rounded to three places."""
    if n == 0:
        return None
    p, z = k / n, 1.96
    center = (p + z * z / (2 * n)) / (1 + z * z / n)
    half = z * math.sqrt(p * (1 - p) / n + z * z / (4 * n * n)) / (1 + z * z / n)
    return [round(center - half, 3), round(center + half, 3)]


def share(k, n):
    return round(k / n, 3) if n else None


def median(values):
    """The middle value, or the mean of the two middle values."""
    values = sorted(values)
    if not values:
        return None
    middle = len(values) // 2
    return values[middle] if len(values) % 2 else (values[middle - 1] + values[middle]) / 2


# ---------------------------------------------------------------------------------------------------------------
# Pages
# ---------------------------------------------------------------------------------------------------------------


def checked_pages(rows):
    """Check every page row and return the rows. The index of a row has to follow from its counts."""
    seen = set()
    for row in rows:
        key = (row.get("set"), row.get("writer"), row.get("id"))
        if row.get("set") not in SETS or row.get("half") not in HALVES:
            fail(f"page with an unknown set or half: {key}")
        if key in seen:
            fail(f"page listed twice: {key}")
        seen.add(key)
        if row["set"] in ("guided", "plain") and row.get("writer") not in WRITERS:
            fail(f"written page without a known writer: {key}")
        for field in ("index", "marks", "rhythm", "earlier"):
            if not isinstance(row.get(field), int) or not 0 <= row[field] <= 100:
                fail(f"page without a whole {field} of 0 to 100: {key}")
        if row["index"] != max(row["marks"], row["rhythm"]):
            fail(f"index that is not the higher of its two readings: {key}")
        if row["set"] == "set-s":
            # The posts of the style-guide set carry their readings and nothing that tells which posts they are.
            if set(row) - {"set", "half", "id", "index", "marks", "rhythm", "earlier"}:
                fail(f"post of the style-guide set with more than its readings: {key}")
            continue
        recount_readings(row, key)
    return rows


def recount_readings(row, key, earlier=True):
    words, signals = row.get("words"), row.get("signals")
    if not isinstance(words, int) or words < 150 or not isinstance(signals, dict) or set(signals) != set(MARKS) | set(RHYTHM):
        fail(f"page without its words or without every signal: {key}")
    marks = MARKS_BIAS + sum(weight * math.log1p(1000 * signals[name][0] / words) for name, weight in MARKS.items())
    rhythm = RHYTHM_BIAS + sum(weight * math.log1p(1000 * signals[name][1] / words) for name, weight in RHYTHM.items())
    if index_of(marks) != row["marks"] or index_of(rhythm) != row["rhythm"]:
        fail(f"readings that do not follow from the counts: {key}")
    if earlier and index_of(marks) != row["earlier"]:
        fail(f"earlier index that is not the first reading: {key}")


def levels(rows, field="index"):
    counted = {"pages": len(rows), "high": 0, "elevated": 0, "low": 0}
    for row in rows:
        counted[level(row[field])] += 1
    values = sorted(row[field] for row in rows)
    counted["median"] = values[len(values) // 2] if values else None
    return counted


def pages_results(rows):
    result = {}
    for half in HALVES:
        part = [row for row in rows if row["half"] == half]
        of = lambda name, writer=None: [r for r in part if r["set"] == name and (writer is None or r.get("writer") == writer)]
        human = of("human-blog")
        styled = of("set-s") + of("guided")
        entry = {"sets": {}, "separation": {}, "separation_earlier": {}}
        groups = {"human-blog": human, "readme": of("readme"), "set-s": of("set-s")}
        for writer in WRITERS:
            groups["guided, " + writer] = of("guided", writer)
            groups["plain, " + writer] = of("plain", writer)
        groups["all under a style guide"] = styled
        for name, group in groups.items():
            entry["sets"][name] = {"build": levels(group), "earlier": levels(group, "earlier")}
            if name not in ("human-blog", "readme"):
                entry["separation"][name] = auc([r["index"] for r in group], [r["index"] for r in human])
                entry["separation_earlier"][name] = auc([r["earlier"] for r in group], [r["earlier"] for r in human])
        people = human + of("readme")
        high = sum(level(r["index"]) == "high" for r in people)
        entry["quiet"] = {"pages": len(people), "high": high, "share": share(high, len(people)), "interval": wilson(high, len(people)),
                          "high_earlier": sum(level(r["earlier"]) == "high" for r in people),
                          "high_by_rhythm": sum(level(r["rhythm"]) == "high" for r in people)}
        result[half] = entry
    return result


# ---------------------------------------------------------------------------------------------------------------
# Judges
# ---------------------------------------------------------------------------------------------------------------

COMPARISONS = {"plain": "admission", "cross": "admission", "guided": "tells", "set": "tells"}
VOTES = {"model": 1, "human": -1, "same": 0, "unreadable": 0}


def panel(votes_by_unit, judges=JUDGES):
    """Sum the votes of each unit: for the model's text +1, for the other text -1."""
    model = human = 0
    for votes in votes_by_unit.values():
        total = sum(value for judge, value in votes if judge in judges)
        model += total > 0
        human += total < 0
    units = len(votes_by_unit)
    return {"units": units, "model": model, "other": human, "tie": units - model - human, "rate": share(model, units),
            "decisive": share(model, model + human), "interval": wilson(model, model + human)}


def checked_judgments(rows):
    for row in rows:
        if row.get("judge") not in JUDGES:
            fail(f"answer of an unknown judge: {row.get('judge')}")
        if row.get("comparison") == "revision":
            if row.get("vote") not in ("page", "revision", "same", "unreadable") or row.get("reviser") not in WRITERS:
                fail(f"malformed answer about a revision: {row.get('unit')}")
        elif row.get("comparison") not in COMPARISONS or row.get("vote") not in VOTES:
            fail(f"malformed answer: {row.get('comparison')} {row.get('unit')}")
    return rows


def judged(rows, comparison, **only):
    """Group the answers of one comparison by unit, and require one answer of each judge on every pair."""
    units, pairs = {}, {}
    for row in rows:
        if row["comparison"] != comparison or any(row.get(key) != value for key, value in only.items()):
            continue
        pair = (row["unit"], row.get("writer"), row.get("against"), row.get("reviser"))
        pairs.setdefault(pair, []).append(row["judge"])
        value = {"page": 1, "revision": -1}.get(row["vote"], 0) if comparison == "revision" else VOTES[row["vote"]]
        unit = (row["unit"], row.get("writer")) if comparison != "set" else row["unit"]
        units.setdefault(unit, []).append((row["judge"], value))
    for pair, judges in pairs.items():
        if sorted(judges) != sorted(JUDGES):
            fail(f"pair without one answer of each judge: {comparison} {pair}")
    return units


def judges_results(rows):
    result = {}
    for comparison in COMPARISONS:
        units = judged(rows, comparison)
        entry = {"panel": panel(units), "judges": {judge: panel(units, (judge,)) for judge in JUDGES},
                 "unreadable": sum(r["comparison"] == comparison and r["vote"] == "unreadable" for r in rows)}
        if comparison != "set":
            entry["writers"] = {}
            for writer in WRITERS:
                mine = judged(rows, comparison, writer=writer)
                others = tuple(j for j in JUDGES if family(j) != family(writer))
                entry["writers"][writer] = {"panel": panel(mine), "other_family": panel(mine, others)}
        if comparison == "guided":
            entry["guides"] = {guide: panel(judged(rows, comparison, guide=guide)) for guide in ("A", "B")}
        result[comparison] = entry
    return result


def family(model):
    return "qwen" if "Qwen" in model else "gpt"


# ---------------------------------------------------------------------------------------------------------------
# Revisions
# ---------------------------------------------------------------------------------------------------------------


def revisions_results(revisions, judgments):
    result = {}
    for reviser in WRITERS:
        mine = [r for r in revisions if r.get("reviser") == reviser]
        if not mine:
            continue
        for row in mine:
            if row.get("outcome") not in ("revised", "no mark listed", "no revision") or row.get("set") not in ("set-s", "guided"):
                fail(f"malformed revision: {row.get('id')}")
            if row["outcome"] == "revised" and not (isinstance(row.get("before"), list) and isinstance(row.get("after"), list)):
                fail(f"revision without its index before and after: {row.get('id')}")
        groups = {"all": lambda r: True, "set-s": lambda r: r["set"] == "set-s"}
        for writer in WRITERS:
            groups["guided, " + writer] = lambda r, writer=writer: r.get("writer") == writer
        entry = {"pages": len(mine), "no_mark_listed": sum(r["outcome"] == "no mark listed" for r in mine),
                 "no_revision": sum(r["outcome"] == "no revision" for r in mine), "groups": {}}
        for name, belongs in groups.items():
            done = [r for r in mine if r["outcome"] == "revised" and belongs(r)]
            lower = sum(r["after"][0] < r["before"][0] for r in done)
            only = {} if name == "all" else {"set": "set-s"} if name == "set-s" else {"writer": name.split(", ")[1]}
            units = judged(judgments, "revision", reviser=reviser, **only)
            if len(units) != len(done):
                fail(f"revised pages and judged pages differ for {reviser}, {name}: {len(done)} and {len(units)}")
            words = [r["words"][1] / r["words"][0] for r in done if "words" in r]
            entry["groups"][name] = {
                "revised": len(done), "lower": lower, "same": sum(r["after"][0] == r["before"][0] for r in done), "lower_share": share(lower, len(done)),
                "before": levels([{"index": r["before"][0]} for r in done]), "after": levels([{"index": r["after"][0]} for r in done]),
                "words_ratio": round(sum(words) / len(words), 3) if words else None,
                "em_dashes_added": sum(r["em_dashes_added"] for r in done),
                "panel": panel(units), "judges": {judge: panel(units, (judge,)) for judge in JUDGES}}
        result[reviser] = entry
    return result


# ---------------------------------------------------------------------------------------------------------------
# The earlier record
# ---------------------------------------------------------------------------------------------------------------


def earlier_results(rows):
    """The held-out reading of the earlier record: pages before a de-slop commit against the README files."""
    people = [r["index"] for r in rows if r["group"] == "human"]
    pairs = {}
    for row in rows:
        if row["index"] != max(row["marks"], row["rhythm"]):
            fail(f"earlier page whose index is not the higher reading: {row['group']} {row['id']}")
        if row["group"] in ("reserve", "web") and row["outside_main"]:
            pairs.setdefault((row["group"], row["id"]), {})[row["side"]] = row
    for key, pair in pairs.items():
        if set(pair) != {"before", "after"}:
            fail(f"earlier pair without both versions: {key}")
    before = [pair["before"]["index"] for pair in pairs.values()]
    lower = sum(pair["after"]["index"] < pair["before"]["index"] for pair in pairs.values())
    higher = sum(pair["after"]["index"] > pair["before"]["index"] for pair in pairs.values())
    first = lambda values: [v["marks"] for v in values]
    return {"pairs": len(pairs), "readme": len(people), "auc": auc(before, people),
            "auc_first_reading": auc(first([p["before"] for p in pairs.values()]), [r["marks"] for r in rows if r["group"] == "human"]),
            "lower": lower, "differ": lower + higher, "lower_share": share(lower, lower + higher),
            "readme_high": sum(v >= HIGH for v in people), "readme_high_share": share(sum(v >= HIGH for v in people), len(people)),
            "readme_high_first_reading": sum(r["marks"] >= HIGH for r in rows if r["group"] == "human")}


# ---------------------------------------------------------------------------------------------------------------
# The control: text that people wrote under a style guide
# ---------------------------------------------------------------------------------------------------------------


def checked_control(rows):
    """Check every page of the control. Its readings have to follow from its counts, as on the other pages."""
    seen = set()
    for row in rows:
        key = (row.get("repo"), row.get("id"))
        if row.get("kind") not in KINDS or not all(isinstance(row.get(name), str) and row[name] for name in ("repo", "path", "commit", "guide")):
            fail(f"control page without its kind or its source: {key}")
        if key in seen:
            fail(f"control page listed twice: {key}")
        seen.add(key)
        for field in ("index", "marks", "rhythm"):
            if not isinstance(row.get(field), int) or not 0 <= row[field] <= 100:
                fail(f"control page without a whole {field} of 0 to 100: {key}")
        if row["index"] != max(row["marks"], row["rhythm"]):
            fail(f"index that is not the higher of its two readings: {key}")
        recount_readings(row, key, earlier=False)
        if not isinstance(row.get("paragraphs"), (int, float)) or not 0 <= row["paragraphs"] <= 1:
            fail(f"control page without its share of prose in paragraphs: {key}")
        sentences = row.get("sentences")
        if not isinstance(sentences, list) or not all(well_formed_sentence(sentence) for sentence in sentences):
            fail(f"control page with a malformed sentence: {key}")
        if any(first[0] > second[0] for first, second in zip(sentences, sentences[1:])):
            fail(f"control page whose stretches are out of order: {key}")
    return rows


def well_formed_sentence(sentence):
    """A sentence is [stretch, words, plain]; a plain one has five to twenty words."""
    if not isinstance(sentence, list) or len(sentence) != 3 or not all(isinstance(value, int) for value in sentence):
        return False
    stretch, words, plain = sentence
    return stretch >= 1 and words >= 1 and plain in (0, 1) and (not plain or 5 <= words <= 20)


def in_runs(sentences):
    """Count the plain sentences that stand in runs of two or more inside one stretch of paragraphs."""
    total, run, last = 0, 0, None
    for stretch, _, plain in sentences + [[None, 0, 0]]:
        if plain and stretch == last:
            run += 1
        else:
            total += run if run >= 2 else 0
            run = 1 if plain else 0
        last = stretch
    return total


def by_chance(sentences):
    """The expected count of plain sentences in runs when the plain ones are placed at random among the sentences."""
    n, k = len(sentences), sum(sentence[2] for sentence in sentences)
    if n < 2 or k < 2:
        return 0.0
    one = (k - 1) / (n - 1)
    two = 1 - ((n - k) / (n - 1)) * ((n - k - 1) / (n - 2)) if n > 2 else one
    expected = 0.0
    for at, (stretch, _, _) in enumerate(sentences):
        neighbors = (at > 0 and sentences[at - 1][0] == stretch) + (at + 1 < n and sentences[at + 1][0] == stretch)
        expected += (k / n) * (one if neighbors == 1 else two if neighbors == 2 else 0.0)
    return expected


def control_group(rows, set_s):
    """The levels of a group of pages of the control, and how well each reading tells the style-guide set from them."""
    pages = len(rows)
    high = sum(row["rhythm"] >= HIGH for row in rows)
    rate = median(1000 * row["signals"]["plain-sentence-run"][1] / row["words"] for row in rows)
    by_index = levels(rows)
    by_index["median"] = median(row["index"] for row in rows)
    return {"pages": pages, "rhythm_high": high, "rhythm_high_share": share(high, pages), "rhythm_high_interval": wilson(high, pages),
            "rhythm_elevated": sum(ELEVATED <= row["rhythm"] < HIGH for row in rows), "index": by_index,
            "marks_above_low": sum(row["marks"] >= ELEVATED for row in rows),
            "high_by_marks": sum(row["marks"] >= HIGH for row in rows),
            "median_rhythm": median(row["rhythm"] for row in rows), "median_rate": None if rate is None else round(rate, 1),
            "auc_index": auc([post["index"] for post in set_s], [row["index"] for row in rows]),
            "auc_rhythm": auc([post["rhythm"] for post in set_s], [row["rhythm"] for row in rows])}


def sentence_figures(row):
    sentences = row["sentences"]
    count = len(sentences)
    return {"count": count, "words": sum(sentence[1] for sentence in sentences) / count,
            "long": sum(sentence[1] > 20 for sentence in sentences) / count, "plain": sum(sentence[2] for sentence in sentences) / count,
            "runs": in_runs(sentences), "chance": by_chance(sentences), "per_thousand": 1000 * count / row["words"]}


def sentences_group(rows):
    """What the sentences of a group of pages are like. A page with too few sentences in paragraphs is left out."""
    figures = [sentence_figures(row) for row in rows if len(row["sentences"]) >= FEWEST_SENTENCES]
    if not figures:
        return {"pages": 0}
    chance = sum(figure["chance"] for figure in figures)
    return {"pages": len(figures), "words_per_sentence": round(median(figure["words"] for figure in figures), 1),
            "over_20_words": round(median(figure["long"] for figure in figures), 3),
            "plain": round(median(figure["plain"] for figure in figures), 3),
            "in_runs": round(median(figure["runs"] / figure["count"] for figure in figures), 3),
            "by_chance": round(median(figure["chance"] / figure["count"] for figure in figures), 3),
            "runs_to_chance": round(sum(figure["runs"] for figure in figures) / chance, 2) if chance else None,
            "paragraph_sentences_per_thousand_words": round(median(figure["per_thousand"] for figure in figures), 1)}


def control_results(rows, pages):
    set_s = [row for row in pages if row["set"] == "set-s" and row["half"] == "measured"]
    matched = [row for row in rows if row["paragraphs"] >= MATCH]
    of = lambda group, kind: [row for row in group if row["kind"] == kind]
    result = {"match": MATCH, "fewest_sentences": FEWEST_SENTENCES, "pages": len(rows),
              "groups": {"people": control_group(rows, set_s), "matched": control_group(matched, set_s)}, "sources": {}}
    for label, group in (("matched", matched), ("all", rows)):
        for kind in KINDS:
            result["groups"][f"{label} {kind}"] = control_group(of(group, kind), set_s)
    for repo in sorted({row["repo"] for row in rows}):
        mine = [row for row in rows if row["repo"] == repo]
        entry = control_group(mine, set_s)
        entry.update(kind=mine[0]["kind"], guide=mine[0]["guide"], commit=mine[0]["commit"], matched=sum(row["paragraphs"] >= MATCH for row in mine))
        result["sources"][repo] = entry
    by_set = levels(set_s)
    by_set["median"] = median(post["index"] for post in set_s)
    result["set_s"] = {"posts": len(set_s), "rhythm_high": sum(post["rhythm"] >= HIGH for post in set_s),
                       "rhythm_elevated": sum(ELEVATED <= post["rhythm"] < HIGH for post in set_s), "index": by_set,
                       "marks_above_low": sum(post["marks"] >= ELEVATED for post in set_s)}
    rebuilt = [in_runs(row["sentences"]) for row in rows]
    built = [row["signals"]["plain-sentence-run"][1] for row in rows]
    result["rebuild"] = {"pages": len(rows), "equal": sum(mine == theirs for mine, theirs in zip(rebuilt, built)),
                         "rebuilt": sum(rebuilt), "build": sum(built)}
    result["sentences"] = {"set_s": SET_S_SENTENCES, "matched articles": sentences_group(of(matched, "articles")),
                           "matched documentation": sentences_group(of(matched, "documentation")), "matched": sentences_group(matched)}
    counted = [row for row in matched if len(row["sentences"]) >= FEWEST_SENTENCES]
    floor = SET_S_SENTENCES["plain_lower_quartile"]
    like = [row for row in counted if sentence_figures(row)["plain"] >= floor]
    result["as_plain_as_set_s"] = {"floor": floor, "pages": len(like), "of": len(counted),
                                   "rhythm_lowest": min((row["rhythm"] for row in like), default=None),
                                   "rhythm_highest": max((row["rhythm"] for row in like), default=None),
                                   "rhythm_high": sum(row["rhythm"] >= HIGH for row in like),
                                   "words_per_sentence": round(median(sentence_figures(row)["words"] for row in like), 1) if like else None}
    return result


# ---------------------------------------------------------------------------------------------------------------
# The whole record
# ---------------------------------------------------------------------------------------------------------------


def results(root):
    pages = checked_pages(read(root, "pages"))
    judgments = checked_judgments(read(root, "judgments"))
    result = {"readings": {"marks": {"weights": MARKS, "bias": MARKS_BIAS}, "rhythm": {"weights": RHYTHM, "bias": RHYTHM_BIAS},
                           "elevated_at": ELEVATED, "high_at": HIGH},
              "bars": BARS, "pages": pages_results(pages), "judges": judges_results(judgments),
              "revisions": revisions_results(read(root, "revisions"), judgments), "earlier": earlier_results(read(root, "earlier")),
              "control": control_results(checked_control(read(root, "control")), pages)}
    result["registered"] = registered(result)
    return result


def registered(result):
    """Read each registered bar on the measured half."""
    judges, measured = result["judges"], result["pages"]["measured"]
    held = {
        "admission, same subject": judges["plain"]["panel"]["rate"] >= BARS["admission"],
        "admission, other subject": judges["cross"]["panel"]["rate"] >= BARS["admission"],
        "guided texts are told from human posts": judges["guided"]["panel"]["rate"] >= BARS["tells"],
        "the style-guide set is told from human posts": judges["set"]["panel"]["rate"] >= BARS["tells"],
        "quiet on human pages": measured["quiet"]["share"] <= BARS["quiet"],
        "earlier result": result["earlier"]["auc"] >= BARS["earlier"],
    }
    for name, value in measured["separation"].items():
        if not name.startswith("plain"):
            held["separation, " + name] = value >= BARS["separation"]
    reviser = result["revisions"].get(WRITERS[0])
    if reviser:
        everything = reviser["groups"]["all"]
        held["index lower after the revision"] = everything["lower_share"] >= BARS["index_lower"]
        held["revision less like a model, all pages"] = (everything["panel"]["decisive"] or 0) > 0.5 and everything["panel"]["interval"][0] > 0.5
        for name, group in reviser["groups"].items():
            if name != "all":
                held["revision less like a model, " + name] = (group["panel"]["decisive"] or 0) > 0.5
    articles = result["control"]["groups"]["matched articles"]
    if articles["pages"]:
        held["control, second reading quiet on people's articles"] = articles["rhythm_high_share"] <= BARS["control_quiet"]
        held["control, the style-guide set apart from people's articles"] = articles["auc_index"] >= BARS["control_separation"]
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
    print(f"style-guide text: {sum(held.values())} of {len(held)} registered bars hold")
    return 0


if __name__ == "__main__":
    sys.exit(main())
