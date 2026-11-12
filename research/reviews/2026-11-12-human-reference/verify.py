#!/usr/bin/env python3
"""Verify the measurement against human references offline.

usage: verify.py [--write] [record directory]

The record holds coordinates and choices, never text. The check covers file
digests, the schema, and every published number: the sentence-level rates of
each system against the sentences an editor changed, the agreement between
editors, the agreement of judge models with the writers' preferences, and the
admission of judges that follows. RESULTS.md must equal the rendering of the
results. With --write the results, their rendering, and the digests are
regenerated from the data files.
"""
import collections
import gzip
import hashlib
import itertools
import json
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
DIGESTS = "files.sha256"
RESULTS = "results.json"
DATA = ("data/sentences.jsonl.gz", "data/preferences.jsonl.gz", "data/expert-pairs.jsonl.gz", "data/readers.jsonl.gz",
        "data/panel.jsonl.gz")
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
WEIGHTS = {TOP_JUDGE: 2}
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
    """The answer with the most votes; the top model has two votes and decides a tie."""
    tally = collections.Counter()
    for judge, vote in votes.items():
        tally[vote] += WEIGHTS.get(judge, 1)
    best = max(tally.values())
    leaders = [vote for vote, count in tally.items() if count == best]
    return leaders[0] if len(leaders) == 1 else votes[TOP_JUDGE]


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
        for name in names:
            entry[f"{name}_meets_rule"] = meets(summaries[name])
        if names == ("reviewed", "control"):
            reviewed, control = summaries["reviewed"], summaries["control"]
            share = reviewed["revised"] / reviewed["decisive"] if reviewed["decisive"] else 0.0
            other = control["revised"] / control["decisive"] if control["decisive"] else 0.0
            entry["ahead_of_control"] = share - other >= LEAD
        result[comparison] = entry
    return result


def summarize(root):
    documents = load_documents(root)
    preferences = preference_results(root)
    experts = expert_results(root)
    return {"sentences": sentence_results(documents), "preferences": preferences, "expert_pairs": experts,
            "judges": admissions(preferences, experts), "readers": reader_results(root), "panel": panel_results(root)}


def normalized(value):
    return json.loads(json.dumps(value, sort_keys=True))


def verify(root=HERE):
    """Raise ValueError when the record at root is inconsistent."""
    root = Path(root)
    check_digests(root)
    published = json.loads((root / RESULTS).read_text())
    if normalized(summarize(root)) != normalized(published):
        fail("results differ from the data files")
    rendered = root / "RESULTS.md"
    if rendered.exists():
        import render
        if rendered.read_text() != render.render(published):
            fail("RESULTS.md differs from the rendering of the results")
    return published


def main(arguments):
    write = "--write" in arguments
    paths = [a for a in arguments if a != "--write"]
    root = Path(paths[0]) if paths else HERE
    if write:
        import render
        results = summarize(root)
        (root / RESULTS).write_text(json.dumps(results, indent=1, sort_keys=True) + "\n")
        (root / "RESULTS.md").write_text(render.render(normalized(results)))
        write_digests(root)
    results = verify(root)
    documents = sum(entry["documents"] for entry in results["sentences"]["sets"].values())
    print(f"record verified: {documents} documents, {results['preferences']['triplets']} preference triplets")


if __name__ == "__main__":
    main(sys.argv[1:])
