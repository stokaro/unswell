package extract

import (
	"fmt"

	"github.com/stokaro/unswell/document"
)

func (r *sourceReader) workflowProgram(mapped document.MappedText, format document.Format) error {
	options := r.options
	options.MaxBlocks = max(1, options.MaxBlocks-len(r.doc.Blocks))
	program, err := Parse(r.ctx, document.Source{Name: r.doc.Name, Format: format, Bytes: []byte(mapped.Text)}, options)
	if err != nil {
		return fmt.Errorf("GitHub Actions %s run at byte %d: %w", format, document.Bounds(mapped.Map).Start, err)
	}
	if len(r.doc.Blocks)+len(program.Blocks) > r.options.MaxBlocks {
		return fmt.Errorf("source exceeds %d prose blocks", r.options.MaxBlocks)
	}
	selected := make([]bool, len(mapped.Text))
	if err := r.workflowBlocks(program, mapped, format, selected); err != nil {
		return err
	}
	if err := r.workflowControls(program, mapped, selected); err != nil {
		return err
	}
	return r.workflowSyntax(mapped, selected)
}

func (r *sourceReader) workflowBlocks(
	program document.Document, mapped document.MappedText, format document.Format, selected []bool,
) error {
	for _, block := range program.Blocks {
		if err := mapWorkflowBlock(mapped, &block, selected); err != nil {
			return err
		}
		if r.options.IncludeStructure {
			block.Context = append([]string{"embedded:" + string(format) + ":" + block.Kind}, block.Context...)
		}
		block.Kind = "string" // The containing YAML scalar remains the outer context.
		r.doc.Blocks = append(r.doc.Blocks, block)
	}
	return nil
}

func (r *sourceReader) workflowControls(program document.Document, mapped document.MappedText, selected []bool) error {
	for _, excluded := range program.Excluded {
		if err := coverWorkflowSpan(selected, excluded.Span); err != nil {
			return err
		}
		r.workflowExclusion(mapped, excluded.Span, excluded.Reason)
	}
	for _, directive := range program.Directives {
		if err := coverWorkflowSpan(selected, directive.Span); err != nil {
			return err
		}
		directive.Span = document.Bounds(mapped.Spans(directive.Span.Start, directive.Span.End))
		r.doc.Directives = append(r.doc.Directives, directive)
	}
	return nil
}

func mapWorkflowBlock(mapped document.MappedText, block *document.Block, selected []bool) error {
	for i, span := range block.Map {
		if err := coverWorkflowSpan(selected, span); err != nil {
			return err
		}
		block.Map[i] = document.Bounds(mapped.Spans(span.Start, span.End))
	}
	block.Span = document.Bounds(block.Map)
	return nil
}

func coverWorkflowSpan(selected []bool, span document.Span) error {
	if !span.Valid(len(selected)) {
		return fmt.Errorf("embedded shell returned an invalid source span")
	}
	for offset := span.Start; offset < span.End; offset++ {
		selected[offset] = true
	}
	return nil
}

func (r *sourceReader) workflowExclusion(mapped document.MappedText, span document.Span, reason string) {
	for _, original := range mapped.Spans(span.Start, span.End) {
		r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: original, Reason: reason})
	}
}

func (r *sourceReader) workflowSyntax(mapped document.MappedText, selected []bool) error {
	for pos := 0; pos < len(selected); {
		if err := r.ctx.Err(); err != nil {
			return err
		}
		if selected[pos] {
			pos++
			continue
		}
		start := pos
		for pos < len(selected) && !selected[pos] {
			pos++
		}
		r.workflowExclusion(mapped, document.Span{Start: start, End: pos}, "actions-shell-syntax")
	}
	return nil
}
