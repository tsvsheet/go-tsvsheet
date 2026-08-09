package tsvt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
)

// parse builds the AST of a formula, failing the test on a syntax error.
func parse(t *testing.T, src string) Expr {
	t.Helper()
	e, _, err := ParseFormula(FormulaText(src))
	require.NoError(t, err)
	return e
}

func TestBuild_Number(t *testing.T) {
	t.Parallel()
	assert.Equal(t, Number{Text: "42"}, parse(t, "42"))
	assert.Equal(t, Number{Text: "3.14"}, parse(t, "3.14"))
}

func TestBuild_String(t *testing.T) {
	t.Parallel()
	assert.Equal(t, StringLit{Value: "hi"}, parse(t, `"hi"`))
}

func TestBuild_Bool(t *testing.T) {
	t.Parallel()
	assert.Equal(t, BoolLit{IsTrue: true}, parse(t, "TRUE"))
	assert.Equal(t, BoolLit{IsTrue: false}, parse(t, "FALSE"))
}

func TestBuild_Error(t *testing.T) {
	t.Parallel()
	assert.Equal(t, ErrorLit{Code: "#N/A"}, parse(t, "#N/A"))
	assert.Equal(t, ErrorLit{Code: "#REF!"}, parse(t, "#REF!"))
}

func TestBuild_Reference(t *testing.T) {
	t.Parallel()
	cell := func(isColAbsolute, isRowAbsolute bool) RefOperand {
		return RefOperand{Ref: RangeRef{
			From: CellRef{Col: "B", Row: 2, IsColAbsolute: isColAbsolute, IsRowAbsolute: isRowAbsolute},
		}}
	}
	assert.Equal(t, cell(false, false), parse(t, "B2"))
	// $-absolute markers are retained (SPECIFICATION §4), each pinning its own
	// coordinate; they carry no positional difference.
	assert.Equal(t, cell(true, true), parse(t, "$B$2"))
	assert.Equal(t, cell(false, true), parse(t, "B$2"))
	assert.Equal(t, cell(true, false), parse(t, "$B2"))
}

func TestBuild_Range(t *testing.T) {
	t.Parallel()
	to := CellRef{Col: "C", Row: 3}
	want := RefOperand{Ref: RangeRef{From: CellRef{Col: "A", Row: 1}, To: &to}}
	assert.Equal(t, want, parse(t, "A1:C3"))
}

func TestBuild_Unary(t *testing.T) {
	t.Parallel()
	assert.Equal(t, Unary{X: Number{Text: "5"}, Op: OpNeg}, parse(t, "-5"))
	assert.Equal(t, Unary{X: Number{Text: "5"}, Op: OpPos}, parse(t, "+5"))
}

func TestBuild_Percent(t *testing.T) {
	t.Parallel()
	assert.Equal(t, Percent{X: Number{Text: "50"}}, parse(t, "50%"))
}

func TestBuild_BinaryOperators(t *testing.T) {
	t.Parallel()
	cases := map[string]BinaryOp{
		"2 ^ 8":     OpPow,
		"1 * 2":     OpMul,
		"1 / 2":     OpDiv,
		"1 + 2":     OpAdd,
		"1 - 2":     OpSub,
		`"a" & "b"`: OpCat,
		"1 = 2":     OpEq,
		"1 <> 2":    OpNe,
		"1 < 2":     OpLt,
		"1 <= 2":    OpLe,
		"1 > 2":     OpGt,
		"1 >= 2":    OpGe,
	}
	for src, op := range cases {
		t.Run(src, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, op, parse(t, src).(Binary).Op)
		})
	}
}

func TestBuild_Precedence(t *testing.T) {
	t.Parallel()
	// (1 + 2) * 3 groups the addition first.
	outer := parse(t, "(1 + 2) * 3").(Binary)
	assert.Equal(t, OpMul, outer.Op)
	assert.Equal(t, OpAdd, outer.Left.(Binary).Op)
}

func TestBuild_Call(t *testing.T) {
	t.Parallel()
	multi := parse(t, "sum(A1, B1)").(Call)
	assert.Equal(t, "sum", multi.Name)
	assert.Len(t, multi.Args, 2)

	assert.Equal(t, "IF", parse(t, "IF(1, 2, 3)").(Call).Name)    // name via COL
	assert.Empty(t, parse(t, "now()").(Call).Args)                // no arguments
	assert.Equal(t, "atan2", parse(t, "atan2(1, 1)").(Call).Name) // trailing digits folded in
	assert.Equal(t, "log10", parse(t, "log10(100)").(Call).Name)
}

func TestBuild_Pipe(t *testing.T) {
	t.Parallel()
	// `x | f(a…)` desugars to `f(x, a…)` with the spelling flag set (§5.4).
	want := Call{
		Name: "round",
		Args: []Expr{
			RefOperand{Ref: RangeRef{From: CellRef{Col: "A", Row: 1}}},
			Number{Text: "2"},
		},
		IsPiped: true,
	}
	assert.Equal(t, want, parse(t, "A1 | round(2)"))
}

func TestBuild_PipeIsTheComposedCall(t *testing.T) {
	t.Parallel()
	// The two spellings are the same formula: identical calls but for the flag.
	piped := parse(t, "A1 | round(2)").(Call)
	piped.IsPiped = false
	assert.Equal(t, parse(t, "round(A1, 2)"), piped)
}

func TestBuild_PipeChainFoldsLeft(t *testing.T) {
	t.Parallel()
	// x | f() | g() ≡ g(f(x)).
	outer := parse(t, "A1 | trim() | len()").(Call)
	assert.Equal(t, "len", outer.Name)
	require.Len(t, outer.Args, 1)
	inner := outer.Args[0].(Call)
	assert.Equal(t, "trim", inner.Name)
	assert.True(t, inner.IsPiped)
}

func TestBuild_PipeBindsLoosest(t *testing.T) {
	t.Parallel()
	// The entire preceding expression is the piped value: A1 & B1 | len() ≡ len(A1 & B1).
	call := parse(t, "A1 & B1 | len()").(Call)
	assert.Equal(t, "len", call.Name)
	require.Len(t, call.Args, 1)
	assert.Equal(t, OpCat, call.Args[0].(Binary).Op)
}

func TestBuild_PipeSyntaxErrors(t *testing.T) {
	t.Parallel()
	// The right-hand side must be a §5.3 call (bare or parenthesized); a missing
	// stage, a non-call stage, a parenthesized expression, or a leading pipe is a
	// syntax error by construction. A bare name (`A1 | len`) is now a valid
	// zero-argument stage and is covered in the build tests, not here.
	for _, src := range []string{"A1 |", "A1 | 5", "A1 | (len())", "| len()"} {
		t.Run(src, func(t *testing.T) {
			t.Parallel()
			_, _, err := ParseFormula(FormulaText(src))
			require.Error(t, err)
			assert.ErrorIs(t, err, constants.ErrSyntax)
		})
	}
}

func TestBuild_PipeOperandErrors(t *testing.T) {
	t.Parallel()
	// A bad reference surfaces through both pipe builder paths: the piped
	// value and a stage's own argument.
	for _, src := range []string{"B2.5 | len()", "A1 | round(B2.5)"} {
		t.Run(src, func(t *testing.T) {
			t.Parallel()
			_, _, err := ParseFormula(FormulaText(src))
			require.Error(t, err)
			assert.ErrorIs(t, err, constants.ErrSyntax)
		})
	}
}

func TestBuild_FractionalRowRejected(t *testing.T) {
	t.Parallel()
	// A fractional A1 row is a syntax error; assert it surfaces through every
	// builder path that can contain a reference.
	for _, src := range []string{"B2.5", "-B2.5", "B2.5%", "B2.5 + 1", "1 + B2.5", "sum(B2.5)", "A1:C3.5"} {
		t.Run(src, func(t *testing.T) {
			t.Parallel()
			_, _, err := ParseFormula(FormulaText(src))
			require.Error(t, err)
			assert.ErrorIs(t, err, constants.ErrSyntax)
		})
	}
}

// TestBuildPipe_EvaluationNeverSeesAPipe pins the desugar's whole point: the
// pipe is not a node the evaluator must know about, it IS the call. If a Pipe
// survived into the AST, every consumer — evaluator, checker, renderer — would
// need a case for it, and the "sugar over §5.3" claim would be false.
func TestBuildPipe_EvaluationNeverSeesAPipe(t *testing.T) {
	t.Parallel()

	expr, _, err := ParseFormula(FormulaText("x | f(a)"))
	require.NoError(t, err)

	call, isCall := expr.(Call)
	require.True(t, isCall, "the pipe built a Call, not a pipe node")
	assert.Equal(t, "f", call.Name)
	assert.True(t, call.IsPiped, "the source spelling is recorded for rendering")
	require.Len(t, call.Args, 2, "the piped value is prepended to the call's arguments")

	// And the invariant Call's own doc states: a piped call always has at least
	// one argument, because the desugar prepends one.
	bare, _, err := ParseFormula(FormulaText("x | f"))
	require.NoError(t, err)
	if piped, ok := bare.(Call); ok && piped.IsPiped {
		assert.NotEmpty(t, piped.Args, "a piped call is never argument-less")
	}
}
