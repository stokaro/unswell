# Whole-page recall on technical documentation

Unswell misses most of the wording problems in this auxiliary review. On 12
sampled Ptah pages, both technical and strict detect **1 of 27** defects marked
before the new diagnostic outputs were opened. On 12 historical pages they detect
**0 of 21**. These are single-assistant development judgments, not qualified
human labels or a population recall estimate.

The two previously discussed pages are separate anchors: PostgreSQL has 2/7
annotated defects detected, and database URLs has 2/2. Pooling those known examples
with the sample would overstate the result on pages not selected for their known
findings. All 26 files pass the existing gate in both profiles.

- [Results](RESULTS.md): page/category counts, diagnostic usefulness, sampling
  sensitivity and provisional clean-unit alarms.
- [Missed passages](MISSES.md): all 52 missed events, pinned source links,
  quotations, rationale and proposed repairs.
- [Protocol](protocol.md): fixed sampling, rubric, reviewer identity and scope.
- [Annotations](annotations.json): complete-file coverage, 57 frozen defects,
  13 uncertain cases and acceptable controls.
- [Dispositions](dispositions.json): all 130 technical and 149 strict findings
  reviewed individually, including related locations.
- [Post-freeze addendum](addendum.json): ten complexity events noticed with the
  detector's help. They affect diagnostic usefulness, not primary recall.

## What the measurements establish

The principal gap is not a gate threshold. Technical emits 84 length warnings
among 130 findings; strict adds 19 unscored dash-density notes and detects no
additional frozen defect. Many missed passages describe why the document or
implementation is supposedly well designed instead of stating behavior. Others
make a broad evaluative claim without defining its scope.

Examples of missed constructions include:

- `p002-d02`: “That is the whole design” followed by a defense of avoiding a
  second output language. The length warning on the sentence does not identify
  the removable justification.
- `p008-d01`: an unqualified zero-cost claim about requiring a flag, followed by
  a decision/default contrast. The actual safety condition is already explained.
- `p007-d02`: declaring a mapping the right answer nearly always, without a
  denominator or workload. The type-specific exception can be stated directly.
- `p012-d01`: the same single-source glossary benefit explained twice, with
  different wording and intervening contributor instructions.
- `p014-d01`: one Literal-documentation change listed twice within the same
  Pydantic release, with the same issue number.

The historical controls expose a second problem: repetition rules group sponsor
acknowledgments across releases and installation introductions across operating
systems. Their form repeats, but their audiences or obligations differ. The
source-provider table in Ptah legitimately repeats a capability in two rows.
Those are not instructions to erase repeated information indiscriminately.

There are also useful findings. The repeated exit-1 explanation is caught; the
known PostgreSQL evaluative closure and checks-versus-trusts slogan are caught.
Detector-assisted review identified ten additional complexity events where a
comparison table or separate command cases would improve readability. Leaving
those outside the original denominator makes the review's own omissions visible.

## Engineering priorities

1. **Contextual framing and certainty.** Evaluate bounded sentence/clause
   constructions for self-justification, self-certification and unscoped evaluation.
   Start from missed IDs, then write counterexamples in which a reason, measurement,
   scope statement or guarantee is necessary. A bag of words such as “right,”
   “design” or “important” is not sufficient. Do not turn this review into a default
   gate qualification.
2. **Repetition scope and short duplicates.** Compare within the relevant release,
   task and table row before treating recurring text as debt. Test the Pydantic
   duplicate alongside changing sponsor lists, platform instructions and capability
   rows. Preserve source segments across the entire repeated event.
3. **Independent confirmation.** Any implementation tuned on these records needs
   separately selected complete pages before reporting an improvement. The current
   sample is development material from now on. Independent human adjudication and
   stable-default qualification remain the work of #22 and #26.

These priorities extend the existing engine. No rule, weight, threshold, profile,
public API or playground runtime changes in this record.

## Limits of the evidence

The reviewer was the Codex assistant in this session, with one annotation pass
before new outputs and one diagnostic-assisted pass. The reviewer knew the
repositories and had prior exposure to related research and the anchors. The
freeze records that sequence and detects later drift; it is not an independently
timestamped or externally blinded annotation service. No human agreement metric
or independent scientific qualification is claimed.

The seven-category rubric is the existing [editorial rubric](../../../docs/editorial-annotation.md).
Required edits, optional alternatives and unresolved cases are separate. The
matching contract is intentionally conservative: a finding must describe the
annotated defect and overlap its source segments. The evaluator enforces category
compatibility and ranges; it cannot independently verify the reviewer's semantic
judgment. This remains an inspectable auxiliary gold set, not exhaustive truth.

Historical sources have dated provenance, not automatic human/acceptable labels.
Ptah has maintainer-reported repository-level AI provenance, not verified authorship
for each passage. The long historical stratum contains three changelogs and one
release-policy page. The cohorts differ in genre, format and repository diversity;
their scores do not measure an AI-origin effect. Bootstrap intervals describe
variation among these selected pages, not annotation uncertainty.

The input is complete source files, including all sections. Imported glossary
entries, generated previews and text embedded in images are not expanded. Thus
full source coverage is not full rendered-site coverage. The 136-page Ptah frame
is pinned to `654eae5591392278e6c8bce8e54737f780766f19`; today's edge site may differ.
Original bytes, root notices, rights and date evidence are retained in
`inputs.tar.gz` and `manifest.json`.

## Reproduce offline

The research tools use Python's standard library. They are not dependencies of
the product runtime or model training. The evaluator does not call a model or
network service, rerun NLP, or invent labels.

From the repository root:

```sh
record=research/reviews/2026-09-17-full-page-recall
python3 "$record/tools/evaluate.py" "$record"
python3 -m unittest discover -s "$record/tools" -p 'test_*.py' -v
python3 "$record/tools/render.py" "$record"
```

`render.py` rebuilds `results.json`, `RESULTS.md` and `MISSES.md`. Repeating it must
leave these outputs unchanged. The evaluator rejects source or annotation drift,
invalid/UTF-8-breaking spans, incomplete coverage, missing dispositions/documents,
unknown or incomplete runs, unmatched credit and semantic category mismatches.
Multiple findings may identify the same event; recall counts that event once.
Uncertain judgments never become positives.

To repeat the public-engine run, build the reference commit
`cc79256188310f9d63a86f259044df4c3a27af1d` with `CGO_ENABLED=0`, `-trimpath`,
`-buildvcs=false`, and
`-ldflags '-X github.com/stokaro/unswell.BuildCommit=cc79256188310f9d63a86f259044df4c3a27af1d'`.
Then supply that binary explicitly:

```sh
python3 "$record/tools/measure.py" --binary /path/to/unswell \
  --output /path/to/new-output --compare-reference
```

The output directory must be new. The runner validates the frozen archive, restores
whole sources, uses the two explicit configurations, records real exit codes and
compares the complete JSON results. Native analysis uses one worker and a five-minute
limit. Both profiles were repeated through this runner and matched the saved
reports exactly; see [reproduction.json](reproduction.json).

`tools/prepare.py` reproduces the fixed sample from the prior pinned Ptah source
directory and the committed long-prose-v3 historical archive. Its manifest stores
the full candidate frame, SHA-256 ranks, cell sizes and selections. It selects by
identity and preexisting length measurements, never finding counts. The archived
selected inputs are sufficient for evaluation without obtaining that larger frame.
