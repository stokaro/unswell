# Editorial annotation rubric

Rubric ID: `unswell-editorial-v1`. Artifact format: `unswell-annotation-v1`.
This is the protocol for [#21](https://github.com/stokaro/unswell/issues/21), under
the [research umbrella](https://github.com/stokaro/unswell/issues/59).

Unswell aims to keep formulaic AI-style wording out of code and documents. The
annotation target is the need for editorial revision under a stated policy.
Origin is a separate research variable. A useful AI-assisted explanation may need
no changes; a human draft may contain empty framing or repetition.

## Decision and scope

Answer this question: does the target need a wording change because it contains
one of the defects below, while retaining its substantive meaning?

| Label | Decision |
| --- | --- |
| `acceptable` | No required wording change under this rubric and profile. Another phrasing may also work. |
| `needs_revision` | A specific defect requires a change. Record its category and explain it using the target and available context. |
| `uncertain` | Context is insufficient, the rubric does not resolve the case, or the distinction remains unclear. Record why. |

These labels concern editorial wording. Spelling, punctuation, factual correctness,
and compliance with an external contract are separate questions. A typo alone is
not a positive label. If a project bans a phrase, Unswell may report it even when
another profile considers the text acceptable.

Distinguish a necessary change from an optional alternative or personal preference.
Length, passive voice, formality, caution, repeated API names, and technical terms
do not establish a defect by themselves. Do not infer a writer's English fluency
or origin from their prose. Record language-background groups only when suitable
metadata is available and its use is permitted.

## Categories and boundaries

The examples below teach the decision boundary. They are not independent human
annotations or measured rule precision. The executable tutorial includes each
category and a corresponding acceptable example.

| ID | Candidate needing revision | Acceptable counterpart | Boundary |
| --- | --- | --- | --- |
| `empty_framing` | `It is important to note that the client may retry.` | `The client may retry.` | The opening merely announces the statement. Retain a necessary warning, attribution, or contractual quotation. |
| `needless_repetition` | `The timeout is 30 seconds. The timeout is 30 seconds.` | `The timeout is 30 seconds. The retry interval is 60 seconds.` | The repeated claim adds no distinction. Similar words can describe different conditions, objects, or values. |
| `wordiness` | `Perform an evaluation of the cache before use.` | `Evaluate the cache before use.` | The instruction needs the action. A named evaluation stage or artifact may require the noun. |
| `unjustified_intensifiers` | `This incredibly powerful and amazingly effective cache stores pages.` | `The cache stores pages for 30 seconds.` | Praise adds no usable information here. Preserve calibrated uncertainty and supported comparisons. |
| `vague_claims` | `The service offers unparalleled performance.` | `The service handled 500 requests per second in this test.` | The claim lacks a scope or measurement. Assess wording and available support, not the truth of an unseen benchmark. |
| `formulaic_transitions` | `Furthermore, the cache stores pages. Moreover, it keeps headers. Additionally, it keeps tags.` | `The cache stores pages, headers, and tags.` | Repeated transitions add no relationship in this list. Retain transitions that express a necessary contrast or dependency. |
| `needless_complexity` | `The execution of the initiation of the reload procedure occurs after validation.` | `Reload begins after validation.` | The extra nesting conveys no additional condition here. Preserve named stages, exact terminology, and necessary qualifications. |

Choose all supported categories, without assigning the same defect several times
to imply greater severity. `needs_revision` requires at least one category;
`acceptable` carries none. An uncertain answer may record plausible categories,
but those are not confirmed defects for category-agreement statistics.

Technical counterexamples require attention:

- `may retry` and `must retry` impose different obligations.
- `enabled` and `disabled` differ despite lexical similarity.
- Two versions, numeric limits, or failure conditions can require similar wording.
- An exact technical term may need repetition to avoid ambiguity.
- A passive sentence may correctly emphasize the object or omit an unknown actor.
- An API table may repeat a template because the rows describe different members.
- `Package document defines source coordinates and the neutral prose model.` and
  `Emit latches validation failures even when a custom rule ignores the error.`
  are valid comments. A POS tagger's noun-stack finding is not a quality label.

## Annotation units and context

| Kind | Target | Limits |
| --- | --- | --- |
| `sentence` | One sentence from one eligible prose block | Judge that sentence with the supplied context. Do not inherit its document's label. |
| `paragraph` | One paragraph or coherent extracted comment block | Judge its internal organization and repetition. A paragraph probability cannot supply sentence probabilities. |
| `fragment` | A coherent source fragment, such as a short string or list item | Record this kind explicitly. It is not a sentence solely to satisfy a model's minimum length. |

Input roles are `documentation`, `readme`, `api_reference`, `doc_comment`, `comment`,
`release_note`, and `string`. Record prose language separately from source syntax.
Version 1 handles English prose only. Source syntax uses the existing
`document.Format` values, including Markdown, code languages, shells, and YAML.

Acquire and group source documents before extracting units. Use Unswell's existing
source-preserving extractor and NLP sentence boundaries. Store its identity,
policy hash, and context selection policy. Keep the source hash and every original
half-open UTF-8 segment; an enclosing span alone cannot represent separated prose.
Excluded code, directives, and other protected content remain excluded.

Do not concatenate unrelated comments to reach a length threshold. If a protected
boundary prevents one coherent target, create separate units. Show any necessary
neighboring material as context with its role clear. Curators must check context
selection without consulting detector results. The same target and context must
reach all independent raters in the round and later match the training protocol.

If context is inadequate, choose `uncertain` and `context: insufficient`. Do not
invent the missing condition. Requests for more context create a new recorded
round with a new context snapshot; retain the original answers.

## Independent review

1. A curator records the source, permissions, grouping, extraction identity,
   profile, and opaque unit IDs. Keep the original administrative artifact private
   to the study team unless publication has been approved.
2. Prepare a packet using the [Go command](../research/annotation/README.md).
   It includes only the target, context, kind, role, and frozen profile. It omits
   origin, author/model identifiers, source paths, detector output, and all answers.
3. Each of at least two human raters reads the same rubric and packet, works
   independently, and records the label, context adequacy, categories, rationale,
   and timestamp. Do not reveal other answers before submissions are locked.
   Each response includes the packet's SHA-256, binding it to the exact target,
   context, and policy shown to that rater.
4. Compute agreement on these original answers. Keep missing responses distinct
   from `uncertain`. An assistant judge belongs in an auxiliary record and cannot
   fill a required human slot.
5. Discuss disagreements only after independent submission. Record a later
   adjudication, its reviewers, final label, and rationale. Preserve every original
   answer. An unresolved case remains `uncertain`; consensus is not mandatory.

Adjudication echoes the same packet hash. Changing target or context requires a
new packet and fresh responses; do not attach old answers by matching short IDs.

Use opaque actor IDs with the identity key stored separately. In version 1, every
primary rater in a round is assigned every unit in that round. An adjudicator who
did not independently label the packet has role `adjudicator`, not `rater`.
For different assignment groups, create separate rounds and preserve their IDs.
An original rater can participate in adjudication after independent submission;
the protocol does not require a third person when two raters can resolve the case.

The [response instructions](../research/annotation/templates/response.md) describe
the exact submission shape. Preserve submitted files before importing records into
the administrative round; never distribute that round as an answer form.

Blinding removes administrative clues, not clues inherent in the target. Do not
rewrite a quotation or remove an AI self-reference to hide its likely origin.
The curator checks that profile instructions and context contain only information
needed for the editorial task. Record unavoidable blinding limitations in the
study report. Opaque IDs are not a promise of anonymity against outside knowledge.

An actor's `human` declaration is administrative metadata, not proof of a person's
participation. Keep consent, assignment, and submission records for audit. The
validator cannot establish identity, independence, permission, or good faith.

## Before and after review

Treat revisions as related source versions, with distinct hashes and unit IDs in
the same split group. Blind the initial editorial ratings of both versions where
possible; use a separate paired review to check preservation of meaning.

Use the [meaning-review worksheet](../research/annotation/templates/meaning-review.md).
Check negation, modal obligations, conditions and exceptions, quantities and units,
versions, identifiers, attribution, and technical requirements. Record each retained,
changed, lost, or unresolved item and the responsible reviewer. Shortening a warning
by deleting its failure condition is not an improvement.

The annotation command neither rewrites prose nor decides semantic equivalence.
The corpus task stores both source versions, transformation history, and the
completed worksheet. A lower Unswell index alone cannot certify an improvement.

## Provenance and permissions

Store provenance separately from quality:

| Label | Required basis |
| --- | --- |
| `human` | A documented basis for human authorship of this unit; absence of AI-use records is insufficient. |
| `generated` | A generation record tied to this unit. |
| `human_ai_edited` | The human original and a recorded AI editing step. |
| `generated_human_edited` | The generated original and a recorded human revision. |
| `mixed` | Documented mixed contributions that cannot be reduced to one of the preceding cases. |
| `unknown` | No adequate unit-level record. This remains usable for editorial annotation. |

The administrative record keeps the scope of evidence. A repository-level claim
cannot establish a unit-level origin label. In particular, the maintainer's report
that `stokaro/ptah` is almost entirely generated supports a repository-level note;
it does not label every fragment or establish its quality. Use `unknown` for
individual units until suitable evidence exists. Never run the detector to create
its own origin ground truth.

A generation record should include the available model identifier and version,
date, parameters, prompt/source relationships, and later transformations. Mark
unavailable fields as unknown in that record. Do not invent a model revision or
publish prompts and private source material automatically.

For each unit retain a source or permitted reference, source hash, document and
repository IDs, template/author grouping when known, and a related-version group.
An empty template or author-group field means unknown, not an independent source.
Record the license, permission evidence, and reviewed allowed uses separately:
annotation, training, evaluation, source redistribution, and annotation
redistribution. No permission is implied by an omitted use. Packet creation
requires recorded annotation permission.

The command records permission decisions; it does not interpret licenses. Curators
verify terms and privacy requirements before each use. A source-code license does
not establish permissions for every linked model, document, or dataset. Publish
only the permitted subset plus manifests, acquisition instructions, and honest
limitations. Private materials stay private even when the format validates.

## Sampling, pilot, and frozen partitions

The [corpus task](https://github.com/stokaro/unswell/issues/22) must obtain at least
5,000 labeled units, at least two independent human annotators per unit,
and at least 1,000 final held-out units. Also publish independent document,
repository, template-family, and class counts. Adjacent sentences are not
independent observations for sampling uncertainty.

Build a sampling matrix across quality and provenance. It must contain acceptable
AI-assisted material and human-authored material needing revision, in addition to
the other combinations. Unknown origin is an explicit group, not the human class.
Include technical English from non-native writers where supported by permitted
metadata, short comments, dense terminology, API templates, instructions, warnings,
and prose mixed with code. Keep difficult examples and report missing strata.

Run a pilot before large-scale annotation. Cover each category, its exceptions,
the unit kinds, and the intended roles. Obtain real independent human responses,
publish agreement and disagreements, revise unclear boundaries, and record the
rubric decision. Pilot examples used to improve the rubric become development
data, never a final test. Freeze the rubric and profile before corpus annotation;
changes require a new version and recorded reannotation decisions.

Use four partitions: training, development, calibration, and final test. Before
splitting, connect document versions, shared templates, prompts and source texts,
paraphrases, translations, expansions, reductions, and before/after versions.
Each connected group belongs wholly to one partition. Use repository or author
holdouts when the data supports them and state the achieved independence level.

The split manifest records group membership, sizes, hashes, seed, code revision,
and the split algorithm. Do not learn vocabulary, scaling, feature selection, or
deduplication parameters from the final test. Choose algorithms and thresholds on
development data; fit calibration on its own partition. Opening final-test results
and then changing the model consumes that test for development and requires new
independent confirmation.

Plan separate checks for new generators, new repositories/domains, and later data.
These are distinct from randomly withheld fragments of known documents. Source
acquisition, leakage validation, pilot execution, and split assignment remain in
#22 and the comparison harness; the current command does not claim to perform them.

## Agreement and uncertainty

The [agreement implementation](../research/annotation/README.md#agreement-statistics)
computes raw pair agreement and nominal Krippendorff alpha on original primary
judgments. Treat `uncertain` as a third quality category. Skip missing answers
when computing agreement. Units need at least two primary answers to contribute
to alpha.
Singleton responses still appear in coverage and uncertainty counts.

Report the number of raters, units, rated and paired units, missing answers, label
counts, and uncertain share. Report unit kinds and input roles separately. For
reason categories, compare selected/absent membership among non-uncertain answers;
show how many uncertain answers were excluded. Do not turn an uncertain answer
into a confident negative for every defect.

Alpha is unavailable when there are no paired ratings or no expected disagreement.
Return `null` with a reason. Retain negative alpha values. All-agreeing constant
labels do not establish a useful classifier or justify a fabricated alpha of one.
Agreement statistics do not measure label correctness or model precision.

The command currently returns point estimates and `sampling_intervals: not_estimated`.
For corpus publication, report sampling intervals using document/repository-group
resampling under a frozen protocol. Preserve all ratings from a sampled group,
declare the group level, seed, replicate count, interval method, and undefined
replicates. Insufficient independent groups means insufficient interval evidence.
The comparison harness owns that computation; do not substitute a confidence
interval that treats neighboring sentences as independent.

After the pilot, preregister any numerical agreement criteria and uncertain-case
tolerance used to qualify a collection round. The command deliberately has no
automatic agreement pass threshold. Publish the original and adjudicated label
distributions, the share left uncertain, and category disagreements even when a
collection decision is unfavorable. No results from the scripted tutorial count
toward the pilot, corpus size, model calibration, or stable-rule qualification.
