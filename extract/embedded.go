package extract

import (
	"fmt"

	"github.com/stokaro/unswell/document"
)

func (r *sourceReader) embeddedBlocks(
	program document.Document, mapped document.MappedText, format document.Format, selected []bool,
) error {
	listIDs := make(map[int]int)
	nextList := nextEmbeddedList(r.doc.Blocks)
	for _, block := range program.Blocks {
		if err := mapEmbeddedBlock(mapped, &block, selected); err != nil {
			return err
		}
		if r.options.IncludeStructure {
			block.Context = append([]string{"embedded:" + string(format) + ":" + block.Kind}, block.Context...)
		}
		if block.List != nil {
			list := *block.List
			if _, ok := listIDs[list.ID]; !ok {
				listIDs[list.ID] = nextList
				nextList++
			}
			list.ID = listIDs[list.ID]
			block.List = &list
		}
		block.Kind = "string" // The containing source literal remains the outer context.
		r.doc.Blocks = append(r.doc.Blocks, block)
	}
	return nil
}

func (r *sourceReader) embeddedControls(program document.Document, mapped document.MappedText, selected []bool) error {
	for _, excluded := range program.Excluded {
		if err := coverEmbeddedSpan(selected, excluded.Span); err != nil {
			return err
		}
		r.embeddedExclusion(mapped, excluded.Span, excluded.Reason)
	}
	for _, directive := range program.Directives {
		if len(r.doc.Directives) >= 1000 {
			return fmt.Errorf("source exceeds 1000 suppression directives")
		}
		if err := coverEmbeddedSpan(selected, directive.Span); err != nil {
			return err
		}
		directive.Span = document.Bounds(mapped.Spans(directive.Span.Start, directive.Span.End))
		r.doc.Directives = append(r.doc.Directives, directive)
	}
	return nil
}

func mapEmbeddedBlock(mapped document.MappedText, block *document.Block, selected []bool) error {
	for i, span := range block.Map {
		if err := coverEmbeddedSpan(selected, span); err != nil {
			return err
		}
		block.Map[i] = document.Bounds(mapped.Spans(span.Start, span.End))
	}
	block.Span = document.Bounds(block.Map)
	return nil
}

func coverEmbeddedSpan(selected []bool, span document.Span) error {
	if !span.Valid(len(selected)) {
		return fmt.Errorf("embedded grammar returned an invalid source span")
	}
	for offset := span.Start; offset < span.End; offset++ {
		selected[offset] = true
	}
	return nil
}

func (r *sourceReader) embeddedExclusion(mapped document.MappedText, span document.Span, reason string) {
	for _, original := range mapped.Spans(span.Start, span.End) {
		r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: original, Reason: reason})
	}
}

func nextEmbeddedList(blocks []document.Block) int {
	next := 0
	for _, block := range blocks {
		if block.List != nil {
			next = max(next, block.List.ID+1)
		}
	}
	return next
}
