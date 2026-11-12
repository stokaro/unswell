# Assistant-panel corpus record

This directory holds the data behind the
[rule qualification record](../../../docs/research/rule-qualification.md) and
the decisions of ADRs
[0046](../../../docs/adr/0046-assistant-panel-corpus.md),
[0047](../../../docs/adr/0047-qualified-defaults-and-assistant-review.md), and
[0048](../../../docs/adr/0048-severity-follows-review.md).

Every annotator, reviewer, and auditor here is an AI assistant. `summary.json`
names the models. No label is a human judgment.

## What is stored

The record stores coordinates and decisions. It stores no page text, no quote,
and no reviewer prose. A page is identified by its repository, path, commit
reference, and SHA-256 digest, so it can be fetched again and checked.

| Path | Content |
| --- | --- |
| `manifests/<set>.json` | Corpus manifest: source, commit reference, license, digest. |
| `manifests/<set>-selection.json` | Selection log: every repository considered, in order, with the decision. |
| `manifests/<set>-parse-failures.json` | Pages that the engine could not parse and that were left out. |
| `items/<set>.jsonl.gz` | One row per document: its identity and the byte segments of each item. |
| `labels/<set>.jsonl.gz` | One row per document and annotator: the items marked, with categories. |
| `reviews/<round>.jsonl.gz` | One row per finding: rule, severity, byte span, and each reviewer's verdict. |
| `audit/*.json` | Outcomes of the blind paired comparisons, with placebo rows. |
| `phrases/*.json` | Phrase candidates, their statistics, the selection decisions, and the rule statements shown to reviewers. |
| `configs/*.yaml` | Engine configurations: every rule enabled with all candidates, and the selected lists. |
| `contextual/<caller>.jsonl.gz` | Flags that a calling assistant raised through the MCP server: categories, byte span, and the verdicts of the other assistants. |
| `contextual/command-<reviewer>.jsonl.gz` | Flags that a reviewer raised through `unswell review` on generated documents, with the verdict of each confirming model. |
| `rounds/*.json` | Acceptance by sentence length, the comparison of defaults before and after round three, the measurement of review in context, and the measurement of the review command. |
| `summary.json` | Panel, document counts, annotator agreement counts, and per-rule review counts. |
| `files.sha256` | Digest of every data file. |
| `scripts/` | The scripts that built the corpus, as they ran. |

An item is a sentence, a paragraph without separate sentences, or a fragment of
six or more words. Segments and spans are byte offsets into the page as it is
at the pinned commit.

## Sets

| Set | Role | Documents | Labels |
| --- | --- | ---: | --- |
| `long-prose` | development | 493 | two annotators |
| `real-dev-1` | development | 164 | two annotators |
| `real-dev-2` | development | 593 | none |
| `real-test-1a` | final test 1 | 127 | two annotators |
| `real-test-1b` | final test 1 | 645 | two annotators |
| `real-test-2` | final test 2 | 589 | none |
| `real-test-3` | final test 3 | 578 | none |
| `real-test-4` | final test 4 | 606 | none |

`long-prose` comes from the archives of
[the long-prose study](../../generation/studies/long-prose-v3). The other sets
are README files and documentation pages of public repositories. No repository
appears under two roles. A test set became development data after it was opened.

## Review rounds

| Round | Findings | Reviewers |
| --- | ---: | --- |
| `development` | 7,315 | `astra`, `claude` |
| `final-test-1` | 1,453 | `astra`, `claude` |
| `final-test-2` | 1,001 | `astra`, `claude`, `qwen` |
| `final-test-3` | 547 | `astra`, `claude`, `qwen` |
| `final-test-3-removed` | 593 | `astra`, `claude` |
| `final-test-4` | 413 | `astra`, `claude` |

`final-test-3` holds the findings of the round-three defaults on the third test
set, as reviewed before two rules became opt-in. `final-test-3-removed` holds
the findings that the earlier defaults reported there and round three dropped.
`rounds/round-3-test.json` compares the two.

A verdict answers one question: is the finding justified under the statement of
its rule? The values are `yes`, `no`, and `uncertain`. A missing verdict means
that the reviewer returned no answer for that finding. Nothing was repaired by
hand.

## Verify

```sh
make check-assistant-panel-corpus
```

The check runs offline. It compares file digests, rejects any field outside the
schema, binds labels and reviews to existing items, checks repository
separation between roles, and recomputes `summary.json` from the data files.
It also recounts `rounds/review-command.json` from the stored flags and labels:
the bound, kept, and dropped flags, the marked items they touch, and the
decision that follows. A run that calls its reference annotator independent is
rejected when that annotator reviews or confirms in it. Its tests break one
property at a time and expect a rejection.

## Rebuild

The scripts read and write a workspace directory outside the repository. They
need the GitHub command-line client and access to the three models. The model
behind `qwen` runs on a private endpoint; its address and key come from the
environment and are not part of this record.

1. Select and fetch pages: `acquire_real.py`, `acquire_real_dev.py`,
   `acquire_more.py`, `acquire_dev2.py`, `acquire_real3.py`, `acquire_real4.py`,
   `acquire_real5.py`, in that order. Each one skips repositories that an earlier selection log
   names.
2. Extract items: `build_docs.py` for the study archives, `prep_real.py` for
   real pages.
3. Annotate: `annotate.py` once per annotator.
4. Run the engine with `configs/broad.yaml`, then review the findings with
   `review_findings.py` once per reviewer.
5. Select phrases and defaults: `rule_table.py`, `select_release.py`,
   `gen_lexicon_go.py`. `relist.py` reads the listing criterion again over all
   opened pages and writes `phrases/relisting.json`.
6. Audit: `audit.py` for items, `audit_spans.py` for findings.
7. Compare defaults between rounds: `round3_evidence.py`, `compare_rounds.py`.
   `round4_evidence.py` computes the decisions that were fixed before final
   test 4 was read.
8. Review in context through the MCP server: `mcp_review.py` once per calling
   assistant, then `ctx_report.py`, `review_findings.py` with the other
   assistant, and `ctx_eval.py`.
9. Review with the command: `review_product.py` runs `unswell review` once per
   document, first with a reviewer alone and then with each confirming model.
   `ask_model.py` is the local program that the command starts; it stores each
   answer by prompt, so every confirming model judges the same flags.
   `review_command_evidence.py` computes the round and the registered decision.
10. Report and export: `gen_docs.py`, `export_record.py`.

`study.py`, `study2.py`, `study3.py`, `embed.py`, `mine.py`, `qualify.py`, and
`excess.py` produced the negative results on fitted models and mined phrases
that ADR 0047 reports.

A rebuild does not reproduce the files byte for byte. Repository rankings
change, and the models do not return the same answers twice.

## Limits

- Assistants wrote every label. The record does not measure agreement with a
  human editor.
- The two annotators agree on which items need revision with a Cohen kappa of
  0.43 on the long-prose set.
- Reviewers saw at most eight findings per rule and page.
- Flags raised in context were judged by one assistant of the other family.
  The third assistant judged 80 of them before its run was stopped for time.
- `real-dev-2`, `real-test-2`, `real-test-3`, and `real-test-4` have reviews of
  findings and no item labels, so they give precision and no recall.
- The auditor's paired comparison penalizes mechanical edits that break
  grammar. The qualification record states the effect.
