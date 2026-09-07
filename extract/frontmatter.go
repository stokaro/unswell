package extract

import (
	"bytes"

	"github.com/stokaro/unswell/document"
)

func maskFrontMatter(doc *document.Document) []byte {
	source := bytes.Clone(doc.Source)
	start := 0
	if bytes.HasPrefix(source, []byte{0xef, 0xbb, 0xbf}) {
		start = 3
		copy(source[:3], "   ")
	}
	if !bytes.HasPrefix(source[start:], []byte("---\n")) && !bytes.HasPrefix(source[start:], []byte("---\r\n")) {
		return source
	}
	end := frontMatterEnd(source, start)
	if end == 0 {
		return source
	}
	doc.Excluded = append(doc.Excluded, document.Exclusion{Span: document.Span{Start: start, End: end}, Reason: "front-matter"})
	maskRange(source, start, end)
	return source
}

func frontMatterEnd(source []byte, start int) int {
	pos := start + bytes.IndexByte(source[start:], '\n') + 1
	for pos < len(source) {
		end := bytes.IndexByte(source[pos:], '\n')
		if end < 0 {
			end = len(source) - pos
		}
		line := bytes.TrimSpace(source[pos : pos+end])
		pos = min(len(source), pos+end+1)
		if bytes.Equal(line, []byte("---")) || bytes.Equal(line, []byte("...")) {
			return pos
		}
	}
	return 0
}

func maskRange(source []byte, start, end int) {
	for i := start; i < end; i++ {
		if source[i] != '\n' && source[i] != '\r' {
			source[i] = ' '
		}
	}
}
