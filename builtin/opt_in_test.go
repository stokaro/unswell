package builtin_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/builtin"
)

// A rule that ships disabled is a policy decision, so the catalog carries the
// reason beside the rule. This test is what keeps the record from drifting:
// a new opt-in rule fails here until its reason is written, and a reason left
// behind by a rule that now ships enabled fails here too.
func TestEveryOptInRuleRecordsItsReason(t *testing.T) {
	c := qt.New(t)
	var optIn int
	for _, r := range builtin.Rules() {
		d := r.Descriptor()
		if d.Defaults.Enabled {
			c.Assert(d.OptInReason, qt.Equals, "",
				qt.Commentf("%s ships enabled and must carry no opt-in reason", d.ID))
			continue
		}
		optIn++
		c.Assert(d.OptInReason, qt.Not(qt.Equals), "",
			qt.Commentf("%s ships disabled and must record why", d.ID))
		c.Assert(len(d.OptInReason) > 40, qt.IsTrue,
			qt.Commentf("%s records a reason too short to be one: %q", d.ID, d.OptInReason))
	}
	c.Assert(optIn, qt.Equals, 23)
}
