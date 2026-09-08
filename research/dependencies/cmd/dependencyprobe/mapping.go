package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/bioshock/gospacy/v3/bundle"
	gdoc "github.com/bioshock/gospacy/v3/doc"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
)

type parsedInput struct {
	Text          string `json:"text"`
	Offset        int    `json:"offset"`
	FirstSentence int    `json:"first_sentence"`
	EndSentence   int    `json:"end_sentence"`
}

type parsedBlock struct {
	sentences []document.Sentence
	inputs    []parsedInput
}

func parseBlock(ctx context.Context, model *bundle.Bundle, mapped document.MappedText) (parsedBlock, error) {
	var result parsedBlock
	offset := 0
	for prose := range strings.SplitSeq(mapped.Text, "\x00") {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		if strings.TrimSpace(prose) != "" {
			doc, err := model.PipeWith(prose, bundle.PipeOptions{SkipNER: true, SkipLemmatizer: true})
			if err != nil {
				return result, err
			}
			parsed, err := mapSentences(ctx, mapped, offset, doc)
			if err != nil {
				return result, err
			}
			result.inputs = append(result.inputs, parsedInput{Text: prose, Offset: offset,
				FirstSentence: len(result.sentences), EndSentence: len(result.sentences) + len(parsed)})
			result.sentences = append(result.sentences, parsed...)
		}
		offset += len(prose) + 1
	}
	return result, nlp.ValidateDependencies(ctx, mapped, result.sentences)
}

func mapSentences(ctx context.Context, mapped document.MappedText, offset int, doc *gdoc.Doc) ([]document.Sentence, error) {
	if doc.NumTokens() == 0 || doc.Tokens[0].SentStart != 1 {
		return nil, fmt.Errorf("parser did not return a sentence start")
	}
	text := doc.Source()
	runeBytes := make([]int, 0, len(text)+1)
	for i := range text {
		runeBytes = append(runeBytes, i)
	}
	runeBytes = append(runeBytes, len(text))
	var result []document.Sentence
	first := 0
	for i := 1; i <= doc.NumTokens(); i++ {
		if i < doc.NumTokens() {
			switch doc.Tokens[i].SentStart {
			case 0:
				continue
			case 1:
			default:
				return nil, fmt.Errorf("parser left an unknown sentence boundary")
			}
		}
		sentence, err := mapSentence(mapped, offset, doc, runeBytes, first, i)
		if err != nil {
			return nil, err
		}
		if err := nlp.ValidateDependencyTree(ctx, mapped, sentence); err != nil {
			return nil, err
		}
		result = append(result, sentence)
		first = i
	}
	return result, nil
}

func mapSentence(mapped document.MappedText, offset int, doc *gdoc.Doc, runeBytes []int, first, end int) (document.Sentence, error) {
	sentence := document.Sentence{Dependencies: &document.DependencyTree{}}
	for i := first; i < end; i++ {
		token := doc.Tokens[i]
		if token.Idx < 0 || token.Idx >= len(runeBytes)-1 {
			return sentence, fmt.Errorf("invalid rune offset at token %d", i)
		}
		start := offset + runeBytes[token.Idx]
		stop := start + len(token.Text)
		if start < 0 || stop > len(mapped.Text) {
			return sentence, fmt.Errorf("token %d lies outside the mapped block", i)
		}
		tag, _ := doc.Vocab.StringStore().Lookup(token.Tag)
		relation, _ := doc.Vocab.StringStore().Lookup(token.Dep)
		head, err := localHead(token, relation, i, first, end)
		if err != nil {
			return sentence, err
		}
		word := document.IsWord(token.Text)
		sentence.Tokens = append(sentence.Tokens, document.Token{Text: token.Text, Normal: document.Normalize(token.Text),
			Tag: tag, Start: start, End: stop, Spans: mapped.Spans(start, stop), Word: word})
		sentence.Dependencies.Arcs = append(sentence.Dependencies.Arcs, document.DependencyArc{Head: head, Relation: relation})
		if word {
			sentence.Words++
		}
	}
	start, stop := sentence.Tokens[0].Start, sentence.Tokens[len(sentence.Tokens)-1].End
	sentence.Text = mapped.Text[start:stop]
	sentence.Spans = mapped.Spans(start, stop)
	sentence.Span = document.Bounds(sentence.Spans)
	return sentence, nil
}

func localHead(token gdoc.Token, relation string, index, first, end int) (int, error) {
	if token.Head == index && relation == "ROOT" {
		return -1, nil
	}
	if token.Head < first || token.Head >= end {
		return 0, fmt.Errorf("token %d has a head outside its sentence", index)
	}
	return token.Head - first, nil
}
