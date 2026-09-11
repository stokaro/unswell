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
const MaxInputs = 512

// Options selects the baseline cohort of the contrasts, what one count is,
// and an optional first-appearance filter. An empty UnitKind counts
// documents; a unit kind counts every retained unit of that kind inside the
// applicable documents. FirstAppearance requires that kind and keeps, of its
// later cohort, only the listed units.
type Options struct {
	Baseline        string
	UnitKind        string
	FirstAppearance *corpus.AppearanceFilter
	Selection       *corpus.UnitSelection
}

// Selection records the unit selection an analysis applied to every cohort:
// each unit text counted once, at its first occurrence in the stated order.
type Selection struct {
	Kind    string            `json:"kind"`
	Order   []string          `json:"order"`
	Cohorts []SelectionCohort `json:"cohorts"`
}

// SelectionCohort counts one cohort's kept and excluded units.
type SelectionCohort struct {
	Cohort   string `json:"cohort"`
	Kept     int    `json:"kept_units"`
	Excluded int    `json:"excluded_units"`
}

// FirstAppearance records the filter an analysis applied to its later cohort.
// Excluded units repeat earlier text or belong to a repository without an
// earlier snapshot; they enter no table.
type FirstAppearance struct {
	Earlier   string   `json:"earlier"`
	Later     string   `json:"later"`
	Kind      string   `json:"kind"`
	Kept      int      `json:"kept_units"`
	Excluded  int      `json:"excluded_units"`
	LaterOnly []string `json:"later_only_repositories"`
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

// RolePrevalence is one rule on one cohort within one document role. The
// counts follow the row's unit: documents in a document analysis, retained
// units of the chosen kind in a unit analysis. The interval is the same
// cluster bootstrap over the components that hold documents of the role.
type RolePrevalence struct {
	Role               string   `json:"role"`
	Documents          int      `json:"documents"`
	Components         int      `json:"components"`
	Words              int      `json:"words"`
	Findings           int      `json:"findings"`
	Counted            int      `json:"counted"`
	CountedWithFinding int      `json:"counted_with_finding"`
	Prevalence         Estimate `json:"prevalence"`
	PerThousandWords   *float64 `json:"per_thousand_words"`
}

// RuleCohort is one rule measured on one cohort. Counted is what the
// prevalence divides: the applicable documents, or in a unit analysis the
// retained units of the chosen kind inside them. Words and findings cover
// the same set.
type RuleCohort struct {
	Cohort               string           `json:"cohort"`
	Documents            int              `json:"documents"`
	Components           int              `json:"components"`
	Words                int              `json:"words"`
	Findings             int              `json:"findings"`
	DocumentsWithFinding int              `json:"documents_with_finding"`
	Counted              int              `json:"counted"`
	CountedWithFinding   int              `json:"counted_with_finding"`
	Prevalence           Estimate         `json:"prevalence"`
	PerThousandWords     *float64         `json:"per_thousand_words"`
	ZeroUpperBound       *float64         `json:"zero_upper_bound"`
	Units                []KindPrevalence `json:"units"`
	Roles                []RolePrevalence `json:"roles"`
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

// CohortSummary is the whole-profile load of one cohort. Failed documents
// are those whose policy run did not finish; they enter no table. In a unit
// analysis the units, words, and findings are those of the retained units,
// and Excluded counts the units a first-appearance filter removed.
type CohortSummary struct {
	Cohort               string   `json:"cohort"`
	Documents            int      `json:"documents"`
	FailedDocuments      int      `json:"failed_documents"`
	Units                int      `json:"units"`
	Excluded             int      `json:"excluded_units"`
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

// Tables is the complete E1 output. Unit names what one count is: "document"
// or a unit kind. Documents without a cohort are counted under Unassigned and
// enter no table.
type Tables struct {
	Version         string                `json:"version"`
	HumanCorpus     string                `json:"human_corpus"`
	Baseline        string                `json:"baseline"`
	Unit            string                `json:"unit"`
	FirstAppearance *FirstAppearance      `json:"first_appearance,omitempty"`
	Selection       *Selection            `json:"selection,omitempty"`
	Policy          corpus.PolicyIdentity `json:"policy"`
	Classes         string                `json:"rule_classes"`
	Inputs          []Input               `json:"inputs"`
	Cohorts         []CohortSummary       `json:"cohorts"`
	Unassigned      int                   `json:"unassigned_documents"`
	Rules           []RuleTable           `json:"rules"`
}

// Analyze joins finding artifacts measured under one policy and builds the
// tables. The unit of independence is the provenance component; intervals
// resample components jointly across cohorts.
func Analyze(ctx context.Context, inputs []corpus.FindingsArtifact, classes corpus.RuleClasses, options Options) (Tables, error) {
	options, err := checkAnalysis(inputs, classes, options)
	if err != nil {
		return Tables{}, err
	}
	frame, err := buildFrame(ctx, inputs, classes, options)
	if err != nil {
		return Tables{}, err
	}
	tables := Tables{Version: Version, HumanCorpus: "not_qualified", Baseline: options.Baseline, Unit: "document",
		Policy: inputs[0].Policy, Classes: fmt.Sprintf("%s revision %d", classes.Format, classes.Revision),
		Inputs: frame.inputs, Unassigned: frame.unassigned}
	if options.UnitKind != "" {
		tables.Unit = options.UnitKind
	}
	tables.FirstAppearance, err = frame.appearance()
	if err != nil {
		return Tables{}, err
	}
	tables.Selection, err = frame.selection()
	if err != nil {
		return Tables{}, err
	}
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

// checkAnalysis fills the default baseline and refuses inputs, classes, or
// options an analysis cannot use.
func checkAnalysis(inputs []corpus.FindingsArtifact, classes corpus.RuleClasses, options Options) (Options, error) {
	if len(inputs) == 0 || len(inputs) > MaxInputs {
		return options, fmt.Errorf("analysis requires 1 through %d finding artifacts", MaxInputs)
	}
	if options.Baseline == "" {
		options.Baseline = DefaultBaseline
	}
	if err := checkOptions(options); err != nil {
		return options, err
	}
	return options, classes.Cover(inputs[0].Policy.Rules)
}

func checkOptions(options Options) error {
	if options.FirstAppearance != nil && options.Selection != nil {
		return fmt.Errorf("a first-appearance filter and a unit selection cannot apply together")
	}
	if options.Selection != nil {
		return checkSelection(options)
	}
	if options.FirstAppearance != nil {
		return checkFilter(options)
	}
	return nil
}

func checkSelection(options Options) error {
	selection := options.Selection
	if options.UnitKind == "" || selection.Kind != options.UnitKind {
		return fmt.Errorf("a unit selection applies to a unit analysis of its own kind %q", selection.Kind)
	}
	if !strictlySorted(selection.Kept) {
		return fmt.Errorf("the unit selection lists a unit twice")
	}
	return nil
}

func checkFilter(options Options) error {
	filter := options.FirstAppearance
	if options.UnitKind == "" || filter.Kind != options.UnitKind {
		return fmt.Errorf("a first-appearance filter applies to a unit analysis of its own kind %q", filter.Kind)
	}
	if filter.Later == options.Baseline {
		return fmt.Errorf("the first-appearance filter's later cohort cannot be the baseline")
	}
	if !strictlySorted(filter.New) {
		return fmt.Errorf("the first-appearance filter lists a unit twice")
	}
	return nil
}

func strictlySorted(keys []string) bool {
	for i := 1; i < len(keys); i++ {
		if keys[i] <= keys[i-1] {
			return false
		}
	}
	return true
}

// component holds one provenance component's documents per cohort.
type component struct {
	id        string
	documents map[string][]document
}

// document is one measured document with the units the analysis retained.
type document struct {
	role   string
	words  int
	byRule map[string]int
	total  int
	units  []unit
}

type unit struct {
	kind   string
	words  int
	total  int
	byRule map[string]int
}

type frame struct {
	options    Options
	keep       map[string]bool
	filtered   map[string]bool
	matched    int
	inputs     []Input
	components []component
	cohorts    []string
	unassigned int
	failedBy   map[string]int
	unitsBy    map[string]int
	excluded   map[string]int
	byID       map[string]int
	cohortSet  map[string]bool
	known      map[string]bool
}

func buildFrame(ctx context.Context, inputs []corpus.FindingsArtifact, classes corpus.RuleClasses, options Options,
) (*frame, error) {
	result := &frame{options: options, unitsBy: map[string]int{}, failedBy: map[string]int{}, excluded: map[string]int{},
		byID: map[string]int{}, cohortSet: map[string]bool{}, known: map[string]bool{}}
	for _, class := range classes.Rules {
		result.known[class.RuleID] = true
	}
	if options.FirstAppearance != nil {
		result.keep = keySet(options.FirstAppearance.New)
		result.filtered = map[string]bool{options.FirstAppearance.Later: true}
	}
	if options.Selection != nil {
		result.keep = keySet(options.Selection.Kept)
		result.filtered = map[string]bool{}
		for _, cohort := range options.Selection.Cohorts {
			result.filtered[cohort.Cohort] = true
		}
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
		f.unitsBy[item.Cohort]++
		if record, retained := f.retain(item); retained {
			unitsBySource[item.SourceID] = append(unitsBySource[item.SourceID], record)
		}
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
		if doc.Status == "failed" {
			f.failedBy[doc.Cohort]++
			continue
		}
		index := f.componentIndex(doc.GroupID)
		record := document{role: doc.Role, words: doc.ProseWords, byRule: doc.ByRule, total: doc.Findings,
			units: unitsBySource[doc.SourceID]}
		f.components[index].documents[doc.Cohort] = append(f.components[index].documents[doc.Cohort], record)
	}
	return nil
}

func keySet(keys []string) map[string]bool {
	set := make(map[string]bool, len(keys))
	for _, key := range keys {
		set[key] = true
	}
	return set
}

// retain decides whether a unit enters the analysis. A document analysis
// keeps every unit for the per-kind breakdown. A unit analysis keeps the
// chosen kind and, in a filtered cohort, only the listed units.
func (f *frame) retain(item corpus.UnitFindings) (unit, bool) {
	if f.options.UnitKind != "" && item.Kind != f.options.UnitKind {
		return unit{}, false
	}
	if f.keep != nil && f.filtered[item.Cohort] {
		if !f.keep[corpus.UnitKey(item.SourceID, item.UnitID)] {
			f.excluded[item.Cohort]++
			return unit{}, false
		}
		f.matched++
	}
	return unitRecord(item), true
}

// appearance checks that the filter's cohorts are present and that every
// listed unit was found exactly once, so a filter built from other artifacts
// cannot silently pass.
func (f *frame) appearance() (*FirstAppearance, error) {
	filter := f.options.FirstAppearance
	if filter == nil {
		return nil, nil
	}
	if !f.cohortSet[filter.Earlier] || !f.cohortSet[filter.Later] {
		return nil, fmt.Errorf("the first-appearance filter names cohorts absent from the inputs")
	}
	if f.matched != len(f.keep) {
		return nil, fmt.Errorf("the first-appearance filter lists %d units but the inputs matched %d", len(f.keep), f.matched)
	}
	return &FirstAppearance{Earlier: filter.Earlier, Later: filter.Later, Kind: filter.Kind, Kept: f.matched,
		Excluded: f.excluded[filter.Later], LaterOnly: slices.Clone(filter.LaterOnly)}, nil
}

// selection checks that the selection's cohorts are exactly the cohorts of
// the inputs and that every kept unit was found exactly once.
func (f *frame) selection() (*Selection, error) {
	selection := f.options.Selection
	if selection == nil {
		return nil, nil
	}
	if len(f.filtered) != len(f.cohortSet) {
		return nil, fmt.Errorf("the unit selection covers %d cohorts but the inputs hold %d", len(f.filtered), len(f.cohortSet))
	}
	result := &Selection{Kind: selection.Kind, Order: slices.Clone(selection.Order)}
	kept := 0
	for _, cohort := range selection.Cohorts {
		if !f.cohortSet[cohort.Cohort] {
			return nil, fmt.Errorf("the unit selection names cohort %q absent from the inputs", cohort.Cohort)
		}
		result.Cohorts = append(result.Cohorts, SelectionCohort{Cohort: cohort.Cohort, Kept: cohort.Kept,
			Excluded: f.excluded[cohort.Cohort]})
		kept += cohort.Kept
	}
	if f.matched != len(f.keep) || kept != f.matched {
		return nil, fmt.Errorf("the unit selection lists %d units but the inputs matched %d", len(f.keep), f.matched)
	}
	return result, nil
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
	return unit{kind: item.Kind, words: item.Words, total: len(item.Findings), byRule: counts}
}

func (f *frame) summaries() []CohortSummary {
	result := make([]CohortSummary, 0, len(f.cohorts))
	for _, cohort := range f.cohorts {
		summary := CohortSummary{Cohort: cohort, FailedDocuments: f.failedBy[cohort], Excluded: f.excluded[cohort]}
		if f.options.UnitKind == "" {
			summary.Units = f.unitsBy[cohort]
		}
		for _, item := range f.components {
			present := false
			for _, doc := range item.documents[cohort] {
				present = f.load(&summary, doc) || present
			}
			if present {
				summary.Components++
			}
		}
		summary.PerThousandWords = perThousand(summary.Findings, summary.Words)
		result = append(result, summary)
	}
	return result
}

// load adds one measured document to a cohort summary and reports whether it
// held anything to count. A unit analysis counts only retained units.
func (f *frame) load(summary *CohortSummary, doc document) bool {
	if f.options.UnitKind == "" {
		summary.Documents++
		summary.Words += doc.words
		summary.Findings += doc.total
		if doc.total > 0 {
			summary.DocumentsWithFinding++
		}
		return true
	}
	if len(doc.units) == 0 {
		return false
	}
	summary.Documents++
	hit := false
	for _, item := range doc.units {
		summary.Units++
		summary.Words += item.words
		summary.Findings += item.total
		hit = hit || item.total > 0
	}
	if hit {
		summary.DocumentsWithFinding++
	}
	return true
}

func perThousand(findings, words int) *float64 {
	if words == 0 {
		return nil
	}
	value := 1000 * float64(findings) / float64(words)
	return &value
}

// counts holds one rule's tallies of one component: what was counted and
// how much of it carried a finding.
type counts struct {
	counted, withFinding int
}

func (f *frame) ruleTable(class corpus.RuleClass, baseline string, draws [][]int) RuleTable {
	table := RuleTable{RuleID: class.RuleID, Class: class.Class, Roles: class.Roles, Cohorts: []RuleCohort{}, Contrasts: []Contrast{}}
	tallies := make(map[string][]counts, len(f.cohorts))
	for _, cohort := range f.cohorts {
		row, tally, roles := f.cohortRow(class, cohort)
		tallies[cohort] = tally
		for _, item := range tally {
			row.Counted += item.counted
			row.CountedWithFinding += item.withFinding
		}
		row.Prevalence = intervalFor(draws, tally, nil, row.Components)
		if row.Findings == 0 && row.Components > 0 {
			bound := zeroUpperBound(row.Components)
			row.ZeroUpperBound = &bound
		}
		row.Roles = roleRows(roles, draws)
		table.Cohorts = append(table.Cohorts, row)
	}
	table.Contrasts = contrasts(f.cohorts, baseline, tallies, draws, table.Cohorts)
	return table
}

// load is what one applicable document adds to a row or a role stratum.
type load struct {
	documents, words, findings, documentsWithFinding int
}

func (l *load) add(other load) {
	l.documents += other.documents
	l.words += other.words
	l.findings += other.findings
	l.documentsWithFinding += other.documentsWithFinding
}

// roleTally accumulates one role of one cohort per component.
type roleTally struct {
	load    load
	tally   []counts
	present []bool
}

// cohortRow tallies one rule on one cohort per component, in total, and per
// document role.
func (f *frame) cohortRow(class corpus.RuleClass, cohort string) (RuleCohort, []counts, map[string]*roleTally) {
	tally := make([]counts, len(f.components))
	row := RuleCohort{Cohort: cohort, Units: []KindPrevalence{}, Roles: []RolePrevalence{}}
	var total load
	kinds := map[string]*KindPrevalence{}
	roles := map[string]*roleTally{}
	for i, item := range f.components {
		present := false
		for _, doc := range item.documents[cohort] {
			if !slices.Contains(class.Roles, doc.role) {
				continue
			}
			var added load
			var count counts
			if !f.tallyDocument(&added, &count, doc, class.RuleID) {
				continue
			}
			present = true
			total.add(added)
			tally[i].counted += count.counted
			tally[i].withFinding += count.withFinding
			stratum, found := roles[doc.role]
			if !found {
				stratum = &roleTally{tally: make([]counts, len(f.components)), present: make([]bool, len(f.components))}
				roles[doc.role] = stratum
			}
			stratum.load.add(added)
			stratum.tally[i].counted += count.counted
			stratum.tally[i].withFinding += count.withFinding
			stratum.present[i] = true
			tallyUnits(kinds, doc.units, class.RuleID)
		}
		if present {
			row.Components++
		}
	}
	row.Documents, row.Words, row.Findings, row.DocumentsWithFinding = total.documents, total.words, total.findings,
		total.documentsWithFinding
	row.PerThousandWords = perThousand(row.Findings, row.Words)
	row.Units = sortedKinds(kinds)
	return row, tally, roles
}

// roleRows turns the per-role tallies into sorted strata with the same
// interval the cohort row carries.
func roleRows(roles map[string]*roleTally, draws [][]int) []RolePrevalence {
	result := make([]RolePrevalence, 0, len(roles))
	for role, stratum := range roles {
		item := RolePrevalence{Role: role, Documents: stratum.load.documents, Words: stratum.load.words,
			Findings: stratum.load.findings}
		for i, count := range stratum.tally {
			item.Counted += count.counted
			item.CountedWithFinding += count.withFinding
			if stratum.present[i] {
				item.Components++
			}
		}
		item.Prevalence = intervalFor(draws, stratum.tally, nil, item.Components)
		item.PerThousandWords = perThousand(item.Findings, item.Words)
		result = append(result, item)
	}
	slices.SortFunc(result, func(a, b RolePrevalence) int { return strings.Compare(a.Role, b.Role) })
	return result
}

// tallyDocument adds one applicable document to a load and reports whether
// it contributed. A document analysis counts the document; a unit analysis
// counts each retained unit, and a document without one contributes nothing.
func (f *frame) tallyDocument(added *load, tally *counts, doc document, rule string) bool {
	if f.options.UnitKind == "" {
		tally.counted++
		added.documents++
		added.words += doc.words
		added.findings += doc.byRule[rule]
		if doc.byRule[rule] > 0 {
			tally.withFinding++
			added.documentsWithFinding++
		}
		return true
	}
	if len(doc.units) == 0 {
		return false
	}
	added.documents++
	hit := false
	for _, item := range doc.units {
		tally.counted++
		added.words += item.words
		added.findings += item.byRule[rule]
		if item.byRule[rule] > 0 {
			tally.withFinding++
			hit = true
		}
	}
	if hit {
		added.documentsWithFinding++
	}
	return true
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
