// Windowed evaluation (spec 016 area 6b): ComputeRows runs the one resolver
// over a sparse budgeted memo with dependencies read lazily through the block
// cache. The capability type and its census policy live in windowed.go.
package engine

import "github.com/tsvsheet/go-tsvsheet/internal/index"

// ComputeRows evaluates the window [from, from+n): literals as their values,
// formulas through the one resolver over a sparse memo bounded by the
// touched-cells budget — the design's cumulative bound on values, memory, and
// source I/O alike; past the budget a cell answers #LIMIT!. The caller's
// ComputeOptions govern the pass exactly as they do a resident ComputeWith:
// its Limits, Loader/Base, Fetcher, clock, and Tick. An array-producing
// formula renders its top-left value (windows do not spill), except an anchor
// the document's own semantics refuses, which shows the resident render's
// #SPILL!. A source failure during evaluation fails the call as ErrReadInput
// rather than serving partial data. Volatile draws (rand and kin) are made in
// this window's evaluation order: one window at a fixed At reproduces
// exactly, but different windows over the same cells may draw differently —
// windowed evaluation has no whole-document order to share.
func (w WindowedSheet) ComputeRows(from, n int, opts ComputeOptions) (Grid, error) {
	if from < 0 {
		from = 0 // mirror ReadRows' clip, or the returned rows would be addressed off by the clip
	}
	rows, err := w.source.reader.ReadRows(index.GridRow(from), index.RowCount(n))
	if err != nil {
		return nil, readFailure(err)
	}
	names, err := windowedBindings(w.source)
	if err != nil {
		return nil, readFailure(err)
	}
	lazy := newLazyCells(w.source)
	comp := lazy.computer(names, effectiveLimits(opts.Limits), opts)
	out := make(Grid, len(rows))
	for r, row := range rows {
		out[r] = make([]string, len(row))
		for c, cl := range row {
			out[r][c] = windowCellText(comp, Address{Row: from + r, Col: c}, cl)
		}
	}
	if *lazy.fail != nil {
		return nil, readFailure(*lazy.fail)
	}
	return out, nil
}

// windowCellText renders one window cell exactly as the resident render does
// (pinned by TestWindowCellTextNeverReformatsALiteralNorServesARejectedSpill):
// a literal VERBATIM — a stored "00" or "4.50" is not reformatted, the
// standing display rule the differential fuzz caught this path violating —
// a formula's value, and, for an array anchor the document's own semantics
// refuses, the resident render's #SPILL!. An unblocked array renders its
// top-left value: windows do not spill.
func windowCellText(comp computer, at Address, cl cell) string {
	if !cl.isFormula() {
		return cl.text
	}
	v := comp.cellValue(rowIndex(at.Row), colIndex(at.Col), cl)
	switch v.kind {
	case kindArray:
		if comp.sheet.spillBlocked(at, v.arr) {
			return string(ErrSpill)
		}
		return v.String()
	default:
		return v.String()
	}
}
