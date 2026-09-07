// Package extract provides source-preserving prose extraction from text and syntax trees.
package extract

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/mapping"
)

// Options controls extraction. Zero values select conservative prose contexts.
type Options struct {
	IncludeQuotes bool
	MaxBytes      int
	MaxBlocks     int
	Policy        Policy
}

// Parse extracts prose without changing or retaining the caller's input buffer.
func Parse(ctx context.Context, src document.Source, options Options) (document.Document, error) {
	doc := document.Document{Name: src.Name, Format: src.Format, Blocks: []document.Block{}, Excluded: []document.Exclusion{}}
	if err := ctx.Err(); err != nil {
		return doc, err
	}
	options = defaultOptions(options)
	if err := ValidatePolicy(options.Policy); err != nil {
		return doc, err
	}
	if err := validateSource(src, options); err != nil {
		return doc, err
	}
	doc.Source = append([]byte(nil), src.Bytes...)
	doc.Hash = fmt.Sprintf("%x", sha256.Sum256(src.Bytes))
	var err error
	switch src.Format {
	case document.Plain:
		plain(&doc, 0, len(src.Bytes), "paragraph")
	case document.Markdown:
		err = markdown(ctx, &doc, options)
	case document.Go, document.JavaScript, document.TypeScript, document.TSX, document.Python,
		document.Rust, document.Java, document.C, document.CPP, document.Bash, document.Shell,
		document.Zsh, document.Fish, document.PowerShell, document.CSharp, document.YAML:
		err = sourceProse(ctx, &doc, options)
	default:
		err = fmt.Errorf("unsupported input format %q", src.Format)
	}
	if err != nil {
		return doc, err
	}
	filterContexts(&doc, options.Policy)
	if len(doc.Blocks) > options.MaxBlocks {
		return doc, fmt.Errorf("source exceeds %d prose blocks", options.MaxBlocks)
	}
	return doc, ctx.Err()
}

func defaultOptions(options Options) Options {
	if options.MaxBytes <= 0 {
		options.MaxBytes = 2 << 20
	}
	if options.MaxBlocks <= 0 {
		options.MaxBlocks = 10000
	}
	return options
}

func validateSource(src document.Source, options Options) error {
	if src.Name == "" {
		return fmt.Errorf("source name is required")
	}
	if len(src.Bytes) > options.MaxBytes {
		return fmt.Errorf("source exceeds %d bytes", options.MaxBytes)
	}
	if !utf8.Valid(src.Bytes) || strings.ContainsRune(string(src.Bytes), 0) {
		return fmt.Errorf("source must be valid UTF-8 without NUL")
	}
	return nil
}

var blankLine = regexp.MustCompile(`\r?\n[\t ]*\r?\n`)
var url = regexp.MustCompile(`(?:https?://|mailto:|www\.)[^\s<>]+`)

func plain(doc *document.Document, start, end int, kind string) {
	pos := start
	for _, gap := range blankLine.FindAllIndex(doc.Source[start:end], -1) {
		plainBlock(doc, pos, start+gap[0], kind)
		pos = start + gap[1]
	}
	plainBlock(doc, pos, end, kind)
}

func plainBlock(doc *document.Document, start, end int, kind string) {
	var builder mapping.Builder
	builder.Source(doc.Source, start, end, false)
	appendBlock(doc, builder.Build(), kind)
}

func appendBlock(doc *document.Document, mapped document.MappedText, kind string) {
	if strings.TrimSpace(strings.ReplaceAll(mapped.Text, "\x00", "")) == "" {
		return
	}
	protectURLs(doc, &mapped)
	doc.Blocks = append(doc.Blocks, document.Block{
		ID: len(doc.Blocks), Kind: kind, Span: document.Bounds(mapped.Spans(0, len(mapped.Text))),
		MappedText: mapped, Sentences: []document.Sentence{},
	})
}

func protectURLs(doc *document.Document, mapped *document.MappedText) {
	matches := url.FindAllStringIndex(mapped.Text, -1)
	if len(matches) == 0 {
		return
	}
	var builder mapping.Builder
	pos := 0
	for _, match := range matches {
		for match[1] > match[0] && strings.ContainsRune(".,;:!?)", rune(mapped.Text[match[1]-1])) {
			match[1]--
		}
		for i := pos; i < match[0]; i++ {
			builder.Add(mapped.Text[i:i+1], mapped.Map[i])
		}
		span := document.Bounds(mapped.Spans(match[0], match[1]))
		builder.Add(" \x00 ", span)
		doc.Excluded = append(doc.Excluded, document.Exclusion{Span: span, Reason: "url"})
		pos = match[1]
	}
	for i := pos; i < len(mapped.Text); i++ {
		builder.Add(mapped.Text[i:i+1], mapped.Map[i])
	}
	*mapped = builder.Build()
}
