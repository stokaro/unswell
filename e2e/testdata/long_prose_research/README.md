# Complete-document research examples

The generated examples are the first two `filler.announced-importance` findings
in Qwen's complete `generate/neutral` arm, ordered by source path and byte offset.
`provenance.json` binds them to the frozen requests, output hashes, and contiguous
source ranges of study `2026-09-15-long-qwen`. The full study lives under
`research/generation/studies/long-prose-v3`. Source licenses are retained here.

The agent-written alternatives remove only the announced-importance prefaces and
capitalize the resulting sentence openings. The benchmark still states 3.423 and
13.031 seconds and that ugrep is unaffected. The maintenance paragraph retains
both file names, volunteer delivery, no fixed release dates, the possibility of a
regression patch, and the absence of a timing promise. The shared engine finds
the original prefaces and no preface in either alternative.

The Haiku example is the first `syntax.nominalization-chain` finding in its
complete `generate/neutral` arm under the same ordering. The edit replaces
`conducts an investigation of` with `investigates`. Every other word remains,
including acknowledgment, acceptance or rejection, the rejection explanation,
the vulnerability-policy condition, and notification that fix work has begun.
The expected nominalization finding disappears. `LICENSE.curl` retains the
factual source's notice.

The historical Moment release-note passage is a technical counterexample to
treating every contrast as needless. Its ordinal return value and `Date` copying
behavior are distinct changes. The contrast rule correctly reports the repeated
form; deleting those distinctions would lose information. It remains a note and
does not fail this fixture's gate.

These are construction and source-mapping regressions, with exact expected
diagnostics and unchanged input checks. They are not human precision estimates or
proof that an agent's rewrite improves a reader's experience. Model generation is
never part of the test.
