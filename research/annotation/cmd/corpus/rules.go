package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
	"github.com/stokaro/unswell/research/annotation/training"
)

const maxRuleConfigBytes = 1 << 20

func ruleOperation(ctx context.Context, name string, options options, artifact corpus.Artifact,
	round *annotation.Round, files map[string][]byte,
) (any, error) {
	configuration, err := commandio.Await(ctx, func() ([]byte, error) { return readRuleConfig(options.ruleConfig) })
	if err != nil {
		return nil, err
	}
	if name == "train" {
		return training.RunRules(ctx, artifact, round, files, options.train, configuration)
	}
	return corpus.JoinRules(ctx, artifact, round, files, options.features, configuration)
}

func readRuleConfig(path string) ([]byte, error) {
	return readLocalArtifact(path, maxRuleConfigBytes, "rule config")
}

func readLocalArtifact(path string, maximum int, description string) ([]byte, error) {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	data, readErr := readLocalRootFile(root, filepath.Base(path), maximum, description)
	closeErr := root.Close()
	if readErr != nil {
		return nil, readErr
	}
	return data, closeErr
}

func readLocalRootFile(root *os.Root, name string, maximum int, description string) ([]byte, error) {
	info, err := root.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > int64(maximum) {
		return nil, fmt.Errorf("%s must be a regular file within its byte limit", description)
	}
	file, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(file, int64(maximum)+1))
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	if len(data) > maximum {
		return nil, fmt.Errorf("%s exceeds its byte limit", description)
	}
	return data, closeErr
}
