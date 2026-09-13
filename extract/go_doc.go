package extract

import (
	"context"
	"fmt"
	"go/ast"
	"go/doc/comment"
	"go/token"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/mapping"
)

type goDocLine struct {
	text       string
	start, end int
}

type goDocReader struct {
	ctx     context.Context
	doc     *document.Document
	options Options
	lines   []goDocLine
	ignore  string
}

func goDocGroup(
	ctx context.Context, doc *document.Document, group *ast.CommentGroup, fset *token.FileSet, options Options, ignore string,
) error {
	r := goDocReader{ctx: ctx, doc: doc, options: options, ignore: ignore}
	for _, item := range group.List {
		start := fset.Position(item.Pos()).Offset
		span := document.Span{Start: start, End: originalCommentEnd(doc.Source, start)}
		if err := r.comment(span, item.Text); err != nil {
			return err
		}
	}
	return r.flush()
}

func (r *goDocReader) comment(span document.Span, text string) error {
	control := directiveText(r.doc.Source, span)
	if strings.HasPrefix(text, "//") && goDocIndented(strings.TrimPrefix(text[2:], " ")) {
		control = "" // A line inside a Go doc example cannot change suppression policy.
	}
	found, err := collectDirective(r.doc, span, control)
	if err != nil {
		return err
	}
	if found || toolDirective(document.Go, control) {
		if !found {
			r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: span, Reason: "directive"})
		}
		return r.flush()
	}
	if contextExclusion(r.doc, r.options.Policy, "comment", span) {
		return r.flush()
	}
	if r.ignore != "" {
		r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: span, Reason: r.ignore})
		return r.flush()
	}
	r.addComment(span, text)
	return nil
}

func (r *goDocReader) addComment(span document.Span, text string) {
	start, end := span.Start+2, span.End
	opener := commentOpener(text)
	if opener != "" {
		end -= 2
	} else if start < end && r.doc.Source[start] == ' ' {
		start++
	}
	for line := range strings.SplitSeq(string(r.doc.Source[start:end]), "\n") {
		contentStart := start
		if opener != "" {
			_, contentStart = blockCommentLine(line, start)
		}
		content := strings.TrimRight(string(r.doc.Source[contentStart:start+len(line)]), " \t\r")
		r.lines = append(r.lines, goDocLine{text: content, start: contentStart, end: start + len(line)})
		start += len(line) + 1
	}
}

func (r *goDocReader) flush() error {
	lines := unindentGoDoc(r.lines)
	r.lines = nil
	spans, err := goDocSpans(r.ctx, lines)
	if err != nil {
		return err
	}
	for _, span := range spans {
		part := lines[span.start:span.end]
		switch span.kind {
		case "code":
			r.exclude(part, "comment-code")
		case "list":
			if err := r.list(part); err != nil {
				return err
			}
		default:
			r.paragraph(part, nil)
		}
		if len(r.doc.Blocks) > r.options.MaxBlocks {
			return fmt.Errorf("source exceeds %d prose blocks", r.options.MaxBlocks)
		}
	}
	return nil
}

func unindentGoDoc(lines []goDocLine) []goDocLine {
	// CommentGroup.Text trims trailing whitespace before the doc grammar runs.
	prefix := ""
	set := false
	for _, line := range lines {
		if line.text == "" {
			continue
		}
		space := line.text[:len(line.text)-len(strings.TrimLeft(line.text, " \t"))]
		if !set {
			prefix, set = space, true
		}
		for !strings.HasPrefix(space, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}
	for i := range lines {
		if lines[i].text != "" {
			lines[i].text = strings.TrimPrefix(lines[i].text, prefix)
			lines[i].start += len(prefix)
		}
	}
	return lines
}

func (r *goDocReader) exclude(lines []goDocLine, reason string) {
	for _, line := range lines {
		if line.start < line.end {
			r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{
				Span: document.Span{Start: line.start, End: line.end}, Reason: reason,
			})
		}
	}
}

func (r *goDocReader) mappedLines(lines []goDocLine) document.MappedText {
	var builder mapping.Builder
	for i, line := range lines {
		if i > 0 && lines[i-1].end < len(r.doc.Source) {
			builder.Add(" ", document.Span{Start: lines[i-1].end, End: lines[i-1].end + 1})
		}
		builder.Source(r.doc.Source, line.start, line.start+len(line.text), false)
	}
	return builder.Build()
}

func (r *goDocReader) paragraph(lines []goDocLine, list *document.ListContext) {
	if goDocLinkDefinitions(lines) {
		r.exclude(lines, "link-definition")
		return
	}
	mapped := r.mappedLines(lines)
	kind := "paragraph"
	if list != nil {
		kind = "list-item"
	} else if len(lines) == 1 && strings.HasPrefix(mapped.Text, "#") && goDocIndented(mapped.Text[1:]) {
		kind = "heading"
		mapped.Text, mapped.Map = mapped.Text[1:], mapped.Map[1:]
	}
	before := len(r.doc.Blocks)
	appendBlock(r.doc, mapped, "comment")
	if len(r.doc.Blocks) > before {
		block := &r.doc.Blocks[before]
		block.List = list
		if r.options.IncludeStructure {
			block.Context = []string{"embedded:godoc:" + kind}
		}
	}
}

func goDocLinkDefinitions(lines []goDocLine) bool {
	if len(lines) == 0 || !strings.HasPrefix(lines[0].text, "[") {
		return false
	}
	text := make([]string, len(lines))
	for i, line := range lines {
		text[i] = line.text
	}
	var parser comment.Parser
	parsed := parser.Parse(strings.Join(text, "\n"))
	return len(parsed.Content) == 0 && len(parsed.Links) > 0
}
