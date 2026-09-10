package unswell

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"slices"
	"strings"
	"unicode"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/internal/terms"
	"github.com/stokaro/unswell/rule"
)

func (e *Engine) analyzeSource(ctx context.Context, source document.Source, identities *sourceIdentities) (RunResult, error) {
	result := e.emptyResult()
	doc, err := extract.Parse(
		ctx,
		source,
		extract.Options{
			IncludeStructure: e.extractionStructure() || identities != nil,
			IncludeQuotes:    e.policy.Analysis.IncludeQuotes,
			MaxBytes:         e.policy.Analysis.MaxFileBytes,
			MaxBlocks:        e.policy.Analysis.MaxBlocks,
			Policy:           e.policy.Extraction,
		},
	)
	if err != nil {
		return result, err
	}
	if err := e.enrich(ctx, &doc); err != nil {
		return result, err
	}
	plan, err := e.suppressionPlan(ctx, doc)
	if err != nil {
		return result, err
	}
	result.Documents = append(result.Documents, e.documentResult(doc))
	if err := e.evaluateRules(ctx, &doc, &result, e.extractionStructure() || identities != nil); err != nil {
		return result, err
	}
	sortFindings(result.Findings)
	result.Findings = deduplicateFindings(result.Findings)
	builder, err := e.identityBuilder(ctx, doc, identities)
	if err != nil {
		return result, err
	}
	suppressionErr := e.applySuppressions(ctx, &result, doc, plan, builder)
	if err := e.summarizeSource(ctx, &result, doc, identities, builder); err != nil {
		return result, err
	}
	return result, suppressionErr
}

// summarizeSource scores units, decides revision probabilities, records source
// identities, and evaluates the gate for one already analyzed document.
func (e *Engine) summarizeSource(ctx context.Context, result *RunResult, doc document.Document,
	identities *sourceIdentities, builder *debtBuilder,
) error {
	structure := e.extractionStructure() || identities != nil
	estimates, err := e.estimateChannel(ctx, &doc, structure, e.revision)
	if err != nil {
		return err
	}
	origins, err := e.estimateChannel(ctx, &doc, structure, e.origin)
	if err != nil {
		return err
	}
	e.assess(result, doc, estimates, origins)
	if builder != nil {
		if err := e.identifySource(ctx, result, doc, identities, builder); err != nil {
			return err
		}
	}
	e.decide(result, doc)
	return nil
}

func (e *Engine) extractionStructure() bool { return e.collectBaseline || e.requiresStructure() }

func (e *Engine) requiresStructure() bool {
	if len(e.featureIDs) > 0 {
		return true
	}
	for _, implementation := range e.rules {
		descriptor := implementation.Descriptor()
		if e.policy.Rules[descriptor.ID].Enabled && descriptor.RequiresStructure {
			return true
		}
	}
	return false
}

func (e *Engine) evaluateRules(ctx context.Context, doc *document.Document, result *RunResult, structure bool) error {
	if err := e.capturePreparedFeatures(ctx, doc, result, structure); err != nil {
		return err
	}
	features, err := e.sharedFeatures(ctx, doc)
	if err != nil {
		return err
	}
	if err := e.captureFeatures(ctx, doc, features, result); err != nil {
		return err
	}
	termMatches, err := e.matchTerms(ctx, doc)
	if err != nil {
		return err
	}
	for _, implementation := range e.rules {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := e.runRule(ctx, implementation, rule.View{Document: doc, Features: features}, termMatches, result); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) matchTerms(ctx context.Context, doc *document.Document) (*rule.TermMatches, error) {
	vocabulary := e.policy.Vocabulary
	if len(vocabulary.TermExemptions) == 0 || len(vocabulary.ResolvedTerms) == 0 {
		return nil, nil
	}
	matcher, err := terms.Compile(vocabulary.ResolvedTerms, vocabulary.CaseSensitive)
	if err != nil {
		return nil, err
	}
	ranges, err := matcher.Find(ctx, doc, e.policy.Analysis.MaxCandidates)
	if err != nil {
		return nil, err
	}
	return rule.NewTermMatches(ranges)
}

func (e *Engine) enrich(ctx context.Context, doc *document.Document) error {
	tokens, sentenceID := 0, 0
	for i := range doc.Blocks {
		block := &doc.Blocks[i]
		if err := englishApplicable(block.Text); err != nil {
			return err
		}
		sentences, err := e.nlp.Analyze(ctx, block.MappedText, e.capabilities)
		if err != nil {
			return err
		}
		for j := range sentences {
			sentence := &sentences[j]
			sentence.ID = sentenceID
			sentence.BlockID = block.ID
			sentenceID++
			tokens += len(sentence.Tokens)
			if tokens > e.policy.Analysis.MaxTokens {
				return fmt.Errorf("source exceeds max_tokens")
			}
			block.Words += sentence.Words
		}
		if err := e.validateDependencies(ctx, block.MappedText, sentences); err != nil {
			return err
		}
		block.Sentences = sentences
		doc.Words += block.Words
	}
	return ctx.Err()
}

func englishApplicable(text string) error {
	letters, nonLatin := 0, 0
	for _, r := range text {
		if !unicode.IsLetter(r) {
			continue
		}
		letters++
		if !unicode.In(r, unicode.Latin) {
			nonLatin++
		}
	}
	if letters >= 20 && nonLatin*2 > letters {
		return fmt.Errorf("block is inapplicable to configured English analysis (predominantly non-Latin prose)")
	}
	return nil
}

func (e *Engine) documentResult(doc document.Document) DocumentResult {
	result := DocumentResult{
		GateMode:         e.selectedGateMode(),
		Name:             doc.Name,
		Format:           doc.Format,
		SourceHash:       doc.Hash,
		ConfigHash:       e.policy.Hash,
		AppliedOverrides: slices.Clone(e.policy.AppliedOverrides),
		Bytes:            len(doc.Source),
		ProseWords:       doc.Words,
		Blocks:           len(doc.Blocks),
		Excluded:         doc.Excluded,
	}
	for _, block := range doc.Blocks {
		result.Sentences += len(block.Sentences)
	}
	if e.includeSource {
		result.Source = string(doc.Source)
	}
	return result
}

type collector struct {
	activations   *feature.ActivationBuilder
	ctx           context.Context
	doc           *document.Document
	descriptor    rule.Descriptor
	settings      rule.Settings
	limit         int
	includeSource bool
	findings      []Finding
	err           error
}

// Emit latches validation failures even when a custom rule ignores the error.
func (c *collector) Emit(evidence rule.Evidence) error {
	if c.err != nil {
		return c.err
	}
	c.err = c.append(evidence)
	return c.err
}

func (c *collector) append(evidence rule.Evidence) error {
	if err := c.ctx.Err(); err != nil {
		return err
	}
	if len(c.findings) >= c.limit {
		return fmt.Errorf("max_findings exceeded")
	}
	if err := validateEvidence(evidence); err != nil {
		return err
	}
	locations := make([]Location, 0, len(evidence.Occurrences))
	identities := []string{c.doc.Name, c.descriptor.ID, c.descriptor.Version}
	for _, occurrence := range evidence.Occurrences {
		location, err := c.occurrenceLocation(occurrence)
		if err != nil {
			return err
		}
		locations = append(locations, location)
		block := c.doc.Blocks[occurrence.BlockID]
		identities = append(identities, strings.Join(strings.Fields(document.Normalize(block.Text)), " "))
	}
	message := evidence.Message
	if message == "" {
		message = c.descriptor.Summary
	}
	for _, location := range locations {
		for _, span := range location.Segments {
			identities = append(identities, document.Normalize(string(c.doc.Source[span.Start:span.End])))
		}
	}
	fingerprint := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(identities, "\n"))))
	instance := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%d", fingerprint, locations[0].Span.Start, locations[0].Span.End))))
	c.findings = append(c.findings, Finding{ID: instance[:16], RuleID: c.descriptor.ID, RuleVersion: c.descriptor.Version,
		Severity: c.settings.Severity, Gate: c.settings.Gate, Group: c.descriptor.Group, Scope: c.descriptor.Scope, Message: message,
		Primary: locations[0], Related: locations[1:], Evidence: evidence, Fingerprint: fingerprint, BaselineState: "untracked"})
	return c.observeOccurrences(evidence)
}

func (c *collector) occurrenceLocation(occurrence rule.Occurrence) (Location, error) {
	if occurrence.BlockID < 0 || occurrence.BlockID >= len(c.doc.Blocks) {
		return Location{}, fmt.Errorf("invalid block ID")
	}
	block := c.doc.Blocks[occurrence.BlockID]
	if !slices.ContainsFunc(block.Sentences, func(s document.Sentence) bool { return s.ID == occurrence.SentenceID }) {
		return Location{}, fmt.Errorf("invalid sentence ID")
	}
	location, err := locate(c.doc, occurrence.Spans, c.includeSource)
	if err != nil {
		return Location{}, err
	}
	if location.Span.Start < block.Span.Start || location.Span.End > block.Span.End {
		return Location{}, fmt.Errorf("evidence lies outside its block")
	}
	return location, nil
}

func validateEvidence(evidence rule.Evidence) error {
	if !slices.Contains([]string{"exact", "heuristic", "statistical"}, evidence.Kind) {
		return fmt.Errorf("invalid evidence kind")
	}
	if evidence.Activation < 0 || evidence.Activation > 1000 {
		return fmt.Errorf("activation must be in [0,1000]")
	}
	if len(evidence.Occurrences) == 0 || len(evidence.Occurrences) > 10000 {
		return fmt.Errorf("invalid occurrence count")
	}
	for _, metric := range evidence.Metrics {
		if !validMetric(metric) {
			return fmt.Errorf("invalid evidence metric")
		}
	}
	return nil
}

func validMetric(metric rule.Metric) bool {
	return metric.Name != "" && metric.Unit != "" && finite(metric.Value) && finite(metric.Onset) && finite(metric.Saturation)
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func locate(doc *document.Document, spans []document.Span, includeSource bool) (Location, error) {
	if len(spans) == 0 || len(spans) > 200000 {
		return Location{}, fmt.Errorf("invalid evidence segments")
	}
	last := -1
	for _, span := range spans {
		if !span.Valid(len(doc.Source)) || span.Start < last {
			return Location{}, fmt.Errorf("invalid or unordered evidence span")
		}
		last = span.Start
	}
	bounds := document.Bounds(spans)
	start, err := document.Locate(doc.Source, bounds.Start)
	if err != nil {
		return Location{}, err
	}
	end, err := document.Locate(doc.Source, bounds.End)
	if err != nil {
		return Location{}, err
	}
	location := Location{Path: doc.Name, Span: bounds, Start: start, End: end, Segments: slices.Clone(spans)}
	if includeSource {
		location.Snippet = string(doc.Source[bounds.Start:bounds.End])
	}
	return location, nil
}

func sortFindings(findings []Finding) {
	slices.SortFunc(findings, func(a, b Finding) int {
		if a.Primary.Span.Start != b.Primary.Span.Start {
			return a.Primary.Span.Start - b.Primary.Span.Start
		}
		if a.RuleID != b.RuleID {
			return strings.Compare(a.RuleID, b.RuleID)
		}
		return strings.Compare(a.Fingerprint, b.Fingerprint)
	})
}

func deduplicateFindings(findings []Finding) []Finding {
	return slices.CompactFunc(findings, func(a, b Finding) bool {
		return a.RuleID == b.RuleID && a.Primary.Span == b.Primary.Span && a.Fingerprint == b.Fingerprint
	})
}

func cloneParameters(parameters rule.Parameters) rule.Parameters {
	parameters.Phrases = slices.Clone(parameters.Phrases)
	parameters.Positions = slices.Clone(parameters.Positions)
	parameters.Verbs = slices.Clone(parameters.Verbs)
	parameters.Nouns = slices.Clone(parameters.Nouns)
	return parameters
}
