package corpus

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"slices"
	"strings"
)

// UnionOptions picks the shards and sources of a dataset that join one
// manifest. Empty cohort, repository, and role lists pick everything. The unit
// kinds narrow the dataset's kinds and default to all of them. A positive
// MaxPerCheckout keeps the first sources of each checkout in ID order. A corpus
// of many repositories then fits one artifact, and no choice depends on the
// text. Checkouts of an uncapped cohort keep every source. A positive
// MaxSourceBytes drops sources above that size, which keeps one large file
// from taking most of a corpus or the measurement budget of its engine run.
type UnionOptions struct {
	ID              string
	Cohorts         []string
	Repositories    []string
	Roles           []string
	UnitKinds       []string
	MaxPerCheckout  int
	UncappedCohorts []string
	MaxSourceBytes  int
}

// UnionManifest joins the pinned sources of chosen shards into one manifest.
// The corpus then spans cohorts. Each source path and notice path gains the
// prefix of its checkout, cohort/owner__repo. That is the layout the
// acquisition driver writes, so one root serves extraction. Partitions stay
// pinned to the dataset's assignment, and the header comes from the dataset.
// The result passes the same checks as a curated manifest. The unit and byte
// limits of one artifact still apply at extraction.
func UnionManifest(ctx context.Context, plan DatasetPlan, pinned map[string][]byte, options UnionOptions) (Manifest, error) {
	if err := ctx.Err(); err != nil {
		return Manifest{}, err
	}
	kinds, err := unionKinds(plan.Dataset.UnitKinds, options.UnitKinds)
	if err != nil {
		return Manifest{}, err
	}
	sources, err := selectedSources(ctx, plan, pinned, options)
	if err != nil {
		return Manifest{}, err
	}
	result := Manifest{Version: Version, ID: options.ID, Seed: plan.Dataset.Seed, Weights: plan.Dataset.Weights,
		Policy: plan.Dataset.Policy, UnitKinds: kinds, Sources: sources}
	return sealedManifest(ctx, result)
}

// selectedSources collects the selected sources of every shard in plan order,
// with their paths prefixed, and refuses a source that two shards declare.
func selectedSources(ctx context.Context, plan DatasetPlan, pinned map[string][]byte, options UnionOptions) ([]Source, error) {
	if options.MaxPerCheckout < 0 {
		return nil, fmt.Errorf("the per-checkout limit cannot be negative")
	}
	var sources []Source
	seen := make(map[string]string)
	for _, shard := range plan.Shards {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		manifest, err := pinnedShard(ctx, shard, pinned)
		if err != nil {
			return nil, err
		}
		for _, source := range manifest.Sources {
			if !selected(source, options) {
				continue
			}
			if previous, exists := seen[source.ID]; exists {
				return nil, fmt.Errorf("source %s appears in shards %s and %s", source.ID, previous, shard.Path)
			}
			seen[source.ID] = shard.Path
			sources = append(sources, prefixed(source))
		}
	}
	sources = capped(sources, options)
	if len(sources) == 0 {
		return nil, fmt.Errorf("no source of the dataset matches the selection")
	}
	return sources, nil
}

// capped keeps the first sources of each capped checkout in source ID order
// and returns all sources in that order. A zero limit keeps everything.
func capped(sources []Source, options UnionOptions) []Source {
	if options.MaxPerCheckout == 0 {
		return sources
	}
	slices.SortFunc(sources, func(a, b Source) int {
		if by := strings.Compare(CheckoutPrefix(a), CheckoutPrefix(b)); by != 0 {
			return by
		}
		return strings.Compare(a.ID, b.ID)
	})
	kept := sources[:0]
	count, previous := 0, ""
	for _, source := range sources {
		if prefix := CheckoutPrefix(source); prefix != previous {
			count, previous = 0, prefix
		}
		count++
		if count <= options.MaxPerCheckout || uncapped(source, options.UncappedCohorts) {
			kept = append(kept, source)
		}
	}
	return kept
}

func uncapped(source Source, cohorts []string) bool {
	return source.Snapshot != nil && slices.Contains(cohorts, source.Snapshot.Cohort)
}

func unionKinds(available, requested []string) ([]string, error) {
	if len(requested) == 0 {
		return slices.Clone(available), nil
	}
	kinds := slices.Clone(requested)
	slices.Sort(kinds)
	kinds = slices.Compact(kinds)
	for _, kind := range kinds {
		if !slices.Contains(available, kind) {
			return nil, fmt.Errorf("unit kind %s is not in the dataset", kind)
		}
	}
	return kinds, nil
}

// pinnedShard decodes one pinned shard and checks it against the digest the
// plan recorded, so a union never reads a manifest the plan did not pin.
func pinnedShard(ctx context.Context, shard DatasetShard, pinned map[string][]byte) (Manifest, error) {
	data, exists := pinned[shard.Path]
	if !exists {
		return Manifest{}, fmt.Errorf("pinned shard %s is missing", shard.Path)
	}
	if hashBytes(data) != shard.PinnedSHA256 {
		return Manifest{}, fmt.Errorf("pinned shard %s does not match the plan", shard.Path)
	}
	return LoadManifest(ctx, data)
}

func selected(source Source, options UnionOptions) bool {
	if len(options.Repositories) > 0 && !slices.Contains(options.Repositories, source.Repository) {
		return false
	}
	if options.MaxSourceBytes > 0 && source.Bytes > options.MaxSourceBytes {
		return false
	}
	if len(options.Roles) > 0 && !slices.Contains(options.Roles, source.Role) {
		return false
	}
	if len(options.Cohorts) == 0 {
		return true
	}
	return source.Snapshot != nil && slices.Contains(options.Cohorts, source.Snapshot.Cohort)
}

// CheckoutPrefix is the directory of a source's checkout under the acquisition
// work root: the cohort, then the repository with its slash replaced.
func CheckoutPrefix(source Source) string {
	repository := strings.ReplaceAll(source.Repository, "/", "__")
	if source.Snapshot != nil && source.Snapshot.Cohort != "" {
		return path.Join(source.Snapshot.Cohort, repository)
	}
	return repository
}

func prefixed(source Source) Source {
	prefix := CheckoutPrefix(source)
	result := source
	result.Path = path.Join(prefix, source.Path)
	result.Notices = slices.Clone(source.Notices)
	for i := range result.Notices {
		result.Notices[i].Path = path.Join(prefix, result.Notices[i].Path)
	}
	return result
}

// sealedManifest round-trips the union through the manifest loader, so it
// carries the same guarantees as a curated manifest.
func sealedManifest(ctx context.Context, manifest Manifest) (Manifest, error) {
	data, err := json.Marshal(manifest)
	if err != nil {
		return Manifest{}, err
	}
	if len(data) > MaxManifestBytes {
		return Manifest{}, fmt.Errorf("union manifest exceeds %d bytes", MaxManifestBytes)
	}
	return LoadManifest(ctx, data)
}
