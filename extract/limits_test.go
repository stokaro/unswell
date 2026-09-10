package extract_test

import (
	"context"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

// Extraction runs on untrusted input, so its guards belong to the contract:
// an unsupported format, a missing name, a broken encoding, or a source past a
// declared limit stops before any prose is produced.
func TestExtractionRefusesUnusableInput(t *testing.T) {
	for _, row := range []struct {
		name    string
		source  document.Source
		options extract.Options
		want    string
	}{
		{"unsupported format", document.Source{Name: "sample.pl", Format: "perl", Bytes: []byte("Visible text.\n")},
			extract.Options{}, `unsupported input format "perl"`},
		{"missing name", document.Source{Format: document.Plain, Bytes: []byte("Visible text.\n")},
			extract.Options{}, "source name is required"},
		{"invalid UTF-8", document.Source{Name: "sample.txt", Format: document.Plain, Bytes: []byte{0xff, 0xfe}},
			extract.Options{}, "source must be valid UTF-8 without NUL"},
		{"NUL byte", document.Source{Name: "sample.txt", Format: document.Plain, Bytes: []byte("Visible\x00text.")},
			extract.Options{}, "source must be valid UTF-8 without NUL"},
		{"source above the byte limit",
			document.Source{Name: "sample.txt", Format: document.Plain, Bytes: []byte(strings.Repeat("word ", 100))},
			extract.Options{MaxBytes: 64}, "source exceeds 64 bytes"},
		{"more blocks than allowed",
			document.Source{Name: "sample.txt", Format: document.Plain,
				Bytes: []byte(strings.Repeat("Visible paragraph text.\n\n", 8))},
			extract.Options{MaxBlocks: 3}, "source exceeds 3 prose blocks"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(), row.source, row.options)
			c.Assert(err, qt.ErrorMatches, row.want)
			c.Assert(doc.Name, qt.Equals, row.source.Name)
		})
	}
}

// A canceled scan must stop extraction rather than finish the document and
// report work the caller no longer wants.
func TestExtractionStopsOnCancellation(t *testing.T) {
	for _, row := range []struct {
		name   string
		format document.Format
		source string
	}{
		{"plain text", document.Plain, "Visible paragraph text.\n"},
		{"markdown", document.Markdown, "# Heading\n\nVisible paragraph text.\n"},
		{"source code", document.Go, "package a\n// Visible comment text.\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			_, err := extract.Parse(ctx,
				document.Source{Name: "sample", Format: row.format, Bytes: []byte(row.source)}, extract.Options{})
			c.Assert(err, qt.ErrorIs, context.Canceled)
		})
	}
}
