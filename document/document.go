// Package document defines source coordinates and the neutral prose model.
package document

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Format identifies an explicitly supported input syntax.
type Format string

// Supported formats never execute the source.
const (
	Plain      Format = "text"
	Markdown   Format = "markdown"
	Go         Format = "go"
	JavaScript Format = "javascript"
	TypeScript Format = "typescript"
	TSX        Format = "tsx"
	Python     Format = "python"
	Rust       Format = "rust"
	Java       Format = "java"
	C          Format = "c"
	CPP        Format = "cpp"
	CSharp     Format = "csharp"
	YAML       Format = "yaml"
	Bash       Format = "bash"
	Shell      Format = "sh"
	Zsh        Format = "zsh"
	Fish       Format = "fish"
	PowerShell Format = "powershell"
)

// Formats returns the supported syntax names in a stable order.
func Formats() []Format {
	return []Format{Plain, Markdown, Go, JavaScript, TypeScript, TSX, Python, Rust, Java, C, CPP,
		CSharp, YAML, Bash, Shell, Zsh, Fish, PowerShell}
}

// Source contains UTF-8 bytes owned by the caller. Do not mutate them during analysis.
type Source struct {
	Name   string `json:"name"`
	Format Format `json:"format"`
	Bytes  []byte `json:"-"`
}

// Span is a half-open range of original UTF-8 bytes.
type Span struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// Valid reports whether a nonempty span falls within the source.
func (s Span) Valid(size int) bool { return s.Start >= 0 && s.End > s.Start && s.End <= size }

// Position uses one-based lines and Unicode code point columns; tabs count once.
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Locate converts a byte offset into a position, rejecting split UTF-8 sequences.
func Locate(src []byte, offset int) (Position, error) {
	if offset < 0 || offset > len(src) || !utf8.Valid(src[:offset]) {
		return Position{}, fmt.Errorf("invalid UTF-8 byte offset %d", offset)
	}
	before := string(src[:offset])
	lineStart := strings.LastIndexByte(before, '\n') + 1
	return Position{Line: strings.Count(before, "\n") + 1, Column: utf8.RuneCountInString(before[lineStart:]) + 1}, nil
}

// MappedText maps each analysis byte back to its original rune or encoded entity.
// A NUL in Text marks a protected boundary and cannot participate in a phrase.
type MappedText struct {
	Text string `json:"text"`
	Map  []Span `json:"-"`
}

// Spans returns the actual source segments for an analysis range.
func (m MappedText) Spans(start, end int) []Span {
	if start < 0 || end <= start || end > len(m.Map) {
		return nil
	}
	result := make([]Span, 0)
	for _, span := range m.Map[start:end] {
		if len(result) > 0 && span.Start <= result[len(result)-1].End {
			result[len(result)-1].End = max(result[len(result)-1].End, span.End)
			continue
		}
		result = append(result, span)
	}
	return result
}

// Bounds covers a set of ordered evidence segments.
func Bounds(spans []Span) Span {
	if len(spans) == 0 {
		return Span{}
	}
	return Span{Start: spans[0].Start, End: spans[len(spans)-1].End}
}

// Block is one structural prose unit. Context contains grammar-derived scope
// labels, not line numbers; an empty context means no named owner was available.
// Slices are immutable during rule evaluation.
type Block struct {
	ID      int      `json:"id"`
	Kind    string   `json:"kind"`
	Span    Span     `json:"span"`
	Context []string `json:"context,omitempty"`
	MappedText
	Sentences []Sentence `json:"sentences"`
	Words     int        `json:"words"`
}

// Token carries original evidence and a Penn Treebank POS tag.
type Token struct {
	Text      string `json:"text"`
	Normal    string `json:"normal"`
	Tag       string `json:"tag"`
	Start     int    `json:"start"`
	End       int    `json:"end"`
	Spans     []Span `json:"spans"`
	Word      bool   `json:"word"`
	Protected bool   `json:"protected"`
}

// Sentence is segmented within a structural block, never across excluded regions.
type Sentence struct {
	ID      int     `json:"id"`
	BlockID int     `json:"block_id"`
	Text    string  `json:"text"`
	Span    Span    `json:"span"`
	Spans   []Span  `json:"spans"`
	Tokens  []Token `json:"tokens"`
	Chunks  []Chunk `json:"chunks"`
	Words   int     `json:"words"`
}

// Chunk is a shallow NP, VP, or PP candidate, not a dependency parse.
type Chunk struct {
	Kind       string `json:"kind"`
	FirstToken int    `json:"first_token"`
	EndToken   int    `json:"end_token"`
}

// Exclusion records an intentionally unexamined region.
type Exclusion struct {
	Span   Span   `json:"span"`
	Reason string `json:"reason"`
}

// Directive is policy-sensitive text extracted from an actual source comment.
// Text omits comment delimiters; Span covers the original comment bytes.
type Directive struct {
	Text string `json:"text"`
	Span Span   `json:"span"`
}

// Document holds extracted prose and explicit analysis coverage.
type Document struct {
	Name       string      `json:"name"`
	Format     Format      `json:"format"`
	Hash       string      `json:"sha256"`
	Blocks     []Block     `json:"blocks"`
	Excluded   []Exclusion `json:"excluded"`
	Directives []Directive `json:"directives,omitempty"`
	Words      int         `json:"words"`
	Source     []byte      `json:"-"`
}

// Normalize folds case and typographic apostrophes without changing source data.
func Normalize(text string) string {
	return strings.ToLower(strings.NewReplacer("’", "'", "‘", "'").Replace(text))
}

// IsWord recognizes prose tokens containing letters or digits.
func IsWord(text string) bool {
	return strings.ContainsFunc(text, func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) })
}
