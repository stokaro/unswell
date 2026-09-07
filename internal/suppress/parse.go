// Package suppress binds explicit source permissions to complete prose units.
package suppress

import (
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/stokaro/unswell/document"
)

// Options contains policy and a shared bound on resolution and matching work.
type Options struct {
	RequireReason bool
	AllowFileWide bool
	RejectUnused  bool
	MaxCandidates int
}

type directive struct {
	kind, reason string
	ids          []string
	span         document.Span
}

func parse(raw document.Directive, catalog map[string]bool, options Options) (directive, error) {
	d := directive{span: raw.Span}
	text := strings.TrimSpace(raw.Text)
	if len(text) > 4096 || strings.ContainsAny(text, "\r\n\x00") {
		return d, fmt.Errorf("suppression must be one logical line of at most 4096 bytes")
	}
	command, reason, hasReason := strings.Cut(text, " -- ")
	var err error
	d.kind, d.ids, err = parseCommand(command, catalog)
	if err != nil {
		return d, err
	}
	d.reason = strings.TrimSpace(reason)
	return d, validateReason(d, hasReason, options)
}

func parseCommand(command string, catalog map[string]bool) (string, []string, error) {
	fields := strings.Fields(command)
	if len(fields) != 2 {
		return "", nil, fmt.Errorf("expected a suppression command, explicit rule IDs, and -- reason")
	}
	kind := strings.TrimPrefix(fields[0], "unswell-")
	if !strings.HasPrefix(fields[0], "unswell-") || !slices.Contains([]string{
		"disable-next-sentence", "disable-next-block", "disable", "enable", "disable-file",
	}, kind) {
		return "", nil, fmt.Errorf("unknown suppression command %q", fields[0])
	}
	ids, err := ruleIDs(fields[1], catalog)
	return kind, ids, err
}

func validateReason(d directive, hasReason bool, options Options) error {
	if d.kind == "enable" {
		if hasReason {
			return fmt.Errorf("enable uses the opening directive's reason")
		}
		return nil
	}
	if options.RequireReason && !substantiveReason(d.reason) {
		return fmt.Errorf("suppression requires -- and a reason with at least two words")
	}
	if d.kind == "disable-file" && !options.AllowFileWide {
		return fmt.Errorf("file-wide suppression requires suppressions.allow_file_wide")
	}
	return nil
}

func ruleIDs(value string, catalog map[string]bool) ([]string, error) {
	ids := strings.Split(value, ",")
	if len(ids) > 32 {
		return nil, fmt.Errorf("suppression exceeds 32 rule IDs")
	}
	seen := make(map[string]bool)
	for _, id := range ids {
		if !catalog[id] || strings.HasPrefix(id, "gate.") {
			return nil, fmt.Errorf("unknown or unsupported suppression rule ID %q", id)
		}
		if seen[id] {
			return nil, fmt.Errorf("duplicate suppression rule ID %q", id)
		}
		seen[id] = true
	}
	slices.Sort(ids)
	return ids, nil
}

func substantiveReason(reason string) bool {
	words := 0
	for word := range strings.FieldsSeq(reason) {
		if strings.ContainsFunc(word, unicode.IsLetter) {
			words++
		}
	}
	return words >= 2
}
