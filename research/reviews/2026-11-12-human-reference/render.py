#!/usr/bin/env python3
"""Render RESULTS.md and RESULTS-words.md from results.json, so that no number in them is typed by hand.

The comparisons at the level of words have a page of their own. One page with
every table is more than the repository's self-check through the MCP server
accepts for a single document.

usage: render.py [record directory]
"""
import json
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
MODELS = {"claude": "`claude-sonnet-5-5`", "astra": "`gpt-6-astra`", "qwen": "`private-host/Qwen3.8-27B`"}
SETS = [("prs", "Documentation pull requests"), ("beemo", "Expert-edited answers (Beemo)"), ("lamp", "Writer-edited fiction (LAMP)")]
# A rule needs findings in this many documents of a set for a row of the rule table.
RULE_ROW_DOCUMENTS = 10
PAIRS = [("Human-edited over AI-generated", "writer-edited or raw"), ("AI-edited over AI-generated", "model-edited or raw"),
         ("Human-edited over AI-edited", "writer-edited or model-edited")]


def pct(part, whole):
    return f"{100 * part / whole:.1f}%" if whole else "n/a"


def points(value):
    return f"{100 * value:+.1f}"


def share(value):
    return f"{100 * value:.1f}%" if value is not None else "n/a"


def difference(report, key, scale=100, digits=1):
    """A difference with its interval, in points or, for a movement, in words per removed word."""
    if key not in report["differences"]:
        return "n/a"
    low, high = report["intervals"][key]
    return (f"{scale * report['differences'][key]:+.{digits}f} "
            f"({scale * low:+.{digits}f} to {scale * high:+.{digits}f})")


def ordered(models):
    """The trimming models in the order of MODELS."""
    return [(model, models[model]) for model in MODELS if model in models]


def trim_tables(add, trims, split):
    """Add the tables of one part: the trims of each model, their distances, and the other commands."""
    add("| Set | Model | Documents with a trim | Trims | Removed words | Precision | Chance in the sentences | Chance in the document "
        "| Over sentence chance, points | Over document chance, points |")
    add("| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- |")
    for key, label in SETS:
        entry = trims["sets"].get(f"{key}/{split}")
        for model, report in ordered(entry["models"] if entry else {}):
            add(f"| {label} | {MODELS[model]} | {report['documents_with_output']} of {report['documents']} | {report['trims']:,} "
                f"| {report['removed_words']:,} | {share(report['precision'])} | {share(report['chance_sentences'])} "
                f"| {share(report['chance_document'])} | {difference(report, 'precision_minus_sentences')} "
                f"| {difference(report, 'precision_minus_document')} |")
    add("")
    add("| Set | Model | Net movement per removed word | Control | Net movement minus control |")
    add("| --- | --- | --- | ---: | --- |")
    for key, label in SETS:
        entry = trims["sets"].get(f"{key}/{split}")
        for model, report in ordered(entry["models"] if entry else {}):
            if report.get("net") is not None:
                add(f"| {label} | {MODELS[model]} | {difference(report, 'net', 1, 3)} | {report['control_net']:+.3f} "
                    f"| {difference(report, 'net_minus_control', 1, 3)} |")
    add("")
    add("Words under the findings of the other commands, against the same chance in the document:" if split == "test" else
        "The other commands on these parts:")
    add("")
    add("| Set | System | Words under findings | Editor changed them | Chance in the document | Difference, points |")
    add("| --- | --- | ---: | ---: | ---: | --- |")
    for key, label in SETS:
        entry = trims["sets"].get(f"{key}/{split}")
        for system, report in sorted(entry["comparisons"].items() if entry else (), key=lambda item: system_order(item[0])):
            add(f"| {label} | {system_name(system)} | {report['removed_words']:,} | {share(report['precision'])} "
                f"| {share(report['chance_document'])} | {difference(report, 'precision_minus_document')} |")


def trim_section(add, trims):
    add("")
    add("## Trims against the words an editor changed")
    add("")
    add("`unswell trim` ran on every development and test document with two models. A removed word is a hit when the")
    add("editor of the reference removed or replaced it. Chance in the sentences is what random removal of the same number")
    add("of words from the same sentences gives, among the words the check allows to remove. Chance in the document is the")
    add("same for random words of the same documents. Net movement is the change of the word-level edit distance to the")
    add("editor's text per removed word: below zero, the trimmed text is closer to the editor's text than the original was.")
    add("The control removes random words of the same sentences. Brackets hold 95% bootstrap intervals over documents,")
    add(f"{trims['bootstrap_draws']:,} draws.")
    add("")
    add("### Test parts")
    add("")
    trim_tables(add, trims, "test")
    add("")
    add("The registered rule, set by set:")
    add("")
    for key, label in SETS:
        entry = trims["sets"].get(f"{key}/test")
        if not entry:
            continue
        rule = entry["rule"]
        line = f"- {label}: trimming " + ("agrees with the editors above chance" if rule["agrees_above_chance"] else
                                         "does not agree with the editors above chance by the rule")
        if "moves_toward_editor" in rule:
            line += (", and it moves the text toward the editor's version" if rule["moves_toward_editor"] else
                     ", and it does not move the text toward the editor's version")
        add(line + ".")
    add("")
    add("### Development parts")
    add("")
    add("These were read before the test parts were run. They decide nothing.")
    add("")
    trim_tables(add, trims, "dev")


def opener_section(add, openers):
    add("")
    add("## A chat opener, listed from what experts removed")
    add("")
    add("`scaffold.chat-preamble` gained the openers `Sure,` and `Sure!` at the start of a document. The table counts")
    add("the findings these two phrases produce and how often the editor removed or replaced every flagged word.")
    add("")
    add("| Set | Part | Documents | Findings | Editor changed every flagged word | 95% interval |")
    add("| --- | --- | ---: | ---: | ---: | --- |")
    for key, label in SETS:
        for split in ("dev", "test"):
            entry = openers["sets"].get(f"{key}/{split}")
            if not entry:
                continue
            low, high = entry["interval"]
            interval = f"{100 * low:.1f}% to {100 * high:.1f}%" if entry["findings"] else "n/a"
            add(f"| {label} | {'development' if split == 'dev' else 'test'} | {entry['documents']} | {entry['findings']} "
                f"| {entry['changed']} of {entry['findings']} | {interval} |")
    add("")
    add("The rule asked for at least 85% on the test part of the expert-edited answers, a lower end of the interval above")
    add(f"70%, and at most one finding on the other two sets, which have {openers['elsewhere']}. "
        + ("The phrases are listed." if openers["listed"] else "The phrases are not listed."))
    load = openers.get("load")
    if load:
        add("")
        add("The rule also looks at the first sentence after the leading headings of a document. With that position it ran")
        add("on texts outside this record: documentation pages of the assistant-panel corpus, and source files of")
        add("open-source packages, where the first comment or string of a file counts as the start of a document.")
        add("")
        add("| Texts | Number | Failed to run | Findings of the rule |")
        add("| --- | ---: | ---: | ---: |")
        names = {"real": "Pages of public repositories", "generated": "Pages that models generated",
                 "code": "Source files of open-source packages"}
        for part, entry in load.items():
            add(f"| {names.get(part, part)} | {entry['pages']:,} | {entry['failed']} | {entry['findings']} |")
        filtered = load.get("code")
        if filtered and "candidates" in filtered:
            add("")
            add(f"A finding of the listed openers needs their text in the file. {filtered['candidates']:,} of the source files hold it,")
            add("and the rule ran on those. The ones that failed exceed the engine's file size limit or do not parse.")


def rule_word_section(add, rule_words):
    add("")
    add("## Each rule against the words an editor changed")
    add("")
    add("Test parts, every rule enabled, before the opener was listed. A row gives the words under the findings of one")
    add("rule, how many of them the editor changed, and what random words of the same documents would give. No rule was")
    add(f"registered for this table. It lists the rules with findings in at least {RULE_ROW_DOCUMENTS} documents of a set;")
    add("`results.json` has every rule.")
    for key, label in SETS:
        rules = rule_words.get(f"{key}/test")
        if not rules:
            continue
        add("")
        add(f"{label}:")
        add("")
        add("| Rule | Documents | Words under findings | Editor changed them | Chance in the document | Difference, points |")
        add("| --- | ---: | ---: | ---: | ---: | --- |")
        for name, report in rules.items():
            if report["documents"] < RULE_ROW_DOCUMENTS:
                continue
            add(f"| `{name}` | {report['documents']} | {report['removed_words']:,} | {share(report['precision'])} "
                f"| {share(report['chance_document'])} | {difference(report, 'precision_minus_document')} |")


def vocabulary_section(add, vocabulary):
    listed, rule = vocabulary["list"], vocabulary["rule"]
    add("")
    add("## A wide list of phrases that chat models favor, on answers the other results did not use")
    add("")
    add(f"The list has {listed['phrases']} words and phrases and is not in the product. It ran on expert-edited answers that no")
    add("other result of this record had used. On the selection part a phrase left the list when it had at least four")
    add(f"findings and the expert removed or replaced every flagged word in fewer than half of them: {len(listed['dropped'])} left,")
    add(f"{listed['kept']} stayed. The confirmation part ran once with those. The second row is high by construction: the list")
    add("was pruned on the answers it counts.")
    add("")
    add("| Part | Answers | With a finding | Findings | Words under findings that the editor changed | Chance in the document "
        "| Difference, points | Findings with every flagged word changed | 95% interval |")
    add("| --- | ---: | ---: | ---: | ---: | ---: | --- | ---: | --- |")
    for part, label in (("selection_all", f"Selection, all {listed['phrases']} phrases"),
                        ("selection", f"Selection, the {listed['kept']} that stayed"), ("confirmation", "Confirmation")):
        report = vocabulary["parts"][part]
        low, high = report["findings_interval"]
        add(f"| {label} | {report['documents']} | {report['documents_with_output']} | {report['findings']} "
            f"| {share(report['precision'])} | {share(report['chance_document'])} | {difference(report, 'precision_minus_document')} "
            f"| {report['findings_changed']} of {report['findings']}, {pct(report['findings_changed'], report['findings'])} "
            f"| {100 * low:.1f}% to {100 * high:.1f}% |")
    add("")
    add("The phrases that left: " + ", ".join(f"`{phrase}`" for phrase in listed["dropped"]) + ".")
    add("")
    add("The rule for the confirmation part asked for two things. Words under findings above the chance of the same answers, "
        "with an interval above zero: " + ("met. " if rule["words_above_chance"] else "not met. ")
        + "At least 70% of findings with every flagged word changed, with the lower end of the interval above 60%: "
        + ("met. " if rule["findings_changed_enough"] else "not met. ")
        + ("The list ships as a rule that is off by default." if rule["ships"] else "The list does not ship."))
    load = vocabulary.get("load")
    if load:
        add("")
        add(f"On {load['pages']:,} pages of public repositories from the assistant-panel corpus the kept phrases produce "
            f"{load['findings']:,} findings in {load['words']:,} words, {1000 * load['findings'] / load['words']:.2f} per 1,000 words, "
            f"on {load['pages_with_finding']:,} of the pages.")


def deslop_section(add, deslop):
    add("")
    add("## Findings before and after commits that remove AI-sounding wording")
    add("")
    add("A pair is one Markdown file before and after a commit whose message says that it removes AI-sounding wording")
    add("and that carries no assistant's trailer. For a set of rules, the findings per 1,000 words are compared between")
    add("the two versions. The last column is the share of the pairs that differ in which the findings are lower after")
    add("the commit; rules unrelated to the edit give 50%. `format.em-dash-frequency` reports a document once, so its")
    add("own row counts the em dashes that it lists; in a set of rules it counts as one finding.")
    labels = [("alpha9_defaults", "Default rules of v0.1.0-alpha.9"),
              ("format.decorative-emoji-with-status-marks", "`format.decorative-emoji` as registered, with check marks and crosses"),
              ("format.decorative-emoji", "`format.decorative-emoji` as it ships, without them"),
              ("format.em-dash-density", "`format.em-dash-density`"), ("format.em-dash-frequency", "`format.em-dash-frequency`"),
              ("with_decorative_emoji", "Default rules with `format.decorative-emoji`"),
              ("current_defaults", "Default rules with `format.decorative-emoji` and `format.em-dash-frequency`"),
              ("shipping_defaults", "The rules that ship on now, without the six that report on pages people wrote")]
    for part, title in (("test", "Test part"), ("development", "Development part, read first")):
        add("")
        add(f"{title}: {deslop['parts'][part]['alpha9_defaults']['pairs']} pairs from {deslop['repositories'][part]} repositories.")
        add("")
        add("| Rules | Pairs where they fire | Lower after | Higher after | Findings before | Findings after | Lower among pairs that differ |")
        add("| --- | ---: | ---: | ---: | ---: | ---: | --- |")
        for name, label in labels:
            entry = deslop["parts"][part][name]
            low, high = entry["interval"]
            add(f"| {label} | {entry['fires']} | {entry['lower']} | {entry['higher']} | {entry['findings_before']:,} "
                f"| {entry['findings_after']:,} | {pct(entry['lower'], entry['lower'] + entry['higher'])} ({100 * low:.1f}% to {100 * high:.1f}%) |")
    rule = deslop["rule"]
    add("")
    add("A rule stays on by default at 65% or more on the test part with the lower end of the interval above 55%: "
        + "; ".join(f"`{name}` " + ("does" if rule[name] else "does not")
                    for name in ("format.decorative-emoji-with-status-marks", "format.em-dash-density", "format.em-dash-frequency"))
        + ". The default rules as a whole were to reach 60% with the same lower end: they "
        + ("do." if rule["defaults_follow_removal"] else "do not."))
    add("")
    add("The emoji rule was registered and tested with check marks and crosses counted. They were then taken out, because")
    add("people use them in checklists and the development part shows them removed about as often as any word. The row of")
    add("the rule as it ships was measured after that, on the same pairs, so it is a second reading of the test part.")
    load = deslop.get("load")
    if load:
        add("")
        add("Findings of the rule as it ships, on documentation:")
        add("")
        add("| Texts | Pages | Pages with a finding | Findings per 1,000 words |")
        add("| --- | ---: | ---: | ---: |")
        names = [("people", "Pages that people wrote before November 2022"), ("repositories", "Pages of public repositories"),
                 ("generated", "Pages that models generated for the assistant-panel corpus"),
                 ("before_deslop", "Texts before a commit that removes AI-sounding wording")]
        for key, label in names:
            if key in load:
                entry = load[key]
                add(f"| {label} | {entry['pages']:,} | {entry['pages_with_finding']:,} ({pct(entry['pages_with_finding'], entry['pages'])}) "
                    f"| {1000 * entry['findings'] / entry['words']:.2f} |")
    frequency = deslop.get("frequency_load")
    if frequency:
        add("")
        add("`format.em-dash-frequency` was registered after the emoji rule had been read on the test part, so it is a second")
        add("use of those pairs. Its bar has one more condition: it may flag at most 5% of the pages that people wrote before")
        add("November 2022. Pages that it flags, and the em dashes that it lists:")
        add("")
        add("| Texts | Pages | Pages flagged | Em dashes listed per 1,000 words |")
        add("| --- | ---: | ---: | ---: |")
        for key, label in names:
            if key in frequency:
                entry = frequency[key]
                add(f"| {label} | {entry['pages']:,} | {entry['pages_with_finding']:,} ({pct(entry['pages_with_finding'], entry['pages'])}) "
                    f"| {1000 * entry['findings'] / entry['words']:.2f} |")
    add("")
    add("Who typed a commit is unknown: one without an assistant's trailer can still have been made by an assistant that")
    add("was told to de-slop. Commits made in GitHub's web editor were most likely typed by a person, and they are few.")
    add("Among them, the pairs in which a rule's count is lower after the commit, of those that differ:")
    add("")
    add("| Rule | Test part | Development part |")
    add("| --- | ---: | ---: |")
    for name in ("format.decorative-emoji", "format.em-dash-frequency", "current_defaults"):
        cells = []
        for part in ("test", "development"):
            entry = deslop["web_edits"][part][name]
            cells.append(f"{entry['lower']} of {entry['lower'] + entry['higher']} ({entry['pairs']} pairs)")
        label = "All default rules" if name == "current_defaults" else f"`{name}`"
        add(f"| {label} | {cells[0]} | {cells[1]} |")
    add("")
    add("These counts are too small to confirm or to reject either rule on edits that a person certainly typed.")
    english = deslop["english"]["test"]
    add("")
    add(f"The pairs were not selected by language, and {deslop['parts']['test']['alpha9_defaults']['pairs'] - english['alpha9_defaults']['pairs']} "
        f"of the {deslop['parts']['test']['alpha9_defaults']['pairs']} test pairs are not English prose. A file counts as English when at least")
    add("10% of the words of its prose are common English function words. The cut was made after the results above. On the")
    add(f"{english['alpha9_defaults']['pairs']} English test pairs:")
    add("")
    add("| Rules | Pairs where they fire | Lower after | Higher after | Lower among pairs that differ |")
    add("| --- | ---: | ---: | ---: | --- |")
    for name, label in (("alpha9_defaults", "Default rules of v0.1.0-alpha.9"), ("format.decorative-emoji", "`format.decorative-emoji`"),
                        ("format.em-dash-frequency", "`format.em-dash-frequency`"),
                        ("current_defaults", "Default rules with both new rules")):
        entry = english[name]
        low, high = entry["interval"]
        add(f"| {label} | {entry['fires']} | {entry['lower']} | {entry['higher']} "
            f"| {pct(entry['lower'], entry['lower'] + entry['higher'])} ({100 * low:.1f}% to {100 * high:.1f}%) |")
    kept = deslop["rule"]["english"]
    add("")
    add("On English files alone the bar for a default rule is "
        + ("met by both new rules." if all(kept.values()) else "not met by both new rules."))


def review_repeat_section(add, repeat):
    add("")
    add("## Whether a reviewer repeats itself")
    add("")
    add("`unswell review` used to ask the reviewer only for the items that need revision. It now asks for a verdict on every")
    add("item. A stricter verdict form, with \"revise\" only when the defect is clear, was tried as well. Each form was run")
    add("twice on the same texts: 40 files before a commit that removes AI-sounding wording, pairs 41 to 80 of the review")
    add("sample, which no earlier run of these forms with the first reviewer had touched. The repeat rate is the share of the")
    add("sentences flagged in one run that the other run flags too. The other columns are from the first run: the flagged")
    add("words that the commit changed, their share of the flagged words, the same share for all words of those documents,")
    add("and the difference with a 95% interval.")
    labels = (("verdicts", "A verdict for every item"), ("strict", "The same, \"revise\" only when clear"),
              ("list", "Only the items that need revision"))
    for reviewer, entry in repeat.items():
        add("")
        add(f"Reviewer `{reviewer}`, no confirming model:")
        add("")
        add("| Answer asked for | Sentences flagged, two runs | In both | Repeat rate | Changed words flagged | Share of flagged words "
            "| Chance | Difference |")
        add("| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |")
        for name, label in labels:
            row = entry[name]
            low, high = row["interval"]
            add(f"| {label} | {row['sentences'][0]} and {row['sentences'][1]} | {row['both']} | {100 * row['repeat_rate']:.1f}% "
                f"| {row['changed_words_flagged']} | {100 * row['precision']:.1f}% | {100 * row['chance']:.1f}% "
                f"| {100 * (row['precision'] - row['chance']):+.1f} points ({100 * low:+.1f} to {100 * high:+.1f}) |")
        add("")
        for name, label in labels[:2]:
            low, high = entry[name]["difference_interval"]
            add(f"{label}: repeats {100 * entry[name]['difference']:+.1f} points against the list form "
                f"({100 * low:+.1f} to {100 * high:+.1f}).")
            add("")
        rule = entry["rule"]
        add("The verdict form was to replace the list form if it repeats more with an interval that excludes zero, its flags fall")
        add("on changed words above chance with an interval that excludes zero, and it flags no fewer changed words: "
            + ("all three hold." if rule["ships"] else "they do not all hold: " + ", ".join(
                f"{name.replace('_', ' ')} {'holds' if kept else 'fails'}" for name, kept in rule.items() if name != "ships") + "."))
        row = entry["verdicts"]
        move, edit = row["movement"], row["movement"]["edit"]
        low, high = move["interval"]
        add("")
        add(f"Its findings number {row['findings'][0]} on the texts before the commit and {move['findings_after']} on the texts after. "
            f"Per 1,000 words they are lower after the commit in {move['lower']} pairs and higher in {move['higher']}: "
            f"{pct(move['lower'], move['lower'] + move['higher'])} ({100 * low:.1f}% to {100 * high:.1f}%). "
            "The review is not shown to report fewer findings after such a commit.")
        add("")
        add(f"What the commit did to the flagged sentences says why. Of the {edit['flagged_before']} sentences flagged before the commit, "
            f"it changed or removed {edit['changed_by_commit']} ({pct(edit['changed_by_commit'], edit['flagged_before'])}); of all "
            f"{edit['sentences']:,} sentences of six or more words in those files it changed or removed "
            f"{pct(edit['sentences_changed'], edit['sentences'])}. Of the {edit['flagged_after']} sentences flagged after the commit, "
            f"{edit['on_new_text']} are new or changed text, {edit['flagged_both_times']} stood unchanged and were flagged before, and "
            f"{edit['flagged_after'] - edit['on_new_text'] - edit['flagged_both_times']} stood unchanged and were not. "
            "So the count stays up mainly because the rewritten text is flagged again, not because the reviewer fails to repeat itself.")
        add("")
        add(f"The review does not find most of the edit: of the {edit['sentences_changed']} sentences that the commits changed or removed, "
            f"it had flagged {edit['changed_by_commit']} ({pct(edit['changed_by_commit'], edit['sentences_changed'])}).")


def defaults_section(add, load, manual):
    add("")
    add("## Default rules on pages that people wrote, and on likely-manual edits")
    add("")
    add(f"{load['pages']} pages that people wrote before November 2022. A default rule may report on at most one page in twenty.")
    add("")
    add("| Rule | Pages with a finding | Share | Ships |")
    add("| --- | ---: | ---: | --- |")
    strict = ("filler.wordy-phrase", "syntax.long-sentence", "filler.instruction-scaffolding", "hype.vague-praise",
              "repetition.adjacent-word", "format.decorative-emoji")
    for name, count in sorted(load["rules"].items(), key=lambda item: -item[1]):
        add(f"| `{name}` | {count} | {pct(count, load['pages'])} | {'off; on in strict' if name in strict else 'on'} |")
    add("")
    add(f"With the rules that ship on, {load['with_default_finding']} pages have a finding "
        f"({pct(load['with_default_finding'], load['pages'])}); with the six others on as well, {load['with_any_finding']} "
        f"({pct(load['with_any_finding'], load['pages'])}). Every rule over the limit ships off: "
        + ("yes." if load["defaults_within_limit"] else "no."))
    if not manual:
        return
    add("")
    add(f"Likely-manual edits: {manual['pairs']} pairs from {manual['repositories']} repositories. A pair is one Markdown file before and")
    add("after a commit made through GitHub's web flow that changes that one file and whose message is one short line saying")
    add("that it removes AI-sounding wording. A commit made through GitHub's API looks the same, and a person can paste text")
    add("into the editor, so these are likely typed by a person, not proven. Findings per 1,000 words are compared as above.")
    labels = (("format.em-dash-frequency", "`format.em-dash-frequency`"), ("format.decorative-emoji", "`format.decorative-emoji`"),
              ("strict_only_rules", "The five other rules that now ship off"), ("default_rules", "The rules that ship on"),
              ("earlier_defaults", "All rules that shipped on before"))
    for key, title in (("all", "All pairs"), ("one_per_repository", "One pair per repository")):
        add("")
        add(f"{title}:")
        add("")
        add("| Rules | Pairs where they report | Lower after | Higher after | Lower among pairs that differ |")
        add("| --- | ---: | ---: | ---: | --- |")
        for name, label in labels:
            entry = manual[key][name]
            low, high = entry["interval"]
            add(f"| {label} | {entry['fires']} | {entry['lower']} | {entry['higher']} "
                f"| {pct(entry['lower'], entry['lower'] + entry['higher'])} ({100 * low:.1f}% to {100 * high:.1f}%) |")
    add("")
    add("A rule is confirmed on these edits at 65% or more with the lower end of the interval above 50%: "
        + "; ".join(f"`{name}` " + ("is" if kept else "is not") for name, kept in manual["confirmed"].items()) + ".")


def page_index_section(add, index, load):
    human, held = index["human"], index["held_out"]
    add("")
    add("## Page index of AI-sounding documentation")
    add("")
    add("For a Markdown page of 150 prose words or more the product counts six marks per 1,000 prose words and combines the")
    add("natural logarithm of one plus each rate with a fixed weight; the index is 100 times the logistic function of the sum,")
    add("as a whole number. Level \"high\" is 70 or more, \"elevated\" 40 or more. The check recomputes the index of every page")
    add("from its counts with the weights below and fails when it differs from what the product reported.")
    add("")
    add("| Signal | Weight |")
    add("| --- | ---: |")
    for name, weight in index["weights"].items():
        add(f"| {name} | {weight:.2f} |")
    add(f"| constant | {index['bias']:.2f} |")
    add("")
    add(f"Pages that people wrote: {human['pages']} README files as they were before November 2022, from repositories that had no")
    add("part in fitting the weights. Pairs: one Markdown file before and after a commit whose message says that it removes")
    add("AI-sounding wording, English, 150 words or more in both versions. The held-out pairs come from repositories with no")
    add("pair in the main set: unused commits of the search and short one-file edits through GitHub's web flow. Nothing about")
    add("them was used to choose a rule or to fit the index. The test part of the main set stands apart, because its pairs were")
    add("used to admit two rules.")
    add("")
    add("| Pairs | Count | AUC against the pages that people wrote | Index lower after | Higher after | Same "
        "| Lower among pairs that differ | High before | High after |")
    add("| --- | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |")
    labels = (("held_out", "Held-out pairs"), ("held_out_web", "Held-out web-flow edits alone"),
              ("web", "All web-flow edits, with repositories of the main set"), ("main_test", "Test part of the main set"))
    for key, label in labels:
        entry = index[key]
        low, high = entry["interval"]
        changed = entry["lower"] + entry["higher"]
        add(f"| {label} | {entry['pairs']} | {entry['auc']:.3f} | {entry['lower']} | {entry['higher']} | {entry['pairs'] - changed} "
            f"| {pct(entry['lower'], changed)} ({100 * low:.1f}% to {100 * high:.1f}%) "
            f"| {pct(entry['high_before'], entry['pairs'])} | {pct(entry['high_after'], entry['pairs'])} |")
    add("")
    add(f"Of the pages that people wrote, {human['high']} are at level \"high\" ({pct(human['high'], human['pages'])}) and "
        f"{human['elevated']} at \"elevated\" ({pct(human['elevated'], human['pages'])}).")
    if load:
        add("")
        add(f"The rules that ship on, over the same {load['pages']} README files: {load['with_any_finding']} have a finding "
            f"({pct(load['with_any_finding'], load['pages'])}). A default rule may report on at most one page in twenty.")
        add("")
        add("| Rule | Pages with a finding | Share |")
        add("| --- | ---: | ---: |")
        for name, count in sorted(load["rules"].items(), key=lambda item: (-item[1], item[0])):
            add(f"| `{name}` | {count} | {pct(count, load['pages'])} |")
        add("")
        add("Every rule that reports is within the limit: " + ("yes." if not load["over_limit"] else "no, over it: "
            + ", ".join(f"`{name}`" for name in load["over_limit"]) + "."))
    add("")
    registered = index["registered"]
    words = {True: "holds", False: "does not hold"}
    add("Registered on issue 525 before the index ran outside the development part: "
        f"AUC of at least 0.80 on the held-out pairs {words[registered['auc']]}; "
        f"index lower after the commit in at least 70% of the held-out pairs that differ {words[registered['pairs']]}; "
        f"the same on the held-out web-flow edits alone {words[registered['web_pairs']]}; "
        f"level \"high\" on at most 5% of the pages that people wrote {words[registered['human_high']]}.")
    add("")
    add(f"The index does not change in {held['pairs'] - held['lower'] - held['higher']} of the {held['pairs']} held-out pairs, so it is "
        f"lower after the commit in {pct(held['lower'], held['pairs'])} of all of them. The web-flow reading rests on "
        f"{index['held_out_web']['lower'] + index['held_out_web']['higher']} pairs that differ. The index is fitted on documentation "
        "that someone set out to clean; it does not say who wrote a page.")


FIX_ARM_NAMES = {"fix-comma": "`unswell fix` at the registered run, no model", "fix": "`unswell fix` as it ships, no model",
                 "report": "A model that follows unswell's report",
                 "google": "The same model with Vale's Google findings", "microsoft": "The same model with Vale's Microsoft findings",
                 "write-good": "The same model with write-good's findings"}
FIX_VERSION_NAMES = {"original": "the original", "fix-comma": "after `unswell fix` at the registered run",
                     "fix": "after `unswell fix` as it ships", "report": "fixed by unswell's report",
                     "google": "fixed by Vale's Google findings", "microsoft": "fixed by Vale's Microsoft findings",
                     "write-good": "fixed by write-good's findings"}
FIX_LINTER_NAMES = {"google": "Vale with the Google style", "microsoft": "Vale with the Microsoft style", "write-good": "write-good"}


def judge_share(entry):
    decisive = entry["before"] + entry["after"]
    low, high = entry["interval"]
    return f"{pct(entry['before'], decisive)} ({100 * low:.1f}% to {100 * high:.1f}%)"


def judges_section(add, judges):
    add("")
    add("## The judges of the fix experiment")
    add("")
    add(f"{judges['pairs']} held-out pairs of one page before and after a de-slop edit by people (unused commits of the search and web-flow")
    add("edits from repositories with no pair in the main set, one per repository, 150 to 1,500 words, at least 3% of the tokens")
    add("changed). A judge sees the two versions as A and B in an order fixed by a hash and names the one that reads more like")
    add("text written by a language model. A judge is admitted if it names the version before the edit in at least 70% of its")
    add("decisive answers.")
    add("")
    add("| Judge | Before the edit named | After the edit named | Same | Share of decisive answers | Admitted |")
    add("| --- | ---: | ---: | ---: | --- | --- |")
    for judge, entry in judges["comparisons"]["edit"].items():
        if entry["before"] + entry["after"]:
            add(f"| `{judge}` | {entry['before']} | {entry['after']} | {entry['same']} | {judge_share(entry)} "
                f"| {'yes' if judges['admitted'][judge] else 'no'} |")
    add("")
    add(f"{judges['likely_manual_pairs']} of the pairs are short one-file edits through GitHub's web flow, which a person most likely typed. On them alone:")
    add("")
    add("| Judge | Before the edit named | After the edit named | Same | Share of decisive answers | At least 70% |")
    add("| --- | ---: | ---: | ---: | --- | --- |")
    for judge, entry in judges["likely_manual"].items():
        if entry["before"] + entry["after"]:
            add(f"| `{judge}` | {entry['before']} | {entry['after']} | {entry['same']} | {judge_share(entry)} "
                f"| {'yes' if judges['admitted_on_likely_manual'][judge] else 'no'} |")
    both, fixed = judges["comparisons"]["edit_both_fixed"], judges["comparisons"]["fixed_against_edit"]
    if not any(entry["before"] + entry["after"] for entry in both.values()):
        return
    add("")
    add("The judges often give as their reason the marks that the index counts. So the same pairs were judged again with")
    add("`unswell fix` applied to both versions, which leaves neither side a mark with a mechanical replacement. A pair that")
    add("the engine cannot parse is left out.")
    add("")
    add("| Judge | Before the edit named | After the edit named | Same | Share of decisive answers |")
    add("| --- | ---: | ---: | ---: | --- |")
    for judge, entry in both.items():
        if entry["before"] + entry["after"]:
            add(f"| `{judge}` | {entry['before']} | {entry['after']} | {entry['same']} | {judge_share(entry)} |")
    if not any(entry["before"] + entry["after"] for entry in fixed.values()):
        return
    add("")
    add("The version before the edit after `unswell fix`, against the version as the person left it:")
    add("")
    add("| Judge | The fixed earlier version named | The person's version named | Same | Share naming the fixed earlier version |")
    add("| --- | ---: | ---: | ---: | --- |")
    for judge, entry in fixed.items():
        if entry["before"] + entry["after"]:
            add(f"| `{judge}` | {entry['before']} | {entry['after']} | {entry['same']} | {judge_share(entry)} |")


def fix_extra_section(add, extra, judge):
    add("")
    add(f"A judge of another model family, `{judge}`, was registered after the run of the panel and judged four of the")
    add("comparisons alone:")
    add("")
    add("| Version under test | Against | Named the other version | Named the version under test | Same | No answer "
        "| Share for the version under test | With the two judges of the panel: for it, against it, tie |")
    add("| --- | --- | ---: | ---: | ---: | ---: | --- | --- |")
    for key, entry in extra.items():
        first, _, second = key.partition("|")
        tested, other = (second, first) if first == "original" else (first, second)
        low, high = entry["interval"]
        decisive = entry["tested_less"] + entry["other_less"]
        three = entry["three_judges"]
        ours, theirs = (three["second_less"], three["first_less"]) if first == "original" else (three["first_less"], three["second_less"])
        add(f"| {FIX_VERSION_NAMES[tested]} | {FIX_VERSION_NAMES[other]} | {entry['tested_less']} | {entry['other_less']} "
            f"| {entry['answers']['same']} | {entry['answers']['missing']} "
            f"| {pct(entry['tested_less'], decisive)} ({100 * low:.1f}% to {100 * high:.1f}%) | {ours}, {theirs}, {three['tie']} |")
    add("")
    words = {True: "holds", False: "does not hold"}
    add("Its bars, registered before its run: more than half of its decisive answers for the version of unswell, with the lower")
    add("end of the interval above 50%. "
        f"`unswell fix` at the registered run against the original: {words[extra['original|fix-comma']['holds']]}. "
        f"The report version against the original: {words[extra['original|report']['holds']]}. "
        f"The report version against write-good's: {words[extra['report|write-good']['holds']]}. "
        f"On the second reading, `unswell fix` as it ships against the original: {words[extra['original|fix']['holds']]}.")


FRESH_VERSION_NAMES = {"original": "the original", "fix": "after `unswell fix`", "report": "fixed by unswell's report",
                       "write-good": "fixed with write-good's findings"}


def fix_fresh_section(add, fresh):
    arms = fresh["arms"]
    add("")
    add("## The released binary on new pages")
    add("")
    add("`unswell fix` changed after the registered run above, and its second reading used pages that had been read once.")
    add(f"This run measures the published `v0.1.0-alpha.10` binary on {fresh['pages']} pages that no earlier run had read. They come from a")
    add("new search of commits whose message states the removal of AI-sounding wording, without the commits that an earlier")
    add("search returned and without any repository of an earlier set, and they are selected as before: English, 150 to 1,500")
    add("words, index 40 or more, one page per repository. The versions are the page after `unswell fix`, the page fixed by a")
    add("model with the report of `unswell check`, and the page fixed by the same model with the findings of write-good, the")
    add("best third-party linter of the registered run. The bars are those of the registered run and were written on issue 525")
    add("before a fixed version of these pages existed.")
    add("")
    add("| Fixed by | Pages fixed | Failed | Index lower | Index higher | High before | High after | Median share of tokens changed |")
    add("| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |")
    for name, label in (("fix", "`unswell fix`, no model"), ("report", "A model that follows unswell's report"),
                        ("write-good", "The same model with write-good's findings")):
        arm = arms[name]
        changed = f"{100 * arm['median_changed']:.1f}%" if arm["median_changed"] is not None else "n/a"
        add(f"| {label} | {arm['fixed']} | {arm['failed']} | {arm['lower']} | {arm['higher']} | {arm['high_before']} "
            f"| {arm['high_after']} | {changed} |")
    extra = fresh.get("extra_judge", {})
    add("")
    add("| Version under test | Against | Panel: the less AI-sounding one on, of the decisive pages | Panel: pages without a verdict "
        "| Third judge alone, of its decisive answers | Third judge: pages it calls the same |")
    add("| --- | --- | --- | ---: | --- | ---: |")
    for key, entry in fresh["comparisons"].items():
        first, _, second = key.partition("|")
        tested, other = (second, first) if first == "original" else (first, second)
        low, high = entry["interval"]
        decisive = entry["tested_less"] + entry["other_less"]
        alone, same = "not read", ""
        if key in extra:
            alone = f"{extra[key]['tested_less']} of {extra[key]['tested_less'] + extra[key]['other_less']}"
            same = str(extra[key]["answers"]["same"])
        add(f"| {FRESH_VERSION_NAMES[tested]} | {FRESH_VERSION_NAMES[other]} | {entry['tested_less']} of {decisive}, "
            f"{pct(entry['tested_less'], decisive)} ({100 * low:.1f}% to {100 * high:.1f}%) | {entry['tie']} | {alone} | {same} |")
    registered = fresh["registered"]
    words = {True: "holds", False: "does not hold"}
    add("")
    add("The bars:")
    add("")
    add(f"- The index is lower after `unswell fix` on at least 90% of the pages: {arms['fix']['lower']} of {fresh['pages']}, "
        f"{words[registered['fix_lowers_index']]}. After the fix by the report: {arms['report']['lower']} of {fresh['pages']}, "
        f"{words[registered['report_lowers_index']]}.")
    add("- The version is the less AI-sounding one on more than half of the decisive pages, with the lower end of the interval "
        f"above 50%. `unswell fix` against the original: {words[registered['original|fix']]}. The report version against the "
        f"original: {words[registered['original|report']]}. The report version against write-good's version: "
        f"{words[registered['report|write-good']]}.")
    if extra:
        add("- The third judge alone, with the same bar on its decisive answers: `unswell fix` against the original "
            f"{words[extra['original|fix']['holds']]}, the report version against the original {words[extra['original|report']['holds']]}, "
            f"the report version against write-good's version {words[extra['report|write-good']['holds']]}.")


def verdict_cell(entry):
    decisive = entry["tested_less"] + entry["other_less"]
    low, high = entry["interval"]
    return f"{entry['tested_less']} of {decisive}, {pct(entry['tested_less'], decisive)} ({100 * low:.1f}% to {100 * high:.1f}%)"


def fix_third_section(add, third):
    arms, extra = third["arms"], third.get("extra_judge", {})
    add("")
    add("## `unswell fix` against the best third-party linter, on a third sample")
    add("")
    add("In both runs above the page after `unswell fix` was the less AI-sounding one against the page fixed with write-good's")
    add("findings, and in neither did that comparison have a bar. The comparison that had one, for the page fixed by a model")
    add("with unswell's report, held in the first run and not in the second. This run tests the command, with the bar written")
    add(f"on issue 525 first, on {third['pages']} more pages that no earlier run had read: the rest of the same new search, without any")
    add("repository of an earlier set, selected as before. `unswell fix` is the build of `main` at b868082e, whose code")
    add("`v0.1.0-alpha.11` ships; on every page of the two earlier runs it writes the same bytes as `v0.1.0-alpha.10`.")
    add("")
    add("| Fixed by | Pages fixed | Failed | Index lower | Index higher | High before | High after | Median share of tokens changed |")
    add("| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |")
    for name, label in (("fix", "`unswell fix`, no model"), ("write-good", "A model with write-good's findings")):
        arm = arms[name]
        changed = f"{100 * arm['median_changed']:.1f}%" if arm["median_changed"] is not None else "n/a"
        add(f"| {label} | {arm['fixed']} | {arm['failed']} | {arm['lower']} | {arm['higher']} | {arm['high_before']} "
            f"| {arm['high_after']} | {changed} |")
    add("")
    add("The panel is the two judges of the earlier runs. All three admitted judges together and the third judge alone are")
    add("read with the same bar and reported beside it.")
    add("")
    add("| Version under test | Against | Panel, of the decisive pages | Panel: no verdict | All three judges, of the decisive pages "
        "| Third judge alone, of its decisive answers | Third judge: the same |")
    add("| --- | --- | --- | ---: | --- | --- | ---: |")
    for key, entry in third["comparisons"].items():
        first, _, second = key.partition("|")
        tested, other = (second, first) if first == "original" else (first, second)
        together = alone = "not read"
        same = ""
        if key in extra:
            together, alone, same = verdict_cell(extra[key]["three_judges"]), verdict_cell(extra[key]), str(extra[key]["answers"]["same"])
        add(f"| {FRESH_VERSION_NAMES[tested]} | {FRESH_VERSION_NAMES[other]} | {verdict_cell(entry)} | {entry['tie']} "
            f"| {together} | {alone} | {same} |")
    registered = third["registered"]
    words = {True: "holds", False: "does not hold"}
    add("")
    add("The bars, for the panel:")
    add("")
    add(f"- The index is lower after `unswell fix` on at least 90% of the pages: {arms['fix']['lower']} of {third['pages']}, "
        f"{words[registered['fix_lowers_index']]}.")
    add("- The page after `unswell fix` is the less AI-sounding one on more than half of the decisive pages, with the lower end "
        f"of the interval above 50%. Against the original: {words[registered['original|fix']]}. Against the page fixed with "
        f"write-good's findings: {words[registered['fix|write-good']]}.")
    if extra:
        add("- The same bar for all three judges together: against the original "
            f"{words[extra['original|fix']['three_judges']['holds']]}, against write-good's page "
            f"{words[extra['fix|write-good']['three_judges']['holds']]}. For the third judge alone: against the original "
            f"{words[extra['original|fix']['holds']]}, against write-good's page {words[extra['fix|write-good']['holds']]}.")


def fix_fourth_section(add, fourth):
    arms, extra = fourth["arms"], fourth.get("extra_judge", {})
    add("")
    add("## Fixing by a report that names the replacement, on a fourth sample")
    add("")
    add("For a model that follows the report of `unswell check`, the bar against write-good's page held on the first sample")
    add("and failed on the second. That report gave the place of each mark and general advice, and the fixer chose its own")
    add("punctuation. The report now names the replacement at each mark, the one that `unswell fix` would make. This run tests")
    add(f"the fix by that report on the {fourth['pages']} pages of the third search's selection that no run had fixed or judged, with the bars")
    add("written on issue 525 first.")
    add("")
    add("| Fixed by | Pages fixed | Failed | Index lower | Index higher | High before | High after | Median share of tokens changed |")
    add("| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |")
    for name, label in (("report", "A model that follows the report"), ("write-good", "The same model with write-good's findings"),
                        ("fix", "`unswell fix`, no model")):
        arm = arms[name]
        changed = f"{100 * arm['median_changed']:.1f}%" if arm["median_changed"] is not None else "n/a"
        add(f"| {label} | {arm['fixed']} | {arm['failed']} | {arm['lower']} | {arm['higher']} | {arm['high_before']} "
            f"| {arm['high_after']} | {changed} |")
    add("")
    add("| Version under test | Against | Panel, of the decisive pages | Panel: no verdict | All three judges, of the decisive pages "
        "| Third judge alone, of its decisive answers | Third judge: the same |")
    add("| --- | --- | --- | ---: | --- | --- | ---: |")
    for key, entry in fourth["comparisons"].items():
        first, _, second = key.partition("|")
        tested, other = (second, first) if first == "original" else (first, second)
        together = alone = "not read"
        same = ""
        if key in extra:
            together, alone, same = verdict_cell(extra[key]["three_judges"]), verdict_cell(extra[key]), str(extra[key]["answers"]["same"])
        add(f"| {FRESH_VERSION_NAMES[tested]} | {FRESH_VERSION_NAMES[other]} | {verdict_cell(entry)} | {entry['tie']} "
            f"| {together} | {alone} | {same} |")
    registered = fourth["registered"]
    words = {True: "holds", False: "does not hold"}
    add("")
    add("The bars, for the panel:")
    add("")
    add(f"- The index is lower after the fix by the report on at least 90% of the pages: {arms['report']['lower']} of {fourth['pages']}, "
        f"{words[registered['report_lowers_index']]}.")
    add("- The report version is the less AI-sounding one on more than half of the decisive pages, with the lower end of the "
        f"interval above 50%. Against the original: {words[registered['original|report']]}. Against the page fixed with "
        f"write-good's findings: {words[registered['report|write-good']]}.")
    if extra:
        add("- The same bar for all three judges together: against the original "
            f"{words[extra['original|report']['three_judges']['holds']]}, against write-good's page "
            f"{words[extra['report|write-good']['three_judges']['holds']]}. For the third judge alone: against the original "
            f"{words[extra['original|report']['holds']]}, against write-good's page {words[extra['report|write-good']['holds']]}.")
    add("")
    add("The rows of `unswell fix` have no bar in this run.")


def fix_section(add, fix):
    add("")
    add("## Fixing what unswell points at")
    add("")
    add(f"{fix['pages']} pages as they were before a de-slop commit, from the held-out pairs: English, 150 to 1,500 words, index 40")
    add("or more, one page per repository. Each page has fixed versions. One is what `unswell fix --write` leaves, with no")
    add("model. For the others one model gets the page and one report with the same instruction: resolve what the report")
    add("points at, change nothing else, return the whole page. The report is the text report of `unswell check` with the")
    add("configuration that ships, or the findings of a third-party linter. A model fix that returns less than half of the")
    add("page, or leaves fewer than 150 prose words, counts as failed. A fixer call that returned nothing was repeated once.")
    add("")
    add("`unswell fix` changed after the registered run. It wrote a comma for an em dash between two full clauses, which is a")
    add("comma splice; it now writes a period and a capital letter there. The registered bars are read on the version of the")
    add("run. The comparisons of the version that ships were judged again on the same pages, which had been read once.")
    add("")
    add("| Fixed by | Pages fixed | Failed | Index lower | Index higher | High before | High after | Median share of tokens changed |")
    add("| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |")
    for name, label in FIX_ARM_NAMES.items():
        arm = fix["arms"][name]
        changed = f"{100 * arm['median_changed']:.1f}%" if arm["median_changed"] is not None else "n/a"
        add(f"| {label} | {arm['fixed']} | {arm['failed']} | {arm['lower']} | {arm['higher']} | {arm['high_before']} "
            f"| {arm['high_after']} | {changed} |")
    add("")
    add("Two judges see two versions of a page as A and B, in an order fixed by a hash, and name the one that reads more like")
    add("text written by a language model. Both were admitted on held-out pairs before and after a de-slop edit by people; the")
    add("section on the judges has their numbers. A page-level verdict adds the two judges: an answer counts for the version it")
    add("did not name, and \"same\" counts for neither.")
    add("")
    add("| Version under test | Against | Pages where it is the less AI-sounding | The other | Tie | Share of decisive pages |")
    add("| --- | --- | ---: | ---: | ---: | --- |")
    for key, entry in fix["comparisons"].items():
        first, _, second = key.partition("|")
        tested, other = (second, first) if first == "original" else (first, second)
        low, high = entry["interval"]
        decisive = entry["tested_less"] + entry["other_less"]
        add(f"| {FIX_VERSION_NAMES[tested]} | {FIX_VERSION_NAMES[other]} | {entry['tested_less']} | {entry['other_less']} | {entry['tie']} "
            f"| {pct(entry['tested_less'], decisive)} ({100 * low:.1f}% to {100 * high:.1f}%) |")
    add("")
    add("Each judge alone, as answers naming the first version, the second, both the same, and no answer:")
    add("")
    judges = list(next(iter(fix["comparisons"].values()))["judges"])
    add("| First version, then second | " + " | ".join(f"`{judge}`" for judge in judges) + " |")
    add("| --- | " + " | ".join("---" for _ in judges) + " |")
    for key, entry in fix["comparisons"].items():
        first, _, second = key.partition("|")
        cells = " | ".join("{first} / {second} / {same} / {missing}".format(**entry["judges"][judge]) for judge in judges)
        add(f"| {FIX_VERSION_NAMES[first]}, then {FIX_VERSION_NAMES[second]} | {cells} |")
    add("")
    registered, best = fix["registered"], fix["best_linter"]
    words = {True: "holds", False: "does not hold"}
    add("Registered on issue 525 before a fixed version of a held-out page existed:")
    add("")
    for name, label, key in (("fix-comma", "`unswell fix`", "fix_lowers_index"), ("report", "the fix by unswell's report", "report_lowers_index")):
        add(f"- The index is lower after {label} on at least 90% of the pages: {fix['arms'][name]['lower']} of {fix['pages']}, "
            f"{words[registered[key]]}.")
    rows = [("original|fix-comma", "The version after `unswell fix`", "the original", registered["original|fix"]),
            ("original|report", "The version fixed by unswell's report", "the original", registered["original|report"]),
            (f"report|{best}", "The version fixed by unswell's report", f"the version fixed by the findings of {FIX_LINTER_NAMES[best]}",
             registered["report|best_linter"])]
    second = fix["second_reading"]
    for key, tested, other, held in rows:
        entry = fix["comparisons"][key]
        low, high = entry["interval"]
        decisive = entry["tested_less"] + entry["other_less"]
        add(f"- {tested} is the less AI-sounding one against {other} on more than half of the decisive pages, with the lower "
            f"end of the interval above 50%: {entry['tested_less']} of {decisive}, {pct(entry['tested_less'], decisive)} "
            f"({100 * low:.1f}% to {100 * high:.1f}%), {words[held]}.")
    entry = fix["comparisons"]["original|fix"]
    low, high = entry["interval"]
    decisive = entry["tested_less"] + entry["other_less"]
    add("")
    add(f"The second reading, with `unswell fix` as it ships: the index is lower on {fix['arms']['fix']['lower']} of {fix['pages']} pages, "
        f"which {words[second['fix_lowers_index']]}, and the fixed page is the less AI-sounding one against the original on "
        f"{entry['tested_less']} of {decisive} decisive pages, {pct(entry['tested_less'], decisive)} ({100 * low:.1f}% to {100 * high:.1f}%), "
        f"which {words[second['original|fix']]}.")
    add("")
    add(f"The best third-party linter is the one whose version does best against the original on these pages: {FIX_LINTER_NAMES[best]}.")
    add("")
    add("The fixer and the two judges of the panel are models of one family, and the fixer is one of the two. The versions")
    add("fixed by a linter's findings come from the same fixer, so a judge that likes any rewrite by its own family would like")
    add("them too; `unswell fix` involves no model. The judges often give the marks that the index counts as their reasons, so")
    add("the panel and the index are not independent readings.")
    if fix.get("extra_judge"):
        fix_extra_section(add, fix["extra_judge"], "gpt-6-astra")


def system_name(name):
    parts = name.split("-")
    if parts[0] == "rules":
        return {"default": "`unswell check`, default rules", "opener": "`unswell check`, default rules with the listed opener",
                "all": "`unswell check`, every rule"}[parts[1]]
    if parts[0] == "review":
        text = f"`unswell review`, {MODELS[parts[1]]}"
        return text + (f", confirmed by {MODELS[parts[2]]}" if len(parts) > 2 else "")
    if parts[0] == "trim":
        return f"`unswell trim`, {MODELS[parts[1]]}"
    if parts[0] == "lamp5":
        return f"Published span prompt, {MODELS[parts[1]]}"
    return name


def system_order(name):
    kinds = ["rules", "review", "trim", "lamp5"]
    kind = name.split("-")[0]
    return (kinds.index(kind) if kind in kinds else len(kinds), name)


def render(results):
    sentences, lines = results["sentences"], []
    add = lines.append
    add("# Results")
    add("")
    add("Generated by `render.py` from `results.json`, which `verify.py` recounts from the data files.")
    add("")
    add("## Sets")
    add("")
    add("| Set | Part | Documents | Sentences | Edited by a person |")
    add("| --- | --- | ---: | ---: | ---: |")
    for key, label in SETS:
        for split in ("dev", "test", "ceiling"):
            entry = sentences["sets"].get(f"{key}/{split}")
            if entry:
                part = {"dev": "development", "test": "test", "ceiling": "three writers"}[split]
                add(f"| {label} | {part} | {entry['documents']:,} | {entry['sentences']:,} | {pct(entry['edited'], entry['sentences'])} |")
    add("")
    add("## Agreement between editors")
    add("")
    ceiling = sentences["ceiling"]
    unedited = ceiling["sentences"] - ceiling["edited"]
    add(f"On {ceiling['paragraphs']} LAMP paragraphs edited by three writers, one writer's edited sentences are measured against")
    add(f"another's, in both directions. A sentence that one writer edited was edited by the other in "
        f"{pct(ceiling['edited_flagged'], ceiling['edited'])} of cases, and a")
    add(f"sentence that one writer left alone was edited by the other in {pct(ceiling['unedited_flagged'], unedited)}. "
        f"The difference is {points(ceiling['j'])} points.")
    words = ceiling.get("words")
    if words:
        marked = words["marked_by"]
        total = sum(marked.values())
        add("")
        add(f"Word by word, Cohen's kappa between two writers is {words['kappa']:.2f}. Of the words in paragraphs with three")
        add(f"writers, {pct(marked['3'], total)} are marked by all three, {pct(marked['2'], total)} by two, "
            f"{pct(marked['1'], total)} by one, and {pct(marked['0'], total)} by none.")
    add("")
    add("## Systems against the sentences a person edited")
    add("")
    add("TPR is the share of edited sentences that a finding touches, FPR the share of unedited sentences. J is their")
    add("difference in points; a system that flags at random has J = 0 at any flagging rate.")
    for split, title in (("test", "Test"), ("dev", "Development")):
        add("")
        add(f"### {title}")
        for key, label in SETS:
            rows = [(name, entry[f"{key}/{split}"]) for name, entry in sorted(sentences["systems"].items(), key=lambda kv: system_order(kv[0]))
                    if f"{key}/{split}" in entry]
            if not rows:
                continue
            add("")
            add(f"{label}:")
            add("")
            add("| System | Documents | Sentences flagged | TPR | FPR | J | Findings |")
            add("| --- | ---: | ---: | ---: | ---: | ---: | ---: |")
            for name, entry in rows:
                add(f"| {system_name(name)} | {entry['documents']:,} | {pct(entry['flagged'], entry['sentences'])} | "
                    f"{pct(entry['edited_flagged'], entry['edited'])} | "
                    f"{pct(entry['unedited_flagged'], entry['sentences'] - entry['edited'])} | {points(entry['j'])} | {entry['findings']:,} |")
    preferences = results["preferences"]
    add("")
    add("## Preferences between versions")
    add("")
    add(f"The LAMP preference study: {preferences['triplets']} triplets, three writers each, {preferences['judgments']} judgments.")
    add("")
    add("| Pair | Judgments preferring the first | Two writers give the same answer |")
    add("| --- | ---: | ---: |")
    for name, label in PAIRS:
        entry = preferences["writers"][name]
        add(f"| {label} | {pct(entry['judgments_preferring_first'], preferences['judgments'])} | "
            f"{pct(entry['writer_pairs_agreeing'], entry['writer_pairs'])} |")
    add("")
    add("Judge models on the same triplets. The first number is how often the model prefers the first version of the")
    add("pair, the second how often its answer equals the majority of the three writers.")
    add("")
    add("| Judge | Triplets | " + " | ".join(label for _, label in PAIRS) + " | Expert-edited answer preferred (Beemo) | Admitted |")
    add("| --- | ---: | " + " | ".join("---:" for _ in PAIRS) + " | ---: | --- |")
    for judge, entry in sorted(preferences["judges"].items()):
        cells = [f"{pct(entry['pairs'][name]['prefers_first'], entry['triplets'])} / {pct(entry['pairs'][name]['agrees_with_majority'], entry['triplets'])}"
                 for name, _ in PAIRS]
        expert = results["expert_pairs"].get(judge)
        expert_cell = f"{pct(expert['prefers_edited'], expert['pairs'])} of {expert['pairs']}" if expert else "not run"
        admitted = "yes" if results["judges"][judge]["admitted"] else "no"
        add(f"| {MODELS.get(judge, judge)} | {entry['triplets']} | " + " | ".join(cells) + f" | {expert_cell} | {admitted} |")
    add("")
    add("A judge is admitted when it agrees with the writers' majority on at least 70% of the pairs \"writer-edited or")
    add("raw\", on at least 65% of the pairs \"model-edited or raw\", and prefers the expert-edited Beemo answer in at least")
    add("65% of pairs.")
    readers = results.get("readers")
    if readers:
        add("")
        add("## Reader answers, withdrawn as a measure")
        add("")
        add(f"The blind comparison of paragraphs before and after revision was first answered by {readers['readers']} reader"
            f"{'' if readers['readers'] == 1 else 's'}. The maintainer")
        add("withdrew these answers as a measure, and no decision rests on them. They are kept as recorded. In the reviewed")
        add("arm the revision follows the findings of `unswell review`; in the control arm the same model rewrote as many")
        add("random other sentences.")
        add("")
        add("| Arm | Revised preferred | Original preferred | No real difference | Revised among decisive answers | 95% interval |")
        add("| --- | ---: | ---: | ---: | ---: | --- |")
        for arm in ("reviewed", "control"):
            entry = readers["arms"][arm]
            low, high = entry["interval"]
            add(f"| {arm} | {entry['revised']} | {entry['original']} | {entry['same']} | {pct(entry['revised'], entry['decisive'])} "
                f"of {entry['decisive']} | {100 * low:.1f}% to {100 * high:.1f}% |")
    panel = results.get("panel")
    if panel:
        add("")
        add("## The two comparisons, judged by a consensus of models")
        add("")
        add("Four models answered every pair by a rule about information: a version is worse if it lacks a fact, condition,")
        add("qualification, or instruction that the other has; otherwise the shorter one is better when the extra words carry")
        add("no information. `claude-fable-5-2` has three votes, `gpt-6-astra` two, and the other two one each; a tie goes to")
        add("`claude-fable-5-2`. The maintainer set these weights after the answers were read. The weights first registered")
        add("(two votes for `claude-fable-5-2`, one for each other model) give the same consensus on every pair.")
        titles = {"first": "Free revision", "trim": "Revision that may only delete words",
                  "product": "The command `unswell trim` as built, on documents outside both comparisons"}
        for comparison in ("first", "trim", "product"):
            entry = panel.get(comparison)
            if not entry:
                continue
            add("")
            add(f"{titles[comparison]}: {entry['pairs']} pairs. All four models agree on {entry['unanimous']}; on {entry['opposed']} at least one")
            add("model prefers the revision and another the original.")
            add("")
            add("| Arm | Revised preferred | Original preferred | No real difference | Revised among decisive answers | 95% interval |")
            add("| --- | ---: | ---: | ---: | ---: | --- |")
            order = [arm for arm in ("reviewed", "control", "product") if arm in entry["arms"]]
            for arm in order:
                row = entry["arms"][arm]
                low, high = row["interval"]
                add(f"| {arm} | {row['revised']} | {row['original']} | {row['same']} | {pct(row['revised'], row['decisive'])} "
                    f"of {row['decisive']} | {100 * low:.1f}% to {100 * high:.1f}% |")
            add("")
            add("Single models, revised preferred / original preferred: " + "; ".join(
                f"{MODELS.get(judge, '`claude-fable-5-2`' if judge == 'fable' else judge)} "
                + ", ".join(f"{entry['arms'][arm]['judges'][judge]['revised']}/{entry['arms'][arm]['judges'][judge]['original']}"
                            for arm in order)
                for judge in ("fable", "claude", "astra", "qwen")) + "."
                + (" Two numbers per model are the reviewed and the control arm." if len(entry["arms"]) == 2 else ""))
            family = entry["one_vote_per_family"]
            add("")
            add("With one equal vote per model family (`claude-fable-5-2`, `gpt-6-astra`, `private-host/Qwen3.8-27B`), revised "
                "among decisive answers: " + "; ".join(
                    f"{arm} {pct(family[arm]['revised'], family[arm]['decisive'])} of {family[arm]['decisive']}" for arm in order)
                + ". " + ("That meets the rule." if all(family[arm]["meets_rule"] for arm in order) else
                          "That does not meet the rule in every arm."))
            add("")
            if comparison == "product":
                add("The command meets the registered rule." if entry["product_meets_rule"] else
                    "The command does not meet the registered rule.")
                continue
            reached = [arm for arm in ("reviewed", "control") if entry[f"{arm}_meets_rule"]]
            if len(reached) == 2:
                floor = "Both arms reach the registered 70% of decisive answers."
            elif reached:
                floor = f"Only the {reached[0]} arm reaches the registered 70% of decisive answers."
            else:
                floor = "Neither arm reaches the registered 70% of decisive answers."
            lead = "The reviewed arm leads the control arm by the registered 15 points." if entry["ahead_of_control"] else \
                "The reviewed arm is not 15 points ahead of the control arm."
            add(f"{floor} {lead}")
    if has_words(results):
        add("")
        add("## Words")
        add("")
        add(f"The comparisons at the level of words are in [{WORDS_PAGE}]({WORDS_PAGE}): the words that trims remove, the")
        add("listed chat opener, and each rule against the words an editor changed.")
    if has_index(results):
        add("")
        add("## Page index and fix")
        add("")
        add(f"The page index of AI-sounding documentation, the fix experiment, and its judges are in [{INDEX_PAGE}]({INDEX_PAGE}).")
    return "\n".join(lines) + "\n"


def has_words(results):
    return any(results.get(name) for name in ("trims", "openers", "rule_words", "vocabulary", "deslop", "review_repeat",
                                              "default_load"))


def has_index(results):
    return any(results.get(name) for name in ("page_index", "fix", "judges_of_edits", "fix_fresh", "fix_third", "fix_fourth"))


def render_words(results):
    """Render the page of the comparisons at the level of words."""
    lines = []
    add = lines.append
    add("# Results at the level of words")
    add("")
    add("Generated by `render.py` from `results.json`, which `verify.py` recounts from the data files.")
    add("[RESULTS.md](RESULTS.md) has the other results.")
    if results.get("trims"):
        trim_section(add, results["trims"])
    if results.get("openers"):
        opener_section(add, results["openers"])
    if results.get("rule_words"):
        rule_word_section(add, results["rule_words"])
    if results.get("vocabulary"):
        vocabulary_section(add, results["vocabulary"])
    if results.get("deslop"):
        deslop_section(add, results["deslop"])
    if results.get("review_repeat"):
        review_repeat_section(add, results["review_repeat"])
    if results.get("default_load"):
        defaults_section(add, results["default_load"], results.get("manual"))
    return "\n".join(lines) + "\n"


def render_index(results):
    """Render the page of the page index, the fix experiment, and its judges."""
    lines = []
    add = lines.append
    add("# Results of the page index and of the fix")
    add("")
    add("Generated by `render.py` from `results.json`, which `verify.py` recounts from the data files.")
    add("[RESULTS.md](RESULTS.md) and [RESULTS-words.md](RESULTS-words.md) have the other results.")
    if results.get("page_index"):
        page_index_section(add, results["page_index"], results.get("readme_load"))
    if results.get("fix"):
        fix_section(add, results["fix"])
    if results.get("judges_of_edits"):
        judges_section(add, results["judges_of_edits"])
    if results.get("fix_fresh"):
        fix_fresh_section(add, results["fix_fresh"])
    if results.get("fix_third"):
        fix_third_section(add, results["fix_third"])
    if results.get("fix_fourth"):
        fix_fourth_section(add, results["fix_fourth"])
    return "\n".join(lines) + "\n"


# Every generated page with its renderer, and whether a record with these results has the page. The second page
# exists only when the record has word-level results, and the third only with the page index or the fix experiment.
WORDS_PAGE, INDEX_PAGE = "RESULTS-words.md", "RESULTS-index.md"
PAGES = {"RESULTS.md": render, WORDS_PAGE: render_words, INDEX_PAGE: render_index}
PRESENT = {"RESULTS.md": lambda results: True, WORDS_PAGE: has_words, INDEX_PAGE: has_index}


if __name__ == "__main__":
    root = Path(sys.argv[1]) if len(sys.argv) > 1 else HERE
    loaded = json.loads((root / "results.json").read_text())
    for page, function in PAGES.items():
        if PRESENT[page](loaded):
            (root / page).write_text(function(loaded))
            print("written", root / page)
