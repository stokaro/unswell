// Package textutil shares text boundaries without changing source mappings.
package textutil

import (
	"strings"
	"unicode"
)

// TrimSpaceBounds returns the UTF-8 byte interval retained by strings.TrimSpace.
// An all-whitespace input retains an empty interval at its end.
func TrimSpaceBounds(text string) (int, int) {
	start := len(text) - len(strings.TrimLeftFunc(text, unicode.IsSpace))
	return start, start + len(strings.TrimSpace(text))
}
