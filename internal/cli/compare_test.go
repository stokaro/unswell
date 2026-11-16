package cli_test

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

const comparedPage = "# Retry policy\n\n" +
	"The client retries a request after a connection failure. The delay doubles after every attempt, up to 30 seconds.\n\n" +
	"The client does not retry a request that the server rejected. Run `make build` before the first start.\n"

const comparedRevision = "# Retry policy\n\n" +
	"The client retries a request after a connection failure, and the delay doubles after every attempt, up to\n" +
	"30 seconds.\n\n" +
	"A request that the server rejected is not retried. Run `make build` before the first start.\n"

const comparedFacts = `{"facts": ["The client retries a request after a connection failure.",
	"The delay doubles after every attempt, up to 30 seconds.", "The client does not retry a request that the server rejected."]}`

const comparedKept = `{"facts": [
	{"fact": 1, "kept": "yes", "quote": "The client retries a request after a connection failure"},
	{"fact": 2, "kept": "yes", "quote": "the delay doubles after every attempt, up to 30 seconds"},
	{"fact": 3, "kept": "yes", "quote": "A request that the server rejected is not retried"}]}`

// comparedControl finds each statement in the page itself.
const comparedControl = `{"facts": [
	{"fact": 1, "kept": "yes", "quote": "The client retries a request after a connection failure"},
	{"fact": 2, "kept": "yes", "quote": "The delay doubles after every attempt, up to 30 seconds"},
	{"fact": 3, "kept": "yes", "quote": "The client does not retry a request that the server rejected"}]}`

// pageNamed answers the blind comparison as a judge that hears the page as
// the one more like a model's text, whichever letter it has.
func pageNamed(call chatCall) string {
	return pageNamedBefore("is not retried")(call)
}

// pageNamedBefore is pageNamed for a revision that is known by other words.
func pageNamedBefore(revisionWords string) func(chatCall) string {
	return func(call chatCall) string {
		letter := "A"
		if strings.Index(call.User, "The client does not retry") > strings.Index(call.User, revisionWords) {
			letter = "B"
		}
		return `{"more_like_a_model": "` + letter + `", "reason": "Its sentences have one length."}`
	}
}

// revisionJudge scripts the model behind compare: the answer to the blind
// comparison, to the request for statements, and to the search for them in the
// revision. In the page itself it finds every statement.
func revisionJudge(sound func(chatCall) string, facts, kept string) func(chatCall, int) (int, string) {
	return scriptedJudge(sound, facts, comparedControl, kept)
}

func scriptedJudge(sound func(chatCall) string, facts, control, kept string) func(chatCall, int) (int, string) {
	return judgeScript{sound: sound, facts: facts, control: control, kept: kept}.answer
}

// judgeScript holds the answers of the model behind compare. again answers
// the repeated question about statements whose quote was not in the revision,
// and sentences the question about the sentences that the revision does not
// keep word for word; without it the judge finds each of them and quotes the
// first words of its passage. The page of every test has the sentence that
// marks the control pass.
type judgeScript struct {
	sound                                  func(chatCall) string
	facts, control, kept, again, sentences string
}

func (s judgeScript) answer(call chatCall, _ int) (int, string) {
	switch {
	case strings.Contains(call.User, "Which text reads MORE like text written by a language model?"):
		return http.StatusOK, s.sound(call)
	case strings.Contains(call.User, "List the statements of the technical page"):
		return http.StatusOK, s.facts
	case strings.Contains(call.User, "===== PAIR 1 =====") && s.sentences != "":
		return http.StatusOK, s.sentences
	case strings.Contains(call.User, "===== PAIR 1 ====="):
		return http.StatusOK, everySentenceFound(call.User)
	case strings.Contains(call.User, "===== STATEMENTS ====="):
		return http.StatusOK, s.lookup(call)
	}
	return http.StatusBadRequest, "unexpected prompt"
}

// passageLine is the line of a pair that holds the sentences of the revision.
var passageLine = regexp.MustCompile(`(?m)^Passage: (.*)$`)

// everySentenceFound answers the question about suspects as a judge that
// finds each sentence and quotes the first five words of its passage.
func everySentenceFound(prompt string) string {
	entries := []string{}
	for index, passage := range passageLine.FindAllStringSubmatch(prompt, -1) {
		words := strings.Fields(passage[1])
		quote, err := json.Marshal(strings.Join(words[:min(5, len(words))], " "))
		if err != nil {
			panic(err)
		}
		entries = append(entries, fmt.Sprintf(`{"sentence": %d, "kept": "yes", "quote": %s}`, index+1, quote))
	}
	return `{"sentences": [` + strings.Join(entries, ", ") + `]}`
}

// lookup answers a search for statements: in the page itself, once more for
// one statement, or in the revision.
func (s judgeScript) lookup(call chatCall) string {
	listed, searched, _ := strings.Cut(call.User, "===== PAGE =====")
	switch {
	case strings.Contains(searched, "The client does not retry"):
		return s.control
	case s.again != "" && !strings.Contains(listed, "\n2. "):
		return s.again
	}
	return s.kept
}

type compareReport struct {
	SchemaVersion string `json:"schema_version"`
	Protocol      string `json:"protocol"`
	Page          struct {
		Path string `json:"path"`
	} `json:"page"`
	Revision struct {
		Path string `json:"path"`
	} `json:"revision"`
	Sound *struct {
		Verdict   string `json:"verdict"`
		PageFirst struct {
			Named, Reason string
		} `json:"page_first"`
		RevisionFirst struct {
			Named, Reason string
		} `json:"revision_first"`
	} `json:"sound"`
	Facts *struct {
		Listed     int      `json:"listed"`
		Dropped    []string `json:"dropped"`
		Kept       int      `json:"kept"`
		Statements []struct {
			Text, Outcome, Quote string
		} `json:"statements"`
	} `json:"facts"`
	Sentences *struct {
		Compared int `json:"compared"`
		Lost     int `json:"lost"`
		Suspects []struct {
			Line                                  int
			Text, Reason, Revised, Outcome, Quote string
		} `json:"suspects"`
	} `json:"sentences"`
	Lost *struct {
		Numbers, Code, Links []string
	} `json:"lost"`
	EmDashes *struct {
		Page, Revision int
	} `json:"em_dashes"`
	Verdict *struct {
		Accepted bool     `json:"accepted"`
		Reasons  []string `json:"reasons"`
	} `json:"verdict"`
	Error string `json:"error"`
}

// comparedFiles writes the page and a revision of it and returns their directory.
func comparedFiles(c *qt.C, t *testing.T, revision string) string {
	t.Helper()
	root := t.TempDir()
	c.Assert(os.WriteFile(filepath.Join(root, "page.md"), []byte(comparedPage), 0o600), qt.IsNil)
	c.Assert(os.WriteFile(filepath.Join(root, "revised.md"), []byte(revision), 0o600), qt.IsNil)
	return root
}

func compareJSON(c *qt.C, t *testing.T, root string, model *fakeModel) (compareReport, reviewOutcome) {
	t.Helper()
	outcome := runCommand(t, "compare", root, "", "page.md", "revised.md", "--report", "json:-", "--quiet",
		"--endpoint", model.server.URL, "--model", "judge")
	var report compareReport
	c.Assert(json.Unmarshal([]byte(outcome.stdout), &report), qt.IsNil, qt.Commentf("stdout: %s\nstderr: %s", outcome.stdout, outcome.stderr))
	return report, outcome
}

func TestCompareAcceptsARevisionThatPassesEveryCheck(t *testing.T) {
	c := qt.New(t)
	model := newFakeModel(t, revisionJudge(pageNamed, comparedFacts, comparedKept))
	report, outcome := compareJSON(c, t, comparedFiles(c, t, comparedRevision), model)
	c.Assert(outcome.code, qt.Equals, 0, qt.Commentf("stderr: %s", outcome.stderr))
	c.Assert(report.SchemaVersion, qt.Equals, "unswell-compare-v2")
	c.Assert(report.Protocol, qt.Equals, "unswell-revision-v2")
	c.Assert(report.Page.Path, qt.Equals, "page.md")
	c.Assert(report.Revision.Path, qt.Equals, "revised.md")
	c.Assert(report.Sound.Verdict, qt.Equals, "better")
	c.Assert(report.Sound.PageFirst.Named, qt.Equals, "page")
	c.Assert(report.Sound.RevisionFirst.Named, qt.Equals, "page")
	c.Assert(report.Facts.Listed, qt.Equals, 3)
	c.Assert(report.Facts.Kept, qt.Equals, 3)
	// Three sentences stand in the revision word for word, two of them joined into one. The fourth is said in
	// another order, so the judge is asked about it and finds it.
	c.Assert(report.Sentences.Compared, qt.Equals, 4)
	c.Assert(report.Sentences.Suspects, qt.HasLen, 1)
	c.Assert(report.Sentences.Suspects[0].Reason, qt.Equals, "reworded")
	c.Assert(report.Sentences.Suspects[0].Text, qt.Equals, "The client does not retry a request that the server rejected.")
	c.Assert(report.Sentences.Suspects[0].Revised, qt.Equals, "A request that the server rejected is not retried.")
	c.Assert(report.Sentences.Suspects[0].Outcome, qt.Equals, "kept")
	c.Assert(report.Sentences.Lost, qt.Equals, 0)
	c.Assert(report.Lost.Numbers, qt.HasLen, 0)
	c.Assert(report.EmDashes.Revision, qt.Equals, 0)
	c.Assert(report.Verdict.Accepted, qt.IsTrue)
	c.Assert(report.Verdict.Reasons, qt.HasLen, 0)

	calls := model.received()
	c.Assert(calls, qt.HasLen, 6)
	// The comparison is asked in both orders, and no prompt says which text is the revision.
	c.Assert(calls[0].User, qt.Contains, "===== TEXT A =====\n"+comparedPage)
	c.Assert(calls[1].User, qt.Contains, "===== TEXT A =====\n"+comparedRevision)
	for _, call := range calls[:2] {
		instructions, _, _ := strings.Cut(call.User, "===== TEXT A =====")
		c.Assert(strings.ToLower(instructions), qt.Not(qt.Contains), "revis")
	}
	// The statements come from the page, are looked for in the page itself, and then in the revision.
	c.Assert(calls[2].User, qt.Contains, "===== PAGE =====\n"+comparedPage)
	c.Assert(calls[3].User, qt.Contains, "===== STATEMENTS =====\n1. The client retries a request after a connection failure.\n")
	c.Assert(calls[3].User, qt.Contains, "===== PAGE =====\n"+comparedPage)
	c.Assert(calls[4].User, qt.Contains, "===== PAGE =====\n"+comparedRevision)
	c.Assert(report.Facts.Dropped, qt.HasLen, 0)
	// The last question shows the reworded sentence beside its counterpart, and neither text as a page.
	c.Assert(calls[5].User, qt.Contains, "===== PAIR 1 =====\nSentence: The client does not retry a request that the server rejected.\n"+
		"Passage: A request that the server rejected is not retried.\n")
	c.Assert(calls[5].User, qt.Not(qt.Contains), "# Retry policy")
}

func TestCompareTextReportNamesEachCheck(t *testing.T) {
	c := qt.New(t)
	model := newFakeModel(t, revisionJudge(pageNamed, comparedFacts, comparedKept))
	root := comparedFiles(c, t, comparedRevision)
	outcome := runCommand(t, "compare", root, "", "page.md", "revised.md", "--endpoint", model.server.URL, "--model", "judge")
	c.Assert(outcome.code, qt.Equals, 0, qt.Commentf("stderr: %s", outcome.stderr))
	host := strings.TrimPrefix(model.server.URL, "http://")
	c.Assert(outcome.stdout, qt.Equals, ""+
		"page.md against its revision revised.md\n"+
		"Sound: the revision reads less like a model's text. The judge named the page in both orders.\n"+
		"    Its sentences have one length.\n"+
		"    Its sentences have one length.\n"+
		"Facts: 3 of 3 statements kept.\n"+
		"Sentences: 1 of 4 compared sentences of the page not in the revision word for word; the judge found each of them there.\n"+
		"Numbers, code, and links: nothing lost.\n"+
		"Em dashes: 0 in the page, 0 in the revision.\n"+
		"Accepted.\n"+
		"Judge: judge at "+host+".\n"+
		"The judge did not see which text is the revision. It is a model: name one other than the model that made the revision.\n")
	c.Assert(outcome.stderr, qt.Contains, "compare: revised.md: accepted\n")
}

func TestCompareRejectsWhenTheJudgeFollowsTheOrder(t *testing.T) {
	c := qt.New(t)
	// A judge that always names text A names the page first and the revision second.
	alwaysFirst := func(chatCall) string { return `{"more_like_a_model": "A", "reason": "The first one."}` }
	model := newFakeModel(t, revisionJudge(alwaysFirst, comparedFacts, comparedKept))
	root := comparedFiles(c, t, comparedRevision)
	outcome := runCommand(t, "compare", root, "", "page.md", "revised.md", "--quiet", "--endpoint", model.server.URL, "--model", "judge")
	c.Assert(outcome.code, qt.Equals, 1, qt.Commentf("stderr: %s", outcome.stderr))
	c.Assert(outcome.stdout, qt.Contains,
		"Sound: no difference. With the page first the judge named the page, and with the revision first the revision.\n")
	c.Assert(outcome.stdout, qt.Contains, "Rejected: the judge does not find the revision less like a model's text than the page.\n")
}

func TestCompareRejectsARevisionForEachKindOfDamage(t *testing.T) {
	revisionNamed := func(call chatCall) string {
		return strings.NewReplacer(`"A"`, `"B"`, `"B"`, `"A"`).Replace(pageNamed(call))
	}
	lostFact := strings.Replace(comparedKept, `{"fact": 3, "kept": "yes", "quote": "A request that the server rejected is not retried"}`,
		`{"fact": 3, "kept": "changed", "quote": ""}`, 1)
	unquoted := strings.Replace(comparedKept, "A request that the server rejected is not retried", "the server never retries", 1)
	cases := []struct {
		name     string
		revision string
		sound    func(chatCall) string
		kept     string
		reason   string
		line     string
	}{
		{
			"the judge prefers the page", comparedRevision, revisionNamed, comparedKept,
			"the judge finds the revision more like a model's text than the page",
			"Sound: the revision reads more like a model's text. The judge named the revision in both orders.\n",
		},
		{
			"a statement is changed", comparedRevision, pageNamed, lostFact,
			"1 of 3 statements of the page are not kept",
			"    changed: The client does not retry a request that the server rejected.\n",
		},
		{
			"a statement is called kept with words that are not in the revision", comparedRevision, pageNamed, unquoted,
			"1 of 3 statements of the page are not kept",
			"    unquoted: The client does not retry a request that the server rejected.\n",
		},
		{
			"a number is lost", strings.Replace(comparedRevision, "30 seconds", "half a minute", 1), pageNamed,
			strings.Replace(comparedKept, "up to 30 seconds", "up to", 1),
			"the revision loses 1 numbers", "Lost numbers: 30.\n",
		},
		{
			"code is lost", strings.Replace(comparedRevision, "`make build`", "the build", 1), pageNamed, comparedKept,
			"the revision loses 1 code", "Lost code: make build.\n",
		},
		{
			"an em dash is added", strings.Replace(comparedRevision, "failure, and the delay", "failure—the delay", 1), pageNamed,
			strings.Replace(comparedKept, "after a connection failure", "after a connection", 1),
			"the revision adds 1 em dashes", "Em dashes: 0 in the page, 1 in the revision.\n",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			// Asked once more about the statement without its words, the model gives the same words.
			script := judgeScript{sound: test.sound, facts: comparedFacts, control: comparedControl, kept: test.kept,
				again: `{"facts": [{"fact": 1, "kept": "yes", "quote": "the server never retries"}]}`}
			model := newFakeModel(t, script.answer)
			root := comparedFiles(c, t, test.revision)
			report, outcome := compareJSON(c, t, root, model)
			c.Assert(outcome.code, qt.Equals, 1, qt.Commentf("stderr: %s", outcome.stderr))
			c.Assert(report.Verdict.Accepted, qt.IsFalse)
			c.Assert(report.Verdict.Reasons, qt.DeepEquals, []string{test.reason})
			text := runCommand(t, "compare", root, "", "page.md", "revised.md", "--quiet", "--endpoint", model.server.URL, "--model", "judge")
			c.Assert(text.code, qt.Equals, 1)
			c.Assert(text.stdout, qt.Contains, test.line)
			c.Assert(text.stdout, qt.Contains, "Rejected: "+test.reason+".\n")
		})
	}
}

func TestCompareLeavesOutAStatementThatThePageItselfDoesNotState(t *testing.T) {
	c := qt.New(t)
	// The model lists a statement and then does not find it in the page it took it from. It is not looked for in the
	// revision, where it would be missing whatever the revision changed.
	control := strings.Replace(comparedControl,
		`{"fact": 2, "kept": "yes", "quote": "The delay doubles after every attempt, up to 30 seconds"}`,
		`{"fact": 2, "kept": "changed", "quote": ""}`, 1)
	kept := `{"facts": [
		{"fact": 1, "kept": "yes", "quote": "The client retries a request after a connection failure"},
		{"fact": 2, "kept": "yes", "quote": "A request that the server rejected is not retried"}]}`
	model := newFakeModel(t, scriptedJudge(pageNamed, comparedFacts, control, kept))
	root := comparedFiles(c, t, comparedRevision)
	report, outcome := compareJSON(c, t, root, model)
	c.Assert(outcome.code, qt.Equals, 0, qt.Commentf("stderr: %s", outcome.stderr))
	c.Assert(report.Facts.Listed, qt.Equals, 3)
	c.Assert(report.Facts.Dropped, qt.DeepEquals, []string{"The delay doubles after every attempt, up to 30 seconds."})
	c.Assert(report.Facts.Statements, qt.HasLen, 2)
	c.Assert(report.Facts.Kept, qt.Equals, 2)
	calls := model.received()
	c.Assert(calls[4].User, qt.Contains, "===== STATEMENTS =====\n1. The client retries a request after a connection failure.\n"+
		"2. The client does not retry a request that the server rejected.\n\n===== PAGE")
	text := runCommand(t, "compare", root, "", "page.md", "revised.md", "--quiet", "--endpoint", model.server.URL, "--model", "judge")
	c.Assert(text.stdout, qt.Contains,
		"Facts: 2 of 2 statements kept. Not counted: 1 statement that the judge did not find in the page itself.\n")
}

func TestCompareRejectsWhenThePageItselfStatesNoListedStatement(t *testing.T) {
	c := qt.New(t)
	control := `{"facts": [{"fact": 1, "kept": "missing"}, {"fact": 2, "kept": "missing"}, {"fact": 3, "kept": "changed"}]}`
	model := newFakeModel(t, scriptedJudge(pageNamed, comparedFacts, control, comparedKept))
	report, outcome := compareJSON(c, t, comparedFiles(c, t, comparedRevision), model)
	// Nothing was checked, and an unchecked revision does not pass.
	c.Assert(outcome.code, qt.Equals, 1)
	c.Assert(report.Facts.Dropped, qt.HasLen, 3)
	c.Assert(report.Verdict.Reasons, qt.DeepEquals, []string{"no statement of the page was checked"})
	// Two questions for the sound, two for the statements, and one for the reworded sentence.
	c.Assert(model.received(), qt.HasLen, 5)
}

func TestCompareWithoutAnAnswerIsAnErrorAndNoVerdict(t *testing.T) {
	cases := []struct {
		name    string
		answers func(chatCall, int) (int, string)
		problem string
	}{
		{
			"the judge does not choose",
			revisionJudge(func(chatCall) string { return "Both texts read well." }, comparedFacts, comparedKept),
			"blind comparison: the answer contains no JSON object",
		},
		{
			"the model lists no statement",
			revisionJudge(pageNamed, `{"facts": []}`, comparedKept),
			"statements of fact: the answer lists no statement",
		},
		{
			"the model skips a statement",
			revisionJudge(pageNamed, comparedFacts, `{"facts": [{"fact": 1, "kept": "yes", "quote": "The client retries a request"}]}`),
			"statements of fact: the answer has no entry for statement 2",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			model := newFakeModel(t, test.answers)
			report, outcome := compareJSON(c, t, comparedFiles(c, t, comparedRevision), model)
			// A check that did not complete is an error, never a pass and never a rejection.
			c.Assert(outcome.code, qt.Equals, 2)
			c.Assert(report.Error, qt.Equals, test.problem)
			c.Assert(report.Verdict, qt.IsNil)
			c.Assert(outcome.stderr, qt.Contains, "the comparison is incomplete: "+test.problem)
		})
	}
}

func TestCompareNeedsAPageAndADifferentRevision(t *testing.T) {
	c := qt.New(t)
	model := newFakeModel(t, revisionJudge(pageNamed, comparedFacts, comparedKept))
	root := comparedFiles(c, t, comparedPage)
	same := runCommand(t, "compare", root, "", "page.md", "revised.md", "--quiet", "--endpoint", model.server.URL, "--model", "judge")
	c.Assert(same.code, qt.Equals, 2)
	c.Assert(same.stderr, qt.Contains, "the revision is the page itself")
	one := runCommand(t, "compare", root, "", "page.md", "--quiet", "--endpoint", model.server.URL, "--model", "judge")
	c.Assert(one.code, qt.Equals, 2)
	c.Assert(one.stderr, qt.Contains, "accepts 2 arg(s), received 1")
	missing := runCommand(t, "compare", root, "", "page.md", "absent.md", "--quiet", "--endpoint", model.server.URL, "--model", "judge")
	c.Assert(missing.code, qt.Equals, 2)
	// Nothing is sent before both texts are read.
	c.Assert(model.received(), qt.HasLen, 0)
}

func TestCompareDoesNotWriteAReportOverItsInput(t *testing.T) {
	c := qt.New(t)
	model := newFakeModel(t, revisionJudge(pageNamed, comparedFacts, comparedKept))
	root := comparedFiles(c, t, comparedRevision)
	outcome := runCommand(t, "compare", root, "", "page.md", "revised.md", "--quiet", "--report", "json:revised.md",
		"--endpoint", model.server.URL, "--model", "judge")
	c.Assert(outcome.code, qt.Equals, 2)
	data, err := fs.ReadFile(os.DirFS(root), "revised.md")
	c.Assert(err, qt.IsNil)
	c.Assert(string(data), qt.Equals, comparedRevision)
	c.Assert(model.received(), qt.HasLen, 0)
}

// withoutRejected is the revision without the sentence about a rejected request.
var withoutRejected = strings.Replace(comparedRevision, "A request that the server rejected is not retried. ", "", 1)

// twoFacts, twoControl, and twoKept are the answers of a model that does not list the statement about a rejected
// request.
const (
	twoFacts = `{"facts": ["The client retries a request after a connection failure.",
		"The delay doubles after every attempt, up to 30 seconds."]}`
	twoControl = `{"facts": [
		{"fact": 1, "kept": "yes", "quote": "The client retries a request after a connection failure"},
		{"fact": 2, "kept": "yes", "quote": "The delay doubles after every attempt, up to 30 seconds"}]}`
	twoKept = `{"facts": [
		{"fact": 1, "kept": "yes", "quote": "The client retries a request after a connection failure"},
		{"fact": 2, "kept": "yes", "quote": "the delay doubles after every attempt, up to 30 seconds"}]}`
)

func TestCompareRejectsARemovedSentenceThatTheStatementsMiss(t *testing.T) {
	c := qt.New(t)
	// The model lists two statements and finds both. The removed sentence is in none of them.
	script := judgeScript{sound: pageNamedBefore("and the delay doubles"), facts: twoFacts, control: twoControl, kept: twoKept,
		sentences: `{"sentences": [{"sentence": 1, "kept": "missing", "quote": ""}]}`}
	model := newFakeModel(t, script.answer)
	root := comparedFiles(c, t, withoutRejected)
	report, outcome := compareJSON(c, t, root, model)
	c.Assert(outcome.code, qt.Equals, 1, qt.Commentf("stderr: %s", outcome.stderr))
	c.Assert(report.Facts.Kept, qt.Equals, 2)
	c.Assert(report.Sentences.Compared, qt.Equals, 4)
	c.Assert(report.Sentences.Lost, qt.Equals, 1)
	c.Assert(report.Sentences.Suspects, qt.HasLen, 1)
	suspect := report.Sentences.Suspects[0]
	c.Assert(suspect.Text, qt.Equals, "The client does not retry a request that the server rejected.")
	c.Assert(suspect.Reason, qt.Equals, "missing")
	c.Assert(suspect.Outcome, qt.Equals, "missing")
	c.Assert(suspect.Line, qt.Equals, 5)
	c.Assert(report.Verdict.Reasons, qt.DeepEquals, []string{"1 sentence of the page is not in the revision"})

	calls := model.received()
	c.Assert(calls, qt.HasLen, 6)
	// The judge sees the sentence beside the sentences that stand in its place, and not the revision as a page.
	c.Assert(calls[5].User, qt.Contains, "===== PAIR 1 =====\nSentence: The client does not retry a request that the server rejected.\n"+
		"Passage: The client retries a request after a connection failure, and the delay doubles")
	c.Assert(calls[5].User, qt.Not(qt.Contains), "# Retry policy")

	text := runCommand(t, "compare", root, "", "page.md", "revised.md", "--quiet", "--endpoint", model.server.URL, "--model", "judge")
	c.Assert(text.stdout, qt.Contains, ""+
		"Sentences: 1 of 4 compared sentences of the page not in the revision word for word; the judge did not find 1 of them there.\n"+
		"    line 5, not in the revision: The client does not retry a request that the server rejected.\n")
	c.Assert(text.stdout, qt.Contains, "Rejected: 1 sentence of the page is not in the revision.\n")
}

func TestCompareRejectsALostNegation(t *testing.T) {
	c := qt.New(t)
	turned := strings.Replace(comparedRevision, "is not retried", "is retried", 1)
	kept := strings.Replace(comparedKept, "is not retried", "is retried", 1)
	script := judgeScript{sound: pageNamedBefore("is retried"), facts: comparedFacts, control: comparedControl, kept: kept,
		sentences: `{"sentences": [{"sentence": 1, "kept": "changed", "quote": ""}]}`}
	model := newFakeModel(t, script.answer)
	root := comparedFiles(c, t, turned)
	report, outcome := compareJSON(c, t, root, model)
	c.Assert(outcome.code, qt.Equals, 1, qt.Commentf("stderr: %s", outcome.stderr))
	// The model called every statement kept. The count of negations did not agree.
	c.Assert(report.Facts.Kept, qt.Equals, 3)
	c.Assert(report.Sentences.Suspects, qt.HasLen, 1)
	c.Assert(report.Sentences.Suspects[0].Reason, qt.Equals, "negation")
	c.Assert(report.Sentences.Suspects[0].Revised, qt.Equals, "A request that the server rejected is retried.")
	c.Assert(report.Verdict.Reasons, qt.DeepEquals, []string{"1 sentence of the page loses a negation"})
	text := runCommand(t, "compare", root, "", "page.md", "revised.md", "--quiet", "--endpoint", model.server.URL, "--model", "judge")
	c.Assert(text.stdout, qt.Contains, "    line 5, a negation is gone: The client does not retry a request that the server rejected.\n")
}

func TestCompareClearsASuspectOnlyWithWordsOfItsPassage(t *testing.T) {
	// The revision says the sentence about a rejected request in other words, so the sentence has no counterpart.
	reworded := strings.Replace(comparedRevision, "A request that the server rejected is not retried.",
		"After the server refuses a call, the same call is never sent again.", 1)
	kept := strings.Replace(comparedKept, "A request that the server rejected is not retried", "the same call is never sent again", 1)
	cases := []struct {
		name, sentences, outcome, line string
		code                           int
	}{
		{
			"the judge quotes the passage", `{"sentences": [{"sentence": 1, "kept": "yes", "quote": "After the server refuses a call"}]}`, "kept",
			"Sentences: 1 of 4 compared sentences of the page not in the revision word for word; the judge found each of them there.\n", 0,
		},
		{
			// "Server" is a word of the sentence that the sentences around it do not have. This quote has none.
			"the judge quotes words without a word of the sentence",
			`{"sentences": [{"sentence": 1, "kept": "yes", "quote": "the same call is never sent again"}]}`, "unquoted",
			"    line 5, not in the revision: The client does not retry a request that the server rejected.\n", 1,
		},
		{
			"the judge quotes words that are not there",
			`{"sentences": [{"sentence": 1, "kept": "yes", "quote": "a refused request is not repeated"}]}`, "unquoted",
			"    line 5, not in the revision: The client does not retry a request that the server rejected.\n", 1,
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			script := judgeScript{sound: pageNamedBefore("never sent again"), facts: comparedFacts, control: comparedControl, kept: kept,
				sentences: test.sentences}
			model := newFakeModel(t, script.answer)
			root := comparedFiles(c, t, reworded)
			report, outcome := compareJSON(c, t, root, model)
			c.Assert(outcome.code, qt.Equals, test.code, qt.Commentf("stderr: %s", outcome.stderr))
			c.Assert(report.Sentences.Suspects, qt.HasLen, 1)
			c.Assert(report.Sentences.Suspects[0].Outcome, qt.Equals, test.outcome)
			c.Assert(report.Sentences.Lost, qt.Equals, test.code)
			text := runCommand(t, "compare", root, "", "page.md", "revised.md", "--quiet", "--endpoint", model.server.URL, "--model", "judge")
			c.Assert(text.stdout, qt.Contains, test.line)
		})
	}
}

func TestCompareAsksOnceMoreAboutAStatementThatWasNotFound(t *testing.T) {
	rejected := `{"fact": 3, "kept": "yes", "quote": "A request that the server rejected is not retried"}`
	found := `{"facts": [{"fact": 1, "kept": "yes", "quote": "A request that the server rejected is not retried"}]}`
	cases := []struct {
		name, first, again string
		kept, code         int
	}{
		{
			"a quote that is not in the revision, then the words", strings.Replace(comparedKept, "A request that the server rejected is not retried",
				"the server never retries", 1), found, 3, 0,
		},
		{"called changed, then found", strings.Replace(comparedKept, rejected, `{"fact": 3, "kept": "changed", "quote": ""}`, 1), found, 3, 0},
		{"called missing, then found", strings.Replace(comparedKept, rejected, `{"fact": 3, "kept": "missing", "quote": ""}`, 1), found, 3, 0},
		{
			"called changed both times", strings.Replace(comparedKept, rejected, `{"fact": 3, "kept": "changed", "quote": ""}`, 1),
			`{"facts": [{"fact": 1, "kept": "changed", "quote": ""}]}`, 2, 1,
		},
		{
			// The first outcome stays when the second answer does not find the statement either.
			"called changed, then missing", strings.Replace(comparedKept, rejected, `{"fact": 3, "kept": "changed", "quote": ""}`, 1),
			`{"facts": [{"fact": 1, "kept": "missing", "quote": ""}]}`, 2, 1,
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			script := judgeScript{sound: pageNamed, facts: comparedFacts, control: comparedControl, kept: test.first, again: test.again}
			model := newFakeModel(t, script.answer)
			report, outcome := compareJSON(c, t, comparedFiles(c, t, comparedRevision), model)
			c.Assert(outcome.code, qt.Equals, test.code, qt.Commentf("stderr: %s", outcome.stderr))
			c.Assert(report.Facts.Kept, qt.Equals, test.kept)
			if test.code == 1 {
				c.Assert(report.Facts.Statements[2].Outcome, qt.Equals, "changed")
			}
			calls := model.received()
			c.Assert(calls, qt.HasLen, 7)
			// Only the statement that was not found is asked again, in the revision.
			c.Assert(calls[5].User, qt.Contains, "===== STATEMENTS =====\n1. The client does not retry a request that the server rejected.\n\n"+
				"===== PAGE =====\n"+comparedRevision)
		})
	}
}

func TestCompareWithoutAnAnswerAboutASuspectIsAnError(t *testing.T) {
	c := qt.New(t)
	script := judgeScript{sound: pageNamed, facts: twoFacts, control: twoControl, kept: twoKept, sentences: `{"sentences": []}`}
	model := newFakeModel(t, script.answer)
	report, outcome := compareJSON(c, t, comparedFiles(c, t, withoutRejected), model)
	c.Assert(outcome.code, qt.Equals, 2)
	c.Assert(report.Error, qt.Equals, "sentences of the page: the answer has no entry for pair 1")
	c.Assert(report.Verdict, qt.IsNil)
}

func TestCompareRejectsARewordedSentenceThatSaysSomethingElse(t *testing.T) {
	c := qt.New(t)
	// One word is turned around. The sentence keeps its content words, so only the order of all its words tells,
	// and the judge has to say what it reads.
	turned := strings.Replace(comparedRevision, "after every attempt", "before every attempt", 1)
	kept := strings.Replace(comparedKept, "after every attempt", "before every attempt", 1)
	script := judgeScript{sound: pageNamed, facts: comparedFacts, control: comparedControl, kept: kept, sentences: `{"sentences": [
		{"sentence": 1, "kept": "changed", "quote": ""},
		{"sentence": 2, "kept": "yes", "quote": "A request that the server rejected is not retried"}]}`}
	model := newFakeModel(t, script.answer)
	root := comparedFiles(c, t, turned)
	report, outcome := compareJSON(c, t, root, model)
	c.Assert(outcome.code, qt.Equals, 1, qt.Commentf("stderr: %s", outcome.stderr))
	c.Assert(report.Facts.Kept, qt.Equals, 3)
	c.Assert(report.Sentences.Suspects, qt.HasLen, 2)
	c.Assert(report.Sentences.Suspects[0].Reason, qt.Equals, "reworded")
	c.Assert(report.Sentences.Suspects[0].Text, qt.Equals, "The delay doubles after every attempt, up to 30 seconds.")
	c.Assert(report.Sentences.Suspects[0].Outcome, qt.Equals, "changed")
	c.Assert(report.Sentences.Lost, qt.Equals, 1)
	c.Assert(report.Verdict.Reasons, qt.DeepEquals, []string{"1 sentence of the page says something else in the revision"})
	text := runCommand(t, "compare", root, "", "page.md", "revised.md", "--quiet", "--endpoint", model.server.URL, "--model", "judge")
	c.Assert(text.stdout, qt.Contains, ""+
		"Sentences: 2 of 4 compared sentences of the page not in the revision word for word; the judge did not find 1 of them there.\n"+
		"    line 3, says something else: The delay doubles after every attempt, up to 30 seconds.\n")
}

func TestCompareAsksAboutManySentencesInSeveralQuestions(t *testing.T) {
	c := qt.New(t)
	// The page has the sentence that marks it for the scripted judge and 22 more. The revision turns a word around
	// in each of the 22 and drops the first, so 23 sentences are not in it word for word.
	var page, revision strings.Builder
	page.WriteString("# Workers\n\nThe client does not retry a request that the server rejected.\n")
	revision.WriteString("# Workers\n")
	names := "amber basalt cedar delta ember flint garnet harbor indigo jasper kestrel lumen marble nectar onyx pebble quartz russet " +
		"saffron timber umber violet"
	for name := range strings.FieldsSeq(names) {
		fmt.Fprintf(&page, "\nThe %s worker drains the %s queue before the %s report is written.\n", name, name, name)
		fmt.Fprintf(&revision, "\nThe %s worker drains the %s queue after the %s report is written.\n", name, name, name)
	}
	root := t.TempDir()
	c.Assert(os.WriteFile(filepath.Join(root, "page.md"), []byte(page.String()), 0o600), qt.IsNil)
	c.Assert(os.WriteFile(filepath.Join(root, "revised.md"), []byte(revision.String()), 0o600), qt.IsNil)
	one := `{"facts": [{"fact": 1, "kept": "yes", "quote": "The amber worker drains the amber queue"}]}`
	script := judgeScript{sound: pageNamedBefore("queue after"), facts: `{"facts": ["The amber worker drains the amber queue."]}`,
		control: one, kept: one}
	model := newFakeModel(t, script.answer)
	report, outcome := compareJSON(c, t, root, model)
	c.Assert(outcome.code, qt.Equals, 1, qt.Commentf("stderr: %s", outcome.stderr))
	c.Assert(report.Sentences.Compared, qt.Equals, 23)
	c.Assert(report.Sentences.Suspects, qt.HasLen, 23)
	// The scripted judge finds every sentence that it is shown. The dropped one is cleared by words of its place
	// only when they carry a word of its own, and they do not.
	c.Assert(report.Sentences.Lost, qt.Equals, 1)
	c.Assert(report.Verdict.Reasons, qt.DeepEquals, []string{"1 sentence of the page is not in the revision"})
	questions := []int{}
	for _, call := range model.received() {
		if pairs := strings.Count(call.User, "===== PAIR "); pairs > 0 {
			questions = append(questions, pairs)
		}
	}
	// Twenty and three sentences go in two questions. The dropped one, found with words that do not clear it, is
	// asked about once more.
	c.Assert(questions, qt.DeepEquals, []int{20, 3, 1})
}

func TestCompareAsksOnceMoreAboutASentenceFoundWithoutItsWords(t *testing.T) {
	// The judge finds the reworded sentence and quotes the sentence of the page that it was asked about. Those words
	// are not in the passage, so the command asks once more about that sentence.
	first := `{"sentences": [{"sentence": 1, "kept": "yes", "quote": "does not retry a request that the server rejected"}]}`
	cases := []struct {
		name, second, outcome string
		code                  int
	}{
		{
			"then the words of the passage", `{"sentences": [{"sentence": 1, "kept": "yes", "quote": "that the server rejected is not retried"}]}`,
			"kept", 0,
		},
		{"then the same words again", first, "unquoted", 1},
		{"then another answer", `{"sentences": [{"sentence": 1, "kept": "changed", "quote": ""}]}`, "unquoted", 1},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			script := judgeScript{sound: pageNamed, facts: comparedFacts, control: comparedControl, kept: comparedKept}
			asked := 0
			model := newFakeModel(t, func(call chatCall, attempt int) (int, string) {
				if !strings.Contains(call.User, "===== PAIR 1 =====") {
					return script.answer(call, attempt)
				}
				asked++
				if asked == 1 {
					return http.StatusOK, first
				}
				return http.StatusOK, test.second
			})
			report, outcome := compareJSON(c, t, comparedFiles(c, t, comparedRevision), model)
			c.Assert(outcome.code, qt.Equals, test.code, qt.Commentf("stderr: %s", outcome.stderr))
			c.Assert(asked, qt.Equals, 2)
			c.Assert(report.Sentences.Suspects, qt.HasLen, 1)
			// The first outcome stays unless the second answer finds the sentence with words of its passage.
			c.Assert(report.Sentences.Suspects[0].Outcome, qt.Equals, test.outcome)
			c.Assert(report.Sentences.Lost, qt.Equals, test.code)
		})
	}
}

func TestCompareDoesNotAskAgainAboutASentenceThatTheJudgeCallsChanged(t *testing.T) {
	c := qt.New(t)
	script := judgeScript{sound: pageNamed, facts: comparedFacts, control: comparedControl, kept: comparedKept,
		sentences: `{"sentences": [{"sentence": 1, "kept": "changed", "quote": ""}]}`}
	model := newFakeModel(t, script.answer)
	report, outcome := compareJSON(c, t, comparedFiles(c, t, comparedRevision), model)
	c.Assert(outcome.code, qt.Equals, 1)
	c.Assert(report.Sentences.Suspects[0].Outcome, qt.Equals, "changed")
	// Two questions for the sound, three for the statements, and one for the sentence.
	c.Assert(model.received(), qt.HasLen, 6)
}
