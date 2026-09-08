package annotation

import (
	"math"
	"testing"

	qt "github.com/frankban/quicktest"
)

// The numeric matrices are examples A-C in Krippendorff's computational note,
// pp. 2-5, linked in README.md. They are not editorial corpus annotations.
func TestPublishedNominalExamples(t *testing.T) {
	cases := []struct {
		name      string
		observers []string
		alpha     float64
		paired    int
	}{
		{"binary", []string{"0100000010", "1110010000"}, 1 - 152.0/168, 20},
		{"five labels", []string{"aabbdcccedda", "babbbccceddd"}, 1 - 138.0/448, 24},
		{"missing answers", []string{"123321412...", "1233224125.3", ".3332342251.", "12332441251."}, 1 - 312.0/1216, 40},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			result := nominal(transpose(tc.observers))
			c.Assert(result.Alpha.Value, qt.IsNotNil)
			c.Assert(math.Abs(*result.Alpha.Value-tc.alpha) < 1e-12, qt.IsTrue)
			c.Assert(result.PairedRatings, qt.Equals, tc.paired)
		})
	}
}

func TestNominalDegenerateAndNegativeValues(t *testing.T) {
	cases := []struct {
		name, reason string
		rows         [][]string
	}{
		{"empty", "no_paired_ratings", nil},
		{"singleton", "no_paired_ratings", [][]string{{"acceptable"}}},
		{"constant", "no_expected_disagreement", [][]string{{"acceptable", "acceptable"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			result := nominal(tc.rows)
			c.Assert(result.Alpha.Value, qt.IsNil)
			c.Assert(result.Alpha.Reason, qt.Equals, tc.reason)
		})
	}
	t.Run("negative is retained", func(t *testing.T) {
		c := qt.New(t)
		result := nominal([][]string{{"a", "b"}, {"a", "b"}})
		c.Assert(*result.Alpha.Value, qt.Equals, -0.5)
	})
	t.Run("uncertain is a label", func(t *testing.T) {
		c := qt.New(t)
		result := nominal([][]string{{"acceptable", "uncertain"}, {"needs_revision", "needs_revision"}})
		c.Assert(*result.Uncertain.Value, qt.Equals, 0.25)
		c.Assert(*result.RawAgreement.Value, qt.Equals, 0.5)
		c.Assert(result.PairedRatings, qt.Equals, 4)
	})
}

func transpose(observers []string) [][]string {
	rows := make([][]string, len(observers[0]))
	for _, observer := range observers {
		for i, label := range observer {
			if label != '.' {
				rows[i] = append(rows[i], string(label))
			}
		}
	}
	return rows
}
