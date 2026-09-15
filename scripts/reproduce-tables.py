#!/usr/bin/env python3
"""Rebuild every table of the evidence release from saved records.

The screening records are committed, so the evidence cards and the
confirmatory result rebuild from a clean checkout. The corpus tables read
artifacts/measurement, which "bash scripts/measure-corpus.sh" produces and
which is not committed; without it those tables report what is missing
instead of printing a number no record supports.
"""

import collections
import json
from pathlib import Path
import sys

SCREENINGS = Path("research/methods/screening")
MEASUREMENT = Path("artifacts/measurement/findings")
BANDS = ((0, 50, "under 50"), (50, 150, "50-149"), (150, 400, "150-399"),
         (400, 1200, "400-1199"), (1200, None, "1200+"))


def band(words):
    for low, high, name in BANDS:
        if words >= low and (high is None or words < high):
            return name
    return BANDS[-1][2]


def arm(cohort):
    if cohort == "controlled":
        return "generated"
    if cohort.startswith("historical"):
        return "human"
    return cohort


def documents(root):
    """Yield one record per measured document, or nothing when the corpus is absent."""
    if not root.is_dir():
        return
    for shard in sorted(root.iterdir()):
        if shard.suffix != ".json":
            continue
        for document in json.loads(shard.read_text()).get("documents", []):
            yield shard.name, document


def rows(counts, header, order=None):
    keys = order or sorted(counts)
    width = max(len(str(k)) for k in keys) if keys else 5
    print("| " + " | ".join([header[0].ljust(width)] + header[1:]) + " |")
    print("| " + " | ".join(["-" * max(3, len(h)) for h in header]) + " |")
    for key in keys:
        n, words, findings, any_finding = counts[key]
        rate = findings / words * 1000 if words else 0
        print(f"| {str(key).ljust(width)} | {n} | {words} | {findings} | "
              f"{findings / n:.2f} | {rate:.3f} | {any_finding / n * 100:.1f}% |")


def composition():
    counts = collections.defaultdict(lambda: [0, 0, 0, 0])
    for _, document in documents(MEASUREMENT):
        entry = counts[document["cohort"]]
        entry[0] += 1
        entry[1] += document.get("prose_words", 0)
        entry[2] += document.get("findings", 0)
        entry[3] += 1 if document.get("findings", 0) else 0
    print("\n## Table 1. Corpus composition\n")
    if not counts:
        print(missing())
        return
    rows(counts, ["Cohort", "Documents", "Prose words", "Findings", "Per document",
                  "Per 1000 words", "Share with any"])


def by_length():
    counts = collections.defaultdict(lambda: [0, 0, 0, 0])
    for _, document in documents(MEASUREMENT):
        key = arm(document["cohort"]) + ", " + band(document.get("prose_words", 0))
        entry = counts[key]
        entry[0] += 1
        entry[1] += document.get("prose_words", 0)
        entry[2] += document.get("findings", 0)
        entry[3] += 1 if document.get("findings", 0) else 0
    print("\n## Table 2. Document-level behavior by length band\n")
    if not counts:
        print(missing())
        return
    names = [b[2] for b in BANDS]
    order = sorted(counts, key=lambda k: (k.split(", ")[0], names.index(k.split(", ")[1])))
    rows(counts, ["Arm and band", "Documents", "Prose words", "Findings", "Per document",
                  "Per 1000 words", "Share with any"], order)


def by_rule():
    counts = collections.defaultdict(collections.Counter)
    words = collections.Counter()
    for _, document in documents(MEASUREMENT):
        side = arm(document["cohort"])
        words[side] += document.get("prose_words", 0)
        for rule, value in (document.get("by_rule") or {}).items():
            counts[side][rule] += value
    print("\n## Table 3. Findings per thousand words, by rule\n")
    if not words:
        print(missing())
        return
    names = sorted({rule for side in counts for rule in counts[side]})
    print("| Rule | Human | Generated | Contemporary |")
    print("| --- | ---: | ---: | ---: |")
    for rule in names:
        cells = []
        for side in ("human", "generated", "contemporary"):
            total = words[side]
            cells.append(f"{counts[side][rule] / total * 1000:.3f}" if total else "n/a")
        print(f"| `{rule}` | " + " | ".join(cells) + " |")


def cards():
    """Print one row per rule, family and screening: the evidence cards."""
    print("\n## Table 4. Evidence cards\n")
    files = sorted(p for p in SCREENINGS.glob("*.json") if "dataset-plan" not in p.name)
    if not files:
        sys.exit("no screening record under " + str(SCREENINGS))
    print("| Screening | Rule | Family | State | Reason | Difference | p |")
    print("| --- | --- | --- | --- | --- | ---: | ---: |")
    for path in files:
        record = json.loads(path.read_text())
        for rule in record.get("rules", []):
            for family in rule.get("families", []):
                state = family.get("state", "")
                if state in ("", "inconclusive") and family.get("reason") == "zero_count":
                    continue
                difference = family.get("difference", {}).get("value")
                shown = f"{difference * 100:+.2f} points" if isinstance(difference, (int, float)) else ""
                print(f"| {path.stem} | `{rule['rule_id']}` | {family.get('family', '')} | {state} | "
                      f"{family.get('reason', '')} | {shown} | {family.get('p_value', ''):.3f} |")


def missing():
    return ("Not rebuilt: `artifacts/measurement` is absent. Run "
            "`bash scripts/measure-corpus.sh` first; it is not committed because "
            "the corpus exceeds what this repository stores.")


def self_test():
    assert band(0) == "under 50" and band(49) == "under 50"
    assert band(50) == "50-149" and band(1200) == "1200+" and band(10 ** 9) == "1200+"
    assert arm("controlled") == "generated"
    assert arm("historical-2012") == "human"
    assert arm("contemporary") == "contemporary"
    assert list(documents(Path("no/such/directory"))) == []
    assert "measure-corpus.sh" in missing()
    print("Band, arm, and absent-corpus handling behave as the tables require.")


def main(arguments):
    if arguments == ["--self-test"]:
        self_test()
        return
    if arguments:
        sys.exit("Usage: reproduce-tables.py [--self-test]")
    print("# Evidence tables\n")
    print("Rebuilt by `python3 scripts/reproduce-tables.py`.")
    composition()
    by_length()
    by_rule()
    cards()


if __name__ == "__main__":
    main(sys.argv[1:])
