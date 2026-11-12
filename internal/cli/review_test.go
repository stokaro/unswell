package cli_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/internal/cli"
)

const reviewDraft = "It is important to note that the client retries after a failure. " +
	"In order to start the service, run the installer.\n"

const (
	draftFlags = `{"flags": [
		{"item": "i001", "categories": ["empty_framing"], "quote": "It is important to note that", "reason": "Announces only."},
		{"item": "i002", "categories": ["wordiness"], "quote": "text that is absent", "reason": "Not a quote."},
		{"item": "i002", "categories": ["wordiness"], "quote": "In order to", "reason": "A shorter form exists."},
		{"item": "i002", "categories": ["wordiness"], "quote": "run the installer", "reason": "A second flag on one item."}]}`
	draftVerdicts = `{"reviews": [{"id": "f001", "justified": "yes", "reason": "Adds nothing."},
		{"id": "f002", "justified": "no", "reason": "The purpose clause is needed."}]}`
)

// chatCall is one request a fake model endpoint received.
type chatCall struct {
	Model         string
	Authorization string
	System        string
	User          string
}

// fakeModel is a chat completions endpoint whose answers a test scripts.
type fakeModel struct {
	server *httptest.Server
	lock   sync.Mutex
	calls  []chatCall
}

// newFakeModel answers each call with the status and text that answer
// returns. The number counts calls from one.
func newFakeModel(t *testing.T, answer func(call chatCall, number int) (int, string)) *fakeModel {
	t.Helper()
	model := &fakeModel{}
	model.server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var body struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		data, _ := io.ReadAll(request.Body)
		if err := json.Unmarshal(data, &body); err != nil || len(body.Messages) != 2 {
			http.Error(writer, "bad request", http.StatusBadRequest)
			return
		}
		call := chatCall{Model: body.Model, Authorization: request.Header.Get("Authorization"),
			System: body.Messages[0].Content, User: body.Messages[1].Content}
		model.lock.Lock()
		model.calls = append(model.calls, call)
		number := len(model.calls)
		model.lock.Unlock()
		status, text := answer(call, number)
		if status != http.StatusOK {
			http.Error(writer, text, status)
			return
		}
		reply := map[string]any{"choices": []map[string]any{{"finish_reason": "stop",
			"message": map[string]string{"role": "assistant", "content": text}}}}
		_ = json.NewEncoder(writer).Encode(reply)
	}))
	t.Cleanup(model.server.Close)
	return model
}

func (m *fakeModel) received() []chatCall {
	m.lock.Lock()
	defer m.lock.Unlock()
	return append([]chatCall(nil), m.calls...)
}

// scripted answers a review with flags and a confirmation with verdicts.
func scripted(call chatCall, _ int) (int, string) {
	if strings.Contains(call.User, "===== FLAGS =====") {
		return http.StatusOK, draftVerdicts
	}
	return http.StatusOK, draftFlags
}

type reviewOutcome struct {
	code   int
	stdout string
	stderr string
}

func runReview(t *testing.T, dir, stdin string, args ...string) reviewOutcome {
	t.Helper()
	var output, stderr bytes.Buffer
	environment := cli.Environment{Dir: dir, In: strings.NewReader(stdin), Out: &output, Err: &stderr,
		Getenv: func(name string) string {
			if name == "REVIEW_KEY" {
				return "secret-key"
			}
			return ""
		}}
	arguments := append([]string{"review", "--retry-delay", "1ms"}, args...)
	code := cli.Run(t.Context(), arguments, environment)
	return reviewOutcome{code: code, stdout: output.String(), stderr: stderr.String()}
}

type reviewReport struct {
	SchemaVersion string `json:"schema_version"`
	Advisory      bool   `json:"advisory"`
	Reviewer      struct {
		Transport, Model, Host, Command string
	} `json:"reviewer"`
	Confirmer *struct {
		Transport, Model, Host, Command string
	} `json:"confirmer"`
	Summary struct {
		Documents    int `json:"documents"`
		Failed       int `json:"failed"`
		WithoutProse int `json:"without_prose"`
		Findings     int `json:"findings"`
		Unconfirmed  int `json:"unconfirmed"`
		Unbound      int `json:"unbound"`
	} `json:"summary"`
	Documents []struct {
		Path     string `json:"path"`
		Items    int    `json:"items"`
		Flags    int    `json:"flags"`
		Error    string `json:"error"`
		Findings []struct {
			Quote string `json:"quote"`
			Span  struct {
				Start int `json:"start"`
				End   int `json:"end"`
			} `json:"span"`
			Confirmation *struct {
				Justified string `json:"justified"`
			} `json:"confirmation"`
		} `json:"findings"`
		Unconfirmed []struct {
			Quote        string `json:"quote"`
			Confirmation *struct {
				Justified string `json:"justified"`
			} `json:"confirmation"`
		} `json:"unconfirmed"`
		Unbound []struct {
			Problem string `json:"problem"`
		} `json:"unbound"`
	} `json:"documents"`
}

func decodeReview(c *qt.C, text string) reviewReport {
	c.Helper()
	var report reviewReport
	c.Assert(json.Unmarshal([]byte(text), &report), qt.IsNil, qt.Commentf("output: %s", text))
	return report
}

func TestReviewKeepsOnlyBoundAndConfirmedFlags(t *testing.T) {
	c := qt.New(t)
	model := newFakeModel(t, scripted)
	outcome := runReview(t, t.TempDir(), reviewDraft, "--stdin", "--filename", "draft.md", "--report", "json:-", "--quiet",
		"--endpoint", model.server.URL+"/v1/chat/completions", "--model", "first", "--api-key-env", "REVIEW_KEY",
		"--confirm-model", "second")
	c.Assert(outcome.code, qt.Equals, 0, qt.Commentf("stderr: %s", outcome.stderr))
	c.Assert(outcome.stderr, qt.Equals, "")
	report := decodeReview(c, outcome.stdout)
	c.Assert(report.SchemaVersion, qt.Equals, "unswell-review-v1")
	c.Assert(report.Advisory, qt.IsTrue)
	c.Assert(report.Reviewer.Model, qt.Equals, "first")
	c.Assert(report.Reviewer.Host, qt.Equals, strings.TrimPrefix(model.server.URL, "http://"))
	c.Assert(report.Confirmer.Model, qt.Equals, "second")
	c.Assert(report.Summary.Findings, qt.Equals, 1)
	c.Assert(report.Summary.Unconfirmed, qt.Equals, 1)
	c.Assert(report.Summary.Unbound, qt.Equals, 2)
	document := report.Documents[0]
	c.Assert(document.Path, qt.Equals, "draft.md")
	c.Assert(document.Items, qt.Equals, 2)
	c.Assert(document.Flags, qt.Equals, 4)
	finding := document.Findings[0]
	c.Assert(reviewDraft[finding.Span.Start:finding.Span.End], qt.Equals, "It is important to note that")
	c.Assert(finding.Confirmation.Justified, qt.Equals, "yes")
	c.Assert(document.Unconfirmed[0].Quote, qt.Equals, "In order to")
	c.Assert(document.Unconfirmed[0].Confirmation.Justified, qt.Equals, "no")
	c.Assert(document.Unbound[0].Problem, qt.Equals, "quote does not occur in one segment of the item")
	c.Assert(document.Unbound[1].Problem, qt.Equals, "duplicate item")

	calls := model.received()
	c.Assert(calls, qt.HasLen, 2)
	c.Assert(calls[0].Model, qt.Equals, "first")
	c.Assert(calls[0].Authorization, qt.Equals, "Bearer secret-key")
	c.Assert(calls[0].User, qt.Contains, "===== DOCUMENT =====\n"+reviewDraft)
	// The confirming model at the same endpoint inherits the address and the key.
	c.Assert(calls[1].Model, qt.Equals, "second")
	c.Assert(calls[1].Authorization, qt.Equals, "Bearer secret-key")
	c.Assert(calls[1].User, qt.Contains, `quoted text: "In order to"`)
	c.Assert(calls[1].User, qt.Not(qt.Contains), "A shorter form exists.")
	c.Assert(outcome.stdout+outcome.stderr, qt.Not(qt.Contains), "secret-key")
}

func TestReviewTextReportNamesTheJudgesAndStaysAdvisory(t *testing.T) {
	c := qt.New(t)
	model := newFakeModel(t, scripted)
	root := t.TempDir()
	c.Assert(os.WriteFile(filepath.Join(root, "draft.md"), []byte(reviewDraft), 0o600), qt.IsNil)
	c.Assert(os.WriteFile(filepath.Join(root, "code.md"), []byte("```sh\nmake build\n```\n"), 0o600), qt.IsNil)
	outcome := runReview(t, root, "", "draft.md", "code.md", "--endpoint", model.server.URL, "--model", "first")
	c.Assert(outcome.code, qt.Equals, 0, qt.Commentf("stderr: %s", outcome.stderr))
	host := strings.TrimPrefix(model.server.URL, "http://")
	c.Assert(outcome.stdout, qt.Equals, ""+
		"draft.md:1:1: review [empty_framing] Announces only.\n    \"It is important to note that\"\n"+
		"draft.md:1:66: review [wordiness] A shorter form exists.\n    \"In order to\"\n"+
		"Reviewed 1 document: 2 findings.\n"+
		"Dropped: 0 flags the confirming model did not accept, 2 flags that do not quote their item.\n"+
		"Without prose to review: 1 document.\n"+
		"Reviewer: first at "+host+". No second model confirmed these findings.\n"+
		"Findings are a model's judgment, bound to source bytes. They are advisory and never gate.\n")
	c.Assert(outcome.stderr, qt.Contains, "review: draft.md: 2 findings, 2 flags dropped\n")
	c.Assert(outcome.stderr, qt.Contains, "review: code.md: no prose to review\n")
	// A document without prose is never sent.
	c.Assert(model.received(), qt.HasLen, 1)
}

func TestReviewTextShowsUnconfirmedFlagsOnlyOnRequest(t *testing.T) {
	c := qt.New(t)
	model := newFakeModel(t, scripted)
	args := []string{"--stdin", "--filename", "draft.md", "--quiet", "--endpoint", model.server.URL, "--model", "first",
		"--confirm-model", "second"}
	outcome := runReview(t, t.TempDir(), reviewDraft, args...)
	c.Assert(outcome.code, qt.Equals, 0, qt.Commentf("stderr: %s", outcome.stderr))
	host := strings.TrimPrefix(model.server.URL, "http://")
	summary := "Reviewed 1 document: 1 finding.\n" +
		"Dropped: 1 flags the confirming model did not accept, 2 flags that do not quote their item.\n" +
		"Reviewer: first at " + host + ". Confirmed by: second at " + host + ".\n" +
		"Findings are a model's judgment, bound to source bytes. They are advisory and never gate.\n"
	confirmed := "draft.md:1:1: review [empty_framing] Announces only.\n    \"It is important to note that\"\n"
	c.Assert(outcome.stdout, qt.Equals, confirmed+summary)
	outcome = runReview(t, t.TempDir(), reviewDraft, append(args, "--show-unconfirmed")...)
	c.Assert(outcome.code, qt.Equals, 0, qt.Commentf("stderr: %s", outcome.stderr))
	c.Assert(outcome.stdout, qt.Equals, confirmed+
		"draft.md:1:66: review, not confirmed [wordiness] A shorter form exists.\n    \"In order to\"\n"+summary)
}

func TestReviewWritesReportsToFiles(t *testing.T) {
	c := qt.New(t)
	model := newFakeModel(t, scripted)
	root := t.TempDir()
	c.Assert(os.WriteFile(filepath.Join(root, "draft.md"), []byte(reviewDraft), 0o600), qt.IsNil)
	outcome := runReview(t, root, "", "draft.md", "--quiet", "--endpoint", model.server.URL, "--model", "first",
		"--report", "json:out/review.json", "--report", "text:-")
	c.Assert(outcome.code, qt.Equals, 0, qt.Commentf("stderr: %s", outcome.stderr))
	c.Assert(outcome.stdout, qt.Contains, "Reviewed 1 document: 2 findings.")
	data, err := fs.ReadFile(os.DirFS(root), "out/review.json")
	c.Assert(err, qt.IsNil)
	c.Assert(decodeReview(c, string(data)).Summary.Findings, qt.Equals, 2)
	// A report cannot replace a reviewed source.
	outcome = runReview(t, root, "", "draft.md", "--endpoint", model.server.URL, "--model", "first", "--report", "json:draft.md")
	c.Assert(outcome.code, qt.Equals, 2)
	c.Assert(outcome.stderr, qt.Contains, "report cannot overwrite source")
}

func TestReviewRepeatsACallThatCanRecover(t *testing.T) {
	c := qt.New(t)
	model := newFakeModel(t, func(call chatCall, number int) (int, string) {
		switch number {
		case 1:
			return http.StatusServiceUnavailable, "overloaded"
		case 2:
			return http.StatusOK, "I will review the document now."
		}
		return scripted(call, number)
	})
	outcome := runReview(t, t.TempDir(), reviewDraft, "--stdin", "--filename", "draft.md", "--report", "json:-", "--quiet",
		"--endpoint", model.server.URL, "--model", "first")
	c.Assert(outcome.code, qt.Equals, 0, qt.Commentf("stderr: %s", outcome.stderr))
	c.Assert(decodeReview(c, outcome.stdout).Summary.Findings, qt.Equals, 2)
	c.Assert(model.received(), qt.HasLen, 3)
}

func TestReviewFailsWhenAModelCannotAnswer(t *testing.T) {
	refusal := func(chatCall, int) (int, string) { return http.StatusOK, "I cannot help with that." }
	cases := []struct {
		name    string
		answer  func(chatCall, int) (int, string)
		args    []string
		calls   int
		problem string
		kept    int
	}{
		{"rejected key", func(chatCall, int) (int, string) { return http.StatusUnauthorized, "invalid key" },
			nil, 1, "answered 401: invalid key", 0},
		{"server error", func(chatCall, int) (int, string) { return http.StatusBadGateway, "upstream" },
			nil, 3, "answered 502: upstream", 0},
		{"no review in the answer", refusal, []string{"--attempts", "2"}, 2, "reviewer: the answer contains no JSON object", 0},
		{"empty answer", func(chatCall, int) (int, string) { return http.StatusOK, " " },
			[]string{"--attempts", "1"}, 1, "returned no text", 0},
		{"confirming model fails", func(call chatCall, number int) (int, string) {
			if strings.Contains(call.User, "===== FLAGS =====") {
				return http.StatusOK, `{"reviews": []}`
			}
			return scripted(call, number)
		}, []string{"--confirm-model", "second", "--attempts", "2"}, 3, "confirming model: the answer judges 0 of 2 flags", 2},
		{"document over the limit", scripted, []string{"--max-document-bytes", "20"}, 0, "--max-document-bytes is 20", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			model := newFakeModel(t, tc.answer)
			args := append([]string{"--stdin", "--filename", "draft.md", "--report", "json:-", "--quiet",
				"--endpoint", model.server.URL, "--model", "first"}, tc.args...)
			outcome := runReview(t, t.TempDir(), reviewDraft, args...)
			c.Assert(outcome.code, qt.Equals, 2)
			c.Assert(outcome.stderr, qt.Contains, "the review is incomplete: 1 of 1 documents failed")
			report := decodeReview(c, outcome.stdout)
			c.Assert(report.Summary.Failed, qt.Equals, 1)
			c.Assert(report.Summary.Findings, qt.Equals, 0)
			c.Assert(report.Summary.Unconfirmed, qt.Equals, tc.kept)
			c.Assert(report.Documents[0].Error, qt.Contains, tc.problem)
			c.Assert(model.received(), qt.HasLen, tc.calls)
		})
	}
}

func TestReviewKeepsTheKeyAndTerminalControlOutOfItsOutput(t *testing.T) {
	c := qt.New(t)
	model := newFakeModel(t, func(call chatCall, _ int) (int, string) {
		if strings.Contains(call.User, "Second page") {
			return http.StatusUnauthorized, "rejected key secret-key\x1b[2J for this model"
		}
		return http.StatusOK, `{"flags": [{"item": "i001", "categories": ["empty_framing"], ` +
			`"quote": "It is important to note that", "reason": "Announces\u001b[31m only.\nSecond line."}]}`
	})
	root := t.TempDir()
	c.Assert(os.WriteFile(filepath.Join(root, "a.md"), []byte(reviewDraft), 0o600), qt.IsNil)
	c.Assert(os.WriteFile(filepath.Join(root, "b.md"), []byte("Second page of the guide explains the retry policy.\n"), 0o600), qt.IsNil)
	outcome := runReview(t, root, "", "a.md", "b.md", "--jobs", "1", "--endpoint", model.server.URL, "--model", "first",
		"--api-key-env", "REVIEW_KEY", "--report", "text:-", "--report", "json:review.json")
	c.Assert(outcome.code, qt.Equals, 2)
	c.Assert(outcome.stdout, qt.Contains, "review [empty_framing] Announces [31m only. Second line.\n")
	c.Assert(outcome.stdout, qt.Contains, "answered 401: rejected key [key] [2J for this model")
	data, err := fs.ReadFile(os.DirFS(root), "review.json")
	c.Assert(err, qt.IsNil)
	for _, text := range []string{outcome.stdout, outcome.stderr, string(data)} {
		c.Assert(text, qt.Not(qt.Contains), "secret-key")
		c.Assert(text, qt.Not(qt.Contains), "\x1b")
	}
}

func TestReviewKeepsTheAddressQueryOutOfErrors(t *testing.T) {
	c := qt.New(t)
	closed := httptest.NewServer(http.NotFoundHandler())
	address := closed.URL + "/v1/chat/completions?api-key=query-secret"
	closed.Close()
	outcome := runReview(t, t.TempDir(), reviewDraft, "--stdin", "--filename", "draft.md", "--report", "json:-",
		"--endpoint", address, "--model", "first", "--attempts", "1")
	c.Assert(outcome.code, qt.Equals, 2)
	c.Assert(decodeReview(c, outcome.stdout).Documents[0].Error, qt.Contains, "request to 127.0.0.1:")
	c.Assert(outcome.stdout+outcome.stderr, qt.Not(qt.Contains), "query-secret")
}

func TestReviewDoesNotFollowARedirect(t *testing.T) {
	c := qt.New(t)
	target := newFakeModel(t, scripted)
	redirect := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, target.server.URL, http.StatusTemporaryRedirect)
	}))
	t.Cleanup(redirect.Close)
	outcome := runReview(t, t.TempDir(), reviewDraft, "--stdin", "--filename", "draft.md", "--quiet", "--report", "json:-",
		"--endpoint", redirect.URL, "--model", "first", "--api-key-env", "REVIEW_KEY")
	c.Assert(outcome.code, qt.Equals, 2)
	c.Assert(decodeReview(c, outcome.stdout).Documents[0].Error, qt.Contains, "answered 307")
	// The key and the document never reach the address the redirect names.
	c.Assert(target.received(), qt.HasLen, 0)
}

func TestReviewReportsOtherDocumentsWhenOneFails(t *testing.T) {
	c := qt.New(t)
	model := newFakeModel(t, func(call chatCall, number int) (int, string) {
		if strings.Contains(call.User, "Second page") {
			return http.StatusForbidden, "blocked"
		}
		return scripted(call, number)
	})
	root := t.TempDir()
	c.Assert(os.WriteFile(filepath.Join(root, "a.md"), []byte(reviewDraft), 0o600), qt.IsNil)
	c.Assert(os.WriteFile(filepath.Join(root, "b.md"), []byte("Second page of the guide explains the retry policy.\n"), 0o600), qt.IsNil)
	outcome := runReview(t, root, "", "a.md", "b.md", "--quiet", "--jobs", "1", "--endpoint", model.server.URL, "--model", "first")
	c.Assert(outcome.code, qt.Equals, 2)
	c.Assert(outcome.stdout, qt.Contains, "a.md:1:1: review [empty_framing]")
	c.Assert(outcome.stdout, qt.Contains, "Reviewed 1 document: 2 findings.")
	c.Assert(outcome.stdout, qt.Contains, "FAILED b.md: reviewer: ")
	c.Assert(outcome.stderr, qt.Contains, "the review is incomplete: 1 of 2 documents failed")
}

func TestReviewRejectsUnsafeOrIncompleteOptionsBeforeAnyCall(t *testing.T) {
	model := newFakeModel(t, scripted)
	endpoint := []string{"--endpoint", model.server.URL, "--model", "first"}
	cases := []struct {
		name    string
		args    []string
		problem string
	}{
		{"no model", nil, "name a model with --endpoint and --model, or a local program with --command"},
		{"endpoint without model", []string{"--endpoint", model.server.URL}, "name a model with --endpoint and --model"},
		{"missing key", append([]string{"--api-key-env", "ABSENT_KEY"}, endpoint...),
			"environment variable ABSENT_KEY named by --api-key-env is empty"},
		{"plain http elsewhere", []string{"--endpoint", "http://models.example.com/v1/chat/completions", "--model", "first"},
			"uses http outside this machine"},
		{"credentials in the address", []string{"--endpoint", "https://user:pass@models.example.com/v1", "--model", "first"},
			"must not carry credentials"},
		{"not a url", []string{"--endpoint", "models.example.com", "--model", "first"}, "must be an http or https URL"},
		{"command with endpoint", append([]string{"--command", "cat"}, endpoint...), "--command cannot be combined with --endpoint"},
		{"argument without command", append([]string{"--command-arg", "-p"}, endpoint...), "--command-arg requires --command"},
		{"missing program", []string{"--command", "no-such-review-program"}, `review command "no-such-review-program" cannot be started`},
		{"confirming command with model", append([]string{"--confirm-command", "cat", "--confirm-model", "second"}, endpoint...),
			"--confirm-command cannot be combined with --confirm-endpoint"},
		{"confirming key alone", append([]string{"--confirm-api-key-env", "REVIEW_KEY"}, endpoint...),
			"name a model with --confirm-endpoint and --confirm-model"},
		{"sarif report", append([]string{"--report", "sarif:-"}, endpoint...), `review writes text or json reports, not "sarif"`},
		{"no documents allowed", append([]string{"--max-documents", "0"}, endpoint...),
			"--max-documents and --max-document-bytes must be positive"},
		{"jobs", append([]string{"--jobs", "0"}, endpoint...), "--jobs must be between 1 and 16"},
		{"attempts", append([]string{"--attempts", "9"}, endpoint...), "--attempts must be between 1 and 5"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			args := append([]string{"--stdin", "--filename", "draft.md"}, tc.args...)
			outcome := runReview(t, t.TempDir(), reviewDraft, args...)
			c.Assert(outcome.code, qt.Equals, 2)
			c.Assert(outcome.stderr, qt.Contains, tc.problem)
			c.Assert(outcome.stdout, qt.Equals, "")
		})
	}
	qt.New(t).Assert(model.received(), qt.HasLen, 0)
}

func TestReviewFailsWhenNothingCanBeReviewed(t *testing.T) {
	c := qt.New(t)
	model := newFakeModel(t, scripted)
	outcome := runReview(t, t.TempDir(), "```sh\nmake build\n```\n", "--stdin", "--filename", "code.md", "--quiet",
		"--endpoint", model.server.URL, "--model", "first")
	c.Assert(outcome.code, qt.Equals, 2)
	c.Assert(outcome.stdout, qt.Contains, "Without prose to review: 1 document.")
	c.Assert(outcome.stderr, qt.Contains, "no selected document has prose to review")
	c.Assert(model.received(), qt.HasLen, 0)
}

func TestReviewReportsFinishedDocumentsWhenTheTimeLimitPasses(t *testing.T) {
	c := qt.New(t)
	release := make(chan struct{})
	model := newFakeModel(t, func(call chatCall, number int) (int, string) {
		if strings.Contains(call.User, "Second page") {
			<-release
		}
		return scripted(call, number)
	})
	// Registered after the server, so it runs first and lets the held request end.
	t.Cleanup(func() { close(release) })
	root := t.TempDir()
	c.Assert(os.WriteFile(filepath.Join(root, "a.md"), []byte(reviewDraft), 0o600), qt.IsNil)
	c.Assert(os.WriteFile(filepath.Join(root, "b.md"), []byte("Second page of the guide explains the retry policy.\n"), 0o600), qt.IsNil)
	// The limit is long enough for the first document on a slow machine and
	// ends the held request of the second one.
	outcome := runReview(t, root, "", "a.md", "b.md", "--quiet", "--jobs", "1", "--timeout", "8s", "--attempts", "1",
		"--endpoint", model.server.URL, "--model", "first")
	c.Assert(outcome.code, qt.Equals, 2)
	c.Assert(outcome.stdout, qt.Contains, "a.md:1:1: review [empty_framing]")
	c.Assert(outcome.stdout, qt.Contains, "FAILED b.md: reviewer: ")
	c.Assert(outcome.stderr, qt.Contains, "the review is incomplete: 1 of 2 documents failed")
}

func TestReviewRefusesMoreDocumentsThanTheLimit(t *testing.T) {
	c := qt.New(t)
	model := newFakeModel(t, scripted)
	root := t.TempDir()
	for _, name := range []string{"a.md", "b.md", "c.md"} {
		c.Assert(os.WriteFile(filepath.Join(root, name), []byte(reviewDraft), 0o600), qt.IsNil)
	}
	outcome := runReview(t, root, "", ".", "--max-documents", "2", "--endpoint", model.server.URL, "--model", "first")
	c.Assert(outcome.code, qt.Equals, 2)
	c.Assert(outcome.stderr, qt.Contains, "the selection holds 3 documents and --max-documents is 2")
	c.Assert(model.received(), qt.HasLen, 0)
}

// TestReviewHelperProgram is not a test. The review command starts this test
// binary as the local review program; it then answers one prompt and exits.
func TestReviewHelperProgram(t *testing.T) {
	mode := os.Getenv("UNSWELL_REVIEW_HELPER")
	if mode == "" {
		t.Skip("helper for the command transport tests")
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
	entries, err := os.ReadDir(".")
	if err != nil || len(entries) != 0 {
		fmt.Fprintln(os.Stderr, "the program must start in an empty directory")
		os.Exit(3)
	}
	answer := "Here is the review.\n" + draftFlags + "\n"
	switch {
	case mode == "fail":
		fmt.Fprintln(os.Stderr, "model quota exhausted")
		os.Exit(4)
	case mode == "silent":
		answer = ""
	case strings.Contains(string(prompt), "===== FLAGS ====="):
		answer = "```json\n" + draftVerdicts + "\n```\n"
	}
	if _, err := io.WriteString(os.Stdout, answer); err != nil {
		os.Exit(3)
	}
	os.Exit(0)
}

func helperProgram(c *qt.C, prefix string) []string {
	c.Helper()
	program, err := os.Executable()
	c.Assert(err, qt.IsNil)
	return []string{"--" + prefix + "command", program, "--" + prefix + "command-arg", "-test.run=^TestReviewHelperProgram$"}
}

func TestReviewRunsALocalProgramAsTheModel(t *testing.T) {
	c := qt.New(t)
	t.Setenv("UNSWELL_REVIEW_HELPER", "answer")
	args := append([]string{"--stdin", "--filename", "draft.md", "--report", "json:-", "--quiet"}, helperProgram(c, "")...)
	args = append(args, helperProgram(c, "confirm-")...)
	outcome := runReview(t, t.TempDir(), reviewDraft, args...)
	c.Assert(outcome.code, qt.Equals, 0, qt.Commentf("stderr: %s", outcome.stderr))
	report := decodeReview(c, outcome.stdout)
	c.Assert(report.Reviewer.Transport, qt.Equals, "command")
	program, err := os.Executable()
	c.Assert(err, qt.IsNil)
	c.Assert(report.Reviewer.Command, qt.Equals, filepath.Base(program))
	c.Assert(report.Confirmer.Transport, qt.Equals, "command")
	c.Assert(report.Summary.Findings, qt.Equals, 1)
	c.Assert(report.Summary.Unconfirmed, qt.Equals, 1)
	c.Assert(report.Documents[0].Findings[0].Quote, qt.Equals, "It is important to note that")
}

func TestReviewFailsWhenTheLocalProgramFails(t *testing.T) {
	cases := []struct{ mode, problem string }{
		{"fail", "exit status 4: model quota exhausted"},
		{"silent", "printed no answer"},
	}
	for _, tc := range cases {
		t.Run(tc.mode, func(t *testing.T) {
			c := qt.New(t)
			t.Setenv("UNSWELL_REVIEW_HELPER", tc.mode)
			args := append([]string{"--stdin", "--filename", "draft.md", "--report", "json:-", "--quiet", "--attempts", "2"},
				helperProgram(c, "")...)
			outcome := runReview(t, t.TempDir(), reviewDraft, args...)
			c.Assert(outcome.code, qt.Equals, 2)
			report := decodeReview(c, outcome.stdout)
			c.Assert(report.Summary.Failed, qt.Equals, 1)
			c.Assert(report.Documents[0].Error, qt.Contains, tc.problem)
		})
	}
}
