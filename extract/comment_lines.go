package extract

import (
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/mapping"
)

// commentOpener returns the block comment opener of a comment, or "" for a
// line comment. A "/**" opener marks a documentation comment.
func commentOpener(text string) string {
	for _, opener := range []string{"/**", "/*", "<#"} {
		if strings.HasPrefix(text, opener) {
			return opener
		}
	}
	return ""
}

// docTagFormats are the formats whose "/**" comments follow the JSDoc and
// Javadoc tag grammar: a line that starts with "@name" opens a tag section
// that runs to the next tag or the end of the comment.
var docTagFormats = []document.Format{document.JavaScript, document.TypeScript, document.TSX, document.Java}

// addCommentLines appends the prose lines of one comment to the builder and
// records the lines it leaves out: indented examples, and the tag sections of
// a documentation comment. A blank line or a tag ends the current block, so
// the description before the first tag keeps its own block.
func addCommentLines(doc *document.Document, builder *mapping.Builder, start, end int, opener string) {
	writer := commentWriter{doc: doc, builder: builder, tags: opener == "/**" && slices.Contains(docTagFormats, doc.Format)}
	block := opener != ""
	pos := start
	for line := range strings.SplitSeq(string(doc.Source[start:end]), "\n") {
		lineStart := pos
		lineSize := len(line)
		if block {
			line, lineStart = blockCommentLine(line, pos)
		}
		writer.line(line, lineStart)
		pos += lineSize + 1
		if pos <= end {
			builder.Add(" ", document.Span{Start: pos - 1, End: pos})
		}
	}
}

type commentWriter struct {
	doc     *document.Document
	builder *mapping.Builder
	tags    bool   // recognize documentation tags
	section string // exclusion reason of the open tag section, or ""
}

func (w *commentWriter) line(line string, start int) {
	span := document.Span{Start: start, End: start + len(line)}
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		w.flush()
		return
	}
	if w.tags {
		if tag := docTag(trimmed); tag != "" {
			w.flush()
			w.section = docTagReason(tag)
		}
		if w.section != "" {
			w.exclude(span, w.section)
			return
		}
	}
	if strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "    ") {
		w.builder.Add(" \x00 ", span)
		w.exclude(span, "comment-code")
		return
	}
	w.builder.Source(w.doc.Source, span.Start, span.End, false)
}

func (w *commentWriter) flush() {
	appendBlock(w.doc, w.builder.Build(), "comment")
	*w.builder = mapping.Builder{}
}

func (w *commentWriter) exclude(span document.Span, reason string) {
	w.doc.Excluded = append(w.doc.Excluded, document.Exclusion{Span: span, Reason: reason})
}

// docTag returns the block tag that opens a documentation comment line, such
// as "param" for "@param {Date} date", or "" when the line is not a tag line.
func docTag(trimmed string) string {
	if len(trimmed) < 2 || trimmed[0] != '@' || !asciiLetter(trimmed[1]) {
		return ""
	}
	end := 1
	for end < len(trimmed) && wordByte(trimmed[end]) {
		end++
	}
	return trimmed[1:end]
}

// docTagReason names the exclusion for a tag section. An example holds code;
// every other tag holds a type, a name, a reference or a short phrase.
func docTagReason(tag string) string {
	if tag == "example" {
		return "doc-example"
	}
	return "doc-tag"
}

func asciiLetter(b byte) bool {
	return ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z')
}

func blockCommentLine(line string, start int) (string, int) {
	trimmed := strings.TrimLeft(line, " \t")
	if trimmed == "*" || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "*\t") {
		removed := len(line) - len(trimmed) + 1
		return line[removed:], start + removed
	}
	return line, start
}
