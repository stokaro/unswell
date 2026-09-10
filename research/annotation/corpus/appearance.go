package corpus

import (
	"context"
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

// AppearanceVersion identifies the first-appearance filter.
const AppearanceVersion = "unswell-first-appearance-v1"

// MaxAppearanceUnits bounds the unit keys one filter lists.
const MaxAppearanceUnits = 2000000

// RepositoryAppearance counts one repository's units of the filtered kind in
// both cohorts.
type RepositoryAppearance struct {
	Repository string `json:"repository"`
	Earlier    int    `json:"earlier"`
	Later      int    `json:"later"`
	Repeated   int    `json:"repeated"`
	New        int    `json:"new"`
}

// AppearanceFilter names, for a later cohort, every unit of one kind whose
// text is not repeated from an earlier cohort of the same repository. An
// unchanged paragraph that a later snapshot repeats is not a new observation;
// an analysis restricted to the listed units counts text at its first
// evidenced appearance. Units of a repository without an earlier snapshot are
// excluded as well, because none of them can be told from repeated ones. Keys
// are "<source id>#<unit id>", sorted.
type AppearanceFilter struct {
	Version        string                 `json:"version"`
	Kind           string                 `json:"kind"`
	Earlier        string                 `json:"earlier"`
	Later          string                 `json:"later"`
	EarlierUnits   int                    `json:"earlier_units"`
	LaterUnits     int                    `json:"later_units"`
	RepeatedUnits  int                    `json:"repeated_units"`
	LaterOnlyUnits int                    `json:"later_only_units"`
	NewUnits       int                    `json:"new_units"`
	Repositories   []RepositoryAppearance `json:"repositories"`
	LaterOnly      []string               `json:"later_only_repositories"`
	New            []string               `json:"new"`
}

// UnitKey names one unit across shards: source IDs are unique within a
// dataset, unit IDs within a source.
func UnitKey(sourceID, unitID string) string {
	return sourceID + "#" + unitID
}

// LoadAppearanceFilter decodes a strict filter and checks its counts.
func LoadAppearanceFilter(ctx context.Context, data []byte) (AppearanceFilter, error) {
	var filter AppearanceFilter
	limits := jsoninput.Limits{Array: MaxAppearanceUnits, Object: 64}
	if err := jsoninput.Decode(ctx, data, MaxArtifactBytes, &filter, limits); err != nil {
		return AppearanceFilter{}, err
	}
	if err := filter.validate(); err != nil {
		return AppearanceFilter{}, err
	}
	return filter, nil
}

func (f AppearanceFilter) validate() error {
	if f.Version != AppearanceVersion || !text(f.Kind) || !text(f.Earlier) || !text(f.Later) || f.Earlier == f.Later {
		return fmt.Errorf("unsupported first-appearance filter version, kind, or cohorts")
	}
	if f.NewUnits != len(f.New) || f.LaterUnits != f.RepeatedUnits+f.LaterOnlyUnits+f.NewUnits {
		return fmt.Errorf("first-appearance filter counts do not add up")
	}
	for i := 1; i < len(f.New); i++ {
		if f.New[i] <= f.New[i-1] {
			return fmt.Errorf("first-appearance unit keys must be sorted and unique")
		}
	}
	return nil
}

// Appearance accumulates candidate artifacts one at a time, every earlier
// artifact before any later one, so a caller need not hold a whole dataset in
// memory. Text is compared exactly, per repository, for one unit kind.
type Appearance struct {
	kind         string
	earlier      string
	later        string
	seen         map[string]bool
	repositories map[string]*RepositoryAppearance
	laterOnly    map[string]bool
	filter       AppearanceFilter
}

// NewAppearance starts a filter for one unit kind.
func NewAppearance(kind string) (*Appearance, error) {
	if !text(kind) {
		return nil, fmt.Errorf("a first-appearance filter needs a unit kind")
	}
	return &Appearance{kind: kind, seen: map[string]bool{}, repositories: map[string]*RepositoryAppearance{},
		laterOnly: map[string]bool{}, filter: AppearanceFilter{Version: AppearanceVersion, Kind: kind}}, nil
}

// AddEarlier indexes the units of one earlier-cohort artifact.
func (a *Appearance) AddEarlier(ctx context.Context, artifact Artifact) error {
	if a.later != "" {
		return fmt.Errorf("every earlier artifact must precede the later ones")
	}
	cohort, err := a.cohortOf(artifact, a.earlier)
	if err != nil {
		return fmt.Errorf("earlier cohort: %w", err)
	}
	a.earlier = cohort
	for _, candidate := range artifact.Units {
		if err := ctx.Err(); err != nil {
			return err
		}
		if candidate.Unit.Kind != a.kind {
			continue
		}
		a.seen[a.textKey(candidate)] = true
		a.repository(candidate).Earlier++
		a.filter.EarlierUnits++
	}
	return nil
}

// AddLater compares the units of one later-cohort artifact with the index.
func (a *Appearance) AddLater(ctx context.Context, artifact Artifact) error {
	if a.earlier == "" {
		return fmt.Errorf("at least one earlier artifact must precede the later ones")
	}
	cohort, err := a.cohortOf(artifact, a.later)
	if err != nil {
		return fmt.Errorf("later cohort: %w", err)
	}
	if cohort == a.earlier {
		return fmt.Errorf("both artifact sets carry cohort %q", cohort)
	}
	a.later = cohort
	for _, candidate := range artifact.Units {
		if err := ctx.Err(); err != nil {
			return err
		}
		if candidate.Unit.Kind != a.kind {
			continue
		}
		a.addLater(candidate)
	}
	return nil
}

func (a *Appearance) addLater(candidate Candidate) {
	record := a.repository(candidate)
	record.Later++
	a.filter.LaterUnits++
	switch {
	case record.Earlier == 0:
		a.laterOnly[record.Repository] = true
		a.filter.LaterOnlyUnits++
	case a.seen[a.textKey(candidate)]:
		record.Repeated++
		a.filter.RepeatedUnits++
	default:
		record.New++
		a.filter.NewUnits++
		a.filter.New = append(a.filter.New, UnitKey(candidate.SourceID, candidate.Unit.ID))
	}
}

// Filter returns the finished filter. Both cohorts must have been added.
func (a *Appearance) Filter() (AppearanceFilter, error) {
	if a.earlier == "" || a.later == "" {
		return AppearanceFilter{}, fmt.Errorf("a first-appearance filter needs artifacts of both cohorts")
	}
	if len(a.filter.New) > MaxAppearanceUnits {
		return AppearanceFilter{}, fmt.Errorf("new units exceed %d", MaxAppearanceUnits)
	}
	filter := a.filter
	filter.Earlier, filter.Later = a.earlier, a.later
	filter.Repositories = make([]RepositoryAppearance, 0, len(a.repositories))
	for _, record := range a.repositories {
		filter.Repositories = append(filter.Repositories, *record)
	}
	slices.SortFunc(filter.Repositories, func(x, y RepositoryAppearance) int { return strings.Compare(x.Repository, y.Repository) })
	filter.LaterOnly = make([]string, 0, len(a.laterOnly))
	for repository := range a.laterOnly {
		filter.LaterOnly = append(filter.LaterOnly, repository)
	}
	slices.Sort(filter.LaterOnly)
	filter.New = slices.Clone(a.filter.New)
	slices.Sort(filter.New)
	return filter, nil
}

// cohortOf returns the single cohort of an artifact's units, which must match
// the cohort already recorded for the side when there is one.
func (a *Appearance) cohortOf(artifact Artifact, expected string) (string, error) {
	if artifact.Version != Version {
		return "", fmt.Errorf("unsupported candidate artifact version")
	}
	cohort := expected
	for _, candidate := range artifact.Units {
		if candidate.Cohort == "" {
			return "", fmt.Errorf("unit %s carries no cohort", candidate.Unit.ID)
		}
		if cohort == "" {
			cohort = candidate.Cohort
		} else if candidate.Cohort != cohort {
			return "", fmt.Errorf("artifacts mix cohorts %q and %q", cohort, candidate.Cohort)
		}
	}
	if cohort == "" {
		return "", fmt.Errorf("the artifact contains no units")
	}
	return cohort, nil
}

func (a *Appearance) textKey(candidate Candidate) string {
	sum := sha256.Sum256([]byte(candidate.Unit.Text))
	return candidate.Unit.Source.RepositoryID + "\x00" + fmt.Sprintf("%x", sum)
}

func (a *Appearance) repository(candidate Candidate) *RepositoryAppearance {
	name := candidate.Unit.Source.RepositoryID
	record, found := a.repositories[name]
	if !found {
		record = &RepositoryAppearance{Repository: name}
		a.repositories[name] = record
	}
	return record
}
