package engine_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// Pasting INTO a selection: one block repeated over a span. The rules here are
// about the span — how a block tiles it, when it does not tile at all, and what
// bounds the area — while paste_test.go covers placing a block once at a point.

func TestPasteInto_OneRowSpreadsOverSelectedRows(t *testing.T) {
	t.Parallel()

	// The reported flow: C1 (=A1/B1) copied, C2:C4 selected, pasted — every
	// selected row gets the formula rebased to ITSELF, exactly as three
	// separate pastes would.
	s := parse(t, "10\t2\t=A1/B1\n20\t0\n15\t3\n8\t4\n")
	got := pasteInto(t, s, span(1, 2, 3, 2), addr(0, 2), block([]string{"=A1/B1"}))

	assert.Equal(t, "=A2 / B2", sourceAt(t, got, 1, 2))
	assert.Equal(t, "=A3 / B3", sourceAt(t, got, 2, 2))
	assert.Equal(t, "=A4 / B4", sourceAt(t, got, 3, 2))
	g := got.Compute()
	assert.Equal(t, "#DIV/0!", cellAt(t, g, 1, 2)) // 20 / 0
	assert.Equal(t, "5", cellAt(t, g, 2, 2))
	assert.Equal(t, "2", cellAt(t, g, 3, 2))
}

func TestPasteInto_OneCellFillsAnyRectangle(t *testing.T) {
	t.Parallel()

	// A 1×1 block divides every span: the whole selection fills, each cell
	// rebased to its own position, pins holding.
	s := parse(t, "1\t2\n3\t4\n")
	got := pasteInto(t, s, span(0, 2, 1, 3), addr(0, 0), block([]string{"=$A$1+A1"}))

	assert.Equal(t, "=$A$1 + C1", sourceAt(t, got, 0, 2))
	assert.Equal(t, "=$A$1 + D1", sourceAt(t, got, 0, 3))
	assert.Equal(t, "=$A$1 + C2", sourceAt(t, got, 1, 2))
	assert.Equal(t, "=$A$1 + D2", sourceAt(t, got, 1, 3))
}

func TestPasteInto_BlockTilesAnExactMultiple(t *testing.T) {
	t.Parallel()

	// A 1×2 block over a 2×2 span tiles twice vertically.
	s := parse(t, "0\n0\n")
	got := pasteInto(t, s, span(0, 0, 1, 1), addr(0, 0), block([]string{"a", "b"}))

	assert.Equal(t, "a", sourceAt(t, got, 0, 0))
	assert.Equal(t, "b", sourceAt(t, got, 0, 1))
	assert.Equal(t, "a", sourceAt(t, got, 1, 0))
	assert.Equal(t, "b", sourceAt(t, got, 1, 1))
}

func TestPasteInto_NonMultipleSpanPlacesOnceAtTopLeft(t *testing.T) {
	t.Parallel()

	// A 2-row block over a 3-row span does not divide: single placement, no
	// half tile — Paste semantics, corners in any order.
	s := parse(t, "x\nx\nx\nkeep\n")
	got := pasteInto(t, s, span(2, 0, 0, 0), addr(0, 0), block([]string{"a"}, []string{"b"}))

	assert.Equal(t, "a", sourceAt(t, got, 0, 0))
	assert.Equal(t, "b", sourceAt(t, got, 1, 0))
	assert.Equal(t, "x", sourceAt(t, got, 2, 0))
	assert.Equal(t, "keep", sourceAt(t, got, 3, 0))
}

func TestPasteInto_SpanEqualToBlockIsExactlyPaste(t *testing.T) {
	t.Parallel()

	s := parse(t, "10\t=A1*2\n20\n")
	viaInto := pasteInto(t, s, span(1, 1, 1, 1), addr(0, 1), block([]string{"=A1*2"}))
	viaPaste := paste(t, s, addr(1, 1), addr(0, 1), block([]string{"=A1*2"}))
	assert.Equal(t, viaPaste.Source(), viaInto.Source())
}

func TestPasteInto_MultiRowBlockTilesByItsOwnHeight(t *testing.T) {
	t.Parallel()

	// A 2-row block over a 4-row span lands twice, whole — the row stride is
	// the BLOCK height, never one (a stride mutation repeats the first row
	// and grows past the span).
	s := parse(t, "x\nx\nx\nx\n")
	got := pasteInto(t, s, span(0, 0, 3, 0), addr(0, 0), block([]string{"a"}, []string{"b"}))

	require.Len(t, got.Source(), 4)
	assert.Equal(t, "a", sourceAt(t, got, 0, 0))
	assert.Equal(t, "b", sourceAt(t, got, 1, 0))
	assert.Equal(t, "a", sourceAt(t, got, 2, 0))
	assert.Equal(t, "b", sourceAt(t, got, 3, 0))
}

func TestPasteInto_TheReceiverSheetIsUntouched(t *testing.T) {
	t.Parallel()

	// Sheet is a value: a SUCCESSFUL tiled paste must never write through
	// into the receiver's rows (dropping the clone shares the backing array).
	s := parse(t, "keep\nkeep\n")
	got := pasteInto(t, s, span(0, 0, 1, 0), addr(0, 0), block([]string{"new"}))

	assert.Equal(t, "new", sourceAt(t, got, 0, 0))
	assert.Equal(t, "keep", sourceAt(t, s, 0, 0))
	assert.Equal(t, "keep", sourceAt(t, s, 1, 0))
}

func TestPasteInto_SpanAreaIsBoundedByResultCells(t *testing.T) {
	t.Parallel()

	// Four caller integers must not buy O(area) work: a span past the
	// result-cells ceiling is refused up front — an OOM inside wasm aborts
	// the runtime fatally and kills the engine for the page's life.
	s := parse(t, "1\n")
	limits := engine.Limits{ResultCells: 100, GridDim: 1000, ResultBytes: 100}
	_, err := s.PasteInto(span(0, 0, 10, 10), addr(0, 0), block([]string{"x"}), limits)
	assert.ErrorIs(t, err, constants.ErrInvalidValue)
	// At the ceiling exactly (10×10 = 100), the paste proceeds.
	_, ok := s.PasteInto(span(0, 0, 9, 9), addr(0, 0), block([]string{"x"}), limits)
	assert.NoError(t, ok)
}

func TestPasteInto_CommentMarkerRefusedInsideATile(t *testing.T) {
	t.Parallel()

	// The comment-marker guard fires for a first-column tile cell, naming it,
	// and the whole tiled paste is refused.
	s := parse(t, "x\nx\n")
	_, err := s.PasteInto(span(0, 0, 1, 0), addr(0, 0), block([]string{"#. note"}), engine.DefaultLimits())
	require.Error(t, err)
	assert.Equal(t, "x", sourceAt(t, s, 0, 0))
	assert.Equal(t, "x", sourceAt(t, s, 1, 0))
}

func TestPasteInto_AtomicAcrossTiles(t *testing.T) {
	t.Parallel()

	// The malformed formula lands only in the SECOND tile; the whole tiled
	// paste is refused, naming that tile's target cell, and nothing changed.
	s := parse(t, "ok\nok\n")
	_, err := s.PasteInto(span(0, 0, 1, 0), addr(5, 0), block([]string{"=1+"}), engine.DefaultLimits())
	require.Error(t, err)
	assert.ErrorIs(t, err, constants.ErrSyntax)
	assert.Equal(t, "ok", sourceAt(t, s, 0, 0))
	assert.Equal(t, "ok", sourceAt(t, s, 1, 0))
}

func TestPasteInto_SpanBeyondTheLimitIsRejected(t *testing.T) {
	t.Parallel()

	s := parse(t, "1\n")
	limits := engine.Limits{ResultCells: 100, GridDim: 4, ResultBytes: 100}
	_, err := s.PasteInto(span(0, 0, 3, 4), addr(0, 0), block([]string{"a"}), limits)
	assert.ErrorIs(t, err, constants.ErrInvalidValue)
	_, err = s.PasteInto(span(-1, 0, 0, 0), addr(0, 0), block([]string{"a"}), limits)
	assert.ErrorIs(t, err, constants.ErrInvalidValue)
	_, err = s.PasteInto(span(0, 0, 0, 0), addr(-1, 0), block([]string{"a"}), limits)
	assert.ErrorIs(t, err, constants.ErrInvalidValue)
}

func TestPasteInto_EmptyBlockNeverTiles(t *testing.T) {
	t.Parallel()

	// A zero-row block cannot tile (it would loop forever); it falls back to
	// Paste, whose empty-block placement is a no-op.
	s := parse(t, "1\t2\n")
	got := pasteInto(t, s, span(0, 0, 1, 1), addr(0, 0), engine.Grid{})
	assert.Equal(t, s.Source(), got.Source())
}

func TestDocument_PasteIntoTilesAndGrowsLayout(t *testing.T) {
	t.Parallel()

	doc, err := engine.ParseDocument([]byte("# note\n=1+1\n"))
	require.NoError(t, err)
	got, err := doc.PasteInto(
		engine.Span{From: engine.Address{Row: 1, Col: 0}, To: engine.Address{Row: 2, Col: 0}},
		engine.Address{Row: 1, Col: 0}, block([]string{"x"}), engine.DefaultLimits(),
	)
	require.NoError(t, err)
	assert.Equal(t, "# note\n=1+1\nx\nx\n", string(got.Text()))

	_, err = doc.PasteInto(
		engine.Span{From: engine.Address{Row: 0, Col: 0}, To: engine.Address{Row: 1, Col: 0}},
		engine.Address{Row: 0, Col: 0}, block([]string{"=("}), engine.DefaultLimits(),
	)
	require.Error(t, err)
	assert.Equal(t, "# note\n=1+1\n", string(doc.Text()))
}

// TestTilesSpanNeverTilesAnEmptyBlock pins the degenerate case. An empty block
// divides into any span an unbounded number of times, so "does it tile?" has no
// useful answer for it; Paste no-ops instead of filling the selection with
// nothing a very large number of times.
func TestTilesSpanNeverTilesAnEmptyBlock(t *testing.T) {
	t.Parallel()
	sheet, err := engine.Parse([]byte("a\tb\nc\td\n"))
	require.NoError(t, err)

	pasted, err := sheet.PasteInto(span(0, 0, 1, 1), engine.Address{}, engine.Grid{}, engine.DefaultLimits())
	require.NoError(t, err)

	assert.Equal(t, sheet.Source(), pasted.Source(), "an empty block changes nothing")
}

// TestSpanBoundsRefusesAnAreaTheCallerCouldNotCatch pins the ceiling. The work
// and the allocation are O(area) while the request is four integers, so an
// unbounded span is an out-of-memory a caller has no way to recover from — in
// the wasm build a fatal abort that takes the page's engine with it.
func TestSpanBoundsRefusesAnAreaTheCallerCouldNotCatch(t *testing.T) {
	t.Parallel()
	sheet, err := engine.Parse([]byte("a\n"))
	require.NoError(t, err)

	_, err = sheet.PasteInto(span(0, 0, 1<<20, 1<<20), engine.Address{}, engine.Grid{{"x"}}, engine.DefaultLimits())

	require.Error(t, err, "four integers must not authorise a trillion cells")
}

// TestPasteTilesFillsAnExactlyDivisibleSpanOverOneCopy pins both halves of the
// tiling contract: the selection is filled tile by tile, and a thousand-cell
// selection is one allocation rather than a thousand — the structural fact that
// makes a later-tile failure unable to leave the sheet half-written.
func TestPasteTilesFillsAnExactlyDivisibleSpanOverOneCopy(t *testing.T) {
	t.Parallel()
	sheet, err := engine.Parse([]byte(".\t.\t.\t.\n.\t.\t.\t.\n"))
	require.NoError(t, err)

	tiled, err := sheet.PasteInto(span(0, 0, 1, 3), engine.Address{}, engine.Grid{{"a", "b"}}, engine.DefaultLimits())
	require.NoError(t, err)

	assert.Equal(t, []string{"a", "b", "a", "b"}, tiled.Source()[0], "the block repeats across the span")
	assert.Equal(t, []string{"a", "b", "a", "b"}, tiled.Source()[1], "and down it")
}

// TestPasteIntoSpanAreaNeverWrapsOnMultiplication pins that the span-area
// bound goes through the dimension-first tooManyCells, never the raw
// rows*cols product: two dimensions near sqrt(MaxInt) multiply to a NEGATIVE
// int, which passed the former comparison and authorised a 9-quintillion-cell
// span under a 10-cell budget.
func TestPasteIntoSpanAreaNeverWrapsOnMultiplication(t *testing.T) {
	t.Parallel()

	s := parse(t, "1\n")
	limits := engine.Limits{ResultCells: 10, GridDim: math.MaxInt, ResultBytes: 100}
	wide := span(0, 0, 3037000499, 3037000499) // 3037000500^2 wraps negative in int64
	done := make(chan error, 1)
	go func() {
		_, err := s.PasteInto(wide, addr(0, 0), block([]string{"a"}), limits)
		done <- err
	}()
	select {
	case err := <-done:
		assert.ErrorIs(t, err, constants.ErrInvalidValue)
	case <-time.After(10 * time.Second):
		t.Fatal("no refusal within 10s — the wrapped area authorised the span")
	}
}
