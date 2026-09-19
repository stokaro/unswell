# Prospective supervised encoder comparison

This is exposed assistant development on the unchanged 157 pages, 146 source
groups and 29,750 eligible-piece targets of the previous encoder comparison.
It is not new confirmation, human annotation, probability qualification or
explained diagnostic recall. Do not change labels, gaps, groups or gates.

Compare H (head-only updates) with FT (head and transformer updates). Both
start from the exact pinned small BERT export and the corresponding E model's
training-only logistic head from PR341. Reconstruct raw head weights as
Weights/Scales and bias as Intercept-sum(Weights*Means/Scales). Attention-mask
mean-pool and L2-normalize the hidden states, as in E. Before updates require
raw-score parity within 1e-4 against retained E on available labeled training
and threshold rows. A parity failure stops the fit before interpretation.

H freezes all encoder variables. FT freezes embedding variables and updates
floating transformer variables. Both update the head. No token remapping,
new pretraining, data augmentation, category guesses or reviewed edits enter.
Inputs with >256 IDs retain sequence_limit; no truncation or zero substitution.
Use batches of four, padded to 32/64/128/256 tokens, no dropped final batch.

Use GoMLX Adam, learning rate 1e-4, beta1=.9, beta2=.999, epsilon=1e-7,
no weight decay or optimizer backoff. Loss is mean binary cross entropy plus
.01/2 times sum((raw_head_weight*original_training_scale)^2), with unpenalized
intercept. The original training scales stay fixed for both arms. No added
transformer penalty or dropout; the exported graph has its inference behavior.
Train exactly five epochs. Shuffle the sorted training rows each epoch with
math/rand seeded by 19092026+fold; use identical order in H and FT. Record
losses, steps, variable identities and execution cost. Maximum 30 minutes per
fold/arm, one fit at a time; a failure yields no partial successful result.

Evaluation fold i stays excluded from training and checkpoint/threshold
selection. Fold (i+1)%5 selects checkpoints at epochs 0,1,3,5. Other folds
provide labeled available training rows. At each checkpoint apply the retained
threshold constraints on available labeled development rows: FPR <=1%, labeled
precision >=85%, maximize TP with score ties together. Select the checkpoint
with greatest development TP, then fewer FP, then earlier epoch. Unavailable
nonempty operating points are explicit abstention. Keep all checkpoints and
the chosen one, including epoch0 when training does not improve the criterion.

Evaluate every held-out row once using the selected checkpoint; unavailable
rows have null scores and remain in full-stream denominators. Retain raw
scores, source identities, all labeled and unlabeled selections, cohort/fold/
length-band/coverage summaries and source-bound event retrieval. Compare both
arms with retained E, ES and lexical baselines. Any extra exploratory view is
labeled as such and cannot change the frozen operating point.

Repeat the chosen training execution deterministically and independently
verify saved-encoder logits against ONNX Runtime plus the saved head before
product consideration. Do not distribute model weights before provenance is
qualified. Generic selection earns no diagnostic credit: a useful result must
still support concrete source-specific explanations and new frozen complete-
page confirmation. Do not infer that a failed small fine-tune disproves all
supervised contextual methods.
