#!/usr/bin/env python3
"""Render docs/research/rule-qualification.md from the workspace measurements.

usage: gen_docs.py <workspace> <out markdown>
"""
import collections, json, sys
from pathlib import Path

W, OUT = Path(sys.argv[1]), Path(sys.argv[2])


def docs(name):
    rows = [json.loads(line) for line in open(W / name)]
    return len(rows), len({d["repository"] for d in rows}), sum(len(d["items"]) for d in rows), sum(len(d["text"].split()) for d in rows)


def reviews(folder, only=None):
    out = {}
    for name in ("claude", "astra"):
        for path in (W / folder / name).glob("*.json"):
            if only and not path.name.startswith(only):
                continue
            for finding in json.loads(path.read_text())["findings"]:
                out.setdefault(finding["id"], {"rule": finding["rule"], "group": finding["group"],
                                                "phrase": finding["phrase"]})[name] = finding.get("justified")
    return [v for v in out.values() if v.get("claude") and v.get("astra")]


def rule_rows(both, minimum, by_group=True):
    table = collections.defaultdict(collections.Counter)
    for v in both:
        key = v["rule"] if not (v["group"] and by_group) else f"{v['rule']} ({v['group']} candidates)"
        c = table[key]
        c["n"] += 1
        c["yes"] += v["claude"] == "yes" and v["astra"] == "yes"
        c["any"] += v["claude"] == "yes" or v["astra"] == "yes"
        c["no"] += v["claude"] == "no" and v["astra"] == "no"
    rows = [(k, c["n"], c["yes"], c["any"], c["no"]) for k, c in table.items() if c["n"] >= minimum]
    return sorted(rows, key=lambda r: (-r[2] / r[1], -r[1]))


def wilson(yes, n, z=1.96):
    if not n:
        return (0.0, 0.0)
    phat = yes / n
    centre = (phat + z * z / (2 * n)) / (1 + z * z / n)
    half = z * ((phat * (1 - phat) / n + z * z / (4 * n * n)) ** 0.5) / (1 + z * z / n)
    return (100 * (centre - half), 100 * (centre + half))


def pct(a, b):
    return f"{100 * a / b:.1f}%" if b else "n/a"


def labels(names):
    both = one = items = 0
    for name in names:
        for line in open(W / name):
            doc = json.loads(line)
            paths = [W / "ann" / n / f"{doc['doc_id']}.json" for n in ("astra", "claude")]
            if not all(p.exists() for p in paths):
                continue
            a, c = [{f["item"]: f for f in json.loads(p.read_text())["flags"]} for p in paths]
            for item in doc["items"]:
                pa = bool(a.get(item["id"]) and a[item["id"]]["label"] == "needs_revision")
                pc = bool(c.get(item["id"]) and c[item["id"]]["label"] == "needs_revision")
                items += 1
                both += pa and pc
                one += pa != pc
    return items, both, one


def coverage(report_name, pairs):
    report = json.loads((W / "engine" / report_name).read_text())
    items = {}
    for folder, name in pairs:
        for line in open(W / name):
            doc = json.loads(line)
            paths = [W / "ann" / n / f"{doc['doc_id']}.json" for n in ("astra", "claude")]
            if not all(p.exists() for p in paths):
                continue
            a, c = [{f["item"]: f for f in json.loads(p.read_text())["flags"]} for p in paths]
            rows = []
            for item in doc["items"]:
                pa = bool(a.get(item["id"]) and a[item["id"]]["label"] == "needs_revision")
                pc = bool(c.get(item["id"]) and c[item["id"]]["label"] == "needs_revision")
                rows.append([item["start"], item["end"], pa and pc, set()])
            items[f"{folder}/{doc['doc_id']}.md"] = rows
    counts = collections.Counter(f["rule_id"] for f in report["findings"])
    for finding in report["findings"]:
        rows = items.get(finding["primary"]["path"], [])
        spans = finding["primary"].get("segments") or [finding["primary"]["span"]]
        for row in rows:
            if any(s["start"] < row[1] and row[0] < s["end"] for s in spans):
                row[3].add(finding["rule_id"])
    flat = [row for rows in items.values() for row in rows if row[2]]
    words = sum(d["prose_words"] for d in report["documents"])
    return {"findings": len(report["findings"]), "long": counts["syntax.long-sentence"], "words": words,
            "consensus": len(flat), "touched": sum(1 for r in flat if r[3]),
            "touched_other": sum(1 for r in flat if r[3] - {"syntax.long-sentence"}), "counts": counts}


v0, dev1, dev2 = docs("docs.jsonl"), docs("docs-realdev.jsonl"), docs("docs-realdev2.jsonl")
real, real2 = docs("docs-real.jsonl"), docs("docs-real2.jsonl")
dev = reviews("rev")
final = reviews("rev-final")
decisions = json.loads((W / "engine/release.decisions.json").read_text())
assignment = json.loads((W / "engine/release.assignment.json").read_text())
old = coverage("report-final-old.json", (("real", "docs-real.jsonl"), ("real2", "docs-real2.jsonl")))
new = coverage("report-final-new.json", (("real", "docs-real.jsonl"), ("real2", "docs-real2.jsonl")))
audit = json.loads((W / "audit-spans-final.json").read_text())
agreement_dev = sum(v["claude"] == v["astra"] for v in dev) / len(dev)
v0_labels = labels(["docs.jsonl"])
final_labels = labels(["docs-real.jsonl", "docs-real2.jsonl"])
lines = []
add = lines.append
add("# Rule qualification on the assistant-panel corpus")
add("")
add("This record supports the rule defaults and phrase lists that changed after")
add("[ADR 0046](../adr/0046-assistant-panel-corpus.md) and [ADR 0047](../adr/0047-qualified-defaults-and-assistant-review.md).")
add("Every reviewer here is an AI assistant. No number is a human judgment, and none")
add("measures agreement with a native-speaking editor. The labels, reviews, and")
add("manifests are in the")
add("[corpus record](../../research/corpus/2026-11-11-assistant-panel/README.md).")
add("")
add("## Documents")
add("")
add("| Set | Role | Documents | Repositories | Items | Words |")
add("| --- | --- | ---: | ---: | ---: | ---: |")
for name, role, d in (("Long-prose study: 38 sources and 455 generated versions", "development", v0),
                      ("Real pages, first development set", "development", dev1),
                      ("Real pages, second development set", "development", dev2),
                      ("Real pages, final test part 1", "final test", real),
                      ("Real pages, final test part 2", "final test", real2)):
    add(f"| {name} | {role} | {d[0]:,} | {d[1]:,} | {d[2]:,} | {d[3]:,} |")
add("")
add("Real pages are README files and documentation pages of public repositories. For four")
add("languages and two creation bands, repositories were taken in order of stars. A")
add("repository named in any earlier research material was skipped. The final test was")
add("selected before any phrase list existed and was opened once, for the lists in this record.")
add("An item is a sentence, a paragraph without separate sentences, or a fragment of six or")
add("more words, with byte ranges from the engine's own extraction.")
add("")
add("## Reviewers")
add("")
add("| Role | Model | Access |")
add("| --- | --- | --- |")
add("| Annotator and reviewer A | `gpt-6-astra`, medium effort | command-line client, tools disabled |")
add("| Annotator and reviewer B | `claude-sonnet-5-5` | command-line client, tools disabled |")
add("| Auditor | `private-host/Qwen3.8-27B` | chat completion endpoint |")
add("")
add("## Two questions, two answers")
add("")
add("Annotators read each complete document and marked the items that **need revision**")
add("under rubric `unswell-editorial-v1`. Reviewers read each engine finding and judged")
add("whether it is **justified under its rule's statement**. The two questions give")
add("different numbers. An annotator asked for required changes leaves most sentences")
add("with a word such as `comprehensive` unmarked; a reviewer asked about the rule accepts")
add("most findings on it. Rule defaults are judged by the second question. Coverage is")
add("reported against the first.")
add("")
add(f"On the long-prose study the annotators marked {v0_labels[1] + v0_labels[2]:,} of {v0_labels[0]:,} items;")
add(f"both marked {v0_labels[1]:,}. On the final test both marked {final_labels[1]:,} of {final_labels[0]:,} items.")
add("Cohen's kappa on the long-prose study is 0.43. Sentence-level necessity is a judgment")
add("on which two strong models often differ, so every recall figure below is a share of")
add("their agreement, not of a ground truth.")
add("")
add("## Development: every rule enabled")
add("")
add("A configuration enabled all 58 rules and extended the phrase lists with 321 candidate")
add("words and phrases. Both reviewers judged each finding, at most eight per rule and")
add(f"document. They agreed on yes, no, or uncertain for {100 * agreement_dev:.1f}% of {len(dev):,} findings.")
add("")
add("| Rule | Reviewed | Both accept | At least one accepts | Both reject |")
add("| --- | ---: | ---: | ---: | ---: |")
for key, n, yes, any_yes, no in rule_rows(dev, 15):
    add(f"| `{key.split(' ')[0]}`{' ' + key.split(' ', 1)[1] if ' ' in key else ''} | {n:,} | {pct(yes, n)} | {pct(any_yes, n)} | {pct(no, n)} |")
add("")
add("## Decisions")
add("")
add("A phrase is listed when both reviewers accepted at least 80% of at least five reviewed")
add("findings of its word family, and the phrase itself did not fall below 60% on three or")
add("more findings. Inflections of one word form a family. A word counts toward a cluster")
add("when at least one reviewer accepted 85% of its findings and both rejected at most 15%.")
add("A rule that ships enabled becomes opt-in when both reviewers rejected at least 60% of")
add("15 or more findings.")
add("")
by_rule = collections.defaultdict(list)
for phrase, row in assignment.items():
    if row["rule"] != "unassigned":
        by_rule[row["rule"]].append(phrase)
add("| Change | Content |")
add("| --- | --- |")
code = lambda items: ", ".join(f"`{item}`" for item in sorted(items))
add(f"| `hype.vague-praise` ships enabled, list extended | {code(by_rule['hype.vague-praise'])} |")
add(f"| `filler.wordy-phrase` list extended | {code(by_rule['filler.wordy-phrase'])} |")
add(f"| `hype.modifier-cluster` list extended | {code(decisions['cluster'])} |")
for rule, (state, n, accept, reject) in sorted(decisions["rules"].items()):
    add(f"| `{rule}` becomes opt-in | both reviewers rejected {round(reject * n)} of {n} findings and accepted {round(accept * n)} |")
add("")
kept = decisions["kept"]
add(f"On the development sets the listed phrases have {sum(k[2] for k in kept)} reviewed findings, and both reviewers")
add(f"accepted {pct(sum(k[3] for k in kept), sum(k[2] for k in kept))}. That figure is optimistic, because the same findings selected the phrases.")
add("The final test below is the unbiased measurement.")
add("")
add("Candidates that failed include stock connectives and summary openers (both reviewers")
add("accepted about 5%), `harness` (a technical term in every reviewed case), `ensure` and")
add("its forms, and `this allows`. Surface measurements stayed opt-in: both reviewers")
add("rejected most findings of the passive, readability, parenthetical, contrast, and")
add("em-dash measurements.")
add("")
add("## Final test")
add("")
add("The engine ran with the technical profile on the 772 final pages, before and after the change.")
add("")
add("| Measure | Before | After |")
add("| --- | ---: | ---: |")
add(f"| Findings | {old['findings']:,} | {new['findings']:,} |")
add(f"| Findings other than `syntax.long-sentence` | {old['findings'] - old['long']:,} | {new['findings'] - new['long']:,} |")
add(f"| Findings per 1,000 prose words | {1000 * old['findings'] / old['words']:.2f} | {1000 * new['findings'] / new['words']:.2f} |")
add(f"| Items both annotators marked that a finding touches | {old['touched']} of {old['consensus']} ({pct(old['touched'], old['consensus'])}) | {new['touched']} of {new['consensus']} ({pct(new['touched'], new['consensus'])}) |")
add(f"| The same without `syntax.long-sentence` | {old['touched_other']} ({pct(old['touched_other'], old['consensus'])}) | {new['touched_other']} ({pct(new['touched_other'], new['consensus'])}) |")
add("")
add("Both reviewers then judged the findings of the changed engine on the final test.")
add("")
add("| Rule | Reviewed | Both accept | At least one accepts | Both reject |")
add("| --- | ---: | ---: | ---: | ---: |")
for key, n, yes, any_yes, no in rule_rows(final, 5, by_group=False):
    add(f"| `{key.split(' ')[0]}` | {n:,} | {pct(yes, n)} | {pct(any_yes, n)} | {pct(no, n)} |")
add("")
final_rows = {key: (n, yes, any_yes, no) for key, n, yes, any_yes, no in rule_rows(final, 5, by_group=False)}
wn, wy = final_rows["filler.wordy-phrase"][:2]
pn, py = final_rows["hype.vague-praise"][:2]
add(f"`filler.wordy-phrase` meets the 85% target of issue 26: both reviewers accepted {wy} of {wn} findings")
add(f"(95% interval {wilson(wy, wn)[0]:.0f}% to {wilson(wy, wn)[1]:.0f}%). `hype.vague-praise` is just below it: {py} of {pn}")
add(f"(interval {wilson(py, pn)[0]:.0f}% to {wilson(py, pn)[1]:.0f}%), with at least one reviewer accepting {pct(final_rows['hype.vague-praise'][2], pn)} and both rejecting {pct(final_rows['hype.vague-praise'][3], pn)}.")
add("It ships enabled as a warning without a gate, and the point estimate is stated here.")
add("")
add("The same table shows that several rules that already shipped enabled do worse on real")
add("pages than on the development sets: `filler.instruction-scaffolding`,")
add("`filler.document-metadiscourse`, and `filler.unscoped-assurance`. The final test reports;")
add("it does not select. Their defaults are unchanged here and are the next subject of issue 26.")
add("`syntax.long-sentence` produces three of four findings on real pages, and both reviewers")
add("accept about six in ten.")
add("")
findings = [r for r in audit if r["kind"] == "finding" and r["answer"]]
placebo = [r for r in audit if r["kind"] == "placebo" and r["answer"]]
add("### Third-assistant audit")
add("")
add("The auditor compared each passage with a version in which the flagged phrase was")
add("removed, or replaced by a plain word for a stock verb. The order was random. Placebo")
add("items removed a random ordinary word elsewhere. An item is confirmed when the auditor")
add("does not prefer the original and reports no lost information.")
add("")
add("| Group | Audited | Confirmed | No information lost |")
add("| --- | ---: | ---: | ---: |")
for label, rows in (("Findings", findings), ("Placebo", placebo)):
    add(f"| {label} | {len(rows)} | {pct(sum(r['confirmed'] for r in rows), len(rows))} | {pct(sum(not r['answer']['information_lost'] for r in rows), len(rows))} |")
for rule in ("filler.wordy-phrase", "hype.vague-praise"):
    rows = [r for r in findings if r["rule"] == rule]
    add(f"| `{rule}` | {len(rows)} | {pct(sum(r['confirmed'] for r in rows), len(rows))} | {pct(sum(not r['answer']['information_lost'] for r in rows), len(rows))} |")
add("")
add("The audit separates findings from placebo edits, so the auditor is not accepting")
add("everything. It confirms fewer findings than the 85% that ADR 0046 set for a fitted")
add("model. About half of its rejections come from the mechanical edit: removing one")
add("adjective can leave `most language models` or `for , API-key-free`, and the auditor then")
add("prefers the grammatical original while stating that the removed word carried no")
add("information. A later audit should let an assistant write the minimal edit.")
add("")
if (W / "rev-real3").exists() and (W / "docs-real3.jsonl").exists():
    real3 = docs("docs-real3.jsonl")
    report3 = json.loads((W / "engine/report-real3-r2.json").read_text())
    severity = {f["id"]: f["severity"] for f in report3["findings"]}
    by_severity = collections.Counter(f["severity"] for f in report3["findings"])
    reviewed3 = {}
    for name in ("claude", "astra", "qwen"):
        for path in (W / "rev-real3" / name).glob("*.json"):
            for finding in json.loads(path.read_text())["findings"]:
                reviewed3.setdefault(finding["id"], {"rule": finding["rule"]})[name] = finding.get("justified")
    # The second reviewer is gpt-6-astra where it answered and the third assistant elsewhere.
    by_second = collections.Counter()
    overlap = [v for v in reviewed3.values() if v.get("claude") and v.get("astra") and v.get("qwen")]
    for v in reviewed3.values():
        if v.get("astra"):
            by_second["astra"] += bool(v.get("claude"))
        elif v.get("qwen"):
            v["astra"] = v["qwen"]
            by_second["qwen"] += bool(v.get("claude"))
    both3 = {k: v for k, v in reviewed3.items() if v.get("claude") and v.get("astra")}
    add("## Second round: severity by evidence")
    add("")
    add("[ADR 0048](../adr/0048-severity-follows-review.md) sets default severity from the reviews above:")
    add("a warning needs both reviewers to accept at least 85% of 50 or more findings on real pages.")
    add(f"A new set of {real3[0]:,} real pages from {real3[1]:,} repositories, selected by the same rule and used")
    add("for nothing before, reports the result.")
    add("")
    add("| Severity | Findings | Reviewed | Both accept | At least one accepts | Both reject |")
    add("| --- | ---: | ---: | ---: | ---: | ---: |")
    for level in ("warning", "note"):
        rows = [v for k, v in both3.items() if severity.get(k) == level]
        yes = sum(v["claude"] == "yes" and v["astra"] == "yes" for v in rows)
        any_yes = sum(v["claude"] == "yes" or v["astra"] == "yes" for v in rows)
        no = sum(v["claude"] == "no" and v["astra"] == "no" for v in rows)
        low, high = wilson(yes, len(rows))
        add(f"| {level} | {by_severity[level]:,} | {len(rows):,} | {pct(yes, len(rows))} ({low:.0f}% to {high:.0f}%) | {pct(any_yes, len(rows))} | {pct(no, len(rows))} |")
    add("")
    add("| Rule | Severity | Reviewed | Both accept | At least one accepts |")
    add("| --- | --- | ---: | ---: | ---: |")
    table3 = collections.defaultdict(collections.Counter)
    for k, v in both3.items():
        c = table3[(v["rule"], severity.get(k, ""))]
        c["n"] += 1
        c["yes"] += v["claude"] == "yes" and v["astra"] == "yes"
        c["any"] += v["claude"] == "yes" or v["astra"] == "yes"
    for (rule, level), c in sorted(table3.items(), key=lambda kv: (kv[0][1] != "warning", -kv[1]["n"])):
        if c["n"] >= 5:
            add(f"| `{rule}` | {level} | {c['n']:,} | {pct(c['yes'], c['n'])} | {pct(c['any'], c['n'])} |")
    add("")
    pages_first = len(list((W / "rev-real3" / "claude").glob("*.json")))
    pages_both = len(list((W / "rev-real3" / "astra").glob("*.json")))
    add("Findings are capped at eight per rule and page for review, so a page with many long")
    add("sentences contributes eight of them.")
    if overlap and not by_second["qwen"]:
        same = sum(v["qwen"] == v["astra"] for v in overlap)
        third_yes = sum(v["qwen"] == "yes" for v in overlap)
        second_yes = sum(v["astra"] == "yes" for v in overlap)
        add("The third assistant, `private-host/Qwen3.8-27B`, judged the same findings as a check. It gave the")
        add(f"same answer as the second reviewer for {pct(same, len(overlap))} of {len(overlap):,} findings. It accepted")
        add(f"{pct(third_yes, len(overlap))} of them and the second reviewer {pct(second_yes, len(overlap))}.")
    elif pages_both < pages_first and by_second["qwen"]:
        same = sum(v["qwen"] == v["astra"] for v in overlap) if overlap else 0
        add(f"`gpt-6-astra` judged {pages_both} of the {pages_first} pages that have findings; its usage quota ended")
        add(f"there. For the other pages the second reviewer is `private-host/Qwen3.8-27B`: {by_second['qwen']:,} of the")
        add(f"{len(both3):,} findings counted above. On the {len(overlap):,} findings that both of them judged, the two")
        add(f"gave the same answer for {pct(same, len(overlap))}.")
    elif pages_both < pages_first:
        add(f"The second reviewer judged {pages_both} of the {pages_first} pages that have findings; its usage")
        add("quota ended there. The pages it covered are the Go and Rust repositories and the recent")
        add("Python ones. The TypeScript pages and the older Python ones have one review only and are")
        add("not counted above.")
    add("")
if (W / "engine/round3-dev.json").exists() and (W / "engine/round3-evidence.json").exists():
    evidence = json.loads((W / "engine/round3-evidence.json").read_text())
    dev = json.loads((W / "engine/round3-dev.json").read_text())

    def rule_rows(side):
        return {(row["rule"], row["severity"]): row for row in side["rules"]}

    def removed_rule(side, rule):
        rows = [row for row in side["rules"] if row["rule"] == rule]
        return sum(r["reviewed"] for r in rows), sum(r["both_accept"] for r in rows)

    def comparison(result, minimum):
        add("| Rule | Reviewed before | Both accept | Reviewed after | Both accept |")
        add("| --- | ---: | ---: | ---: | ---: |")
        before_rules, after_rules = collections.defaultdict(collections.Counter), collections.defaultdict(collections.Counter)
        for side, target in ((result["old"], before_rules), (result["new"], after_rules)):
            for row in side["rules"]:
                target[row["rule"]]["reviewed"] += row["reviewed"]
                target[row["rule"]]["both_accept"] += row["both_accept"]
        for rule in sorted(before_rules, key=lambda r: -before_rules[r]["reviewed"]):
            before, after = before_rules[rule], after_rules[rule]
            if before["reviewed"] < minimum:
                continue
            shown = pct(after["both_accept"], after["reviewed"]) if after["reviewed"] else "not reported"
            add(f"| `{rule}` | {before['reviewed']:,} | {pct(before['both_accept'], before['reviewed'])} | {after['reviewed']:,} | {shown} |")
        for label, key in (("all notes", "note"), ("all warnings", "warning")):
            before, after = result["old"]["severity"][key], result["new"]["severity"][key]
            add(f"| {label} | {before['reviewed']:,} | {pct(before['both_accept'], before['reviewed'])} | "
                f"{after['reviewed']:,} | {pct(after['both_accept'], after['reviewed'])} |")
        before, after = result["old"]["all"], result["new"]["all"]
        add(f"| all findings | {before['reviewed']:,} | {pct(before['both_accept'], before['reviewed'])} | "
            f"{after['reviewed']:,} | {pct(after['both_accept'], after['reviewed'])} |")
        add("")

    add("## Third round: thresholds, matcher forms, and two rules")
    add("")
    add("[ADR 0049](../adr/0049-defaults-follow-reviewed-forms.md) reads the reviews inside a rule.")
    add("Split by sentence length, the long-sentence findings on real pages show where the rule holds.")
    add("")
    add("| Sentence length in prose words | Reviewed | Both accept |")
    add("| --- | ---: | ---: |")
    for band in evidence["bands"]:
        label = f"{band['low']} to {band['high']}" if band["high"] else f"{band['low']} or more"
        add(f"| {label} | {band['reviewed']:,} | {pct(band['both_accept'], band['reviewed'])} |")
    add("")
    table_rows, other_rows = evidence["rows"]["table row"], evidence["rows"]["other"]
    forms = removed_rule(dev["removed"], "filler.instruction-scaffolding")
    words = removed_rule(dev["removed"], "repetition.adjacent-word")
    meta = removed_rule(dev["removed"], "filler.document-metadiscourse")
    assurance = removed_rule(dev["removed"], "filler.unscoped-assurance")
    add(f"In table rows both reviewers accepted {pct(table_rows['both_accept'], table_rows['reviewed'])} of {table_rows['reviewed']:,} findings,")
    add(f"against {pct(other_rows['both_accept'], other_rows['reviewed'])} of {other_rows['reviewed']:,} elsewhere.")
    add("In `filler.instruction-scaffolding`, the conditional reader goal and the `Be sure to` reminder gave")
    add(f"{forms[0]:,} reviewed findings, and both reviewers accepted {pct(forms[1], forms[0])} of them.")
    add(f"In `repetition.adjacent-word`, both reviewers accepted none of the {words[0]:,} findings whose two")
    add("copies differ in case, repeat a capitalized word inside a sentence, or stand on two sides of Markdown markup.")
    add(f"They accepted {pct(meta[1], meta[0])} of {meta[0]:,} `filler.document-metadiscourse` findings and")
    add(f"{pct(assurance[1], assurance[0])} of {assurance[0]:,} `filler.unscoped-assurance` findings. Both rules are now opt-in.")
    add("")
    add(f"The next table compares the earlier defaults with the changed ones on the {dev['documents']:,} real pages")
    add("reviewed before this round. A finding keeps its verdict when its page, rule, and byte span are unchanged.")
    add("")
    comparison(dev, 10)
    add(f"The default profile reports {dev['new']['findings']:,} findings on these pages in place of {dev['old']['findings']:,}.")
    add("These pages chose the changes, so the right columns overstate the effect.")
    add("")
    if (W / "engine/round3-test.json").exists() and (W / "docs-real4.jsonl").exists():
        test = json.loads((W / "engine/round3-test.json").read_text())
        real4 = docs("docs-real4.jsonl")
        add(f"A third set of {real4[0]:,} new real pages from {real4[1]:,} repositories, selected by the same rule and used")
        add("for nothing before, reports the result. Both reviewers judged the findings of the earlier")
        add("defaults and of the changed ones on the same pages.")
        add("")
        comparison(test, 5)
        after = test["new"]
        add("| Severity, changed defaults | Findings | Reviewed | Both accept | At least one accepts | Both reject |")
        add("| --- | ---: | ---: | ---: | ---: | ---: |")
        for level in ("warning", "note"):
            c = after["severity"][level]
            low, high = wilson(c["both_accept"], c["reviewed"])
            add(f"| {level} | {after['reported'][level]:,} | {c['reviewed']:,} | {pct(c['both_accept'], c['reviewed'])} ({low:.0f}% to {high:.0f}%) | "
                f"{pct(c['any_accept'], c['reviewed'])} | {pct(c['both_reject'], c['reviewed'])} |")
        add("")
        gone = test["removed"]["all"]
        add(f"The changed defaults report {after['findings']:,} findings on these pages in place of {test['old']['findings']:,},")
        add(f"or {1000 * after['findings'] / real4[3]:.2f} per thousand words in place of {1000 * test['old']['findings'] / real4[3]:.2f}.")
        add(f"Both reviewers accepted {pct(gone['both_accept'], gone['reviewed'])} of the {gone['reviewed']:,} dropped findings they reviewed.")
        add("The gain is smaller than on the pages that chose the changes, as expected. `syntax.long-sentence`")
        add("holds two findings in three above 45 words, and one in two above 70, where the sentence is often a list.")
        add("")
        third = {}
        if (W / "rev-real4" / "qwen").exists():
            second = {}
            for review_path in (W / "rev-real4" / "astra").glob("*.json"):
                for finding in json.loads(review_path.read_text())["findings"]:
                    if finding.get("justified"):
                        second[finding["id"]] = finding["justified"]
            for review_path in (W / "rev-real4" / "qwen").glob("*.json"):
                for finding in json.loads(review_path.read_text())["findings"]:
                    if finding.get("justified") and finding["id"] in second:
                        third[finding["id"]] = finding["justified"]
        if third:
            same = sum(third[k] == second[k] for k in third)
            third_yes = sum(v == "yes" for v in third.values())
            second_yes = sum(second[k] == "yes" for k in third)
            add("The third assistant, `private-host/Qwen3.8-27B`, judged findings of this round as a check. It gave the")
            add(f"same answer as the second reviewer for {pct(same, len(third))} of {len(third):,} findings. It accepted")
            add(f"{pct(third_yes, len(third))} of them and the second reviewer {pct(second_yes, len(third))}.")
            add("")
add("## Limits")
add("")
add("- Reviewers and auditor are assistants. Models can share a preference that readers do not have.")
add("- The phrase lists cover prose in paragraphs, comments, and strings. List items, headings, and table cells are not checked by the praise rule.")
add("- Coverage of the annotators' agreement stays low. Most items they mark are contextual: a claim without scope, a sentence that repeats an earlier one, a frame that adds nothing. Fixed rules name a small part of that. The MCP review tools let the calling assistant judge the rest.")
add("- The corpus tool left out 17 of 2,713 real pages that it could not parse; issue 522 tracks them. The engine now analyzes such a page and excludes the rejected block.")
add("- The third round matches findings between defaults by page, rule, and byte span. Review stays capped at eight findings per rule and page.")
OUT.parent.mkdir(parents=True, exist_ok=True)
OUT.write_text("\n".join(lines) + "\n")
print("written", OUT, len(lines), "lines; final reviews by both:", len(final))
