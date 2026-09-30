package tetra

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
)

type parsedDocument struct {
	metadata map[string]string
	sections []Section
}

type leaf struct {
	before   string
	after    string
	edit     bool
	typeText *string
	comment  *string
}

type xmlSource struct {
	context context.Context
	decoder *xml.Decoder
	raw     []byte
}

func (source xmlSource) next() (xml.Token, error) {
	if err := source.context.Err(); err != nil {
		return nil, err
	}
	start := source.decoder.InputOffset()
	token, err := source.decoder.Token()
	if err != nil {
		return nil, err
	}
	if _, ok := token.(xml.Directive); ok {
		return nil, fmt.Errorf("XML directives and DTD/entity declarations are unsupported")
	}
	if _, ok := token.(xml.StartElement); ok {
		if literalAttributeWhitespace(source.raw[start:source.decoder.InputOffset()]) {
			return nil, fmt.Errorf("literal tabs or line breaks in XML attribute values are unsupported")
		}
	}
	return token, nil
}

// The standard decoder does not apply XML's attribute whitespace normalization.
// Reject its unsupported literal form explicitly rather than changing a revision.
// Numeric character references remain supported and retain their decoded values.
func literalAttributeWhitespace(raw []byte) bool {
	var quote byte
	for _, value := range raw {
		if quote == 0 {
			if value == '\'' || value == '"' {
				quote = value
			}
			continue
		}
		switch value {
		case quote:
			quote = 0
		case '\t', '\r', '\n':
			return true
		}
	}
	return false
}

func parseXML(ctx context.Context, raw []byte, paper, editor string) (parsedDocument, error) {
	source := xmlSource{context: ctx, decoder: xml.NewDecoder(bytes.NewReader(raw)), raw: raw}
	root, err := source.root()
	if err != nil {
		return parsedDocument{}, err
	}
	metadata, err := attributes(root, []string{"id", "editor", "format", "position", "region"})
	if err != nil {
		return parsedDocument{}, err
	}
	if _, ok := metadata["id"]; !ok {
		return parsedDocument{}, fmt.Errorf("document lacks an explicit paper identity")
	}
	if _, ok := metadata["editor"]; !ok {
		return parsedDocument{}, fmt.Errorf("document lacks an explicit editor identity")
	}
	sections, err := source.sections(paper, editor)
	if err != nil {
		return parsedDocument{}, err
	}
	if err := source.finish(); err != nil {
		return parsedDocument{}, err
	}
	return parsedDocument{metadata: metadata, sections: sections}, nil
}

func (source xmlSource) root() (xml.StartElement, error) {
	for {
		token, err := source.next()
		if err != nil {
			return xml.StartElement{}, fmt.Errorf("read XML root: %w", err)
		}
		switch node := token.(type) {
		case xml.StartElement:
			if node.Name.Space != "" || node.Name.Local != "doc" {
				return xml.StartElement{}, fmt.Errorf("unsupported XML root")
			}
			return node, nil
		default:
			if err := preamble(token); err != nil {
				return xml.StartElement{}, err
			}
		}
	}
}

func preamble(token xml.Token) error {
	if node, ok := token.(xml.ProcInst); ok && node.Target == "xml" {
		return nil
	}
	return indentation(token)
}

func attributes(node xml.StartElement, allowed []string) (map[string]string, error) {
	values := make(map[string]string)
	for _, attr := range node.Attr {
		_, duplicate := values[attr.Name.Local]
		if attr.Name.Space != "" || duplicate || !slices.Contains(allowed, attr.Name.Local) {
			return nil, fmt.Errorf("unknown, namespaced or repeated XML attribute")
		}
		values[attr.Name.Local] = attr.Value
	}
	return values, nil
}

func (source xmlSource) sections(paper, editor string) ([]Section, error) {
	names := []string{"title", "abstract", "introduction"}
	sections := make([]Section, 0, len(names))
	for {
		token, err := source.next()
		if err != nil {
			return nil, fmt.Errorf("read XML document: %w", err)
		}
		switch node := token.(type) {
		case xml.StartElement:
			section, err := source.readSection(node, names, len(sections), paper, editor)
			if err != nil {
				return nil, err
			}
			sections = append(sections, section)
		case xml.EndElement:
			if node.Name.Local != "doc" || len(sections) != len(names) {
				return nil, fmt.Errorf("missing XML sections")
			}
			return sections, nil
		default:
			if err := indentation(token); err != nil {
				return nil, err
			}
		}
	}
}

func (source xmlSource) readSection(node xml.StartElement, names []string, index int, paper, editor string) (Section, error) {
	if index == len(names) || node.Name.Space != "" || node.Name.Local != names[index] || len(node.Attr) != 0 {
		return Section{}, fmt.Errorf("unknown, reordered or repeated XML section")
	}
	leaves, err := source.leaves(node.Name)
	if err != nil {
		return Section{}, err
	}
	return renderSection(paper, editor, node.Name.Local, leaves), nil
}

func (source xmlSource) leaves(section xml.Name) ([]leaf, error) {
	leaves := make([]leaf, 0)
	for {
		token, err := source.next()
		if err != nil {
			return nil, fmt.Errorf("read XML section: %w", err)
		}
		switch node := token.(type) {
		case xml.StartElement:
			part, err := source.typedLeaf(node)
			if err != nil {
				return nil, err
			}
			leaves = append(leaves, part)
		case xml.EndElement:
			if node.Name != section {
				return nil, fmt.Errorf("unexpected XML section boundary")
			}
			return leaves, nil
		default:
			if err := indentation(token); err != nil {
				return nil, err
			}
		}
	}
}

func (source xmlSource) typedLeaf(node xml.StartElement) (leaf, error) {
	if node.Name.Space != "" || (node.Name.Local != "text" && node.Name.Local != "edit") {
		return leaf{}, fmt.Errorf("unknown XML text/edit element")
	}
	return source.readLeaf(node)
}

func (source xmlSource) readLeaf(node xml.StartElement) (leaf, error) {
	allowed := []string{}
	if node.Name.Local == "edit" {
		allowed = []string{"type", "crr", "comments"}
	}
	attrs, err := attributes(node, allowed)
	if err != nil {
		return leaf{}, err
	}
	text, err := source.leafText(node.Name)
	if err != nil {
		return leaf{}, err
	}
	part := leaf{before: text, after: text, edit: node.Name.Local == "edit"}
	if !part.edit {
		return part, nil
	}
	corrected, present := attrs["crr"]
	if !present {
		return leaf{}, fmt.Errorf("edit lacks corrected text; empty correction must be explicit")
	}
	part.after = corrected
	if value, ok := attrs["type"]; ok {
		part.typeText = &value
	}
	if value, ok := attrs["comments"]; ok {
		part.comment = &value
	}
	return part, nil
}

func (source xmlSource) leafText(name xml.Name) (string, error) {
	var text strings.Builder
	for {
		token, err := source.next()
		if err != nil {
			return "", fmt.Errorf("read XML leaf: %w", err)
		}
		switch node := token.(type) {
		case xml.CharData:
			text.Write(node)
		case xml.Comment:
		case xml.EndElement:
			if node.Name != name {
				return "", fmt.Errorf("unexpected XML leaf boundary")
			}
			return text.String(), nil
		default:
			return "", fmt.Errorf("nested elements or processing instructions in XML leaves are unsupported")
		}
	}
}

func indentation(token xml.Token) error {
	switch node := token.(type) {
	case xml.Comment:
		return nil
	case xml.CharData:
		if strings.Trim(string(node), " \t\r\n") == "" {
			return nil
		}
	}
	return fmt.Errorf("unexpected content outside XML text/edit leaves")
}

func (source xmlSource) finish() error {
	for {
		token, err := source.next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read XML suffix: %w", err)
		}
		if err := indentation(token); err != nil {
			return err
		}
	}
}

func renderSection(paper, editor, name string, leaves []leaf) Section {
	var original, revised strings.Builder
	section := Section{Name: name, Edits: make([]Edit, 0)}
	for index, part := range leaves {
		if index > 0 {
			original.WriteByte(' ')
			revised.WriteByte(' ')
		}
		before := document.Span{Start: original.Len(), End: original.Len() + len(part.before)}
		after := document.Span{Start: revised.Len(), End: revised.Len() + len(part.after)}
		original.WriteString(part.before)
		revised.WriteString(part.after)
		if part.edit {
			edit := Edit{ID: fmt.Sprintf("%s-%s/%s/%04d", paper, editor, name, index), ElementIndex: index,
				Before: part.before, After: part.after, OriginalSpan: before, RevisedSpan: after,
				Types: []string{}, Comment: part.comment, Unchanged: part.before == part.after}
			if part.typeText != nil {
				edit.TypeAvailable, edit.OriginalType = true, *part.typeText
				edit.Types = editTypes(*part.typeText)
			}
			section.Edits = append(section.Edits, edit)
		}
	}
	section.Original, section.Revised = original.String(), revised.String()
	section.OriginalSHA256, section.RevisedSHA256 = digest([]byte(section.Original)), digest([]byte(section.Revised))
	return section
}

func editTypes(value string) []string {
	types := make([]string, 0)
	for kind := range strings.SplitSeq(value, ";") {
		kind = strings.ToLower(strings.TrimSpace(kind))
		if kind != "" && !slices.Contains(types, kind) {
			types = append(types, kind)
		}
	}
	slices.Sort(types)
	return types
}
