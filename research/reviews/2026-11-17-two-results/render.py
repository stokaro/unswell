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
        if bar["holds"] is None:
            # The judge did not see every copy that the bar is about.
            number = f"{bar['value']['count']} of {bar['value']['of']} judged, of {bar['of_all']}"
        holds = "not read" if bar["holds"] is None else ("yes" if bar["holds"] else "**no**")
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


def render(result):
    lines = ["# Results: two results of the check of a revision", "",
             "Generated by `verify.py --write` from the data of this record. Do not edit.", "",
             '"Nothing lost" is the first result of `unswell compare`: the statements, the sentences, the numbers, code, and links of '
             "the page are kept, and no em dash is added. The second is the blind comparison of the sound. With all required, a "
             "revision is accepted when both hold.", "",
             "## The bars", "",
             f"{result['bars_read']} of the {result['bars_total']} bars can be read, and {result['bars_held']} of them hold. They were "
             "written on issue 605 before the run on the measured part.", ""]
    lines += bars_table(result)
    for judge in result["stopped"]:
        found = result["copies"][judge]
        lines += ["", f"The run of the copies with `{judge}` was stopped after {found['judged']} of {found['copies']} copies, when "
                  "that judge was no longer available. A bar that needs all the copies of a group is not read for "
                  "it, and the counts beside such a bar are of the copies that were judged. They were judged in the order of the file, "
                  "so they are no sample of the rest."]
    lines += ["", "## Copies of the measured pages", "",
              "One sentence of a page is changed in each copy. A damaged copy lost something, and a harmless one did not.", ""]
    lines += copies_table(result)
    lines += [""] + failures_table(result)
    accepted = ", ".join(f"{result['copies'][judge]['harmless_accepted_all']} of {result['copies'][judge]['harmless']['of']} with `{judge}`"
                         for judge in result["judges"])
    lines += ["", f"With all required, the check accepts these of the harmless copies that it judged: {accepted}."]
    lines += ["", "## Revisions, checked twice", "",
              "Each judge checked the revisions of the other reviser: 27 of posts and 30 of fresh texts, in two runs of the same build.",
              ""]
    lines += revisions_table(result)
    lines += [""] + repeat_table(result)
    lines += [""] + pages_table(result)
    first = ", ".join(f"{cell(result['copies'][judge]['first_attempt'])} of the copies with `{judge}`" for judge in result["judges"])
    lines += ["", f"Verdicts on the first attempt: {first}."]
    lines += ["", "## The fitting half", "",
              "The 40 revisions of each reviser in two runs, and 75 harmless copies of the fitting posts.", ""]
    lines += fitting_tables(result)
    return "\n".join(lines) + "\n"
