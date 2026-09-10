package probability

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
)

// MinCalibrationKnots keeps a mapping usable over an interval, not one score.
const MinCalibrationKnots = 2

// MaxMinimumWords bounds the declared applicability floor.
const MaxMinimumWords = 10000

func validate(file File) error {
	if file.Version != Version {
		return fmt.Errorf("unsupported probability pack version %q", file.Version)
	}
	if file.Task != Task {
		return fmt.Errorf("unsupported probability pack task %q", file.Task)
	}
	if err := validateDeclarations(file); err != nil {
		return err
	}
	if err := validateContract(file); err != nil {
		return err
	}
	if err := validateNumerical(file); err != nil {
		return err
	}
	return validateDigest(file)
}

// validateDeclarations checks author statements for internal consistency only.
// An accepted pack must name its qualified corpus and published evaluation.
func validateDeclarations(file File) error {
	if !validText(file.ID, 128) || !validText(file.Rubric, 128) {
		return fmt.Errorf("probability pack requires a printable ID and rubric of at most 128 bytes")
	}
	if !slices.Contains([]string{"sentence", "paragraph"}, file.Kind) {
		return fmt.Errorf("unsupported probability pack unit kind %q", file.Kind)
	}
	if !slices.Contains([]string{"experimental", "accepted"}, file.DeclaredStatus) {
		return fmt.Errorf("probability pack status must be experimental or accepted")
	}
	if !slices.Contains([]string{"not_qualified", "qualified"}, file.HumanCorpus) {
		return fmt.Errorf("probability pack corpus must be not_qualified or qualified")
	}
	if err := validateAcceptance(file); err != nil {
		return err
	}
	if file.Limits.MinWords < 1 || file.Limits.MinWords > MaxMinimumWords {
		return fmt.Errorf("probability pack requires a declared minimum word count within 1..%d", MaxMinimumWords)
	}
	return nil
}

func validateAcceptance(file File) error {
	accepted := file.DeclaredStatus == "accepted"
	if accepted && (file.HumanCorpus != "qualified" || !validText(file.Evaluation, 256)) {
		return fmt.Errorf("an accepted probability pack requires a qualified corpus and a published evaluation reference")
	}
	if !accepted && file.Evaluation != "" && !validText(file.Evaluation, 256) {
		return fmt.Errorf("probability pack evaluation reference must be printable and at most 256 bytes")
	}
	return nil
}

// validateContract requires columns this build computes identically. Rule
// activations and dependency features are not part of this pack contract.
func validateContract(file File) error {
	contract := file.Contract
	if contract.FeatureContract != feature.UnitContract || contract.UnitContract != nlp.UnitContract {
		return fmt.Errorf("probability pack requires feature contract %s and unit contract %s",
			feature.UnitContract, nlp.UnitContract)
	}
	if !validHash(contract.PreparationHash) || !validIdentity(contract.NLP) {
		return fmt.Errorf("probability pack requires a preparation hash and a complete NLP identity")
	}
	if err := validateColumns(file); err != nil {
		return err
	}
	return validateCapabilities(contract)
}

func validateColumns(file File) error {
	columns := file.Contract.Columns
	if len(columns) < 1 || len(columns) > MaxColumns {
		return fmt.Errorf("probability pack requires 1 to %d columns", MaxColumns)
	}
	catalog, err := feature.UnitCatalog(file.Kind)
	if err != nil {
		return err
	}
	seen := make(map[string]bool, len(columns))
	for _, column := range columns {
		if seen[column.ID] {
			return fmt.Errorf("probability pack repeats column %q", column.ID)
		}
		seen[column.ID] = true
		index := slices.IndexFunc(catalog, func(d feature.Descriptor) bool { return d.ID == column.ID })
		if index < 0 || !equalDescriptor(catalog[index], column) {
			return fmt.Errorf("probability pack column %q does not match this build's %s measurement", column.ID, file.Kind)
		}
	}
	return validateColumnDigest(file.Contract)
}

func validateColumnDigest(contract Contract) error {
	encoded, err := json.Marshal(contract.Columns)
	if err != nil {
		return err
	}
	if contract.ColumnsSHA256 != fmt.Sprintf("%x", sha256.Sum256(encoded)) {
		return fmt.Errorf("probability pack column digest does not cover its declared columns")
	}
	return nil
}

func validateCapabilities(contract Contract) error {
	supported := preparedCapabilities()
	if len(contract.Capabilities) == 0 || len(contract.Capabilities) > len(supported) {
		return fmt.Errorf("probability pack requires the capabilities its columns need")
	}
	for i, capability := range contract.Capabilities {
		if !slices.Contains(supported, capability) || slices.Index(contract.Capabilities, capability) != i {
			return fmt.Errorf("probability pack capability %q is unknown or repeated", capability)
		}
	}
	for _, column := range contract.Columns {
		for _, required := range column.Requires {
			if !capabilityCovered(contract.Capabilities, required) {
				return fmt.Errorf("probability pack column %q requires the %s capability", column.ID, required)
			}
		}
	}
	return nil
}

// validateNumerical accepts one regularized linear estimator with separate
// calibration. Other estimators are rejected until they are qualified.
func validateNumerical(file File) error {
	if file.Estimator != "logistic" || file.Logistic == nil {
		return fmt.Errorf("probability pack requires the logistic estimator and its parameters")
	}
	width := len(file.Contract.Columns)
	if len(file.Logistic.Weights) != width || len(file.Logistic.Means) != width || len(file.Logistic.Scales) != width {
		return fmt.Errorf("probability pack logistic parameters must match its %d columns", width)
	}
	if file.Calibration.Algorithm != "isotonic" {
		return fmt.Errorf("probability pack requires isotonic calibration")
	}
	knots := len(file.Calibration.Scores)
	if knots < MinCalibrationKnots || knots != len(file.Calibration.Responses) {
		return fmt.Errorf("probability pack calibration requires at least %d matching knots", MinCalibrationKnots)
	}
	return nil
}

func validateDigest(file File) error {
	expected, err := Digest(file)
	if err != nil {
		return err
	}
	if file.SHA256 != expected {
		return fmt.Errorf("probability pack digest does not cover its contents")
	}
	return nil
}

// equalDescriptor compares complete serialized definitions, so a later field
// addition cannot silently accept a differently computed column.
func equalDescriptor(catalog, column feature.Descriptor) bool {
	return equalJSON(catalog, column)
}

func equalJSON(left, right any) bool {
	encodedLeft, err := json.Marshal(left)
	if err != nil {
		return false
	}
	encodedRight, err := json.Marshal(right)
	if err != nil {
		return false
	}
	return bytes.Equal(encodedLeft, encodedRight)
}

// preparedCapabilities lists the representations prepared targets can request.
func preparedCapabilities() []nlp.Capability {
	return []nlp.Capability{nlp.Tokens, nlp.Sentences, nlp.POS, nlp.Chunks, nlp.Dependencies}
}

func validIdentity(identity nlp.Identity) bool {
	return validText(identity.Name, 128) && validText(identity.Version, 128) && validText(identity.License, 128)
}

func validHash(value string) bool {
	if len(value) != 64 {
		return false
	}
	return strings.IndexFunc(value, func(r rune) bool {
		return !strings.ContainsRune("0123456789abcdef", r)
	}) < 0
}

func validText(value string, maximum int) bool {
	if value == "" || len(value) > maximum {
		return false
	}
	return strings.IndexFunc(value, func(r rune) bool {
		return r == unicode.ReplacementChar || !unicode.IsPrint(r)
	}) < 0
}
