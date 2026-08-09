package tsvt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
)

func TestParseFormula_Expression(t *testing.T) {
	t.Parallel()

	// A formula is the expression sublanguage: references, operators, and calls
	// compose into the typed Expr AST.
	expr, _, err := ParseFormula("B2 + C2")
	require.NoError(t, err)
	binary, ok := expr.(Binary)
	require.True(t, ok, "expected Binary, got %T", expr)
	assert.Equal(t, OpAdd, binary.Op)
}

func TestParseFormula_Reference(t *testing.T) {
	t.Parallel()

	// A bare reference is a valid formula: an operand with no operator.
	expr, _, err := ParseFormula("D2:D4")
	require.NoError(t, err)
	_, ok := expr.(RefOperand)
	assert.True(t, ok, "expected RefOperand, got %T", expr)
}

func TestParseFormula_SyntaxError(t *testing.T) {
	t.Parallel()

	_, _, err := ParseFormula("sum(")
	require.Error(t, err)
	assert.ErrorIs(t, err, constants.ErrSyntax)
}

func TestParseFormula_TrailingInput(t *testing.T) {
	t.Parallel()

	// A complete expression followed by more tokens is rejected — the whole
	// cell must be one formula, not a formula plus leftovers.
	_, _, err := ParseFormula("1 2")
	require.Error(t, err)
	assert.ErrorIs(t, err, constants.ErrSyntax)
}
