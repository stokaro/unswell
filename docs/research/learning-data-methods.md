# Learning editorial decisions from assistant annotations

Training in Python can produce a model for an offline Go runtime. The immediate
research question is whether better supervision can teach where an edit is
needed, its purpose, and what information must survive. Classifying authorship,
detecting familiar phrases, and selecting already generated criticisms are
different tasks.

At published core commit `8a56d13f2ac0432dc1704e12453c4973290ff3a9`, the
[complete development replay](discourse-stance-context.md) has 37/123 FULL events
and 94/154 useful actual criticisms. The requirements remain at least 80% FULL
events and 85% useful deliveries in **each** cohort, with no unsafe or unresolved
supplied advice. This investigation changes neither those denominators nor the
published diagnostics.

## Relevant research and Python implementations

The following are primary papers and their authors' implementations, checked on
October 8, 2026. Their reported results concern their own tasks; none qualifies
Unswell on technical documentation.

| Approach | Evidence and implementation | Use in Unswell | Limitation |
| --- | --- | --- | --- |
| Find the span and edit intent | [DElIteraTeR, EMNLP 2022](https://aclanthology.org/2022.emnlp-main.678/) and [Python code and checkpoints](https://github.com/vipulraheja/iterater#deliterater-emnlp-2022) | Learn localization and purpose from source text, with surrounding sentences. The paper compares token intent detection, edit detection, and context settings. | A proposed clarity or style edit does not establish that editing is necessary. The paper's strongest overall span detector uses multiple sentences and a single task; multiple heads are a hypothesis to test, not a guaranteed improvement. |
| Distill labels and explanations | [Distilling Step-by-Step, ACL 2023](https://aclanthology.org/2023.findings-acl.507/) and [Python implementation](https://github.com/google-research/distilling-step-by-step) | Train explanation prediction as an auxiliary **output** task, alongside the decision. At inference, the source alone supplies the input. | This differs from appending a generated criticism's rationale to a selection model's input. Results on reasoning and classification benchmarks do not establish editorial performance. |
| Combine weak annotation sources | [Snorkel, VLDB 2017](https://pmc.ncbi.nlm.nih.gov/articles/PMC5951191/) and [Snorkel](https://github.com/snorkel-team/snorkel) | Let rules and annotators emit a label or abstain; inspect coverage and conflicts before fitting a source-text model. | Repeated prompts, variants of one complaint, and annotators with shared biases are not independent votes. The current [LabelModel implementation](https://github.com/snorkel-team/snorkel/blob/main/snorkel/labeling/model/label_model.py) assumes conditional independence; it does not automatically repair arbitrary correlated sources. |
| Audit label disagreement | [Confident Learning](https://arxiv.org/abs/1911.00068) and [cleanlab](https://github.com/cleanlab/cleanlab) | Rank disputed examples using predictions made outside the example's training project. Review the source before changing training labels. | A classifier can confidently disagree with a correct label. Uncalibrated scores and subjective class boundaries limit the interpretation. |
| Learn with few labels | [Efficient Few-Shot Learning Without Prompts](https://arxiv.org/abs/2209.11055) and [SetFit](https://github.com/huggingface/setfit) | Contrastive fine-tuning and a small classification head provide a useful baseline for source-plus-context decisions. | Text classification alone supplies neither precise spans nor preservation guarantees. Keep all examples from one source project and all derived pairs in the same split. |
| Train a classifier on LLM annotations | [Promises and Pitfalls of LLM Annotations, NAACL 2025](https://aclanthology.org/2025.findings-naacl.75/) and [annotation pipeline](https://github.com/Media-Bias-Group/llm-annotations-annomatic) | A direct precedent for generating annotations with several LLMs and training a smaller classifier. | The media-bias experiment's trained classifier also loses robustness on invariance tests. Its gains cannot be transferred to our task by assumption. |
| Match teaching complexity to the student | [Small Models Struggle to Learn from Strong Reasoners, ACL 2025](https://aclanthology.org/2025.findings-acl.1301/) | Compare concise structured explanations against longer explanations while holding data, encoder, and decision task fixed. | The evidence concerns reasoning tasks. It motivates an ablation, not a conclusion that shorter editorial rationales always work better. |
| Test editing behavior separately | [EditEval, CoNLL 2024](https://aclanthology.org/2024.conll-1.7/) and [evaluation code](https://github.com/facebookresearch/EditEval); [CheckList, ACL 2020](https://aclanthology.org/2020.acl-main.442/) and [behavioral tests](https://github.com/marcotcr/checklist) | Measure specific editing skills and test whether irrelevant changes affect a decision. | Similarity to a reference rewrite does not prove necessity, safety, or complete event detection. Behavioral controls supplement the unchanged complete-document evaluation. |

The current PyPI versions observed were cleanlab 2.9.0, Snorkel 0.10.0, and
SetFit 1.2.0. Only cleanlab was installed and executed in this investigation;
Snorkel and SetFit remain candidates. Context7 supplied current cleanlab and
SetFit documentation; Snorkel was absent from its search index, so its official
documentation and source were inspected directly.

## What earlier Unswell experiments already tested

The completed TETRA auxiliary-transfer configuration was rejected. On the six
context pages, its fixed cutoff produced 4/46 FULL events, below its earlier
7/46 baseline. Even its optimistic post hoc cutoff sweep did not meet the
unchanged recall and usefulness requirements. Professional revision preferences
are useful auxiliary information, but they are not automatic labels for
mandatory editing or technical preservation.

The October 3 contextual cross-encoder selection screen also terminated without
an admissible operating point. Its record retains all 36 pages, 4,549 units,
804 judgments and 123 events; it did not establish FULL-event performance.

The October 6 rationale-preservation experiment appended a diagnostician's
explanation to its public prediction input. It trained a criticism-selection
head, not a source-text model that predicts explanations as a separate output
task. Its cached whole-document result was 34/57 FULL events and 867/1,149
useful deliveries, with five unresolved suggestions. It failed acceptance.
The results justify investigating a different supervision contract rather than
repeating that configuration or calling it a completed Step-by-Step replication.

The corresponding local records are retained under the named TETRA warmup,
contextual cross-encoder selection, and rationale-preservation studies. These
aggregate results do not imply public reproduction of the private judgments.

## Executed cleanlab audit

The [measurement record](learning-data-methods-result.json) preserves the
registration, input, script, dependency, review, and completion hashes. The
audit used the existing rationale-preservation experiment's cached outer-project
scores. It verified source byte slices, labels, unique prediction ownership,
and the exclusion of the tested project from both training and calibration.

| Audit item | Result |
| --- | ---: |
| Verified outer projects | 13 |
| Cached predictor-input identities | 8,291 |
| Actual criticism aliases | 3,566 |
| Scored unique criticisms | 3,520 |
| Scored judgments: accepted / uncertain / rejected | 2,566 / 888 / 66 |
| Missing scored aliases | 0 |
| Model disagreements with frozen judgments | 1,029 |
| cleanlab-ranked review candidates | 840 |
| Ranked candidates: accepted / uncertain / rejected | 421 / 379 / 40 |
| First ranked criticisms reviewed against complete sources | 12 |
| Distinct source targets in that review | 7 |
| Confirmed label errors in that review | 0 |
| Changed frozen labels, removed training rows, new fits, provider calls | 0 |

Softmax was applied to the saved three-class logits. These are uncalibrated,
class-weighted model estimates, not product probabilities. The cleanlab call
used `filter_by="both"`, self-confidence ranking, and one worker. The completed
audit took 35.12 seconds with a 284,295,168-byte peak resident set.

The first 12 criticisms concern an established selection idiom, architectural
responsibilities, algorithm continuation, an explanation of sample-output
omissions, optional environment-variable capabilities, and algorithm iteration.
The complete source supports retaining their frozen judgments. Several are
variants of the same source target. They remain separate criticism records;
they are not 12 independent language examples.

Review was performed by the root Codex assistant under
[ADR 0041](../adr/0041-assistant-review-acceptance.md). It is not independent human
annotation. The deliberately selected top-ranked sample estimates neither
population label noise nor the full queue's error rate. The other 828 ranked
criticisms remain unreviewed. No automatic correction, pruning, or accuracy gain
is justified by this audit.

## Source-annotation seed

The completed seed copies the 123 unchanged reference events into 129 exact
source spans and adds seven explicit retention anchors from the reviewed
criticisms. Each span was checked against the original UTF-8 bytes and mapped
to the engine's units. The source groups, event components, and annotation
lineage remain attached.

These annotations overlap 128 of the 4,549 original units. The other **4,421
units remain unlabeled**, not negative examples. Positive events cover 13
projects; the seven retention anchors cover only four. The seed is neither
complete nor balanced, and no model was fit from it.

A retention anchor means that the specific criticized construction should
remain. It does not declare its sentence or document flawless. Reference-event
reasons and review judgments are annotation metadata, not prediction inputs.
The measurement record preserves the seed hash and project counts without
publishing the private annotation packet.

## Subsequent executed comparison

The [October 9 source-unit comparison](source-student-learning.md) trains 39
heads on an expanded source-annotation packet while keeping every original
event. Auxiliary outputs predict explanation embeddings. The fixed diagnostic
streams fail the unchanged recall screen; a decision-precision proxy and the
next annotation queue are not recorded as detection gains.

## Learning design

The next packet should teach editorial necessity from source text, then compare
supervision choices. This is a design, not a claim that a new model has been
trained or qualified.

1. Preserve raw UTF-8 source, engine units, protected spans, context, source
   group, and all components of an event. A missing criticism is **unlabeled**;
   it is not evidence that an entire unit is clean.
2. Have the assistant annotator identify location, intent, necessity, a concise
   reason, and information to retain. Keep `necessary`, `optional`, `retain`,
   and `unresolved` distinct. Review dispositions on generated complaints are
   a separate label space. A negative requires an explicit source review.
3. Include difficult retention examples: necessary contrasts, capability
   statements, algorithm transitions, architectural role introductions, output
   scope explanations, and literal control loops. Root assistant review is
   sufficient; retain the actual reviewer identity and the checks performed.
4. Select additional annotation work using disagreement, uncovered units,
   underrepresented projects, and missing event components. Do not sample only
   examples the current classifier already finds easy. Record correlated
   annotation sources and abstentions before any Snorkel experiment.
5. Compare a plain source/context classifier with a span-and-intent model and
   auxiliary explanation supervision. SetFit is a small-data baseline, not the
   required architecture. The student must not receive the evaluation judgment
   or the teacher's explanation as an inference input. Concise and full
   explanations form separate registered arms.
6. Freeze source identities, annotation packet, explicit negatives, project
   splits, model and tokenizer hashes, hyperparameters, resource limits, and
   failure policy before fitting. Group revisions, paraphrases, duplicate
   source spans, and generated complaint variants with their original project.
   A prose fact retained by a repair does not become a negative quality label.
7. Evaluate delivered criticisms and every unchanged original event component,
   including misses and unavailable outputs. Report FULL recall, usefulness,
   and preservation separately for all three cohorts. Keep the 80%/85%/zero
   unsafe-advice requirements, then use separate prospective confirmation.

The present 804 judgments assess generated criticisms; they do not provide
exhaustive positive and negative labels for all 4,549 source units. A selector
also cannot create a diagnostic that no candidate stream proposes. These are
the reasons to improve annotation coverage and the learning target before
changing another classifier head.

Python libraries remain research tools. Any candidate intended for ordinary
analysis still needs export and numerical parity with the pure-Go runtime,
complete source localization, resource measurements, and CLI/MCP/WASM checks.
Experimental LLM review remains separate. This investigation sent no source
documents to providers, started no new model requests, and changed no gate.
The full detection objective remains open in
[#349](https://github.com/stokaro/unswell/issues/349).

The [expanded-label, attention-adaptation, and pretrained edit-span comparisons](encoder-adaptation.md)
are complete. The record includes the negative results, all-event failure-stage
accounting, and primary work on learning token edits from reviewed text pairs.
None changes the published runtime or achieves product acceptance.

The separately registered [reviewed source-span study](reviewed-source-spans.md)
now completes the original/repair reviews and 26 matched unit/span head fits.
It records historical supervision conflicts, unavailable calibration groups,
and zero byte-covering witnesses among all 123 original events. Further primary
papers and Python implementations cover revision-necessity prediction,
automatically labeled edit intents, and contextual connective preferences.
