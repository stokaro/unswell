package unswell

import (
	"slices"

	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/rule"
)

func cloneDescriptors(source []rule.Descriptor) []rule.Descriptor {
	result := slices.Clone(source)
	for i := range result {
		d := &result[i]
		d.Contexts = slices.Clone(d.Contexts)
		d.Requires = slices.Clone(d.Requires)
		d.Parameters = slices.Clone(d.Parameters)
		d.Examples = slices.Clone(d.Examples)
		d.Defaults.Parameters = cloneParameters(d.Defaults.Parameters)
		if d.Origin != nil {
			origin := *d.Origin
			d.Origin = &origin
		}
	}
	return result
}

func cloneOverrides(source []config.OverrideIdentity) []config.OverrideIdentity {
	result := slices.Clone(source)
	for i := range result {
		result[i].Files = slices.Clone(result[i].Files)
	}
	return result
}
