package unswell_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func BenchmarkRepetitionScaling(b *testing.B) {
	for _, id := range []string{"repetition.ngram-density", "repetition.syntax-template", "repetition.paragraph-overlap"} {
		for _, blocks := range []int{128, 512, 2048} {
			b.Run(fmt.Sprintf("%s/blocks=%d", id, blocks), func(b *testing.B) {
				benchmarkRepetition(b, id, blocks)
			})
		}
	}
}

func benchmarkRepetition(b *testing.B, id string, blocks int) {
	b.Helper()
	var text strings.Builder
	for i := range blocks {
		fmt.Fprintf(&text, "The client records value %d during the planned operation and confirms the local result.\n\n", i)
	}
	config := "version: 1\nextends: [builtin:custom]\nanalysis: {max_candidates: 8000000}\nrules:\n  " + id + ": {enabled: true}\n"
	engine, err := unswell.New(unswell.Options{Config: []byte(config)})
	if err != nil {
		b.Fatal(err)
	}
	source := document.Source{Name: "scaling.txt", Format: document.Plain, Bytes: []byte(text.String())}
	b.SetBytes(int64(len(source.Bytes)))
	b.ReportAllocs()
	for b.Loop() {
		result, err := engine.Analyze(b.Context(), source)
		if err != nil {
			b.Fatal(err)
		}
		if len(result.Findings) != 0 {
			b.Fatal("different numeric facts must not form repetition clusters")
		}
	}
}
