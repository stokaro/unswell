"""Render RESULTS.md of the record from the recounted results. verify.py compares the page on disk with this one."""

GPT, QWEN35, QWEN27 = "gpt-6-astra", "private-host/Qwen3.6-35B-A3B", "private-host/Qwen3.8-27B"
REVISERS = (GPT, QWEN35)
JUDGES = (GPT, QWEN27, QWEN35)
HUMAN = {"blog": "human blog posts", "styled": "articles of people under a style guide"}


def percent(value):
    return "n/a" if value is None else f"{100 * value:.1f}%"


def interval(pair):
    return "" if not pair else f" ({percent(pair[0])} to {percent(pair[1])})"


def code(name):
    return name if name == "none" else f"`{name}`"


def plainly(value):
    return "n/a" if value is None else f"{value:g}"


def verdict(result, name):
    held = result["registered"][name]
    return "does not apply" if held is None else "holds" if held else "**does not hold**"


def picked(entry):
    return f"{entry['post']} of {entry['units']}, {percent(entry['rate'])}"


def published_section(result):
    published, bars = result["published"], result["bars"]
    lines = ["## Can a published post be told from human text", "",
             "A writer who follows the product publishes the revision when the check accepts it, and the post as it was otherwise.",
             "Three blind judges read the opening of such a post and of a human text, and name the one that reads more like text written",
             "by a language model. A post is a unit with its three pairs; the panel picks it when the votes for it outnumber the votes",
             "for the human texts, and a tie counts against the rate.", "",
             "| Reviser | Human text | Panel picks the post | " + " | ".join(code(judge) + " alone" for judge in JUDGES) +
             f" | Panel at most {percent(bars['panel'])} | {code(GPT)} at most {percent(bars['strict_judge'])} |",
             "| --- | --- | --- |" + " --- |" * len(JUDGES) + " --- | --- |"]
    for reviser in ("none",) + REVISERS:
        for human, name in HUMAN.items():
            entry = published[reviser][human]
            alone = " | ".join(f"{entry['judges'][judge]['post']} of {entry['judges'][judge]['units']}" for judge in JUDGES)
            bar = ("| no bar | no bar |" if reviser == "none" else
                   f"| {verdict(result, f'A1, panel, {reviser}, {human}')} | {verdict(result, f'A2, {GPT} alone, {reviser}, {human}')} |")
            who = "No revision" if reviser == "none" else code(reviser)
            lines.append(f"| {who} | {name} | {picked(entry['panel'])}{interval(entry['panel']['interval'])} | {alone} {bar}")
    accepted = ", ".join(f"{result['revisions'][reviser]['check']['accepted_set_s']} for {code(reviser)}" for reviser in REVISERS)
    lines += ["", f"Of the posts of the style-guide set, the check accepted the revision of {accepted}. The other posts are published as they were, "
              "so the rows of a reviser differ from the first rows only where a revision was accepted.", ""]
    return lines


def third(entry):
    return f"{entry['page']} page, {entry['revision']} revision, {entry['same'] + entry['unreadable']} neither"


def check_section(result):
    revisions, bars = result["revisions"], result["bars"]
    lines = ["## What the check does with a revision", "",
             "`unswell compare` on each page and its revision, with a judge of the other family than the reviser. The third model",
             f"is {code(QWEN27)}: it answers one order for each pair and names the page or the revision as the one that reads more like a",
             "model's text.", "",
             "| Reviser | Judge of the check | Pairs | Verdicts | Accepted | Sound better, no difference, worse | Second text both times |"
             " Statements kept | Revisions under 98% | Lost a number, code, or a link |",
             "| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |"]
    for reviser in REVISERS:
        check = revisions[reviser]["check"]
        sound = check["sound"]
        lines.append(f"| {code(reviser)} | {code(check['judge'])} | {check['pairs']} | {check['verdicts']}, {percent(check['verdict_share'])} | "
                     f"{check['accepted']} | {sound['better']}, {sound['no difference']}, {sound['worse']} | {check['second_text_both_times']} | "
                     f"{check['kept']:,} of {check['statements']:,}, {percent(check['kept_share'])} | {check['under_the_bar']} | {check['lost_something']} |")
    lines += ["", "\"Second text both times\" counts the pairs where the judge named text B in both orders, which is the page once and the "
              "revision once. Such a judge follows the order of the texts, and the command reads that as no difference.", "",
              "The third model on the same pairs:", "",
              "| Reviser | All revisions | Accepted by the check | Rejected by the check |", "| --- | --- | --- | --- |"]
    for reviser in REVISERS:
        mine = revisions[reviser]
        lines.append(f"| {code(reviser)} | {third(mine['third'])} | {third(mine['check']['third_on_accepted'])} | {third(mine['check']['third_on_rejected'])} |")
    together = revisions["together"]
    lines += ["", f"Bars. The check reaches a verdict on at least {percent(bars['verdicts'])} of the pairs: "
              + "; ".join(f"{code(reviser)}, {verdict(result, f'B1, verdicts, {reviser}')}" for reviser in REVISERS) + ". "
              f"On the accepted revisions of both revisers the third model finds the page more like a model's text on at least "
              f"{percent(bars['accepted_better'])} of its decisive answers, when at least {bars['fewest_accepted']} are accepted: "
              f"{together['accepted']} are accepted, {verdict(result, 'B2, accepted revisions are better')}. "
              f"On the rejected ones that share is lower: {verdict(result, 'B3, rejected revisions are less often better')}.", ""]
    return lines


def spoil_section(result):
    revisions, bars, damage, known = result["revisions"], result["bars"], result["damage"], result["known"]
    lines = ["## What a revision may not spoil", "",
             "| Reviser | Pages | Revised | No mark listed | No usable revision | Index lower | Median index before and after |"
             " `high` before and after | Revisions that add an em dash | Em dashes added | Accepted with an em dash added |",
             "| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |"]
    for reviser in REVISERS:
        mine = revisions[reviser]
        lines.append(f"| {code(reviser)} | {mine['pages']} | {mine['revised']} | {mine['no_mark_listed']} | {mine['no_revision']} | {mine['index_lower']} | "
                     f"{plainly(mine['median_index'][0])} and {plainly(mine['median_index'][1])} | {mine['high'][0]} and {mine['high'][1]} | "
                     f"{mine['add_em_dash']}, {percent(mine['add_em_dash_share'])} | {mine['em_dashes_added']} | {mine['check']['accepted_add_em_dash']} |")
    lines += ["", f"Bars. At most {percent(bars['em_dash_revisions'])} of the revisions add an em dash, and no accepted one does: "
              + "; ".join(f"{code(reviser)}, {verdict(result, f'C1, em dashes, {reviser}')}" for reviser in REVISERS) + ". "
              f"The check finds at least {percent(bars['statements_kept'])} of the statements kept over all revisions with a verdict: "
              + "; ".join(f"{code(reviser)}, {verdict(result, f'C2, statements kept, {reviser}')}" for reviser in REVISERS) + ".", "",
              "Damage on purpose. Each post of the style-guide set gets a copy with one sentence removed and a copy with the word \"not\"",
              "removed from one sentence. The sentence has no digit, code, or link, so only the search for the statements can notice it.", "",
              "| Damage | Copies | The check reports a statement as not kept |", "| --- | --- | --- |",
              f"| A sentence removed | {damage['sentence']['copies']} | {damage['sentence']['noticed']} |",
              f"| \"not\" removed | {damage['negation']['copies']} | {damage['negation']['noticed']} |",
              f"| Both | {damage['copies']} | {damage['noticed']}, {percent(damage['noticed_share'])} |", "",
              f"{damage['no_verdict']} copies got no verdict, and the check accepted {damage['accepted']} damaged copies. Bar: at least "
              f"{percent(bars['damage_noticed'])} of the copies, {verdict(result, 'C3, damage is noticed')}.", "",
              "## Can the check find a revision better", "",
              "A difference that is known to be large: a text that a model wrote under a style guide stands as the page, and the human post",
              "on the same subject as its revision. Only the blind comparison is read; the two texts do not hold the same statements.", "",
              "| Pairs | No verdict | The comparison calls the revision better | Page in both orders | Revision in both orders | Second text both times |",
              "| --- | --- | --- | --- | --- | --- |",
              f"| {known['pairs']} | {known['no_verdict']} | {known['better']}, {percent(known['better_share'])} | {known['orders']['page in both orders']} | "
              f"{known['orders']['revision in both orders']} | {known['orders']['second text both times']} |", "",
              f"Bar: at least {percent(bars['known_difference'])}, {verdict(result, 'D, a known difference is found')}.", ""]
    return lines


def fitting_section(result):
    fitting = result["fitting"]
    lines = ["## What the fitting half showed", "",
             "These runs chose what the product does. They read the 40 posts of the style-guide set in the fitting half and carry no bar.", "",
             "A post against human text, after a revision of every post by one report or instruction:", "",
             "| Revision | Reviser | Human text | Panel picks the post | " + " | ".join(code(judge) + " alone" for judge in JUDGES) + " |",
             "| --- | --- | --- | --- |" + " --- |" * len(JUDGES)]
    for entry in fitting["against_human"]:
        alone = " | ".join(f"{entry['judges'][judge]['post']} of {entry['judges'][judge]['units']}" for judge in JUDGES)
        lines.append(f"| {entry['variant']} | {code(entry['reviser'])} | {HUMAN[entry['human']]} | {picked(entry['panel'])} | {alone} |")
    lines += ["", "A post against its own revision. Each judge answers one order for each pair; a cell gives the pairs where it names the page, "
              "the revision, and neither:", "",
              "| Report and instruction | Reviser | Pairs | Panel: page, revision, tie | " + " | ".join(code(judge) for judge in JUDGES) + " |",
              "| --- | --- | --- | --- |" + " --- |" * len(JUDGES)]
    for entry in fitting["page_against_revision"]:
        cells = []
        for judge in JUDGES:
            counted = entry["judges"].get(judge)
            cells.append("not asked" if counted is None else f"{counted['page']}, {counted['revision']}, {counted['same'] + counted['unreadable']}")
        tally = entry["panel"]
        lines.append(f"| {entry['variant']} | {code(entry['reviser'])} | {entry['pairs']} | {tally['page']}, {tally['revision']}, {tally['tie']} | "
                     + " | ".join(cells) + " |")
    lines += ["", "The index and the em dashes of the same revisions:", "",
              "| Report and instruction | Reviser | Revised | Median index before and after | `high` before and after | Revisions that add an em dash |"
              " Em dashes added |", "| --- | --- | --- | --- | --- | --- | --- |"]
    for entry in fitting["revisions"]:
        lines.append(f"| {entry['variant']} | {code(entry['reviser'])} | {entry['revised']} | {plainly(entry['median_index'][0])} and "
                     f"{plainly(entry['median_index'][1])} | {entry['high'][0]} and {entry['high'][1]} | {entry['add_em_dash']} | {entry['em_dashes_added']} |")
    lines += ["", "One pair in both orders. A judge that names the second text both times follows the order:", "",
              "| Pairs | Wording | Judge | Pairs asked | Page in both orders | Revision in both orders | Second text both times | First text both times |"
              " Other |", "| --- | --- | --- | --- | --- | --- | --- | --- | --- |"]
    for entry in fitting["both_orders"]:
        lines.append(f"| {entry['pairs_of']} | {entry['wording']} | {code(entry['judge'])} | {entry['pairs']} | {entry['page in both orders']} | "
                     f"{entry['revision in both orders']} | {entry['second text both times']} | {entry['first text both times']} | {entry['other']} |")
    lines += ["", "A vote by section was tried and not kept. A page and its revision were cut at their headings, and each section was asked in "
              "both orders:", "",
              "| Pairs | Judge | Pairs asked | Sections | For the revision | Against | Unsure | Pairs better, worse, no difference |",
              "| --- | --- | --- | --- | --- | --- | --- | --- |"]
    for entry in fitting["sections"]:
        lines.append(f"| {entry['pairs_of']} | {code(entry['judge'])} | {entry['pairs']} | {entry['sections']} | {entry['for_the_revision']} | "
                     f"{entry['against']} | {entry['unsure']} | {entry['better']}, {entry['worse']}, {entry['no_difference']} |")
    damage = fitting["damage"]
    lines += ["", f"Damage on purpose, with {code(GPT)} as the judge: the check reported a statement as not kept on {damage['noticed']} of "
              f"{damage['copies']} copies, and {damage['no_verdict']} got no verdict.", ""]
    return lines


def page(result):
    applied = [value for value in result["registered"].values() if value is not None]
    lines = ["# The revision check: results", "",
             "Generated by `verify.py --write` from the data of this record. Do not edit it by hand.", "",
             f"{sum(applied)} of {len(applied)} registered bars hold, and {len(result['registered']) - len(applied)} do not apply.", ""]
    lines += published_section(result) + check_section(result) + spoil_section(result) + fitting_section(result)
    return "\n".join(lines).rstrip("\n") + "\n"
