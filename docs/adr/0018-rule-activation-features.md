# ADR 0018: Collect rule activations with explicit applicability

Status: accepted for implementation.

## Decision

Extend the existing feature request set with `activation/<rule-id>`. The engine
validates the rule against its registered catalog. Requesting an activation never
enables a disabled rule or bypasses its policy. Existing block requests retain
their definitions, and collection remains absent unless explicitly requested.

Each selected activation is a block feature: the maximum raw fixed-point
activation among this rule's evidence occurrences in that block, divided by 1000.
Maximum aggregation counts a cluster once and cannot multiply identical emitted
evidence. It does not apply severity, score weights, rule/group caps, suppressions,
accepted debt, changed-unit selection, or derived gate diagnostics. It is not a
probability, a document score copied to sentences, or a quality label.

The input is validated emitted evidence before diagnostic deduplication. A
stronger repeated emission sets the maximum; adding the same emission again
cannot increase it.

An absent finding does not establish a measured zero. Add an optional observer to
the existing rule View. A rule reports each block it evaluated or found
inapplicable while doing its normal computation. The observer validates original
block IDs, states, reasons, duplicates, contradictions, and bounds; errors latch
like emitter errors and cannot be ignored to obtain a passing run. The rule API
retains its existing Evaluate and Emit signatures.

A descriptor can declare complete block observations. A successful evaluation
with that declaration must account for every extracted block. An evaluated block
with no evidence has an observed zero. An inapplicable block has no value and an
explicit reason. Existing rules without that declaration remain compatible:
positive evidence establishes an observed activation; missing observations and
missing evidence remain `applicability_unknown`, never an invented zero.

Disabled rules have absent values with `disabled`. Rules not reached after an
earlier failure remain `not_evaluated`. A failed evaluation or observer/emitter
error makes its values absent with `evaluation_failed`, including blocks with
partial evidence. Previously completed values may remain in an incomplete run;
the enclosing completion state still prevents treating it as complete model data.

Bind values to the existing source, context, policy, vocabulary, NLP and rule
identities. Preserve original extraction boundaries. Activation values do not
restore excluded prose, expose raw source, or require another extraction/NLP pass.
Limit stored observations and collected values before allocation, preserve
deterministic ordering, support cancellation, and own all returned data.

## Integration and evidence

Use the same feature collection and repeated CLI/MCP startup flags. Reporters
only render completed engine data. Save the activation formula contract and
participating rule versions explicitly; readers reject incompatible definitions
and contradictory applicability/value states. Old block-only reports remain
readable with their original contract.

Instrument the existing readability computations first: unsupported block kinds,
empty prose, and the ARI minimum word/sentence requirements must produce explicit
absence reasons; a measured value below the configured onset produces zero.
Other rule families require their own applicability accounting before their
negative observations can serve as dense training inputs. This remaining work is
part of #56 and cannot be replaced by assuming every silent rule was applicable.

Acceptance includes positive and zero activations, each absence state, disabled
rules, uninstrumented external rules, ignored observer errors, partial failures,
cluster occurrences and duplicates, policy overrides, source protection,
concurrent ownership, saved-report roundtrips, all five formats, root CLI golden
cases, and CLI/MCP self-checking. Existing diagnostic and scoring behavior must
remain unchanged. Model training and qualification remain separate requirements.
