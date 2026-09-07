// Package mapping builds source-preserving normalized prose.
package mapping

import (
	"html"
	"strings"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
)

// Builder accumulates text with an origin for every analysis byte.
type Builder struct {
	text    strings.Builder
	origins []document.Span
}

// Add appends transformed text sharing one original source span.
func (b *Builder) Add(text string, span document.Span) {
	b.text.WriteString(text)
	for range len(text) {
		b.origins = append(b.origins, span)
	}
}

// Source copies a source range, optionally decoding Markdown escapes and entities.
func (b *Builder) Source(src []byte, start, end int, decode bool) {
	for i := start; i < end; {
		text, size := next(src[i:end], decode)
		b.Add(text, document.Span{Start: i, End: i + size})
		i += size
	}
}

func next(src []byte, decode bool) (string, int) {
	if decode && escapedPunctuation(src) {
		return string(src[1:2]), 2
	}
	if decode && src[0] == '&' {
		end := strings.IndexByte(string(src[:min(len(src), 34)]), ';')
		if end > 0 {
			encoded := string(src[:end+1])
			decoded := html.UnescapeString(encoded)
			if decoded != encoded {
				return decoded, end + 1
			}
		}
	}
	r, size := utf8.DecodeRune(src)
	if r == '\r' || r == '\ufeff' {
		return " ", size
	}
	return string(src[:size]), size
}

func escapedPunctuation(src []byte) bool {
	return len(src) > 1 && src[0] == '\\' && strings.ContainsRune("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", rune(src[1]))
}

// Build returns an independent immutable mapping.
func (b *Builder) Build() document.MappedText {
	return document.MappedText{Text: b.text.String(), Map: b.origins}
}
