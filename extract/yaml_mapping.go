package extract

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/mapping"
)

func mapYAMLScalar(source []byte, span document.Span, kind, value string) (document.MappedText, error) {
	span = yamlContentSpan(source, span, kind)
	var raw mapping.Builder
	segment := string(source[span.Start:span.End])
	for pos := 0; pos < len(segment); {
		text, size, err := yamlRune(segment[pos:], kind)
		if err != nil {
			return document.MappedText{}, err
		}
		raw.Add(text, document.Span{Start: span.Start + pos, End: span.Start + pos + size})
		pos += size
	}
	return mapYAMLWhitespace(raw.Build(), value)
}

func yamlContentSpan(source []byte, span document.Span, kind string) document.Span {
	switch kind {
	case "double_quote_scalar", "single_quote_scalar":
		span.Start++
		span.End--
	case "block_scalar":
		newline := bytes.IndexByte(source[span.Start:span.End], '\n')
		if newline < 0 {
			span.Start = span.End
			return span
		}
		span.Start += newline + 1
		// The syntax node ends at the last content character. Chomping also
		// consumes following line breaks, which must retain their own origins.
		for end := span.End; end < len(source) && strings.ContainsRune(" \t\r\n", rune(source[end])); end++ {
			if source[end] == '\n' || source[end] == '\r' {
				span.End = end + 1
			}
		}
	}
	return span
}

func yamlRune(source, kind string) (string, int, error) {
	if kind == "single_quote_scalar" && strings.HasPrefix(source, "''") {
		return "'", 2, nil
	}
	if kind == "double_quote_scalar" && source[0] == '\\' {
		return yamlEscape(source)
	}
	_, size := utf8.DecodeRuneInString(source)
	return source[:size], size, nil
}

func yamlEscape(source string) (string, int, error) {
	if len(source) < 2 {
		return "", 0, fmt.Errorf("unfinished YAML escape")
	}
	escapes := map[byte]string{
		'0': "\x00", 'a': "\a", 'b': "\b", 't': "\t", '\t': "\t", 'n': "\n", 'v': "\v", 'f': "\f",
		'r': "\r", 'e': "\x1b", ' ': " ", '"': "\"", '/': "/", '\\': "\\", 'N': "\u0085", '_': "\u00a0", 'L': "\u2028", 'P': "\u2029",
	}
	if text, ok := escapes[source[1]]; ok {
		return text, 2, nil
	}
	if source[1] == '\r' || source[1] == '\n' {
		return "", continuationSize(source, "yaml"), nil
	}
	if source[1] == 'x' {
		return numericEscape(source, "yaml")
	}
	if source[1] == 'u' || source[1] == 'U' {
		return unicodeEscape(source)
	}
	return "", 0, fmt.Errorf("unsupported YAML escape")
}

// Align only whitespace transformations. Non-whitespace content must match
// exactly; disagreement between the grammar and decoder is an operational error.
func mapYAMLWhitespace(raw document.MappedText, value string) (document.MappedText, error) {
	var builder mapping.Builder
	pos := 0
	for target := 0; target < len(value); {
		end := whitespaceEnd(raw.Text, pos)
		wanted := whitespaceEnd(value, target)
		if wanted > target {
			if end == pos {
				return document.MappedText{}, fmt.Errorf("decoded YAML whitespace has no source")
			}
			builder.Add(value[target:wanted], document.Bounds(raw.Spans(pos, end)))
		}
		pos, target = end, wanted
		if target == len(value) {
			break
		}
		char, size := utf8.DecodeRuneInString(value[target:])
		if !strings.HasPrefix(raw.Text[pos:], value[target:target+size]) {
			return document.MappedText{}, fmt.Errorf("decoded YAML content differs from the syntax span")
		}
		text := value[target : target+size]
		if unicode.IsControl(char) {
			text = " \x00 "
		}
		builder.Add(text, document.Bounds(raw.Spans(pos, pos+size)))
		pos += size
		target += size
	}
	if whitespaceEnd(raw.Text, pos) != len(raw.Text) {
		return document.MappedText{}, fmt.Errorf("YAML syntax span contains unconsumed content")
	}
	return builder.Build(), nil
}

func whitespaceEnd(text string, start int) int {
	for start < len(text) {
		char, size := utf8.DecodeRuneInString(text[start:])
		if !unicode.IsSpace(char) {
			break
		}
		start += size
	}
	return start
}
