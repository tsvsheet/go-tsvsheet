// The lazy cell backing a windowed evaluation reads through (spec 016 6b):
// on-demand block reads with a shared failure latch, and the sparse-memo
// computer construction that carries it.
package engine

import "github.com/tsvsheet/go-tsvsheet/internal/index"

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
		env: embedEnv{
			loader:   opts.Loader,
			base:     opts.Base,
			visiting: map[Path]boolResult{opts.Base: true},
		},
		tick: opts.Tick,
	}
}
