# Matched training with partial intent labels

The [source-review contract](partial-source-labels.md) preserves unknown edit
purposes. A completed offline comparison now measures how that change affects
training on the same 256 original source cases. Masking unknown dimensions
increases target-byte coverage, but the resulting diagnostic stream fails the
recall screen and has not undergone semantic usefulness review.

## Completed comparison

Both arms use the same frozen ModernBERT features, source cases, positive edit
spans, project splits, initialization, optimizer, and global-necessity loss.
Only their intent labels differ:

- `default_negatives` reproduces the historical assumption that an unmentioned
  purpose is absent. This is an experimental control, not the accepted contract.
- `partial_masks` learns from observed binary intent dimensions. Optional and
  unknown dimensions contribute no intent loss; an empty mask skips that loss.

The packet contains 156 admitted historical cases and 100 newly reviewed cases.
Its 77 historical quarantines remain excluded. No repaired variants enter
training. Global necessity has 79 positive and 153 negative labels; 24 cases
remain optional or unknown.

The masked arm has 88 positive, 11 negative, and 1,949 unobserved intent
dimensions. The control has 88 positive, 1,887 negative, and 73 unobserved
dimensions. Thus 1,876 inferred negative dimensions differ between the arms.
The original documents and evaluation labels remain unchanged.

Thirteen project folds produce 26 heads. Each fold excludes its tested project
from both training and calibration. The token head has 1,536 inputs, 64 hidden
units, and nine outputs: global necessity and eight nonexclusive purposes.
The registered run uses 120 epochs, learning rate 0.002, weight decay 0.01,
two CPU threads, and deterministic execution. It completed in 181.44 seconds
with a 538,574,848-byte peak child resident set. It made no provider request,
encoder pass, or encoder fit.

Calibration uses the unchanged 17-point grid, both known global-necessity
classes, at least ten selected known cases, and at least 85% decision precision.
That precision assesses a **whole-unit necessity proxy**. It does not assess
the correctness of a delivered purpose or the usefulness of a complaint.
The same cutoff gates pooled unit necessity and token intent scores.
Delivery permits at most two purposes per unit and eight byte runs per purpose.
Protected source is excluded; no rewrite or probability is supplied.

Both arms introduce a global-necessity output. Their comparison isolates intent
masking within this design, not the effect of the global output against the
preceding eight-output experiment.

| Measurement | Default negatives | Partial masks |
| --- | ---: | ---: |
| Folds with an admissible proxy cutoff | 2/13 | 3/13 |
| Available original units | 295/4,549 | 2,411/4,549 |
| Proposed criticisms | 57 | 1,262 |
| Events with complete target-byte union, any purpose | 1/123 | 11/123 |
| Events with complete target-byte union and matching purpose | 0/123 | 5/123 |
| Semantically reviewed criticisms | 0 | 0 |
| Semantic FULL recall and actual usefulness | Unmeasured | Unmeasured |
| Product acceptance credit | 0 | 0 |

The masked arm emits 1,232 of its 1,262 proposals on Ptah. Removing false
negative labels makes more outputs possible, but does not make those outputs
correct. The global calibration proxy supplies no per-purpose precision
guarantee. Promoting this stream or reporting 11 events as FULL would be wrong.

Every original event remains in the accounting, including unavailable outputs.
The byte-union check allows several delivered spans to cover an event's
components; it does not require one large enclosing span.

| Original cohort | Events | Default byte union | Masked byte union | Masked matching-purpose union |
| --- | ---: | ---: | ---: | ---: |
| Context | 46 | 0 | 5 | 0 |
| Whole documents | 57 | 1 | 4 | 4 |
| Historical exposed confirmation | 20 | 0 | 2 | 1 |

Byte coverage still requires semantic review. The existing 80% FULL and 85%
usefulness requirements apply to every cohort, followed by separate untouched
confirmation. No quality score is inferred from an unreviewed stream.

## Remaining supervision gaps

| Purpose | Observed positive units | Observed negative units | Optional or unknown units |
| --- | ---: | ---: | ---: |
| Clarity | 4 | 0 | 252 |
| Empty framing | 36 | 3 | 217 |
| Fluency | 26 | 0 | 230 |
| Needless complexity | 1 | 1 | 254 |
| Needless repetition | 1 | 2 | 253 |
| Unjustified intensifiers | 4 | 2 | 250 |
| Vague claims | 5 | 1 | 250 |
| Wordiness | 11 | 2 | 243 |

Known global retention is not eight explicit intent negatives. A unit containing
one reviewed repair also does not label every untouched token as background.
The packet therefore lacks explicit retention evidence for most intent and
token decisions. The positive span loss supervises 100 purpose-expanded edit
targets but supplies no explicit negative token spans.

The Python [support report](../../research/annotation/source_review/label_support.py)
counts observed classes and rejects missing dimensions or inconsistent masks.
It reports class support without declaring a model available. Its regression
tests preserve unknown and optional decisions, nonexclusive positives, and
explicit negatives. Completing a class count is still insufficient for
calibration or editorial qualification.

Further acquisition should explicitly review every purpose while permitting
unknown answers, include difficult retention controls, and review contextual
relations. Another feature-distance sample alone is unsupported by the preceding
acquisition result. A registered uncertainty-versus-random comparison can test
a different selection mechanism before more fitting.

The subsequent [complete-purpose acquisition and learning experiment](complete-source-acquisition.md)
finished that comparison and an additional 26-head annotation-effect study. It
adds explicit retention evidence but fails the complete-target-byte recall screen.
The historical counts above describe the unchanged 256-case comparison.

## Research and Python implementations

Primary sources were inspected on October 9, 2026.

| Source | Relevant method | Boundary |
| --- | --- | --- |
| [LLMaAA, Findings of EMNLP 2023](https://aclanthology.org/2023.findings-emnlp.872/), [paper](https://arxiv.org/html/2310.19596v2), [author Python code](https://github.com/ridiculouz/LLMaAA) | Acquire LLM annotations with a student uncertainty score, compare with random selection, and use retrieved demonstrations. | The experiments concern named entities and relations. Automatic reweighting also assumes a separately trusted validation set; the paper is not replicated here and supplies no Unswell accuracy claim. |
| [Active Learning Design Choices for NER with Transformers, LREC-COLING 2024](https://aclanthology.org/2024.lrec-main.30/), [paper](https://aclanthology.org/2024.lrec-main.30.pdf) | Its partial-token experiments retain sentence context while masking unknown supervision, with a fully annotated initial seed. | NER labels differ from editorial necessity. Masking unknowns cannot replace explicit retention evidence; no paper model was trained here. |
| [small-text query strategies](https://small-text.readthedocs.io/en/latest/components/query_strategies.html), [initialization](https://small-text.readthedocs.io/en/latest/components/initialization.html) | Python implementations of uncertainty, contrastive, diversity, and random acquisition support a controlled annotation loop. | Context7 returned no matching library, so official documentation was used. The library was not installed or run. Its classification interface does not establish our partial-token contract or project exclusions. |

## Verification and product status

The [measurement record](partial-label-learning-result.json) binds the packet,
all 26 saved heads, native completion, deliveries, calibration grids, and
independent byte-union accounting. An initial reconstruction from memory-mapped
weights differed only in floating-point logits, by at most
`2.8312206268310547e-07`; spans, quotes, purposes, and cutoff decisions agreed.
Loading the same weights into cloned Parameter storage then reproduced every
saved output exactly. Both attempts are retained. No training, threshold,
saved prediction, or tolerance was changed to make the audit pass.

A draft registration was superseded before execution to make calibration and
delivery use the same pooled necessity gate. It performed zero fits. The
completed packet records that preexecution amendment explicitly.

Publishing aggregate records does not claim public reproduction of private
annotation judgments. The source cases, reviews, cached features, and prediction
streams remain in the local research archive.

The published analyzer remains at 37/123 FULL events and 94/154 useful actual
criticisms across all 36 pages, 4,549 units, and 804 judgments. This comparison
changes no runtime, model distribution, gate, or playground build.
[Issue #349](https://github.com/stokaro/unswell/issues/349) remains open.
