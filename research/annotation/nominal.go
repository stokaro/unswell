package annotation

// nominal implements the coincidence-weighted nominal alpha definition cited in
// README.md. Singleton ratings contribute to coverage, not alpha's marginals.
func nominal(rows [][]string) AgreementStats {
	result := AgreementStats{Units: len(rows), Counts: []Count{}}
	allCounts, pairedCounts := make(map[string]int), make(map[string]int)
	var pairs, matchingPairs, observed float64
	for _, row := range rows {
		counts := countLabels(row)
		for label, count := range counts {
			allCounts[label] += count
		}
		result.Ratings += len(row)
		if len(row) > 0 {
			result.RatedUnits++
		}
		if len(row) < 2 {
			continue
		}
		result.PairedUnits++
		result.PairedRatings += len(row)
		m := float64(len(row))
		rowMatches := 0.0
		for _, label := range sortedKeys(counts) {
			count := counts[label]
			pairedCounts[label] += count
			rowMatches += float64(count * (count - 1))
		}
		pairs += m * (m - 1)
		matchingPairs += rowMatches
		observed += (m*(m-1) - rowMatches) / (m - 1)
	}
	for _, label := range sortedKeys(allCounts) {
		result.Counts = append(result.Counts, Count{Label: label, Count: allCounts[label]})
	}
	result.RawAgreement = ratio(matchingPairs, pairs, "no_paired_ratings")
	result.Uncertain = ratio(float64(allCounts["uncertain"]), float64(result.Ratings), "no_ratings")
	result.Alpha = alpha(observed, result.PairedRatings, pairedCounts)
	return result
}

func countLabels(row []string) map[string]int {
	counts := make(map[string]int)
	for _, label := range row {
		counts[label]++
	}
	return counts
}

func alpha(observed float64, n int, marginals map[string]int) Estimate {
	if n < 2 {
		return Estimate{Reason: "no_paired_ratings"}
	}
	expectedPairs := 0.0
	for _, label := range sortedKeys(marginals) {
		count := float64(marginals[label])
		expectedPairs += count * (float64(n) - count)
	}
	if expectedPairs == 0 {
		return Estimate{Reason: "no_expected_disagreement"}
	}
	value := 1 - observed*float64(n-1)/expectedPairs
	return Estimate{Value: &value}
}

func ratio(numerator, denominator float64, reason string) Estimate {
	if denominator == 0 {
		return Estimate{Reason: reason}
	}
	value := numerator / denominator
	return Estimate{Value: &value}
}
