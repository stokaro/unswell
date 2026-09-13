package extract

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"unicode/utf8"

	ts "github.com/stokaro/gotreesitter"
	"go.yaml.in/yaml/v3"

	"github.com/stokaro/unswell/document"
)

type yamlPosition struct{ line, column int }

type yamlValue struct {
	text, tag string
	key       bool
}

// The grammar supplies structural spans. The existing YAML decoder supplies
// scalar semantics, including folding, chomping, and explicit tags. Decoding
// into Node preserves aliases as references without constructing application data.
func yamlValues(ctx context.Context, source []byte) (map[int]yamlValue, error) {
	positions := make(map[yamlPosition]yamlValue)
	decoder := yaml.NewDecoder(bytes.NewReader(source))
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var root yaml.Node
		if err := decoder.Decode(&root); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("decode YAML scalars: %w", err)
		}
		if err := collectYAMLValues(ctx, &root, false, 0, positions); err != nil {
			return nil, err
		}
	}
	return indexYAMLPositions(ctx, source, positions)
}

func collectYAMLValues(ctx context.Context, node *yaml.Node, key bool, depth int, positions map[yamlPosition]yamlValue) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if depth > 128 {
		return fmt.Errorf("YAML nesting exceeds 128")
	}
	if node.Kind == yaml.ScalarNode {
		positions[yamlPosition{node.Line, node.Column}] = yamlValue{text: node.Value, tag: node.ShortTag(), key: key}
	}
	for i, child := range node.Content {
		if err := collectYAMLValues(ctx, child, key || node.Kind == yaml.MappingNode && i%2 == 0, depth+1, positions); err != nil {
			return err
		}
	}
	return nil
}

func indexYAMLPositions(ctx context.Context, source []byte, positions map[yamlPosition]yamlValue) (map[int]yamlValue, error) {
	result := make(map[int]yamlValue, len(positions))
	position := yamlPosition{1, 1}
	for offset := 0; offset < len(source); {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if value, ok := positions[position]; ok {
			result[offset] = value
		}
		char, size := utf8.DecodeRune(source[offset:])
		offset += size
		position.column++
		if char == '\n' || char == '\r' && (offset == len(source) || source[offset] != '\n') {
			position.line++
			position.column = 1
		}
		if char == '\ufeff' && offset == size {
			position.column = 1
		}
	}
	return result, nil
}

func (r *sourceReader) yamlNode(node *ts.Node) (bool, error) {
	kind := node.Type(r.syntax.lang)
	if kind == "alias" {
		r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: syntaxSpan(node, 0), Reason: "yaml-alias"})
		return true, nil
	}
	if !slices.Contains([]string{"plain_scalar", "double_quote_scalar", "single_quote_scalar", "block_scalar"}, kind) {
		return false, nil
	}
	value, err := r.yamlValue(node)
	if err != nil {
		return true, err
	}
	reason := yamlExclusion(value)
	if reason == "" {
		reason = r.exception("string", node)
	}
	if reason != "" {
		r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: syntaxSpan(node, 0), Reason: reason})
		return true, nil
	}
	if contextExclusion(r.doc, r.options.Policy, "string", syntaxSpan(node, 0)) {
		return true, nil
	}
	mapped, err := mapYAMLScalar(r.doc.Source, syntaxSpan(node, 0), kind, value.text)
	if err != nil {
		return true, fmt.Errorf("YAML scalar at byte %d: %w", node.StartByte(), err)
	}
	appendBlock(r.doc, mapped, "string")
	return true, nil
}

func (r *sourceReader) yamlValue(node *ts.Node) (yamlValue, error) {
	for parent := node; parent != nil; parent = parent.Parent() {
		if value, ok := r.yamlValues[int(parent.StartByte())]; ok {
			return value, nil
		}
		if slices.Contains([]string{"flow_node", "block_node"}, parent.Type(r.syntax.lang)) {
			break
		}
	}
	return yamlValue{}, fmt.Errorf("YAML grammar and scalar decoder disagree at byte %d", node.StartByte())
}

func yamlExclusion(value yamlValue) string {
	if value.key {
		return "yaml-key"
	}
	if value.tag != "!!str" {
		return "yaml-non-string:" + value.tag
	}
	return ""
}
