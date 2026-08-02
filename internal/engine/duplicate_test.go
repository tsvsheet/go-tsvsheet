package engine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// Duplicating a row or a column: the copy rebases one line over, the grid grows
// by one, and everything below or right of the insertion shifts exactly as an
// insert shifts it.

func TestDuplicateRow_RebasesTheDuplicateAndShiftsTheGrid(t *testing.T) {
	t.Parallel()

	// Duplicating row 2: the duplicate's A1 rebases to A2 (fill semantics),
	// while the range below shifts per the existing insert semantics.
	s := parse(t, "10\n=A1*2\n=sum(A1:A2)\n")
	got := s.DuplicateRow(addr(1, 0))

	require.Len(t, got.Source(), 4)
	assert.Equal(t, "=A1*2", sourceAt(t, got, 1, 0)) // source row untouched
	assert.Equal(t, "=A2 * 2", sourceAt(t, got, 2, 0))
	assert.Equal(t, "=sum(A1:A2)", sourceAt(t, got, 3, 0))
	g := got.Compute()
	assert.Equal(t, "20", cellAt(t, g, 1, 0))
	assert.Equal(t, "40", cellAt(t, g, 2, 0)) // doubles the duplicate above it
	assert.Equal(t, "30", cellAt(t, g, 3, 0)) // 10 + 20
}

func TestDuplicateRow_KeepsWidthAndPins(t *testing.T) {
	t.Parallel()

	// The duplicate keeps the source row's exact width (no widest-row padding
	// that would add trailing tabs), and pinned references hold.
	s := parse(t, "1\t2\t3\n=$A$1+B1\n")
	got := s.DuplicateRow(addr(1, 0))

	require.Len(t, got.Source()[2], 1)
	assert.Equal(t, "=$A$1 + B2", sourceAt(t, got, 2, 0))
}

func TestDuplicateRow_OutOfRangeIsNoOp(t *testing.T) {
	t.Parallel()

	s := parse(t, "1\n2\n")
	assert.Equal(t, s.Source(), s.DuplicateRow(addr(-1, 0)).Source())
	assert.Equal(t, s.Source(), s.DuplicateRow(addr(9, 0)).Source())
}

func TestDuplicateCol_RebasesTheDuplicateAndShiftsTheGrid(t *testing.T) {
	t.Parallel()

	// Duplicating column A: the duplicate's B1 rebases to C1 — which is where
	// the insert moved the data it referenced. The source cell's own moved
	// reference re-renders canonically, exactly as structural edits already
	// re-render moved formulas.
	s := parse(t, "=B1*2\t7\nx\ty\n")
	got := s.DuplicateCol(addr(0, 0))

	assert.Equal(t, "=C1 * 2", sourceAt(t, got, 0, 0)) // ref followed its data
	assert.Equal(t, "=D1 * 2", sourceAt(t, got, 0, 1))
	assert.Equal(t, "7", sourceAt(t, got, 0, 2))
	g := got.Compute()
	assert.Equal(t, "14", cellAt(t, g, 0, 0))
}

func TestDuplicateCol_RaggedRowTooShortStaysUntouched(t *testing.T) {
	t.Parallel()

	// Duplicating column B: the one-cell row never reaches it, so it stays
	// untouched, mirroring InsertCol.
	s := parse(t, "1\t2\t3\nonly-a\n")
	got := s.DuplicateCol(addr(0, 1))

	require.Len(t, got.Source()[1], 1)
	assert.Equal(t, "2", sourceAt(t, got, 0, 2)) // the duplicate
}

func TestDuplicateCol_OutOfRangeIsNoOp(t *testing.T) {
	t.Parallel()

	s := parse(t, "1\t2\n")
	assert.Equal(t, s.Source(), s.DuplicateCol(addr(0, -1)).Source())
	assert.Equal(t, s.Source(), s.DuplicateCol(addr(0, 9)).Source())
}

func TestDocument_DuplicateRowSplicesMarker(t *testing.T) {
	t.Parallel()

	// The duplicate lands where the next row was — after any comment written
	// between the rows, which keeps its gap exactly as InsertRow anchors it; a
	// sheet no-op is a document no-op.
	doc, err := engine.ParseDocument([]byte("=1*2\n# tail\n5\n"))
	require.NoError(t, err)

	got := doc.DuplicateRow(engine.Address{Row: 0, Col: 0})
	assert.Equal(t, "=1*2\n# tail\n=1 * 2\n5\n", string(got.Text()))

	noop := doc.DuplicateRow(engine.Address{Row: 9, Col: 0})
	assert.Equal(t, string(doc.Text()), string(noop.Text()))
}

func TestDocument_DuplicateColKeepsLayout(t *testing.T) {
	t.Parallel()

	doc, err := engine.ParseDocument([]byte("# note\n=1*3\t9\n"))
	require.NoError(t, err)
	got := doc.DuplicateCol(engine.Address{Row: 0, Col: 0})

	assert.Equal(t, "# note\n=1*3\t=1 * 3\t9\n", string(got.Text()))
}
