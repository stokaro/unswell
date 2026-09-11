package training

import (
	"context"
	"fmt"
	"slices"

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
	joined, decisions, err := referenceRoundInputs(ctx, candidates, round, files, options.Features)
	if err != nil {
		return Artifact{}, err
	}
	source, err := compressionSource(options.Kind, bank)
	if err != nil {
		return Artifact{}, err
	}
	return fitReference(ctx, candidates, files, decisions, options, source, joined)
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
	joined, err := referenceDecisionInputs(ctx, candidates, decisions, files, options.Features)
	if err != nil {
		return Artifact{}, err
	}
	source, err := compressionSource(options.Kind, bank)
	if err != nil {
		return Artifact{}, err
	}
	return fitReference(ctx, candidates, files, decisions, options, source, joined)
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
	if err := validateReferenceFeatures(options.Features); err != nil {
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
	references := referenceColumns(a)
	if references == 0 || references%2 != 0 {
		return fmt.Errorf("compression artifact requires paired reference columns")
	}
	return nil
}
