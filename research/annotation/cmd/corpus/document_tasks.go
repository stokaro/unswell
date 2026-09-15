package main

import (
	"context"
	"fmt"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/generation"
)

func addSelectedTasks(ctx context.Context, sampler *generation.Sampler, artifact corpus.Artifact,
	read generation.Reader, briefsPath string,
) error {
	if briefsPath == "" {
		return sampler.Add(ctx, artifact, read)
	}
	var input struct {
		Version string                     `json:"version"`
		Briefs  []generation.DocumentBrief `json:"briefs"`
	}
	if _, err := loadJSON(ctx, briefsPath, corpus.MaxManifestBytes, &input); err != nil {
		return err
	}
	if input.Version != "unswell-document-briefs-v1" || len(input.Briefs) == 0 || len(input.Briefs) > generation.MaxTasks {
		return fmt.Errorf("document briefs require version 1 and 1 through %d entries", generation.MaxTasks)
	}
	briefs := make(map[string]generation.DocumentBrief, len(input.Briefs))
	for _, brief := range input.Briefs {
		if _, duplicate := briefs[brief.SourceID]; duplicate || brief.SourceID == "" {
			return fmt.Errorf("document briefs require unique nonempty source IDs")
		}
		briefs[brief.SourceID] = brief
	}
	return sampler.AddDocuments(ctx, artifact, read, briefs)
}
