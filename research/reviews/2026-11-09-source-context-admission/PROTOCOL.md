# Source-context admission screen for #349

Date: November 9, 2026. This is an exposed, offline feasibility screen of the
archived six-stream candidate bank. It cannot qualify the product, does not
change frozen labels, and makes no authorship claim. No generative model call
or new editorial judgment is permitted during this run.

The hypothesis is narrow: the exact original source around a proposed
diagnosis may improve transfer of accepted-versus-nonaccepted admission beyond
the candidate's structured critique alone. The previous candidate-only
[screen](../2026-11-09-structured-admission/README.md) is the fixed control.
Its 0.5 operating point delivered 37/57 whole-page FULL events and 1,183/1,358
accepted findings. Its optimistic 85% envelope reached 45/57 FULL. Do not
change that control or select features from this run's grades.

Use the same five SHA-256-pinned input files, 2,847 selected aliases, 123
reference events, 36 pages, 13 total repository groups, and three cohort
boundaries as the control. Train and score only the 2,101 whole-page aliases.
Keep 548 contextual and 198 historical confirmation aliases unfiltered. For
each of the 12 repositories represented in the whole-page cohort, train on
the other 11 and predict only its held-out candidates. Never include page,
repository, path, source hash, review rationale, witness, or grade in features.

Require every original page SHA-256 and every candidate quote to match its
recorded original UTF-8 byte span. For each candidate, sort target spans by
byte start and use at most the first three. For each target add: the previous,
target, and next physical lines (each clipped to 400 Unicode characters), the
nearest preceding Markdown heading within 2,000 bytes (clipped to 120
characters), and fixed structural tags for a heading, list item, table row,
blockquote, code fence, and HTML line. Also add a target-count bucket (one,
two, three, or four-or-more). These are research features; no parser or source
extraction change is proposed for Unswell.

Vectorize the original structured candidate text with the unchanged control
TF-IDF settings. Vectorize the new source-context text separately using word
unigrams/bigrams, `min_df=2`, `max_features=30000`, and `sublinear_tf=True`.
Concatenate the two L2-normalized sparse matrices without reweighting. Fit
the unchanged `LogisticRegression(C=1.0, class_weight="balanced",
solver="liblinear", max_iter=500, random_state=0)` per repository fold under
`scikit-learn==1.7.2`, `numpy==2.5.4`, and `scipy==1.18.1`.

Report held-out AUC, fixed `score >= 0.5`, and the optimistic post-hoc
threshold envelope over all distinct held-out scores. The directional
question is whether the fixed point improves both whole-page FULL recall and
accepted delivery over the control, with no increase in unresolved supplied
edits. Report whether any optimistic threshold reaches at least 46/57 FULL
with 85% accepted delivery. A positive result is only a lead: candidates and
assistant grades are exposed, no new-page transfer or edit safety is tested,
and candidate generation is not standalone. A negative result retires this
exact combined feature representation without threshold sweeps or variants on
these records.
