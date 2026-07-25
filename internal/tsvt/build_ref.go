package tsvt

import (
	grammar "github.com/tsvsheet/go-tsvsheet/internal/grammar"
)

// buildReference builds an A1 reference: a single cell or a two-cell range,
// optionally qualified by a "file"! sheet path.
func buildReference(ctx grammar.IReferenceContext) (Reference, error) {
	file := buildQualifier(ctx.SheetQualifier())
	cells := ctx.AllCellRef()
	from, err := buildCellRef(cells[0])
	if err != nil {
		return nil, err
	}
	if len(cells) == 1 {
		return RangeRef{From: from, File: file}, nil
	}
	to, err := buildCellRef(cells[1])
	if err != nil {
		return nil, err
	}
	return RangeRef{From: from, To: &to, File: file}, nil
}

// buildQualifier extracts the sheet path from a `"file"!` qualifier, or "" when
// the reference has none (the current sheet).
func buildQualifier(ctx grammar.ISheetQualifierContext) string {
	if ctx == nil {
		return ""
	}
	return unquote(quoted(ctx.STRING().GetText()))
}

// buildCellRef builds one A1 cell (column letters + 1-based row), retaining
// its `$` absolute markers (SPECIFICATION §4).
func buildCellRef(ctx grammar.ICellRefContext) (CellRef, error) {
	row, err := intToken(ctx.NUMBER())
	if err != nil {
		return CellRef{}, err
	}
	isColAbsolute, isRowAbsolute := pins(ctx)
	return CellRef{
		Col: ctx.COL().GetText(), Row: row,
		IsColAbsolute: isColAbsolute, IsRowAbsolute: isRowAbsolute,
	}, nil
}

// pins reads a cell's `$` markers (cellRef: DOLLAR? COL DOLLAR? NUMBER): a
// dollar before the column letters pins the column, one after pins the row.
func pins(ctx grammar.ICellRefContext) (isColAbsolute, isRowAbsolute bool) {
	col := ctx.COL().GetSymbol().GetTokenIndex()
	for _, dollar := range ctx.AllDOLLAR() {
		if dollar.GetSymbol().GetTokenIndex() < col {
			isColAbsolute = true
		} else {
			isRowAbsolute = true
		}
	}
	return isColAbsolute, isRowAbsolute
}
