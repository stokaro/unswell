package unswell_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestTutorialNarration(t *testing.T) {
	for _, text := range []string{
		"Let us explore the records returned by the client.",
		"Let's see the parser in action.",
		"Let us look at the configuration file.",
		"Next, we will demonstrate the query interface.",
		"Now we will configure the client to read these records.",
		"Next we will examine the response body.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 1)
		})
	}
}

func TestDetachedDiscourseEvaluation(t *testing.T) {
	for _, text := range []string{
		"Interestingly, the parser accepts an empty record.",
		"Surprisingly, the query returns no records when the filter is reversed.",
		"Thankfully, the client can override the default limit.",
		"Of course, the client retains the connection until the request completes.",
		"As you can see, the parser preserves the original offset.",
		"As you can infer from the output, the client does retry when the connection closes.",
		"As you know, the client reuses its connection.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.evaluative-closure", text, "", "").Findings, qt.HasLen, 1)
		})
	}
}

func TestDegreeQualityPredicate(t *testing.T) {
	for _, text := range []string{
		"The configuration is extremely simple.",
		"Opening a connection is remarkably easy.",
		"The feedback is particularly useful.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.unscoped-assurance", text, "", "").Findings, qt.HasLen, 1)
		})
	}
}

func TestDiscoursePatternControls(t *testing.T) {
	tests := map[string][]string{
		"filler.instruction-scaffolding": {
			"Let us know which adapter failed.",
			"Let us configure the server together.",
			"We will configure the server after receiving approval.",
			"We will examine the database failure tomorrow.",
			"Next we will demonstrate the adapter in the release meeting.",
			"Let us examine the records together.",
			"We will support this interface in the next release.",
			"Now we will configure the server only if approval is granted.",
			"Let us examine whether the credentials grant access.",
			"The example contains `let us explore the records` as a string.",
		},
		"filler.evaluative-closure": {
			"The operator responded thankfully to the maintainer.",
			"Interestingly, a connection.",
			"As you can see the dashboard, you have access to the metrics.",
			"If you can infer the record type, select the matching decoder.",
			"As you know the key, decrypt the payload.",
			"Of course, I can review the patch.",
			"The manual says: Interestingly, the parser accepts empty records.",
			"`Interestingly`, the parser accepts empty records.",
			"\"Interestingly, the parser accepts empty records.\"",
		},
		"filler.unscoped-assurance": {
			"The configuration is extremely simple because it contains one field.",
			"The feedback is useful when reproducing the failure.",
			"The configuration is not particularly useful.",
			"The comparison includes a very small buffer.",
			"The output contains `extremely useful` as a quoted label.",
		},
	}
	for id, texts := range tests {
		for _, text := range texts {
			t.Run(id+"/"+text, func(t *testing.T) {
				qt.New(t).Assert(singleRuleResult(t, id, text, "", "").Findings, qt.HasLen, 0)
			})
		}
	}
}

func TestDiscourseSourceMapping(t *testing.T) {
	c := qt.New(t)
	source := document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("Résumé.\r\n\r\nAs you can infer from [the output](./output), the **client** does retry when the connection closes.")}
	r, err := singleRuleEngine(t, "filler.evaluative-closure", "", "").Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(r.Findings, qt.HasLen, 1)
	f := r.Findings[0]
	c.Assert(string(source.Bytes[f.Primary.Span.Start:f.Primary.Span.End]), qt.Equals, f.Primary.Snippet)
	c.Assert(f.Primary.Snippet, qt.Contains, "[the output](./output)")
	c.Assert(f.Primary.Snippet, qt.Contains, "when the connection closes")
}
