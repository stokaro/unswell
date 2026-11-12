package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/review"
)

// ReviewProtocol names the request and answer shapes of the review tools.
const ReviewProtocol = review.Protocol

// MaxReviewFlags bounds one submitted review.
const MaxReviewFlags = review.MaxFlags

// reviewInstructions is the shared rubric followed by how this server takes
// the answer. The judgment belongs to the calling assistant.
const reviewInstructions = review.Instructions + `

Submit your answer with unswell_review_submit: one flag per item, with the item ID, at least one category,
a quote copied character for character from that item, and a one-sentence reason.`

// ReviewRequestInput asks for the items of one document.
type ReviewRequestInput struct {
	Source Source `json:"source" jsonschema:"The document to review. Its text is analyzed as data."`
}

// ReviewItem is one sentence, paragraph, or longer fragment of eligible prose.
type ReviewItem struct {
	ID   string `json:"id"   jsonschema:"Item identifier to cite in a flag."`
	Kind string `json:"kind" jsonschema:"sentence, paragraph, or fragment."`
	Line int    `json:"line" jsonschema:"One-based line of the item's first byte."`
	Text string `json:"text" jsonschema:"Original text of the item. Separate segments are joined by one space."`
}

// ReviewRequestOutput gives the calling assistant a rubric and exact items.
type ReviewRequestOutput struct {
	Protocol     string       `json:"protocol"`
	SourceHash   string       `json:"source_hash"  jsonschema:"SHA-256 of the supplied text."`
	Instructions string       `json:"instructions" jsonschema:"Rubric for the calling assistant."`
	Categories   []string     `json:"categories"`
	Items        []ReviewItem `json:"items"`
}

// ReviewFlag is one judgment of the calling assistant.
type ReviewFlag struct {
	Item       string   `json:"item"       jsonschema:"Item identifier from unswell_review_request."`
	Categories []string `json:"categories" jsonschema:"One or more defect categories."`
	Quote      string   `json:"quote"      jsonschema:"Text copied from the item that shows the defect."`
	Reason     string   `json:"reason"     jsonschema:"One sentence explaining the defect."`
}

// ReviewSubmitInput carries the same document and the assistant's flags.
type ReviewSubmitInput struct {
	Source Source       `json:"source" jsonschema:"The same document that was sent to unswell_review_request."`
	Flags  []ReviewFlag `json:"flags"  jsonschema:"Between 0 and 256 flags."`
}

// ReviewFinding is a flag bound to exact source coordinates.
type ReviewFinding struct {
	Item       string            `json:"item"`
	Categories []string          `json:"categories"`
	Quote      string            `json:"quote" jsonschema:"The quoted text exactly as it occurs in the source."`
	Reason     string            `json:"reason"`
	Path       string            `json:"path"`
	Span       document.Span     `json:"span" jsonschema:"Half-open byte range in the original source."`
	Start      document.Position `json:"start"`
	End        document.Position `json:"end"`
}

// ReviewRejection reports a flag that could not be bound to the source.
type ReviewRejection struct {
	Flag    ReviewFlag `json:"flag"`
	Problem string     `json:"problem"`
}

// ReviewSubmitOutput separates bound findings from rejected flags. Findings are
// the calling assistant's judgments. They are advisory and never gate.
type ReviewSubmitOutput struct {
	Protocol   string            `json:"protocol"`
	SourceHash string            `json:"source_hash"`
	Reviewer   string            `json:"reviewer" jsonschema:"Always the calling assistant. The server makes no editorial judgment here."`
	Accepted   []ReviewFinding   `json:"accepted"`
	Rejected   []ReviewRejection `json:"rejected"`
}

// ReviewCategories lists the defect categories of the editorial rubric.
func ReviewCategories() []string {
	return review.Categories()
}

type reviewer struct {
	prepare *review.Reviewer
}

func newReviewer(options Options) (*reviewer, error) {
	prepare, err := review.New(review.Options{Config: options.Config, ConfigBundle: options.ConfigBundle, NLP: options.NLP})
	if err != nil {
		return nil, err
	}
	return &reviewer{prepare: prepare}, nil
}

func (r *reviewer) document(ctx context.Context, input Source) (*review.Document, error) {
	return r.prepare.Prepare(ctx, document.Source{Name: input.Name, Format: input.Format, Bytes: []byte(input.Text)})
}

func (r *reviewer) request(
	ctx context.Context, _ *mcp.CallToolRequest, input ReviewRequestInput,
) (*mcp.CallToolResult, ReviewRequestOutput, error) {
	prepared, err := r.document(ctx, input.Source)
	if err != nil {
		return nil, ReviewRequestOutput{}, err
	}
	items := make([]ReviewItem, 0, len(prepared.Items))
	for _, item := range prepared.Items {
		items = append(items, ReviewItem(item))
	}
	return nil, ReviewRequestOutput{Protocol: ReviewProtocol, SourceHash: prepared.SourceHash,
		Instructions: reviewInstructions, Categories: ReviewCategories(), Items: items}, nil
}

func (r *reviewer) submit(
	ctx context.Context, _ *mcp.CallToolRequest, input ReviewSubmitInput,
) (*mcp.CallToolResult, ReviewSubmitOutput, error) {
	flags := make([]review.Flag, 0, len(input.Flags))
	for _, flag := range input.Flags {
		flags = append(flags, review.Flag(flag))
	}
	prepared, err := r.document(ctx, input.Source)
	if err != nil {
		return nil, ReviewSubmitOutput{}, err
	}
	accepted, rejected, err := prepared.Bind(flags)
	if err != nil {
		return nil, ReviewSubmitOutput{}, err
	}
	result := ReviewSubmitOutput{Protocol: ReviewProtocol, SourceHash: prepared.SourceHash, Reviewer: "calling assistant",
		Accepted: make([]ReviewFinding, 0, len(accepted)), Rejected: make([]ReviewRejection, 0, len(rejected))}
	for _, finding := range accepted {
		result.Accepted = append(result.Accepted, ReviewFinding(finding))
	}
	for _, rejection := range rejected {
		result.Rejected = append(result.Rejected, ReviewRejection{Flag: ReviewFlag(rejection.Flag), Problem: rejection.Problem})
	}
	return nil, result, nil
}
