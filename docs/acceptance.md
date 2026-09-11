# Acceptance audit

This page records the state of every acceptance requirement the project states,
with a link to evidence a reader can reproduce. It is a status record, not a
claim that the product is finished. Requirements that need human-labeled data or
a live release stay open here until that data or that release exists.

Each row uses one of three states.

- **Met** means evidence exists for a merged commit and the repository can
  reproduce it.
- **Partly met** means the implementation and its tests are complete, while one
  named part of the evidence is still missing.
- **Open** means the requirement is not satisfied.

## Stages 0 and 1: alpha acceptance

| Requirement | State | Evidence |
| --- | --- | --- |
| Public repository and project setup | Met | [GitHub settings](repository-settings.md), `LICENSE`, `CONTRIBUTING.md`, `SECURITY.md`, pinned actions and tool modules |
| Standalone public Go API | Met | `examples/consumer` in CI, the [API policy](public_api.md), and [`scripts/verify-published-module.sh`](../scripts/verify-published-module.sh) with its [recorded check](release/v0.1.0-alpha.1-module-check.json) |
| Pure Go offline runtime | Met | Native Linux, macOS and Windows jobs; `CGO_ENABLED=0` release builds; [SBOM and notices](sbom.md) |
| Source mapping | Met | `extract` tests, including escape decoding and scalar styles, and the annotated `e2e/testdata` fixtures |
| NLP baseline | Met | `nlp` tests and [ADR 0001](adr/0001-source-mapping-and-nlp.md); capability failures are explicit |
| Required rule catalog | Met | `builtin` rules with positive, negative and boundary examples; [scoring](scoring.md) documents what a finding does and does not claim |
| Index and local gate | Met | [scoring](scoring.md) and the root engine tests for caps, nondilution and threshold diagnostics |
| Strict configuration | Met | `config` validation tests reject unknown fields, unknown rule IDs, out-of-range parameters, duplicates and unavailable models before analysis |
| CLI | Met | `internal/cli` tests and the `e2e` scenarios for files, directories, stdin, explicit formats, inspection and saved-report transformation |
| Reports | Met | [reports](reports.md), [SARIF](sarif.md) and the end-to-end format checks on the three native platforms |
| Operational behavior | Met | Cancellation, byte and block limits, empty input and failed-output tests in `extract`, `internal/cli` and the root package |
| Determinism and ownership | Met | Root package tests for reordered inputs, worker counts and concurrent calls |
| Project quality | Partly met | `make check`, the repository policy tests and their negative cases; race detection, active fuzzing and coverage collection stay deferred to [#123](https://github.com/stokaro/unswell/issues/123), the last roadmap item |
| Native platforms and release | Met | The three native CI jobs, the [release audit](release/v0.1.0-alpha.1-audit.json) and [reproducible builds](reproducible-builds.md) |
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
| Extended signal catalogs | Partly met | The [editorial](editorial-patterns.md), [repetition](repetition-signals.md) and [surface](surface-signals.md) catalogs are implemented and opt-in; their corpus qualification needs the data in stage 3 |

## Stage 3: calibrated revision probability

| Requirement | State | Evidence |
| --- | --- | --- |
| Annotation rubric, provenance and licenses | Met | [editorial annotation](editorial-annotation.md) and [ADR 0013](adr/0013-annotation-protocol.md) |
| Labeled corpus of at least 5,000 units | Open | The protocol, tooling and validation exist; no human-labeled corpus has been collected. [#22](https://github.com/stokaro/unswell/issues/22) is on hold for resource reasons with this requirement unchanged |
| Held-out splits by document, repository and template family | Partly met | [ADR 0014](adr/0014-corpus-acquisition.md) and the corpus commands freeze the partitions; no real corpus has been split |
| Reproducible Go training and separate calibration | Met | [ADR 0020](adr/0020-logistic-numerical-core.md), [ADR 0021](adr/0021-isotonic-calibration.md), [ADR 0025](adr/0025-corpus-training.md) and [training and evaluation](training.md) |
| Published probability evaluation | Deferred | The harness reports the #25 metrics, a risk-coverage curve and generated figures, and measures its own stage cost; the numbers need the human-labeled corpus, on hold with [#22](https://github.com/stokaro/unswell/issues/22) and [#25](https://github.com/stokaro/unswell/issues/25) |
| Applicability and calibrated gating | Met | [ADR 0034](adr/0034-probability-pack.md) and [scoring](scoring.md) define the pack, the statuses and the gate; no accepted pack exists, so revision probability stays unavailable with a reason and no build is gated, which is the required behavior |

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
| C. Comparable experiment | In progress | Amendment 1 of the protocol runs generation through session agents with no paid call. The [pilot record](../research/generation/runs/2026-09-11-pilot/README.md) and the [second run](../research/generation/runs/2026-09-11-run2/README.md) hold 800 saved responses of one family on 200 tasks. Each run keeps its controlled shards and paired tables. A second family and protocol version 2 remain open |
| D. Confirmatory study | Open | Waits on stage C |
| E. Evidence release | Open | Waits on stage D; the stage B records and their digests are published already |

## Next stage: diagnostics on real texts

[ADR 0037](adr/0037-diagnostics-not-authorship.md) names this stage and
the [roadmap](roadmap.md) carries its table. The rows below record its
state.

| Item | State | Evidence |
| --- | --- | --- |
| Regression fixtures per rule | Met | `builtin/testdata/rewrites-v1.json` holds one pair for every builtin rule: a text that carries the construction and a rewrite that keeps its identifiers, numbers, and at least half of its words. `TestRewriteFixturesStopFiringAfterTheRewrite` in the root package asserts that the rule fires on the first and stays silent on the second in ordinary CI |
| Role stratum | Met | Every row of the pattern tables carries `roles`, the same counts and interval per document role; the [role run](../research/acquisition/runs/2026-09-11-roles/README.md) records the tables of the current measurement with it |
| Frequency command | Met | `corpus frequencies` counts constructions per cohort and role and contrasts each stratum with its baseline; the [frequency run](../research/acquisition/runs/2026-09-11-frequencies/README.md) records the output on the current corpus |
| Review on real texts | Partly met | The [Ptah review](../research/reviews/2026-09-11-ptah/README.md) runs every catalog rule on one real repository and judges a sample of ten findings per rule; the 800 controlled responses and this repository's own documentation remain |
| Sources beyond repositories | Open | The corpus holds GitHub repositories only |

## Stage 4: first product release

| Requirement | State | Evidence |
| --- | --- | --- |
| Stable rule qualification against a real corpus | Deferred | Needs a human decision per finding on a labeled corpus, on hold with [#22](https://github.com/stokaro/unswell/issues/22) and [#26](https://github.com/stokaro/unswell/issues/26); the alpha claims no precision figure |
| Coverage and failure-path targets | Partly met | Failure-path tests and the [measurement recorded on #27](https://github.com/stokaro/unswell/issues/27): configuration 91.9%, engine and scoring 91.9%, rules 92.7%, source mapping 90.9%, 86.7% overall; the CI gate that enforces them is deferred to #123 |
| Reproducible performance and resource limits | Met | [scan cost](performance.md): the synthetic corpus and three real trees meet the target on the 2-vCPU host after [#181](https://github.com/stokaro/unswell/issues/181); the earlier records keep the failure the bounds removed. The published `v0.1.0-alpha.1` binaries and an experimental origin model have records there too |
| SARIF consumers, reproducible releases and formats | Met | [SARIF](sarif.md), [reproducible builds](reproducible-builds.md), the [release audit](release/v0.1.0-alpha.1-audit.json) and the format checks in [reports](reports.md) |
| Documentation set | Met | [configuration](configuration.md), [custom rules](custom-rules.md), [API policy](public_api.md), [MCP](mcp.md), [containers](containers.md), [research](research.md), [editorial annotation](editorial-annotation.md), [training and evaluation](training.md), [reports](reports.md) and [installation](installation.md) |
| Release inventory and notices | Met | [SBOM and notices](sbom.md) and the [release audit](release/v0.1.0-alpha.1-audit.json) |
| Acceptance audit | Met | This page, kept current with the [roadmap](roadmap.md) |
| Race detection, active fuzzing and coverage in CI | Open | Deferred by design to [#123](https://github.com/stokaro/unswell/issues/123), which runs after every other issue has a recorded disposition |

## What this alpha does not claim

The rule catalog is experimental. No precision or recall figure is published for
any rule, because that number requires the labeled corpus stage 3 defines. A
revision probability is unavailable without an accepted pack, and an absent
estimate carries a reason rather than a zero. The origin channel never gates a
build. Distribution automation remains unproven until a real release exercises
it end to end.
