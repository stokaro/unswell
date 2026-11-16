"""Render RESULTS.md of the record from the recounted results. verify.py compares the page on disk with this one."""

GPT, QWEN35 = "gpt-6-astra", "private-host/Qwen3.6-35B-A3B"
JUDGES = (GPT, QWEN35)
COPY_PARTS = ("fitting", "measured")
PAGES = {"set-s": "posts of the style-guide set", "fresh": "texts that no earlier run read"}
DAMAGE = {"sentence": "A sentence removed", "negation": "\"not\" removed", "addnot": "\"not\" put in", "swap": "A word turned around",
          "clause": "A clause cut off"}
HARMLESS = {"join": "Two sentences joined", "contract": "Contractions"}
REVISION_PARTS = {"fitting": "fitting half", "read": "measured half, the revisions of the first record", "fresh": "texts that no earlier run read"}


def percent(value):
    return "n/a" if value is None else f"{100 * value:.1f}%"


def code(name):
    return f"`{name}`"


def verdict(result, name):
    return "holds" if result["registered"][name] else "**does not hold**"


def of(k, n):
    return f"{k} of {n}" if n else "none"


def bars(result, name, cells):
    """One sentence part for a bar read on several cells: each cell with whether the bar holds there."""
    return "; ".join(f"{label}, {verdict(result, name.format(**keys))}" for label, keys in cells)


def copies_section(result):
    copies = result["copies"]
    lines = ["## Without a model", "",
             "The command compares each sentence of a page with its revision. A damaged copy has one sentence changed in one of five",
             "ways; the count is the copies where that sentence is a suspect. A harmless copy has two sentences joined with \", and\", or",
             "its negations written as contractions; the count is the copies with any suspect.", "",
             "| Part | Pages | " + " | ".join(DAMAGE.values()) + " | All damaged | " + " | ".join(HARMLESS.values()) + " | Harmless copies with a suspect |",
             "| --- | --- |" + " --- |" * (len(DAMAGE) + len(HARMLESS) + 2)]
    for part in COPY_PARTS:
        for pages, name in PAGES.items():
            entry = copies.get(part, {}).get(pages)
            if entry is None:
                continue
            cell = lambda kind, field: of(entry[kind][field], entry[kind]["copies"]) if kind in entry else "none made"
            lines.append(f"| {part} | {name} | " + " | ".join(cell(kind, "named") for kind in DAMAGE) +
                         f" | {of(entry['damaged']['named'], entry['damaged']['copies'])}, {percent(entry['damaged']['named_share'])} | "
                         + " | ".join(cell(kind, "with_a_suspect") for kind in HARMLESS) +
                         f" | {of(entry['harmless']['with_a_suspect'], entry['harmless']['copies'])} |")
    cells = [(name, {"pages": pages}) for pages, name in PAGES.items()]
    lines += ["", f"Bar H. On the measured part, at most {percent(result['bars']['harmless_suspected'])} of the harmless copies have a suspect: "
              + bars(result, "H, harmless copies, {pages}", cells) + ".", ""]
    return lines


def damage_section(result):
    damage, limits = result["damage"], result["bars"]
    lines = ["## Damaged copies under a judge", "",
             "The whole check on each damaged copy. A cell of a kind of damage counts the copies where the damaged sentence is a suspect",
             "that the judge did not clear. \"By the statements\" counts those where the search for statements found one not kept, which",
             "need not be the damaged one. A copy whose check ended without a verdict was run once more; the counts after the first",
             "column of verdicts are those after that retry.", "",
             "| Part | Pages | Judge | Copies | Verdicts on the first attempt | Verdicts after one retry | " + " | ".join(DAMAGE.values()) +
             " | All damaged | By the statements | Sound better, no difference, worse | Damaged copies accepted |",
             "| --- | --- | --- | --- | --- | --- |" + " --- |" * (len(DAMAGE) + 4)]
    for part in COPY_PARTS:
        for pages, name in PAGES.items():
            for judge in JUDGES:
                entry = damage.get(part, {}).get(pages, {}).get(judge)
                if entry is None:
                    continue
                sound = entry["sound"]
                lines.append(f"| {part} | {name} | {code(judge)} | {entry['copies']} | {entry['first_verdicts']} | {entry['verdicts']} | "
                             + " | ".join(of(entry[kind]["reported"], entry[kind]["copies"]) for kind in DAMAGE) +
                             f" | {of(entry['reported'], entry['copies'])}, {percent(entry['reported_share'])} | {entry['by_statements']} | "
                             f"{sound['better']}, {sound['no difference']}, {sound['worse']} | {entry['accepted']} |")
    lines += ["", "Each kind of damage on the measured part, the copies of both kinds of pages together:", "",
              "| Judge | " + " | ".join(DAMAGE.values()) + " |", "| --- |" + " --- |" * len(DAMAGE)]
    for judge in JUDGES:
        kinds = result["kinds"][judge]
        lines.append(f"| {code(judge)} | " + " | ".join(f"{of(kinds[kind]['reported'], kinds[kind]['copies'])}, {percent(kinds[kind]['reported_share'])}"
                                                      for kind in DAMAGE) + " |")
    cells = [(f"{name}, {code(judge)}", {"pages": pages, "judge": judge}) for pages, name in PAGES.items() for judge in JUDGES]
    kinds = [(f"{DAMAGE[kind].lower()}, {code(judge)}", {"kind": kind, "judge": judge}) for judge in JUDGES for kind in DAMAGE]
    lines += ["", f"Bar D1. The check reports the damage on at least {percent(limits['damage_reported'])} of the damaged copies of a kind of pages: "
              + bars(result, "D1, damage reported, {pages}, {judge}", cells) + ".", "",
              f"Bar D2. It reports each kind of damage on at least {percent(limits['damage_reported'])} of the copies of that kind: "
              + bars(result, "D2, {kind} reported, {judge}", kinds) + ".", "",
              f"Bar D3. It accepts at most {limits['damaged_accepted']} damaged copies of a kind of pages: "
              + bars(result, "D3, damaged copies accepted, {pages}, {judge}", cells) + ".", "",
              f"Bar V. It reaches a verdict on at least {percent(limits['verdicts'])} of the copies, after one retry of a copy without a verdict: "
              + bars(result, "V, verdicts on damaged copies, {pages}, {judge}", cells) + ". On the first attempt the shares were "
              + "; ".join(f"{name}, {code(judge)}, {percent(damage['measured'][pages][judge]['first_verdict_share'])}"
                          for pages, name in PAGES.items() for judge in JUDGES) + ".", ""]
    return lines


def revisions_section(result):
    revisions, limits = result["revisions"], result["bars"]
    lines = ["## Revisions", "",
             "The whole check on each revision by the report, with a judge of the other family than the reviser.", "",
             "| Part | Reviser | Judge | Pairs | Verdicts on the first attempt | Verdicts after one retry | Accepted |"
             " Sound better, no difference, worse | Statements kept | Changed, missing, quote not in the revision | Revisions under 98% |",
             "| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |"]
    rows = [(part, name, reviser, revisions[part][reviser]) for part, name in REVISION_PARTS.items() for reviser in JUDGES
            if reviser in revisions.get(part, {})]
    for part, name, reviser, entry in rows:
        sound, not_kept = entry["sound"], entry["not_kept"]
        lines.append(f"| {name} | {code(reviser)} | {code(entry['judge'])} | {entry['pairs']} | {entry['first_verdicts']} | {entry['verdicts']} | "
                     f"{entry['accepted']} | {sound['better']}, {sound['no difference']}, {sound['worse']} | {entry['kept']:,} of "
                     f"{entry['statements']:,}, {percent(entry['kept_share'])} | {not_kept['changed']}, {not_kept['missing']}, {not_kept['unquoted']} | "
                     f"{entry['under_the_bar']} |")
    lines += ["", "The sentences of the same revisions. A suspect is a sentence of the page that does not stand in the revision word for word.", "",
              "| Part | Reviser | Revisions with a suspect | Suspects | Cleared by the judge | Revisions with a sentence not cleared |"
              " Sentences not cleared: missing, negation, reworded | Rejected by a sentence alone | Revisions that add an em dash |",
              "| --- | --- | --- | --- | --- | --- | --- | --- | --- |"]
    for part, name, reviser, entry in rows:
        lost = entry["lost"]
        lines.append(f"| {name} | {code(reviser)} | {of(entry['with_a_suspect'], entry['verdicts'])} | {entry['suspects']} | {entry['cleared']} | "
                     f"{entry['with_a_lost_sentence']} | {lost['missing']}, {lost['negation']}, {lost['reworded']} | "
                     f"{entry['rejected_by_sentences_alone']} | {entry['add_em_dash']} |")
    measured = [(f"{name}, {code(reviser)}", {"part": part, "reviser": reviser}) for part, name, reviser, _ in rows if part in ("read", "fresh")]
    lines += ["", "\"Rejected by a sentence alone\" counts the revisions that keep at least 98% of their statements and have a sentence that was "
              "not cleared. No bar was set for it, nor for the em dashes: the check rejects a revision that adds one.", "",
              f"Bar R. The check finds at least {percent(limits['statements_kept'])} of the statements kept: "
              + bars(result, "R, statements kept, {part}, {reviser}", measured) + ".", "",
              f"Bar V. It reaches a verdict on at least {percent(limits['verdicts'])} of the pairs, after one retry of a pair without a verdict: "
              + bars(result, "V, verdicts on revisions, {part}, {reviser}", measured) + ". On the first attempt the shares were "
              + "; ".join(f"{label}, {percent(revisions[keys['part']][keys['reviser']]['first_verdict_share'])}" for label, keys in measured) + ".", ""]
    return lines


def page(result):
    held = result["registered"]
    lines = ["# Every changed sentence: results", "",
             "Generated by `verify.py --write` from the data of this record. Do not edit it by hand.", "",
             f"{sum(held.values())} of {len(held)} registered bars hold.", ""]
    lines += copies_section(result) + damage_section(result) + revisions_section(result)
    return "\n".join(lines).rstrip("\n") + "\n"
