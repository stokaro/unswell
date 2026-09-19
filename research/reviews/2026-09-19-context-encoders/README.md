# Frozen encoder features do not improve review retrieval

A small frozen BERT encoder does not outperform the wider lexical model on
these exposed assistant reviews. It can run locally in pure Go, but this
experiment provides no reason to install it in Unswell.

| Features | Positive units selected / 872 | Explicit controls selected / 1,234 | Reviewed events retrieved / 994 |
| --- | ---: | ---: | ---: |
| Prior SW8192 | 119 | 21 | 134 |
| E: frozen encoder | 88 | 14 | 98 |
| ES: encoder and shared numeric features | 81 | 12 | 88 |

The Ptah subset has 347 positive units. SW8192 selects 46, E selects 32 and
ES selects 29. Their false selections among 701 explicit Ptah controls are
four, four and three. No arm approaches the recall objective.

These are source-group evaluation folds within exposed development data.
All labels come from the retained assistant reviews. Unmarked text remains
unlabeled. E selects 160 unlabeled units and ES selects 169; their quality is
unknown, so labeled precision does not describe the full selected stream.
E's overall explicit-control FPR is 1.13%; ES's is 0.97%. The 1% constraint
was selected on separate development folds and is not a guaranteed evaluation
bound. Historical controls still show more false selections than Ptah controls.

The [protocol](protocol.md) and [split plan](split-plan.json) were fixed before
real-data encoding or prediction. Each original paragraph or protected-gap
fragment is encoded separately. This experiment tests attention within that
piece; it does not test whole-page reasoning, cross-block context, fine-tuning,
a larger encoder or an editorially pretrained encoder.

Of 29,750 units, 21 exceed 256 token IDs: two positives, one explicit control
and 18 unlabeled units. They retain null scores and remain in full-stream
recall denominators. No token truncation or zero-vector substitution is used.
The encoder receives no labels, provenance metadata, review edits or code.

Selection earns no new diagnostic credit. A raw logistic score neither
explains an editorial defect nor proves machine authorship. The retained
[rule reference](rule-reference.json) still has 144 fully diagnosed events;
these proposals do not increase that count. No gates or product models change.

## Runtime and numerical evidence

The isolated runner uses Hugot v0.7.8 and the pinned 128-dimensional,
two-layer BERT export in [the model manifest](feasibility/tiny-model-manifest.json).
No native inference library is linked into the Go executable. Model files
are loaded explicitly from local disk; the runner does not fetch them.
The export does not separately declare its license. Distribution is not
qualified, and no weights are committed here.

A first batch run failed because Hugot's Go backend used the explicit bucket
sizes to limit its graph cache without applying sequence padding. The
[research patch](runtime-patch.json) enables the configured sequence padding
for Go. The failed log and exact patched source are retained. No installed
module cache or product dependency was modified.

The patched encoder processed 29,750 rows, with 23,385 unique text hashes,
in 115.01 seconds and a peak RSS of 335,200,256 bytes on this macOS arm64 host
with GOMAXPROCS=2. This includes retained vectors and JSON output; it is not
a streaming runtime qualification or the roadmap's 2-vCPU Linux benchmark.
The ten Go logistic fits took about nine seconds in the initial run.

A deterministic 32-text sample was independently encoded with ONNX Runtime
and its tokenizer. Maximum absolute vector difference was 1.92e-7 against
the previously fixed 1e-4 tolerance. The sample contains 3–93 token IDs;
separate synthetic longer-input probes are in `feasibility/`. Numerical
compatibility does not establish editorial validity.

Preliminary MiniLM probes also ran in Go but were slower. They were resource
probes only: no MiniLM editorial fit is reported. The negative BERT result
cannot establish that all contextual encoders fail. Pretraining overlap with
historical source material is unknown.

## Reproduction

Ordinary repository checks test the export, fitting and rejection paths with
synthetic vectors. They do not download or run an encoder and do not need Python.
The frozen vectors allow Go training reproduction without model weights:

```sh
CGO_ENABLED=0 go -C research/annotation build -o /tmp/encodedreview ./cmd/encodedreview
gzip -dc research/reviews/2026-09-19-context-encoders/embeddings.json.gz > /tmp/embeddings.json
/tmp/encodedreview /tmp/embeddings.json < /tmp/review-input.json > /tmp/encoder-predictions.json
```

Obtain `review-input.json` with the retained
[prepared-input procedure](../2026-09-19-learned-proposals/README.md).
Its uncompressed SHA256 is in `split-plan.json`. `reviewtext` exports the
verified eligible text from that same input; it excludes labels. Rebuilding
an encoder run is a separate optional research operation:

1. Unpack `encoder-source.tar.gz` in an isolated directory.
2. Fetch Hugot v0.7.8 into a separate source directory; verify its recorded
   original `backends/backend.go` hash before applying the archived replacement.
   Keep the local module replacement from the archive.
3. Obtain the model files from the manifest's pinned Hugging Face revision and
   verify every size and hash. Place `onnx/model.onnx` at `tiny-model/model.onnx`.
4. Build `batch.go` with CGO_ENABLED=0 and run it with the local model directory,
   manifest and `reviewtext` export as its three arguments. Use GOMAXPROCS=2.
5. Optionally run `reference.py` in an isolated environment with the reference
   package versions recorded in `feasibility/manifest.json`.

[The measurement record](measurement-record.json) binds the sources, binaries,
artifacts and repeat comparison. An offline evidence check is:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 research/reviews/2026-09-19-context-encoders/tools/check.py
```

The next model hypothesis needs stronger task supervision or a different
representation, followed by source-specific diagnostic review. Existing
exposed folds can guide development; they cannot replace new complete-page
confirmation before a product claim.
