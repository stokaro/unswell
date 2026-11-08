# #349 source inventory preparation

Status: offline preparation. No model calls or detector quality gains have been
measured for this hypothesis.

## Motivation

The preceding four-page source-function pilot returned 16 accepted diagnoses
out of 17 but fully recalled only 3 of 8 frozen events. Missed cases included
unsupported comparisons, an absolute behavior claim, a performance tautology,
and a malformed screenshot lead. Most output concerned grammar. A threshold
change cannot recover events absent from the model's response.

## Source contract

`reviewtext --source-inventory` uses the existing technical-v1 extraction and
NLP unit preparation. Inputs pin full UTF-8 source bytes and their SHA-256
hashes. Output pins the input hash and every eligible unit's page-local ordinal,
binding, prepared text, original byte segments, and raw source envelope. A
sentence is exported where the unit API can select one. Protected boundaries
produce fragments, which must be accounted for separately. Source envelopes
can contain Markdown syntax outside the eligible segments. The model must
never treat such syntax or excluded code as an editable prose target.

The command does not claim semantic completeness: extraction policy can omit
quotes or structure, and the sentence API cannot join prose across a protected
boundary. A later scorer must count every eligible exported unit and report
excluded material separately. Repeated text is distinguished by its source
hash, ordinal, and original segments.

## Next experiment

Select new repository-disjoint pages before any model call. Freeze sources,
source-only assistant annotations, accepted counterexamples, prompt, schema,
model, effort, call budget, failure policy, and scoring code. Require exactly
one inventory record per exported unit ID and reject missing or extra IDs.
Then make a separate semantic diagnosis pass over the full inventory. Grammar
only changes are not credited to the semantic treatment. Report delivered
calls, valid and abstained pages, full-event recall, accepted diagnosis fraction,
and edit safety when edits exist. A small pilot only screens this mechanism;
the 36-page/123-event development target and separate prospective confirmation
remain required for #349.
