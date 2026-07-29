package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

func TestArray_SequenceSpills(t *testing.T) {
	t.Parallel()

	g := compute(t, "=sequence(2, 3)\n")
	require.Len(t, g, 2)
	assert.Equal(t, []string{"1", "2", "3"}, g[0]) // spills right
	assert.Equal(t, []string{"4", "5", "6"}, g[1]) // and down (grid grew)
}

func TestArray_SequenceSingleArg(t *testing.T) {
	t.Parallel()

	g := compute(t, "=sequence(3)\n") // one column
	assert.Equal(t, "1", cellAt(t, g, 0, 0))
	assert.Equal(t, "2", cellAt(t, g, 1, 0))
	assert.Equal(t, "3", cellAt(t, g, 2, 0))
}

func TestArray_SpillBlocked(t *testing.T) {
	t.Parallel()

	// A literal below the anchor blocks the spill.
	g := compute(t, "=sequence(3, 1)\nX\n")
	assert.Equal(t, string(engine.ErrSpill), cellAt(t, g, 0, 0))
	assert.Equal(t, "X", cellAt(t, g, 1, 0)) // the blocker is untouched
}

func TestArray_TransposeUniqueSortFilterFlatten(t *testing.T) {
	t.Parallel()

	// Transpose a row into a column.
	tr := compute(t, "1\t2\t3\n=transpose(A1:C1)\n")
	assert.Equal(t, "1", cellAt(t, tr, 1, 0))
	assert.Equal(t, "3", cellAt(t, tr, 3, 0))

	// Unique keeps first occurrences.
	uq := compute(t, "a\na\nb\n=unique(A1:A3)\n")
	assert.Equal(t, "a", cellAt(t, uq, 3, 0))
	assert.Equal(t, "b", cellAt(t, uq, 4, 0))

	// Sort a numeric column ascending, and a text column.
	sn := compute(t, "3\n1\n2\n=sort(A1:A3)\n")
	assert.Equal(t, "1", cellAt(t, sn, 3, 0))
	assert.Equal(t, "3", cellAt(t, sn, 5, 0))
	st := compute(t, "c\na\nb\n=sort(A1:A3)\n")
	assert.Equal(t, "a", cellAt(t, st, 3, 0)) // lexicographic

	// Filter keeps rows whose condition is truthy.
	fl := compute(t, "10\t1\n20\t0\n30\t1\n=filter(A1:A3, B1:B3)\n")
	assert.Equal(t, "10", cellAt(t, fl, 3, 0))
	assert.Equal(t, "30", cellAt(t, fl, 4, 0))

	// Flatten stacks all cells into a column.
	ft := compute(t, "1\t2\n3\t4\n=flatten(A1:B2)\n")
	assert.Equal(t, "1", cellAt(t, ft, 2, 0))
	assert.Equal(t, "4", cellAt(t, ft, 5, 0))
}

func TestArray_ScalarContext(t *testing.T) {
	t.Parallel()

	// An array in a scalar context reduces to its top-left value.
	assert.Equal(t, "1", formula1(t, "sequence(2, 2) + 0"))
}

func TestArray_Errors(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"sequence()":          string(engine.ErrValue), // arity
		"sequence(0)":         string(engine.ErrValue), // rows < 1
		"sequence(1, 0)":      string(engine.ErrValue), // cols < 1
		"sequence(6000000)":   string(engine.ErrValue), // exceeds the default cell budget (OOM guard)
		"sequence(3000,3000)": string(engine.ErrValue), // rows×cols (9M) exceeds the budget
		"transpose(A1, A2)":   string(engine.ErrValue), // arity
		"unique(A1, A2)":      string(engine.ErrValue),
		"sort(A1, A2)":        string(engine.ErrValue),
		"filter(A1)":          string(engine.ErrValue),
		"flatten(A1, A2)":     string(engine.ErrValue),
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, formula1(t, expr))
		})
	}

	// FILTER with no truthy condition is #N/A.
	assert.Equal(t, string(engine.ErrNA),
		cellAt(t, compute(t, "10\t0\n20\t0\n=filter(A1:A2, B1:B2)\n"), 2, 0))
	// A short condition range leaves later rows unmatched.
	assert.Equal(t, "10",
		cellAt(t, compute(t, "10\t1\n20\t1\n=filter(A1:A2, B1:B1)\n"), 2, 0))

	// Non-numeric dimension arguments propagate #VALUE! (A1 holds text).
	for _, expr := range []string{"=sequence(A1)", "=sequence(1, A1)"} {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, "#VALUE!", cellAt(t, compute(t, "hi\t"+expr+"\n"), 0, 1))
		})
	}
}

// Array-valued arguments (ADR 0004 §2, go-tsvsheet#1): a nested call that
// evaluates to an array is consumed by the enclosing function exactly as if it
// were a range of the same shape.

func TestArray_NestedInAggregates(t *testing.T) {
	t.Parallel()

	// Data column {3, 1, 2, 3}: aggregates consume the inner array flattened.
	const data = "3\n1\n2\n3\n"
	cases := map[string]string{
		"=sum(sort(A1:A4))":      "9", // not the top-left element
		"=count(unique(A1:A4))":  "3", // counts distinct numbers, not 0
		"=max(unique(A1:A4))":    "3",
		"=avg(flatten(A1:A4))":   "2.25",
		"=counta(unique(A1:A4))": "3",
		"=sum(sequence(4))":      "10", // 1+2+3+4
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, cellAt(t, compute(t, data+expr+"\n"), 4, 0))
		})
	}
}

func TestArray_PipeChainCountsDistinct(t *testing.T) {
	t.Parallel()

	// The §5.4 pipe-operator spec's own chain, and its composed spelling —
	// the two are one formula by normalization and must agree.
	g := compute(t, "3\n1\n2\n3\n=A1:A4 | sort() | unique() | count()\n=count(unique(sort(A1:A4)))\n")
	assert.Equal(t, "3", cellAt(t, g, 4, 0))
	assert.Equal(t, cellAt(t, g, 4, 0), cellAt(t, g, 5, 0))
}

func TestArray_NestedInLookupAndArrayFns(t *testing.T) {
	t.Parallel()

	// INDEX reads the inner array's 2-D shape: row 2 of sorted {3,1,2} is 2.
	assert.Equal(t, "2", cellAt(t, compute(t, "3\n1\n2\n=index(sort(A1:A3), 2)\n"), 3, 0))

	// TRANSPOSE of TRANSPOSE round-trips a row through the 2-D path and spills.
	g := compute(t, "1\t2\t3\n=transpose(transpose(A1:C1))\n")
	assert.Equal(t, []string{"1", "2", "3"}, g[1])

	// UNIQUE over SORT keeps the sorted first occurrences.
	sorted := compute(t, "3\n1\n3\n=unique(sort(A1:A3))\n")
	assert.Equal(t, "1", cellAt(t, sorted, 3, 0))
	assert.Equal(t, "3", cellAt(t, sorted, 4, 0))
}

func TestArray_NestedInCriteria(t *testing.T) {
	t.Parallel()

	// COUNTIF/SUMIF consume the inner array like a range: distinct {3,1,2}.
	g := compute(t, "3\n1\n2\n3\n=countif(unique(A1:A4), \">1\")\n=sumif(unique(A1:A4), \">1\")\n")
	assert.Equal(t, "2", cellAt(t, g, 4, 0))
	assert.Equal(t, "5", cellAt(t, g, 5, 0))
}

func TestArray_ErrorElementShortCircuits(t *testing.T) {
	t.Parallel()

	// An error element inside the array propagates exactly as an error cell in
	// a range argument would.
	g := compute(t, "3\n=1/0\n=sum(sort(A1:A2))\n")
	assert.Equal(t, string(engine.ErrDiv), cellAt(t, g, 2, 0))
}

func TestArray_ScalarContextReducesToTopLeft(t *testing.T) {
	t.Parallel()

	// Operators and IF's condition reduce an array operand to its top-left
	// element (no broadcasting — pinned in ADR 0004 §2).
	g := compute(t, "=sequence(2, 2) + 10\n\n=if(sequence(2, 2), \"y\", \"n\")\n")
	assert.Equal(t, "11", cellAt(t, g, 0, 0))
	assert.Equal(t, "y", cellAt(t, g, 2, 0))
}
