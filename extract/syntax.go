package extract

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	ts "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"

	"github.com/stokaro/unswell/document"
)

type syntaxTree struct {
	tree *ts.Tree
	lang *ts.Language
}

func parseSyntax(ctx context.Context, source []byte, name string) (syntaxTree, error) {
	syntax, err := parseSyntaxTree(ctx, source, name)
	if err == nil && syntax.tree.RootNode().HasErrorOrMissing() {
		syntax.tree.Release()
		return syntaxTree{}, invalidSyntax(name)
	}
	return syntax, err
}

// Parser budgets. The C# grammar spends about two seconds on a 19 KB file on
// an idle machine, so a fixed five seconds failed such files under load. The
// budget grows with the input and stays a bound: a file gets the floor plus
// half a second for every KiB of source.
const (
	parseBudgetFloor  = 5 * time.Second
	parseBudgetPerKiB = 500 * time.Millisecond
)

// parseBudget returns the parser timeout for one source of the given size.
func parseBudget(size int) time.Duration {
	return parseBudgetFloor + time.Duration(size/1024)*parseBudgetPerKiB
}

func invalidSyntax(name string) error {
	return fmt.Errorf("parse %s: %s grammar returned an incomplete or invalid syntax tree", name, name)
}

// parsers keeps idle parsers by grammar name. Building a parser derives tables
// from its language that cost more than parsing a short block, and a Markdown
// document parses every inline block on its own. A fresh parser per block
// dominated the scan of a documentation tree; a parser that finished cleanly
// is kept for the next block of the same grammar instead.
var parsers sync.Map

func acquireParser(name string) (*ts.Parser, error) {
	pool, _ := parsers.LoadOrStore(name, &sync.Pool{})
	if parser, ok := pool.(*sync.Pool).Get().(*ts.Parser); ok {
		return parser, nil
	}
	lang, err := syntaxLanguage(name)
	if err != nil {
		return nil, err
	}
	parser := ts.NewParser(lang)
	parser.SetLogger(nil)
	return parser, nil
}

// releaseParser returns a parser whose parse completed. A parser that timed
// out, was canceled, or failed is dropped: nothing in this package relies on
// its state after such a run, and a fresh one costs less than doubt.
func releaseParser(name string, parser *ts.Parser) {
	parser.SetCancellationFlag(nil)
	if pool, ok := parsers.Load(name); ok {
		pool.(*sync.Pool).Put(parser)
	}
}

// parseSyntaxTree preserves error nodes only for bounded grammar normalization.
// Callers must validate the complete final tree before extracting any prose.
func parseSyntaxTree(ctx context.Context, source []byte, name string) (syntaxTree, error) {
	if err := ctx.Err(); err != nil {
		return syntaxTree{}, err
	}
	parser, err := acquireParser(name)
	if err != nil {
		return syntaxTree{}, err
	}
	// #nosec G115 -- The budget is a positive duration computed from a non-negative size.
	parser.SetTimeoutMicros(uint64(parseBudget(len(source)) / time.Microsecond))
	var canceled uint32
	parser.SetCancellationFlag(&canceled)
	stop := context.AfterFunc(ctx, func() { atomic.StoreUint32(&canceled, 1) })
	// Use each grammar's DFA and attached scanner. The registry's C-family
	// token factory misclassifies C++ raw literals as concatenated strings.
	tree, err := parser.ParseStrict(source)
	stop()
	if ctx.Err() != nil {
		err = ctx.Err()
	}
	if err == nil && (tree == nil || tree.RootNode() == nil) {
		err = fmt.Errorf("%s grammar returned an incomplete or invalid syntax tree", name)
	}
	if err != nil {
		if tree != nil {
			tree.Release()
		}
		return syntaxTree{}, fmt.Errorf("parse %s: %w", name, err)
	}
	lang := parser.Language()
	releaseParser(name, parser)
	return syntaxTree{tree: tree, lang: lang}, nil
}

func syntaxLanguage(name string) (*ts.Language, error) {
	if name == "bash" || name == "sh" || name == "zsh" {
		return bashSyntaxLanguage()
	}
	if name == "markdown" {
		return markdownSyntaxLanguage()
	}
	loaders := map[string]func() *ts.Language{
		"markdown_inline": grammars.MarkdownInlineLanguage,
		"go":              grammars.GoLanguage, "javascript": grammars.JavascriptLanguage,
		"typescript": grammars.TypescriptLanguage, "tsx": grammars.TsxLanguage,
		"python": grammars.PythonLanguage, "rust": grammars.RustLanguage,
		"java": grammars.JavaLanguage, "c": grammars.CLanguage, "cpp": grammars.CppLanguage,
		"csharp": grammars.CSharpLanguage, "yaml": grammars.YamlLanguage,
		"fish": grammars.FishLanguage, "powershell": grammars.PowershellLanguage,
	}
	load := loaders[name]
	if load == nil {
		return nil, fmt.Errorf("unsupported grammar %q", name)
	}
	return load(), nil
}

func walkSyntax(ctx context.Context, node *ts.Node, depth int, visit func(*ts.Node) (bool, error)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if depth > 128 {
		return fmt.Errorf("syntax nesting exceeds 128")
	}
	skip, err := visit(node)
	if err != nil || skip {
		return err
	}
	for i := 0; i < node.ChildCount(); i++ {
		if err := walkSyntax(ctx, node.Child(i), depth+1, visit); err != nil {
			return err
		}
	}
	return nil
}

func syntaxSpan(node *ts.Node, offset int) document.Span {
	return document.Span{Start: offset + int(node.StartByte()), End: offset + int(node.EndByte())}
}
