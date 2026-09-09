package nlp_test

import (
	"context"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
)

// unitTreeFixture supplies a structurally valid synthetic tree for copying tests.
// It makes no linguistic claim and is not a production dependency backend.
type unitTreeFixture struct {
	nlp.Provider
	omit bool
}

func (p unitTreeFixture) Identity() nlp.Identity {
	i := p.Provider.Identity()
	i.Capabilities = append(i.Capabilities, nlp.Dependencies)
	i.DependencyScheme = "synthetic-test-v1"
	return i
}

func (p unitTreeFixture) Analyze(ctx context.Context, mapped document.MappedText,
	capabilities []nlp.Capability,
) ([]document.Sentence, error) {
	capabilities = slices.DeleteFunc(slices.Clone(capabilities), func(c nlp.Capability) bool { return c == nlp.Dependencies })
	sentences, err := p.Provider.Analyze(ctx, mapped, capabilities)
	if err != nil || p.omit {
		return sentences, err
	}
	for i := range sentences {
		arcs := make([]document.DependencyArc, len(sentences[i].Tokens))
		for j := range arcs {
			arcs[j] = document.DependencyArc{Head: 0, Relation: "dep"}
		}
		arcs[0] = document.DependencyArc{Head: -1, Relation: "root"}
		sentences[i].Dependencies = &document.DependencyTree{Arcs: arcs}
	}
	return sentences, nil
}

func TestPreparedDependencyTreesAreRequiredAndDetached(t *testing.T) {
	c := qt.New(t)
	base, err := english.New()
	c.Assert(err, qt.IsNil)
	provider := unitTreeFixture{Provider: base}
	options := unitOptions()
	options.Capabilities = append(options.Capabilities, nlp.Dependencies)
	units, err := nlp.PrepareUnits(t.Context(), unitBlock("The cache may retry. The cache cannot wait.", "comment"), provider, options)
	c.Assert(err, qt.IsNil)
	block := units[1].Block()
	c.Assert(nlp.ValidateDependencies(t.Context(), block.MappedText, block.Sentences), qt.IsNil)
	block.Sentences[0].Dependencies.Arcs[0].Head = 99
	c.Assert(units[1].Block().Sentences[0].Dependencies.Arcs[0].Head, qt.Equals, -1)
	provider.omit = true
	units, err = nlp.PrepareUnits(t.Context(), unitBlock("The cache may retry.", "comment"), provider, options)
	c.Assert(err, qt.ErrorMatches, ".*requested dependency tree is missing.*")
	c.Assert(units, qt.IsNil)
}

func TestPreparedMappingIdentityDistinguishesSameBounds(t *testing.T) {
	c := qt.New(t)
	first := unitBlock("Red blue.", "comment")
	for i := 4; i < len(first.Map); i++ {
		first.Map[i].Start += 8
		first.Map[i].End += 8
	}
	first.Span.End += 8
	second := first
	second.Map = slices.Clone(first.Map)
	second.Map[3] = document.Span{Start: 17, End: 18}
	a := prepared(t, first, unitOptions())[0].Binding()
	b := prepared(t, second, unitOptions())[0].Binding()
	c.Assert(document.Bounds(a.Segments), qt.Equals, document.Bounds(b.Segments))
	c.Assert(a.TextSHA256, qt.Equals, b.TextSHA256)
	c.Assert(a.Segments, qt.Not(qt.DeepEquals), b.Segments)
	second.Context = []string{"Different grammar scope"}
	d := prepared(t, second, unitOptions())[0].Binding()
	c.Assert(d.ContextSHA256, qt.Equals, b.ContextSHA256)
	c.Assert(d.GrammarSHA256, qt.Not(qt.Equals), b.GrammarSHA256)
	options := unitOptions()
	options.Limits.MaxSegments = 1
	base, err := english.New()
	c.Assert(err, qt.IsNil)
	units, err := nlp.PrepareUnits(t.Context(), first, base, options)
	c.Assert(err, qt.ErrorMatches, ".*source segment limit.*")
	c.Assert(units, qt.IsNil)
}
