# Dependency backend evaluation

Issue #20 evaluates a real parser and defines the tree contract consumed by the
existing engine. The experiment uses the public extraction and NLP packages.
It does not add a second linter or change the builtin English provider.

**Decision: retain GoSpacy as an experimental candidate. Do not ship its current
high-level loader as an Unswell provider or enable dependency-based rules.**
Its output can represent real dependencies, but its loading, cancellation, mutable
state, and failure behavior require further adapter work. The small model also
makes questionable attachments on technical constructions. This evaluation does
not qualify passive-density, long-subject, or nested-clause defaults.

## Candidates and resources

| Candidate | Pinned code | Model and terms | Decision |
| --- | --- | --- | --- |
| [GoSpacy](https://github.com/bioshock/gospacy/tree/e2766da9ab71ffc55a5967a76a474046a96402bc) | `v3.8.14-port.2`, `e2766da9ab71ffc55a5967a76a474046a96402bc`, MIT | Official `en_core_web_sm` 3.8.0; model MIT, WordNet lookup data carries its own notice | Reproduced on the recorded probes; experimental |
| [Lingo](https://github.com/chewxy/lingo/tree/491e816b48d421cddb666db8c288cb62e4f617d7) | `491e816b48d421cddb666db8c288cb62e4f617d7`, MIT | Its dependency README explicitly withholds company models and requires users to train their own | Reviewed; defer a separate training experiment |
| [YAP](https://github.com/OnlpLab/yap/tree/188e0046317f61bfe64af9210f91bcb9d88d561a) | `188e0046317f61bfe64af9210f91bcb9d88d561a`, Apache-2.0 code | Documented distributed setup targets Hebrew; BGU lexicon is outside the code license | Reviewed; do not import the Hebrew bundle for English analysis |

No Lingo/YAP model was run or assigned an invented size, hash, or accuracy. YAP's
pinned README lists 6 GB RAM as an upstream requirement; this is not our
measurement. These decisions concern the inspected distributions, not a claim
that the algorithms cannot support English.

The archive downloaded from the
[official English model release](https://github.com/explosion/spacy-models/releases/tag/en_core_web_sm-3.8.0)
had this SHA-256:

```text
en_core_web_sm-3.8.0.tar.gz
14a2f31bc476af87019819ea8c9948fabdfd473a442edd6b1cba62bf0c2c0f55
```

The downloaded archive was 12,806,159 bytes; the model directory contains
15,231,350 bytes. `testdata/observation.json` records a SHA-256 for every loaded
model file. The model's `LICENSE` is MIT; retain it and `LICENSES_SOURCES` when
redistributing the bundle. Its metadata identifies OntoNotes 5 as commercial
training data licensed by Explosion, ClearNLP conversion as a reference rather
than packaged code, and WordNet 3.0 under Princeton's separate terms. Permission
to use the model does not grant access to or redistribution of OntoNotes.
No model weights or training data are committed here.

GoSpacy's selected runtime dependencies are recorded in this module's `go.mod`
and `go.sum`: MessagePack and its tag parser (BSD-2-Clause), regexp2 (MIT), and
Gonum (BSD-3-Clause), in addition to Unswell. Tests use quicktest. The pinned Go
module cache retains their original notices. This developer command is not a
published release artifact; a future distributed model adapter must package the
notices of its exact selected graph.

## Reproduce the Go experiment

Obtain and extract the official archive explicitly. Verify its SHA-256 against
the value above before using it. Set `model_dir` to the extracted directory that
contains `meta.json` and `config.cfg`, not the outer source-distribution directory.
From this directory:

```sh
CGO_ENABLED=0 go test ./...
probe_commit=$(git rev-parse HEAD)
CGO_ENABLED=0 go build -ldflags "-X main.buildCommit=$probe_commit" -o dependencyprobe ./cmd/dependencyprobe
./dependencyprobe --model "$model_dir" --input testdata/cases.json --repeat 100 > observation.json
```

The command never downloads resources or launches a model server. It accepts
only explicit model/input paths, rejects unknown case fields and duplicate IDs,
and bounds case bytes, file bytes, model directory entries, and repetitions.
It checks parser/tagger resources before calling the upstream high-level pipeline
and validates each mapped dependency tree. It runs sequentially and checks
cancellation between calls. The upstream inference call itself has no context
parameter: this is a research limitation, not a claim of bounded cancellation
latency for a production provider.

Each recorded `inputs` entry contains the exact text passed to the parser and its
sentence range. This preserves context for reference comparison. Protected NUL
boundaries divide calls; code excluded by extraction never returns through the
parser. The adapter converts Unicode rune offsets to mapped UTF-8 byte offsets
and document-level heads to sentence-local indices. It preserves the English
model's labels, including `nsubjpass`; these are not relabeled as UD v2.

## Python reference, outside ordinary CI

Use a separate Python 3.12 virtual environment and install
`reference-requirements.txt`. The reference requires spaCy 3.8.14 and the same
model directory. It uses the recorded parser inputs, not a different sentence
context or raw Markdown:

```sh
python reference.py --model "$model_dir" --observation observation.json > reference.json
```

`testdata/reference.json` records the installed versions, observation hash,
reference outputs, and comparison counts. This checks compatibility with another
implementation of the same model. It does not establish grammatical accuracy.
The reference exits unsuccessfully when tokenization or any compared field differs.
Python is not used by `make check`, product tests, or the Go command.

The reference wheel is pinned by SHA-256 in `testdata/reference-wheel.json`.
All 1,032 installed spaCy package files matched that official wheel. The source
tag advertised by the port was not found during verification, so the record uses
the verified package artifact rather than an unverified source revision.

## Observations and limits

On September 8, 2026, macOS 26.5 / Apple M3 Pro / Go 1.27.1, the recorded Go run
processed ten sources with 100 repetitions: 8,400 counted words and 10,300 tokens.
It took 6.10 seconds wall time with 203,948,032 bytes (194.5 MiB) maximum resident
set size. The command recorded 30.5 ms model loading and 5.15 seconds for extraction,
parsing, and tree/source validation. Lazy initialization occurs in the analysis
interval. Report encoding is included in wall time. Allocated bytes and retained
Go heap are recorded separately; neither is peak RSS. The source commit is
`efb31808df9556fa64b460c1fa04b382572cb06b`; `testdata/resources.json` links the
process measurements to the observation hash, and `testdata/process-time.txt`
preserves the original timing output.

All 103 unique tokens matched Python in token text, head, relation label, fine POS,
and byte start on these sources. A separate run of the upstream GoSpacy parser
suite, with the real small model installed and `CGO_ENABLED=0`, matched all 68
heads and labels in its eight frozen reference examples. Medium/large-model tests
were skipped because those models were not installed. The committed experiment
is the ten-source comparison; neither result supports a corpus-wide accuracy
claim or the 100,000-word / 2-vCPU Linux acceptance target.

The `technical-3.txt` and `technical-4.txt` cases illustrate a limitation: the
model chooses `report` and `list` as roots where a technical editor would expect
the main predicates `accepted` and `remains`. Python reproduces these choices.
These are agent-authored structural probes, not independently annotated data.
The root e2e replay retains the observed `server` passive-subject label in the
third case; it does not rewrite the prediction to manufacture a better result.

## Requirements before a product adapter

- Load immutable, explicitly supplied model data at the application boundary.
  The current upstream bundle and component constructors read files, including
  lazy reads on the first inference call, and may write warnings to stderr.
- Turn every missing or incompatible required component into an error. Upstream
  `bundle.ensureComponents` logs and skips parser construction errors.
- Bound model decoding, per-call scratch space, vocabulary growth, and work.
  The command's directory/input limits are not a hardened model-pack loader.
- Provide cancellation within inference and safe concurrent reuse. Upstream
  bundles mutate vocabularies and layer scratch; clones require a quiescent
  source and have their own memory cost.
- Evaluate supported grammatical constructions, labels, and uncertainty on an
  independently reviewed technical corpus. Add dependency rules only when their
  representation and error behavior have that evidence. A valid tree is not a
  correct tree, and a passive construction is not automatically bad prose.

The [tree ADR](../../docs/adr/0012-dependency-contract.md) and executable public
contract tests are the product changes from this evaluation. Shared feature
qualification, training, and broader resource acceptance remain in #56, #23,
#26, and #28.
