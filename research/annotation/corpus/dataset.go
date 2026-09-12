package corpus

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

// DatasetVersion identifies the dataset manifest and dataset plan contract.
const DatasetVersion = "unswell-corpus-dataset-v1"

// Dataset limits bound the union of shards without raising any shard's limit.
const (
	MaxDatasetBytes   = 4 << 20
	MaxShards         = 1024
	MaxDatasetSources = 200000
	maxDatasetKeys    = 2000000
)

// Dataset freezes one grouping plan over several shard manifests. Every shard
// repeats the seed, weights, extraction policy, and unit kinds declared here,
// and RuleClassesSHA256 pins the rule-class list the measurements will use.
type Dataset struct {
	Version           string         `json:"version"`
	ID                string         `json:"id"`
	Seed              string         `json:"seed"`
	Weights           Weights        `json:"weights"`
	Policy            extract.Policy `json:"extraction_policy"`
	UnitKinds         []string       `json:"unit_kinds"`
	RuleClassesSHA256 string         `json:"rule_classes_sha256"`
	Shards            []Notice       `json:"shards"`
}

// DatasetPlan assigns every connected component of the union of shards to one
// partition. Shards record both the unpinned manifest the curator wrote and
// the pinned manifest the plan derives from it; sources list the global
// group of every source, which a shard-local plan cannot know.
type DatasetPlan struct {
	Version       string          `json:"version"`
	Algorithm     string          `json:"algorithm"`
	Dataset       Dataset         `json:"dataset"`
	DatasetSHA256 string          `json:"dataset_sha256"`
	Shards        []DatasetShard  `json:"shards"`
	Groups        []DatasetGroup  `json:"groups"`
	Sources       []DatasetSource `json:"sources"`
}

// DatasetShard identifies one shard manifest before and after pinning.
type DatasetShard struct {
	Path         string `json:"path"`
	ID           string `json:"id"`
	SHA256       string `json:"sha256"`
	Bytes        int    `json:"bytes"`
	PinnedSHA256 string `json:"pinned_sha256"`
	PinnedBytes  int    `json:"pinned_bytes"`
	Sources      int    `json:"sources"`
}

// DatasetGroup is one connected component of the union. A group may span
// shards and never spans partitions.
type DatasetGroup struct {
	ID        string   `json:"id"`
	Partition string   `json:"partition"`
	Pinned    bool     `json:"pinned"`
	Shards    []string `json:"shards"`
	Sources   int      `json:"sources"`
}

// DatasetSource maps one source to its shard and global group.
type DatasetSource struct {
	ID        string `json:"id"`
	Shard     string `json:"shard"`
	Group     string `json:"group"`
	Partition string `json:"partition"`
}

// LoadDatasetPlan decodes strict JSON; consumers recompute the plan from the
// shards before trusting it.
func LoadDatasetPlan(ctx context.Context, data []byte) (DatasetPlan, error) {
	var plan DatasetPlan
	limits := jsoninput.Limits{Array: MaxDatasetSources, Object: 64, Arrays: map[string]int{"shards": MaxShards}}
	if err := jsoninput.Decode(ctx, data, MaxArtifactBytes, &plan, limits); err != nil {
		return DatasetPlan{}, err
	}
	if plan.Version != DatasetVersion || plan.Algorithm != assignmentAlgorithm {
		return DatasetPlan{}, fmt.Errorf("unsupported dataset plan version or algorithm")
	}
	if err := plan.Dataset.validate(); err != nil {
		return DatasetPlan{}, err
	}
	return plan, nil
}

// LoadDataset decodes strict JSON and validates the dataset header.
func LoadDataset(ctx context.Context, data []byte) (Dataset, error) {
	var dataset Dataset
	if err := jsoninput.Decode(ctx, data, MaxDatasetBytes, &dataset, jsoninput.Limits{Array: MaxShards, Object: 64}); err != nil {
		return Dataset{}, err
	}
	if err := dataset.validate(); err != nil {
		return Dataset{}, err
	}
	return dataset, nil
}

func (d Dataset) validate() error {
	if err := d.validateHeader(); err != nil {
		return err
	}
	return d.validateShards()
}

func (d Dataset) validateHeader() error {
	if d.Version != DatasetVersion || !text(d.ID) || !text(d.Seed) || len(d.Seed) > 128 || !validHash(d.RuleClassesSHA256) {
		return fmt.Errorf("invalid dataset version, ID, seed, or rule-class hash")
	}
	if err := d.Weights.validate(); err != nil {
		return err
	}
	if err := extract.ValidatePolicy(d.Policy); err != nil {
		return err
	}
	return choices(d.UnitKinds, []string{"sentence", "paragraph", "fragment"}, 3)
}

func (d Dataset) validateShards() error {
	if len(d.Shards) == 0 || len(d.Shards) > MaxShards {
		return fmt.Errorf("shard count must be 1 through %d", MaxShards)
	}
	seen := make(map[string]bool, len(d.Shards))
	for _, shard := range d.Shards {
		if !ValidPath(shard.Path) || !validHash(shard.SHA256) || shard.Bytes <= 0 || shard.Bytes > MaxManifestBytes {
			return fmt.Errorf("invalid shard path, SHA-256, or byte count")
		}
		if seen[shard.Path] {
			return fmt.Errorf("duplicate shard %s", shard.Path)
		}
		seen[shard.Path] = true
	}
	return nil
}

// MakeDatasetPlan loads every shard manifest, checks that it repeats the
// dataset header, connects sources across shards, and assigns whole
// components with the existing algorithm. Explicit pins in any shard pin the
// whole global component; conflicting pins are errors.
func MakeDatasetPlan(ctx context.Context, dataset Dataset, shards map[string][]byte) (DatasetPlan, error) {
	if err := dataset.validate(); err != nil {
		return DatasetPlan{}, err
	}
	canonical, hash, err := canonicalDataset(dataset)
	if err != nil {
		return DatasetPlan{}, err
	}
	manifests, err := loadShards(ctx, canonical, shards)
	if err != nil {
		return DatasetPlan{}, err
	}
	union, owners, err := unionSources(ctx, canonical, manifests)
	if err != nil {
		return DatasetPlan{}, err
	}
	groups, err := componentsWithin(ctx, union, maxDatasetKeys)
	if err != nil {
		return DatasetPlan{}, err
	}
	plan := DatasetPlan{Version: DatasetVersion, Algorithm: assignmentAlgorithm, Dataset: canonical, DatasetSHA256: hash}
	plan.Groups, plan.Sources = datasetGroups(groups, owners)
	partitions := make(map[string]string, len(plan.Sources))
	for _, source := range plan.Sources {
		partitions[source.ID] = source.Partition
	}
	for _, shard := range canonical.Shards {
		record, err := pinShard(shard, manifests[shard.Path], partitions)
		if err != nil {
			return DatasetPlan{}, err
		}
		plan.Shards = append(plan.Shards, record)
	}
	return plan, ctx.Err()
}

// PinShards returns the pinned manifest bytes of every shard, exactly as the
// plan recorded them, so per-shard plans inherit the global assignment.
func PinShards(ctx context.Context, plan DatasetPlan, shards map[string][]byte) (map[string][]byte, error) {
	expected, err := MakeDatasetPlan(ctx, plan.Dataset, shards)
	if err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(plan, expected) {
		return nil, fmt.Errorf("dataset plan does not match its shards and assignment algorithm")
	}
	manifests, err := loadShards(ctx, plan.Dataset, shards)
	if err != nil {
		return nil, err
	}
	partitions := make(map[string]string, len(plan.Sources))
	for _, source := range plan.Sources {
		partitions[source.ID] = source.Partition
	}
	pinned := make(map[string][]byte, len(plan.Shards))
	for _, shard := range plan.Shards {
		data, err := pinnedManifest(manifests[shard.Path], partitions)
		if err != nil {
			return nil, err
		}
		if hashBytes(data) != shard.PinnedSHA256 {
			return nil, fmt.Errorf("pinned shard %s does not reproduce its recorded digest", shard.Path)
		}
		pinned[shard.Path] = data
	}
	return pinned, ctx.Err()
}

// VerifyDatasetPlan recomputes the plan from the curator's original shard
// files and rejects any difference. Pinned copies are checked separately.
func VerifyDatasetPlan(ctx context.Context, plan DatasetPlan, shards map[string][]byte) error {
	expected, err := MakeDatasetPlan(ctx, plan.Dataset, shards)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(plan, expected) {
		return fmt.Errorf("dataset plan does not match its shards and assignment algorithm")
	}
	return nil
}

// VerifyPinnedShards checks that every pinned shard file carries exactly the
// bytes the plan recorded, so a per-shard plan built from it inherits the
// global assignment.
func VerifyPinnedShards(ctx context.Context, plan DatasetPlan, pinned map[string][]byte) error {
	if plan.Version != DatasetVersion || len(plan.Shards) == 0 {
		return fmt.Errorf("invalid dataset plan")
	}
	for _, shard := range plan.Shards {
		if err := ctx.Err(); err != nil {
			return err
		}
		data, found := pinned[shard.Path]
		if !found || len(data) != shard.PinnedBytes || hashBytes(data) != shard.PinnedSHA256 {
			return fmt.Errorf("pinned shard %s is missing or differs from its recorded digest", shard.Path)
		}
	}
	return nil
}

func canonicalDataset(dataset Dataset) (Dataset, string, error) {
	data, err := json.Marshal(dataset)
	if err != nil {
		return Dataset{}, "", err
	}
	var result Dataset
	if err := json.Unmarshal(data, &result); err != nil {
		return Dataset{}, "", err
	}
	slices.Sort(result.UnitKinds)
	slices.Sort(result.Policy.Contexts)
	for _, language := range result.Policy.Languages {
		slices.Sort(language.Contexts)
	}
	slices.SortFunc(result.Shards, func(a, b Notice) int { return strings.Compare(a.Path, b.Path) })
	hash, err := digest(result)
	if err != nil {
		return Dataset{}, "", err
	}
	return result, hash, nil
}

func loadShards(ctx context.Context, dataset Dataset, shards map[string][]byte) (map[string]Manifest, error) {
	manifests := make(map[string]Manifest, len(dataset.Shards))
	for _, shard := range dataset.Shards {
		data, found := shards[shard.Path]
		if !found || len(data) != shard.Bytes || hashBytes(data) != shard.SHA256 {
			return nil, fmt.Errorf("shard %s is missing or differs from its declared bytes", shard.Path)
		}
		manifest, err := LoadManifest(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("shard %s: %w", shard.Path, err)
		}
		if err := sameHeader(dataset, manifest); err != nil {
			return nil, fmt.Errorf("shard %s: %w", shard.Path, err)
		}
		canonical, err := canonicalManifest(manifest)
		if err != nil {
			return nil, err
		}
		manifests[shard.Path] = canonical
	}
	return manifests, nil
}

func sameHeader(dataset Dataset, manifest Manifest) error {
	header := Dataset{Version: DatasetVersion, ID: dataset.ID, Seed: manifest.Seed, Weights: manifest.Weights,
		Policy: manifest.Policy, UnitKinds: manifest.UnitKinds, RuleClassesSHA256: dataset.RuleClassesSHA256,
		Shards: dataset.Shards}
	canonical, _, err := canonicalDataset(header)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(canonical, dataset) {
		return fmt.Errorf("shard header differs from the dataset seed, weights, extraction policy, or unit kinds")
	}
	return nil
}

// unionSources joins every shard's sources into one manifest in ID order and
// records which shard owns each source. IDs are unique across shards; paths
// and notices are checked within each shard, because every shard may have
// its own source root.
func unionSources(ctx context.Context, dataset Dataset, manifests map[string]Manifest) (Manifest, map[string]string, error) {
	union := Manifest{Version: Version, ID: dataset.ID, Seed: dataset.Seed, Weights: dataset.Weights,
		Policy: dataset.Policy, UnitKinds: dataset.UnitKinds}
	owners := make(map[string]string)
	for _, shard := range dataset.Shards {
		manifest := manifests[shard.Path]
		for _, source := range manifest.Sources {
			if err := ctx.Err(); err != nil {
				return Manifest{}, nil, err
			}
			if _, found := owners[source.ID]; found {
				return Manifest{}, nil, fmt.Errorf("source %s appears in more than one shard", source.ID)
			}
			owners[source.ID] = shard.Path
			union.Sources = append(union.Sources, source)
		}
		if len(union.Sources) > MaxDatasetSources {
			return Manifest{}, nil, fmt.Errorf("dataset exceeds %d sources", MaxDatasetSources)
		}
	}
	slices.SortFunc(union.Sources, func(a, b Source) int { return strings.Compare(a.ID, b.ID) })
	return union, owners, nil
}

func datasetGroups(groups []Group, owners map[string]string) ([]DatasetGroup, []DatasetSource) {
	records := make([]DatasetGroup, 0, len(groups))
	sources := make([]DatasetSource, 0, len(owners))
	for _, group := range groups {
		record := DatasetGroup{ID: group.ID, Partition: group.Partition, Pinned: group.Pinned, Sources: len(group.Sources)}
		for _, id := range group.Sources {
			record.Shards = append(record.Shards, owners[id])
			sources = append(sources, DatasetSource{ID: id, Shard: owners[id], Group: group.ID, Partition: group.Partition})
		}
		slices.Sort(record.Shards)
		record.Shards = slices.Compact(record.Shards)
		records = append(records, record)
	}
	slices.SortFunc(sources, func(a, b DatasetSource) int { return strings.Compare(a.ID, b.ID) })
	return records, sources
}

func pinShard(shard Notice, manifest Manifest, partitions map[string]string) (DatasetShard, error) {
	data, err := pinnedManifest(manifest, partitions)
	if err != nil {
		return DatasetShard{}, err
	}
	return DatasetShard{Path: shard.Path, ID: manifest.ID, SHA256: shard.SHA256, Bytes: shard.Bytes,
		PinnedSHA256: hashBytes(data), PinnedBytes: len(data), Sources: len(manifest.Sources)}, nil
}

// pinnedManifest sets every source's partition to its global assignment. The
// encoding is fixed so the plan can record the pinned digest.
func pinnedManifest(manifest Manifest, partitions map[string]string) ([]byte, error) {
	pinned := manifest
	pinned.Sources = slices.Clone(manifest.Sources)
	for i := range pinned.Sources {
		partition, found := partitions[pinned.Sources[i].ID]
		if !found {
			return nil, fmt.Errorf("source %s has no global assignment", pinned.Sources[i].ID)
		}
		pinned.Sources[i].Partition = partition
	}
	data, err := json.MarshalIndent(pinned, "", "  ")
	if err != nil {
		return nil, err
	}
	if len(data)+1 > MaxManifestBytes {
		return nil, fmt.Errorf("pinned shard %s exceeds the manifest byte limit", manifest.ID)
	}
	return append(data, '\n'), nil
}
