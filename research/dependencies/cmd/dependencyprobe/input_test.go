package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestProbeInputValidation(t *testing.T) {
	for _, row := range []struct{ name, input, err string }{
		{"empty", `[]`, ".*1 through 1000.*"},
		{"unknown field", `[{"id":"a","extra":1}]`, ".*unknown field.*"},
		{"trailing value", `[{"id":"a"}] null`, ".*one JSON array"},
		{"duplicate", `[{"id":"a"},{"id":"a"}]`, ".*unique.*"},
		{"missing ID", `[{}]`, ".*nonempty.*"},
		{"too long", `[{"id":"a","text":"` + strings.Repeat("a", 8193) + `"}]`, ".*8192 bytes"},
		{"valid", `[{"id":"a","format":"text","text":"Clients retry."}]`, ""},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			path := filepath.Join(t.TempDir(), "cases.json")
			c.Assert(os.WriteFile(path, []byte(row.input), 0o600), qt.IsNil)
			cases, hash, err := loadCases(path)
			if row.err != "" {
				c.Assert(err, qt.ErrorMatches, row.err)
				return
			}
			c.Assert(err, qt.IsNil)
			c.Assert(cases, qt.HasLen, 1)
			c.Assert(hash, qt.HasLen, 64)
		})
	}
}

func TestModelInventoryAndLimits(t *testing.T) {
	c := qt.New(t)
	path := filepath.Join(t.TempDir(), "model")
	c.Assert(os.Mkdir(path, 0o700), qt.IsNil)
	file := filepath.Join(path, "model.bin")
	c.Assert(os.WriteFile(file, []byte("abc"), 0o600), qt.IsNil)
	files, size, err := inventoryModel(path)
	c.Assert(err, qt.IsNil)
	c.Assert(size, qt.Equals, int64(3))
	c.Assert(files["model.bin"], qt.Equals, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad")
	_, err = readLimited(file, 2)
	c.Assert(err, qt.ErrorMatches, ".*exceeds 2 bytes")
}

func TestProbeRequiresExplicitResources(t *testing.T) {
	for _, args := range [][]string{nil, {"--model", "x"}, {"--input", "x"}, {"--repeat", "0"}, {"--unknown"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			c := qt.New(t)
			var out, stderr bytes.Buffer
			c.Assert(run(context.Background(), args, &out, &stderr), qt.IsNotNil)
			c.Assert(out.Len(), qt.Equals, 0)
		})
	}
}
