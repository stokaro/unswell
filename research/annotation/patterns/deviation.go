// Package patterns also compares an outside tree with a baseline table. The
// comparison names constructions a tree uses more than the baseline does. It
// establishes nothing about how the tree was written: a construction can stand
// out because a project has a house style, because one author wrote most of
// it, or because the subject demands it.
package patterns

import (
	"context"
	"fmt"
	"slices"
	"sort"

	"github.com/stokaro/unswell/nlp/english"
)

// DeviationVersion identifies the proposal artifact.
const DeviationVersion = "unswell-construction-deviation-v1"

const (
	// A tree of a few hundred thousand prose words holds tens of thousands of
	// distinct closed-class trigrams. The bound is a memory guard on a map of
	// short strings, not a statement about how many a corpus should have, and
	// it is set where an ordinary documentation tree fits and a runaway input
	// does not.
	deviationMaxTerms    = 200000
	defaultDeviationMin  = 5
	defaultDeviationLift = 4
	defaultDeviationTop  = 40
)

// DeviationOptions bound a comparison. MinCount is the count a key needs in the
// tree before a rate is computed from it, MinLift the factor over the baseline
// rate a key needs to enter the proposal, and Top the number kept.
type DeviationOptions struct {
	Measure  string
	MinCount int
	MinLift  float64
	Top      int
}

// DeviationTerm is one construction the tree uses more than the baseline.
// BaselineRate is zero when the baseline holds the key below its own minimum,
// which Status records, because a rate of zero and an unmeasured rate are not
// the same claim.
type DeviationTerm struct {
	Key           string  `json:"key"`
	Count         int     `json:"count"`
	TreeRate      float64 `json:"tree_per_thousand_words"`
	BaselineRate  float64 `json:"baseline_per_thousand_words"`
	Lift          float64 `json:"lift"`
	BaselineCount int     `json:"baseline_count"`
	Status        string  `json:"status"`
}

// Deviation is the proposal: what was compared, against what, and what stood
// out. It proposes candidates for a person to judge and decides nothing.
type Deviation struct {
	Version       string          `json:"version"`
	Measure       string          `json:"measure"`
	BaselineRole  string          `json:"baseline_role"`
	BaselineWords int             `json:"baseline_words"`
	TreeWords     int             `json:"tree_words"`
	TreeSentences int             `json:"tree_sentences"`
	MinCount      int             `json:"min_count"`
	MinLift       float64         `json:"min_lift"`
	Terms         []DeviationTerm `json:"terms"`
}

// Statuses a term carries. A measured baseline rate is the ordinary case; the
// other two say why a lift is a lower bound rather than a ratio of two rates.
const (
	DeviationMeasured = "baseline_measured"
	DeviationBelowMin = "baseline_below_minimum"
	DeviationAbsent   = "absent_from_baseline"
)

// Counter accumulates the keys of one tree, one text at a time.
type Counter struct {
	frequencies   *Frequencies
	measure       string
	counts        map[string]int
	words         int
	sentences     int
	baselineWords int
}

// NewCounter prepares a counter for one measure over one tree.
func NewCounter(measure string) (*Counter, error) {
	if !validMeasure(measure) {
		return nil, fmt.Errorf("unknown measure %q", measure)
	}
	provider, err := english.New()
	if err != nil {
		return nil, err
	}
	return &Counter{frequencies: &Frequencies{provider: provider}, measure: measure,
		counts: map[string]int{}}, nil
}

func validMeasure(measure string) bool { return slices.Contains(frequencyMeasures, measure) }

// Add counts one piece of prose. Words and sentences accumulate so the rates
// have the same denominator the baseline used.
func (c *Counter) Add(ctx context.Context, text string) error {
	if len(c.counts) > deviationMaxTerms {
		return fmt.Errorf("tree holds more than %d distinct keys", deviationMaxTerms)
	}
	keys, err := c.frequencies.keys(ctx, text)
	if err != nil {
		return err
	}
	for _, key := range keys[c.measure] {
		c.counts[key]++
	}
	words, sentences, err := c.frequencies.size(ctx, text)
	if err != nil {
		return err
	}
	c.words += words
	c.sentences += sentences
	return nil
}

// Compare names the keys the tree uses more than the baseline does.
func (c *Counter) Compare(baseline BaselineTable, options DeviationOptions) (Deviation, error) {
	if baseline.Measure != c.measure {
		return Deviation{}, fmt.Errorf("baseline measures %q, not %q", baseline.Measure, c.measure)
	}
	if baseline.Words <= 0 || c.words <= 0 {
		return Deviation{}, fmt.Errorf("a comparison needs words on both sides")
	}
	options = defaultDeviation(options)
	c.baselineWords = baseline.Words
	rates := make(map[string]BaselineTerm, len(baseline.Terms))
	for _, term := range baseline.Terms {
		rates[term.Key] = term
	}
	result := Deviation{Version: DeviationVersion, Measure: c.measure, BaselineRole: baseline.Role,
		BaselineWords: baseline.Words, TreeWords: c.words, TreeSentences: c.sentences,
		MinCount: options.MinCount, MinLift: options.MinLift, Terms: []DeviationTerm{}}
	keys := make([]string, 0, len(c.counts))
	for key := range c.counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		term, keep := c.term(key, rates, options)
		if keep {
			result.Terms = append(result.Terms, term)
		}
	}
	sort.SliceStable(result.Terms, func(i, j int) bool { return result.Terms[i].Lift > result.Terms[j].Lift })
	if len(result.Terms) > options.Top {
		result.Terms = result.Terms[:options.Top]
	}
	return result, nil
}

// term builds one comparison. A key the baseline does not hold is reported with
// the rate one occurrence would have had, so its lift is a lower bound rather
// than a division by zero.
func (c *Counter) term(key string, rates map[string]BaselineTerm, options DeviationOptions) (DeviationTerm, bool) {
	count := c.counts[key]
	if count < options.MinCount {
		return DeviationTerm{}, false
	}
	tree := 1000 * float64(count) / float64(c.words)
	base, found := rates[key]
	status, baseRate := DeviationMeasured, base.PerThousandWords
	if !found {
		status, baseRate = DeviationAbsent, 0
	}
	floor := baseRate
	if floor <= 0 {
		// One occurrence is the smallest rate the baseline could have shown.
		floor = 1000 / float64(c.baselineWords)
	}
	lift := tree / floor
	if lift < options.MinLift {
		return DeviationTerm{}, false
	}
	return DeviationTerm{Key: key, Count: count, TreeRate: tree, BaselineRate: baseRate,
		Lift: lift, BaselineCount: base.Count, Status: status}, true
}

func defaultDeviation(options DeviationOptions) DeviationOptions {
	if options.MinCount <= 0 {
		options.MinCount = defaultDeviationMin
	}
	if options.MinLift <= 0 {
		options.MinLift = defaultDeviationLift
	}
	if options.Top <= 0 {
		options.Top = defaultDeviationTop
	}
	return options
}
