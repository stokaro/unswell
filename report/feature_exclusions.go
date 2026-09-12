package report

import (
	"fmt"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
)

func excludedFeatureSpans(doc unswell.DocumentResult) map[document.Span]bool {
	spans := make(map[document.Span]bool)
	for _, exclusion := range doc.Excluded {
		if exclusion.Reason == "non-latin-prose" {
			spans[exclusion.Span] = true
		}
	}
	return spans
}

func validateExcludedFeature(value feature.Value, definition feature.Descriptor) error {
	reason := "excluded_unit"
	if definition.Family == "rule-activation" {
		reason = "inapplicable/excluded_unit"
	}
	if value.Number != nil || value.Reason != reason {
		return fmt.Errorf("excluded feature value must be absent with its exclusion reason")
	}
	return nil
}
