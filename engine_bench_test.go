package unswell_test

import (
	"testing"

	"github.com/stokaro/unswell"
)

// BenchmarkEngineConstruction measures policy setup without analyzing prose.
func BenchmarkEngineConstruction(b *testing.B) {
	for _, profile := range []string{"default", "strict"} {
		b.Run(profile, func(b *testing.B) {
			var policy []byte
			if profile == "strict" {
				policy = []byte("version: 1\nextends: [builtin:strict-v1]\n")
			}
			b.ReportAllocs()
			for b.Loop() {
				if _, err := unswell.New(unswell.Options{Config: policy}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
