package unswell

import (
	"context"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/probability"
)

// ProbabilityModel repeats the declarations of the pack behind reported
// estimates. The engine checks compatibility and consistency, never the truth
// of an acceptance, corpus, or evaluation statement.
type ProbabilityModel struct {
	PackID         string `json:"pack_id"`
	SHA256         string `json:"sha256"`
	Version        string `json:"version"`
	Task           string `json:"task"`
	Rubric         string `json:"rubric"`
	Kind           string `json:"kind"`
	DeclaredStatus string `json:"declared_status"`
	HumanCorpus    string `json:"human_corpus"`
	Evaluation     string `json:"evaluation,omitempty"`
	MinWords       int    `json:"min_words"`
}

// probabilityRun holds one source's decisions. A nonempty status applies to
// every unit; otherwise units without a prepared target remain unsupported.
type probabilityRun struct {
	kind   string
	status string
	values map[document.Span]probability.Estimate
}

func (e *Engine) configureProbability(options Options) error {
	calibration := e.policy.Calibration
	if calibration == nil || calibration.Model != "pack" {
		if len(options.Model) != 0 {
			return fmt.Errorf("an explicit model pack requires calibration.model: pack")
		}
		return nil
	}
	if len(options.Model) == 0 {
		// Configuration inspection stays available; analysis cannot silently
		// abstain while its policy requests an explicitly supplied model.
		e.packMissing = true
		return nil
	}
	pack, err := probability.Load(context.Background(), options.Model)
	if err != nil {
		return err
	}
	if !pack.Accepted() && !calibration.AcceptExperimental {
		return fmt.Errorf("probability pack %s declares experimental status; set calibration.accept_experimental", pack.File().ID)
	}
	e.pack, e.packColumns = pack, pack.Columns()
	return e.planPackCapabilities()
}

func (e *Engine) planPackCapabilities() error {
	e.packCapabilities = slices.Clone(e.pack.File().Contract.Capabilities)
	supported := e.nlp.Identity().Capabilities
	for _, capability := range e.packCapabilities {
		if !slices.Contains(supported, capability) {
			return fmt.Errorf("probability pack requires unavailable capability %s", capability)
		}
	}
	return nil
}

// ProbabilityModelIdentity returns the configured pack's declarations, or nil.
func (e *Engine) ProbabilityModelIdentity() *ProbabilityModel {
	if e.pack == nil {
		return nil
	}
	file := e.pack.File()
	return &ProbabilityModel{PackID: file.ID, SHA256: file.SHA256, Version: file.Version, Task: file.Task,
		Rubric: file.Rubric, Kind: file.Kind, DeclaredStatus: file.DeclaredStatus, HumanCorpus: file.HumanCorpus,
		Evaluation: file.Evaluation, MinWords: file.Limits.MinWords}
}

// requireProbabilityModel keeps a run from analyzing prose without the model
// its policy explicitly requests.
func (e *Engine) requireProbabilityModel() error {
	if e.packMissing {
		return fmt.Errorf("calibration.model: pack requires an explicit model pack")
	}
	return nil
}

// estimateProbability decides one source's revision probabilities. Preparation
// runs separately from optional feature collection, with exactly the pack's
// capabilities, so its measurements match the pack's declared contract.
func (e *Engine) estimateProbability(ctx context.Context, doc *document.Document, structure bool) (probabilityRun, error) {
	if e.pack == nil {
		return probabilityRun{status: probability.StatusUnavailable}, nil
	}
	source, err := e.preparedSource(doc, structure)
	if err != nil {
		return probabilityRun{}, err
	}
	run := probability.Run{FeatureContract: feature.UnitContract, UnitContract: nlp.UnitContract, NLP: source.NLP,
		Capabilities: e.packCapabilities, PreparationHash: source.PreparationHash,
		IncludeQuotes: source.IncludeQuotes, IncludeStructure: source.IncludeStructure}
	if err := e.pack.Compatible(run); err != nil {
		if e.policy.Calibration.OnIncompatible == "fail" {
			return probabilityRun{}, err
		}
		return probabilityRun{status: probability.StatusIncompatible}, nil
	}
	values, err := e.measureProbability(ctx, doc, source)
	if err != nil {
		return probabilityRun{}, err
	}
	return probabilityRun{kind: e.pack.Kind(), values: values}, nil
}

func (e *Engine) measureProbability(ctx context.Context, doc *document.Document,
	source PreparedFeatureSource,
) (map[document.Span]probability.Estimate, error) {
	identity := feature.Identity{NLP: source.NLP, Capabilities: e.packCapabilities, Source: doc.Hash,
		Policy: source.PolicyHash, Vocabulary: source.VocabularyHash, Preprocessing: source.PreparationHash}
	values := make(map[document.Span]probability.Estimate)
	budget := e.policy.Analysis.MaxCandidates
	for _, block := range doc.Blocks {
		units, err := nlp.PrepareUnits(ctx, block, e.nlp, nlp.UnitOptions{Kinds: []string{e.pack.Kind()},
			Capabilities: e.packCapabilities, Limits: e.packUnitLimits(&budget)})
		if err != nil {
			return nil, fmt.Errorf("probability preparation: %w", err)
		}
		for _, unit := range units {
			if err := e.estimateUnit(ctx, unit, identity, &budget, values); err != nil {
				return nil, err
			}
		}
	}
	return values, ctx.Err()
}

func (e *Engine) packUnitLimits(budget *int) nlp.UnitLimits {
	return nlp.UnitLimits{MaxBytes: min(e.policy.Analysis.MaxFileBytes, 16<<20), MaxContextBytes: 65536,
		MaxUnits: min(max(*budget, 1), 10000), MaxTokens: min(e.policy.Analysis.MaxTokens, 1000000), MaxSegments: 4096}
}

func (e *Engine) estimateUnit(ctx context.Context, unit nlp.PreparedUnit, identity feature.Identity,
	budget *int, values map[document.Span]probability.Estimate,
) error {
	block := unit.Block()
	*budget -= len(e.packColumns)
	if *budget < 0 {
		return fmt.Errorf("probability measurements exceed max_candidates")
	}
	measurements, err := feature.MeasureUnit(ctx, unit, identity, feature.Limits{MaxTokens: *budget,
		MaxUniqueWords: *budget, MaxBytes: e.policy.Analysis.MaxFileBytes, MaxBlocks: 1})
	if err != nil {
		return err
	}
	*budget -= measurements.Counts().TokenVisits
	if *budget < 0 {
		return fmt.Errorf("probability measurements exceed max_candidates")
	}
	vector := make([]feature.Value, 0, len(e.packColumns))
	for _, column := range e.packColumns {
		value, err := measurements.Value(column.ID)
		if err != nil {
			return err
		}
		vector = append(vector, value)
	}
	estimate, err := e.pack.Estimate(ctx, probability.Unit{Kind: unit.Binding().Kind, Words: block.Words, Values: vector})
	if err != nil {
		return err
	}
	values[block.Span] = estimate
	return nil
}

// probabilityFor keeps an estimate out of a unit the pack does not qualify.
func (run probabilityRun) probabilityFor(scope string, span document.Span) (*float64, string, string) {
	if run.status != "" {
		return nil, run.status, ""
	}
	if scope != run.kind {
		return nil, probability.StatusUnsupportedUnit, run.kind
	}
	estimate, exists := run.values[span]
	if !exists {
		return nil, probability.StatusUnsupportedUnit, "unprepared_target"
	}
	return estimate.Probability, estimate.Status, estimate.Detail
}

// probabilityModelHash keeps accepted debt from carrying across a model change.
func (e *Engine) probabilityModelHash(hasher *debtHasher) string {
	if e.pack == nil {
		return hasher.hash("calibration.model:none")
	}
	return hasher.hash(e.ProbabilityModelIdentity())
}
