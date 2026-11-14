# Findings against what people edited

November 12, 2026. Every earlier number in this repository compares a model
with a model: two assistant annotators, two assistant reviewers, an assistant
auditor. This record measures the tool against what people did. It uses three
public sources of human edits and one public study of human preferences, and
it keeps coordinates and choices only. No text of any source is stored here.

[RESULTS.md](RESULTS.md), [RESULTS-words.md](RESULTS-words.md), and
[RESULTS-index.md](RESULTS-index.md) hold every
number. They are generated from `results.json`, and `verify.py` recounts
`results.json` from the data files.

## Sources

| Source | What people did | Genre | Use here |
| --- | --- | --- | --- |
| Documentation pull requests | Maintainers reworded pages of their own project | Technical documentation, written by people | Does a finding fall on a sentence that a maintainer changed? |
| [Beemo](https://huggingface.co/datasets/toloka/beemo) (MIT) | Experts edited answers that models gave to instructions | Explanations, summaries, answers | The same question, on model-written text |
| [LAMP](https://github.com/salesforce/creativity_eval/tree/main/Writing_Alignment) at `3d02987` (BSD-3-Clause) | Professional writers edited paragraphs that models wrote; three writers edited the same 54 paragraphs; writers ranked three versions of 201 paragraphs | Fiction and creative non-fiction | Agreement between editors; whether a judge model shares the writers' preferences |

The pull requests come from GitHub search: merged before November 2022, with
an editorial word in the title (`copy edit`, `copyedit`, `reword`, `rephrase`,
`wording`, `concise`, `wordy`, `verbose`, `redundant`, `tighten`,
`readability`, `simplify docs`). The searches returned 3,392 pull requests. The
first 1,200 by hash were examined. 238 change at most three files, all
Markdown, and rewrite 1 to 40 lines of a file of at most 40,000 bytes; 184
remain with one per repository; 121 of those change at least one prose
sentence as the engine extracts it. Repositories are split by hash into 60 for
development and 61 for the test.

The study that is closest to this problem,
[Measuring AI "Slop" in Text](https://arxiv.org/abs/2509.19163), has span
annotations by professional editors. Its data was not released when this
record was made.

## Method

The plan, the measures, and the decision rules were written on
[issue 525](https://github.com/stokaro/unswell/issues/525) before the
corresponding results were read. Two amendments were posted there before any
result they affect.

A text is split into sentences: by one fixed splitter for LAMP and Beemo, and
by the engine's own prose sentences for documentation, so that code and tables
are outside. A sentence is *edited* when a person changed any part of it and
*flagged* when a finding touches it. TPR is the share of edited sentences that
are flagged, FPR the share of unedited sentences that are flagged, and J is
TPR minus FPR. A system that flags at random has J = 0 whatever its rate.

Systems: `unswell check` with the default rules and with every rule,
`unswell review` as merged in #544 with two reviewer models, alone and with
the other model confirming, and the span-detection prompt that the LAMP
authors published, with the same two models.

For preferences, each judge model ranked the three versions of every LAMP
triplet with labels hidden, and chose between the raw and the expert-edited
answer of 300 Beemo test rows. A model was to be admitted as a judge of
writing only if it agreed with the writers' majority.

`unswell trim` changes text, so it is compared with the editor word by word.
The command ran on every development and test document with two models. A
removed word is a hit when the editor removed or replaced that word. Precision
is compared with two chance values: random removal of the same number of
words from the same sentences, among the words the check allows to remove, and
from the same documents. For Beemo and LAMP, where the editor's text exists,
the net movement is the change of the word-level edit distance to that text
per removed word, compared with the same random removal, 20 draws per
document. Intervals come from 2,000 bootstrap draws over documents. The rule,
two amendments, and the change of the parser that the development parts led
to were posted on issue 525 before any test document ran.

## What the results say

1. **Editors do not agree on which places to change.** One writer's edited
   sentences predict another's barely above chance. At the word level Cohen's
   kappa between writers is 0.11, and 11.9% of words are marked by all three
   (`ceiling_words.py`). The LAMP paper reports the same data as a span
   precision of 0.57, which does not correct for writers marking a third to a
   half of the text. "Finds what an editor would mark" is therefore not a
   stable target, for a tool or for a second editor.

2. **Editors do agree on which version is better.** Two writers give the same
   answer for four pairs in five. They prefer text that a model pipeline
   edited to the raw model text in 71.6% of judgments. That is the published
   human evidence that finding and rewriting helps, in fiction.

3. **Judge models do not share the writers' preferences.** On the pair
   "writer-edited or raw model text" the three judge models agree with the
   writers' majority about half the time, and they prefer the raw model text
   to the writer's edit in about two thirds of the triplets. Their reasons
   show why: they penalize the typos of a hand edit and reward smooth stock
   phrasing. No model was admitted as a judge. This matches published work on
   model evaluators favoring model text
   ([Laurito et al.](https://arxiv.org/abs/2407.12856)).

4. **The fixed rules do not find what people edit.** On documentation that
   maintainers reworded, the default rules flag edited and unedited sentences
   at the same rate, and enabling every rule does not help.

5. **A weighted consensus of models prefers trimmed paragraphs.** Paragraphs
   of generated documentation were compared before and after a model's
   revision. A first comparison had one reader; the maintainer withdrew those
   answers as a measure and accepted a consensus of models in their place, in
   which a stronger model weighs more. Four models answered by a rule about
   lost information, not about taste. Where the revision could only delete
   words, the consensus prefers the trimmed paragraph in 16 of 20 decisive
   answers after review findings and in 17 of 20 after trimming random
   sentences. Where the model rewrote freely, it prefers the revision in
   15 of 20 after review findings and in 6 of 9 for random sentences.

6. **The consensus hides a split between model families.** On the trim-only
   pairs after review findings the single models prefer the trimmed paragraph
   in 16, 18, 9, 6 of 20 pairs: the two models of one family approve, and the
   other two find that about half of the trims lose information. The outcome
   rests on counting both models of that family: with one equal vote per
   family the rule is not met. `RESULTS.md` lists every
   count, and `data/panel.jsonl.gz` every answer.

7. **Review by a model does a little.** It flags edited sentences more often
   than unedited ones in all three genres, by a margin that is small on
   documentation and fiction and larger on expert-edited answers. Confirmation
   by a second model shortens the list. J stays about the same on
   documentation and falls or stays flat on the other two.

8. **Trimming removes words that experts also changed, on model answers.**
   On Beemo's test part, 71.2% and 68.0% of the words that
   `unswell trim` removes with the two models are words the expert removed or
   replaced, against 62.7% and 59.3% for random words of the same
   sentences. Removal as such moves these answers toward the expert's text,
   because the experts removed or replaced more than half of the words: under
   random removal from the same sentences the edit distance falls by 0.15 and
   0.09 words per removed word. Under the command it falls by 0.28 and 0.23,
   so the command adds 0.13 and 0.15. This is the
   first result in this record in which a part of the tool changes text in the
   direction that people changed it.

9. **On fiction and on documentation written by people, that is not shown.**
   On LAMP the differences over chance are +2.6 and +3.1 points and the
   intervals reach zero; the distance to the writer's text does not fall. On
   documentation pull requests they are +2.0 and +2.2 points, on pages
   where a maintainer changed about 4.0% of the words. The development
   parts had shown larger differences on fiction than the test part did.

10. **At the level of words, review beats chance and the fixed rules do not.**
    The words under the quotes of `unswell review` are words an editor changed
    more often than random words of the same documents, with both reviewer
    models on all three test parts. The words under the findings of the
    default rules are not, on any of them. No rule was registered for this
    comparison.

11. **One opener is a phrase that experts always removed.** 27 of 150
    development answers open with `Sure,` or `Sure!`, and the experts removed
    or replaced that word in every one. The rule for chat openers did not list
    it. With the two phrases listed, the rule reports 39 of the 300 test
    answers, and the expert changed the flagged word in 39 of 39. The phrases
    report nothing on the documentation pages and the fiction paragraphs.
    The rule also looks at the first sentence after the leading headings of a
    document, because a pasted answer usually sits under a title; with that
    position it reports nothing on 3,302 pages of public repositories and
    455 generated pages outside this record.

12. **Rule by rule, little else stands out.** `results.json` has every rule
    against the words editors changed on the test parts, and
    `RESULTS-words.md` lists those with findings in at least ten documents of
    a set. A wide list of words that models favor, enabled in the run
    with every rule, falls on words the experts changed more often than
    chance on the answers, and not on the other two sets. No rule was
    registered for that table.

13. **A wide list of phrases that chat models favor beats chance and falls
    short of a rule.** The list of result 12 ran on expert-edited answers that
    no other result had used: 850 to prune it and 883 to confirm it.
    10 of its 160 phrases left on the first part. On the second, the words
    under its findings are words the expert changed in 69.2% of cases,
    against 56.3% for random words of the same answers
    (+12.9 points, +8.7 to +17.4). The expert removed or replaced the
    whole flagged phrase in 227 of 343 (66.2%) findings, below the
    70% that was registered. The figure on the part that had been read before
    was 75.8%; the fresh answers put it at 69.2%.

14. **On documentation that people de-slop, the default rules follow what
    they remove, and emoji and em dashes follow it best.** Commits whose
    message says that they remove AI-sounding wording give 1,556 pairs of a
    Markdown file before and after. On the 597 test pairs the findings of the
    alpha.9 defaults are lower after the commit in 64.0% (56.1% to 71.2%)
    of the pairs that differ. A new rule for emoji in headings and list items
    gives 77.7% (69.8% to 84.0%), a new rule for documents with many em
    dashes 70.4% (64.1% to 76.0%), and em-dash density in one paragraph
    62.5% (51.0% to 72.8%). Passive density, noun stacks, parenthetical
    load, and paired contrasts sit near 50% on the development part. Who typed
    these commits is unknown, and the 78 pairs from GitHub's web editor are too
    few to confirm either rule. 97 of the 597 test pairs are not English
    prose; on the 500 English ones both rules still meet the bar and the
    defaults fire in 300 pairs against 141.

15. **A page index of six surface marks separates pages that someone later
    cleaned from pages that people wrote, and stays low on the latter.** The
    index counts em dashes, emoji in headings and at the start of list items,
    arrows, stock words, and bold labels in lists per 1,000 prose words. On
    189 held-out pairs and 566 README files as they were before November
    2022 its AUC is 0.867. It is lower after the commit in
    82 of 105, 78.1% (69.3% to 84.9%) of the pairs that differ, and
    in 31 of 44, 70.5% (55.8% to 81.8%) on the web-flow edits among
    them. 18 of the README files (3.2%) reach the level `high`. The index
    does not change in 84 of the 189 pairs, and 43.4% of the pages are still
    `high` after the commit.

16. **Removing the marks makes a page read less AI-sounding to model judges
    than the original and than a page fixed with a third-party linter's
    findings. That a model guided by unswell's report beats one guided by
    such a linter is not replicated.** On 83 held-out pages `unswell fix`, which uses no model,
    lowers the index on 83, and two blind model judges find its page the
    less AI-sounding one against the original on
    72 of 73, 98.6% (92.6% to 99.8%) of the decisive pages.
    A model that follows the report of `unswell check` gives
    64 of 68, 94.1% (85.8% to 97.7%) against the original. The
    same model with the findings of write-good, the best of three third-party
    linters by this measure, gives 45 of 71, 63.4% (51.8% to 73.6%),
    and against its page the page fixed with unswell's report is the less
    AI-sounding one on 48 of 70, 68.6% (57.0% to 78.2%).
    With Vale's two styles the same model makes pages that the judges find
    more AI-sounding than the original. The judges were admitted on 80
    held-out pairs before and after a de-slop edit; on the 38 of them that are
    web-flow edits, which a person most likely typed, they name the version
    before the edit in 74.2%, 78.4%, 89.7% of their decisive answers.
    A judge of another model family,
    added afterwards, never prefers the original or the linter's page where
    it decides, and calls the two versions the same on 40 of the 83 pages
    for `unswell fix` and on 31 for the report version. With the marks
    removed from both versions of a pair, all three judges still name the
    version before a person's edit in 75.3%, 72.7%, 77.3% of their decisive
    answers, so they hear more than the marks. `unswell fix` changed after the
    registered run (a period between two full clauses in place of a comma);
    its numbers here are a second reading of the same pages, and the first
    was 72 of 73 against the original.

17. **The result of 16 holds for the command on new pages, and fails there
    for the report.** A second run, on 89 pages that no run had read and with
    the published `v0.1.0-alpha.10` binary, repeats every bar against the
    original: the index lower after `unswell fix` on 85 pages, its page
    preferred on 67 of 68, 98.5% (92.1% to 99.7%). The page fixed
    by a model with unswell's report against the page fixed with write-good's
    findings gives 39 of 72, 54.2% (42.7% to 65.2%), which does
    not meet its bar. A third run, on 250 more new pages, gave the command the
    bar that it had lacked: against write-good's page its page is the less
    AI-sounding one on 146 of 209, 69.9% (63.3% to 75.7%), and
    against the original on 231 of 233, 99.1% (96.9% to 99.8%); the
    index is lower on 247 pages.

18. **A report that names the replacement at each mark brings the fix by the
    report back over its bar.** The report of the first two runs gave the
    place of a mark and general advice; the fixer chose its own punctuation.
    With a report that names the replacement `unswell fix` would make, a
    fourth run on 131 pages that nothing had fixed or judged gives, for the
    page fixed by a model with the report against write-good's page,
    73 of 108, 67.6% (58.3% to 75.7%), and against the original
    111 of 113, 98.2% (93.8% to 99.5%); the index is lower on
    131 pages. On 61 of the pages the model's page is the same as the
    command's.

## Decisions

Each rule was registered on issue 525 before its result was read.

1. *Report the test numbers against people next to the numbers against model
   annotation.* Done in `docs/review.md`, the README, and ADR 0052.
2. *Revise the rubric if the published span prompt beats `unswell review` by 5
   points of J with both reviewer models on both development sets.* The gaps
   are +3.5, -0.4, +14.7, +3.6 points (fiction and answers for each model). Not all four reach 5 points, so the rubric stays.
   The published prompt flags most sentences of a text, between
   56.8% and 82.9%.
3. *Read a model-free linter's rules as candidates if it beats the default
   rules on both development sets.* Not decided: the linters were not run.
4. *State that no available method finds what editors change if no system
   reaches half of the human ceiling on the test.* Half of the ceiling is
   3.5 points. A system reaches it on the fiction test, so the rule does not apply as written.
   The ceiling itself is the finding: at +7.1 points, editors do not find what
   another editor changes either, and ADR 0052 says so in place of this rule.
5. *Admit a model as a judge of writing only if it agrees with the writers.*
   None was admitted: agreement with the writers' majority on "writer-edited
   or raw" is 50.7%, 49.2%, 51.7%, against a floor of 70%. The comparison
   of revisions by model judges was therefore not run. A blind comparison for
   readers was registered in its place.
6. *On documentation, present review as the way to find what maintainers
   tighten if it reaches a J of 10 points on the test and the default rules do
   not.* The review reaches +7.2 and +4.7 with the two reviewer models, the
   default rules -0.3. Neither reaches 10 points, so the documentation says that neither finds what maintainers tighten.

7. *Add a revision step if readers prefer the revised paragraph.* One reader
   answered the first comparison. The maintainer withdrew those answers as a
   measure, so this rule was not applied to them.
8. *Let a model panel stand in if it reproduces the reader.* Set aside with
   the reader's answers: agreement with withdrawn answers shows nothing.
9. *Maintainer's decision: a weighted consensus of four models judges both
   comparisons, by the 70% rule.* Trim-only meets it in both arms
   (16 of 20 and 17 of 20), and trimming after review findings is not ahead of
   random trimming. A trim-only fix enters the product, not tied to review
   findings. Free revision meets the rule after review findings only
   (15 of 20); it is admitted and not built.
10. *Ship `unswell trim` as built if the consensus prefers its trims in at
    least 70% of decisive answers on documents outside both comparisons.* It
    prefers them in 33 of 40; single models: 31, 34, 21, 16 of 40. The command
    ships as built.

11. *State that trimming agrees with editors above chance on a kind of text
    only if, with both models, precision exceeds both chance values and the
    net movement is below that of random removal, each with an interval that
    excludes zero; state that it moves text toward the editor's version only
    if the net movement itself is below zero.* Met on expert-edited answers,
    both parts of the rule. Not met on fiction and on documentation pull
    requests. [ADR 0054](../../../docs/adr/0054-trim-measured-against-editors.md)
    records what the documentation states.

12. *List `Sure,` and `Sure!` in `scaffold.chat-preamble` if, on the test
    answers, the expert removed or replaced the flagged opener in at least 85%
    of the findings, with the lower end of the interval above 70%, and the
    phrases report at most one finding on the other two sets.* 39 of 39
    (91.0% to 100.0%), and no finding elsewhere. The phrases are listed;
    [ADR 0055](../../../docs/adr/0055-phrase-listed-on-editor-reference.md)
    makes this a route into a default list.

13. *Ship the pruned list as a rule that is off by default if, on the
    confirmation part, the words under its findings are changed by the expert
    more often than random words of the same answers, with an interval above
    zero, and the expert changed every flagged word in at least 70% of the
    findings, with the lower end of the interval above 60%.* The first holds.
    The second does not: 227 of 343 (66.2%). The list does not ship.

14. *Turn a rule on by default if, on the test pairs, its findings are lower
    after the commit in at least 65% of the pairs that differ, with the lower
    end of the interval above 55%.* `format.decorative-emoji` meets it and is
    on. `format.em-dash-frequency` meets it, flags 2 of the 121 pages that
    people wrote before November 2022 against a limit of 5%, and is on.
    `format.em-dash-density` does not and stays off outside the strict
    profile. [ADR 0056](../../../docs/adr/0056-rules-tested-on-de-slop-commits.md)
    records the reference and the bar.

15. *Ship the page index if, on held-out pairs, its AUC against README files
    written before November 2022 is at least 0.80, it is lower after the
    commit in at least 70% of the pairs that differ, the same holds on the
    web-flow edits alone, and at most 5% of the README files are `high`.*
    All four hold, the third by one pair. The index ships in `unswell check`.

16. *Say that fixing what unswell points at lowers AI-sounding if the index
    falls on at least 90% of the pages after `unswell fix` and after the fix
    by unswell's report, each version is the less AI-sounding one against the
    original on more than half of the decisive pages with the lower end of
    the interval above 50%, and the report version meets the same bar against
    the version of the best third-party linter.* All hold. `unswell fix` ships.
    [ADR 0059](../../../docs/adr/0059-page-index-and-fix.md) records both
    decisions.

17. *On new pages, keep each claim of 16 only if its bar holds again; give
    `unswell fix` the bar against the best third-party linter that it lacked.*
    The bars against the original hold again. The bar of the report version
    against write-good's version does not, and that claim is withdrawn. The
    command meets its new bar on the third sample, so the claim against a
    third-party linter is made for `unswell fix` only.

18. *Make the claim for a model that follows the report again only if, with
    the report that names the replacement, its bar against write-good's page
    holds on pages that nothing has read.* It holds on the fourth sample. The
    claim is made for the report as it is now, and the failure of the
    earlier report stays in result 17.

## Limits

- An unedited sentence is not a clean sentence. A person changed what they
  came to change. FPR therefore overstates false alarms, and J is a lower
  bound on how well a system separates wording that needs work.
- The pull requests are purpose-enriched by title, and contributors are
  maintainers, not editors. Their source text was written by people, so this
  set says nothing about wording typical of models.
- LAMP is fiction, and Beemo's answers come from older and smaller models than
  those that write documentation today.
- Agreement between editors was measured on fiction only. No set gives it for
  documentation.
- Open-source linters for model-style wording were registered as a comparison
  for results 1 to 13 and were not run there: installing and executing
  third-party packages was not permitted in the session that made that part
  of the record. For result 16 Vale and write-good ran over the pages in a
  container without network. One of them publishes an
  audit of its own in which its phrase sets fired more often on text written
  by people than on model text.
- Each configuration ran once. A model's answers differ between runs.
- An editor who kept a word did not judge it necessary, and the editors of
  this record agree with each other on words only slightly. A word that one
  editor kept may be one that another would drop, so precision against one
  editor understates agreement with editors as a group. The chance values
  correct for the base rate, not for this.
- Beemo's experts also corrected facts and restructured answers, and they
  changed more than half of the words. Being closer to their text by edit
  distance is not the same as reading better; no reader judged the trimmed
  text.
- Every expert-edited answer of Beemo has now been read by some result of
  this record. A later phrase list, or a change to this one, needs other
  texts.
- The test parts were read for the review before, and now for trimming. A
  later change to the command can be checked on them for a loss, not credited
  with a gain.
- The two comparisons and the check of the command were judged by models.
  Result 3 applies to them: these model families did not share professional
  writers' preferences on fiction. The judging rule here asks about lost
  information and not about taste, which is a narrower question, and it was
  not validated against people.
- Two of the four judges belong to one model family and hold four of the seven
  votes together. The maintainer set the weights after the answers were read;
  the weights registered first give the same consensus on every pair.
- In the free-revision comparison a model rewrote the flagged sentences and
  deleted some of them. The trim-only comparison and the command never delete
  a sentence.

- The fixer and the two judges of the panel of result 16 are models of one
  family, and the fixer is one of them. The judges often give as their reason
  the marks that the index counts. A judge of another family read part of the
  comparisons; no person took part, and nobody judged the grammar of the
  fixed text.
- The version of `unswell fix` that ships was judged in result 16 on pages
  that had been read once with its earlier version. Result 17 reads it on
  new pages.
- The pages of results 16 and 17 have an index of 40 or more. The command
  leaves pages below that alone unless it is told otherwise, and nothing here
  measures it on them.
- The pages of result 16 are the held-out pages with an index of 40 or more.
  On the others the report lists no mark and the result says nothing.

## Files

| Path | Content |
| --- | --- |
| `data/sentences.jsonl.gz` | Per document: sentence spans, the spans a person changed, each system's finding spans, and a digest of the text |
| `data/preferences.jsonl.gz` | Per LAMP triplet: the three writers' orders and each judge's order |
| `data/expert-pairs.jsonl.gz` | Per Beemo pair: each judge's choice and the position it was shown in |
| `data/readers.jsonl.gz` | The one reader's answers to the first comparison, withdrawn as a measure and kept as recorded |
| `data/panel.jsonl.gz` | Per pair of the two comparisons and of the check of the command: the arm, the model that revised, and each judge model's answer under the information rule |
| `data/trims.jsonl.gz` | Per document and trimming model: each accepted trim with the byte spans of the words of its sentence, the words the check never removes, the words the trim removed, and the edit distances to the editor's text before, after, and under random removal |
| `data/word-counts.jsonl.gz` | Per document: its words, the words an editor changed, and both counts under the findings of the default rules and of the review |
| `data/openers.jsonl.gz` | Every finding of the listed chat opener, with the byte spans of its words |
| `data/opener-load.jsonl.gz` | The number of texts and of findings of the chat opener rule outside this record: documentation pages, and source files of open-source packages |
| `data/vocabulary.jsonl.gz`, `data/vocabulary-phrases.jsonl.gz`, `data/vocabulary-load.jsonl.gz` | The wide phrase list, the phrases that the selection part dropped, per unused answer the findings of the list with the index of the phrase and the byte spans of the flagged words, and the count of findings of the kept phrases on pages of public repositories |
| `data/deslop-pairs.jsonl.gz`, `data/emoji-load.jsonl.gz`, `data/frequency-load.jsonl.gz` | Per file before and after a commit that removes AI-sounding wording: repository, commit, path, words, and findings per rule in both versions; and the reports of the two new rules on four kinds of documentation |
| `data/default-load.jsonl.gz` | Per page that people wrote before November 2022: its words and the rules that report on it |
| `data/page-index.jsonl.gz` | Per page behind the reading of the page index (README files as they were before November 2022, and both versions of de-slop pairs outside the main set and in its test part): repository, commit, path, prose words, the count of each signal, and the index that the product reported |
| `data/fix-panel.jsonl.gz` | Per page of the fix experiment: repository, commit, path, the index of the original, for each fixed version its index and the share of tokens changed, and each judge's answer per pair of versions. No page text and no model output |
| `data/fix-fresh.jsonl.gz` | Per page of the run of the released binary on new pages: repository, commit, path, the index of the original, for each fixed version its index and the share of tokens changed, and each judge's answer per pair of versions. No page text and no model output |
| `data/fix-third.jsonl.gz` | Per page of the third sample: the same fields as `data/fix-fresh.jsonl.gz`, for the page after `unswell fix` and the page fixed with write-good's findings |
| `data/fix-fourth.jsonl.gz` | Per page of the fourth sample: the same fields as `data/fix-fresh.jsonl.gz`, for the page fixed by a model with the report that names the replacement, the page fixed with write-good's findings, and the page after `unswell fix` |
| `data/judge-pairs.jsonl.gz` | Per held-out pair before and after a de-slop edit by people: repository, commit, path, and each judge's answer for the pair as it is, for the pair with `unswell fix` applied to both versions, and for the fixed earlier version against the person's version |
| `data/readme-load.jsonl.gz` | Per README file of the page index: its words and the rules that report on it with the configuration that ships |
| `data/manual-pairs.jsonl.gz` | Per likely-manual de-slop edit (a short one-file commit through GitHub's web flow): repository, commit, path, words, and findings per rule in both versions |
| `data/review-repeat.jsonl.gz` | Per file before a de-slop commit and per answer form of the review: findings, flagged words, flagged words that the commit changed, and flagged sentences in each of two runs; the sentences flagged in both; and what the commit did to the flagged sentences |
| `data/rule-words.jsonl.gz` | Per document and rule of the run with every rule: the words under the rule's findings and how many the editor changed |
| `results.json`, `RESULTS.md`, `RESULTS-words.md`, `RESULTS-index.md` | Every published number, recounted and rendered; the second page holds the comparisons at the level of words, the third the page index, the fix experiment, and its judges |
| `verify.py`, `test_verify.py`, `render.py` | The check, its negative tests, and the renderer |
| `scripts/` | The scripts that fetched, prepared, ran, and measured, as they ran. `scripts/index/` holds those of the page index and of the fix experiment; `fix_pairs.py` there takes the path of the binary as an argument, where the runs had it as a constant. The judge `gpt-6-astra` ran through Codex with the settings of `annotate.py` in the assistant-panel corpus and the reasoning effort set to low |

```sh
make check-human-reference
```

The check runs offline. To rebuild the data, the scripts need the two public
datasets, the GitHub command-line client, and access to the models.
