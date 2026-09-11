package training

import (
	"context"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

// RunLexical fits a training-only vocabulary and the shared Go classifier.
// It uses the same reproduced NLP units, annotation selection, permissions, and
// separate calibration as Run. Options.Features must be empty: column IDs are
// learned from permitted, resolved training targets only. The returned vocabulary
// exposes source-derived keys explicitly; normal scan reports are unchanged.
func RunLexical(ctx context.Context, candidates corpus.Artifact, round *annotation.Round, files map[string][]byte,
	options Options, lexical LexicalOptions,
) (Artifact, error) {
	if err := validateOptions(ctx, candidates, options); err != nil {
		return Artifact{}, err
	}
	if err := lexical.validate(); err != nil {
		return Artifact{}, err
	}
	if len(options.Features) != 0 {
		return Artifact{}, fmt.Errorf("lexical training learns columns; do not supply features")
	}
	prepared, err := corpus.Prepare(ctx, candidates, files)
	if err != nil {
		return Artifact{}, err
	}
	expected := make([]annotation.Unit, len(candidates.Units))
	for i, candidate := range candidates.Units {
		expected[i] = candidate.Unit
	}
	if err := round.MatchTargets(ctx, expected); err != nil {
		return Artifact{}, err
	}
	decisions, err := round.Decisions(ctx)
	if err != nil {
		return Artifact{}, err
	}
	if decisions.Basis == "simulation" && !options.AllowSimulation {
		return Artifact{}, fmt.Errorf("tutorial training requires explicit allow_simulation")
	}
	return fitLexical(ctx, candidates, prepared, decisions, options, lexical)
}

// RunLexicalDecisions fits the lexical baseline from a prepared decision set,
// such as the provenance labels of the origin task, instead of a round.
func RunLexicalDecisions(ctx context.Context, candidates corpus.Artifact, decisions annotation.DecisionSet,
	files map[string][]byte, options Options, lexical LexicalOptions,
) (Artifact, error) {
	if err := validateOptions(ctx, candidates, options); err != nil {
		return Artifact{}, err
	}
	if err := lexical.validate(); err != nil {
		return Artifact{}, err
	}
	if len(options.Features) != 0 {
		return Artifact{}, fmt.Errorf("lexical training learns columns; do not supply features")
	}
	if err := corpus.MatchDecisionTargets(ctx, candidates, decisions); err != nil {
		return Artifact{}, err
	}
	prepared, err := corpus.Prepare(ctx, candidates, files)
	if err != nil {
		return Artifact{}, err
	}
	if decisions.Basis == "simulation" && !options.AllowSimulation {
		return Artifact{}, fmt.Errorf("tutorial training requires explicit allow_simulation")
	}
	return fitLexical(ctx, candidates, prepared, decisions, options, lexical)
}

func fitLexical(ctx context.Context, candidates corpus.Artifact, prepared corpus.Prepared,
	decisions annotation.DecisionSet, options Options, lexical LexicalOptions,
) (Artifact, error) {
	selector, bindings, err := lexicalSelector(ctx, candidates, prepared, decisions, options, nil, nil)
	if err != nil {
		return Artifact{}, err
	}
	selected, err := selectMeasuredRows(ctx, candidates.Plan, decisions, bindings, selector)
	if err != nil {
		return Artifact{}, err
	}
	counts, err := collectLexical(ctx, prepared, selectedLexicalIDs(selected), lexical.Counts)
	if err != nil {
		return Artifact{}, err
	}
	vocabulary, err := learnVocabulary(ctx, selected, lexical, counts)
	if err != nil {
		return Artifact{}, err
	}
	for _, column := range lexicalColumns(vocabulary, options.Kind) {
		options.Features = append(options.Features, column.ID)
	}
	selector, bindings, err = lexicalSelector(ctx, candidates, prepared, decisions, options, &vocabulary, counts)
	if err != nil {
		return Artifact{}, err
	}
	selected, err = selectMeasuredRows(ctx, candidates.Plan, decisions, bindings, selector)
	if err != nil {
		return Artifact{}, err
	}
	joinedHash, err := hashJSON(struct {
		Corpus, Round, Vocabulary string
		Bindings                  []corpus.FeatureBinding
	}{candidates.SHA256, decisions.RoundSHA256, vocabulary.SHA256, bindings})
	if err != nil {
		return Artifact{}, err
	}
	result, err := fitSelected(ctx, candidates, decisions, joinedHash, options, selected)
	if err != nil {
		return Artifact{}, err
	}
	result.Lexical, result.SHA256 = &vocabulary, ""
	return finish(ctx, result)
}

func selectedLexicalIDs(selected selection) map[string]bool {
	result := make(map[string]bool)
	for _, partition := range selected.partitions {
		for _, row := range partition.Rows {
			result[row.UnitID] = true
		}
	}
	return result
}

func learnVocabulary(ctx context.Context, selected selection, options LexicalOptions,
	histograms map[string][]feature.LexicalTerm,
) (Vocabulary, error) {
	training := selected.partitions[0].Rows
	counts := make(map[string]int)
	bytes := 0
	for _, row := range training {
		if err := ctx.Err(); err != nil {
			return Vocabulary{}, err
		}
		if err := countVocabularyTargets(counts, histograms[row.UnitID], &bytes); err != nil {
			return Vocabulary{}, err
		}
	}
	return freezeVocabulary(ctx, counts, len(training), options)
}

func validateLexicalArtifact(a Artifact) error {
	if a.Identity.FeatureSource != "lexical_ngrams" {
		if a.Lexical != nil || a.Identity.LexicalVocabularyHash != "" {
			return fmt.Errorf("nonlexical artifacts cannot contain a lexical vocabulary")
		}
		return nil
	}
	if a.Lexical == nil {
		return fmt.Errorf("lexical artifacts require an explicit vocabulary")
	}
	if err := validateVocabulary(*a.Lexical); err != nil {
		return err
	}
	return validateLexicalIdentity(a)
}

func validateLexicalIdentity(a Artifact) error {
	columns := lexicalColumns(*a.Lexical, a.Options.Kind)
	hash, err := hashJSON(columns)
	if err != nil || hash != a.Identity.ColumnsSHA256 || a.Identity.LexicalVocabularyHash != a.Lexical.SHA256 ||
		a.Identity.FeatureContract != feature.LexicalCountContract || a.Identity.Context != "prepared_piece" ||
		a.Identity.Preprocessing != lexicalPreprocessing {
		return fmt.Errorf("lexical artifact representation mismatch")
	}
	return validateLexicalTargets(a)
}

func validateLexicalTargets(a Artifact) error {
	if len(a.Partitions) == 0 || a.Lexical.TrainingTargets != len(a.Partitions[0].Rows) ||
		!slices.IsSorted(a.Options.Features) {
		return fmt.Errorf("lexical vocabulary training-target mismatch")
	}
	return nil
}

func countVocabularyTargets(counts map[string]int, terms []feature.LexicalTerm, bytes *int) error {
	for _, term := range terms {
		if counts[term.Key] == 0 {
			*bytes += len(term.Key)
		}
		if len(term.Key) > 1<<16 || (counts[term.Key] == 0 && len(counts) >= 100000) || *bytes > 16<<20 {
			return fmt.Errorf("training vocabulary exceeds candidate-key budget")
		}
		counts[term.Key]++
	}
	return nil
}
