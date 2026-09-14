package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

// A worked example sits in "<pre><code>" markup. The tag grammar does not
// cover it, because the lines belong to the description and come before any
// tag. The lines hold Java, not English. So the region is held back, and the
// prose on either side of it stays apart.
func TestJavadocCodeMarkupIsExcluded(t *testing.T) {
	c := qt.New(t)
	source := "/**\n" +
		" * Keeps the elements whose value does not match any given value.\n" +
		" * <p>\n" +
		" * As often, an example helps:\n" +
		" * <pre><code class='java'> Employee yoda = new Employee(1L, new Name(\"Yoda\"), 800);\n" +
		" * Employee obiwan = new Employee(2L, new Name(\"Obi\"), 800);</code></pre>\n" +
		" * The filter reads the property once per element.\n" +
		" */\n" +
		"class Filters { }\n"
	doc := parseJava(t, source)
	c.Assert(doc.Blocks, qt.HasLen, 1)
	c.Assert(fields(doc.Blocks[0].Text), qt.Equals,
		"Keeps the elements whose value does not match any given value. <p> As often, an example helps: "+
			"\x00 \x00 The filter reads the property once per element.")
	assertSourceMap(c, doc, source)
	c.Assert(excluded(doc, source), qt.DeepEquals, []string{
		"comment-code: <pre><code class='java'> Employee yoda = new Employee(1L, new Name(\"Yoda\"), 800);",
		"comment-code: Employee obiwan = new Employee(2L, new Name(\"Obi\"), 800);</code></pre>",
	})
}

// An inline "{@code}" tag holds a fragment of code on a prose line. The tag
// nests braces, so the region ends at the brace that balances its opener and
// the rest of the sentence is read as prose.
func TestJavadocInlineCodeTagIsExcluded(t *testing.T) {
	c := qt.New(t)
	source := "/**\n" +
		" * Returns {@code new Name(\"Yoda\", new int[]{1, 2})} when the queue is empty.\n" +
		" * A {@literal <T>} parameter names the element type.\n" +
		" */\n" +
		"class Names { }\n"
	doc := parseJava(t, source)
	c.Assert(doc.Blocks, qt.HasLen, 1)
	c.Assert(fields(doc.Blocks[0].Text), qt.Equals,
		"Returns \x00 when the queue is empty. A \x00 parameter names the element type.")
	assertSourceMap(c, doc, source)
	c.Assert(excluded(doc, source), qt.DeepEquals, []string{
		"comment-code: {@code new Name(\"Yoda\", new int[]{1, 2})}",
		"comment-code: {@literal <T>}",
	})
}

// A C# documentation comment is a run of line comments, so a "<code>" region
// opened on one line closes on another. The "<c>" element is its inline form.
func TestCSharpDocCodeMarkupIsExcluded(t *testing.T) {
	c := qt.New(t)
	source := "/// <summary>\n" +
		"/// Reads the settings file and returns the parsed document.\n" +
		"/// The <c>path</c> argument names the file to read.\n" +
		"/// <example>\n" +
		"/// <code>\n" +
		"/// var reader = new SettingsReader(path);\n" +
		"/// var parsed = reader.Read();\n" +
		"/// </code>\n" +
		"/// </example>\n" +
		"/// </summary>\n" +
		"class SettingsReader { }\n"
	doc, err := extract.Parse(t.Context(),
		document.Source{Name: "sample.cs", Format: document.CSharp, Bytes: []byte(source)}, extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Blocks, qt.HasLen, 1)
	c.Assert(fields(doc.Blocks[0].Text), qt.Equals,
		"<summary> Reads the settings file and returns the parsed document. The \x00 argument names the file to read. "+
			"<example> \x00 \x00 \x00 \x00 </example> </summary>")
	assertSourceMap(c, doc, source)
	c.Assert(excluded(doc, source), qt.DeepEquals, []string{
		"comment-code: <c>path</c>",
		"comment-code: <code>",
		"comment-code: var reader = new SettingsReader(path);",
		"comment-code: var parsed = reader.Read();",
		"comment-code: </code>",
	})
}

// A name that merely starts with a region name is not a region.
func TestCommentMarkupNeedsAWholeElementName(t *testing.T) {
	c := qt.New(t)
	source := "/**\n" +
		" * The <precondition> element and the {@codepoint} reference stay prose.\n" +
		" */\n" +
		"class Names { }\n"
	doc := parseJava(t, source)
	c.Assert(doc.Blocks, qt.HasLen, 1)
	c.Assert(fields(doc.Blocks[0].Text), qt.Equals,
		"The <precondition> element and the {@codepoint} reference stay prose.")
	c.Assert(doc.Excluded, qt.HasLen, 0)
}

func parseJava(t *testing.T, source string) document.Document {
	t.Helper()
	doc, err := extract.Parse(t.Context(),
		document.Source{Name: "Sample.java", Format: document.Java, Bytes: []byte(source)}, extract.Options{})
	qt.New(t).Assert(err, qt.IsNil)
	return doc
}

func fields(text string) string { return strings.Join(strings.Fields(text), " ") }

func excluded(doc document.Document, source string) []string {
	var lines []string
	for _, item := range doc.Excluded {
		lines = append(lines, item.Reason+": "+strings.TrimSpace(source[item.Span.Start:item.Span.End]))
	}
	return lines
}
