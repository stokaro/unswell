# LLM-associated pattern protocol

Protocol ID: `unswell-llm-patterns-v1`. Decision date: September 10, 2026.
Authority: [ADR 0036](../../docs/adr/0036-llm-pattern-evidence.md) for
[#154](https://github.com/stokaro/unswell/issues/154).

Protocol version 1 is the development-pilot protocol. It fixes cohorts, roles,
the historical boundary, the controlled experiment, partitions, measures,
statistics, decision rules, budget caps, and stage outputs before anyone
acquires a corpus. The pilot permits exploration. Version 2 of this protocol
freezes the confirmatory hypotheses, the sample size, the minimums, and the
exact comparison list, and the team commits it before confirmatory data
exists. A committed protocol is an internal record, not an external
preregistration.

The editorial protocol [`unswell-research-v1`](protocol-v1.md) is unchanged.
This protocol produces no editorial label, no origin estimate, and no
probability. Every artifact it produces keeps `human_corpus: not_qualified`.

## Claims and non-claims

The branch measures two things: whether a rule recognizes a formally defined
construction, and whether that construction occurs at a different rate in the
studied LLM cohorts than in the historical cohort. Reader benefit from removing
a construction is a separate human judgment under #22 and is never inferred.

An observed rate in the historical cohort is a baseline prevalence. It is not an
editorial false-positive rate. The share of LLM outputs that carry a finding is
an association, not the recall of unwanted patterns. Without a separate record
of whether each unit contains the pattern, precision and recall are unknown and
the report says so.

## Results v0 and v1

The branch delivers two results with separate completion criteria.

| Result | Content | Completes when |
| --- | --- | --- |
| v0 | The historical corpus of stage B, the E1 baseline map, and this executable comparison protocol | Stage B artifacts verify and E1 tables reproduce from saved records |
| v1 | The confirmatory study of stages C and D on comparable, permitted LLM data | Protocol version 2 has run once on the confirmation partition and stage D artifacts reproduce |

When comparable LLM data is unavailable, v0 completes on its own. The team keeps
the artifacts, and every comparison that did not run carries the status
`unrun` with its reason. That outcome closes nothing in v1, blocks no other
roadmap work, and never becomes a project-wide blocker or a CI failure.

## Literature

The [sources record](llm-patterns-sources-v1.json) pins the reviewed version,
target, unit of analysis, data terms, method, limitations, and decision for
each study. It was verified on September 10, 2026 from the listed pages.
Publisher pages for PNAS and Science Advances returned HTTP 403, so journal
metadata for those two studies comes from the arXiv listing and Europe PMC.

| Study | Pinned version | Decision and scope |
| --- | --- | --- |
| Reinhart et al., PNAS 2025 | arXiv:2410.16107v2 (August 21, 2025); PNAS 122, e2422455122 | Adapt: participial clauses, nominalizations, subject `that` clauses, phrasal coordination, agentless passives, and instruction-tuned lexical overuse become candidate constructions; no data reuse, availability unknown |
| Kobak et al., Science Advances 2025 | arXiv:2406.07016v5 (July 3, 2025); Sci. Adv. 11(27), CC BY | Adapt: the excess-frequency method against a projected trend is the temporal check; population shifts identify no document; biomedical data not reused |
| RAID, ACL 2024 | 2024.acl-long.674 | Adopt the robustness dimensions already recorded in the [registry](README.md); dataset use stays deferred until its terms are read |
| DetectRL-X, ACL 2026 | 2026.acl-long.1773 | Adopt the polish, expand, and condense operations and the length dimension; dataset use stays deferred because its stated terms conflict |
| Liang et al., 2023 | arXiv:2304.02819v3 (July 10, 2023); published version in Patterns not fetched | Adopt as the language-background limitation check; no background is inferred from prose or names |
| Rallapalli et al., 2026 | arXiv:2604.14111v1 (April 15, 2026); preprint, no venue listed | Adapt: role is stratified before source, one generator family is held out, and chat variants of one model are one family; publication status is rechecked before protocol version 2 |

None of the six studies reports a result on code comments or technical
documentation. Reinhart, Kobak, and Liang work on general registers,
biomedical abstracts, and essays. The domains of RAID and DetectRL-X and the
genres of Rallapalli et al. were not read; the sources record says so, and the
team reads those papers in full before protocol version 2. No number from any
study is an Unswell result. Novelty claims in the final report must compare
against these six studies and the registry entries.

The statistical methods below rest on references that the sources record
lists as bibliographic entries. Efron and Tibshirani, An Introduction to the
Bootstrap, 1993, covers percentile intervals. Davison and Hinkley, Bootstrap
Methods and Their Application, 1997, covers resampling by cluster. Cameron,
Gelbach, and Miller, 2008, covers behavior with few clusters. Clopper and
Pearson, 1934, gives the exact binomial bound. Benjamini and Hochberg, 1995,
gives false-discovery control, and Holm, 1979, gives the confirmatory
correction.

## Cohorts and roles

| Cohort | ID | Purpose | Admissible evidence |
| --- | --- | --- | --- |
| Historical technical English | `historical` | Baseline prevalence and diversity | Snapshot bytes with a recorded confidence level in their date; never a human label |
| Controlled LLM outputs | `controlled` | Association under known conditions | Saved request, response, model identity, and the source task |
| Natural AI-heavy projects | `natural` | Applicability to real workflows | Repository-level statements at the level actually confirmed |
| Contemporary technical English | `contemporary` | Prevalence and distribution shift | `mixed` or `unknown` unless independent evidence exists |

Roles are `documentation`, `readme`, `api_reference`, `doc_comment`, and
`comment` in the primary analysis. `release_note` and `string` are separate
layers. The manifest's `error_message`, `log_message`, `ui_text`, and
`other_string` roles stay apart from strings meant for readers, and an
undetermined role stays unknown. Prose language, programming-language
ecosystem, topic, and length are separate dimensions in every table.

### Historical boundary

The boundary dates are design decisions of this protocol. The primary
historical cohort `H0` contains text whose public appearance is evidenced on
or before December 31, 2020. That date precedes the GitHub Copilot technical
preview of June 2021 and the open availability of the GPT-3 API in November
2021. `H0` therefore predates the tools that first put LLM text into code
repositories at scale. It is conservative because the mass use that followed
the public ChatGPT release began almost two years later. The sensitivity
cohort `H1` ends on December 31, 2018, before the GPT-2 release of February
2019. The transition period from January 1, 2021 to November 29, 2022 is
reported separately and never merged into the baseline. The contemporary
cohort starts on November 30, 2022, the day of the public ChatGPT release.

Each source records its date confidence: `corroborated` when a commit date
agrees with an independently published release or archive, `vcs_only` when
only version-control metadata exists, and `unknown` otherwise. `H0` and `H1`
admit `corroborated` sources only. `vcs_only` sources form the sensitivity
cohort `H0+vcs`, and `unknown` sources enter no historical cohort. Tables
report unit and group counts for each. A timestamp alone never proves
authorship or the time of public appearance.

### Sampling frame

The frame covers public repositories with a permissive license that allows
analysis, English documentation, and an available history, in the ecosystems
the extractor supports. Source selection looks at license, technical
diversity, and history availability, never at how human the text reads. The
planning target is tens of thousands of unique units, with 50,000 to 100,000
available from dozens of independent source groups. That target is a resource
volume, not a sufficiency criterion.

One repository contributes at most 10% of the units of a cohort, and one
ecosystem at most 40%. Seeded random subsampling of whole documents enforces
each cap before any measurement; the record keeps the seed, the dropped
document IDs, and the uncapped totals. Predefined rules exclude vendored and
generated directories, generated API pages, license and boilerplate text,
translations, and files marked as templates, and the pipeline counts every
exclusion. No rule excludes text because Unswell fires on it.

A unit counts once per content hash across snapshots, copies, and identical
document versions. A sentence inside a counted paragraph is a nested
observation of that paragraph, never a second independent unit. Full snapshots
and newly added or changed text are separate analyses; the newly-added
analysis counts a unit at its first evidenced appearance, so an unchanged
paragraph never becomes a new observation in a later snapshot.

### Natural corpora

Ptah stays a development case study because it was inspected while building
Unswell. Its component keeps the explicit `development` assignment of
[ADR 0014](../../docs/adr/0014-corpus-acquisition.md). Its maintainer's
statement of heavy AI use is repository-level provenance and never a
per-fragment label. No contemporary repository receives a human or AI label
from a detector, from the absence of a disclosure, or from style. Unknown
provenance supports a prevalence measurement and nothing more.

## Controlled experiment

The pilot uses one task type: documentation for a unit of code or a structured
fact sheet drawn from the historical cohort. Seeded random sampling,
stratified by role and ecosystem, picks the tasks: pilot tasks from the
training and development partitions, and confirmatory tasks from the
confirmation partition after the version 2 freeze. Every sampling record is
saved. A deterministic, versioned extractor builds each fact sheet from
signatures, identifiers, parameter lists, and numbers. It uses no model and
copies no sentence of the original, and its version and output hash enter
every response record.

Two operations are in scope for the pilot. `generate` writes the text from the
fact sheet without seeing the original wording. `polish` edits the original
text, and the pair is kept. `expand` and `condense` follow the DetectRL-X
taxonomy in the sources record and use the opening lines in
[prompts/README.md](prompts/README.md); they enter when budget allows.

Two prompt conditions are fixed before the pilot and stored verbatim under
`research/methods/prompts/`: `neutral`, an ordinary request for the text, and
`plain`, an ordinary request for concise plain technical English. Neither
prompt asks for typical AI text, names a construction under study, or asks the
model to avoid AI patterns. Prompt condition is a stratification factor with
two fixed levels; it has no holdout axis in protocol version 1.

A generator family is a distinct pretraining lineage from a distinct
organization. Versions of one model, including its chat variants, are one
family. At least two families take part in the pilot; version 2 requires at
least three, with one family held out from feature selection. Generation for the
held-out family happens only after the version 2 freeze. If it happens
earlier, a sealed status blocks measurement until that freeze, and the seal
is part of the exposure history. Each family runs with
one recorded decoding setting, the provider default, for every task. The
budget authorization record names the families.

The requested length for `generate` and `polish` equals the original's word
count with a 30% tolerance. `condense` requests 50% and `expand` 150% of the
original count, fixed here so nobody tunes them later. The generation tooling
must keep refusals, errors, truncations, and off-length responses with their
status, and it must never regenerate a response for style or length. The
primary analysis at the requested budget includes every response with status
`complete`, whatever its realized length. Refused and truncated responses
count in coverage and in the share-of-documents denominator as `no text`.
Excluding off-length responses is a listed sensitivity analysis, never the
primary analysis. A transport or rate error permits one retry after a recorded
delay; a content outcome permits none. An Unswell score never selects or
filters a response.

Each response record holds every visible input message, the model identifier
or fingerprint, the date, the parameters, and the task, operation, and prompt
IDs. It also holds the raw response, each transformation with its version, the
status, the cost, and SHA-256 hashes of input and output. It holds no secret.
Linguistic input excludes the dataset wrapper and any assistant label, and
target text goes through the same versioned transformations as historical
text.

Comparability checks cover identifiers, numbers, signatures, and required
elements between a response and its source. They report coverage and failures
and never claim semantic equivalence. The overlap statistic is the share of
response tokens inside word 8-gram matches with the source; a response above
0.5 stays in the primary analysis and leaves only in a listed sensitivity
analysis. When a model's training set is unknown, the record marks possible
contamination by public training data as unknown.

## Budget

No paid budget is authorized on the decision date, so the money cap is zero
and no tooling may make a paid call. When a maintainer records an amendment
that names the amount and the authorizer, the pilot runs within these caps,
and a lower cap in the amendment wins.

| Cap | Pilot value |
| --- | --- |
| Tasks | 200 |
| Requests | tasks x operations x prompts x families, at most 2,400 including retries |
| Input tokens per request | 8,000 |
| Output tokens per request | 4,000 |
| Total tokens per family | 9,600,000 (800 requests x 12,000 tokens) |
| Retries | one per request, transport errors only, counted as requests |
| Wall-clock time | eight hours per family |
| Money | the authorized amount; zero without an authorization record |

Dry run, response caching, resumption, and a stop at any cap are mandatory.
Local or open-weight models run through the same research tooling outside the
product and its ordinary build and tests; they count against time and carry
the same identity fields. Reproduction of an analysis reads saved outputs and
never repeats generation.

## Partitions and independence

The existing planner assigns each connected component to one of the four
partitions. In this branch `training` is the discovery partition,
`development` selects features and rules, `calibration` is the selection
partition for the confirmatory list and any operating threshold, and
`final_test` is the confirmation partition. No probability calibration is
required by the name. This branch uses the editorial seed
`unswell-research-v1` and the editorial weights of 50%, 15%, 15%, and 20%, so
a component that both protocols touch lands in the same partition under both.
The editorial final test is therefore preserved by construction, and the
exposure history records every viewing on top of that.

Provenance links connect a source with every derivative: rewrites, related
versions, forks, exact and near duplicates, shared text templates, and known
author links. The planner today connects repository, document, author,
template, related-version, generation-task, and exact-hash keys. Forks and
near duplicates therefore need a curated related-version or template key, and
a versioned near-duplicate detector with a threshold fixed before extraction
is a stage B adapter. A shared generator or a shared prompt is a design
factor, never a provenance link, and it never merges the corpus into one
component. A controlled response inherits the component of its source task.

Generator family has its own holdout axis. Protocol version 2 checks four
combinations: new sources with studied families, new sources with the
held-out family, known ecosystems on new projects, and new roles or topics
where support exists. Threshold tuning on a studied family never counts as an
unseen-family result.

Discovery data alone fits vocabulary, feature selection, normalization,
deduplication parameters, and compression references. The selection partition
chooses the confirmatory list and any operating threshold. The confirmation
partition is measured once per protocol version, and the run hash of that
measurement is recorded. Several LLM versions of one source form one linked
group. Every artifact records the number of independent groups, the largest
components, and the developer exposure history of each group. A group with
viewed final results cannot confirm a changed rule.

## Measures

Each candidate construction declares its exact condition, scope, applicability,
denominator, unit kind, and required NLP capabilities before measurement.
Lexical, POS, dependency, and discourse features are distinct; an absent
backend reports unavailability, never a heuristic substitute.

A `document` is one documentation file or section for `documentation`,
`readme`, and `api_reference`; one doc comment block for `doc_comment`; one
comment block for `comment`; one release note entry for `release_note`; and
one string literal for the string roles. An applicable unit is a unit of the
kind the construction declares, within a role the construction admits.

| Measure | Definition |
| --- | --- |
| Prevalence | Share of applicable units with the construction, per cohort, and where declared per 1,000 words or per opportunity |
| Absolute difference `D` | Prevalence difference between two arms, with its interval |
| Prevalence ratio | Ratio of the two prevalences; `undefined` when either count is zero |
| Paired change | Difference within an original and its edited response, by operation and family |
| Warning load | Findings per document and per 1,000 words, and the share of documents with at least one finding, for each rule and for the whole profile |
| Coverage | Units analyzed, unavailable, excluded, and failed, with reasons |
| Cost | CPU time, peak memory, wall-clock time, and resource size per stage, from the existing cost measurement |

Tables show micro and group-macro values, the number of sources, groups, units,
and words, and the counts of gaps and exclusions. They break effects down by
family, operation, role, and source group, and they report length both at the
requested budget and in a separate length-standardized analysis.

### Primary estimand

Every confirmatory hypothesis fills one template before protocol version 2:
construction ID, unit kind, arm A as cohort, operation, prompt, and family,
arm B, role stratum, estimand (`D` or paired change), and the declared
direction. The primary comparison for a construction is `D` between
`controlled` outputs of `generate` under `neutral` and `H0`, within the
declared role stratum, reported per family. The paired change under `polish`
is the secondary estimand for the same construction. Other cohorts,
operations, and prompt levels are exploratory unless the template names them.

## Statistics

The unit of independence is the provenance component in every cohort; a
controlled response belongs to the component of its source task. Paired
responses and repeated generations are linked observations inside that
component. Intervals come from a cluster bootstrap that resamples components
jointly, so a component's historical units and all of its responses enter or
leave a replicate together, with 10,000 replicates, seed 17, and percentile
bounds. Where an arm has fewer than 30 components, the report adds a BCa
interval and a coverage check on a simulated reference. No hierarchical model
is part of the pilot: the expected number of components per arm is too small
for stable variance components, and the research module has no random-effects
implementation.

A zero count in either arm makes the construction `inconclusive`. The report
shows that arm's prevalence with a one-sided 97.5% Clopper-Pearson upper bound
computed on the number of components that could carry the construction, and
the ratio cell reads `undefined`. This bound is a stage B adapter with its own
reference calculation.

Exploratory screening on the development partition controls the false
discovery rate at 0.10 with the Benjamini-Hochberg procedure. Protocol version
2 freezes the confirmatory list: at most twelve hypotheses, one per
construction, each with the single primary comparison from its template. Holm
correction at 0.05 applies across the whole list. A construction that looks
good on the confirmation partition cannot join the list afterward.

The pilot fixes these design thresholds:

- Minimum useful difference `MID`: three percentage points in document-level
  prevalence of the declared unit kind.
- Acceptable baseline load for a default rule: one finding per 1,000
  historical words.
- Support: the construction occurs in at least five components per arm.
- Cluster minimum: at least 20 components per arm.

The 95% cluster interval of `D` decides the state of an evidence card:

- `supported-within-scope` when the lower bound exceeds zero, the point
  estimate reaches `MID`, support holds, and the baseline load is acceptable.
- `restricted` when the rule above holds in some declared roles or families
  and fails in others; the card names the supported scope.
- `unsupported` when the upper bound is below `MID`, including an interval
  around zero. A difference that excludes zero and stays below `MID` is
  `unsupported`.
- `inconclusive` when the interval spans both zero and `MID`, the cluster
  minimum fails, or either arm has a zero count.

These numbers are design decisions. Version 2 restates them with the pilot's
observed frequencies and within-group dependence. It fixes the sample size
and the stopping rule. It records the minimums for independent groups, roles,
comparable tasks, and unit and word volume. No sample grows after that to
reach significance.

The confounders under check are role, topic, length, project, ecosystem, era,
and prompt or model family. Sensitivity analyses vary the historical boundary,
the cap on large projects, near-duplicate exclusion, extraction failures, and
known templates. Nobody infers a language background from a name or from
prose; a subgroup without confirmed metadata stays an open limitation.

## Mandatory experiments

| ID | Experiment | Output |
| --- | --- | --- |
| E1 | Baseline map: the pinned commit, profile, and rule versions measured on every cohort, including rules that do not separate cohorts | Per-rule prevalence and load tables |
| E2 | Controlled comparison by family and operation, then deferred confirmation on new source groups | Paired change tables; every unrun family or operation listed as `untested` |
| E3 | Added value: current rules, the frozen phrase list baseline, and structural candidates on identical inputs, with lexical, grammatical, and repetition ablations | Comparison tables; no authorship classifier is required |
| E4 | Historical and temporal check across periods and snapshots with composition controls and placebo comparisons inside the historical period | Period tables; a temporal change is never reported as a causal LLM effect |
| E5 | Linter behavior on formally defined minimal pairs, counterexamples, and a formatting-only control | Regression fixtures with expected spans and capabilities |

E1 takes the rule classes as input. Before measuring any cohort, stage B
publishes the pinned commit's rule list with a class for each rule: general
style, explicit prohibition, or LLM-associated candidate. The list names each
rule's applicable roles, and its hash enters the dataset manifest. A class
change is an amendment.

E3's phrase list is the pinned commit's phrase catalog plus the Reinhart and
Kobak word lists in the sources record. Version 2 records its hash before the
confirmation partition is measured.

E4's placebo comparisons use pseudo-boundaries at December 31, 2012 and
December 31, 2016 inside `H0` with the same composition controls. The report
shows the post-2022 change against the distribution of placebo changes.

E5's counterexample classes are necessary technical repetition, a concrete
comparison in place of an empty contrast, code and identifiers, quotations,
and word matches without the required construction.

A text from which a construction was removed must stop triggering its rule;
that is correct linter behavior. Texts with inserted cliches or a prompt to
remove AI patterns are interventions, stored apart from the prevalence sample.
A lower score is never evidence of preserved meaning.

## Evidence cards and rule decisions

Each studied construction gets one evidence card. The card holds the
definition, the implementing rule or feature with its version, corpus and run
hashes, and the original hypothesis. It also holds prevalence and intervals
per cohort, holdout results, permitted roles, limitations, examples and
counterexamples, cost, and the decision. The states are
`supported-within-scope`, `restricted`, `unsupported`, and `inconclusive`,
decided by the rules above. An association never makes one occurrence an
error. The branch keeps a useful general rule that fails to separate cohorts,
and a strong marker may still be unfit as a prohibition in a given context.

A product rule that cites a card also carries regression tests. Severity and
gate policy are set apart from effect size. A profile or default-gate change
is a separate product decision, and public API and saved-report semantics do
not change through this branch.

## Implementation and reproducibility

Stage B first inventories the corpus, feature, training, and evaluation
functions on main and records the result. The existing limits on sources,
units, and artifact bytes stay in force. A large corpus lives in shards under
one dataset manifest, and grouping, deduplication, and partition checks run
across all shards. A group may span shards and never spans partitions. Reruns
and parallel runs produce identical samples and no duplicates.

The existing research module gains adapters for acquire, verify, plan,
extract, generate or import, measure, analyze, and reproduce-report. Model
calls and acquisition stay in research tooling. Every adapter has tests for
hashes, schemas, permitted state transitions, leak absence, feature
unavailability, analysis determinism, cancellation, and partial failure.
Ordinary CI checks every adapter on frozen fixtures without network access. A
separate reproducible run with a published command and published hashes
checks the full corpus. Known numerical examples and small independent
reference calculations check every statistical routine. Results from
different pipeline versions never merge without a recorded compatibility
check.

## Stage outputs

| Stage | Verifiable artifact |
| --- | --- |
| A | ADR 0036, this protocol with its budget caps, the sources record, the pilot prompts, and #22 on hold |
| B | Dataset manifest, source and snapshot records with dates and confidence, rule classes, global plan, shards that verify, and E1 tables |
| C | Saved permitted generation records, pilot tables, sample-size record, and protocol version 2 |
| D | E2 to E5 results with intervals, holdouts, ablations, and negative results |
| E | Versioned data set, data card, research report, table reproduction, and evidence cards with decisions |

A plan does not close data collection. A collection without measurements does
not close validation. Results without accessible records do not close
reproducibility. Under a resource limit, completed stages stay completed and
each unrun comparison carries a status and a reason. The epic completes when
the frozen protocol has run and the permitted artifacts are public, including
null results, whether or not Unswell beats the baseline.

## Amendments

An amendment increments the version, records its date, states the reason,
lists the changed fields, and records which partitions the team had viewed at
that time. After the team has viewed the selection or confirmation partition,
a threshold or a minimum moves upward only. A downward change requires new
independent groups that nobody has measured, and the report shows both the
frozen and the amended value. A change to a confirmatory hypothesis after
viewing the confirmation partition requires new independent groups. The
budget section changes only through a recorded authorization.
