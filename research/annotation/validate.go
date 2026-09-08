package annotation

import (
	"context"
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/stokaro/unswell/document"
)

func (r *roundData) validate(ctx context.Context) error {
	if fmt.Sprintf("%x", sha256.Sum256([]byte(r.Profile.Instructions))) != r.Profile.SHA256 {
		return fmt.Errorf("profile instructions do not match sha256")
	}
	units := make(map[string]Unit, len(r.Units))
	for _, unit := range r.Units {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, exists := units[unit.ID]; exists {
			return fmt.Errorf("duplicate unit %s", unit.ID)
		}
		if err := validateUnit(unit); err != nil {
			return fmt.Errorf("unit %s: %w", unit.ID, err)
		}
		units[unit.ID] = unit
	}
	packet, err := r.packetView(ctx)
	if err != nil {
		return err
	}
	r.packetSHA256 = packet.SHA256
	actors, err := r.validateActors()
	if err != nil {
		return err
	}
	judgments, err := r.validateJudgments(ctx, units, actors)
	if err != nil {
		return err
	}
	return r.validateAdjudications(ctx, actors, judgments)
}

func validateUnit(unit Unit) error {
	if !slices.Contains(document.Formats(), unit.Source.Language) {
		return fmt.Errorf("unsupported source language %q", unit.Source.Language)
	}
	end := 0
	for _, span := range unit.Source.Segments {
		if !span.Valid(unit.Source.Bytes) || span.Start < end {
			return fmt.Errorf("source segments must be ordered, nonoverlapping byte ranges")
		}
		end = span.End
	}
	if strings.ContainsRune(unit.Text, 0) || strings.ContainsRune(unit.Context, 0) {
		return fmt.Errorf("protected boundaries require separate units or explicit context, not NUL text")
	}
	return validateOrigin(unit.Origin)
}

func validateOrigin(origin Origin) error {
	if origin.Label != "unknown" && origin.Scope != "unit" {
		return fmt.Errorf("a source-level origin claim cannot establish a unit label")
	}
	if origin.Label != "human" && origin.Label != "unknown" && origin.GenerationRecord == "" {
		return fmt.Errorf("generation or editing provenance requires a generation record reference")
	}
	return nil
}

func (r roundData) primaryKind() string {
	if r.Purpose == "tutorial" {
		return "simulation"
	}
	return "human"
}

func (r roundData) validateActors() (map[string]Actor, error) {
	actors := make(map[string]Actor, len(r.Actors))
	primary := 0
	for _, actor := range r.Actors {
		if _, exists := actors[actor.ID]; exists {
			return nil, fmt.Errorf("duplicate actor %s", actor.ID)
		}
		if (r.Purpose == "tutorial" && actor.Kind == "human") || (r.Purpose != "tutorial" && actor.Kind == "simulation") {
			return nil, fmt.Errorf("actor %s: simulated examples and human evidence require separate rounds", actor.ID)
		}
		if actor.Kind == r.primaryKind() && actor.Role == "rater" {
			primary++
		}
		actors[actor.ID] = actor
	}
	if primary < 2 {
		return nil, fmt.Errorf("round requires at least two %s raters", r.primaryKind())
	}
	return actors, nil
}

func validateDecision(label string, categories []string) error {
	if label == "needs_revision" && len(categories) == 0 {
		return fmt.Errorf("needs_revision requires at least one rubric category")
	}
	if label == "acceptable" && len(categories) != 0 {
		return fmt.Errorf("acceptable cannot carry required revision categories")
	}
	return nil
}

func (r roundData) validateJudgments(ctx context.Context, units map[string]Unit, actors map[string]Actor) (
	map[string][]Judgment, error,
) {
	seen := make(map[string]bool)
	byUnit := make(map[string][]Judgment)
	for _, judgment := range r.Judgments {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		actor, exists := actors[judgment.ActorID]
		if !exists || actor.Role != "rater" {
			return nil, fmt.Errorf("judgment has unknown or non-rater actor %s", judgment.ActorID)
		}
		if _, exists := units[judgment.UnitID]; !exists {
			return nil, fmt.Errorf("judgment has unknown unit %s", judgment.UnitID)
		}
		key := judgment.UnitID + ":" + judgment.ActorID
		if seen[key] {
			return nil, fmt.Errorf("duplicate independent judgment %s", key)
		}
		seen[key] = true
		if err := r.validateJudgment(judgment); err != nil {
			return nil, err
		}
		if actor.Kind == r.primaryKind() {
			byUnit[judgment.UnitID] = append(byUnit[judgment.UnitID], judgment)
		}
	}
	return byUnit, nil
}

func (r roundData) validateJudgment(judgment Judgment) error {
	if judgment.PacketSHA256 != r.packetSHA256 {
		return fmt.Errorf("judgment does not match the frozen annotation packet")
	}
	if err := validateDecision(judgment.Label, judgment.Categories); err != nil {
		return err
	}
	if judgment.Context == "insufficient" && judgment.Label != "uncertain" {
		return fmt.Errorf("insufficient context requires uncertain")
	}
	_, err := time.Parse(time.RFC3339, judgment.RecordedAt)
	return err
}

func (r roundData) validateAdjudications(ctx context.Context, actors map[string]Actor, judgments map[string][]Judgment) error {
	seen := make(map[string]bool)
	for _, decision := range r.Adjudications {
		if err := ctx.Err(); err != nil {
			return err
		}
		if seen[decision.UnitID] || len(judgments[decision.UnitID]) < 2 {
			return fmt.Errorf("adjudication %s requires two original judgments and one final record", decision.UnitID)
		}
		seen[decision.UnitID] = true
		if err := validateDecision(decision.Label, decision.Categories); err != nil {
			return err
		}
		if err := r.validateReviewers(decision, actors, judgments[decision.UnitID]); err != nil {
			return err
		}
	}
	return nil
}

func (r roundData) validateReviewers(decision Adjudication, actors map[string]Actor, judgments []Judgment) error {
	if decision.PacketSHA256 != r.packetSHA256 {
		return fmt.Errorf("adjudication does not match the frozen annotation packet")
	}
	for _, id := range decision.Reviewers {
		if actors[id].Kind != r.primaryKind() {
			return fmt.Errorf("adjudication requires known %s reviewers", r.primaryKind())
		}
	}
	recorded, err := time.Parse(time.RFC3339, decision.RecordedAt)
	if err != nil {
		return err
	}
	for _, judgment := range judgments {
		original, err := time.Parse(time.RFC3339, judgment.RecordedAt)
		if err != nil {
			return err
		}
		if recorded.Before(original) {
			return fmt.Errorf("adjudication predates independent judgment")
		}
	}
	return nil
}
