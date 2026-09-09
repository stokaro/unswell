package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/training"
)

func TestEvaluationFlagsSeparateLabelsFromPrediction(t *testing.T) {
	for _, args := range [][]string{
		{"predict"}, {"evaluate"}, {"predict", "--round", "round.json"},
		{"evaluate", "--root", "."}, {"evaluate", "--model", "model.json"},
		{"predict", "--root", ".", "--model", "model.json", "--plan", "trial.json"},
		{"evaluate", "--corpus", "corpus.json"}, {"evaluate", "--round", "round.json"},
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

func TestPredictionInputLimitsAndOutputErrors(t *testing.T) {
	c := qt.New(t)
	directory := t.TempDir()
	path := filepath.Join(directory, "model.json")
	c.Assert(os.WriteFile(path, []byte("12345"), 0o600), qt.IsNil)
	_, err := loadLocalArtifact(t.Context(), path, 4, "model")
	c.Assert(err, qt.ErrorMatches, ".*byte limit")
	_, err = loadLocalArtifact(t.Context(), directory, 4, "model")
	c.Assert(err, qt.ErrorMatches, ".*regular file.*")
	var output bytes.Buffer
	args := []string{"evaluate", "--corpus", path, "--round", path}
	c.Assert(run(t.Context(), args, badReader{}, &output), qt.ErrorIs, errIO)
	c.Assert(output.Len(), qt.Equals, 0)
	result := training.Predictions{Status: "experimental_predictions"}
	c.Assert(writeResult(t.Context(), "predict", result, badWriter{}), qt.ErrorIs, errIO)
	c.Assert(writeResult(t.Context(), "evaluate", result, shortWriter{}), qt.ErrorIs, io.ErrShortWrite)
}
