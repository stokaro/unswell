package extract

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/stokaro/unswell/document"
)

func directiveText(source []byte, span document.Span) string {
	content := commentContent(source, span)
	text := string(source[content.Start:content.End])
	if bytes.HasPrefix(source[span.Start:span.End], []byte("/*")) {
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			lines[i], _ = blockCommentLine(line, 0)
		}
		text = strings.Join(lines, "\n")
	}
	// Indented examples retain the same protection as other comment code.
	if strings.HasPrefix(text, "\t") || strings.HasPrefix(text, "    ") {
		return ""
	}
	return strings.TrimSpace(text)
}

func directiveCandidate(text string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(text)), "unswell-")
}

func collectDirective(doc *document.Document, span document.Span, text string) (bool, error) {
	if !directiveCandidate(text) {
		return false, nil
	}
	if len(doc.Directives) >= 1000 {
		return true, fmt.Errorf("source exceeds 1000 suppression directives")
	}
	doc.Directives = append(doc.Directives, document.Directive{Text: strings.TrimSpace(text), Span: span})
	doc.Excluded = append(doc.Excluded, document.Exclusion{Span: span, Reason: "suppression-directive"})
	return true, nil
}

func htmlDirectives(doc *document.Document, span document.Span) error {
	for pos := span.Start; pos < span.End; {
		raw := doc.Source[pos:span.End]
		if pos == 0 {
			raw = bytes.TrimPrefix(raw, []byte("\ufeff"))
		}
		raw = bytes.TrimLeft(raw, " \t\r\n")
		start := span.End - len(raw)
		// Raw HTML is already excluded by the Markdown grammar. Only comments at
		// this boundary can carry policy; comment-like strings inside it cannot.
		if !bytes.HasPrefix(raw, []byte("<!--")) {
			return nil
		}
		offset := bytes.Index(doc.Source[start+4:span.End], []byte("-->"))
		if offset < 0 {
			if directiveCandidate(string(doc.Source[start+4 : span.End])) {
				return fmt.Errorf("unterminated suppression comment at byte %d", start)
			}
			return nil
		}
		end := start + 4 + offset + 3
		if _, err := collectDirective(doc, document.Span{Start: start, End: end}, string(doc.Source[start+4:end-3])); err != nil {
			return err
		}
		pos = end
	}
	return nil
}
