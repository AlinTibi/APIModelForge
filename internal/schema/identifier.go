package schema

import (
	"strings"
	"unicode"
)

// PascalCase converts an arbitrary JSON field/type name (snake_case,
// kebab-case, space separated, already camelCase, ...) into a PascalCase
// Go-style identifier. It is used for generated type names, which must be
// valid and idiomatic across every supported target language.
func PascalCase(s string) string {
	s = strings.Map(func(r rune) rune {
		if r > 127 {
			return '_'
		}
		return r
	}, s)
	words := splitWords(s)
	var b strings.Builder
	for _, w := range words {
		if w == "" {
			continue
		}
		runes := []rune(w)
		b.WriteRune(unicode.ToUpper(runes[0]))
		for _, r := range runes[1:] {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	out := b.String()
	if out == "" {
		return "Model"
	}
	if !unicode.IsLetter(rune(out[0])) {
		out = "N" + out
	}
	return out
}

// splitWords breaks an identifier-ish string into words at separators
// (anything that isn't a letter or digit) and at lower-to-upper case
// boundaries (camelCase).
func splitWords(s string) []string {
	var words []string
	var cur strings.Builder
	runes := []rune(s)

	flush := func() {
		if cur.Len() > 0 {
			words = append(words, cur.String())
			cur.Reset()
		}
	}

	for i, r := range runes {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if i > 0 && unicode.IsUpper(r) && unicode.IsLower(runes[i-1]) {
				flush()
			}
			cur.WriteRune(r)
		default:
			flush()
		}
	}
	flush()
	return words
}

// singularize applies a light heuristic to turn a plural-looking name into
// a singular one, used when naming the element type of an array (e.g. a
// field called "items" should name its element type "Item", not "Items").
// It intentionally only handles common, low-risk suffixes.
func singularize(s string) string {
	lower := strings.ToLower(s)
	switch {
	case strings.HasSuffix(lower, "ies") && len(s) > 3:
		return s[:len(s)-3] + "y"
	case strings.HasSuffix(lower, "ses") || strings.HasSuffix(lower, "xes") ||
		strings.HasSuffix(lower, "ches") || strings.HasSuffix(lower, "shes"):
		return s[:len(s)-2]
	case strings.HasSuffix(lower, "s") && !strings.HasSuffix(lower, "ss") && len(s) > 1:
		return s[:len(s)-1]
	default:
		return s
	}
}

// IsValidBareIdentifier reports whether s can be used as-is as a property
// identifier in TypeScript/Kotlin/Python-like languages: non-empty,
// letters/digits/underscore only, and not starting with a digit.
func IsValidBareIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if r == '_' || unicode.IsLetter(r) {
			continue
		}
		if unicode.IsDigit(r) && i > 0 {
			continue
		}
		return false
	}
	return true
}

// SanitizeIdentifier produces a valid bare identifier from an arbitrary
// JSON field name by replacing invalid characters with underscores and
// prefixing with "_" if the result would start with a digit or be empty.
func SanitizeIdentifier(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	out := b.String()
	if strings.Trim(out, "_") == "" {
		return "field"
	}
	if out[0] >= '0' && out[0] <= '9' {
		out = "_" + out
	}
	return out
}
