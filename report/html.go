package report

import (
	_ "embed"
	"html/template"
	"io"
	"slices"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

type piece struct {
	Text      string
	Highlight bool
}
type paragraph struct {
	Assessment unswell.Assessment
	Pieces     []piece
	Anchor     string
}
type htmlData struct {
	PreparedRows []featureRow
	FeatureRows  []featureRow
	Result       unswell.RunResult
	Verdict      string
	Probability  string
	Origin       string
	Findings     []unswell.Finding
	Paragraphs   []paragraph
	Omitted      int
}

// identityEscape passes values through; html/template escapes on output.
func identityEscape(value string) string { return value }

func htmlReport(writer io.Writer, result unswell.RunResult, options Options) error {
	data := htmlData{PreparedRows: preparedRows(result), FeatureRows: featureRows(result),
		Result: result, Verdict: verdict(result), Probability: probabilitySummary(result, identityEscape),
		Origin:   originSummary(result, identityEscape),
		Findings: visible(result, options)}
	data.Omitted = len(result.Findings) - len(data.Findings)
	for _, assessment := range result.Assessments {
		if assessment.Scope != "paragraph" || assessment.SlopScore == 0 {
			continue
		}
		data.Paragraphs = append(data.Paragraphs, paragraph{Assessment: assessment, Pieces: paragraphPieces(result, assessment)})
	}
	compiled, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return err
	}
	return compiled.Execute(writer, data)
}

func paragraphPieces(result unswell.RunResult, assessment unswell.Assessment) []piece {
	var source string
	for _, doc := range result.Documents {
		if doc.Name == assessment.Path {
			source = doc.Source
			break
		}
	}
	span := assessment.Span
	if !span.Valid(len(source)) {
		return []piece{{Text: "Source text was not included in this saved result."}}
	}
	hits := paragraphHits(result, assessment)
	return highlightPieces(source, span, hits)
}

func paragraphHits(result unswell.RunResult, assessment unswell.Assessment) []document.Span {
	span := assessment.Span
	hits := make([]document.Span, 0)
	for _, finding := range result.Findings {
		if finding.Primary.Path != assessment.Path || finding.Derived {
			continue
		}
		for _, occurrence := range finding.Evidence.Occurrences {
			for _, hit := range occurrence.Spans {
				hit.Start = max(hit.Start, span.Start)
				hit.End = min(hit.End, span.End)
				if hit.End > hit.Start {
					hits = append(hits, hit)
				}
			}
		}
	}
	slices.SortFunc(hits, func(a, b document.Span) int { return a.Start - b.Start })
	return hits
}

func highlightPieces(source string, span document.Span, hits []document.Span) []piece {
	pieces := make([]piece, 0)
	pos := span.Start
	for _, hit := range hits {
		if hit.End <= pos {
			continue
		}
		if hit.Start > pos {
			pieces = append(pieces, piece{Text: source[pos:hit.Start]})
		}
		pieces = append(pieces, piece{Text: source[max(pos, hit.Start):hit.End], Highlight: true})
		pos = hit.End
	}
	if pos < span.End {
		pieces = append(pieces, piece{Text: source[pos:span.End]})
	}
	return pieces
}

//go:embed template.html
var htmlTemplate string
