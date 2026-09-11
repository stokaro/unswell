package patterns

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"unicode"

	srcdoc "github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

// FrequencyVersion identifies the frequency tables. The tables propose
// candidate constructions from the corpus; they decide nothing.
const FrequencyVersion = "unswell-frequency-tables-v1"

// Frequency measures, counted per sentence unit.
const (
	MeasureWord1     = "word-1gram"
	MeasureWord2     = "word-2gram"
	MeasureWord3     = "word-3gram"
	MeasureOpener    = "opener-3"
	MeasureTemplate  = "pos-template"
	maxTemplateTags  = 12
	maxStrata        = 64
	frequencyUnit    = "sentence"
	defaultMinCount  = 5
	defaultTopItems  = 200
	defaultSupport   = 3
	frequencyNoRatio = "absent_in_baseline"
	frequencyLowBase = "baseline_below_minimum"
)

var frequencyMeasures = []string{MeasureWord1, MeasureWord2, MeasureWord3, MeasureOpener, MeasureTemplate}

// FrequencyOptions selects the baseline cohort, the target cohorts that get
// contrasts (every other cohort when empty), the count a key needs in a
// stratum to enter the tables, the components a key needs in the target
// stratum to enter a contrast, and the number of contrasts kept per target
// stratum and measure.
type FrequencyOptions struct {
	Baseline      string
	Targets       []string
	MinCount      int
	MinComponents int
	Top           int
}

// Stratum is one cohort and role with the sentences, words, and provenance
// components it holds.
type Stratum struct {
	Cohort     string `json:"cohort"`
	Role       string `json:"role"`
	Sentences  int    `json:"sentences"`
	Words      int    `json:"words"`
	Components int    `json:"components"`
}

// FrequencyCell is one key in one stratum: its count, its rate per 1,000
// words of the stratum, and the components that carry it.
type FrequencyCell struct {
	Count            int     `json:"count"`
	PerThousandWords float64 `json:"per_thousand_words"`
	Components       int     `json:"components"`
}

// FrequencyItem is one key with its cell in every stratum where it reached
// the minimum count.
type FrequencyItem struct {
	Key   string                   `json:"key"`
	Cells map[string]FrequencyCell `json:"cells"`
}

// FrequencyContrast compares one key in a target stratum with the same role
// of the baseline cohort. The ratio divides the rates when both sides reach
// the minimum count; the status says why it is absent otherwise.
type FrequencyContrast struct {
	Cohort      string        `json:"cohort"`
	Role        string        `json:"role"`
	Key         string        `json:"key"`
	Baseline    FrequencyCell `json:"baseline"`
	Target      FrequencyCell `json:"target"`
	Ratio       *float64      `json:"ratio"`
	RatioStatus string        `json:"ratio_status"`
}

// FrequencyMeasure holds the items and contrasts of one measure.
type FrequencyMeasure struct {
	Measure   string              `json:"measure"`
	Items     []FrequencyItem     `json:"items"`
	Contrasts []FrequencyContrast `json:"contrasts"`
}

// FrequencyTables is the output: strata, then the measures. Items list the
// keys that enter a contrast, with their cell in every stratum where they
// reach the minimum count.
type FrequencyTables struct {
	Version       string             `json:"version"`
	HumanCorpus   string             `json:"human_corpus"`
	Unit          string             `json:"unit"`
	Baseline      string             `json:"baseline"`
	Targets       []string           `json:"targets"`
	MinCount      int                `json:"min_count"`
	MinComponents int                `json:"min_components"`
	Top           int                `json:"top"`
	NLP           nlp.Identity       `json:"nlp"`
	Pairing       *FrequencyPairing  `json:"pairing,omitempty"`
	Inputs        []Input            `json:"inputs"`
	Strata        []Stratum          `json:"strata"`
	Measures      []FrequencyMeasure `json:"measures"`
}

// Frequencies accumulates counts over candidate artifacts in two passes:
// Add counts every key, Select keeps the keys that reach the minimum, and
// Support counts the components of the kept keys.
type Frequencies struct {
	options  FrequencyOptions
	provider *english.Provider
	inputs   []Input
	strata   []Stratum
	index    map[string]int
	groups   []map[string]bool
	counts   map[string]map[string][]int
	selected map[string]map[string]bool
	support  map[string]map[string][]map[string]bool
	pairs    *pairing
}

// NewFrequencies prepares an accumulator with the English provider.
func NewFrequencies(options FrequencyOptions) (*Frequencies, error) {
	if options.Baseline == "" {
		options.Baseline = DefaultBaseline
	}
	if options.MinCount <= 0 {
		options.MinCount = defaultMinCount
	}
	if options.Top <= 0 {
		options.Top = defaultTopItems
	}
	if options.MinComponents <= 0 {
		options.MinComponents = defaultSupport
	}
	provider, err := english.New()
	if err != nil {
		return nil, err
	}
	counts := map[string]map[string][]int{}
	for _, measure := range frequencyMeasures {
		counts[measure] = map[string][]int{}
	}
	return &Frequencies{options: options, provider: provider, index: map[string]int{}, counts: counts}, nil
}

func stratumKey(cohort, role string) string { return cohort + "/" + role }

// stratum returns the index of a cohort and role, creating it on first sight.
func (f *Frequencies) stratum(cohort, role string) (int, error) {
	key := stratumKey(cohort, role)
	if i, found := f.index[key]; found {
		return i, nil
	}
	if len(f.strata) == maxStrata {
		return 0, fmt.Errorf("more than %d cohort and role strata", maxStrata)
	}
	f.index[key] = len(f.strata)
	f.strata = append(f.strata, Stratum{Cohort: cohort, Role: role})
	f.groups = append(f.groups, map[string]bool{})
	return len(f.strata) - 1, nil
}

// Add counts every sentence unit of one artifact.
func (f *Frequencies) Add(ctx context.Context, artifact corpus.Artifact) error {
	f.inputs = append(f.inputs, Input{SHA256: artifact.SHA256, Documents: len(artifact.Sources), Units: len(artifact.Units)})
	for _, candidate := range artifact.Units {
		if err := ctx.Err(); err != nil {
			return err
		}
		if candidate.Unit.Kind != frequencyUnit || candidate.Cohort == "" {
			continue
		}
		i, err := f.stratum(f.pairs.label(candidate), candidate.Unit.Role)
		if err != nil {
			return err
		}
		f.strata[i].Sentences++
		f.strata[i].Words += candidate.Words
		f.groups[i][candidate.GroupID] = true
		keys, err := f.keys(ctx, candidate.Unit.Text)
		if err != nil {
			return err
		}
		for measure, list := range keys {
			for _, key := range list {
				f.bump(measure, key, i)
			}
		}
	}
	return nil
}

func (f *Frequencies) bump(measure, key string, i int) {
	row := f.counts[measure][key]
	for len(row) <= i {
		row = append(row, 0)
	}
	row[i]++
	f.counts[measure][key] = row
}

// keys tags one sentence and returns its keys per measure.
func (f *Frequencies) keys(ctx context.Context, text string) (map[string][]string, error) {
	mapped := srcdoc.MappedText{Text: text, Map: make([]srcdoc.Span, len(text))}
	for i := range mapped.Map {
		mapped.Map[i] = srcdoc.Span{Start: i, End: i + 1}
	}
	sentences, err := f.provider.Analyze(ctx, mapped, []nlp.Capability{nlp.Tokens, nlp.Sentences, nlp.POS})
	if err != nil {
		return nil, err
	}
	result := map[string][]string{}
	for _, sentence := range sentences {
		words, tags := sentenceWords(sentence.Tokens)
		result[MeasureWord1] = append(result[MeasureWord1], grams(words, 1)...)
		result[MeasureWord2] = append(result[MeasureWord2], grams(words, 2)...)
		result[MeasureWord3] = append(result[MeasureWord3], grams(words, 3)...)
		if opener := opener(words); opener != "" {
			result[MeasureOpener] = append(result[MeasureOpener], opener)
		}
		if template := template(tags); template != "" {
			result[MeasureTemplate] = append(result[MeasureTemplate], template)
		}
	}
	return result, nil
}

// sentenceWords lowercases the alphabetic word tokens of a sentence. A token
// that is not such a word breaks the n-gram window, so a phrase never spans
// punctuation, a number, or an identifier. Tags follow the words.
func sentenceWords(tokens []srcdoc.Token) (words, tags []string) {
	for _, token := range tokens {
		if !token.Word || token.Protected || !alphabetic(token.Text) {
			words = append(words, "")
			tags = append(tags, "")
			continue
		}
		words = append(words, strings.ToLower(token.Text))
		tags = append(tags, coarse(token.Tag))
	}
	return words, tags
}

func alphabetic(text string) bool {
	if text == "" {
		return false
	}
	for _, r := range text {
		if !unicode.IsLetter(r) && r != '\'' && r != '-' {
			return false
		}
	}
	return true
}

func coarse(tag string) string {
	if len(tag) > 2 {
		return tag[:2]
	}
	return tag
}

// grams lists the n-grams of consecutive words; an empty word is a break.
func grams(words []string, n int) []string {
	var result []string
	for i := 0; i+n <= len(words); i++ {
		window := words[i : i+n]
		if slices.Contains(window, "") {
			continue
		}
		result = append(result, strings.Join(window, " "))
	}
	return result
}

// opener is the first three words of a sentence when it starts with three
// consecutive words.
func opener(words []string) string {
	if len(words) < 3 || words[0] == "" || words[1] == "" || words[2] == "" {
		return ""
	}
	return strings.Join(words[:3], " ")
}

// template is the coarse tag sequence of the sentence's words, cut at
// twelve tags with a marker.
func template(tags []string) string {
	kept := make([]string, 0, maxTemplateTags+1)
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		if len(kept) == maxTemplateTags {
			kept = append(kept, "...")
			break
		}
		kept = append(kept, tag)
	}
	if len(kept) < 2 {
		return ""
	}
	return strings.Join(kept, " ")
}

// Select keeps the keys that reach the minimum count in any stratum and
// prepares the support pass for them.
func (f *Frequencies) Select() {
	f.selected = map[string]map[string]bool{}
	f.support = map[string]map[string][]map[string]bool{}
	for measure, keys := range f.counts {
		f.selected[measure] = map[string]bool{}
		f.support[measure] = map[string][]map[string]bool{}
		for key, row := range keys {
			if slices.Max(row) >= f.options.MinCount {
				f.selected[measure][key] = true
			}
		}
	}
}

// Support counts, for the kept keys, the components of each stratum that
// carry them. It runs over the same artifacts as Add, after Select.
func (f *Frequencies) Support(ctx context.Context, artifact corpus.Artifact) error {
	if f.selected == nil {
		return fmt.Errorf("select keys before counting their support")
	}
	for _, candidate := range artifact.Units {
		if err := ctx.Err(); err != nil {
			return err
		}
		if candidate.Unit.Kind != frequencyUnit || candidate.Cohort == "" {
			continue
		}
		if err := f.supportUnit(ctx, candidate); err != nil {
			return err
		}
	}
	return nil
}

func (f *Frequencies) supportUnit(ctx context.Context, candidate corpus.Candidate) error {
	i := f.index[stratumKey(f.pairs.label(candidate), candidate.Unit.Role)]
	keys, err := f.keys(ctx, candidate.Unit.Text)
	if err != nil {
		return err
	}
	for measure, list := range keys {
		for _, key := range list {
			if f.selected[measure][key] {
				f.mark(measure, key, i, candidate.GroupID)
			}
		}
	}
	return nil
}

func (f *Frequencies) mark(measure, key string, i int, group string) {
	sets := f.support[measure][key]
	for len(sets) <= i {
		sets = append(sets, nil)
	}
	if sets[i] == nil {
		sets[i] = map[string]bool{}
	}
	sets[i][group] = true
	f.support[measure][key] = sets
}

// Tables assembles the output after both passes.
func (f *Frequencies) Tables() FrequencyTables {
	for i := range f.strata {
		f.strata[i].Components = len(f.groups[i])
	}
	tables := FrequencyTables{Version: FrequencyVersion, HumanCorpus: "not_qualified", Unit: frequencyUnit,
		Baseline: f.options.Baseline, Targets: f.targets(), MinCount: f.options.MinCount,
		MinComponents: f.options.MinComponents, Top: f.options.Top, NLP: f.provider.Identity(),
		Pairing: f.pairs.summary(), Inputs: f.inputs, Strata: slices.Clone(f.strata), Measures: []FrequencyMeasure{}}
	for _, measure := range frequencyMeasures {
		tables.Measures = append(tables.Measures, f.measure(measure))
	}
	return tables
}

func (f *Frequencies) cell(measure, key string, i int) FrequencyCell {
	row := f.counts[measure][key]
	cell := FrequencyCell{}
	if i < len(row) {
		cell.Count = row[i]
	}
	if words := f.strata[i].Words; words > 0 {
		cell.PerThousandWords = 1000 * float64(cell.Count) / float64(words)
	}
	if sets := f.support[measure][key]; i < len(sets) {
		cell.Components = len(sets[i])
	}
	return cell
}

func (f *Frequencies) measure(measure string) FrequencyMeasure {
	result := FrequencyMeasure{Measure: measure, Items: []FrequencyItem{}, Contrasts: []FrequencyContrast{}}
	keys := make([]string, 0, len(f.selected[measure]))
	for key := range f.selected[measure] {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	targets := f.targets()
	for i, stratum := range f.strata {
		if !slices.Contains(targets, stratum.Cohort) {
			continue
		}
		base, found := f.index[stratumKey(f.options.Baseline, stratum.Role)]
		if !found {
			continue
		}
		result.Contrasts = append(result.Contrasts, f.contrasts(measure, keys, i, base)...)
	}
	listed := map[string]bool{}
	for _, contrast := range result.Contrasts {
		listed[contrast.Key] = true
	}
	for _, key := range keys {
		if listed[key] {
			result.Items = append(result.Items, f.item(measure, key))
		}
	}
	return result
}

// targets lists the cohorts that get contrasts: the configured ones, or
// every cohort but the baseline, in sorted order.
func (f *Frequencies) targets() []string {
	if len(f.options.Targets) > 0 {
		targets := slices.Clone(f.options.Targets)
		slices.Sort(targets)
		return slices.Compact(targets)
	}
	var targets []string
	for _, stratum := range f.strata {
		if stratum.Cohort != f.options.Baseline && !slices.Contains(targets, stratum.Cohort) {
			targets = append(targets, stratum.Cohort)
		}
	}
	slices.Sort(targets)
	return targets
}

func (f *Frequencies) item(measure, key string) FrequencyItem {
	item := FrequencyItem{Key: key, Cells: map[string]FrequencyCell{}}
	for i, stratum := range f.strata {
		if cell := f.cell(measure, key, i); cell.Count >= f.options.MinCount {
			item.Cells[stratumKey(stratum.Cohort, stratum.Role)] = cell
		}
	}
	return item
}

// contrasts ranks the keys of one target stratum against the baseline of
// the same role. A key enters when the target reaches the minimum count and
// the minimum component support. Defined ratios come first, descending; then
// the keys the baseline holds below the minimum, then the keys it never
// holds, each by target count. The list is cut at the top count.
func (f *Frequencies) contrasts(measure string, keys []string, target, base int) []FrequencyContrast {
	var result []FrequencyContrast
	for _, key := range keys {
		t, b := f.cell(measure, key, target), f.cell(measure, key, base)
		if t.Count < f.options.MinCount || t.Components < f.options.MinComponents {
			continue
		}
		contrast := FrequencyContrast{Cohort: f.strata[target].Cohort, Role: f.strata[target].Role, Key: key,
			Baseline: b, Target: t, RatioStatus: frequencyNoRatio}
		switch {
		case b.Count >= f.options.MinCount && b.PerThousandWords > 0:
			ratio := t.PerThousandWords / b.PerThousandWords
			contrast.Ratio, contrast.RatioStatus = &ratio, "defined"
		case b.Count > 0:
			contrast.RatioStatus = frequencyLowBase
		}
		result = append(result, contrast)
	}
	slices.SortStableFunc(result, compareContrasts)
	if len(result) > f.options.Top {
		result = result[:f.options.Top]
	}
	return result
}

func compareContrasts(a, b FrequencyContrast) int {
	if a.Ratio != nil && b.Ratio != nil && *a.Ratio != *b.Ratio {
		if *a.Ratio > *b.Ratio {
			return -1
		}
		return 1
	}
	if rank := statusRank(a.RatioStatus) - statusRank(b.RatioStatus); rank != 0 {
		return rank
	}
	if a.Target.Count != b.Target.Count {
		return b.Target.Count - a.Target.Count
	}
	return strings.Compare(a.Key, b.Key)
}

func statusRank(status string) int {
	switch status {
	case "defined":
		return 0
	case frequencyLowBase:
		return 1
	default:
		return 2
	}
}
