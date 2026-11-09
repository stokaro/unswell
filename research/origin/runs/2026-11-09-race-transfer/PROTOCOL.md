# RACE source-matched transfer: preregistered feasibility screen

Date: November 9, 2026. This is an origin-channel experiment, separate from
editorial issue [#349](https://github.com/stokaro/unswell/issues/349). It uses
the published, non-generative [RACE](https://aclanthology.org/2026.acl-long.235/)
encoder and RST parser. It makes no generative-model service calls and no
authorship or editorial judgment about an individual document.

V1 stopped at a pre-inference artifact check: one manually transcribed
tokenizer SHA-256 differed from the actual downloaded file. No selected
document was parsed or scored. [The terminal record](preflight-v1.json)
preserves the error. V2 changes only that expected hash and the versioned
freeze; the source selection, score, and advance rule below are unchanged.

## Fixed source and method

- ARB `giper45/ARB-Dataset`, commit
  `e03387f3d53efcaec9ca068c8e902e5f55611fe9`,
  `data/arb.parquet`: 27,439,482 bytes, SHA-256
  `66e48240618b92ce876e2097d8df52f06661a70e31791b91a95b15688a099ab0`.
  Reuse the existing verified loader and the source-disjoint `train`,
  `calibration`, and `test` assignment of the November 8 interpretable
  ensemble. No `train` or `calibration` ARB source enters this screen.
- Upstream RACE code: `ICTMCG/RACE` commit
  `61ce528430a8bb84fd1bfb3285ee34827182ef59`. Use its released
  `best_model.pth` checkpoint (603,067,257 bytes, SHA-256
  `ae9ba214ad47fd6834f82a951ad0b7a54f0903e171fa28b7f7608d0dea969b60`),
  not new training or tuning. Its embedded arguments specify four classes,
  512 tokens, RoBERTa-base, and seed 3407.
- RST parser: `tchewik/isanlp_rst` commit
  `cc8789f44c1341c3e7752b472634d75e81b62518`, `isanlp_rst==3.2.1`,
  `iinemo/isanlp` commit `2a102e59f9718acc7fe259dd8d83c66d5da39794`,
  English `rstdt` model revision `5141450ac3eb7decad0b8735fdbf5ec6928ec97a`.
  Its `best_weights.pt` SHA-256 is
  `c4008a42605d6fc2657be72ba1de50977b164f88bbc5f53b62a922f43fff3c32`.
- RoBERTa-base: `FacebookAI/roberta-base` revision
  `e2da8e2f811d1448a5b465c236feacd80ffbac7b`, local
  `model.safetensors` SHA-256
  `5bde1d28afb363d0103324efeb5afc8b2b397fe5e04beabb9b1ef355255ade81`.
  Pin tokenizer files and hashes in the execution freeze.
- Keep the complete original ARB text, including punctuation and line breaks.
  Parse exactly the same text that RACE tokenizes. RACE itself truncates to
  512 RoBERTa tokens; report how many documents exceed that limit. Do not
  truncate or normalize only one member of a source pair.

For each ARB domain (`owt`, `wp`, `xsum`), order only the previously assigned
`test` source IDs by `(SHA-256("unswell-race-v1|" + pair_id), pair_id)` and take
the first 20. Include each human original and all four `h2l` rewrites. This
fixed selection has 60 human source controls and 240 paired machine rewrites
before parse or model eligibility. Preserve failures as abstentions; never
replace a failed source or variant with the next ID. Keep text outside Git.

The predeclared score is `1 - softmax(logits)[human_written]`, an index of
similarity to the three machine-involved classes, not a calibrated authorship
probability. The `human_ai_polished` class probability is secondary for `h2l`.
Use checkpoint label order from the released RACE code; reject a conflicting
checkpoint order. Do not choose a new class combination after seeing results.
The comparison baseline is the unchanged full-document richness score from
`2026-11-07-richness-screen/run.py` on exactly the same eligible pairs. Report
both the original and first-100-token richness measures as named secondary
contrasts, with no hindsight selection.

## Fixed measurements and decision

Record every source's raw-text SHA-256, parser status, EDU count, tokenizer
length and truncation, inference status, and four logits locally. Publish
hashes, coverage, and aggregate scores; do not publish raw text or per-source
authorship verdicts. Count a pair only when both human and rewrite receive a
finite score. Report coverage, pooled AUROC with source clustering, matched
pair win fraction, and each domain and generator separately. A human source
appears four times in pooled AUROC, so do not interpret pair-level precision
as independent document evidence. Report bootstrap intervals by resampling
source IDs, not rows, with fixed seed 3492026 and 2,000 draws.

This **feasibility screen** advances the exact checkpoint to a larger,
independently calibrated transfer only if at least 54/60 human sources and
216/240 rewrites are scored, pooled source-matched AUROC is at least 0.75,
matched-pair win fraction is at least 0.70, and no domain's paired win
fraction is below 0.60. These thresholds are a directional filter, not the
product gate. If any condition fails, retire this checkpoint as an Unswell
origin candidate without retuning on these 60 sources. Regardless of outcome,
60 human controls cannot establish a 1% false-positive operating point.

The ARB corpus and split were used in earlier Unswell studies. This is exposed
development evidence. It does not prove transfer to software documentation,
mixed authorship, or later generators. The paper's main HART table uses a
variant-random split; its group-aware appendix and leave-one-domain-out
results are the relevant caveats. A pass here would justify a larger frozen
external screen with separate human calibration and test sources, including
scientific and technical documents. It would add no FULL editorial event or
accepted diagnosis to #349.

## Execution boundary

Commit the source selection, runner, negative integrity tests, exact request
configuration, checkpoint/model hashes, and failure policy before the first
RACE inference. Stop on the first infrastructure or integrity error; preserve
the trace and version any repair. Limit to one checkpoint, one parser version,
one pass over the 300 selected documents, at most four hours of CPU work
(two hours for parsing and two for checkpoint inference).
Do not start a run until the freeze is complete. Do not reinterpret a failed
or partial run as a quality result.

The bounded CPU container uses `python:3.12-slim` digest
`sha256:5e2bdb1b0d542d5039cffb8181c65930093949194a65c365dfb1b0fe13a4fa5b`,
two CPUs and 6 GiB memory. The exact installed environment is in
`environment.txt` (including `torch==2.6.0+cpu`, `torch-geometric==2.6.1`,
`transformers==4.51.3`, and `isanlp_rst==3.2.1`). This is an execution
record, not a dependency requirement for the Unswell library.
