package unswell_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

// The rule's name promises a construction family. It used to recognize one
// stock sentence pair. A claim that defines itself against an alternative makes
// the same move whether the alternative fills a second sentence or a clause, so
// the in-sentence forms count too. One such clause reads as ordinary English.
// The window decides when repeating the shape deserves a report.
func TestPairedContrastCountsTheConstructionFamily(t *testing.T) {
	const id = "syntax.paired-contrast-density"
	for _, row := range []struct {
		name  string
		text  string
		fires bool
	}{
		{"one contrast is ordinary English",
			"The request is refused rather than retried, because the checksum did not match.", false},
		{"two in a passage is the habit",
			"The plan is refused rather than applied. The row is dropped rather than rewritten.", true},
		{"the two forms combine",
			"The plan is refused rather than applied. The name is reported instead of guessed.", true},
		{"the sentence pair still counts",
			"It is not about speed. It is about impact. It is not about tools. It is about outcomes.", true},
		{"a pair and a clause combine",
			"It is not about speed. It is about impact. The row is dropped rather than rewritten.", true},
		{"a protected marker does not count",
			"The plan is refused rather than applied. The row is dropped `rather than` rewritten.", false},
		{"the words apart are not the construction",
			"The plan is refused. The rate is higher than before. Instead, the row is dropped.", false},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			result := editorialResult(t, id, row.text, "")
			if row.fires {
				c.Assert(len(result.Findings) > 0, qt.IsTrue,
					qt.Commentf("expected the repeated contrast to be reported"))
				return
			}
			c.Assert(result.Findings, qt.HasLen, 0)
		})
	}
}
