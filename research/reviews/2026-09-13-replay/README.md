# Replay of the September 11 diagnostic reviews

The newer engine removes several observed sources of noise while keeping
useful findings. It also exposes an eligibility loss in paragraph-openers
and leaves structural prose inside source comments and workflow code unresolved.
This record compares identical inputs. It does not measure precision or infer
authorship.

The subject snapshots and original records remain unchanged:

| Subject | Pinned source | Old engine |
| --- | --- | --- |
| [Ptah](../2026-09-11-ptah/README.md) | `7b47e7cfb5d4ff32a38375345069f5533bff892f` | `d3c9f231f372694c2f38307e209e65503bf3b1e1` |
| [800 responses](../2026-09-11-responses/README.md) | records/tasks/shards at `a1c6d7af67a89ab573997aba9cec7b51a09f4dbb` | `a1c6d7af67a89ab573997aba9cec7b51a09f4dbb` |
| [Unswell](../2026-09-11-unswell/README.md) | `a1c6d7af67a89ab573997aba9cec7b51a09f4dbb` | `a1c6d7af67a89ab573997aba9cec7b51a09f4dbb` |

The new engine is `06b38615d4483970b65a6fbe2187aeb4864830dc`.
It includes the alpha.3 repairs, shared-feature exclusion repair, and later
parser repairs through #235. This is a comparison of the combined engines;
it does not attribute every change to one pull request. It is also not a
benchmark of a published release; that work remains in #219.

## Measurement

| Input and policy | Old documents / words / findings | New documents / words / findings |
| --- | --- | --- |
| Responses, all rules | 800 / 15,920 / 39 | 800 / 15,920 / 39 |
| Ptah, default | 4,110 / 2,564,578 / 11,067 | 4,111 / 2,568,097 / 11,031 |
| Ptah, all rules | 4,110 / 2,564,578 / 24,227 | 4,110 / 2,560,685 / 21,438 |
| Ptah, comments and Markdown | 4,110 / 1,922,635 / 22,160 | 4,110 / 1,920,741 / 19,532 |
| Unswell, all rules | 785 / 149,157 / 488 | 785 / 148,104 / 402 |

Every old count matches its dated record. All runs are complete except the
old Ptah default scan: one non-Latin source failed, producing a source error
and a wrapper error, with exit code 2. The new default scan includes that file
and completes. Both all-rules Ptah policies retain the original file exclusion
so that their selection remains comparable. No failed document becomes a
zero-risk observation.

[comparison.json](comparison.json) records report and binary hashes, the
compiled policy, rule versions, NLP resources, errors, exclusions, abstentions,
and execution limits. The configuration bytes are identical within each pair;
no YAML migration was needed. Compiled identities change with the engine.
The POS and sentence-model hashes remain the same, while the NLP version adds
`boundaries-v1`. The grade-metric, noun-stack, and parenthetical rule versions
also change. Grade-metric v2 reports separate prose-only ARI fields; those schema
changes count as changed evidence even where the finding remains in place.

All binaries were built with Go 1.27.1 on darwin/arm64 and `CGO_ENABLED=0`.
Each scan uses `--no-gate --include-source --timeout 20m --jobs 2`, with
`GOMAXPROCS=2`. The original reviews used four workers. Replaying them with two
reproduces their counts and samples. This is a load limit, not a CPU quota or
performance claim. The default candidate budget is 100,000; the copied
all-rules configurations retain their original 2,000,000 budget.

## Matching and review

[Per-rule tables](tables.md) distinguish retained, changed, removed, and added
findings. Matching starts with rule, path, and source span, then uses a unique
identical excerpt or a mutually unique overlapping range within the same rule
and file. IDs and fingerprints never decide whether a finding is new.
Metric values, metric schemas, policy, and occurrence counts are material;
local block IDs and source mapping are tracked separately.

Each compressed JSONL archive listed in `comparison.json` contains every
alignment, original locations and excerpts, both fingerprints, and hashes of
the complete diagnostic content and mapping. Raw reports can be regenerated
with the runner below. It verifies every reported document against the source
manifest before comparing anything. The archives use a fixed gzip timestamp.

One old long sentence in Ptah's MySQL version note becomes two long sentences
after a correct split following `26.7.`. The matcher leaves this one-to-two
relationship unresolved instead of assigning an arbitrary partner. The review
records the split explicitly. It appears in all three Ptah profiles.

[review.json](review.json) records this Codex agent's reading of all 221 original
sample entries and 51 additional changed/added/unresolved rows. Two Ptah sample
entries lack a line number and each match two locations. Both locations are
retained, making 223 original locations and 274 review rows. The delta sample
uses a fixed SHA-256 ordering with the `217-v1` salt, up to five rows per
rule/status/input after excluding original samples. It is a development sample,
not an exhaustive judgment of every change.

Maintainer confirmation is pending. No independent human annotation occurred.
These judgments are not training labels, a precision estimate, or qualification
under #22/#26. Some old judgments are revised after reading the source:

- `(rule (c))` is a reference label, so removing its nesting-only parenthetical
  finding is appropriate despite the original positive judgment.
- The five disputed purpose phrases in Ptah's style material actually use
  `in order to` in the surrounding sentence. The phrase is not the cited banned
  term. String contracts can still require preserving the wording.
- The OCI help value has several paragraphs and a YAML example. Its old
  description as a one-line help string missed the lost structure.

## What survived and what disappeared

All 25 previously justified Unswell sample findings remain. The 18 removals
from its previously unjustified sample cover embedded shell programs, false
noun stacks at predicates, and short reference/license insertions. Nine sampled
readability findings remain with changed evidence; a workflow-code finding,
a state candidate, and an API-template opener also remain.

All 19 previously justified response findings remain, including the three with
changed readability metadata. The two joined sentences after `chi.Router.` no
longer fire as long sentences. Two Cobra responses gain passive-candidate
findings after sentence repair; their text includes `is shown`, `is started`
or `is launched`, and `is pressed` with the original Windows, zero-duration,
and empty-string conditions intact.

Eighteen response findings still concern JSDoc examples or tag lists. The old
review deliberately stored these as unfenced Markdown, so they remain prose
under that representation. They do not test the source-comment repair.
The root e2e control now pairs a real JavaScript `@example` comment with adjacent
eligible prose. Changing the historical response bytes would hide this
representation limit.

Ptah retains 74 of the 77 originally justified sample findings, counting changed
findings as retained constructions. Of the other three:

- The reference label above should have been a negative example.
- Prose-only ARI stops flagging the `collectPages` comment, but its two actual
  long sentences still produce findings at 40 and 38 words. Source words,
  sentence counts, and exclusions are unchanged.
- Three paragraphs still start `Measured on PostgreSQL`. Correctly splitting
  the short first sentence of one paragraph makes it fail the paragraph rule's
  sentence-based `min_words` check. This is an eligibility loss, tracked in
  [#239](https://github.com/stokaro/unswell/issues/239), not removed noise.

The comments-only comparison still carries most Ptah findings. Removing string
contexts reduces code/test-string noise but cannot resolve lists inside comments,
formula sensitivity to terminology, or necessary repetition in API descriptions.
Use explicit owner/context exceptions for known SQL, regex, and test literals;
this review provides no basis for excluding all human-readable strings.

## Rule dispositions

These are proposed engineering dispositions, not changes to default policy.

| Rule or group | Disposition | Reason |
| --- | --- | --- |
| `filler.wordy-phrase` | Retain | The inspected purpose constructions survive; preserve external wording contracts. |
| `filler.weak-intensifiers` | Leave opt-in | The words occur, but emphasis can distinguish technical cases. |
| `format.em-dash-density` | Leave opt-in | The measured marks remain; punctuation density is a policy choice. |
| `readability.grade-metric` | Leave opt-in | Prose-only measurement removes some identifier effects; necessary vocabulary still drives findings. |
| `readability.long-paragraph` | Narrow | Keep actual long prose; preserve paragraph boundaries in help values and comments (#237). |
| `syntax.long-sentence` | Narrow | Keep long conditional sentences; do not merge comment lists into one sentence (#237). |
| `syntax.noun-stack` | Leave opt-in | Predicate errors are repaired; established technical noun sequences remain common. |
| `syntax.parenthetical-load` | Narrow | Reference labels disappear and real qualifications survive; embedded workflow code remains (#238). |
| `syntax.passive-candidate-density` | Leave opt-in | The POS shape can describe a state; the descriptor already says it does not establish grammatical voice. |
| `repetition.exact-sentence`, `repetition.near-sentence` | Leave opt-in | Shared API obligations and expected strings are not necessarily redundant. |
| `repetition.sentence-openers` | Leave opt-in | Repeated openings occur in both prose and required reference templates. |
| `repetition.paragraph-openers` | Investigate | Decide paragraph eligibility without undoing correct sentence splitting (#239). |
| `repetition.ngram-density` | Leave opt-in | Repeated API wording and unfenced example output need context. |
| `repetition.syntax-template` | Leave opt-in | Similar test names and error shapes need distinct technical obligations. |
| `repetition.paragraph-overlap` | Leave opt-in | Lexical overlap does not prove equivalent meaning. |
| `gate.paragraph-score` | Retain | It remains a derived threshold result; do not count it as independent evidence. |

Rules absent from the reviewed sample receive no qualification from this replay.

## Controls and follow-ups

The root [review_controls fixture](../../../e2e/testdata/review_controls/case.json)
checks six files through the compiled CLI. Seven positive diagnostics remain
beside negative predicate/reference examples, a protected Python heredoc, and a
real JSDoc example. It retains conditions, negation, `RetryLimit`, and numeric
limits. Goldens check original byte locations, exclusions, JSON, SARIF, and
terminal output with BOM and CRLF. The harness also checks that source files
remain unchanged. These are constructed regression controls, not corpus labels.

[probes.json](probes.json) records compact reproductions of remaining limitations.
The probe sources are under [probes](probes/policy.yaml). They record observed
behavior; they do not require a false positive to remain in future engines.

- [#237](https://github.com/stokaro/unswell/issues/237): paragraph/list structure
  inside source comments and selected multiline prose values.
- [#238](https://github.com/stokaro/unswell/issues/238): shell syntax in Actions
  YAML `run` scalars, with source mapping through both grammars.
- [#239](https://github.com/stokaro/unswell/issues/239): paragraph-opener eligibility
  after the sentence repair.

The state probe demonstrates the documented POS-candidate limitation. It is not
a new claim that Unswell determines grammatical voice. The fixtures derived from
Ptah and Unswell retain their repositories' MIT licensing; source commits and
locations are identified above and in the review rows. Other excerpts in the
alignment archives remain attributed to their pinned source manifests and
original generation records.

## Reproduce

Use local clones containing the pinned commits, Go 1.27.1, and Python 3.11 or
newer. The research helpers use the standard library. Python is needed only for
this offline replay; the shipped tool and its normal scans remain pure Go.
Build preparation may fetch Go dependencies. Scanning does not download models,
call generation services, or modify source repositories' tracked files.

```sh
python3 research/reviews/compare_test.py
GOCACHE=/path/to/go-cache python3 research/reviews/reproduce.py \
  --unswell /path/to/unswell --ptah /path/to/ptah \
  --output /path/to/new-replay-directory
python3 research/reviews/compare.py /path/to/new-replay-directory \
  /path/to/new-comparison-directory
python3 research/reviews/verify.py /path/to/initial-replay-directory \
  /path/to/new-replay-directory research/reviews/2026-09-13-replay
python3 research/reviews/probe.py \
  --old /path/to/new-replay-directory/bin/old-d3c9f23 \
  --new /path/to/new-replay-directory/bin/new \
  --output /path/to/probes.json
go test -count=1 ./e2e -run '^TestCLI$/^review_controls$'
```

The runner builds all three engines, creates temporary linked worktrees,
reconstructs response files as `record.text` plus one LF, and verifies their
original shard hashes and byte counts. It runs both engines on every policy and
removes its linked worktrees. Input manifests, binaries, raw reports, and stderr
remain in the chosen artifact directory for inspection. No current generation
cohort or neighboring agent output participates.

The complete runner was executed in a fresh directory after the initial replay.
[reproduction.json](reproduction.json) records independent report and binary
identity comparisons. The nine comparator tests cover identity-only changes,
source mapping, material metrics, ambiguous splits/merges, duplicate anchors,
cross-file/rule separation, nonfinite values, source integrity, and incomplete
runs. Race detection, active fuzzing, and coverage remain deferred to #123.
