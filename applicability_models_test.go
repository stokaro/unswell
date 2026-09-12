package unswell_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
	"github.com/stokaro/unswell/probability"
)

type exclusionRejectingNLP struct{ nlp.Provider }

func (p exclusionRejectingNLP) Analyze(
	ctx context.Context, text document.MappedText, required []nlp.Capability,
) ([]document.Sentence, error) {
	if strings.Contains(text.Text, cyrillicParagraph) {
		return nil, fmt.Errorf("excluded prose reached the NLP provider")
	}
	return p.Provider.Analyze(ctx, text, required)
}

func TestNonLatinExclusionReachesPreparedAndModelChannels(t *testing.T) {
	c := qt.New(t)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	engine, err := unswell.New(unswell.Options{
		Config: []byte(packPolicy + "origin: {model: pack, accept_experimental: true}\n"),
		NLP:    exclusionRejectingNLP{Provider: provider},
		Model:  packFixture(c, "sentence", nil), OriginModel: originPack(c, nil),
		PreparedFeatures: []string{"prose-words"},
		PreparedKinds:    []string{"sentence", "paragraph"},
	})
	c.Assert(err, qt.IsNil)
	source := packSource()
	source.Bytes = []byte(cyrillicParagraph + "\n\n" + packProse)
	result, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Manifest.Complete, qt.IsTrue)
	c.Assert(result.Documents[0].Excluded, qt.HasLen, 1)
	c.Assert(paragraphIDs(result), qt.DeepEquals, []int{1, 2})
	statuses := []string{probability.StatusAvailable, probability.StatusInsufficientEvidence, probability.StatusAvailable}
	c.Assert(packStatuses(result.Assessments, "sentence"), qt.DeepEquals, statuses)
	c.Assert(originStatuses(result.Assessments, "sentence"), qt.DeepEquals, statuses)
	for _, unit := range result.PreparedFeatures.Sources[0].Units {
		c.Assert(unit.Binding.BlockID > 0, qt.IsTrue)
	}
}
