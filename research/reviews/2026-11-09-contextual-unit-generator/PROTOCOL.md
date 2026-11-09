# #349 source-context candidate generator: frozen offline screen

This is one exposed-development screen of a new candidate generator. It does
not diagnose an editorial problem, assess author origin, propose an edit, or
qualify a product rule. The hypothesis is that a discriminative classifier
trained on exact source units plus their immediately adjacent eligible units
can locate missed editorial targets across several categories better than an
external sentence score or a selector of already written diagnoses.

## Frozen inputs and boundaries

Use the SHA-256-pinned 36-page `joint-component-sol-high-evaluation/originals.json`
archive, all 123 frozen events, and the existing technical-v1
`reviewtext --source-inventory` exporter built from the pinned main tree. Verify
original UTF-8 source hashes, quotes, target byte spans, unit raw slices, and
eligible segment disjointness from protected spans. Export each page once. A
page or export error ends this run; do not retry or substitute a source.

Use the pinned current-review event ledger and category-frontier additions to
identify previously FULL events. Preserve 57 whole-page, 46 contextual, and
20 historical event denominators. These labels are exposed development data,
not prospective confirmation. No reserved prospective page is opened.

No hosted model, network lookup, runtime package, or model download is used.
The local research-only Python environment has pinned versions in the freeze
manifest. Its fitted logistic regressions are non-generative classifiers.

## Fixed candidate method

Assign each page to exactly one of the archive's 13 repository groups. For
each source unit, form one text field from the immediately preceding eligible
unit in the same source block, the target unit, and the immediately following
eligible unit in that block. Keep explicit `PREV`, `TARGET`, and `NEXT` markers.
Protected source text is never added as context.

Train three independent binary classifiers, for `empty_framing`, `wordiness`,
and `vague_claims`. A unit is positive for a category when any eligible byte
segment overlaps an alphanumeric byte of that category's frozen target. Other
units are unlabeled controls treated as negatives only for this exploratory
ranking; they are not adjudicated editorial negatives.

Use leave-one-repository-group-out fitting: for each held-out group, fit only
on units from the other 12 groups, then score held-out units. Fit word TF-IDF
with 1-2 grams, `min_df=2`, `max_features=30000`, and `sublinear_tf=True`,
plus character TF-IDF with 3-5 grams, `min_df=3`, `max_features=30000`, and
`sublinear_tf=True`. Concatenate the two sparse channels. Fit each category
with `LogisticRegression(solver="liblinear", C=0.5,
class_weight="balanced", max_iter=500, random_state=0)`. A missing class,
convergence failure, invalid source binding, or other operational error stops
the run. Use one process and fixed BLAS thread limits. There is no threshold
search or adaptive feature change.

For each page, select the highest-scoring held-out unit for each of the three
categories, in the category order above. If a unit was already selected,
choose the next highest-scoring unused unit. Break ties by source ordinal.
This yields at most three distinct candidate locations per page and at most
108 across the archive. Keep the selected unit's exact eligible byte segments;
do not expand it to a whole paragraph or include a protected byte.

## Fixed screen and next action

Count an event as location-covered only if the union of selected eligible
segments contains **all alphanumeric bytes of every frozen target**. Report
coverage and previously non-FULL gains by complete cohort and category, plus
candidate counts and abstentions. Partial overlap receives no FULL credit.

The method earns a separate editorial-review experiment only if at least
eight previously non-FULL events are location-covered, with at least one gain
in each cohort and at least two categories represented. This is a *location*
screen, not FULL-event recall or accepted delivery. A passing screen would
still require a newly frozen, repository-disjoint transfer, full combined
finding review, suggestion safety, and the original 80% FULL / 85% accepted
gates. A failing screen retires this exact configuration on these pages.

Freeze this protocol, script, source identities, current main tree, exporter
binary, review ledgers, and dependency versions before the first run. Publish
aggregate counts only; retain per-unit scores and event IDs in a local
untracked file. Do not alter these inputs after seeing the result.
