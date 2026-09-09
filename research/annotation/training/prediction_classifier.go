package training

import "fmt"

func validRawPrediction(row Prediction) bool {
	if row.ForestResponse != nil {
		return validResponse(row.ForestResponse) && row.LinearScore == nil && row.LogisticResponse == nil
	}
	return finitePointer(row.LinearScore) && validResponse(row.LogisticResponse)
}

func consistentRawResponse(row Prediction, channel string) bool {
	switch channel {
	case "logistic":
		return row.LogisticResponse != nil && *row.Response == *row.LogisticResponse
	case "forest":
		return row.ForestResponse != nil && *row.Response == *row.ForestResponse
	default:
		return true
	}
}

func validatePredictionEstimator(rows []Prediction, estimator string) error {
	for _, row := range rows {
		if (row.ForestResponse != nil && estimator != "forest") ||
			(row.LogisticResponse != nil && estimator != "logistic") {
			return fmt.Errorf("prediction raw channel does not match the fitted estimator")
		}
	}
	return nil
}
