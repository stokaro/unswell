package feature_test

import (
	"bytes"
	"compress/zlib"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/feature"
)

func TestCompressionRetainsCountsAtAStreamBoundary(t *testing.T) {
	c := qt.New(t)
	reference, before, after := compressionBoundary(t)
	m, err := feature.NewCompression(t.Context(), reference, feature.CompressionOptions{Level: 9, MaxInputBytes: 1 << 20})
	c.Assert(err, qt.IsNil)
	unit := featureUnits(t, testBlock("a"), false)[0]
	result, err := m.Measure(t.Context(), unit, unitIdentity(unit), testLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(result.Counts.TargetBytes, qt.Equals, 1)
	c.Assert(result.Counts.SeedBytes, qt.Equals, before)
	c.Assert(result.Counts.SeedTargetBytes, qt.Equals, after)
	c.Assert(result.Counts.ControlTargetBytes > result.Counts.ControlBytes, qt.IsTrue)
	for _, value := range result.Values {
		c.Assert(value.Number, qt.IsNil)
		c.Assert(value.Reason, qt.Equals, "compression_boundary")
	}
}

// Find a real framing edge through an independent complete-stream calculation.
// Go 1.25.0 reaches it at 1405 repeats; Go 1.27.1 at 1164. Exact compressed
// bytes are not portable across Go releases, so do not force one byte golden.
func compressionBoundary(t *testing.T) (string, int, int) {
	t.Helper()
	c := qt.New(t)
	var output bytes.Buffer
	writer, err := zlib.NewWriterLevel(&output, 9)
	c.Assert(err, qt.IsNil)
	size := func(text string) int {
		output.Reset()
		writer.Reset(&output)
		_, err := writer.Write([]byte(text))
		c.Assert(err, qt.IsNil)
		c.Assert(writer.Close(), qt.IsNil)
		return output.Len()
	}
	for n := 1; n <= 4096; n++ {
		reference := strings.Repeat("a ", n)
		before, after := size(reference+"\n"), size(reference+"\na")
		if after <= before {
			return reference, before, after
		}
	}
	c.Fatal("review compressor behavior: no nonpositive increment found within the bounded framing fixture")
	return "", 0, 0
}
