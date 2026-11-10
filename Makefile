SHELL := /bin/bash

.PHONY: check test race fuzz lint lint-shell tidy policy build build-mcp release fmt schema dogfood dogfood-mcp
.PHONY: check-registry check-mirror-policy check-sbom-policy check-research-cost research-cost
.PHONY: check-contextual-admission check-external-controls check-current-baseline check-current-editorial-review
.PHONY: check-source-review-contract check-source-bound-disposition-pilot check-paired-source-necessity check-complementary-semantic-pilot
.PHONY: check-unit-relation-pilot check-source-role-rescue check-source-role-selector check-structured-admission check-source-context-admission check-joint-coverage-triage
.PHONY: check-ape-screen check-dual-gate-triage check-dual-gate-feasibility check-length-triage-audit check-scoped-degree check-quality-claim-screen check-hope-framing-screen check-contextual-unit-slices check-span-example-screen check-aiwritingeffect-screen check-arb-transfer check-heterogeneous-spans-transfer check-gen-edit-transfer check-interpretable-ensemble check-counterfactual-admission check-evidence-flip-pilot check-real-source-transfer check-rubric-transfer check-operator-triage check-source-function-pilot check-exhaustive-inventory-pilot

check: policy tidy test lint lint-shell check-mirror-policy check-sbom-policy schema dogfood-mcp check-performance check-research-cost check-reproducible check-contextual-admission check-source-review-contract check-external-controls check-current-baseline check-current-editorial-review check-ape-screen check-dual-gate-triage check-dual-gate-feasibility check-length-triage-audit check-scoped-degree check-quality-claim-screen check-hope-framing-screen check-contextual-unit-slices check-span-example-screen check-aiwritingeffect-screen check-arb-transfer check-heterogeneous-spans-transfer check-gen-edit-transfer check-gen-technical-transfer check-interpretable-ensemble check-counterfactual-admission check-evidence-flip-pilot check-real-source-transfer check-rubric-transfer check-operator-triage check-source-function-pilot check-exhaustive-inventory-pilot check-source-bound-disposition-pilot check-paired-source-necessity check-complementary-semantic-pilot check-unit-relation-pilot check-source-role-rescue check-source-role-selector check-structured-admission check-source-context-admission check-joint-coverage-triage

check-joint-coverage-triage:
	python3 -B research/reviews/2026-11-09-joint-coverage-triage/run.py verify
	python3 -B -m unittest discover -s research/reviews/2026-11-09-joint-coverage-triage -p 'test_*.py'

check-source-context-admission:
	python3 -B research/reviews/2026-11-09-source-context-admission/verify_result.py
	python3 -B -m unittest discover -s research/reviews/2026-11-09-source-context-admission -p 'test_*.py'

check-structured-admission:
	python3 -B research/reviews/2026-11-09-structured-admission/verify_result.py
	python3 -B -m unittest discover -s research/reviews/2026-11-09-structured-admission -p 'test_*.py'

check-source-role-selector:
	python3 -B research/reviews/2026-11-09-source-role-selector/verify.py
	python3 -B -m unittest discover -s research/reviews/2026-11-09-source-role-selector -p 'test_*.py'

check-source-role-rescue:
	python3 -B research/reviews/2026-11-08-source-role-rescue/verify.py
	python3 -B research/reviews/2026-11-08-source-role-rescue/verify_baseline.py
	python3 -B -m unittest discover -s research/reviews/2026-11-08-source-role-rescue -p 'test_*.py'

check-unit-relation-pilot:
	python3 -B research/reviews/2026-11-08-unit-relation-pilot/verify_terminal.py --check
	python3 -B -m unittest discover -s research/reviews/2026-11-08-unit-relation-pilot -p 'test_*.py'
	python3 -B research/reviews/2026-11-08-unit-relation-pilot-v2/verify_terminal.py --check
	python3 -B -m unittest discover -s research/reviews/2026-11-08-unit-relation-pilot-v2 -p 'test_*.py'

check-complementary-semantic-pilot:
	python3 -B research/reviews/2026-11-08-complementary-semantic-pilot/verify_result.py
	python3 -B -m unittest discover -s research/reviews/2026-11-08-complementary-semantic-pilot -p 'test_*.py'

check-dual-gate-feasibility:
	python3 -B research/reviews/2026-11-08-dual-gate-feasibility/analyze.py --check
	python3 -B research/reviews/2026-11-08-dual-gate-feasibility/category_frontier.py
	python3 -B -m unittest discover -s research/reviews/2026-11-08-dual-gate-feasibility -p 'test_*.py'

check-paired-source-necessity:
	python3 -B research/reviews/2026-11-08-paired-source-necessity/run.py verify >/dev/null
	python3 -B research/reviews/2026-11-08-paired-source-necessity/verify_result.py verify
	python3 -B -m unittest discover -s research/reviews/2026-11-08-paired-source-necessity -p 'test_*.py'

check-source-bound-disposition-pilot:
	python3 -B research/reviews/2026-11-08-source-bound-disposition-pilot/verify_result.py
	python3 -B -m unittest discover -s research/reviews/2026-11-08-source-bound-disposition-pilot -p 'test_*.py'

check-exhaustive-inventory-pilot:
	python3 -B research/reviews/2026-11-08-exhaustive-inventory-pilot/verify_result.py
	python3 -B -m unittest discover -s research/reviews/2026-11-08-exhaustive-inventory-pilot -p 'test_*.py'

check-source-function-pilot:
	python3 -B research/reviews/2026-11-08-source-function-pilot/verify_result.py
	python3 -B -m unittest discover -s research/reviews/2026-11-08-source-function-pilot -p 'test_*.py'

check-operator-triage:
	python3 -B research/reviews/2026-11-08-operator-triage/verify_result.py
	python3 -B -m unittest discover -s research/reviews/2026-11-08-operator-triage -p 'test_*.py'

check-rubric-transfer:
	python3 -B research/reviews/2026-11-08-rubric-transfer/run.py verify >/dev/null
	python3 -B research/reviews/2026-11-08-rubric-transfer/verify_result.py >/dev/null
	python3 -B -m unittest discover -s research/reviews/2026-11-08-rubric-transfer -p 'test_*.py'

check-real-source-transfer:
	python3 -B research/reviews/2026-11-08-real-source-admission-transfer/run.py verify >/dev/null
	python3 -B research/reviews/2026-11-08-real-source-admission-transfer-v2/run.py verify >/dev/null
	python3 -B research/reviews/2026-11-08-real-source-admission-transfer-v2/verify_result.py
	python3 -B -m unittest discover -s research/reviews/2026-11-08-real-source-admission-transfer -p 'test_*.py'
	python3 -B -m unittest discover -s research/reviews/2026-11-08-real-source-admission-transfer-v2 -p 'test_*.py'

check-evidence-flip-pilot:
	python3 -B research/reviews/2026-11-08-evidence-flip-pilot/run.py verify >/dev/null
	python3 -B research/reviews/2026-11-08-evidence-flip-pilot/score.py check
	python3 -B research/reviews/2026-11-08-evidence-flip-pilot/verify_result.py
	python3 -B -m unittest discover -s research/reviews/2026-11-08-evidence-flip-pilot -p 'test_*.py'

check-counterfactual-admission:
	python3 -B research/reviews/2026-11-08-counterfactual-admission/verify.py
	python3 -B research/reviews/2026-11-08-counterfactual-admission/verify_result.py
	python3 -B -m unittest discover -s research/reviews/2026-11-08-counterfactual-admission -p 'test_*.py'

check-interpretable-ensemble:
	python3 -B research/origin/runs/2026-11-08-interpretable-ensemble/verify.py
	python3 -B -m unittest discover -s research/origin/runs/2026-11-08-interpretable-ensemble -p 'test_*.py'

.PHONY: check-gen-technical-transfer
check-gen-technical-transfer:
	python3 -B research/origin/runs/2026-11-08-gen-technical-transfer/verify.py
	python3 -B -m unittest discover -s research/origin/runs/2026-11-08-gen-technical-transfer -p 'test_*.py'

check-gen-edit-transfer:
	python3 -B research/origin/runs/2026-11-08-gen-edit-transfer/verify.py
	python3 -B -m unittest discover -s research/origin/runs/2026-11-08-gen-edit-transfer -p 'test_*.py'

check-aiwritingeffect-screen:
	python3 -B research/origin/runs/2026-11-08-aiwritingeffect-screen/verify.py
	python3 -B -m unittest discover -s research/origin/runs/2026-11-08-aiwritingeffect-screen -p 'test_*.py'
check-arb-transfer:
	python3 -B research/origin/runs/2026-11-08-arb-transfer/verify.py
	python3 -B -m unittest discover -s research/origin/runs/2026-11-08-arb-transfer -p 'test_*.py'

check-heterogeneous-spans-transfer:
	python3 -B research/origin/runs/2026-11-08-heterogeneous-spans-transfer/verify.py
	python3 -B -m unittest discover -s research/origin/runs/2026-11-08-heterogeneous-spans-transfer -p 'test_*.py'

check-span-example-screen:
	python3 -B research/origin/runs/2026-11-07-span-example-screen/verify.py
	python3 -B -m unittest discover -s research/origin/runs/2026-11-07-span-example-screen -p 'test_*.py'

check-contextual-unit-slices:
	python3 -B research/reviews/2026-11-07-contextual-unit-slices/unit_slices.py --check
	python3 -B -m unittest discover -s research/reviews/2026-11-07-contextual-unit-slices -p 'test_*.py'

check-hope-framing-screen:
	python3 -B research/reviews/2026-11-07-hope-framing/analyze.py --check
	python3 -B -m unittest discover -s research/reviews/2026-11-07-hope-framing -p 'test_*.py'

check-quality-claim-screen:
	python3 -B research/reviews/2026-11-07-quality-claim-screen/analyze.py --check
	python3 -B -m unittest discover -s research/reviews/2026-11-07-quality-claim-screen -p 'test_*.py'

check-scoped-degree:
	python3 -B research/reviews/2026-11-07-scoped-degree/analyze.py --check
	python3 -B -m unittest discover -s research/reviews/2026-11-07-scoped-degree -p 'test_*.py'

check-dual-gate-triage:
	python3 -B research/reviews/2026-11-07-dual-gate-triage/analyze.py --check
	python3 -B -m unittest discover -s research/reviews/2026-11-07-dual-gate-triage -p 'test_*.py'

check-length-triage-audit:
	python3 -B research/reviews/2026-11-08-length-triage-audit/analyze.py --check
	python3 -B -m unittest discover -s research/reviews/2026-11-08-length-triage-audit -p 'test_*.py'

check-ape-screen:
	python3 -B -m unittest discover -s research/reviews/2026-11-07-ape-proposition-screen -p 'test_*.py'

check-current-editorial-review:
	python3 -B research/reviews/2026-11-07-current-main-editorial-review/verify.py
	python3 -B -m unittest research/reviews/2026-11-07-current-main-editorial-review/test_verify.py

check-current-baseline:
	python3 -B research/reviews/2026-11-07-current-main-baseline/analyze.py --run research/reviews/2026-11-07-current-main-baseline --output research/reviews/2026-11-07-current-main-baseline/result.json --check
	python3 -B -m unittest research/reviews/2026-11-07-current-main-baseline/test_analyze.py

check-external-controls:
	python3 -B research/methods/audits/2026-11-07-external-controls/verify.py
	python3 -B -m unittest research/methods/audits/2026-11-07-external-controls/test_verify.py

check-source-review-contract:
	cd research/annotation/source_review && python3 -B -m unittest test_annotation_contract.py

check-contextual-admission:
	cd research/reviews/2026-10-07-separated-admission && python3 -B verify.py && python3 -B -m unittest test_verify.py

check-sbom-policy:
	bash scripts/prepare-sbom.sh --self-test

check-mirror-policy:
	bash scripts/check-image-mirrors.sh --self-test

check-performance: build
	@temporary=$$(mktemp -d); trap 'rm -rf "$$temporary"' EXIT; \
	  bash scripts/measure-performance.sh --words 2000 --artifacts "$$temporary/evidence" --self-test && \
	  python3 scripts/summarize-performance.py "$$temporary/evidence" >"$$temporary/summary.json" && \
	  ! bash scripts/measure-performance.sh --words 100 --artifacts "$$temporary/evidence" >"$$temporary/overwrite.log" 2>&1
	python3 scripts/summarize-performance.py --self-test

check-research-cost:
	bash scripts/measure-research-cost.sh --self-test

research-cost:
	bash scripts/measure-research-cost.sh --label local

performance: build
	bash scripts/measure-performance.sh --words 100000 --label local

check-reproducible:
	bash scripts/test-reproducible-build.sh
	bash scripts/verify-reproducible-build.sh

audit-release:
	bash scripts/verify-release-artifacts.sh

check-published-module:
	bash scripts/verify-published-module.sh

reproducible:
	bash scripts/verify-reproducible-build.sh

check-registry:
	bash scripts/check-registry.sh

test:
	bash scripts/modules.sh test

# Retained for final validation (#123); deferred from routine alpha development.
race:
	bash scripts/modules.sh test -race

fuzz:
	go test ./feature -run '^$$' -fuzz '^FuzzLexicalKey$$' -fuzztime 10s -parallel 2
	cd research/annotation && go test ./llmdet -run '^$$' -fuzz '^FuzzLoadNumericalPack$$' -fuzztime 10s -parallel 2
	cd research/annotation && go test ./corpus -run '^$$' -fuzz '^FuzzLoadManifest$$' -fuzztime 10s -parallel 2
	cd research/annotation && go test ./corpus -run '^$$' -fuzz '^FuzzLoadArtifact$$' -fuzztime 10s -parallel 2
	cd research/annotation && go test . -run '^$$' -fuzz FuzzLoad -fuzztime 10s -parallel 2
	go test ./nlp -run '^$$' -fuzz FuzzDependencyTree -fuzztime 10s -parallel 2
	go test ./nlp -run '^$$' -fuzz FuzzPrepareUnits -fuzztime 10s -parallel 2
	go test ./baseline -run '^$$' -fuzz FuzzLoad -fuzztime 10s -parallel 2
	go test ./extract -run '^$$' -fuzz Fuzz -fuzztime 10s -parallel 2
	go test ./report -run '^$$' -fuzz Fuzz -fuzztime 10s -parallel 2
	go test ./config -run '^$$' -fuzz Fuzz -fuzztime 10s -parallel 2
	go test ./ruleset -run '^$$' -fuzz FuzzLoad -fuzztime 10s -parallel 2
	go test ./internal/suppress -run '^$$' -fuzz FuzzDirectives -fuzztime 10s -parallel 2

lint:
	bash scripts/lint.sh

lint-shell:
	bash scripts/lint-shell.sh
	bash scripts/lint-shell.sh --self-test

tidy:
	bash scripts/modules.sh tidy -diff

policy:
	go run ./cmd/repocheck
	go test ./internal/repopolicy

schema:
	@temporary=$$(mktemp); trap 'rm -f "$$temporary"' EXIT; go run ./cmd/genschema > "$$temporary" && cmp report/schema.json "$$temporary"

build:
	@commit=$$(git rev-parse HEAD); \
	if [ -n "$$(git status --porcelain)" ]; then commit="$$commit-dirty"; fi; \
	CGO_ENABLED=0 go build -trimpath -ldflags "-X github.com/stokaro/unswell.BuildCommit=$$commit" -o bin/unswell ./cmd/unswell

dogfood: build
	bash scripts/dogfood.sh

build-mcp:
	@commit=$$(git rev-parse HEAD); \
	if [ -n "$$(git status --porcelain)" ]; then commit="$$commit-dirty"; fi; \
	cd mcp && CGO_ENABLED=0 go build -trimpath -ldflags "-X github.com/stokaro/unswell.BuildCommit=$$commit" -o ../bin/unswell-mcp ./cmd/unswell-mcp

dogfood-mcp: dogfood build-mcp
	cd mcp && go run ./cmd/mcp-selfcheck --timeout 5m --expected ../artifacts/dogfood/result.json \
	  --output ../artifacts/dogfood/mcp-result.json -- ../bin/unswell-mcp --config ../.unswell.yaml \
	  --prepared-feature prose-words --prepared-feature noun-token-ratio \
	  --prepared-kind sentence --prepared-kind paragraph --prepared-kind fragment \
	  --feature prose-words --feature type-token-ratio --feature activation/readability.long-paragraph \
	  --feature activation/filler.announced-importance --feature activation/syntax.noun-stack \
	  --feature activation/filler.section-announcement --feature activation/filler.stacked-hedging \
	  --feature activation/syntax.not-only-density --feature activation/syntax.paired-contrast-density \
	  --feature activation/syntax.triad-density --feature activation/syntax.whether-preface-density \
	  --feature activation/syntax.rhetorical-question-density --feature activation/syntax.passive-candidate-density \
	  --feature activation/repetition.exact-sentence --feature activation/repetition.sentence-openers \
	  --feature activation/repetition.paragraph-openers \
	  --feature activation/repetition.near-sentence --feature activation/repetition.ngram-density \
	  --feature activation/repetition.syntax-template --feature activation/repetition.paragraph-overlap \
	  --feature activation/repetition.heading-echo --feature activation/repetition.summary-echo \
	  --feature activation/format.list-fragmentation

release:
	bash scripts/release.sh

fmt:
	@formatter=$$(cd tools && go tool -n golangci-lint); root=$$PWD; \
	while read -r directory role; do \
	  if [ "$$role" != tools ]; then (cd "$$directory" && "$$formatter" fmt --config "$$root/.golangci.yml" ./...) || exit; fi; \
	done <.gomodules
