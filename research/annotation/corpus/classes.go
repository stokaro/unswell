package corpus

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
	"github.com/stokaro/unswell/rule"
)

// RuleClassesVersion identifies the rule-class input of experiment E1.
const RuleClassesVersion = "unswell-rule-classes-v1"

// MaxRuleClassBytes bounds a rule-class file.
const MaxRuleClassBytes = 1 << 20

// RuleClass records one rule's class before any cohort is measured. Roles
// list where the rule's count enters a prevalence table. HypothesisSource
// names the sources-record entry behind a candidate, or nothing.
type RuleClass struct {
	RuleID           string   `json:"rule_id"`
	Class            string   `json:"class"`
	Roles            []string `json:"roles"`
	HypothesisSource *string  `json:"hypothesis_source"`
	Reason           string   `json:"reason"`
}

// RuleClasses is the committed class list. A class is a design input of the
// pattern protocol, never a measured result; changing one is an amendment.
type RuleClasses struct {
	Format    string            `json:"format"`
	Revision  int               `json:"revision"`
	DecidedOn string            `json:"decided_on"`
	Protocol  string            `json:"protocol"`
	Scope     string            `json:"scope"`
	Classes   map[string]string `json:"classes"`
	Rules     []RuleClass       `json:"rules"`
}

// LoadRuleClasses decodes strict JSON and checks the list's own consistency.
// Coverage of an actual rule catalog is a separate check.
func LoadRuleClasses(ctx context.Context, data []byte) (RuleClasses, error) {
	var classes RuleClasses
	if err := jsoninput.Decode(ctx, data, MaxRuleClassBytes, &classes, jsoninput.Limits{Array: 1000, Object: 1000}); err != nil {
		return RuleClasses{}, err
	}
	if err := classes.validate(); err != nil {
		return RuleClasses{}, err
	}
	return classes, nil
}

func (classes RuleClasses) validate() error {
	if err := classes.validateHeader(); err != nil {
		return err
	}
	for name, meaning := range classes.Classes {
		if !slices.Contains(ruleClassNames(), name) || !text(meaning) {
			return fmt.Errorf("unknown or undescribed rule class %q", name)
		}
	}
	return classes.validateRules()
}

func (classes RuleClasses) validateHeader() error {
	for _, value := range []string{classes.DecidedOn, classes.Protocol, classes.Scope} {
		if !text(value) {
			return fmt.Errorf("rule classes require a decision date, a protocol, and a scope")
		}
	}
	if classes.Format != RuleClassesVersion || classes.Revision <= 0 || len(classes.Rules) == 0 {
		return fmt.Errorf("rule classes require the %s format, a positive revision, and rules", RuleClassesVersion)
	}
	return nil
}

func (classes RuleClasses) validateRules() error {
	previous := ""
	for _, entry := range classes.Rules {
		if err := entry.validate(classes.Classes); err != nil {
			return fmt.Errorf("rule class %q: %w", entry.RuleID, err)
		}
		if strings.Compare(previous, entry.RuleID) >= 0 {
			return fmt.Errorf("rule classes must list each rule once in ascending order")
		}
		previous = entry.RuleID
	}
	return nil
}

func (entry RuleClass) validate(described map[string]string) error {
	if !text(entry.RuleID) || !text(entry.Reason) {
		return fmt.Errorf("a rule ID and a reason are required")
	}
	if _, found := described[entry.Class]; !found {
		return fmt.Errorf("class %q is not described", entry.Class)
	}
	if err := choices(entry.Roles, roles(), len(roles())); err != nil {
		return err
	}
	if entry.HypothesisSource != nil && !text(*entry.HypothesisSource) {
		return fmt.Errorf("a hypothesis source must be nonempty text when present")
	}
	return nil
}

// Cover checks that the list names every catalog rule exactly once and no
// rule outside the catalog, so a new rule cannot enter E1 unclassified.
func (classes RuleClasses) Cover(catalog []rule.Descriptor) error {
	expected := make([]string, 0, len(catalog))
	for _, descriptor := range catalog {
		expected = append(expected, descriptor.ID)
	}
	slices.Sort(expected)
	actual := make([]string, 0, len(classes.Rules))
	for _, entry := range classes.Rules {
		actual = append(actual, entry.RuleID)
	}
	if !slices.Equal(expected, actual) {
		return fmt.Errorf("rule classes cover %d rules while the catalog has %d, or the sets differ", len(actual), len(expected))
	}
	return nil
}

func ruleClassNames() []string {
	return []string{"general_style", "explicit_prohibition", "llm_associated_candidate"}
}
