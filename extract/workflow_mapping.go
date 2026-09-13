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
	if err := r.embeddedBlocks(program, mapped, format, selected); err != nil {
		return err
	}
	if err := r.embeddedControls(program, mapped, selected); err != nil {
		return err
	}
	return r.workflowSyntax(mapped, selected)
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
		r.embeddedExclusion(mapped, document.Span{Start: start, End: pos}, "actions-shell-syntax")
	}
	return nil
}
