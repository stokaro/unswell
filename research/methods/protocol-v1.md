# Editorial comparison protocol

Protocol ID: `unswell-research-v1`. Decision date: September 8, 2026.
Authority: [ADR 0015](../../docs/adr/0015-research-methodology.md), #55 under #59.

This version fixes the design before model fitting or final evaluation. There is
no registered execution, trained model, or benchmark result yet. Dataset and
implementation hashes become concrete in a committed run manifest under #57.
Unfilled manifest fields prevent that run from starting; they are not defaults
that a runner may infer after seeing results.

## Targets and observations

The primary task predicts `needs_revision` against `acceptable` under the
[editorial rubric](../../docs/editorial-annotation.md) and a frozen profile.
Exclude `uncertain`, missing responses, inadequate context, and unresolved
adjudications from binary fitting and metrics. Publish their counts and reasons
in the acquisition denominator, with separate error review. Never convert them
to negative examples to improve measured performance.

Sentence, paragraph, and fragment are distinct targets. Qualify models per kind
and role; paragraph qualification supplies no sentence probabilities. Report
documentation, README, API references, comments, release notes, and string roles
separately. Origin, actor identity, generator, repository, and author language
metadata are for grouping and audit, never editorial feature inputs.

The separate origin task first compares documented `human` and `generated` units.
Keep `human_ai_edited`, `generated_human_edited`, `mixed`, and `unknown` outside
that binary fit. Evaluate the known editing classes separately. An origin
estimate is disabled by default and cannot fail a gate by itself. A feature
derived from origin research enters the editorial model only through its own
quality-label fit and ablation.

## Data, context, and partitioning

Use at least 5,000 human-labeled units, with at least two independent humans and
at least 1,000 final-test units, as required by #21–#22. Keep original submissions,
packet hashes, adjudication history, permissions, and source identities. An LLM
judge may supply auxiliary observations; it cannot fill a human annotation slot.
Include acceptable AI-assisted text, poor human text, and permitted examples of
technical English from non-native writers. Unknown provenance stays unknown.

Use the [existing planner](../annotation/corpus/README.md) before extraction.
Connect repositories, documents, authors, templates, original/rewritten versions,
generation tasks, and exact duplicates. Curators review unrecorded near duplicates
without detector scores. Each component belongs to one partition. Prefer the
highest independent grouping available; disclose when repository independence
cannot be established. Existing inspected Ptah material remains development-only.

Target shares are training 50%, development 15%, calibration 15%, and final test
20%, using seed `unswell-research-v1`. These are acquisition targets, not promised
outputs of hashing connected components. Pin whole components as needed before
labels are exposed. Record achieved counts and increase acquisition when minimums
fail; never split a related group or count overlapping units as independent.

Apply the same frozen extraction policy and target/context bytes to each compared
method. The primary comparison uses target-only model input; raters retain the
declared context needed for judgment. Units that require unavailable model context
must be counted as an applicability limit. A separate context-enabled experiment
may use the same recorded coherent block for every method, with its own manifest
and evaluation. Never compare a full-document score to a short-comment score as
if they had the same input. Keep protected regions and unrelated comments apart.

Fit vocabularies, IDF, normalizers, deduplication parameters, and feature selection
only on training. Create derived examples only after the source group is assigned.
Development selects models and procedures. Calibration fits the selected
calibration and selects operating thresholds. Final test supplies one locked
measurement. No refitting on development or calibration after selection in v1.

Reserve additional unseen-generator, domain, time, and repository cohorts where
permitted data supports them. Their group IDs and intended uses must precede
analysis. External origin benchmarks remain separate from editorial qualification.

## Comparisons A–G

| ID | Method | Question |
| --- | --- | --- |
| A | Existing rules and editorial index | What does the current product find? |
| B | Regularized logistic model on shared rule activations | Does learning weights help without new NLP features? |
| C | Character/word n-grams with a linear classifier | What does a simple lexical model add? |
| D | Structure, lexical diversity, repetition, and POS with logistic regression | Does linguistic structure help beyond A–C? |
| E | D features with a small tree ensemble | Does nonlinearity justify extra cost? |
| F | LLMDet origin scores, plus a separate quality fit using its features | Is there origin signal, and does it improve editorial recall? |
| G | Fast-DetectGPT; Binoculars if the budget permits | What changes with model-based reference computation? |

A pins Unswell commit `a412b43d86923702ac2c1e9f2d70c8ccc0f6081e` and
`builtin:technical-v1`. Record its resolved policy hash and rule versions in the
run manifest. Preserve actual policy outcomes; also report index operating curves
without relabeling them as the default gate. B uses underlying rule activations,
not threshold findings or annotation answers. E–G cannot delay the B/D Go baseline.

B–D use an unpenalized intercept and L2-regularized logistic loss. The training
implementation in #23 must publish its exact loss scaling, optimizer, stopping
rule, missing-value handling, and feature contract before a run. The regularization
grid is `0.0001, 0.001, 0.01, 0.1, 1, 10`. Stable sorted feature and example order
is mandatory; worker count must not change an artifact in a pinned environment.

C compares character 3–5-grams and word 1–2-grams, separately and combined, with
training-only TF-IDF. Bound each vocabulary to 50,000 entries; ties use byte order.
The contract in #56 specifies Unicode, case, token boundaries, and exclusions
before fitting. D uses the versioned shared structure, diversity, repetition,
POS, and rule-signal families. No author-origin metadata or duplicate NLP formulas.

Each B–E family gets at most 24 predeclared configurations, five grouped
development folds, and two CPU-hours per configuration on the declared host.
List exact configurations and fold IDs before fitting. Record timeouts as failed
trials, rather than granting the best-looking method extra tuning. Use seed 17
for stochastic choices; seeds 29 and 43 are fixed stability runs, not a search for
a better winner. Deterministic models still record the seed and ordering policy.

Select the best simple comparator from A–C on development at the same coverage
and false-positive constraint. D is the primary candidate; E and F are secondary
hypotheses. Break ties by lower resource use, then stable configuration ID.
Freeze the winner and comparator before final labels are opened. Publish all
trials, including failures. Do not select a new winner on final test.

G receives a fixed paired subset selected by group and ID before any predictions,
with at most 1,000 targets and 24 GPU-hours per method on a named device. The
manifest records subset hashes, exact model/tokenizer revisions, precision, and
token limits. Record a truncated or unsupported target as inapplicable; silent
prefix scoring cannot stand in for its full-target result. E–G may be deferred
with a reason. Published paper numbers never substitute for an unexecuted trial.

## Calibration and applicability

Choose between logistic recalibration and isotonic calibration using five grouped
development folds of the frozen model's predictions, selecting lower mean Brier
score; break ties in favor of logistic recalibration. Fit that chosen calibrator
on calibration only. Select the highest-recall threshold satisfying the declared
FPR constraint there, then freeze it before final evaluation. No final-test bin,
length, threshold, or calibration adjustment is allowed.

Record training and target class prevalence, any class weights, normalization,
raw score, classifier transform, and calibration identity separately. Compare
against a constant probability estimated from training prevalence. A sigmoid of
the heuristic index does not qualify. Calibration must be measured independently
of ranking performance; see [Guo et al.](https://proceedings.mlr.press/v70/guo17a.html).

Report length bins of 1–7, 8–19, 20–49, 50–99, 100–249, and at least 250 words,
using the shared tokenizer identity. These are reporting bins, not reliability
claims. Choose supported lengths and roles on development, then freeze them.
Missing capabilities, incompatible contracts, unknown domains, resource limits,
low LLMDet coverage, and short targets require explicit reasons. Language metadata
and its verification limits belong in the manifest; Latin script is not proof
of English.

Report availability, disabled, inapplicable, and operational-error states. Null
does not mean zero risk. Rules can still evaluate an unsupported statistical
target. If a policy requires an estimate, missing coverage cannot silently pass;
#24 retains the existing distinction between policy failure and incomplete work.

## Metrics and decision rule

Report revision precision/recall, clean-block FPR, counts by defect, and the share
of clean files or PRs receiving any false block. Micro metrics count all eligible
units; group-macro metrics give each independent source component equal weight.
Publish both, with class counts, group counts, missing labels, and abstentions.

For calibrated outputs publish Brier score, the constant baseline, and reliability
tables/plots. ECE uses ten fixed equal-width bins on [0, 1], left-inclusive and
right-exclusive except the final bin. Weight each bin's absolute difference
between mean prediction and observed positive rate by its sample share. Empty
bins have zero weight and display no estimated rate. The roadmap target is
ECE at most 0.05; this is not a measured result.

Version 1 adopts a minimum five-percentage-point recall gain over the selected
simple comparator at clean-block FPR at most 1%, on identical supported targets.
Require the gain and FPR limits for both micro and group-macro estimates. Report
95% uncertainty intervals; the primary recall difference interval must exclude
zero. Require the one-sided 95% upper FPR bound to meet 1% before qualifying a
blocking default. A small or dependent sample that cannot support that bound
leaves qualification open, even when it contains no observed false positives.

Resample the highest independent source groups with replacement, retaining all
their units, paired predictions, and labels. Use 10,000 bootstrap replicates,
seed 17, and percentile intervals; report invalid replicates with a missing
class. Inadequate independent groups make an interval insufficient evidence.
Predeclare any stratification in the manifest, preserving original grouping.
When zero false positives produce a degenerate bootstrap, do not publish zero
as an upper risk bound. Report a one-sided exact binomial bound for the chance
that an independent clean group has any false flag. This conservatively bounds
group-macro block FPR, not micro FPR; micro qualification needs an appropriate
independent confirmation sample or remains unproven in v1. Define that sample
before labels are exposed: one randomly chosen eligible target per independently
sampled source group, using seed 17, with its own declared target population.
An exact binomial bound applies only when those Bernoulli observations are
independent samples from that population. It cannot silently certify a differently
weighted full-corpus micro rate. Preserve that scope in the qualification claim.

Report coverage over the full eligible flow and performance on the common covered
subset together. Refusing difficult targets cannot alone establish an improvement.
Compare missed positives and false flags over the full flow, counting abstained
positives as not detected, without inventing probabilities for them. No new
default may reduce coverage relative to its comparator within the claimed scope.

Predeclare role, length, domain, origin, and permitted author-language slices.
For each supported slice, require the upper 95% bound on recall loss to be below
five percentage points and the FPR bound to meet the same 1% limit. Otherwise
restrict the qualification claim or collect more data. Sparse slices stay
unqualified; an average gain cannot conceal a known subgroup regression.

Only D versus the chosen simple comparator is confirmatory in the first run.
Treat other comparisons and feature ablations as exploratory. A later product
choice based on them needs an independent confirmation cohort. Report per-family
ablation of rules, length/readability, lexical, POS, repetition, and LLMDet inputs.
Remove expensive families without useful additional evidence in the next protocol.

Origin metrics are separate: TPR at fixed FPR, precision at declared prevalence,
ROC-AUC/PR-AUC, calibration, and abstention coverage. Where independent negatives
are scarce, do not claim extreme FPR. For example, zero errors in 1,000 independent
negatives gives an approximate one-sided 95% upper bound of 0.3%, not 0.01%.

## Meaning, resources, and run records

Use blinded before/after editorial judgments plus the existing meaning worksheet
for the draft → CLI/MCP → external revision → rescan workflow. Retain negation,
obligations, quantities, terms, and conditions. A lower index alone does not prove
improvement. Include valid repeated API descriptions and lexical matches with
different facts. External agents and research model calls stay outside ordinary CI.

Measure full scan, cold start, loading, extraction/NLP, features, inference, and
reporting separately. The ordinary product target remains 100,000 words in ten
seconds and 512 MiB on a described 2-vCPU Linux host. Include tokenizers, tables,
and temporary buffers in peak memory. Record binary/model-pack sizes. #28 must
set their delivery budgets before qualification. Reference GPU costs are separate
and cannot satisfy the product budget. This protocol claims no measured costs.

A committed run manifest must include these concrete values:

- Protocol/registry revisions and hashes; task, rubric, profile, and allowed uses.
- Source/packet/label hashes, independent group IDs, partition counts and hashes,
  context policy, exposure history, and final-test custodian/access record.
- Feature, extraction, NLP, vocabulary, model, and calibration identities;
  configurations, fold membership, seeds, threshold procedure, and compute limits.
- Units/roles/length support, primary and exploratory comparisons, confidence
  methods, missing-value policy, and the exact selection and rejection criteria.
- Code commit, Go toolchain, OS/architecture, hardware, resource accounting,
  commands, and allowed numerical tolerance across platforms.

Before final scoring, add hashes of selected artifacts, support rules, fixed
thresholds, comparator, and the locked evaluation command. Keep all raw predictions
and statuses, including failed trials; derive tables and plots mechanically.
Publish permitted data, acquisition instructions, limitations, a model card, and
the adoption ADR. Model-free behavior remains until the accepted evidence exists.

Any amendment increments the protocol version, states the reason and exposure,
and identifies affected runs. Editing after final-test exposure requires new
independent confirmation. Negative scientific results are valid; absent human
data, failed calibration, or missing measurements do not close product acceptance.
