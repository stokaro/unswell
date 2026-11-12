package review_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/review"
)

func TestPromptCarriesTheRubricTheDocumentAndEveryItem(t *testing.T) {
	c := qt.New(t)
	prepared := prepare(c, t, "guide.md", guide)
	prompt := prepared.Prompt()
	c.Assert(strings.HasPrefix(prompt, review.Instructions), qt.IsTrue)
	c.Assert(prompt, qt.Contains, "===== DOCUMENT =====\n"+guide)
	for _, item := range prepared.Items {
		c.Assert(prompt, qt.Contains, item.ID+" (line ")
		c.Assert(prompt, qt.Contains, item.Text)
	}
	c.Assert(prompt, qt.Contains, `{"flags": []}`)
	// The whole document is shown as context, code included; only items can be flagged.
	c.Assert(prompt, qt.Contains, "install-the-service-on-every-host")
}

func TestParseFlagsReadsAnObjectInsideOtherText(t *testing.T) {
	c := qt.New(t)
	answer := "Here is the review.\n```json\n" +
		`{"flags": [{"item": "i001", "categories": ["wordiness"], "quote": "in order to", "reason": "Shorter."}]}` +
		"\n```\n"
	flags, err := review.ParseFlags(answer)
	c.Assert(err, qt.IsNil)
	c.Assert(flags, qt.DeepEquals, []review.Flag{{Item: "i001", Categories: []string{"wordiness"},
		Quote: "in order to", Reason: "Shorter."}})
	flags, err = review.ParseFlags(`{"flags": []}`)
	c.Assert(err, qt.IsNil)
	c.Assert(flags, qt.HasLen, 0)
}

func TestParseFlagsTakesTheLastObjectOfACorrectedAnswer(t *testing.T) {
	c := qt.New(t)
	answer := `{"flags": [{"item": "i001", "categories": ["wordiness"], "quote": "in order", "reason": "Shorter."}]}` +
		"\n\nWait, the quote is incomplete. Corrected:\n\n" +
		`{"flags": [{"item": "i001", "categories": ["wordiness"], "quote": "in order to", "reason": "Shorter."}]}`
	flags, err := review.ParseFlags(answer)
	c.Assert(err, qt.IsNil)
	c.Assert(flags, qt.DeepEquals, []review.Flag{{Item: "i001", Categories: []string{"wordiness"},
		Quote: "in order to", Reason: "Shorter."}})

	_, err = review.ParseFlags(`{"flags": []}` + "\nCorrected:\n" + `{"flags": [{"item": "i001"`)
	c.Assert(err, qt.ErrorMatches, "the answer is not the requested JSON object: .*")
}

func TestParseFlagsRejectsAnswersThatAreNotAReview(t *testing.T) {
	cases := []struct{ name, answer, problem string }{
		{"refusal", "I cannot review this document.", "the answer contains no JSON object"},
		{"empty", "", "the answer contains no JSON object"},
		{"other object", `{"result": "fine"}`, "the answer has no flags field"},
		{"null flags", `{"flags": null}`, "the answer has no flags field"},
		{"wrong type", `{"flags": [{"item": "i001", "categories": "wordiness"}]}`, "the answer is not the requested JSON object: .*"},
		{"truncated", `{"flags": [{"item": "i001"}`, "the answer is not the requested JSON object: .*"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			_, err := review.ParseFlags(tc.answer)
			c.Assert(err, qt.ErrorMatches, tc.problem)
		})
	}
}

func bound(c *qt.C, prepared *review.Document) []review.Finding {
	c.Helper()
	first := itemContaining(c, prepared.Items, "It is important to note")
	command := itemContaining(c, prepared.Items, "before the first start")
	accepted, rejected, err := prepared.Bind([]review.Flag{
		{Item: first.ID, Categories: []string{"empty_framing", "wordiness"},
			Quote: "It is important to note that", Reason: "The first reviewer's private reason."},
		{Item: command.ID, Categories: []string{"vague_claims"}, Quote: "before the first start", Reason: "Unclear."},
	})
	c.Assert(err, qt.IsNil)
	c.Assert(rejected, qt.HasLen, 0)
	return accepted
}

func TestConfirmationPromptListsFlagsWithoutTheFirstReasons(t *testing.T) {
	c := qt.New(t)
	prepared := prepare(c, t, "guide.md", guide)
	prompt := prepared.ConfirmationPrompt(bound(c, prepared))
	c.Assert(prompt, qt.Contains, "===== DOCUMENT =====\n"+guide)
	c.Assert(prompt, qt.Contains, `[f001] defect: The opening only announces the statement`)
	c.Assert(prompt, qt.Contains, " Or: A longer phrase where a shorter one states the same thing.")
	c.Assert(prompt, qt.Contains, `quoted text: "It is important to note that"`)
	c.Assert(prompt, qt.Contains, "item (line 3): It is important to note that the client retries after a failure.")
	c.Assert(prompt, qt.Contains, "[f002] defect: A claim with no scope or measurement.")
	c.Assert(prompt, qt.Not(qt.Contains), "private reason")
}

func TestParseVerdictsKeepsTheOrderOfThePrompt(t *testing.T) {
	c := qt.New(t)
	answer := `{"reviews": [{"id": "[f002]", "justified": "no", "reason": " A real condition. "},
		{"id": "f001", "justified": "yes", "reason": "Announces only."},
		{"id": "f001", "justified": "no", "reason": "A repeated entry is ignored."},
		{"id": "f009", "justified": "yes", "reason": "An unknown entry is ignored."}]}`
	verdicts, err := review.ParseVerdicts(answer, 2)
	c.Assert(err, qt.IsNil)
	c.Assert(verdicts, qt.DeepEquals, []review.Verdict{
		{Justified: review.Justified, Reason: "Announces only."},
		{Justified: review.Unjustified, Reason: "A real condition."},
	})
}

func TestParseVerdictsTakesTheLastObjectOfACorrectedAnswer(t *testing.T) {
	c := qt.New(t)
	answer := `{"reviews": [{"id": "f001", "justified": "yes", "reason": "Announces only."}]}` +
		"\nOn a second reading:\n" +
		`{"reviews": [{"id": "f001", "justified": "no", "reason": "It states a condition."}]}`
	verdicts, err := review.ParseVerdicts(answer, 1)
	c.Assert(err, qt.IsNil)
	c.Assert(verdicts, qt.DeepEquals, []review.Verdict{{Justified: review.Unjustified, Reason: "It states a condition."}})
}

func TestParseVerdictsLeavesASkippedFlagUnjudged(t *testing.T) {
	c := qt.New(t)
	reviews := make([]string, 0, 4)
	for _, id := range []string{"f001", "f002", "f003", "f004"} {
		reviews = append(reviews, `{"id": "`+id+`", "justified": "uncertain", "reason": ""}`)
	}
	verdicts, err := review.ParseVerdicts(`{"reviews": [`+strings.Join(reviews, ",")+`]}`, 5)
	c.Assert(err, qt.IsNil)
	c.Assert(verdicts, qt.HasLen, 5)
	c.Assert(verdicts[3].Justified, qt.Equals, review.Uncertain)
	c.Assert(verdicts[4], qt.Equals, review.Verdict{})
}

func TestParseVerdictsRejectsShortOrMalformedAnswers(t *testing.T) {
	cases := []struct{ name, answer, problem string }{
		{"too few", `{"reviews": [{"id": "f001", "justified": "yes", "reason": ""}]}`, "the answer judges 1 of 5 flags"},
		{"bad verdict", `{"reviews": [{"id": "f001", "justified": "maybe", "reason": ""}]}`, "the answer judges 0 of 5 flags"},
		{"no field", `{"flags": []}`, "the answer has no reviews field"},
		{"prose", "All flags look right.", "the answer contains no JSON object"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			_, err := review.ParseVerdicts(tc.answer, 5)
			c.Assert(err, qt.ErrorMatches, tc.problem)
		})
	}
}
