package main

import (
	"flag"
	"fmt"
	"strings"
)

func lexicalFlags(flags *flag.FlagSet, options *options) {
	flags.BoolVar(&options.lexical, "lexical", false, "Fit a vocabulary from selected training targets")
	v := &options.lexicalOptions
	flags.IntVar(&v.Counts.WordMin, "lexical-word-min", 1, "Minimum word n-gram order; 0/0 disables words")
	flags.IntVar(&v.Counts.WordMax, "lexical-word-max", 2, "Maximum word n-gram order (at most 3)")
	flags.IntVar(&v.Counts.CharMin, "lexical-char-min", 3, "Minimum character n-gram order; 0/0 disables characters")
	flags.IntVar(&v.Counts.CharMax, "lexical-char-max", 5, "Maximum character n-gram order (at most 6)")
	flags.IntVar(&v.MaxFeatures, "lexical-max-features", 128, "Maximum retained vocabulary columns (at most 128)")
	flags.IntVar(&v.MinTargets, "lexical-min-targets", 1, "Minimum number of fitted training targets containing a term")
}

func validateLexicalFlags(flags *flag.FlagSet, options options) error {
	var err error
	flags.Visit(func(f *flag.Flag) {
		if strings.HasPrefix(f.Name, "lexical-") && !options.lexical {
			err = fmt.Errorf("lexical options require --lexical")
		}
	})
	return err
}

func (result options) validateLexical(name string) error {
	if name == "train" && result.train.Kind == "" {
		return fmt.Errorf("train requires --kind")
	}
	if result.lexical && (name != "train" || len(result.features) != 0 || result.ruleConfig != "") {
		return fmt.Errorf("lexical training rejects explicit features and rule configuration")
	}
	return nil
}
