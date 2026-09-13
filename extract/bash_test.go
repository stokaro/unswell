package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	ts "github.com/stokaro/gotreesitter"
	"github.com/stokaro/gotreesitter/grammars"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestBashEmptyAssignmentsPreserveProse(t *testing.T) {
	for _, format := range []document.Format{document.Bash, document.Shell, document.Zsh} {
		t.Run(string(format), func(t *testing.T) {
			t.Parallel()
			for _, command := range []string{
				"! A=x B= command", "A=x B= command", "! A= command", "! A=x B=\tcommand",
				"! A=x B= C= command", "A= B= C= command", "! A=x B= command >output",
				"! A=x B= command | next", "! A=x B= command && next", "! A=x B= \\\ncommand",
				"if ! GOBIN=\"$bindir\" GOFLAGS= $command; then :; fi",
				"case \"$x\" in a) A= command;; esac", "A=(one two)",
			} {
				t.Run(command, func(t *testing.T) {
					assertBashProse(t, format, command)
				})
			}
		})
	}
}

func assertBashProse(t *testing.T, format document.Format, command string) {
	t.Helper()
	c := qt.New(t)
	const message = "Keep café coordinates."
	source := []byte("# Check the input.\r\n" + command + "\r\nmessage=\"" + message + "\"\r\n")
	original := string(source)
	doc, err := extract.Parse(t.Context(), document.Source{Name: "sample.sh", Format: format, Bytes: source},
		extract.Options{IncludeStructure: true})
	c.Assert(err, qt.IsNil)
	c.Assert(string(source), qt.Equals, original)
	c.Assert(strings.TrimSpace(doc.Blocks[0].Text), qt.Equals, "Check the input.")
	var found bool
	for _, block := range doc.Blocks {
		if block.Text != message {
			continue
		}
		found = true
		start := strings.Index(original, message)
		c.Assert(document.Bounds(block.Spans(0, len(block.Text))), qt.DeepEquals, document.Span{Start: start, End: start + len(message)})
		c.Assert(block.Context, qt.Contains, "variable_assignment:name:message")
	}
	c.Assert(found, qt.IsTrue)
	for _, excluded := range doc.Excluded {
		c.Assert(excluded.Span.Valid(len(source)), qt.IsTrue)
	}
}

func TestBashEmptyAssignmentPreservesSubstitutionBoundaries(t *testing.T) {
	c := qt.New(t)
	source := "! A=x B= command \"It is ${important} to $(printf 'note') that.\"\n"
	doc, err := extract.Parse(t.Context(), document.Source{Name: "sample.bash", Format: document.Bash, Bytes: []byte(source)},
		extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Blocks, qt.HasLen, 2)
	c.Assert(doc.Blocks[0].Text, qt.Equals, "It is  \x00  to  \x00  that.")
	c.Assert(doc.Blocks[1].Text, qt.Equals, "note")
	var protected []string
	for _, excluded := range doc.Excluded {
		if excluded.Reason == "interpolation" {
			protected = append(protected, source[excluded.Span.Start:excluded.Span.End])
		}
	}
	c.Assert(protected, qt.DeepEquals, []string{"${important}", "$(printf 'note')"})
}

func TestBashEmptyAssignmentDoesNotHideSyntaxErrors(t *testing.T) {
	for _, source := range []string{
		"# A visible comment.\nif ! A=x B= command; then\n",
		"# A visible comment.\n! A=x B=\"unfinished\n",
		"# A visible comment.\n! A=x B= command\n\xff",
	} {
		c := qt.New(t)
		_, err := extract.Parse(t.Context(), document.Source{Name: "sample.bash", Format: document.Bash, Bytes: []byte(source)},
			extract.Options{})
		c.Assert(err, qt.IsNotNil)
	}
}

func TestBashGrammarAdapterDoesNotChangeSharedGrammar(t *testing.T) {
	c := qt.New(t)
	source := []byte("! A=x B= command\n# Keep the input.\n")
	lang := grammars.BashLanguage()
	before, err := ts.NewParser(lang).ParseStrict(source)
	c.Assert(err, qt.IsNil)
	defer before.Release()
	c.Assert(before.RootNode().HasErrorOrMissing(), qt.IsTrue)
	doc, err := extract.Parse(t.Context(), document.Source{Name: "sample.bash", Format: document.Bash, Bytes: source}, extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Blocks, qt.HasLen, 1)
	c.Assert(strings.TrimSpace(doc.Blocks[0].Text), qt.Equals, "Keep the input.")
	after, err := ts.NewParser(lang).ParseStrict(source)
	c.Assert(err, qt.IsNil)
	defer after.Release()
	c.Assert(after.RootNode().HasErrorOrMissing(), qt.IsTrue)
	c.Assert(after.RootNode().SExpr(lang), qt.Equals, before.RootNode().SExpr(lang))
}
