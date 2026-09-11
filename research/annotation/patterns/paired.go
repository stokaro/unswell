package patterns

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/generation"
)

// PairedVersion identifies the E2 paired tables.
const PairedVersion = "unswell-paired-tables-v1"

// Arm is one operation under one prompt of a run.
type Arm struct {
	Operation string `json:"operation"`
	Prompt    string `json:"prompt"`
}

// ArmCoverage counts what one arm could pair.
type ArmCoverage struct {
	Arm
	Responses           int `json:"responses"`
	Complete            int `json:"complete"`
	Measured            int `json:"measured"`
	UnmeasuredOriginals int `json:"unmeasured_originals"`
	MissingResponses    int `json:"missing_responses"`
	Components          int `json:"components"`
}

// ArmRule is one rule measured on one arm: the responses, their originals,
// and the paired change between them, all over the same tasks.
type ArmRule struct {
	Arm
	Pairs               int      `json:"pairs"`
	Components          int      `json:"components"`
	OriginalWithFinding int      `json:"original_with_finding"`
	ResponseWithFinding int      `json:"response_with_finding"`
	OriginalPrevalence  Estimate `json:"original_prevalence"`
	ResponsePrevalence  Estimate `json:"response_prevalence"`
	PairedChange        Estimate `json:"paired_change"`
	DifferenceFromH0    Estimate `json:"difference_from_h0"`
	ResponseZeroUpper   *float64 `json:"response_zero_upper_bound"`
}

// PairedRule is the E2 row set of one rule.
type PairedRule struct {
	RuleID       string    `json:"rule_id"`
	Class        string    `json:"class"`
	H0Documents  int       `json:"h0_documents"`
	H0WithFindng int       `json:"h0_documents_with_finding"`
	H0Prevalence Estimate  `json:"h0_prevalence"`
	Arms         []ArmRule `json:"arms"`
}

// Paired is the complete E2 output of one run. Every number is a rule
// outcome under one policy; a paired change is the share of tasks whose
// response carries the rule minus the share whose original does.
type Paired struct {
	Version     string                `json:"version"`
	HumanCorpus string                `json:"human_corpus"`
	Run         string                `json:"run"`
	Family      string                `json:"family"`
	Model       string                `json:"model"`
	Role        string                `json:"role"`
	Policy      corpus.PolicyIdentity `json:"policy"`
	Classes     string                `json:"rule_classes"`
	Inputs      []Input               `json:"inputs"`
	H0Cohort    string                `json:"h0_cohort"`
	Arms        []ArmCoverage         `json:"arms"`
	Rules       []PairedRule          `json:"rules"`
}

// pair holds one task's original unit and its response document.
type pair struct {
	component int
	original  map[string]int
	response  map[string]int
}

type pairedFrame struct {
	components []string
	byGroup    map[string]int
	arms       map[Arm][]pair
	coverage   map[Arm]*ArmCoverage
	h0         []map[string]int
	h0Groups   []int
}

// AnalyzePaired joins a run's generation records with the finding artifacts
// of the originals and the responses. Originals are the tasks' units in the
// historical cohort; responses are the controlled documents the records
// name. H0 documents of the tasks' role give the difference estimand.
func AnalyzePaired(ctx context.Context, records generation.Generation, tasks generation.Tasks,
	inputs []corpus.FindingsArtifact, classes corpus.RuleClasses,
) (Paired, error) {
	if err := checkPaired(tasks, inputs, classes); err != nil {
		return Paired{}, err
	}
	role := tasks.Tasks[0].Role
	frame, meta, err := buildPairedFrame(ctx, records, tasks, inputs, role)
	if err != nil {
		return Paired{}, err
	}
	result := Paired{Version: PairedVersion, HumanCorpus: "not_qualified", Run: records.Run, Family: records.Family,
		Model: records.Model, Role: role, Policy: inputs[0].Policy,
		Classes: fmt.Sprintf("%s revision %d", classes.Format, classes.Revision), Inputs: meta, H0Cohort: tasks.Cohort}
	for _, arm := range frame.armOrder() {
		result.Arms = append(result.Arms, *frame.coverage[arm])
	}
	draws, err := drawComponents(ctx, len(frame.components))
	if err != nil {
		return Paired{}, err
	}
	for _, class := range classes.Rules {
		if err := ctx.Err(); err != nil {
			return Paired{}, err
		}
		if slices.Contains(class.Roles, role) {
			result.Rules = append(result.Rules, frame.rule(class, draws))
		}
	}
	return result, ctx.Err()
}

func checkPaired(tasks generation.Tasks, inputs []corpus.FindingsArtifact, classes corpus.RuleClasses) error {
	if len(inputs) == 0 || len(inputs) > MaxInputs || len(tasks.Tasks) == 0 {
		return fmt.Errorf("paired analysis requires tasks and 1 through %d finding artifacts", MaxInputs)
	}
	return classes.Cover(inputs[0].Policy.Rules)
}

func buildPairedFrame(ctx context.Context, records generation.Generation, tasks generation.Tasks,
	inputs []corpus.FindingsArtifact, role string,
) (*pairedFrame, []Input, error) {
	frame := &pairedFrame{byGroup: map[string]int{}, arms: map[Arm][]pair{}, coverage: map[Arm]*ArmCoverage{}}
	byTask := map[string]generation.Task{}
	for _, task := range tasks.Tasks {
		byTask[task.ID] = task
	}
	originals := map[string]map[string]int{}
	responses := map[string]map[string]int{}
	responseGroups := map[string]string{}
	meta := []Input{}
	for _, input := range inputs {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		if err := checkInput(input, inputs[0]); err != nil {
			return nil, nil, err
		}
		meta = append(meta, Input{SHA256: input.SHA256, Documents: len(input.Documents), Units: len(input.Units)})
		frame.collect(input, tasks.Cohort, role, originals, responses, responseGroups)
	}
	for _, record := range records.Records {
		if err := frame.pairRecord(record, byTask, originals, responses); err != nil {
			return nil, nil, err
		}
	}
	for arm, pairs := range frame.arms {
		seen := map[int]bool{}
		for _, item := range pairs {
			seen[item.component] = true
		}
		frame.coverage[arm].Components = len(seen)
	}
	return frame, meta, nil
}

// pairRecord counts one record in its arm's coverage and, when both the
// original unit and the response document were measured, adds the pair.
func (f *pairedFrame) pairRecord(record generation.Record, byTask map[string]generation.Task,
	originals, responses map[string]map[string]int,
) error {
	arm := Arm{Operation: record.Operation, Prompt: record.Prompt}
	coverage := f.armCoverage(arm)
	coverage.Responses++
	if record.Status != "complete" {
		return nil
	}
	coverage.Complete++
	task, found := byTask[record.TaskID]
	if !found {
		return fmt.Errorf("record %s names task %s outside the task set", record.ResponseID, record.TaskID)
	}
	original, measured := originals[task.SourceID+"#"+task.UnitID]
	response, present := responses["generated/"+record.ResponseID+".md"]
	switch {
	case !present:
		coverage.MissingResponses++
	case !measured:
		coverage.UnmeasuredOriginals++
	default:
		coverage.Measured++
		f.arms[arm] = append(f.arms[arm], pair{component: f.component(task.GroupID), original: original, response: response})
	}
	return nil
}

// collect indexes the originals' unit findings, the responses' document
// findings, and the H0 documents of the role.
func (f *pairedFrame) collect(input corpus.FindingsArtifact, h0 string, role string, originals, responses map[string]map[string]int,
	responseGroups map[string]string,
) {
	for _, unit := range input.Units {
		if unit.Cohort == h0 && !unit.Unmeasured {
			originals[unit.SourceID+"#"+unit.UnitID] = ruleCounts(unit)
		}
	}
	for _, doc := range input.Documents {
		if doc.Status != "measured" {
			continue
		}
		switch {
		case doc.Cohort == "controlled" && strings.HasPrefix(doc.Path, "generated/"):
			responses[doc.Path] = doc.ByRule
			responseGroups[doc.Path] = doc.GroupID
		case doc.Cohort == h0 && doc.Role == role:
			f.h0 = append(f.h0, doc.ByRule)
			f.h0Groups = append(f.h0Groups, f.component(doc.GroupID))
		}
	}
}

func ruleCounts(unit corpus.UnitFindings) map[string]int {
	counts := map[string]int{}
	for _, finding := range unit.Findings {
		counts[finding.RuleID]++
	}
	return counts
}

func (f *pairedFrame) component(group string) int {
	index, found := f.byGroup[group]
	if !found {
		index = len(f.components)
		f.byGroup[group] = index
		f.components = append(f.components, group)
	}
	return index
}

func (f *pairedFrame) armCoverage(arm Arm) *ArmCoverage {
	coverage, found := f.coverage[arm]
	if !found {
		coverage = &ArmCoverage{Arm: arm}
		f.coverage[arm] = coverage
	}
	return coverage
}

func (f *pairedFrame) armOrder() []Arm {
	arms := make([]Arm, 0, len(f.coverage))
	for arm := range f.coverage {
		arms = append(arms, arm)
	}
	sort.Slice(arms, func(a, b int) bool {
		if arms[a].Operation != arms[b].Operation {
			return arms[a].Operation < arms[b].Operation
		}
		return arms[a].Prompt < arms[b].Prompt
	})
	return arms
}

func (f *pairedFrame) rule(class corpus.RuleClass, draws [][]int) PairedRule {
	row := PairedRule{RuleID: class.RuleID, Class: class.Class, Arms: []ArmRule{}}
	h0 := make([]counts, len(f.components))
	h0Components := map[int]bool{}
	for i, doc := range f.h0 {
		h0[f.h0Groups[i]].counted++
		h0Components[f.h0Groups[i]] = true
		row.H0Documents++
		if doc[class.RuleID] > 0 {
			h0[f.h0Groups[i]].withFinding++
			row.H0WithFindng++
		}
	}
	row.H0Prevalence = intervalFor(draws, h0, nil, len(h0Components))
	for _, arm := range f.armOrder() {
		row.Arms = append(row.Arms, f.armRule(class.RuleID, arm, draws, h0, len(h0Components)))
	}
	return row
}

func (f *pairedFrame) armRule(rule string, arm Arm, draws [][]int, h0 []counts, h0Components int) ArmRule {
	result := ArmRule{Arm: arm}
	original := make([]counts, len(f.components))
	response := make([]counts, len(f.components))
	seen := map[int]bool{}
	for _, item := range f.arms[arm] {
		seen[item.component] = true
		result.Pairs++
		original[item.component].counted++
		response[item.component].counted++
		if item.original[rule] > 0 {
			original[item.component].withFinding++
			result.OriginalWithFinding++
		}
		if item.response[rule] > 0 {
			response[item.component].withFinding++
			result.ResponseWithFinding++
		}
	}
	result.Components = len(seen)
	result.OriginalPrevalence = intervalFor(draws, original, nil, len(seen))
	result.ResponsePrevalence = intervalFor(draws, response, nil, len(seen))
	result.PairedChange = intervalFor(draws, response, original, len(seen))
	result.DifferenceFromH0 = intervalFor(draws, response, h0, min(len(seen), h0Components))
	if result.Pairs > 0 && result.ResponseWithFinding == 0 && len(seen) > 0 {
		bound := zeroUpperBound(len(seen))
		result.ResponseZeroUpper = &bound
	}
	return result
}
