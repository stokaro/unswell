package builtin

import (
	"slices"
	"strings"

	"github.com/jdkato/prose/v3/tokenize"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func phraseRules() []rule.Rule {
	definitions := []struct {
		id, summary        string
		positions, phrases []string
		forbid             bool
	}{
		{"policy.banned-phrases", "Remove a phrase forbidden by this editorial policy.", []string{"any"}, []string{}, true},
		{
			"scaffold.chat-preamble",
			"Start with the subject instead of a conversational acknowledgment.",
			// A pasted chat answer usually sits under a title, so the opener is
			// also sought in the first sentence after the leading headings.
			[]string{"document-start", "prose-start"},
			// "Sure," and "Sure!" come from what experts removed from model answers;
			// see ADR 0055.
			[]string{"Certainly!", "Great question!", "Absolutely!", "Sure,", "Sure!"},
			true,
		},
		{
			"scaffold.ai-self-reference",
			"Remove the assistant's self-reference.",
			[]string{"sentence-start"},
			[]string{"As an AI language model", "As an AI assistant"},
			true,
		},
		{
			"scaffold.follow-up-offer",
			"Remove the conversational follow-up offer from this document.",
			[]string{"document-end"},
			[]string{"Let me know if you'd like", "Let me know if you need"},
			false,
		},
		{
			"scaffold.dive-in",
			"Start with the subject instead of announcing the explanation.",
			[]string{"sentence-start"},
			[]string{"Let's dive in", "Let's dive into", "Let's delve into"},
			false,
		},
		{
			"filler.announced-importance",
			"State the fact without announcing its importance.",
			[]string{"sentence-start"},
			[]string{"It is important to note that", "It is worth noting that", "It is crucial to note that"},
			false,
		},
		{
			"filler.modern-world-opening",
			"Replace the generic opening with the document's subject.",
			[]string{"sentence-start"},
			[]string{"In today's rapidly evolving digital landscape", "In today's fast-paced digital world"},
			false,
		},
		{
			"filler.wordy-phrase",
			"Consider a shorter phrase that preserves the condition or meaning.",
			[]string{"any"},
			append([]string{"due to the fact that", "in order to", "at this point in time"}, qualifiedWordyPhrases()...),
			false,
		},
	}
	result := make([]rule.Rule, 0, len(definitions))
	for _, def := range definitions {
		d := descriptor(def.id, def.summary, "scaffolding", "sentence", 15)
		d.TermExemptions = true
		d.BlockObservations = true
		d.Defaults.Parameters = rule.Parameters{Phrases: def.phrases, Positions: def.positions}
		d.Parameters = []string{"phrases", "positions"}
		if def.id == "scaffold.chat-preamble" {
			d.Limitations += " In source code the first comment or string of a file is read as the start of a document." +
				" A file that holds chat replies as product text can open with a listed phrase;" +
				" exclude such a file in the extraction policy."
		}
		if def.forbid {
			d.Defaults.Gate = "forbid"
			d.Defaults.Severity = "error"
		}
		if len(def.phrases) > 0 {
			d.Examples = []rule.Example{
				{Text: def.phrases[0] + " the client opens connections.", Match: true},
				{Text: "The client opens connections.", Match: false},
			}
		} else {
			policy := "version: 1\nrules:\n  policy.banned-phrases:\n    parameters:\n      phrases: [magic solution]\n"
			d.Examples = []rule.Example{
				{Text: "This is a magic solution.", Match: true, Config: policy},
				{Text: "This is a pragmatic solution.", Match: false, Config: policy},
			}
		}
		result = append(result, check{descriptor: d, evaluate: phraseEvaluate})
	}
	return result
}

func phraseTokens(text string) []string {
	result := make([]string, 0)
	for _, token := range tokenize.New().Tokenize(document.Normalize(text)) {
		result = append(result, document.Normalize(token.Text))
	}
	return result
}

func matches(tokens []document.Token, pattern []string) bool {
	for i, token := range tokens {
		if token.Protected || token.Normal != pattern[i] {
			return false
		}
	}
	return true
}

// sentencePlace locates one sentence among the sentences of its document.
type sentencePlace struct {
	// index counts sentences from the start of the document.
	index int
	// prose is the index of the first sentence outside a heading, or -1 when
	// a document has headings only.
	prose int
	total int
}

func phrasePosition(positions []string, start int, place sentencePlace) bool {
	for _, position := range positions {
		if atPosition(position, start, place) {
			return true
		}
	}
	return false
}

func atPosition(position string, start int, place sentencePlace) bool {
	switch position {
	case "any":
		return true
	case "sentence-start":
		return start == 0
	case "document-start":
		return place.index == 0 && start == 0
	case "prose-start":
		// The first sentence after the leading headings. Without a heading it
		// is the start of the document.
		return place.index == place.prose && start == 0
	case "document-end":
		return place.index == place.total-1 && start == 0
	}
	return false
}

// firstProseSentence returns the index of the first sentence outside a
// heading, or -1 when a document has headings only.
func firstProseSentence(blocks []document.Block) int {
	index := 0
	for _, block := range blocks {
		if block.Kind != "heading" && len(block.Sentences) > 0 {
			return index
		}
		index += len(block.Sentences)
	}
	return -1
}

func allSentences(doc *document.Document) []document.Sentence {
	result := make([]document.Sentence, 0)
	for _, block := range doc.Blocks {
		result = append(result, block.Sentences...)
	}
	return result
}

func tokenOccurrence(sentence document.Sentence, start, end int) rule.Occurrence {
	spans := make([]document.Span, 0)
	for _, token := range sentence.Tokens[start:end] {
		spans = append(spans, token.Spans...)
	}
	return rule.Occurrence{BlockID: sentence.BlockID, SentenceID: sentence.ID, Spans: spans}
}

func sentenceOccurrence(sentence document.Sentence) rule.Occurrence {
	return rule.Occurrence{BlockID: sentence.BlockID, SentenceID: sentence.ID, Spans: slices.Clone(sentence.Spans)}
}

func normalizedWords(sentence document.Sentence) []string {
	result := make([]string, 0, sentence.Words)
	for _, token := range sentence.Tokens {
		if token.Word {
			result = append(result, token.Normal)
		}
	}
	return result
}

func sentenceKey(sentence document.Sentence) string {
	parts := make([]string, 0, len(sentence.Tokens))
	for _, token := range sentence.Tokens {
		if token.Protected {
			return ""
		}
		parts = append(parts, token.Normal)
	}
	return strings.Join(parts, " ")
}
