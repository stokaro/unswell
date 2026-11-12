package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/review"
)

// trimSchema names the JSON report of the trim command.
const trimSchema = "unswell-trim-v1"

// maxSentenceText bounds a sentence in the text report. The JSON report keeps
// it whole.
const maxSentenceText = 600

type trimOptions struct {
	modelRunOptions
	model modelOptions
	write bool
}

// trimCommand asks a model for sentences with words removed and keeps the
// proposals that pass a deterministic check. It changes a file only with
// --write.
func trimCommand(environment Environment) *cobra.Command {
	var options trimOptions
	command := &cobra.Command{
		Use:   "trim [paths...]",
		Short: "Propose sentences with empty words removed, using a model you name",
		Long: "Trim sends each selected document to a model you name and asks for sentences with words removed.\n" +
			"A proposal is kept only if it deletes words and nothing else: no word added, replaced, or moved, at least half\n" +
			"of the sentence left, and no negation, number, modal word, code, or link removed.\n" +
			"The document text leaves this machine for the endpoint or program you name. Nothing else does.\n" +
			"Files change only with --write. Read each trim before you keep it: a deleted word can carry meaning.",
		RunE: func(command *cobra.Command, args []string) error {
			return runTrim(command.Context(), environment, options, args)
		},
	}
	modelFlags(command, &options.model, "", "trimming")
	modelRunFlags(command, &options.modelRunOptions)
	command.Flags().BoolVar(&options.write, "write", false, "Write the accepted trims into the selected files")
	return command
}

// trimResult is the report of one trim run.
type trimResult struct {
	SchemaVersion string            `json:"schema_version"`
	Protocol      string            `json:"protocol"`
	Version       string            `json:"version"`
	Model         modelIdentity     `json:"model"`
	Summary       trimSummary       `json:"summary"`
	Documents     []trimmedDocument `json:"documents"`
}

type trimSummary struct {
	Documents    int `json:"documents"`
	Failed       int `json:"failed"`
	WithoutProse int `json:"without_prose"`
	Trims        int `json:"trims"`
	Rejected     int `json:"rejected"`
	Written      int `json:"written"`
}

// trimmedDocument holds the accepted trims of one document and the proposals
// that did not pass the check.
type trimmedDocument struct {
	Path       string                 `json:"path"`
	SourceHash string                 `json:"source_hash,omitempty"`
	Sentences  int                    `json:"sentences"`
	Proposed   int                    `json:"proposed"`
	Trims      []review.TrimFinding   `json:"trims"`
	Rejected   []review.TrimRejection `json:"rejected"`
	Written    bool                   `json:"written"`
	Error      string                 `json:"error,omitempty"`

	revised []byte
}

func (d trimmedDocument) outcome() string {
	switch {
	case d.Error != "":
		return "failed: " + excerpt(d.Error, maxReasonText)
	case d.Sentences == 0:
		return "no sentence to trim"
	}
	return fmt.Sprintf("%s, %d proposals rejected", plural(len(d.Trims), "trim"), len(d.Rejected))
}

// trimModel validates the options and builds the model client.
func trimModel(environment Environment, options trimOptions) (modelClient, error) {
	if err := validateModelRun(options.modelRunOptions); err != nil {
		return nil, err
	}
	if options.write && options.sources.stdin {
		return nil, fmt.Errorf("--write changes files and cannot be combined with --stdin")
	}
	return newModelClient(environment, options.model, "", options.insecureEndpoint)
}

func runTrim(ctx context.Context, environment Environment, options trimOptions, args []string) error {
	client, err := trimModel(environment, options)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, options.sources.timeout)
	defer cancel()
	sources, inputs, root, prepare, err := modelSources(ctx, environment, options.modelRunOptions, args)
	if err != nil {
		return err
	}
	targets, err := reviewTargets(environment.Dir, options.sources.reports, inputs)
	if err != nil {
		return err
	}
	progress := progressWriter(environment, options.quiet)
	documents := eachSource(options.sources.jobs, sources, func(source document.Source) trimmedDocument {
		result := trimOne(ctx, prepare, client, options.modelRunOptions, source)
		progress("trim", result.Path, result.outcome())
		return result
	})
	// An interrupted run reports and writes nothing.
	if err := ctx.Err(); errors.Is(err, context.Canceled) {
		return err
	}
	if options.write {
		writeTrims(root, documents)
	}
	result := newTrimResult(client, documents)
	if err := writeTrimReports(environment, targets, result, options.write); err != nil {
		return err
	}
	return result.Summary.incomplete()
}

// incomplete reports why a run did not trim what was selected, or nil.
func (s trimSummary) incomplete() error {
	switch {
	case s.Failed > 0:
		return fmt.Errorf("the run is incomplete: %d of %d documents failed", s.Failed, s.Documents)
	case s.WithoutProse == s.Documents:
		return fmt.Errorf("no selected document has a sentence to trim")
	}
	return nil
}

// trimOne asks the model about one document and keeps the proposals that pass
// the check. A failure is recorded on the document.
func trimOne(
	ctx context.Context, prepare *review.Reviewer, client modelClient, options modelRunOptions, source document.Source,
) trimmedDocument {
	result := trimmedDocument{Path: source.Name, Trims: []review.TrimFinding{}, Rejected: []review.TrimRejection{}}
	if len(source.Bytes) > options.maxDocumentBytes {
		result.Error = fmt.Sprintf("the document has %d bytes and --max-document-bytes is %d", len(source.Bytes), options.maxDocumentBytes)
		return result
	}
	prepared, err := prepare.Prepare(ctx, source)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.SourceHash, result.Sentences = prepared.SourceHash, len(prepared.Trimmable())
	if result.Sentences == 0 {
		return result
	}
	var proposals []review.Trim
	err = askModel(ctx, client, prepared.TrimPrompt(), options, func(answer string) (parseErr error) {
		proposals, parseErr = review.ParseTrims(answer)
		return parseErr
	})
	if err == nil {
		result.Proposed = len(proposals)
		result.Trims, result.Rejected, err = prepared.BindTrims(proposals)
	}
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if len(result.Trims) > 0 {
		result.revised = prepared.ApplyTrims(result.Trims)
	}
	return result
}

// writeTrims replaces each file that has accepted trims. A file that changed
// since it was read is left alone and its document fails.
func writeTrims(root string, documents []trimmedDocument) {
	for index := range documents {
		entry := &documents[index]
		if entry.Error != "" || len(entry.Trims) == 0 {
			continue
		}
		if err := rewriteFile(filepath.Join(root, filepath.FromSlash(entry.Path)), entry.SourceHash, entry.revised); err != nil {
			entry.Error = err.Error()
			continue
		}
		entry.Written = true
	}
}

// rewriteFile replaces a file's content and keeps its mode. It refuses when
// the file no longer has the content the trims were made for.
func rewriteFile(path, expected string, data []byte) (err error) {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	// #nosec G304 -- The path is a source the user selected, read again before it is replaced.
	current, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if sum := sha256.Sum256(current); hex.EncodeToString(sum[:]) != expected {
		return fmt.Errorf("the file changed during the run and was not written")
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".unswell-trim-*")
	if err != nil {
		return err
	}
	name := file.Name()
	defer func() {
		if removeErr := os.Remove(name); removeErr != nil && !os.IsNotExist(removeErr) {
			err = errors.Join(err, removeErr)
		}
	}()
	_, writeErr := file.Write(data)
	modeErr := file.Chmod(info.Mode().Perm())
	closeErr := file.Close()
	if err := errors.Join(writeErr, modeErr, closeErr); err != nil {
		return err
	}
	return os.Rename(name, path)
}

func newTrimResult(client modelClient, documents []trimmedDocument) trimResult {
	result := trimResult{SchemaVersion: trimSchema, Protocol: review.TrimProtocol, Version: unswell.Version,
		Model: client.identity(), Documents: documents}
	result.Summary.Documents = len(documents)
	for _, entry := range documents {
		if entry.Error != "" {
			result.Summary.Failed++
		} else if entry.Sentences == 0 {
			result.Summary.WithoutProse++
		}
		if entry.Written {
			result.Summary.Written++
		}
		result.Summary.Trims += len(entry.Trims)
		result.Summary.Rejected += len(entry.Rejected)
	}
	return result
}

func writeTrimReports(environment Environment, targets []reportTarget, result trimResult, written bool) error {
	var failures []error
	for _, target := range targets {
		var buffer bytes.Buffer
		err := writeTrimText(&buffer, result, written)
		if target.format == "json" {
			buffer.Reset()
			encoder := json.NewEncoder(&buffer)
			encoder.SetIndent("", "  ")
			err = encoder.Encode(result)
		}
		if err == nil && target.path == "-" {
			_, err = environment.Out.Write(buffer.Bytes())
		} else if err == nil {
			err = atomicWrite(target.path, writeBytes(buffer.Bytes()))
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("write %s report: %w", target.format, err))
		}
	}
	return errors.Join(failures...)
}

func writeTrimText(writer io.Writer, result trimResult, written bool) error {
	var text strings.Builder
	for _, entry := range result.Documents {
		for _, trim := range entry.Trims {
			fmt.Fprintf(&text, "%s:%d:%d: trim\n    - %s\n    + %s\n", trim.Path, trim.Start.Line, trim.Start.Column,
				excerpt(trim.Original, maxSentenceText), excerpt(trim.Trimmed, maxSentenceText))
		}
	}
	summary := result.Summary
	fmt.Fprintf(&text, "%s in %s.", plural(summary.Trims, "trim"), plural(summary.Documents-summary.Failed, "document"))
	if summary.Rejected > 0 {
		fmt.Fprintf(&text, " %s did not pass the check.", plural(summary.Rejected, "proposal"))
	}
	text.WriteString("\n")
	if summary.WithoutProse > 0 {
		fmt.Fprintf(&text, "Without a sentence to trim: %s.\n", plural(summary.WithoutProse, "document"))
	}
	for _, entry := range result.Documents {
		if entry.Error != "" {
			fmt.Fprintf(&text, "FAILED %s: %s\n", entry.Path, excerpt(entry.Error, maxReasonText))
		}
	}
	fmt.Fprintf(&text, "Model: %s.\n", result.Model)
	switch {
	case written:
		fmt.Fprintf(&text, "Written to %s. Review the change before you keep it.\n", plural(summary.Written, "file"))
	case summary.Trims > 0:
		text.WriteString("Nothing was written. Use --write to apply these trims.\n")
	}
	text.WriteString("A trim only deletes words. A deleted word can still carry meaning that the check cannot see.\n")
	_, err := io.WriteString(writer, text.String())
	return err
}
