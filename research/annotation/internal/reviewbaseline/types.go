// Package reviewbaseline compares numerical proposals on exposed assistant labels.
// It does not create a product model pack or qualify editorial probabilities.
package reviewbaseline

import (
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/nlp"
)

type input struct {
	Version int    `json:"version"`
	Basis   string `json:"basis"`
	Pages   []page `json:"pages"`
}

type page struct {
	ID         string          `json:"id"`
	Repository string          `json:"repository"`
	Source     string          `json:"source_group"`
	Cohort     string          `json:"cohort"`
	Format     document.Format `json:"format"`
	Hash       string          `json:"sha256"`
	Text       string          `json:"text"`
	Group      string          `json:"group"`
	Fold       int             `json:"fold"`
	Units      []expectedUnit  `json:"units"`
	Excluded   []int           `json:"excluded_blocks"`
}

type expectedUnit struct {
	Binding nlp.UnitBinding `json:"binding"`
	Label   *int            `json:"label"`
}

type row struct {
	page, group, cohort, textHash string
	text                          string
	unit, fold, words             int
	label                         *int
	numeric                       []float64
	terms                         map[string]bool
}

type prediction struct {
	Page     string  `json:"page"`
	Unit     int     `json:"unit"`
	Words    int     `json:"words"`
	Cohort   string  `json:"cohort"`
	Label    *int    `json:"label"`
	Score    float64 `json:"raw_score"`
	Selected bool    `json:"selected"`
}

type operatingPoint struct {
	Available bool     `json:"available"`
	Threshold *float64 `json:"threshold"`
	Reason    string   `json:"reason,omitempty"`
	Positive  int      `json:"development_positives"`
	Negative  int      `json:"development_negatives"`
	True      int      `json:"development_true"`
	False     int      `json:"development_false"`
}

type fitted struct {
	Kind           string           `json:"kind"`
	Fold           int              `json:"evaluation_fold"`
	Features       []string         `json:"features"`
	Parameters     model.Parameters `json:"parameters"`
	TrainingHash   string           `json:"training_sha256"`
	TrainingGroups []string         `json:"training_groups"`
	TrainingRows   int              `json:"training_rows"`
	Options        model.FitOptions `json:"options"`
	Iterations     int              `json:"iterations"`
	Operations     int64            `json:"operations"`
	Gradient       float64          `json:"gradient_norm,omitempty"`
	OperatingPoint operatingPoint   `json:"operating_point"`
	Predictions    []prediction     `json:"predictions"`
}

type output struct {
	Version         int          `json:"version"`
	Basis           string       `json:"basis"`
	InputHash       string       `json:"input_sha256"`
	ModelAlgorithm  string       `json:"model_algorithm"`
	UnitContract    string       `json:"unit_contract"`
	FeatureContract string       `json:"feature_contract"`
	LexicalContract string       `json:"lexical_contract"`
	Provider        nlp.Identity `json:"provider"`
	Rows            int          `json:"rows"`
	Models          []fitted     `json:"models"`
}
