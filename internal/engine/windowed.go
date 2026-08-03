// The census policy and the windowed document (spec 016 area 6): a source of
// any size opens through the one indexed path; the census decides which
// capability comes back. At or under the resident budget the caller gets the
// same fully resident Sheet Parse builds — editable, computable, byte-for-byte
// today's contract. Over it the caller gets a WindowedSheet: view/compute-only
// by ruling, serving bounded row windows through the block cache while the
// source stays on disk. One implementation either way; the type is the policy
// made visible.
package engine

import (
	"github.com/tsvsheet/go-tsvsheet/internal/index"
)

// SheetCensus is what one open-time scan learned about a document — the
// numbers the load policy and a frontend's chrome read.
type SheetCensus struct {
	Rows     int
	MaxWidth int
	Cells    int64
	Formulas int64
}

// WindowedSheet is the over-budget capability: bounded row windows over an
// indexed source. Values share the underlying reader (the mutex-guarded cache
// lives behind sheetSource's pointer), so it copies safely; OpenSheet returns
// a pointer only as the is-it-windowed signal.
type WindowedSheet struct {
	source sheetSource
	limits Limits
}

// OpenSheet loads a source of any size under limits: one scan builds the
// census, and the resident budget decides the capability. Exactly one of the
// returns is populated: an in-budget source materializes into a Sheet
// identical to Parse's; an over-budget one stands up a WindowedSheet without
// materializing anything.
func OpenSheet(src ByteSource, limits Limits) (Sheet, *WindowedSheet, error) {
	source, err := newSheetSource(readerAtSize{at: src.ReadAt, size: index.SourceSize(src.Size)})
	if err != nil {
		return Sheet{}, nil, err
	}
	effective := effectiveLimits(limits)
	if cellBudget(source.ix.Census().Cells) <= effective.residentBudget() {
		sheet, err := materializeSheet(source)
		return sheet, nil, err
	}
	return Sheet{}, &WindowedSheet{source: source, limits: effective}, nil
}

// ByteSource is an any-size byte source: a file, a spooled stream, or
// in-memory bytes, with its length.
type ByteSource struct {
	ReadAt readAtSource
	Size   int64
}

// Census reports the windowed document's totals.
func (w WindowedSheet) Census() SheetCensus {
	c := w.source.ix.Census()
	return SheetCensus{Rows: c.Rows, MaxWidth: c.MaxWidth, Cells: int64(c.Cells), Formulas: int64(c.Formulas)}
}

// Rows returns the source texts of the window [from, from+n), clipped to the
// grid on both ends — the viewport read: literals and formulas as written,
// nothing computed. Formula evaluation for a window arrives with ComputeRows
// (area 6b); a caller needing whole-document computation raises the resident
// budget instead (the owner's ruling: budgets are policy, never ceilings).
func (w WindowedSheet) Rows(from, n int) (Grid, error) {
	rows, err := w.source.reader.ReadRows(index.GridRow(from), index.RowCount(n))
	if err != nil {
		return nil, readFailure(err)
	}
	out := make(Grid, len(rows))
	for r, row := range rows {
		out[r] = make([]string, len(row))
		for c, cl := range row {
			out[r][c] = cl.text
		}
	}
	return out, nil
}

// lazyCells reads cells on demand through the block reader for a windowed
// evaluation, latching the first source failure behind a shared pointer: a
// read that fails answers out-of-grid to keep evaluation total, and the
// latched error fails the whole ComputeRows call afterwards — a broken source
// is an error, never wrong data (pinned by
// TestLazyCellsNeverServeWrongDataAfterAFailure).
type lazyCells struct {
	fail   *error
	source sheetSource
}

// newLazyCells stands a lazy backing with a fresh latch.
func newLazyCells(source sheetSource) lazyCells {
	return lazyCells{fail: new(error), source: source}
}

// at reads one cell through the block cache.
func (l lazyCells) at(row rowIndex, col colIndex) (cell, boolResult) {
	if *l.fail != nil || row < 0 || col < 0 {
		return cell{}, false
	}
	rows, err := l.source.reader.ReadRows(index.GridRow(row), 1)
	if err != nil {
		*l.fail = err
		return cell{}, false
	}
	if len(rows) == 0 || int(col) >= len(rows[0]) {
		return cell{}, false
	}
	return rows[0][col], true
}

// ComputeRows evaluates the window [from, from+n): literals as their values,
// formulas through the one resolver over a sparse memo bounded by the
// touched-cells budget — the design's cumulative bound, so computing the
// visible rows stays bounded regardless of what they reference; past the
// budget a cell answers #LIMIT!. An array-producing formula renders its
// top-left value (windows do not spill). A source failure during evaluation
// fails the call as ErrReadInput rather than serving partial data.
func (w WindowedSheet) ComputeRows(from, n int, opts ComputeOptions) (Grid, error) {
	rows, err := w.source.reader.ReadRows(index.GridRow(from), index.RowCount(n))
	if err != nil {
		return nil, readFailure(err)
	}
	lazy := newLazyCells(w.source)
	comp := lazy.computer(effectiveLimits(w.limits), opts)
	out := make(Grid, len(rows))
	for r, row := range rows {
		out[r] = make([]string, len(row))
		for c, cl := range row {
			out[r][c] = comp.cellValue(rowIndex(from+r), colIndex(c), cl).String()
		}
	}
	if *lazy.fail != nil {
		return nil, readFailure(*lazy.fail)
	}
	return out, nil
}

// computer builds the sparse-memo computer a window evaluates under this lazy
// backing: the touched-cells budget and the pass clock/limits the options
// carry; every value copy shares the latch through its pointer. Constructed
// directly — newComputer's dense memo sizes slabs by the grid height, which on
// a windowed document is the very allocation this path exists to avoid.
func (l lazyCells) computer(limits Limits, opts ComputeOptions) computer {
	return computer{
		now:     opts.At,
		rng:     newPassRNG(prngSeed(opts.At.UnixNano())),
		sheet:   Sheet{lazy: &l},
		memo:    newSparseMemo(limits.touchedBudget()),
		limits:  limits,
		fetcher: opts.Fetcher,
		tick:    opts.Tick,
	}
}
