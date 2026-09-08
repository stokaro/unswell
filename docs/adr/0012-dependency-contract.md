# ADR 0012: validate dependency trees at the NLP boundary

Status: accepted for the experimental alpha; dependency provider adoption deferred.

The NLP interface has a `dependencies` capability but no representation for it.
A custom provider can currently advertise that capability and return only POS
annotations. The engine accepts the response, so an active dependency rule could
complete without the representation it requested.

## Contract

Add an optional basic dependency tree to each sentence. Every sentence token,
including punctuation, has one incoming arc in token order. A head is a
sentence-local token index; `-1` denotes the single root. Trees must be complete,
acyclic, and confined to one sentence and one uninterrupted extracted prose
region. Nonprojective edges are allowed. Enhanced graphs, empty nodes, and
multiword-token nodes are outside this representation.

The provider identity declares a dependency label scheme. Rules may require an
exact scheme in addition to the dependency capability. An empty rule scheme
means the rule uses only graph structure or handles schemes itself. Label names
are opaque: spaCy's English `nsubjpass` must not silently become UD's
`nsubj:pass`. The provider and model hashes remain part of the existing manifest
and baseline identity.

Validate requested trees before evaluating any rules. Missing trees, cycles,
invalid heads, incompatible schemes, invalid token coordinates, and edges across
protected boundaries are operational errors. Validation is linear in tokens and
checks cancellation. An absent tree is unavailable analysis, not a collection of
zero dependency features.

These additive document and identity fields do not change the default Prose
provider or existing surface rules. The builtin provider still does not advertise
dependencies. No new dependency rule is enabled by this contract.

## Candidate evaluation

The [recorded experiment](../../research/dependencies/README.md) reproduces
GoSpacy with the official `en_core_web_sm` 3.8.0 model. Its heads, labels, POS, and
byte positions match Python on 103 tokens from ten source cases, including
questionable attachments in technical sentences. These are compatibility probes,
not an independently annotated accuracy evaluation.

Retain GoSpacy as an experimental candidate. Do not expose its current high-level
loader as a product provider: it reads files lazily, logs and skips some required
component failures, lacks inference cancellation, and mutates shared state.
The isolated developer command checks required parser/tagger resources, converts
rune offsets and sentence-local heads, and uses the same public tree validator.
Its model and Python reference remain outside ordinary product execution.
Lingo's inspected distribution supplies no pretrained model, and YAP's documented
bundle targets Hebrew with separate lexicon terms. Exact revisions, licenses,
resource measurements, and the remaining adapter requirements are in the report.

The alternative of advertising dependencies from POS tags or interpreting spaCy
labels as UD would give rules a different representation from the one requested.
The builtin backend therefore remains unchanged. The contract and frozen replay
are usable by independently supplied real providers; dependency-based editorial
rules still require their own qualification before adoption.
