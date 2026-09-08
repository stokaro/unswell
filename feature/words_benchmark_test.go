package feature_test

import (
	"strconv"
	"testing"

	"github.com/stokaro/unswell/feature"
)

func BenchmarkComparePreparedWords(b *testing.B) {
	leftWords, rightWords := make([]string, 512), make([]string, 512)
	for i := range leftWords {
		leftWords[i] = "word" + strconv.Itoa(i)
		rightWords[i] = "word" + strconv.Itoa(i+256)
	}
	limits := feature.WordLimits{MaxWords: 512, MaxUniqueWords: 512, MaxBytes: 10000}
	left, err := feature.NewWordSet(b.Context(), leftWords, limits)
	if err != nil {
		b.Fatal(err)
	}
	right, err := feature.NewWordSet(b.Context(), rightWords, limits)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		value, compareErr := feature.CompareWords(b.Context(), left, right)
		if compareErr != nil || value.Intersection() != 256 || value.Union() != 768 {
			b.Fatalf("unexpected comparison: %+v (%v)", value, compareErr)
		}
	}
}
