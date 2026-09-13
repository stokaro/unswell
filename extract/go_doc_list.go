package extract

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in licenses/go_LICENSE.
// List item boundaries follow go/doc/comment/parse.go at Go 1.25.0,
// commit 6e676ab2b809d46623acb5988248d95d1eb7939c, with source line ranges.

import (
	"fmt"
	"strings"

	"github.com/stokaro/unswell/document"
)

func (r *goDocReader) list(lines []goDocLine) error {
	number, _, _ := goDocListMarker(lines[0].text)
	ordered := number != ""
	id, items := nextEmbeddedList(r.doc.Blocks), goDocListItems(lines, ordered)
	var paragraph []goDocLine
	item := -1
	flush := func() {
		if len(paragraph) > 0 {
			r.paragraph(paragraph, &document.ListContext{ID: id, Item: item, Items: items, Depth: 1, Ordered: ordered})
			paragraph = nil
		}
	}
	for _, line := range lines {
		if err := r.ctx.Err(); err != nil {
			return err
		}
		if len(r.doc.Blocks) > r.options.MaxBlocks {
			return fmt.Errorf("source exceeds %d prose blocks", r.options.MaxBlocks)
		}
		if n, rest, ok := goDocListMarker(line.text); ok && (n != "") == ordered {
			flush()
			item++
			trimmed := strings.TrimSpace(line.text)
			line.start += strings.Index(line.text, trimmed) + len(trimmed) - len(rest)
			line.text = rest
		}
		trimmed := strings.TrimSpace(line.text)
		if trimmed == "" {
			flush()
			continue
		}
		line.start += strings.Index(line.text, trimmed)
		line.text = trimmed
		paragraph = append(paragraph, line)
	}
	flush()
	return nil
}

func goDocListItems(lines []goDocLine, ordered bool) int {
	items := 0
	for _, line := range lines {
		if n, _, ok := goDocListMarker(line.text); ok && (n != "") == ordered {
			items++
		}
	}
	return items
}
