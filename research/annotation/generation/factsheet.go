package generation

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// FactSheetVersion identifies the deterministic extractor. A change to what
// it reads or how it formats the sheet is a new version.
const FactSheetVersion = "unswell-fact-sheet-v1"

// FactSheet is what a generator sees instead of the original wording: the
// declaration that follows the documented unit, the identifiers and
// parameter list in it, and the numbers the original states. It copies no
// sentence of the original.
type FactSheet struct {
	Version     string   `json:"version"`
	Signature   string   `json:"signature"`
	Identifiers []string `json:"identifiers"`
	Parameters  string   `json:"parameters"`
	Numbers     []string `json:"numbers"`
	SHA256      string   `json:"sha256"`
}

var (
	identifierPattern = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)
	numberPattern     = regexp.MustCompile(`\b\d+(?:\.\d+)?\b`)
	declarationStart  = regexp.MustCompile(`^(?:pub(?:\([a-z]+\))?\s+|public\s+|private\s+|protected\s+|internal\s+|static\s+|export\s+|` +
		`async\s+|final\s+|abstract\s+|virtual\s+|override\s+|inline\s+|extern\s+|template\s*<[^>]*>\s*|unsafe\s+|const\s+)*` +
		`(?:func|fn|def|class|struct|enum|trait|impl|interface|type|function|void|int|bool|string|char|float|double|long|` +
		`size_t|unsigned|signed|auto|var|let|const|static|namespace|module|mod|record|delegate|event|property|` +
		`[A-Za-z_][A-Za-z0-9_:<>\[\]\*&,\. ]*\s+[A-Za-z_][A-Za-z0-9_]*\s*\()`)
	commentStart    = regexp.MustCompile(`^\s*(//|#|/\*|\*|--|'''|"""|<#|;)`)
	trailingComment = regexp.MustCompile(`\s+(//|#)[^"'\x60]*$`)
	callOpen        = regexp.MustCompile(`[A-Za-z0-9_>\)\]]\s*\(`)
	keywords        = []string{"func", "fn", "def", "class", "struct", "enum", "trait", "impl", "interface", "type", "function",
		"void", "int", "bool", "string", "char", "float", "double", "long", "auto", "var", "let", "const", "static", "pub",
		"public", "private", "protected", "internal", "export", "async", "final", "abstract", "virtual", "override",
		"return", "self", "this", "new", "namespace", "module", "mod", "record", "delegate", "event", "property", "unsigned",
		"signed", "size_t", "template", "typename", "extern", "inline", "unsafe", "readonly", "sealed", "partial"}
)

// ExtractFactSheet reads the declaration that begins after byte offset end
// of the source. It returns false when no declaration follows within the
// next lines, so the unit is not a documentation task.
func ExtractFactSheet(source []byte, end int, original string) (FactSheet, bool) {
	if end < 0 || end > len(source) {
		return FactSheet{}, false
	}
	signature, ok := declarationAfter(string(source[end:]))
	if !ok {
		return FactSheet{}, false
	}
	sheet := FactSheet{Version: FactSheetVersion, Signature: signature, Identifiers: identifiers(signature),
		Parameters: parameters(signature), Numbers: numbers(original)}
	sheet.SHA256 = fmt.Sprintf("%x", sha256.Sum256([]byte(sheet.Text())))
	return sheet, true
}

// declarationAfter returns the declaration header after the unit. The header
// must start within the first three non-empty lines. It runs to the first
// line that ends in "{", ":", ";", "=", or ")", or to the end of the block.
// Comment lines before the declaration are skipped. A blank line before any
// declaration means none follows.
func declarationAfter(rest string) (string, bool) {
	lines := strings.Split(rest, "\n")
	if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" && !strings.Contains(lines[0], "*/") {
		// The unit ends mid-line: the rest of that line is its own terminator.
		lines[0] = ""
	}
	start, ok := declarationLine(lines)
	if !ok {
		return "", false
	}
	signature := strings.TrimRight(strings.Join(headerLines(lines, start), " "), " {:;=")
	if len(signature) < 3 || len(signature) > 600 {
		return "", false
	}
	return signature, true
}

// declarationLine finds the line that starts the declaration: the first
// line after the unit that is neither blank nor a comment, within twelve
// lines, provided no blank line comes first.
func declarationLine(lines []string) (int, bool) {
	for i, line := range lines {
		if i > 12 {
			break
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if i > 0 {
				return 0, false
			}
			continue
		}
		if trimmed == "*/" || trimmed == "#>" || commentStart.MatchString(trimmed) {
			continue
		}
		return i, declarationStart.MatchString(trimmed)
	}
	return 0, false
}

// headerLines collects the declaration header from its first line to the
// line that ends it, with trailing line comments removed.
func headerLines(lines []string, start int) []string {
	header := []string{}
	for i := start; i < len(lines) && i < start+8; i++ {
		trimmed := trailingComment.ReplaceAllString(strings.TrimSpace(lines[i]), "")
		if trimmed == "" {
			break
		}
		header = append(header, trimmed)
		if headerEnds(trimmed) {
			break
		}
	}
	return header
}

func headerEnds(line string) bool {
	for _, suffix := range []string{"{", ":", ";", "=", ")", "{}"} {
		if strings.HasSuffix(line, suffix) {
			return true
		}
	}
	return false
}

func identifiers(signature string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, name := range identifierPattern.FindAllString(signature, -1) {
		if seen[name] || slices.Contains(keywords, name) || len(name) < 2 {
			continue
		}
		seen[name] = true
		result = append(result, name)
	}
	return result
}

// parameters returns the list inside the first call-like parentheses, those
// that follow a name or a type; a pattern or a grouping in an initializer is
// not a parameter list.
func parameters(signature string) string {
	search := signature
	if strings.HasPrefix(signature, "func (") {
		// A Go receiver comes first; the parameter list follows the method name.
		if close := strings.Index(signature, ")"); close >= 0 {
			search = signature[close+1:]
		}
	}
	match := callOpen.FindStringIndex(search)
	if match == nil {
		return ""
	}
	signature = search
	open := match[1] - 1
	depth := 0
	for i := open; i < len(signature); i++ {
		switch signature[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return strings.TrimSpace(signature[open+1 : i])
			}
		}
	}
	return strings.TrimSpace(signature[open+1:])
}

func numbers(original string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range numberPattern.FindAllString(original, -1) {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

// Text renders the sheet as the material a generator receives.
func (f FactSheet) Text() string {
	var b strings.Builder
	b.WriteString("Signature:\n")
	b.WriteString(f.Signature)
	b.WriteString("\n\nIdentifiers: ")
	if len(f.Identifiers) == 0 {
		b.WriteString("none")
	} else {
		b.WriteString(strings.Join(f.Identifiers, ", "))
	}
	b.WriteString("\nParameters: ")
	if f.Parameters == "" {
		b.WriteString("none")
	} else {
		b.WriteString(f.Parameters)
	}
	b.WriteString("\nNumbers: ")
	if len(f.Numbers) == 0 {
		b.WriteString("none")
	} else {
		b.WriteString(strings.Join(f.Numbers, ", "))
	}
	b.WriteString("\n")
	return b.String()
}
