package patterns

import (
	"slices"
	"sort"

	"github.com/stokaro/unswell/research/annotation/generation"
)

func taskRoles(tasks generation.Tasks) []string {
	var roles []string
	for _, task := range tasks.Tasks {
		if !slices.Contains(roles, task.Role) {
			roles = append(roles, task.Role)
		}
	}
	sort.Strings(roles)
	return roles
}
