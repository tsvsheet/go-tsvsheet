package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBareNullaryAndPipe covers the drop-the-empty-parens affordance: a
// zero-argument call needs no parentheses, as a standalone value or as a pipe
// stage, and a bare column+row stays a cell reference rather than a call.
func TestBareNullaryAndPipe(t *testing.T) {
	t.Parallel()

	cases := map[string]struct{ expr, want string }{
		"bare nullary computes":        {"round(pi, 5)", "3.14159"},
		"bare nullary alone":           {"round(pi(), 5)", "3.14159"}, // parens still accepted
		"bare pipe stage computes":     {"A1 | len", "1"},             // len of A1's text "2"
		"parenthesized pipe unchanged": {"A1 | len()", "1"},
		"bare pipe chains":             {"A1 | trim | len", "1"},
		"A1 stays a reference":         {"A1", "2"}, // not a call named "A"
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, formula1(t, tc.expr))
		})
	}
}

// §5.4: the pipe is sugar for a function call — both spellings are the same
// computation, chains fold left, error values propagate as ordinary arguments,
// and an unknown stage is the usual #NAME?.
func TestPipe_EvaluatesAsTheComposedCall(t *testing.T) {
	t.Parallel()
	// A1=3.4 B1=2; the formula computes in C1.
	cases := map[string]string{
		"=A1 | round(0)":         "3",      // one stage with an argument
		"=round(A1, 0)":          "3",      // …identical to its composed spelling
		"=A1:B1 | sum()":         "5.4",    // a range pipes as the first argument
		"=A1 | round(0) | len()": "1",      // chains fold left: len(round(A1, 0))
		"=A1 & B1 | len()":       "4",      // pipe binds loosest: len(A1 & B1) = len("3.42")
		"=#N/A | len()":          "#N/A",   // an error value propagates through the pipe
		"=A1 | bogus()":          "#NAME?", // an unknown stage is an unknown function
	}
	for formula, want := range cases {
		t.Run(formula, func(t *testing.T) {
			t.Parallel()
			sheet := parse(t, "3.4\t2\t"+formula+"\n")
			assert.Equal(t, want, cellAt(t, sheet.Compute(), 0, 2))
		})
	}
}

// A structural edit re-renders formulas from the AST; a formula written with
// the pipe operator keeps its pipe spelling (§5.4) — it is never silently
// rewritten into the composed call — and still recomputes correctly.
func TestPipe_SpellingSurvivesStructuralEdit(t *testing.T) {
	t.Parallel()
	s := parse(t, "3.4\n=A1 | round(0)\n")
	ins := s.InsertRow(addr(0, 0)) // blank row at top: every row shifts down one
	assert.Equal(t, "=A2 | round(0)", sourceAt(t, ins, 2, 0))
	assert.Equal(t, "3", cellAt(t, ins.Compute(), 2, 0))
}
