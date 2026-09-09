# Implementation and acceptance roadmap

The September 7, 2026 technical specification defines five stages. The requested
first alpha implements stages 0 and 1. Later stages remain product requirements;
the alpha does not claim completion of the full product specification.

## Issue queue

Work proceeds in the order below. An issue is complete only after its acceptance
evidence passes on the merged commit. Human annotation and measured quality require
real data; implementation alone cannot satisfy those requirements.

During alpha implementation, race detection, active fuzzing, and coverage are deferred.
Their tests and commands remain available. Run #123 last, after the other
implementation tasks have a recorded disposition, to restore these CI checks and
fix discovered issues. An earlier audit must keep this final validation outstanding.
Alpha API compatibility with previous releases is not required; current schemas,
consumer tests, and the public package ledger remain maintained.

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
| [#46](https://github.com/stokaro/unswell/issues/46) | Containers: automate verified Docker Hub mirrors |
| [#36](https://github.com/stokaro/unswell/issues/36) | MCP Registry: publish and verify the Unswell server |
| [#37](https://github.com/stokaro/unswell/issues/37) | Distribution: publish and verify a dedicated Homebrew tap |
| [#38](https://github.com/stokaro/unswell/issues/38) | Distribution: publish and verify the Unswell GitHub Action |
| [#66](https://github.com/stokaro/unswell/issues/66) | Distribution: verify automatic release update PRs, checked merges, and Action publication |
| [#48](https://github.com/stokaro/unswell/issues/48) | Release: distinguish the project license from bundled notices in SBOM metadata |
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
| [#55](https://github.com/stokaro/unswell/issues/55) | Research: fix targets, methodology, and a versioned method registry |
| [#56](https://github.com/stokaro/unswell/issues/56) | Research: share versioned features across rules, training, and inference |
| [#22](https://github.com/stokaro/unswell/issues/22) | Stage 3: collect and validate the human-labeled corpus |
| [#57](https://github.com/stokaro/unswell/issues/57) | Research: execute reproducible comparisons with grouped evaluation |
| [#50](https://github.com/stokaro/unswell/issues/50) | Research: compare lexical and stylometric quality baselines |
| [#23](https://github.com/stokaro/unswell/issues/23) | Stage 3: implement reproducible Go training and calibration |
| [#24](https://github.com/stokaro/unswell/issues/24) | Stage 3: add probability applicability and calibrated gating |
| [#25](https://github.com/stokaro/unswell/issues/25) | Stage 3: publish held-out probability evaluation |
| [#51](https://github.com/stokaro/unswell/issues/51) | Research: evaluate an offline pure-Go LLMDet port |
| [#52](https://github.com/stokaro/unswell/issues/52) | Research: test compression similarity as an optional feature |
| [#53](https://github.com/stokaro/unswell/issues/53) | Research: define optional model-based detector comparisons |
| [#26](https://github.com/stokaro/unswell/issues/26) | Stage 4: qualify stable rules against a real corpus |
| [#27](https://github.com/stokaro/unswell/issues/27) | Stage 4: meet coverage and failure-path acceptance |
| [#28](https://github.com/stokaro/unswell/issues/28) | Stage 4: establish reproducible performance and resource limits |
| [#29](https://github.com/stokaro/unswell/issues/29) | Stage 4: verify SARIF consumers and reproducible release artifacts |
| [#30](https://github.com/stokaro/unswell/issues/30) | Stage 4: close the product documentation and acceptance audit |
| [#123](https://github.com/stokaro/unswell/issues/123) | Final task: restore race detection, active fuzzing, and coverage; fix findings and record acceptance evidence |

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

The [alpha distribution evidence](alpha-distribution-evidence.md) records the
published version and its verified installation paths. Automatic updates remain
open in #66; later changes on main do not alter the published alpha.

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

The [published annotation protocol](editorial-annotation.md) and its
[research tools](../research/annotation/README.md) support collection and review.
Real human annotation, the corpus minimums, and final held-out evidence remain
acceptance requirements; the tutorial fixtures do not satisfy them.

The [corpus preparation tool](../research/annotation/corpus/README.md) supplies
source manifests, grouped partition plans, original ranges, and re-extraction
verification for #22. Its 378 Ptah candidates remain unlabeled and entirely in
development. The human-labeled corpus and final-test acceptance remain open.

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
The [research plan](research.md) defines baseline comparisons, feature ablations,
and optional detector experiments. Quality and provenance labels stay independent.
Provenance research does not determine the initial CI gate, and none of these
experiments is reported as a completed benchmark.

## Stage 4: first product release

- Complete the full specification's 13 final acceptance criteria, including stages 2 and 3.
- Validate stable rules with at least 15 positive and 15 negative examples each and
  corpus evidence. Target 98% precision for hard defaults and 85% for soft defaults,
  with sample sizes and uncertainty; measure clean-block false positives separately.
- Meet the stated coverage goals: 90% for engine/rules/scoring/config/source mapping
  and 85% overall, without hiding difficult packages.
- Publish repeatable performance measurements against the proposed 100,000 words
  in 10 seconds and 512 MiB target on a specified 2-vCPU Linux host.
- Verify SARIF import in a real consumer, release reproducibility, current API consumers,
  all report/input formats and platform behavior.
- Complete the full documentation set, model/data notices and release inventory.
- Finally, restore race detection, active fuzzing, and coverage in CI, fix findings, and retain
  passing evidence for the merged commit as required by #123.

Additional filler/hype signals and repeated rhetorical patterns are available as
[opt-in experiments](editorial-patterns.md); their corpus qualification remains open.
[Repetition experiments](repetition-signals.md) cover n-grams, templates, paragraph
overlap, and heading/summary echoes, with the same qualification requirement.
[Surface experiments](surface-signals.md) cover nominalization, noun stacks,
passive candidates, parenthetical load, readability, and formatting. These are
unqualified candidates; they remain disabled in builtin profiles.
The [dependency evaluation](../research/dependencies/README.md) records GoSpacy
reference parity, component terms, and adapter limitations. The engine now validates
requested trees and label schemes; the builtin provider remains surface-only.
Dependency-based passive, long-subject and nested-clause rules require a real
dependency backend. They must never silently fall back to surface approximations.
