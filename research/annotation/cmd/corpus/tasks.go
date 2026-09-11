package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/generation"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
)

type tasksOptions struct {
	exclude    []string
	plan       string
	sources    string
	work       string
	cohort     string
	partitions string
	roles      string
	count      int
	seed       string
	protocol   string
	candidates []string
}

// runTasks samples the controlled experiment's tasks from the historical
// cohort: one candidate artifact at a time, reading each documented unit's
// source file from the cohort's checkout under the work directory.
func runTasks(ctx context.Context, args []string, _ io.Reader, output io.Writer) error {
	options, err := tasksFlags(args[1:])
	if err != nil {
		return err
	}
	plan, err := loadPlanFile(ctx, options.plan)
	if err != nil {
		return err
	}
	ecosystems, err := loadEcosystems(ctx, options.sources)
	if err != nil {
		return err
	}
	excluded, err := excludedTasks(ctx, options.exclude)
	if err != nil {
		return err
	}
	sampler, err := generation.NewSampler(generation.Options{Protocol: options.protocol, Seed: options.seed,
		Cohort: options.cohort, Partitions: strings.Split(options.partitions, ","), Roles: strings.Split(options.roles, ","),
		Count: options.count, Ecosystems: ecosystems, Excluded: excluded}, plan)
	if err != nil {
		return err
	}
	for _, path := range options.candidates {
		if err := addTaskCandidates(ctx, sampler, path, options); err != nil {
			return err
		}
	}
	tasks, err := sampler.Sample()
	if err != nil {
		return err
	}
	return writeResult(ctx, "tasks", tasks, output)
}

func addTaskCandidates(ctx context.Context, sampler *generation.Sampler, path string, options tasksOptions) error {
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return readLocalArtifact(path, corpus.MaxArtifactBytes, "candidate artifact")
	})
	if err != nil {
		return err
	}
	artifact, err := corpus.LoadArtifact(ctx, data)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if len(artifact.Units) == 0 {
		return nil
	}
	repository := artifact.Units[0].Unit.Source.RepositoryID
	root, err := os.OpenRoot(filepath.Join(options.work, options.cohort, strings.ReplaceAll(repository, "/", "__")))
	if err != nil {
		return err
	}
	read := func(name string) ([]byte, error) {
		return readLocalRootFile(root, filepath.FromSlash(name), corpus.MaxSourceBytes, "checkout file")
	}
	addErr := sampler.Add(ctx, artifact, read)
	closeErr := root.Close()
	if addErr != nil {
		return fmt.Errorf("%s: %w", path, addErr)
	}
	return closeErr
}

func loadPlanFile(ctx context.Context, path string) (corpus.DatasetPlan, error) {
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return readLocalArtifact(path, corpus.MaxArtifactBytes, "dataset plan")
	})
	if err != nil {
		return corpus.DatasetPlan{}, err
	}
	return corpus.LoadDatasetPlan(ctx, data)
}

// loadEcosystems maps repository names to ecosystems from a sources frame.
func loadEcosystems(ctx context.Context, path string) (map[string]string, error) {
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return readLocalArtifact(path, corpus.MaxManifestBytes, "sources frame")
	})
	if err != nil {
		return nil, err
	}
	var frame struct {
		Sources []struct {
			Name      string `json:"name"`
			Ecosystem string `json:"ecosystem"`
		} `json:"sources"`
	}
	if err := json.Unmarshal(data, &frame); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	ecosystems := make(map[string]string, len(frame.Sources))
	for _, source := range frame.Sources {
		if source.Name == "" || source.Ecosystem == "" {
			return nil, fmt.Errorf("%s: every source needs a name and an ecosystem", path)
		}
		ecosystems[source.Name] = source.Ecosystem
	}
	return ecosystems, nil
}

func tasksFlags(args []string) (tasksOptions, error) {
	var options tasksOptions
	flags := flag.NewFlagSet("corpus tasks", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.plan, "plan", "", "Dataset plan naming each source's partition")
	flags.StringVar(&options.sources, "sources", "", "Sources frame naming each repository's ecosystem")
	flags.StringVar(&options.work, "work", "", "Directory holding the checkouts by cohort and repository slug")
	flags.StringVar(&options.cohort, "cohort", "historical", "Cohort the tasks come from")
	flags.StringVar(&options.partitions, "partitions", "training,development", "Comma-separated partitions the tasks come from")
	flags.StringVar(&options.roles, "roles", "comment", "Comma-separated roles of the documented units")
	flags.IntVar(&options.count, "count", 0, "Number of tasks to draw")
	flags.StringVar(&options.seed, "seed", "", "Sampling seed")
	flags.StringVar(&options.protocol, "protocol", "unswell-llm-patterns-v1", "Protocol the tasks serve")
	flags.Func("candidates", "Candidate artifact; repeat for every shard", func(value string) error {
		options.candidates = append(options.candidates, value)
		return nil
	})
	flags.Func("exclude-tasks", "Task set of an earlier run whose tasks leave the pool; repeat for a set", func(value string) error {
		options.exclude = append(options.exclude, value)
		return nil
	})
	if err := flags.Parse(args); err != nil {
		return tasksOptions{}, err
	}
	if flags.NArg() != 0 || options.plan == "" || options.sources == "" || options.work == "" || options.seed == "" ||
		options.count < 1 || len(options.candidates) == 0 || len(options.candidates) > corpus.MaxShards {
		return tasksOptions{}, fmt.Errorf("tasks requires --plan, --sources, --work, --seed, --count, and 1 through %d --candidates",
			corpus.MaxShards)
	}
	return options, nil
}

// excludedTasks collects the task IDs of earlier runs' task sets.
func excludedTasks(ctx context.Context, paths []string) (map[string]bool, error) {
	excluded := map[string]bool{}
	for _, path := range paths {
		tasks, _, err := loadTasks(ctx, path)
		if err != nil {
			return nil, err
		}
		for _, task := range tasks.Tasks {
			excluded[task.ID] = true
		}
	}
	return excluded, nil
}
