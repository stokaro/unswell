package cli

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/review"
)

// modelRunOptions are shared by the commands that call a model: which
// documents, how many at once, and how long and how often to ask.
type modelRunOptions struct {
	sources          checkOptions
	insecureEndpoint bool
	attempts         int
	retryDelay       time.Duration
	requestTimeout   time.Duration
	maxDocuments     int
	maxDocumentBytes int
	quiet            bool
}

type reviewOptions struct {
	modelRunOptions
	reviewer        modelOptions
	confirmer       modelOptions
	showUnconfirmed bool
}

// reviewCommand reads documents with a model the user names. It is the only
// command that sends text off the machine or starts another program to judge
// text, and it does so only with explicit flags.
func reviewCommand(environment Environment) *cobra.Command {
	var options reviewOptions
	command := &cobra.Command{
		Use:   "review [paths...]",
		Short: "Review documents in context with a model you name; findings are advisory",
		Long: "Review sends each selected document to a model you name and binds the model's flags to source bytes.\n" +
			"A flag that does not quote its item is dropped. With a second model, only flags that model confirms are shown.\n" +
			"The document text leaves this machine for the endpoint or program you name. Nothing else does.\n" +
			"Findings are a model's judgment. They are advisory: the command exits 0 when the review completed.",
		RunE: func(command *cobra.Command, args []string) error {
			return runReview(command.Context(), environment, options, args)
		},
	}
	reviewFlags(command, &options)
	return command
}

func reviewFlags(command *cobra.Command, options *reviewOptions) {
	modelFlags(command, &options.reviewer, "", "reviewing")
	modelFlags(command, &options.confirmer, "confirm-", "confirming")
	modelRunFlags(command, &options.modelRunOptions)
	command.Flags().BoolVar(&options.showUnconfirmed, "show-unconfirmed", false,
		"Also list in the text report the flags the confirming model did not accept")
}

func modelRunFlags(command *cobra.Command, options *modelRunOptions) {
	flags := command.Flags()
	flags.BoolVar(&options.insecureEndpoint, "insecure-endpoint", false, "Permit an http endpoint outside this machine")
	flags.StringVar(&options.sources.config, "config", "", "Use this exact YAML configuration file")
	configurationFlags(command, &options.sources, false)
	flags.StringVar(&options.sources.profile, "profile", "", "Select a builtin profile (cannot be combined with a config file)")
	flags.BoolVar(&options.sources.stdin, "stdin", false, "Read source bytes from stdin")
	flags.StringVar(&options.sources.filename, "filename", "", "Logical filename for stdin")
	flags.StringVar(&options.sources.format, "format", "", "Input syntax, as for check")
	flags.StringArrayVar(&options.sources.reports, "report", nil, "Report format:path with format text or json; repeat for several")
	flags.IntVar(&options.sources.jobs, "jobs", 2, "Documents sent at the same time (1–16)")
	flags.DurationVar(&options.sources.timeout, "timeout", 2*time.Hour, "Timeout for the whole run")
	flags.DurationVar(&options.requestTimeout, "request-timeout", 5*time.Minute, "Timeout for one model call")
	flags.IntVar(&options.attempts, "attempts", 3, "Calls per question before a document fails (1–5)")
	flags.DurationVar(&options.retryDelay, "retry-delay", 2*time.Second, "Delay before a repeated call; it grows with each attempt")
	flags.IntVar(&options.maxDocuments, "max-documents", 50, "Refuse to send more documents than this in one run")
	flags.IntVar(&options.maxDocumentBytes, "max-document-bytes", 200_000, "A larger document fails instead of being sent")
	flags.BoolVar(&options.quiet, "quiet", false, "Do not print progress to standard error")
}

func modelFlags(command *cobra.Command, options *modelOptions, prefix, role string) {
	flags := command.Flags()
	flags.StringVar(&options.endpoint, prefix+"endpoint", "", "Chat completions URL of the "+role+" model (OpenAI-compatible)")
	flags.StringVar(&options.model, prefix+"model", "", "Name of the "+role+" model at the endpoint")
	flags.StringVar(&options.apiKeyEnv, prefix+"api-key-env", "", "Environment variable that holds the endpoint's API key")
	flags.IntVar(&options.maxTokens, prefix+"max-tokens", 0, "Send max_tokens with each request; zero leaves it to the endpoint")
	flags.StringVar(&options.command, prefix+"command", "",
		"Local program as the "+role+" model: it reads the prompt on stdin and prints the answer")
	flags.StringArrayVar(&options.commandArgs, prefix+"command-arg", nil, "Argument for the "+role+" program; repeat in order")
}

func validateModelRun(options modelRunOptions) error {
	switch {
	case options.sources.jobs < 1 || options.sources.jobs > 16:
		return fmt.Errorf("--jobs must be between 1 and 16")
	case options.attempts < 1 || options.attempts > 5:
		return fmt.Errorf("--attempts must be between 1 and 5")
	case options.sources.timeout <= 0 || options.requestTimeout <= 0 || options.retryDelay < 0:
		return fmt.Errorf("timeouts must be positive and the retry delay cannot be negative")
	case options.maxDocuments < 1 || options.maxDocumentBytes < 1:
		return fmt.Errorf("--max-documents and --max-document-bytes must be positive")
	}
	return nil
}

// confirmerOptions lets a second model at the same endpoint be named by its
// model alone.
func confirmerOptions(options reviewOptions) modelOptions {
	confirmer := options.confirmer
	if confirmer.command != "" || confirmer.model == "" || confirmer.endpoint != "" || options.reviewer.endpoint == "" {
		return confirmer
	}
	confirmer.endpoint = options.reviewer.endpoint
	if confirmer.apiKeyEnv == "" {
		confirmer.apiKeyEnv = options.reviewer.apiKeyEnv
	}
	return confirmer
}

type reviewRun struct {
	options   reviewOptions
	prepare   *review.Reviewer
	reviewer  modelClient
	confirmer modelClient
}

func newReviewRun(environment Environment, options reviewOptions) (*reviewRun, error) {
	if err := validateModelRun(options.modelRunOptions); err != nil {
		return nil, err
	}
	reviewer, err := newModelClient(environment, options.reviewer, "", options.insecureEndpoint)
	if err != nil {
		return nil, err
	}
	run := &reviewRun{options: options, reviewer: reviewer}
	if confirmer := confirmerOptions(options); !confirmer.empty() {
		if run.confirmer, err = newModelClient(environment, confirmer, "confirm-", options.insecureEndpoint); err != nil {
			return nil, err
		}
	}
	return run, nil
}

func runReview(ctx context.Context, environment Environment, options reviewOptions, args []string) error {
	run, err := newReviewRun(environment, options)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, options.sources.timeout)
	defer cancel()
	sources, inputs, _, prepare, err := modelSources(ctx, environment, options.modelRunOptions, args)
	if err != nil {
		return err
	}
	run.prepare = prepare
	targets, err := reviewTargets(environment.Dir, options.sources.reports, inputs)
	if err != nil {
		return err
	}
	progress := progressWriter(environment, options.quiet)
	documents := eachSource(options.sources.jobs, sources, func(source document.Source) reviewedDocument {
		result := run.one(ctx, source)
		progress("review", result.Path, result.outcome())
		return result
	})
	// An interrupted run reports nothing. A run that reached its time limit
	// still reports the documents it finished; the rest count as failed.
	if err := ctx.Err(); errors.Is(err, context.Canceled) {
		return err
	}
	result := newReviewResult(run, documents)
	if err := writeReviewReports(environment, targets, result, options.showUnconfirmed); err != nil {
		return err
	}
	return result.Summary.incomplete()
}

// modelSources loads the policy, discovers the sources as check does, and
// prepares a reviewer under the same policy. It also returns the paths that a
// report must not overwrite and the root that source names are relative to.
func modelSources(
	ctx context.Context, environment Environment, options modelRunOptions, args []string,
) ([]document.Source, []string, string, *review.Reviewer, error) {
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
	if len(sources) > options.maxDocuments {
		return nil, nil, "", nil, fmt.Errorf("the selection holds %d documents and --max-documents is %d; "+
			"name fewer paths or raise the limit", len(sources), options.maxDocuments)
	}
	prepare, err := review.New(review.Options{ConfigBundle: &loaded.Bundle})
	return sources, append(paths, loaded.Paths...), loaded.Root, prepare, err
}

// eachSource calls work for every source with a fixed number of workers and
// returns the results in the order of the sources.
func eachSource[T any](jobs int, sources []document.Source, work func(document.Source) T) []T {
	results := make([]T, len(sources))
	indexes := make(chan int)
	var workers sync.WaitGroup
	for range min(jobs, len(sources)) {
		workers.Go(func() {
			for index := range indexes {
				results[index] = work(sources[index])
			}
		})
	}
	for index := range sources {
		indexes <- index
	}
	close(indexes)
	workers.Wait()
	return results
}

// one reviews one document. A failure is recorded on the document, so the
// other documents still report and the run ends as incomplete.
func (r *reviewRun) one(ctx context.Context, source document.Source) reviewedDocument {
	result := reviewedDocument{Path: source.Name, Findings: []reviewedFinding{},
		Unconfirmed: []reviewedFinding{}, Unbound: []review.Rejection{}}
	if len(source.Bytes) > r.options.maxDocumentBytes {
		result.Error = fmt.Sprintf("the document has %d bytes and --max-document-bytes is %d", len(source.Bytes), r.options.maxDocumentBytes)
		return result
	}
	prepared, err := r.prepare.Prepare(ctx, source)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.SourceHash, result.Items = prepared.SourceHash, len(prepared.Items)
	if len(prepared.Items) == 0 {
		return result
	}
	if err := r.judge(ctx, prepared, &result); err != nil {
		result.Error = err.Error()
	}
	return result
}

// judge asks the reviewer, binds its flags, and asks the confirming model
// about the bound flags. When the confirming model fails, the bound flags are
// kept as unconfirmed and the document counts as failed.
func (r *reviewRun) judge(ctx context.Context, prepared *review.Document, result *reviewedDocument) error {
	var flags []review.Flag
	err := askModel(ctx, r.reviewer, prepared.Prompt(), r.options.modelRunOptions, func(answer string) (parseErr error) {
		flags, parseErr = review.ParseFlags(answer)
		return parseErr
	})
	if err != nil {
		return fmt.Errorf("reviewer: %w", err)
	}
	result.Flags = len(flags)
	bound, rejected, err := prepared.Bind(flags)
	if err != nil {
		return fmt.Errorf("reviewer: %w", err)
	}
	result.Unbound = rejected
	if r.confirmer == nil || len(bound) == 0 {
		for _, finding := range bound {
			result.Findings = append(result.Findings, reviewedFinding{Finding: finding})
		}
		return nil
	}
	var verdicts []review.Verdict
	err = askModel(ctx, r.confirmer, prepared.ConfirmationPrompt(bound), r.options.modelRunOptions, func(answer string) (parseErr error) {
		verdicts, parseErr = review.ParseVerdicts(answer, len(bound))
		return parseErr
	})
	if err != nil {
		for _, finding := range bound {
			result.Unconfirmed = append(result.Unconfirmed, reviewedFinding{Finding: finding})
		}
		return fmt.Errorf("confirming model: %w", err)
	}
	for index, finding := range bound {
		entry := reviewedFinding{Finding: finding, Confirmation: &verdicts[index]}
		if verdicts[index].Justified == review.Justified {
			result.Findings = append(result.Findings, entry)
		} else {
			result.Unconfirmed = append(result.Unconfirmed, entry)
		}
	}
	return nil
}

// askModel calls a model until its answer parses, at most the configured
// number of times. A permanent failure or a canceled run ends the attempts at
// once.
func askModel(ctx context.Context, client modelClient, prompt string, options modelRunOptions, parse func(string) error) error {
	var last error
	for attempt := range options.attempts {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * options.retryDelay):
			}
		}
		call, cancel := context.WithTimeout(ctx, options.requestTimeout)
		answer, err := client.ask(call, prompt)
		cancel()
		if err == nil {
			if err = parse(answer); err == nil {
				return nil
			}
		}
		last = err
		if _, permanent := errors.AsType[permanentError](err); permanent || ctx.Err() != nil {
			break
		}
	}
	return last
}

// progressWriter prints one line per finished document to standard error.
func progressWriter(environment Environment, quiet bool) func(command, path, outcome string) {
	if quiet {
		return func(string, string, string) {}
	}
	var lock sync.Mutex
	return func(command, path, outcome string) {
		lock.Lock()
		defer lock.Unlock()
		// The stream records a failed write itself and the run then exits 2.
		_, _ = fmt.Fprintln(environment.Err, command+":", path+":", outcome)
	}
}
