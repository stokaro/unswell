// Package pathglob compiles bounded project-relative path patterns.
package pathglob

import (
	"fmt"
	"regexp"
	"strings"
)

// Compile validates and compiles project-relative *, **, and ? patterns.
func Compile(patterns []string) ([]*regexp.Regexp, error) {
	result := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if len(pattern) > 500 || strings.ContainsAny(pattern, "[]{}\\") {
			return nil, fmt.Errorf("unsupported glob %q; use *, **, and ?", pattern)
		}
		expression := globExpression(pattern)
		compiled, err := regexp.Compile(expression)
		if err != nil {
			return nil, err
		}
		result = append(result, compiled)
	}
	return result, nil
}

// Matches reports whether any compiled pattern matches a slash-separated path.
func Matches(name string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(name) {
			return true
		}
	}
	return false
}

func globExpression(pattern string) string {
	var expression strings.Builder
	expression.WriteByte('^')
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '*':
			fragment, end := globStar(pattern, i)
			expression.WriteString(fragment)
			i = end
		case '?':
			expression.WriteString("[^/]")
		default:
			expression.WriteString(regexp.QuoteMeta(pattern[i : i+1]))
		}
	}

	expression.WriteByte('$')
	return expression.String()
}

func globStar(pattern string, index int) (string, int) {
	if index+1 >= len(pattern) || pattern[index+1] != '*' {
		return "[^/]*", index
	}
	index++
	if index+1 < len(pattern) && pattern[index+1] == '/' {
		return "(?:.*/)?", index + 1
	}
	return ".*", index
}
