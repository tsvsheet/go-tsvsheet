package engine_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// sourceAt reads the source text of the cell at a 0-based (row, col).
func sourceAt(t *testing.T, s engine.Sheet, row, col int) string {
	t.Helper()
	g := s.Source()
	require.Less(t, row, len(g))
	require.Less(t, col, len(g[row]))
	return g[row][col]
}

func TestInsertRow_ShiftsReferencesAndInsertsBlank(t *testing.T) {
	t.Parallel()

	// B1 = A1 (row above the insert, unchanged); B2 = sum(A1:A2) (range whose
	// lower endpoint follows its data down).
	s := parse(t, "10\t=A1\n20\t=sum(A1:A2)\n")
	got := s.InsertRow(addr(1, 0)) // blank row before row index 1

	assert.Equal(t, "=A1", sourceAt(t, got, 0, 1))         // A1 is above the insert
	assert.Equal(t, "=sum(A1:A3)", sourceAt(t, got, 2, 1)) // range grew over the blank
	g := got.Compute()
	assert.Equal(t, "10", cellAt(t, g, 0, 1))
	assert.Equal(t, "30", cellAt(t, g, 2, 1)) // 10 + 0(blank) + 20
	assert.Equal(t, "", cellAt(t, g, 1, 0))   // the inserted row is empty
	assert.Equal(t, "", cellAt(t, g, 1, 1))
}

func TestInsertRow_PastEndAppendsBlank(t *testing.T) {
	t.Parallel()

	// An index beyond the grid clamps to the end: a trailing blank row of empty
	// cells (as wide as the widest row).
	got := parse(t, "1\n2\n").InsertRow(addr(9, 0))
	assert.Len(t, got.Source(), 3)
	assert.Equal(t, []string{""}, got.Source()[2])
}

func TestDeleteRow_ReferenceToDeletedBecomesRef(t *testing.T) {
	t.Parallel()

	// A3 = A2 (single ref to the deleted row) → #REF!; A4 = sum(A1:A2) (range
	// whose lower endpoint is deleted) shrinks to sum(A1:A1).
	s := parse(t, "10\n20\n=A2\n=sum(A1:A2)\n")
	got := s.DeleteRow(addr(1, 0)) // delete row A2

	assert.Equal(t, "=#REF!", sourceAt(t, got, 1, 0))      // old A3, ref deleted
	assert.Equal(t, "=sum(A1:A1)", sourceAt(t, got, 2, 0)) // range endpoint clamped
	g := got.Compute()
	assert.Equal(t, "#REF!", cellAt(t, g, 1, 0))
	assert.Equal(t, "10", cellAt(t, g, 2, 0)) // sum(A1:A1)
}

func TestDeleteRow_WholeRangeDeletedCollapses(t *testing.T) {
	t.Parallel()

	// B1 = sum(A2:A2) references only row A2; deleting that row collapses the
	// range argument to #REF! (lo > hi). The formula sits in row 0, which
	// survives; only the reference (not the whole call) becomes #REF!.
	got := parse(t, "10\t=sum(A2:A2)\n20\n").DeleteRow(addr(1, 0))
	assert.Equal(t, "=sum(#REF!)", sourceAt(t, got, 0, 1))
	assert.Equal(t, "#REF!", cellAt(t, got.Compute(), 0, 1))
}

func TestDeleteRow_EveryShiftBranch(t *testing.T) {
	t.Parallel()

	// Row 0 probes every remapping of a delete at row A2: a single ref above
	// (A1, unchanged), on (A2 → #REF!), and below (A3 → A2) the deletion; and
	// range endpoints below (A3:A3), straddling (A1:A3), and above (A1:A1) it.
	s := parse(t, "10\t=A1\t=A2\t=A3\t=sum(A3:A3)\t=sum(A1:A3)\t=sum(A1:A1)\n20\n30\n")
	got := s.DeleteRow(addr(1, 0))

	assert.Equal(t, "=A1", sourceAt(t, got, 0, 1))         // above → unchanged
	assert.Equal(t, "=#REF!", sourceAt(t, got, 0, 2))      // on → deleted
	assert.Equal(t, "=A2", sourceAt(t, got, 0, 3))         // below → shifts up
	assert.Equal(t, "=sum(A2:A2)", sourceAt(t, got, 0, 4)) // range below → shifts up
	assert.Equal(t, "=sum(A1:A2)", sourceAt(t, got, 0, 5)) // straddling → high shifts
	assert.Equal(t, "=sum(A1:A1)", sourceAt(t, got, 0, 6)) // above → unchanged
}

func TestDeleteRow_OutOfRangeIsNoOp(t *testing.T) {
	t.Parallel()

	s := parse(t, "1\n=A1\n")
	assert.Equal(t, s.Source(), s.DeleteRow(addr(9, 0)).Source())
	assert.Equal(t, s.Source(), s.DeleteRow(addr(-1, 0)).Source())
}

func TestInsertCol_ShiftsColumnReferences(t *testing.T) {
	t.Parallel()

	// C1 = A1 + B1: inserting a column before B pushes B1's data (and its
	// reference) to C, so B1 becomes C1 and C1 becomes D1.
	got := parse(t, "1\t2\t=A1 + B1\n").InsertCol(addr(0, 1))
	assert.Equal(t, "=A1 + C1", sourceAt(t, got, 0, 3))
	assert.Equal(t, "3", cellAt(t, got.Compute(), 0, 3)) // 1 + 2
}

func TestDeleteCol_ReferenceToDeletedBecomesRef(t *testing.T) {
	t.Parallel()

	// Deleting column B (the 2) makes =A1+B1 read a deleted cell → #REF!.
	got := parse(t, "1\t2\t=A1 + B1\n").DeleteCol(addr(0, 1))
	assert.Equal(t, "=A1 + #REF!", sourceAt(t, got, 0, 1))
	assert.Equal(t, "#REF!", cellAt(t, got.Compute(), 0, 1))
}

func TestDeleteCol_OutOfRangeIsNoOp(t *testing.T) {
	t.Parallel()

	s := parse(t, "1\t=A1\n")
	assert.Equal(t, s.Source(), s.DeleteCol(addr(0, 9)).Source())
}

func TestStructuralEdits_RaggedRowsAndAllNodeForms(t *testing.T) {
	t.Parallel()

	// A ragged grid: row 0 reaches column C, row 1 has a single short cell.
	// The formula exercises every mapRefs branch — unary, percent, binary,
	// call, and a bare literal (default) — so a column edit rewrites through
	// all of them without disturbing the literal.
	s := parse(t, "1\t=-A1 + sum(A1:A1) & B1% & \"x\"\t7\n9\n")
	const formula = "=-A1 + sum(A1:A1) & B1% & \"x\""

	ins := s.InsertCol(addr(0, 2))                   // insert before column C; short row 1 cannot reach it
	assert.Len(t, ins.Source()[1], 1)                // ragged row untouched
	assert.Equal(t, formula, sourceAt(t, ins, 0, 1)) // no ref is at/after C, so unchanged
	assert.Len(t, ins.Source()[0], 4)                // a blank column was spliced in

	del := s.DeleteCol(addr(0, 2))                   // drop column C (the 7); short row 1 cannot reach it
	assert.Len(t, del.Source()[1], 1)                // ragged row untouched
	assert.Len(t, del.Source()[0], 2)                // row 0 now ends at the formula
	assert.Equal(t, formula, sourceAt(t, del, 0, 1)) // literal "x" and all nodes preserved
}

func TestStructuralEdits_RetainAbsoluteMarkers(t *testing.T) {
	t.Parallel()

	// SPECIFICATION §4: `$` pins are retained through re-rendering, and — as in
	// Excel — a structural edit shifts a pinned reference exactly like an
	// unpinned one (the data it names moved; `$` pins only under copy/fill).
	s := parse(t, "10\t=$A$1*2\t=sum($A$1:A1)\t=A$1+$A1\n")
	got := s.InsertRow(addr(0, 0)) // blank row above: every row-1 ref follows to row 2

	assert.Equal(t, "=$A$2 * 2", sourceAt(t, got, 1, 1))
	assert.Equal(t, "=sum($A$2:A2)", sourceAt(t, got, 1, 2))
	assert.Equal(t, "=A$2 + $A2", sourceAt(t, got, 1, 3))
	g := got.Compute()
	assert.Equal(t, "20", cellAt(t, g, 1, 1)) // pins carry no positional difference
	assert.Equal(t, "10", cellAt(t, g, 1, 2))
	assert.Equal(t, "20", cellAt(t, g, 1, 3))
}

func TestStructuralEdits_UnmovedPinnedFormulaKeepsSpelling(t *testing.T) {
	t.Parallel()

	// A formula none of whose references moved is not re-rendered, so its
	// original `$` spelling (and spacing) survives byte-for-byte.
	s := parse(t, "10\t=$A$1*2\n")
	got := s.InsertRow(addr(1, 0)) // insert below: nothing at/after row 2 is referenced

	assert.Equal(t, "=$A$1*2", sourceAt(t, got, 0, 1))
}

// TestSameReferenceComparesWhatAReferenceNamesNotHowItWasWritten pins the
// comparison used to decide whether a shift changed anything. It is done in
// rendered form because that form is canonical for references: two spellings
// that name the same cell must compare equal, or an edit would report changes
// it did not make.
func TestSameReferenceComparesWhatAReferenceNamesNotHowItWasWritten(t *testing.T) {
	t.Parallel()
	sheet, err := engine.Parse([]byte("1\t2\n=$A$1+1\t=A1+1\n"))
	require.NoError(t, err)

	edited := sheet.InsertCol(engine.Address{Col: 5})

	// Unchanged means untouched: the formula keeps the author's spacing rather
	// than being re-rendered into canonical form, which is the visible payoff
	// of comparing what a reference NAMES instead of how it was written.
	assert.Equal(t, "=$A$1+1", edited.Source()[1][0], "a pinned reference past the edit is left as written")
	assert.Equal(t, "=A1+1", edited.Source()[1][1], "and so is a relative one")
}

// TestShiftPointAndShiftSpanCannotTouchAnUnaddressableColumn pins the rewrite rule for
// a reference whose column letters exceed the addressable bound (lettersToIndex
// refuses the run): no insert or delete at a real index can affect such a
// column, so the formula text is left byte-for-byte as written — never
// re-rendered through a wrapped index — and it computes as #REF! regardless.
// The range case keeps the WHOLE reference verbatim even though its in-bounds
// endpoint would otherwise shift: a reference that is already an error is not
// half-rewritten.
func TestShiftPointAndShiftSpanCannotTouchAnUnaddressableColumn(t *testing.T) {
	t.Parallel()

	huge := strings.Repeat("Z", 9)
	src := "1\t=" + huge + "1\t=sum(A1:" + huge + "1)\n"
	for name, edit := range map[string]func(engine.Sheet) engine.Sheet{
		"insert col": func(s engine.Sheet) engine.Sheet { return s.InsertCol(addr(0, 0)) },
		"delete col": func(s engine.Sheet) engine.Sheet { return s.DeleteCol(addr(0, 0)) },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := edit(parse(t, src))
			cells := got.Source()[0]
			assert.Contains(t, cells, "="+huge+"1")
			assert.Contains(t, cells, "=sum(A1:"+huge+"1)")
		})
	}
}

// TestShiftNeverMintsAnUnaddressableReference pins the output guard: EFWFSWHTP
// is the LAST addressable column, so an insert that would shift a reference to
// it one further must collapse the reference to #REF! — never render EFWFSWHTQ,
// a spelling the library's own parser, resolver, and edit language all refuse.
func TestShiftNeverMintsAnUnaddressableReference(t *testing.T) {
	t.Parallel()

	s := parse(t, "=EFWFSWHTP1\t=sum(A1:EFWFSWHTP1)\n")
	got := s.InsertCol(addr(0, 0))

	cells := got.Source()[0]
	assert.Contains(t, cells, "=#REF!", "the point reference shifted past the bound collapses")
	for _, cell := range cells {
		assert.NotContains(t, cell, "EFWFSWHTQ", "no cell may hold the unparseable spelling")
	}
}

// TestShiftToExactlyTheLastAddressableColumnIsRendered pins the TOP of the
// axis bound from below: EFWFSWHTO shifts to EFWFSWHTP — the LAST addressable
// column — and must be rendered, not collapsed. An off-by-one narrowing of
// axis.inBounds (`<=` to `<`) collapses exactly this shift and nothing the
// one-past tests can see.
func TestShiftToExactlyTheLastAddressableColumnIsRendered(t *testing.T) {
	t.Parallel()

	got := parse(t, "=EFWFSWHTO1\n").InsertCol(addr(0, 0))
	assert.Equal(t, "=EFWFSWHTP1", sourceAt(t, got, 0, 1))
}
