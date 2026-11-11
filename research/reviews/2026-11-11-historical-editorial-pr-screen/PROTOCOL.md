# Historical editorial PR source-only screen

Status: frozen acquisition selection before complete-source review or patch
inspection, November 11, 2026. This is an exposed development screen for #349,
not a detector qualification or a claim of human authorship.

## Question

Can older merged documentation PRs selected for an explicit editorial purpose
yield source-visible, necessary #349 diagnoses and nearby retain controls? The
previous 18-PR title screen retained only two necessary cases, both repetition,
and its reviewer had seen after text before judging necessity. This screen
uses broader editorial titles and reviews source-only first.

## Frozen selection

GitHub title searches used `is:pr is:merged merged:<2023-01-01` with
`"redundant phrasing"`, `"repetitive wording"`, `"remove repetition" docs`,
and `"simplify wording" docs`, plus `"wordy" docs`. We selected the 13 PRs in
`selection.tsv` from distinct repositories on title and PR metadata. Each was
merged before 2023, has a named non-bot account and at most three changed
files, and does not overlap the previous 18-PR screen. The selection favors
title relevance; it is not a representative sample of all docs PRs.

The GitHub connector's PR metadata response contains a diff field, but no
patch was printed or inspected for the selected 13 PRs before this freeze.
The PR bodies for `freedomofpress/encryption-works#102` and
`trinodb/trino#14520` do describe their proposed after text. Those two are
an explicitly exposed secondary stratum and do not count toward the blind
source-only yield. The primary stratum contains the other 11 PRs.
Two other searched PRs (`docker/docs#15803` and `coin-or/python-mip#232`)
had partial patch text displayed during metadata inspection; they are
excluded from source-only review. Changed filenames were inspected to locate
base files. The exact base commit, path, PR URL, and author are frozen in
`selection.tsv`.

## Review order

1. Fetch the full pre-edit file at each frozen base SHA. Record bytes, SHA-256,
   and Git blob ID. Reject a file if its content cannot be retrieved or the
   historical source cannot be pinned. Keep full public source locally; only
   quote short excerpts in the result.
2. The root Codex assistant reviews each complete pre-edit source without
   viewing the PR's after text. Record whether it contains a necessary,
   source-visible editorial diagnosis in the six #349 purposes, its exact
   source span, explanation, and at least one nearby cue-matched span that
   should be retained. An edit is necessary only when leaving the source as
   written impairs clarity, accuracy, or usefulness under its own context.
   Optional style preferences, changed product facts, typos, and formatting
   alone do not qualify. Reviewer identity is an AI assistant per ADR 0041,
   never a human annotation.
3. Seal source-only decisions before viewing additions or PR discussion.
   Then inspect the patch and determine whether the actual edit supports,
   contradicts, or is irrelevant to each diagnosis. Preserve disagreements.
4. Replay the current Unswell CLI only on eligible Markdown sources. Record
   exact source/tree hashes and diagnosis judgments. A location overlap is
   not a FULL diagnosis. Non-Markdown files remain acquisition evidence, not
   native CLI recall evidence.

If fewer than four distinct source-visible necessary diagnoses remain among
the 11 primary PRs, stop: the acquisition strategy has insufficient yield to make
a new selector. Even if this screen passes, do not fit or qualify a selector
on these exposed pages. A separate repository-disjoint, prospectively frozen
confirmation cohort and the unchanged #349 complete-cohort 80% FULL / 85%
accepted-delivery gates remain required. Missing capabilities and errors fail.
