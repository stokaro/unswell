package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

// directiveFrames hold the text before and after the comment under test in
// the syntax of each format: an ordinary comment and a string, so a test can
// check that both neighbors stay checked. Formats without a frame use the
// C-style default.
var directiveFrames = map[document.Format][2]string{
	document.Go:     {"package sample\n// Keep the input.\n", "\nconst message = \"Keep the message.\"\n"},
	document.Python: {"# Keep the input.\n", "\nmessage = \"Keep the message.\"\n"},
	document.Bash:   {"# Keep the input.\n", "\nmessage='Keep the message.'\n"},
	document.Shell:  {"# Keep the input.\n", "\nmessage='Keep the message.'\n"},
	document.YAML:   {"# Keep the input.\n", "\nmessage: Keep the message.\n"},
	document.C:      {"// Keep the input.\n", "\nconst char *message = \"Keep the message.\";\n"},
	document.Java:   {"// Keep the input.\n", "\nclass Sample { String message = \"Keep the message.\"; }\n"},
	document.CSharp: {"// Keep the input.\n", "\nclass Sample { string message = \"Keep the message.\"; }\n"},
}

func directiveSource(format document.Format, comment string) string {
	frame, ok := directiveFrames[format]
	if !ok {
		frame = [2]string{"// Keep the input.\n", "\nconst message = \"Keep the message.\";\n"}
	}
	return frame[0] + comment + frame[1]
}

// A comment that starts with a listed tool prefix is a directive for that
// tool, not prose. It is excluded with its full comment span, and the prose
// around it is still checked.
func TestToolDirectiveCommentsAreExcluded(t *testing.T) {
	for _, row := range []struct {
		format  document.Format
		comment string
	}{
		{document.Python, "# mypy: allow-untyped-defs"}, {document.Python, "# pylint: disable=missing-docstring"},
		{document.Python, "# noqa"}, {document.Python, "# type: ignore[arg-type]"}, {document.Python, "# pragma: no cover"},
		{document.Python, "# fmt: off"}, {document.Python, "# ruff: noqa"}, {document.Python, "# nosec B101"},
		{document.Python, "# -*- coding: utf-8 -*-"}, {document.Python, "# isort: skip_file"},
		{document.Go, "//nolint:errcheck"}, {document.Go, "//go:generate go run ./gen"}, {document.Go, "//lint:ignore SA1019 kept"},
		{document.Go, "// +build linux"}, {document.Go, "// #nosec G204 -- Runs the fixed binary."},
		{document.Go, "// SPDX-License-Identifier: MIT"},
		{document.JavaScript, "// eslint-disable-next-line no-console"}, {document.JavaScript, "/* eslint-disable */"},
		{document.JavaScript, "// @ts-ignore"}, {document.JavaScript, "// prettier-ignore"},
		{document.JavaScript, "/* istanbul ignore next */"}, {document.JavaScript, "/* c8 ignore next */"},
		{document.JavaScript, "// biome-ignore lint/suspicious/noExplicitAny: kept"},
		{document.TypeScript, "// @ts-expect-error"}, {document.TSX, "// @ts-nocheck"},
		{document.C, "// NOLINT"}, {document.C, "// NOLINTNEXTLINE(readability)"}, {document.C, "// clang-format off"},
		{document.Java, "// CHECKSTYLE:OFF"}, {document.Java, "// noinspection unchecked"}, {document.Java, "// NOSONAR"},
		{document.CSharp, "// ReSharper disable All"},
		{document.Bash, "# shellcheck disable=SC2034"}, {document.Shell, "# shfmt: -i 2"},
		{document.YAML, "# yamllint disable-line rule:line-length"}, {document.YAML, "# yaml-language-server: $schema=schema.json"},
	} {
		t.Run(string(row.format)+" "+row.comment, func(t *testing.T) {
			c := qt.New(t)
			source := directiveSource(row.format, row.comment)
			doc, err := extract.Parse(t.Context(),
				document.Source{Name: "sample", Format: row.format, Bytes: []byte(source)}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(blockTexts(doc), qt.DeepEquals, []string{"Keep the input.", "Keep the message."})
			directives := excludedTexts(doc, source, "directive")
			c.Assert(directives, qt.DeepEquals, []string{row.comment})
			// YAML also records the mapping key of the message as excluded data.
			c.Assert(len(doc.Excluded)-len(directives), qt.Equals, len(excludedTexts(doc, source, "yaml-key")))
		})
	}
}

func excludedTexts(doc document.Document, source, reason string) []string {
	var texts []string
	for _, excluded := range doc.Excluded {
		if excluded.Reason == reason {
			texts = append(texts, source[excluded.Span.Start:excluded.Span.End])
		}
	}
	return texts
}

// A sentence that mentions a tool, a prefix in another case, or a prefix that
// continues into a longer word is prose.
func TestToolMentionsStayProse(t *testing.T) {
	for _, row := range []struct {
		format  document.Format
		comment string
	}{
		{document.JavaScript, "// Our eslint setup rejects console calls in production code."},
		{document.JavaScript, "// The prettier-ignore marker is documented below."},
		{document.Python, "# The pylint checks run in the pipeline."},
		{document.Python, "# Pragma: the client retries on failure."},
		{document.Python, "# Type: ignore the cache and read the file."},
		{document.C, "// NOLINTER is the name of a tool."},
		{document.Bash, "# shellcheck-style tools exist for other languages."},
		{document.Go, "// nolinters are not a thing."},
	} {
		t.Run(string(row.format)+" "+row.comment, func(t *testing.T) {
			c := qt.New(t)
			source := directiveSource(row.format, row.comment)
			doc, err := extract.Parse(t.Context(),
				document.Source{Name: "sample", Format: row.format, Bytes: []byte(source)}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Excluded, qt.HasLen, 0)
			c.Assert(strings.Join(blockTexts(doc), "\n"), qt.Contains, strings.TrimLeft(row.comment, "/# "))
		})
	}
}

// A directive between two prose comment lines ends the first block and does
// not join the lines around it.
func TestDirectiveLinesDoNotJoinNeighboringProse(t *testing.T) {
	for _, row := range []struct {
		format document.Format
		source string
	}{
		{document.Python, "# The first paragraph ends here.\n# noqa: E501\n# The second paragraph starts here.\n"},
		{document.JavaScript, "// The first paragraph ends here.\n// @ts-ignore\n// The second paragraph starts here.\n"},
		{document.Go, "package sample\n// The first paragraph ends here.\n//nolint:lll\n// The second paragraph starts here.\n"},
	} {
		t.Run(string(row.format), func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(),
				document.Source{Name: "sample", Format: row.format, Bytes: []byte(row.source)}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(blockTexts(doc), qt.DeepEquals, []string{"The first paragraph ends here.", "The second paragraph starts here."})
			c.Assert(doc.Excluded, qt.HasLen, 1)
			c.Assert(doc.Excluded[0].Reason, qt.Equals, "directive")
		})
	}
}
