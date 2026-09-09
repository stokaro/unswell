package report_test

import (
	"bytes"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/report"
)

func TestFeatureReportsReadLegacyMissingBindings(t *testing.T) {
	c := qt.New(t)
	result := featureReport(t)
	for i := range result.Features.Sources[0].Units {
		result.Features.Sources[0].Units[i].Binding = nil
	}
	var output bytes.Buffer
	c.Assert(report.Write(&output, "json", result, report.Options{}), qt.IsNil)
	c.Assert(output.String(), qt.Not(qt.Contains), `"binding"`)
	loaded, err := report.Read(&output)
	c.Assert(err, qt.IsNil)
	c.Assert(loaded, qt.DeepEquals, result)
}
