# Expanded annotations, encoder adaptation, and edit-span models

Three completed offline comparisons failed to produce an admissible diagnostic
configuration. Additional source annotation reduced the source-only classifier's
matching-intent byte witnesses from nine to two among 123 original events.
Attention adaptation with LoRA also supplied two, as did its control. A published
edit-span checkpoint covered seven events even when its proposals were expanded
to enclosing units. None of these witnesses establishes semantic FULL acceptance.

The [measurement record](encoder-adaptation-result.json) preserves native
completion, frozen protocols, source identity, and verification hashes. These
studies change no runtime rule, model, or gate. Published core `8a56d13f2ac0`
remains at 37/123 FULL events and 94/154 useful actual criticisms on the
[complete development replay](discourse-stance-context.md). The targets remain
at least 80% FULL recall and 85% useful actual criticisms in every cohort,
zero unsafe or unresolved advice, and separate prospective confirmation.

## Expanded source annotation

The [original source-unit study](source-student-learning.md) completed 39 head
fits on frozen ModernBERT representations. A separate acquisition pilot added
78 complete-unit reviews by the root Codex assistant under
[ADR 0041](../adr/0041-assistant-review-acceptance.md). The pilot found 18 necessary
edits, predominantly fluency problems, 50 retain decisions, nine optional edits,
and one unresolved case. All original reference events remain unchanged.

The expanded inventory has 140 known necessary units, 93 known units with no
mandatory edit, and 4,316 unknown units. One unresolved review stays masked.
Reviewed intent decisions and 77 new explanation outputs join the supervision;
the 4,549 source representations and 56 previous explanation outputs remain
byte-identical. Explanation text is an auxiliary training output and never an
inference input.

The newly registered comparison retains the original encoder, head architecture,
hyperparameters, 13 source-group folds, three calibration groups per fold, and
delivery rules. All 39 new head fits completed. The annotation bundle changes
more than sample count: it also adds explicit intent negatives and explanations.
Its comparison with the original study cannot isolate a sample-size effect.

| Fixed configuration | Available source groups | Proposed criticisms | Matching-intent byte witnesses |
| --- | ---: | ---: | ---: |
| Original source-only head | 6/13 | 1,314 | 9/123 |
| Expanded lexical baseline | 1/13 | 0 | 0/123 |
| Expanded source-only head | 5/13 | 951 | 2/123 |
| Expanded auxiliary-reason head | 5/13 | 950 | 2/123 |

Selected known-decision precision is 44/50 for the expanded source-only head
and 43/50 for the auxiliary arm. These sparse annotation proxies cannot establish
usefulness throughout the complete population or precision of the proposed
criticisms. No candidate stream received semantic acceptance credit.

## Attention adaptation

[LoRA](https://arxiv.org/abs/2106.09685) offers low-rank updates without fitting
every pretrained weight. The study used the
[PEFT Python implementation](https://huggingface.co/docs/peft/developer_guides/custom_models)
at version 0.21.2. The ModernBERT-base revision and base weights remain the same
as in the source-unit study.

Each source-group fold starts from its exact expanded-study source-only head.
The control receives 40 additional head updates. The adaptation arm receives
the same batches and head updates, plus rank-eight adapters on all 22 attention
`Wqkv` projections: 540,672 trainable adapter parameters, alpha 16, and no adapter
dropout. The 149,014,272 base encoder parameters remain frozen. No complete base
weight fit is claimed.

Both arms preserve labels, warm starts, batch order, effective batch size eight,
microbatch size two, calibration, and delivery. The head learning rate is 0.0002,
the adapter rate 0.0001, AdamW weight decay 0.01, and maximum gradient norm 1.0.
The source-only control runs on CPU and the adaptation arm on MPS. Device
arithmetic therefore differs; small score differences cannot be attributed
solely to adaptation. Both fixed configurations fail the acceptance screen.

All 26 head fits and 13 adapter fits completed. Verification confirms changed
adapter weights, exact target modules, matched warm starts and batches, all
4,549 source inputs, and unchanged evaluation scope. Native training took
851.34 seconds and peaked at 4.71 GiB of GPU memory. These fits do not train a
decoder that generates editorial reasons.

| Fixed configuration | Available source groups | Proposed criticisms | Matching-intent byte witnesses |
| --- | ---: | ---: | ---: |
| Further-updated frozen control | 5/13 | 954 | 2/123 |
| Attention LoRA | 5/13 | 973 | 2/123 |

Every matching-intent witness is still only a structural screen. Separate
partial proposals are not pooled into FULL credit. Unknown labels, unavailable
folds, and misses stay in all original denominators.

## Where the adaptation stream fails

A post hoc audit assigns each of the 123 original events to its first failed
stage. It changes no model, cutoff, or delivery rule.

| First stage | Frozen control | Attention LoRA |
| --- | ---: | ---: |
| No single source unit contains all event components | 6 | 6 |
| No admissible calibration point | 45 | 45 |
| Necessity head does not select the unit | 31 | 28 |
| Matching intent logit is negative | 37 | 40 |
| Matching intent falls outside the two delivered intents | 2 | 2 |
| Matching intent and complete byte witness delivered | 2 | 2 |

The calibration protocol requires ten selected explicitly labeled units at 85%
known-decision precision and both decision classes. More annotations made four
previously count-infeasible folds numerically possible, but eight learned folds
still lack an admissible operating point. Numerical capacity is not a successful
calibration. `ripgrep` still has only seven known necessary calibration units.

The training packet is also imbalanced across intents: it has 54 `empty_framing`
associations, 27 wordiness associations, and only one needless-complexity
association. Most new positive examples are grammar cases. Additional source
labels alone do not establish coverage of rhetorical constructions.

## Existing Python span methods

[DElIteraTeR, EMNLP 2022](https://aclanthology.org/2022.emnlp-main.678/) explicitly
learns where to edit and the edit's intent. Its
[author implementation](https://github.com/vipulraheja/iterater) supplies a
RoBERTa token classifier trained on revision data. This is closer to source
localization than classifying a whole unit. Its intents concern clarity, fluency,
coherence, and style; they do not directly identify Unswell's rhetorical categories.

The pinned author checkpoint, revision
`a1e43d5c8b414ba74e0074684437260ee68b8024`, ran once offline over all 36 original
pages. It covered 200,794 tokens in 534 overlapping windows without truncation.
Protected code remains input context and supplies no primary evidence. The fixed
decoder averages overlapping logits and uses the author's six-class argmax.
Neither a cutoff nor a category mapping was tuned. The native process completed
in 68.47 seconds; independent reconstruction verified source bytes, token coverage,
protected boundaries, delivery, and all 123 original events.

Two delivery configurations were registered before original-source inference:
minimal contiguous positive spans, and one enclosing-unit diagnosis per author
intent with the predicted spans as evidence. They produce 100 and 86 proposals.

| Delivery | Context, 46 events | Whole documents, 57 events | Historical confirmation, 20 events |
| --- | ---: | ---: | ---: |
| Minimal-span byte witness | 0/46 | 0/57 | 0/20 |
| Enclosing-unit byte witness | 1/46 | 5/57 | 1/20 |

These are optimistic byte-envelope bounds, not actual FULL recall or useful
diagnoses. They already rule out the fixed configurations at the 80% target.
The checkpoint carries CC-BY-NC-4.0 and remains a private research reference,
not an asset added to the runtime. Its model smoke probe succeeded before a
packet-preparation key error; that native failure is preserved. Source registration
was corrected without another model forward pass, and one original-source pass ran.

[GLiNER](https://arxiv.org/abs/2311.08526) and its
[Python source](https://github.com/urchade/GLiNER) offer label-conditioned spans.
An inspected published small-model configuration caps spans at 12 splitter words.
Only 59/123 original target envelopes fit that default width: 26/46 context,
23/57 whole-document, and 10/20 historical-confirmation events. This is a
representation check with no GLiNER inference or fitting, and it does not reject
the method with a different representation. Wider spans, complete windows, and
explicit unknown-label masks need a separately registered comparison. The stock
processor's unmarked spans become background, which does not match our partial
annotation contract.

[discopy](https://aclanthology.org/2021.codi-main.12/) supplies Python discourse
relations and argument spans, not edit necessity. Its
[author repository](https://github.com/rknaebel/discopy) pins substantially older
Transformers and FastAPI dependencies. It was source-inspected and not installed
into the learning environment. A structural relation cannot receive editorial
credit without distinguishing a necessary technical contrast from needless framing.

## Next learning contract

Reviewed before-and-after pairs can supply a stronger localization target.
[ERRANT, ACL 2017](https://aclanthology.org/P17-1074/) and its
[Python implementation](https://github.com/chrisjbryant/errant) extract and classify
edits from such pairs. This can reduce alignment work; extracting an edit does
not prove that it was necessary. Its grammatical taxonomy also cannot replace
the rhetorical intent review, and token positions need a verified map back to
the original UTF-8 source.

[LaserTagger, EMNLP 2019](https://aclanthology.org/D19-1510/) and
[GECToR, BEA 2020](https://aclanthology.org/2020.bea-1.16/) learn source-token edit
operations instead of predicting a whole-unit score. The
[LaserTagger Python code](https://github.com/google-research/lasertagger) and
[GECToR PyTorch code](https://github.com/grammarly/gector) offer relevant designs
for learning what to keep, remove, or change. Neither was installed or fit here.
Their published task scores are not evidence of Unswell's required recall or
usefulness on technical rhetoric.

The results motivate training on exact original-source edits and contextual
associations, with separate intent-specific retain decisions. A complete-unit
necessity label cannot teach which words form the problem or why it matters.
The source review should record editable spans, necessary intent, relevant nearby
evidence, retained technical information, and explicit abstentions. These records
must remain separate from authorship labels and from unreviewed model proposals.

Learning should distinguish source-span proposals from their contextual diagnosis,
preserve unknown losses, and keep all original source-group exclusions. Broad
style or discourse labels cannot stand in for an actionable editorial reason.
The [learning-data investigation](learning-data-methods.md) and
[source acquisition study](source-student-learning.md) retain relevant primary
work on weak supervision, explanation distillation, and active learning. SetFit,
GLiNER, and discopy are source-inspected options, not completed successful fits.

The next candidate still needs the unchanged complete-source semantic evaluation
and separate prospective pages. These negative studies establish limits of the
tested configurations; they do not establish that learned contextual diagnosis
is impossible or that the product has reached its goal.
