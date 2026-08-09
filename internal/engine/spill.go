// Spilling: rendering a computed sheet's scalars and dynamic-array results
// into the output grid. fillScalars lays down each cell's own value;
// spillArrays then writes each array result from its anchor, or #SPILL! when a
// target cell is occupied — the one rule both the resident render and the
// windowed render answer to.
package engine

// fillScalars renders each cell at its own position: a literal verbatim, a
// formula's computed value (an array renders its top-left anchor here;
// spillArrays overwrites the spilled cells).
func (s Sheet) fillScalars(values [][]Value, d dims) Grid {
	out := make(Grid, d.rows)
	for r := range out {
		out[r] = make([]string, d.cols)
		for c := range out[r] {
			out[r][c] = s.scalarText(values, Address{Row: r, Col: c})
		}
	}
	return out
}

// scalarText renders one cell's own text: a literal verbatim, a formula's value,
// empty when beyond the source grid.
func (s Sheet) scalarText(values [][]Value, at Address) string {
	if at.Row >= len(values) || at.Col >= len(values[at.Row]) {
		return ""
	}
	if cl := s.rowsView()[at.Row][at.Col]; !cl.isFormula() {
		return cl.text
	}
	return values[at.Row][at.Col].String()
}

// spillArrays writes each array result into the output grid.
func (s Sheet) spillArrays(out Grid, values [][]Value) {
	for r := range values {
		for c := range values[r] {
			switch values[r][c].kind {
			case kindArray:
				s.spill(out, Address{Row: r, Col: c}, values[r][c].arr)
			default:
			}
		}
	}
}

// spill writes an array from its anchor, or #SPILL! at the anchor when a target
// cell is occupied.
func (s Sheet) spill(out Grid, anchor Address, arr [][]Value) {
	if s.spillBlocked(anchor, arr) {
		out[anchor.Row][anchor.Col] = string(ErrSpill)
		return
	}
	for i := range arr {
		for j := range arr[i] {
			out[anchor.Row+i][anchor.Col+j] = arr[i][j].String()
		}
	}
}

// spillBlocked reports whether any non-anchor target cell already holds content.
func (s Sheet) spillBlocked(anchor Address, arr [][]Value) boolResult {
	for i := range arr {
		for j := range arr[i] {
			target := Address{Row: anchor.Row + i, Col: anchor.Col + j}
			if s.blocksSpill(anchor, target) {
				return true
			}
		}
	}
	return false
}

// blocksSpill reports whether target (a spill destination other than the anchor
// itself) already holds content and so blocks the spill.
func (s Sheet) blocksSpill(anchor, target Address) boolResult {
	return boolResult(target != anchor && !bool(s.isEmptyCell(target)))
}

// isEmptyCell reports whether a source cell is empty (spillable): out of the
// source grid, or a blank non-formula cell. It reads through at(), the one
// seam both backings share, so spill blocking answers identically for
// resident and windowed sheets.
func (s Sheet) isEmptyCell(at Address) boolResult {
	cl, inGrid := s.at(rowIndex(at.Row), colIndex(at.Col))
	if !inGrid {
		return true
	}
	return boolResult(cl.text == "" && !bool(cl.isFormula()))
}
