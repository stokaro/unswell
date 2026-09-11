package patterns

import (
	"fmt"
	"strings"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/generation"
)

// FrequencyPairing says which generation runs relabel the strata: how many
// tasks and responses were registered and the cohort the originals come from.
type FrequencyPairing struct {
	Cohort    string `json:"cohort"`
	Tasks     int    `json:"tasks"`
	Responses int    `json:"responses"`
}

// pairing relabels the sentence units of registered tasks and responses, so a
// contrast can put the responses against the sentences of the paragraphs they
// answer instead of a whole cohort.
type pairing struct {
	cohort    string
	tasks     int
	originals map[string]map[string]bool
	responses map[string]string
}

// Pair registers the tasks and records of one generation run; call it before
// Add, once per run. A response whose task is registered counts under
// "controlled-<operation>", and the sentences of a task's original paragraph
// count under "<cohort>-paired". Every other unit keeps its cohort.
func (f *Frequencies) Pair(tasks generation.Tasks, records generation.Generation) error {
	if f.pairs == nil {
		f.pairs = &pairing{originals: map[string]map[string]bool{}, responses: map[string]string{}}
	}
	if f.pairs.cohort != "" && f.pairs.cohort != tasks.Cohort {
		return fmt.Errorf("paired tasks must share one cohort, not %s and %s", f.pairs.cohort, tasks.Cohort)
	}
	f.pairs.cohort = tasks.Cohort
	known := map[string]bool{}
	for _, task := range tasks.Tasks {
		known[task.ID] = true
		texts := f.pairs.originals[task.SourceID]
		if texts == nil {
			texts = map[string]bool{}
			f.pairs.originals[task.SourceID] = texts
		}
		texts[squash(task.Text)] = true
	}
	f.pairs.tasks += len(tasks.Tasks)
	for _, record := range records.Records {
		if !known[record.TaskID] {
			return fmt.Errorf("record %s names task %s outside its task set", record.ResponseID, record.TaskID)
		}
		f.pairs.responses[record.ResponseID] = record.Operation
	}
	return nil
}

// label names the cohort a unit counts under: the paired label when its
// response or original paragraph is registered, its own cohort otherwise.
func (p *pairing) label(candidate corpus.Candidate) string {
	if p == nil {
		return candidate.Cohort
	}
	if operation, found := p.responses[responseID(candidate.Unit.Source.Reference)]; found {
		return candidate.Cohort + "-" + operation
	}
	if candidate.Cohort == p.cohort && p.originals[candidate.SourceID][squash(candidate.Unit.Context)] {
		return candidate.Cohort + "-paired"
	}
	return candidate.Cohort
}

func (p *pairing) summary() *FrequencyPairing {
	if p == nil {
		return nil
	}
	return &FrequencyPairing{Cohort: p.cohort, Tasks: p.tasks, Responses: len(p.responses)}
}

// responseID reads the response of a generated source from its reference,
// "generation:<run>/<response id>"; any other reference names no response.
func responseID(reference string) string {
	if !strings.HasPrefix(reference, "generation:") {
		return ""
	}
	return reference[strings.LastIndex(reference, "/")+1:]
}

// squash joins the words of a text with single spaces, so a paragraph and
// the context of its sentences compare equal whatever the source spacing.
func squash(text string) string { return strings.Join(strings.Fields(text), " ") }
