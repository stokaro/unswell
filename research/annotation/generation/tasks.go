// Package generation builds the controlled-experiment artifacts of the
// LLM-associated pattern protocol: sampled tasks with deterministic fact
// sheets, generation records, and the controlled cohort's manifests. It calls
// no model; the generation itself happens in research tooling outside this
// package, and every output comes back as a saved record.
package generation

import (
	"context"
	"crypto/sha256"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/stokaro/unswell/research/annotation/corpus"
)

// TasksVersion identifies the sampled task set.
const TasksVersion = "unswell-tasks-v1"

// MaxTasks is the protocol's pilot cap on tasks.
const MaxTasks = 200

// Task is one documentation unit of the historical cohort with the fact
// sheet a generator sees instead of the original wording.
type Task struct {
	ID         string    `json:"id"`
	SourceID   string    `json:"source_id"`
	UnitID     string    `json:"unit_id"`
	GroupID    string    `json:"group_id"`
	Partition  string    `json:"partition"`
	Cohort     string    `json:"cohort"`
	Repository string    `json:"repository"`
	Ecosystem  string    `json:"ecosystem"`
	Path       string    `json:"path"`
	Role       string    `json:"role"`
	Words      int       `json:"words"`
	Text       string    `json:"text"`
	TextSHA256 string    `json:"text_sha256"`
	FactSheet  FactSheet `json:"fact_sheet"`
}

// Stratum records the sampling of one ecosystem and role.
type Stratum struct {
	Ecosystem string `json:"ecosystem"`
	Role      string `json:"role"`
	Eligible  int    `json:"eligible"`
	Selected  int    `json:"selected"`
}

// Tasks is the sampled task set with its sampling record.
type Tasks struct {
	Version    string    `json:"version"`
	Protocol   string    `json:"protocol"`
	Seed       string    `json:"seed"`
	Cohort     string    `json:"cohort"`
	Partitions []string  `json:"partitions"`
	Roles      []string  `json:"roles"`
	Requested  int       `json:"requested"`
	Excluded   int       `json:"excluded,omitempty"`
	Strata     []Stratum `json:"strata"`
	Tasks      []Task    `json:"tasks"`
}

// MinWords is the shortest documentation unit a task may use: a shorter
// paragraph is a fragment of code or a label, not a document to write.
const MinWords = 12

// Options fixes the sampling before any candidate is read.
type Options struct {
	Protocol   string
	Seed       string
	Cohort     string
	Partitions []string
	Roles      []string
	Count      int
	Ecosystems map[string]string
	// Excluded names the tasks of earlier runs by ID. They leave the eligible
	// pool before allocation, so a later run draws new tasks under the same
	// seed instead of regenerating the same texts.
	Excluded map[string]bool
}

// Sampler accumulates eligible units one candidate artifact at a time.
type Sampler struct {
	options   Options
	partition map[string]string
	eligible  map[string][]Task
	excluded  int
}

// NewSampler validates the options and the dataset plan the partitions come
// from.
func NewSampler(options Options, plan corpus.DatasetPlan) (*Sampler, error) {
	if options.Protocol == "" || options.Seed == "" || options.Cohort == "" || len(options.Partitions) == 0 ||
		len(options.Roles) == 0 || options.Count < 1 || options.Count > MaxTasks {
		return nil, fmt.Errorf("task sampling needs a protocol, seed, cohort, partitions, roles, and 1 through %d tasks", MaxTasks)
	}
	partition := make(map[string]string, len(plan.Sources))
	for _, source := range plan.Sources {
		partition[source.ID] = source.Partition
	}
	return &Sampler{options: options, partition: partition, eligible: map[string][]Task{}}, nil
}

// Reader returns the bytes of one checkout file by its repository-relative
// path, or an error when the file cannot be read.
type Reader func(path string) ([]byte, error)

// Add indexes the eligible paragraphs of one candidate artifact. A unit is
// eligible when it is a paragraph of an admitted role in the cohort and an
// admitted partition, and when the source bytes after it yield a fact
// sheet. The reader loads files on demand.
func (s *Sampler) Add(ctx context.Context, artifact corpus.Artifact, read Reader) error {
	if artifact.Version != corpus.Version {
		return fmt.Errorf("unsupported candidate artifact version")
	}
	cache := map[string][]byte{}
	for _, candidate := range artifact.Units {
		if err := ctx.Err(); err != nil {
			return err
		}
		task, ok, err := s.eligibleTask(candidate, read, cache)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		if s.options.Excluded[task.ID] {
			s.excluded++
			continue
		}
		key := task.Ecosystem + "\x00" + task.Role
		s.eligible[key] = append(s.eligible[key], task)
	}
	return nil
}

// admits reports whether a candidate is a paragraph of an admitted role,
// cohort, partition, and length whose repository has a known ecosystem.
func (s *Sampler) admits(candidate corpus.Candidate) bool {
	unit := candidate.Unit
	if unit.Kind != "paragraph" || candidate.Cohort != s.options.Cohort || !slices.Contains(s.options.Roles, unit.Role) ||
		candidate.Words < MinWords || len(unit.Source.Segments) == 0 {
		return false
	}
	if !slices.Contains(s.options.Partitions, s.partition[candidate.SourceID]) {
		return false
	}
	_, known := s.options.Ecosystems[unit.Source.RepositoryID]
	return known
}

func (s *Sampler) eligibleTask(candidate corpus.Candidate, read Reader, cache map[string][]byte) (Task, bool, error) {
	if !s.admits(candidate) {
		return Task{}, false, nil
	}
	unit := candidate.Unit
	partition := s.partition[candidate.SourceID]
	ecosystem := s.options.Ecosystems[unit.Source.RepositoryID]
	path := strings.TrimPrefix(unit.Source.DocumentID, unit.Source.RepositoryID+"/")
	source, cached := cache[path]
	if !cached {
		data, err := read(path)
		if err != nil {
			return Task{}, false, fmt.Errorf("%s: %w", path, err)
		}
		source, cache[path] = data, data
	}
	end := unit.Source.Segments[len(unit.Source.Segments)-1].End
	sheet, ok := ExtractFactSheet(source, end, unit.Text)
	if !ok {
		return Task{}, false, nil
	}
	return Task{ID: taskID(candidate.SourceID, unit.ID), SourceID: candidate.SourceID, UnitID: unit.ID,
		GroupID: candidate.GroupID, Partition: partition, Cohort: candidate.Cohort, Repository: unit.Source.RepositoryID,
		Ecosystem: ecosystem, Path: path, Role: unit.Role, Words: candidate.Words, Text: unit.Text,
		TextSHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(unit.Text))), FactSheet: sheet}, true, nil
}

func taskID(sourceID, unitID string) string {
	sum := sha256.Sum256([]byte("task\x00" + sourceID + "\x00" + unitID))
	return fmt.Sprintf("t%x", sum[:8])
}

// Sample draws the requested number of tasks. Every stratum with an eligible
// unit gets at least one task; the rest are allocated in proportion to the
// eligible counts, largest remainders first. Within a stratum the order is a
// seeded hash of the task ID, so the draw depends on nothing but the seed and
// the eligible set.
func (s *Sampler) Sample() (Tasks, error) {
	keys := make([]string, 0, len(s.eligible))
	for key := range s.eligible {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	total := 0
	for _, key := range keys {
		total += len(s.eligible[key])
	}
	if total == 0 {
		return Tasks{}, fmt.Errorf("no eligible unit for the requested cohort, partitions, and roles")
	}
	result := Tasks{Version: TasksVersion, Protocol: s.options.Protocol, Seed: s.options.Seed, Cohort: s.options.Cohort,
		Partitions: slices.Clone(s.options.Partitions), Roles: slices.Clone(s.options.Roles), Requested: s.options.Count,
		Excluded: s.excluded, Strata: []Stratum{}, Tasks: []Task{}}
	allocation := allocate(keys, s.eligible, min(s.options.Count, total))
	for _, key := range keys {
		pool := slices.Clone(s.eligible[key])
		sort.Slice(pool, func(a, b int) bool { return s.rank(pool[a].ID) < s.rank(pool[b].ID) })
		chosen := pool[:allocation[key]]
		ecosystem, role, _ := strings.Cut(key, "\x00")
		result.Strata = append(result.Strata, Stratum{Ecosystem: ecosystem, Role: role, Eligible: len(pool), Selected: len(chosen)})
		result.Tasks = append(result.Tasks, chosen...)
	}
	sort.Slice(result.Tasks, func(a, b int) bool { return result.Tasks[a].ID < result.Tasks[b].ID })
	return result, nil
}

func (s *Sampler) rank(id string) string {
	sum := sha256.Sum256([]byte(s.options.Seed + "\x00" + id))
	return fmt.Sprintf("%x", sum)
}

// allocate gives every stratum one task, then the remainder in proportion to
// its eligible count, and never more than it holds.
func allocate(keys []string, eligible map[string][]Task, count int) map[string]int {
	allocation := map[string]int{}
	remaining := count
	total := 0
	for _, key := range keys {
		if remaining > 0 {
			allocation[key] = 1
			remaining--
		}
		total += len(eligible[key])
	}
	type share struct {
		key       string
		remainder float64
	}
	shares := []share{}
	for _, key := range keys {
		exact := float64(remaining) * float64(len(eligible[key])) / float64(total)
		whole := int(exact)
		allocation[key] = min(allocation[key]+whole, len(eligible[key]))
		shares = append(shares, share{key, exact - float64(whole)})
	}
	sort.SliceStable(shares, func(a, b int) bool { return shares[a].remainder > shares[b].remainder })
	given := 0
	for _, item := range allocation {
		given += item
	}
	for _, item := range shares {
		if given >= count {
			break
		}
		if allocation[item.key] < len(eligible[item.key]) {
			allocation[item.key]++
			given++
		}
	}
	return allocation
}
