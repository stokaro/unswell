#!/usr/bin/env python3
"""Verify the measurement against human references offline.

usage: verify.py [--write] [record directory]

The record holds coordinates and choices, never text. The check covers file
digests, the schema, and every published number: the sentence-level rates of
each system against the sentences an editor changed, the agreement between
editors, the agreement of judge models with the writers' preferences, the
admission of judges that follows, and the words that trims remove against the
words an editor changed. RESULTS.md must equal the rendering of the
results. With --write the results, their rendering, and the digests are
regenerated from the data files.
"""
import collections
import gzip
import hashlib
import itertools
import json
import math
import random
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
DIGESTS = "files.sha256"
RESULTS = "results.json"
DATA = ("data/sentences.jsonl.gz", "data/preferences.jsonl.gz", "data/expert-pairs.jsonl.gz", "data/readers.jsonl.gz",
        "data/panel.jsonl.gz", "data/trims.jsonl.gz", "data/word-counts.jsonl.gz", "data/openers.jsonl.gz",
        "data/rule-words.jsonl.gz", "data/opener-load.jsonl.gz", "data/vocabulary.jsonl.gz",
        "data/vocabulary-phrases.jsonl.gz", "data/vocabulary-load.jsonl.gz", "data/deslop-pairs.jsonl.gz",
        "data/emoji-load.jsonl.gz", "data/frequency-load.jsonl.gz", "data/review-repeat.jsonl.gz")
REVIEW_REPEAT = "data/review-repeat.jsonl.gz"
REVIEW_REPEAT_KEYS = {"document", "reviewer", "protocol", "words", "changed", "findings", "flagged", "flagged_changed", "sentences",
                      "both", "after_words", "after_findings", "edit"}
REVIEW_EDIT_KEYS = ("sentences", "sentences_changed", "flagged_before", "changed_by_commit", "flagged_after", "on_new_text",
                    "flagged_both_times")
# The seed that the registered comparison of the answer protocols was run with.
REVIEW_REPEAT_SEED = 20261113
# A verdict for every item; the same with "revise" only when the defect is clear; only the items that need revision.
REVIEW_PROTOCOLS = ("verdicts", "strict", "list")
EMOJI_LOAD, FREQUENCY_LOAD = "data/emoji-load.jsonl.gz", "data/frequency-load.jsonl.gz"
DESLOP = "data/deslop-pairs.jsonl.gz"
DESLOP_KEYS = {"id", "repository", "commit", "path", "part", "web_edit", "english", "words_before", "words_after", "before",
               "after"}
# The rules that were tried, and the bars registered for a rule and for the default set.
DESLOP_NEW = ("format.decorative-emoji", "format.em-dash-density", "format.em-dash-frequency")
# This rule reports a document once. Its count in a pair is the em dashes it lists, which is its registered measure.
DESLOP_DOCUMENT = "format.em-dash-frequency"
# Its registered limit: it may flag at most one in twenty of the pages that people wrote before November 2022.
DESLOP_HUMAN_PAGES = 20
# The emoji rule as it was registered and tested first, with check marks and crosses counted.
DESLOP_REGISTERED = "format.decorative-emoji-with-status-marks"
DESLOP_RULE_BAR, DESLOP_SET_BAR, DESLOP_FLOOR = 0.65, 0.60, 0.55
VOCABULARY, VOCABULARY_PHRASES = "data/vocabulary.jsonl.gz", "data/vocabulary-phrases.jsonl.gz"
VOCABULARY_LOAD = "data/vocabulary-load.jsonl.gz"
VOCABULARY_PARTS = ("selection", "confirmation")
VOCABULARY_KEYS = {"id", "part", "text_sha256", "words", "changed", "reference", "findings"}
# The confirmation rule: findings with every flagged word changed, and the lower end of their interval.
VOCABULARY_SHARE, VOCABULARY_FLOOR = 0.70, 0.60
# The selection rule: a phrase with this many findings leaves when fewer than this share of them are fully changed.
VOCABULARY_MIN_FINDINGS, VOCABULARY_KEEP = 4, 0.5
TRIMS, WORD_COUNTS = "data/trims.jsonl.gz", "data/word-counts.jsonl.gz"
OPENERS, RULE_WORDS = "data/openers.jsonl.gz", "data/rule-words.jsonl.gz"
# Counts of the rule's findings on documentation pages outside this record's sets.
OPENER_PAGES = "data/opener-load.jsonl.gz"
# The run of the default rules with the build whose chat-preamble rule lists the opener.
OPENER_SYSTEM = "rules-opener"
# The set on which the opener was registered, and the sets on which it must stay silent.
OPENER_SET, OPENER_ELSEWHERE = "beemo", ("prs", "lamp")
OPENER_SHARE, OPENER_FLOOR, OPENER_LOAD = 0.85, 0.70, 1
TRIM_MODELS = ("claude", "astra")
TRIM_KEYS = {"id", "model", "sentences", "proposed", "rejected", "trims"}
TRIM_FIELDS = {"span", "words", "fixed", "removed"}
COUNT_KEYS = {"id", "words", "changed", "systems"}
WORD_SYSTEMS = ("rules-default", "review-claude", "review-astra")
TRIM_SPLITS = ("dev", "test")
# Sets whose record of the editor's text allows an edit distance.
DISTANCE_SETS = ("beemo", "lamp")
BOOTSTRAP, SEED = 2000, 20261112
# The per-document sums behind every trim statistic, in the order the bootstrap adds them.
SUMS = ("removed", "hits", "sentences", "document", "moved", "control", "measured")
SETS = {"lamp": {"dev", "test", "ceiling"}, "beemo": {"dev", "test"}, "prs": {"dev", "test"}}
DOCUMENT_KEYS = {"id", "set", "split", "editor", "kind", "generator", "text_sha256", "units", "reference", "systems"}
OPTIONAL_KEYS = {"repository", "pull_request", "path", "base", "words"}
VERSIONS = ["AI-edited", "AI-generated", "Human-edited"]
PAIRS = [("Human-edited", "AI-generated"), ("AI-edited", "AI-generated"), ("Human-edited", "AI-edited")]
ADMISSION = {"Human-edited over AI-generated": 0.70, "AI-edited over AI-generated": 0.65}
EXPERT_FLOOR = 0.65
READER_KEYS = {"reader", "pair", "arm", "reviewer", "outcome"}
MIN_DECISIVE = 15
PREFERRED = 0.70
LEAD = 0.15
PANEL_KEYS = {"comparison", "pair", "arm", "reviewer", "judges"}
OUTCOMES = ("revised", "original", "same")
JUDGE_MODELS = ("fable", "claude", "astra", "qwen")
COMPARISONS = {"first": ("reviewed", "control"), "trim": ("reviewed", "control"), "product": ("product",)}
TOP_JUDGE = "fable"
WEIGHTS = {TOP_JUDGE: 3, "astra": 2}
FAMILIES = ("fable", "astra", "qwen")
COUNTS = ("documents", "sentences", "edited", "flagged", "edited_flagged", "unedited_flagged", "findings",
          "findings_on_edit", "edits", "edits_touched")


def fail(message):
    raise ValueError(message)


def rows(path):
    with gzip.open(path, "rt") as handle:
        for line in handle:
            yield json.loads(line)


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def tracked(root):
    names = [name for name in DATA if (root / name).exists()]
    return sorted(names + ([RESULTS] if (root / RESULTS).exists() else []))


def write_digests(root):
    (root / DIGESTS).write_text("".join(f"{digest(root / name)}  {name}\n" for name in tracked(root)))


def check_digests(root):
    listed = dict(reversed(line.split("  ", 1)) for line in (root / DIGESTS).read_text().splitlines())
    if sorted(listed) != tracked(root):
        fail("data files differ from the digest list")
    for name, value in listed.items():
        if digest(root / name) != value:
            fail(f"digest mismatch for {name}")


def check_spans(spans, where):
    for span in spans:
        if not (isinstance(span, list) and len(span) == 2 and all(isinstance(v, int) for v in span) and 0 <= span[0] < span[1]):
            fail(f"invalid span in {where}")


def overlaps(span, others):
    return any(span[0] < end and start < span[1] for start, end in others)


def load_documents(root):
    documents, seen = [], set()
    for row in rows(root / DATA[0]):
        extra = set(row) - DOCUMENT_KEYS
        if set(row) - extra != DOCUMENT_KEYS or not extra <= OPTIONAL_KEYS:
            fail(f"unexpected document fields: {sorted(set(row) ^ DOCUMENT_KEYS)}")
        if row["id"] in seen:
            fail(f"duplicate document {row['id']}")
        seen.add(row["id"])
        if row["split"] not in SETS.get(row["set"], ()):
            fail(f"unknown set or split for {row['id']}")
        if len(row["text_sha256"]) != 64:
            fail(f"document {row['id']} is not identified by a digest")
        check_spans(row["units"], row["id"])
        check_spans(row["reference"], row["id"])
        for system, spans in row["systems"].items():
            check_spans(spans, f"{row['id']} for {system}")
        check_spans(row.get("words", []), row["id"])
        documents.append(row)
    return documents


def count(flags, reference, units, counter):
    for unit in units:
        edited, flagged = overlaps(unit, reference), overlaps(unit, flags)
        counter["sentences"] += 1
        counter["edited"] += edited
        counter["flagged"] += flagged
        counter["edited_flagged"] += edited and flagged
        counter["unedited_flagged"] += flagged and not edited
    counter["findings"] += len(flags)
    counter["findings_on_edit"] += sum(overlaps(span, reference) for span in flags)
    counter["edits"] += len(reference)
    counter["edits_touched"] += sum(overlaps(span, flags) for span in reference)
    counter["documents"] += 1


def rates(counter):
    entry = {name: counter[name] for name in COUNTS}
    unedited = entry["sentences"] - entry["edited"]
    entry["tpr"] = entry["edited_flagged"] / entry["edited"] if entry["edited"] else 0.0
    entry["fpr"] = entry["unedited_flagged"] / unedited if unedited else 0.0
    entry["j"] = entry["tpr"] - entry["fpr"]
    return entry


def sentence_results(documents):
    """Return base rates, the ceiling between editors, and each system's rates per set and split."""
    result = {"sets": {}, "ceiling": {}, "systems": {}}
    groups = collections.defaultdict(list)
    for row in documents:
        groups[(row["set"], row["split"])].append(row)
    for (name, split), group in sorted(groups.items()):
        base = collections.Counter()
        for row in group:
            count([], row["reference"], row["units"], base)
        result["sets"][f"{name}/{split}"] = rates(base)
        if split == "ceiling":
            continue
        for system in sorted({s for row in group for s in row["systems"]}):
            counter = collections.Counter()
            for row in group:
                if system in row["systems"]:
                    count(row["systems"][system], row["reference"], row["units"], counter)
            result["systems"].setdefault(system, {})[f"{name}/{split}"] = rates(counter)
    paragraphs = collections.defaultdict(dict)
    for row in documents:
        if row["split"] == "ceiling":
            if row["editor"] in paragraphs[row["text_sha256"]]:
                fail(f"editor {row['editor']} edited one paragraph twice")
            paragraphs[row["text_sha256"]][row["editor"]] = row
    pooled = collections.Counter()
    for editors in paragraphs.values():
        for first, second in itertools.permutations(sorted(editors), 2):
            count(editors[first]["reference"], editors[second]["reference"], editors[first]["units"], pooled)
    if pooled:
        result["ceiling"] = rates(pooled)
        result["ceiling"]["paragraphs"] = len(paragraphs)
        result["ceiling"]["words"] = word_agreement(paragraphs)
    return result


def word_agreement(paragraphs):
    """Agreement between editors on single words: pooled ordered pairs, Cohen's kappa, and how many editors marked a word."""
    both = first_only = second_only = neither = 0
    votes = collections.Counter()
    for editors in paragraphs.values():
        marks = {name: [overlaps(word, row["reference"]) for word in row.get("words", [])] for name, row in editors.items()}
        for first, second in itertools.permutations(sorted(marks), 2):
            for a, b in zip(marks[first], marks[second]):
                both += a and b
                first_only += a and not b
                second_only += b and not a
                neither += not a and not b
        if len(marks) == 3:
            for word in zip(*marks.values()):
                votes[sum(word)] += 1
    total = both + first_only + second_only + neither
    if not total:
        return {}
    observed = (both + neither) / total
    marked_first, marked_second = (both + first_only) / total, (both + second_only) / total
    expected = marked_first * marked_second + (1 - marked_first) * (1 - marked_second)
    return {"word_pairs": total, "both": both, "first_only": first_only, "second_only": second_only, "neither": neither,
            "kappa": (observed - expected) / (1 - expected) if expected < 1 else 0.0,
            "marked_by": {str(count): votes[count] for count in range(4)}}


def preference_results(root):
    """Return how writers agree with each other and how each judge agrees with their majority."""
    writers, judges, triplets = {}, collections.defaultdict(dict), 0
    for row in rows(root / DATA[1]):
        if set(row) != {"triplet", "writers", "judges"}:
            fail(f"unexpected preference fields: {sorted(row)}")
        orders = [entry["order"] for entry in row["writers"]]
        for order in orders + list(row["judges"].values()):
            if sorted(order) != VERSIONS:
                fail(f"ranking of {row['triplet']} is not an order of the three versions")
        if len(orders) != 3:
            fail(f"triplet {row['triplet']} does not have three writers")
        writers[row["triplet"]] = orders
        for judge, order in row["judges"].items():
            judges[judge][row["triplet"]] = order
        triplets += 1
    result = {"triplets": triplets, "judgments": 3 * triplets, "writers": {}, "judges": {}}
    for first, second in PAIRS:
        name = f"{first} over {second}"
        agree = total = prefer = 0
        for orders in writers.values():
            votes = [order.index(first) < order.index(second) for order in orders]
            prefer += sum(votes)
            for a, b in itertools.combinations(votes, 2):
                total += 1
                agree += a == b
        result["writers"][name] = {"judgments_preferring_first": prefer, "writer_pairs": total, "writer_pairs_agreeing": agree}
    for judge, orders in sorted(judges.items()):
        entry = {"triplets": len(orders), "pairs": {}}
        for first, second in PAIRS:
            prefers = majority = 0
            for triplet, order in orders.items():
                choice = order.index(first) < order.index(second)
                votes = sum(o.index(first) < o.index(second) for o in writers[triplet])
                prefers += choice
                majority += choice == (votes >= 2)
            entry["pairs"][f"{first} over {second}"] = {"prefers_first": prefers, "agrees_with_majority": majority}
        result["judges"][judge] = entry
    return result


def expert_results(root):
    judges = collections.defaultdict(collections.Counter)
    for row in rows(root / DATA[2]):
        if set(row) != {"id", "judges"}:
            fail(f"unexpected expert-pair fields: {sorted(row)}")
        for judge, answer in row["judges"].items():
            if answer.get("better") not in ("edited", "raw") or answer.get("position") not in ("A", "B") or set(answer) != {"better", "position"}:
                fail(f"invalid choice of {judge} for {row['id']}")
            judges[judge]["pairs"] += 1
            judges[judge]["prefers_edited"] += answer["better"] == "edited"
            judges[judge]["chooses_first_position"] += answer["position"] == "A"
    return {judge: dict(counter) for judge, counter in sorted(judges.items())}


def admissions(preferences, experts):
    """A judge is admitted when it meets all three registered floors."""
    result = {}
    for judge, entry in preferences["judges"].items():
        checks = {name: entry["pairs"][name]["agrees_with_majority"] >= floor * entry["triplets"] for name, floor in ADMISSION.items()}
        expert = experts.get(judge, {})
        checks["expert-edited over raw"] = bool(expert) and expert["prefers_edited"] >= EXPERT_FLOOR * expert["pairs"]
        result[judge] = {"checks": checks, "admitted": all(checks.values())}
    return result


def wilson(yes, total, z=1.96):
    if not total:
        return [0.0, 0.0]
    share = yes / total
    centre = (share + z * z / (2 * total)) / (1 + z * z / total)
    half = z * (share * (1 - share) / total + z * z / (4 * total * total)) ** 0.5 / (1 + z * z / total)
    return [centre - half, centre + half]


def reader_results(root):
    """Count the reader answers to the blind comparison by arm. They are withdrawn as a measure and decide nothing."""
    path = root / DATA[3]
    if not path.exists():
        return {}
    arms, readers, seen = collections.defaultdict(collections.Counter), set(), set()
    for row in rows(path):
        if set(row) != READER_KEYS:
            fail(f"unexpected reader fields: {sorted(row)}")
        if row["arm"] not in ("reviewed", "control") or row["outcome"] not in ("revised", "original", "same"):
            fail(f"invalid reader answer for pair {row['pair']}")
        if (row["reader"], row["pair"]) in seen:
            fail(f"reader {row['reader']} answered pair {row['pair']} twice")
        seen.add((row["reader"], row["pair"]))
        readers.add(row["reader"])
        arms[row["arm"]][row["outcome"]] += 1
    result = {"readers": len(readers), "arms": {}}
    for arm in ("reviewed", "control"):
        counter = arms[arm]
        decisive = counter["revised"] + counter["original"]
        result["arms"][arm] = {"revised": counter["revised"], "original": counter["original"], "same": counter["same"],
                               "decisive": decisive, "interval": wilson(counter["revised"], decisive)}
    return result


def consensus(votes):
    """The answer with the most votes. The top model has three votes, the second two, and the top model decides a tie."""
    tally = collections.Counter()
    for judge, vote in votes.items():
        tally[vote] += WEIGHTS.get(judge, 1)
    best = max(tally.values())
    leaders = [vote for vote, count in tally.items() if count == best]
    return leaders[0] if len(leaders) == 1 else votes[TOP_JUDGE]


def family_answer(votes):
    """One equal vote per model family: the answer two of the three give, else no real difference."""
    answer, count = collections.Counter(votes[judge] for judge in FAMILIES).most_common(1)[0]
    return answer if count >= 2 else "same"


def arm_summary(counter):
    decisive = counter["revised"] + counter["original"]
    return {"revised": counter["revised"], "original": counter["original"], "same": counter["same"], "decisive": decisive,
            "interval": wilson(counter["revised"], decisive)}


def meets(entry):
    return (entry["decisive"] >= MIN_DECISIVE and entry["revised"] >= PREFERRED * entry["decisive"]
            and entry["interval"][0] > 0.5)


def panel_results(root):
    """Read the comparisons on the weighted consensus of the judge models and apply the registered rule."""
    path = root / DATA[4]
    if not path.exists():
        return {}
    arms = {comparison: collections.defaultdict(collections.Counter) for comparison in COMPARISONS}
    families = {comparison: collections.defaultdict(collections.Counter) for comparison in COMPARISONS}
    single = collections.defaultdict(lambda: collections.defaultdict(collections.Counter))
    counts, seen = collections.defaultdict(collections.Counter), set()
    for row in rows(path):
        if set(row) != PANEL_KEYS:
            fail(f"unexpected panel fields: {sorted(row)}")
        if row["arm"] not in COMPARISONS.get(row["comparison"], ()):
            fail(f"invalid panel row for pair {row['pair']}")
        if set(row["judges"]) != set(JUDGE_MODELS) or any(vote not in OUTCOMES for vote in row["judges"].values()):
            fail(f"pair {row['pair']} does not carry one valid answer of every judge")
        if (row["comparison"], row["pair"]) in seen:
            fail(f"pair {row['pair']} appears twice in the panel answers")
        seen.add((row["comparison"], row["pair"]))
        arms[row["comparison"]][row["arm"]][consensus(row["judges"])] += 1
        families[row["comparison"]][row["arm"]][family_answer(row["judges"])] += 1
        for judge, vote in row["judges"].items():
            single[(row["comparison"], row["arm"])][judge][vote] += 1
        counts[row["comparison"]]["pairs"] += 1
        counts[row["comparison"]]["unanimous"] += len(set(row["judges"].values())) == 1
        counts[row["comparison"]]["opposed"] += {"revised", "original"} <= set(row["judges"].values())
    result = {}
    for comparison, names in COMPARISONS.items():
        if not counts[comparison]["pairs"]:
            continue
        summaries = {name: arm_summary(arms[comparison][name]) for name in names}
        for name in names:
            summaries[name]["judges"] = {judge: {vote: single[(comparison, name)][judge][vote] for vote in OUTCOMES}
                                         for judge in JUDGE_MODELS}
        entry = {**dict(counts[comparison]), "arms": summaries}
        entry["one_vote_per_family"] = {name: arm_summary(families[comparison][name]) for name in names}
        for name in names:
            entry[f"{name}_meets_rule"] = meets(summaries[name])
            entry["one_vote_per_family"][name]["meets_rule"] = meets(entry["one_vote_per_family"][name])
        if names == ("reviewed", "control"):
            reviewed, control = summaries["reviewed"], summaries["control"]
            share = reviewed["revised"] / reviewed["decisive"] if reviewed["decisive"] else 0.0
            other = control["revised"] / control["decisive"] if control["decisive"] else 0.0
            entry["ahead_of_control"] = share - other >= LEAD
        result[comparison] = entry
    return result


def within(span, others):
    return any(start <= span[0] and span[1] <= end for start, end in others)


def load_word_counts(root, documents):
    """Return, per document, its words, the words an editor changed, and both counts under each earlier system."""
    known = {row["id"] for row in documents if row["split"] in TRIM_SPLITS}
    counts = {}
    for row in rows(root / WORD_COUNTS):
        if set(row) != COUNT_KEYS:
            fail(f"unexpected word count fields: {sorted(row)}")
        if row["id"] not in known or row["id"] in counts:
            fail(f"word counts for an unknown or repeated document {row['id']}")
        if not (isinstance(row["words"], int) and isinstance(row["changed"], int) and 0 <= row["changed"] <= row["words"]):
            fail(f"invalid word counts for {row['id']}")
        for system, pair in row["systems"].items():
            valid = isinstance(pair, list) and len(pair) == 2 and all(isinstance(v, int) for v in pair)
            if system not in WORD_SYSTEMS or not valid or not 0 <= pair[1] <= pair[0] <= row["words"]:
                fail(f"invalid counts of {system} for {row['id']}")
        counts[row["id"]] = row
    if set(counts) != known:
        fail("word counts do not cover the development and test documents")
    return counts


def trim_counts(row, document):
    """Recount one document's accepted trims from their coordinates."""
    removed = hits = 0
    chance = 0.0
    spans = document["systems"].get(f"trim-{row['model']}", [])
    last = -1
    for trim in row["trims"]:
        if set(trim) != TRIM_FIELDS:
            fail(f"unexpected trim fields in {row['id']}")
        check_spans([trim["span"], *trim["words"]], row["id"])
        if trim["span"] not in spans or trim["span"][0] < last:
            fail(f"a trim of {row['id']} is not a finding of trim-{row['model']} or overlaps another")
        last, position = trim["span"][1], trim["span"][0]
        for word in trim["words"]:
            if word[0] < position or word[1] > trim["span"][1]:
                fail(f"words of a trim in {row['id']} are out of order or outside the sentence")
            position = word[1]
        size = len(trim["words"])
        fixed, gone = set(trim["fixed"]), trim["removed"]
        indexes = [*trim["fixed"], *gone]
        if not all(isinstance(i, int) and 0 <= i < size for i in indexes) or len(set(gone)) != len(gone) or not gone:
            fail(f"invalid word indexes in a trim of {row['id']}")
        if fixed & set(gone):
            fail(f"a trim of {row['id']} removes a word that the check never removes")
        changed = [within(word, document["reference"]) for word in trim["words"]]
        pool = [index for index in range(size) if index not in fixed]
        removed += len(gone)
        hits += sum(changed[index] for index in gone)
        chance += len(gone) * sum(changed[index] for index in pool) / len(pool)
    return removed, hits, chance


def trim_entry(row, document, counts):
    """Return one document's sums: what was removed, what an editor changed, and the edit distances."""
    if set(row) - {"distance"} != TRIM_KEYS:
        fail(f"unexpected trim row fields: {sorted(row)}")
    if not all(isinstance(row[name], int) and row[name] >= 0 for name in ("sentences", "proposed", "rejected")):
        fail(f"invalid counts in the trim row of {row['id']}")
    removed, hits, chance = trim_counts(row, document)
    share = counts["changed"] / counts["words"] if counts["words"] else 0.0
    entry = {"removed": removed, "hits": hits, "sentences": chance, "document": removed * share, "trims": len(row["trims"]),
             "rejected": row["rejected"], "proposed": row["proposed"], "offered": row["sentences"],
             "moved": 0.0, "control": 0.0, "measured": 0, "words": counts["words"], "changed": counts["changed"]}
    measured = document["set"] in DISTANCE_SETS and removed > 0
    if measured != ("distance" in row):
        fail(f"the trim row of {row['id']} must carry a distance exactly when its set has the editor's text and words were removed")
    if measured:
        distance = row["distance"]
        if set(distance) != {"before", "after", "control"} or not all(isinstance(distance[k], int) for k in ("before", "after")):
            fail(f"invalid distance in the trim row of {row['id']}")
        moved, control = distance["after"] - distance["before"], distance["control"] - distance["before"]
        # Removing a word changes an edit distance by at most one.
        if distance["before"] < 0 or abs(moved) > removed or abs(control) > removed:
            fail(f"the distance of {row['id']} moves further than the removed words allow")
        entry.update(moved=float(moved), control=float(control), measured=removed)
    return entry


def totals(table):
    """Sum the columns of SUMS over the rows of a table.

    The sums are exactly rounded. The built-in sum of floats changed in Python 3.12, and a check that
    compares published numbers must not depend on the interpreter that recounts them.
    """
    return [math.fsum(column) for column in zip(*table)] if table else [0.0] * len(SUMS)


def trim_differences(total, sentences, distance):
    """Return precision against both chance values and, where measured, the net movement against its control."""
    removed, hits, in_sentences, in_document, moved, control, measured = total
    if not removed:
        return {}
    result = {"precision_minus_document": (hits - in_document) / removed}
    if sentences:
        result["precision_minus_sentences"] = (hits - in_sentences) / removed
    if distance and measured:
        result["net"] = moved / measured
        result["net_minus_control"] = (moved - control) / measured
    return result


def trim_intervals(table, sentences, distance, label):
    """Return 95% percentile intervals of the differences, resampling documents."""
    generator = random.Random(f"{SEED}|{label}")
    draws, size = collections.defaultdict(list), len(table)
    for _ in range(BOOTSTRAP if size else 0):
        a = b = c = d = e = f = g = 0.0
        for _ in range(size):
            row = table[generator.randrange(size)]
            a += row[0]
            b += row[1]
            c += row[2]
            d += row[3]
            e += row[4]
            f += row[5]
            g += row[6]
        for key, value in trim_differences([a, b, c, d, e, f, g], sentences, distance).items():
            draws[key].append(value)
    result = {}
    for key, values in draws.items():
        values.sort()
        result[key] = [values[int(0.025 * len(values))], values[min(len(values) - 1, int(0.975 * len(values)))]]
    return result


def trim_report(entries, sentences, distance, label):
    """Summarize one system on one part of one set."""
    table = [tuple(float(entry.get(name, 0)) for name in SUMS) for entry in entries]
    removed, hits, in_sentences, in_document, moved, control, measured = totals(table)
    words = sum(entry["words"] for entry in entries if entry["removed"])
    report = {"documents": len(entries), "documents_with_output": sum(1 for entry in entries if entry["removed"]),
              "removed_words": int(removed), "precision": hits / removed if removed else None,
              "chance_document": in_document / removed if removed else None,
              "pooled_share": sum(entry["changed"] for entry in entries if entry["removed"]) / words if words else None,
              "differences": trim_differences([removed, hits, in_sentences, in_document, moved, control, measured],
                                              sentences, distance),
              "intervals": trim_intervals(table, sentences, distance, label)}
    if sentences:
        report["chance_sentences"] = in_sentences / removed if removed else None
        for name in ("trims", "rejected", "proposed", "offered"):
            report[name] = sum(entry[name] for entry in entries)
    if distance:
        report["net"] = moved / measured if measured else None
        report["control_net"] = control / measured if measured else None
    return report


def trim_rule(models, distance):
    """Apply the registered rule to the test part of one set."""
    def low(report, key):
        return report["intervals"].get(key, [float("-inf"), float("inf")])[0]

    def high(report, key):
        return report["intervals"].get(key, [float("-inf"), float("inf")])[1]

    above = all(low(report, "precision_minus_sentences") > 0 and low(report, "precision_minus_document") > 0
                and (not distance or high(report, "net_minus_control") < 0) for report in models.values())
    outcome = {"agrees_above_chance": above}
    if distance:
        outcome["moves_toward_editor"] = above and all(high(report, "net") < 0 for report in models.values())
    return outcome


def trim_results(root, documents):
    """Return, per set and part, how the words that trims remove compare with the words an editor changed."""
    if not (root / TRIMS).exists():
        return {}
    counts = load_word_counts(root, documents)
    by_id = {row["id"]: row for row in documents if row["split"] in TRIM_SPLITS}
    groups, seen = collections.defaultdict(list), set()
    for row in rows(root / TRIMS):
        document = by_id.get(row.get("id"))
        if document is None or row.get("model") not in TRIM_MODELS or (row["id"], row["model"]) in seen:
            fail(f"a trim row names an unknown or repeated document and model: {row.get('id')}")
        seen.add((row["id"], row["model"]))
        groups[(document["set"], document["split"], row["model"])].append(trim_entry(row, document, counts[row["id"]]))
    if seen != {(name, model) for name in by_id for model in TRIM_MODELS}:
        fail("trim rows do not cover every development and test document with every model")
    result = {"bootstrap_draws": BOOTSTRAP, "sets": {}}
    for name, split in sorted({(row["set"], row["split"]) for row in by_id.values()}):
        distance = name in DISTANCE_SETS
        models = {model: trim_report(groups[(name, split, model)], True, distance, f"{name}|{split}|{model}")
                  for model in TRIM_MODELS}
        entry = {"models": models, "comparisons": {}}
        selected = [row["id"] for row in by_id.values() if (row["set"], row["split"]) == (name, split)]
        for system in WORD_SYSTEMS:
            entries = []
            for identifier in sorted(selected):
                if system in counts[identifier]["systems"]:
                    under, hits = counts[identifier]["systems"][system]
                    share = counts[identifier]["changed"] / counts[identifier]["words"] if counts[identifier]["words"] else 0.0
                    entries.append({"removed": under, "hits": hits, "document": under * share,
                                    "words": counts[identifier]["words"], "changed": counts[identifier]["changed"]})
            if entries:
                entry["comparisons"][system] = trim_report(entries, False, False, f"{name}|{split}|{system}")
        if split == "test":
            entry["rule"] = trim_rule(models, distance)
        result["sets"][f"{name}/{split}"] = entry
    return result


def opener_results(root, documents):
    """Count the findings of the listed opener and how often the editor changed every flagged word."""
    if not (root / OPENERS).exists():
        return {}
    by_id = {row["id"]: row for row in documents if row["split"] in TRIM_SPLITS}
    found, seen = collections.defaultdict(collections.Counter), set()
    for row in rows(root / OPENERS):
        document = by_id.get(row.get("id"))
        if set(row) != {"id", "span", "words"} or document is None:
            fail(f"invalid opener row: {row.get('id')}")
        check_spans([row["span"], *row["words"]], row["id"])
        key = (row["id"], *row["span"])
        if key in seen or row["span"] not in document["systems"].get(OPENER_SYSTEM, []):
            fail(f"an opener of {row['id']} is repeated or is not a finding of {OPENER_SYSTEM}")
        seen.add(key)
        if not row["words"] or not all(within(word, [row["span"]]) for word in row["words"]):
            fail(f"an opener of {row['id']} needs words inside its span")
        counter = found[(document["set"], document["split"])]
        counter["findings"] += 1
        counter["changed"] += all(within(word, document["reference"]) for word in row["words"])
    result = {"sets": {}}
    for name, split in sorted({(row["set"], row["split"]) for row in by_id.values()}):
        counter = found[(name, split)]
        result["sets"][f"{name}/{split}"] = {
            "documents": sum(1 for row in by_id.values() if (row["set"], row["split"]) == (name, split)),
            "findings": counter["findings"], "changed": counter["changed"],
            "interval": wilson(counter["changed"], counter["findings"])}
    test = result["sets"].get(f"{OPENER_SET}/test", {"findings": 0, "changed": 0, "interval": [0.0, 0.0]})
    elsewhere = sum(entry["findings"] for key, entry in result["sets"].items() if key.split("/")[0] in OPENER_ELSEWHERE)
    result["elsewhere"] = elsewhere
    if (root / OPENER_PAGES).exists():
        result["load"] = opener_load(root)
    result["listed"] = bool(test["findings"]) and test["changed"] >= OPENER_SHARE * test["findings"] \
        and test["interval"][0] > OPENER_FLOOR and elsewhere <= OPENER_LOAD
    return result


def opener_load(root):
    """Return the counts of the rule's findings on documentation pages outside the record's sets."""
    load = {}
    for row in rows(root / OPENER_PAGES):
        # A part that was filtered by text first also says how many of its texts were run.
        names = ("pages", "failed", "findings", "candidates") if "candidates" in row else ("pages", "failed", "findings")
        counts = [row.get(name) for name in names]
        if set(row) != {"part", *names} or row["part"] in load:
            fail(f"invalid or repeated opener load row: {row.get('part')}")
        run = row.get("candidates", row.get("pages"))
        if not all(isinstance(value, int) and value >= 0 for value in counts) or row["failed"] > run or run > row["pages"]:
            fail(f"invalid counts in the opener load of {row['part']}")
        load[row["part"]] = {name: row[name] for name in names}
    return load


def rule_word_results(root, documents):
    """Return, for every rule on the test parts, the words under its findings against the words an editor changed."""
    if not (root / RULE_WORDS).exists():
        return {}
    counts = load_word_counts(root, documents)
    by_id = {row["id"]: row for row in documents if row["split"] in TRIM_SPLITS}
    groups, seen = collections.defaultdict(lambda: collections.defaultdict(list)), set()
    for row in rows(root / RULE_WORDS):
        document = by_id.get(row.get("id"))
        if set(row) != {"id", "rules"} or document is None or row["id"] in seen:
            fail(f"invalid or repeated rule word row: {row.get('id')}")
        seen.add(row["id"])
        total = counts[row["id"]]
        for name, pair in row["rules"].items():
            valid = isinstance(pair, list) and len(pair) == 2 and all(isinstance(v, int) for v in pair)
            if not valid or not 0 <= pair[1] <= pair[0] <= total["words"] or not pair[0]:
                fail(f"invalid counts of {name} for {row['id']}")
            if document["split"] == "test":
                share = total["changed"] / total["words"]
                groups[document["set"]][name].append({"removed": pair[0], "hits": pair[1], "document": pair[0] * share,
                                                      "words": total["words"], "changed": total["changed"]})
    if seen != set(by_id):
        fail("rule word rows do not cover the development and test documents")
    return {f"{name}/test": {rule: trim_report(entries, False, False, f"{name}|test|{rule}")
                             for rule, entries in sorted(rules.items())}
            for name, rules in sorted(groups.items())}


def vocabulary_phrases(root):
    """Return the list and the phrases that the selection part dropped."""
    listed = list(rows(root / VOCABULARY_PHRASES))
    if len(listed) != 1 or set(listed[0]) != {"phrases", "dropped"}:
        fail("the phrase list of the vocabulary test needs one row with phrases and dropped")
    phrases, dropped = listed[0]["phrases"], listed[0]["dropped"]
    if len(set(phrases)) != len(phrases) or not set(dropped) <= set(phrases) or len(set(dropped)) != len(dropped):
        fail("the dropped phrases of the vocabulary test are not distinct phrases of the list")
    return phrases, dropped


def vocabulary_document(row, phrases, gone, tallies):
    """Recount one answer: its findings by part, and the words under them that the editor changed."""
    counts, by_phrase, whole = tallies
    names = [name for name in counts if name.split("_")[0] == row["part"]]
    under, last = {name: {} for name in names}, -1
    for finding in row["findings"]:
        number = finding.get("phrase")
        if set(finding) != {"span", "phrase", "words"} or not finding["words"] or number not in range(len(phrases)):
            fail(f"invalid finding in vocabulary document {row['id']}")
        check_spans([finding["span"], *finding["words"]], row["id"])
        if finding["span"][0] < last or not all(within(word, [finding["span"]]) for word in finding["words"]):
            fail(f"findings of vocabulary document {row['id']} are out of order or hold words outside their span")
        last = finding["span"][0]
        changed = [within(word, row["reference"]) for word in finding["words"]]
        if row["part"] == "confirmation" and number in gone:
            fail(f"vocabulary document {row['id']} of the confirmation part has a finding of a dropped phrase")
        if row["part"] == "selection":
            by_phrase[number] += 1
            whole[number] += all(changed)
        # The selection part counts twice: with every phrase, and with the phrases that stayed.
        for name in names:
            if name == "selection_all" or number not in gone:
                counts[name]["findings"] += 1
                counts[name]["changed"] += all(changed)
                under[name].update({tuple(word): flag for word, flag in zip(finding["words"], changed)})
    share = row["changed"] / row["words"]
    return {name: {"removed": len(under[name]), "hits": sum(under[name].values()), "document": len(under[name]) * share,
                   "words": row["words"], "changed": row["changed"]} for name in names}


def vocabulary_load(root):
    loaded = list(rows(root / VOCABULARY_LOAD))
    keys = {"pages", "failed", "pages_with_finding", "findings", "words"}
    if len(loaded) != 1 or set(loaded[0]) != keys or not all(isinstance(v, int) and v >= 0 for v in loaded[0].values()):
        fail("the load of the vocabulary test needs one row of counts")
    if loaded[0]["pages_with_finding"] > loaded[0]["pages"] or not loaded[0]["words"]:
        fail("the load of the vocabulary test is inconsistent")
    return loaded[0]


def vocabulary_results(root, documents):
    """Recount the findings of a wide phrase list on answers that the other results did not use."""
    if not (root / VOCABULARY).exists():
        return {}
    phrases, dropped = vocabulary_phrases(root)
    gone = {phrases.index(phrase) for phrase in dropped}
    used = {row["text_sha256"] for row in documents}
    names = (*VOCABULARY_PARTS, "selection_all")
    entries = {name: [] for name in names}
    tallies = ({name: collections.Counter() for name in names}, collections.Counter(), collections.Counter())
    seen = set()
    for row in rows(root / VOCABULARY):
        if set(row) != VOCABULARY_KEYS or row["part"] not in VOCABULARY_PARTS or row["id"] in seen:
            fail(f"invalid or repeated vocabulary row: {row.get('id')}")
        seen.add(row["id"])
        if row["text_sha256"] in used:
            fail(f"vocabulary document {row['id']} is a text of the other results")
        valid = isinstance(row["words"], int) and isinstance(row["changed"], int) and 0 <= row["changed"] <= row["words"]
        if not valid or not row["words"]:
            fail(f"invalid word counts for vocabulary document {row['id']}")
        check_spans(row["reference"], row["id"])
        for name, entry in vocabulary_document(row, phrases, gone, tallies).items():
            entries[name].append(entry)
    counts, by_phrase, whole = tallies
    expected = {number for number in by_phrase
                if by_phrase[number] >= VOCABULARY_MIN_FINDINGS and whole[number] < VOCABULARY_KEEP * by_phrase[number]}
    if expected != gone:
        fail("the dropped phrases do not follow from the selection part")
    result = {"list": {"phrases": len(phrases), "dropped": sorted(dropped), "kept": len(phrases) - len(dropped)}, "parts": {}}
    for name in names:
        report = trim_report(entries[name], False, False, f"vocabulary|{name}")
        report.update(findings=counts[name]["findings"], findings_changed=counts[name]["changed"],
                      findings_interval=wilson(counts[name]["changed"], counts[name]["findings"]))
        result["parts"][name] = report
    confirmed = result["parts"]["confirmation"]
    words = confirmed["intervals"].get("precision_minus_document", [float("-inf")])[0] > 0
    enough = bool(confirmed["findings"]) and confirmed["findings_changed"] >= VOCABULARY_SHARE * confirmed["findings"] \
        and confirmed["findings_interval"][0] > VOCABULARY_FLOOR
    result["rule"] = {"words_above_chance": words, "findings_changed_enough": enough, "ships": words and enough}
    if (root / VOCABULARY_LOAD).exists():
        result["load"] = vocabulary_load(root)
    return result


def deslop_share(pairs, keep, listed=False):
    """Compare the findings of the kept rules per 1,000 words before and after each commit.

    The document-level rule adds one finding to a set. With listed, it is counted by the em dashes it lists.
    """
    def total(counts):
        return sum(count if listed or name != DESLOP_DOCUMENT else 1 for name, count in counts.items() if keep(name))

    lower = higher = same = before = after = 0
    for pair in pairs:
        b, a = total(pair["before"]), total(pair["after"])
        if not a and not b:
            continue
        before += b
        after += a
        # Cross-multiplied, so that the comparison of two rates stays in whole numbers.
        left, right = a * pair["words_before"], b * pair["words_after"]
        lower += left < right
        higher += left > right
        same += left == right
    return {"pairs": len(pairs), "fires": lower + higher + same, "lower": lower, "higher": higher, "same": same,
            "findings_before": before, "findings_after": after, "interval": wilson(lower, lower + higher)}


def deslop_load(path, label):
    """Read the counts of one rule on the kinds of documentation."""
    load = {}
    for row in rows(path):
        counts = [row.get(name) for name in ("pages", "pages_with_finding", "findings", "words")]
        valid = set(row) == {"corpus", "pages", "pages_with_finding", "findings", "words"} and row["corpus"] not in load
        if not valid or not all(isinstance(v, int) and v >= 0 for v in counts) or row["pages_with_finding"] > row["pages"] or not row["words"]:
            fail(f"invalid {label} load row: {row.get('corpus')}")
        load[row["corpus"]] = {name: row[name] for name in ("pages", "pages_with_finding", "findings", "words")}
    return load


def deslop_results(root):
    """Recount, per part, how the findings of rule sets change across commits that remove AI-sounding wording."""
    if not (root / DESLOP).exists():
        return {}
    parts, seen = {"development": [], "test": []}, set()
    for row in rows(root / DESLOP):
        if set(row) != DESLOP_KEYS or row["part"] not in parts or row["id"] in seen:
            fail(f"invalid or repeated de-slop pair: {row.get('id')}")
        seen.add(row["id"])
        counts = [*row["before"].values(), *row["after"].values(), row["words_before"], row["words_after"]]
        if not all(isinstance(value, int) and value > 0 for value in counts) or not isinstance(row["english"], bool):
            fail(f"invalid counts in de-slop pair {row['id']}")
        parts[row["part"]].append(row)
    extra = (*DESLOP_NEW, DESLOP_REGISTERED)
    sets = {"alpha9_defaults": lambda name: name not in extra,
            "with_decorative_emoji": lambda name: name not in ("format.em-dash-density", DESLOP_DOCUMENT, DESLOP_REGISTERED),
            "current_defaults": lambda name: name not in ("format.em-dash-density", DESLOP_REGISTERED),
            "format.decorative-emoji": lambda name: name == "format.decorative-emoji",
            DESLOP_REGISTERED: lambda name: name == DESLOP_REGISTERED,
            "format.em-dash-density": lambda name: name == "format.em-dash-density"}
    def shares(pairs):
        table = {name: deslop_share(pairs, keep) for name, keep in sets.items()}
        table[DESLOP_DOCUMENT] = deslop_share(pairs, lambda name: name == DESLOP_DOCUMENT, listed=True)
        return table

    result = {"parts": {part: shares(pairs) for part, pairs in parts.items()},
              "repositories": {part: len({pair["repository"] for pair in pairs}) for part, pairs in parts.items()},
              "web_edits": {part: shares([p for p in pairs if p["web_edit"]]) for part, pairs in parts.items()},
              # The pairs were not selected by language. The tool is for English, so the same tables are kept for English files.
              "english": {part: shares([p for p in pairs if p["english"]]) for part, pairs in parts.items()}}

    def passes(entry, bar):
        changed = entry["lower"] + entry["higher"]
        return bool(changed) and entry["lower"] >= bar * changed and entry["interval"][0] > DESLOP_FLOOR

    test = result["parts"]["test"]
    result["rule"] = {name: passes(test[name], DESLOP_RULE_BAR) for name in (*DESLOP_NEW, DESLOP_REGISTERED)}
    if (root / EMOJI_LOAD).exists():
        result["load"] = deslop_load(root / EMOJI_LOAD, "emoji")
    if (root / FREQUENCY_LOAD).exists():
        result["frequency_load"] = deslop_load(root / FREQUENCY_LOAD, "em dash")
    people = result.get("frequency_load", {}).get("people")
    result["rule"]["frequency_within_human_limit"] = bool(people) and people["pages_with_finding"] * DESLOP_HUMAN_PAGES <= people["pages"]
    result["rule"][DESLOP_DOCUMENT] = result["rule"][DESLOP_DOCUMENT] and result["rule"]["frequency_within_human_limit"]
    result["rule"]["defaults_follow_removal"] = passes(test["with_decorative_emoji"], DESLOP_SET_BAR)
    english = result["english"]["test"]
    result["rule"]["english"] = {name: passes(english[name], DESLOP_RULE_BAR) for name in ("format.decorative-emoji", DESLOP_DOCUMENT)}
    return result


def review_row_valid(row):
    """Report whether the counts of one text under one answer protocol are consistent."""
    pairs = [row["findings"], row["flagged"], row["flagged_changed"], row["sentences"]]
    if not all(isinstance(pair, list) and len(pair) == 2 and all(isinstance(v, int) and v >= 0 for v in pair) for pair in pairs):
        return False
    if not all(isinstance(row[name], int) for name in ("words", "changed", "both")) or not 0 <= row["changed"] <= row["words"]:
        return False
    later = (row["after_words"], row["after_findings"])
    if row["protocol"] == "list":
        if later != (None, None) or row["edit"] is not None:
            return False
    elif not all(isinstance(v, int) and v >= 0 for v in later) or not row["after_words"] or not review_edit_valid(row["edit"]):
        return False
    return (row["words"] > 0 and 0 <= row["both"] <= min(row["sentences"])
            and all(c <= f <= row["words"] for c, f in zip(row["flagged_changed"], row["flagged"])))


def review_edit_valid(edit):
    """Report whether the counts of what a commit did to the flagged sentences of one text are consistent."""
    if not isinstance(edit, dict) or set(edit) != set(REVIEW_EDIT_KEYS) or not all(isinstance(v, int) and v >= 0 for v in edit.values()):
        return False
    return (edit["sentences_changed"] <= edit["sentences"] and edit["changed_by_commit"] <= edit["flagged_before"] <= edit["sentences"]
            and edit["on_new_text"] + edit["flagged_both_times"] <= edit["flagged_after"]
            and edit["changed_by_commit"] <= edit["sentences_changed"])


def repeat_rate(sample):
    """Of the sentences flagged in one run, the share flagged in the other run of the same text."""
    flagged = sum(sum(row["sentences"]) for row in sample)
    return 2 * sum(row["both"] for row in sample) / flagged if flagged else 0.0


def over_chance(sample):
    """Share of the flagged words of the first run that the commit changed, and the chance of the same documents."""
    flagged = sum(row["flagged"][0] for row in sample)
    if not flagged:
        return None
    hits = sum(row["flagged_changed"][0] for row in sample)
    chance = sum(row["flagged"][0] * row["changed"] / row["words"] for row in sample)
    return hits / flagged, chance / flagged


def draws_of(sample, statistic):
    """Bootstrap over documents, in the way the registered comparison drew them."""
    generator, values = random.Random(REVIEW_REPEAT_SEED), []
    for _ in range(BOOTSTRAP):
        value = statistic([sample[generator.randrange(len(sample))] for _ in sample])
        if value is not None:
            values.append(value)
    return sorted(values)


def review_protocol(sample):
    precision, chance = over_chance(sample) or (0.0, 0.0)
    draws = draws_of(sample, lambda drawn: None if over_chance(drawn) is None else over_chance(drawn)[0] - over_chance(drawn)[1])
    return {"documents": len(sample), "sentences": [sum(row["sentences"][run] for row in sample) for run in (0, 1)],
            "changed_words_flagged": sum(row["flagged_changed"][0] for row in sample),
            "both": sum(row["both"] for row in sample), "repeat_rate": repeat_rate(sample),
            "findings": [sum(row["findings"][run] for row in sample) for run in (0, 1)],
            "flagged_words": sum(row["flagged"][0] for row in sample), "precision": precision, "chance": chance,
            "interval": [draws[int(0.025 * len(draws))], draws[int(0.975 * len(draws)) - 1]] if draws else [0.0, 0.0]}


def review_movement(sample):
    """Compare the findings per 1,000 words of the first run with the findings on the text after the commit."""
    lower = higher = same = after = 0
    for row in sample:
        before, later = row["findings"][0], row["after_findings"]
        after += later
        if not before and not later:
            continue
        left, right = later * row["words"], before * row["after_words"]
        lower += left < right
        higher += left > right
        same += left == right
    edit = {name: sum(row["edit"][name] for row in sample) for name in REVIEW_EDIT_KEYS}
    return {"lower": lower, "higher": higher, "same": same, "findings_after": after, "interval": wilson(lower, lower + higher),
            "edit": edit}


def review_repeat_results(root):
    """Recount how often a reviewer flags the same sentences twice under each of two answer protocols."""
    if not (root / REVIEW_REPEAT).exists():
        return {}
    table, seen = {}, set()
    for row in rows(root / REVIEW_REPEAT):
        key = (row.get("reviewer"), row.get("protocol"), row.get("document"))
        if set(row) != REVIEW_REPEAT_KEYS or row["protocol"] not in REVIEW_PROTOCOLS or key in seen:
            fail(f"invalid or repeated review row: {row.get('document')}")
        seen.add(key)
        if not review_row_valid(row):
            fail(f"inconsistent review row: {row['document']}")
        table.setdefault(row["reviewer"], {}).setdefault(row["protocol"], {})[row["document"]] = row
    result = {}
    for reviewer, protocols in sorted(table.items()):
        if set(protocols) != set(REVIEW_PROTOCOLS) or any(set(protocols[name]) != set(protocols["list"]) for name in REVIEW_PROTOCOLS):
            fail(f"review rows of {reviewer} do not pair the protocols")
        documents = sorted(protocols["list"])
        old = [protocols["list"][name] for name in documents]
        entry = {"list": review_protocol(old)}
        for protocol in ("verdicts", "strict"):
            new = [protocols[protocol][name] for name in documents]
            draws = draws_of(list(zip(new, old)), lambda drawn: repeat_rate([p[0] for p in drawn]) - repeat_rate([p[1] for p in drawn]))
            entry[protocol] = review_protocol(new)
            entry[protocol].update(movement=review_movement(new), difference=repeat_rate(new) - repeat_rate(old),
                                   difference_interval=[draws[49], draws[BOOTSTRAP - 51]])
        chosen = entry["verdicts"]
        # The verdict form replaces the list form when it repeats more, still points at changed words, and finds no less.
        entry["rule"] = {"repeats_more": chosen["difference_interval"][0] > 0, "on_changed_words": chosen["interval"][0] > 0,
                         "finds_no_less": chosen["changed_words_flagged"] >= entry["list"]["changed_words_flagged"]}
        entry["rule"]["ships"] = all(entry["rule"].values())
        result[reviewer] = entry
    return result


def summarize(root):
    documents = load_documents(root)
    preferences = preference_results(root)
    experts = expert_results(root)
    return {"sentences": sentence_results(documents), "preferences": preferences, "expert_pairs": experts,
            "judges": admissions(preferences, experts), "readers": reader_results(root), "panel": panel_results(root),
            "trims": trim_results(root, documents), "openers": opener_results(root, documents),
            "rule_words": rule_word_results(root, documents), "vocabulary": vocabulary_results(root, documents),
            "deslop": deslop_results(root), "review_repeat": review_repeat_results(root)}


def normalized(value):
    return json.loads(json.dumps(value, sort_keys=True))


def verify(root=HERE):
    """Raise ValueError when the record at root is inconsistent."""
    root = Path(root)
    check_digests(root)
    published = json.loads((root / RESULTS).read_text())
    if normalized(summarize(root)) != normalized(published):
        fail("results differ from the data files")
    import render
    for page, function in render.PAGES.items():
        rendered = root / page
        if rendered.exists() and rendered.read_text() != function(published):
            fail(f"{page} differs from the rendering of the results")
        if page == render.WORDS_PAGE and render.has_words(published) and (root / "RESULTS.md").exists() and not rendered.exists():
            fail(f"{page} is missing")
    return published


def main(arguments):
    write = "--write" in arguments
    paths = [a for a in arguments if a != "--write"]
    root = Path(paths[0]) if paths else HERE
    if write:
        import render
        results = summarize(root)
        (root / RESULTS).write_text(json.dumps(results, indent=1, sort_keys=True) + "\n")
        for page, function in render.PAGES.items():
            if page == "RESULTS.md" or render.has_words(results):
                (root / page).write_text(function(normalized(results)))
        write_digests(root)
    results = verify(root)
    documents = sum(entry["documents"] for entry in results["sentences"]["sets"].values())
    print(f"record verified: {documents} documents, {results['preferences']['triplets']} preference triplets")


if __name__ == "__main__":
    main(sys.argv[1:])
