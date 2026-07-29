package tsvt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUnquote_StripsExactlyTheEnclosingQuotes pins what the doc claims the
// STRING token guarantees: the outer pair goes, and nothing inside is touched.
// A naive strings.Trim would eat a legitimate leading or trailing quote from
// the content, and an escape-aware unquote would rewrite the body — neither is
// what this does, and a string literal that changed under parsing would make
// `="he said ""hi"""`-style content unrepresentable.
func TestUnquote_StripsExactlyTheEnclosingQuotes(t *testing.T) {
	t.Parallel()

	cases := map[quoted]string{
		`""`:           "",
		`"a"`:          "a",
		`"a b"`:        "a b",
		`"say ""hi"""`: `say ""hi""`,
		`""""`:         `""`,
		`"trailing "`:  "trailing ",
		`"  padded  "`: "  padded  ",
		`"tab	inside"`: "tab\tinside",
	}
	for in, want := range cases {
		assert.Equal(t, want, unquote(in), string(in))
	}
}
