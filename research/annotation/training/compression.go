package training

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

// compressionPreprocessing names what a compression measurement reads: the
// UTF-8 bytes of the prepared target, framed by the shared primitive.
const compressionPreprocessing = "shared-prepared-target/utf8-bytes-v1"

// RunCompression fits the compression baseline. Every target of the fitted
// kind is measured against each reference cohort of a bank built on the same
// frozen corpus, and the bank's reserved groups leave the fit. Options.Features
// may add prepared features to the same rows, so a fit can show what the
// references add over them. The bank's prose never enters the artifact.
func RunCompression(ctx context.Context, candidates corpus.Artifact, round *annotation.Round, files map[string][]byte,
	options Options, bank corpus.CompressionBank,
) (Artifact, error) {
	options, err := compressionGuards(ctx, candidates, options, bank)
	if err != nil {
		return Artifact{}, err
	}
	joined, decisions, err := compressionRoundInputs(ctx, candidates, round, files, options.Features)
	if err != nil {
		return Artifact{}, err
	}
	return fitCompression(ctx, candidates, files, decisions, options, bank, joined)
}

// RunCompressionDecisions fits the compression baseline from a prepared
// decision set, such as the cohort labels of the pattern protocol.
func RunCompressionDecisions(ctx context.Context, candidates corpus.Artifact, decisions annotation.DecisionSet,
	files map[string][]byte, options Options, bank corpus.CompressionBank,
) (Artifact, error) {
	options, err := compressionGuards(ctx, candidates, options, bank)
	if err != nil {
		return Artifact{}, err
	}
	joined, err := compressionDecisionInputs(ctx, candidates, decisions, files, options.Features)
	if err != nil {
		return Artifact{}, err
	}
	return fitCompression(ctx, candidates, files, decisions, options, bank, joined)
}

// compressionGuards checks the bank against the corpus and the options, and
// binds the bank's reservation to the fit.
func compressionGuards(ctx context.Context, candidates corpus.Artifact, options Options,
	bank corpus.CompressionBank,
) (Options, error) {
	if bank.Version != corpus.CompressionBankVersion || !validDigest(bank.SHA256) ||
		bank.Verification.ArtifactSHA256 != candidates.SHA256 || bank.Options.Kind != options.Kind {
		return Options{}, fmt.Errorf("compression training requires a sealed bank of the fitted kind built on the frozen corpus")
	}
	if err := validateCompressionFeatures(options.Features); err != nil {
		return Options{}, err
	}
	reservation := BankReservation(bank)
	if options.Reservation != nil && !slices.Equal(options.Reservation.Groups, reservation.Groups) {
		return Options{}, fmt.Errorf("compression training reserves the bank's groups and no others")
	}
	options.Reservation = reservation
	if err := validateOptions(ctx, candidates, options); err != nil {
		return Options{}, err
	}
	return options, zeroPolicyForRules(options, "compression_bank")
}

// validateCompressionFeatures admits prepared features beside the reference
// columns and nothing else.
func validateCompressionFeatures(features []string) error {
	for _, id := range features {
		if strings.HasPrefix(id, "activation/") || strings.HasPrefix(id, "compression.") {
			return fmt.Errorf("compression training adds reference columns to prepared features only")
		}
	}
	return nil
}

// BankReservation is the sorted set of groups a bank reserved, keyed by the
// bank's digest. A baseline compared with a compression fit carries it, so
// both train on the same rows.
func BankReservation(bank corpus.CompressionBank) *Reservation {
	groups := make([]string, 0, len(bank.ReservedGroups))
	for _, group := range bank.ReservedGroups {
		groups = append(groups, group.ID)
	}
	slices.Sort(groups)
	return &Reservation{BankSHA256: bank.SHA256, Groups: slices.Compact(groups)}
}

func compressionRoundInputs(ctx context.Context, candidates corpus.Artifact, round *annotation.Round,
	files map[string][]byte, features []string,
) (*corpus.JoinedArtifact, annotation.DecisionSet, error) {
	if len(features) != 0 {
		joined, err := corpus.Join(ctx, candidates, round, files, features)
		if err != nil {
			return nil, annotation.DecisionSet{}, err
		}
		return &joined, joined.Decisions, nil
	}
	expected := make([]annotation.Unit, len(candidates.Units))
	for i, candidate := range candidates.Units {
		expected[i] = candidate.Unit
	}
	if err := round.MatchTargets(ctx, expected); err != nil {
		return nil, annotation.DecisionSet{}, err
	}
	decisions, err := round.Decisions(ctx)
	return nil, decisions, err
}

func compressionDecisionInputs(ctx context.Context, candidates corpus.Artifact, decisions annotation.DecisionSet,
	files map[string][]byte, features []string,
) (*corpus.JoinedArtifact, error) {
	if len(features) != 0 {
		joined, err := corpus.JoinDecisions(ctx, candidates, decisions, files, features)
		if err != nil {
			return nil, err
		}
		return &joined, nil
	}
	return nil, corpus.MatchDecisionTargets(ctx, candidates, decisions)
}

func fitCompression(ctx context.Context, candidates corpus.Artifact, files map[string][]byte,
	decisions annotation.DecisionSet, options Options, bank corpus.CompressionBank, joined *corpus.JoinedArtifact,
) (Artifact, error) {
	if decisions.Basis == "simulation" && !options.AllowSimulation {
		return Artifact{}, fmt.Errorf("tutorial training requires explicit allow_simulation")
	}
	prepared, err := corpus.Prepare(ctx, candidates, files)
	if err != nil {
		return Artifact{}, err
	}
	selector, bindings, columns, err := compressionSelector(ctx, candidates, prepared, decisions, options, bank, joined)
	if err != nil {
		return Artifact{}, err
	}
	options.Features = make([]string, 0, len(columns))
	for _, column := range columns {
		options.Features = append(options.Features, column.ID)
	}
	selected, err := selectMeasuredRows(ctx, candidates.Plan, decisions, bindings, selector)
	if err != nil {
		return Artifact{}, err
	}
	joinedHash, err := hashJSON(struct {
		Corpus, Round, Bank string
		Bindings            []corpus.FeatureBinding
	}{candidates.SHA256, decisions.RoundSHA256, bank.SHA256, bindings})
	if err != nil {
		return Artifact{}, err
	}
	return fitSelected(ctx, candidates, decisions, joinedHash, options, selected)
}

// preparedFeatureIDs returns the prepared features of a compression fit,
// without its reference columns.
func preparedFeatureIDs(features []string) []string {
	result := make([]string, 0, len(features))
	for _, id := range features {
		if !strings.HasPrefix(id, "compression.") {
			result = append(result, id)
		}
	}
	return result
}

func validateCompressionArtifact(a Artifact) error {
	if a.Identity.FeatureSource != "compression_bank" {
		if a.Identity.CompressionBankHash != "" {
			return fmt.Errorf("noncompression artifacts cannot name a compression bank")
		}
		return nil
	}
	reservation := a.Options.Reservation
	if !validDigest(a.Identity.CompressionBankHash) || reservation == nil ||
		reservation.BankSHA256 != a.Identity.CompressionBankHash || a.Identity.Context != "prepared_piece" ||
		a.Identity.Preprocessing != compressionPreprocessing {
		return fmt.Errorf("compression artifact representation mismatch")
	}
	references := len(a.Options.Features) - len(preparedFeatureIDs(a.Options.Features))
	if references == 0 || references%2 != 0 {
		return fmt.Errorf("compression artifact requires paired reference columns")
	}
	return nil
}
