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

// comment appends the prose lines of one comment to the builder and records
// the lines it leaves out: indented examples, and the tag sections of a
// documentation comment. A blank line or a tag ends the current block, so the
// description before the first tag keeps its own block.
func (w *commentWriter) comment(start, end int, opener string) {
	w.tags = opener == "/**" && slices.Contains(docTagFormats, w.doc.Format)
	w.section = ""
	block := opener != ""
	if block {
		w.markup, w.depth = "", 0
	}
	pos := start
	for line := range strings.SplitSeq(string(w.doc.Source[start:end]), "\n") {
		lineStart := pos
		lineSize := len(line)
		if block {
			line, lineStart = blockCommentLine(line, pos)
		}
		w.line(line, lineStart)
		pos += lineSize + 1
		if pos <= end {
			w.builder.Add(" ", document.Span{Start: pos - 1, End: pos})
		}
	}
}

type commentWriter struct {
	doc     *document.Document
	builder *mapping.Builder
	tags    bool   // recognize documentation tags
	section string // exclusion reason of the open tag section, or ""
	markup  string // token that closes the open code region, or ""
	depth   int    // open braces of a "{@code}" region
}

func (w *commentWriter) line(line string, start int) {
	span := document.Span{Start: start, End: start + len(line)}
	if w.markup != "" {
		w.emit(line, start)
		return
	}
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
		w.hold(span)
		return
	}
	w.emit(line, start)
}

func (w *commentWriter) flush() {
	appendBlock(w.doc, w.builder.Build(), "comment")
	*w.builder = mapping.Builder{}
}

func (w *commentWriter) exclude(span document.Span, reason string) {
	w.doc.Excluded = append(w.doc.Excluded, document.Exclusion{Span: span, Reason: reason})
}

// hold replaces a code region with a break, so that the prose before it and
// the prose after it do not read as one sentence.
func (w *commentWriter) hold(span document.Span) {
	w.builder.Add(" \x00 ", span)
	w.exclude(span, "comment-code")
}

// markupRegions are the documentation markup regions that carry code rather
// than prose. Each is named by the token that opens it and the token that
// closes it; a region stays open across lines until its closer appears.
var markupRegions = []struct{ open, close string }{
	{"<pre", "</pre>"},
	{"<code", "</code>"},
	{"<c", "</c>"},
	{"{@code", "}"},
	{"{@literal", "}"},
}

// emit writes one comment line and holds back the code regions in it.
// Javadoc and JSDoc put examples in the HTML "pre" and "code" elements. C#
// documentation comments use "code" and the inline "c". Javadoc also has the
// inline tags "{@code}" and "{@literal}".
func (w *commentWriter) emit(line string, start int) {
	for pos := 0; pos < len(line); {
		if w.markup == "" {
			offset, closer := openMarkup(line[pos:])
			if offset < 0 {
				w.builder.Source(w.doc.Source, start+pos, start+len(line), false)
				return
			}
			w.builder.Source(w.doc.Source, start+pos, start+pos+offset, false)
			pos += offset
			w.markup, w.depth = closer, 0
		}
		end := len(line)
		if closed := w.closeMarkup(line[pos:]); closed >= 0 {
			end = pos + closed
			w.markup = ""
		}
		w.hold(document.Span{Start: start + pos, End: start + end})
		pos = end
	}
}

// openMarkup returns the offset of the next code region in line and the token
// that closes it, or -1 when the line opens none.
func openMarkup(line string) (int, string) {
	for i := range len(line) {
		if line[i] != '<' && line[i] != '{' {
			continue
		}
		for _, region := range markupRegions {
			if markupOpen(line[i:], region.open) {
				return i, region.close
			}
		}
	}
	return -1, ""
}

// markupOpen reports whether text opens the named region. A name ends at a
// delimiter, so "<pre>" and "<pre class='java'>" open the "pre" region while
// "<precondition>" opens nothing.
func markupOpen(text, name string) bool {
	if len(text) < len(name) || !strings.EqualFold(text[:len(name)], name) {
		return false
	}
	rest := text[len(name):]
	return rest == "" || rest[0] == '>' || rest[0] == '}' || rest[0] == ' ' || rest[0] == '\t'
}

// closeMarkup returns the offset past the token that closes the open region
// in line, or -1 when the region stays open past the end of the line. Braces
// nest, so a "{@code}" region ends at the brace that balances its opener.
func (w *commentWriter) closeMarkup(line string) int {
	if w.markup != "}" {
		if i := indexFold(line, w.markup); i >= 0 {
			return i + len(w.markup)
		}
		return -1
	}
	for i := range len(line) {
		switch line[i] {
		case '{':
			w.depth++
		case '}':
			w.depth--
			if w.depth == 0 {
				return i + 1
			}
		}
	}
	return -1
}

// indexFold returns the offset of the first case-insensitive occurrence of
// name in text, or -1. Markup names are ASCII, so folding one byte at a time
// keeps the returned offset in the original text.
func indexFold(text, name string) int {
	for i := 0; i+len(name) <= len(text); i++ {
		if strings.EqualFold(text[i:i+len(name)], name) {
			return i
		}
	}
	return -1
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
