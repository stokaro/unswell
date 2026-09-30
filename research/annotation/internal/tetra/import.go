package tetra

import (
	"context"
	"crypto/sha1" // #nosec G505 -- Required to verify existing Git SHA-1 object identities.
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
)

var filenamePattern = regexp.MustCompile(`^original/([PW][0-9]{2}-[0-9]{4})-([ABC])\.xml$`)

// Import validates the complete input before parsing any source. Invalid XML and
// identity conflicts preserve their records and quarantine the whole paper group.
// Such a report returns ErrIncomplete; operational failures return no report.
func Import(ctx context.Context, input Input) (Report, error) {
	if err := validateInput(ctx, input); err != nil {
		return Report{}, err
	}
	out := Report{Version: Version, Repository: input.Repository, Revision: input.Revision, License: input.License,
		Rendering: "Join decoded XML text/edit leaves with one ASCII space; spans refer only to reconstructed UTF-8 sections.",
		Target:    "Upstream editor preference; neither revision necessity nor technical safety is inferred.",
		Files:     make([]File, 0, len(input.Files)), Groups: make([]Group, 0), Complete: true}
	for _, source := range input.Files {
		file, err := importFile(ctx, source, input.IncludeXML)
		if err != nil {
			return Report{}, err
		}
		out.Files = append(out.Files, file)
	}
	slices.SortFunc(out.Files, func(a, b File) int { return strings.Compare(a.Path, b.Path) })
	groupFiles(&out)
	countFiles(&out)
	if !out.Complete {
		return out, ErrIncomplete
	}
	return out, nil
}

func validateInput(ctx context.Context, input Input) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateIdentity(input); err != nil {
		return err
	}
	if len(input.Files) == 0 || len(input.Files) > maximumFiles {
		return fmt.Errorf("missing or excessive TETRA source inventory")
	}
	seen, total := make(map[string]bool), 0
	for _, file := range input.Files {
		if err := ctx.Err(); err != nil {
			return err
		}
		if seen[file.Path] {
			return fmt.Errorf("repeated TETRA source path %q", file.Path)
		}
		seen[file.Path] = true
		if err := validateFile(file); err != nil {
			return err
		}
		total += len(file.XML)
		if total > maximumSourceBytes {
			return fmt.Errorf("TETRA source byte limit exceeded")
		}
	}
	return nil
}

func validateIdentity(input Input) error {
	if input.Version != Version || input.Repository != "https://github.com/chemicaltree/tetra" ||
		!validHex(input.Revision, 40) || input.License != "CC-BY-4.0" {
		return fmt.Errorf("unsupported TETRA input identity or license declaration")
	}
	return nil
}

func validateFile(file SourceFile) error {
	if !filenamePattern.MatchString(file.Path) {
		return fmt.Errorf("invalid TETRA source path %q", file.Path)
	}
	if len(file.XML) > maximumXMLBytes {
		return fmt.Errorf("TETRA source byte limit exceeded")
	}
	if !validHex(file.SHA256, 64) || !validHex(file.GitBlob, 40) ||
		file.SHA256 != digest(file.XML) || file.GitBlob != gitBlob(file.XML) {
		return fmt.Errorf("TETRA source hashes do not match %q", file.Path)
	}
	return nil
}

func validHex(value string, length int) bool {
	if len(value) != length || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func digest(raw []byte) string {
	value := sha256.Sum256(raw)
	return hex.EncodeToString(value[:])
}

func gitBlob(raw []byte) string {
	// #nosec G401 -- This matches Git's existing object format; SHA-256 also validates every source.
	hash := sha1.New()
	_, _ = fmt.Fprintf(hash, "blob %d\x00", len(raw))
	_, _ = hash.Write(raw)
	return hex.EncodeToString(hash.Sum(nil))
}

func importFile(ctx context.Context, source SourceFile, includeXML bool) (File, error) {
	parts := filenamePattern.FindStringSubmatch(source.Path)
	file := File{Path: source.Path, Paper: parts[1], Editor: parts[2], SHA256: source.SHA256,
		GitBlob: source.GitBlob, Bytes: len(source.XML), DeclaredMetadata: map[string]string{},
		Status: "valid", SourceEligible: true, Issues: []Issue{}, Sections: []Section{}}
	if includeXML {
		encoded := base64.StdEncoding.EncodeToString(source.XML)
		file.XMLBase64 = &encoded
	}
	parsed, err := parseXML(ctx, source.XML, file.Paper, file.Editor)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return File{}, err
		}
		file.Status, file.SourceEligible = "invalid_xml", false
		file.Issues = append(file.Issues, xmlIssue(err))
		return file, nil
	}
	file.DeclaredMetadata, file.Sections = parsed.metadata, parsed.sections
	if parsed.metadata["id"] != file.Paper || parsed.metadata["editor"] != file.Editor {
		file.Status, file.SourceEligible = "identity_conflict", false
		file.Issues = append(file.Issues, Issue{Code: "identity_conflict",
			Reason: "Declared paper/editor differs from filename identity; no identity is repaired or renamed."})
	}
	return file, nil
}

func xmlIssue(err error) Issue {
	code := "unsupported_xml"
	var syntax *xml.SyntaxError
	if errors.As(err, &syntax) || errors.Is(err, io.EOF) {
		code = "xml_syntax_error"
	}
	return Issue{Code: code, Reason: err.Error()}
}

func groupFiles(out *Report) {
	index := make(map[string]int)
	for _, file := range out.Files {
		n, exists := index[file.Paper]
		if !exists {
			n = len(out.Groups)
			index[file.Paper] = n
			out.Groups = append(out.Groups, Group{Paper: file.Paper, Files: []string{}, SourceEligible: true})
		}
		group := &out.Groups[n]
		group.Files = append(group.Files, file.Path)
		group.SourceEligible = group.SourceEligible && file.SourceEligible
	}
	for i := range out.Files {
		file := &out.Files[i]
		if !out.Groups[index[file.Paper]].SourceEligible {
			out.Complete, file.SourceEligible = false, false
			if file.Status == "valid" {
				file.Status = "quarantined"
				file.Issues = append(file.Issues, Issue{Code: "paper_group_quarantined",
					Reason: "Another filename version of this paper has an XML or identity failure."})
			}
		}
	}
}

func countFiles(out *Report) {
	out.Counts.Files, out.Counts.Groups = len(out.Files), len(out.Groups)
	for _, group := range out.Groups {
		if group.SourceEligible {
			out.Counts.EligibleGroups++
		}
	}
	for _, file := range out.Files {
		switch file.Status {
		case "invalid_xml":
			out.Counts.InvalidFiles++
		case "identity_conflict":
			out.Counts.IdentityConflicts++
		}
		if file.Status != "invalid_xml" {
			out.Counts.ParsedFiles++
		}
		if !file.SourceEligible {
			out.Counts.QuarantinedFiles++
		}
		for _, section := range file.Sections {
			out.Counts.Edits += len(section.Edits)
			if file.SourceEligible {
				out.Counts.EligibleEdits += len(section.Edits)
			}
		}
	}
}
