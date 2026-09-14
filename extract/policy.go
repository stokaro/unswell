package extract

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	ts "github.com/stokaro/gotreesitter"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/pathglob"
)

// Policy selects prose contexts and explicit comment and string exceptions.
// GitHubActions accepts shell (also the zero-value default) or strings.
// The zero value stays out of the serialized form. A policy is hashed to
// identify a frozen corpus artifact. Writing an unstated default there would
// renumber every artifact frozen before the field, for no change in behavior.
// GoComments accepts godoc (also the zero-value default) or plain.
type Policy struct {
	GoComments      string                             `json:"go_comments" yaml:"go_comments"`
	MarkdownStrings []MarkdownString                   `json:"markdown_strings,omitempty" yaml:"markdown_strings,omitempty"`
	GitHubActions   string                             `json:"github_actions,omitempty" yaml:"github_actions,omitempty"`
	Contexts        []string                           `json:"contexts"   yaml:"contexts"`
	Languages       map[document.Format]LanguagePolicy `json:"languages"  yaml:"languages"`
	Exceptions      []Exception                        `json:"exceptions" yaml:"exceptions"`
}

// MarkdownString selects static string literals for Markdown extraction.
// Paths is required. Nonempty selectors must all match; their values are alternatives.
// Symbols uses the same enclosing declaration and keyed-field names as Exception.
type MarkdownString struct {
	ID      string            `json:"id" yaml:"id"`
	Paths   []string          `json:"paths" yaml:"paths"`
	Formats []document.Format `json:"formats,omitempty" yaml:"formats,omitempty"`
	Symbols []string          `json:"symbols,omitempty" yaml:"symbols,omitempty"`
}

// LanguagePolicy replaces the global context set for one input format.
// A non-nil empty set disables every prose context for that format.
type LanguagePolicy struct {
	Contexts []string `json:"contexts" yaml:"contexts"`
}

// Exception omits a region only when every nonempty selector matches.
// Paths and Kinds are required. Symbols selects enclosing named declarations
// or keyed fields and is supported for strings only. Values within a selector
// are alternatives. Reasons and IDs are retained in document exclusions.
type Exception struct {
	ID      string            `json:"id"      yaml:"id"`
	Paths   []string          `json:"paths"   yaml:"paths"`
	Formats []document.Format `json:"formats" yaml:"formats"`
	Kinds   []string          `json:"kinds"   yaml:"kinds"`
	Symbols []string          `json:"symbols" yaml:"symbols"`
	Reason  string            `json:"reason"  yaml:"reason"`
}

type compiledException struct {
	value Exception
	paths []*regexp.Regexp
}

// ValidatePolicy checks every exception before any source is analyzed.
func ValidatePolicy(policy Policy) error {
	if policy.GoComments != "" && policy.GoComments != "godoc" && policy.GoComments != "plain" {
		return fmt.Errorf("extraction go_comments must be godoc or plain")
	}
	if policy.GitHubActions != "" && policy.GitHubActions != "shell" && policy.GitHubActions != "strings" {
		return fmt.Errorf("extraction github_actions must be shell or strings")
	}
	if err := validateContexts(policy); err != nil {
		return err
	}
	if _, err := compileExceptions(policy); err != nil {
		return err
	}
	_, err := compileMarkdownStrings(policy)
	return err
}

func compileMarkdownStrings(policy Policy) ([]compiledException, error) {
	// Both policies select source literals with the same grammar owners and glob rules.
	selectors := Policy{Exceptions: make([]Exception, 0, len(policy.MarkdownStrings))}
	for _, value := range policy.MarkdownStrings {
		selectors.Exceptions = append(selectors.Exceptions, Exception{
			ID: value.ID, Paths: value.Paths, Formats: value.Formats, Symbols: value.Symbols,
			Kinds: []string{"string"}, Reason: "Markdown selection.",
		})
	}
	compiled, err := compileExceptions(selectors)
	if err != nil {
		return nil, fmt.Errorf("markdown_strings: %w", err)
	}
	return compiled, nil
}

func compileExceptions(policy Policy) ([]compiledException, error) {
	if len(policy.Exceptions) > 100 {
		return nil, fmt.Errorf("extraction allows at most 100 exceptions")
	}
	seen := make(map[string]bool)
	result := make([]compiledException, 0, len(policy.Exceptions))
	for _, value := range policy.Exceptions {
		if err := validateException(value); err != nil {
			return nil, fmt.Errorf("extraction exception %q: %w", value.ID, err)
		}
		if seen[value.ID] {
			return nil, fmt.Errorf("duplicate extraction exception %q", value.ID)
		}
		seen[value.ID] = true
		paths, err := pathglob.Compile(value.Paths)
		if err != nil {
			return nil, fmt.Errorf("extraction exception %q: %w", value.ID, err)
		}
		result = append(result, compiledException{value: value, paths: paths})
	}
	return result, nil
}

func validateException(value Exception) error {
	validID := regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)
	if !validID.MatchString(value.ID) {
		return fmt.Errorf("id must contain lowercase letters, digits, or hyphens and start with a letter")
	}
	if strings.TrimSpace(value.Reason) == "" || len(value.Reason) > 500 {
		return fmt.Errorf("a reason of 1 to 500 bytes is required")
	}
	if len(value.Paths) == 0 || len(value.Paths) > 100 || len(value.Kinds) == 0 {
		return fmt.Errorf("paths (1 to 100) and kinds are required")
	}
	if err := validateExceptionKinds(value); err != nil {
		return err
	}
	return validateExceptionSymbols(value)
}

func validateExceptionKinds(value Exception) error {
	for _, kind := range value.Kinds {
		if kind != "comment" && kind != "string" {
			return fmt.Errorf("unknown extraction kind %q", kind)
		}
	}
	for _, format := range value.Formats {
		if !slices.Contains(document.Formats(), format) || format == document.Plain || format == document.Markdown || format == document.MDX {
			return fmt.Errorf("unknown source format %q", format)
		}
	}
	return nil
}

func validateExceptionSymbols(value Exception) error {
	if len(value.Symbols) > 100 {
		return fmt.Errorf("at most 100 symbols are allowed")
	}
	if len(value.Symbols) > 0 && slices.Contains(value.Kinds, "comment") {
		return fmt.Errorf("symbols can select strings only")
	}
	for _, name := range value.Symbols {
		if strings.TrimSpace(name) == "" || len(name) > 200 {
			return fmt.Errorf("symbols must contain 1 to 200 bytes")
		}
	}
	return validateExceptionPaths(value.Paths)
}

func validateExceptionPaths(paths []string) error {
	for _, name := range paths {
		if name == "" || strings.HasPrefix(name, "/") || slices.Contains(strings.Split(name, "/"), "..") {
			return fmt.Errorf("paths must be nonempty project-relative patterns")
		}
	}
	return nil
}

func (r *sourceReader) exception(kind string, node *ts.Node) string {
	for _, exception := range r.exceptions {
		value := exception.value
		if !slices.Contains(value.Kinds, kind) || !pathglob.Matches(r.doc.Name, exception.paths) {
			continue
		}
		if len(value.Formats) > 0 && !slices.Contains(value.Formats, r.doc.Format) {
			continue
		}
		if len(value.Symbols) > 0 && !r.withinSymbol(node, value.Symbols) {
			continue
		}
		return "config:" + value.ID + ": " + value.Reason
	}
	return ""
}

func (r *sourceReader) withinSymbol(node *ts.Node, symbols []string) bool {
	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		kind := parent.Type(r.syntax.lang)
		if !symbolOwner(kind) {
			continue
		}
		for _, field := range []string{"name", "left", "key", "declarator", "pattern"} {
			if r.matchesSymbol(parent.ChildByFieldName(field, r.syntax.lang), symbols) {
				return true
			}
		}
	}
	return false
}

func symbolOwner(kind string) bool {
	for _, part := range []string{"declaration", "definition", "declarator", "assignment"} {
		if strings.Contains(kind, part) {
			return true
		}
	}
	return slices.Contains([]string{
		"keyed_element", "const_spec", "var_spec", "const_item", "static_item", "function_item", "block_mapping_pair", "flow_pair",
	}, kind)
}

func (r *sourceReader) matchesSymbol(node *ts.Node, symbols []string) bool {
	for depth := 0; node != nil && depth < 128; depth++ {
		span := syntaxSpan(node, 0)
		if slices.Contains(symbols, strings.TrimSpace(string(r.doc.Source[span.Start:span.End]))) {
			return true
		}
		node = node.ChildByFieldName("declarator", r.syntax.lang)
	}
	return false
}
