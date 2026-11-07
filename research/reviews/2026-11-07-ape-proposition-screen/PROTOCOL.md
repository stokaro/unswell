# APE proposition-overlap feasibility screen for #349

Status: frozen before any APE inference, November 7, 2026. This is an
exposed-development screen, not prospective qualification or a runtime rule.

## Inputs and boundaries

- Unswell source tree: `6e724618005623eddbd325653c0e04e7c38b317d`.
- Use the 36 source documents returned by
  `research/reviews/2026-09-26-outcome-announcements/replay.py::sources`.
  That loader verifies each original SHA-256. Keep all 123 frozen events and
  the three complete cohort denominators. Do not open reserved prospective
  pages or modify reference labels.
- APE code: `SapienzaNLP/APE` commit
  `a6723d56e70a5fc92577858c9c2a2f17a6cbfe2a`, with its committed
  `uv.lock`. Checkpoint: `sapienzanlp/ape-bge-small` revision
  `e6dc09deb8c9e18323f17bbfd35b9025d89d8044`. Record downloaded file
  SHA-256 values before inference. Its base tokenizer and architecture config
  come from `BAAI/bge-small-en-v1.5` revision
  `5c38ec7c405ec4b44b94cc5a9bb96e735b38267a`. In a local copy of the
  checkpoint config, replace only `encoder_name` with the absolute path of
  those pinned local assets; record both config hashes. Use the checkpoint
  weights as published, without training, tuning, remote inference, or
  generated text. Require offline mode during inference.
- APE is a local discriminative encoder. Its training data include generated
  propositions; that provenance does not turn its output into editorial
  ground truth. It cannot represent causal or conditional relations between
  propositions. Do not use it to approve a textual rewrite.

## Source-local candidate construction

Use Unswell's current technical profile to extract prose and the current
English NLP provider to segment it. Retain original UTF-8 source ranges.
Include every eligible sentence in paragraph and fragment blocks with at least
five prose words. Do not analyze protected text, headings, or excluded blocks.
For each sentence, compare its APE proposition vectors with at most two prior
and two following eligible sentences in the same source and extracted prose
piece (block plus piece ordinal).
Never pair a sentence with itself. Do not cross a protected or structural
boundary. Skip a sentence that exceeds APE's default 128-token input limit;
record the skip rather than truncating it. Use the published activity threshold
of `0.6` and directional `h_min_to_p` score with the candidate as hypothesis.
The candidate score is the maximum over those neighbor comparisons. Preserve
all scored pairs, including low-scoring pairs; rank candidates by score within
each page, breaking ties by original byte start, and take the top three per
page. Pages with fewer than three eligible candidates contribute all of them.
No minimum similarity cutoff, event identity, reviewer grade, or target quote
may influence candidate construction. Run one local CPU pass with batch size
32 and a three-hour ceiling. MPS was unavailable in the pinned environment.

## Evaluation and stop decision

First inspect source-local pair scores for frozen positive events and frozen
legitimate controls, including contrasts involving negation, conditions,
numbers, versions, and identifiers. These examples test whether a semantic
echo signal is useful; they do not estimate population precision. Then review
every top-three candidate in its complete source context under ADR 0041,
recording accepted, rejected, or uncertain editorial necessity. Identify the
reviewer as a Codex assistant. A candidate overlapping a frozen event only
earns credit if its diagnosis matches the event, not merely its byte range.

Admit a follow-up source-local operator only if this frozen screen yields at
least eight genuinely new credited FULL events across the complete cohorts,
at least 85% accepted new candidate findings, and no unreviewed candidate.
Any unsafe suggested edit, failed source-map check, incomplete document, model
error, or hidden truncation stops qualification. These are admission criteria
for another experiment, not #349 completion criteria. Report gained and lost
event IDs, per-cohort FULL-event recall, the full delivered-stream acceptance
fraction, unavailable pages, and supplied-advice safety separately. The #349
thresholds remain 80% FULL-event recall and 85% accepted actual delivered
findings in each complete cohort. A source-local signal that fails this screen
stays a negative research result; do not retune its threshold on these pages.

The comparison is against the unchanged current-main baseline in
`research/reviews/2026-11-07-current-main-editorial-review`. Do not interpret
APE's published paraphrase or entailment benchmark scores as Unswell results.
