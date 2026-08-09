package tsvt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
)

// parseMeta parses src and returns its meta clause, requiring a clean parse.
func parseMeta(t *testing.T, src string) MetaClause {
	t.Helper()
	_, meta, err := ParseFormula(FormulaText(src))
	require.NoError(t, err)
	return meta
}

// TestParseFormulaMetaClause states the clause's parsed shape (SPECIFICATION
// §5.6): the function name and argument case-preserved, the author's
// parenthesization recorded, and the absent clause the zero value.
func TestParseFormulaMetaClause(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		src  string
		want MetaClause
	}{
		"canonical declaration": {
			src:  `sum(C3:C4) |@ named(Total)`,
			want: MetaClause{Fn: "named", Args: []NameText{"Total"}, HasParens: true},
		},
		"wider charset than a function name": {
			src:  `1 |@ named(q1_revenue)`,
			want: MetaClause{Fn: "named", Args: []NameText{"q1_revenue"}, HasParens: true},
		},
		"case preserved for rendering": {
			src:  `1 |@ NAMED(TOTAL)`,
			want: MetaClause{Fn: "NAMED", Args: []NameText{"TOTAL"}, HasParens: true},
		},
		"bare clause: dropped parentheses, no arguments": {
			src:  `1 |@ named`,
			want: MetaClause{Fn: "named", HasParens: false},
		},
		"empty parentheses parse; arity is semantic": {
			src:  `1 |@ named()`,
			want: MetaClause{Fn: "named", HasParens: true},
		},
		"surplus arguments parse; arity is semantic": {
			src:  `1 |@ named(a, b)`,
			want: MetaClause{Fn: "named", Args: []NameText{"a", "b"}, HasParens: true},
		},
		"no space after the marker": {
			src:  `1 |@named(T)`,
			want: MetaClause{Fn: "named", Args: []NameText{"T"}, HasParens: true},
		},
		"a bare name reads as the clause's function": {
			src:  `1 |@Total`,
			want: MetaClause{Fn: "Total", HasParens: false},
		},
		"after a pipe chain": {
			src:  `sum(A1:A5) | round(2) |@ named(Total)`,
			want: MetaClause{Fn: "named", Args: []NameText{"Total"}, HasParens: true},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, parseMeta(t, tc.src))
		})
	}
}

// TestParseFormulaWithoutMetaIsZero states the absent clause: every ordinary
// formula parses with the zero MetaClause, so no caller needs a second signal.
func TestParseFormulaWithoutMetaIsZero(t *testing.T) {
	t.Parallel()

	meta := parseMeta(t, `sum(A1:B2) + 1`)
	assert.True(t, meta.IsZero())
}

// TestParseFormulaMetaUnwritableShapes states what the grammar makes
// unwritable rather than merely invalid (SPECIFICATION §5.6): no expression of
// any kind can occupy the clause, and the clause cannot nest, repeat, sit
// mid-pipeline, or precede more expression.
func TestParseFormulaMetaUnwritableShapes(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"a number is not a name":          `1 |@ named(1)`,
		"a string is not a name":          `1 |@ named("x")`,
		"indirection cannot be written":   `1 |@ named(@Other)`,
		"a computation cannot be written": `1 |@ named(concat(a, b))`,
		"two identifiers need a comma":    `1 |@ named(a b)`,
		"the clause cannot nest":          `sum(A1:A5 @named(T))`,
		"the clause cannot sit mid-pipe":  `A1 | @named(T)`,
		"the clause cannot repeat":        `1 |@ named(T) |@ named(U)`,
		"the clause must be last":         `1 |@ named(T) | round(2)`,
		"the clause cannot be grouped":    `(1 |@ named(T))`,
		"the marker needs a function":     `1 |@`,
		"a name is not a range endpoint":  `@Total:B2`,
	}
	for name, src := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, _, err := ParseFormula(FormulaText(src))
			require.ErrorIs(t, err, constants.ErrSyntax)
		})
	}
}

// TestParseFormulaNameRef states the `@name` operand: the sigil is stripped,
// the spelling is case-preserved, and the reference composes as an ordinary
// operand.
func TestParseFormulaNameRef(t *testing.T) {
	t.Parallel()

	expr, meta, err := ParseFormula(`@Total * @rate_2`)
	require.NoError(t, err)
	assert.True(t, meta.IsZero())

	bin, ok := expr.(Binary)
	require.True(t, ok)
	assert.Equal(t, NameRef{Name: "Total"}, bin.Left)
	assert.Equal(t, NameRef{Name: "rate_2"}, bin.Right)
}

// TestParseFormulaNameRefAsArgument states that a name reference is an
// ordinary operand inside a call's argument list.
func TestParseFormulaNameRefAsArgument(t *testing.T) {
	t.Parallel()

	expr, _, err := ParseFormula(`sum(@Total, 1)`)
	require.NoError(t, err)
	call, ok := expr.(Call)
	require.True(t, ok)
	require.Len(t, call.Args, 2)
	assert.Equal(t, NameRef{Name: "Total"}, call.Args[0])
}
