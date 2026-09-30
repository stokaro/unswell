# Prometheus 2: local editorial-necessity admission screen

A fixed Prometheus 2 configuration failed the constructed technical admission
screen. It retained 9 of 12 useful edits and 4 of 12 unsafe edits, accepted a
case-sensitive identifier change, and missed a meaningful failure-condition
distinction. Full-page development inference did not start. This adds negative
evidence to [#349](https://github.com/stokaro/unswell/issues/349); it does not
measure document recall or qualify a product model.

## Hypothesis and fixed decision

The [Prometheus 2 publication](https://aclanthology.org/2024.emnlp-main.248/)
and [author implementation](https://github.com/prometheus-eval/prometheus-eval)
provide a judge that accepts a custom rubric. This experiment tests whether a
reference-free necessity and safety rubric can distinguish useful technical
editorial proposals from optional paraphrases and edits that change meaning.
The task differs from authorship detection and general fluency evaluation.

The model receives complete constructed sources and proposed diagnoses/actions,
using the pinned upstream `ABSOLUTE_PROMPT_WO_REF` and system prompt. No reference
answer, expected score, group label, event map, or prospective page enters a
request. The [protocol](protocol.json), [requests](controls-requests.json), and
[expected dispositions](control-expectations.json) were frozen before inference.
The expectations are assistant-authored controls, not independent human labels.

Retain only a complete EOS-terminated response with exactly one terminal
`[RESULT]` integer at least 4 on the 1–5 ordinal scale. Feedback must be
nonempty and substantive. Invalid outputs are unavailable; no retry, output
repair, fitted cutoff, source truncation, or fallback is allowed. Ordinal scores
are not probabilities. All six fixed criteria must pass before any development
inference can begin.

## Complete result

All 35 cases produced complete parseable responses. The
[raw outputs](predictions.json), [summary](preflight-summary.json), and
[terminal run record](run-record.json) support these counts:

| Group | Result | Fixed criterion | Outcome |
| --- | --- | --- | --- |
| Useful edits | 9/12 retained | At least 10 | Fail |
| Unsafe edits | 4/12 retained | At most 1 | Fail |
| Optional paraphrases | 1/4 retained | At most 1 | Pass |
| Case-sensitive identifier pair | 1/2 correct | Both correct | Fail |
| Grammar-only repair leaves rhetoric | Correctly rejected | Must reject | Pass |
| Local contextual distinctions | 3/4 correct | At least 3 | Pass |

The judge recognizes the exposed unsupported numerical-inventory hook and the
later redundant checkpoint summary. It also accepts removal of a distinct
permanent-error behavior as redundant. Other unsafe admissions change a
validation condition, cache condition, component responsibility, or database
role. It accepts `parseURL` becoming `parseUrl`. A lower cutoff cannot repair
these already-high grades.

The four local context controls are short constructed examples, not the six
original Ptah pages or the frozen 123-event evaluation. No event credit,
combined-stream precision, calibration claim, or probability follows from them.
The unchanged 804-candidate payload remains local and was never evaluated by
this configuration because preflight failed. The reserve remains unused.

## Checkpoint and execution evidence

The [author checkpoint](https://huggingface.co/prometheus-eval/prometheus-7b-v2.0)
is pinned at `66ffb1fc20beebfb60a3964a957d9011723116c5`. Its model card declares
Apache 2.0. The upstream prompt source is pinned at
`dcfb44272d5d0414832f5dbb8c2a05ebc2614234`, also under Apache 2.0.
The [manifest](model-manifest.json) records all resource sizes and hashes.
Weights are not distributed in this record.

The original unquantized BF16 weights run locally on Apple Metal through
MLX 0.32.3 and MLX-LM 0.31.3. Every one of eight acquired shards is verified
against its pinned byte size and SHA-256. All 291 tensor names and shapes match
the pinned Mistral configuration; strict loading retains the complete
14,483,464,192-byte BF16 parameter inventory. The published tokenizer and chat
template are retained without remote code or duplicate BOS insertion.

Two setup failures occurred before any model response: insufficient scratch
disk space, then an assumption that the MergeKit index contained optional
`total_size` metadata. Both [failed setup records](setup-records.json) remain
separate. Before the first inference, the loader was amended to derive the
BF16 inventory from every expected tensor shape. Inputs, rubric, thresholds,
decoding, checkpoint, and budgets did not change. No failed model call was retried.

The complete run took 1,065.9 seconds including acquisition. Peak MLX memory
was 13.76 GiB, below the fixed 24 GiB guard; recorded peak process RSS was
3.76 GiB. These are different measurements, not interchangeable totals.
Public resource acquisition preceded inference; Python socket access was
disabled during inference. External model calls and paid provider calls were zero.

This is a local MLX implementation with checked checkpoint integrity and
architecture, not numerical parity against the authors' original backend or
reproduction of their publication's evaluation. The conclusion is limited to
this fixed rubric, prompt format, backend, and checkpoint. Newer
[M-Prometheus](https://arxiv.org/abs/2504.04953) models were source-reviewed but
not executed; their separate checkpoint terms remain outside this experiment.

## Replay and limits

Run `python3 -B verify.py` and `python3 -B -m unittest test_verify.py` here.
The model-free verifier independently parses raw grades, checks request hashes,
recomputes all group decisions, and verifies the source archive and checkpoint
record. Negative tests reject missing cases, duplicate IDs, unavailable outputs,
truncation, changed grades, ambiguous results, and altered group denominators.

`runner-source.tar.gz` preserves the actual preparation/execution code, initial
loader, eight offline loader/decision tests, pinned tokenizer metadata, author
prompt source, and its license. It contains no weights, full-page development
payload, or private annotations. The original freeze names the private payload
by hash; exact full-run preparation therefore requires the preserved local
packet. The public raw controls and verifier are independently inspectable.

Do not integrate this configuration as an editorial selector or safety check.
Remaining work needs source-localized, separately preserved claims and evidence
that actually distinguishes necessary edits from valid technical exposition.
The [rule-evidence child replay](../2026-09-30-rule-evidence/README.md) addresses
one representation loss; it does not make these semantic judgments correct.
