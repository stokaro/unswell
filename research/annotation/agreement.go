package annotation

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

// Estimate is absent with a reason when a statistic has no denominator.
type Estimate struct {
	Value  *float64 `json:"value"`
	Reason string   `json:"reason"`
}

// Count retains a category's support alongside aggregate agreement.
type Count struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// Agreement contains point estimates from independent primary judgments only.
// These are agreement statistics, not model performance or corpus acceptance.
type Agreement struct {
	Version            string       `json:"version"`
	RoundID            string       `json:"round_id"`
	Purpose            string       `json:"purpose"`
	Basis              string       `json:"basis"`
	Rubric             string       `json:"rubric"`
	ProfileSHA256      string       `json:"profile_sha256"`
	PacketSHA256       string       `json:"packet_sha256"`
	PrimaryRaters      int          `json:"primary_raters"`
	MissingAnswers     int          `json:"missing_answers"`
	AuxiliaryJudgments int          `json:"auxiliary_judgments"`
	SamplingIntervals  string       `json:"sampling_intervals"`
	Groups             []GroupStats `json:"groups"`
}

// GroupStats reports each unit kind and input role separately from the overall round.
type GroupStats struct {
	Group      string          `json:"group"`
	Quality    AgreementStats  `json:"quality"`
	Categories []CategoryStats `json:"categories"`
}

// CategoryStats treats unselected reasons as absent only for non-uncertain judgments.
type CategoryStats struct {
	Category          string         `json:"category"`
	UncertainExcluded int            `json:"uncertain_judgments_excluded"`
	Agreement         AgreementStats `json:"agreement"`
}

// AgreementStats retains missingness and support instead of inventing perfect agreement.
type AgreementStats struct {
	Units         int      `json:"units"`
	RatedUnits    int      `json:"rated_units"`
	PairedUnits   int      `json:"paired_units"`
	Ratings       int      `json:"ratings"`
	PairedRatings int      `json:"paired_ratings"`
	Counts        []Count  `json:"counts"`
	RawAgreement  Estimate `json:"raw_pair_agreement"`
	Alpha         Estimate `json:"nominal_alpha"`
	Uncertain     Estimate `json:"uncertain_share"`
}

// Agreement computes statistics without consulting origin or final adjudication.
// All primary raters are assigned every unit in one v1 round; omissions stay missing.
func (r *Round) Agreement(ctx context.Context) (Agreement, error) {
	if r == nil || r.data.packetSHA256 == "" {
		return Agreement{}, fmt.Errorf("load a validated annotation round before measuring agreement")
	}
	result := Agreement{Version: "unswell-annotation-agreement-v1", RoundID: r.data.ID,
		Purpose: r.data.Purpose, Basis: r.data.primaryKind(), Rubric: r.data.Rubric,
		ProfileSHA256: r.data.Profile.SHA256, SamplingIntervals: "not_estimated"}
	result.PacketSHA256 = r.data.packetSHA256
	actors := make(map[string]bool)
	for _, actor := range r.data.Actors {
		if actor.Kind == r.data.primaryKind() && actor.Role == "rater" {
			actors[actor.ID] = true
			result.PrimaryRaters++
		}
	}
	byUnit := make(map[string][]Judgment)
	for _, judgment := range r.data.Judgments {
		if actors[judgment.ActorID] {
			byUnit[judgment.UnitID] = append(byUnit[judgment.UnitID], judgment)
		} else {
			result.AuxiliaryJudgments++
		}
	}
	primaryRatings := len(r.data.Judgments) - result.AuxiliaryJudgments
	result.MissingAnswers = len(r.data.Units)*result.PrimaryRaters - primaryRatings
	groups := r.groupUnits()
	for _, name := range sortedKeys(groups) {
		if err := ctx.Err(); err != nil {
			return Agreement{}, err
		}
		result.Groups = append(result.Groups, summarizeGroup(name, groups[name], byUnit))
	}
	return result, nil
}

func (r *Round) groupUnits() map[string][]string {
	groups := make(map[string][]string)
	for _, unit := range r.data.Units {
		for _, group := range []string{"all", "kind:" + unit.Kind, "role:" + unit.Role} {
			groups[group] = append(groups[group], unit.ID)
		}
	}
	for _, ids := range groups {
		slices.Sort(ids)
	}
	return groups
}

func summarizeGroup(name string, ids []string, byUnit map[string][]Judgment) GroupStats {
	rows := make([][]string, len(ids))
	for i, id := range ids {
		for _, judgment := range byUnit[id] {
			rows[i] = append(rows[i], judgment.Label)
		}
	}
	result := GroupStats{Group: name, Quality: nominal(rows)}
	for _, category := range Categories() {
		result.Categories = append(result.Categories, summarizeCategory(category, ids, byUnit))
	}
	return result
}

func summarizeCategory(category string, ids []string, byUnit map[string][]Judgment) CategoryStats {
	result := CategoryStats{Category: category}
	rows := make([][]string, len(ids))
	for i, id := range ids {
		for _, judgment := range byUnit[id] {
			if judgment.Label == "uncertain" {
				result.UncertainExcluded++
				continue
			}
			label := "absent"
			if slices.Contains(judgment.Categories, category) {
				label = "present"
			}
			rows[i] = append(rows[i], label)
		}
	}
	result.Agreement = nominal(rows)
	return result
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, strings.Compare)
	return keys
}
