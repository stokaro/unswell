"""Render RESULTS.md of the record from the recounted results. verify.py compares the page on disk with this one."""

GPT, QWEN35 = "gpt-6-astra", "private-host/Qwen3.6-35B-A3B"
JUDGES = (GPT, QWEN35)
COPY_PARTS = ("fitting, first copies", "fitting, second copies", "measured")
PAGES = {"set-s": "posts of the style-guide set", "fresh": "texts that no earlier run read"}
REVISION_PARTS = {"fitting, earlier build": "fitting half, before every statement that was not found was asked again", "fitting": "fitting half",
                  "read": "measured half, the revisions of the earlier record", "fresh": "texts that no earlier run read"}


def percent(value):
    return "n/a" if value is None else f"{100 * value:.1f}%"


def code(name):
    return f"`{name}`"


def verdict(result, name):
    return "holds" if result["registered"][name] else "**does not hold**"


def of(k, n):
    return f"{k} of {n}" if n else "none"


def copies_section(result):
    copies, bars = result["copies"], result["bars"]
    lines = ["## Without a model", "",
             "The command compares each sentence of a page with its revision. A damaged copy has one sentence removed, or the word \"not\"",
             "removed from one sentence; the count is the copies where that sentence is a suspect. A harmless copy has two sentences",
             "joined with \", and\", or its negations written as contractions; the count is the copies with any suspect.", "",
             "| Part | Pages | A sentence removed | \"not\" removed | Both | Two sentences joined | Contractions | Harmless copies with a suspect |",
             "| --- | --- | --- | --- | --- | --- | --- | --- |"]
    for part in COPY_PARTS:
        for pages, name in PAGES.items():
            entry = copies.get(part, {}).get(pages)
            if entry is None:
                continue
            cell = lambda kind, field: of(entry[kind][field], entry[kind]["copies"]) if kind in entry else "none made"
            lines.append(f"| {part} | {name} | {cell('sentence', 'named')} | {cell('negation', 'named')} | "
                         f"{of(entry['damaged']['named'], entry['damaged']['copies'])}, {percent(entry['damaged']['named_share'])} | "
                         f"{cell('join', 'with_a_suspect')} | {cell('contract', 'with_a_suspect')} | "
                         f"{of(entry['harmless']['with_a_suspect'], entry['harmless']['copies'])} |")
    lines += ["", f"Bar. On the measured part, at most {percent(bars['harmless_suspected'])} of the harmless copies have a suspect: "
              + "; ".join(f"{name}, {verdict(result, f'H, harmless copies, {pages}')}" for pages, name in PAGES.items()) + ".", ""]
    return lines


def damage_section(result):
    damage, bars = result["damage"], result["bars"]
    lines = ["## Damaged copies under a judge", "",
             "The whole check on each damaged copy. \"Reported by the sentences\" counts the copies where the damaged sentence is a suspect",
             "that the judge did not clear. \"By the statements\" counts those where the search for statements found one not kept, which",
             "need not be the damaged one. A copy whose check ended without a verdict was run once more; the counts after the first",
             "column of verdicts are those after that retry.", "",
             "| Part | Pages | Judge | Copies | Verdicts on the first attempt | Verdicts after one retry |"
             " Reported by the sentences: removed, \"not\" removed, both | By the statements |"
             " By either | Sound better, no difference, worse | Damaged copies accepted |",
             "| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |"]
    for part in COPY_PARTS:
        for pages, name in PAGES.items():
            for judge in JUDGES:
                entry = damage.get(part, {}).get(pages, {}).get(judge)
                if entry is None:
                    continue
                sound = entry["sound"]
                lines.append(f"| {part} | {name} | {code(judge)} | {entry['copies']} | {entry['first_verdicts']} | {entry['verdicts']} | "
                             f"{of(entry['sentence']['reported'], entry['sentence']['copies'])}, "
                             f"{of(entry['negation']['reported'], entry['negation']['copies'])}, "
                             f"{of(entry['reported'], entry['copies'])}, {percent(entry['reported_share'])} | {entry['by_statements']} | "
                             f"{entry['by_either']} | {sound['better']}, {sound['no difference']}, {sound['worse']} | {entry['accepted']} |")
    cells = [(pages, name, judge) for pages, name in PAGES.items() for judge in JUDGES]
    lines += ["", f"Bars, on the measured part. The sentences report the damage on at least {percent(bars['damage_reported'])} of the copies: "
              + "; ".join(f"{name}, {code(judge)}, {verdict(result, f'D1, damage reported, {pages}, {judge}')}" for pages, name, judge in cells)
              + f". The check accepts at most {bars['damaged_accepted']} damaged copy: "
              + "; ".join(f"{name}, {code(judge)}, {verdict(result, f'D2, damaged copies accepted, {pages}, {judge}')}" for pages, name, judge in cells)
              + f". It reaches a verdict on at least {percent(bars['verdicts'])} of the copies, after one retry of a copy without a verdict: "
              + "; ".join(f"{name}, {code(judge)}, {verdict(result, f'V, verdicts on damaged copies, {pages}, {judge}')}" for pages, name, judge in cells)
              + ". On the first attempt the shares were "
              + "; ".join(f"{name}, {code(judge)}, {percent(damage['measured'][pages][judge]['first_verdict_share'])}" for pages, name, judge in cells)
              + ".", ""]
    return lines


def revisions_section(result):
    revisions, bars = result["revisions"], result["bars"]
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
                     f"{entry['accepted']} | "
                     f"{sound['better']}, {sound['no difference']}, {sound['worse']} | {entry['kept']:,} of {entry['statements']:,}, "
                     f"{percent(entry['kept_share'])} | {not_kept['changed']}, {not_kept['missing']}, {not_kept['unquoted']} | {entry['under_the_bar']} |")
    lines += ["", "The sentences of the same revisions. A suspect is a sentence of the page without a counterpart, or with one negation less.", "",
              "| Part | Reviser | Revisions with a suspect | Suspects | Cleared by the judge | Revisions with a sentence not cleared |"
              " Sentences not cleared: missing, negation | Rejected by a sentence alone | Revisions that add an em dash |",
              "| --- | --- | --- | --- | --- | --- | --- | --- | --- |"]
    for part, name, reviser, entry in rows:
        lines.append(f"| {name} | {code(reviser)} | {of(entry['with_a_suspect'], entry['verdicts'])} | {entry['suspects']} | {entry['cleared']} | "
                     f"{entry['with_a_lost_sentence']} | {entry['lost']['missing']}, {entry['lost']['negation']} | "
                     f"{entry['rejected_by_sentences_alone']} | {entry['add_em_dash']} |")
    measured = [(part, name, reviser) for part, name, reviser, _ in rows if part in ("read", "fresh")]
    lines += ["", "\"Rejected by a sentence alone\" counts the revisions that keep at least 98% of their statements and have a sentence that was "
              "not cleared. No bar was set for it, nor for the em dashes: the check rejects a revision that adds one.", "",
              f"Bars. The check finds at least {percent(bars['statements_kept'])} of the statements kept: "
              + "; ".join(f"{name}, {code(reviser)}, {verdict(result, f'R, statements kept, {part}, {reviser}')}" for part, name, reviser in measured)
              + f". It reaches a verdict on at least {percent(bars['verdicts'])} of the pairs, after one retry of a pair without a verdict: "
              + "; ".join(f"{name}, {code(reviser)}, {verdict(result, f'V, verdicts on revisions, {part}, {reviser}')}"
                          for part, name, reviser in measured) + ". On the first attempt the shares were "
              + "; ".join(f"{name}, {code(reviser)}, {percent(revisions[part][reviser]['first_verdict_share'])}"
                          for part, name, reviser in measured) + ".", ""]
    return lines


def variants_section(result):
    lines = ["## How the rule grew on the fitting half", "",
             "Each step on the first damaged copies of the fitting half and on its ordinary revisions, without a model.", "",
             "| Rule | A removed sentence is a suspect | A sentence without its \"not\" is a suspect | "
             + " | ".join(f"Revisions of {code(reviser)} with a suspect, and suspects" for reviser in JUDGES) + " |",
             "| --- | --- | --- |" + " --- |" * len(JUDGES)]
    for entry in result["variants"]:
        damaged = entry["damaged"]
        cells = []
        for reviser in JUDGES:
            mine = entry["revisions"].get(reviser)
            cells.append("not run" if mine is None else f"{of(mine['with_a_suspect'], mine['revisions'])}, {mine['suspects']}")
        lines.append(f"| {entry['variant']} | {of(damaged['sentence']['named'], damaged['sentence']['copies'])} | "
                     f"{of(damaged['negation']['named'], damaged['negation']['copies'])} | " + " | ".join(cells) + " |")
    return lines + [""]


def page(result):
    held = result["registered"]
    lines = ["# The sentence comparison: results", "",
             "Generated by `verify.py --write` from the data of this record. Do not edit it by hand.", "",
             f"{sum(held.values())} of {len(held)} registered bars hold.", ""]
    lines += copies_section(result) + damage_section(result) + revisions_section(result) + variants_section(result)
    return "\n".join(lines).rstrip("\n") + "\n"
