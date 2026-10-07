package schema

import (
	"encoding/json"
	"strconv"
	"strings"
)

// UniqueIdentifier allocates a deterministic name within a generated scope.
func UniqueIdentifier(base string, used map[string]bool) string {
	name := base
	for suffix := 2; used[name]; suffix++ {
		name = base + strconv.Itoa(suffix)
	}
	used[name] = true
	return name
}

// QuoteString uses JSON-compatible escapes supported by C#, TypeScript and
// safe single-line source comments. It never emits literal control characters.
func QuoteString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// Model type names must not shadow the runtime types or imports emitted by
// any renderer. Type names are shared across all five output targets.
func reservedTypeNames() map[string]*Node {
	result := map[string]*Node{}
	for _, name := range strings.Fields("System List String Long Double Boolean Any Object Optional ModelCount JsonPropertyName JsonPropertyNameAttribute Dataclass True False None") {
		result[name] = nil
	}
	return result
}
