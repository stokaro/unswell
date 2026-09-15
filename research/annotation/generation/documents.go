package generation

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

// DocumentScope identifies a task whose original is the complete source file.
const DocumentScope = "document"

// DocumentFact is a curator's factual note with the original byte range that
// supports it. The range is evidence, not text substituted into a prompt.
type DocumentFact struct {
	Text string        `json:"text"`
	Span document.Span `json:"span"`
}

// DocumentBrief supplies facts independently of the source's prose style.
// Reviewer records human or agent curation; neither implies a quality label.
type DocumentBrief struct {
	Version      string         `json:"version"`
	SourceID     string         `json:"source_id"`
	SourceSHA256 string         `json:"source_sha256"`
	Reviewer     string         `json:"reviewer"`
	Purpose      string         `json:"purpose"`
	Facts        []DocumentFact `json:"facts"`
	SHA256       string         `json:"sha256"`
}

// Seal validates a brief against its source and returns an owned copy with a
// digest. Changing facts or evidence changes the digest, not the source bytes.
func (b DocumentBrief) Seal(source []byte) (DocumentBrief, error) {
	if !b.validHeader(source) {
		return DocumentBrief{}, fmt.Errorf("document brief requires a matching source, purpose, reviewer, and 1 through 200 facts")
	}
	for _, fact := range b.Facts {
		if !fact.valid(source) {
			return DocumentBrief{}, fmt.Errorf("document fact needs text and a nonempty UTF-8 source range")
		}
	}
	b.Facts = slices.Clone(b.Facts)
	b.SHA256 = ""
	data, err := json.Marshal(b)
	if err != nil {
		return DocumentBrief{}, err
	}
	b.SHA256 = hash(data)
	return b, nil
}

func (b DocumentBrief) validHeader(source []byte) bool {
	return b.Version == "unswell-document-brief-v1" && b.SourceID != "" && b.SourceSHA256 == hash(source) &&
		slices.Contains([]string{"human", "agent"}, b.Reviewer) && validBriefText(b.Purpose) &&
		len(b.Facts) > 0 && len(b.Facts) <= 200 && utf8.Valid(source)
}

func validBriefText(text string) bool {
	return strings.TrimSpace(text) != "" && utf8.ValidString(text) && len(text) <= 8192
}

func (f DocumentFact) valid(source []byte) bool {
	return validBriefText(f.Text) && f.Span.Start >= 0 && f.Span.End > f.Span.Start && f.Span.End <= len(source) &&
		utf8.Valid(source[f.Span.Start:f.Span.End])
}

// Text renders only the purpose and factual notes, without the original prose
// or curation metadata. Source locations remain in the saved task artifact.
func (b DocumentBrief) Text() string {
	var text strings.Builder
	text.WriteString("Purpose: " + b.Purpose + "\n\nFacts:\n")
	for _, fact := range b.Facts {
		text.WriteString("- " + fact.Text + "\n")
	}
	return text.String()
}

// AddDocuments admits complete Markdown or plain-text sources with supplied
// briefs. It reuses the extraction artifact and the sampler's global groups.
// Paragraph and fragment words count once; nested sentences do not count again.
func (s *Sampler) AddDocuments(ctx context.Context, artifact corpus.Artifact, read Reader, briefs map[string]DocumentBrief) error {
	if artifact.Version != corpus.Version || artifact.Status != "unlabeled_candidates" {
		return fmt.Errorf("document tasks require a complete candidate artifact")
	}
	words := documentWords(artifact)
	for _, source := range artifact.Plan.Manifest.Sources {
		if err := ctx.Err(); err != nil {
			return err
		}
		brief, found := briefs[source.ID]
		if !found || !s.admitsDocument(source, words[source.ID]) {
			continue
		}
		task, err := s.uniqueDocumentTask(source, words[source.ID], brief, read)
		if err != nil {
			return err
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

func (s *Sampler) uniqueDocumentTask(source corpus.Source, words int, brief DocumentBrief, read Reader) (Task, error) {
	if s.documents[source.ID] {
		return Task{}, fmt.Errorf("document task source %s was already added", source.ID)
	}
	task, err := s.documentTask(source, words, brief, read)
	if err != nil {
		return Task{}, err
	}
	s.documents[source.ID] = true
	return task, nil
}

func documentWords(artifact corpus.Artifact) map[string]int {
	words := map[string]int{}
	for _, unit := range artifact.Units {
		if unit.Unit.Kind == "paragraph" || unit.Unit.Kind == "fragment" {
			words[unit.SourceID] += unit.Words
		}
	}
	return words
}

func (s *Sampler) admitsDocument(source corpus.Source, words int) bool {
	return (source.Format == document.Markdown || source.Format == document.Plain) && source.Snapshot != nil &&
		source.Snapshot.Cohort == s.options.Cohort && s.options.Ecosystems[source.Repository] != "" &&
		slices.Contains(s.options.Roles, source.Role) && slices.Contains(s.options.Partitions, s.partition[source.ID]) &&
		words >= EligibleFloor(s.options) && (s.options.MaxWords == 0 || words <= s.options.MaxWords)
}

func (s *Sampler) documentTask(source corpus.Source, words int, brief DocumentBrief, read Reader) (Task, error) {
	data, err := read(source.Path)
	if err != nil {
		return Task{}, err
	}
	if len(data) != source.Bytes || hash(data) != source.SHA256 || brief.SourceID != source.ID {
		return Task{}, fmt.Errorf("document task %s has mismatched source bytes or brief identity", source.ID)
	}
	sealed, err := brief.Seal(data)
	if err != nil {
		return Task{}, fmt.Errorf("document %s: %w", source.ID, err)
	}
	if brief.SHA256 != "" && brief.SHA256 != sealed.SHA256 {
		return Task{}, fmt.Errorf("document %s: brief digest mismatch", source.ID)
	}
	return Task{ID: taskID(source.ID, DocumentScope), SourceID: source.ID, UnitID: DocumentScope, Scope: DocumentScope,
		GroupID: s.groups[source.ID], Partition: s.partition[source.ID], Cohort: s.options.Cohort, Repository: source.Repository,
		Ecosystem: s.options.Ecosystems[source.Repository], Path: source.Path, Role: source.Role, Words: words,
		Text: string(data), TextSHA256: source.SHA256, Brief: &sealed}, nil
}

func validateDocumentTask(task Task) error {
	if task.Scope == "" && task.Brief == nil {
		return nil
	}
	if task.Scope != DocumentScope || task.Brief == nil || task.UnitID != DocumentScope ||
		task.Brief.SourceID != task.SourceID || task.TextSHA256 != hash([]byte(task.Text)) {
		return fmt.Errorf("task %s has an invalid document scope or source binding", task.ID)
	}
	sealed, err := task.Brief.Seal([]byte(task.Text))
	if err != nil {
		return err
	}
	if sealed.SHA256 != task.Brief.SHA256 {
		return fmt.Errorf("task %s has an invalid brief digest", task.ID)
	}
	return nil
}
