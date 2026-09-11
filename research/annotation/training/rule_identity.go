package training

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/rule"
)

func ruleColumnIdentity(joined corpus.RuleJoinedArtifact, kind, configHash string) (Identity, error) {
	columns := make([]feature.Descriptor, 0, len(joined.Features.Requested))
	for _, id := range joined.Features.Requested {
		ruleID, ok := strings.CutPrefix(id, "activation/")
		index := slices.IndexFunc(joined.Rules, func(d rule.Descriptor) bool { return d.ID == ruleID })
		if !ok || index < 0 {
			return Identity{}, fmt.Errorf("unknown training activation %s", id)
		}
		column := feature.ActivationDescriptor(ruleID)
		column.Requires = slices.Clone(joined.Rules[index].Requires)
		columns = append(columns, column)
	}
	encoded, err := json.Marshal(columns)
	if err != nil {
		return Identity{}, err
	}
	task, err := annotation.TaskForRubric(joined.Decisions.Rubric)
	if err != nil {
		return Identity{}, err
	}
	return Identity{Task: task, Rubric: joined.Decisions.Rubric, ProfileSHA256: joined.Decisions.ProfileSHA256,
		Kind: kind, FeatureContract: joined.Features.BlockContract, UnitContract: unswell.FeatureBlockBindingContract,
		Columns: columns, ColumnsSHA256: fmt.Sprintf("%x", sha256.Sum256(encoded)), FeatureSource: "rule_activations",
		Context: joined.Context, ActivationContract: joined.Features.ActivationContract, RuleConfigSHA256: configHash,
		IncludeStructure: true}, nil // Original block collection requests source structure in the engine.
}

func ruleSourceIdentity(identity Identity, source unswell.FeatureSource) Identity {
	identity.NLP, identity.Capabilities = source.NLP, source.Capabilities
	identity.PolicyHash, identity.VocabularyHash = source.PolicyHash, source.VocabularyHash
	identity.RulesetHash, identity.Preprocessing = source.RulesetHash, source.Preprocessing
	return identity
}
