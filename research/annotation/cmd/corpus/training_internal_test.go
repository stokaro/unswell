package main

// White-box tests: Exercise private training argument validation and inject a failing output writer;
// the executable interface cannot accept the writer used to verify error propagation.

import (
	"bytes"
	"io"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/training"
)

func TestTrainingFlagsRequireExplicitInputs(t *testing.T) {
	for _, args := range [][]string{
		{"train"}, {"train", "--root", "."},
		{"train", "--root", ".", "--round", "round.json", "--feature", "prose-words"},
		{"plan", "--kind", "paragraph"}, {"join", "--allow-simulation"},
		{"train", "--max-operations", "invalid"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			c := qt.New(t)
			var output bytes.Buffer
			err := run(t.Context(), args, badReader{}, &output)
			c.Assert(err, qt.IsNotNil)
			c.Assert(err, qt.Not(qt.ErrorIs), errIO)
			c.Assert(output.Len(), qt.Equals, 0)
		})
	}
}

func TestTrainingOutputErrors(t *testing.T) {
	c := qt.New(t)
	result := training.Artifact{Status: "experimental_numerical_fit"}
	c.Assert(writeResult(t.Context(), "train", result, badWriter{}), qt.ErrorIs, errIO)
	c.Assert(writeResult(t.Context(), "train", result, shortWriter{}), qt.ErrorIs, io.ErrShortWrite)
	var output bytes.Buffer
	c.Assert(writeResult(t.Context(), "train", make(chan int), &output), qt.IsNotNil)
	c.Assert(output.Len(), qt.Equals, 0)
	result.Status = strings.Repeat("x", training.MaxArtifactBytes)
	c.Assert(writeResult(t.Context(), "train", result, &output), qt.ErrorMatches, ".*size limit.*")
	c.Assert(output.Len(), qt.Equals, 0)
}
