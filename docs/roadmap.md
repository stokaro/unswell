# Implementation and acceptance roadmap

The September 7, 2026 technical specification defines five stages. The requested
first alpha implements stages 0 and 1. Later stages remain product requirements;
the alpha does not claim completion of the full product specification.

## Alpha acceptance: stages 0 and 1

| Requirement | Evidence required before release |
| --- | --- |
| Public repository and project setup | Public GitHub state; license, contribution and security guidance; protected main; pinned CI and tools |
| Standalone public Go API | External consumer module analyzes bytes, registers a custom rule, writes JSON, and passes CI |
| Pure Go offline runtime | Native tests and `CGO_ENABLED=0` release builds; dependency and model notices |
| Source mapping | Exact UTF-8 spans for Markdown, GFM, Go comments, CRLF, BOM, Unicode, entities, escapes, and emphasis |
| NLP baseline | Tokens, sentences, Penn Treebank POS, tested NP/VP/PP chunks; capability failures; backend comparison ADR |
| Required 16-rule catalog | Executable positive, negative, boundary and technical-prose examples; no unsupported quality claims |
| Index and local gate | Fixed-point trace, correlated-evidence and group caps, nondilution, threshold diagnostics, independent severity |
| Strict configuration | Unknown fields, rule IDs, parameters, duplicates, bad ranges and unavailable models fail before analysis |
| CLI | Files, directories, stdin, explicit formats, configuration inspection, rules, doctor, explain, saved-report transformation |
| Reports | Text, JSON, SARIF 2.1.0, standalone HTML and Markdown from one result; schema, escaping and write-error tests |
| Operational behavior | Empty input, cancellation, resource exhaustion and failed output cannot pass; input files remain unchanged |
| Determinism and ownership | Reordered inputs, worker counts and concurrent calls return equivalent results without shared result buffers |
| Project quality | Strict lint, qtlint, nolintguard, architecture/API/module gates and their negative tests; reproducible `make check` |
| Native platforms and release | Linux/macOS/Windows test results, minimum Go compiler without auto-upgrade, release binaries, checksums and SBOM |

Acceptance is unproven until the listed evidence exists for the released commit.
Passing a subset of local unit tests does not establish release readiness.

## Stage 2: context and integration policy

- Extend and tune contextual signals using a documented evaluation corpus.
- Add the bounded custom rule DSL: phrase, RE2, token/POS sequences, gaps,
  positions, counters, density, Boolean combinations, exceptions and shared features.
- Add local inheritance, ordered file overrides, dictionary terms and explicit term exemptions.
- Add reasoned next-sentence, next-block and paired-region suppressions, rejecting
  unknown IDs, missing targets, missing reasons and unused directives.
- Add explicit baseline create/update/check with stable structural fingerprints.
  Preserve raw findings and scores when accepting existing debt.
- Add committed changed-unit analysis from merge-base to HEAD, clean-source
  verification, whole-paragraph rescanning and document context for repetition.
- Add trusted base-policy comparison and policy-change handling. Document the CI
  trust boundary when a pull request can modify its own checker or workflow.
- Add a compile-tested `go/analysis` adapter with `analysistest`.

## Stage 3: calibrated revision probability

- Publish an editorial annotation rubric, data provenance and licenses.
- Collect at least 5,000 labeled sentence/paragraph units, including good
  AI-assisted and poor human-written prose, with at least two annotators.
- Reserve at least 1,000 held-out units. Split by document, repository and template
  family; keep related rewrites together and separate training, development,
  calibration and final evaluation.
- Implement reproducible Go training, regularized logistic inference and separate
  calibration. Record feature contracts, model/NLP hashes, seeds and split manifests.
- Publish precision/recall, clean-block false positives, Brier comparison,
  reliability plots and uncertainty. Target ECE at most 0.05 on a declared protocol.
- Return null with an applicability reason for missing, incompatible, short or
  unsupported-domain cases. Add probability gating only with a compatible model.

These requirements cannot be replaced by a sigmoid applied to the heuristic index.

## Stage 4: first product release

- Complete the full specification's 13 final acceptance criteria, including stages 2 and 3.
- Validate stable rules with at least 15 positive and 15 negative examples each and
  corpus evidence. Target 98% precision for hard defaults and 85% for soft defaults,
  with sample sizes and uncertainty; measure clean-block false positives separately.
- Meet the stated coverage goals: 90% for engine/rules/scoring/config/source mapping
  and 85% overall, without hiding difficult packages.
- Publish repeatable performance measurements against the proposed 100,000 words
  in 10 seconds and 512 MiB target on a specified 2-vCPU Linux host.
- Verify SARIF import in a real consumer, release reproducibility, API compatibility,
  all report/input formats and platform behavior.
- Complete the full documentation set, model/data notices and release inventory.

The remaining catalog includes additional filler/hype signals, repeated rhetorical
patterns, nominalization/noun stacks, parenthetical load, n-grams, template and
paragraph overlap, heading/summary echoes, readability and experimental formatting.
Dependency-based passive, long-subject and nested-clause rules require a real
dependency backend. They must never silently fall back to surface approximations.
