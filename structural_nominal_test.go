package unswell_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestNominalSupportWording(t *testing.T) {
	for _, text := range []string{
		"The reader makes use of the version marker.",
		"Make use of `route_pattern` when configuring the listener.",
		"The reader avoids duplicates through the use of a version marker.",
		"The reader avoids duplicates through the use of `entry_id`.",
		"The loader refreshes its configuration on a periodic basis.",
		"The loaders refresh their configuration in a parallel manner.",
		"The reader does not make use of the optional cache.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.nominal-support", text, "", "").Findings, qt.HasLen, 1)
		})
	}
}

func TestNominalSupportControls(t *testing.T) {
	for _, text := range []string{
		"The reader makes full use of the available buffer.",
		"The reader makes no use of the optional cache.",
		"The license restricts use of the renderer.",
		"The tutorial explains the use of version markers.",
		"The right to make use of the license is granted to the recipient.",
		"The reader makes `use of` the optional cache.",
		"The reader `makes use of` the optional cache.",
		"The reader operates on a per-request basis.",
		"The policy evaluates requests on a case-by-case basis.",
		"The task runs in a manner defined by the scheduler.",
		"The worker runs on a basis of 24 hours per task.",
		"The specification says the reader makes use of the cache.",
		"\"The reader makes use of the cache.\"",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.nominal-support", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}
