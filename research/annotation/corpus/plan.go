package corpus

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

const assignmentAlgorithm = "connected-sources-sha256-v1"

// MakePlan groups original sources before extraction and assigns whole components.
// The returned manifest is detached from the caller. Source/group/context sets
// are sorted; extraction exceptions retain their declared selector order.
func MakePlan(ctx context.Context, manifest Manifest) (Plan, error) {
	if err := manifest.validate(ctx); err != nil {
		return Plan{}, err
	}
	canonical, err := canonicalManifest(manifest)
	if err != nil {
		return Plan{}, err
	}
	hash, err := digest(canonical)
	if err != nil {
		return Plan{}, err
	}
	groups, err := components(ctx, canonical)
	if err != nil {
		return Plan{}, err
	}
	return Plan{Version: Version, Algorithm: assignmentAlgorithm, Manifest: canonical,
		ManifestSHA256: hash, Groups: groups}, nil
}

// ValidatePlan recomputes all components and assignments instead of trusting IDs.
func ValidatePlan(ctx context.Context, plan Plan) error {
	expected, err := MakePlan(ctx, plan.Manifest)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(plan, expected) {
		return fmt.Errorf("plan does not match its frozen manifest and assignment algorithm")
	}
	return nil
}

func canonicalManifest(manifest Manifest) (Manifest, error) {
	data, err := json.Marshal(manifest)
	if err != nil {
		return Manifest{}, err
	}
	if len(data) > MaxManifestBytes {
		return Manifest{}, fmt.Errorf("manifest exceeds byte limit")
	}
	var result Manifest
	if err := json.Unmarshal(data, &result); err != nil {
		return Manifest{}, err
	}
	slices.SortFunc(result.Sources, func(a, b Source) int { return strings.Compare(a.ID, b.ID) })
	slices.Sort(result.UnitKinds)
	slices.Sort(result.Policy.Contexts)
	for _, language := range result.Policy.Languages {
		slices.Sort(language.Contexts)
	}
	for i := range result.Sources {
		source := &result.Sources[i]
		for _, values := range [][]string{source.Authors, source.Templates, source.Related, source.GenerationTasks, source.Rights.AllowedUses} {
			slices.Sort(values)
		}
		slices.SortFunc(source.Notices, func(a, b Notice) int { return strings.Compare(a.Path, b.Path) })
	}
	return result, nil
}

type disjoint struct{ parent []int }

func (d disjoint) root(index int) int {
	for d.parent[index] != index {
		d.parent[index] = d.parent[d.parent[index]]
		index = d.parent[index]
	}
	return index
}

func (d disjoint) join(a, b int) {
	a, b = d.root(a), d.root(b)
	d.parent[max(a, b)] = min(a, b)
}

func connectSources(ctx context.Context, m Manifest) (disjoint, error) {
	sets := disjoint{parent: make([]int, len(m.Sources))}
	owners := make(map[string]int)
	for i, source := range m.Sources {
		if err := ctx.Err(); err != nil {
			return disjoint{}, err
		}
		sets.parent[i] = i
		for _, key := range sourceKeys(source) {
			if owner, found := owners[key]; found {
				sets.join(i, owner)
			} else {
				owners[key] = i
			}
		}
		if len(owners) > 100000 {
			return disjoint{}, fmt.Errorf("grouping exceeds 100000 distinct keys")
		}
	}
	return sets, nil
}

func components(ctx context.Context, m Manifest) ([]Group, error) {
	sets, err := connectSources(ctx, m)
	if err != nil {
		return nil, err
	}
	members := make(map[int][]Source)
	for i, source := range m.Sources {
		members[sets.root(i)] = append(members[sets.root(i)], source)
	}
	groups := make([]Group, 0, len(members))
	roots := make([]int, 0, len(members))
	for root := range members {
		roots = append(roots, root)
	}
	slices.Sort(roots)
	for _, root := range roots {
		sources := members[root]
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		group, err := makeGroup(m, sources)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	slices.SortFunc(groups, func(a, b Group) int { return strings.Compare(a.ID, b.ID) })
	return groups, nil
}

func sourceKeys(source Source) []string {
	keys := []string{"repository:" + source.Repository, "document:" + source.Document, "source-sha256:" + source.SHA256}
	for _, family := range []struct {
		kind   string
		values []string
	}{
		{"author", source.Authors}, {"template", source.Templates}, {"related", source.Related}, {"generation-task", source.GenerationTasks},
	} {
		for _, value := range family.values {
			keys = append(keys, family.kind+":"+value)
		}
	}
	slices.Sort(keys)
	return keys
}

func makeGroup(manifest Manifest, sources []Source) (Group, error) {
	group := Group{Sources: []string{}, Keys: []string{}}
	for _, source := range sources {
		group.Sources = append(group.Sources, source.ID)
		group.Keys = append(group.Keys, sourceKeys(source)...)
		if source.Partition != "" {
			if group.Partition != "" && group.Partition != source.Partition {
				return Group{}, fmt.Errorf("connected sources have conflicting partition pins: %s and %s", group.Partition, source.Partition)
			}
			group.Partition, group.Pinned = source.Partition, true
		}
	}
	slices.Sort(group.Sources)
	slices.Sort(group.Keys)
	group.Keys = slices.Compact(group.Keys)
	identity, err := digest(group.Keys)
	if err != nil {
		return Group{}, err
	}
	group.ID = identity
	if !group.Pinned {
		group.Partition = assign(manifest.Seed, identity, manifest.Weights)
	}
	return group, nil
}

func assign(seed, identity string, weights Weights) string {
	hash := sha256.Sum256([]byte(assignmentAlgorithm + "\x00" + seed + "\x00" + identity))
	bucket := int(binary.BigEndian.Uint64(hash[:8]) % 10000)
	limits := []int{weights.Training, weights.Development, weights.Calibration, weights.FinalTest}
	for i, weight := range limits {
		if bucket < weight {
			return partitions()[i]
		}
		bucket -= weight
	}
	return ""
}

// Files returns the exact, deduplicated local files required by a valid plan.
// Shared notice paths must declare the same hash and byte count everywhere.
func Files(ctx context.Context, plan Plan) ([]Notice, error) {
	if err := ValidatePlan(ctx, plan); err != nil {
		return nil, err
	}
	return requiredFiles(plan.Manifest)
}

func requiredFiles(manifest Manifest) ([]Notice, error) {
	byPath := make(map[string]Notice)
	total := 0
	for _, source := range manifest.Sources {
		files := append([]Notice{{Path: source.Path, SHA256: source.SHA256, Bytes: source.Bytes}}, source.Notices...)
		for _, file := range files {
			if previous, found := byPath[file.Path]; found {
				if previous != file {
					return nil, fmt.Errorf("conflicting file declarations for %s", file.Path)
				}
				continue
			}
			byPath[file.Path] = file
			total += file.Bytes
			if total > MaxTotalBytes {
				return nil, fmt.Errorf("source and notice bytes exceed %d", MaxTotalBytes)
			}
		}
	}
	result := make([]Notice, 0, len(byPath))
	for _, value := range byPath {
		result = append(result, value)
	}
	slices.SortFunc(result, func(a, b Notice) int { return strings.Compare(a.Path, b.Path) })
	return result, nil
}
