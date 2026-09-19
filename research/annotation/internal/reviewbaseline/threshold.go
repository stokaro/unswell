package reviewbaseline

import (
	"cmp"
	"fmt"
	"slices"
)

func threshold(predictions []prediction) (operatingPoint, error) {
	point, err := countLabels(predictions)
	if err != nil {
		return operatingPoint{}, err
	}
	return chooseThreshold(predictions, point), nil
}

func countLabels(predictions []prediction) (operatingPoint, error) {
	point := operatingPoint{Reason: "no_operating_point"}
	for _, row := range predictions {
		if row.Label == nil {
			return operatingPoint{}, fmt.Errorf("threshold selection requires explicit labels")
		}
		if *row.Label == 1 {
			point.Positive++
		} else {
			point.Negative++
		}
	}
	if point.Positive == 0 || point.Negative == 0 {
		return operatingPoint{}, fmt.Errorf("development threshold requires both classes")
	}
	return point, nil
}

func chooseThreshold(predictions []prediction, point operatingPoint) operatingPoint {
	ordered := slices.Clone(predictions)
	slices.SortFunc(ordered, func(a, b prediction) int { return cmp.Compare(b.Score, a.Score) })
	trueCount, falseCount := 0, 0
	for i, row := range ordered {
		if *row.Label == 1 {
			trueCount++
		} else {
			falseCount++
		}
		if i+1 < len(ordered) && ordered[i+1].Score == row.Score {
			continue
		}
		// Integer comparisons avoid floating boundary errors and include all ties.
		if falseCount*100 > point.Negative || trueCount*100 < 85*(trueCount+falseCount) || trueCount == 0 {
			continue
		}
		if !point.Available || trueCount > point.True {
			score := row.Score
			point.Available, point.Reason, point.Threshold = true, "", &score
			point.True, point.False = trueCount, falseCount
		}
	}
	return point
}
