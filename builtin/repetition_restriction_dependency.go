package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

// The possessive dependency role is explicit: OWNER's dependencies and OWNER
// depends on code. No other noun/verb synonym or omitted scope is inferred.
func dependencyRestriction(tokens []document.Token) (restriction, bool) {
	if len(tokens) > 3 && matches(tokens[:3], []string{"in", "particular", ","}) {
		tokens = tokens[3:]
	}
	actor, rest := dependencyOwner(tokens)
	if !restrictionActor(actor) {
		return restriction{}, false
	}
	if len(rest) > 5 && matches(rest[:5], []string{"(", "direct", "and", "transitive", ")"}) {
		rest = rest[5:]
	}
	tense, start := dependencyRestrictionPredicate(rest)
	if start == 0 {
		return restriction{}, false
	}
	property, object, ok := qualifiedRestrictionObject(rest[start:])
	if !ok || object != "licenses" {
		return restriction{}, false
	}
	return restriction{actor, document.Token{Normal: "depend", Tag: "VB"}, property, "licenses", tense}, true
}

func dependencyOwner(tokens []document.Token) ([]document.Token, []document.Token) {
	for i := 1; i+2 < len(tokens) && i < 8; i++ {
		if frameWord(tokens[i], "'s", "’s") && frameWord(tokens[i+1], "dependencies") {
			return tokens[:i], tokens[i+2:]
		}
	}
	return nil, nil
}

func dependencyRestrictionPredicate(tokens []document.Token) (string, int) {
	if len(tokens) > 5 && matches(tokens[:5], []string{"will", "always", "be", "limited", "to"}) {
		return "will", 5
	}
	if len(tokens) > 3 && matches(tokens[:3], []string{"are", "limited", "to"}) {
		return "present", 3
	}
	return "", 0
}

func licenseProperty(tokens []document.Token) ([]document.Token, string, bool) {
	if len(tokens) != 2 || !frameWord(tokens[1], "licensed") || !strings.HasSuffix(tokens[0].Normal, "ly") {
		return nil, "", false
	}
	// Licensed code and its licenses name the same property here. Retain the
	// complete adjective; do not equate permissive with open or free licenses.
	property := tokens[0]
	property.Normal = strings.TrimSuffix(property.Normal, "ly")
	property.Tag = "JJ"
	return []document.Token{property}, "licenses", true
}
