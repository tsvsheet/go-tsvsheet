// Duplicating a row or a column.
//
// A duplicate is a fill of one line onto a new one, plus the grid shift an
// insert performs — so it shares fill.go's rebasing machinery and differs in
// what it does to the grid's shape. Keeping it here makes that difference the
// subject of its own file: fill.go changes cells inside the grid it was given,
// this one changes how many lines the grid has.
package engine

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
