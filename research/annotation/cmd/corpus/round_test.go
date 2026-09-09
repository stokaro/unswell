package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation"
)

func TestJoinRoundFileBoundaries(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	directory, err := os.OpenRoot(root)
	c.Assert(err, qt.IsNil)
	t.Cleanup(func() { c.Assert(directory.Close(), qt.IsNil) })
	_, err = loadRound(t.Context(), root)
	c.Assert(err, qt.ErrorMatches, ".*regular file.*")
	_, err = loadRound(t.Context(), filepath.Join(root, "absent"))
	c.Assert(err, qt.IsNotNil)
	path := filepath.Join(root, "round.json")
	c.Assert(os.WriteFile(path, []byte(`{"version":"a","version":"b"}`), 0o600), qt.IsNil)
	_, err = loadRound(t.Context(), path)
	c.Assert(err, qt.IsNotNil)
	c.Assert(os.Truncate(path, annotation.MaxBytes+1), qt.IsNil)
	_, err = loadRound(t.Context(), path)
	c.Assert(err, qt.ErrorMatches, ".*byte limit.*")
	data, err := os.ReadFile("../../testdata/tutorial.json")
	c.Assert(err, qt.IsNil)
	c.Assert(directory.WriteFile("round.json", data, 0o600), qt.IsNil)
	round, err := loadRound(t.Context(), path)
	c.Assert(err, qt.IsNil)
	c.Assert(round, qt.IsNotNil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = loadRound(ctx, path)
	c.Assert(err, qt.ErrorIs, context.Canceled)
}
