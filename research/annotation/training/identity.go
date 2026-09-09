package training

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"slices"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

func columnIdentity(joined corpus.JoinedArtifact, kind string) (Identity, error) {
	catalog, err := feature.UnitCatalog(kind)
	if err != nil {
		return Identity{}, err
	}
	columns := make([]feature.Descriptor, 0, len(joined.Features.Requested))
	for _, id := range joined.Features.Requested {
		index := slices.IndexFunc(catalog, func(d feature.Descriptor) bool { return d.ID == id })
		if index < 0 {
			return Identity{}, fmt.Errorf("unknown training feature %s", id)
		}
		columns = append(columns, catalog[index])
	}
	encoded, err := json.Marshal(columns)
	if err != nil {
		return Identity{}, err
	}
	return Identity{Task: "editorial_needs_revision", Rubric: joined.Decisions.Rubric, ProfileSHA256: joined.Decisions.ProfileSHA256,
		Kind: kind, FeatureContract: joined.Features.FeatureContract, UnitContract: joined.Features.UnitContract,
		Columns: columns, ColumnsSHA256: fmt.Sprintf("%x", sha256.Sum256(encoded))}, nil
}

func sourceIdentity(identity Identity, source unswell.PreparedFeatureSource) Identity {
	identity.NLP, identity.Capabilities = source.NLP, source.Capabilities
	identity.PolicyHash, identity.VocabularyHash = source.PolicyHash, source.VocabularyHash
	identity.ExtractionPolicyHash, identity.PreparationHash = source.ExtractionPolicyHash, source.PreparationHash
	identity.IncludeQuotes, identity.IncludeStructure = source.IncludeQuotes, source.IncludeStructure
	return identity
}

func (s *rowSelector) acceptIdentity(identity Identity) error {
	if !s.result.identitySet {
		s.result.identity, s.result.identitySet = identity, true
		return nil
	}
	if !reflect.DeepEqual(s.result.identity, identity) {
		return fmt.Errorf("selected rows have incompatible feature, policy, or NLP identities")
	}
	return nil
}

func numericValues(values []feature.Value, columns []feature.Descriptor) ([]float64, string, error) {
	if len(values) != len(columns) {
		return nil, "", fmt.Errorf("training measurement does not match the column count")
	}
	result := make([]float64, 0, len(values))
	missing := ""
	for i, value := range values {
		column := columns[i]
		if err := validateColumnValue(value, column); err != nil {
			return nil, "", err
		}
		if value.Number == nil {
			if missing == "" {
				missing = value.ID + "/" + value.Reason
			}
		} else {
			result = append(result, *value.Number)
		}
	}
	if missing != "" {
		return nil, missing, nil
	}
	return result, "", nil
}

func validateColumnValue(value feature.Value, column feature.Descriptor) error {
	if value.ID != column.ID || value.Version != column.Version || value.Unit != column.Unit {
		return fmt.Errorf("training measurement does not match column %s", column.ID)
	}
	if err := validNumber(value); err != nil {
		return err
	}
	if column.Family == "rule-activation" {
		return feature.ValidateActivation(value)
	}
	return nil
}

func validNumber(value feature.Value) error {
	if value.Number == nil {
		if value.Reason != "" {
			return nil
		}
	} else if value.Reason == "" && !math.IsNaN(*value.Number) && !math.IsInf(*value.Number, 0) {
		return nil
	}
	return fmt.Errorf("feature %s requires a finite value or an explicit absence reason", value.ID)
}
