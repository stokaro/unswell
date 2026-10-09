# Source-unit learning with auxiliary explanations

The completed offline comparison did not produce an admissible diagnostic
configuration. The source-only classifier selected 61 necessary units among
66 selected units with known decisions, but its actual diagnostic stream
supplied only nine witnesses with the original byte coverage and matching
intent among 123 unchanged reference events. Auxiliary explanation supervision
supplied eight. These witnesses are not semantic FULL acceptance.

The [measurement record](source-student-learning-result.json) retains the
registration, native completion, and independent verification hashes. No
candidate was added to Unswell. Published core `8a56d13f2ac0` remains at 37/123
FULL events and 94/154 useful actual criticisms on the
[complete development replay](discourse-stance-context.md). Acceptance still
requires at least 80% FULL recall and 85% useful actual criticisms in each
cohort, zero unsafe or unresolved advice, and separate prospective confirmation.

## Annotations and matched training

The study extends the [source-annotation investigation](learning-data-methods.md).
Its frozen packet preserves 123 reference events and adds 39 complete-unit
reviews and seven construction-retention anchors. The root Codex assistant
reviewed the additional units under
[ADR 0041](../adr/0041-assistant-review-acceptance.md); this is exposed development
evidence, not independent human annotation.

The 39 reviews identify four necessary edits, 31 retain decisions, three
optional edits, and one unresolved case. Seventeen original-event explanations
have supported teaching associations. Four disputed original teaching
associations and the unresolved review are masked for training; all original
events remain in evaluation. A retained construction cannot label its whole
unit negative.

The resulting supervision covers 122 necessary units and 34 units with no
mandatory edit. The other 4,393 units remain unknown. There are 56 eligible
unit-associated explanation outputs. Other intents in a positive unit remain
unknown unless explicitly reviewed.

The three arms predict from original source: a hashed lexical linear baseline,
a source-only neural head, and the same neural head with an auxiliary
explanation-embedding output. The latter is motivated by
[Distilling Step-by-Step](https://aclanthology.org/2023.findings-acl.507/); it does
not replicate that paper's generated-rationale decoder.

ModernBERT-base is frozen at revision
`8949b909ec900327062f0ebf497f51aef5e6f0c8`. Target-unit pooling uses complete
neighboring units and hash-verified local and whole-document encodings. All
tokens are covered. The inference input receives neither the teacher's
explanation nor its judgment, and no encoder parameters are fit.

The unchanged 13 source groups define outer test folds, with three other
groups reserved for calibration. These are source groups, not 13 independent
repositories. Supervised training, auxiliary outputs, and lexical IDF estimation
exclude both test and calibration groups. All 4,549 units receive exactly one
outer-test prediction per arm. Matched neural heads use identical initialization
and 200 epochs. All 39 head fits completed without provider requests.

Calibration requires ten selected, explicitly labeled units at 85% decision
precision and both decision classes. A missing feasible point is unavailable.
These sparse-label scores are not product probabilities or actual usefulness.
Delivery requires both a selected unit and a supported intent with a nonnegative
logit, emitting at most two intent messages per unit without replacement text.

## Complete result

| Fixed arm | Available groups | Selected units | Proposed criticisms | Single-witness byte coverage | Coverage with matching intent |
| --- | ---: | ---: | ---: | ---: | ---: |
| Hashed lexical | 6/13 | 1,213 | 52 | 2/123 | 1/123 |
| Source only | 6/13 | 1,305 | 1,314 | 22/123 | 9/123 |
| Auxiliary reason output | 6/13 | 1,322 | 1,277 | 23/123 | 8/123 |

A byte-covering witness independently contains every target component. Separate
partial diagnostics are not pooled into FULL credit. Even this optimistic
coverage is below 80% in every cohort:

| Arm | Context, 46 events | Whole documents, 57 events | Historical confirmation, 20 events |
| --- | ---: | ---: | ---: |
| Hashed lexical | 0/46 | 2/57 | 0/20 |
| Source only | 2/46 | 9/57 | 11/20 |
| Auxiliary reason output | 3/46 | 9/57 | 11/20 |

The candidate streams remain unreviewed. No actual usefulness or semantic FULL
result is claimed. Every miss and unavailable output stays in the original
denominators. Precision on selected known units is 51/58 for the lexical arm
and 61/66 for both neural arms; known-positive decision recall is 51/122 and
61/122, respectively. This selected annotation subset cannot estimate usefulness
throughout the complete 4,549-unit population.

Five calibration folds are infeasible from annotation counts alone: `bat`,
`containerd`, `imgui`, `prometheus`, and `ripgrep` have only six, eight, eight,
seven, and five known necessary calibration units. Reaching 85% among ten
selected units requires at least nine positives. The other two unavailable
folds fail the registered score-grid screen. This data-coverage limitation does
not establish that every unavailable model is intrinsically poor.

The result rejects these fixed delivery configurations. It does not reject
source learning or explanation supervision. Necessity selection and useful,
intent-specific diagnosis remain different learning targets.

## Registered annotation acquisition

A separate frozen pilot contains 78 previously unreviewed source units: three
random-control and three uncertainty/disagreement units from each source group.
It preserves complete units, neighbors, source identities, and access to the
complete original pages. Scores and selection arms are excluded from the mixed
review packet. Prior benchmark exposure still makes this development evidence.

[Small-Text, EACL 2023](https://aclanthology.org/2023.eacl-demo.11/) and its
[Python implementation](https://github.com/webis-de/small-text) offer relevant
pool-based active-learning methods. This pilot uses a local sampling protocol;
it does not install Small-Text or create pseudo-labels.
[skweak, ACL 2021](https://aclanthology.org/2021.acl-demo.40/) offers contextual
span aggregation, but its
[author repository](https://github.com/NorskRegnesentral/skweak) states that it
is no longer actively maintained. Both sources were checked on October 9, 2026.

The root assistant completed all 78 source-only reviews: 18 necessary edits,
50 retain decisions, nine optional edits, and one unresolved case. The necessary
edits comprise 12 fluency cases, three wordiness cases, two clarity cases, and
one formatting case, with 22 exact source targets. This is predominantly
additional grammar supervision, not an established gain in rhetorical-pattern
detection. The original frozen study was not refit on these reviews. The separately
registered [expanded-label and encoder-adaptation studies](encoder-adaptation.md)
are now complete; neither produced an admissible configuration.

The random control found ten necessary units among 39; the uncertainty and
disagreement queue found eight among 39. This small exposed pilot does not
establish superiority of the acquisition method. Both queues preserve meaningful
negation, algorithmic contrasts, API introductions, and community guidance.
Reviews distinguish mandatory intent decisions from optional or unresolved
ones; they do not treat every construction associated with LLMs as a defect.

The expanded annotation inventory now has 140 known necessary units and 93
known no-mandatory-edit units; 4,316 remain unknown. Four of the five previously
count-infeasible calibration folds now have at least nine positives, while
`ripgrep` still has only seven. Sufficient counts do not establish a feasible
learned operating point. A new fit requires its own frozen protocol.

The reviews keep source groups, original bytes, reasons, retained information,
and intent abstentions. All 36 pages, 4,549 units, 804 judgments, and 123 events
remain in subsequent evaluation. Annotation acquisition is not product acceptance.
