package extract

import (
	"bytes"
	"fmt"
	"strings"

	ts "github.com/stokaro/gotreesitter"
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
	m.remaining -= len(input)
	if m.remaining < 0 {
		return false, fmt.Errorf("MDX syntax exceeds the parse-input budget (16 times source bytes plus 64 KiB)")
	}
	syntax, err := parseSyntaxTree(m.ctx, []byte(input), "javascript")
	if err != nil {
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
