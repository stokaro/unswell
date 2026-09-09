package nlp

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// PreparationHash binds the unit contract, extraction policy, and source-selection
// switches. It is shared by engine measurements and explicit research consumers.
// The policy identity must already describe the effective extraction policy.
func PreparationHash(policy string, includeQuotes, includeStructure bool) (string, error) {
	data, err := json.Marshal(struct {
		Contract, Policy                string
		IncludeQuotes, IncludeStructure bool
	}{UnitContract, policy, includeQuotes, includeStructure})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}
