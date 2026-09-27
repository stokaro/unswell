package extract

import (
	"fmt"
	"slices"
	"sync"

	ts "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
)

var markdownSyntaxLanguage = sync.OnceValues(func() (*ts.Language, error) {
	lang, err := grammars.LoadLanguage("markdown", grammars.BlobByName("markdown"))
	if err != nil {
		return nil, fmt.Errorf("load markdown grammar: %w", err)
	}
	scanner, ok := lang.ExternalScanner.(grammars.MarkdownExternalScanner)
	if !ok {
		return nil, fmt.Errorf("markdown grammar has an unsupported external scanner")
	}
	line, table := -1, -1
	for index, symbol := range lang.ExternalSymbols {
		if int(symbol) >= len(lang.SymbolNames) {
			continue
		}
		switch lang.SymbolNames[symbol] {
		case "_line_ending":
			line = index
		case "_pipe_table_line_ending":
			table = index
		}
	}
	if line < 0 || table < 0 {
		return nil, fmt.Errorf("markdown grammar lacks table boundary tokens")
	}
	lang.ExternalScanner = markdownBoundaryScanner{MarkdownExternalScanner: scanner, line: line, table: table}
	return lang, nil
})

type markdownBoundaryScanner struct {
	grammars.MarkdownExternalScanner
	line  int
	table int
}

// SupportsIncrementalReuse disables reuse because the copied lookahead cursor
// does not extend the runtime's recorded lookahead frontier. Unswell uses full parses.
func (markdownBoundaryScanner) SupportsIncrementalReuse() bool { return false }

// Scan lets the grammar end a table before a line containing a single pipe.
func (s markdownBoundaryScanner) Scan(payload any, lexer *ts.ExternalLexer, valid []bool) bool {
	if s.line < len(valid) && s.table < len(valid) && valid[s.line] && valid[s.table] && markdownLonePipeAhead(lexer) {
		valid = slices.Clone(valid)
		valid[s.table] = false
	}
	return s.MarkdownExternalScanner.Scan(payload, lexer, valid)
}

func markdownLonePipeAhead(lexer *ts.ExternalLexer) bool {
	// ExternalLexer contains a read-only source slice and scalar cursor state.
	// Look ahead on a value copy without moving the scanner's real cursor.
	probe := *lexer
	if !markdownNextLine(&probe) {
		return false
	}
	markdownBoundarySpaces(&probe)
	for probe.Lookahead() == '>' {
		probe.Advance(false)
		markdownBoundarySpaces(&probe)
	}
	if probe.Lookahead() != '|' {
		return false
	}
	probe.Advance(false)
	markdownBoundarySpaces(&probe)
	return probe.Lookahead() == 0 || probe.Lookahead() == '\r' || probe.Lookahead() == '\n'
}

func markdownNextLine(probe *ts.ExternalLexer) bool {
	markdownBoundarySpaces(probe)
	switch probe.Lookahead() {
	case '\r':
		probe.Advance(false)
		if probe.Lookahead() == '\n' {
			probe.Advance(false)
		}
	case '\n':
		probe.Advance(false)
	default:
		return false
	}
	return true
}

func markdownBoundarySpaces(probe *ts.ExternalLexer) {
	for probe.Lookahead() == ' ' || probe.Lookahead() == '\t' {
		probe.Advance(false)
	}
}
