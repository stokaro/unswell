# Small-encoder training falls short of the recall target

Fine-tuning the small encoder improves retrieval over its frozen version but
does not reach the retained lexical comparison's recall. The result does not
justify adding this model to Unswell.

| Method | Positive pieces selected / 872 | Explicit controls selected / 1,234 | Reviewed events retrieved / 994 |
| --- | ---: | ---: | ---: |
| Prior SW8192 lexical model | 119 | 21 | 134 |
| Prior E frozen encoder | 88 | 14 | 98 |
| H: head-only training | 88 | 14 | 98 |
| FT: head and transformer training | 108 | 17 | 119 |

FT selects 12.4% of positive pieces and 1.38% of explicit controls. Its control
FPR is lower than the lexical model's 1.70%, so neither result establishes
dominance across operating points. Both remain far below the recall objective.
The 1% FPR constraint applies to development threshold selection; the evaluation
results show that it is not a guaranteed error bound on another source group.

On Ptah, FT selects 40/347 positives and 5/701 controls, compared with the
lexical model's 46/347 and 4/701. H keeps epoch zero in all five folds; FT
selects epochs 3, 3, 1, 3 and 3. FT also selects 243 unlabeled pieces, whose
quality is unknown. Its precision on labeled pieces does not measure the
precision of that full selected stream.

All ten fits completed and all ten saved snapshots passed the independent
numerical reference. The machine-readable [summary](summary.json) retains
cohort, fold, length-band and event-category results. The
[rule reference](rule-reference.json) still credits 144/994 fully diagnosed
events. These selections add no concrete diagnostic credit, and product
rules, models and gates remain unchanged.

This experiment tests whether editorial training improves the small BERT
encoder used in the [frozen-feature comparison](../2026-09-19-context-encoders/README.md).
H updates the logistic head. FT updates that head and the 32 transformer
parameters while keeping embeddings fixed. Both begin with the same encoder
and the corresponding fold's trained E head.

The source remains 157 exposed complete pages, 146 source groups and 29,750
eligible prose pieces. There are 872 positive pieces, 1,234 explicit controls
and 27,644 unlabeled pieces. These are the retained assistant judgments.
Unmarked prose remains unlabeled. No new confirmation or human annotation is
claimed, and a selected piece earns no diagnostic credit by itself.

## Protocol and implementation corrections

The [protocol](protocol.md) was frozen before real-data training. Each fold
holds out one source group partition for evaluation, uses the next partition
for checkpoint and threshold selection, and trains on the remaining three.
Neither evaluation labels nor outputs select the checkpoint. The threshold
constraints are development FPR at most 1% and labeled precision at least 85%.
The full evaluation denominators retain unavailable and unselected pieces.

Training runs in Go with CGO_ENABLED=0. The encoder mean-pools hidden states
with an attention mask and normalizes the resulting vector. Each arm uses
five epochs, batches of four, fixed ordering seeds, Adam at 1e-4 and the
retained head regularizer. Checkpoints 0, 1, 3 and 5 compete under the frozen
selection rule. The initial checkpoint can win. Raw scores are not calibrated
probabilities, and this experiment does not provide concrete editorial reasons.

Two failed attempts are preserved rather than interpreted as model results:

- The first trainable-variable selector included floating graph constants
  exposed by the importer. Its name-based embedding exclusion also froze a
  synthesized attention projection. [The correction](implementation-correction.json)
  disables fusion and restricts updates to original transformer parameters.
  The original graph and initializer order are preserved during export.
  Unequal snapshot bytes in the first synthetic repeat came from initializer
  ordering; the corresponding tensor values were equal.
- The corrected run encountered a worker-pool deadlock during a backward
  matrix multiplication. The retained goroutine dump shows nested callers
  waiting on the same bounded pool. [The scheduling amendment](scheduling-amendment.json)
  uses the supported `go:ops_sequential` backend option. Kernels retain their
  normal implementation. Both arms use the same amended execution settings.

Neither correction changes the data, labels, splits, optimizer settings,
training budget or selection criterion. The failed outputs do not contribute
to the comparison. The changes affect only the isolated research runner;
the installed module cache and product dependencies remain unchanged.

## Numerical checks

Before updates, every fit compares its reconstructed head scores with the
retained frozen encoder on available labeled training and development pieces.
The maximum allowed absolute error is 1e-4. After choosing a checkpoint, the
runner reloads its saved model and checks all development scores within 1e-5
and requires identical threshold decisions before scoring evaluation pieces.

The observed maximum warm-start discrepancy is 9.69e-6. Reloaded Go scores
match exactly. ONNX Runtime checks 32 fixed, hash-ordered available evaluation
pieces per fit; the maximum discrepancy across those 320 checks is 6.20e-6.
All ten initial fits took 1,379.75 seconds in total on this macOS arm64 host
with GOMAXPROCS=2. This is research execution time, not the product's target-host
scan benchmark.

Short synthetic training repeats produced byte-identical weights and heads.
A separate synthetic case exercises mixed 32- and 256-token padded batches.
Snapshot audits compare the original computation graph, unchanged embeddings
and constants, and the intended transformer updates. Numerical agreement does
not establish editorial quality.

The fixed FT fold-zero real-data repeat reproduced every numerical report
field, encoder byte and head byte. It took 187.77 seconds with a peak RSS of
472,858,624 bytes on this host. This is one repeated fit, not an all-fit or
cross-platform reproducibility claim. The full record is in
[real-repeat.json](real-repeat.json).

## Reproduction

Ordinary builds and product tests do not need an encoder, Python, network
access or model weights. All model execution below is optional research.
No encoder weights are distributed: the parent model manifest still leaves
the export's licensing provenance unqualified.

1. Reconstruct the parent `review-input.json` using the
   [retained input procedure](../2026-09-19-learned-proposals/README.md).
   Export its verified text with the Go `reviewtext` command.
2. Run `tools/prepare.py --input INPUT --texts TEXTS --out INPUTS` to reconstruct
   the exact `data.json` and `warm.json` hashes in `plan.json`.
3. Unpack `trainer-source.tar.gz` in an isolated directory. Obtain Hugot v0.7.8
   sources for the local replacement and overlay the two archived backend
   files. All runner, module and modified-file hashes are in `execution.json`.
4. Obtain the six model files pinned by the parent study's
   `feasibility/tiny-model-manifest.json`, verify their hashes, and put the ONNX
   file at `MODEL/model.onnx`. Uncompress the parent's `embeddings.json.gz`.
5. Build the archived `./supervised` package with CGO_ENABLED=0 and run
   `tools/run.py --binary BINARY --sources SOURCES --inputs INPUTS --vectors VECTORS
   --model MODEL --out OUTPUT`. It validates identities, refuses an existing
   output directory, and runs the ten fits sequentially with GOMAXPROCS=2.
6. Independently check a saved fit with `tools/reference.py FIT INPUTS/data.json`
   in the parent's isolated ONNX Runtime, NumPy and tokenizers environment.

`tools/test_runner.py --binary BINARY --model MODEL` runs seven blackbox cases
against the archived synthetic inputs, including a full training repeat,
long backward batches, rejected source bindings and explicit length abstention.

The offline evidence verifier recomputes thresholds and checkpoint selection,
checks source-group boundaries and full-stream denominators, and rejects
changed availability or training parameter identities. Its negative tests
deliberately corrupt those properties.

This result covers one small encoder, one training schedule and exposed
assistant reviews. It does not establish a limit for larger models or different
supervision. Further work must demonstrate useful source-specific diagnostics
and transfer to fresh complete pages before a model can change product behavior.
