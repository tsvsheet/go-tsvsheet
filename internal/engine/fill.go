// Fill and duplicate: the copy-time half of the reference model (spec 007).
// Fill copies one cell across a target span with Excel fill semantics — each
// unpinned reference coordinate shifts by the target's offset from the source,
// `$`-pinned coordinates hold — and DuplicateRow/DuplicateCol compose the
// existing insert with a one-line fill, so a whole line duplicates with its
// references rebased and the rest of the grid shifted exactly as inserts
// already shift it. Like the structural quartet (structural.go), everything is
// an AST transform re-serialized through the canonical renderer.
package engine

import "github.com/tsvsheet/go-tsvsheet/internal/tsvt"

// offset is a fill displacement: how many rows down and columns right a target
// sits from its source (either may be negative).
type offset struct {
	rows int
	cols int
}

// normalized returns the span with From at the top-left and To at the
// bottom-right, so a fill target's corners may be given in any order (the same
// tolerance Span.contains applies).
func normalized(sp Span) Span {
	lo := Address{Row: min(sp.From.Row, sp.To.Row), Col: min(sp.From.Col, sp.To.Col)}
	hi := Address{Row: max(sp.From.Row, sp.To.Row), Col: max(sp.From.Col, sp.To.Col)}
	return Span{From: lo, To: hi}
}

// Fill returns a new sheet with the cell at from copied into every cell of to:
// a literal copies verbatim, a formula rebases each reference by the target's
// offset (pinned coordinates hold; a coordinate rebased off the grid renders
// as #REF!; cross-sheet references copy unshifted). The target equal to from
// is skipped, so a span containing the source fills around it. Targets beyond
// the grid grow it, as Set does. A negative from or span corner is a no-op,
// mirroring the structural quartet; a from beyond the grid is an empty source
// and clears its targets.
func (s Sheet) Fill(from Address, to Span) Sheet {
	span := normalized(to)
	if from.Row < 0 || from.Col < 0 || span.From.Row < 0 || span.From.Col < 0 {
		return s
	}
	source := lineCell(s.cells, from)
	cells := mapRows(s.cells, cloneRow)
	for r := span.From.Row; r <= span.To.Row; r++ {
		for c := span.From.Col; c <= span.To.Col; c++ {
			cells = fillTarget(cells, source, from, Address{Row: r, Col: c})
		}
	}
	return Sheet{cells: cells}
}

// fillTarget places the source cell rebased to one target, skipping the source
// position itself.
func fillTarget(cells [][]cell, source cell, from, at Address) [][]cell {
	if at == from {
		return cells
	}
	d := offset{rows: at.Row - from.Row, cols: at.Col - from.Col}
	return place(cells, at, rebaseCell(source, d))
}

// DuplicateRow returns a new sheet with row at.Row duplicated below itself:
// the existing InsertRow shifts every reference past the new line, then the
// source row fills the blank line with its references rebased one row down
// (pins hold). The duplicate keeps the source row's exact width. An
// out-of-range row is a no-op. Only the row coordinate of at is used.
func (s Sheet) DuplicateRow(at Address) Sheet {
	if at.Row < 0 || at.Row >= len(s.cells) {
		return s
	}
	cells := mapRows(s.InsertRow(Address{Row: at.Row + 1}).cells, cloneRow)
	cells[at.Row+1] = rebaseLine(cells[at.Row], offset{rows: 1})
	return Sheet{cells: cells}
}

// DuplicateCol returns a new sheet with column at.Col duplicated to its right:
// the existing InsertCol shifts every reference past the new column, then each
// row that has a source cell fills its inserted blank with the source rebased
// one column right (pins hold). Rows too short to reach the column stay
// untouched, mirroring InsertCol. An out-of-range column is a no-op. Only the
// column coordinate of at is used.
func (s Sheet) DuplicateCol(at Address) Sheet {
	if at.Col < 0 || at.Col >= widestRow(s.cells) {
		return s
	}
	return Sheet{cells: mapRows(s.InsertCol(Address{Col: at.Col + 1}).cells, func(row []cell) []cell {
		return duplicateInto(cloneRow(row), colIndex(at.Col))
	})}
}

// duplicateInto fills the blank cell inserted after col from the cell at col,
// when the row reaches that far.
func duplicateInto(row []cell, col colIndex) []cell {
	if int(col)+1 >= len(row) {
		return row
	}
	row[col+1] = rebaseCell(row[col], offset{cols: 1})
	return row
}

// rebaseLine rebases every cell of a row by d, keeping the row's exact width.
func rebaseLine(row []cell, d offset) []cell {
	out := make([]cell, len(row))
	for i, cl := range row {
		out[i] = rebaseCell(cl, d)
	}
	return out
}

// rebaseCell rebases one cell by d: a literal (or empty) cell copies verbatim;
// a formula's references shift and the cell re-serializes canonically.
func rebaseCell(cl cell, d offset) cell {
	if !cl.isFormula() {
		return cl
	}
	rebased := mapRefs(cl.formula, func(ref tsvt.Reference) tsvt.Expr {
		return rebaseReference(ref, d)
	})
	return cell{formula: rebased, text: formulaMarker + renderExpr(rebased)}
}

// rebaseReference rebases one A1 reference by d, returning #REF! when any
// unpinned coordinate leaves the grid. A cross-sheet reference addresses
// another sheet and copies unshifted (the structural-edit ruling).
func rebaseReference(ref tsvt.Reference, d offset) tsvt.Expr {
	rangeRef := ref.(tsvt.RangeRef)
	if rangeRef.File != "" {
		return tsvt.RefOperand{Ref: rangeRef}
	}
	from, fromOK := rebaseCellRef(rangeRef.From, d)
	if rangeRef.To == nil {
		return rebasedOperand(tsvt.RangeRef{From: from}, fromOK)
	}
	to, toOK := rebaseCellRef(*rangeRef.To, d)
	return rebasedOperand(tsvt.RangeRef{From: from, To: &to}, fromOK && toOK)
}

// rebasedOperand wraps a rebased range as an operand, or #REF! when a
// coordinate left the grid.
func rebasedOperand(ref tsvt.RangeRef, isOnGrid boolResult) tsvt.Expr {
	if !isOnGrid {
		return refError()
	}
	return tsvt.RefOperand{Ref: ref}
}

// rebaseCellRef shifts one cell's unpinned coordinates by d; ok is false when
// a shifted coordinate leaves the grid (row < 1 or column before A).
func rebaseCellRef(c tsvt.CellRef, d offset) (tsvt.CellRef, boolResult) {
	row, rowOK := rebasedRow(c, d)
	col, colOK := rebasedCol(c, d)
	c.Row, c.Col = row, col
	return c, rowOK && colOK
}

// rebasedRow is the cell's row after the shift — unchanged when pinned.
func rebasedRow(c tsvt.CellRef, d offset) (int, boolResult) {
	if c.IsRowAbsolute {
		return c.Row, true
	}
	return c.Row + d.rows, boolResult(c.Row+d.rows >= 1)
}

// rebasedCol is the cell's column letters after the shift — unchanged when
// pinned.
func rebasedCol(c tsvt.CellRef, d offset) (string, boolResult) {
	if c.IsColAbsolute {
		return c.Col, true
	}
	idx := lettersToIndex(columnLetters(c.Col)) + d.cols
	if idx < 0 {
		return c.Col, false
	}
	return indexToLetters(colIndex(idx)), true
}

// cloneRow copies one row, so a fill or duplicate never aliases the source
// grid.
func cloneRow(row []cell) []cell { return append([]cell(nil), row...) }

// lineCell reads the cell at a non-negative address, empty when the grid does
// not reach it.
func lineCell(cells [][]cell, at Address) cell {
	if at.Row >= len(cells) || at.Col >= len(cells[at.Row]) {
		return cell{}
	}
	return cells[at.Row][at.Col]
}

// place grows the (already copied) grid so at is addressable and sets the cell
// there — rows and the target row's width extend minimally, as Set's growth
// does.
func place(cells [][]cell, at Address, cl cell) [][]cell {
	for len(cells) <= at.Row {
		cells = append(cells, nil)
	}
	for len(cells[at.Row]) <= at.Col {
		cells[at.Row] = append(cells[at.Row], cell{})
	}
	cells[at.Row][at.Col] = cl
	return cells
}
