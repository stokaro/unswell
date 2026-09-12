package rule_test

import (
	"errors"
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/rule"
)

func TestAbstentionCarriesReasonAndCause(t *testing.T) {
	cause := errors.New("editorial pattern checks exceed max_candidates")
	for _, tc := range []struct {
		name    string
		err     error
		message string
		detail  string
	}{
		{"with cause", rule.Abstain(rule.ReasonBudgetExhausted, cause),
			"rule abstained: budget_exhausted: editorial pattern checks exceed max_candidates",
			"editorial pattern checks exceed max_candidates"},
		{"without cause", rule.Abstain(rule.ReasonBudgetExhausted, nil), "rule abstained: budget_exhausted", ""},
		{"wrapped", fmt.Errorf("outer: %w", rule.Abstain(rule.ReasonBudgetExhausted, cause)),
			"outer: rule abstained: budget_exhausted: editorial pattern checks exceed max_candidates",
			"editorial pattern checks exceed max_candidates"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			var abstention *rule.Abstention
			c.Assert(tc.err, qt.ErrorAs, &abstention)
			c.Assert(tc.err.Error(), qt.Equals, tc.message)
			c.Assert(abstention.Reason, qt.Equals, rule.ReasonBudgetExhausted)
			c.Assert(abstention.Detail(), qt.Equals, tc.detail)
			c.Assert(feature.ValidApplicabilityReason(abstention.Reason), qt.IsTrue)
			if tc.detail != "" {
				c.Assert(tc.err, qt.ErrorIs, cause)
			}
		})
	}
}
