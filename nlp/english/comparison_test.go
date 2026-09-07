package english_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/jdkato/prose/v3/segment"
	sentences "gopkg.in/neurosnap/sentences.v1"
	"gopkg.in/neurosnap/sentences.v1/data"
)

func TestSegmentationBackendComparison(t *testing.T) {
	c := qt.New(t)
	model, err := data.Asset("data/english.json")
	c.Assert(err, qt.IsNil)
	training, err := sentences.LoadTraining(model)
	c.Assert(err, qt.IsNil)
	generic := sentences.NewSentenceTokenizer(training)
	selected, err := segment.New()
	c.Assert(err, qt.IsNil)
	cases := []struct {
		name, text string
		want       int
	}{
		{"abbreviation", "Dr. Smith configures the client. The client retries.", 2},
		{"version", "Install version 1.2.3 first. Then restart the client.", 2},
		{"quotation", "The server returned \"try again.\" The client waited.", 2},
		{"Unicode", "The client uses UTF-8. The label is café.", 2},
	}
	for _, row := range cases {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			actual := selected.Segment(row.text)
			c.Assert(actual, qt.HasLen, row.want)
			for _, sentence := range actual {
				c.Assert(row.text[sentence.Start:sentence.End()], qt.Equals, sentence.Text)
			}
			t.Logf("expected=%d prose=%d generic-punkt=%d", row.want, len(actual), len(generic.Tokenize(row.text)))
		})
	}
}

func BenchmarkProseSegmentation(b *testing.B) {
	segmenter, err := segment.New()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		segmenter.Segment("Dr. Smith configures version 1.2.3. The client retries after a connection failure.")
	}
}
