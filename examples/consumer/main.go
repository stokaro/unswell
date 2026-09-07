// Command consumer demonstrates the public Unswell extension API from another module.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/report"
	"github.com/stokaro/unswell/rule"
)

type teamRule struct{}

// Descriptor defines a custom policy without importing Unswell internals.
func (teamRule) Descriptor() rule.Descriptor {
	return rule.Descriptor{
		ID: "team.avoid-magic", Version: "1", Summary: "Describe the mechanism instead of calling it magic.",
		Description: "A team-specific exact token policy.", Limitations: "Fiction and quotations need a different policy.",
		Scope: "sentence", Group: "team-policy", Status: "experimental", Contexts: []string{"paragraph"},
		Requires: []nlp.Capability{nlp.Tokens, nlp.Sentences}, Parameters: []string{}, Examples: []rule.Example{},
		Defaults: rule.Settings{Enabled: true, Severity: "warning", Gate: "forbid", Score: rule.Score{Weight: 10, Cap: 10}},
	}
}

// Evaluate shares the engine's tokenization and returns emitter errors.
func (teamRule) Evaluate(ctx context.Context, view rule.View, emit rule.Emitter) error {
	for _, block := range view.Document.Blocks {
		for _, sentence := range block.Sentences {
			if err := checkSentence(ctx, sentence, emit); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkSentence(ctx context.Context, sentence document.Sentence, emit rule.Emitter) error {
	for _, token := range sentence.Tokens {
		if err := ctx.Err(); err != nil {
			return err
		}
		if token.Normal != "magic" {
			continue
		}
		evidence := rule.Evidence{Kind: "exact", Activation: 1000, Metrics: []rule.Metric{{Name: "matches", Value: 1, Unit: "tokens"}},
			Occurrences: []rule.Occurrence{{BlockID: sentence.BlockID, SentenceID: sentence.ID, Spans: token.Spans}}}
		if err := emit.Emit(evidence); err != nil {
			return err
		}
	}
	return nil
}

func analyze(ctx context.Context) (unswell.RunResult, error) {
	engine, err := unswell.New(unswell.Options{Rules: append(builtin.Rules(), teamRule{})})
	if err != nil {
		return unswell.RunResult{}, err
	}
	return engine.Analyze(ctx, document.Source{Name: "memory.txt", Format: document.Plain, Bytes: []byte("The service uses magic.")})
}

func run() error {
	result, err := analyze(context.Background())
	if err != nil {
		return err
	}
	return report.Write(os.Stdout, "json", result, report.Options{})
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
