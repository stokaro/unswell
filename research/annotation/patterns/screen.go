package patterns

import (
	"cmp"
	"context"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/generation"
)

// ScreeningVersion identifies the exploratory screening output.
const ScreeningVersion = "unswell-screening-v1"

// Screening defaults: the partition, the unit kind, the role stratum, the
// false discovery rate, and the minimum useful difference the protocol
// fixed for the pilot.
const (
	DefaultPartition = "development"
	DefaultUnitKind  = "paragraph"
	DefaultRole      = "comment"
	DefaultFDR       = 0.10
	DefaultMID       = 0.03
)

// Card thresholds and sample-size inputs fixed by the protocol: components
// with the construction per arm, components per arm, the confirmatory list
// size behind the Holm alpha, the two-sided alpha, and the power.
const (
	supportMinimum   = 5
	clusterMinimum   = 20
	confirmatoryList = 12
	alpha            = 0.05
	power            = 0.80
	screenOperation  = "generate"
	screenPrompt     = "neutral"
)

// ScreenOptions selects the partition, the unit kind, the role stratum, and
// the thresholds of one screening. Empty fields take the defaults.
type ScreenOptions struct {
	Partition string
	UnitKind  string
	Role      string
	FDR       float64
	MID       float64
}

// ScreenRun is one generation run that enters the controlled arm: its
// records and the task set they answer. The records name the family and
// the model.
type ScreenRun struct {
	Records generation.Generation
	Tasks   generation.Tasks
}

// ScreenFamily is one generator family the screening saw, with its models
// and runs. Responses counts the complete generate/neutral responses of the
// role. Documents counts the response documents with a counted unit in the
// partition, and Units counts those units.
type ScreenFamily struct {
	Family    string   `json:"family"`
	Models    []string `json:"models"`
	Runs      []string `json:"runs"`
	Responses int      `json:"responses"`
	Documents int      `json:"documents"`
	Units     int      `json:"units"`
}

// ScreenArm tallies one rule on one arm: units of the kind, units with a
// finding, the components they span, and the components with a finding.
type ScreenArm struct {
	Units       int      `json:"units"`
	WithFinding int      `json:"with_finding"`
	Components  int      `json:"components"`
	Support     int      `json:"support"`
	Prevalence  *float64 `json:"prevalence"`
}

// ScreenModel is the controlled arm of one model inside its family.
type ScreenModel struct {
	Model string `json:"model"`
	ScreenArm
}

// SampleRequirement is the sample one arm needs at one two-sided alpha:
// units before the design effect, components after it.
type SampleRequirement struct {
	Alpha                float64 `json:"alpha"`
	UnitsPerArm          int     `json:"units_per_arm"`
	ControlledComponents int     `json:"controlled_components"`
	H0Components         int     `json:"h0_components"`
}

// SampleSize records what the observed dependence implies for a confirmatory
// sample. The design effect is the bootstrap variance of D over the binomial
// variance of the two unit counts. The requirements give the units and
// components per arm that detect MID over the observed H0 prevalence. Holm
// restates them at alpha over twelve.
type SampleSize struct {
	H0Prevalence                float64           `json:"h0_prevalence"`
	Alternative                 float64           `json:"alternative"`
	BootstrapVariance           float64           `json:"bootstrap_variance"`
	BinomialVariance            float64           `json:"binomial_variance"`
	DesignEffect                float64           `json:"design_effect"`
	ControlledUnitsPerComponent float64           `json:"controlled_units_per_component"`
	H0UnitsPerComponent         float64           `json:"h0_units_per_component"`
	Unadjusted                  SampleRequirement `json:"unadjusted"`
	Holm                        SampleRequirement `json:"holm"`
}

// ScreenTest is one rule on one family: the two arms, the difference D, its
// bootstrap p-value and q-value, and the card state the thresholds give.
// PStatus explains a missing p-value.
type ScreenTest struct {
	Family     string        `json:"family"`
	Controlled ScreenArm     `json:"controlled"`
	Models     []ScreenModel `json:"models"`
	H0         ScreenArm     `json:"h0"`
	Difference Estimate      `json:"difference"`
	PValue     *float64      `json:"p_value"`
	PStatus    string        `json:"p_status"`
	QValue     *float64      `json:"q_value"`
	PassesFDR  bool          `json:"passes_fdr"`
	State      string        `json:"state"`
	Reason     string        `json:"reason"`
	SampleSize *SampleSize   `json:"sample_size"`
}

// ScreenRule is the screening of one rule across the families.
type ScreenRule struct {
	RuleID   string       `json:"rule_id"`
	Class    string       `json:"class"`
	Families []ScreenTest `json:"families"`
}

// ScreenCandidate is a rule and family that passed the false discovery rate
// with a positive difference.
type ScreenCandidate struct {
	RuleID     string  `json:"rule_id"`
	Family     string  `json:"family"`
	Difference float64 `json:"difference"`
	QValue     float64 `json:"q_value"`
}

// Screening is the exploratory screening of one partition. Every number is a
// rule outcome under one policy; a candidate is a rule to write a hypothesis
// template for, not a confirmed construction. BaselineLoad says that the
// supported state does not check the baseline load here.
type Screening struct {
	Version       string                `json:"version"`
	HumanCorpus   string                `json:"human_corpus"`
	Partition     string                `json:"partition"`
	Unit          string                `json:"unit"`
	Role          string                `json:"role"`
	H0Cohort      string                `json:"h0_cohort"`
	Dataset       string                `json:"dataset"`
	DatasetSHA256 string                `json:"dataset_sha256"`
	Families      []ScreenFamily        `json:"families"`
	Replicates    int                   `json:"replicates"`
	Seed          int                   `json:"seed"`
	FDR           float64               `json:"fdr"`
	MID           float64               `json:"mid"`
	BaselineLoad  string                `json:"baseline_load"`
	Policy        corpus.PolicyIdentity `json:"policy"`
	Classes       string                `json:"rule_classes"`
	Inputs        []Input               `json:"inputs"`
	Components    int                   `json:"components"`
	Tests         int                   `json:"tests"`
	Rules         []ScreenRule          `json:"rules"`
	Candidates    []ScreenCandidate     `json:"candidates"`
}

// AnalyzeScreen screens every rule of the role on one partition: the
// controlled units of each family's generate/neutral responses against the
// H0 units of the tasks' cohort. It reports the difference D with its
// cluster interval, a two-sided bootstrap p-value, Benjamini-Hochberg
// q-values across every test, the card state, and a sample-size record. The
// partition of every source comes from the dataset plan, never from the
// artifacts.
func AnalyzeScreen(ctx context.Context, plan corpus.DatasetPlan, runs []ScreenRun, inputs []corpus.FindingsArtifact,
	classes corpus.RuleClasses, options ScreenOptions,
) (Screening, error) {
	options, err := checkScreen(runs, inputs, classes, options)
	if err != nil {
		return Screening{}, err
	}
	frame, meta, err := buildScreenFrame(ctx, plan, runs, inputs, options)
	if err != nil {
		return Screening{}, err
	}
	result := Screening{Version: ScreeningVersion, HumanCorpus: "not_qualified", Partition: options.Partition,
		Unit: options.UnitKind, Role: options.Role, H0Cohort: frame.cohort, Dataset: plan.Dataset.ID,
		DatasetSHA256: plan.DatasetSHA256, Families: frame.familyOrder(), Replicates: replicates, Seed: seed,
		FDR: options.FDR, MID: options.MID, BaselineLoad: "not_checked", Policy: inputs[0].Policy,
		Classes: fmt.Sprintf("%s revision %d", classes.Format, classes.Revision), Inputs: meta,
		Components: len(frame.components), Rules: []ScreenRule{}}
	draws, err := drawComponents(ctx, len(frame.components))
	if err != nil {
		return Screening{}, err
	}
	for _, class := range classes.Rules {
		if err := ctx.Err(); err != nil {
			return Screening{}, err
		}
		if slices.Contains(class.Roles, options.Role) {
			result.Rules = append(result.Rules, frame.rule(class, draws))
		}
	}
	result.Tests = adjustScreening(result.Rules, options.FDR)
	result.Candidates = screenCandidates(result.Rules)
	return result, ctx.Err()
}

// checkScreen fills the defaults and refuses runs, inputs, classes, or
// options a screening cannot use.
func checkScreen(runs []ScreenRun, inputs []corpus.FindingsArtifact, classes corpus.RuleClasses, options ScreenOptions,
) (ScreenOptions, error) {
	if len(runs) == 0 || len(inputs) == 0 || len(inputs) > MaxInputs {
		return options, fmt.Errorf("screening requires generation runs and 1 through %d finding artifacts", MaxInputs)
	}
	if err := checkRuns(runs); err != nil {
		return options, err
	}
	options = screenDefaults(options)
	if options.FDR <= 0 || options.FDR >= 1 || options.MID <= 0 || options.MID >= 1 {
		return options, fmt.Errorf("the false discovery rate and the minimum useful difference lie strictly between 0 and 1")
	}
	return options, classes.Cover(inputs[0].Policy.Rules)
}

func checkRuns(runs []ScreenRun) error {
	for _, run := range runs {
		if len(run.Tasks.Tasks) == 0 || run.Tasks.Cohort != runs[0].Tasks.Cohort {
			return fmt.Errorf("every run needs tasks of one cohort, not %q and %q", runs[0].Tasks.Cohort, run.Tasks.Cohort)
		}
		if run.Records.Run == "" || run.Records.Family == "" || run.Records.Model == "" {
			return fmt.Errorf("every generation record names its run, family, and model")
		}
	}
	return nil
}

func screenDefaults(options ScreenOptions) ScreenOptions {
	options.Partition = cmp.Or(options.Partition, DefaultPartition)
	options.UnitKind = cmp.Or(options.UnitKind, DefaultUnitKind)
	options.Role = cmp.Or(options.Role, DefaultRole)
	options.FDR = cmp.Or(options.FDR, DefaultFDR)
	options.MID = cmp.Or(options.MID, DefaultMID)
	return options
}

// screenUnit is one counted unit: its component, its document, the model
// that wrote it when controlled, and its findings per rule.
type screenUnit struct {
	component int
	document  string
	model     string
	byRule    map[string]int
}

// screenResponse names the family and model behind one response document.
type screenResponse struct {
	family, model string
}

type screenFrame struct {
	options    ScreenOptions
	cohort     string
	partition  map[string]string
	group      map[string]string
	responses  map[string]screenResponse
	families   map[string]*ScreenFamily
	documents  map[string]map[string]bool
	controlled map[string][]screenUnit
	h0         []screenUnit
	components []string
	byGroup    map[string]int
}

func buildScreenFrame(ctx context.Context, plan corpus.DatasetPlan, runs []ScreenRun, inputs []corpus.FindingsArtifact,
	options ScreenOptions,
) (*screenFrame, []Input, error) {
	frame := &screenFrame{options: options, cohort: runs[0].Tasks.Cohort, partition: map[string]string{},
		group: map[string]string{}, responses: map[string]screenResponse{}, families: map[string]*ScreenFamily{},
		documents: map[string]map[string]bool{}, controlled: map[string][]screenUnit{}, byGroup: map[string]int{}}
	for _, source := range plan.Sources {
		frame.partition[source.ID] = source.Partition
		frame.group[source.ID] = source.Group
	}
	for _, run := range runs {
		if err := frame.indexRun(run); err != nil {
			return nil, nil, err
		}
	}
	meta := []Input{}
	for _, input := range inputs {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		if err := checkInput(input, inputs[0]); err != nil {
			return nil, nil, err
		}
		meta = append(meta, Input{SHA256: input.SHA256, Documents: len(input.Documents), Units: len(input.Units)})
		if err := frame.collect(input); err != nil {
			return nil, nil, err
		}
	}
	return frame, meta, nil
}

// indexRun registers the complete generate/neutral responses of a run whose
// task has the role, by the source ID their documents carry, so a unit of
// that source enters the family's arm.
func (f *screenFrame) indexRun(run ScreenRun) error {
	byTask := map[string]generation.Task{}
	for _, task := range run.Tasks.Tasks {
		byTask[task.ID] = task
	}
	family := f.family(run.Records.Family)
	family.Runs = append(family.Runs, run.Records.Run)
	family.Models = append(family.Models, run.Records.Model)
	for _, record := range run.Records.Records {
		if record.Status != "complete" || record.Operation != screenOperation || record.Prompt != screenPrompt {
			continue
		}
		task, found := byTask[record.TaskID]
		if !found {
			return fmt.Errorf("record %s names task %s outside the task set", record.ResponseID, record.TaskID)
		}
		if task.Role != f.options.Role {
			continue
		}
		path := generation.ResponsePath(run.Records.Run, record.ResponseID)
		f.responses[generation.ControlledID(run.Records.Run, task.Repository, path)] = screenResponse{
			family: run.Records.Family, model: run.Records.Model}
		family.Responses++
	}
	slices.Sort(family.Runs)
	slices.Sort(family.Models)
	family.Runs, family.Models = slices.Compact(family.Runs), slices.Compact(family.Models)
	return nil
}

func (f *screenFrame) family(name string) *ScreenFamily {
	family, found := f.families[name]
	if !found {
		family = &ScreenFamily{Family: name, Models: []string{}, Runs: []string{}}
		f.families[name] = family
		f.documents[name] = map[string]bool{}
	}
	return family
}

// collect adds the units of one artifact. A controlled unit of a registered
// response joins its family's arm. A measured unit of the H0 cohort and the
// role joins the H0 arm. Both need a source that the plan puts in the
// screened partition.
func (f *screenFrame) collect(input corpus.FindingsArtifact) error {
	for _, unit := range input.Units {
		response, controlled := f.responses[unit.SourceID]
		if !f.admits(unit, controlled) {
			continue
		}
		selected, err := f.inPartition(unit.SourceID)
		if err != nil {
			return err
		}
		if !selected {
			continue
		}
		record := screenUnit{component: f.component(f.group[unit.SourceID]), document: unit.SourceID, model: response.model,
			byRule: ruleCounts(unit)}
		if controlled {
			f.addControlled(response.family, record)
			continue
		}
		f.h0 = append(f.h0, record)
	}
	return nil
}

// admits keeps measured units of the kind that answer a registered response
// or belong to the H0 cohort in the role.
func (f *screenFrame) admits(unit corpus.UnitFindings, controlled bool) bool {
	if unit.Unmeasured || unit.Kind != f.options.UnitKind {
		return false
	}
	return controlled || (unit.Cohort == f.cohort && unit.Role == f.options.Role)
}

// inPartition reports whether the plan puts a source in the screened
// partition. A source outside the plan is an error, not a silent drop.
func (f *screenFrame) inPartition(source string) (bool, error) {
	partition, found := f.partition[source]
	if !found {
		return false, fmt.Errorf("source %s of a finding artifact is not in the dataset plan", source)
	}
	return partition == f.options.Partition, nil
}

func (f *screenFrame) addControlled(family string, record screenUnit) {
	f.controlled[family] = append(f.controlled[family], record)
	coverage := f.family(family)
	coverage.Units++
	if !f.documents[family][record.document] {
		f.documents[family][record.document] = true
		coverage.Documents++
	}
}

func (f *screenFrame) component(group string) int {
	index, found := f.byGroup[group]
	if !found {
		index = len(f.components)
		f.byGroup[group] = index
		f.components = append(f.components, group)
	}
	return index
}

func (f *screenFrame) familyNames() []string {
	names := make([]string, 0, len(f.families))
	for name := range f.families {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func (f *screenFrame) familyOrder() []ScreenFamily {
	result := make([]ScreenFamily, 0, len(f.families))
	for _, name := range f.familyNames() {
		result = append(result, *f.families[name])
	}
	return result
}

func (f *screenFrame) rule(class corpus.RuleClass, draws [][]int) ScreenRule {
	row := ScreenRule{RuleID: class.RuleID, Class: class.Class, Families: []ScreenTest{}}
	h0, h0Arm := f.tally(f.h0, class.RuleID)
	for _, family := range f.familyNames() {
		row.Families = append(row.Families, f.test(class.RuleID, family, draws, h0, h0Arm))
	}
	return row
}

// tally counts one rule over units per component and summarizes the arm.
func (f *screenFrame) tally(units []screenUnit, rule string) ([]counts, ScreenArm) {
	tallies := make([]counts, len(f.components))
	arm := ScreenArm{}
	for _, unit := range units {
		tallies[unit.component].counted++
		arm.Units++
		if unit.byRule[rule] > 0 {
			tallies[unit.component].withFinding++
			arm.WithFinding++
		}
	}
	for _, tally := range tallies {
		if tally.counted > 0 {
			arm.Components++
		}
		if tally.withFinding > 0 {
			arm.Support++
		}
	}
	if arm.Units > 0 {
		prevalence := float64(arm.WithFinding) / float64(arm.Units)
		arm.Prevalence = &prevalence
	}
	return tallies, arm
}

func (f *screenFrame) test(rule, family string, draws [][]int, h0 []counts, h0Arm ScreenArm) ScreenTest {
	units := f.controlled[family]
	controlled, arm := f.tally(units, rule)
	result := ScreenTest{Family: family, Controlled: arm, Models: f.models(family, units, rule), H0: h0Arm}
	var values []float64
	result.Difference, values = intervalWithReplicates(draws, controlled, h0, min(arm.Components, h0Arm.Components))
	result.PValue = bootstrapP(values)
	result.PStatus = result.Difference.Status
	if result.PValue != nil {
		result.PStatus = "bootstrap_two_sided"
	}
	result.State, result.Reason = screenState(result, f.options.MID)
	result.SampleSize = sampleSize(arm, h0Arm, values, f.options.MID)
	return result
}

// models tallies the family's units per model, in the family's model order.
func (f *screenFrame) models(family string, units []screenUnit, rule string) []ScreenModel {
	byModel := map[string][]screenUnit{}
	for _, unit := range units {
		byModel[unit.model] = append(byModel[unit.model], unit)
	}
	result := []ScreenModel{}
	for _, model := range f.families[family].Models {
		_, arm := f.tally(byModel[model], rule)
		result = append(result, ScreenModel{Model: model, ScreenArm: arm})
	}
	return result
}

// bootstrapP is the two-sided bootstrap p-value of a difference: twice the
// smaller share of replicates on either side of zero, never below one
// replicate's share and never above one. Nil when there is no replicate
// distribution.
func bootstrapP(values []float64) *float64 {
	if len(values) < minimumCount {
		return nil
	}
	below, above := 0, 0
	for _, value := range values {
		if value <= 0 {
			below++
		}
		if value >= 0 {
			above++
		}
	}
	share := float64(len(values))
	p := min(1, max(2*float64(min(below, above))/share, 1/share))
	return &p
}

// screenState applies the protocol's card rules to one test. The baseline
// load is not checked here: supported within scope rests on the interval,
// the point estimate, the support, and the cluster minimum alone.
func screenState(test ScreenTest, mid float64) (string, string) {
	if reason := inconclusiveReason(test); reason != "" {
		return "inconclusive", reason
	}
	lower, value, upper := *test.Difference.Lower, *test.Difference.Value, *test.Difference.Upper
	switch {
	case upper < mid:
		return "unsupported", "upper_below_mid"
	case lower <= 0 || value < mid:
		return "inconclusive", "interval_spans_zero_or_mid"
	case min(test.Controlled.Support, test.H0.Support) < supportMinimum:
		return "inconclusive", "support_below_minimum"
	}
	return "supported-within-scope", "interval_above_zero_and_point_at_mid"
}

// inconclusiveReason names what stops a card before its interval is read: a
// missing interval, a zero count, or an arm below the cluster minimum.
func inconclusiveReason(test ScreenTest) string {
	switch {
	case test.Difference.Status != "cluster_percentile":
		return test.Difference.Status
	case min(test.Controlled.WithFinding, test.H0.WithFinding) == 0:
		return "zero_count"
	case min(test.Controlled.Components, test.H0.Components) < clusterMinimum:
		return "cluster_minimum"
	}
	return ""
}

// adjustScreening applies the Benjamini-Hochberg procedure across every test
// with a p-value and returns how many tests entered it.
func adjustScreening(rules []ScreenRule, fdr float64) int {
	var p []float64
	var tests []*ScreenTest
	for r := range rules {
		for t := range rules[r].Families {
			test := &rules[r].Families[t]
			if test.PValue != nil {
				p = append(p, *test.PValue)
				tests = append(tests, test)
			}
		}
	}
	for i, q := range BenjaminiHochberg(p) {
		value := q
		tests[i].QValue = &value
		tests[i].PassesFDR = value <= fdr
	}
	return len(tests)
}

// BenjaminiHochberg returns the adjusted q-value of every p-value, in the
// given order: the smallest step-up bound m*p/rank over the ranks at or above
// each one, capped at one.
func BenjaminiHochberg(p []float64) []float64 {
	m := len(p)
	order := make([]int, m)
	for i := range order {
		order[i] = i
	}
	slices.SortStableFunc(order, func(a, b int) int { return cmp.Compare(p[a], p[b]) })
	q := make([]float64, m)
	bound := 1.0
	for rank := m; rank >= 1; rank-- {
		index := order[rank-1]
		bound = min(bound, float64(m)*p[index]/float64(rank))
		q[index] = bound
	}
	return q
}

// screenCandidates lists the tests that pass the false discovery rate with a
// positive difference, sorted by q-value, then rule and family.
func screenCandidates(rules []ScreenRule) []ScreenCandidate {
	result := []ScreenCandidate{}
	for _, rule := range rules {
		for _, test := range rule.Families {
			if test.PassesFDR && *test.Difference.Value > 0 {
				result = append(result, ScreenCandidate{RuleID: rule.RuleID, Family: test.Family,
					Difference: *test.Difference.Value, QValue: *test.QValue})
			}
		}
	}
	slices.SortFunc(result, func(a, b ScreenCandidate) int {
		return cmp.Or(cmp.Compare(a.QValue, b.QValue), strings.Compare(a.RuleID, b.RuleID), strings.Compare(a.Family, b.Family))
	})
	return result
}

// sampleSize turns the observed H0 prevalence and the bootstrap dependence
// into the units and components per arm a confirmatory test would need. Nil
// when either arm counted nothing.
func sampleSize(controlled, h0 ScreenArm, values []float64, mid float64) *SampleSize {
	if controlled.Units == 0 || h0.Units == 0 {
		return nil
	}
	observed, p0 := *controlled.Prevalence, *h0.Prevalence
	record := &SampleSize{H0Prevalence: p0, Alternative: min(p0+mid, 1), BootstrapVariance: variance(values),
		BinomialVariance:            observed*(1-observed)/float64(controlled.Units) + p0*(1-p0)/float64(h0.Units),
		DesignEffect:                1,
		ControlledUnitsPerComponent: float64(controlled.Units) / float64(controlled.Components),
		H0UnitsPerComponent:         float64(h0.Units) / float64(h0.Components)}
	if record.BootstrapVariance > 0 && record.BinomialVariance > 0 {
		record.DesignEffect = record.BootstrapVariance / record.BinomialVariance
	}
	record.Unadjusted = requirement(alpha, record, mid)
	record.Holm = requirement(alpha/confirmatoryList, record, mid)
	return record
}

func requirement(alpha float64, record *SampleSize, mid float64) SampleRequirement {
	units := RequiredUnits(record.H0Prevalence, mid, alpha)
	inflated := units * record.DesignEffect
	return SampleRequirement{Alpha: alpha, UnitsPerArm: int(math.Ceil(units)),
		ControlledComponents: int(math.Ceil(inflated / record.ControlledUnitsPerComponent)),
		H0Components:         int(math.Ceil(inflated / record.H0UnitsPerComponent))}
}

// RequiredUnits is the units per arm that detect a difference of mid over the
// prevalence p0 between two proportions at the two-sided alpha with power
// 0.80: (z(1-alpha/2) + z(0.8))^2 (p0(1-p0) + p1(1-p1)) / mid^2 with p1 =
// p0 + mid.
func RequiredUnits(p0, mid, alpha float64) float64 {
	p1 := min(p0+mid, 1)
	z := normalQuantile(1-alpha/2) + normalQuantile(power)
	return z * z * (p0*(1-p0) + p1*(1-p1)) / (mid * mid)
}

// normalQuantile inverts the standard normal distribution.
func normalQuantile(probability float64) float64 { return math.Sqrt2 * math.Erfinv(2*probability-1) }

// variance is the sample variance of the replicate values, zero below two.
func variance(values []float64) float64 {
	if len(values) < minimumCount {
		return 0
	}
	mean := 0.0
	for _, value := range values {
		mean += value
	}
	mean /= float64(len(values))
	sum := 0.0
	for _, value := range values {
		sum += (value - mean) * (value - mean)
	}
	return sum / float64(len(values)-1)
}
