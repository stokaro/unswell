package feature_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
)

// These repeated ASCII fixtures measure allocation and processing cost. They
// carry no authorship or quality labels and do not represent a prose corpus.
func BenchmarkCompression(b *testing.B) {
	for _, size := range []int{48, 1024, 65536} {
		unit := compressionBenchmarkUnit(b, size)
		for _, level := range []int{0, 9} {
			for _, prefix := range []int{32, 32768} {
				name := fmt.Sprintf("target=%d/level=%d/prefix=%d", size, level, prefix)
				b.Run(name, func(b *testing.B) {
					benchmarkCompressionCalls(b, unit, level, prefix)
				})
			}
		}
	}
}

func benchmarkCompressionCalls(b *testing.B, unit nlp.PreparedUnit, level, prefix int) {
	b.Helper()
	reference := compressionBenchmarkText(prefix - 1)
	options := feature.CompressionOptions{Level: level, MaxInputBytes: 1 << 20}
	identity := unitIdentity(unit)
	limits := feature.Limits{MaxTokens: 20000, MaxUniqueWords: 20000, MaxBytes: 65536, MaxBlocks: 1}
	for _, mode := range []string{"reuse-reference", "new-reference"} {
		b.Run(mode, func(b *testing.B) {
			compressor := newBenchmarkCompression(b, reference, options)
			b.ReportAllocs()
			b.SetBytes(int64(len(unit.Block().Text)))
			for b.Loop() {
				if mode == "new-reference" {
					compressor = newBenchmarkCompression(b, reference, options)
				}
				result, err := compressor.Measure(b.Context(), unit, identity, limits)
				if err != nil || result.Hash == "" || len(result.Values) != 2 {
					b.Fatalf("incomplete compression measurement: %v", err)
				}
			}
		})
	}
}

func newBenchmarkCompression(b *testing.B, reference string, options feature.CompressionOptions) *feature.Compression {
	b.Helper()
	compressor, err := feature.NewCompression(b.Context(), reference, options)
	if err != nil {
		b.Fatal(err)
	}
	return compressor
}

func compressionBenchmarkUnit(b *testing.B, size int) nlp.PreparedUnit {
	b.Helper()
	doc, err := extract.Parse(b.Context(), document.Source{Name: "compression.txt", Format: document.Plain,
		Bytes: []byte(compressionBenchmarkText(size))}, extract.Options{})
	if err != nil || len(doc.Blocks) != 1 {
		b.Fatalf("prepare one source block: %v", err)
	}
	provider, err := english.New()
	if err != nil {
		b.Fatal(err)
	}
	units, err := nlp.PrepareUnits(b.Context(), doc.Blocks[0], provider, nlp.UnitOptions{
		Kinds: []string{"paragraph"}, Capabilities: []nlp.Capability{nlp.Tokens, nlp.Sentences},
		Limits: nlp.UnitLimits{MaxBytes: 65536, MaxContextBytes: 65536, MaxUnits: 1, MaxTokens: 20000, MaxSegments: 1024},
	})
	if err != nil || len(units) != 1 || len(units[0].Block().Text) != size {
		b.Fatalf("prepare one target with %d bytes: %v", size, err)
	}
	return units[0]
}

func compressionBenchmarkText(size int) string {
	const sentence = "The cache retries after a connection failure. "
	return strings.Repeat(sentence, size/len(sentence)+1)[:size-1] + "."
}
