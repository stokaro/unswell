# ADR 0028: Paired evaluation of frozen research trials

Status: accepted for numerical developer tooling; scientific acceptance remains open.

## Decision

Extend the existing research `evaluation` package and `corpus` command with
`compare`. Its inputs are saved predictions, their frozen comparison plan, the
original candidate artifact, and an independent annotation round. Reuse prediction
loading, target matching, decision selection, and metric definitions. There is no
source extraction, model execution, fitting, threshold search, or winner selection
during comparison. Product engine, report, and calibration contracts stay intact.

The plan pins both prediction hashes and their shared protocol hash before labels
are opened. Reject different corpora, partitions, target kinds, contexts, rubric,
profile, feature/preparation contracts, extraction policies, vocabulary, or NLP.
Feature columns, requested capabilities, fitted weights, calibration choice, and
frozen thresholds may differ. Required capabilities describe feature computation;
the provider identity and its supported capabilities remain equal. These
constraints intentionally reject a comparison between the current
`prepared_piece` and `source_document` trials. Neither implements the protocol's
target-only comparison. A complete run manifest and exposure record remain required.

## Numerical contract

Retain per-model summaries for every eligible labeled target and separately for
the intersection of covered targets. Count candidate-only, comparator-only, and
neither-covered targets. Missing labels remain acquisition exclusions; abstentions
remain eligible and supply no probability or threshold decision. Full-flow recall
counts abstained positives as undetected. Full-flow false-flag rate divides flags
on negative targets by all eligible negatives, including abstentions. It measures
observed flags in the flow, not the quality of unavailable predictions.

Micro estimates weight targets equally. Group-macro estimates average each defined
source-group metric equally. Each estimate records its contributing group count.
The macro difference averages within-group differences on groups where both
metrics are defined. It can differ from subtracting two marginal macro estimates
whose support differs. This distinction also applies inside each resample.

Use 10,000 unstratified replicates. Sort groups by ID; draw the original number of
groups with replacement using Go PCG(17,0). Retain both models, all member targets,
labels, and abstentions for every selected group. Precomputed sufficient statistics
bound work to at most 100 million group draws per scope for 10,000 input targets.
Cancel during draws; no partial report is returned. Compare scopes independently,
each starting from the same fixed seed. Cross-scope intervals are not paired.

Report two-sided numerical percentile intervals at 0.025 and 0.975 using linear
interpolation at index `(n-1)*p` in sorted valid replicates. Missing denominators
produce invalid replicates, counted per estimate. Never substitute zero. Require
at least two original groups defining the metric and two valid replicates to
display numerical bounds. This is an identifiability minimum, not a declaration
of adequate scientific sample size or reliable nominal coverage. Sparse data and
many invalid replicates still need the protocol's independent evidence review.

An observed zero false-flag rate suppresses its degenerate bootstrap risk interval.
Separately compute `1 - 0.05^(1/n)` for zero flagged groups among groups containing
eligible negative targets. State its requirement of independent Bernoulli sampling
from the declared group population. This conditional exact bound concerns any
false flag within a negative-containing group; it does not certify micro FPR.
The zero-event formula is unavailable when a group has a false flag. General
nonzero-event exact bounds and independent micro confirmation remain requirements.

## Evidence and limits

The versioned result binds plans, model identities, response kinds, training
prevalence, training and evaluation bases, decisions, and exclusions. It retains
unavailable-reason counts over each full prediction run and the comparison
binary's compiler, OS, architecture, and available VCS metadata. It retains
`unavailable_unqualified_model`. Hashes cannot establish honest metadata, source
independence, preregistration, or human annotation. No automatic qualification
or CI policy change follows from a favorable interval.

Tests cover paired versus independent resampling, unequal group sizes, different
coverage, missing classes, nulls, invalid plans, unchanged prediction files, and
cancellation. Compiled root e2e exercises two real prediction runs at different
frozen thresholds, then compares them after removing the external model files.
The prepared comparator also adds a POS feature to the word-count baseline; this
regression failed before allowing different requested capability sets.
Tutorial judgments remain explicitly simulated.

Issues #57, #25, and #59 still require matched A–G experiments, a real human corpus,
registered resource and exposure records, predeclared slices, ablations, figures,
qualified bounds, and independent final evaluation. This ADR fixes numerical
implementation choices allowed by the existing percentile protocol; it registers
no experiment and changes none of its qualification targets.
