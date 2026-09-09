package evaluation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
	"github.com/stokaro/unswell/research/annotation/training"
)

// ComparisonVersion fixes pairing, support, resampling, and interval semantics.
const ComparisonVersion = "unswell-research-comparison-v1"

// ComparisonPlan identifies the selected trials before evaluation labels open.
// The fixed numerical method is described by ComparisonVersion. Content hashes
// do not attest to preregistration, independent sampling, or annotation provenance.
type ComparisonPlan struct {
	Version          string `json:"version"`
	ID               string `json:"id"`
	ProtocolSHA256   string `json:"protocol_sha256"`
	CandidateSHA256  string `json:"candidate_predictions_sha256"`
	ComparatorSHA256 string `json:"comparator_predictions_sha256"`
}

// LoadComparisonPlan validates a bounded, explicit selection of two saved trials.
func LoadComparisonPlan(ctx context.Context, data []byte) (ComparisonPlan, error) {
	var plan ComparisonPlan
	if err := jsoninput.Decode(ctx, data, 16<<10, &plan, jsoninput.Limits{Object: 8}); err != nil {
		return ComparisonPlan{}, err
	}
	if plan.Version != ComparisonVersion || strings.TrimSpace(plan.ID) == "" || len(plan.ID) > 128 {
		return ComparisonPlan{}, fmt.Errorf("comparison requires a supported version and an ID of 1 to 128 bytes")
	}
	for _, digest := range []string{plan.ProtocolSHA256, plan.CandidateSHA256, plan.ComparatorSHA256} {
		decoded, err := hex.DecodeString(digest)
		if err != nil || len(decoded) != sha256.Size || strings.ToLower(digest) != digest {
			return ComparisonPlan{}, fmt.Errorf("comparison digests must be 64 lowercase hexadecimal characters")
		}
	}
	return plan, ctx.Err()
}

func comparisonPlanIdentity(ctx context.Context, plan ComparisonPlan, a, b training.Predictions) (string, error) {
	data, err := json.Marshal(plan)
	if err != nil {
		return "", err
	}
	if _, err := LoadComparisonPlan(ctx, data); err != nil {
		return "", err
	}
	if err := matchingComparisonScope(plan, a, b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

func matchingComparisonScope(plan ComparisonPlan, a, b training.Predictions) error {
	if plan.CandidateSHA256 != a.SHA256 || plan.ComparatorSHA256 != b.SHA256 {
		return fmt.Errorf("comparison trials differ from the frozen selection")
	}
	if plan.ProtocolSHA256 != a.Plan.ProtocolSHA256 || plan.ProtocolSHA256 != b.Plan.ProtocolSHA256 {
		return fmt.Errorf("compared trials require the same frozen protocol")
	}
	if a.Plan.CorpusSHA256 != b.Plan.CorpusSHA256 || a.Plan.Partition != b.Plan.Partition || a.Plan.Context != b.Plan.Context {
		return fmt.Errorf("compared trials require identical corpus, partition, and available context")
	}
	if !compatibleComparisonIdentity(a.Model.Identity, b.Model.Identity) {
		return fmt.Errorf("compared trials have different targets, rubric, profile, preprocessing, or NLP")
	}
	return nil
}

func compatibleComparisonIdentity(a, b training.Identity) bool {
	// Feature columns and fitted parameters may differ; source preprocessing and
	// the editorial target must stay fixed for this controlled numerical comparison.
	a.Columns, b.Columns = nil, nil
	a.ColumnsSHA256, b.ColumnsSHA256 = "", ""
	// Required capabilities describe feature computation, not additional source
	// context. The provider identity, including its supported capabilities, stays
	// identical; each trial still records the capabilities its columns requested.
	a.Capabilities, b.Capabilities = nil, nil
	return reflect.DeepEqual(a, b)
}
