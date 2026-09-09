package annotation

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// OriginClaim keeps curated provenance separate from an editorial judgment.
type OriginClaim struct {
	UnitID string         `json:"unit_id"`
	Target DecisionTarget `json:"target"`
	Origin Origin         `json:"origin"`
}

// OriginSet binds declared provenance to a validated round, without source prose.
// It does not establish the truth of a claim or qualify a human corpus.
type OriginSet struct {
	Version     string        `json:"version"`
	SHA256      string        `json:"sha256,omitempty"`
	RoundID     string        `json:"round_id"`
	RoundSHA256 string        `json:"round_sha256"`
	Purpose     string        `json:"purpose"`
	Basis       string        `json:"basis"`
	Units       []OriginClaim `json:"units"`
}

// Origins exports detached, source-bound claims without selecting quality labels.
// Evidence and acquisition metadata must not be shown to blinded annotators.
func (r *Round) Origins(ctx context.Context) (OriginSet, error) {
	if err := ctx.Err(); err != nil {
		return OriginSet{}, err
	}
	if r == nil || r.inputSHA256 == "" || r.data.packetSHA256 == "" {
		return OriginSet{}, fmt.Errorf("load a validated annotation round before exporting origins")
	}
	result := OriginSet{Version: "unswell-origin-claims-v1", RoundID: r.data.ID, RoundSHA256: r.inputSHA256,
		Purpose: r.data.Purpose, Basis: "declared_provenance", Units: make([]OriginClaim, 0, len(r.data.Units))}
	if r.data.Purpose == "tutorial" {
		result.Basis = "simulation"
	}
	for _, unit := range r.data.Units {
		if err := ctx.Err(); err != nil {
			return OriginSet{}, err
		}
		result.Units = append(result.Units, OriginClaim{UnitID: unit.ID, Target: decisionTarget(unit), Origin: unit.Origin})
	}
	slices.SortFunc(result.Units, func(a, b OriginClaim) int { return strings.Compare(a.UnitID, b.UnitID) })
	return finishOrigins(ctx, result)
}

func finishOrigins(ctx context.Context, result OriginSet) (OriginSet, error) {
	data, err := json.Marshal(result)
	if err != nil {
		return OriginSet{}, err
	}
	if len(data)+128 > MaxDecisionBytes {
		return OriginSet{}, fmt.Errorf("origin export exceeds %d bytes", MaxDecisionBytes)
	}
	result.SHA256 = fmt.Sprintf("%x", sha256.Sum256(data))
	if err := ctx.Err(); err != nil {
		return OriginSet{}, err
	}
	return result, nil
}
