package nlp_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
)

func FuzzPrepareUnits(f *testing.F) {
	provider, err := english.New()
	if err != nil {
		f.Fatal(err)
	}
	f.Add([]byte("The café may retry.\r\nThe service cannot wait."), uint8(0))
	f.Add([]byte("Keep\x00 unchanged."), uint8(1))
	f.Add([]byte(""), uint8(2))
	f.Fuzz(func(t *testing.T, data []byte, mode uint8) {
		if len(data) > 4096 {
			return
		}
		c := qt.New(t)
		block := unitBlock(string(data), "comment")
		if len(block.Map) > 1 {
			switch mode % 3 {
			case 1:
				block.Map = block.Map[1:]
			case 2:
				block.Map[0], block.Map[len(block.Map)-1] = block.Map[len(block.Map)-1], block.Map[0]
			}
		}
		options := unitOptions()
		options.Capabilities = []nlp.Capability{nlp.Tokens, nlp.Sentences}
		units, err := nlp.PrepareUnits(t.Context(), block, provider, options)
		if err != nil {
			c.Assert(units, qt.IsNil)
			return
		}
		for _, unit := range units {
			target := unit.Block()
			c.Assert(strings.ContainsRune(target.Text, 0), qt.IsFalse)
			c.Assert(target.Map, qt.HasLen, len(target.Text))
			c.Assert(unit.Binding().Segments, qt.DeepEquals, target.Spans(0, len(target.Text)))
		}
	})
}
