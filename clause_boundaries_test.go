package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestQualityRequiresMainSubject(t *testing.T) {
	for _, text := range []string{
		"Raise `--connect-timeout` for databases that are slow to accept connections.",
		"Use a pool for workers that are expensive to initialize.",
		"Select drivers which are easy to replace.",
		"The controller handles tasks that are difficult to schedule.",
		"`CLOCK_MONOTONIC_COARSE` has approximately the resolution corresponding to epoll, and is much faster to invoke than `CLOCK_MONOTONIC`.",
		"The query skips a full scan, and is faster.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.unscoped-assurance", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestQualityAfterGoal(t *testing.T) {
	for _, text := range []string{
		"To use these secrets in an application the process is simple.",
		"To configure the client, the process is straightforward.",
		"To deploy the service, opening the dashboard is easy.",
		"The query is slow to execute.",
		"The API is easy to use.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.unscoped-assurance", text, "", "").Findings, qt.HasLen, 1)
		})
	}
}

func TestGoalConditionsRemainScoped(t *testing.T) {
	for _, text := range []string{
		"To use these secrets, the process is simple when the provider is installed.",
		"To configure the client, the process is straightforward because it imports existing settings.",
		"The goal is to use drivers that are easy to replace.",
		"Use the documentation to select drivers that are simple to configure.",
		"According to the manual, the API is easy to use.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.unscoped-assurance", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestRelativeOperationHead(t *testing.T) {
	for _, text := range []string{
		"The library includes a standalone utility function `parse_obj_as` that can be used to apply the parsing logic.",
		"The client includes a method `Fetch` which can be used to read records.",
		"There are a number of configure options that can be used to reduce the binary size.",
		"The documentation lists configuration options that can be used to set the output format.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 1)
		})
	}
}

func TestRelativeIdentifiersDoNotSupplyRoleWords(t *testing.T) {
	for _, text := range []string{
		"The response includes a checksum `function` that can be used to cache the result.",
		"The response includes a token `reader` that can be used to fetch records.",
		"The object carries a label `configuration options` that can be used to store additional information.",
		"The request includes a filename `method` that can be used to open the database.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestGoalJudgmentSourceMapping(t *testing.T) {
	text := "Résumé.\r\n\r\nTo configure the client, the **process** is straightforward."
	r := singleRuleResult(t, "filler.unscoped-assurance", text, "", "")
	c := qt.New(t)
	c.Assert(r.Findings, qt.HasLen, 1)
	c.Assert(r.Findings[0].Primary.Span.Start, qt.Equals, strings.Index(text, "the **process**"))
	c.Assert(r.Findings[0].Primary.Snippet, qt.Equals, "the **process** is straightforward")
}
