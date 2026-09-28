package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestInfinitiveDocumentEndorsement(t *testing.T) {
	for _, text := range []string{
		"I hope this guide helped to get you up and running with the client.",
		"We hope that this tutorial has helped to show readers how retries work.",
		"I trust this chapter helps explain the retry policy.",
		"We hope this guide helped configure the client.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.evaluative-closure", text, "", "").Findings, qt.HasLen, 1)
		})
	}
}

func TestInfinitiveDocumentEndorsementControls(t *testing.T) {
	for _, text := range []string{
		"I hope this guide helped to prevent data loss.",
		"I hope this guide helped, but the failure remains unexplained.",
		"I hope this guide helped to.",
		"I hope this guide helped to not mislead readers.",
		"I hope this guide helped `to get you` started.",
		"The user said: I hope this guide helped to configure the client.",
		"I hope this guide helps with database recovery.",
		"I hope this service helped to configure the client.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.evaluative-closure", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestGenericDemonstrationAnnouncement(t *testing.T) {
	for _, text := range []string{
		"Let's see it in action.",
		"Let us see this in action.",
		"Let’s see it in action.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.document-metadiscourse", text, "", "").Findings, qt.HasLen, 1)
		})
	}
}

func TestGenericDemonstrationControls(t *testing.T) {
	for _, text := range []string{
		"Let us see whether the command fails.",
		"Let's see the retry counter in action.",
		"Let's see it in action when the network fails.",
		"Let's see it in action with a failed request.",
		"Do you want to see it in action?",
		"The author said: let's see it in action.",
		"The message is \"Let's see it in action.\"",
		"The button reads `Let's see it in action.`",
		"Let's `see it` in action.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.document-metadiscourse", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestReaderFramingSourceMapping(t *testing.T) {
	for _, tc := range []struct{ rule, text string }{
		{"filler.document-metadiscourse", "Let’s **see** it in action."},
		{"filler.evaluative-closure", "I hope this guide helped to **get** you started."},
	} {
		text := "Résumé.\r\n\r\n" + tc.text
		r := singleRuleResult(t, tc.rule, text, "", "")
		c := qt.New(t)
		c.Assert(r.Findings, qt.HasLen, 1)
		c.Assert(r.Findings[0].Primary.Span.Start, qt.Equals, strings.Index(text, tc.text))
		c.Assert(r.Findings[0].Primary.Snippet, qt.Equals, strings.TrimSuffix(tc.text, "."))
	}
}
