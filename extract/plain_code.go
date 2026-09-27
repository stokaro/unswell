package extract

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/stokaro/unswell/document"
)

const (
	plainCodeBytes      = 16 << 10
	plainCodeLines      = 128
	plainCodeCandidates = 100
)

type plainLine struct {
	text       string
	start, end int
	indent     int
}

// Plain-text examples need both a prose introducer and a complete C parse.
// Unknown, oversized, and malformed regions remain prose. Indentation alone
// never removes content, and an excluded example splits the surrounding prose.
func plainText(ctx context.Context, doc *document.Document) error {
	lines := plainLines(doc.Source)
	start, previous, candidates := 0, -1, 0
	for i := 0; i < len(lines); {
		if err := ctx.Err(); err != nil {
			return err
		}
		line := lines[i]
		if line.text == "" {
			i++
			continue
		}
		if !plainIntroduced(lines, previous, i) {
			previous = i
			i++
			continue
		}
		end := plainRegionEnd(lines, i, lines[previous].indent)
		span := document.Span{Start: line.start, End: lines[end-1].end}
		eligible := plainCandidateEligible(lines, i, end)
		if eligible {
			candidates++
		}
		if candidates > plainCodeCandidates {
			return fmt.Errorf("plain-text code recognition exceeds %d candidates", plainCodeCandidates)
		}
		if eligible && plainCExample(ctx, doc.Source[span.Start:span.End]) {
			plain(doc, start, span.Start, "paragraph")
			doc.Excluded = append(doc.Excluded, document.Exclusion{Span: span, Reason: "plain-code"})
			start = span.End
		}
		// Consume a candidate once even when it is unrecognized. Nested colons
		// cannot cause quadratic reparsing or hide a subset of a failed example.
		previous = end - 1
		i = end
	}
	plain(doc, start, len(doc.Source), "paragraph")
	return ctx.Err()
}

func plainCandidateEligible(lines []plainLine, start, end int) bool {
	return end-start <= plainCodeLines && lines[end-1].end-lines[start].start <= plainCodeBytes && plainCodeAnchor(lines[start].text)
}

func plainIntroduced(lines []plainLine, previous, current int) bool {
	return previous >= 0 && strings.HasSuffix(lines[previous].text, ":") && lines[current].indent > lines[previous].indent
}

func plainLines(source []byte) []plainLine {
	var lines []plainLine
	start := 0
	for raw := range strings.SplitAfterSeq(string(source), "\n") {
		lines = append(lines, plainLine{
			text: strings.TrimSpace(raw), start: start, end: start + len(raw), indent: plainIndent(raw),
		})
		start += len(raw)
	}
	return lines
}

func plainIndent(line string) int {
	indent := 0
	for _, char := range line {
		switch char {
		case ' ':
			indent++
		case '\t':
			indent += 8 - indent%8
		default:
			return indent
		}
	}
	return indent
}

func plainRegionEnd(lines []plainLine, start, indent int) int {
	end := start + 1
	for end < len(lines) && (lines[end].text == "" || lines[end].indent > indent) {
		end++
	}
	for end > start+1 && lines[end-1].text == "" {
		end--
	}
	return end
}

var plainCDeclaration = regexp.MustCompile(`^(?:(?:static|extern|inline|const|volatile)\s+)*` +
	`(?:void|int|char|short|long|float|double|unsigned|signed|struct|union|enum)\b`)
var plainCCall = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9]*\([^(){}\r\n]*\);?$`)

func plainCodeAnchor(text string) bool {
	return strings.HasPrefix(text, "#include ") || strings.HasPrefix(text, "#define ") ||
		plainDeclarationAnchor(text) || plainCCall.MatchString(text)
}

func plainDeclarationAnchor(text string) bool {
	return plainCDeclaration.MatchString(text) && strings.ContainsAny(text, "*([=")
}

func plainCExample(ctx context.Context, source []byte) bool {
	text := string(source)
	// A list of bare call signatures is common in release notes. Require two
	// unless the C statement terminator itself provides the missing boundary.
	if calls, ok := plainCallList(text); ok {
		return validPlainC(ctx, "void unswell_example(void) {\n"+calls+"\n}")
	}
	first := strings.TrimSpace(strings.SplitN(text, "\n", 2)[0])
	if !plainDeclarationAnchor(first) && !strings.HasPrefix(first, "#include ") && !strings.HasPrefix(first, "#define ") {
		return false
	}
	text = plainEllipses(text)
	return validPlainC(ctx, text) || validPlainC(ctx, "void unswell_example(void) {\n"+text+"\n}")
}

func plainCallList(text string) (string, bool) {
	var result strings.Builder
	count, terminated := 0, true
	for line := range strings.SplitSeq(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !plainCCall.MatchString(line) {
			return "", false
		}
		count++
		terminated = terminated && strings.HasSuffix(line, ";")
		result.WriteString(strings.TrimSuffix(line, ";"))
		result.WriteString(";\n")
	}
	return result.String(), count >= 2 || (count == 1 && terminated)
}

func plainEllipses(text string) string {
	// Only conventional omitted-statement markers are normalized in this
	// private parsing buffer. The original bytes and exclusions stay exact.
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "..." {
			lines[i] = ""
		} else {
			lines[i] = strings.ReplaceAll(line, "{ ... }", "{ ; }")
		}
	}
	return strings.Join(lines, "\n")
}

func validPlainC(ctx context.Context, text string) bool {
	syntax, err := parseSyntax(ctx, []byte(text), "c")
	if err != nil {
		return false
	}
	syntax.tree.Release()
	return true
}
