package corpus

import (
	"context"
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

// SelectionVersion identifies a unit selection that counts each content
// hash once.
const SelectionVersion = "unswell-unit-selection-v1"

// CohortSelection counts one cohort's units of the selected kind. Kept units
// are the first occurrence of their text. RepeatedWithin units repeat text
// kept earlier in the same cohort; RepeatedEarlier units repeat text kept in
// an earlier cohort.
type CohortSelection struct {
	Cohort          string `json:"cohort"`
	Units           int    `json:"units"`
	Kept            int    `json:"kept"`
	RepeatedWithin  int    `json:"repeated_within"`
	RepeatedEarlier int    `json:"repeated_earlier"`
}

// UnitSelection keeps, across every cohort in a stated order, the first
// occurrence of each unit text of one kind. The protocol counts a unit once
// per content hash across snapshots, copies, and identical document
// versions. An analysis restricted to the kept units does that for every
// cohort at once, whatever repository a copy sits in. Within a cohort the
// first occurrence is the first by repository, source, and unit ID. Keys
// have the form "<source id>#<unit id>" and come sorted.
type UnitSelection struct {
	Version string            `json:"version"`
	Kind    string            `json:"kind"`
	Order   []string          `json:"order"`
	Cohorts []CohortSelection `json:"cohorts"`
	Kept    []string          `json:"kept"`
}

// LoadUnitSelection decodes a strict selection and checks its counts.
func LoadUnitSelection(ctx context.Context, data []byte) (UnitSelection, error) {
	var selection UnitSelection
	limits := jsoninput.Limits{Array: MaxAppearanceUnits, Object: 64}
	if err := jsoninput.Decode(ctx, data, MaxArtifactBytes, &selection, limits); err != nil {
		return UnitSelection{}, err
	}
	if err := selection.validate(); err != nil {
		return UnitSelection{}, err
	}
	return selection, nil
}

func (s UnitSelection) validate() error {
	rank, err := s.validateOrder()
	if err != nil {
		return err
	}
	kept := 0
	for _, cohort := range s.Cohorts {
		if _, known := rank[cohort.Cohort]; !known || cohort.Units != cohort.Kept+cohort.RepeatedWithin+cohort.RepeatedEarlier {
			return fmt.Errorf("unit selection counts do not add up for cohort %q", cohort.Cohort)
		}
		kept += cohort.Kept
	}
	if kept != len(s.Kept) || !strictlySortedKeys(s.Kept) {
		return fmt.Errorf("unit selection keys must match the kept counts, sorted and unique")
	}
	return nil
}

func (s UnitSelection) validateOrder() (map[string]int, error) {
	if s.Version != SelectionVersion || !text(s.Kind) || len(s.Order) == 0 {
		return nil, fmt.Errorf("unsupported unit selection version, kind, or order")
	}
	rank := map[string]int{}
	for i, cohort := range s.Order {
		if _, seen := rank[cohort]; seen || !text(cohort) {
			return nil, fmt.Errorf("unit selection order must list distinct cohorts")
		}
		rank[cohort] = i
	}
	return rank, nil
}

func strictlySortedKeys(keys []string) bool {
	for i := 1; i < len(keys); i++ {
		if keys[i] <= keys[i-1] {
			return false
		}
	}
	return true
}

// Dedupe accumulates candidate artifacts of any cohort in the stated order
// and keeps the first occurrence of each unit text of one kind.
type Dedupe struct {
	kind  string
	order []string
	rank  map[string]int
	seen  map[string]bool
	units []dedupeUnit
}

type dedupeUnit struct {
	rank       int
	cohort     string
	repository string
	sourceID   string
	unitID     string
	hash       [sha256.Size]byte
}

// NewDedupe starts a selection for one unit kind over cohorts in order.
func NewDedupe(kind string, order []string) (*Dedupe, error) {
	if !text(kind) || len(order) == 0 {
		return nil, fmt.Errorf("a unit selection needs a unit kind and a cohort order")
	}
	rank := make(map[string]int, len(order))
	for i, cohort := range order {
		if _, seen := rank[cohort]; seen || !text(cohort) {
			return nil, fmt.Errorf("the cohort order must list distinct cohorts")
		}
		rank[cohort] = i
	}
	return &Dedupe{kind: kind, order: slices.Clone(order), rank: rank, seen: map[string]bool{}}, nil
}

// Add indexes the units of one candidate artifact. Every unit must carry a
// cohort of the order, and no unit key may repeat across the artifacts: the
// selection names units by source and unit ID, which one dataset plan keeps
// unique.
func (d *Dedupe) Add(ctx context.Context, artifact Artifact) error {
	if artifact.Version != Version {
		return fmt.Errorf("unsupported candidate artifact version")
	}
	for _, candidate := range artifact.Units {
		if err := ctx.Err(); err != nil {
			return err
		}
		if candidate.Unit.Kind != d.kind {
			continue
		}
		rank, known := d.rank[candidate.Cohort]
		if !known {
			return fmt.Errorf("unit %s carries cohort %q outside the order", candidate.Unit.ID, candidate.Cohort)
		}
		if len(d.units) >= MaxAppearanceUnits {
			return fmt.Errorf("units exceed %d", MaxAppearanceUnits)
		}
		key := UnitKey(candidate.SourceID, candidate.Unit.ID)
		if d.seen[key] {
			return fmt.Errorf("unit %s appears twice; source IDs must be unique across the artifacts", key)
		}
		d.seen[key] = true
		d.units = append(d.units, dedupeUnit{rank: rank, cohort: candidate.Cohort, repository: candidate.Unit.Source.RepositoryID,
			sourceID: candidate.SourceID, unitID: candidate.Unit.ID, hash: sha256.Sum256([]byte(candidate.Unit.Text))})
	}
	return nil
}

// Selection returns the finished selection. At least one unit is required.
func (d *Dedupe) Selection() (UnitSelection, error) {
	if len(d.units) == 0 {
		return UnitSelection{}, fmt.Errorf("a unit selection needs at least one unit of kind %q", d.kind)
	}
	slices.SortFunc(d.units, compareDedupe)
	firstRank := map[[sha256.Size]byte]int{}
	counts := map[string]*CohortSelection{}
	selection := UnitSelection{Version: SelectionVersion, Kind: d.kind, Order: slices.Clone(d.order), Kept: []string{}}
	for _, unit := range d.units {
		record, found := counts[unit.cohort]
		if !found {
			record = &CohortSelection{Cohort: unit.cohort}
			counts[unit.cohort] = record
		}
		record.Units++
		seen, repeated := firstRank[unit.hash]
		switch {
		case !repeated:
			firstRank[unit.hash] = unit.rank
			record.Kept++
			selection.Kept = append(selection.Kept, UnitKey(unit.sourceID, unit.unitID))
		case seen == unit.rank:
			record.RepeatedWithin++
		default:
			record.RepeatedEarlier++
		}
	}
	for _, cohort := range d.order {
		if record, found := counts[cohort]; found {
			selection.Cohorts = append(selection.Cohorts, *record)
		}
	}
	slices.Sort(selection.Kept)
	return selection, nil
}

func compareDedupe(a, b dedupeUnit) int {
	if a.rank != b.rank {
		return a.rank - b.rank
	}
	if c := strings.Compare(a.repository, b.repository); c != 0 {
		return c
	}
	if c := strings.Compare(a.sourceID, b.sourceID); c != 0 {
		return c
	}
	return strings.Compare(a.unitID, b.unitID)
}
