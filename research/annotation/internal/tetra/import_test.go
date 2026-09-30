package tetra_test

import (
	"bytes"
	"context"
	"crypto/sha1" // #nosec G505 -- Construct source fixtures in the upstream Git blob format.
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/research/annotation/internal/tetra"
)

func sourceFile(path, source string) tetra.SourceFile {
	raw := []byte(source)
	sum := sha256.Sum256(raw)
	// #nosec G401 -- Git fixture identity; no security or authentication decision is made.
	git := sha1.New()
	_, _ = fmt.Fprintf(git, "blob %d\x00", len(raw))
	_, _ = git.Write(raw)
	return tetra.SourceFile{Path: path, SHA256: hex.EncodeToString(sum[:]), GitBlob: hex.EncodeToString(git.Sum(nil)), XML: raw}
}

func xmlDocument(paper, editor, abstract string) string {
	return fmt.Sprintf(`<doc id=%q editor=%q format="Conf" position="S" region="NN">`+
		`<title><text>Title</text></title><abstract>%s</abstract><introduction><text>Body.</text></introduction></doc>`,
		paper, editor, abstract)
}

func input(files ...tetra.SourceFile) tetra.Input {
	return tetra.Input{Version: tetra.Version, Repository: "https://github.com/chemicaltree/tetra",
		Revision: "054ba8b308b7ed5a3a2aca9619427bc8f8fbd405", License: "CC-BY-4.0", IncludeXML: true, Files: files}
}

func TestImportPreservesCoupledUnicodeInsertionDeletionAndIdentity(t *testing.T) {
	c := qt.New(t)
	content := `<text>Café</text><edit type="Clarity; grammar; clarity" crr="can&#10;retry" comments="">may</edit>` +
		`<edit type="redundancy" crr="">very </edit><edit crr="again"></edit><edit crr="." comments=" ">.</edit>`
	file := sourceFile("original/P01-1001-A.xml", xmlDocument("P01-1001", "A", content))
	report, err := tetra.Import(context.Background(), input(file))
	c.Assert(err, qt.IsNil)
	c.Assert(report.Complete, qt.IsTrue)
	c.Assert(report.TaskLabelsAssigned, qt.IsFalse)
	c.Assert(report.ProductQualified, qt.IsFalse)
	c.Assert(report.Counts.Edits, qt.Equals, 4)
	section := report.Files[0].Sections[1]
	c.Assert(section.Original, qt.Equals, "Café may very   .")
	c.Assert(section.Revised, qt.Equals, "Café can\nretry  again .")
	c.Assert(section.Edits[0].OriginalSpan, qt.Equals, document.Span{Start: 6, End: 9})
	c.Assert(section.Edits[0].RevisedSpan, qt.Equals, document.Span{Start: 6, End: 15})
	c.Assert(section.Edits[1].RevisedSpan, qt.Equals, document.Span{Start: 16, End: 16})
	c.Assert(section.Edits[2].OriginalSpan, qt.Equals, document.Span{Start: 16, End: 16})
	c.Assert(section.Edits[2].RevisedSpan, qt.Equals, document.Span{Start: 17, End: 22})
	c.Assert(section.Edits[3].Unchanged, qt.IsTrue)
	c.Assert(section.Edits[0].Types, qt.DeepEquals, []string{"clarity", "grammar"})
	c.Assert(section.Edits[0].OriginalType, qt.Equals, "Clarity; grammar; clarity")
	c.Assert(*section.Edits[0].Comment, qt.Equals, "")
	c.Assert(section.Edits[2].Comment, qt.IsNil)
	c.Assert(section.Edits[2].TypeAvailable, qt.IsFalse)
	c.Assert(report.Files[0].XMLBase64, qt.IsNotNil)
	for _, edit := range section.Edits {
		c.Assert(section.Original[edit.OriginalSpan.Start:edit.OriginalSpan.End], qt.Equals, edit.Before)
		c.Assert(section.Revised[edit.RevisedSpan.Start:edit.RevisedSpan.End], qt.Equals, edit.After)
	}
}

func TestImportQuarantinesWholeFilenameGroupWithoutRenaming(t *testing.T) {
	c := qt.New(t)
	valid := sourceFile("original/P01-1001-A.xml", xmlDocument("P01-1001", "A", `<text>Text.</text>`))
	conflict := sourceFile("original/P01-1001-B.xml", xmlDocument("P01-1001 ", "C", `<edit crr="after">before</edit>`))
	other := sourceFile("original/P02-1002-A.xml", xmlDocument("P02-1002", "A", `<text>Other.</text>`))
	report, err := tetra.Import(context.Background(), input(other, conflict, valid))
	c.Assert(err, qt.ErrorIs, tetra.ErrIncomplete)
	c.Assert(report.Complete, qt.IsFalse)
	c.Assert(report.Counts.IdentityConflicts, qt.Equals, 1)
	c.Assert(report.Counts.QuarantinedFiles, qt.Equals, 2)
	c.Assert(report.Counts.EligibleGroups, qt.Equals, 1)
	c.Assert(report.Files[0].Status, qt.Equals, "quarantined")
	c.Assert(report.Files[1].Status, qt.Equals, "identity_conflict")
	c.Assert(report.Files[1].Paper, qt.Equals, "P01-1001")
	c.Assert(report.Files[1].DeclaredMetadata["id"], qt.Equals, "P01-1001 ")
	c.Assert(report.Files[1].DeclaredMetadata["editor"], qt.Equals, "C")
	c.Assert(report.Files[1].Sections[1].Edits[0].Before, qt.Equals, "before")
	c.Assert(report.Files[2].SourceEligible, qt.IsTrue)
	c.Assert(report.Groups[0].Files, qt.DeepEquals, []string{valid.Path, conflict.Path})
}

func TestImportPreservesInvalidSourceAndItsValidPeers(t *testing.T) {
	c := qt.New(t)
	valid := sourceFile("original/P01-1001-A.xml", xmlDocument("P01-1001", "A", `<text>Text.</text>`))
	broken := sourceFile("original/P01-1001-B.xml", `<doc><edit type= crr="after">before</edit></doc>`)
	report, err := tetra.Import(context.Background(), input(broken, valid))
	c.Assert(err, qt.ErrorIs, tetra.ErrIncomplete)
	c.Assert(report.Counts.Files, qt.Equals, 2)
	c.Assert(report.Counts.InvalidFiles, qt.Equals, 1)
	c.Assert(report.Counts.ParsedFiles, qt.Equals, 1)
	c.Assert(report.Counts.QuarantinedFiles, qt.Equals, 2)
	c.Assert(report.Files[1].Bytes, qt.Equals, len(broken.XML))
	c.Assert(report.Files[1].SHA256, qt.Equals, broken.SHA256)
	c.Assert(report.Files[1].GitBlob, qt.Equals, broken.GitBlob)
	c.Assert(report.Files[1].XMLBase64, qt.IsNotNil)
	c.Assert(report.Files[1].Sections, qt.HasLen, 0)
}

func TestImportRejectsUnsupportedXMLWithoutRepair(t *testing.T) {
	source := xmlDocument("P01-1001", "A", `<edit crr="after">before</edit>`)
	tests := map[string]string{
		"external entity":        `<!DOCTYPE doc [<!ENTITY remote SYSTEM "file:///not-read">]>` + source,
		"duplicate attribute":    strings.Replace(source, `editor="A"`, `editor="A" editor="A"`, 1),
		"unknown attribute":      strings.Replace(source, `editor="A"`, `editor="A" plugin="run"`, 1),
		"nested markup":          strings.Replace(source, "before", `<b>before</b>`, 1),
		"missing correction":     strings.Replace(source, ` crr="after"`, "", 1),
		"repeated section":       strings.Replace(source, "<abstract>", "<title></title><abstract>", 1),
		"unknown section":        strings.ReplaceAll(source, "abstract", "conclusion"),
		"namespace":              strings.Replace(source, "<doc ", `<doc xmlns="unknown" `, 1),
		"extra root":             source + source,
		"unexpected prose":       strings.Replace(source, "<abstract>", "<abstract>unowned", 1),
		"root stray marker":      strings.Replace(source, "<title>", ">\n<title>", 1),
		"leaf stray punctuation": strings.Replace(source, "</text></title>", "</text>,\n</title>", 1),
		"unowned Unicode space":  strings.Replace(source, "</text></title>", "</text>\u00a0</title>", 1),
		"processing instruction": strings.Replace(source, "before", `<?run task?>before`, 1),
		"attribute whitespace":   strings.Replace(source, `crr="after"`, "crr=\"after\nnext\"", 1),
	}
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			report, err := tetra.Import(context.Background(), input(sourceFile("original/P01-1001-A.xml", raw)))
			c.Assert(err, qt.ErrorIs, tetra.ErrIncomplete)
			c.Assert(report.Files[0].Status, qt.Equals, "invalid_xml")
			c.Assert(report.Files[0].Issues, qt.HasLen, 1)
			c.Assert(report.Files[0].SourceEligible, qt.IsFalse)
			c.Assert(report.Files[0].Sections, qt.HasLen, 0)
		})
	}
}

func TestDocumentedTutorialMatchesSourceBoundGolden(t *testing.T) {
	c := qt.New(t)
	data, err := os.ReadFile("testdata/tutorial-input.json")
	c.Assert(err, qt.IsNil)
	want, err := os.ReadFile("testdata/tutorial-output.json")
	c.Assert(err, qt.IsNil)
	var output bytes.Buffer
	c.Assert(tetra.Run(context.Background(), bytes.NewReader(data), &output), qt.IsNil)
	var got, expected tetra.Report
	c.Assert(json.Unmarshal(output.Bytes(), &got), qt.IsNil)
	c.Assert(json.Unmarshal(want, &expected), qt.IsNil)
	c.Assert(got, qt.DeepEquals, expected)
	xml, err := os.ReadFile("testdata/tutorial.xml")
	c.Assert(err, qt.IsNil)
	var in tetra.Input
	c.Assert(json.Unmarshal(data, &in), qt.IsNil)
	c.Assert(in.Files[0].XML, qt.DeepEquals, xml)
}

func TestImportChecksKnownGitBlobIdentity(t *testing.T) {
	c := qt.New(t)
	file := tetra.SourceFile{Path: "original/P01-1001-A.xml", XML: []byte("abc"),
		SHA256:  "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
		GitBlob: "f2ba8f84ab5c1bce84a7b441cb1959cfc7093b7f"}
	report, err := tetra.Import(context.Background(), input(file))
	c.Assert(err, qt.ErrorIs, tetra.ErrIncomplete)
	c.Assert(report.Files[0].Status, qt.Equals, "invalid_xml")
}

func TestImportValidatesEntireInventoryBeforeReturningPairs(t *testing.T) {
	valid := sourceFile("original/P01-1001-A.xml", xmlDocument("P01-1001", "A", `<text>Text.</text>`))
	tests := map[string]func(*tetra.Input){
		"SHA256":          func(in *tetra.Input) { in.Files[0].SHA256 = strings.Repeat("0", 64) },
		"Git blob":        func(in *tetra.Input) { in.Files[0].GitBlob = strings.Repeat("0", 40) },
		"changed bytes":   func(in *tetra.Input) { in.Files[0].XML = []byte("changed") },
		"repeated path":   func(in *tetra.Input) { in.Files = append(in.Files, in.Files[0]) },
		"path traversal":  func(in *tetra.Input) { in.Files[0].Path = "../P01-1001-A.xml" },
		"uppercase hash":  func(in *tetra.Input) { in.Files[0].SHA256 = strings.ToUpper(in.Files[0].SHA256) },
		"version":         func(in *tetra.Input) { in.Version = "unsupported" },
		"license":         func(in *tetra.Input) { in.License = "MIT" },
		"revision":        func(in *tetra.Input) { in.Revision = "main" },
		"empty inventory": func(in *tetra.Input) { in.Files = nil },
		"file bytes":      func(in *tetra.Input) { in.Files[0] = sourceFile(valid.Path, strings.Repeat("x", 262145)) },
	}
	for name, change := range tests {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			in := input(valid)
			change(&in)
			report, err := tetra.Import(context.Background(), in)
			c.Assert(err, qt.IsNotNil)
			c.Assert(err, qt.Not(qt.ErrorIs), tetra.ErrIncomplete)
			c.Assert(report.Files, qt.IsNil)
		})
	}
}

func TestImportIsDeterministicForReorderedFiles(t *testing.T) {
	c := qt.New(t)
	files := []tetra.SourceFile{
		sourceFile("original/P01-1001-A.xml", xmlDocument("P01-1001", "A", `<text>A.</text>`)),
		sourceFile("original/P01-1001-B.xml", xmlDocument("P01-1001", "B", `<text>B.</text>`)),
	}
	first, err := tetra.Import(context.Background(), input(files...))
	c.Assert(err, qt.IsNil)
	slices.Reverse(files)
	second, err := tetra.Import(context.Background(), input(files...))
	c.Assert(err, qt.IsNil)
	c.Assert(first, qt.DeepEquals, second)
}

type badWriter struct{}

func (badWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

type shortWriter struct{}

func (shortWriter) Write([]byte) (int, error) { return 0, nil }

func TestCommandStrictInputAndWriterFailure(t *testing.T) {
	c := qt.New(t)
	in := input(sourceFile("original/P01-1001-A.xml", xmlDocument("P01-1001", "A", `<text>Text.</text>`)))
	encoded, err := json.Marshal(in)
	c.Assert(err, qt.IsNil)
	valid := string(encoded)
	for _, raw := range []string{
		strings.Replace(valid, `"version":`, `"unknown":true,"version":`, 1),
		strings.Replace(valid, `"version":`, `"version":"other","version":`, 1),
		strings.Replace(valid, `"version":`, `"Version":`, 1),
		strings.Replace(valid, `"include_xml":true`, `"include_xml":null`, 1),
		strings.Replace(valid, `"license":`, `"extra":null,"license":`, 1),
		valid + "{}",
	} {
		var output bytes.Buffer
		c.Assert(tetra.Run(context.Background(), strings.NewReader(raw), &output), qt.IsNotNil)
		c.Assert(output.Len(), qt.Equals, 0)
	}
	c.Assert(tetra.Run(context.Background(), strings.NewReader(valid), badWriter{}), qt.ErrorMatches, "write failed")
	c.Assert(tetra.Run(context.Background(), strings.NewReader(valid), shortWriter{}), qt.Equals, io.ErrShortWrite)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c.Assert(tetra.Run(ctx, strings.NewReader(valid), badWriter{}), qt.Equals, context.Canceled)
}

func TestCommandEmitsIncompleteQuarantineRecord(t *testing.T) {
	c := qt.New(t)
	in := input(sourceFile("original/P01-1001-A.xml", "not XML"))
	in.IncludeXML = false
	encoded, err := json.Marshal(in)
	c.Assert(err, qt.IsNil)
	var output bytes.Buffer
	c.Assert(tetra.Run(context.Background(), bytes.NewReader(encoded), &output), qt.Equals, tetra.ErrIncomplete)
	var report tetra.Report
	c.Assert(json.Unmarshal(output.Bytes(), &report), qt.IsNil)
	c.Assert(report.Complete, qt.IsFalse)
	c.Assert(report.Files[0].XMLBase64, qt.IsNil)
	c.Assert(report.Counts.InvalidFiles, qt.Equals, 1)
}
