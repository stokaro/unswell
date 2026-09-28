# Reader-facing document framing

This bounded repair addresses [#364](https://github.com/stokaro/unswell/issues/364).
It does not qualify a contextual detector or close
[#349](https://github.com/stokaro/unswell/issues/349).

The existing document-endorsement rule recognized `helped you configure` but
missed the infinitive in `helped to get you started`. Its new alternative accepts
bare and to-infinitive actions after `help`, `helps`, or `helped`, with an operand
or reader complement. Negative, conditional, and safety-related complements stay
outside this alternative. Existing reader-object behavior is unchanged.

The document-metadiscourse rule now recognizes an unqualified invitation to see
`it`, `this`, or `that` in action. It accepts `let us` and straight/curly apostrophe
forms. A named object, condition, quotation, or protected code does not qualify.
The rule remains an unscored, nonblocking note: an announcement can help a reader.
Neither construction establishes authorship or automatically warrants deletion.

## Complete-page comparison

Baseline: `0f43d6f60daeba4f9eb8b4bf394728490229c108`. The candidate was built from
that revision with the recorded working-tree changes. Rule versions advance to
`filler.evaluative-closure` 10 and `filler.document-metadiscourse` 2; their severity,
score, and gate defaults are unchanged.

Four complete public Markdown documents were scanned with each binary and both
profiles. The exposed containerd guide adds two findings in each profile, with
none removed. Three separate confirmation documents add and remove none. These
confirmation pages were pinned before rule edits or output inspection; they have
no positive examples of the added constructions and do not establish recall.

| Added location in the containerd guide | Assistant disposition | Suggested treatment |
| --- | --- | --- |
| `[7149, 7171)` — “Let's see it in action” | Accepted as a generic demonstration announcement | Omit the invitation if the following example already introduces the action; retain it when navigation benefits the reader. |
| `[14884, 14950)` — “I hope this guide helped to get you up and running with containerd” | Accepted as a document-outcome endorsement | Remove the author's endorsement; keep the following further-reading link. |

Reviewer: Codex assistant under ADR 0041. The complete guide supplies the context
for both judgments. This is two accepted additions on one exposed page, with
zero additions on three other pages, not an estimate of population precision or
a claim of broad recall. The historical source snapshot does not establish its
author's identity. All byte spans refer to the pinned original UTF-8 file.

The [measurement record](measurement.json) retains every added/removed finding,
binary/report hashes, commands, and timing. The public [input manifest](inputs.json)
pins repository revisions and file hashes. Full reports remain reproducible by
the script; third-party documents are retrieved separately rather than bundled.

## Reproduction

Retrieve each manifest URL into a separate directory using its `file` name.
Build the baseline and candidate with `make build` in separate worktrees, then run:

```sh
python3 -B research/reviews/2026-09-28-reader-framing/measure.py \
  --inputs /absolute/public-inputs \
  --before /absolute/baseline/unswell \
  --after /absolute/candidate/unswell \
  --output /new/output
```

The script verifies every source hash, refuses an existing output directory, and
rejects incomplete scans or rule abstentions. It saves complete JSON reports and
compares messages, source locations, related locations, evidence, severity, and
gate behavior; version-dependent identities are intentionally excluded from the
comparison. No model call or network access occurs during measurement.

The original confirmation selection included a Requests RST quickstart. Before
tuning or scanning, that unsupported input was replaced with the same project's
Markdown README. The manifest records this amendment. No output was used to
choose the replacement.

Blackbox tests cover both grammatical variants, malformed complements, protected
code, reported speech, conditions, real demonstration objects, Unicode, emphasis,
and CRLF source coordinates. The new positive tests failed before the repair.
