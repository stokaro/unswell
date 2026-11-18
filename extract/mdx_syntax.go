package extract

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"

	ts "github.com/odvcencio/gotreesitter"
)

// syntaxEnd tries only delimiter boundaries, with both per-region and total
// parse-input bounds. A damaged construct fails extraction instead of exposing
// its tail as prose or scanning an unbounded number of prefixes.
func (m *mdxMasker) syntaxEnd(start int, delimiter byte, kind string) (int, error) {
	pos := start
	for attempts := 0; attempts < 128 && pos < len(m.source); {
		next := bytes.IndexByte(m.source[pos:], delimiter)
		end := len(m.source)
		if next >= 0 {
			end = pos + next + 1
		}
		pos = end
		if kind == "esm" && end < len(m.source) && !mdxBlankLine(m.source[end:]) {
			continue
		}
		attempts++
		ok, err := m.candidate(m.source[start:end], kind)
		if err != nil {
			return 0, err
		}
		if ok {
			return end, nil
		}
	}
	return 0, fmt.Errorf("invalid or unterminated MDX %s (at most 128 candidate boundaries)", kind)
}

func mdxBlankLine(source []byte) bool {
	line, _, _ := bytes.Cut(source, []byte{'\n'})
	return len(bytes.TrimSpace(line)) == 0
}

func (m *mdxMasker) candidate(source []byte, kind string) (bool, error) {
	input := string(source)
	if kind == "expression" {
		input = "<_>" + input + "</_>"
	}
	if kind == "tag" {
		input = mdxTagProgram(input)
	}
	ok, err := m.accepts(source, input, kind)
	if ok || err != nil || kind != "tag" {
		return ok, err
	}
	// HTML names attributes "class" and "for", and JSX allows them. The
	// JavaScript grammar reads such a name as a keyword and rejects the tag, so
	// a tag that failed is tried once more with those names respelled.
	respelled := mdxKeywordAttributes(input)
	if respelled == input {
		return false, nil
	}
	return m.accepts(source, respelled, kind)
}

// accepts parses one candidate program and reports whether it is the complete
// construct of its kind. The source is the candidate as the page wrote it.
func (m *mdxMasker) accepts(source []byte, input, kind string) (bool, error) {
	m.remaining -= len(input)
	if m.remaining < 0 {
		return false, fmt.Errorf("MDX syntax exceeds the parse-input budget (16 times source bytes plus 64 KiB)")
	}
	syntax, err := parseSyntaxTree(m.ctx, []byte(input), "javascript")
	if err != nil {
		var stopped *ts.ParseStoppedEarlyError
		if errors.As(err, &stopped) && stopped.Reason == ts.ParseStopAcceptedPrefix {
			// This candidate closes an inner JavaScript construct, not the MDX
			// region. Try the next delimiter without admitting a partial tree.
			return false, nil
		}
		return false, err
	}
	defer syntax.tree.Release()
	root := syntax.tree.RootNode()
	if root.HasErrorOrMissing() {
		return false, nil
	}
	if kind == "esm" {
		return mdxESMProgram(root, syntax.lang), nil
	}
	if kind == "tag" {
		return m.tag(source, root, syntax.lang)
	}
	return true, nil
}

// javascriptWords are the words that JavaScript reserves or reads in a special
// way. Each is a legal name of a JSX attribute.
var javascriptWords = []string{"async", "await", "break", "case", "catch", "class", "const", "continue", "debugger",
	"default", "delete", "do", "else", "enum", "export", "extends", "false", "finally", "for", "function", "get", "if",
	"import", "in", "instanceof", "let", "new", "null", "of", "return", "set", "static", "super", "switch", "this",
	"throw", "true", "try", "typeof", "var", "void", "while", "with", "yield"}

// mdxKeywordAttributes respells each attribute name of a tag that is a word of
// JavaScript: its first letter becomes a capital, so the length of the tag
// stays the same. Only a name counts: a word after white space, outside
// quotes and braces, and before "=", white space, "/", or ">". A word inside
// an attribute value is left as it is, so a damaged expression stays damaged.
func mdxKeywordAttributes(tag string) string {
	out := []byte(tag)
	var values tagValues
	for pos := 1; pos < len(out); pos++ {
		outside, escaped := values.read(out[pos])
		if escaped {
			pos++
		} else if outside {
			pos = respellAttributeName(out, pos)
		}
	}
	return string(out)
}

// tagValues follows the attribute values of a tag: its strings and the braces
// of its expressions.
type tagValues struct {
	depth int
	quote byte
}

// read takes the next byte of a tag. It reports whether the byte stands
// outside every value, and whether it escapes the byte after it.
func (v *tagValues) read(char byte) (outside, escapes bool) {
	if v.quote != 0 {
		return false, v.quoted(char)
	}
	switch char {
	case '"', '\'':
		v.quote = char
	case '`':
		if v.depth == 0 {
			return true, false
		}
		v.quote = char
	case '{':
		v.depth++
	case '}':
		v.depth--
	default:
		return v.depth == 0, false
	}
	return false, false
}

// quoted takes a byte inside a string and reports whether it escapes the byte
// after it. A backslash does that in a string of an expression, and stands for
// itself in the string of a plain attribute.
func (v *tagValues) quoted(char byte) bool {
	if char == '\\' && v.depth > 0 {
		return true
	}
	if char == v.quote {
		v.quote = 0
	}
	return false
}

// respellAttributeName capitalizes the lowercase word at pos when it is an
// attribute name and a word of JavaScript, and returns the place of its last
// byte, or pos when no word starts there.
func respellAttributeName(tag []byte, pos int) int {
	const space = " \t\r\n"
	lower := func(char byte) bool { return char >= 'a' && char <= 'z' }
	if !lower(tag[pos]) || strings.IndexByte(space, tag[pos-1]) < 0 {
		return pos
	}
	end := pos
	for end < len(tag) && lower(tag[end]) {
		end++
	}
	if end < len(tag) && strings.IndexByte(space+"=/>", tag[end]) >= 0 && slices.Contains(javascriptWords, string(tag[pos:end])) {
		tag[pos] -= 'a' - 'A'
	}
	return end - 1
}

func mdxESMProgram(root *ts.Node, lang *ts.Language) bool {
	statements := 0
	for i := 0; i < root.NamedChildCount(); i++ {
		switch root.NamedChild(i).Type(lang) {
		case "import_statement", "export_statement":
			statements++
		case "comment":
		default:
			return false
		}
	}
	return statements > 0
}

func mdxTagProgram(tag string) string {
	if tag == "<>" || tag == "</>" {
		return "<></>"
	}
	if strings.HasPrefix(tag, "</") {
		return "<" + tag[2:] + tag
	}
	if strings.HasSuffix(tag, "/>") {
		return tag
	}
	return strings.TrimSuffix(tag, ">") + "/>"
}

func (m *mdxMasker) tag(source []byte, root *ts.Node, lang *ts.Language) (bool, error) {
	if root.NamedChildCount() != 1 || root.NamedChild(0).Type(lang) != "expression_statement" {
		return false, nil
	}
	expression := root.NamedChild(0)
	if expression.NamedChildCount() != 1 {
		return false, nil
	}
	node := expression.NamedChild(0)
	if node.Type(lang) != "jsx_element" && node.Type(lang) != "jsx_self_closing_element" {
		return false, nil
	}
	return m.balanceTag(source)
}

func (m *mdxMasker) balanceTag(source []byte) (bool, error) {
	// The grammar has validated the name and attributes. The original opening
	// name identifies the matching closing tag without retaining parser nodes.
	tag := strings.TrimPrefix(string(source), "<")
	closing := strings.HasPrefix(tag, "/")
	tag = strings.TrimPrefix(tag, "/")
	end := strings.IndexAny(tag, " \t\r\n/>")
	if end < 0 {
		return false, nil
	}
	name := tag[:end]
	if closing {
		if len(m.stack) == 0 || m.stack[len(m.stack)-1] != name {
			return false, fmt.Errorf("mismatched JSX closing tag %q", name)
		}
		m.stack = m.stack[:len(m.stack)-1]
	} else if !bytes.HasSuffix(source, []byte("/>")) {
		if len(m.stack) >= 128 {
			return false, fmt.Errorf("JSX nesting exceeds 128")
		}
		m.stack = append(m.stack, name)
	}
	return true, nil
}
