package report

import (
	"fmt"
	"strconv"

	"github.com/stokaro/unswell"
)

type featureRow struct {
	Location, ID, Value, Unit string
}

func featureRows(result unswell.RunResult) []featureRow {
	var rows []featureRow
	if result.Features == nil {
		return rows
	}
	for _, source := range result.Features.Sources {
		for _, unit := range source.Units {
			location := fmt.Sprintf("%s [block %d, %s]", source.Path, unit.UnitID, unit.Kind)
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

func featureLines(result unswell.RunResult) []string {
	return append(blockFeatureLines(result), preparedLines(result)...)
}

func blockFeatureLines(result unswell.RunResult) []string {
	if result.Features == nil {
		return nil
	}
	lines := []string{fmt.Sprintf("Shared measurements: %s / %s; analysis %s.",
		result.Features.Version, result.Features.BlockContract, result.Status)}
	if result.Features.ActivationContract != "" {
		lines = append(lines, "Rule activations: "+result.Features.ActivationContract+"; values are signals, not probabilities.")
	}
	for _, row := range featureRows(result) {
		lines = append(lines, fmt.Sprintf("%s %s: %s %s", row.Location, row.ID, row.Value, row.Unit))
	}
	return lines
}
