package extract

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in licenses/go_LICENSE.
// Adapted from go/doc/comment/parse.go at Go 1.25.0, commit
// 6e676ab2b809d46623acb5988248d95d1eb7939c. The block scanner retains source
// line ranges; prose headings and paragraphs share one analysis path.

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

type goDocSpan struct {
	start, end int
	kind       string
}

type goDocScanner struct {
	lines       []goDocLine
	forceIndent int
}

func goDocSpans(ctx context.Context, lines []goDocLine) ([]goDocSpan, error) {
	scanner := goDocScanner{lines: lines}
	var result []goDocSpan
	remaining := 2 * len(lines)
	for pos := 0; pos < len(lines); {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if lines[pos].text == "" {
			pos++
			continue
		}
		remaining--
		if remaining < 0 {
			return nil, fmt.Errorf("go comment block scanner made no progress")
		}
		span := scanner.next(pos)
		if span.end > pos {
			result = append(result, span)
		}
		pos = span.end
	}
	return result, nil
}

func (s *goDocScanner) next(start int) goDocSpan {
	if start < s.forceIndent || goDocIndented(s.lines[start].text) {
		return s.indented(start)
	}
	end := start + 1
	for end < len(s.lines) && s.lines[end].text != "" && !goDocIndented(s.lines[end].text) {
		end++
	}
	return goDocSpan{start: start, end: s.backtrack(start, end), kind: "paragraph"}
}

func (s *goDocScanner) indented(start int) goDocSpan {
	allowList := goDocIsList(s.lines[start].text) && start < s.forceIndent
	end := start + 1
	for end < len(s.lines) {
		line := s.lines[end].text
		if !s.acceptIndented(end, allowList) {
			break
		}
		if line == "" {
			allowList = false
		}
		end++
	}
	for end > start && s.lines[end-1].text == "" {
		end--
	}
	if end < len(s.lines) && strings.HasPrefix(s.lines[end].text, "}") {
		end++
	}
	kind := "code"
	if goDocIsList(s.lines[start].text) {
		kind = "list"
	}
	return goDocSpan{start: start, end: end, kind: kind}
}

func (s *goDocScanner) acceptIndented(pos int, allowList bool) bool {
	line := s.lines[pos].text
	return line == "" || pos < s.forceIndent || goDocIndented(line) || allowList && goDocIsList(line)
}

func (s *goDocScanner) backtrack(start, end int) int {
	if end == len(s.lines) || s.lines[end].text == "" || goDocIsList(s.lines[end].text) {
		return end
	}
	last := s.lines[end-1].text
	if goDocIsList(last) {
		s.forceIndent = end
		for end > start && goDocIsList(s.lines[end-1].text) {
			end--
		}
	} else if strings.HasSuffix(last, "{") || strings.HasSuffix(last, `\`) {
		s.forceIndent = end
		end--
	}
	return end
}

func goDocIndented(line string) bool {
	return line != "" && (line[0] == ' ' || line[0] == '\t')
}

func goDocIsList(line string) bool {
	_, _, ok := goDocListMarker(line)
	return ok
}

func goDocListMarker(line string) (number, rest string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", false
	}
	char, size := utf8.DecodeRuneInString(line)
	if char == '\u2022' || char == '*' || char == '+' || char == '-' {
		rest = line[size:]
	} else {
		var found bool
		number, rest, found = goDocNumberMarker(line)
		if !found {
			return "", "", false
		}
	}
	return number, rest, goDocIndented(rest) && strings.TrimSpace(rest) != ""
}

func goDocNumberMarker(line string) (number, rest string, ok bool) {
	n := 0
	for n < len(line) && line[n] >= '0' && line[n] <= '9' {
		n++
	}
	if n == 0 || n == len(line) || (line[n] != '.' && line[n] != ')') {
		return "", "", false
	}
	return line[:n], line[n+1:], true
}
