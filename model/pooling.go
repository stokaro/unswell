package model

import "context"

type calibrationPool struct {
	start, end      int
	count, positive int
}

func poolCalibration(ctx context.Context, groups []calibrationGroup) (IsotonicParameters, int, float64, error) {
	pools := make([]calibrationPool, 0, len(groups))
	for i, group := range groups {
		if err := ctx.Err(); err != nil {
			return IsotonicParameters{}, 0, 0, err
		}
		pools = append(pools, calibrationPool{start: i, end: i + 1, count: group.count, positive: group.positive})
		for len(pools) >= 2 {
			last := len(pools) - 1
			left, right := &pools[last-1], pools[last]
			if int64(left.positive)*int64(right.count) < int64(right.positive)*int64(left.count) {
				break
			}
			left.end = right.end
			left.count += right.count
			left.positive += right.positive
			pools = pools[:last]
		}
	}
	parameters, loss, err := expandCalibration(ctx, groups, pools)
	return parameters, len(pools), loss, err
}

func expandCalibration(ctx context.Context, groups []calibrationGroup, pools []calibrationPool) (IsotonicParameters, float64, error) {
	p := IsotonicParameters{Scores: make([]float64, len(groups)), Responses: make([]float64, len(groups))}
	loss := 0.0
	for _, pool := range pools {
		mean := float64(pool.positive) / float64(pool.count)
		loss += float64(pool.positive)*(1-mean)*(1-mean) + float64(pool.count-pool.positive)*mean*mean
		for i := pool.start; i < pool.end; i++ {
			if err := ctx.Err(); err != nil {
				return IsotonicParameters{}, 0, err
			}
			p.Scores[i], p.Responses[i] = groups[i].score, mean
		}
	}
	return p, loss, nil
}
