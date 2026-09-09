package report

import (
	"fmt"
	"strconv"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func preparedRows(result unswell.RunResult) []featureRow {
	var rows []featureRow
	if result.PreparedFeatures == nil {
		return rows
	}
	for _, source := range result.PreparedFeatures.Sources {
		for _, unit := range source.Units {
			span := document.Bounds(unit.Binding.Segments)
			location := fmt.Sprintf("%s [%s, bytes %d:%d]", source.Path, unit.Binding.Kind, span.Start, span.End)
			for _, value := range unit.Values {
				text := "unavailable (" + value.Reason + ")"
				if value.Number != nil {
					text = strconv.FormatFloat(*value.Number, 'g', -1, 64)
				}
				rows = append(rows, featureRow{Location: location, ID: value.ID, Value: text, Unit: value.Unit})
			}
		}
	}
	return rows
}

func preparedLines(result unswell.RunResult) []string {
	if result.PreparedFeatures == nil {
		return nil
	}
	lines := []string{fmt.Sprintf("Prepared measurements: %s / %s; analysis %s.",
		result.PreparedFeatures.Version, result.PreparedFeatures.FeatureContract, result.Status)}
	for _, row := range preparedRows(result) {
		lines = append(lines, fmt.Sprintf("%s %s: %s %s", row.Location, row.ID, row.Value, row.Unit))
	}
	return lines
}
