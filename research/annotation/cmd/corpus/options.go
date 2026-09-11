package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
	"github.com/stokaro/unswell/research/annotation/training"
)

type options struct {
	root, round    string
	labels         string
	ruleConfig     string
	policy         string
	features       []string
	train          training.Options
	lexical        bool
	lexicalOptions training.LexicalOptions
}

func commandOptions(args []string) (options, error) {
	var result options
	flags := flag.NewFlagSet("corpus", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	registerFlags(args[0], flags, &result)
	if err := flags.Parse(args[1:]); err != nil {
		return options{}, err
	}
	if args[0] == "train" {
		if err := selectTrainingEstimator(flags, &result.train); err != nil {
			return options{}, err
		}
	}
	if err := validateLexicalFlags(flags, result); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 || (args[0] == "plan") != (result.root == "") {
		return options{}, fmt.Errorf("extract, verify, join, measure, and train require --root; plan does not accept it")
	}
	result.train.Features = result.features
	return result, result.validate(args[0])
}

func registerFlags(name string, flags *flag.FlagSet, result *options) {
	flags.StringVar(&result.root, "root", "", "Local directory containing exactly pinned source and notice files")
	flags.StringVar(&result.round, "round", "", "Explicit local annotation round for join or train")
	flags.StringVar(&result.labels, "labels", "",
		"Label source instead of a round: provenance labels the origin task, cohort the cohort task, from the corpus")
	flags.Func("feature", "Feature ID; repeat to select a set for join or train", func(value string) error {
		result.features = append(result.features, value)
		return nil
	})
	switch name {
	case "join":
		flags.StringVar(&result.ruleConfig, "rule-config", "", "Explicit inline policy file for original-block rule activations")
	case "measure":
		flags.StringVar(&result.policy, "policy", "", "Pinned policy file whose rule findings are measured over every candidate")
	case "train":
		flags.StringVar(&result.ruleConfig, "rule-config", "", "Explicit inline policy file for original-block rule activations")
		trainingFlags(flags, &result.train)
		lexicalFlags(flags, result)
	}
}

func (result options) validate(name string) error {
	if err := result.validateSelectors(name); err != nil {
		return err
	}
	if name == "measure" && result.policy == "" {
		return fmt.Errorf("measure requires --policy")
	}
	return result.validateLexical(name)
}

func (result options) validateSelectors(name string) error {
	if name == "join" || name == "train" {
		return result.validateLabelSource()
	}
	if result.round != "" || result.labels != "" || len(result.features) != 0 {
		return fmt.Errorf("only join and train accept --round, --labels, and --feature")
	}
	return nil
}

// validateLabelSource requires exactly one of a round and a label source.
func (result options) validateLabelSource() error {
	if (result.round == "") == (result.labels == "") || (len(result.features) == 0 && !result.lexical) {
		return fmt.Errorf("join and train require --round or --labels, and at least one --feature")
	}
	return validateLabels(result.labels)
}

// labelSources are the decision sets a command can build from the corpus
// itself: provenance labels the origin task, cohort labels the cohort task.
var labelSources = []string{"provenance", "cohort"}

func validateLabels(source string) error {
	if source != "" && !slices.Contains(labelSources, source) {
		return fmt.Errorf("--labels accepts provenance or cohort")
	}
	return nil
}

// labelDecisions builds the decision set a label source names from the
// artifact's declared cohorts and origins.
func labelDecisions(ctx context.Context, source string, artifact corpus.Artifact) (annotation.DecisionSet, error) {
	switch source {
	case "provenance":
		return corpus.OriginDecisions(ctx, artifact)
	case "cohort":
		return corpus.CohortDecisions(ctx, artifact)
	}
	return annotation.DecisionSet{}, fmt.Errorf("--labels accepts provenance or cohort")
}

func loadRound(ctx context.Context, path string) (*annotation.Round, error) {
	data, err := commandio.Await(ctx, func() ([]byte, error) { return readRound(path) })
	if err != nil {
		return nil, err
	}
	return annotation.Load(ctx, data)
}

func readRound(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > annotation.MaxBytes {
		return nil, fmt.Errorf("annotation round must be a regular file within its byte limit")
	}
	// #nosec G304 -- The operator explicitly selects this local round; reads require a regular file and a byte limit.
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(file, annotation.MaxBytes+1))
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	return data, closeErr
}
