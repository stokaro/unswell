package evaluation

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

// Attributes are the per-unit facts a corpus artifact already records and an
// evaluation can break down by. Domain, generator family and editing workflow
// are not among them; those strata need annotation the corpus does not carry.
type Attributes struct {
	Words         int
	Role          string
	Language      string
	ProseLanguage string
	Origin        string
}

// Stratum is the full metric set for one subgroup. Counts inside it are the
// sample size; a sparse stratum reports its numbers with that size beside them
// rather than hiding them.
type Stratum struct {
	Dimension string  `json:"dimension"`
	Value     string  `json:"value"`
	Metrics   Metrics `json:"metrics"`
}

// maxStrataPerDimension bounds a breakdown so corpus metadata cannot turn the
// record into one row per unit.
const maxStrataPerDimension = 64

var wordBuckets = []struct {
	label string
	upper int
}{{"1-19", 19}, {"20-49", 49}, {"50-99", 99}, {"100-199", 199}, {"200+", -1}}

func wordBucket(words int) string {
	for _, bucket := range wordBuckets {
		if bucket.upper < 0 || words <= bucket.upper {
			return bucket.label
		}
	}
	return wordBuckets[len(wordBuckets)-1].label
}

// stratumAccumulator counts one stratum and the source groups it drew from.
type stratumAccumulator struct {
	accumulator *accumulator
	groupIDs    map[string]bool
}

func accumulateDimension(dimension string, rows []Observation, attributes map[string]Attributes,
	constant float64,
) (map[string]*stratumAccumulator, error) {
	groups := make(map[string]*stratumAccumulator)
	for _, row := range rows {
		attribute, exists := attributes[row.UnitID]
		if !exists {
			return nil, fmt.Errorf("unit %s has no corpus attributes", row.UnitID)
		}
		value := dimensionValue(dimension, attribute)
		group := groups[value]
		if group == nil {
			if len(groups) >= maxStrataPerDimension {
				return nil, fmt.Errorf("dimension %s exceeds %d strata", dimension, maxStrataPerDimension)
			}
			group = &stratumAccumulator{accumulator: newAccumulator(), groupIDs: make(map[string]bool)}
			groups[value] = group
		}
		group.accumulator.add(row, constant)
		group.groupIDs[row.GroupID] = true
	}
	return groups, nil
}

func labelOrUnknown(value string) string {
	if value == "" {
		return "unknown"
	}
	return value
}

func dimensionValue(dimension string, attributes Attributes) string {
	switch dimension {
	case "words":
		return wordBucket(attributes.Words)
	case "role":
		return labelOrUnknown(attributes.Role)
	case "language":
		return labelOrUnknown(attributes.Language)
	case "prose_language":
		return labelOrUnknown(attributes.ProseLanguage)
	default:
		return labelOrUnknown(attributes.Origin)
	}
}

// Stratify computes the metric set for every value of every dimension. A row
// without attributes is an error: silently dropping it would shrink a stratum
// without saying so.
func Stratify(ctx context.Context, rows []Observation, attributes map[string]Attributes,
	constant float64,
) ([]Stratum, error) {
	if !responseValue(constant) {
		return nil, fmt.Errorf("constant baseline must be finite and within [0,1]")
	}
	strata := []Stratum{}
	for _, dimension := range []string{"words", "role", "language", "prose_language", "origin"} {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		groups, err := accumulateDimension(dimension, rows, attributes, constant)
		if err != nil {
			return nil, err
		}
		for value, group := range groups {
			strata = append(strata, Stratum{Dimension: dimension, Value: value,
				Metrics: group.accumulator.finish(len(group.groupIDs), false)})
		}
	}
	slices.SortFunc(strata, func(a, b Stratum) int {
		if a.Dimension != b.Dimension {
			return strings.Compare(a.Dimension, b.Dimension)
		}
		return strings.Compare(a.Value, b.Value)
	})
	return strata, nil
}
