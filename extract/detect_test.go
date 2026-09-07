package extract_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestFormatDetection(t *testing.T) {
	cases := []struct {
		name, prefix string
		format       document.Format
	}{
		{"docs/README.markdown", "", document.Markdown},
		{"program.C", "", document.CPP},
		{"source.c", "", document.C},
		{"src/Client.cs", "", document.CSharp},
		{"script.csx", "", document.CSharp},
		{"config.yaml", "", document.YAML},
		{"pipeline.YML", "", document.YAML},
		{"home/.bashrc", "", document.Bash},
		{"home/.zshrc", "", document.Zsh},
		{"home/.profile", "", document.Shell},
		{"tool.fish", "", document.Fish},
		{"script.PS1", "", document.PowerShell},
		{"run", "#!/usr/bin/env -S bash -eu\necho text", document.Bash},
		{"run", "#!/bin/dash\n", document.Shell},
		{"run.sh", "#!/bin/bash\n", document.Bash},
		{"run", "#!/usr/bin/env python3\n", document.Python},
		{"run", "#!/usr/bin/env pwsh\n", document.PowerShell},
		{"run", "echo text", ""},
		{"run", "#!/usr/bin/env --unknown bash\n", ""},
		{"run", "#!/usr/bin/perl\n", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name+tc.prefix, func(t *testing.T) {
			c := qt.New(t)
			format, found := extract.Detect(tc.name, []byte(tc.prefix))
			c.Assert(format, qt.Equals, tc.format)
			c.Assert(found, qt.Equals, tc.format != "")
		})
	}
}
