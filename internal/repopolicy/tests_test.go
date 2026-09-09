package repopolicy_test

import (
	"testing"
	"testing/fstest"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/internal/repopolicy"
)

func TestBlackboxDefaultAndDocumentedWhiteboxException(t *testing.T) {
	for _, row := range []struct {
		name, path, source, want string
	}{
		{"external", "engine_test.go", "package unswell_test", ""},
		{"undocumented internal", "engine_test.go", "package unswell", "whitebox test requires _internal_test.go.*"},
		{"named but undocumented", "engine_internal_test.go", "package unswell", "whitebox test requires a nonempty.*"},
		{"reason before package", "engine_internal_test.go", "// White-box tests: Inspect a private bound.\npackage unswell",
			"whitebox test requires a nonempty.*"},
		{"same package line", "engine_internal_test.go", "package unswell // White-box tests: Inspect a private bound.",
			"whitebox test requires a nonempty.*"},
		{"inside package clause", "engine_internal_test.go", "package\n// White-box tests: Inspect a private bound.\nunswell\n",
			"whitebox test requires a nonempty.*"},
		{"multiline package", "engine_internal_test.go", "package\nunswell\n// White-box tests: Inspect a private bound.\n", ""},
		{"empty reason", "engine_internal_test.go", "package unswell\n// White-box tests:  \n", "whitebox test requires a nonempty.*"},
		{"after declaration", "engine_internal_test.go", "package unswell\nvar x int\n// White-box tests: Inspect a private bound.\n",
			"whitebox test requires a nonempty.*"},
		{"after import", "engine_internal_test.go", "package unswell\nimport \"testing\"\n// White-box tests: Inspect a private bound.\n",
			"whitebox test requires a nonempty.*"},
		{"valid internal", "engine_internal_test.go", "package unswell\n// White-box tests: Inspect a private bound.\nimport \"testing\"\n", ""},
		{"header only", "engine_internal_test.go", "package unswell\n// White-box tests: Inspect a private bound.\n", ""},
		{"line directive", "engine_internal_test.go", "package unswell\n//line generated.go:1\n" +
			"// White-box tests: Inspect a private bound.\n", ""},
		{"line directive before package", "engine_internal_test.go", "// White-box tests: Inspect a private bound.\n" +
			"//line generated.go:1\npackage unswell\n", "whitebox test requires a nonempty.*"},
		{"build constraint", "engine_internal_test.go", "//go:build linux\n\npackage unswell\n\n" +
			"// White-box tests: Inject a private limit that the public API cannot set.\nimport \"testing\"\n", ""},
		{"CRLF", "engine_internal_test.go", "package unswell\r\n// White-box tests: Inspect a private bound.\r\nimport \"testing\"\r\n", ""},
		{"external reserved name", "engine_internal_test.go", "package unswell_test", "blackbox test must not use.*"},
		{"fixture", "testdata/broken_test.go", "not Go code", ""},
		{"fixture path lookalike", "mytestdata/broken_test.go", "package private", "whitebox test requires.*"},
		{"malformed source", "engine_internal_test.go", "package", ".*expected.*"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			tree := fixture()
			tree[row.path] = &fstest.MapFile{Data: []byte(row.source)}
			err := repopolicy.Check(tree)
			if row.want == "" {
				c.Assert(err, qt.IsNil)
			} else {
				c.Assert(err, qt.ErrorMatches, row.want)
			}
		})
	}
}

func TestBlackboxPolicyCoversNestedModulesAndCommands(t *testing.T) {
	for _, directory := range []string{"mcp", "research/annotation", "examples/consumer", "goanalysis", "tools", "cmd/helper"} {
		t.Run(directory, func(t *testing.T) {
			c := qt.New(t)
			tree := fixture()
			if directory != "cmd/helper" && directory != "tools" {
				tree[".gomodules"].Data = append(tree[".gomodules"].Data, []byte(directory+" runtime\n")...)
				tree[directory+"/go.mod"] = &fstest.MapFile{Data: []byte("module example.com/nested\n")}
			}
			tree[directory+"/main_test.go"] = &fstest.MapFile{Data: []byte("package main")}
			c.Assert(repopolicy.Check(tree), qt.ErrorMatches, "whitebox test requires _internal_test.go.*")
		})
	}
}
