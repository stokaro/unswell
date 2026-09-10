// Package report renders a completed result without analyzing or reading sources.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/stokaro/unswell"
)

// Options affects presentation only. JSON and SARIF always retain all findings.
type Options struct {
	MinSeverity string
	MaxFindings int
}

// Formats returns the supported report names in a stable order. Write accepts
// exactly these names and rejects every other value.
func Formats() []string { return []string{"text", "json", "sarif", "html", "markdown"} }

// Write serializes a saved result to one of text, json, sarif, html, or markdown.
// It returns every serialization and I/O error to the caller.
func Write(writer io.Writer, format string, result unswell.RunResult, options Options) error {
	if result.SchemaVersion != unswell.SchemaVersion {
		return fmt.Errorf("unsupported result schema %q", result.SchemaVersion)
	}
	if err := validateFeatures(result); err != nil {
		return err
	}
	switch format {
	case "json":
		return encodeJSON(writer, result)
	case "sarif":
		return sarif(writer, result)
	case "text":
		return text(writer, result, options)
	case "markdown":
		return markdown(writer, result, options)
	case "html":
		return htmlReport(writer, result, options)
	default:
		return fmt.Errorf("unknown report format %q", format)
	}
}

// MaxBytes bounds one saved result that Read accepts. A repository scan
// that saves every prepared and activation feature with its source text
// grows with the repository; the bound keeps the reader finite without
// cutting such a scan short.
const MaxBytes = 256 << 20

// Read decodes one bounded, versioned saved result. Source files are never read.
func Read(reader io.Reader) (unswell.RunResult, error) {
	var result unswell.RunResult
	data, err := io.ReadAll(io.LimitReader(reader, MaxBytes+1))
	if err != nil {
		return result, err
	}
	if len(data) > MaxBytes {
		return result, fmt.Errorf("saved result exceeds %d MiB", MaxBytes>>20)
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return result, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return result, fmt.Errorf("expected exactly one saved result")
	}
	if result.SchemaVersion != unswell.SchemaVersion {
		return result, fmt.Errorf("unsupported result schema %q", result.SchemaVersion)
	}
	if result.Status != "complete" && result.Status != "incomplete" {
		return result, fmt.Errorf("invalid result status")
	}
	if result.Status == "incomplete" && result.Gate.Passed {
		return result, fmt.Errorf("incomplete result cannot pass policy")
	}
	return result, validateSaved(data, result)
}

func encodeJSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func visible(result unswell.RunResult, options Options) []unswell.Finding {
	findings := make([]unswell.Finding, 0)
	for _, finding := range result.Findings {
		if rank(finding.Severity) < rank(options.MinSeverity) {
			continue
		}
		if options.MaxFindings > 0 && len(findings) >= options.MaxFindings {
			break
		}
		findings = append(findings, finding)
	}
	return findings
}

func rank(severity string) int {
	switch severity {
	case "error":
		return 3
	case "warning":
		return 2
	case "note":
		return 1
	}
	return 0
}

func verdict(result unswell.RunResult) string {
	if result.Status != "complete" {
		return "INCOMPLETE"
	}
	if result.Manifest.NoGate {
		return "ADVISORY"
	}
	if result.Gate.Passed {
		return "PASS"
	}
	return "FAIL"
}

func terminal(text string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' {
			return ' '
		}
		if unicode.IsControl(r) {
			return '�'
		}
		if unicode.In(r, unicode.Cf) {
			return '�'
		}
		return r
	}, text)
}

func text(writer io.Writer, result unswell.RunResult, options Options) error {
	var output strings.Builder
	for _, finding := range visible(result, options) {
		fmt.Fprintf(
			&output,
			"%s:%d:%d: %s [%s] %s\n",
			terminal(finding.Primary.Path),
			finding.Primary.Start.Line,
			finding.Primary.Start.Column,
			finding.Severity,
			finding.RuleID,
			terminal(findingMessage(finding)),
		)
		if finding.Primary.Snippet != "" {
			fmt.Fprintf(&output, "  %s\n", terminal(finding.Primary.Snippet))
		}
		for _, location := range finding.Related {
			fmt.Fprintf(&output, "  related: %s:%d:%d\n", terminal(location.Path), location.Start.Line, location.Start.Column)
		}
	}
	fmt.Fprintf(
		&output,
		"%s: %d documents, %d findings; analysis %s.\n",
		verdict(result),
		len(result.Documents),
		len(result.Findings),
		result.Status,
	)
	for _, reason := range result.Gate.Reasons {
		fmt.Fprintf(&output, "  gate: %s: %s\n", terminal(reason.Path), terminal(reason.Message))
	}
	for _, failure := range result.Errors {
		fmt.Fprintf(&output, "  error: %s: %s\n", terminal(failure.Path), terminal(failure.Message))
	}
	for _, line := range append(auditLines(result), featureLines(result)...) {
		fmt.Fprintf(&output, "%s\n", terminal(line))
	}
	writeModelLines(&output, result, terminal, "Revision probability: unavailable (no calibrated alpha model).\n",
		"Revision probability: %s\n", "Origin estimate: %s\n")
	if len(result.Manifest.ConfigSources) > 1 || len(result.Manifest.ConfigOverrides) > 0 {
		fmt.Fprintf(&output, "Configuration: %s (%s).\n", terminal(result.Manifest.ConfigHash), terminal(result.Manifest.ConfigIdentity))
	}
	_, err := io.WriteString(writer, output.String())
	return err
}

func markdown(writer io.Writer, result unswell.RunResult, options Options) error {
	var output strings.Builder
	fmt.Fprintf(
		&output,
		"**Unswell: %s** — %d documents, %d findings.\n\n",
		verdict(result),
		len(result.Documents),
		len(result.Findings),
	)
	output.WriteString("| Location | Rule | Finding |\n| --- | --- | --- |\n")
	findings := visible(result, options)
	for _, finding := range findings {
		fmt.Fprintf(
			&output,
			"| %s:%d:%d | %s | %s |\n",
			markdownEscape(finding.Primary.Path),
			finding.Primary.Start.Line,
			finding.Primary.Start.Column,
			markdownEscape(finding.RuleID),
			markdownEscape(findingMessage(finding)),
		)
	}
	if len(findings) < len(result.Findings) {
		fmt.Fprintf(
			&output,
			"\n%d findings omitted by display options. The gate uses all findings.\n",
			len(result.Findings)-len(findings),
		)
	}
	for _, reason := range result.Gate.Reasons {
		fmt.Fprintf(&output, "\n- Gate: %s — %s\n", markdownEscape(reason.Path), markdownEscape(reason.Message))
	}
	for _, failure := range result.Errors {
		fmt.Fprintf(&output, "\n- Error: %s — %s\n", markdownEscape(failure.Path), markdownEscape(failure.Message))
	}
	for _, line := range append(auditLines(result), featureLines(result)...) {
		fmt.Fprintf(&output, "\n%s\n", markdownEscape(line))
	}
	writeModelLines(&output, result, markdownEscape, "\nRevision probability: unavailable; this alpha has no calibrated model.\n",
		"\nRevision probability: %s\n", "\nOrigin estimate: %s\n")
	if len(result.Manifest.ConfigSources) > 1 || len(result.Manifest.ConfigOverrides) > 0 {
		fmt.Fprintf(&output, "\nConfiguration: %s (%s).\n",
			markdownEscape(result.Manifest.ConfigHash), markdownEscape(result.Manifest.ConfigIdentity))
	}
	_, err := io.WriteString(writer, output.String())
	return err
}

// probabilitySummary describes a configured model and how many units it
// estimated. It is empty when no pack applies, keeping model-free wording.
// The caller supplies the escaping its format needs for declared values.
func probabilitySummary(result unswell.RunResult, escape func(string) string) string {
	model := result.Manifest.Probability
	if model == nil {
		return ""
	}
	estimated, total := 0, 0
	for _, assessment := range result.Assessments {
		if assessment.Scope != model.Kind {
			continue
		}
		total++
		if assessment.SlopProbability != nil {
			estimated++
		}
	}
	return fmt.Sprintf("%s (%s, corpus %s); estimated for %d of %d %s units, and every other unit keeps a status.",
		escape(model.PackID), escape(model.DeclaredStatus), escape(model.HumanCorpus), estimated, total, escape(model.Kind))
}

// writeModelLines states both model channels in one place, so the text and
// Markdown writers describe them identically.
func writeModelLines(output *strings.Builder, result unswell.RunResult, escape func(string) string,
	modelFree, revision, origin string,
) {
	if summary := probabilitySummary(result, escape); summary != "" {
		fmt.Fprintf(output, revision, summary)
	} else {
		output.WriteString(modelFree)
	}
	if summary := originSummary(result, escape); summary != "" {
		fmt.Fprintf(output, origin, summary)
	}
}

// originSummary describes the separate experimental origin channel. It is empty
// unless a policy configures one, and it never reports a share of a text.
func originSummary(result unswell.RunResult, escape func(string) string) string {
	model := result.Manifest.Origin
	if model == nil {
		return ""
	}
	estimated, total := 0, 0
	for _, assessment := range result.Assessments {
		if assessment.Scope != model.Kind {
			continue
		}
		total++
		if assessment.OriginEstimate != nil {
			estimated++
		}
	}
	return fmt.Sprintf("%s (%s, corpus %s); estimated for %d of %d %s units. "+
		"It is similarity to a training class, not a quality judgment, and it decides no gate.",
		escape(model.PackID), escape(model.DeclaredStatus), escape(model.HumanCorpus), estimated, total,
		escape(model.Kind))
}

func markdownEscape(text string) string {
	var output strings.Builder
	for _, r := range terminal(text) {
		switch {
		case strings.ContainsRune("\\`*_{}[]<>()#+-.!|&", r):
			fmt.Fprintf(&output, "&#%d;", r)
		case r == '\n', r == '\r':
			output.WriteByte(' ')
		default:
			output.WriteRune(r)
		}
	}
	return output.String()
}
