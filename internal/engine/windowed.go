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
	effective := effectiveLimits(limits)
	source, err := newSheetSource(
		readerAtSize{at: src.ReadAt, size: index.SourceSize(src.Size)},
		index.CellCount(effective.residentBudget()),
	)
	if err != nil {
		return Sheet{}, nil, err
	}
	if cellBudget(source.ix.Census().Cells) <= effective.residentBudget() {
		sheet, err := materializeSheet(source)
		return sheet, nil, err
	}
	return Sheet{}, &WindowedSheet{source: source, limits: effective}, nil
}

// ByteSource is an any-size byte source: a file, a spooled stream, or
// in-memory bytes, with its length. Size is trusted: bytes past it are not
// read, and a Size shorter than the actual data truncates at that byte — a
// caller pairing a stat with an open owns that pairing.
type ByteSource struct {
	ReadAt readAtSource
	Size   int64
}

// CachedCells reports the cells currently resident in the windowed block
// cache — bounded by the resident budget, and honest telemetry for a
// frontend's status line.
func (w WindowedSheet) CachedCells() int64 {
	return int64(w.source.reader.CachedCells())
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
