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
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

// fixSchema names the JSON report of the fix command.
const fixSchema = "unswell-fix-v1"

// maxFixPasses bounds the analyses of one document. One analysis lists a
// bounded number of marks, and a fix can leave out a mark that it overlaps.
const maxFixPasses = 8

type fixOptions struct {
	sources    checkOptions
	write      bool
	quiet      bool
	includeLow bool
}

// fixCommand removes the marks of the AI-sounding index that have a
// mechanical replacement. It calls no model and changes a file only with
// --write.
func fixCommand(environment Environment) *cobra.Command {
	var options fixOptions
	command := &cobra.Command{
		Use:   "fix [paths...]",
		Short: "Remove the marks that the AI-sounding index counts, without a model",
		Long: "Fix replaces the marks of the AI-sounding index that have a mechanical replacement in Markdown pages.\n" +
			"An em dash becomes a comma, a period between two full clauses, or a colon after the short label of a list item.\n" +
			"An emoji leaves a heading or the start of a list item.\n" +
			"An arrow becomes \"->\". A bold label at the start of a list item loses its asterisks.\n" +
			"A stock word stays: only you can say what the thing does.\n" +
			"A page at the level low is left as it is, because the check lists no marks there; --include-low changes it too.\n" +
			"No model is called and no text leaves this machine. Files change only with --write.\n" +
			"Read the change before you keep it: a comma or a period is not always what the sentence needs.",
		RunE: func(command *cobra.Command, args []string) error {
			return runFix(command.Context(), environment, options, args)
		},
	}
	flags := command.Flags()
	flags.StringVar(&options.sources.config, "config", "", "Use this exact YAML configuration file")
	configurationFlags(command, &options.sources, false)
	flags.StringVar(&options.sources.profile, "profile", "", "Select a builtin profile (cannot be combined with a config file)")
	flags.StringArrayVar(&options.sources.reports, "report", nil, "Report format:path with format text or json; repeat for several")
	flags.IntVar(&options.sources.jobs, "jobs", 2, "Documents handled at the same time (1–16)")
	flags.DurationVar(&options.sources.timeout, "timeout", 10*time.Minute, "Timeout for the whole run")
	flags.BoolVar(&options.write, "write", false, "Write the fixed pages into the selected files")
	flags.BoolVar(&options.includeLow, "include-low", false, "Also replace the marks of pages at the level low")
	flags.BoolVar(&options.quiet, "quiet", false, "Do not print progress to standard error")
	return command
}

// fixResult is the report of one fix run.
type fixResult struct {
	SchemaVersion string          `json:"schema_version"`
	Version       string          `json:"version"`
	Summary       fixSummary      `json:"summary"`
	Documents     []fixedDocument `json:"documents"`
}

type fixSummary struct {
	Documents int `json:"documents"`
	Failed    int `json:"failed"`
	Unindexed int `json:"without_index"`
	Low       int `json:"left_at_low"`
	Changed   int `json:"changed"`
	Removed   int `json:"marks_removed"`
	Left      int `json:"marks_left"`
	Written   int `json:"written"`
}

// fixedDocument holds what the fix did to one document: the index before and
// after, the marks it removed, and the marks it left for a person.
type fixedDocument struct {
	Path       string         `json:"path"`
	SourceHash string         `json:"source_hash,omitempty"`
	Indexed    bool           `json:"indexed"`
	LeftLow    bool           `json:"left_at_low"`
	Before     int            `json:"index_before"`
	After      int            `json:"index_after"`
	Removed    map[string]int `json:"marks_removed"`
	Left       map[string]int `json:"marks_left"`
	Written    bool           `json:"written"`
	Error      string         `json:"error,omitempty"`

	revised []byte
}

func (d fixedDocument) removed() int { return total(d.Removed) }

func (d fixedDocument) outcome() string {
	switch {
	case d.Error != "":
		return "failed: " + excerpt(d.Error, maxReasonText)
	case !d.Indexed:
		return "no index: not a Markdown page of 150 prose words or more"
	case d.LeftLow:
		return fmt.Sprintf("index %d, low: left as it is", d.Before)
	}
	return fmt.Sprintf("index %d -> %d, %s removed", d.Before, d.After, plural(d.removed(), "mark"))
}

func total(counts map[string]int) int {
	sum := 0
	for _, count := range counts {
		sum += count
	}
	return sum
}

func runFix(ctx context.Context, environment Environment, options fixOptions, args []string) error {
	if options.sources.jobs < 1 || options.sources.jobs > 16 {
		return fmt.Errorf("--jobs must be between 1 and 16")
	}
	if options.sources.timeout <= 0 {
		return fmt.Errorf("--timeout must be positive")
	}
	ctx, cancel := context.WithTimeout(ctx, options.sources.timeout)
	defer cancel()
	engine, sources, root, targets, err := fixSources(ctx, environment, options, args)
	if err != nil {
		return err
	}
	progress := progressWriter(environment, options.quiet)
	documents := eachSource(options.sources.jobs, sources, func(source document.Source) fixedDocument {
		result := fixOne(ctx, engine, source, options.includeLow)
		progress("fix", result.Path, result.outcome())
		return result
	})
	// An interrupted run reports and writes nothing.
	if err := ctx.Err(); errors.Is(err, context.Canceled) {
		return err
	}
	if options.write {
		writeFixes(root, documents)
	}
	result := newFixResult(documents)
	if err := writeFixReports(environment, targets, result, options.write); err != nil {
		return err
	}
	if result.Summary.Failed > 0 {
		return fmt.Errorf("the run is incomplete: %d of %d documents failed", result.Summary.Failed, result.Summary.Documents)
	}
	return nil
}

// fixSources builds the engine of the project and selects the documents and
// the report targets of a run.
func fixSources(
	ctx context.Context, environment Environment, options fixOptions, args []string,
) (*unswell.Engine, []document.Source, string, []reportTarget, error) {
	loaded, err := configuration(ctx, environment, options.sources)
	if err != nil {
		return nil, nil, "", nil, err
	}
	engine, err := unswell.New(unswell.Options{ConfigBundle: &loaded.Bundle})
	if err != nil {
		return nil, nil, "", nil, err
	}
	sources, paths, _, err := selectSources(ctx, environment, options.sources, engine, loaded.Root, args)
	if err != nil {
		return nil, nil, "", nil, err
	}
	if len(sources) == 0 {
		return nil, nil, "", nil, fmt.Errorf("the selection holds no documents")
	}
	targets, err := reviewTargets(environment.Dir, options.sources.reports, append(paths, loaded.Paths...))
	return engine, sources, loaded.Root, targets, err
}

// fixOne fixes one document and records what was done. A failure is recorded
// on the document.
func fixOne(ctx context.Context, engine *unswell.Engine, source document.Source, includeLow bool) fixedDocument {
	sum := sha256.Sum256(source.Bytes)
	result := fixedDocument{Path: source.Name, SourceHash: hex.EncodeToString(sum[:]), Removed: map[string]int{}, Left: map[string]int{}}
	revised, err := result.fix(ctx, engine, source, includeLow)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if result.removed() > 0 {
		result.revised = revised
	}
	return result
}

// fix analyzes the document, applies the fixes of its marks, and analyzes the
// result again until no fix is left. A page that starts at the level low is
// left as it is unless includeLow is set: the check lists no marks there. It
// returns the text after the last pass.
func (d *fixedDocument) fix(ctx context.Context, engine *unswell.Engine, source document.Source, includeLow bool) ([]byte, error) {
	current := source
	for pass := 0; ; pass++ {
		index, err := pageIndex(ctx, engine, current)
		if err != nil || index == nil {
			// A fix can take a short page below the words that the index needs.
			return current.Bytes, err
		}
		if pass == 0 {
			d.Indexed, d.Before = true, index.Index
			d.LeftLow = index.Level == unswell.AISoundingLow && !includeLow
		}
		d.After = index.Index
		if d.LeftLow {
			return current.Bytes, nil
		}
		revised, applied := unswell.RemoveAIMarks(current.Bytes, index.Marks)
		if applied == 0 || pass == maxFixPasses {
			d.Left = markCounts(index.Marks, false)
			return current.Bytes, nil
		}
		for name, count := range markCounts(index.Marks, true) {
			d.Removed[name] += count
		}
		current.Bytes = revised
	}
}

// pageIndex returns the index of one document, or nil when it has none.
func pageIndex(ctx context.Context, engine *unswell.Engine, source document.Source) (*unswell.AISounding, error) {
	analysis, err := engine.Analyze(ctx, source)
	if err != nil {
		return nil, err
	}
	if len(analysis.Documents) == 0 || analysis.Documents[0].AISounding == nil || analysis.Documents[0].AISounding.Status != "available" {
		return nil, nil
	}
	return analysis.Documents[0].AISounding, nil
}

// markCounts counts the marks with a fix, or the marks without one, by signal.
func markCounts(marks []unswell.AIMark, fixable bool) map[string]int {
	counts := map[string]int{}
	for _, mark := range marks {
		if (mark.Fix != nil) == fixable {
			counts[mark.Signal]++
		}
	}
	return counts
}

// writeFixes replaces each file that the fix changed. A file that changed
// since it was read is left alone and its document fails.
func writeFixes(root string, documents []fixedDocument) {
	for index := range documents {
		entry := &documents[index]
		if entry.Error != "" || entry.revised == nil {
			continue
		}
		if err := rewriteFile(filepath.Join(root, filepath.FromSlash(entry.Path)), entry.SourceHash, entry.revised); err != nil {
			entry.Error = err.Error()
			continue
		}
		entry.Written = true
	}
}

func newFixResult(documents []fixedDocument) fixResult {
	result := fixResult{SchemaVersion: fixSchema, Version: unswell.Version, Documents: documents}
	result.Summary.Documents = len(documents)
	for _, entry := range documents {
		switch {
		case entry.Error != "":
			result.Summary.Failed++
		case !entry.Indexed:
			result.Summary.Unindexed++
		case entry.LeftLow:
			result.Summary.Low++
		case entry.revised != nil:
			result.Summary.Changed++
		}
		if entry.Written {
			result.Summary.Written++
		}
		result.Summary.Removed += entry.removed()
		result.Summary.Left += total(entry.Left)
	}
	return result
}

func writeFixReports(environment Environment, targets []reportTarget, result fixResult, written bool) error {
	var failures []error
	for _, target := range targets {
		var buffer bytes.Buffer
		err := writeFixText(&buffer, result, written)
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

// fixSignalNames gives the singular and the plural that a reader sees for
// each signal, in the order of the report.
var fixSignalNames = [][3]string{
	{"em-dash", "em dash", "em dashes"},
	{"emoji-in-heading", "emoji in a heading", "emoji in headings"},
	{"emoji-at-list-item", "emoji at the start of a list item", "emoji at the start of list items"},
	{"arrow", "arrow", "arrows"},
	{"stock-word", "stock word", "stock words"},
	{"bold-label-in-list", "bold label", "bold labels"},
}

// markText lists counted marks by signal, such as "3 em dashes, 1 arrow".
func markText(counts map[string]int) string {
	var parts []string
	for _, name := range fixSignalNames {
		switch count := counts[name[0]]; {
		case count == 1:
			parts = append(parts, "1 "+name[1])
		case count > 1:
			parts = append(parts, fmt.Sprintf("%d %s", count, name[2]))
		}
	}
	return strings.Join(parts, ", ")
}

// fixLine reports one document in the text report, or nothing for a document
// that the fix had nothing to say about.
func fixLine(entry fixedDocument) string {
	switch {
	case entry.Error != "":
		return fmt.Sprintf("FAILED %s: %s\n", entry.Path, excerpt(entry.Error, maxReasonText))
	case entry.revised != nil:
		line := fmt.Sprintf("%s: AI-sounding index %d -> %d; removed %s.", entry.Path, entry.Before, entry.After, markText(entry.Removed))
		if left := markText(entry.Left); left != "" {
			line += fmt.Sprintf(" Left for you: %s.", left)
		}
		return line + "\n"
	case entry.Indexed && total(entry.Left) > 0:
		return fmt.Sprintf("%s: AI-sounding index %d; nothing to replace. Left for you: %s.\n", entry.Path, entry.Before, markText(entry.Left))
	}
	return ""
}

func writeFixText(writer io.Writer, result fixResult, written bool) error {
	var text strings.Builder
	for _, entry := range result.Documents {
		text.WriteString(fixLine(entry))
	}
	summary := result.Summary
	fmt.Fprintf(&text, "%s removed in %s of %s.\n", plural(summary.Removed, "mark"), plural(summary.Changed, "page"),
		plural(summary.Documents, "document"))
	if summary.Unindexed > 0 {
		fmt.Fprintf(&text, "Without an index (not a Markdown page of 150 prose words or more): %s.\n", plural(summary.Unindexed, "document"))
	}
	if summary.Low > 0 {
		fmt.Fprintf(&text, "Left as they are at the level low: %s. Use --include-low to replace their marks too.\n",
			plural(summary.Low, "page"))
	}
	switch {
	case written:
		fmt.Fprintf(&text, "Written to %s. Review the change before you keep it.\n", plural(summary.Written, "file"))
	case summary.Removed > 0:
		text.WriteString("Nothing was written. Use --write to apply these replacements.\n")
	}
	text.WriteString("A replacement is mechanical. Read each one: a comma or a period is not always what the sentence needs, " +
		"and a stock word needs your own words.\n")
	_, err := io.WriteString(writer, text.String())
	return err
}
