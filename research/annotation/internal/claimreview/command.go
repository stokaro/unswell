package claimreview

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

//go:embed request.schema.json
var requestSchema []byte

type sourceInput struct {
	Path   string          `json:"path"`
	Format document.Format `json:"format"`
	Text   string          `json:"text"`
	Hash   string          `json:"sha256"`
}

type replayInput struct {
	Version    string          `json:"version"`
	Source     sourceInput     `json:"source"`
	Claims     []Specification `json:"claims"`
	Stages     []Stage         `json:"stages"`
	Approvals  []Approval      `json:"approvals"`
	IncludeRaw bool            `json:"include_raw"`
}

type replayOutput struct {
	Version            string               `json:"version"`
	SourceHash         string               `json:"source_sha256"`
	Units              []unit               `json:"units"`
	Protected          []document.Exclusion `json:"protected"`
	Claims             []Claim              `json:"claims"`
	Stages             []Account            `json:"stages"`
	Complete           bool                 `json:"complete"`
	EditorialQualified bool                 `json:"editorial_qualified"`
}

type unit struct {
	Block    int           `json:"block"`
	Span     document.Span `json:"span"`
	Excluded bool          `json:"excluded"`
}

// Run prepares or replays one local research inventory. It never calls a model.
// With no stages it returns original claim IDs and an incomplete accounting.
// Raw quotes, replacements, and review explanations require explicit include_raw.
func Run(ctx context.Context, reader io.Reader, writer io.Writer) error {
	input, err := readInput(ctx, reader)
	if err != nil {
		return err
	}
	doc, err := extract.Parse(ctx, document.Source{
		Name: input.Source.Path, Format: input.Source.Format, Bytes: []byte(input.Source.Text),
	}, extract.Options{})
	if err != nil {
		return err
	}
	if doc.Hash != input.Source.Hash {
		return fmt.Errorf("source hash does not match the supplied original")
	}
	inventory, err := Bind(ctx, doc, input.Claims)
	if err != nil {
		return err
	}
	output, err := replay(ctx, inventory, input)
	if err != nil {
		return err
	}
	_, err = commandio.Await(ctx, func() (struct{}, error) {
		return struct{}{}, json.NewEncoder(writer).Encode(output)
	})
	if err == nil && len(input.Stages) > 0 && !output.Complete {
		return fmt.Errorf("required claim accounting remains uncertain")
	}
	return err
}

func readInput(ctx context.Context, reader io.Reader) (replayInput, error) {
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return io.ReadAll(io.LimitReader(reader, (32<<20)+1))
	})
	if err != nil {
		return replayInput{}, err
	}
	var input replayInput
	err = jsoninput.Decode(ctx, data, 32<<20, &input, jsoninput.Limits{Array: maximumClaims, Object: 32})
	if err != nil {
		return replayInput{}, err
	}
	if err := jsoninput.Schema(data, requestSchema, "urn:unswell:editorial-claims:v1"); err != nil {
		return replayInput{}, err
	}
	if input.Version != Version || input.Claims == nil || input.Stages == nil || input.Approvals == nil ||
		len(input.Stages) > 32 || len(input.Approvals) > maximumClaims {
		return replayInput{}, fmt.Errorf("unsupported version or missing or excessive review inventory")
	}
	return input, nil
}

func replay(ctx context.Context, inventory Inventory, input replayInput) (replayOutput, error) {
	out := replayOutput{Version: Version, SourceHash: inventory.document.Hash, Claims: inventory.Claims(),
		Stages: make([]Account, 0, len(input.Stages)), Complete: len(input.Stages) > 0,
		Units: make([]unit, 0, len(inventory.document.Blocks)), Protected: inventory.document.Excluded}
	for _, block := range inventory.document.Blocks {
		out.Units = append(out.Units, unit{Block: block.ID, Span: block.Span, Excluded: block.Excluded})
	}
	seen := make(map[string]bool)
	usedApprovals := 0
	for _, stage := range input.Stages {
		if seen[stage.ID] {
			return replayOutput{}, fmt.Errorf("duplicate stage identity")
		}
		seen[stage.ID] = true
		approvals := make([]Approval, 0)
		for _, approval := range input.Approvals {
			if approval.StageID == stage.ID {
				approvals = append(approvals, approval)
			}
		}
		usedApprovals += len(approvals)
		account, err := inventory.Review(ctx, stage, approvals)
		if err != nil {
			return replayOutput{}, fmt.Errorf("stage %s: %w", stage.ID, err)
		}
		out.Complete = out.Complete && account.Complete
		out.Stages = append(out.Stages, account)
	}
	if usedApprovals != len(input.Approvals) {
		return replayOutput{}, fmt.Errorf("review approval belongs to an unknown stage")
	}
	if !input.IncludeRaw {
		redact(&out)
	}
	return out, nil
}

func redact(out *replayOutput) {
	redactClaims(out.Claims)
	for i := range out.Stages {
		stage := &out.Stages[i]
		redactClaims(stage.Claims)
		for j := range stage.Decisions {
			stage.Decisions[j].Reason = ""
		}
		for j := range stage.Edits {
			stage.Edits[j].Target.Quote = ""
			stage.Edits[j].Replacement = ""
		}
		for j := range stage.Approvals {
			stage.Approvals[j].Reason = ""
		}
		for j := range stage.Duplicates {
			stage.Duplicates[j].Reason = ""
		}
	}
}

func redactClaims(claims []Claim) {
	for i := range claims {
		claim := &claims[i]
		claim.Diagnostic, claim.Reason, claim.Suggestion = "", "", ""
		for j := range claim.Targets {
			claim.Targets[j].Quote = ""
		}
		for j := range claim.Support {
			claim.Support[j].Quote = ""
		}
	}
}
