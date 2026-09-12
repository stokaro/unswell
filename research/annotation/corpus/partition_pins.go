package corpus

import (
	"context"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

// PartitionPinsVersion identifies the partition pin file.
const PartitionPinsVersion = "unswell-partition-pins-v1"

// Partition pin limits bound one pin file.
const (
	MaxPartitionPinBytes = 1 << 20
	MaxPartitionPins     = 4096
)

// PartitionPins fixes the partition of whole repositories before grouping.
// The hash assignment moves a component when a later run or cohort adds
// keys to it; a pin keeps the global split where the protocol froze it.
// SHA256 is the digest of the file bytes and enters the dataset plan.
type PartitionPins struct {
	Format       string            `json:"format"`
	DecidedOn    string            `json:"decided_on"`
	Protocol     string            `json:"protocol"`
	Scope        string            `json:"scope,omitempty"`
	Repositories map[string]string `json:"repositories"`
	SHA256       string            `json:"-"`
}

// LoadPartitionPins decodes strict JSON and checks the format, the entry
// count, and the partition names. Keys are repository names as the shard
// manifests carry them.
func LoadPartitionPins(ctx context.Context, data []byte) (PartitionPins, error) {
	var pins PartitionPins
	if err := jsoninput.Decode(ctx, data, MaxPartitionPinBytes, &pins, jsoninput.Limits{Object: MaxPartitionPins}); err != nil {
		return PartitionPins{}, err
	}
	if err := pins.validate(); err != nil {
		return PartitionPins{}, err
	}
	pins.SHA256 = hashBytes(data)
	return pins, nil
}

func (p PartitionPins) validate() error {
	if p.Format != PartitionPinsVersion || !text(p.DecidedOn) || !text(p.Protocol) {
		return fmt.Errorf("partition pins require the %s format, a decision date, and a protocol", PartitionPinsVersion)
	}
	if len(p.Repositories) == 0 || len(p.Repositories) > MaxPartitionPins {
		return fmt.Errorf("partition pins name 1 through %d repositories", MaxPartitionPins)
	}
	for repository, partition := range p.Repositories {
		if !text(repository) {
			return fmt.Errorf("partition pins name an empty or unbounded repository")
		}
		if !slices.Contains(partitions(), partition) {
			return fmt.Errorf("repository %s pins unknown partition %q", repository, partition)
		}
	}
	return nil
}

// apply sets the pinned partition on every source of a pinned repository.
// It returns the count of repositories that matched a source and, sorted,
// the repositories that matched none. A source that already carries another
// pin is an error naming the source and both values.
func (p *PartitionPins) apply(union *Manifest) (int, []string, error) {
	matched := make(map[string]bool, len(p.Repositories))
	for i := range union.Sources {
		source := &union.Sources[i]
		partition, found := p.Repositories[source.Repository]
		if !found {
			continue
		}
		if source.Partition != "" && source.Partition != partition {
			return 0, nil, fmt.Errorf("source %s pins partition %s while the partition file pins %s to %s",
				source.ID, source.Partition, source.Repository, partition)
		}
		source.Partition = partition
		matched[source.Repository] = true
	}
	var unmatched []string
	for repository := range p.Repositories {
		if !matched[repository] {
			unmatched = append(unmatched, repository)
		}
	}
	slices.Sort(unmatched)
	return len(matched), unmatched, nil
}

// pin applies the partition pins to the union before grouping and records
// their digest and effect on the plan. Nil pins leave the plan unpinned.
func (plan *DatasetPlan) pin(union *Manifest, pins *PartitionPins) error {
	if pins == nil {
		return nil
	}
	if !validHash(pins.SHA256) {
		return fmt.Errorf("partition pins carry no file digest")
	}
	pinned, unmatched, err := pins.apply(union)
	if err != nil {
		return err
	}
	plan.PartitionPinsSHA256, plan.PinnedRepositories, plan.UnmatchedPins = pins.SHA256, pinned, unmatched
	return nil
}
