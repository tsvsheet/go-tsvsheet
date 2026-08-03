// Package engine's compute entry points: the public Compute family, and the
// rendering that turns a computed value matrix back into a text grid.
package engine

import "time"

// Compute evaluates every formula in dependency order and returns the value
// grid: literal cells pass through verbatim, formula cells are replaced by their
// computed value. Volatile functions (TODAY/NOW) sample the wall clock once for
// the whole pass.
func (s Sheet) Compute() Grid { return s.ComputeAt(time.Now()) }

// ComputeAt is Compute with the clock injected, so volatile functions are
// deterministic within a pass (and testable). It computes every cell's value,
// then renders — spilling dynamic-array results into empty neighbours.
func (s Sheet) ComputeAt(at time.Time) Grid {
	return s.ComputeAtTick(at, 0)
}

// ComputeAtTick evaluates the sheet against clock at with the recompute-pass
// ordinal tick injected for tick()/frame(). A frontend that re-renders a
// volatile sheet increments tick each pass; ComputeAt uses 0.
func (s Sheet) ComputeAtTick(at time.Time, tick Tick) Grid {
	comp := newComputer(s, at)
	comp.tick = tick
	return s.computeGrid(comp)
}

// computeGrid evaluates every cell through comp and renders the value grid,
// spilling any dynamic-array results. It is the shared body of ComputeAt (plain)
// and ComputeWith (with an embedded-sheet loader).
func (s Sheet) computeGrid(comp computer) Grid {
	values := make([][]Value, s.height())
	for r, row := range s.rowsView() {
		values[r] = make([]Value, len(row))
		for c, cl := range row {
			values[r][c] = comp.cellValue(rowIndex(r), colIndex(c), cl)
		}
	}
	return s.render(values)
}

// dims is an output grid's row and column extent.
type dims struct{ rows, cols int }

// render turns the computed value grid into the output string grid, spilling any
// dynamic-array result down-and-right from its anchor.
func (s Sheet) render(values [][]Value) Grid {
	out := s.fillScalars(values, outputExtent(values))
	s.spillArrays(out, values)
	return out
}

// outputExtent is the grid extent needed once every array result has spilled.
func outputExtent(values [][]Value) dims {
	rows, cols := len(values), 0
	for r := range values {
		cols = max(cols, len(values[r]))
		for c := range values[r] {
			switch a := values[r][c]; a.kind {
			case kindArray:
				rows = max(rows, r+len(a.arr))
				cols = max(cols, c+len(a.arr[0]))
			default:
			}
		}
	}
	return dims{rows: rows, cols: cols}
}
