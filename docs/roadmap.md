# Implementation and acceptance roadmap

The September 7, 2026 technical specification defines five stages. The requested
first alpha implements stages 0 and 1. Later stages remain product requirements;
the alpha does not claim completion of the full product specification.

## Issue queue

Work proceeds in the order below. An issue is complete only after its acceptance
evidence passes on the merged commit. Human annotation and measured quality require
real data; implementation alone cannot satisfy those requirements.

| Issue | Work |
| --- | --- |
| [#5](https://github.com/stokaro/unswell/issues/5) | Alpha: enforce Bash syntax, ShellCheck and shfmt in CI |
| [#6](https://github.com/stokaro/unswell/issues/6) | Alpha: parse source languages and Markdown with gotreesitter |
| [#7](https://github.com/stokaro/unswell/issues/7) | Alpha: select prose contexts globally and per language with reasoned exceptions |
| [#8](https://github.com/stokaro/unswell/issues/8) | Alpha: document the AI-style reduction motivation and workflows |
| [#32](https://github.com/stokaro/unswell/issues/32) | MCP: expose offline self-checking through the shared engine |
| [#33](https://github.com/stokaro/unswell/issues/33) | MCP: prove agent self-checking and repository dogfooding |
| [#34](https://github.com/stokaro/unswell/issues/34) | Containers: publish a standalone Unswell CLI image |
| [#35](https://github.com/stokaro/unswell/issues/35) | Containers: publish a separate Unswell MCP image |
| [#36](https://github.com/stokaro/unswell/issues/36) | MCP Registry: publish and verify the Unswell server |
| [#37](https://github.com/stokaro/unswell/issues/37) | Distribution: publish and verify a dedicated Homebrew tap |
| [#38](https://github.com/stokaro/unswell/issues/38) | Distribution: publish and verify the Unswell GitHub Action |
| [#9](https://github.com/stokaro/unswell/issues/9) | Alpha: verify and publish the first prerelease |
| [#10](https://github.com/stokaro/unswell/issues/10) | Stage 2: add the bounded custom rule DSL |
| [#11](https://github.com/stokaro/unswell/issues/11) | Stage 2: add local inheritance, file overrides and terminology |
| [#12](https://github.com/stokaro/unswell/issues/12) | Stage 2: add reasoned structural suppressions |
| [#13](https://github.com/stokaro/unswell/issues/13) | Stage 2: add explicit baseline debt management |
| [#14](https://github.com/stokaro/unswell/issues/14) | Stage 2: analyze committed changes with full paragraph context |
| [#15](https://github.com/stokaro/unswell/issues/15) | Stage 2: enforce trusted base policy in change checks |
| [#16](https://github.com/stokaro/unswell/issues/16) | Stage 2: add a public go/analysis adapter |
| [#17](https://github.com/stokaro/unswell/issues/17) | Stage 2: expand filler, hype and rhetorical signals |
| [#18](https://github.com/stokaro/unswell/issues/18) | Stage 2: expand repetition beyond sentences |
| [#19](https://github.com/stokaro/unswell/issues/19) | Stage 2: add surface syntax and readability signals |
| [#20](https://github.com/stokaro/unswell/issues/20) | NLP: evaluate a real dependency backend and capability contract |
| [#21](https://github.com/stokaro/unswell/issues/21) | Stage 3: define the editorial annotation and data protocol |
| [#22](https://github.com/stokaro/unswell/issues/22) | Stage 3: collect and validate the human-labeled corpus |
| [#23](https://github.com/stokaro/unswell/issues/23) | Stage 3: implement reproducible Go training and calibration |
| [#24](https://github.com/stokaro/unswell/issues/24) | Stage 3: add probability applicability and calibrated gating |
| [#25](https://github.com/stokaro/unswell/issues/25) | Stage 3: publish held-out probability evaluation |
| [#26](https://github.com/stokaro/unswell/issues/26) | Stage 4: qualify stable rules against a real corpus |
| [#27](https://github.com/stokaro/unswell/issues/27) | Stage 4: meet coverage and failure-path acceptance |
| [#28](https://github.com/stokaro/unswell/issues/28) | Stage 4: establish reproducible performance and resource limits |
| [#29](https://github.com/stokaro/unswell/issues/29) | Stage 4: verify SARIF consumers and reproducible release artifacts |
| [#30](https://github.com/stokaro/unswell/issues/30) | Stage 4: close the product documentation and acceptance audit |

## Alpha acceptance: stages 0 and 1

| Requirement | Evidence required before release |
| --- | --- |
| Public repository and project setup | Public GitHub state; license, contribution and security guidance; protected main; pinned CI and tools |
| Standalone public Go API | External consumer module analyzes bytes, registers a custom rule, writes JSON, and passes CI |
| Pure Go offline runtime | Native tests and `CGO_ENABLED=0` release builds; dependency and model notices |
| Source mapping | Exact UTF-8 spans for Markdown, GFM, comments and string literals, CRLF, BOM, Unicode, entities, escapes, and emphasis |
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
| Dogfooding | The built CLI checks owned Markdown, Go code and Bash scripts with the committed strict policy; CI retains reports and proves a negative case fails |

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
