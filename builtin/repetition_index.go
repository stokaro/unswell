package builtin

import "slices"

type overlapIndex struct {
	postings map[string][]int
	budget   *repetitionBudget
}

func (index *overlapIndex) add(keys []string, at int) error {
	if err := index.budget.spend(len(keys)); err != nil {
		return err
	}
	for _, key := range keys {
		index.postings[key] = append(index.postings[key], at)
	}
	return nil
}

func (index *overlapIndex) expire(keys []string, at int) error {
	if err := index.budget.spend(len(keys)); err != nil {
		return err
	}
	for _, key := range keys {
		ids := index.postings[key]
		if len(ids) > 0 && ids[0] == at {
			ids = ids[1:]
		}
		if len(ids) == 0 {
			delete(index.postings, key)
		} else {
			index.postings[key] = ids
		}
	}
	return nil
}

func (index *overlapIndex) candidates(keys []string) ([]int, error) {
	seen := make(map[int]bool)
	for _, key := range keys {
		ids := index.postings[key]
		if err := index.budget.spend(len(ids) + 1); err != nil {
			return nil, err
		}
		for _, id := range ids {
			seen[id] = true
		}
	}
	result := make([]int, 0, len(seen))
	for id := range seen {
		result = append(result, id)
	}
	slices.Sort(result)
	return result, nil
}
