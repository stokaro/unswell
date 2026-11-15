package cli_test

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
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
	letter := "A"
	if strings.Index(call.User, "The client does not retry") > strings.Index(call.User, "is not retried") {
		letter = "B"
	}
	return `{"more_like_a_model": "` + letter + `", "reason": "Its sentences have one length."}`
}

// revisionJudge scripts the model behind compare: the answer to the blind
// comparison, to the request for statements, and to the search for them in the
// revision. In the page itself it finds every statement.
func revisionJudge(sound func(chatCall) string, facts, kept string) func(chatCall, int) (int, string) {
	return scriptedJudge(sound, facts, comparedControl, kept)
}

func scriptedJudge(sound func(chatCall) string, facts, control, kept string) func(chatCall, int) (int, string) {
	return func(call chatCall, _ int) (int, string) {
		_, searched, _ := strings.Cut(call.User, "===== PAGE =====")
		switch {
		case strings.Contains(call.User, "Which version reads MORE like text written by a language model?"):
			return http.StatusOK, sound(call)
		case strings.Contains(call.User, "List the statements of the technical page"):
			return http.StatusOK, facts
		case strings.Contains(call.User, "===== STATEMENTS =====") && strings.Contains(searched, "The client does not retry"):
			return http.StatusOK, control
		case strings.Contains(call.User, "===== STATEMENTS ====="):
			return http.StatusOK, kept
		}
		return http.StatusBadRequest, "unexpected prompt"
	}
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
	c.Assert(report.SchemaVersion, qt.Equals, "unswell-compare-v1")
	c.Assert(report.Protocol, qt.Equals, "unswell-revision-v1")
	c.Assert(report.Page.Path, qt.Equals, "page.md")
	c.Assert(report.Revision.Path, qt.Equals, "revised.md")
	c.Assert(report.Sound.Verdict, qt.Equals, "better")
	c.Assert(report.Sound.PageFirst.Named, qt.Equals, "page")
	c.Assert(report.Sound.RevisionFirst.Named, qt.Equals, "page")
	c.Assert(report.Facts.Listed, qt.Equals, 3)
	c.Assert(report.Facts.Kept, qt.Equals, 3)
	c.Assert(report.Lost.Numbers, qt.HasLen, 0)
	c.Assert(report.EmDashes.Revision, qt.Equals, 0)
	c.Assert(report.Verdict.Accepted, qt.IsTrue)
	c.Assert(report.Verdict.Reasons, qt.HasLen, 0)

	calls := model.received()
	c.Assert(calls, qt.HasLen, 5)
	// The comparison is asked in both orders, and no prompt says which version is the revision.
	c.Assert(calls[0].User, qt.Contains, "===== VERSION A =====\n"+comparedPage)
	c.Assert(calls[1].User, qt.Contains, "===== VERSION A =====\n"+comparedRevision)
	for _, call := range calls[:2] {
		instructions, _, _ := strings.Cut(call.User, "===== VERSION A =====")
		c.Assert(strings.ToLower(instructions), qt.Not(qt.Contains), "revis")
	}
	// The statements come from the page, are looked for in the page itself, and then in the revision.
	c.Assert(calls[2].User, qt.Contains, "===== PAGE =====\n"+comparedPage)
	c.Assert(calls[3].User, qt.Contains, "===== STATEMENTS =====\n1. The client retries a request after a connection failure.\n")
	c.Assert(calls[3].User, qt.Contains, "===== PAGE =====\n"+comparedPage)
	c.Assert(calls[4].User, qt.Contains, "===== PAGE =====\n"+comparedRevision)
	c.Assert(report.Facts.Dropped, qt.HasLen, 0)
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
		"Numbers, code, and links: nothing lost.\n"+
		"Em dashes: 0 in the page, 0 in the revision.\n"+
		"Accepted.\n"+
		"Judge: judge at "+host+".\n"+
		"The judge did not see which text is the revision. It is a model: name one other than the model that made the revision.\n")
	c.Assert(outcome.stderr, qt.Contains, "compare: revised.md: accepted\n")
}

func TestCompareRejectsWhenTheJudgeFollowsTheOrder(t *testing.T) {
	c := qt.New(t)
	// A judge that always names version A names the page first and the revision second.
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
			model := newFakeModel(t, revisionJudge(test.sound, comparedFacts, test.kept))
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
	c.Assert(model.received(), qt.HasLen, 4)
}

func TestCompareWithoutAnAnswerIsAnErrorAndNoVerdict(t *testing.T) {
	cases := []struct {
		name    string
		answers func(chatCall, int) (int, string)
		problem string
	}{
		{
			"the judge does not choose",
			revisionJudge(func(chatCall) string { return "Both versions read well." }, comparedFacts, comparedKept),
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
