package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// checkDiags parses src and returns its Check findings.
func checkDiags(t *testing.T, src string) []engine.Diagnostic {
	t.Helper()
	s, err := engine.Parse([]byte(src))
	require.NoError(t, err)
	return engine.Check(s)
}

// onlyDiag requires exactly one finding and returns it: the verdict tests each
// build a sheet with one defect, so a second finding is itself a finding.
func onlyDiag(t *testing.T, src string) engine.Diagnostic {
	t.Helper()
	diags := checkDiags(t, src)
	require.Len(t, diags, 1)
	return diags[0]
}

// TestCheckNames_CleanSheetIsClean states the baseline: the canonical named
// sheet — every name bound once, used, and descriptive — reports nothing.
func TestCheckNames_CleanSheetIsClean(t *testing.T) {
	t.Parallel()

	assert.Empty(t, checkDiags(t, "=0.08 |@ named(Rate)\t=@Rate * 100\n"))
}

// TestCheckNames_DuplicateBindingIsFatal is the verdict table's first row: the
// second binding is the error, at its own cell, citing where the name was
// first bound — and name identity is case-insensitive, so a case-variant
// duplicate is still a duplicate.
func TestCheckNames_DuplicateBindingIsFatal(t *testing.T) {
	t.Parallel()

	d := onlyDiag(t, "=1 |@ named(X)\t=2 |@ named(x)\t=@X + A1\n")
	assert.Equal(t, "B1", d.Cell)
	assert.True(t, d.IsFatal)
	assert.Equal(t, "name x is already bound at A1", d.Message)
}

// TestCheckNames_NamedArityIsFatal states the arity verdicts: the name
// argument is required — never optional — and singular. Both the bare clause a
// dropped-parentheses stage permits and a surplus argument are the same
// finding.
func TestCheckNames_NamedArityIsFatal(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"bare clause":       "=1 |@ named\n",
		"empty parentheses": "=1 |@ named()\n",
		"surplus arguments": "=1 |@ named(a, b)\n",
	}
	for name, src := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			d := onlyDiag(t, src)
			assert.True(t, d.IsFatal)
			assert.Equal(t, "named takes exactly one argument, the name", d.Message)
			assert.Equal(t, "A1", d.Cell)
		})
	}
}

// TestCheckNames_UnknownMetaFunctionIsFatal covers both misspellings the
// grammar admits: a wrong function in the clause, and the `|@Total` shape
// where the intended NAME lands in the function position.
func TestCheckNames_UnknownMetaFunctionIsFatal(t *testing.T) {
	t.Parallel()
	tests := map[string]struct{ src, message string }{
		"wrong function": {
			src:     "=1 |@ label(X)\n",
			message: "unknown meta function: label (the one meta function is named)",
		},
		"name in the function position": {
			src:     "=1 |@Total\n",
			message: "unknown meta function: Total (the one meta function is named)",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			d := onlyDiag(t, tc.src)
			assert.True(t, d.IsFatal)
			assert.Equal(t, tc.message, d.Message)
		})
	}
}

// TestCheckNames_MisplacedNamedIsFatal states the other side of the meta
// boundary: `named` written as an ordinary call or pipe stage is refused with
// the spelling that was wanted — never reported as an unknown function, which
// would send its author hunting the catalog for a function that exists.
func TestCheckNames_MisplacedNamedIsFatal(t *testing.T) {
	t.Parallel()
	for name, src := range map[string]string{
		"ordinary call": "=named(1, 2)\n",
		"pipe stage":    "=1 | named\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			d := onlyDiag(t, src)
			assert.True(t, d.IsFatal)
			assert.Contains(t, d.Message, "named is a meta function")
			assert.NotContains(t, d.Message, "unknown function")
		})
	}
}

// TestCheckNames_UnknownNameIsAdvisory states the unbound-reference finding:
// the same class as an unknown function — it computes #NAME? — and reported
// the same way.
func TestCheckNames_UnknownNameIsAdvisory(t *testing.T) {
	t.Parallel()

	d := onlyDiag(t, "=@Nowhere\n")
	assert.False(t, d.IsFatal)
	assert.Equal(t, "unknown name: @Nowhere", d.Message)
}

// TestCheckNames_UnusedNameIsAdvisory states the unused-binding advisory, at
// the binding cell.
func TestCheckNames_UnusedNameIsAdvisory(t *testing.T) {
	t.Parallel()

	d := onlyDiag(t, "=1 |@ named(Lonely)\n")
	assert.False(t, d.IsFatal)
	assert.Equal(t, "A1", d.Cell)
	assert.Equal(t, "name Lonely is bound but never used", d.Message)
}

// TestCheckNames_CellShapedNameIsAdvisory states the 023 amendment: a name
// spelled like a canonical A1 reference is legal — the sigil disambiguates —
// but earns an advisory, while `rate2`, whose lowercase spelling collides with
// nothing, deliberately does not.
func TestCheckNames_CellShapedNameIsAdvisory(t *testing.T) {
	t.Parallel()

	diags := checkDiags(t, "=1 |@ named(A1)\t=@A1\n")
	require.Len(t, diags, 1)
	assert.False(t, diags[0].IsFatal)
	assert.Equal(t, "name A1 reads as a cell reference; consider a descriptive name", diags[0].Message)

	assert.Empty(t, checkDiags(t, "=1 |@ named(rate2)\t=@rate2\n"),
		"a lowercase spelling collides with nothing and is clean")
}

// TestCheckNames_DuplicateSuppressesUnknownNameNoise states diagnostic
// discipline: uses of a duplicated name compute #NAME?, but the finding is the
// duplicate itself — flagging every use as an unknown name would bury the one
// defect under its symptoms.
func TestCheckNames_DuplicateSuppressesUnknownNameNoise(t *testing.T) {
	t.Parallel()

	diags := checkDiags(t, "=1 |@ named(X)\t=2 |@ named(X)\t=@X\n")
	require.Len(t, diags, 1)
	assert.Equal(t, "name X is already bound at A1", diags[0].Message)
}

// TestCheckNames_UnusedAdvisoriesAreOrderedByBindingCell states determinism:
// several unused names report in binding order — row-major, columns breaking
// row ties — never in a map's per-run shuffle.
func TestCheckNames_UnusedAdvisoriesAreOrderedByBindingCell(t *testing.T) {
	t.Parallel()

	src := "=1 |@ named(bee)\t=2 |@ named(cee)\n=3 |@ named(aye)\n"
	diags := checkDiags(t, src)
	require.Len(t, diags, 3)
	assert.Equal(t, "A1", diags[0].Cell)
	assert.Equal(t, "B1", diags[1].Cell, "same row: the column breaks the tie")
	assert.Equal(t, "A2", diags[2].Cell)
}
