package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestInstructionWording(t *testing.T) {
	for _, row := range []struct{ id, text string }{
		{"filler.instruction-scaffolding", "It is possible to display a label."},
		{"filler.instruction-scaffolding", "It is also possible to customize the dashboard."},
		{"filler.instruction-scaffolding", "It is possible to set `retry_limit` to `3`."},
		{"filler.instruction-scaffolding", "It is possible to activate and deactivate a worker."},
		{"filler.instruction-scaffolding", "If you want to display a label, you can set the label option."},
		{"filler.instruction-scaffolding", "If you wish to customize the dashboard, select a theme."},
		{"repetition.redundant-predicate", "The default path for the configuration file is located at `/etc/example.conf`."},
		{"repetition.redundant-predicate", "The location of the archive is situated in the storage directory."},
		{"repetition.redundant-predicate", "The reason for the delay is because the queue is full."},
	} {
		t.Run(row.text, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, row.id, row.text, "", "")
			c.Assert(result.Findings, qt.HasLen, 1)
			f := result.Findings[0]
			c.Assert(f.Evidence.Suggestion, qt.Not(qt.Equals), "")
			c.Assert(row.text[f.Primary.Span.Start:f.Primary.Span.End], qt.Equals, f.Primary.Snippet)
		})
	}
}

func TestInstructionWordingContext(t *testing.T) {
	id := "filler.instruction-scaffolding"
	for _, text := range []string{
		"It is possible to display a label. This is done by using the label option.",
		"It is possible to display a label. This can be achieved through the label option.",
		"Sometimes you might want to change the actor order. It is possible to specify the order by declaring the actors.",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, id, text, "", "")
			c.Assert(result.Findings, qt.HasLen, 1)
			c.Assert(result.Findings[0].Related, qt.HasLen, 1)
			c.Assert(result.Findings[0].Evidence.Metrics[0].Value, qt.Equals, float64(1))
		})
	}
	text := "It is possible to display a label.\n\nThis is done by using the label option."
	result := singleRuleResult(t, id, text, "", "")
	// An adjacent method retains its real source range as related evidence.
	qt.New(t).Assert(result.Findings, qt.HasLen, 1)
	qt.New(t).Assert(result.Findings[0].Related, qt.HasLen, 1)
}

func TestInstructionWordingControls(t *testing.T) {
	for id, texts := range map[string][]string{
		"filler.instruction-scaffolding": {
			"It is possible that the server closed the connection.",
			"It is possible to lose data during a failed write.",
			"It is possible to get an error from the server.",
			"It is not possible to change the queue.",
			"It is possible to change the queue only before startup.",
			"It is possible to change the queue when the worker is stopped.",
			"It is possible to change the queue if the worker is stopped.",
			"It is possible to configure the queue without administrator permission.",
			"It is possible to accidentally disable the queue.",
			"Is it possible to configure the queue?",
			"The specification says \"It is possible to configure the queue\".",
			"It is `possible to configure` the queue.",
			"It is possible to `configure` the queue.",
			"If you configure the queue, set its maximum size.",
			"If the worker wants to reconnect, the server checks its token.",
			"If you want to change the queue, you must obtain permission.",
			"If you want to change the queue, set its size only after draining it.",
			"Sometimes you might want to change the queue. The worker might fail.",
			"This is done by the next worker.",
			"You can configure the queue.",
		},
		"repetition.redundant-predicate": {
			"The configuration file is located at `/etc/example.conf`.",
			"The path resolver is located in the startup module.",
			"The path is not located in this directory.",
			"The `path` is located in the directory.",
			"The path is `located` in the directory.",
			"The reason for the delay is queue saturation.",
			"The delay is because the queue is full.",
			"The handler logs the reason because the worker failed.",
			"The control flow takes this path when the queue is full.",
			"Where is the path located?",
		},
	} {
		for _, text := range texts {
			t.Run(id+"/"+text, func(t *testing.T) {
				qt.New(t).Assert(singleRuleResult(t, id, text, "", "").Findings, qt.HasLen, 0)
			})
		}
	}
}

func TestInstructionWordingSourceAndPolicy(t *testing.T) {
	c := qt.New(t)
	id := "filler.instruction-scaffolding"
	phrase := "It is possible to display the café label"
	text := "\ufeffIt is possible to display the **café** label.\r\n"
	result := singleRuleResult(t, id, text, "", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Primary.Span.Start, qt.Equals, len("\ufeff"))
	c.Assert(result.Findings[0].Primary.Snippet, qt.Equals, strings.TrimSuffix(strings.TrimPrefix(text, "\ufeff"), ".\r\n"))
	c.Assert(singleRuleResult(t, id, phrase+".", "", windowTerm(id, phrase)).Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, phrase+".", "{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings, qt.HasLen, 0)
	for _, source := range []document.Source{
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte("- " + phrase + ".")},
		{Name: "sample.go", Format: document.Go, Bytes: []byte("package sample\n// " + phrase + ".\nfunc Label() {}\n")},
		{Name: "sample.py", Format: document.Python, Bytes: []byte("message = '" + phrase + ".'\n")},
	} {
		result, err := singleRuleEngine(t, id, "", "").Analyze(t.Context(), source)
		c.Assert(err, qt.IsNil)
		c.Assert(result.Findings, qt.HasLen, 1)
	}
	for _, id := range []string{id, "repetition.redundant-predicate"} {
		checkWindowAbsence(t, id, []windowAbsenceCase{{"The connection closes.", "", "", ""},
			{"# The connection closes", "", "", "unsupported_unit"}})
	}
}
