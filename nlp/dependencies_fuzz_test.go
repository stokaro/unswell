package nlp_test

import (
	"strconv"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/nlp"
)

func FuzzDependencyTree(f *testing.F) {
	f.Add([]byte{1, 2, 255})
	f.Add([]byte{255, 2, 1})
	f.Add([]byte{255})
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		data = data[:min(len(data), 32)]
		heads := make([]int, len(data))
		for i, value := range data {
			heads[i] = int(value)
			if value >= 128 {
				heads[i] -= 256
			}
		}
		mapped, sentence := dependencySentence(strings.TrimSpace(strings.Repeat("a ", len(data))), heads)
		if err := nlp.ValidateDependencyTree(t.Context(), mapped, sentence); err != nil {
			return
		}
		assertDependencyRoot(t, heads)
	})
}

// Independent bounded walks must all terminate at the same root.
func assertDependencyRoot(t *testing.T, heads []int) {
	t.Helper()
	c := qt.New(t)
	c.Assert(len(heads) > 0, qt.IsTrue)
	root := -1
	for start := range heads {
		i := start
		for steps := 0; heads[i] != -1; steps++ {
			c.Assert(steps < len(heads), qt.IsTrue)
			i = heads[i]
			c.Assert(i >= 0 && i < len(heads), qt.IsTrue)
		}
		if root == -1 {
			root = i
		}
		c.Assert(i, qt.Equals, root)
	}
}

func BenchmarkDependencyChain(b *testing.B) {
	for _, count := range []int{1000, 10000} {
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			heads := make([]int, count)
			for i := range heads {
				heads[i] = i - 1
			}
			mapped, sentence := dependencySentence(strings.TrimSpace(strings.Repeat("a ", count)), heads)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				if err := nlp.ValidateDependencyTree(b.Context(), mapped, sentence); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
