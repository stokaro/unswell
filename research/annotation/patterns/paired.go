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
const PairedVersion = "unswell-paired-tables-v3"

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
	AbstainedPairs      int      `json:"abstained_pairs,omitempty"`
	Components          int      `json:"components"`
	OriginalSupport     int      `json:"original_support_components"`
	ResponseSupport     int      `json:"response_support_components"`
	OriginalWithFinding int      `json:"original_with_finding"`
	ResponseWithFinding int      `json:"response_with_finding"`
	OriginalPrevalence  Estimate `json:"original_prevalence"`
	ResponsePrevalence  Estimate `json:"response_prevalence"`
	PairedChange        Estimate `json:"paired_change"`
	PairedPValue        *float64 `json:"paired_p_value"`
	DifferenceFromH0    Estimate `json:"difference_from_h0"`
	ResponseZeroUpper   *float64 `json:"response_zero_upper_bound"`
}

// PairedRule is the E2 row set of one rule.
type PairedRule struct {
	RuleID       string    `json:"rule_id"`
	Class        string    `json:"class"`
	H0Documents  int       `json:"h0_documents"`
	H0WithFindng int       `json:"h0_documents_with_finding"`
	H0Abstained  int       `json:"h0_abstained_documents,omitempty"`
	H0Prevalence Estimate  `json:"h0_prevalence"`
	Arms         []ArmRule `json:"arms"`
}

// Paired is the complete E2 output of one run. Every number is a rule
// outcome under one policy; a paired change is the share of tasks whose
// response carries the rule minus the share whose original does.
type Paired struct {
	Version     string                `json:"version"`
	Dataset     string                `json:"dataset_sha256"`
	HumanCorpus string                `json:"human_corpus"`
	Run         string                `json:"run"`
	Family      string                `json:"family"`
	Model       string                `json:"model"`
	Role        string                `json:"role"`
	Roles       []string              `json:"roles,omitempty"`
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
	original  pairedObservation
	response  pairedObservation
}

type pairedObservation struct {
	counts map[string]int
	absent map[string]string
	role   string
}

type pairedFrame struct {
	components []string
	byGroup    map[string]int
	arms       map[Arm][]pair
	coverage   map[Arm]*ArmCoverage
	h0         []pairedObservation
	h0Groups   []int
	groups     map[string]string
}

// AnalyzePaired joins a run's generation records with the finding artifacts
// of the originals and the responses. Originals are the tasks' units in the
// historical cohort; responses are the controlled documents the records
// name. H0 documents of the tasks' role give the difference estimand.
func AnalyzePaired(ctx context.Context, plan corpus.DatasetPlan, records generation.Generation, tasks generation.Tasks,
	inputs []corpus.FindingsArtifact, classes corpus.RuleClasses,
) (Paired, error) {
	if err := checkPaired(tasks, inputs, classes); err != nil {
		return Paired{}, err
	}
	roles := taskRoles(tasks)
	role := roles[0]
	if len(roles) > 1 {
		role = "mixed"
	}
	frame, meta, err := buildPairedFrame(ctx, plan, records, tasks, inputs, roles)
	if err != nil {
		return Paired{}, err
	}
	result := Paired{Version: PairedVersion, Dataset: plan.DatasetSHA256, HumanCorpus: "not_qualified",
		Run: records.Run, Family: records.Family,
		Model: records.Model, Role: role, Policy: inputs[0].Policy,
		Classes: fmt.Sprintf("%s revision %d", classes.Format, classes.Revision), Inputs: meta, H0Cohort: tasks.Cohort}
	if len(roles) > 1 {
		result.Roles = roles
	}
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
		if slices.ContainsFunc(roles, func(role string) bool { return slices.Contains(class.Roles, role) }) {
			result.Rules = append(result.Rules, frame.rule(class, draws))
		}
	}
	return result, ctx.Err()
}

func checkPaired(tasks generation.Tasks, inputs []corpus.FindingsArtifact, classes corpus.RuleClasses) error {
	if len(inputs) == 0 || len(inputs) > MaxInputs || len(tasks.Tasks) == 0 {
		return fmt.Errorf("paired analysis requires tasks and 1 through %d finding artifacts", MaxInputs)
	}
	seen := map[string]bool{}
	for _, task := range tasks.Tasks {
		if task.ID == "" || seen[task.ID] || task.Role == "" ||
			!slices.Contains([]string{"", generation.DocumentScope}, task.Scope) ||
			(task.Scope == generation.DocumentScope) != (task.UnitID == generation.DocumentScope) {
			return fmt.Errorf("paired tasks require unique IDs, a role, and a valid unit or document scope")
		}
		seen[task.ID] = true
	}
	return classes.Cover(inputs[0].Policy.Rules)
}

func buildPairedFrame(ctx context.Context, plan corpus.DatasetPlan, records generation.Generation, tasks generation.Tasks,
	inputs []corpus.FindingsArtifact, roles []string,
) (*pairedFrame, []Input, error) {
	frame := &pairedFrame{byGroup: map[string]int{}, arms: map[Arm][]pair{}, coverage: map[Arm]*ArmCoverage{}}
	groups, err := pairedGroups(plan, tasks, inputs)
	if err != nil {
		return nil, nil, err
	}
	frame.groups = groups
	byTask := map[string]generation.Task{}
	for _, task := range tasks.Tasks {
		byTask[task.ID] = task
	}
	originals := map[string]pairedObservation{}
	responses := map[string]pairedObservation{}
	meta := []Input{}
	for _, input := range inputs {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		if err := checkInput(input, inputs[0]); err != nil {
			return nil, nil, err
		}
		meta = append(meta, Input{SHA256: input.SHA256, Documents: len(input.Documents), Units: len(input.Units)})
		frame.collect(input, tasks.Cohort, roles, originals, responses)
	}
	for _, record := range records.Records {
		if err := frame.pairRecord(records.Run, record, byTask, originals, responses); err != nil {
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
func (f *pairedFrame) pairRecord(run string, record generation.Record, byTask map[string]generation.Task,
	originals, responses map[string]pairedObservation,
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
	responseID := generation.ControlledID(run, task.Repository, generation.ResponsePath(run, record.ResponseID))
	response, present := responses[responseID]
	switch {
	case !present:
		coverage.MissingResponses++
	case !measured:
		coverage.UnmeasuredOriginals++
	default:
		if f.groups[task.SourceID] != f.groups[responseID] {
			return fmt.Errorf("task %s and response %s belong to different global groups", task.ID, record.ResponseID)
		}
		coverage.Measured++
		original.role, response.role = task.Role, task.Role
		f.arms[arm] = append(f.arms[arm], pair{component: f.component(f.groups[task.SourceID]), original: original, response: response})
	}
	return nil
}

// collect indexes the originals' unit findings, the responses' document
// findings by source ID, and the H0 documents of the role. Two runs that
// answer the same task write the same response path, so the path alone
// cannot name this run's response.
func (f *pairedFrame) collect(input corpus.FindingsArtifact, h0 string, roles []string, originals, responses map[string]pairedObservation) {
	absent := map[string]map[string]string{}
	for _, doc := range input.Documents {
		absent[doc.SourceID] = doc.Abstained
	}
	for _, unit := range input.Units {
		if unit.Cohort == h0 && !unit.Unmeasured {
			originals[unit.SourceID+"#"+unit.UnitID] = pairedObservation{ruleCounts(unit), absent[unit.SourceID], unit.Role}
		}
	}
	f.collectDocuments(input.Documents, h0, roles, originals, responses)
}

func (f *pairedFrame) collectDocuments(docs []corpus.DocumentFindings, h0 string, roles []string,
	originals, responses map[string]pairedObservation,
) {
	for _, doc := range docs {
		if doc.Status != "measured" {
			continue
		}
		if doc.Cohort == h0 {
			originals[doc.SourceID+"#"+generation.DocumentScope] = pairedObservation{doc.ByRule, doc.Abstained, doc.Role}
		}
		switch {
		case doc.Cohort == "controlled" && strings.HasPrefix(doc.Path, "generated/"):
			responses[doc.SourceID] = pairedObservation{doc.ByRule, doc.Abstained, doc.Role}
		case doc.Cohort == h0 && slices.Contains(roles, doc.Role):
			f.h0 = append(f.h0, pairedObservation{doc.ByRule, doc.Abstained, doc.Role})
			f.h0Groups = append(f.h0Groups, f.component(f.groups[doc.SourceID]))
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
		if !slices.Contains(class.Roles, doc.role) {
			continue
		}
		if _, absent := doc.absent[class.RuleID]; absent {
			row.H0Abstained++
			continue
		}
		h0[f.h0Groups[i]].counted++
		h0Components[f.h0Groups[i]] = true
		row.H0Documents++
		if doc.counts[class.RuleID] > 0 {
			h0[f.h0Groups[i]].withFinding++
			row.H0WithFindng++
		}
	}
	row.H0Prevalence = intervalFor(draws, h0, nil, len(h0Components))
	for _, arm := range f.armOrder() {
		row.Arms = append(row.Arms, f.armRule(class, arm, draws, h0, len(h0Components)))
	}
	return row
}

func (f *pairedFrame) armRule(class corpus.RuleClass, arm Arm, draws [][]int, h0 []counts, h0Components int) ArmRule {
	result := ArmRule{Arm: arm}
	rule := class.RuleID
	original := make([]counts, len(f.components))
	response := make([]counts, len(f.components))
	seen := map[int]bool{}
	for _, item := range f.arms[arm] {
		if !slices.Contains(class.Roles, item.original.role) {
			continue
		}
		_, originalAbsent := item.original.absent[rule]
		_, responseAbsent := item.response.absent[rule]
		if originalAbsent || responseAbsent {
			result.AbstainedPairs++
			continue
		}
		seen[item.component] = true
		result.Pairs++
		original[item.component].counted++
		response[item.component].counted++
		if item.original.counts[rule] > 0 {
			original[item.component].withFinding++
			result.OriginalWithFinding++
		}
		if item.response.counts[rule] > 0 {
			response[item.component].withFinding++
			result.ResponseWithFinding++
		}
	}
	result.Components = len(seen)
	result.OriginalSupport, result.ResponseSupport = supportComponents(original), supportComponents(response)
	result.OriginalPrevalence = intervalFor(draws, original, nil, len(seen))
	result.ResponsePrevalence = intervalFor(draws, response, nil, len(seen))
	var replicates []float64
	result.PairedChange, replicates = intervalWithReplicates(draws, response, original, len(seen))
	result.PairedPValue = bootstrapP(replicates)
	result.DifferenceFromH0 = intervalFor(draws, response, h0, min(len(seen), h0Components))
	if result.Pairs > 0 && result.ResponseWithFinding == 0 && len(seen) > 0 {
		bound := zeroUpperBound(len(seen))
		result.ResponseZeroUpper = &bound
	}
	return result
}

func supportComponents(groups []counts) int {
	support := 0
	for _, group := range groups {
		if group.withFinding > 0 {
			support++
		}
	}
	return support
}
