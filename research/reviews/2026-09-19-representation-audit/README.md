# Most reviewed defects survive extraction

On 157 previously reviewed complete pages, **920 of 994 annotated defects
(92.6%) have all their target prose in one prepared unit**. Another 54 span
protected gaps within one extracted block, and 20 span multiple blocks. These
are representation counts, not detector recall. They do not establish that a
model can identify the defect or explain a useful repair.

The result rules out missing target prose as the main explanation for the low
recall in earlier experiments. The next experiment should test a learned
representation on these exposed judgments before another sequence of narrow
surface-rule additions. Improving context across protected gaps remains useful:
54 target spans explicitly require it. Other events may also depend on context
outside their annotated targets. Code must remain protected.

| Cohort | Defects | One prepared unit | Split by protected gaps | Multiple blocks |
| --- | --- | --- | --- | --- |
| Ptah | 389 | 362 | 18 | 9 |
| Historical repositories | 605 | 558 | 36 | 11 |
| Total | 994 | 920 | 54 | 20 |

All 994 targets retain some unprotected Word tokens, and the prepared units
collectively retain all measured target token bytes. One original extracted
block contains 974 targets. The audit found no complete loss of a target's
eligible prose in this sample.

Only 439 targets fit a prepared unit whose declared kind is `sentence`. This
does not mean the other targets lack sentences in ordinary English. Under
`eligible-piece-v1`, a paragraph containing an inline protected separator is
prepared as separate `fragment` units; it supplies no `sentence` units. List
items also use fragment units. An experiment restricted to prepared sentences
would therefore omit much of the actual technical prose.

## What these measurements cover

The input includes every page in the original whole-page review and subsequent
complete-category reviews through structural-wordiness. Two earlier targeted
reviews, framing-recall and repetition-scope, have narrower annotation coverage
and are excluded. No source, label or proposed repair is changed. Pages without
defects remain in the input. Repository/path identities cannot repeat.

All labels come from the named Codex assistant and are already exposed
development evidence under ADR 0041. This audit does not supply new confirmation,
independent human annotation, authorship labels or product model qualification.
The 556 controls and 185 uncertain judgments retain their original status.
Unmarked text is not automatically a confirmed clean training example.

The measurement intersects annotated target ranges with the source segments of
unprotected Word tokens. It asks whether the whole resulting set fits one
block, one prepared enclosing piece, or one prepared sentence. A multi-location
event counts once. Markup, whitespace, punctuation and protected operands are
outside the measured token set, so this is an optimistic bound. Necessary
context may extend beyond the annotated target or the enclosing unit.

The original token inventory covers every nonexcluded extraction kind. It does
not infer missing text from the older block-feature catalog's `unsupported_unit`
status: that catalog does not measure list items, while prepared fragment
features do. The exporter binds its second extraction to the engine's original
block hashes and preserves the engine's exclusions, including non-Latin prose.

Twenty-three defects share an original block with an explicit technical control;
sixteen share a prepared piece. Such co-location prevents using all control
pieces as unconditional negative training examples. It also illustrates why a
paragraph score cannot be applied indiscriminately to every sentence in it.

## Reproduce

The audit command reuses the public engine, extraction, English provider and
prepared-unit contract. It performs a complete technical-profile run, requires
no rule abstentions or skipped rules, and exports only the representation
projection needed for this audit. It changes no runtime detector or gate.

From the repository root:

```sh
python3 research/reviews/2026-09-19-representation-audit/tools/prepare.py --output /tmp/representation-inputs.json.gz
cd research/annotation
CGO_ENABLED=0 go build -trimpath -o /tmp/representationaudit ./cmd/representationaudit
gzip -dc /tmp/representation-inputs.json.gz | /tmp/representationaudit | gzip -n > /tmp/representation-report.json.gz
```

The preparation output must be new. Check the command exit status when using a
pipeline; the verification below rejects incomplete JSON or analysis results.
From this review directory:

```sh
python3 tools/evaluate.py --inputs /tmp/representation-inputs.json.gz --report /tmp/representation-report.json.gz --output /tmp/representation-measurements.json
python3 tools/verify.py
python3 tools/test_evidence.py
```

`measurement-record.json` binds the source manifests, exporter sources, executable
and retained report. The current detector base is
`fa1743891a3e82d2d5361f41762604767b497b8f`; the report marks the separate audit
instrumentation in its build identity. Go blackbox tests cover source drift,
duplicate IDs, cancellation, protected pieces and post-extraction exclusions.
Evidence tests reject missing documents and distinguish multi-block targets,
protected-gap splits, token loss and empty eligible ranges.

## Next experiment

The reproducible `supervision.json.gz` projection contains 872 positive units
and 1,234 negative units. A positive contains an entire reviewed edit target.
A negative's eligible tokens are completely covered by an explicit control,
with no defect or uncertain target overlapping it. The remaining 27,644 units
stay unlabeled. This derivation selects candidate units; it does not turn their
whole text into a diagnosed defect or supply a human-adjudicated training set.

Length is a confound: the 1–9-word band has 72 positives and 559 negatives;
10–39 words has 529 and 545; 40 or more has 271 and 130. A learned candidate must
therefore beat a length-only baseline and report these bands separately. Overall
accuracy could otherwise reward the same length preference that the user found
unhelpful. Units from one source remain dependent observations.

Compare the existing rules with a compact Go model using shared lexical and
structural features. Fit only on explicitly reviewed, unambiguous units; keep
source groups and exact reused prose together. Use the current reviews only
for development. Measure selection of the right source units separately from
actual defect localization and useful diagnostic explanations. A new complete-page
confirmation sample is still required before claiming improved transfer.

Keep missing features explicit, fit vocabulary and normalization only on training
groups, and choose thresholds on development groups. Do not treat raw model
responses as calibrated probabilities, route experimental scores into the default
gate, or reinterpret the reviews as human training annotations.
