package unswell

import (
	"context"
	"fmt"
	"sort"

	"github.com/stokaro/unswell/baseline"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/mapping"
	"github.com/stokaro/unswell/rule"
)

type debtBlock struct {
	structure string
	content   string
}

type debtUnit struct {
	block int
	index int
}

type debtBuilder struct {
	debtHasher
	doc       document.Document
	blocks    []debtBlock
	sentences map[int]debtUnit
}

func newDebtBuilder(ctx context.Context, doc document.Document) (*debtBuilder, error) {
	builder := &debtBuilder{doc: doc, sentences: make(map[int]debtUnit)}
	for _, block := range doc.Blocks {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		content, err := mapping.Canonical(block.MappedText, doc.Source)
		if err != nil {
			return nil, err
		}
		structure := builder.hash(struct {
			Format  document.Format
			Kind    string
			Context []string
		}{doc.Format, block.Kind, block.Context})
		builder.blocks = append(builder.blocks, debtBlock{structure: structure, content: builder.hash(content)})
		for i, sentence := range block.Sentences {
			builder.sentences[sentence.ID] = debtUnit{block: block.ID, index: i}
		}
	}
	return builder, builder.err
}

type debtOccurrence struct {
	Structure, Content, Prefix, Evidence string
}

func (b *debtBuilder) occurrence(occurrence rule.Occurrence) debtOccurrence {
	block := b.doc.Blocks[occurrence.BlockID]
	span := document.Bounds(occurrence.Spans)
	start := sort.Search(len(block.Map), func(i int) bool { return block.Map[i].End > span.Start })
	end := sort.Search(len(block.Map), func(i int) bool { return block.Map[i].Start >= span.End })
	if end <= start {
		b.err = fmt.Errorf("baseline evidence has no mapped prose")
		return debtOccurrence{}
	}
	identity := b.blocks[block.ID]
	return debtOccurrence{Structure: identity.structure, Content: identity.content,
		Prefix: b.canonicalRange(block.MappedText, 0, start), Evidence: b.canonicalRange(block.MappedText, start, end)}
}

func (b *debtBuilder) canonicalRange(mapped document.MappedText, start, end int) string {
	if b.err != nil {
		return ""
	}
	canonical, err := mapping.Canonical(document.MappedText{Text: mapped.Text[start:end], Map: mapped.Map[start:end]}, b.doc.Source)
	if err != nil {
		b.err = err
		return ""
	}
	return b.hash(canonical)
}

func (b *debtBuilder) finding(finding Finding) baseline.Identity {
	occurrences := make([]debtOccurrence, 0, len(finding.Evidence.Occurrences))
	structures, contents := []string{}, []string{}
	for _, occurrence := range finding.Evidence.Occurrences {
		identity := b.occurrence(occurrence)
		occurrences = append(occurrences, identity)
		structures = append(structures, identity.Structure)
		contents = append(contents, identity.Content)
	}
	evidence := b.hash(struct {
		Kind        string
		Activation  int
		Metrics     []rule.Metric
		Occurrences []debtOccurrence
	}{finding.Evidence.Kind, finding.Evidence.Activation, finding.Evidence.Metrics, occurrences})
	return baseline.Identity{Path: b.doc.Name, Kind: "finding", RuleID: finding.RuleID, RuleVersion: finding.RuleVersion,
		StructureHash: b.hash(structures), ContentHash: b.hash(contents), EvidenceHash: evidence}
}

func (b *debtBuilder) unit(scope string, id int) baseline.Identity {
	blockID, index := id, -1
	if scope == "sentence" {
		unit := b.sentences[id]
		blockID, index = unit.block, unit.index
	}
	block := b.blocks[blockID]
	return baseline.Identity{Path: b.doc.Name, Kind: scope, ContentHash: block.content,
		StructureHash: b.hash(struct {
			Block string
			Index int
		}{block.structure, index})}
}

func (b *debtBuilder) fingerprint(identity baseline.Identity) string {
	if b.err != nil {
		return ""
	}
	fingerprint, err := baseline.Fingerprint(identity)
	if err != nil {
		b.err = err
	}
	return fingerprint
}
