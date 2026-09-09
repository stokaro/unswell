package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestNonUTF8LiteralPreservesNeighboringProse(t *testing.T) {
	for _, row := range []struct {
		format document.Format
		source string
		value  string
	}{
		{document.Go, "package sample\n// Visible comment.\nconst payload = \"\\xff\"\nconst message = \"Visible café text.\"\n", `"\xff"`},
		{document.C, "// Visible comment.\nconst char *payload = \"\\xff\";\nconst char *message = \"Visible café text.\";\n", `"\xff"`},
		{document.CPP, "// Visible comment.\nconst char *payload = \"\\377\";\nconst char *message = \"Visible café text.\";\n", `"\377"`},
		{document.Bash, "# Visible comment.\npayload=$'\\xff'\nmessage='Visible café text.'\n", `$'\xff'`},
	} {
		t.Run(string(row.format), func(t *testing.T) {
			c := qt.New(t)
			source := []byte(row.source)
			doc, err := extract.Parse(t.Context(), document.Source{Name: "sample", Format: row.format, Bytes: source}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(string(source), qt.Equals, row.source)
			c.Assert(doc.Blocks, qt.HasLen, 2)
			c.Assert(strings.TrimSpace(doc.Blocks[0].Text), qt.Equals, "Visible comment.")
			c.Assert(doc.Blocks[1].Text, qt.Equals, "Visible café text.")
			c.Assert(doc.Excluded, qt.HasLen, 1)
			excluded := doc.Excluded[0]
			c.Assert(excluded.Reason, qt.Equals, "non-utf8-literal")
			c.Assert(string(source[excluded.Span.Start:excluded.Span.End]), qt.Equals, row.value)
		})
	}
}

func TestGoByteEscapesAreClassifiedAfterCompleteDecoding(t *testing.T) {
	for _, row := range []struct{ literal, text string }{
		{`"caf\xc3\xa9"`, "café"},
		{`"caf\303\251"`, "café"},
		{`"caf\u00e9"`, "café"},
		{`"\ufffd"`, "\ufffd"},
		{`"Visible \x00 text."`, "Visible  \x00  text."},
	} {
		t.Run(row.literal, func(t *testing.T) {
			c := qt.New(t)
			source := "package sample\nconst text = " + row.literal + "\n"
			doc, err := extract.Parse(t.Context(), document.Source{Name: "sample.go", Format: document.Go, Bytes: []byte(source)},
				extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Blocks, qt.HasLen, 1)
			block := doc.Blocks[0]
			c.Assert(block.Text, qt.Equals, row.text)
			c.Assert(block.Map, qt.HasLen, len(block.Text))
			start := strings.Index(source, row.literal) + 1
			c.Assert(document.Bounds(block.Spans(0, len(block.Text))), qt.DeepEquals,
				document.Span{Start: start, End: start + len(row.literal) - 2})
			c.Assert(doc.Excluded, qt.HasLen, 0)
		})
	}
}

func TestBinaryPayloadSelectionStillValidatesSource(t *testing.T) {
	policy := extract.Policy{Exceptions: []extract.Exception{{
		ID: "wire-data", Paths: []string{"*.go"}, Kinds: []string{"string"}, Symbols: []string{"payload"}, Reason: "Fixed protocol data.",
	}}}
	for _, suffix := range []string{"\xff\n", "const payload = \"\\xZZ\"\n", "const payload = \"unfinished\n"} {
		c := qt.New(t)
		_, err := extract.Parse(t.Context(), document.Source{Name: "sample.go", Format: document.Go,
			Bytes: []byte("package sample\n" + suffix)}, extract.Options{Policy: policy})
		c.Assert(err, qt.IsNotNil)
	}
	c := qt.New(t)
	source := "package sample\nconst payload = \"\\xff\"\nconst message = \"Keep this message.\"\n"
	doc, err := extract.Parse(t.Context(), document.Source{Name: "sample.go", Format: document.Go, Bytes: []byte(source)},
		extract.Options{Policy: policy})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Blocks, qt.HasLen, 1)
	c.Assert(doc.Blocks[0].Text, qt.Equals, "Keep this message.")
	c.Assert(doc.Excluded, qt.HasLen, 1)
	c.Assert(doc.Excluded[0].Reason, qt.Equals, "config:wire-data: Fixed protocol data.")
}

func TestEmbeddedDataRequiresAnExplicitSelectionPolicy(t *testing.T) {
	c := qt.New(t)
	data := []struct{ name, literal string }{
		{"query", "`SELECT name FROM records WHERE enabled = 1;`"},
		{"payload", "`{\"timeout\":30,\"retries\":false}`"},
		{"script", "`#!/bin/sh\nprintf '%s' \"$HOME\"\n`"},
		{"identifier", `"cache_entry_id"`},
		{"protocol", `"HTTP/1.1 101 Switching Protocols"`},
	}
	source := "package sample\n// Check the connection before retrying.\n"
	policy := extract.Policy{}
	for _, row := range data {
		source += "const " + row.name + " = " + row.literal + "\n"
		policy.Exceptions = append(policy.Exceptions, extract.Exception{
			ID: row.name, Paths: []string{"sample.go"}, Formats: []document.Format{document.Go}, Kinds: []string{"string"},
			Symbols: []string{row.name}, Reason: "Fixed " + row.name + " data, not a user-facing message.",
		})
	}
	source += "const message = \"Check the connection before retrying.\"\n"
	input := document.Source{Name: "sample.go", Format: document.Go, Bytes: []byte(source)}
	broad, err := extract.Parse(t.Context(), input, extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(broad.Blocks, qt.HasLen, len(data)+2)
	c.Assert(broad.Excluded, qt.HasLen, 0)
	focused, err := extract.Parse(t.Context(), input, extract.Options{Policy: policy})
	c.Assert(err, qt.IsNil)
	c.Assert(focused.Blocks, qt.HasLen, 2)
	c.Assert(focused.Blocks[0], qt.DeepEquals, broad.Blocks[0])
	c.Assert(focused.Blocks[1].MappedText, qt.DeepEquals, broad.Blocks[len(broad.Blocks)-1].MappedText)
	c.Assert(focused.Excluded, qt.HasLen, len(data))
	for i, row := range data {
		excluded := focused.Excluded[i]
		c.Assert(source[excluded.Span.Start:excluded.Span.End], qt.Equals, row.literal)
		c.Assert(excluded.Reason, qt.Equals, "config:"+row.name+": "+policy.Exceptions[i].Reason)
	}
}
