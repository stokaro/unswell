package extract

import (
	"bytes"
	"context"
	"fmt"

	ts "github.com/odvcencio/gotreesitter"

	"github.com/stokaro/unswell/document"
)

// inlineDirectiveSyntax recognizes Unswell's double-hyphen reason separator
// when the upstream inline grammar treats that comment as literal prose.
// A trial parse must identify the entire candidate as an HTML tag. Candidates
// inside code, attribute values, or escaped markup never satisfy that check.
func inlineDirectiveSyntax(ctx context.Context, original syntaxTree, source []byte) (syntaxTree, error) {
	if !bytes.Contains(source, []byte("<!--")) {
		return original, nil
	}
	trial := bytes.Clone(source)
	candidates, err := inlineDirectiveCandidates(ctx, trial)
	if err != nil {
		original.tree.Release()
		return syntaxTree{}, err
	}
	if len(candidates) == 0 {
		return original, nil
	}
	boundaries, err := parsedDirectiveBoundaries(ctx, trial, candidates)
	if err != nil {
		original.tree.Release()
		return syntaxTree{}, err
	}
	if len(boundaries) == 0 {
		return original, nil
	}
	defer original.tree.Release()
	for span := range boundaries {
		maskDirectiveHyphens(source[span.Start+4 : span.End-3])
	}
	parsed, err := parseSyntax(ctx, source, "markdown_inline")
	if err != nil {
		return syntaxTree{}, err
	}
	if err := verifyInlineBoundaries(ctx, parsed, boundaries); err != nil {
		parsed.tree.Release()
		return syntaxTree{}, err
	}
	return parsed, nil
}

func inlineDirectiveCandidates(ctx context.Context, source []byte) (map[document.Span]bool, error) {
	candidates := make(map[document.Span]bool)
	for pos := 0; pos < len(source); {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		start := bytes.IndexByte(source[pos:], '<')
		if start < 0 {
			break
		}
		start += pos
		if !bytes.HasPrefix(source[start:], []byte("<!--")) {
			pos = inlineMarkupEnd(source, start)
			continue
		}
		end := bytes.Index(source[start+4:], []byte("-->"))
		if end < 0 {
			break
		}
		end += start + 4
		pos = end + 3
		body := source[start+4 : end]
		if !directiveCandidate(string(body)) || !bytes.Contains(body, []byte("--")) {
			continue
		}
		if len(candidates) == 1000 {
			return nil, fmt.Errorf("inline directive candidates exceed 1000")
		}
		maskDirectiveHyphens(body)
		candidates[document.Span{Start: start, End: pos}] = true
	}
	return candidates, nil
}

// inlineMarkupEnd skips a possible tag, including quoted attribute values.
// Invalid markup can be literal to the upstream grammar; its attribute text
// must still not become policy when a trial repairs an embedded comment.
func inlineMarkupEnd(source []byte, start int) int {
	if start+1 == len(source) || !bytes.ContainsAny(source[start+1:start+2], "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ/!?") {
		return start + 1
	}
	var quote byte
	for pos := start + 1; pos < len(source); pos++ {
		switch char := source[pos]; {
		case quote != 0:
			if char == quote {
				quote = 0
			}
		case char == '\'' || char == '"':
			quote = char
		case char == '>':
			return pos + 1
		}
	}
	return len(source)
}

func maskDirectiveHyphens(body []byte) {
	for i := 0; i+1 < len(body); i++ {
		if body[i] == '-' && body[i+1] == '-' {
			body[i], body[i+1] = ' ', ' '
		}
	}
}

func parsedDirectiveBoundaries(ctx context.Context, source []byte, candidates map[document.Span]bool) (map[document.Span]string, error) {
	parsed, err := parseSyntax(ctx, source, "markdown_inline")
	if err != nil {
		return nil, err
	}
	defer parsed.tree.Release()
	boundaries := make(map[document.Span]string)
	err = walkSyntax(ctx, parsed.tree.RootNode(), 0, func(node *ts.Node) (bool, error) {
		if node.Type(parsed.lang) != "html_tag" {
			return false, nil
		}
		span := syntaxSpan(node, 0)
		if candidates[span] {
			boundaries[span] = "html_tag"
		}
		return true, nil
	})
	return boundaries, err
}
