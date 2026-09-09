package extract

import (
	"context"
	"fmt"
	"sync/atomic"

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

func invalidSyntax(name string) error {
	return fmt.Errorf("parse %s: %s grammar returned an incomplete or invalid syntax tree", name, name)
}

// parseSyntaxTree preserves error nodes only for bounded grammar normalization.
// Callers must validate the complete final tree before extracting any prose.
func parseSyntaxTree(ctx context.Context, source []byte, name string) (syntaxTree, error) {
	if err := ctx.Err(); err != nil {
		return syntaxTree{}, err
	}
	lang, err := syntaxLanguage(name)
	if err != nil {
		return syntaxTree{}, err
	}
	parser := ts.NewParser(lang)
	parser.SetLogger(nil)
	parser.SetTimeoutMicros(5_000_000)
	var canceled uint32
	parser.SetCancellationFlag(&canceled)
	stop := context.AfterFunc(ctx, func() { atomic.StoreUint32(&canceled, 1) })
	defer stop()
	// Use each grammar's DFA and attached scanner. The registry's C-family
	// token factory misclassifies C++ raw literals as concatenated strings.
	tree, err := parser.ParseStrict(source)
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
