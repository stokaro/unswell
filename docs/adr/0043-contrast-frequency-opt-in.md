# ADR 0043: keep contrast frequency opt-in

Status: accepted for default editorial policy; not detector qualification.

The `syntax.paired-contrast-density` matcher counts repeated surface forms such
as `rather than`, `instead of`, and `X, not Y`. It does not compare the
propositions carried by those alternatives. Distinct technical conditions can
therefore produce a note even when every contrast supplies needed information.
Zero score and `gate: none` keep that note out of the gate, but do not remove
its review burden from the delivered stream.

On the 36 complete exposed development pages, the unchanged frozen assistant
judgments accepted one of 30 findings, rejected 27, and left two uncertain.
Keep all 804 original judgments and 123 reference events unchanged. These are
assistant judgments under ADR 0041, not independent human annotation or an
authorship experiment. The complete before/after comparison and the accepted
finding lost from the default stream are recorded in
[the policy review](../../research/reviews/2026-10-04-contrast-defaults/README.md).

Version 3 keeps the matcher, candidate bounds, window parameters, measurements,
source mapping, severity, score, and gate. It makes the rule opt-in in every
builtin profile, with the reason carried by the existing catalog. Explicit
project configuration continues to enable it, including for shared activation
features and the repository's own CLI/MCP checks. No threshold is fitted and
no noisy diagnostic is relabeled as accepted.

The existing construction rules still diagnose their stated structures. Removing
a surface note does not establish better full-event recall or finish contextual
research. Retain the accepted loss, the complete source denominator, and all
other findings when assessing this precision decision. Broad qualification still
requires the separate combined-stream and confirmation evaluations.
