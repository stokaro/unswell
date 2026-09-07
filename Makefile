SHELL := /bin/bash

.PHONY: check test race fuzz lint tidy policy api build release fmt schema

check: policy tidy test lint api schema

test:
	bash scripts/modules.sh test

race:
	bash scripts/modules.sh test -race

fuzz:
	go test ./extract -run '^$$' -fuzz Fuzz -fuzztime 10s -parallel 2
	go test ./report -run '^$$' -fuzz Fuzz -fuzztime 10s -parallel 2
	go test ./config -run '^$$' -fuzz Fuzz -fuzztime 10s -parallel 2

lint:
	bash scripts/lint.sh

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
	CGO_ENABLED=0 go build -trimpath -o bin/unswell ./cmd/unswell

release:
	bash scripts/release.sh

fmt:
	cd tools && go tool golangci-lint fmt --config ../.golangci.yml ..
