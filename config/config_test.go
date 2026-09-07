package config_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/rule"
)

func FuzzPolicy(f *testing.F) {
	f.Add("version: 1\nextends: [builtin:strict]\n")
	f.Add("version: 1\nrules:\n  unknown.rule: {enabled: true}\n")
	f.Add("version: 1\nversion: 2\n")
	catalog := []rule.Descriptor{}
	for _, implementation := range builtin.Rules() {
		catalog = append(catalog, implementation.Descriptor())
	}
	f.Fuzz(func(t *testing.T, text string) {
		if len(text) > 8192 {
			t.Skip()
		}
		policy, err := config.Load([]byte(text), catalog)
		if err != nil {
			return
		}
		c := qt.New(t)
		c.Assert(policy.Hash, qt.HasLen, 64)
		again, err := config.Load([]byte(text), catalog)
		c.Assert(err, qt.IsNil)
		c.Assert(again, qt.DeepEquals, policy)
	})
}
