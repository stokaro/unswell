package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/review"
)

// reviewSchema names the JSON report of the review command.
const reviewSchema = "unswell-review-v1"

// maxReasonText bounds a model's reason in the text report. The JSON report
// keeps the reason whole.
const maxReasonText = 400

// reviewResult is the report of one review run.
type reviewResult struct {
	SchemaVersion string             `json:"schema_version"`
	Protocol      string             `json:"protocol"`
	Version       string             `json:"version"`
	Advisory      bool               `json:"advisory"`
	Reviewer      modelIdentity      `json:"reviewer"`
	Confirmer     *modelIdentity     `json:"confirmer,omitempty"`
	Summary       reviewSummary      `json:"summary"`
	Documents     []reviewedDocument `json:"documents"`
}

type reviewSummary struct {
	Documents    int `json:"documents"`
	Failed       int `json:"failed"`
	WithoutProse int `json:"without_prose"`
	Findings     int `json:"findings"`
	Unconfirmed  int `json:"unconfirmed"`
	Unbound      int `json:"unbound"`
}

// reviewedDocument separates what is reported from what was dropped. Findings
// are bound flags, confirmed when a confirming model is set. Unconfirmed holds
// bound flags the confirming model did not accept. Unbound holds flags that
// do not quote their item.
type reviewedDocument struct {
	Path        string             `json:"path"`
	SourceHash  string             `json:"source_hash,omitempty"`
	Items       int                `json:"items"`
	Flags       int                `json:"flags"`
	Findings    []reviewedFinding  `json:"findings"`
	Unconfirmed []reviewedFinding  `json:"unconfirmed"`
	Unbound     []review.Rejection `json:"unbound"`
	Error       string             `json:"error,omitempty"`
}

type reviewedFinding struct {
	review.Finding
	Confirmation *review.Verdict `json:"confirmation,omitempty"`
}

// incomplete reports why a run did not review what was selected, or nil. A
// selection in which no document has prose is an error, as it is for check: a
// run that reviewed nothing must not read as a clean result.
func (s reviewSummary) incomplete() error {
	switch {
	case s.Failed > 0:
		return fmt.Errorf("the review is incomplete: %d of %d documents failed", s.Failed, s.Documents)
	case s.WithoutProse == s.Documents:
		return fmt.Errorf("no selected document has prose to review")
	}
	return nil
}

func (d reviewedDocument) outcome() string {
	switch {
	case d.Error != "":
		return "failed: " + excerpt(d.Error, maxReasonText)
	case d.Items == 0:
		return "no prose to review"
	}
	return fmt.Sprintf("%s, %d flags dropped", plural(len(d.Findings), "finding"), len(d.Unconfirmed)+len(d.Unbound))
}

func plural(count int, noun string) string {
	if count == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", count, noun)
}

func newReviewResult(run *reviewRun, documents []reviewedDocument) reviewResult {
	result := reviewResult{SchemaVersion: reviewSchema, Protocol: review.Protocol, Version: unswell.Version,
		Advisory: true, Reviewer: run.reviewer.identity(), Documents: documents}
	if run.confirmer != nil {
		confirmer := run.confirmer.identity()
		result.Confirmer = &confirmer
	}
	result.Summary.Documents = len(documents)
	for _, entry := range documents {
		if entry.Error != "" {
			result.Summary.Failed++
		} else if entry.Items == 0 {
			result.Summary.WithoutProse++
		}
		result.Summary.Findings += len(entry.Findings)
		result.Summary.Unconfirmed += len(entry.Unconfirmed)
		result.Summary.Unbound += len(entry.Unbound)
	}
	return result
}

func reviewTargets(root string, specs, inputs []string) ([]reportTarget, error) {
	if len(specs) == 0 {
		specs = []string{"text:-"}
	}
	targets, err := reportTargets(root, specs, inputs)
	if err != nil {
		return nil, err
	}
	for _, target := range targets {
		if target.format != "text" && target.format != "json" {
			return nil, fmt.Errorf("review writes text or json reports, not %q", target.format)
		}
	}
	return targets, nil
}

func writeReviewReports(environment Environment, targets []reportTarget, result reviewResult, showUnconfirmed bool) error {
	var failures []error
	for _, target := range targets {
		var buffer bytes.Buffer
		err := writeReviewText(&buffer, result, showUnconfirmed)
		if target.format == "json" {
			buffer.Reset()
			encoder := json.NewEncoder(&buffer)
			encoder.SetIndent("", "  ")
			err = encoder.Encode(result)
		}
		if err == nil && target.path == "-" {
			_, err = environment.Out.Write(buffer.Bytes())
		} else if err == nil {
			err = atomicWrite(target.path, writeBytes(buffer.Bytes()))
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("write %s report: %w", target.format, err))
		}
	}
	return errors.Join(failures...)
}

// writeReviewText lists the findings and then says what was left out. With
// showUnconfirmed it also lists the flags the confirming model did not accept,
// marked as such, after the findings of their document.
func writeReviewText(writer io.Writer, result reviewResult, showUnconfirmed bool) error {
	var text strings.Builder
	for _, entry := range result.Documents {
		for _, finding := range entry.Findings {
			writeReviewFinding(&text, "review", finding)
		}
		if showUnconfirmed {
			for _, finding := range entry.Unconfirmed {
				writeReviewFinding(&text, "review, not confirmed", finding)
			}
		}
	}
	summary := result.Summary
	fmt.Fprintf(&text, "Reviewed %s: %s.\n", plural(summary.Documents-summary.Failed-summary.WithoutProse, "document"),
		plural(summary.Findings, "finding"))
	if summary.Unconfirmed+summary.Unbound > 0 {
		fmt.Fprintf(&text, "Dropped: %d flags the confirming model did not accept, %d flags that do not quote their item.\n",
			summary.Unconfirmed, summary.Unbound)
	}
	if summary.WithoutProse > 0 {
		fmt.Fprintf(&text, "Without prose to review: %s.\n", plural(summary.WithoutProse, "document"))
	}
	for _, entry := range result.Documents {
		if entry.Error != "" {
			fmt.Fprintf(&text, "FAILED %s: %s\n", entry.Path, excerpt(entry.Error, maxReasonText))
		}
	}
	fmt.Fprintf(&text, "Reviewer: %s.", result.Reviewer)
	if result.Confirmer != nil {
		fmt.Fprintf(&text, " Confirmed by: %s.\n", *result.Confirmer)
	} else {
		text.WriteString(" No second model confirmed these findings.\n")
	}
	text.WriteString("Findings are a model's judgment, bound to source bytes. They are advisory and never gate.\n")
	_, err := io.WriteString(writer, text.String())
	return err
}

func writeReviewFinding(text *strings.Builder, label string, finding reviewedFinding) {
	fmt.Fprintf(text, "%s:%d:%d: %s [%s] %s\n    %q\n", finding.Path, finding.Start.Line, finding.Start.Column, label,
		strings.Join(finding.Categories, ", "), excerpt(finding.Reason, maxReasonText),
		strings.Join(strings.Fields(finding.Quote), " "))
}
