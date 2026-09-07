package extract

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/mapping"
)

func literalSource(builder *mapping.Builder, source []byte, start, end int, mode string) error {
	segment := string(source[start:end])
	for pos := start; pos < end; {
		text, size, err := literalRune(segment[pos-start:], mode)
		if err != nil {
			return fmt.Errorf("string escape at byte %d: %w", pos, err)
		}
		for _, char := range text {
			if unicode.IsControl(char) && char != '\n' && char != '\r' && char != '\t' {
				text = " \x00 "
				break
			}
		}
		builder.Add(text, document.Span{Start: pos, End: pos + size})
		pos += size
	}
	return nil
}

func literalRune(source, mode string) (string, int, error) {
	quotes := map[string]string{"csharp-verbatim": "\"", "powershell-single": "'"}
	if quote := quotes[mode]; quote != "" && strings.HasPrefix(source, quote+quote) {
		return quote, 2, nil
	}
	if mode == "powershell" && source[0] == '`' {
		return powershellEscape(source)
	}
	if !slices.Contains([]string{"raw", "csharp-verbatim", "powershell-single", "powershell"}, mode) && source[0] == '\\' {
		return backslashEscape(source, mode)
	}
	r, size := utf8.DecodeRuneInString(source)
	if r == '\r' || r == '\ufeff' {
		return " ", size, nil
	}
	return source[:size], size, nil
}

func backslashEscape(source, mode string) (string, int, error) {
	if len(source) < 2 {
		return "", 0, fmt.Errorf("unfinished escape")
	}
	if slices.Contains([]string{"shell", "fish-single", "fish-double"}, mode) {
		return shellEscape(source, mode)
	}
	if javascriptIdentityEscape(source, mode) {
		return otherEscape(source, mode)
	}
	if source[1] == '\n' || source[1] == '\r' {
		return "", continuationSize(source, mode), nil
	}
	if source[1] == 'u' || source[1] == 'U' {
		return unicodeEscape(source)
	}
	if strings.ContainsRune("x01234567", rune(source[1])) {
		return numericEscape(source, mode)
	}
	escapes := map[byte]string{
		'a':  "\a",
		'b':  "\b",
		'f':  "\f",
		'n':  "\n",
		'r':  "\r",
		't':  "\t",
		'v':  "\v",
		'\\': "\\",
		'\'': "'",
		'"':  "\"",
	}
	if value, ok := escapes[source[1]]; ok {
		return value, 2, nil
	}
	return otherEscape(source, mode)
}

func shellEscape(source, mode string) (string, int, error) {
	allowed := "$`\"\\\n"
	if mode == "fish-single" {
		allowed = "'\\"
	}
	if mode == "fish-double" {
		allowed = "$\"\\\n"
	}
	if !strings.ContainsRune(allowed, rune(source[1])) {
		return "\\", 1, nil
	}
	if source[1] == '\n' {
		return "", 2, nil
	}
	return source[1:2], 2, nil
}

func continuationSize(source, mode string) int {
	size := 2
	if source[1] == '\r' && len(source) > 2 && source[2] == '\n' {
		size++
	}
	if mode == "rust" {
		for size < len(source) && strings.ContainsRune(" \t\r\n", rune(source[size])) {
			size++
		}
	}
	return size
}

func otherEscape(source, mode string) (string, int, error) {
	if slices.Contains([]string{"javascript", "typescript", "tsx"}, mode) {
		_, size := utf8.DecodeRuneInString(source[1:])
		return source[1 : size+1], size + 1, nil
	}
	if mode == "ansi" && (source[1] == 'e' || source[1] == 'E') {
		return "\x1b", 2, nil
	}
	if mode == "python" && source[1] != 'N' {
		return "\\", 1, nil
	}
	if mode == "java" && source[1] == 's' {
		return " ", 2, nil
	}
	return "", 0, fmt.Errorf("unsupported %s escape %q", mode, source[:min(len(source), 2)])
}

func javascriptIdentityEscape(source, mode string) bool {
	if !slices.Contains([]string{"javascript", "typescript", "tsx"}, mode) {
		return false
	}
	return !strings.ContainsRune("bfnrtv01234567xu\\\"'\n\r", rune(source[1]))
}

func unicodeEscape(source string) (string, int, error) {
	start, end, size := unicodeDigits(source)
	if end > len(source) || end <= start || size > len(source) {
		return "", 0, fmt.Errorf("invalid Unicode escape")
	}
	digits := strings.ReplaceAll(source[start:end], "_", "")
	value, err := strconv.ParseUint(digits, 16, 32)
	if err != nil {
		return "", 0, fmt.Errorf("invalid Unicode escape: %w", err)
	}
	if value >= 0xd800 && value <= 0xdbff && strings.HasPrefix(source[size:], "\\u") {
		return surrogateEscape(source, uint32(value), size)
	}
	if value > utf8.MaxRune || utf16.IsSurrogate(rune(value)) {
		return "", 0, fmt.Errorf("invalid Unicode scalar value")
	}
	return string(rune(value)), size, nil
}

func unicodeDigits(source string) (int, int, int) {
	if len(source) > 2 && source[2] == '{' {
		end := strings.IndexByte(source, '}')
		return 3, end, end + 1
	}
	width := 4
	if source[1] == 'U' {
		width = 8
	}
	return 2, width + 2, width + 2
}

func surrogateEscape(source string, high uint32, size int) (string, int, error) {
	if high < 0xd800 || high > 0xdbff {
		return "", 0, fmt.Errorf("invalid leading Unicode surrogate")
	}
	if len(source) < size+6 {
		return "", 0, fmt.Errorf("unfinished Unicode surrogate pair")
	}
	low, err := strconv.ParseUint(source[size+2:size+6], 16, 32)
	if err != nil || low < 0xdc00 || low > 0xdfff {
		return "", 0, fmt.Errorf("invalid Unicode surrogate pair")
	}
	return string(utf16.DecodeRune(rune(high), rune(low))), size + 6, nil
}

func numericEscape(source, mode string) (string, int, error) {
	if mode == "csharp" && source[1] == '0' {
		return "\x00", 2, nil
	}
	base, start, limit := numericDigits(source, mode)
	end := start
	for end < min(limit, len(source)) && digitInBase(source[end], base) {
		end++
	}
	value, err := strconv.ParseUint(source[start:end], base, 32)
	if err != nil || value > utf8.MaxRune {
		return "", 0, fmt.Errorf("invalid numeric escape")
	}
	if utf16.IsSurrogate(rune(value)) {
		return "", 0, fmt.Errorf("numeric escape is not a Unicode scalar value")
	}
	if slices.Contains([]string{"go", "c", "cpp", "ansi"}, mode) {
		if value > 255 {
			return "", 0, fmt.Errorf("byte escape exceeds 255")
		}
		return string([]byte{byte(value)}), end, nil
	}
	return string(rune(value)), end, nil
}

func numericDigits(source, mode string) (int, int, int) {
	if source[1] != 'x' {
		return 8, 1, 4
	}
	if mode == "c" || mode == "cpp" {
		return 16, 2, len(source)
	}
	if mode == "csharp" {
		return 16, 2, 6
	}
	return 16, 2, 4
}

func digitInBase(char byte, base int) bool {
	if base == 8 {
		return char >= '0' && char <= '7'
	}
	return char >= '0' && char <= '9' || char >= 'a' && char <= 'f' || char >= 'A' && char <= 'F'
}

func powershellEscape(source string) (string, int, error) {
	if len(source) < 2 {
		return "", 0, fmt.Errorf("unfinished PowerShell escape")
	}
	if strings.HasPrefix(source, "`u{") {
		return unicodeEscape(source)
	}
	escapes := map[byte]string{'0': "\x00", 'a': "\a", 'b': "\b", 'e': "\x1b", 'f': "\f", 'n': "\n", 'r': "\r", 't': "\t", 'v': "\v"}
	if value, ok := escapes[source[1]]; ok {
		return value, 2, nil
	}
	if source[1] == '\n' || source[1] == '\r' {
		return "", continuationSize(source, "powershell"), nil
	}
	_, size := utf8.DecodeRuneInString(source[1:])
	return source[1 : size+1], size + 1, nil
}
