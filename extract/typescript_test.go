package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestTypeOnlyWildcardExportsPreserveProse(t *testing.T) {
	for _, format := range []document.Format{document.TypeScript, document.TSX} {
		for _, declaration := range []string{
			`export type * from "./types.ts";`,
			`export type * as names from "./types.ts";`,
			`export type * as type from "./types.ts";`,
			"export type\n* from \"./types.ts\"",
			`export /* Before keyword. */ type /* After keyword. */ * from "./types.ts";`,
			"export // Before keyword.\ntype // After keyword.\n* as names from \"./types.ts\";",
			`export * from "./types.ts";`,
			`export type { Thing } from "./types.ts";`,
		} {
			t.Run(string(format)+"/"+declaration, func(t *testing.T) {
				c := qt.New(t)
				source := "\ufeff// Before export.\n" + declaration + "\n// After export.\n" +
					`export const message = "Visible caf\u00e9 text.";` + "\n"
				source = strings.ReplaceAll(source, "\n", "\r\n")
				input := []byte(source)
				doc, err := extract.Parse(t.Context(), document.Source{Name: "sample", Format: format, Bytes: input},
					extract.Options{IncludeStructure: true})
				c.Assert(err, qt.IsNil)
				c.Assert(string(input), qt.Equals, source)
				c.Assert(string(doc.Source), qt.Equals, source)
				texts := []string{}
				for _, block := range doc.Blocks {
					texts = append(texts, strings.TrimSpace(block.Text))
				}
				want := []string{"Before export."}
				if strings.Contains(source, "Before keyword.") {
					want = append(want, "Before keyword.", "After keyword.")
				}
				want = append(want, "After export.", "Visible café text.")
				c.Assert(texts, qt.DeepEquals, want)
				assertSourceMap(c, doc, source)
				for _, block := range doc.Blocks {
					text := strings.TrimSpace(block.Text)
					original := strings.ReplaceAll(text, "é", `\u00e9`)
					start := strings.Index(source, original)
					textStart := strings.Index(block.Text, text)
					c.Assert(document.Bounds(block.Spans(textStart, textStart+len(text))), qt.DeepEquals,
						document.Span{Start: start, End: start + len(original)})
				}
				c.Assert(doc.Excluded, qt.HasLen, 1)
				c.Assert(doc.Excluded[0].Reason, qt.Equals, "import-path")
				span := doc.Excluded[0].Span
				c.Assert(source[span.Start:span.End], qt.Equals, `"./types.ts"`)
			})
		}
	}
}

func TestTypeOnlyExportsDoNotHideInvalidSyntax(t *testing.T) {
	for _, format := range []document.Format{document.TypeScript, document.TSX} {
		for _, source := range []string{
			`export type type * from "./types.ts";`,
			`export default type * from "./types.ts";`,
			`export type nonsense * from "./types.ts";`,
			`export type ** from "./types.ts";`,
			`export type * as from "./types.ts";`,
			`export type *;`,
			`export type * from;`,
			`export type * from "./types.ts;`,
			`export type * from "./types.ts"; const = 3;`,
			`const = 3; export type * from "./types.ts";`,
			`export type * from "./types.ts"; /* unfinished`,
		} {
			t.Run(string(format)+"/"+source, func(t *testing.T) {
				c := qt.New(t)
				for _, contexts := range [][]string{nil, {}} {
					doc, err := extract.Parse(t.Context(), document.Source{Name: "sample", Format: format, Bytes: []byte(source)},
						extract.Options{Policy: extract.Policy{Contexts: contexts}})
					c.Assert(err, qt.ErrorMatches, `parse `+string(format)+`: .*incomplete or invalid syntax tree`)
					c.Assert(doc.Blocks, qt.HasLen, 0)
				}
			})
		}
	}
}

func TestExportModulePathsExcludeOnlyTheSourceField(t *testing.T) {
	for _, format := range []document.Format{document.JavaScript, document.TypeScript, document.TSX} {
		t.Run(string(format), func(t *testing.T) {
			c := qt.New(t)
			source := `export * from "Certainly!";
export * as names from "It is important to note that.";
export { Thing } from "Let's dive in.";
export const message = "Visible export text.";
export default "Visible default text.";
export function describe() { return "Visible function text."; }
`
			doc, err := extract.Parse(t.Context(), document.Source{Name: "sample", Format: format, Bytes: []byte(source)},
				extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Blocks, qt.HasLen, 3)
			for i, text := range []string{"Visible export text.", "Visible default text.", "Visible function text."} {
				c.Assert(doc.Blocks[i].Text, qt.Equals, text)
			}
			c.Assert(doc.Excluded, qt.HasLen, 3)
			for _, excluded := range doc.Excluded {
				c.Assert(excluded.Reason, qt.Equals, "import-path")
			}
		})
	}
}
