package extract

// White-box tests: the parser pool is an implementation detail of
// parseSyntaxTree, and the property under test is that a parser kept from an
// earlier parse produces the same tree as a fresh one.

import (
	"context"
	"strings"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"
	ts "github.com/stokaro/gotreesitter"
)

func sExpression(ctx context.Context, c *qt.C, source []byte, name string) string {
	c.Helper()
	syntax, err := parseSyntaxTree(ctx, source, name)
	c.Assert(err, qt.IsNil)
	defer syntax.tree.Release()
	var out strings.Builder
	err = walkSyntax(ctx, syntax.tree.RootNode(), 0, func(node *ts.Node) (bool, error) {
		out.WriteString(node.Type(syntax.lang))
		out.WriteByte('[')
		out.WriteString(string(source[node.StartByte():node.EndByte()]))
		out.WriteString("] ")
		return false, nil
	})
	c.Assert(err, qt.IsNil)
	return out.String()
}

func TestParserPoolReusesParsersWithoutChangingTrees(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	sources := map[string][]byte{
		"markdown_inline": []byte("A sentence with `code`, *emphasis*, and a [link](https://example.test)."),
		"go":              []byte("package p\n\n// Comment one.\nfunc F() { _ = \"a string\" }\n"),
		"yaml":            []byte("key: A value with words.\nlist:\n  - one\n  - two\n"),
	}
	first := map[string]string{}
	for name, source := range sources {
		first[name] = sExpression(ctx, c, source, name)
		c.Assert(first[name], qt.Not(qt.Equals), "")
	}
	// The second pass takes the parsers the first pass released.
	for name, source := range sources {
		c.Assert(sExpression(ctx, c, source, name), qt.Equals, first[name])
	}
	// Concurrent parses of one grammar share a pool and never a parser.
	var wg sync.WaitGroup
	results := make([]string, 16)
	for i := range results {
		wg.Go(func() { results[i] = sExpression(ctx, c, sources["markdown_inline"], "markdown_inline") })
	}
	wg.Wait()
	for _, result := range results {
		c.Assert(result, qt.Equals, first["markdown_inline"])
	}
}

func TestParserPoolDropsCanceledParsers(t *testing.T) {
	c := qt.New(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := parseSyntaxTree(ctx, []byte("package p\n"), "go")
	c.Assert(err, qt.ErrorIs, context.Canceled)
	// A parse after cancellation still succeeds on a fresh parser.
	syntax, err := parseSyntaxTree(context.Background(), []byte("package p\n"), "go")
	c.Assert(err, qt.IsNil)
	syntax.tree.Release()
}
