package training

import (
	"context"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

// compressionSource is the reference source of a bank: two columns per
// cohort, measured by the shared compression primitive.
func compressionSource(kind string, bank corpus.CompressionBank) (referenceSource, error) {
	columns, err := compressionColumns(kind, bank)
	if err != nil {
		return referenceSource{}, err
	}
	return referenceSource{source: "compression_bank", contract: feature.CompressionContract,
		preprocessing: compressionPreprocessing, hash: bank.SHA256, columns: columns,
		measure: func(ctx context.Context, prepared corpus.Prepared, kind string) (map[string]map[string]feature.Value, error) {
			return measureCompression(ctx, prepared, kind, bank)
		}}, nil
}

// compressionColumns describes the two measurements of every reference
// cohort, with the cohort ID as a column suffix.
func compressionColumns(kind string, bank corpus.CompressionBank) ([]feature.Descriptor, error) {
	catalog, err := feature.CompressionCatalog(kind)
	if err != nil {
		return nil, err
	}
	columns := make([]feature.Descriptor, 0, len(catalog)*len(bank.Cohorts))
	for _, cohort := range bank.Cohorts {
		for _, descriptor := range catalog {
			column := descriptor
			column.ID = descriptor.ID + "/" + cohort.ID
			column.Requires = slices.Clone(descriptor.Requires)
			columns = append(columns, column)
		}
	}
	return columns, nil
}

// measureCompression measures every prepared target of the kind against each
// reference cohort. The result maps a unit to its values, keyed by column ID.
func measureCompression(ctx context.Context, prepared corpus.Prepared, kind string,
	bank corpus.CompressionBank,
) (map[string]map[string]feature.Value, error) {
	compressors, err := bankCompressors(ctx, bank)
	if err != nil {
		return nil, err
	}
	limits := feature.Limits{MaxTokens: 1000000, MaxUniqueWords: 65536, MaxBytes: corpus.MaxSourceBytes, MaxBlocks: 1}
	result := make(map[string]map[string]feature.Value)
	for _, target := range prepared.Targets {
		if target.Unit.Binding().Kind != kind {
			continue
		}
		values, err := measureTarget(ctx, compressors, bank, target, limits)
		if err != nil {
			return nil, err
		}
		result[target.UnitID] = values
	}
	return result, ctx.Err()
}

// bankCompressors initializes one compressor per reference cohort and checks
// each against the identity the bank recorded.
func bankCompressors(ctx context.Context, bank corpus.CompressionBank) ([]*feature.Compression, error) {
	compressors := make([]*feature.Compression, len(bank.Cohorts))
	for i, cohort := range bank.Cohorts {
		compressor, err := feature.NewCompression(ctx, cohort.Reference, bank.Options.Compression)
		if err != nil {
			return nil, err
		}
		if compressor.Identity().ReferenceSHA256 != cohort.Identity.ReferenceSHA256 {
			return nil, fmt.Errorf("compression reference %s does not match the bank's identity", cohort.ID)
		}
		compressors[i] = compressor
	}
	return compressors, nil
}

func measureTarget(ctx context.Context, compressors []*feature.Compression, bank corpus.CompressionBank,
	target corpus.PreparedTarget, limits feature.Limits,
) (map[string]feature.Value, error) {
	identity := feature.Identity{NLP: target.Unit.Identity(), Capabilities: target.Unit.Capabilities(),
		Source: "verified-corpus", Policy: "frozen-corpus-extraction", Vocabulary: "reference-bank",
		Preprocessing: compressionPreprocessing}
	values := make(map[string]feature.Value)
	for i, cohort := range bank.Cohorts {
		measured, err := compressors[i].Measure(ctx, target.Unit, identity, limits)
		if err != nil {
			return nil, fmt.Errorf("unit %s against reference %s: %w", target.UnitID, cohort.ID, err)
		}
		for _, value := range measured.Values {
			value.ID += "/" + cohort.ID
			values[value.ID] = value
		}
	}
	return values, nil
}

// compressionPredictionMeasurements measures every target the same way the
// fit did; the predictor keeps the planned partition.
func compressionPredictionMeasurements(ctx context.Context, candidates corpus.Artifact, files map[string][]byte,
	fitted Artifact, bank *corpus.CompressionBank,
) (rowSelector, []corpus.FeatureBinding, corpus.Verification, error) {
	if bank == nil || bank.SHA256 != fitted.Identity.CompressionBankHash {
		return rowSelector{}, nil, corpus.Verification{}, fmt.Errorf("compression prediction requires the fitted reference bank")
	}
	source, err := compressionSource(fitted.Options.Kind, *bank)
	if err != nil {
		return rowSelector{}, nil, corpus.Verification{}, err
	}
	return referencePredictionMeasurements(ctx, candidates, files, fitted, source)
}
