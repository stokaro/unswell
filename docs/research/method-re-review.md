# Published methods and open-source projects, read again

This page answers [issue 523](https://github.com/stokaro/unswell/issues/523). It reads the earlier method reviews of this repository again, now that the screen of 123 events is retired, and adds the sources that the issue names. One row per method gives five answers, and two lists at the end say what is worth a trial.

## How to read the table

The columns are the five questions of the issue.

| Column | Question |
| --- | --- |
| Claim | What the method claims, on which data, with which metric. |
| Earlier test here | What this repository ran, on how many documents, and why it stopped. |
| Does the conclusion follow? | Whether the test was too small, or tied to the retired screen. |
| Can it be tried now? | On the [assistant-panel corpus](../../research/corpus/2026-11-11-assistant-panel/README.md) or on a later reference. |
| What it would add | A rule, a feature for the reviewer tools, or an evaluation method. |

Every annotator, reviewer, and judge in these records is a language model ([ADR 0041](../adr/0041-assistant-review-acceptance.md)). An "event" is one of the 123 defects of the retired screen. "Not recorded" means that no file of the repository states the value. "Own" marks a method of this project.

## Statistical detection of origin

An origin score names no wording to change, and [ADR 0037](../adr/0037-diagnostics-not-authorship.md) keeps authorship out of the product, so the last column is short.

| Method | Claim | Earlier test here | Does the conclusion follow? | Can it be tried now? | What it would add |
| --- | --- | --- | --- | --- | --- |
| Lexical and stylometric baselines, [StyloAI](https://arxiv.org/abs/2405.10129v1) (issues 50, 176) | 31 stylometric features, random forest. Data, metric: not recorded ([registry](../../research/methods/registry-v1.json)). | 3,398 paragraphs: no arm separated old snapshots from new ([pilot](../../research/baselines/runs/2026-09-11-pilot/README.md)). | Only for snapshot date. | Done for word n-grams: no recall on 164 real pages at 85% precision ([ADR 0047](../adr/0047-qualified-defaults-and-assistant-review.md)). | Nothing. |
| [LLMDet](https://aclanthology.org/2023.findings-emnlp.139/) (51, 177) | Proxy perplexity from stored n-gram tables. Data, metric: not recorded. | Six of eleven tables, same paragraphs: 56.2% scored, no separation ([run](../../research/baselines/runs/2026-09-11-llmdet/README.md)). | Partly: five models did not run. The tables take 3.7 GB. | Yes, on 455 generated documents. | Nothing. |
| Compression, [ZipPy](https://github.com/thinkst/zippy) (52, 178) | Compression gain against reference text. Metric: not recorded. | One zlib bank, same paragraphs: Brier difference -0.0039 (-0.0110 to +0.0035) ([run](../../research/baselines/runs/2026-09-11-compression/README.md)). | Partly: one bank, one level. | Yes, under [ADR 0033](../adr/0033-compression-reference-bank.md). | Nothing. |
| [Fast-DetectGPT](https://proceedings.iclr.cc/paper_files/paper/2024/hash/6b8c6f846c3575e1d1ad496abea28826-Abstract-Conference.html), [Binoculars](https://proceedings.mlr.press/v235/hans24a.html), [DivEye](https://arxiv.org/abs/2509.18880v3) (53) | Scores from a language model's token probabilities. Data, metric: not recorded. | None: not run, for model cost and license ([methods](../../research/methods/README.md)). | No test, no conclusion. | Outside the engine only. | Nothing. |
| Origin channel (58, 179) | Own: word n-grams separate generated documentation from human code comments ([ADR 0035](../adr/0035-origin-channel.md)). | Held-out recall 0.875 at a false-flag rate of 0.052; polished paragraphs 0.342 ([run](../../research/origin/runs/2026-09-14-confirmation/README.md)). | Yes, for that contrast of formats. | On the generated set only. | Nothing. |
| Richness and entropy, [Shan et al.](https://aclanthology.org/2026.emnlp-main.1725/) (GEN, ARB, TextSight, AITDNA) | Both separate generated from human text in five domains. Metric: not recorded ([review](../../research/methods/non-llm-text-methods-2026.md)). | AUROC 0.880 on GEN ([run](../../research/origin/runs/2026-11-08-gen-edit-transfer/README.md)). That threshold flagged 6 of 228 generated and 0 of 228 polished technical documents ([transfer](../../research/origin/runs/2026-11-08-gen-technical-transfer/README.md), [others](../../research/origin/README.md)). | Yes; frozen gates. | Yes, at no model cost. | Nothing. |
| [EditLens](https://arxiv.org/abs/2510.03154), [MELD](https://arxiv.org/abs/2605.06903) | Encoders trained on edited pairs. MELD: 99.7% detected at 1% false positives on its own pool (same review). EditLens: not recorded. | Of 228 polished documents EditLens flagged 4, MELD 31. By generator, MELD flagged 67, 8, and 72 of 76 generated ones ([EditLens](../../research/origin/runs/2026-11-10-editlens-new-controls/RESULT.md), [MELD](../../research/origin/runs/2026-11-11-meld-feasibility/RESULT.md)). | Yes; frozen thresholds. | Yes. The weights take 1.42 GB ([record](../../research/origin/runs/2026-11-10-editlens-five-window/RESULT.md)) and 4.1 GB. | Nothing. |
| Span detectors: LLMTrace, GigaCheck | They mark spans a model wrote (issue comment). Data, metric, URL: not recorded. | None. Richness on known replaced spans detected 5.65% at 1% false positives ([run](../../research/origin/runs/2026-11-08-heterogeneous-spans-transfer/README.md)). | No conclusion exists. | On cleanup-commit pairs. | A reviewer-tool feature at most. |

## Contextual rules, learned selection, and pipelines with a model

| Method | Claim | Earlier test here | Does the conclusion follow? | Can it be tried now? | What it would add |
| --- | --- | --- | --- | --- | --- |
| Contextual diagnostics on longer prose (218) | Own: windowed rules for repeated constructions ([page](contextual-prose.md)). | 133 pages of one tree, then 38 documents and 455 generated versions; one of four primary constructions met the criterion. | Yes; all arms kept 20 groups. | Done: `filler.instruction-scaffolding` was accepted at 39.0% on real pages, 69.0% on generated ones ([report](../../research/report-v2.md)). | Rules, shipped. |
| [TETRA](https://aclanthology.org/2024.bea-1.21/) editor revisions (365) | Editors' revisions of academic papers. Metric: not recorded. | 2,984 edits imported ([import](../../research/reviews/2026-09-30-tetra-import/README.md)); the transfer reached 4 of 46 events, baseline 7 ([record](learning-data-methods.md)). | Weakly: six pages of the screen. An editor's preference still labels no needed edit. | Done as a reference: on 3,876 sentences the default rules flag 0.6% of the edited sentences and 1.0% of the others ([record](../../research/reviews/2026-11-18-tetra-rules/README.md)). | A reference for academic text. No rule. |
| Offline ensembles over the event pool (421) | Own: select saved candidates to reach 80% of events at 85% accepted. | 36 pages: without long sentences, 25 of 37 accepted, 21 of 57 events ([record](../../research/reviews/2026-11-03-issue421/README.md)). | For the screen only. | Done as review inside a rule ([ADR 0049](../adr/0049-defaults-follow-reviewed-forms.md)). | An evaluation practice. |
| Semantic admission, [APE](https://aclanthology.org/2026.emnlp-main.2302/) encoder (349) | Own: a local encoder admits candidates. Encoder metric: not recorded. | 2,101 candidates: AUC 0.709; 39 of 57 events at 84.0% accepted, gate 46 at 85% ([record](../../research/reviews/2026-11-09-semantic-admission/README.md)). | For the screen only. | Yes, on the findings of the corpus, which carry two reviews. | A reviewer-tool filter. |
| NLI evidence state (349) | Own: an entailment model checks support for a flagged claim. Model metric: not recorded. | 16 built pairs: 9 of 16, then 15 of 16 with one contradiction missed ([first](../../research/reviews/2026-11-10-nli-evidence-state/RESULT.md), [second](../../research/reviews/2026-11-11-explicit-proposition-nli/RESULT.md)). | No: too few pairs. | Only after a step that finds the evidence. | A reviewer-tool filter. |
| [WikiSQE](https://arxiv.org/abs/2305.05928) vague-sentence classifier (349) | Wikipedia sentences tagged as vague. Metric: not recorded. | 108 candidates on 36 pages covered 3 missed events, gate 8 ([record](../../research/reviews/2026-11-09-wikisqe-vague-screen/README.md)). | For the screen only. | Yes, on items with two labels. | A rule candidate. |

## Word lists and phrase rules

| Method | Claim | Earlier test here | Does the conclusion follow? | Can it be tried now? | What it would add |
| --- | --- | --- | --- | --- | --- |
| Excess vocabulary, [Kobak et al.](https://arxiv.org/abs/2406.07016) | 900 excess words in biomedical abstracts, 407 of them style words (issue comments). | 38 sources against 455 generated versions: 190 candidates ([data](../../research/corpus/2026-11-11-assistant-panel/phrases/excess.json)), about a dozen style words. With the external lists, 51 phrases were left ([report](../../research/report-v2.md)). | Yes: the excess measured the task of the generated versions. | Yes, on 959 pairs around a cleanup commit ([ADR 0056](../adr/0056-rules-tested-on-de-slop-commits.md)). | Stock-word candidates. |
| [Antislop](https://arxiv.org/abs/2510.15061), [slop-forensics](https://github.com/sam-paech/slop-forensics) | Words and n-grams that models overuse (issue comment). Metric: not recorded. | Lists read: 1,000 words, 200 bigrams, 200 trigrams, from creative writing. No review run. | Yes for the lists; the method is untested. | Yes, as above. | Phrase candidates. |
| [Wikipedia: Signs of AI writing](https://en.wikipedia.org/wiki/Wikipedia:Signs_of_AI_writing) | Editors' field guide; single signs are weak evidence (issue comment). | Read for ideas; it overlaps the 321 reviewed candidates (same report). | Yes: connectives and summary openers were accepted at about 5%. | Yes, per construction. | Rule candidates. |

## Linters

| Method | Claim | Earlier test here | Does the conclusion follow? | Can it be tried now? | What it would add |
| --- | --- | --- | --- | --- | --- |
| Linters of issue 434: Harper, LanguageTool, textlint, retext, proselint, slopscore | Rules for repetition, phrases, readability. Benchmark: not recorded. | 18 built controls on this engine; no new rule; no linter ran ([audit](../../research/methods/audits/2026-11-07-external-controls/README.md)). | For the engine only. | Yes, in a container. | A baseline. |
| [Vale](https://github.com/vale-cli/vale), [write-good](https://github.com/btford/write-good) | General style rules. | 83 pages: fixed by a model with write-good's findings, a page was the less AI-sounding one against its original on 45 of 71 decisive pages; with Vale's two styles on 21 of 55 and 16 of 68 ([ADR 0059](../adr/0059-page-index-and-fix.md)). | Yes, by model judges. | Done. | A baseline. |
| De-slop linters: [slopless](https://socket.dev/npm/package/slopless), [dslop](https://github.com/aaazzam/dslop), [unslopify](https://pypi.org/project/unslopify/), [slop-guard](https://pypi.org/project/slop-guard/), [plainly](https://github.com/Mapika/plainly), [slop](https://github.com/bheijden/slop), deslop, defluff | Phrase, punctuation, and sentence-length checks (issue comment). slop's own audit: 1.96 hits per 1,000 words on human text, 1.01 on model text ([review](../../research/methods/non-llm-text-methods-2026.md)). | None: third-party packages could not be installed and run where that part was measured ([record](../../research/reviews/2026-11-12-human-reference/README.md)). | No conclusion exists. | Yes, in a container without network. | Rule candidates, a baseline. |

## References made by people

| Method | Claim | Earlier test here | Does the conclusion follow? | Can it be tried now? | What it would add |
| --- | --- | --- | --- | --- | --- |
| Slop taxonomy, [Shaib et al.](https://arxiv.org/abs/2509.19163) | Editors marked spans in 150 news articles and 100 answer passages; alpha 0.34 to 0.45 (issue comment, from the abstract). | None: the data was not released ([record](../../research/reviews/2026-11-12-human-reference/README.md)). | No conclusion exists. | When the data is public. | An evaluation reference. |
| [LAMP](https://arxiv.org/abs/2409.14509), [Beemo](https://huggingface.co/datasets/toloka/beemo) | People edited model text; writers prefer model-edited to raw text in 71.6% of judgments. | Run: writers agree at kappa 0.11 by word; judge models match them in about half of the pairs ([results](../../research/reviews/2026-11-12-human-reference/RESULTS.md)). | Yes, for fiction and chat answers. | Done. | Evaluation references. |
| Cleanup commits ([ADR 0056](../adr/0056-rules-tested-on-de-slop-commits.md)) | Own: 1,556 Markdown files before and after a commit that says it removes AI-sounding wording. | 597 test pairs: the default rules were lower after the commit in 64.0% of the pairs that differ, the em dash rule in 70.4%. | Yes, for commits found by message. | Done. | An evaluation reference. |

## What changed since the issue was opened

- The screen is retired. One annotator marked 123 events on 36 pages. More than 60 frozen pilots failed, and the engine reaches 45 events ([ADR 0046](../adr/0046-assistant-panel-corpus.md)). Rows marked "for the screen only" lost their gate.
- Editors agree weakly. Three writers agree at kappa 0.11 by word on where to edit ([results](../../research/reviews/2026-11-12-human-reference/RESULTS.md)). Matching one editor is an unstable target ([ADR 0052](../adr/0052-claims-follow-human-references.md)).
- Rules fire as often on kept sentences as on reworded ones. On 61 documentation pull requests the default rules flag 4.0% of the reworded sentences and 4.2% of the others. Review by a model gives 18.8% against 11.6% (same results). Academic papers that editors revised give the same answer for the default rules: 0.6% against 1.0% ([record](../../research/reviews/2026-11-18-tetra-rules/README.md)).
- Model judges do not match writers. Three judges agree with the writers' majority in 50.7%, 49.2%, and 51.7% of pairs, against a floor of 70% (same results).
- Surface marks follow what people clean. A page index of six marks tells pages before a cleanup commit from older README files with an AUC of 0.867 (ADR 0059, extended in [ADR 0060](../adr/0060-rhythm-reading.md)).

## Worth a trial

1. De-slop linters on the cleanup-commit reference. Expected gain: small. Their phrase sets will probably fail, as phrase rules did here and in slop's own audit. Their punctuation and sentence-length checks resemble the marks that passed and may add a signal to the page index. Cost: each package in a container without network, new pairs, and the bar of ADR 0056 (65%, lower end above 55%).
2. Excess frequency on cleanup-commit pairs. Expected gain: a few stock-word candidates, measured against what people removed. A list of 160 chat phrases fell short of its bar, 66.2% against 70% ([words](../../research/reviews/2026-11-12-human-reference/RESULTS-words.md)). Cost: one script over the 959 development pairs, a test on new pairs, no model calls.
3. The editor-annotated taxonomy, once released. Expected gain: an editor reference for review findings outside fiction and chat answers, capped by editor agreement of 0.34 to 0.45. Cost: two review runs over 250 texts.

## Not worth a trial, and why

- Document-level origin detectors. They miss polished text (0, 4, and 31 of 228) and weigh up to 4.1 GB. The page index covers the page level with counted marks.
- Fitted sentence selectors (n-gram baselines, WikiSQE, semantic admission, TETRA transfer). A fitted word model had no recall on real pages (ADR 0047), and the label is unstable: annotators agree at kappa 0.43, editors at 0.11.
- An NLI filter. The earlier test was too small to reject it. A filter is a confirmation step, and confirmation by a second model moved the margin on documentation from +7.2 to +7.6 points (ADR 0052).
- Longer phrase lists from other registers. They left 51 candidates, most already tested.
- Offline ensembles over the event pool. The pool belongs to the retired screen.
- Span detectors, for now. Nothing was measured and no cost is recorded. One bounded run on cleanup-commit pairs would settle it.

## Limits of this review

- Read from abstracts or search summaries only: Shaib et al., Antislop, [Almost AI, Almost Human](https://aclanthology.org/2025.findings-acl.1303/), three 2026 papers on stylistic variation and lexical overuse, and the pages of the de-slop linters.
- Run by no record: the de-slop linters, the span detectors, the detectors of issue 53, five LLMDet models, and six linters of issue 434.
- This page ran nothing of its own. Each number comes from the linked record.
- The human references are fiction, chat answers, pull requests on human-written pages, and cleanup commits by unknown hands. Model judges read every comparison of versions.
