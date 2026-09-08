package unswell

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/suppress"
)

func (e *Engine) suppressionPlan(ctx context.Context, doc document.Document) (*suppress.Plan, error) {
	ids := make(map[string]bool, len(e.descriptors))
	for _, descriptor := range e.descriptors {
		ids[descriptor.ID] = true
	}
	policy := e.policy.Suppressions
	return suppress.Build(ctx, doc, ids, suppress.Options{
		RequireReason: policy.RequireReason, AllowFileWide: policy.AllowFileWide,
		RejectUnused: policy.RejectUnused, MaxCandidates: e.policy.Analysis.MaxCandidates,
	})
}

func (e *Engine) applySuppressions(
	ctx context.Context, result *RunResult, doc document.Document, plan *suppress.Plan, builder *debtBuilder,
) error {
	ids := make([]string, len(plan.Entries))
	for i, entry := range plan.Entries {
		identity, err := json.Marshal(struct {
			Path string
			Span document.Span
			Text string
		}{doc.Name, entry.Span, string(doc.Source[entry.Span.Start:entry.Span.End])})
		if err != nil {
			return err
		}
		ids[i] = fmt.Sprintf("%x", sha256.Sum256(identity))
	}
	err := matchSuppressions(ctx, result.Findings, plan, ids)
	for i, entry := range plan.Entries {
		result.Suppressions = append(result.Suppressions, e.suppressionRecord(doc, entry, ids[i]))
	}
	return errors.Join(err, e.filterTrustedSuppressions(ctx, result, doc, builder))
}

func matchSuppressions(ctx context.Context, findings []Finding, plan *suppress.Plan, ids []string) error {
	for i := range findings {
		finding := &findings[i]
		var spans []document.Span
		for _, occurrence := range finding.Evidence.Occurrences {
			spans = append(spans, occurrence.Spans...)
		}
		matched, err := plan.Match(ctx, finding.RuleID, finding.ID, spans)
		if err != nil {
			return err
		}
		finding.Suppressed = len(matched) != 0
		for _, index := range matched {
			finding.SuppressionIDs = append(finding.SuppressionIDs, ids[index])
		}
	}
	return plan.ValidateUse()
}

func (e *Engine) suppressionRecord(doc document.Document, entry suppress.Entry, id string) Suppression {
	record := Suppression{ID: id, Kind: entry.Kind, RuleIDs: slices.Clone(entry.Rules), Reason: entry.Reason,
		Directive: e.suppressionLocation(doc, entry.Span), Targets: []SuppressionTarget{},
		FindingIDs: append([]string{}, entry.Findings...), UsedRules: append([]string{}, entry.UsedRules...), Status: "used"}
	if len(entry.UsedRules) == 0 {
		record.Status = "unused"
	} else if len(entry.UsedRules) != len(entry.Rules) {
		record.Status = "partially_used"
	}
	if entry.End != nil {
		location := e.suppressionLocation(doc, *entry.End)
		record.End = &location
	}
	for _, target := range entry.Targets {
		record.Targets = append(record.Targets, SuppressionTarget{Scope: target.Scope, UnitID: target.ID, Span: target.Span})
	}
	return record
}

func (e *Engine) suppressionLocation(doc document.Document, span document.Span) Location {
	location := Location{Path: doc.Name, Span: span, Segments: []document.Span{span}}
	// The resolver validates directive ranges against the extracted document.
	if position, err := document.Locate(doc.Source, span.Start); err == nil {
		location.Start = position
	}
	if position, err := document.Locate(doc.Source, span.End); err == nil {
		location.End = position
	}
	if e.includeSource {
		location.Snippet = string(doc.Source[span.Start:span.End])
	}
	return location
}
