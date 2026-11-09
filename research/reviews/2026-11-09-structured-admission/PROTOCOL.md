# Exploratory structured-candidate admission screen for #349

Date: November 8, 2026. This is an exposed, offline method screen. It does
not change product policy, frozen event labels, or the prospective boundary.
No generative model call is permitted. The question is whether the archived
candidate's own structured critique fields carry a repository-transferable
signal about the prior Codex assistant's accepted versus nonaccepted grade.
This differs from the earlier source-only retrieval and MiniLM cross-encoder
experiments; it cannot by itself establish source-grounded necessity.

Use the five SHA-256-pinned inputs from
`research/reviews/2026-11-08-operator-triage/audit.py` under the archived
development root. Include every alias selected in that audit, with exactly
the original 123 reference events and three cohorts. Train and predict only
for the 2,101 whole-page aliases. Keep the 548 context and 198 historical
confirmation aliases unfiltered, so their measured scores cannot be silently
exchanged for a new test. For each of the 13 repository groups, fit on the
other 12 and score the excluded group. Prior grades are training labels only
outside that group. Never include a repository or page ID as a feature.

Fit one fixed `scikit-learn==1.7.2` pipeline per fold: word unigram/bigram
TF-IDF (`min_df=2`, `max_features=30000`, `sublinear_tf=True`) followed by
`LogisticRegression(C=1.0, class_weight="balanced", solver="liblinear",
max_iter=500, random_state=0)`. Text features are category, operator, stream,
the exact target quotes (up to 2,000 characters total), diagnostic,
`why_revision_needed`, `wording_burden`, `retention_case`, and
`meaning_preservation`. Do not use review rationales, accepted status, event
witnesses, issue identifiers, or whole source paths in features.

Report held-out AUC, a fixed `score >= 0.5` operating point, and the *optimistic
post-hoc threshold envelope* over all distinct held-out scores. For every
threshold report whole-page FULL events out of 57 and useful actual findings
out of delivered aliases. Also report the unchanged context and historical
confirmation cohorts, and the original five unresolved supplied edits.
The feasibility question is whether **any** held-out-score threshold can
reach at least 46/57 whole FULL events and 85% useful actual findings.
Even a positive envelope is only a lead: threshold selection on pooled
held-out scores is exposed and optimistic; it requires a new page-disjoint
development test, separate prospective confirmation, and independent advice
safety or diagnosis-only output before product use. If the envelope fails,
retire this exact feature family without threshold or feature tuning here.
