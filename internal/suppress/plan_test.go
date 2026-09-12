package suppress_test

import (
	"context"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/suppress"
)

func options() suppress.Options {
	return suppress.Options{RequireReason: true, RejectUnused: true, MaxCandidates: 1000}
}

func fixture(directives ...document.Directive) document.Document {
	return document.Document{Source: []byte(strings.Repeat(" ", 256)), Directives: directives,
		Blocks: []document.Block{{ID: 0, Kind: "paragraph", Words: 10, Span: document.Span{Start: 100, End: 180},
			Sentences: []document.Sentence{
				{ID: 0, BlockID: 0, Words: 4, Span: document.Span{Start: 100, End: 130}},
				{ID: 1, BlockID: 0, Words: 6, Span: document.Span{Start: 140, End: 180}},
			}}}}
}

func raw(text string, start int) document.Directive {
	return document.Directive{Text: text, Span: document.Span{Start: start, End: start + 10}}
}

func TestRegionInsideBlockSelectsOnlyCompleteSentences(t *testing.T) {
	c := qt.New(t)
	doc := fixture(raw("unswell-disable rule.one -- Required contract wording.", 0), raw("unswell-enable rule.one", 132))
	plan, err := suppress.Build(t.Context(), doc, map[string]bool{"rule.one": true}, options())
	c.Assert(err, qt.IsNil)
	c.Assert(plan.Entries[0].Targets, qt.HasLen, 1)
	c.Assert(plan.Entries[0].Targets[0].Scope, qt.Equals, "sentence")
	c.Assert(plan.Entries[0].Targets[0].ID, qt.Equals, 0)
	c.Assert(plan.Entries[0].Targets[0].Span, qt.Equals, document.Span{Start: 100, End: 130})
	matched, err := plan.Match(t.Context(), "rule.one", "crosses", []document.Span{{Start: 110, End: 115}, {Start: 150, End: 155}})
	c.Assert(err, qt.IsNil)
	c.Assert(matched, qt.HasLen, 0)
	c.Assert(plan.ValidateUse(), qt.IsNotNil)
	matched, err = plan.Match(t.Context(), "rule.one", "covered", []document.Span{{Start: 110, End: 115}})
	c.Assert(err, qt.IsNil)
	c.Assert(matched, qt.DeepEquals, []int{0})
	c.Assert(plan.ValidateUse(), qt.IsNil)
}

func TestEveryRulePermissionMustBeUsed(t *testing.T) {
	c := qt.New(t)
	doc := fixture(raw("unswell-disable-next-block rule.one,rule.two -- Required contract wording.", 0))
	plan, err := suppress.Build(t.Context(), doc, map[string]bool{"rule.one": true, "rule.two": true}, options())
	c.Assert(err, qt.IsNil)
	_, err = plan.Match(t.Context(), "rule.one", "finding", []document.Span{{Start: 110, End: 115}})
	c.Assert(err, qt.IsNil)
	c.Assert(plan.ValidateUse(), qt.ErrorMatches, ".*unused suppression for rule rule.two")
	c.Assert(plan.Entries[0].UsedRules, qt.DeepEquals, []string{"rule.one"})
}

// A permission for a rule that abstained on the document had nothing to cover;
// excusing that rule keeps the audit record while the other rules stay checked.
func TestExcusedRulesAreNotReportedUnused(t *testing.T) {
	c := qt.New(t)
	doc := fixture(raw("unswell-disable-next-block rule.one,rule.two -- Required contract wording.", 0))
	plan, err := suppress.Build(t.Context(), doc, map[string]bool{"rule.one": true, "rule.two": true}, options())
	c.Assert(err, qt.IsNil)
	plan.Excuse([]string{"rule.two"})
	c.Assert(plan.ValidateUse(), qt.ErrorMatches, ".*unused suppression for rule rule.one")
	_, err = plan.Match(t.Context(), "rule.one", "finding", []document.Span{{Start: 110, End: 115}})
	c.Assert(err, qt.IsNil)
	c.Assert(plan.ValidateUse(), qt.IsNil)
	c.Assert(plan.Entries[0].UsedRules, qt.DeepEquals, []string{"rule.one"})
}

func TestDirectivePairingTargetsAndLimits(t *testing.T) {
	for _, directives := range [][]document.Directive{
		{raw("unswell-disable-next-block rule.one -- Required contract wording.", 200)},
		{raw("unswell-disable-next-block rule.one -- Required contract wording.", 0),
			raw("unswell-disable-next-block rule.one -- Required contract wording.", 20)},
		{raw("unswell-disable rule.one -- Required contract wording.", 0), raw("unswell-enable rule.two", 200)},
		{raw("unswell-disable rule.one -- Required contract wording.", 0), raw("unswell-disable rule.two -- Required contract wording.", 20),
			raw("unswell-enable rule.one", 200), raw("unswell-enable rule.two", 220)},
		{raw("unswell-disable-next-block rule.one,rule.one -- Required contract wording.", 0)},
		{raw("unswell-enable rule.one -- Unneeded closing reason.", 200)},
		{raw("unswell-disable-next-block rule.one -- Required\ncontract wording.", 0)},
	} {
		c := qt.New(t)
		_, err := suppress.Build(t.Context(), fixture(directives...), map[string]bool{"rule.one": true, "rule.two": true}, options())
		c.Assert(err, qt.IsNotNil, qt.Commentf("%v", directives))
	}
	c := qt.New(t)
	limited := options()
	limited.MaxCandidates = 1
	_, err := suppress.Build(t.Context(), fixture(raw("unswell-disable-next-block rule.one -- Required contract wording.", 0)),
		map[string]bool{"rule.one": true}, limited)
	c.Assert(err, qt.ErrorMatches, ".*max_candidates")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = suppress.Build(ctx, fixture(), nil, options())
	c.Assert(err, qt.ErrorIs, context.Canceled)
}

func FuzzDirectives(f *testing.F) {
	f.Add("unswell-disable-next-block rule.one -- Required contract wording.", "")
	f.Add("unswell-disable rule.one -- Required contract wording.", "unswell-enable rule.one")
	f.Add("unswell-disable-next-sentence rule.one --", "unswell-enabel rule.one")
	f.Fuzz(func(t *testing.T, first, second string) {
		if len(first)+len(second) > 16384 {
			t.Skip()
		}
		directives := []document.Directive{raw(first, 0)}
		if second != "" {
			directives = append(directives, raw(second, 200))
		}
		plan, err := suppress.Build(t.Context(), fixture(directives...), map[string]bool{"rule.one": true}, options())
		if err != nil {
			return
		}
		c := qt.New(t)
		_, err = plan.Match(t.Context(), "rule.one", "finding", []document.Span{{Start: 105, End: 120}, {Start: 150, End: 160}})
		c.Assert(err, qt.IsNil)
	})
}
