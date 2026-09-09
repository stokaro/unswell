package main

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestComparisonCommandInputs(t *testing.T) {
	valid := []string{"compare", "--plan", "plan.json", "--protocol", "protocol.md", "--corpus", "corpus.json",
		"--comparator", "other.json", "--round", "round.json"}
	for _, args := range [][]string{{"compare"}, append(valid, "--root", "."), append(valid, "--model", "model.json"),
		append(valid, "extra"), valid[:len(valid)-2]} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			c := qt.New(t)
			var output bytes.Buffer
			err := run(t.Context(), args, badReader{}, &output)
			c.Assert(err, qt.IsNotNil)
			c.Assert(err, qt.Not(qt.ErrorIs), errIO)
			c.Assert(output.Len(), qt.Equals, 0)
		})
	}
	c := qt.New(t)
	var output bytes.Buffer
	c.Assert(run(t.Context(), valid, badReader{}, &output), qt.ErrorIs, errIO)
	c.Assert(writeResult(t.Context(), "compare", struct{}{}, badWriter{}), qt.ErrorIs, errIO)
	c.Assert(writeResult(t.Context(), "compare", struct{}{}, shortWriter{}), qt.ErrorIs, io.ErrShortWrite)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	c.Assert(run(ctx, valid, strings.NewReader("{}"), &output), qt.ErrorIs, context.Canceled)
	c.Assert(output.Len(), qt.Equals, 0)
}
