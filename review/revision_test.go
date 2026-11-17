package review_test

import (
	"fmt"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/review"
)

const servicePage = "# Retry policy\n\n" +
	"The client retries a request after a connection failure. The delay doubles after every attempt, up to 30 seconds.\n\n" +
	"The client does not retry a request that the server rejected. Run `make build` before the first start.\n\n" +
	"```sh\nmake build\nmake start\n```\n\n" +
	"The [protocol](https://example.com/protocol) has version 2.1.\n"

const serviceRevision = "# Retry policy\n\n" +
	"The client retries a request after a connection failure, and the delay doubles after every attempt, up to\n" +
	"30 seconds.\n\n" +
	"A request that the server rejected is **not** retried. Run `make build` before the first start.\n\n" +
	"```sh\nmake build\nmake start\n```\n\n" +
	"The [protocol](https://example.com/protocol) has version 2.1.\n"

// revisionOf pairs the service page with a revision of it.
func revisionOf(c *qt.C, t *testing.T, revised string) *review.Revision {
	t.Helper()
	revision, err := review.NewRevision(prepare(c, t, "page.md", servicePage), prepare(c, t, "revised.md", revised))
	c.Assert(err, qt.IsNil)
	return revision
}

func TestNewRevisionNeedsProseOnBothSides(t *testing.T) {
	c := qt.New(t)
	page := prepare(c, t, "page.md", servicePage)
	empty := prepare(c, t, "empty.md", "```sh\nmake build\n```\n")
	_, err := review.NewRevision(page, nil)
	c.Assert(err, qt.ErrorMatches, "a revision check needs a page and its revision")
	_, err = review.NewRevision(empty, page)
	c.Assert(err, qt.ErrorMatches, "the page has no prose")
	_, err = review.NewRevision(page, empty)
	c.Assert(err, qt.ErrorMatches, "the revision has no prose")
}

func TestSoundPromptHidesWhichVersionIsTheRevision(t *testing.T) {
	c := qt.New(t)
	revision := revisionOf(c, t, serviceRevision)
	pageFirst, revisionFirst := revision.SoundPrompt(false), revision.SoundPrompt(true)
	c.Assert(pageFirst, qt.Contains, "===== TEXT A =====\n"+servicePage+"\n===== TEXT B =====\n"+serviceRevision)
	c.Assert(revisionFirst, qt.Contains, "===== TEXT A =====\n"+serviceRevision+"\n===== TEXT B =====\n"+servicePage)
	c.Assert(pageFirst, qt.Contains, "Below are the openings of two technical texts, A and B.")
	for _, prompt := range []string{pageFirst, revisionFirst} {
		instructions, _, found := strings.Cut(prompt, "===== TEXT A =====")
		c.Assert(found, qt.IsTrue)
		c.Assert(strings.ToLower(instructions), qt.Not(qt.Contains), "revis")
		c.Assert(strings.ToLower(instructions), qt.Not(qt.Contains), "original")
		// The form of the answer stands before the texts, as in the measured protocol.
		c.Assert(instructions, qt.Contains, `{"more_like_a_model": "A" | "B" | "same", "reason": "<one sentence>"}`)
	}
}

func TestOpeningKeepsWholeBlocksUpToTheWordLimit(t *testing.T) {
	c := qt.New(t)
	paragraph := strings.TrimSpace(strings.Repeat("The client retries a request after a failure. ", 25))
	source := "---\ntitle: Retry policy\n---\n\n# Retry policy\n\n" + paragraph + "\n\n" +
		"```sh\nmake build\n\nmake start\n```\n\n" + paragraph + "\n\n" + paragraph + "\n\nThe last paragraph is not read.\n"
	opening := review.Opening([]byte(source))
	// Front matter is left out. Each paragraph has 200 words, and the code block counts none: the third paragraph
	// reaches the limit and is kept whole.
	c.Assert(opening, qt.Equals, "# Retry policy\n\n"+paragraph+"\n\n```sh\nmake build\n\nmake start\n```\n\n"+paragraph+"\n\n"+paragraph+"\n")
	c.Assert(len(strings.Fields(opening)) > review.SoundWords, qt.IsTrue)
	// A source that is shorter than the limit is its own opening.
	c.Assert(review.Opening([]byte(servicePage)), qt.Equals, servicePage)
	c.Assert(review.Opening([]byte("+++\ntitle = \"x\"\n+++\nOne line.\r\nAnother line.\n")), qt.Equals, "One line.\nAnother line.\n")
	// An unclosed front matter mark is text.
	c.Assert(review.Opening([]byte("---\nA rule, then text.\n")), qt.Equals, "---\nA rule, then text.\n")
}

func TestSoundPromptShowsOnlyTheOpenings(t *testing.T) {
	c := qt.New(t)
	long := strings.TrimSpace(strings.Repeat("The client retries a request after a failure. ", 70))
	page := "# Retry policy\n\n" + long + "\n\nThe page ends with this sentence about the limit of 30 seconds.\n"
	revised := "# Retry policy\n\n" + long + "\n\nThe revision ends with other words about the limit of 30 seconds.\n"
	revision, err := review.NewRevision(prepare(c, t, "page.md", page), prepare(c, t, "revised.md", revised))
	c.Assert(err, qt.IsNil)
	c.Assert(revision.SoundPrompt(false), qt.Not(qt.Contains), "ends with")
	// The statements are still listed from the whole page.
	c.Assert(revision.FactsPrompt(), qt.Contains, "The page ends with this sentence")
}

func TestParseSoundNamesTheTextBehindTheLetter(t *testing.T) {
	cases := []struct {
		name          string
		answer        string
		revisionFirst bool
		named         string
	}{
		{"A with the page first", `{"more_like_a_model": "A", "reason": "uniform"}`, false, review.NamedPage},
		{"B with the page first", `{"more_like_a_model": "B", "reason": "uniform"}`, false, review.NamedRevision},
		{"A with the revision first", `{"more_like_a_model": "A", "reason": "uniform"}`, true, review.NamedRevision},
		{"B with the revision first", `{"more_like_a_model": " b ", "reason": "uniform"}`, true, review.NamedPage},
		{"same", "Here it is:\n{\"more_like_a_model\": \"same\", \"reason\": \"alike\"}", true, review.NamedNeither},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			answer, err := review.ParseSound(test.answer, test.revisionFirst)
			c.Assert(err, qt.IsNil)
			c.Assert(answer.Named, qt.Equals, test.named)
			c.Assert(answer.Reason, qt.Not(qt.Equals), "")
		})
	}
}

func TestParseSoundRejectsAnAnswerWithoutAChoice(t *testing.T) {
	cases := map[string]string{
		"I cannot compare these texts.":                         "the answer contains no JSON object",
		`{"reason": "both read well"}`:                          "the answer has no more_like_a_model field",
		`{"more_like_a_model": "the second", "reason": "long"}`: "more_like_a_model must be A, B, or same",
	}
	for answer, problem := range cases {
		t.Run(problem, func(t *testing.T) {
			c := qt.New(t)
			_, err := review.ParseSound(answer, false)
			c.Assert(err, qt.ErrorMatches, problem)
		})
	}
}

func TestSoundVerdictNeedsTheSameTextInBothOrders(t *testing.T) {
	named := func(text string) review.SoundAnswer { return review.SoundAnswer{Named: text} }
	cases := []struct {
		name                     string
		pageFirst, revisionFirst string
		verdict                  string
	}{
		{"the page twice", review.NamedPage, review.NamedPage, review.SoundBetter},
		{"the revision twice", review.NamedRevision, review.NamedRevision, review.SoundWorse},
		// A judge that always names text A names the page and then the revision.
		{"the first letter twice", review.NamedPage, review.NamedRevision, review.SoundSame},
		{"the second letter twice", review.NamedRevision, review.NamedPage, review.SoundSame},
		{"once the page, once neither", review.NamedPage, review.NamedNeither, review.SoundSame},
		{"neither twice", review.NamedNeither, review.NamedNeither, review.SoundSame},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(review.SoundVerdict(named(test.pageFirst), named(test.revisionFirst)), qt.Equals, test.verdict)
		})
	}
}

func TestFactsAreListedFromThePageAndLookedForInTheRevision(t *testing.T) {
	c := qt.New(t)
	revision := revisionOf(c, t, serviceRevision)
	c.Assert(revision.FactsPrompt(), qt.Contains, "===== PAGE =====\n"+servicePage)
	c.Assert(revision.FactsPrompt(), qt.Contains, "its facts and its instructions to the reader")
	c.Assert(revision.FactsPrompt(), qt.Not(qt.Contains), "is **not** retried")
	prompt := revision.KeptPrompt([]string{"The delay doubles after every attempt.", "The protocol has version 2.1."})
	c.Assert(prompt, qt.Contains, "===== STATEMENTS =====\n1. The delay doubles after every attempt.\n2. The protocol has version 2.1.\n")
	c.Assert(prompt, qt.Contains, "===== PAGE =====\n"+serviceRevision)
	c.Assert(prompt, qt.Not(qt.Contains), "The client does not retry")
}

func TestParseFactsDropsEmptyAndRepeatedStatements(t *testing.T) {
	c := qt.New(t)
	facts, err := review.ParseFacts(`{"facts": ["The delay doubles.", " ", "The delay  doubles.", "The limit is 30 seconds."]}`)
	c.Assert(err, qt.IsNil)
	c.Assert(facts, qt.DeepEquals, []string{"The delay doubles.", "The limit is 30 seconds."})
}

func TestParseFactsRejectsAnAnswerWithoutStatements(t *testing.T) {
	many := make([]string, review.MaxFacts+1)
	for index := range many {
		many[index] = fmt.Sprintf("%q", fmt.Sprintf("Statement %d.", index))
	}
	cases := map[string]string{
		`{"facts": []}`:                                "the answer lists no statement",
		`{"facts": ["", "  "]}`:                        "the answer lists no statement",
		`{"statements": ["The delay doubles."]}`:       "the answer has no facts field",
		"The page states nothing.":                     "the answer contains no JSON object",
		`{"facts": [` + strings.Join(many, ",") + `]}`: "the answer lists more than 300 statements",
	}
	for answer, problem := range cases {
		t.Run(problem, func(t *testing.T) {
			c := qt.New(t)
			_, err := review.ParseFacts(answer)
			c.Assert(err, qt.ErrorMatches, problem)
		})
	}
}

func TestParseKeptReturnsOneVerdictForEachStatement(t *testing.T) {
	c := qt.New(t)
	verdicts, err := review.ParseKept(`{"facts": [
		{"fact": 2, "kept": "Changed", "quote": ""},
		{"fact": 1, "kept": "yes", "quote": "the delay doubles after every attempt"}]}`, 2)
	c.Assert(err, qt.IsNil)
	c.Assert(verdicts, qt.DeepEquals, []review.FactVerdict{
		{Fact: 1, Kept: "yes", Quote: "the delay doubles after every attempt"},
		{Fact: 2, Kept: review.FactChanged},
	})
}

func TestParseKeptRejectsAnIncompleteAnswer(t *testing.T) {
	first := `{"fact": 1, "kept": "yes", "quote": "a b c"}`
	cases := map[string]string{
		`{"facts": [` + first + `]}`:                                    "the answer has no entry for statement 2",
		`{"facts": [` + first + `, {"fact": 1, "kept": "missing"}]}`:    "the answer names statement 1 twice",
		`{"facts": [` + first + `, {"fact": 3, "kept": "missing"}]}`:    "the answer names statement 3 of 2",
		`{"facts": [` + first + `, {"fact": 2, "kept": "mostly"}]}`:     "kept must be yes, changed, or missing for statement 2",
		`{"verdicts": [` + first + `, {"fact": 2, "kept": "missing"}]}`: "the answer has no facts field",
	}
	for answer, problem := range cases {
		t.Run(problem, func(t *testing.T) {
			c := qt.New(t)
			_, err := review.ParseKept(answer, 2)
			c.Assert(err, qt.ErrorMatches, problem)
		})
	}
}

func TestBindFactsKeepsOnlyAFactThatIsQuotedFromTheRevision(t *testing.T) {
	c := qt.New(t)
	revision := revisionOf(c, t, serviceRevision)
	statements := []string{
		"The delay doubles after every attempt, up to 30 seconds.",
		"The client does not retry a request that the server rejected.",
		"The protocol has version 2.1.",
		"The client retries after a connection failure.",
		"The delay starts at one second.",
		"The server rejects some requests.",
	}
	facts, err := revision.BindFacts(statements, []review.FactVerdict{
		// The quote runs over a line break of the revision.
		{Fact: 1, Kept: "yes", Quote: "the delay doubles after every attempt, up to 30 seconds"},
		// The revision marks a word as bold; the quote does not.
		{Fact: 2, Kept: "yes", Quote: "A request that the server rejected is not retried"},
		// Words that are not in the revision.
		{Fact: 3, Kept: "yes", Quote: "the protocol is at version 2.1"},
		// Too few words to bind.
		{Fact: 4, Kept: "yes", Quote: "client retries"},
		{Fact: 5, Kept: review.FactMissing},
		{Fact: 6, Kept: review.FactChanged, Quote: "is not retried"},
	})
	c.Assert(err, qt.IsNil)
	outcomes := make([]string, len(facts))
	for index, fact := range facts {
		outcomes[index] = fact.Outcome
		c.Assert(fact.Text, qt.Equals, statements[index])
	}
	c.Assert(outcomes, qt.DeepEquals, []string{review.FactKept, review.FactKept, review.FactUnquoted, review.FactUnquoted,
		review.FactMissing, review.FactChanged})
	c.Assert(facts[0].Quote, qt.Equals, "the delay doubles after every attempt, up to 30 seconds")
	// An unquoted fact keeps the words that the model gave, so that a reader can see what was claimed.
	c.Assert(facts[2].Quote, qt.Equals, "the protocol is at version 2.1")
	c.Assert(facts[4].Quote, qt.Equals, "")
	c.Assert(review.KeptFacts(facts), qt.Equals, 2)

	_, err = revision.BindFacts(statements, nil)
	c.Assert(err, qt.ErrorMatches, "6 statements have 0 verdicts")

	// A quote reads a link as its words, without the address.
	linked, err := revision.BindFacts(statements[2:3], []review.FactVerdict{{Fact: 1, Kept: "yes", Quote: "The protocol has version 2.1"}})
	c.Assert(err, qt.IsNil)
	c.Assert(linked[0].Outcome, qt.Equals, review.FactKept)
}

func TestBindFactsReadsAContractionAndACapitalAsTheSameWords(t *testing.T) {
	statement := []string{"The client does not retry a request that the server rejected."}
	contracted := strings.Replace(serviceRevision, "is **not** retried", "isn't retried", 1)
	qt.New(t).Assert(contracted, qt.Contains, "rejected isn't retried")
	cases := []struct {
		name    string
		revised string
		quote   string
		outcome string
	}{
		{
			"the quote contracts what the revision writes out", serviceRevision,
			"A request that the server rejected isn't retried", review.FactKept,
		},
		{
			"the quote writes out what the revision contracts", contracted,
			"A request that the server rejected is not retried", review.FactKept,
		},
		{"a typographic apostrophe", contracted, "A request that the server rejected isn’t retried", review.FactKept},
		{"another case of the letters", serviceRevision, "a request that the Server rejected is not retried", review.FactKept},
		// The quote leaves the negation out, so its words do not stand in either revision.
		{
			"a quote without the negation, written out", serviceRevision,
			"A request that the server rejected is retried", review.FactUnquoted,
		},
		{"a quote without the negation, contracted", contracted, "A request that the server rejected is retried", review.FactUnquoted},
		{"other words", contracted, "A request that the server refused isn't retried", review.FactUnquoted},
		// Written out, the quote has three words, and it still has two as the model gave it.
		{"too few words as quoted", contracted, "isn't retried", review.FactUnquoted},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			facts, err := revisionOf(c, t, test.revised).BindFacts(statement, []review.FactVerdict{{Fact: 1, Kept: "yes", Quote: test.quote}})
			c.Assert(err, qt.IsNil)
			c.Assert(facts[0].Outcome, qt.Equals, test.outcome)
			// The report keeps the words as the model gave them.
			c.Assert(facts[0].Quote, qt.Equals, test.quote)
		})
	}
	c := qt.New(t)
	// "cannot" and "can't" are "can not".
	unable := strings.Replace(serviceRevision, "is **not** retried", "cannot be retried", 1)
	c.Assert(unable, qt.Contains, "rejected cannot be retried")
	quotes := []string{
		"the server rejected can't be retried", "the server rejected can not be retried", "the server rejected cannot be retried",
	}
	for _, quote := range quotes {
		facts, err := revisionOf(c, t, unable).BindFacts(statement, []review.FactVerdict{{Fact: 1, Kept: "yes", Quote: quote}})
		c.Assert(err, qt.IsNil)
		c.Assert(facts[0].Outcome, qt.Equals, review.FactKept, qt.Commentf("quote %q", quote))
	}
}

func TestConfirmedKeepsOnlyStatementsThatThePageItselfStates(t *testing.T) {
	c := qt.New(t)
	revision := revisionOf(c, t, serviceRevision)
	statements := []string{
		"The client does not retry a request that the server rejected.",
		"The delay doubles after every attempt and never exceeds one minute.",
		"The server rejects a request after 30 seconds.",
	}
	c.Assert(revision.ControlPrompt(statements), qt.Contains, "===== PAGE =====\n"+servicePage)
	c.Assert(revision.KeptPrompt(statements), qt.Contains, "===== PAGE =====\n"+serviceRevision)
	confirmed, dropped, err := revision.Confirmed(statements, []review.FactVerdict{
		// The words stand in the page and not in the revision: the control reads the page.
		{Fact: 1, Kept: "yes", Quote: "The client does not retry a request"},
		// The model worded the statement so that the page does not state it.
		{Fact: 2, Kept: review.FactChanged},
		// The model calls it stated and quotes words that are not in the page.
		{Fact: 3, Kept: "yes", Quote: "the server rejects a request after 30 seconds"},
	})
	c.Assert(err, qt.IsNil)
	c.Assert(confirmed, qt.DeepEquals, statements[:1])
	c.Assert(dropped, qt.DeepEquals, statements[1:])
	_, _, err = revision.Confirmed(statements, nil)
	c.Assert(err, qt.ErrorMatches, "3 statements have 0 verdicts")
}

func TestLostFindsWhatTheRevisionNoLongerCarries(t *testing.T) {
	cases := []struct {
		name     string
		revised  string
		expected review.Lost
	}{
		{"nothing", serviceRevision, review.Lost{Numbers: []string{}, Code: []string{}, Links: []string{}}},
		{
			"a number written as a word", strings.Replace(serviceRevision, "30 seconds", "thirty seconds", 1),
			review.Lost{Numbers: []string{"30"}, Code: []string{}, Links: []string{}},
		},
		{
			"a version", strings.Replace(serviceRevision, "version 2.1", "version 2", 1),
			review.Lost{Numbers: []string{"2.1"}, Code: []string{}, Links: []string{}},
		},
		{
			// The same code still stands in the code block.
			"inline code that a block still holds", strings.Replace(serviceRevision, "Run `make build` before", "Build it before", 1),
			review.Lost{Numbers: []string{}, Code: []string{}, Links: []string{}},
		},
		{
			// The block is named by its first line, which is also the inline code.
			"code that is nowhere else", strings.ReplaceAll(serviceRevision, "make build", "make all"),
			review.Lost{Numbers: []string{}, Code: []string{"make build"}, Links: []string{}},
		},
		{
			"a line of a code block", strings.Replace(serviceRevision, "make start\n", "make run\n", 1),
			review.Lost{Numbers: []string{}, Code: []string{"make build"}, Links: []string{}},
		},
		{
			"a link", strings.Replace(serviceRevision, "[protocol](https://example.com/protocol)", "protocol", 1),
			review.Lost{Numbers: []string{}, Code: []string{}, Links: []string{"https://example.com/protocol"}},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			lost := revisionOf(c, t, test.revised).Lost()
			c.Assert(lost, qt.DeepEquals, test.expected)
			c.Assert(lost.Any(), qt.Equals, len(test.expected.Numbers)+len(test.expected.Code)+len(test.expected.Links) > 0)
		})
	}
}

func TestLostReadsCodeSpansOfAnyLength(t *testing.T) {
	c := qt.New(t)
	page := "# Notes\n\nA span of two backticks holds a backtick: ``a ` b``. After it `first` and `second` are spans, and the words " +
		"between them are prose. An unclosed ` backtick opens nothing.\n"
	prepared := prepare(c, t, "page.md", page)
	// The revision keeps every span and rewrites the prose between them.
	kept, err := review.NewRevision(prepared, prepare(c, t, "revised.md",
		"# Notes\n\nTwo backticks hold one: ``a ` b``. Then come `first` and `second`, with prose around them. A lone ` opens nothing.\n"))
	c.Assert(err, qt.IsNil)
	c.Assert(kept.Lost().Code, qt.HasLen, 0)
	// A revision that drops one span loses that span and nothing else.
	dropped, err := review.NewRevision(prepared, prepare(c, t, "revised.md",
		"# Notes\n\nTwo backticks hold one: ``a ` b``. Then comes `first`, with prose around it. A lone ` opens nothing.\n"))
	c.Assert(err, qt.IsNil)
	c.Assert(dropped.Lost().Code, qt.DeepEquals, []string{"second"})
}

func TestEmDashesAreCountedInProseOnly(t *testing.T) {
	c := qt.New(t)
	revised := strings.Replace(serviceRevision, "failure, and the delay", "failure—the delay", 1)
	revised = strings.Replace(revised, "make start\n", "make start # the last step — optional\n", 1)
	page, after := revisionOf(c, t, revised).EmDashes()
	c.Assert(page, qt.Equals, 0)
	c.Assert(after, qt.Equals, 1)
}

func kept(count int) []review.Fact {
	facts := make([]review.Fact, count)
	for index := range facts {
		facts[index] = review.Fact{Text: fmt.Sprintf("Statement %d.", index), Outcome: review.FactKept}
	}
	return facts
}

// suspect is a sentence that the revision may have lost, with what a judge said of it.
func suspect(reason, outcome string) review.Suspect {
	return review.Suspect{Line: 3, Text: "A sentence of the page.", Reason: reason, Revised: "Words of the revision.", Outcome: outcome}
}

func TestJudgeAcceptsOnlyARevisionThatPassesEveryCheck(t *testing.T) {
	withLoss := func(facts []review.Fact, lost int) []review.Fact {
		for index := range lost {
			facts[index].Outcome = review.FactChanged
		}
		return facts
	}
	dashed := strings.Replace(serviceRevision, "failure, and the delay", "failure—the delay", 1)
	cases := []struct {
		name      string
		revised   string
		sound     string
		facts     []review.Fact
		sentences []review.Suspect
		reasons   []string
	}{
		{"every check passes", serviceRevision, review.SoundBetter, kept(50), nil, []string{}},
		{"one statement in fifty is lost", serviceRevision, review.SoundBetter, withLoss(kept(50), 1), nil, []string{}},
		{
			"two statements in fifty are lost", serviceRevision, review.SoundBetter, withLoss(kept(50), 2), nil,
			[]string{"2 of 50 statements of the page are not kept"},
		},
		{
			"the judge hears no difference", serviceRevision, review.SoundSame, kept(4), nil,
			[]string{"the judge does not find the revision less like a model's text than the page"},
		},
		{
			"the judge prefers the page", serviceRevision, review.SoundWorse, kept(4), nil,
			[]string{"the judge finds the revision more like a model's text than the page"},
		},
		{"no statement was checked", serviceRevision, review.SoundBetter, nil, nil, []string{"no statement of the page was checked"}},
		{"an em dash is added", dashed, review.SoundBetter, kept(4), nil, []string{"the revision adds 1 em dashes"}},
		{
			"a number and a link are lost",
			strings.NewReplacer("30 seconds", "half a minute", "[protocol](https://example.com/protocol)", "protocol").Replace(serviceRevision),
			review.SoundBetter, kept(4), nil, []string{"the revision loses 1 numbers, 1 links"},
		},
		{
			"every suspect is cleared", serviceRevision, review.SoundBetter, kept(4),
			[]review.Suspect{suspect(review.SentenceMissing, review.FactKept), suspect(review.SentenceNegation, review.FactKept)}, []string{},
		},
		{
			"one sentence is gone", serviceRevision, review.SoundBetter, kept(50),
			[]review.Suspect{suspect(review.SentenceMissing, review.FactMissing), suspect(review.SentenceMissing, review.FactKept)},
			[]string{"1 sentence of the page is not in the revision"},
		},
		{
			"a suspect that nobody looked at counts as lost", serviceRevision, review.SoundBetter, kept(50),
			[]review.Suspect{suspect(review.SentenceNegation, "")}, []string{"1 sentence of the page loses a negation"},
		},
		{
			"sentences are gone and negations too", serviceRevision, review.SoundBetter, kept(50),
			[]review.Suspect{
				suspect(review.SentenceMissing, review.FactUnquoted), suspect(review.SentenceMissing, review.FactChanged),
				suspect(review.SentenceNegation, review.FactChanged), suspect(review.SentenceNegation, review.FactMissing),
			},
			[]string{"2 sentences of the page are not in the revision, and 2 sentences of the page lose a negation"},
		},
		{
			"a reworded sentence says something else", serviceRevision, review.SoundBetter, kept(50),
			[]review.Suspect{suspect(review.SentenceReworded, review.FactChanged), suspect(review.SentenceReworded, review.FactKept)},
			[]string{"1 sentence of the page says something else in the revision"},
		},
		{
			"each kind of lost sentence", serviceRevision, review.SoundBetter, kept(50),
			[]review.Suspect{
				suspect(review.SentenceReworded, review.FactChanged), suspect(review.SentenceReworded, review.FactMissing),
				suspect(review.SentenceMissing, review.FactMissing), suspect(review.SentenceNegation, review.FactChanged),
			},
			[]string{"1 sentence of the page is not in the revision, and 1 sentence of the page loses a negation, " +
				"and 2 sentences of the page say something else in the revision"},
		},
		{
			"everything fails", strings.Replace(dashed, "version 2.1", "version 2", 1), review.SoundWorse, withLoss(kept(4), 4),
			[]review.Suspect{suspect(review.SentenceMissing, review.FactMissing)},
			[]string{
				"the judge finds the revision more like a model's text than the page",
				"4 of 4 statements of the page are not kept", "1 sentence of the page is not in the revision",
				"the revision loses 1 numbers", "the revision adds 1 em dashes",
			},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			verdict, err := revisionOf(c, t, test.revised).Judge(review.RequireAll, test.sound, test.facts, test.sentences)
			c.Assert(err, qt.IsNil)
			c.Assert(verdict.Require, qt.Equals, "all")
			c.Assert(verdict.Reasons, qt.DeepEquals, test.reasons)
			c.Assert(verdict.Accepted, qt.Equals, len(test.reasons) == 0)
			// Under this requirement a revision is accepted exactly when both results pass.
			c.Assert(verdict.Kept.Passed && verdict.Sound.Passed, qt.Equals, verdict.Accepted)
		})
	}
}

func TestJudgeGivesTheTwoResultsApart(t *testing.T) {
	const noDifference = "the judge does not find the revision less like a model's text than the page"
	const morelikeAModel = "the judge finds the revision more like a model's text than the page"
	gone := []review.Suspect{suspect(review.SentenceMissing, review.FactMissing)}
	dashed := strings.Replace(serviceRevision, "failure, and the delay", "failure—the delay", 1)
	cases := []struct {
		name      string
		revised   string
		sound     string
		sentences []review.Suspect
		kept      []string
		heard     []string
	}{
		{"nothing is lost and the revision reads better", serviceRevision, review.SoundBetter, nil, []string{}, []string{}},
		{"nothing is lost and the judge hears no difference", serviceRevision, review.SoundSame, nil, []string{}, []string{noDifference}},
		{"nothing is lost and the revision reads worse", serviceRevision, review.SoundWorse, nil, []string{}, []string{morelikeAModel}},
		{
			"a sentence is gone and the revision reads better", serviceRevision, review.SoundBetter, gone,
			[]string{"1 sentence of the page is not in the revision"}, []string{},
		},
		{
			"a sentence is gone, a dash is added, and the judge hears no difference", dashed, review.SoundSame, gone,
			[]string{"1 sentence of the page is not in the revision", "the revision adds 1 em dashes"}, []string{noDifference},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			revision := revisionOf(c, t, test.revised)
			all, err := revision.Judge(review.RequireAll, test.sound, kept(50), test.sentences)
			c.Assert(err, qt.IsNil)
			onlyKept, err := revision.Judge(review.RequireKept, test.sound, kept(50), test.sentences)
			c.Assert(err, qt.IsNil)
			// The two results do not depend on what is required.
			for _, verdict := range []review.RevisionVerdict{all, onlyKept} {
				c.Assert(verdict.Kept, qt.DeepEquals, review.RevisionResult{Passed: len(test.kept) == 0, Reasons: test.kept})
				c.Assert(verdict.Sound, qt.DeepEquals, review.RevisionResult{Passed: len(test.heard) == 0, Reasons: test.heard})
			}
			// Everything is required: the reasons of the sound come first, then those of what is lost.
			c.Assert(all.Reasons, qt.DeepEquals, append(append([]string{}, test.heard...), test.kept...))
			c.Assert(all.Accepted, qt.Equals, len(test.heard)+len(test.kept) == 0)
			// Only nothing lost is required: the blind comparison neither rejects the revision nor gives a reason.
			c.Assert(onlyKept.Require, qt.Equals, "kept")
			c.Assert(onlyKept.Reasons, qt.DeepEquals, test.kept)
			c.Assert(onlyKept.Accepted, qt.Equals, len(test.kept) == 0)
		})
	}
}

func TestJudgeRefusesARequirementItDoesNotKnow(t *testing.T) {
	c := qt.New(t)
	for _, require := range []string{"", "sound", "ALL", "kept "} {
		verdict, err := revisionOf(c, t, serviceRevision).Judge(require, review.SoundBetter, kept(4), nil)
		c.Assert(err, qt.ErrorMatches, `unknown requirement ".*": use "all" or "kept"`, qt.Commentf("requirement %q", require))
		// A requirement that is not known accepts nothing.
		c.Assert(verdict.Accepted, qt.IsFalse)
	}
}

func TestLostReadsWhatFollowsAnUnclosedFenceAsCode(t *testing.T) {
	c := qt.New(t)
	// The fence is not closed, so the text after it is code to the end of the page. No sentence is read there, and
	// a change in it has to show as lost code.
	page := "# Deploy\n\nRun the deploy job after the migration has finished.\n\n```sh\nmake migrate\nmake deploy after the migration\n"
	kept, err := review.NewRevision(prepare(c, t, "page.md", page), prepare(c, t, "revised.md", page+"\n"))
	c.Assert(err, qt.IsNil)
	c.Assert(kept.Lost().Any(), qt.IsFalse)
	suspects, compared := kept.Suspects()
	c.Assert(suspects, qt.HasLen, 0)
	c.Assert(compared, qt.Equals, 2)

	changed := strings.Replace(page, "make deploy after the migration", "make deploy before the migration", 1)
	revision, err := review.NewRevision(prepare(c, t, "page.md", page), prepare(c, t, "revised.md", changed))
	c.Assert(err, qt.IsNil)
	c.Assert(revision.Lost().Code, qt.DeepEquals, []string{"make migrate"})
	suspects, _ = revision.Suspects()
	c.Assert(suspects, qt.HasLen, 0)
}
