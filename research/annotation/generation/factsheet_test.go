package generation_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/generation"
)

func TestFactSheetReadsTheDeclarationAfterTheUnit(t *testing.T) {
	for _, row := range []struct {
		name, source, original, signature, parameters string
		identifiers                                   []string
		numbers                                       []string
	}{
		{"go method", "// GetLevel returns the logger level.\nfunc (logger *Logger) GetLevel() Level {\n\treturn logger.level\n}\n",
			"GetLevel returns the logger level.", "func (logger *Logger) GetLevel() Level", "",
			[]string{"logger", "Logger", "GetLevel", "Level"}, nil},
		{"go method with parameters", "// Add adds.\nfunc (m *Mock) AssertExpectations(t TestingT) bool {\n",
			"Add adds two values in 3 steps.", "func (m *Mock) AssertExpectations(t TestingT) bool", "t TestingT",
			[]string{"Mock", "AssertExpectations", "TestingT"}, []string{"3"}},
		{"python trailing comment", "# Type ignored.\ndef __enter__(self) -> \"WarningsRecorder\":  # type: ignore\n    pass\n",
			"Type ignored.", "def __enter__(self) -> \"WarningsRecorder\"", "self", []string{"__enter__", "WarningsRecorder"}, nil},
		{"rust multi-line", "/// Creates it.\npub fn with_name(\n    n: &'a str,\n) -> Self {\n",
			"Creates it.", "pub fn with_name( n: &'a str, ) -> Self", "n: &'a str,", []string{"with_name", "str", "Self"}, nil},
		{"csharp after xml comment", "/// <summary>Filters.</summary>\n/// <param name=\"p\">x</param>\n" +
			"public LoggerConfiguration ByExcluding(Func<LogEvent, bool> p)\n{\n",
			"Filters.", "public LoggerConfiguration ByExcluding(Func<LogEvent, bool> p)", "Func<LogEvent, bool> p",
			[]string{"LoggerConfiguration", "ByExcluding", "Func", "LogEvent"}, nil},
		{"javascript initializer", "/** Used to match. */\nvar reEscapedHtml = /&(?:amp|lt);/g;\n",
			"Used to match 2 kinds.", "var reEscapedHtml = /&(?:amp|lt);/g", "", []string{"reEscapedHtml", "amp", "lt"}, []string{"2"}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			// Every fixture keeps its comment on the first line, so the unit ends there.
			sheet, ok := generation.ExtractFactSheet([]byte(row.source), indexAfterComment(row.source), row.original)
			c.Assert(ok, qt.IsTrue)
			c.Assert(sheet.Version, qt.Equals, generation.FactSheetVersion)
			c.Assert(sheet.Signature, qt.Equals, row.signature)
			c.Assert(sheet.Parameters, qt.Equals, row.parameters)
			c.Assert(sheet.Identifiers, qt.DeepEquals, row.identifiers)
			if row.numbers == nil {
				c.Assert(sheet.Numbers, qt.HasLen, 0)
			} else {
				c.Assert(sheet.Numbers, qt.DeepEquals, row.numbers)
			}
			c.Assert(sheet.SHA256, qt.HasLen, 64)
			c.Assert(sheet.Text(), qt.Contains, "Signature:\n"+row.signature)
		})
	}
}

// indexAfterComment returns the offset just past the first line, which every
// fixture uses for its comment.
func indexAfterComment(source string) int {
	for i, r := range source {
		if r == '\n' {
			return i
		}
	}
	return len(source)
}

func TestFactSheetRefusesUnitsWithoutADeclaration(t *testing.T) {
	c := qt.New(t)
	for _, row := range []struct{ name, source string }{
		{"blank line then code", "// A note.\n\nfunc later() {}\n"},
		{"only comments follow", "// A note.\n// Another note.\n"},
		{"prose follows", "// A note.\nthe end of the file\n"},
		{"end of file", "// A note."},
	} {
		end := indexAfterComment(row.source)
		_, ok := generation.ExtractFactSheet([]byte(row.source), end, "A note.")
		c.Assert(ok, qt.IsFalse, qt.Commentf("%s", row.name))
	}
	_, ok := generation.ExtractFactSheet([]byte("x"), 5, "A note.")
	c.Assert(ok, qt.IsFalse)
}
