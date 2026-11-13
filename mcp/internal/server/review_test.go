package server_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell/mcp/internal/server"
)

const reviewText = "# Guide\n\n" +
	"It is important to note that the client retries after a failure. The retry\ninterval is 30 seconds.\n\n" +
	"Run `make build` before the first start of the service.\n"

func reviewRequest(c *qt.C, t *testing.T, session *mcp.ClientSession) server.ReviewRequestOutput {
	c.Helper()
	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_review_request",
		Arguments: server.ReviewRequestInput{Source: server.Source{Name: "guide.md", Text: reviewText}}})
	c.Assert(err, qt.IsNil)
	c.Assert(result.IsError, qt.IsFalse)
	return output[server.ReviewRequestOutput](c, result)
}

// verdicts sends the flags as items to revise, in the given order, and marks
// every other item of the request as acceptable.
func verdicts(items []server.ReviewItem, flags ...server.ReviewFlag) []server.ReviewVerdict {
	result := make([]server.ReviewVerdict, 0, len(items)+len(flags))
	flagged := map[string]bool{}
	for _, flag := range flags {
		flagged[flag.Item] = true
		result = append(result, revise(flag))
	}
	for _, item := range items {
		if !flagged[item.ID] {
			result = append(result, server.ReviewVerdict{Item: item.ID, Verdict: "ok"})
		}
	}
	return result
}

func revise(flag server.ReviewFlag) server.ReviewVerdict {
	return server.ReviewVerdict{Item: flag.Item, Verdict: "revise", Categories: flag.Categories, Quote: flag.Quote, Reason: flag.Reason}
}

func reviewSubmit(c *qt.C, t *testing.T, session *mcp.ClientSession, text string, flags ...server.ReviewFlag) server.ReviewSubmitOutput {
	c.Helper()
	items := reviewRequest(c, t, session).Items
	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_review_submit",
		Arguments: server.ReviewSubmitInput{Source: server.Source{Name: "guide.md", Text: text}, Items: verdicts(items, flags...)}})
	c.Assert(err, qt.IsNil)
	c.Assert(result.IsError, qt.IsFalse)
	return output[server.ReviewSubmitOutput](c, result)
}

func itemContaining(c *qt.C, items []server.ReviewItem, text string) server.ReviewItem {
	c.Helper()
	for _, item := range items {
		if strings.Contains(item.Text, text) {
			return item
		}
	}
	c.Fatalf("no item contains %q in %+v", text, items)
	return server.ReviewItem{}
}

func TestReviewRequestListsSentencesWithTheRubric(t *testing.T) {
	c := qt.New(t)
	session := connect(c, t.Context(), server.Options{})
	out := reviewRequest(c, t, session)
	c.Assert(out.Protocol, qt.Equals, server.ReviewProtocol)
	c.Assert(out.SourceHash, qt.HasLen, 64)
	c.Assert(out.Categories, qt.DeepEquals, server.ReviewCategories())
	c.Assert(out.Instructions, qt.Contains, "unswell_review_submit")
	first := itemContaining(c, out.Items, "It is important to note")
	c.Assert(first.Kind, qt.Equals, "sentence")
	c.Assert(first.Line, qt.Equals, 3)
	second := itemContaining(c, out.Items, "interval is 30 seconds")
	c.Assert(second.ID, qt.Not(qt.Equals), first.ID)
	for index, item := range out.Items {
		c.Assert(item.ID, qt.Not(qt.Equals), "")
		if index > 0 {
			c.Assert(item.ID > out.Items[index-1].ID, qt.IsTrue)
		}
	}
}

func TestReviewSubmitBindsQuotesToOriginalBytes(t *testing.T) {
	c := qt.New(t)
	session := connect(c, t.Context(), server.Options{})
	items := reviewRequest(c, t, session).Items
	first := itemContaining(c, items, "It is important to note")
	second := itemContaining(c, items, "interval is 30 seconds")
	out := reviewSubmit(c, t, session, reviewText,
		server.ReviewFlag{Item: first.ID, Categories: []string{"empty_framing"},
			Quote: "It is important to note that", Reason: "The opening only announces the statement."},
		// The quote was copied with a space where the source wraps the line.
		server.ReviewFlag{Item: second.ID, Categories: []string{"wordiness"},
			Quote: "The retry interval", Reason: "Example of a quote across a line break."})
	c.Assert(out.Reviewer, qt.Equals, "calling assistant")
	c.Assert(out.Rejected, qt.HasLen, 0)
	c.Assert(out.Accepted, qt.HasLen, 2)
	finding := out.Accepted[0]
	c.Assert(reviewText[finding.Span.Start:finding.Span.End], qt.Equals, "It is important to note that")
	c.Assert(finding.Quote, qt.Equals, "It is important to note that")
	c.Assert(finding.Path, qt.Equals, "guide.md")
	c.Assert(finding.Start.Line, qt.Equals, 3)
	c.Assert(finding.Start.Column, qt.Equals, 1)
	wrapped := out.Accepted[1]
	c.Assert(reviewText[wrapped.Span.Start:wrapped.Span.End], qt.Equals, "The retry\ninterval")
	c.Assert(wrapped.End.Line, qt.Equals, 4)
}

func TestReviewSubmitRejectsFlagsItCannotBind(t *testing.T) {
	c := qt.New(t)
	session := connect(c, t.Context(), server.Options{})
	items := reviewRequest(c, t, session).Items
	first := itemContaining(c, items, "It is important to note")
	command := itemContaining(c, items, "before the first start")
	reason := "A reason."
	out := reviewSubmit(c, t, session, reviewText,
		server.ReviewFlag{Item: "i999", Categories: []string{"wordiness"}, Quote: "It is", Reason: reason},
		server.ReviewFlag{Item: first.ID, Categories: []string{}, Quote: "It is", Reason: reason},
		server.ReviewFlag{Item: first.ID, Categories: []string{"wordiness"}, Quote: "text that is absent", Reason: reason},
		server.ReviewFlag{Item: first.ID, Categories: []string{"wordiness"}, Quote: "It is", Reason: " "},
		// A quote must not cross protected inline code.
		server.ReviewFlag{Item: command.ID, Categories: []string{"wordiness"}, Quote: "Run make build before", Reason: reason},
		server.ReviewFlag{Item: first.ID, Categories: []string{"wordiness"}, Quote: "It is", Reason: reason},
		server.ReviewFlag{Item: first.ID, Categories: []string{"wordiness"}, Quote: "the client", Reason: reason})
	c.Assert(out.Accepted, qt.HasLen, 1)
	problems := make([]string, 0, len(out.Rejected))
	for _, rejection := range out.Rejected {
		problems = append(problems, rejection.Problem)
	}
	c.Assert(problems, qt.DeepEquals, []string{"unknown item", "no category",
		"quote does not occur in one segment of the item", "no reason",
		"quote does not occur in one segment of the item", "duplicate item"})
}

func TestReviewSubmitRejectsAnUnknownCategoryAtTheSchema(t *testing.T) {
	c := qt.New(t)
	session := connect(c, t.Context(), server.Options{})
	items := reviewRequest(c, t, session).Items
	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_review_submit",
		Arguments: server.ReviewSubmitInput{Source: server.Source{Name: "guide.md", Text: reviewText},
			Items: verdicts(items, server.ReviewFlag{Item: items[0].ID, Categories: []string{"authorship"}, Quote: "It", Reason: "A reason."})}})
	if err == nil {
		c.Assert(result.IsError, qt.IsTrue)
	}
}

func TestReviewSubmitRefusesASubmissionThatSkipsItems(t *testing.T) {
	c := qt.New(t)
	session := connect(c, t.Context(), server.Options{})
	items := reviewRequest(c, t, session).Items
	c.Assert(len(items) >= 3, qt.IsTrue)
	for _, tc := range []struct {
		name  string
		items []server.ReviewVerdict
	}{
		{"nothing", nil},
		{"one verdict", []server.ReviewVerdict{{Item: items[0].ID, Verdict: "ok"}}},
		{"only a flag", []server.ReviewVerdict{{Item: items[0].ID, Verdict: "revise", Categories: []string{"empty_framing"},
			Quote: "It is important to note that", Reason: "Announces only."}}},
		{"one item many times", []server.ReviewVerdict{{Item: items[0].ID, Verdict: "ok"}, {Item: items[0].ID, Verdict: "ok"},
			{Item: items[0].ID, Verdict: "ok"}, {Item: items[0].ID, Verdict: "ok"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_review_submit",
				Arguments: server.ReviewSubmitInput{Source: server.Source{Name: "guide.md", Text: reviewText}, Items: tc.items}})
			if err == nil {
				c.Assert(result.IsError, qt.IsTrue)
				c.Assert(result.Content[0].(*mcp.TextContent).Text, qt.Contains, "send a verdict for every item")
			}
		})
	}
}

func TestReviewSubmitAcceptsADocumentWithNothingToRevise(t *testing.T) {
	c := qt.New(t)
	session := connect(c, t.Context(), server.Options{})
	out := reviewSubmit(c, t, session, reviewText)
	c.Assert(out.Accepted, qt.HasLen, 0)
	c.Assert(out.Rejected, qt.HasLen, 0)
}

func TestReviewSubmitRejectsAnUnknownVerdictAtTheSchema(t *testing.T) {
	c := qt.New(t)
	session := connect(c, t.Context(), server.Options{})
	items := reviewRequest(c, t, session).Items
	all := verdicts(items)
	all[0].Verdict = "fine"
	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_review_submit",
		Arguments: server.ReviewSubmitInput{Source: server.Source{Name: "guide.md", Text: reviewText}, Items: all}})
	if err == nil {
		c.Assert(result.IsError, qt.IsTrue)
	}
}

func TestReviewRequestReturnsNoItemsWithoutProse(t *testing.T) {
	c := qt.New(t)
	session := connect(c, t.Context(), server.Options{})
	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_review_request",
		Arguments: server.ReviewRequestInput{Source: server.Source{Name: "code.md", Text: "```sh\nmake build\n```\n"}}})
	c.Assert(err, qt.IsNil)
	c.Assert(result.IsError, qt.IsFalse)
	out := output[server.ReviewRequestOutput](c, result)
	c.Assert(out.Items, qt.HasLen, 0)
	c.Assert(out.SourceHash, qt.HasLen, 64)
}
