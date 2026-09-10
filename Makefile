SHELL := /bin/bash

.PHONY: check test race fuzz lint lint-shell tidy policy build build-mcp release fmt schema dogfood dogfood-mcp
.PHONY: check-registry check-mirror-policy check-sbom-policy check-research-cost research-cost

check: policy tidy test lint lint-shell check-mirror-policy check-sbom-policy schema dogfood-mcp check-performance check-research-cost check-reproducible

check-sbom-policy:
	bash scripts/prepare-sbom.sh --self-test

check-mirror-policy:
	bash scripts/check-image-mirrors.sh --self-test

check-performance: build
	bash scripts/measure-performance.sh --words 2000 --self-test

check-research-cost:
	bash scripts/measure-research-cost.sh --self-test

research-cost:
	bash scripts/measure-research-cost.sh --label local

performance: build
	bash scripts/measure-performance.sh --words 100000 --label local

check-reproducible:
	bash scripts/verify-reproducible-build.sh

audit-release:
	bash scripts/verify-release-artifacts.sh

check-published-module:
	bash scripts/verify-published-module.sh --platforms "linux/amd64"

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
	cd mcp && go run ./cmd/mcp-selfcheck --expected ../artifacts/dogfood/result.json \
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
