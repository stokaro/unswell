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
    return "\n".join(lines) + "\n"


def has_words(results):
    return any(results.get(name) for name in ("trims", "openers", "rule_words", "vocabulary"))


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
    return "\n".join(lines) + "\n"


# Every generated page and its renderer. The second page exists only when the record has word-level results.
WORDS_PAGE = "RESULTS-words.md"
PAGES = {"RESULTS.md": render, WORDS_PAGE: render_words}


if __name__ == "__main__":
    root = Path(sys.argv[1]) if len(sys.argv) > 1 else HERE
    loaded = json.loads((root / "results.json").read_text())
    for page, function in PAGES.items():
        if page == "RESULTS.md" or has_words(loaded):
            (root / page).write_text(function(loaded))
            print("written", root / page)
