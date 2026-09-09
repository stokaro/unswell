package evaluation_test

import (
	"fmt"
	"testing"

	"github.com/stokaro/unswell/research/annotation/evaluation"
)

func BenchmarkPairedSourceGroups(b *testing.B) {
	for _, groups := range []int{100, 1000} {
		b.Run(fmt.Sprint(groups), func(b *testing.B) {
			rows := make([]evaluation.PairedObservation, groups*2)
			for i := range rows {
				value, positive := 0.8, true
				observation := evaluation.Observation{UnitID: fmt.Sprint(i), GroupID: fmt.Sprint(i / 2), Label: i % 2,
					Response: &value, Positive: &positive}
				rows[i] = evaluation.PairedObservation{observation, observation}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				if _, err := evaluation.Compare(b.Context(), rows, [2]float64{0.5, 0.5}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
