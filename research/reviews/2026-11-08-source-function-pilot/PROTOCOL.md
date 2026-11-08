# #349 source-function pilot, preregistered local packet

Status: source-only review frozen before candidate generation. This is an
exploratory pilot on four repository-disjoint public README pages, not a
qualification run on the 36 exposed development pages or the reserved
confirmation pages. The source revisions and exact bytes are in `sources/`.

The reviewer is the Codex assistant in this session on November 8, 2026,
using the complete page without candidate findings. The eight events in
`labels.json` are judgments about independently useful editorial corrections,
not AI-authorship labels or human annotations. Uncertain candidate diagnoses
remain nonaccepted in the delivered-finding denominator. A later correction
to a label must be a dated amendment with the old label retained.

Hypothesis: a complete-source model asked to identify the function of each
claim (definition, measurable capability, qualification, transition, or
unsupported promotion) and to apply a minimal-deletion test can find source-
specific omissions without inheriting the archived candidate bank's broad
lexical operators. It should preserve supported technical statements even
when they contain adjectives. The model receives no frozen labels or reviewer
rationales. It must quote exact UTF-8 source substrings, give a concrete
diagnosis and source-specific evidence, and omit optional style preferences.

Protocol before execution:

- Route: existing Codex subscription, `gpt-6-astra`, medium reasoning, one
  ephemeral tool-free CLI process for each complete source page.
- Budget: at most four sequential calls, at most one attempt per page, at most
  10 minutes per call and two hours total. Stop on the first CLI, schema,
  source-binding, trace, or resource error; no retry or response repair.
- The exact four request files, response schema, runner, scorer, source
  identities, and frozen labels must be SHA-256 pinned in `freeze.json`
  before the first model call. Any packet change creates a new version.
- A valid answer may return zero findings. Maximum 20 findings per page;
  exceeding the limit is an invalid response, not a truncation. Exact quote
  occurrences must be unique or disambiguated by the returned original source
  line. The binder computes byte offsets from the source; the model does not
  calculate them. Reject protected code, identifiers, and technical distinctions.
- No replacement text is requested in this diagnostic-only screen. Report
  suggestion safety as not applicable, with zero supplied edits. A later
  advice-producing treatment needs separate safety review.
- Scoring: an event is fully recalled only if at least one delivered finding
  identifies its specific defect at its original source location. Report
  events found / 8 overall and per page, accepted actual diagnoses / all
  delivered diagnoses, abstentions, invalid pages, and whole-page coverage.
  Source review of the *entire* combined stream precedes any success claim.
  Delivery on these pages is directional only. An exploratory advance
  requires at least 7/8 events, at least 85% accepted findings, no invalid
  page, and no protected technical counterexample flagged. This does not
  replace the original 80% full-event and 85% accepted-delivery gates on
  each required complete cohort, followed by separate confirmation.

Source-only counterexamples to preserve: Caddy's scale statement supplies
request and certificate counts; restic's cryptography and restore claims
describe concrete behavior; Gitea's platform list names actual platforms;
the OpenTelemetry Collector's OTLP version and Go compatibility rules
specify exact versions and conditions. These are not defects merely because
they contain adjectives or claims.
