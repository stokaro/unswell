package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestBoundedHopeFraming(t *testing.T) {
	for _, tc := range []struct {
		name, text, target, suggestion string
	}{
		{
			name: "document and process hope",
			text: "Stability is a top goal for this project and we hope that this document " +
				"and the processes it entails will help to achieve that.",
			target:     "we hope that this document and the processes it entails will help to achieve that",
			suggestion: "Keep the stated goal and document scope; remove the hope about the document.",
		},
		{
			name: "indirect conditional requirement",
			text: "It is also hoped that if some implementations can provide better cache capabilities, " +
				"the solver would take advantage of that without requiring significant modification.",
			target: "It is also hoped that if some implementations can provide better cache capabilities, " +
				"the solver would take advantage of that without requiring significant modification",
			suggestion: "State the intended behavior directly; preserve the condition, actor and modification limit.",
		},
		{
			name:       "undated feature wish",
			text:       "Hopefully, some day, the `simd-accel` feature will similarly become unnecessary.",
			target:     "Hopefully, some day, the `simd-accel` feature will similarly become unnecessary",
			suggestion: "Keep the current feature behavior and warnings; omit the hope unless there is a concrete plan to report.",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			text := "Résumé.\r\n\r\n" + tc.text
			result := singleRuleResult(t, "filler.evaluative-closure", text, "", "")
			c := qt.New(t)
			c.Assert(result.Findings, qt.HasLen, 1)
			finding := result.Findings[0]
			c.Assert(finding.Primary.Span.Start, qt.Equals, strings.Index(text, tc.target))
			c.Assert(finding.Primary.Snippet, qt.Equals, tc.target)
			c.Assert(finding.Evidence.Suggestion, qt.Equals, tc.suggestion)
		})
	}
}

func TestBoundedHopeFramingControls(t *testing.T) {
	for _, text := range []string{
		"We hope that the migration succeeds if the database is available.",
		"We hope this document and its tests will help prevent a regression.",
		"It is hoped that if the error rate drops below 1%, the service would accept the traffic.",
		"It is hoped that the retry error will be fixed.",
		"Hopefully, on November 10, the upgrade will finish.",
		"Hopefully, some day, the `simd-accel` feature will become unnecessary because the compiler supports it.",
		"The maintainer said: Hopefully, some day, the feature will become unnecessary.",
		"The note says \"It is hoped that if caches improve, the solver would adapt.\"",
		"It is not hoped that if caches improve, the solver would adapt.",
		"The solver should use better cache implementations when they become available.",
		"The feature currently increases compilation time; it may be removed after the dependency update.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.evaluative-closure", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}
