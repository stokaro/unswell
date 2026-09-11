package main

import (
	"context"
	"crypto/sha256"
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

// runRequests builds the request list of a run from the sampled tasks and
// the frozen prompt files, so the generator and the records share one text.
func runRequests(ctx context.Context, args []string, _ io.Reader, output io.Writer) error {
	flags := flag.NewFlagSet("corpus requests", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var tasksPath, promptsDir, run, operations string
	flags.StringVar(&tasksPath, "tasks", "", "Sampled task set")
	flags.StringVar(&promptsDir, "prompts", "", "Directory of the frozen prompt files")
	flags.StringVar(&run, "run", "", "Run identifier")
	flags.StringVar(&operations, "operations", "generate,polish", "Comma-separated operations")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 || tasksPath == "" || promptsDir == "" || run == "" {
		return fmt.Errorf("requests requires --tasks, --prompts, and --run")
	}
	tasks, tasksSHA, err := loadTasks(ctx, tasksPath)
	if err != nil {
		return err
	}
	prompts, err := loadPrompts(ctx, promptsDir)
	if err != nil {
		return err
	}
	requests, err := generation.BuildRequests(run, tasks, tasksSHA, prompts, strings.Split(operations, ","))
	if err != nil {
		return err
	}
	return writeResult(ctx, "requests", requests, output)
}

// loadPrompts reads the two frozen prompt conditions of the pilot.
func loadPrompts(ctx context.Context, directory string) ([]generation.PromptCondition, error) {
	prompts := []generation.PromptCondition{}
	for _, id := range []string{"neutral", "plain"} {
		file := id + "-v1.md"
		data, err := commandio.Await(ctx, func() ([]byte, error) {
			return readLocalArtifact(filepath.Join(directory, file), corpus.MaxManifestBytes, "prompt")
		})
		if err != nil {
			return nil, err
		}
		prompt, err := generation.ParsePrompt(id, file, data)
		if err != nil {
			return nil, err
		}
		prompts = append(prompts, prompt)
	}
	return prompts, nil
}

func loadTasks(ctx context.Context, path string) (generation.Tasks, string, error) {
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return readLocalArtifact(path, corpus.MaxArtifactBytes, "task set")
	})
	if err != nil {
		return generation.Tasks{}, "", err
	}
	var tasks generation.Tasks
	if err := json.Unmarshal(data, &tasks); err != nil {
		return generation.Tasks{}, "", fmt.Errorf("%s: %w", path, err)
	}
	if tasks.Version != generation.TasksVersion || len(tasks.Tasks) == 0 || len(tasks.Tasks) > generation.MaxTasks {
		return generation.Tasks{}, "", fmt.Errorf("%s: unsupported task set", path)
	}
	return tasks, fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

type generationsOptions struct {
	tasks, requests, responses, records, work, output, historical, suffix string
}

// runGenerations joins a run's responses with its requests and tasks and
// writes the generation record. Every complete response becomes a source of
// the controlled cohort. The import writes one file per response under the
// cohort's checkout directory and one shard manifest per repository.
func runGenerations(ctx context.Context, args []string, _ io.Reader, output io.Writer) error {
	options, err := generationsFlags(args[1:])
	if err != nil {
		return err
	}
	tasks, _, err := loadTasks(ctx, options.tasks)
	if err != nil {
		return err
	}
	var requests generation.Requests
	requestsSHA, err := loadJSON(ctx, options.requests, corpus.MaxArtifactBytes, &requests)
	if err != nil {
		return err
	}
	var responses generation.Responses
	if _, err := loadJSON(ctx, options.responses, corpus.MaxArtifactBytes, &responses); err != nil {
		return err
	}
	records, err := generation.BuildRecords(requests, requestsSHA, tasks, responses)
	if err != nil {
		return err
	}
	imported, err := generation.BuildManifests(records, tasks, importOptions(ctx, options))
	if err != nil {
		return err
	}
	if err := writeImports(ctx, options, imported); err != nil {
		return err
	}
	if err := writeJSONFile(options.records, records); err != nil {
		return err
	}
	return writeResult(ctx, "generations", records.Coverage, output)
}

func importOptions(ctx context.Context, options generationsOptions) generation.ImportOptions {
	historical := map[string]corpus.Acquisition{}
	return generation.ImportOptions{RecordsPath: options.records, Suffix: options.suffix,
		Historical: func(repository string) (corpus.Acquisition, error) {
			if record, found := historical[repository]; found {
				return record, nil
			}
			var record corpus.Acquisition
			file := filepath.Join(options.historical, strings.ReplaceAll(repository, "/", "__")+".json")
			if _, err := loadJSON(ctx, file, corpus.MaxManifestBytes, &record); err != nil {
				return corpus.Acquisition{}, err
			}
			historical[repository] = record
			return record, nil
		},
		Notice: func(repository, notice string) ([]byte, error) {
			return readCheckoutNotice(filepath.Join(options.work, "historical", strings.ReplaceAll(repository, "/", "__")), notice)
		}}
}

func readCheckoutNotice(directory, notice string) ([]byte, error) {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	data, readErr := readLocalRootFile(root, notice, corpus.MaxSourceBytes, "notice")
	closeErr := root.Close()
	if readErr != nil {
		return nil, readErr
	}
	return data, closeErr
}

func writeImports(ctx context.Context, options generationsOptions, imported []generation.Imported) error {
	for _, item := range imported {
		if err := ctx.Err(); err != nil {
			return err
		}
		directory := filepath.Join(options.work, "controlled", item.Slug)
		for name, data := range item.Files {
			path := filepath.Join(directory, filepath.FromSlash(name))
			if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
				return err
			}
			if err := os.WriteFile(path, data, 0o600); err != nil {
				return err
			}
		}
		if err := writeJSONFile(filepath.Join(options.output, item.Manifest.ID+".json"), item.Manifest); err != nil {
			return err
		}
	}
	return nil
}

func loadJSON(ctx context.Context, path string, maximum int, target any) (string, error) {
	data, err := commandio.Await(ctx, func() ([]byte, error) { return readLocalArtifact(path, maximum, "file") })
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return "", fmt.Errorf("%s: %w", path, err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

func writeJSONFile(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

func generationsFlags(args []string) (generationsOptions, error) {
	var options generationsOptions
	flags := flag.NewFlagSet("corpus generations", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.tasks, "tasks", "", "Sampled task set")
	flags.StringVar(&options.requests, "requests", "", "Request list of the run")
	flags.StringVar(&options.responses, "responses", "", "Raw responses of the run")
	flags.StringVar(&options.records, "records", "", "Generation record to write")
	flags.StringVar(&options.work, "work", "", "Directory holding the checkouts by cohort and repository slug")
	flags.StringVar(&options.output, "output", "", "Directory for the controlled cohort's shard manifests")
	flags.StringVar(&options.historical, "historical", "", "Directory of the historical cohort's acquisition records")
	flags.StringVar(&options.suffix, "shard-suffix", "", "Suffix of every shard ID, so a later run's shards sit beside an earlier run's")
	if err := flags.Parse(args); err != nil {
		return generationsOptions{}, err
	}
	if flags.NArg() != 0 || options.tasks == "" || options.requests == "" || options.responses == "" || options.records == "" ||
		options.work == "" || options.output == "" || options.historical == "" {
		return generationsOptions{}, fmt.Errorf("generations requires --tasks, --requests, --responses, --records, --work, --output, " +
			"and --historical")
	}
	return options, nil
}
