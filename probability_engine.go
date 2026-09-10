package unswell

import (
	"context"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/config"
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

// modelChannel is one configured estimation channel. The revision channel can
// gate a build; the origin channel is separate research and never can.
type modelChannel struct {
	pack         *probability.Pack
	columns      []feature.Descriptor
	capabilities []nlp.Capability
	incompatible string
	setting      string
	absent       string
	missing      bool
}

// channelSelection is the policy shape both channels share.
type channelSelection struct {
	model              string
	onIncompatible     string
	acceptExperimental bool
}

func (e *Engine) configureModels(options Options) error {
	revision, err := configureChannel(channelFromCalibration(e.policy.Calibration), options.Model,
		"calibration", probability.Task, e.nlp)
	revision.absent = probability.StatusUnavailable
	if err != nil {
		return err
	}
	origin, err := configureChannel(channelFromOrigin(e.policy.Origin), options.OriginModel,
		"origin", probability.TaskOrigin, e.nlp)
	if err != nil {
		return err
	}
	e.revision, e.origin = revision, origin
	return nil
}

func channelFromCalibration(calibration *config.Calibration) channelSelection {
	if calibration == nil {
		return channelSelection{}
	}
	return channelSelection{model: calibration.Model, onIncompatible: calibration.OnIncompatible,
		acceptExperimental: calibration.AcceptExperimental}
}

func channelFromOrigin(origin *config.Origin) channelSelection {
	if origin == nil {
		return channelSelection{}
	}
	return channelSelection{model: origin.Model, onIncompatible: origin.OnIncompatible,
		acceptExperimental: origin.AcceptExperimental}
}

// configureChannel loads one explicitly supplied pack. Construction stays
// available without the pack, so a configuration remains diagnosable; analysis
// then fails rather than abstaining while its policy requests a model.
func configureChannel(selection channelSelection, data []byte, setting, task string,
	provider nlp.Provider,
) (modelChannel, error) {
	channel := modelChannel{setting: setting, incompatible: selection.onIncompatible}
	if selection.model != "pack" {
		if len(data) != 0 {
			return modelChannel{}, fmt.Errorf("an explicit model pack requires %s.model: pack", setting)
		}
		return channel, nil
	}
	if len(data) == 0 {
		channel.missing = true
		return channel, nil
	}
	pack, err := probability.Load(context.Background(), data)
	if err != nil {
		return modelChannel{}, err
	}
	if pack.Task() != task {
		return modelChannel{}, fmt.Errorf("%s.model expects a %s pack; %s estimates %s",
			setting, task, pack.File().ID, pack.Task())
	}
	if !pack.Accepted() && !selection.acceptExperimental {
		return modelChannel{}, fmt.Errorf("probability pack %s declares experimental status; set %s.accept_experimental",
			pack.File().ID, setting)
	}
	channel.pack, channel.columns = pack, pack.Columns()
	channel.capabilities = slices.Clone(pack.File().Contract.Capabilities)
	supported := provider.Identity().Capabilities
	for _, capability := range channel.capabilities {
		if !slices.Contains(supported, capability) {
			return modelChannel{}, fmt.Errorf("probability pack requires unavailable capability %s", capability)
		}
	}
	return channel, nil
}

// ProbabilityModelIdentity returns the revision pack's declarations, or nil.
func (e *Engine) ProbabilityModelIdentity() *ProbabilityModel { return e.revision.identity() }

// OriginModelIdentity returns the origin pack's declarations, or nil. An origin
// estimate is a separate experimental channel that no gate consults.
func (e *Engine) OriginModelIdentity() *ProbabilityModel { return e.origin.identity() }

func (c modelChannel) identity() *ProbabilityModel {
	if c.pack == nil {
		return nil
	}
	file := c.pack.File()
	return &ProbabilityModel{PackID: file.ID, SHA256: file.SHA256, Version: file.Version, Task: file.Task,
		Rubric: file.Rubric, Kind: file.Kind, DeclaredStatus: file.DeclaredStatus, HumanCorpus: file.HumanCorpus,
		Evaluation: file.Evaluation, MinWords: file.Limits.MinWords}
}

// requireModels keeps a run from analyzing prose without a model its policy
// explicitly requests, for either channel.
func (e *Engine) requireModels() error {
	for _, channel := range []modelChannel{e.revision, e.origin} {
		if channel.missing {
			return fmt.Errorf("%s.model: pack requires an explicit model pack", channel.setting)
		}
	}
	return nil
}

// estimateChannel decides one source's estimates for one channel. Preparation
// runs separately from optional feature collection, with exactly the pack's
// capabilities, so its measurements match the pack's declared contract.
func (e *Engine) estimateChannel(ctx context.Context, doc *document.Document, structure bool,
	channel modelChannel,
) (probabilityRun, error) {
	if channel.pack == nil {
		// A configured revision channel keeps the existing model-free status.
		// The origin channel stays silent, so a run without it is unchanged.
		return probabilityRun{status: channel.absent}, nil
	}
	source, err := e.preparedSource(doc, structure)
	if err != nil {
		return probabilityRun{}, err
	}
	run := probability.Run{FeatureContract: feature.UnitContract, UnitContract: nlp.UnitContract, NLP: source.NLP,
		Capabilities: channel.capabilities, PreparationHash: source.PreparationHash,
		IncludeQuotes: source.IncludeQuotes, IncludeStructure: source.IncludeStructure}
	if err := channel.pack.Compatible(run); err != nil {
		if channel.incompatible == "fail" {
			return probabilityRun{}, err
		}
		return probabilityRun{status: probability.StatusIncompatible}, nil
	}
	values, err := e.measureChannel(ctx, doc, source, channel)
	if err != nil {
		return probabilityRun{}, err
	}
	return probabilityRun{kind: channel.pack.Kind(), values: values}, nil
}

func (e *Engine) measureChannel(ctx context.Context, doc *document.Document, source PreparedFeatureSource,
	channel modelChannel,
) (map[document.Span]probability.Estimate, error) {
	identity := feature.Identity{NLP: source.NLP, Capabilities: channel.capabilities, Source: doc.Hash,
		Policy: source.PolicyHash, Vocabulary: source.VocabularyHash, Preprocessing: source.PreparationHash}
	values := make(map[document.Span]probability.Estimate)
	budget := e.policy.Analysis.MaxCandidates
	for _, block := range doc.Blocks {
		units, err := nlp.PrepareUnits(ctx, block, e.nlp, nlp.UnitOptions{Kinds: []string{channel.pack.Kind()},
			Capabilities: channel.capabilities, Limits: e.packUnitLimits(&budget)})
		if err != nil {
			return nil, fmt.Errorf("%s preparation: %w", channel.setting, err)
		}
		for _, unit := range units {
			if err := e.estimateUnit(ctx, unit, identity, channel, &budget, values); err != nil {
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
	channel modelChannel, budget *int, values map[document.Span]probability.Estimate,
) error {
	block := unit.Block()
	*budget -= len(channel.columns)
	if *budget < 0 {
		return fmt.Errorf("%s measurements exceed max_candidates", channel.setting)
	}
	measurements, err := feature.MeasureUnit(ctx, unit, identity, feature.Limits{MaxTokens: *budget,
		MaxUniqueWords: *budget, MaxBytes: e.policy.Analysis.MaxFileBytes, MaxBlocks: 1})
	if err != nil {
		return err
	}
	*budget -= measurements.Counts().TokenVisits
	if *budget < 0 {
		return fmt.Errorf("%s measurements exceed max_candidates", channel.setting)
	}
	vector := make([]feature.Value, 0, len(channel.columns))
	for _, column := range channel.columns {
		value, err := measurements.Value(column.ID)
		if err != nil {
			return err
		}
		vector = append(vector, value)
	}
	estimate, err := channel.pack.Estimate(ctx, probability.Unit{Kind: unit.Binding().Kind, Words: block.Words,
		Values: vector})
	if err != nil {
		return err
	}
	values[block.Span] = estimate
	return nil
}

// probabilityFor keeps an estimate out of a unit the pack does not qualify.
// A channel that is off reports nothing at all.
func (run probabilityRun) probabilityFor(scope string, span document.Span) (*float64, string, string) {
	if run.status == "" && run.kind == "" {
		return nil, "", ""
	}
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
// Only the revision channel enters it: an origin estimate changes no finding,
// no index, and no gate decision, so it cannot invalidate accepted debt.
func (e *Engine) probabilityModelHash(hasher *debtHasher) string {
	identity := e.ProbabilityModelIdentity()
	if identity == nil {
		return hasher.hash("calibration.model:none")
	}
	return hasher.hash(identity)
}
