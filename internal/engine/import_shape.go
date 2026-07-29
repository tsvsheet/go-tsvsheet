// Package engine's import shaping: the strict rectangle rules each IMPORT*
// result must satisfy. A shape mismatch is #IMPORT!, never a best-effort
// salvage — a silently reshaped import would put wrong numbers in a grid that
// looks right.
package engine

// shapeImport enforces the shape each import media type requires and returns the
// scalar (cell) or spilling array (row/column/range/sheet) result; a shape or
// size mismatch is #IMPORT!, never a salvage (ADR 0006 §4).
func shapeImport(cells [][]Value, media MediaType, limits Limits) Value {
	switch media {
	case mediaCell:
		return importScalar(cells)
	case mediaRow:
		return importRow(cells, limits)
	case mediaColumn:
		return importColumn(cells, limits)
	default: // mediaRange, mediaSheet
		return importRange(cells, limits)
	}
}

// importScalar shapes IMPORTCELL: the grid must be exactly one row of one cell,
// returned as that scalar value.
func importScalar(cells [][]Value) Value {
	if len(cells) != 1 || len(cells[0]) != 1 {
		return errorValue(ErrImport)
	}
	return cells[0][0]
}

// importRow shapes IMPORTROW: exactly one row (of one or more columns), returned
// as a 1×N array that spills horizontally.
func importRow(cells [][]Value, limits Limits) Value {
	if len(cells) != 1 {
		return errorValue(ErrImport)
	}
	row := cells[0]
	if oversize(limits, 1, resultDim(len(row))) {
		return errorValue(ErrImport)
	}
	return arrayValue([][]Value{row})
}

// importColumn shapes IMPORTCOLUMN: one or more rows, each exactly one cell,
// returned as an N×1 array that spills vertically.
func importColumn(cells [][]Value, limits Limits) Value {
	if !allWidth(cells, 1) {
		return errorValue(ErrImport)
	}
	if oversize(limits, resultDim(len(cells)), 1) {
		return errorValue(ErrImport)
	}
	return arrayValue(cells)
}

// importRange shapes IMPORTRANGE and IMPORTSHEET: a non-empty rectangular grid
// (every row the same width), returned as an R×C array that spills. For this
// engine chunk IMPORTSHEET behaves like IMPORTRANGE (a spilling values grid);
// the "nested grid inside one cell" rendering distinction is deferred to the
// frontend chunk — only the requested Accept media type (the handshake) differs.
func importRange(cells [][]Value, limits Limits) Value {
	width := resultDim(len(cells[0]))
	if !allWidth(cells, width) {
		return errorValue(ErrImport)
	}
	if oversize(limits, resultDim(len(cells)), width) {
		return errorValue(ErrImport)
	}
	return arrayValue(cells)
}

// allWidth reports whether every row of cells has exactly width columns.
func allWidth(cells [][]Value, width resultDim) boolResult {
	for _, row := range cells {
		if resultDim(len(row)) != width {
			return false
		}
	}
	return true
}

// oversize reports whether a rows×cols import result exceeds the injected cell
// budget (ADR 0006 §4) — an over-large response is #IMPORT!.
func oversize(limits Limits, rows, cols resultDim) boolResult {
	return boolResult(limits.tooManyCells(rows, cols))
}
