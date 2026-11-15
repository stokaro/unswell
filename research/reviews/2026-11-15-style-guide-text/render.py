"""Render RESULTS.md of the record from the recounted results. verify.py compares the page on disk with this one."""

WRITERS = ("gpt-6-astra", "private-host/Qwen3.6-35B-A3B")
JUDGES = ("gpt-6-astra", "private-host/Qwen3.8-27B", "private-host/Qwen3.6-35B-A3B")


def percent(value):
    return "n/a" if value is None else f"{100 * value:.1f}%"


def interval(pair):
    return "" if not pair else f" ({percent(pair[0])} to {percent(pair[1])})"


def code(name):
    return f"`{name}`"


def held(result, name):
    return "holds" if result["registered"][name] else "**does not hold**"


def rate(entry):
    return f"{entry['model']} of {entry['units']}, {percent(entry['rate'])}"


def decisive(entry):
    return f"{entry['model']} of {entry['model'] + entry['other']}, {percent(entry['decisive'])}{interval(entry['interval'])}"


def title(name):
    """The words for a group of pages in a table or a sentence."""
    if name == "all":
        return "All"
    if name == "set-s":
        return "Style-guide set"
    if name == "all under a style guide":
        return "All texts under a style guide"
    kind, writer = name.split(", ")
    return f"{kind.capitalize()} texts of {code(writer)}"


def judges_section(result):
    judges = result["judges"]
    rows = [("Admission, same subject", "plain", "admission, same subject", "admission"),
            ("Admission, other subject", "cross", "admission, other subject", "admission"),
            ("Guided, same subject", "guided", "guided texts are told from human posts", "tells"),
            ("Style-guide set", "set", "the style-guide set is told from human posts", "tells")]
    lines = ["## Do the judges tell such text from human text", "",
             "Three blind judges read the openings of two posts and name the one that reads more like text written by a",
             "language model. The panel picks a text when the votes for it outnumber the votes for the other. The rate is",
             "over all units, so a tie counts against.", "",
             "| Comparison | Panel picks the model's text | Decisive units | Bar | |", "| --- | --- | --- | --- | --- |"]
    for label, key, bar, limit in rows:
        entry = judges[key]["panel"]
        lines.append(f"| {label} | {rate(entry)} | {decisive(entry)} | {percent(result['bars'][limit])} | {held(result, bar)} |")
    lines += ["", "Each judge alone, units where it picks the model's text:", "",
              "| Comparison | " + " | ".join(code(j) for j in JUDGES) + " |", "| --- |" + " --- |" * len(JUDGES)]
    for label, key, _, _ in rows:
        lines.append(f"| {label} | " + " | ".join(f"{judges[key]['judges'][j]['model']} of {judges[key]['judges'][j]['units']}" for j in JUDGES) + " |")
    lines += ["", "Judges of the other model family than the writer, alone:", "", "| Comparison | Writer | Picks the model's text | Ties |",
              "| --- | --- | --- | --- |"]
    for label, key, _, _ in rows[:3]:
        for writer in WRITERS:
            entry = judges[key]["writers"][writer]["other_family"]
            lines.append(f"| {label} | {code(writer)} | {entry['model']} of {entry['units']} | {entry['tie']} |")
    guides = judges["guided"]["guides"]
    unreadable = sum(judges[key]["unreadable"] for _, key, _, _ in rows)
    lines += ["", f"Under the first style guide the panel picks the model's text on {rate(guides['A'])} of the pairs, under the second on "
              f"{rate(guides['B'])}. Answers that could not be read: {unreadable}; such an answer counts as no vote.", ""]
    return lines


def pages_section(result):
    measured = result["pages"]["measured"]
    names = ["set-s"] + ["guided, " + w for w in WRITERS] + ["all under a style guide"]
    label = lambda name: (name.split(", ")[0].capitalize() + " texts of " + code(name.split(", ")[1]) + ", no style guide"
                          if name.startswith("plain") else title(name))
    lines = ["## Does the index separate such text from human posts", "",
             "The area under the ROC curve of the index, each set against the human blog posts of the measured half. The earlier",
             "release is `v0.1.0-alpha.12`, whose index is the first reading alone.", "",
             "| Set | Pages | This build | Earlier release | Bar | |", "| --- | --- | --- | --- | --- | --- |"]
    for name in names:
        lines.append(f"| {label(name)} | {measured['sets'][name]['build']['pages']} | {measured['separation'][name]:.3f} | "
                     f"{measured['separation_earlier'][name]:.3f} | {result['bars']['separation']:.2f} | {held(result, 'separation, ' + name)} |")
    for writer in WRITERS:
        name = "plain, " + writer
        lines.append(f"| {label(name)} | {measured['sets'][name]['build']['pages']} | {measured['separation'][name]:.3f} | "
                     f"{measured['separation_earlier'][name]:.3f} | no bar | |")
    lines += ["", "Levels with this build:", "", "| Set | Pages | Median index | `high` | `elevated` | `low` | `high` with the earlier release |",
              "| --- | --- | --- | --- | --- | --- | --- |"]
    for name in ["human-blog", "readme"] + names[:-1] + ["plain, " + w for w in WRITERS]:
        build, earlier = measured["sets"][name]["build"], measured["sets"][name]["earlier"]
        shown = {"human-blog": "Human blog posts", "readme": "README files"}.get(name) or label(name)
        lines.append(f"| {shown} | {build['pages']} | {build['median']} | {build['high']} | {build['elevated']} | {build['low']} | {earlier['high']} |")
    quiet = measured["quiet"]
    lines += ["", f"Human pages together: {quiet['high']} of {quiet['pages']} at `high`, {percent(quiet['share'])}{interval(quiet['interval'])}; "
              f"bar {percent(result['bars']['quiet'])}, {held(result, 'quiet on human pages')}. The earlier release puts {quiet['high_earlier']} of them at "
              f"`high`, and {quiet['high_by_rhythm']} are `high` by the second reading.", ""]
    fitting = result["pages"]["fitting"]
    lines += ["The same on the fitting half, where the second reading was fitted:", "", "| Set | Pages | AUC |", "| --- | --- | --- |"]
    lines += [f"| {label(name)} | {fitting['sets'][name]['build']['pages']} | {fitting['separation'][name]:.3f} |" for name in names]
    lines += ["", f"Human pages of the fitting half at `high`: {fitting['quiet']['high']} of {fitting['quiet']['pages']}, of which "
              f"{fitting['quiet']['high_by_rhythm']} by the second reading.", ""]
    return lines


def earlier_section(result):
    earlier = result["earlier"]
    return ["## Does the earlier result stand", "",
            f"The held-out reading of the earlier record, with this build: the pages before a de-slop commit separate from the "
            f"README files with an AUC of {earlier['auc']:.3f} on {earlier['pairs']} pages and {earlier['readme']} README files "
            f"(first reading alone: {earlier['auc_first_reading']:.3f}); bar {result['bars']['earlier']:.2f}, {held(result, 'earlier result')}. "
            f"The index is lower after the commit in {earlier['lower']} of {earlier['differ']} pairs that differ, {percent(earlier['lower_share'])}. "
            f"README files at `high`: {earlier['readme_high']} of {earlier['readme']}, {percent(earlier['readme_high_share'])} "
            f"(first reading alone: {earlier['readme_high_first_reading']}).", ""]


def revisions_section(result):
    revisers = [(position, reviser, result["revisions"][reviser]) for position, reviser in enumerate(WRITERS) if reviser in result["revisions"]]
    lines = ["## Does a revision by the report help", "",
             "A model gets the page and the lines of the report about the index, with the instruction to resolve what the report",
             "points at and to change nothing else. The judges then compare the opening of the page with the opening of its",
             f"revision. The bars are for the reviser {code(WRITERS[0])}; the other reviser is reported beside them.", "",
             "| Reviser | Pages | Revised | Index lower | Median before and after | `high` before and after |"
             " Panel finds the page more like a model | Ties | Words, revision to page | Em dashes added |",
             "| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |"]
    for _, reviser, entry in revisers:
        for name, group in entry["groups"].items():
            ratio = "not kept" if group["words_ratio"] is None else f"{group['words_ratio']:.3f}"
            lines.append(f"| {code(reviser)} | {title(name)} | {group['revised']} | {group['lower']}, {percent(group['lower_share'])} | "
                         f"{group['before']['median']} and {group['after']['median']} | {group['before']['high']} and {group['after']['high']} | "
                         f"{decisive(group['panel'])} | {group['panel']['tie']} | {ratio} | {group['em_dashes_added']} |")
    lines += ["", "Of the pages given to a reviser, some have a report without a listed mark, and some came back without a usable revision: "
              + "; ".join(f"{code(reviser)} {entry['pages']} given, {entry['no_mark_listed']} and {entry['no_revision']}" for _, reviser, entry in revisers) + ".",
              "", "Each judge alone, pages where it finds the page more like a model, of its decisive answers:", "",
              "| Reviser | Pages | " + " | ".join(code(j) for j in JUDGES) + " |", "| --- | --- |" + " --- |" * len(JUDGES)]
    for _, reviser, entry in revisers:
        for name, group in entry["groups"].items():
            lines.append(f"| {code(reviser)} | {title(name)} | " + " | ".join(
                f"{group['judges'][j]['model']} of {group['judges'][j]['model'] + group['judges'][j]['other']}" for j in JUDGES) + " |")
    lines.append("")
    if revisers and revisers[0][0] == 0:
        groups = [name for name in revisers[0][2]["groups"] if name != "all"]
        lines += [f"Bars: the index is lower on {percent(result['bars']['index_lower'])} or more of the revised pages, "
                  f"{held(result, 'index lower after the revision')}. The panel finds the page more like a model than its revision on more "
                  f"than half of the decisive pages with the lower end of the interval above 50%, {held(result, 'revision less like a model, all pages')}. "
                  "The same on more than half for each kind of page: "
                  + "; ".join(f"{title(name)[0].lower() + title(name)[1:]}, {held(result, 'revision less like a model, ' + name)}" for name in groups) + ".", ""]
    return lines


def page(result):
    readings = result["readings"]
    weights = lambda entry: ", ".join(f"`{name}` {value}" for name, value in entry["weights"].items()) + f", constant {entry['bias']}"
    lines = ["# Text written under a style guide: results", "",
             "Generated by `verify.py --write` from the data of this record. Do not edit it by hand.", "",
             f"{sum(result['registered'].values())} of {len(result['registered'])} registered bars hold.", "",
             "The index is the higher of two readings, each 100 times the logistic function of a constant plus the weighted natural",
             "logarithms of one plus a rate per 1,000 prose words. First reading, the marks: " + weights(readings["marks"]) + ".",
             "Second reading, the rhythm, over the sentences in runs: " + weights(readings["rhythm"]) + ".", ""]
    lines += judges_section(result) + pages_section(result) + earlier_section(result) + revisions_section(result)
    return "\n".join(lines).rstrip("\n") + "\n"
