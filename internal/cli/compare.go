package cli

import (
	"bytes"
	"context"
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
	"github.com/stokaro/unswell/review"
)

// compareSchema names the JSON report of the compare command.
const compareSchema = "unswell-compare-v1"

// maxListedLosses bounds each list of lost numbers, code, and links in the
// text report. The JSON report keeps them whole.
const maxListedLosses = 12

type compareOptions struct {
	modelRunOptions
	judge modelOptions
}

// compareCommand checks a revision of a page with a model the user names. The
// model compares the two texts without knowing which is the revision, and the
// command accepts or rejects the revision. It does not use the page index: an
// index falls after almost any edit.
func compareCommand(environment Environment, code *int) *cobra.Command {
	var options compareOptions
	command := &cobra.Command{
		Use:   "compare <page> <revision>",
		Short: "Check a revision of a page blind, with a model you name; exit 1 when it is rejected",
		Long: "Compare sends a page and a revision of it to a model you name. The model reads the opening of both, about\n" +
			"450 words, and is asked which reads more like text written by a language model, in both orders and\n" +
			"without a hint of which is the revision.\n" +
			"It is also asked for the statements of fact of the page, then for each of them in the page itself, and\n" +
			"then for each that it found there in the revision.\n" +
			"A revision is accepted when the judge names the page in both orders, at least 98% of the statements are\n" +
			"kept, no number, code, or link is lost, and no em dash is added. The page index plays no part.\n" +
			"Both texts leave this machine for the endpoint or program you name. Nothing else does.\n" +
			"The judge is a model. Name one other than the model that made the revision.",
		Args: cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			return runCompare(command.Context(), environment, options, args, code)
		},
	}
	modelFlags(command, &options.judge, "", "judging")
	compareFlags(command, &options.modelRunOptions)
	return command
}

func compareFlags(command *cobra.Command, options *modelRunOptions) {
	flags := command.Flags()
	flags.BoolVar(&options.insecureEndpoint, "insecure-endpoint", false, "Permit an http endpoint outside this machine")
	flags.StringVar(&options.sources.config, "config", "", "Use this exact YAML configuration file")
	configurationFlags(command, &options.sources, false)
	flags.StringVar(&options.sources.profile, "profile", "", "Select a builtin profile (cannot be combined with a config file)")
	flags.StringVar(&options.sources.format, "format", "", "Input syntax of both texts, as for check")
	flags.StringArrayVar(&options.sources.reports, "report", nil, "Report format:path with format text or json; repeat for several")
	flags.DurationVar(&options.sources.timeout, "timeout", 30*time.Minute, "Timeout for the whole run")
	flags.DurationVar(&options.requestTimeout, "request-timeout", 5*time.Minute, "Timeout for one model call")
	flags.IntVar(&options.attempts, "attempts", 3, "Calls per question before the comparison fails (1–5)")
	flags.DurationVar(&options.retryDelay, "retry-delay", 2*time.Second, "Delay before a repeated call; it grows with each attempt")
	flags.IntVar(&options.maxDocumentBytes, "max-document-bytes", 200_000, "A larger text fails instead of being sent")
	flags.BoolVar(&options.quiet, "quiet", false, "Do not print progress to standard error")
}

// compareResult is the report of one comparison. The parts after Revision
// are present as far as the run got.
type compareResult struct {
	SchemaVersion string                  `json:"schema_version"`
	Protocol      string                  `json:"protocol"`
	Version       string                  `json:"version"`
	Judge         modelIdentity           `json:"judge"`
	Page          comparedSource          `json:"page"`
	Revision      comparedSource          `json:"revision"`
	Sound         *comparedSound          `json:"sound,omitempty"`
	Facts         *comparedFacts          `json:"facts,omitempty"`
	Lost          *review.Lost            `json:"lost,omitempty"`
	EmDashes      *comparedDashes         `json:"em_dashes,omitempty"`
	Verdict       *review.RevisionVerdict `json:"verdict,omitempty"`
	Error         string                  `json:"error,omitempty"`
}

type comparedSource struct {
	Path       string `json:"path"`
	SourceHash string `json:"source_hash,omitempty"`
}

// comparedSound holds the verdict of the blind comparison and the judge's
// answer to each order.
type comparedSound struct {
	Verdict       string             `json:"verdict"`
	PageFirst     review.SoundAnswer `json:"page_first"`
	RevisionFirst review.SoundAnswer `json:"revision_first"`
}

// comparedFacts counts the statements. Listed are those that the model named
// for the page. Dropped are those of them that the same model did not find in
// the page itself; they are not looked for in the revision. Statements are the
// others, with their outcome in the revision.
type comparedFacts struct {
	Listed     int           `json:"listed"`
	Dropped    []string      `json:"dropped"`
	Kept       int           `json:"kept"`
	Statements []review.Fact `json:"statements"`
}

type comparedDashes struct {
	Page     int `json:"page"`
	Revision int `json:"revision"`
}

func (r compareResult) outcome() string {
	switch {
	case r.Error != "":
		return "failed: " + excerpt(r.Error, maxReasonText)
	case r.Verdict.Accepted:
		return "accepted"
	}
	return "rejected"
}

// compareRun holds what one comparison needs before the first question.
type compareRun struct {
	client  modelClient
	prepare *review.Reviewer
	targets []reportTarget
	sources []document.Source
}

func newCompareRun(ctx context.Context, environment Environment, options compareOptions, args []string) (compareRun, error) {
	var run compareRun
	if err := validateModelRun(options.modelRunOptions); err != nil {
		return run, err
	}
	var err error
	if run.client, err = newModelClient(environment, options.judge, "", options.insecureEndpoint); err != nil {
		return run, err
	}
	loaded, err := configuration(ctx, environment, options.sources)
	if err != nil {
		return run, err
	}
	if run.prepare, err = review.New(review.Options{ConfigBundle: &loaded.Bundle}); err != nil {
		return run, err
	}
	paths := absoluteArguments(environment.Dir, args)
	if run.targets, err = reviewTargets(environment.Dir, options.sources.reports, append(paths, loaded.Paths...)); err != nil {
		return run, err
	}
	run.sources, err = comparedSources(loaded.Root, paths, options.modelRunOptions)
	return run, err
}

func runCompare(ctx context.Context, environment Environment, options compareOptions, args []string, code *int) error {
	// One pair is compared, so the limits of a run over many documents do not apply.
	options.sources.jobs, options.maxDocuments = 1, 2
	ctx, cancel := context.WithTimeout(ctx, options.sources.timeout)
	defer cancel()
	run, err := newCompareRun(ctx, environment, options, args)
	if err != nil {
		return err
	}
	result := compareSources(ctx, run.prepare, run.client, options.modelRunOptions, run.sources[0], run.sources[1])
	// An interrupted run reports nothing.
	if err := ctx.Err(); errors.Is(err, context.Canceled) {
		return err
	}
	progressWriter(environment, options.quiet)("compare", result.Revision.Path, result.outcome())
	if err := writeCompareReports(environment, run.targets, result); err != nil {
		return err
	}
	if result.Error != "" {
		return fmt.Errorf("the comparison is incomplete: %s", excerpt(result.Error, maxReasonText))
	}
	if !result.Verdict.Accepted {
		*code = 1
	}
	return nil
}

// comparedSources reads the page and its revision. A text over the byte
// limit fails here, before anything is sent.
func comparedSources(root string, paths []string, options modelRunOptions) ([]document.Source, error) {
	sources := make([]document.Source, 0, len(paths))
	for _, path := range paths {
		data, err := readLimited(path, options.maxDocumentBytes)
		if err != nil {
			return nil, err
		}
		format, err := inputFormat(path, options.sources.format, data)
		if err != nil {
			return nil, err
		}
		name := path
		if relative, relErr := filepath.Rel(root, path); relErr == nil && !escapesRoot(relative) {
			name = relative
		}
		sources = append(sources, document.Source{Name: filepath.ToSlash(name), Format: format, Bytes: data})
	}
	return sources, nil
}

// compareSources asks the judge about one pair. A failure is recorded on the
// result, which then has no verdict.
func compareSources(
	ctx context.Context, prepare *review.Reviewer, client modelClient, options modelRunOptions, page, revised document.Source,
) compareResult {
	result := compareResult{SchemaVersion: compareSchema, Protocol: review.RevisionProtocol, Version: unswell.Version,
		Judge: client.identity(), Page: comparedSource{Path: page.Name}, Revision: comparedSource{Path: revised.Name}}
	revision, err := preparedRevision(ctx, prepare, page, revised, &result)
	if err == nil {
		err = judgeRevision(ctx, client, options, revision, &result)
	}
	if err != nil {
		result.Error = err.Error()
	}
	return result
}

func preparedRevision(
	ctx context.Context, prepare *review.Reviewer, page, revised document.Source, result *compareResult,
) (*review.Revision, error) {
	before, err := prepare.Prepare(ctx, page)
	if err != nil {
		return nil, err
	}
	after, err := prepare.Prepare(ctx, revised)
	if err != nil {
		return nil, err
	}
	result.Page.SourceHash, result.Revision.SourceHash = before.SourceHash, after.SourceHash
	if before.SourceHash == after.SourceHash {
		return nil, fmt.Errorf("the revision is the page itself")
	}
	return review.NewRevision(before, after)
}

// judgeRevision runs the four questions and the checks that need no model.
func judgeRevision(
	ctx context.Context, client modelClient, options modelRunOptions, revision *review.Revision, result *compareResult,
) error {
	sound, err := askSound(ctx, client, options, revision)
	if err != nil {
		return fmt.Errorf("blind comparison: %w", err)
	}
	result.Sound = &sound
	checked, err := askFacts(ctx, client, options, revision)
	if err != nil {
		return fmt.Errorf("statements of fact: %w", err)
	}
	result.Facts = &checked
	facts := checked.Statements
	lost := revision.Lost()
	result.Lost = &lost
	page, revised := revision.EmDashes()
	result.EmDashes = &comparedDashes{Page: page, Revision: revised}
	verdict := revision.Judge(sound.Verdict, facts)
	result.Verdict = &verdict
	return nil
}

func askSound(ctx context.Context, client modelClient, options modelRunOptions, revision *review.Revision) (comparedSound, error) {
	var answers [2]review.SoundAnswer
	for index, revisionFirst := range []bool{false, true} {
		err := askModel(ctx, client, revision.SoundPrompt(revisionFirst), options, func(answer string) (parseErr error) {
			answers[index], parseErr = review.ParseSound(answer, revisionFirst)
			return parseErr
		})
		if err != nil {
			return comparedSound{}, err
		}
	}
	return comparedSound{Verdict: review.SoundVerdict(answers[0], answers[1]), PageFirst: answers[0], RevisionFirst: answers[1]}, nil
}

// askFacts lists the statements of the page, keeps those that the model finds
// in the page itself, and looks for them in the revision.
func askFacts(ctx context.Context, client modelClient, options modelRunOptions, revision *review.Revision) (comparedFacts, error) {
	var statements []string
	err := askModel(ctx, client, revision.FactsPrompt(), options, func(answer string) (parseErr error) {
		statements, parseErr = review.ParseFacts(answer)
		return parseErr
	})
	if err != nil {
		return comparedFacts{}, err
	}
	lookup := func(prompt string, count int) (verdicts []review.FactVerdict, err error) {
		err = askModel(ctx, client, prompt, options, func(answer string) (parseErr error) {
			verdicts, parseErr = review.ParseKept(answer, count)
			return parseErr
		})
		return verdicts, err
	}
	verdicts, err := lookup(revision.ControlPrompt(statements), len(statements))
	if err != nil {
		return comparedFacts{}, err
	}
	confirmed, dropped, err := revision.Confirmed(statements, verdicts)
	if err != nil {
		return comparedFacts{}, err
	}
	result := comparedFacts{Listed: len(statements), Dropped: dropped, Statements: []review.Fact{}}
	if len(confirmed) == 0 {
		return result, nil
	}
	if verdicts, err = lookup(revision.KeptPrompt(confirmed), len(confirmed)); err != nil {
		return comparedFacts{}, err
	}
	result.Statements, err = revision.BindFacts(confirmed, verdicts)
	result.Kept = review.KeptFacts(result.Statements)
	return result, err
}

func writeCompareReports(environment Environment, targets []reportTarget, result compareResult) error {
	var failures []error
	for _, target := range targets {
		var buffer bytes.Buffer
		err := writeCompareText(&buffer, result)
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

func writeCompareText(writer io.Writer, result compareResult) error {
	var text strings.Builder
	fmt.Fprintf(&text, "%s against its revision %s\n", result.Page.Path, result.Revision.Path)
	if result.Sound != nil {
		text.WriteString(soundText(*result.Sound))
	}
	if result.Facts != nil {
		text.WriteString(factsText(*result.Facts))
	}
	if result.Lost != nil {
		text.WriteString(lostText(*result.Lost))
	}
	if result.EmDashes != nil {
		fmt.Fprintf(&text, "Em dashes: %d in the page, %d in the revision.\n", result.EmDashes.Page, result.EmDashes.Revision)
	}
	switch {
	case result.Error != "":
		fmt.Fprintf(&text, "FAILED: %s\n", excerpt(result.Error, maxReasonText))
	case result.Verdict.Accepted:
		text.WriteString("Accepted.\n")
	default:
		fmt.Fprintf(&text, "Rejected: %s.\n", strings.Join(result.Verdict.Reasons, "; "))
	}
	fmt.Fprintf(&text, "Judge: %s.\n", result.Judge)
	text.WriteString("The judge did not see which text is the revision. " +
		"It is a model: name one other than the model that made the revision.\n")
	_, err := io.WriteString(writer, text.String())
	return err
}

func soundText(sound comparedSound) string {
	named := map[string]string{review.NamedPage: "the page", review.NamedRevision: "the revision", review.NamedNeither: "neither"}
	var text strings.Builder
	switch sound.Verdict {
	case review.SoundBetter:
		text.WriteString("Sound: the revision reads less like a model's text. The judge named the page in both orders.\n")
	case review.SoundWorse:
		text.WriteString("Sound: the revision reads more like a model's text. The judge named the revision in both orders.\n")
	default:
		fmt.Fprintf(&text, "Sound: no difference. With the page first the judge named %s, and with the revision first %s.\n",
			named[sound.PageFirst.Named], named[sound.RevisionFirst.Named])
	}
	for _, answer := range []review.SoundAnswer{sound.PageFirst, sound.RevisionFirst} {
		if answer.Reason != "" {
			fmt.Fprintf(&text, "    %s\n", excerpt(answer.Reason, maxReasonText))
		}
	}
	return text.String()
}

func factsText(facts comparedFacts) string {
	var text strings.Builder
	fmt.Fprintf(&text, "Facts: %d of %s kept.", facts.Kept, plural(len(facts.Statements), "statement"))
	if len(facts.Dropped) > 0 {
		fmt.Fprintf(&text, " Not counted: %s that the judge did not find in the page itself.", plural(len(facts.Dropped), "statement"))
	}
	text.WriteString("\n")
	for _, fact := range facts.Statements {
		if fact.Outcome != review.FactKept {
			fmt.Fprintf(&text, "    %s: %s\n", fact.Outcome, excerpt(fact.Text, maxReasonText))
		}
	}
	return text.String()
}

func lostText(lost review.Lost) string {
	if !lost.Any() {
		return "Numbers, code, and links: nothing lost.\n"
	}
	var text strings.Builder
	for _, group := range []struct {
		name   string
		values []string
	}{{"numbers", lost.Numbers}, {"code", lost.Code}, {"links", lost.Links}} {
		if len(group.values) == 0 {
			continue
		}
		shown := group.values
		rest := ""
		if len(shown) > maxListedLosses {
			rest = fmt.Sprintf(", and %d more", len(shown)-maxListedLosses)
			shown = shown[:maxListedLosses]
		}
		fmt.Fprintf(&text, "Lost %s: %s%s.\n", group.name, strings.Join(shown, ", "), rest)
	}
	return text.String()
}
