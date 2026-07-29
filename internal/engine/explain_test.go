package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

func TestExplain_Formula(t *testing.T) {
	t.Parallel()

	// C1 = A1 + B1 over 2 and 3.
	trace, err := engine.Explain(parse(t, "2\t3\t=A1 + B1\n"), engine.Address{Row: 0, Col: 2})
	require.NoError(t, err)
	assert.Equal(t, "C1", trace.Cell)
	assert.Equal(t, "5", trace.Value)
	assert.Equal(t, "A1 + B1", trace.Formula)
	assert.Equal(t, []engine.TraceInput{{Ref: "A1", Value: "2"}, {Ref: "B1", Value: "3"}}, trace.Inputs)
}

func TestExplain_RangeInput(t *testing.T) {
	t.Parallel()

	// A range operand renders as a two-cell A1 range whose value lists the
	// range's cells — not the #VALUE! that scalar reduction would yield.
	trace, err := engine.Explain(parse(t, "1\t2\t=sum(A1:B1)\n"), engine.Address{Row: 0, Col: 2})
	require.NoError(t, err)
	require.Len(t, trace.Inputs, 1)
	assert.Equal(t, "A1:B1", trace.Inputs[0].Ref)
	assert.Equal(t, "1, 2", trace.Inputs[0].Value)
}

func TestExplain_Literal(t *testing.T) {
	t.Parallel()

	trace, err := engine.Explain(parse(t, "hello\t=A1\n"), engine.Address{Row: 0, Col: 0})
	require.NoError(t, err)
	assert.Equal(t, "hello", trace.Value)
	assert.Empty(t, trace.Formula)
	assert.Empty(t, trace.Inputs)
}

func TestExplain_OutOfGrid(t *testing.T) {
	t.Parallel()

	_, err := engine.Explain(parse(t, "1\t2\n"), engine.Address{Row: 9, Col: 9})
	require.Error(t, err)
	assert.ErrorIs(t, err, constants.ErrNotFound)
}

func TestExplain_RendersEveryExpressionForm(t *testing.T) {
	t.Parallel()

	// Each formula exercises one renderExpr branch; the rendered form round-trips.
	cases := map[string]string{
		"=42":              "42",
		`="hi"`:            `"hi"`,
		"=TRUE":            "TRUE",
		"=FALSE":           "FALSE",
		"=#N/A":            "#N/A",
		"=-A1":             "-A1",
		"=A1%":             "A1%",
		"=A1 + 1":          "A1 + 1",
		"=abs(A1)":         "abs(A1)",
		"=pi()":            "pi",              // a nullary call canonicalizes to bare
		"=pi":              "pi",              // and the bare form round-trips
		"=now()":           "now",             // holds for any zero-argument call
		`="other.tsvt"!A1`: `"other.tsvt"!A1`, // cross-sheet single cell
		`="d.tsvt"!A1:B2`:  `"d.tsvt"!A1:B2`,  // cross-sheet range
		// The pipe spelling is preserved (§5.4): a piped call renders as the
		// author's pipe, a chain stays a chain, and an operator capturing a
		// piped call parenthesizes it. A stage with no explicit arguments
		// renders bare — the canonical form drops its empty parentheses.
		"=A1 | len()":            "A1 | len",
		"=A1 | len":              "A1 | len", // bare stage round-trips unchanged
		"=A1 | round(2)":         "A1 | round(2)",
		"=A1 | trim() | len()":   "A1 | trim | len",
		"=(A1 | len()) + 1":      "(A1 | len) + 1",
		"=sum(A1 | round(2), 1)": "sum(A1 | round(2),1)", // piped call as an argument needs no parens
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			t.Parallel()
			trace, err := engine.Explain(parse(t, "5\t"+src+"\n"), engine.Address{Row: 0, Col: 1})
			require.NoError(t, err)
			assert.Equal(t, want, trace.Formula)
		})
	}
}
