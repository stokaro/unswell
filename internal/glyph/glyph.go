// Package glyph classifies single characters that rules and page-level
// measurements treat alike.
package glyph

// Pictograph reports whether a rune is a decorative emoji or pictographic
// symbol: the emoji blocks, the miscellaneous symbols and dingbats, the
// regional indicators, and a few stars and arrows that render as emoji.
// Check marks, crosses, and ballot boxes are left out: people mark ordinary
// checklists with them, and people who remove AI-sounding wording leave them
// in place about as often as any other word.
func Pictograph(r rune) bool {
	switch r {
	case 0x2705, 0x2714, 0x2713, 0x2611, 0x274C, 0x2717, 0x2718, 0x274E, 0x2610:
		return false
	case 0x2B50, 0x2B55, 0x2B06, 0x2B07:
		return true
	}
	return (r >= 0x1F300 && r <= 0x1FAFF) || (r >= 0x1F1E6 && r <= 0x1F1FF) || (r >= 0x2600 && r <= 0x27BF)
}
