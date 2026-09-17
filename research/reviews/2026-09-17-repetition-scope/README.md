# Repetition scope and short duplicate changes

This change removes eleven nonactionable development diagnostics and detects
the duplicated Pydantic `Literal` documentation entry, including its issue ID.
It also loses one previously actionable repetition warning. Development event
credit stays **12/57** in both profiles; within the repetition category it stays
**1/6**. These results show improved specificity on the reviewed changes, with
no net recall improvement.

On six separately selected confirmation pages, both versions miss all five
annotated repetition defects. One false comparison between shell completion
commands disappears. No new diagnostic appears. All development and confirmation
gates pass. This iteration does not establish that Unswell catches the full set
of editorial defects in technical documentation.

## Implementation and decision

The four existing sentence/opener rules now compare within grammar-derived
sections and explicit leading task conditions. Separate table cells have separate
comparison scopes; repeated sentences inside one cell remain eligible. Successive
headings with identical titles still introduce separate sections. Source-language
declaration names alone do not partition prose. Existing thresholds stay unchanged.

The new `repetition.duplicate-list-item` rule compares complete single-sentence
items in one simple unordered list and section. It preserves identifier case,
numbers, negation, punctuation and issue references. A single inline-code
identifier can match the same plain identifier, but commands, URLs and arbitrary
code cannot. Different link destinations do not collapse into one visible label.
Ordered, task and complex nested lists are outside this rule's current scope.

The rule has experimental warning defaults: weight 15, cap 30, no prohibition.
Default qualification remains separate under #26. Existing rule versions change
because their comparison scope changes; the catalog and current research class
manifest record the new rule. No gate threshold or global sentence minimum changes.

Accepting the narrower comparison scopes trades one known detection for fewer
unrelated comparisons. On p004, an old paragraph-opener cluster combined an exit-0
explanation with two repeated exit-1 explanations in another section. Splitting
the sections leaves two occurrences, below the unchanged opener threshold. The
two exit-1 passages still deserve editing: the frozen label remains actionable,
and the loss is counted. A future predicate-level repetition rule should address
that case without restoring the unrelated exit-0 comparison.

## Glossary paraphrase

The separate p012-d01 judgment remains missed. Both passages express the benefit
of defining a concept once and linking to that definition, with contributor
instructions between them. Their different wording is not full-item identity.
Sharing a glossary subject, a map metaphor or the word "page" does not establish
the repeated proposition. This implementation makes no semantic-equivalence
claim and gives the new rule no credit for that event.

## Evidence and limits

- [Protocol](protocol.md), its [scope clarification](protocol-amendment.md),
  [resource amendment](resource-amendment.md), [input freeze](input-freeze.json)
  and [annotation freeze](annotation-freeze.json).
- [Results](RESULTS.md), including separate repetition counts and resource costs.
- [Every diagnostic change](CHANGES.md) and its
  [machine-readable disposition](dispositions.json).
- [Confirmation misses](CONFIRMATION-MISSES.md) and
  [annotations](confirmation/annotations.json): five defects, five uncertain
  judgments and fourteen acceptable controls, with complete source coverage.
- Eight compressed reports under `reports/`, plus binary hashes, revisions,
  configuration identities, commands and per-process time and memory.

Before is `6f854b634657f45206221b205fe6abb8b4bb7d6d`; after is
`2ded5ff7e3a15b68d722a6d2b8e1599eca9351e7`. The reproduced baseline findings
match the preceding framing study exactly. The separate resource amendment
records the context-serialization budget added after the initial measurements.

The same Codex assistant implemented and reviewed the diagnostics. The maintainer
accepts this review under ADR 0041. It is not independent, blinded or human
annotation, and no inter-reviewer agreement is claimed. Existing labels were
preserved; uncertain judgments receive no positive credit. Every added or removed
diagnostic was reviewed, including the lost actionable finding. Unchanged
development findings inherit the earlier dispositions. Other confirmation
diagnostics were not relabeled to estimate overall precision.

Confirmation sources were selected before implementation by pinned identity,
one per cohort/length cell, excluding the 26 development pages and nine previous
confirmation pages. The assistant read all six source files before opening their
diagnostics and froze the annotations before scanning. No matcher was tuned from
these outcomes. Source files are distinct, but repositories may overlap; this is
not a population recall estimate. Imports, embedded components and images were
not expanded, so complete source coverage does not establish rendered-site
coverage. Source archives retain their notices and rights records.

## Reproduce

The standard-library Python scripts are optional research tools. They are not
runtime, ordinary Go-test or Go-training dependencies. From the repository root:

```sh
record=research/reviews/2026-09-17-repetition-scope
python3 -B "$record/tools/render.py"
python3 -B -m unittest discover -s "$record/tools" -p 'test_*.py' -v
python3 -B "$record/tools/measure.py" --binary /path/to/unswell \
  --set confirmation --output /new/output/directory
```

Use `--set development` for the original 26-page archive. Output directories must
be new. To replay selection, use `tools/select_inputs.py --ptah-sources PATH
--output NEW_PATH` with the pinned Ptah snapshot in the manifest. The evaluator
checks frozen hashes, source spans, notices, disjoint source identities, complete
reports, configuration, all diagnostic changes and source-bound event credit.
Repeated-event credit requires the actual related occurrences and counts once.

Blackbox tests cover section and task boundaries, independent table roles,
within-cell repetition, protected identities, differing IDs and negation, links,
list structure, source ranges, budgets and overrides. The end-to-end fixtures
retain the source examples and licenses. Baseline, changed-unit and trusted
suppression tests exercise within-section repetitions under the new contract.

Further work includes short repeated predicates, adjacent-word duplicates and
the glossary paraphrase. Their source-bound misses remain explicit; a new matcher
iteration needs new confirmation pages rather than tuning on these six.
