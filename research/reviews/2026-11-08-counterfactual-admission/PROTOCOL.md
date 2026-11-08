# Counterfactual admission preflight for #349

Status: offline preparation. No candidate model has been executed.

This study tests whether a proposed editorial diagnosis follows its actual
source evidence when one relevant context fact changes. It does not score a
generic writing preference or infer AI authorship. All controls are synthetic
technical prose; none is an original #349 page or a prospective source.

The method adapts the paired evidence-intervention evaluation design of
[EviScope](https://aclanthology.org/2026.groundlm-1.20/). EviScope's explicit
evidence-action gate underperformed its vanilla comparator on two local models;
this protocol borrows the benchmark design, not an effectiveness claim.
The distinction between feedback goal and exact target location follows
[FOXGLOVE](https://aclanthology.org/2026.emnlp-main.781/).

## Frozen question

Each of four quartets keeps one target sentence and one narrowly worded
diagnosis fixed. The surrounding source is changed to make that particular
diagnosis apply, cease to apply, or become a different problem. The candidate
must identify the target using original UTF-8 byte offsets and make an explicit
admit/reject decision on the fixed diagnosis. Another valid criticism at the
same span does not count as an admission of that diagnosis.

The four families are redundant table announcement, unsupported comparison,
unsupported reader-need claim, and repeated release-status framing. Every
variant is labeled only for its fixed diagnosis. All claims about fictional
products or user studies are fixture text, not assertions about real systems.

## Offline baseline and later decision

Before any new model execution, freeze the complete fixture bytes, labels,
source hashes, verifier source, model request contract, and failure policy.
Run the existing exact-main Go engine on all 16 variants and record complete
outputs, exact spans, source hashes, and zero operational errors. Existing
detector behavior is a descriptive baseline; no finding is automatically
treated as a verdict on the fixed diagnosis without source review.

A later candidate may be tested only as a separately registered treatment.
Its preflight must decide the fixed diagnosis correctly for at least 14/16
variants, including both members of every evidence-flip pair, with zero
invented source coordinates and zero unsafe or unresolved advice. Failures,
abstentions, and invalid output stay in the denominator. Passing permits only
a full exposed 36-page/123-event development screen under the unchanged
80% FULL-event recall and 85% accepted-delivery requirements, then separate
prospective confirmation. It is not product qualification.

No original #349 labels, historical grades, responses, prospective sources,
or maintainer pair choices are changed or used as inference inputs. No model
call, runtime rule, threshold, or default policy change is part of this
preparation. Any future experimental model execution retains the issue's
separate authorization and budget requirement.
