# Pretrained edit-quality and grounding references

Two published pretrained references were reproduced locally and failed their
fixed technical preflight screens. Neither supplies a qualified editorial
selector for Unswell. The results identify separate failures in revision benefit
and meaning preservation; the broader contextual-recall objective remains open.

| Reference and measured role | Constructed-control result | Decision |
| --- | --- | --- |
| LENS-SALSA: improvement over the unchanged source | 9/12 useful edits, 5/12 unsafe edits, and 3/4 optional paraphrases exceed the fixed gain margin | Reject this configuration as a general technical editorial selector |
| MiniCheck-Flan-T5-Large: support in both directions | Retains 10/12 safe edits and 3/12 unsafe edits | Reject the fixed standalone technical safety screen |

These are different tasks, not comparable accuracy numbers. All controls were
constructed and reviewed by the OpenAI assistant, not independent human
annotators. They provide a bounded rejection test, not an estimate of errors
on real documentation. No model was trained, no threshold was fitted to these
outcomes, and no new full-event detection credit is claimed.

## Published methods and available implementations

[LENS-SALSA](https://aclanthology.org/2023.emnlp-main.211/) evaluates individual
simplification edits and predicts error tags. Its
[Python package and model](https://github.com/davidheineman/salsa#lens-salsa)
support reference-free source/revision scoring. The
[model card](https://huggingface.co/davidheineman/lens-salsa) limits the measured
domain to English Wikipedia simplification; technical documentation was untested.
Code and model declare Apache-2.0, while the base RoBERTa tokenizer declares MIT.
These separate declarations and downloaded hashes are retained in
[records.json](records.json). No weights are bundled.

[MiniCheck](https://aclanthology.org/2024.emnlp-main.499/) verifies whether a
claim is supported by a grounding document. Its
[Python implementation](https://github.com/Liyan06/MiniCheck) and
[Flan-T5 model](https://huggingface.co/lytang/MiniCheck-Flan-T5-Large) are available.
The code declares Apache-2.0 and this model declares MIT. Grounding support
does not establish editorial necessity, and support in one direction does
not establish preservation of every original condition. The two-direction
policy here is an explicit experimental composition, not an upstream guarantee.

Several foundations address other parts of the task:

| Work | Useful concept | Execution status in this record |
| --- | --- | --- |
| [SARI, Xu et al., 2016](https://aclanthology.org/Q16-1029/) | Evaluate text simplification as a separate task with appropriate references | Reviewed; not executed |
| [Rhetorical Structure Theory, Mann and Thompson](https://www.sfu.ca/rst/pdfs/Mann_Thompson_1987.pdf) | Represent relations among discourse spans | Reviewed; not a normative defect detector |
| [Entity-based coherence, Barzilay and Lapata, 2008](https://aclanthology.org/J08-1001/) | Model entity transitions across sentences | Reviewed; not executed |
| [Prometheus 2, Kim et al., 2024](https://aclanthology.org/2024.emnlp-main.248/) | Evaluate against an explicit user-defined rubric with direct or pairwise judgments | Publication and [available Python code](https://github.com/prometheus-eval/prometheus-eval) reviewed; models not acquired or executed |

Available research implementations therefore cover editing, grounding, discourse,
and rubric-based evaluation. Their existence does not establish a ready detector
for our rubric. A next editorial critic must assess the necessity of a change
separately from its safety and retain independently actionable claims. The
[claim-accounting task](https://github.com/stokaro/unswell/issues/367) addresses
that representation boundary. Prometheus 2 is a hypothesis for a different
critic, not a selected replacement or a qualified model.

## Frozen preflight

The [constructed controls](testdata/controls.json) contain 12 source groups,
each with an identity baseline, a useful revision, and a meaning-changing
revision. Four clear sentences have optional paraphrases. A separate case-sensitive
identifier probe distinguishes `parseURL` from `parseUrl`.

The LENS protocol was frozen before any control predictions. It compares
`score(source, revision) - score(source, source)` against a fixed 2-point margin
on the published 100-point scale. That margin is exploratory, not published
calibration. Acceptance required at least 10/12 useful gains, at most 1/4 neutral
gains, useful edits ranked above unsafe alternatives in at least 10/12 groups,
and at most 1/12 unsafe gains. All four criteria failed. The underlying regression
score and its scaled output are not probabilities of needing revision.

The MiniCheck protocol was frozen before its inference, using the same unchanged
controls. Each ordinary source and revision contains one sentence. The complete
short source is passed as one chunk to the unmodified published numerical helper;
long-document splitting is outside this reproduction. A pair is retained only
when both directional support scores exceed the upstream 0.5 cutoff. Acceptance
required at least 10/12 safe pairs and at most 1/12 unsafe pairs. The latter failed.
The identifier probe uses forward checks only because its original has two sentences.

The public [protocol projection](protocol.json) records both fixed screens and
the hashes of their complete original local protocols. It was exported after
execution; it does not independently prove when those local protocols were frozen.
The local originals and unchanged execution scripts remain preserved.

## Observed failure cases

LENS rewards the unsafe `any` to `every` revision by 9.89 points, above the
correct revision's 6.70-point gain. It also rewards the changes from seconds to
milliseconds and the exchange of parser and scheduler responsibilities. Useful
edits rank above unsafe alternatives in only 9/12 groups. These observations
do not depend on labeling the 2-point margin a product threshold.

The published wrapper lowercases both inputs. The correct and incorrect
case-sensitive identifier revisions produce identical scores and token inputs.
It cannot preserve this distinction without a separate check.

MiniCheck retains the unsafe `any` to `every`, required to optional, and seconds
to milliseconds revisions under the two-direction policy. It also accepts the
identifier-case error with a forward support score of 0.933. Two useful revisions,
including a preserved numeric limit and a privacy-boundary explanation, are
rejected. All four neutral paraphrases are supported, as expected of a grounding
model; that support is not evidence that the original sentences needed editing.

Combining these metrics has not been qualified. A conjunction inspected after
these outcomes still retains three unsafe revisions and three optional changes.
It is a descriptive post hoc check, not a new preregistered screen or a trained
product policy.

## Numerical compatibility and resources

| Component | Pinned identity |
| --- | --- |
| SALSA repository | `ce5ddba3d338704a20ceeaffbbb6382810c0c3e7` |
| Official LENS wheel | `lens-metric==0.2.0`, SHA-256 `e8bdc18a87a66875035187eb88e5f8f835aee975df8089e93f85102c45000f80` |
| LENS checkpoint | `davidheineman/lens-salsa@cc04094d084493e68bcd3eb196ff74e894cca00d` |
| RoBERTa tokenizer/configuration | `FacebookAI/roberta-large@722cf37b1afa9454edce342e7895e588b6ff1d59` |
| MiniCheck code | `b58b9fa69acbd1015ec970fa65dd752413a053d2` |
| MiniCheck checkpoint/tokenizer | `lytang/MiniCheck-Flan-T5-Large@96eafd01cee2d16cf81aaa2fb226b14f422a37b3` |

The LENS author example reproduced 72.40906 against 72.40909 published, within
the fixed 0.001 tolerance. Both MiniCheck author examples match within the same
tolerance. The local loaders instantiate architectures from pinned configuration
and strictly restore every parameter from verified checkpoints. PyTorch uses
`weights_only=True`; no checkpoint code or remote model code executes. One old
LENS positional-ID buffer is verified against the identical regenerated value.

The LENS wheel's effective layer mixture is softmax despite sparsemax in its
hyperparameter metadata. The reviewed training constructor also omits forwarding
that setting. This run preserves the published inference behavior and matching
smoke value; it does not silently correct that discrepancy or reproduce all
scientific results in the paper. The Git repository alone lacks an encoder module
needed for inference, so the official wheel is the executed implementation.

LENS has 354,706,205 parameters and a 1,418,984,537-byte checkpoint. The complete
331-score run, including the private transfer below, took 74.78 seconds and
2,098,790,400 bytes peak RSS on the local macOS arm64 host with four CPU threads.
MiniCheck has 783,150,080 parameters and a 3,132,786,242-byte checkpoint. Its 60
scores took 39.60 seconds and 3,807,969,280 bytes peak RSS on the same host.
These are reference costs, not the product's target-host scan benchmark.

Inference ran offline with outbound socket connections rejected. An initial
MiniCheck import attempted to acquire missing NLTK data; the socket guard stopped
it before model construction or scoring. Genuine public Punkt data was then
acquired separately, and the unchanged frozen script completed its first model
inference attempt. The acquisition and setup identities are retained in the record.
No private research data was transferred to a provider or remote host.

## Transfer to existing technical revisions

LENS also scored the unchanged 142 previously proposed revisions from six exposed
technical pages, without changing their labels or source units. It evaluated
138 pairs and explicitly abstained on four empty replacement texts. None of
the available pairs was silently truncated. Ninety-seven gains were positive;
78 exceeded the fixed margin. These counts do not measure editorial precision
or detection recall: no new rewrite labels or event witnesses were created.

This portion is published as aggregates only. The original pages, judgments,
full predictions, token traces, and input bindings remain local. The public
verifier checks aggregate arithmetic but cannot reproduce private task-level
transfer or independently validate those judgments. This limitation is explicit
in the record and verifier output. Reserved prospective pages were not used.

## Offline verification and optional reproduction

Run from this directory with an ordinary Python interpreter:

```sh
python3 -B verify.py
python3 -B -m unittest test_verify.py
```

The verifier checks archive/file hashes, exact constructed-control identities,
numerical smoke results, both grounding directions, availability, and the two
fixed decisions. Eight negative tests reject missing or duplicated predictions,
changed availability, NaN, a one-direction substitution, an incorrect cutoff
label, a missing group, and an invalid second direction hidden by the first.
It has no model-package, GPU, or network dependency.

The [runner archive](runner-source.tar.gz) contains the actually executed local
loaders, scoring scripts, freezer, and evaluator. Its exact members and hashes
are in [manifest.json](manifest.json). Extract it only into an isolated research
directory; obtain the public resources at the pinned revisions above and compare
their byte sizes and SHA-256 hashes in `records.json`. LENS uses the official wheel,
PyTorch 2.10.0, Transformers 4.36.2, NumPy 1.26.4, pandas 1.5.3, SciPy 1.16.3,
torchmetrics 1.3.2, and Setuptools 80.9.0 under Python 3.11. MiniCheck additionally
imports NLTK 3.9.2 and its pinned source checkout. Obtain Punkt data before import
and set `NLTK_DATA` to the isolated data directory; do not enable downloads during
inference.

The archived acquisition script preserves the original acquisition procedure,
which discovered the base-tokenizer revision then pinned its downloads. For a new
reproduction, use the exact recorded base revision rather than resolving current
main. Only tokenizer/configuration files are needed from that base model; the
complete trained encoder is already in the LENS checkpoint.

The public control file can be placed at
`unswell-lens-salsa-20260930/controls.json` in that isolated layout. The archived
MiniCheck runner can freeze and execute its full constructed preflight with
`python -B run.py --freeze` followed by `python -B run.py`. The LENS smoke script
and `LocalModel.score` reproduce the author example and any supplied public
source/revision pair. Replaying the archived full LENS run additionally requires
the intentionally excluded 142-pair local packet and its original frozen protocol.

Python model environments and downloads remain optional research infrastructure.
No weights, native inference dependency, product probability, gate change,
new authorship channel, or qualified editorial model is introduced here.
