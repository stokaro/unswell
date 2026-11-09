# WikiSQE vague-label transfer screen for #349

Status: frozen before dataset download, training, scoring, or target-overlap assessment on November 9, 2026. This is an exposed development screen, not a prospective qualification or an AI-origin detector.

## Question and fixed inputs

Can a small, non-generative sentence classifier trained only on the WikiSQE `vague` label rank source-bound candidates that overlap currently non-FULL #349 editorial events? The source is the authors' `ando55/WikiSQE_experiment` dataset at revision `e6a3da39b44df3efa2c75706ccba1dc26dd17076`. Download only `vague/{train,val,test}.parquet` and record each file's SHA-256. The dataset's label `1` means a Wikipedia cleanup tag was present; label `0` means an untagged sentence from the same revision. Neither is an Unswell editorial judgment.

Use the already frozen 4,139 source-bound sentence rows in `research/reviews/2026-11-07-ape-proposition-screen/scores.json.gz`, retaining their original page IDs and byte spans. These rows cover the complete 36-page #349 cohort and were exported before this screen. Use the 123 unchanged events in `research/reviews/2026-11-07-current-main-editorial-review/events.json`, joined by ID to the frozen labels through the existing `2026-09-26-outcome-announcements/verify.py` loader. Treat only the seven independently reviewed additions in `research/reviews/2026-11-08-dual-gate-feasibility/category-frontier.json` as already FULL after the baseline. Record SHA-256 for every local input. Do not use any #349 event, page, score, or reviewer judgment for training, parameter selection, or threshold selection.

## Fixed model and ranking

Fit `TfidfVectorizer` on the official WikiSQE `vague/train` text only: lowercase, Unicode accent stripping disabled, word tokens of at least two word characters, word 1- and 2-grams, `min_df=2`, `max_features=150000`, `sublinear_tf=True`, L2 normalization. Fit `LogisticRegression(C=1.0, max_iter=1000, class_weight=None, random_state=0)` on the corresponding binary labels. No feature selection, fine-tuning, calibration, or model search on Unswell. Record the installed Python, scikit-learn, and Parquet reader versions. Report accuracy, precision, recall, and F1 on the untouched official `test` split as an internal sanity check only; its balanced class distribution does not estimate deployment precision.

Score every nonempty source sentence or fragment by class-1 probability. Choose exactly the three highest-scoring rows per page, breaking ties by earliest original byte start and then row index. This fixes 108 candidates. Record all scores and the selected rows. A selected row is a possible span witness for a frozen event only if it belongs to the same page and overlaps a target byte span. An event's *span ceiling* requires the selected candidates to overlap every target span. This is an optimistic upper bound, not an editorial diagnosis or safe repair.

## Decision rule

Count previously non-FULL events satisfying the span ceiling by cohort and category. If fewer than eight distinct events satisfy it, stop before editorial review and reject this exact `vague` classifier and top-three configuration as a #349 candidate. If at least eight do, review all 108 selected candidates in their full source pages under ADR 0041, retaining all accepted and nonaccepted delivery outcomes. No candidate is admitted into the product until its complete cohort and prospective gates pass. Do not move the eight-event gate or change parameters after seeing results. Do not turn a topical sentence score into a finding without source-local diagnosis.

The exposed sentence list may split prose around protected inline content, so this is a transfer screen of the existing source-bound rows rather than a claim about a final production parser. WikiSQE's Wikipedia domain and artificial 1:1 test balance are additional limits. No external generative model calls are part of this protocol.
