// Package patterns builds the E1 baseline tables of the LLM-associated pattern
// protocol from label-free finding artifacts. Every number is a rule outcome
// under one pinned policy on one cohort; none is a false-positive rate, a
// recall, or a quality judgment.
package patterns

import (
	"context"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"

	"github.com/stokaro/unswell/research/annotation/corpus"
)

// Version identifies the table contract.
const Version = "unswell-pattern-tables-v1"

// DefaultBaseline is the cohort every other cohort is contrasted with.
const DefaultBaseline = "historical"

// MaxInputs bounds the finding artifacts one analysis joins.
const MaxInputs = 64

// Options selects the baseline cohort of the contrasts.
type Options struct {
	Baseline string
}

// Estimate is a point value with a cluster bootstrap interval over provenance
// components. Status explains a missing interval.
type Estimate struct {
	Value      *float64 `json:"value"`
	Lower      *float64 `json:"lower"`
	Upper      *float64 `json:"upper"`
	Status     string   `json:"status"`
	Replicates int      `json:"valid_replicates"`
}

// KindPrevalence is the share of applicable units of one kind with a finding.
type KindPrevalence struct {
	Kind        string  `json:"kind"`
	Units       int     `json:"units"`
	WithFinding int     `json:"with_finding"`
	Prevalence  float64 `json:"prevalence"`
}

// RuleCohort is one rule measured on one cohort.
type RuleCohort struct {
	Cohort               string           `json:"cohort"`
	Documents            int              `json:"documents"`
	Components           int              `json:"components"`
	Words                int              `json:"words"`
	Findings             int              `json:"findings"`
	DocumentsWithFinding int              `json:"documents_with_finding"`
	Prevalence           Estimate         `json:"prevalence"`
	PerThousandWords     *float64         `json:"per_thousand_words"`
	ZeroUpperBound       *float64         `json:"zero_upper_bound"`
	Units                []KindPrevalence `json:"units"`
}

// Contrast compares one cohort with the baseline for one rule.
type Contrast struct {
	Cohort      string   `json:"cohort"`
	Difference  Estimate `json:"difference"`
	Ratio       *float64 `json:"ratio"`
	RatioStatus string   `json:"ratio_status"`
}

// RuleTable is the E1 row set of one rule.
type RuleTable struct {
	RuleID    string       `json:"rule_id"`
	Class     string       `json:"class"`
	Roles     []string     `json:"roles"`
	Cohorts   []RuleCohort `json:"cohorts"`
	Contrasts []Contrast   `json:"contrasts"`
}

// CohortSummary is the whole-profile load of one cohort.
type CohortSummary struct {
	Cohort               string   `json:"cohort"`
	Documents            int      `json:"documents"`
	Units                int      `json:"units"`
	Words                int      `json:"words"`
	Components           int      `json:"components"`
	Findings             int      `json:"findings"`
	DocumentsWithFinding int      `json:"documents_with_finding"`
	PerThousandWords     *float64 `json:"per_thousand_words"`
}

// Input identifies one finding artifact that entered the tables.
type Input struct {
	SHA256    string `json:"sha256"`
	Documents int    `json:"documents"`
	Units     int    `json:"units"`
}

// Tables is the complete E1 output. Documents without a cohort are counted
// under Unassigned and enter no table.
type Tables struct {
	Version     string                `json:"version"`
	HumanCorpus string                `json:"human_corpus"`
	Baseline    string                `json:"baseline"`
	Policy      corpus.PolicyIdentity `json:"policy"`
	Classes     string                `json:"rule_classes"`
	Inputs      []Input               `json:"inputs"`
	Cohorts     []CohortSummary       `json:"cohorts"`
	Unassigned  int                   `json:"unassigned_documents"`
	Rules       []RuleTable           `json:"rules"`
}

// Analyze joins finding artifacts measured under one policy and builds the
// tables. The unit of independence is the provenance component; intervals
// resample components jointly across cohorts.
func Analyze(ctx context.Context, inputs []corpus.FindingsArtifact, classes corpus.RuleClasses, options Options) (Tables, error) {
	if len(inputs) == 0 || len(inputs) > MaxInputs {
		return Tables{}, fmt.Errorf("analysis requires 1 through %d finding artifacts", MaxInputs)
	}
	if options.Baseline == "" {
		options.Baseline = DefaultBaseline
	}
	if err := classes.Cover(inputs[0].Policy.Rules); err != nil {
		return Tables{}, err
	}
	frame, err := buildFrame(ctx, inputs, classes)
	if err != nil {
		return Tables{}, err
	}
	tables := Tables{Version: Version, HumanCorpus: "not_qualified", Baseline: options.Baseline, Policy: inputs[0].Policy,
		Classes: fmt.Sprintf("%s revision %d", classes.Format, classes.Revision), Inputs: frame.inputs,
		Unassigned: frame.unassigned}
	tables.Cohorts = frame.summaries()
	draws, err := drawComponents(ctx, len(frame.components))
	if err != nil {
		return Tables{}, err
	}
	for _, class := range classes.Rules {
		if err := ctx.Err(); err != nil {
			return Tables{}, err
		}
		tables.Rules = append(tables.Rules, frame.ruleTable(class, options.Baseline, draws))
	}
	return tables, ctx.Err()
}

// component holds one provenance component's documents per cohort.
type component struct {
	id        string
	documents map[string][]document
}

type document struct {
	role   string
	words  int
	byRule map[string]int
	total  int
	units  []unit
}

type unit struct {
	kind   string
	byRule map[string]int
}

type frame struct {
	inputs     []Input
	components []component
	cohorts    []string
	unassigned int
	unitsBy    map[string]int
	byID       map[string]int
	cohortSet  map[string]bool
	known      map[string]bool
}

func buildFrame(ctx context.Context, inputs []corpus.FindingsArtifact, classes corpus.RuleClasses) (*frame, error) {
	result := &frame{unitsBy: map[string]int{}, byID: map[string]int{}, cohortSet: map[string]bool{}, known: map[string]bool{}}
	for _, class := range classes.Rules {
		result.known[class.RuleID] = true
	}
	for _, input := range inputs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := checkInput(input, inputs[0]); err != nil {
			return nil, err
		}
		if err := result.addInput(input); err != nil {
			return nil, err
		}
	}
	for cohort := range result.cohortSet {
		result.cohorts = append(result.cohorts, cohort)
	}
	slices.Sort(result.cohorts)
	sort.Slice(result.components, func(a, b int) bool { return result.components[a].id < result.components[b].id })
	return result, nil
}

func (f *frame) addInput(input corpus.FindingsArtifact) error {
	f.inputs = append(f.inputs, Input{SHA256: input.SHA256, Documents: len(input.Documents), Units: len(input.Units)})
	unitsBySource := map[string][]unit{}
	for _, item := range input.Units {
		unitsBySource[item.SourceID] = append(unitsBySource[item.SourceID], unitRecord(item))
		f.unitsBy[item.Cohort]++
	}
	for _, doc := range input.Documents {
		for rule := range doc.ByRule {
			if !f.known[rule] {
				return fmt.Errorf("finding artifact names rule %q outside the rule classes", rule)
			}
		}
		if doc.Cohort == "" {
			f.unassigned++
			continue
		}
		f.cohortSet[doc.Cohort] = true
		index := f.componentIndex(doc.GroupID)
		record := document{role: doc.Role, words: doc.ProseWords, byRule: doc.ByRule, total: doc.Findings,
			units: unitsBySource[doc.SourceID]}
		f.components[index].documents[doc.Cohort] = append(f.components[index].documents[doc.Cohort], record)
	}
	return nil
}

func (f *frame) componentIndex(group string) int {
	index, found := f.byID[group]
	if !found {
		index = len(f.components)
		f.byID[group] = index
		f.components = append(f.components, component{id: group, documents: map[string][]document{}})
	}
	return index
}

func checkInput(input, first corpus.FindingsArtifact) error {
	if input.Version != corpus.FindingsVersion || input.HumanCorpus != "not_qualified" {
		return fmt.Errorf("finding artifact has an unsupported version or claims a qualified corpus")
	}
	if input.Policy.ConfigHash != first.Policy.ConfigHash || input.Policy.RulesetHash != first.Policy.RulesetHash {
		return fmt.Errorf("finding artifacts were measured under different policies")
	}
	return nil
}

func unitRecord(item corpus.UnitFindings) unit {
	counts := map[string]int{}
	for _, finding := range item.Findings {
		counts[finding.RuleID]++
	}
	return unit{kind: item.Kind, byRule: counts}
}

func (f *frame) summaries() []CohortSummary {
	result := make([]CohortSummary, 0, len(f.cohorts))
	for _, cohort := range f.cohorts {
		summary := CohortSummary{Cohort: cohort, Units: f.unitsBy[cohort]}
		for _, item := range f.components {
			docs := item.documents[cohort]
			if len(docs) == 0 {
				continue
			}
			summary.Components++
			for _, doc := range docs {
				summary.Documents++
				summary.Words += doc.words
				summary.Findings += doc.total
				if doc.total > 0 {
					summary.DocumentsWithFinding++
				}
			}
		}
		summary.PerThousandWords = perThousand(summary.Findings, summary.Words)
		result = append(result, summary)
	}
	return result
}

func perThousand(findings, words int) *float64 {
	if words == 0 {
		return nil
	}
	value := 1000 * float64(findings) / float64(words)
	return &value
}

// counts holds one rule's applicable-document tallies of one component.
type counts struct {
	documents, withFinding int
}

func (f *frame) ruleTable(class corpus.RuleClass, baseline string, draws [][]int) RuleTable {
	table := RuleTable{RuleID: class.RuleID, Class: class.Class, Roles: class.Roles, Cohorts: []RuleCohort{}, Contrasts: []Contrast{}}
	tallies := make(map[string][]counts, len(f.cohorts))
	for _, cohort := range f.cohorts {
		row, tally := f.cohortRow(class, cohort)
		tallies[cohort] = tally
		row.Prevalence = intervalFor(draws, tally, nil, row.Components)
		if row.Findings == 0 && row.Components > 0 {
			bound := zeroUpperBound(row.Components)
			row.ZeroUpperBound = &bound
		}
		table.Cohorts = append(table.Cohorts, row)
	}
	table.Contrasts = contrasts(f.cohorts, baseline, tallies, draws, table.Cohorts)
	return table
}

// cohortRow tallies one rule on one cohort per component and in total.
func (f *frame) cohortRow(class corpus.RuleClass, cohort string) (RuleCohort, []counts) {
	tally := make([]counts, len(f.components))
	row := RuleCohort{Cohort: cohort, Units: []KindPrevalence{}}
	kinds := map[string]*KindPrevalence{}
	for i, item := range f.components {
		present := false
		for _, doc := range item.documents[cohort] {
			if !slices.Contains(class.Roles, doc.role) {
				continue
			}
			present = true
			tally[i].documents++
			row.Documents++
			row.Words += doc.words
			row.Findings += doc.byRule[class.RuleID]
			if doc.byRule[class.RuleID] > 0 {
				tally[i].withFinding++
				row.DocumentsWithFinding++
			}
			tallyUnits(kinds, doc.units, class.RuleID)
		}
		if present {
			row.Components++
		}
	}
	row.PerThousandWords = perThousand(row.Findings, row.Words)
	row.Units = sortedKinds(kinds)
	return row, tally
}

func tallyUnits(kinds map[string]*KindPrevalence, units []unit, rule string) {
	for _, item := range units {
		record, found := kinds[item.kind]
		if !found {
			record = &KindPrevalence{Kind: item.kind}
			kinds[item.kind] = record
		}
		record.Units++
		if item.byRule[rule] > 0 {
			record.WithFinding++
		}
	}
}

func sortedKinds(kinds map[string]*KindPrevalence) []KindPrevalence {
	result := make([]KindPrevalence, 0, len(kinds))
	for _, record := range kinds {
		if record.Units > 0 {
			record.Prevalence = float64(record.WithFinding) / float64(record.Units)
		}
		result = append(result, *record)
	}
	slices.SortFunc(result, func(a, b KindPrevalence) int { return strings.Compare(a.Kind, b.Kind) })
	return result
}

func contrasts(cohorts []string, baseline string, tallies map[string][]counts, draws [][]int, rows []RuleCohort) []Contrast {
	base, found := tallies[baseline]
	result := []Contrast{}
	if !found {
		return result
	}
	byCohort := map[string]RuleCohort{}
	for _, row := range rows {
		byCohort[row.Cohort] = row
	}
	for _, cohort := range cohorts {
		if cohort == baseline {
			continue
		}
		contrast := Contrast{Cohort: cohort, RatioStatus: "undefined"}
		components := min(byCohort[cohort].Components, byCohort[baseline].Components)
		contrast.Difference = intervalFor(draws, tallies[cohort], base, components)
		if byCohort[cohort].Findings > 0 && byCohort[baseline].Findings > 0 {
			ratio := *byCohort[cohort].Prevalence.Value / *byCohort[baseline].Prevalence.Value
			contrast.Ratio, contrast.RatioStatus = &ratio, "defined"
		}
		result = append(result, contrast)
	}
	return result
}

// zeroUpperBound is the one-sided 97.5% Clopper-Pearson bound on the share of
// components that could carry a construction when none was observed in n.
func zeroUpperBound(components int) float64 {
	return 1 - math.Pow(0.025, 1/float64(components))
}
