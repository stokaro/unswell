package corpus

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/research/annotation"
)

// compressionLabels is the class of every unit a bank may seed with, under
// one label source: unit-scoped human or generated claims of a round, or the
// historical and contemporary labels of a cohort decision set.
type compressionLabels struct {
	basis, round, set string
	units             map[string]string
	origins           map[string]annotation.Origin
	policies          []string
	missing           string
}

type compressionBankBuilder struct {
	result   CompressionBank
	targets  map[string]Candidate
	prepared map[string]nlp.PreparedUnit
	labels   compressionLabels
	sources  map[string]Source
	reserved map[string]bool
	bytes    int
}

func newCompressionBankBuilder(artifact Artifact, prepared Prepared, labels compressionLabels,
	options CompressionBankOptions,
) *compressionBankBuilder {
	b := &compressionBankBuilder{
		result: CompressionBank{Version: CompressionBankVersion, HumanCorpus: "not_qualified", Basis: labels.basis,
			RoundSHA256: labels.round, OriginsSHA256: labels.set, Verification: prepared.Verification,
			Options: options, Cohorts: []CompressionReference{}, ReservedGroups: []Group{}, ReservedTargets: []CompressionReservation{}},
		targets: make(map[string]Candidate), prepared: make(map[string]nlp.PreparedUnit), labels: labels,
		sources: make(map[string]Source), reserved: make(map[string]bool), bytes: MaxCompressionSelectionBytes,
	}
	for _, target := range artifact.Units {
		b.targets[target.Unit.ID] = target
	}
	for _, target := range prepared.Targets {
		b.prepared[target.UnitID] = target.Unit
	}
	for _, source := range artifact.Plan.Manifest.Sources {
		b.sources[source.ID] = source
	}
	return b
}

func (b *compressionBankBuilder) addCohort(ctx context.Context, cohort CompressionCohort) error {
	result := CompressionReference{ID: cohort.ID, Origin: cohort.Origin, Units: []CompressionReferenceUnit{}}
	var reference strings.Builder
	labels := make(map[string]bool)
	for _, id := range cohort.UnitIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		entry, text, err := b.referenceUnit(id, cohort.Origin)
		if err != nil {
			return err
		}
		if err := appendReference(&reference, &entry, text); err != nil {
			return err
		}
		if err := b.charge(entry); err != nil {
			return err
		}
		result.Units = append(result.Units, entry)
		labels[entry.Label], b.reserved[entry.GroupID] = true, true
	}
	if cohort.Origin == "mixed" && len(labels) != 2 {
		return fmt.Errorf("mixed reference cohort %s requires both %s and %s targets", cohort.ID,
			b.labels.policies[0], b.labels.policies[1])
	}
	result.Reference = reference.String()
	if err := b.charge(result.Reference); err != nil {
		return err
	}
	compressor, err := feature.NewCompression(ctx, result.Reference, b.result.Options.Compression)
	if err != nil {
		return err
	}
	result.Identity = compressor.Identity()
	b.result.Cohorts = append(b.result.Cohorts, result)
	return nil
}

func (b *compressionBankBuilder) referenceUnit(id, policy string) (CompressionReferenceUnit, string, error) {
	target, exists := b.targets[id]
	if !exists || target.Partition != "training" || target.Unit.Kind != b.result.Options.Kind {
		return CompressionReferenceUnit{}, "", fmt.Errorf("reference %s requires a training target of the selected kind", id)
	}
	label, exists := b.labels.units[id]
	if !exists {
		return CompressionReferenceUnit{}, "", fmt.Errorf("reference %s requires %s", id, b.labels.missing)
	}
	if policy != "mixed" && policy != label {
		return CompressionReferenceUnit{}, "", fmt.Errorf("reference %s does not match cohort origin %s", id, policy)
	}
	if policy == "mixed" && !slices.Contains(b.labels.policies, label) {
		return CompressionReferenceUnit{}, "", fmt.Errorf("reference %s is outside the classes of the bank's label source", id)
	}
	if !slices.Contains(target.Unit.Rights.AllowedUses, "training") {
		return CompressionReferenceUnit{}, "", fmt.Errorf("reference %s lacks a declared training permission", id)
	}
	source := b.sources[target.SourceID]
	unit := b.prepared[id]
	entry := CompressionReferenceUnit{UnitID: id, SourceID: target.SourceID, GroupID: target.GroupID,
		Path: source.Path, Source: target.Unit.Source, Binding: unit.Binding(), Origin: b.labels.origins[id], Label: label,
		Rights: target.Unit.Rights, Notices: source.Notices}
	return entry, unit.Block().Text, nil
}

func appendReference(reference *strings.Builder, entry *CompressionReferenceUnit, text string) error {
	separator := 0
	if reference.Len() > 0 {
		separator = 1
	}
	if len(text) == 0 || reference.Len()+separator+len(text)+1 > 32<<10 {
		return fmt.Errorf("complete reference targets and separators exceed the 32KiB prefix limit")
	}
	if separator != 0 {
		reference.WriteByte('\n')
	}
	entry.Start = reference.Len()
	reference.WriteString(text)
	entry.End = reference.Len()
	return nil
}

func (b *compressionBankBuilder) charge(value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	b.bytes += len(encoded)
	if b.bytes > MaxCompressionBankBytes-1024 {
		return fmt.Errorf("compression bank exceeds metadata budget")
	}
	return nil
}
