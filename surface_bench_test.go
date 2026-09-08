package unswell_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
)

func BenchmarkSurfaceScaling(b *testing.B) {
	for _, implementation := range builtin.Rules() {
		d := implementation.Descriptor()
		if !surfaceRuleID(d.ID) {
			continue
		}
		for _, blocks := range []int{128, 512} {
			b.Run(fmt.Sprintf("%s/blocks=%d", d.ID, blocks), func(b *testing.B) {
				benchmarkSurface(b, d.ID, d.Examples[0].Text, blocks)
			})
		}
	}
}

func benchmarkSurface(b *testing.B, id, example string, blocks int) {
	b.Helper()
	text := strings.Repeat(example+"\n\n# Next\n\n", blocks)
	policy := "version: 1\nextends: [builtin:custom]\nanalysis: {max_candidates: 8000000}\nrules:\n  " + id + ": {enabled: true}\n"
	engine, err := unswell.New(unswell.Options{Config: []byte(policy)})
	if err != nil {
		b.Fatal(err)
	}
	source := document.Source{Name: "scaling.md", Format: document.Markdown, Bytes: []byte(text)}
	if _, err := engine.Analyze(b.Context(), source); err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(source.Bytes)))
	b.ReportAllocs()
	for b.Loop() {
		result, err := engine.Analyze(b.Context(), source)
		if err != nil {
			b.Fatal(err)
		}
		if len(result.Findings) != blocks {
			b.Fatalf("got %d findings; expected one per independent section (%d)", len(result.Findings), blocks)
		}
	}
}
