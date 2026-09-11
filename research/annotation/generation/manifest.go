package generation

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

// Imported is the controlled cohort's share of one repository: the files to
// write under the cohort's checkout directory and the shard manifest that
// names them.
type Imported struct {
	Repository string
	Slug       string
	Files      map[string][]byte
	Manifest   corpus.Manifest
}

// ImportOptions names the run and where its records live, and supplies the
// historical acquisition record and notice bytes of each repository.
type ImportOptions struct {
	RecordsPath string
	Historical  func(repository string) (corpus.Acquisition, error)
	Notice      func(repository, notice string) ([]byte, error)
}

// BuildManifests turns every complete response into a source of the
// controlled cohort. A source takes the repository, topic, purpose, rights,
// and notices of its task. Its origin names the generation record, and its
// task link names the task. Refused, truncated, and failed responses produce
// no source; they stay in the record as coverage.
func BuildManifests(generation Generation, tasks Tasks, options ImportOptions) ([]Imported, error) {
	byTask := map[string]Task{}
	for _, task := range tasks.Tasks {
		byTask[task.ID] = task
	}
	repositories := map[string]*Imported{}
	for _, record := range generation.Records {
		if record.Status != "complete" {
			continue
		}
		task, found := byTask[record.TaskID]
		if !found {
			return nil, fmt.Errorf("record %s names task %s outside the task set", record.ResponseID, record.TaskID)
		}
		imported, err := repositoryImport(repositories, task.Repository, options)
		if err != nil {
			return nil, err
		}
		path := "generated/" + record.ResponseID + ".md"
		data := []byte(record.Text + "\n")
		imported.Files[path] = data
		imported.Manifest.Sources = append(imported.Manifest.Sources, controlledSource(generation, record, task, imported, path, data, options))
	}
	result := make([]Imported, 0, len(repositories))
	for _, imported := range repositories {
		sort.Slice(imported.Manifest.Sources, func(a, b int) bool { return imported.Manifest.Sources[a].ID < imported.Manifest.Sources[b].ID })
		result = append(result, *imported)
	}
	sort.Slice(result, func(a, b int) bool { return result[a].Repository < result[b].Repository })
	return result, nil
}

func repositoryImport(repositories map[string]*Imported, repository string, options ImportOptions) (*Imported, error) {
	if imported, found := repositories[repository]; found {
		return imported, nil
	}
	record, err := options.Historical(repository)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", repository, err)
	}
	slug := strings.ReplaceAll(repository, "/", "__")
	imported := &Imported{Repository: repository, Slug: slug, Files: map[string][]byte{},
		Manifest: corpus.Manifest{Version: corpus.Version, ID: "controlled-" + slug, Seed: record.Manifest.Seed,
			Weights: record.Manifest.Weights, Policy: record.Manifest.Policy, UnitKinds: slices.Clone(record.Manifest.UnitKinds),
			Sources: []corpus.Source{}}}
	for _, notice := range record.Repository.Notices {
		data, err := options.Notice(repository, notice)
		if err != nil {
			return nil, fmt.Errorf("%s notice %s: %w", repository, notice, err)
		}
		imported.Files[notice] = data
	}
	repositories[repository] = imported
	return imported, nil
}

func controlledSource(generation Generation, record Record, task Task, imported *Imported, path string, data []byte,
	options ImportOptions,
) corpus.Source {
	historical, _ := options.Historical(task.Repository)
	repository := historical.Repository
	label := "generated"
	if record.Operation == "polish" {
		label = "human_ai_edited"
	}
	notices := make([]corpus.Notice, 0, len(repository.Notices))
	for _, notice := range repository.Notices {
		bytes := imported.Files[notice]
		notices = append(notices, corpus.Notice{Path: notice, SHA256: hash(bytes), Bytes: len(bytes)})
	}
	evidence := fmt.Sprintf("Response %s of run %s: %s under prompt %s by %s (%s); record %s",
		record.ResponseID, generation.Run, record.Operation, record.Prompt, generation.Family, generation.Model, options.RecordsPath)
	return corpus.Source{ID: controlledID(generation.Run, task.Repository, path), Path: path, SHA256: hash(data), Bytes: len(data),
		Format: document.Markdown, ProseLanguage: "en", Repository: task.Repository, Document: task.Repository + "/" + path,
		Authors: []string{}, Templates: []string{}, Related: []string{}, GenerationTasks: []string{task.ID},
		Reference: "generation:" + generation.Run + "/" + record.ResponseID, Topic: repository.Topic, Purpose: repository.Purpose,
		Role: task.Role, Roles: []corpus.RoleRegion{},
		Origin: annotation.Origin{Label: label, Scope: "document", Evidence: evidence,
			GenerationRecord: options.RecordsPath + "#" + record.ResponseID},
		Rights: annotation.Rights{License: repository.Rights.License,
			Evidence: fmt.Sprintf("Generated from the facts of %s at %s under %s; the response is this project's own output",
				task.Repository, repository.Commit, repository.Rights.License),
			AllowedUses: slices.Clone(repository.Rights.AllowedUses)},
		Notices: notices,
		Snapshot: &corpus.Snapshot{Date: record.GeneratedOn, Confidence: "corroborated",
			Evidence: "Generation record " + options.RecordsPath + " dates the response", Cohort: "controlled"}}
}

func controlledID(run, repository, path string) string {
	sum := sha256.Sum256([]byte("controlled\x00" + run + "\x00" + repository + "\x00" + path))
	return fmt.Sprintf("%x", sum[:12])
}

func hash(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }
