package engine_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// TestTextJoin_ExcelSemantics pins TEXTJOIN against Excel's three-argument
// signature: a delimiter, an ignore-empty flag, then any number of operands
// whose ranges flatten — the fold that makes a chart-pack sheet variable-arity
// instead of fixed at input(1)…input(8).
func TestTextJoin_ExcelSemantics(t *testing.T) {
	t.Parallel()

	// A 2x2 block with one empty cell, joined both ways: ignore_empty decides
	// whether the gap earns a separator, exactly as Excel does.
	src := "a\tb\nc\t\n=textjoin(\"-\", TRUE, A1:B2)\t=textjoin(\"-\", FALSE, A1:B2)\n"
	grid := compute(t, src)
	assert.Equal(t, "a-b-c", cellAt(t, grid, 2, 0), "an empty operand contributes nothing, separator included")
	assert.Equal(t, "a-b-c-", cellAt(t, grid, 2, 1), "kept, the empty operand still earns its separator")

	assert.Equal(t, "ab", formula1(t, `textjoin("", TRUE, "a", "b")`), "an empty delimiter is legal")
	assert.Equal(t, "1|TRUE|x", formula1(t, `textjoin("|", TRUE, 1, TRUE, "x")`),
		"operands render in the sheet's own canonical form")
	assert.Equal(t, "only", formula1(t, `textjoin("-", TRUE, "only")`), "one operand needs no separator")
}

// TestTextJoin_RefusalsAndArity pins the failure contract: Excel's arity, error
// propagation from any operand, and the byte budget — a fold over a column is
// exactly the unbounded-string shape a budget exists to stop.
func TestTextJoin_RefusalsAndArity(t *testing.T) {
	t.Parallel()

	assert.Equal(t, string(engine.ErrValue), formula1(t, `textjoin("-")`), "one argument is not the signature")
	assert.Equal(t, string(engine.ErrValue), formula1(t, `textjoin("-", TRUE)`), "two arguments name no text")
	assert.Equal(t, string(engine.ErrDiv), formula1(t, `textjoin("-", TRUE, 1/0, "x")`),
		"an operand error propagates rather than stringifying")
	assert.Equal(t, string(engine.ErrDiv), formula1(t, `textjoin(1/0, TRUE, "x")`),
		"the delimiter is an operand too")

	// Every operand is individually in budget (100 KB against 1 MB) — only
	// their SUM is not. Without this the assertion would pass on rept's own
	// refusal propagating, testing nothing about the fold.
	assert.Equal(t, string(engine.ErrLimit), foldOverLongColumn(t, `textjoin("", TRUE, A1:A20)`),
		"a fold whose operands each fit but whose result does not is #LIMIT!")
}

// foldOverLongColumn computes expr against a column of twenty 100 KB cells:
// each operand sits well inside the 1 MB byte budget while their concatenation
// is twice it, so only the fold's own ceiling can refuse.
func foldOverLongColumn(t *testing.T, expr string) string {
	t.Helper()
	var src strings.Builder
	for range 20 {
		_, _ = src.WriteString(`=rept("a", 100000)` + "\n")
	}
	_, _ = src.WriteString("=" + expr + "\n")
	return cellAt(t, compute(t, src.String()), 20, 0)
}

// TestConcat_FoldsRangesUnderTheBudget pins the sibling fold: CONCAT already
// flattened ranges (that is how the chart-pack example works today) but did so
// with no ceiling at all — a whole-column fold could build an unbounded string,
// the exact door R16 closes. Behaviour is otherwise unchanged.
func TestConcat_FoldsRangesUnderTheBudget(t *testing.T) {
	t.Parallel()

	src := "a\tb\nc\td\n=concat(A1:B2)\n"
	assert.Equal(t, "abcd", cellAt(t, compute(t, src), 2, 0), "a range flattens into the fold")
	assert.Equal(t, "ab", formula1(t, `concat("a", "b")`), "scalars are unchanged")
	assert.Equal(t, string(engine.ErrValue), formula1(t, `concat()`), "no operands is not a fold")
	assert.Equal(t, string(engine.ErrDiv), formula1(t, `concat("a", 1/0)`), "an operand error propagates")
	assert.Equal(t, string(engine.ErrLimit), foldOverLongColumn(t, `concat(A1:A20)`),
		"the fold is now bounded by the byte budget")
}

// TestTextJoin_IsAKnownFunction pins the checker/evaluator agreement a lazy
// builtin needs: a name the evaluator owns but the checker does not know would
// be reported as an unknown function on a sheet that computes fine.
func TestTextJoin_IsAKnownFunction(t *testing.T) {
	t.Parallel()

	sheet, err := engine.Parse([]byte("=textjoin(\"-\", TRUE, \"a\")\t=concat(\"a\")\n"))
	assert.NoError(t, err)
	assert.Empty(t, engine.Check(sheet), "neither fold may be reported as an unknown function")
}
