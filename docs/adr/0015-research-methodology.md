# ADR 0015: separate editorial qualification from origin research

Status: accepted for protocol version 1; no statistical model is qualified.

## Decision

Unswell exists to reduce formulaic AI-style wording in code and documentation.
Train the editorial model to predict `needs_revision` under
`unswell-editorial-v1`, using independent human judgments. Good AI-assisted prose
can be acceptable; human prose can need revision. Origin metadata is never an
editorial input or a substitute for a quality judgment.

Keep four results separate:

| Result | Meaning | Policy |
| --- | --- | --- |
| Rule finding | A stated rule condition matched | Existing severity and gate settings apply |
| Editorial index | Capped contributions from the scoring model | Existing local thresholds apply |
| Revision probability | Calibrated estimate of the rubric label in a qualified population | Available only after #23–#25 acceptance |
| Origin estimate | Similarity to specified provenance classes in an evaluated corpus | Experimental, opt-in, and unable to block CI by itself |

An index is not a percentage. A binary origin probability cannot measure the
fraction of words written by AI. A document classifier cannot locate a defective
word or assign its result to every sentence. Keep exact findings, heuristic
evidence, and statistical estimates distinguishable in explanations.

The [versioned protocol](../../research/methods/protocol-v1.md) adopts the proposed
minimum recall improvement of five percentage points at a clean-block false
positive rate of at most 1%. It also requires uncertainty, coverage, subgroup,
calibration, and resource evidence. Adoption here fixes a research selection rule;
it asserts no measured property of Unswell. Rule qualification retains the
separate roadmap targets of 98% precision for hard defaults and 85% for soft ones.

The [method registry](../../research/methods/registry-v1.json) records the sources
reviewed, pinned code and resource identities, component terms, costs, limits,
and scoped adoption decisions. Reviewing, reproducing, porting, and qualifying
have independent evidence records. Port compatibility does not establish useful
editorial performance. A negative experiment can finish a research task while
leaving model qualification open.

## Architecture and implementation boundaries

Use the existing `document`, `extract`, `nlp`, `rule`, `config`, and `report`
packages and public engine. #56 defines the shared feature API before adding
public interfaces. #23 owns Go training and calibration; #57 owns comparative
execution. No second extractor, NLP provider, index, or reporter belongs in the
research implementation.

The feature contract must distinguish unavailable values from zero, require real
dependency capabilities where needed, and identify normalization, vocabulary,
NLP, extraction, profile, and model compatibility. Preserve protected boundaries
and original source segments. Share appropriate computations between rules and
models, without changing input buffers or exposing mutable shared state.

Ordinary analysis remains local Go with `CGO_ENABLED=0`. It runs no Python,
Node.js, child process, API, or model server and downloads no assets. Primary
editorial training and calibration also run on Go. Python is allowed only in
isolated reference experiments and production of port comparison data, outside
ordinary builds, tests, and scans. Existing product tests need no GPU.

Load explicit local model packs at the application boundary. The engine receives
data and policy; it does not discover models through the filesystem, environment,
or network. Inference treats model packs as bounded data, never executable code,
pickle objects, or plugins. Reporters and MCP consume the same completed result.

Preserve model-free rules and index behavior, including `calibration.model: none`.
A missing estimate has a reason; it is never serialized as zero. An explicitly
required model that is absent, corrupt, incompatible, or unable to cover required
input cannot silently pass. #24 will define configuration and result additions
with documented compatibility. Existing public fields, reports, and exit codes
remain unchanged until that implementation is accepted.

Keep local nondilution, correlation caps, and independent severity/gate behavior.
Do not score a derived threshold finding again or count a model's aggregate and
its underlying signals twice without a documented policy. Baselines and
suppressions preserve raw estimates; accepted debt is not a clean training label.
Changed-unit checks retain block context and trusted base-policy identities.

## Alternatives and consequences

Using origin as the editorial target would confuse acceptable AI assistance with
poor wording. Training a second prose pipeline would let research and shipped
diagnostics diverge. Importing a model-based reference runtime would violate the
offline Go contract. These approaches are excluded from this decision.

A regularized logistic model is the required initial learned baseline. Stylistic
features, trees, and LLMDet must show additional value on the same task. Expensive
references can remain unevaluated when resources or terms prevent execution;
their published accuracy cannot fill an empty result cell. This keeps #23 moving
without making all tokenizers or a GPU experiment prerequisites.

The current Ptah material provides 378 unlabeled development candidates from
one repository group. It supplies workflow evidence only. The human corpus,
held-out data, executable comparisons, trained artifacts, calibrated estimates,
and product acceptance remain outstanding in #22–#30 and #50–#58.

## Acceptance and change control

This ADR and its protocol satisfy the design work in #55 under #59. They do not
close the implementation or evaluation tasks. Commit a reviewed run manifest
before fitting candidate models, then freeze selected models and thresholds
before exposing final labels. A protocol amendment records its reason, changed
fields, and affected data exposure. Once test results influence a choice, that
test becomes development data and independent confirmation requires a new test.
