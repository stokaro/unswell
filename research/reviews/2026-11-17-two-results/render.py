#!/usr/bin/env python3
"""Write RESULTS.md of the record of the two results from the recounted numbers.

verify.py calls render() with what it recounted from the data; this file holds no number of its own.
"""

PAGES = {"set-s": "posts", "fresh": "fresh texts"}
KINDS = {
    "sentence": "A sentence removed", "negation": '"not" removed', "addnot": '"not" put in', "swap": "A word turned around",
    "clause": "A clause cut off", "tail": "Last words cut off", "join": "Two sentences joined (harmless)",
    "contract": "Contractions (harmless)", "trim": "Trimmed (harmless)",
}
CHECKS = {"statements": "statements", "sentences": "sentences", "losses": "numbers, code, or links", "dashes": "em dashes"}
BARS = {"G1": "Harmless edits pass", "G2": "Damage does not", "G3": "The result repeats", "G4": "Verdicts"}


def percent(value):
    return f"{100 * value:.1f}%"


def cell(value):
    return f"{value['count']} of {value['of']}, {percent(value['share'])}"


def limit(bar):
    return f"{bar['kind']} {percent(bar['limit'])}" if isinstance(bar["limit"], float) else f"{bar['kind']} {bar['limit']}"


def bars_table(result):
    lines = ["| Bar | Judge | Read on | Number | Limit | Holds |", "| --- | --- | --- | --- | --- | --- |"]
    for bar in result["bars"]:
        read_on = bar["read_on"]
        for key, label in PAGES.items():
            read_on = read_on.replace(key, "the " + label)
        number = cell(bar["value"]) if bar["bar"] != "G2" else f"{bar['value']['count']} of {bar['value']['of']}"
        holds = "yes" if bar["holds"] else "**no**"
        lines.append(f"| {bar['bar']}. {BARS[bar['bar']]} | `{bar['judge']}` | {read_on} | {number} | {limit(bar)} | {holds} |")
    return lines


def copies_table(result):
    judges = result["judges"]
    lines = ["| Kind of copy | " + " | ".join(f"Nothing lost, `{judge}`" for judge in judges) + " |", "| --- |" + " ---: |" * len(judges)]
    for kind, label in KINDS.items():
        row = []
        for judge in judges:
            found = result["copies"][judge]["by_kind"][kind]
            text = f"{found['nothing_lost']} of {found['copies']}"
            if found["without_verdict"]:
                text += f" ({found['without_verdict']} without a verdict)"
            row.append(text)
        lines.append(f"| {label} | " + " | ".join(row) + " |")
    return lines


def failures_table(result):
    judges = result["judges"]
    lines = ["| Harmless copies that lost something, by the check that failed | " + " | ".join(f"`{judge}`" for judge in judges) + " |",
             "| --- |" + " ---: |" * len(judges)]
    for check, label in CHECKS.items():
        row = [str(sum(result["copies"][judge]["by_kind"][kind]["failed"][check] for kind in ("join", "contract", "trim"))) for judge in judges]
        lines.append(f"| {label[0].upper() + label[1:]} | " + " | ".join(row) + " |")
    return lines


def revisions_table(result):
    lines = ["| Judge | Run | Revisions | Nothing lost | Accepted with all required | Statements found kept |",
             "| --- | ---: | ---: | ---: | ---: | --- |"]
    for judge in result["judges"]:
        for run, found in result["revisions"][judge]["runs"].items():
            lines.append(f"| `{judge}` | {run} | {found['verdicts']['of']} | {found['nothing_lost']} | {found['accepted_all']} | "
                         f"{cell(found['statements'])} |")
    return lines


def repeat_table(result):
    lines = ["| Judge | Same result on nothing lost | Nothing lost in both runs | Same verdict of the blind comparison | "
             "Accepted with all required in both runs |", "| --- | --- | ---: | --- | ---: |"]
    for judge in result["judges"]:
        found = result["revisions"][judge]
        lines.append(f"| `{judge}` | {cell(found['repeat'])} | {found['nothing_lost_both']} | {cell(found['sound_repeat'])} | "
                     f"{found['accepted_all_both']} |")
    return lines


def pages_table(result):
    lines = ["| Judge | Run | Pages | Revisions | Nothing lost | Accepted with all required |", "| --- | ---: | --- | ---: | ---: | ---: |"]
    for judge in result["judges"]:
        for run, found in result["revisions"][judge]["runs"].items():
            for pages, label in PAGES.items():
                part = found["by_pages"][pages]
                lines.append(f"| `{judge}` | {run} | {label} | {part['revisions']} | {part['nothing_lost']} | {part['accepted_all']} |")
    return lines


def fitting_tables(result):
    fitting = result["fitting"]
    lines = ["| Judge | Nothing lost, first run | Second run | Same result in both | Accepted with all required, first run | Second run | In both |",
             "| --- | ---: | ---: | --- | ---: | ---: | ---: |"]
    for judge in result["judges"]:
        found = fitting["revisions"][judge]
        lines.append(f"| `{judge}` | {found['runs']['1']['nothing_lost']} | {found['runs']['2']['nothing_lost']} | {cell(found['repeat'])} | "
                     f"{found['runs']['1']['accepted_all']} | {found['runs']['2']['accepted_all']} | {found['accepted_all_both']} |")
    lines += ["", "| Judge | Build | Harmless copies with a verdict | Nothing lost | Contractions | Two sentences joined | "
              "Accepted with all required |", "| --- | --- | ---: | ---: | --- | --- | ---: |"]
    for found in fitting["harmless"]:
        build = "this record" if found["build"] == "record" else "`v0.1.0-alpha.17`"
        kinds = found["by_kind"]
        lines.append(f"| `{found['judge']}` | {build} | {found['with_verdict']} | {found['nothing_lost']} | "
                     f"{kinds['contract']['nothing_lost']} of {kinds['contract']['copies']} | "
                     f"{kinds['join']['nothing_lost']} of {kinds['join']['copies']} | {found['accepted_all']} |")
    return lines


def parts_text(judge, found):
    """What a run that was made in parts found in each of them."""
    parts = found["parts"]
    sizes = ", ".join(str(part["copies"]) for part in parts[:-1]) + f", and {parts[-1]['copies']}"
    text = (f"The run of the copies with `{judge}` was made in {len(parts)} parts, of {sizes} copies. Between them that judge was "
            "not available, and the run went on with the same build when it was again. A copy belongs to the part that gave it "
            "its verdict. The copies were judged in the order of the file, so the parts hold different pages.")
    if found["retried"]:
        text += (f" {found['retried']} checks got no verdict on the first attempt and were run once more, as the rule for a pair "
                 "without a verdict says.")
    lines = [text, "", "| Part | Copies | Harmless copies that lose nothing | Damaged copies that lose nothing | With a verdict |",
             "| ---: | ---: | --- | --- | --- |"]
    for part in parts:
        lines.append(f"| {part['part']} | {part['copies']} | {part['harmless']['count']} of {part['harmless']['of']} | "
                     f"{part['damaged']['count']} of {part['damaged']['of']} | {part['verdicts']['count']} of {part['verdicts']['of']} |")
    return lines


def first_attempts(result):
    """The verdicts before any retry, under the bars that are read after it."""
    copies = ", ".join(f"{cell(result['copies'][judge]['first_attempt'])} of the copies with `{judge}`" for judge in result["judges"])
    runs = [(judge, run, found["first_attempt"]) for judge in result["judges"] for run, found in result["revisions"][judge]["runs"].items()]
    if all(first["count"] == first["of"] for _, _, first in runs):
        revisions = "every revision in each run of each judge"
    else:
        revisions = ", ".join(f"{cell(first)} of the revisions in run {run} of `{judge}`" for judge, run, first in runs)
    return ("The bars of verdicts are read after the one retry that a pair without a verdict gets. A verdict on the first attempt: "
            f"{copies}, and {revisions}.")


def controls_table(result):
    lines = ["| Revisions of | Judge | Family of the judge | Revisions | Nothing lost |", "| --- | --- | --- | ---: | ---: |"]
    for found in result["controls"]:
        lines.append(f"| `{found['reviser']}` | `{found['judge_of_other_family']}` | the other | {found['revisions']} | {found['other_family']} |")
        lines.append(f"| `{found['reviser']}` | `{found['judge_of_own_family']}` | its own | {found['revisions']} | {found['own_family']} |")
    return lines


def controls_text(result):
    lines = []
    for found in result["controls"]:
        pages = " ".join(f"Of the {part['revisions']} revisions of {PAGES[key]}, they find it in {part['other_family']} and in "
                         f"{part['own_family']}." for key, part in found["by_pages"].items())
        missing = f" {found['without_verdict']} of its checks ended without a verdict and are left out." if found["without_verdict"] else ""
        lines += ["", f"On the revisions of `{found['reviser']}` the two judges agree on {cell(found['agree'])}, and both find nothing "
                  f"lost in {found['both']}. {pages}{missing}"]
    return lines


def render(result):
    lines = ["# Results: two results of the check of a revision", "",
             "Generated by `verify.py --write` from the data of this record. Do not edit.", "",
             '"Nothing lost" is the first result of `unswell compare`: the statements, the sentences, the numbers, code, and links of '
             "the page are kept, and no em dash is added. The second is the blind comparison of the sound. With all required, a "
             "revision is accepted when both hold.", "",
             "## The bars", "",
             f"{result['bars_held']} of the {result['bars_total']} bars hold. They were written on issue 605 before the run on the "
             "measured part.", ""]
    lines += bars_table(result)
    lines += ["", first_attempts(result)]
    for judge in result["judges"]:
        if len(result["copies"][judge]["parts"]) > 1:
            lines += [""] + parts_text(judge, result["copies"][judge])
    lines += ["", "## Copies of the measured pages", "",
              "One sentence of a page is changed in each copy. A damaged copy lost something, and a harmless one did not.", ""]
    lines += copies_table(result)
    lines += [""] + failures_table(result)
    accepted = ", ".join(f"{result['copies'][judge]['harmless_accepted_all']} of {result['copies'][judge]['harmless']['of']} with `{judge}`"
                         for judge in result["judges"])
    lines += ["", f"With all required, the check accepts these of the harmless copies: {accepted}."]
    lines += ["", "## Revisions, checked twice", "",
              "Each judge checked the revisions of the other reviser: 27 of posts and 30 of fresh texts, in two runs of the same build.",
              ""]
    lines += revisions_table(result)
    lines += [""] + repeat_table(result)
    lines += [""] + pages_table(result)
    lines += ["", "## The same revisions under both judges", "",
              "The two judges differ much in how many revisions lose nothing, and each had checked the revisions of the other "
              "reviser. As controls after the run, each judge checked the revisions of its own family that the other judge had "
              "checked in its first run. This is no check of a revision: a judge of the reviser's family is what the command warns "
              "against.", ""]
    lines += controls_table(result)
    lines += controls_text(result)
    lines += ["", "## The fitting half", "",
              "The 40 revisions of each reviser in two runs, and 75 harmless copies of the fitting posts.", ""]
    lines += fitting_tables(result)
    return "\n".join(lines) + "\n"
