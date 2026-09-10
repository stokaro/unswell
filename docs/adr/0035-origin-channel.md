# ADR 0035: Report origin estimates in a separate, ungated channel

Status: accepted for the channel and its contract; qualification, incremental
value, and any change to the default remain open.

## Decision

Add #58's origin channel beside the revision-probability channel rather than
inside it. Both load the same explicitly supplied pack format, and a pack
declares which of the two targets it estimates. The engine refuses a pack in the
wrong channel: an origin pack cannot answer whether a unit needs editing, and a
revision pack cannot answer where prose came from.

The channel is off unless a policy asks for it. `origin.model: pack` plus
`Options.OriginModel`, `check --origin-model`, or `unswell-mcp --origin-model`
turn it on, and the policy records `origin` only when it does, so a run without
it keeps its existing hashes, reports, and accepted debt. An unconfigured
channel reports nothing at all: the assessment fields are absent rather than
carrying a status, which keeps model-free results byte for byte identical.

An origin estimate cannot decide a build. The origin policy has no threshold to
configure, no gate consults the estimate, and accepted debt records only the
revision pack, because an origin estimate changes no finding, no index, and no
gate decision. A test asserts that a run with both channels and a required
probability gate still passes on the origin channel alone.

What the estimate means stays narrow. It is a similarity to a defined training
class, reported per unit for the single unit kind its pack qualifies. It is not
a quality judgment, it is not evidence that a unit needs revision, and it is
never a share of a text written by any tool. A pack estimates one unit kind, so
a document-level number is never attached to a sentence, and no span is invented
for a unit the pack does not qualify.

Abstention follows the existing contract: a unit below the pack's declared
minimum words, a unit of another kind, an unavailable column, a score outside
the fitted calibration range, or a source whose measurement inputs differ from
the pack each produce an explicit status and no value. An explicitly requested
pack that is corrupt or in the wrong channel fails at construction, and
`origin.on_incompatible: fail` turns an incompatible source into an operational
failure.

## Acceptance and follow-up

Every writer carries the same decisions. JSON and SARIF serialize the assessment
fields and the manifest; the text, Markdown and HTML writers name the configured
pack, how many units of its kind it estimated, and that the estimate judges no
quality and decides no gate.

Confirmed endpoint labels, mixed and edited cases reported separately, measured
incremental value before any origin feature enters editorial training, and the
product decision that would change the default all remain open under #58 and
#57. Nothing here qualifies an origin model, and a configured pack is not
evidence that its classes describe real authorship.
