SHELL := /bin/bash

.PHONY: check test race fuzz lint lint-shell tidy policy api build build-mcp release fmt schema dogfood dogfood-mcp
.PHONY: check-registry check-mirror-policy check-sbom-policy

check: policy tidy test lint lint-shell check-mirror-policy check-sbom-policy api schema dogfood-mcp

check-sbom-policy:
	bash scripts/prepare-sbom.sh --self-test

check-mirror-policy:
	bash scripts/check-image-mirrors.sh --self-test

check-registry:
	bash scripts/check-registry.sh

test:
	bash scripts/modules.sh test

race:
	bash scripts/modules.sh test -race

fuzz:
	cd research/annotation && go test ./corpus -run '^$$' -fuzz '^FuzzLoadManifest$$' -fuzztime 10s -parallel 2
	cd research/annotation && go test ./corpus -run '^$$' -fuzz '^FuzzLoadArtifact$$' -fuzztime 10s -parallel 2
	cd research/annotation && go test . -run '^$$' -fuzz FuzzLoad -fuzztime 10s -parallel 2
	go test ./nlp -run '^$$' -fuzz FuzzDependencyTree -fuzztime 10s -parallel 2
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

api:
	bash scripts/api.sh
	bash scripts/api.sh --self-test

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
	  --feature prose-words --feature type-token-ratio --feature activation/readability.long-paragraph \
	  --feature activation/filler.announced-importance --feature activation/syntax.noun-stack \
	  --feature activation/filler.section-announcement --feature activation/filler.stacked-hedging \
	  --feature activation/syntax.not-only-density --feature activation/syntax.paired-contrast-density \
	  --feature activation/syntax.triad-density --feature activation/syntax.whether-preface-density \
	  --feature activation/syntax.rhetorical-question-density --feature activation/syntax.passive-candidate-density

release:
	bash scripts/release.sh

fmt:
	@formatter=$$(cd tools && go tool -n golangci-lint); root=$$PWD; \
	while read -r directory role; do \
	  if [ "$$role" != tools ]; then (cd "$$directory" && "$$formatter" fmt --config "$$root/.golangci.yml" ./...) || exit; fi; \
	done <.gomodules
