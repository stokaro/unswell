# Acceptance audit

This page records the state of every acceptance requirement the project states,
with a link to evidence a reader can reproduce. It is a status record, not a
claim that the product is finished. Implemented tooling, published artifacts,
research observations, diagnostic review and model qualification have separate
acceptance criteria.

This September 13, 2026 snapshot uses merged source through
[`06b3861`](https://github.com/stokaro/unswell/commit/06b38615d4483970b65a6fbe2187aeb4864830dc),
including the [protocol freeze](https://github.com/stokaro/unswell/pull/229),
[performance records](https://github.com/stokaro/unswell/pull/232), and
[TypeScript repair](https://github.com/stokaro/unswell/pull/233) and
[Markdown repair](https://github.com/stokaro/unswell/pull/235). Published-release
claims refer to alpha.3 at `2a2a6d4`, as recorded in the
[distribution evidence](alpha-distribution-evidence.md). Later engineering does
not change those release results. Local generation runs are not evidence here.

The alpha.4 supplement below records the published release at `7e5b933`,
its reproducibility audit, and complete reruns of the pinned performance inputs.
The historical source and research snapshot above remains separately dated.

Each row uses one of five states.

- **Met** means evidence exists for a merged commit and the repository can
  reproduce it.
- **Partly met** means some required implementation or evidence exists, with
  specific remaining work identified.
- **Open** means the requirement is unsatisfied and remains in scope.
- **In progress** means work has started but acceptance is not complete.
- **Deferred** means the requirement is unsatisfied and a recorded decision puts
  it on hold. Its criteria remain intact; deferral is not completion.

Under [ADR 0037](adr/0037-diagnostics-not-authorship.md), the active release path
improves explainable diagnostics. [ADR 0041](adr/0041-assistant-review-acceptance.md)
records the September 17 decision to accept assistant review for diagnostic
changes without an independent human reviewer. Calibrated probabilities, origin
analysis and detector comparisons remain deferred. #123 has its own explicit
start restriction and remains required for final validation.

## Stages 0 and 1: alpha acceptance

| Requirement | State | Evidence |
| --- | --- | --- |
| Public repository and project setup | Met | [GitHub settings](repository-settings.md), `LICENSE`, `CONTRIBUTING.md`, `SECURITY.md`, pinned actions and tool modules |
| Standalone public Go API | Met | `examples/consumer` in CI, the [API policy](public_api.md), and [`scripts/verify-published-module.sh`](../scripts/verify-published-module.sh) with its [recorded check](release/v0.1.0-alpha.3-module-check.json) |
| Pure Go offline runtime | Met | Native Linux, macOS and Windows jobs; `CGO_ENABLED=0` release builds; [SBOM and notices](sbom.md) |
| Source mapping | Met | `extract` tests and annotated `e2e/testdata` fixtures cover source ranges. [Alpha.4 reruns](performance/alpha4/README.md) process the original FastAPI and date-fns failures after #230/#231, with unchanged source identities and recorded extraction exclusions |
| NLP baseline | Met | `nlp` tests and [ADR 0001](adr/0001-source-mapping-and-nlp.md); capability failures are explicit |
| Required rule catalog | Met | `builtin` rules with positive, negative and boundary examples; [scoring](scoring.md) documents what a finding does and does not claim |
| Index and local gate | Met | [scoring](scoring.md) and the root engine tests for caps, nondilution and threshold diagnostics |
| Strict configuration | Met | `config` validation tests reject unknown fields, unknown rule IDs, out-of-range parameters, duplicates and unavailable models before analysis |
| CLI | Met | `internal/cli` tests and the `e2e` scenarios for files, directories, stdin, explicit formats, inspection and saved-report transformation |
| Reports | Met | [reports](reports.md), [SARIF](sarif.md) and the end-to-end format checks on the three native platforms |
| Operational behavior | Met | Cancellation, byte and block limits, empty input and failed-output tests in `extract`, `internal/cli` and the root package |
| Determinism and ownership | Met | Root package tests for reordered inputs, worker counts and concurrent calls |
| Project quality | Partly met | `make check`, the repository policy tests and their negative cases; race detection, active fuzzing and coverage collection stay deferred to [#123](https://github.com/stokaro/unswell/issues/123), the last roadmap item |
| Native platforms and release | Met | The three native CI jobs, the [release audit](release/v0.1.0-alpha.3-audit.json) and [reproducible builds](reproducible-builds.md) |
| Dogfooding | Met | The built CLI checks the repository's own Markdown, Go and Bash in every CI run; the MCP self-check compares the same evidence |

## Stage 2: context and integration policy

| Requirement | State | Evidence |
| --- | --- | --- |
| Bounded custom rule DSL | Met | [custom rules](custom-rules.md), `ruleset` tests and [ADR 0002](adr/0002-bounded-custom-rules.md) |
| Inheritance, file overrides and terminology | Met | [configuration](configuration.md) and [ADR 0003](adr/0003-configuration-bundles.md) |
| Reasoned structural suppressions | Met | [suppressions](suppressions.md) and [ADR 0004](adr/0004-structural-suppressions.md) |
| Explicit baseline debt | Met | [baselines](baseline.md) and [ADR 0005](adr/0005-baseline-debt.md) |
| Committed changed-unit analysis | Met | [changes](changes.md) and [ADR 0006](adr/0006-committed-changes.md) |
| Trusted base policy | Met | [trusted policy](trusted-policy.md) and [ADR 0007](adr/0007-trusted-policy.md) |
| Public `go/analysis` adapter | Met | [Go analysis adapter](go-analysis.md) and [ADR 0008](adr/0008-go-analysis-adapter.md) |
| Extended signal catalogs | Partly met | The [editorial](editorial-patterns.md), [repetition](repetition-signals.md), and [surface](surface-signals.md) catalogs are implemented. The replay in [#217](https://github.com/stokaro/unswell/issues/217) and the [complete-document study](../research/generation/studies/long-prose-v3/results.md) provide construction evidence. Rule acceptance proceeds with maintainer-accepted assistant review under [ADR 0041](adr/0041-assistant-review-acceptance.md); per-rule evidence remains required in [#26](https://github.com/stokaro/unswell/issues/26) |

## Stage 3: calibrated revision probability

| Requirement | State | Evidence |
| --- | --- | --- |
| Annotation rubric, provenance and licenses | Met | [editorial annotation](editorial-annotation.md) and [ADR 0013](adr/0013-annotation-protocol.md) |
| Labeled corpus of at least 5,000 units | Deferred | The protocol, tooling and validation exist; no human-labeled corpus has been collected. [#22](https://github.com/stokaro/unswell/issues/22) is on hold for resource reasons with this requirement unchanged |
| Held-out splits by document, repository and template family | Partly met | [ADR 0014](adr/0014-corpus-acquisition.md) and the corpus commands freeze partitions. Historical and generated cohorts use them, but the human-labeled editorial corpus and its final test remain absent under #22 and [#25](https://github.com/stokaro/unswell/issues/25) |
| Reproducible Go training and separate calibration | Met | Engineering only: [ADR 0020](adr/0020-logistic-numerical-core.md), [ADR 0021](adr/0021-isotonic-calibration.md), [ADR 0025](adr/0025-corpus-training.md) and [training and evaluation](training.md) document tested Go tools. They do not supply an accepted editorial model |
| Published probability evaluation | Deferred | The harness reports the #25 metrics, a risk-coverage curve and generated figures, and measures its own stage cost; the numbers need the human-labeled corpus, on hold with [#22](https://github.com/stokaro/unswell/issues/22) and [#25](https://github.com/stokaro/unswell/issues/25) |
| Applicability and calibrated gating | Met | Engineering only: [ADR 0034](adr/0034-probability-pack.md) and [scoring](scoring.md) define the pack, statuses, and gate. No accepted pack exists, so ordinary scans have no revision probability or probability gate. An explicitly required but unavailable estimate cannot pass |

Quality and origin stay independent. The [origin channel](adr/0035-origin-channel.md)
is opt-in, ungated and experimental. [ADR 0036](adr/0036-llm-pattern-evidence.md)
defines the active research branch in
[#154](https://github.com/stokaro/unswell/issues/154) to measure pattern
prevalence by cohort. Its measurements are association evidence that
satisfies no row in this table; the stages below record their state.

### Research stages of #154

| Stage | State | Evidence |
| --- | --- | --- |
| A. Scope and methodology | Met | [ADR 0036](adr/0036-llm-pattern-evidence.md), the [protocol](../research/methods/llm-patterns-v1.md), the [sources record](../research/methods/llm-patterns-sources-v1.json), the [prompts](../research/methods/prompts/README.md); #22 on hold |
| B. Historical corpus | Met | Five dated cohorts of 43 repositories under one global plan, verifiable shards, baseline measurements, first-appearance and one-count-per-text analyses, and placebo comparisons in the [run records](../research/acquisition/README.md) |
| C. Comparable experiment | Met | The [generation index](../research/generation/README.md) retains twenty earlier runs and their development screenings. The separate [version 3 study](../research/generation/studies/long-prose-v3/README.md) freezes 38 complete documents from 20 fresh repositories, four arms, three model families, and 456 requests. The original input freeze and the declared dispatch-reservation amendment are both published |
| D. Confirmatory study | Met | [Version 3 results](../research/generation/studies/long-prose-v3/results.md) execute all 12 primary tests with 20 measured groups in every arm. There are 455 complete responses and one retained Qwen truncation. ARI meets the frozen cross-family association criterion but exceeds the historical review-load limit; the other three primary constructions do not establish cross-family association. No blocking rule is qualified. Version 2 remains the earlier underpowered result with 16 groups; its null results and freeze are unchanged |
| E. Evidence release | Met | [Version 1](../research/report-v1.md), published in [#287](https://github.com/stokaro/unswell/pull/287), retains the historical corpus, temporal comparisons, and earlier null results. The [version 3 record](../research/generation/studies/long-prose-v3/README.md) adds permitted frozen inputs and outputs, all measurements, 40 evidence cards, 118 source-bound cases, a data card, and offline reproduction of the complete analysis. The dispatch amendment and actual token usage are explicit |

## Next stage: diagnostics on real texts

[ADR 0037](adr/0037-diagnostics-not-authorship.md) names this stage and
the [roadmap](roadmap.md) carries its table. The rows below record its
state.

| Item | State | Evidence |
| --- | --- | --- |
| Regression fixtures per rule | Met | `builtin/testdata/rewrites-v1.json` holds one pair for every builtin rule: a text that carries the construction and a rewrite that keeps its identifiers, numbers, and at least half of its words. `TestRewriteFixturesStopFiringAfterTheRewrite` in the root package asserts that the rule fires on the first and stays silent on the second in ordinary CI |
| Role stratum | Met | Every row of the pattern tables carries `roles`, the same counts and interval per document role; the [role run](../research/acquisition/runs/2026-09-11-roles/README.md) records the tables of the current measurement with it |
| Frequency command | Met | `corpus frequencies` counts constructions per cohort and role and contrasts each stratum with its baseline; the [frequency run](../research/acquisition/runs/2026-09-11-frequencies/README.md) records the output on the current corpus |
| Review on real texts | Partly met | The dated reviews cover [Ptah](../research/reviews/2026-09-15-contextual/README.md), [complete technical documents and generated responses](../research/generation/studies/long-prose-v3/inspection.md), and [this repository](../research/reviews/2026-09-11-unswell/README.md). The [offline fixtures](../e2e/testdata/long_prose_research/README.md) preserve concrete edits and technical counterexamples. The [whole-page audit](../research/reviews/2026-09-17-full-page-recall/README.md) adds missed-event recall and complete finding dispositions. The maintainer accepted its assistant judgments under ADR 0041. The [contextual wording follow-up](../research/reviews/2026-09-17-context-recall/README.md) raises detection on the twelve exposed pages from 1/78 to 7/78. Its eleven new pages remain at 1/42; broad recall and #305 remain open |
| Sources beyond repositories | Met | `scripts/acquire-documents.sh` brings published specifications into the historical cohort under the corpus contract; the [document run](../research/acquisition/runs/2026-09-11-documents/README.md) records three RFC sets of eight texts with their tables |

## Stage 4: first product release

| Requirement | State | Evidence |
| --- | --- | --- |
| Stable rule qualification against a real corpus | Open | [#26](https://github.com/stokaro/unswell/issues/26) accepts assistant-reviewed evidence under [ADR 0041](adr/0041-assistant-review-acceptance.md). Each rule still needs contextual positives/negatives, confirmation results and a recorded default decision. Independent human review is no longer a blocker; the whole-page audit does not qualify all defaults |
| Coverage and failure-path targets | Partly met | Failure-path tests and the [measurement recorded on #27](https://github.com/stokaro/unswell/issues/27): configuration 91.9%, engine and scoring 91.9%, rules 92.7%, source mapping 90.9%, 86.7% overall; the CI gate that enforces them is deferred to #123 |
| Reproducible performance and resource limits | Met | [Alpha.4 measurements](performance/alpha4/README.md): all 24 scans complete within 10 seconds and 512 MiB under a two-CPU quota. Synthetic, pytest, and FastAPI each exceed 100,000 prose words. Date-fns is smaller; its row establishes completeness and measured cost, not the word-count target. Historical failures and coverage changes remain recorded |
| SARIF consumers, reproducible releases and formats | Met | [SARIF](sarif.md), [reproducible builds](reproducible-builds.md), the [release audit](release/v0.1.0-alpha.3-audit.json) and the format checks in [reports](reports.md) |
| Documentation set | Met | [configuration](configuration.md), [custom rules](custom-rules.md), [API policy](public_api.md), [MCP](mcp.md), [containers](containers.md), [research](research.md), [editorial annotation](editorial-annotation.md), [training and evaluation](training.md), [reports](reports.md) and [installation](installation.md) |
| Release inventory and notices | Met | [SBOM and notices](sbom.md) and the [release audit](release/v0.1.0-alpha.3-audit.json) |
| Acceptance audit | Met | This page, kept current with the [roadmap](roadmap.md) |
| Race detection, active fuzzing and coverage in CI | Deferred | [#123](https://github.com/stokaro/unswell/issues/123) is on hold and may start only on an explicit maintainer request, after every other implementation task has a recorded disposition. Prior measurements do not replace validation of the final tree |

## What this alpha does not claim

The rule catalog is experimental. Diagnostic review metrics describe agreement
with the named reviewer on the declared sample. The whole-page audit publishes
recall against assistant annotations; it establishes neither population accuracy
nor independent human agreement. A revision probability is unavailable without
an accepted pack, and an absent estimate carries a reason rather than a zero. The origin channel never gates a
build. Distribution automation completed the alpha.3 cycle documented in
[#66](https://github.com/stokaro/unswell/issues/66) and the
[release evidence](alpha-distribution-evidence.md). That success does not verify
future releases. [Alpha.4 evidence](release/v0.1.0-alpha.4.md) adds complete corpus scans and
resource measurements for #219. Tap and Action update reviews remain pending;
final validation remains deferred in #123.
