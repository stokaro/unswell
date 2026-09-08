package corpus

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"slices"
	"strings"
	"unicode"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
	"github.com/stokaro/unswell/research/annotation"
)

const pipelineVersion = "unswell-corpus-extraction-v1"

// Build verifies all declared source and notice bytes, then extracts unlabeled
// units from the frozen plan. It returns no partial artifact on error.
func Build(ctx context.Context, plan Plan, files map[string][]byte) (Artifact, error) {
	if err := ValidatePlan(ctx, plan); err != nil {
		return Artifact{}, err
	}
	if err := verifyFiles(ctx, plan.Manifest, files); err != nil {
		return Artifact{}, err
	}
	owned, err := MakePlan(ctx, plan.Manifest)
	if err != nil {
		return Artifact{}, err
	}
	plan = owned
	provider, err := english.New()
	if err != nil {
		return Artifact{}, err
	}
	result := Artifact{Version: Version, Status: "unlabeled_candidates", Plan: plan,
		Pipeline: pipeline(provider), Sources: []SourceResult{}, Units: []Candidate{}}
	policyHash, err := digest(plan.Manifest.Policy)
	if err != nil {
		return Artifact{}, err
	}
	if err := collectSources(ctx, &result, files, provider, policyHash); err != nil {
		return Artifact{}, err
	}
	if len(result.Units) == 0 {
		return Artifact{}, fmt.Errorf("no eligible corpus units")
	}
	for i := range result.Units {
		result.Units[i].Unit.ID = fmt.Sprintf("u%06d", i+1)
	}
	result.SHA256, err = digest(result)
	if err != nil {
		return Artifact{}, err
	}
	return result, nil
}

func collectSources(ctx context.Context, result *Artifact, files map[string][]byte, provider *english.Provider, policyHash string) error {
	plan := result.Plan
	groups := sourceGroups(plan)
	budget := MaxManifestBytes
	for _, source := range plan.Manifest.Sources {
		if err := ctx.Err(); err != nil {
			return err
		}
		doc, err := extract.Parse(ctx, document.Source{Name: source.Path, Format: source.Format, Bytes: files[source.Path]},
			extract.Options{MaxBytes: MaxSourceBytes, MaxBlocks: MaxUnits, Policy: plan.Manifest.Policy})
		if err != nil {
			return fmt.Errorf("extract %s: %w", source.ID, err)
		}
		units, err := sourceUnits(ctx, provider, source, doc, plan.Manifest.UnitKinds, groups[source.ID], policyHash)
		if err != nil {
			return fmt.Errorf("units for %s: %w", source.ID, err)
		}
		out := SourceResult{ID: source.ID, Blocks: len(doc.Blocks), Units: len(units), Excluded: doc.Excluded}
		if len(units) == 0 {
			out.EmptyReason = "no_eligible_requested_units"
		}
		if err := account(&budget, out, units); err != nil {
			return err
		}
		result.Units = append(result.Units, units...)
		if len(result.Units) > MaxUnits {
			return fmt.Errorf("candidate count exceeds %d", MaxUnits)
		}
		result.Sources = append(result.Sources, out)
	}
	return nil
}

func account(budget *int, source SourceResult, units []Candidate) error {
	data, err := json.Marshal(source)
	if err != nil {
		return err
	}
	*budget += len(data)
	for _, unit := range units {
		data, err := json.Marshal(unit)
		if err != nil {
			return err
		}
		*budget += len(data) + 32
		if *budget > MaxArtifactBytes/2 {
			return fmt.Errorf("candidate serialization exceeds preparation budget")
		}
	}
	if *budget > MaxArtifactBytes/2 {
		return fmt.Errorf("candidate serialization exceeds preparation budget")
	}
	return nil
}

func verifyFiles(ctx context.Context, manifest Manifest, files map[string][]byte) error {
	required, err := requiredFiles(manifest)
	if err != nil {
		return err
	}
	if len(files) != len(required) {
		return fmt.Errorf("supplied files must exactly match source and notice declarations")
	}
	for _, file := range required {
		if err := ctx.Err(); err != nil {
			return err
		}
		data, found := files[file.Path]
		if !found || len(data) != file.Bytes || hashBytes(data) != file.SHA256 {
			return fmt.Errorf("source or notice %s does not match its recorded bytes and hash", file.Path)
		}
	}
	return nil
}

func sourceGroups(plan Plan) map[string]Group {
	result := make(map[string]Group)
	for _, group := range plan.Groups {
		for _, source := range group.Sources {
			result[source] = group
		}
	}
	return result
}

func sourceUnits(ctx context.Context, provider *english.Provider, source Source, doc document.Document,
	kinds []string, group Group, policyHash string,
) ([]Candidate, error) {
	result := []Candidate{}
	for _, block := range doc.Blocks {
		kind := "fragment"
		if !strings.ContainsRune(block.Text, 0) && slices.Contains([]string{"paragraph", "comment"}, block.Kind) {
			kind = "paragraph"
		}
		for _, piece := range pieces(block.MappedText) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			units, err := pieceUnits(ctx, provider, source, block.Kind, kind, piece, kinds, group, policyHash)
			if err != nil {
				return nil, err
			}
			result = append(result, units...)
			if len(result) > MaxUnits {
				return nil, fmt.Errorf("source exceeds candidate limit")
			}
		}
	}
	return result, nil
}

func pieces(mapped document.MappedText) []document.MappedText {
	result := []document.MappedText{}
	start := 0
	for part := range strings.SplitSeq(mapped.Text, "\x00") {
		left := len(part) - len(strings.TrimLeftFunc(part, unicode.IsSpace))
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, document.MappedText{Text: trimmed, Map: mapped.Map[start+left : start+left+len(trimmed)]})
		}
		start += len(part) + 1
	}
	return result
}

func pieceUnits(ctx context.Context, provider *english.Provider, source Source, blockKind, kind string,
	piece document.MappedText, kinds []string, group Group, policyHash string,
) ([]Candidate, error) {
	sentences, err := analyzePiece(ctx, provider, piece)
	if err != nil {
		return nil, err
	}
	result := []Candidate{}
	words := 0
	for _, sentence := range sentences {
		words += sentence.Words
		if kind != "paragraph" || !slices.Contains(kinds, "sentence") || sentence.Words == 0 {
			continue
		}
		unit, err := candidate(source, group, blockKind, "sentence", sentence.Text, piece.Text, sentence.Spans, sentence.Words, policyHash)
		if err != nil {
			return nil, err
		}
		result = append(result, unit)
	}
	if slices.Contains(kinds, kind) && words > 0 {
		unit, err := candidate(source, group, blockKind, kind, piece.Text, piece.Text, piece.Spans(0, len(piece.Text)), words, policyHash)
		if err != nil {
			return nil, err
		}
		result = append(result, unit)
	}
	return result, nil
}

func analyzePiece(ctx context.Context, provider *english.Provider, piece document.MappedText) ([]document.Sentence, error) {
	if len(piece.Text) > 65536 {
		return nil, fmt.Errorf("eligible context exceeds 65536 bytes")
	}
	return provider.Analyze(ctx, piece, []nlp.Capability{nlp.Tokens, nlp.Sentences})
}

func candidate(source Source, group Group, blockKind, kind, target, surrounding string,
	spans []document.Span, words int, policyHash string,
) (Candidate, error) {
	if len(spans) == 0 || len(spans) > 1024 {
		return Candidate{}, fmt.Errorf("unit requires 1 through 1024 source segments")
	}
	role, err := unitRole(source, blockKind, document.Bounds(spans))
	if err != nil {
		return Candidate{}, err
	}
	origin := source.Origin
	origin.Label = "unknown"
	return Candidate{SourceID: source.ID, GroupID: group.ID, Partition: group.Partition, Words: words,
		Unit: annotation.Unit{Kind: kind, Role: role, Text: target, Context: surrounding,
			Source: annotation.Source{DocumentID: source.Document, RepositoryID: source.Repository,
				RelatedGroup: group.ID, Reference: source.Reference, SHA256: source.SHA256, Bytes: source.Bytes,
				Language: source.Format, ProseLanguage: source.ProseLanguage, Segments: spans},
			Extraction: annotation.Extraction{Identity: pipelineVersion, PolicySHA256: policyHash, ContextPolicy: "eligible-piece-v1"},
			Origin:     origin, Rights: source.Rights}}, nil
}

func unitRole(source Source, blockKind string, span document.Span) (string, error) {
	role := source.Role
	if blockKind == "comment" {
		role = "comment"
	}
	if blockKind == "string" {
		role = "string"
	}
	for _, region := range source.Roles {
		if region.Span.End <= span.Start || region.Span.Start >= span.End {
			continue
		}
		if region.Span.Start > span.Start || region.Span.End < span.End {
			return "", fmt.Errorf("role region cuts through a candidate unit")
		}
		role = region.Role
	}
	return role, nil
}

func pipeline(provider *english.Provider) Pipeline {
	result := Pipeline{Version: pipelineVersion, Revision: "unknown", GoVersion: "unknown",
		Dependencies: []Dependency{}, NLP: provider.Identity()}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return result
	}
	result.GoVersion = info.GoVersion
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			result.Revision = setting.Value
		}
		if setting.Key == "vcs.modified" {
			result.Modified = setting.Value == "true"
		}
	}
	tracked := []string{"github.com/odvcencio/gotreesitter", "github.com/jdkato/prose/v3", "github.com/stokaro/unswell"}
	for _, dependency := range info.Deps {
		if !slices.Contains(tracked, dependency.Path) {
			continue
		}
		version, sum := dependency.Version, dependency.Sum
		if dependency.Replace != nil {
			version, sum = dependency.Replace.Version, dependency.Replace.Sum
		}
		result.Dependencies = append(result.Dependencies, Dependency{Path: dependency.Path, Version: version, Sum: sum})
	}
	slices.SortFunc(result.Dependencies, func(a, b Dependency) int { return strings.Compare(a.Path, b.Path) })
	return result
}
