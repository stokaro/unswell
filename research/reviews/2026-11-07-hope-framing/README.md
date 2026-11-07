# Bounded hope framing: exposed gain, prospective rejection

**Final decision:** reject the proposed default rule. A separately frozen
external transfer screen found one new, nonactionable warning on a council
meeting record and no new accepted finding. The final PR diff retains this
research record and its exposed-study verifier; it restores the runtime,
tests, catalog, and golden files to the merged-main behavior.

The [protocol](PROTOCOL.md) was fixed before implementation and replay. It
tests three source-local constructions that express a document endorsement,
design goal, or feature-removal wish through uncommitted hope. Existing direct
document-help hopes remain covered by the same rule. This is exposed
development: the Codex assistant had already read the source pages and frozen
reference events. No model call or new event label was used.

The comparator is merged main `0b5433f8289d27f97c905590b44acfaa221792c6`.
The complete technical and strict reports retain all 36 original pages and
123 reference events. `filler.evaluative-closure` advances from version 12 to
13. Three findings are added in both profiles; no prior finding disappears or
changes content. Prior findings of the updated rule receive new version-based
IDs. All 18 previously frozen external controls have unchanged findings.

| Profile and cohort | FULL events / total | Accepted / delivered | Required FULL | Required accepted |
| --- | ---: | ---: | ---: | ---: |
| Technical, whole page | 30/57 | 84/128 | 46 | 109 |
| Technical, context | 17/46 | 22/28 | 37 | 24 |
| Technical, historical confirmation | 4/20 | 6/9 | 16 | 8 |
| Strict, whole page | 30/57 | 84/147 | 46 | 125 |
| Strict, context | 17/46 | 22/29 | 37 | 25 |
| Strict, historical confirmation | 4/20 | 6/10 | 16 | 9 |

The three new source-context decisions are in [review.json](review.json):

| Frozen event | New target | Decision |
| --- | --- | --- |
| `whole/p016-d01` | `we hope that this document and the processes it entails will help to achieve that` | Accepted: keep the stated stability goal and release-policy scope while removing the document hope. |
| `whole/p017-d03` | `Hopefully, some day, the simd-accel feature will similarly become unnecessary` | Accepted: preserve current feature behavior and compilation warning; omit the undated wish. |
| `context/c04-d03` | `It is also hoped that if ... the solver would ...` | Accepted: state the conditional caching design goal directly, preserving the actor and modification limit. |

The rule supplies conditional editing guidance, not an applied edit. The
assistant reviewed each suggestion against the complete containing paragraph
and neighboring text; no separate revision was run through a meaning
preservation test. Every new primary span maps to original UTF-8 bytes.
Synthetic blackbox tests cover matching, near misses, repaired wording,
attribution, quotation, numeric and dated conditions, protected identifiers,
and CRLF/Unicode source mapping. The full-stream [verifier](analyze.py) checks
source hashes, versions, unchanged findings, target spans, review decisions,
and cohort arithmetic. Its negative tests reject a missing new finding, a
changed old finding, a changed source, and a promoted uncertain decision.

The predeclared development screen passed: three newly FULL events, three
accepted new findings, no lost witness, unchanged external controls, and no
unsafe advice identified in the reviewed contexts. The unchanged 80% FULL and
85% accepted-delivery gates still fail in all three cohorts and both profiles.
That result was not sufficient for default promotion. It does not identify AI
authorship or qualify #349.

## Prospective transfer screen

The [external protocol](PROSPECTIVE.md) and candidate commit `35e3350` were
fixed before the GitHub source search. Search metadata, [source selection](prospective-selection.json),
and [assistant judgments](prospective-reference.json) were frozen before either
binary ran. Nine repository-disjoint complete pages yielded 13 `hope`-family
passages: 11 acceptable unchanged, two uncertain, and **zero definite
actionable**. The sample therefore cannot estimate recall for the proposed
operator. It was trigger-enriched and was not a representative #349 cohort.

On both technical and strict profiles, the pinned main binary delivered 59
findings and the candidate delivered 60. The sole new finding targeted a
[parish council meeting record](https://github.com/rbirkby/Ilmer.org/blob/1aabf3894d77fc5094d30764446932deb967d3f3/parish/parishcouncil/2021-04-28.md):
the minutes ask interested landowners to contact the council so gates can be
ordered more economically. The assistant had marked that passage acceptable
before replay. The candidate called it a buried design goal and advised
preserving a nonexistent modification limit. No prior finding content was
lost or changed. The rule-specific no-false-positive condition fails; this
candidate must not become default behavior.

The [run identity](prospective-run.json), [derived result](prospective-result.json),
and [verifier](verify_prospective.py) bind all nine upstream Git blobs and
source hashes, the 13 exact UTF-8 spans, four complete local reports, and the
one-finding diff. Four [negative tests](test_verify_prospective.py) reject a
changed source byte, changed report byte, changed old finding, and promoted
control. Full third-party source bytes and reports containing them remain in
the local ignored archive; the public record contains pinned URLs, hashes,
judgments, and the result. To reproduce after fetching the pinned source
revisions and running the four commands in `prospective-run.json`:

```sh
python3 research/reviews/2026-11-07-hope-framing/verify_prospective.py \
  --sources /path/to/pinned/sources --reports /path/to/complete/reports
```

No model call was made. The implementing Codex assistant reviewed the passages
under ADR 0041; this is not a human or independent review. The search API
was queried twice for the first stratum because the initial response was
inspected without archiving all result metadata; the first three hit identities
and order matched. That deviation is recorded in `prospective-search.json`.

Check the exposed-development archive without model or network calls:

```sh
make check-hope-framing-screen
```

The [run manifest](runs.json) binds the candidate binary, policies, exact
commands, and compressed complete reports. The [result](result.json) is
regenerated by the verifier.
