package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/bioshock/gospacy/v3/bundle"
	"github.com/bioshock/gospacy/v3/pipeline"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

type parsedSource struct {
	ID        string              `json:"id"`
	Format    document.Format     `json:"format"`
	Text      string              `json:"text"`
	Sentences []document.Sentence `json:"sentences"`
	Inputs    []parsedInput       `json:"inputs"`
}

func evaluate(ctx context.Context, opts options) (observation, error) {
	result := newObservation(opts.repeat)
	cases, hash, err := loadCases(opts.input)
	if err != nil {
		return result, err
	}
	result.InputHash = hash
	result.ModelFiles, result.ModelBytes, err = inventoryModel(opts.model)
	if err != nil {
		return result, err
	}
	start := time.Now()
	model, err := loadModel(opts.model)
	if err != nil {
		return result, err
	}
	result.LoadNS = elapsedNS(start)
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	start = time.Now()
	for repeat := range opts.repeat {
		for _, item := range cases {
			parsed, err := parseSource(ctx, model, item)
			if err != nil {
				return result, fmt.Errorf("case %s: %w", item.ID, err)
			}
			if repeat == 0 {
				result.Sources = append(result.Sources, parsed)
			}
		}
	}
	result.AnalysisNS = elapsedNS(start)
	runtime.ReadMemStats(&after)
	result.AllocBytes = after.TotalAlloc - before.TotalAlloc
	runtime.GC()
	runtime.ReadMemStats(&after)
	result.HeapRetained = after.HeapAlloc
	runtime.KeepAlive(model)
	return result, ctx.Err()
}

func parseSource(ctx context.Context, model *bundle.Bundle, item sourceCase) (parsedSource, error) {
	result := parsedSource{ID: item.ID, Format: item.Format, Text: item.Text}
	doc, err := extract.Parse(ctx, document.Source{Name: item.ID, Format: item.Format, Bytes: []byte(item.Text)},
		extract.Options{MaxBytes: 8192, MaxBlocks: 128})
	if err != nil {
		return result, err
	}
	for _, block := range doc.Blocks {
		parsed, err := parseBlock(ctx, model, block.MappedText)
		if err != nil {
			return result, err
		}
		for _, input := range parsed.inputs {
			input.FirstSentence += len(result.Sentences)
			input.EndSentence += len(result.Sentences)
			result.Inputs = append(result.Inputs, input)
		}
		sentences := parsed.sentences
		for i := range sentences {
			sentences[i].ID = len(result.Sentences)
			sentences[i].BlockID = block.ID
			result.Sentences = append(result.Sentences, sentences[i])
		}
	}
	if len(result.Sentences) == 0 {
		return result, fmt.Errorf("no parsed sentences")
	}
	return result, nil
}

func loadModel(path string) (*bundle.Bundle, error) {
	if err := checkModelMetadata(path); err != nil {
		return nil, err
	}
	model, err := bundle.FromDisk(path)
	if err != nil {
		return nil, err
	}
	// The high-level loader otherwise logs and skips a missing parser. Require
	// actual resources before inference, including lazy cfg/moves files.
	if _, err := pipeline.NewParser(model); err != nil {
		return nil, err
	}
	if _, err := pipeline.NewTagger(model); err != nil {
		return nil, err
	}
	return model, nil
}
