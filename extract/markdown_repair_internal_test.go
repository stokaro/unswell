package extract

// White-box tests: Check three guards of the HTML rewrite that no public input
// reaches with the pinned grammar. The grammar ends a comment, a declaration,
// and a "pre" element itself, so the rewrite never sees one. Its only error in
// an HTML block follows a blank line. And the rewrite always gives an HTML
// block, so the tree check never refuses one.

import (
	"bytes"
	"testing"

	qt "github.com/frankban/quicktest"
	ts "github.com/odvcencio/gotreesitter"

	"github.com/stokaro/unswell/document"
)

func TestHTMLRewriteNeedsABlankLineBeforeTheDefect(t *testing.T) {
	c := qt.New(t)
	source := []byte("x\n\n- Item.\n\n  <div>x</div>\n\n  More of it.\n")
	loose, err := parseSyntaxTree(t.Context(), bytes.Clone(source), "markdown")
	c.Assert(err, qt.IsNil)
	defer loose.tree.Release()
	var defect *ts.Node
	err = walkSyntax(t.Context(), loose.tree.RootNode(), 0, func(node *ts.Node) (bool, error) {
		if defect == nil && node.IsError() {
			defect = node
		}
		return false, nil
	})
	c.Assert(err, qt.IsNil)
	c.Assert(defect, qt.IsNotNil)

	// The same tree over bytes that have no blank line before the defect.
	joined := bytes.Clone(source)
	joined[bytes.Index(joined, []byte("</div>\n\n"))+len("</div>\n")] = ' '
	_, changed := maskOpenHTML(loose, defect, joined)
	c.Assert(changed, qt.IsFalse)
	c.Assert(string(joined), qt.Contains, "<div>x</div>")

	run, changed := maskOpenHTML(loose, defect, source)
	c.Assert(changed, qt.IsTrue)
	c.Assert(string(source[run.Start:run.End]), qt.Equals, "<?xxxxxxxx?>")
}

func TestOnlyHTMLThatABlankLineEndsIsRewritten(t *testing.T) {
	for open, want := range map[string]bool{
		"<div>":        true,
		"<p>":          true,
		"</pre>":       true,
		"<preview>":    true,
		"<pre-set x>":  true,
		"<pre>":        false,
		"<PRE class>":  false,
		"<pre":         false,
		"<script>":     false,
		"<style\n":     false,
		"<textarea>":   false,
		"<!-- note":    false,
		"<!DOCTYPE x>": false,
		"<?php":        false,
		"<":            false,
		"text":         false,
	} {
		qt.New(t).Assert(blankLineEndsHTML([]byte(open)), qt.Equals, want, qt.Commentf("%q", open))
	}
}

func TestRecoveryRefusesARewrittenRunOutsideHTML(t *testing.T) {
	c := qt.New(t)
	source := []byte("x\n\n- Item.\n\n  <?xxx?>\n\n  More of it.\n")
	span := func(text string) document.Span {
		start := bytes.Index(source, []byte(text))
		c.Assert(start >= 0, qt.IsTrue)
		return document.Span{Start: start, End: start + len(text)}
	}
	html, prose := span("<?xxx?>"), span("More of it.")

	accepted := markdownRecovery{source: source, html: []document.Span{html}}
	syntax, err := accepted.parse(t.Context())
	c.Assert(err, qt.IsNil)
	syntax.tree.Release()

	refused := markdownRecovery{source: source, html: []document.Span{html, prose}}
	_, err = refused.parse(t.Context())
	c.Assert(err, qt.ErrorIs, errInvalidSyntax)

	// A rewritten run that was blanked later is not expected in the tree.
	blanked := markdownRecovery{source: source, html: []document.Span{html, prose}, blanked: []document.Span{prose}}
	syntax, err = blanked.parse(t.Context())
	c.Assert(err, qt.IsNil)
	syntax.tree.Release()
}
