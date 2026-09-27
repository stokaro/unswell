package extract

// Empty-value recognition follows tree-sitter-bash at a06c2e4415e9bc0346c6b86d401879ffb44058f7.
// Copyright (c) 2017 Max Brunsfeld. Licensed under the MIT License; see
// licenses/tree-sitter-bash_LICENSE. The surrounding adapter is Unswell code.

import (
	"fmt"
	"sync"

	ts "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
)

var bashSyntaxLanguage = sync.OnceValues(func() (*ts.Language, error) {
	// Decode an owned grammar instance so other users of gotreesitter's cached
	// BashLanguage retain their scanner and parser configuration.
	lang, err := grammars.LoadLanguage("bash", grammars.BlobByName("bash"))
	if err != nil {
		return nil, fmt.Errorf("load bash grammar: %w", err)
	}
	empty, emptySymbol, err := bashExternalSymbol(lang, "_empty_value")
	if err != nil {
		return nil, err
	}
	concat, _, err := bashExternalSymbol(lang, "_concat")
	if err != nil {
		return nil, err
	}
	if lang.ExternalScanner == nil {
		return nil, fmt.Errorf("bash grammar has no external scanner")
	}
	lang.ExternalScanner = bashEmptyValueScanner{ExternalScanner: lang.ExternalScanner, empty: empty, concat: concat, symbol: emptySymbol}
	return lang, nil
})

func bashExternalSymbol(lang *ts.Language, name string) (int, ts.Symbol, error) {
	for index, symbol := range lang.ExternalSymbols {
		if int(symbol) < len(lang.SymbolNames) && lang.SymbolNames[symbol] == name {
			return index, symbol, nil
		}
	}
	return 0, 0, fmt.Errorf("bash grammar lacks external token %q", name)
}

type bashEmptyValueScanner struct {
	ts.ExternalScanner
	empty  int
	concat int
	symbol ts.Symbol
}

// Scan preserves empty assignment boundaries before delegating other tokens.
func (s bashEmptyValueScanner) Scan(payload any, lexer *ts.ExternalLexer, valid []bool) bool {
	// The v0.52.0 scanner's opening-parenthesis probe can consume spaces before
	// EMPTY_VALUE runs. Preserve the original grammar's zero-width token at this
	// boundary. A valid CONCAT retains the scanner's existing precedence.
	if s.empty < len(valid) && s.concat < len(valid) && valid[s.empty] && !valid[s.concat] {
		switch lexer.Lookahead() {
		case ' ', '\t', '\n', '\r', '\v', '\f':
			lexer.SetResultSymbol(s.symbol)
			return true
		}
	}
	return s.ExternalScanner.Scan(payload, lexer, valid)
}
