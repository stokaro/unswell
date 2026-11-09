# Frozen offline semantic admission screen for #349

This is one exploratory screen on the exposed six-stream candidate archive. It
tests whether a locally pinned, frozen encoder improves admission of useful
whole-page editorial findings. It cannot qualify a detector, establish advice
safety, or close #349. No hosted model calls or new grades are allowed in this
screen.

## Frozen inputs and decision

`scan.py` verifies the five archived source/judgment SHA-256 identities in the
November 9 structured-admission control. It also verifies the control script
and result, original source hashes, exact UTF-8 target byte spans, local APE
checkpoint and tokenizer identities. All candidate streams are pooled without
relabeling or editing. The 13 groups contain 12 whole-page repository groups.

The APE checkpoint's BGE-small-en backbone supplies mean-pooled, normalized
384-dimensional vectors. Encode three strings per candidate, each capped at
128 tokens: (1) target quote, diagnosis, revision need, and wording burden;
(2) the exact target line, closest preceding Markdown heading, preceding line,
and following line; (3) retention case and meaning-preservation explanation.
No labels enter embeddings. Use the first two targets in source order. Embed
offline on CPU, batch 32, four Torch threads. Each vector is unit normalized.

For each whole-page candidate concatenate the three vectors, the absolute
difference between candidate and source vectors, their cosine similarity, and
one-hot category, operator, and stream. Multiply each vector block by sqrt(384).
In each leave-one-repository-out fold, fit one-hot categories and a balanced
liblinear logistic classifier on the other repositories only (`C=0.01`,
`max_iter=1000`, `random_state=0`). Score the held-out repository. No feature,
threshold, category, or model tuning may follow the result.

Report AUC, the fixed 0.5 threshold, and the optimistic post-hoc threshold
envelope at at least 85% accepted whole-page findings. The fixed policy would
need at least 46/57 whole-page FULL events and 85% accepted findings. Keep the
contextual and historical candidate banks unfiltered to expose, rather than
hide, their separate ceilings. Report unresolved supplied edits and preserve
the full 57/46/20 denominators. The optimistic envelope is exploratory and
cannot be deployed. Any apparent positive requires a new page-disjoint natural
transfer with fresh grades, separate edit-safety review, and a complete-stream
check before changing the product. A negative retires this exact feature and
model configuration on these exposed pages.

Freeze `PROTOCOL.md` and `scan.py` hashes before the first encoder run. The
public result contains aggregates and input identities, not private pages,
individual grades, or model weights.
