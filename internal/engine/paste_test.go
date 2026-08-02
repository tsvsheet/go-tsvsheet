package engine_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// block builds a paste block from raw cell texts.
func block(rows ...[]string) engine.Grid { return rows }

// paste applies Sheet.Paste under the default limits, requiring success.
func paste(t *testing.T, s engine.Sheet, at, origin engine.Address, b engine.Grid) engine.Sheet {
	t.Helper()
	got, err := s.Paste(at, origin, b, engine.DefaultLimits())
	require.NoError(t, err)
	return got
}

func TestPaste_RebasesUnpinnedReferencesByTheDelta(t *testing.T) {
	t.Parallel()

	// A block copied at B1 pastes at B3: two rows down, so A1 follows to A3.
	s := parse(t, "10\t=A1*2\n20\n30\n")
	got := paste(t, s, addr(2, 1), addr(0, 1), block([]string{"=A1*2"}))

	assert.Equal(t, "=A3 * 2", sourceAt(t, got, 2, 1))
	assert.Equal(t, "60", cellAt(t, got.Compute(), 2, 1))
}

func TestPaste_PinPermutationsShiftPerAxis(t *testing.T) {
	t.Parallel()

	// One cell of each pin form, pasted one row down and one column right:
	// each axis honors its own pin independently, and a range endpoint pins
	// independently of the other endpoint.
	s := parse(t, "1\t2\n3\t4\n")
	b := block([]string{"=A1", "=$A1", "=A$1", "=$A$1", "=sum($B$1:B1)"})
	got := paste(t, s, addr(1, 1), addr(0, 0), b)

	assert.Equal(t, "=B2", sourceAt(t, got, 1, 1))
	assert.Equal(t, "=$A2", sourceAt(t, got, 1, 2))
	assert.Equal(t, "=B$1", sourceAt(t, got, 1, 3))
	assert.Equal(t, "=$A$1", sourceAt(t, got, 1, 4))
	assert.Equal(t, "=sum($B$1:C2)", sourceAt(t, got, 1, 5))
}

func TestPaste_NegativeDeltaOffGridBecomesRefError(t *testing.T) {
	t.Parallel()

	// Pasting above/left of the copy origin pushes A1 off the grid on each
	// axis: the reference renders #REF! and computes #REF!.
	s := parse(t, "1\t2\n3\t=A1\n")
	up := paste(t, s, addr(0, 1), addr(1, 1), block([]string{"=A1"}))
	left := paste(t, s, addr(1, 0), addr(1, 1), block([]string{"=A1"}))

	assert.Equal(t, "=#REF!", sourceAt(t, up, 0, 1))
	assert.Equal(t, "=#REF!", sourceAt(t, left, 1, 0))
	assert.Equal(t, "#REF!", cellAt(t, up.Compute(), 0, 1))
}

func TestPaste_CrossSheetReferencesCopyUnshifted(t *testing.T) {
	t.Parallel()

	s := parse(t, "x\ny\n")
	got := paste(t, s, addr(1, 0), addr(0, 0), block([]string{`="rates.tsvt"!B2`}))

	assert.Equal(t, `="rates.tsvt"!B2`, sourceAt(t, got, 1, 0))
}

func TestPaste_LiteralsCopyVerbatimAndEmptyCellsClear(t *testing.T) {
	t.Parallel()

	// The block's footprint overwrites wholesale: literals land verbatim (no
	// canonical re-render of a literal), and an empty block cell clears the
	// occupied target beneath it.
	s := parse(t, "old\tkeep\nstale\t=A1\n")
	got := paste(t, s, addr(0, 0), addr(5, 5), block([]string{"new"}, []string{""}))

	assert.Equal(t, "new", sourceAt(t, got, 0, 0))
	assert.Equal(t, "keep", sourceAt(t, got, 0, 1))
	assert.Equal(t, "", sourceAt(t, got, 1, 0))
	assert.Equal(t, "=A1", sourceAt(t, got, 1, 1))
}

func TestPaste_RaggedBlockClearsItsWholeFootprint(t *testing.T) {
	t.Parallel()

	// A ragged external block ("a\nb\tc") still overwrites its bounding
	// rectangle: the short first row pads with an empty cell, so the stale
	// value beneath it clears instead of surviving inside the pasted area.
	s := parse(t, "x\tKEEP\ny\tKEEP2\n")
	got := paste(t, s, addr(0, 0), addr(0, 0), engine.ParseBlock("a\nb\tc"))

	assert.Equal(t, "a", sourceAt(t, got, 0, 0))
	assert.Equal(t, "", sourceAt(t, got, 0, 1))
	assert.Equal(t, "b", sourceAt(t, got, 1, 0))
	assert.Equal(t, "c", sourceAt(t, got, 1, 1))
}

func TestPaste_ZeroDeltaParsesVerbatim(t *testing.T) {
	t.Parallel()

	// origin == at is the external-clipboard path: formulas parse with no
	// rebasing (canonical spacing is the only change).
	s := parse(t, "1\n2\n")
	got := paste(t, s, addr(0, 1), addr(0, 1), block([]string{"=A1+A2"}))

	assert.Equal(t, "=A1 + A2", sourceAt(t, got, 0, 1))
	assert.Equal(t, "3", cellAt(t, got.Compute(), 0, 1))
}

func TestPaste_GrowsTheGridWithinLimits(t *testing.T) {
	t.Parallel()

	s := parse(t, "1\n")
	got := paste(t, s, addr(2, 2), addr(0, 0), block([]string{"a", "b"}, []string{"c"}))

	require.Len(t, got.Source(), 4)
	assert.Equal(t, "a", sourceAt(t, got, 2, 2))
	assert.Equal(t, "b", sourceAt(t, got, 2, 3))
	assert.Equal(t, "c", sourceAt(t, got, 3, 2))
}

func TestPaste_FootprintBeyondTheLimitIsRejected(t *testing.T) {
	t.Parallel()

	// The whole footprint is bounded, on each axis: a block reaching GridDim
	// rows down or (via its widest row) columns right is refused up front.
	s := parse(t, "1\n")
	limits := engine.Limits{ResultCells: 100, GridDim: 4, ResultBytes: 100}

	_, err := s.Paste(addr(3, 0), addr(0, 0), block([]string{"a"}, []string{"b"}), limits)
	assert.ErrorIs(t, err, constants.ErrInvalidValue)

	_, err = s.Paste(addr(0, 3), addr(0, 0), block([]string{"a"}, []string{"b", "c"}), limits)
	assert.ErrorIs(t, err, constants.ErrInvalidValue)

	_, ok := s.Paste(addr(3, 3), addr(0, 0), block([]string{"z"}), limits)
	assert.NoError(t, ok)
}

func TestPaste_NegativeCornersAreRejected(t *testing.T) {
	t.Parallel()

	s := parse(t, "1\n")
	_, err := s.Paste(addr(-1, 0), addr(0, 0), block([]string{"a"}), engine.DefaultLimits())
	assert.ErrorIs(t, err, constants.ErrInvalidValue)

	_, err = s.Paste(addr(0, 0), addr(0, -1), block([]string{"a"}), engine.DefaultLimits())
	assert.ErrorIs(t, err, constants.ErrInvalidValue)
}

func TestPaste_MalformedFormulaIsAtomicNamingTheTargetCell(t *testing.T) {
	t.Parallel()

	// The second block cell is malformed: the error names the cell it would
	// have landed in (C2), and the sheet is unchanged — the valid first cell
	// did not land.
	s := parse(t, "1\t2\n3\t4\n")
	_, err := s.Paste(addr(1, 1), addr(0, 0), block([]string{"=A1", "=1+"}), engine.DefaultLimits())

	require.Error(t, err)
	assert.True(t, errors.Is(err, constants.ErrSyntax))
	assert.Contains(t, err.Error(), "C2")
}

func TestPaste_EmptyBlockGridIsANoOp(t *testing.T) {
	t.Parallel()

	// A Go caller passing a zero-row block pastes nothing. (ParseBlock never
	// produces one — its empty decode is a single empty cell.)
	s := parse(t, "1\t2\n")
	got := paste(t, s, addr(0, 0), addr(0, 0), engine.Grid{})
	assert.Equal(t, s.Source(), got.Source())
}

func TestPaste_EmptyTextClearsOneCell(t *testing.T) {
	t.Parallel()

	// The single-cell clear a UI's Delete key performs: paste the empty text
	// at the cell, origin = target.
	s := parse(t, "1\t2\n")
	got := paste(t, s, addr(0, 1), addr(0, 1), engine.ParseBlock(""))
	assert.Equal(t, "", sourceAt(t, got, 0, 1))
	assert.Equal(t, "1", sourceAt(t, got, 0, 0))
}

func TestPaste_RoundTripsItsOwnCopy(t *testing.T) {
	t.Parallel()

	// The 010 contract in one motion: copy B1:B2's source texts (what the
	// editor puts on the clipboard), paste two columns right, and the pasted
	// formulas read their own new column.
	s := parse(t, "1\t=A1*2\n2\t=A2*2\n")
	src := s.Source()
	b := block([]string{src[0][1]}, []string{src[1][1]})
	got := paste(t, s, addr(0, 3), addr(0, 1), b)

	g := got.Compute()
	assert.Equal(t, "=C1 * 2", sourceAt(t, got, 0, 3))
	assert.Equal(t, "=C2 * 2", sourceAt(t, got, 1, 3))
	assert.Equal(t, "0", cellAt(t, g, 0, 3)) // C1 is empty → 0
}

// pasteInto applies Sheet.PasteInto under the default limits, requiring success.
func pasteInto(t *testing.T, s engine.Sheet, target engine.Span, origin engine.Address, b engine.Grid) engine.Sheet {
	t.Helper()
	got, err := s.PasteInto(target, origin, b, engine.DefaultLimits())
	require.NoError(t, err)
	return got
}

func TestDocument_PasteGrowsLayoutAndPreservesComments(t *testing.T) {
	t.Parallel()

	// Growth appends row markers after trailing comments, as Fill's growth
	// does, and comment lines survive serialization.
	doc, err := engine.ParseDocument([]byte("# note\n=1+1\n"))
	require.NoError(t, err)

	got, err := doc.Paste(engine.Address{Row: 1, Col: 0}, engine.Address{Row: 0, Col: 0},
		block([]string{"x"}, []string{"y"}), engine.DefaultLimits())
	require.NoError(t, err)
	assert.Equal(t, "# note\n=1+1\nx\ny\n", string(got.Text()))
}

func TestDocument_PasteErrorLeavesTheDocumentUnchanged(t *testing.T) {
	t.Parallel()

	doc, err := engine.ParseDocument([]byte("1\n"))
	require.NoError(t, err)

	_, err = doc.Paste(engine.Address{Row: 0, Col: 0}, engine.Address{Row: 0, Col: 0},
		block([]string{"=("}), engine.DefaultLimits())
	assert.ErrorIs(t, err, constants.ErrSyntax)
	assert.Equal(t, "1\n", string(doc.Text()))
}

// TestPaddedLeavesNoStaleDataUnderAPastedRectangle pins the footprint rule. A
// ragged block still overwrites its whole rectangle, so a short row cannot
// leave the old value showing through in the gap — which would read as data the
// paste had brought, and is the worst kind of wrong: plausible.
func TestPaddedLeavesNoStaleDataUnderAPastedRectangle(t *testing.T) {
	t.Parallel()
	sheet, err := engine.Parse([]byte("old\told\nold\told\n"))
	require.NoError(t, err)

	ragged := engine.Grid{{"new", "new"}, {"new"}}
	pasted, err := sheet.Paste(engine.Address{}, engine.Address{}, ragged, engine.DefaultLimits())
	require.NoError(t, err)

	assert.Equal(t, "", pasted.Source()[1][1], "the short row's gap is cleared, not left holding old")
}

// TestPasteBoundsNeverWrapsOnAnUntrustedRow pins that a row at Atoi's ceiling
// cannot wrap `at.Row+len(block)` negative and slip past GridDim: the bound is
// subtraction-shaped, refusing immediately, never after allocation. The call
// runs under a deadline so a regression FAILS with a diagnosis instead of
// allocating until the runner dies.
func TestPasteBoundsNeverWrapsOnAnUntrustedRow(t *testing.T) {
	t.Parallel()

	s := parse(t, "x\n")
	done := make(chan error, 1)
	go func() {
		_, err := s.Paste(
			addr(math.MaxInt-1, 0),
			addr(0, 0),
			block([]string{"a"}, []string{"b"}),
			engine.DefaultLimits(),
		)
		done <- err
	}()
	select {
	case err := <-done:
		assert.ErrorIs(t, err, constants.ErrInvalidValue)
	case <-time.After(10 * time.Second):
		t.Fatal("no refusal within 10s — the wrapped bound authorised the paste")
	}
}
