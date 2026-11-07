# AITDNA mixed-authorship feature pilot

Date: 2026-11-07. This is a frozen, offline research screen. No generative
model is called, no Unswell runtime rule is changed, and no result is credited
to the editorial event gates in issue #349.

## Question

Do the 50-token lexical-richness and normalized token-entropy features from
the published Unswell long-prose-v3 screen separate human-only documents from
documents written in an AI-assisted setting when the provenance labels come
from a user study? AITDNA is a different, nontechnical domain. Human-only and
AI-assisted are the study settings, not verified token-level authorship labels.

## Frozen input and identity

Use the public UKPLab/AITDNA document configuration, test split, 362 rows, at
the dataset repository HEAD observed before this protocol:
21d77f7d44833507750c0da309295c4147c1bda5. Retrieve the four fixed
datasets-server pages at offsets 0, 100, 200, and 300, each with length 100.
The dataset server does not provide a commit binding in these responses; raw
page SHA-256 hashes and a second repository HEAD check must be recorded. If
the HEAD changes or row indices/counts disagree, stop without scoring. A
single offset-zero row was viewed for schema preflight before this freeze.
No raw participant text is published by this pilot.

## Feature and evaluation contract

Use the unchanged Unswell research feature function in
research/origin/runs/2026-11-07-richness-screen/run.py, verified by SHA-256.
Join the ordered text snippets in each document with a newline. Reject malformed
rows and record documents below 100 prose tokens as abstentions. Evaluate the
same two features on a fixed 100-prose-token prefix for every admitted row;
the existing 50-token, stride-25 windows then cover three positions. Report
raw prose-token length and Bot-snippet word fraction separately.

The positive class is the AI-assisted study setting (human_only=false); the
negative class is human_only=true. Compute directional rank AUROC, where a
higher feature predicts the positive class, for each feature on all admitted
documents and each declared task. Do not fit a model or choose a threshold.
Within each author who has at least one document in each class, compare their
mean feature values and report positive, negative, and tied paired directions.
Report class/task/author counts and the median length by class. No author is
treated as an independent paper or repository.

The predeclared feasibility screen requires both directional AUROCs at least
0.60 overall, at least two task-level AUROCs at least 0.60 for each feature,
and positive author-pair differences for at least 60% of eligible authors for
each feature. If any condition fails, retire this two-feature recipe for the
AI-assisted setting. Passing would only justify a separate provenance-aware
technical-domain validation, not an origin verdict or #349 qualification.

## Integrity and stopping

The scoring program must verify its own SHA-256, this protocol, the feature
source, all four raw page hashes, contiguous row indices 0..361, equal server
feature schemas and total counts, unique document rows, and the unchanged
dataset HEAD. Malformed data or a failed assertion is terminal; do not repair
or rerun against different data after seeing scores. Keep the result and its
input hashes locally; publish only aggregate results and method limitations.
