package extract

import (
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
)

// Shell strings and heredocs commonly carry programs for other interpreters:
// jq filters, awk scripts, SQL statements and Python heredocs. Those literals
// are excluded when their lines look like code. The test is a share of lines,
// not a parse, so it can be wrong in both directions; the floor keeps short
// literals and one-line messages out of it, and a configured exception
// outranks it.
const (
	programLineFloor = 3
	programCodeShare = 60 // percent of nonblank lines
)

// programFormat reports whether a format's string literals are tested for
// embedded programs. Other grammars keep every literal unless an exception
// selects it.
func programFormat(format document.Format) bool {
	return slices.Contains([]document.Format{
		document.Bash, document.Shell, document.Zsh, document.Fish, document.PowerShell,
	}, format)
}

// programLiteral reports whether a literal's content is a program rather than
// prose. Content that starts with a shebang is a program. Otherwise it needs
// at least programLineFloor nonblank lines, of which at least programCodeShare
// percent look like code.
func programLiteral(content string) bool {
	lines, code := 0, 0
	for line := range strings.SplitSeq(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if lines == 0 && strings.HasPrefix(trimmed, "#!") {
			return true
		}
		lines++
		if codeLine(trimmed) {
			code++
		}
	}
	return lines >= programLineFloor && code*100 >= programCodeShare*lines
}

// codeKeywords open statements in the interpreters that shell scripts embed.
// Prose lines start with a capital letter, so the lowercase spellings rarely
// collide with English; the SQL keywords are listed in their usual uppercase.
var codeKeywords = []string{
	"case", "class", "const", "def", "do", "done", "echo", "elif", "else", "esac", "except", "exit", "export", "fi",
	"for", "from", "function", "if", "import", "let", "local", "print", "printf", "raise", "return", "then", "try",
	"var", "while", "with", "yield",
	"CREATE", "DELETE", "FROM", "INSERT", "SELECT", "UPDATE", "WHERE",
}

// codeOperators are token sequences that English prose does not contain.
var codeOperators = []string{"=>", "->", "::", "&&", "||", "$(", "${", "==", "!=", "<=", ">=", "+=", "-=", " = ", " | "}

// codeLine reports whether one trimmed, nonblank line looks like code. A line
// that ends a sentence is never code. A line that ends with a brace or a
// semicolon is code, and so is a line that starts with a pipe. So is a line
// that holds an operator, an assignment or a program keyword. So is a line
// with two or more brackets, pipes, semicolons or equals signs.
func codeLine(line string) bool {
	if sentenceEnd(line) {
		return false
	}
	if strings.ContainsAny(line[len(line)-1:], "{};") || strings.HasPrefix(line, "|") {
		return true
	}
	for _, operator := range codeOperators {
		if strings.Contains(line, operator) {
			return true
		}
	}
	return assignmentLine(line) || keywordLine(line) || punctuationCount(line) >= 2
}

// punctuationCount counts the brackets, pipes, semicolons and equals signs
// that programs use far more often than prose.
func punctuationCount(line string) int {
	count := 0
	for i := range len(line) {
		if strings.IndexByte("(){}[]<>|;=", line[i]) >= 0 {
			count++
		}
	}
	return count
}

func sentenceEnd(line string) bool {
	line = strings.TrimRight(line, "\"'")
	return line != "" && strings.ContainsAny(line[len(line)-1:], ".?!")
}

// assignmentLine matches "name=value" with no space before the equals sign.
func assignmentLine(line string) bool {
	equals := strings.IndexByte(line, '=')
	if equals <= 0 {
		return false
	}
	for i := range equals {
		if !wordByte(line[i]) && line[i] != '_' {
			return false
		}
	}
	return true
}

func keywordLine(line string) bool {
	word := line
	if end := strings.IndexAny(line, " \t(:;"); end >= 0 {
		word = line[:end]
	}
	return slices.Contains(codeKeywords, word)
}
