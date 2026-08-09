// Package engine is the tsvsheet spreadsheet engine: a .tsvt file IS the sheet —
// a TAB-separated grid whose cells are literal values or `=formulas`, and whose
// formulas address other cells by A1 reference (`B2`, `D2:D4`) exactly like a
// conventional spreadsheet. It parses the grid, evaluates every formula in
// dependency order (memoized, with cycle detection), and emits the computed
// grid. The expression sublanguage and its AST come from internal/tsvt (the
// ANTLR-generated parser); this package resolves A1 references and evaluates.
//
// This is the implementation behind the thin public github.com/tsvsheet/go-tsvsheet
// facade, which re-exports engine's surface unchanged; callers use that package.
package engine

import (
	"bytes"
	"strings"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/index"
	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// cell is one spreadsheet cell: a verbatim literal, or a compiled formula with
// its optional meta clause (SPECIFICATION §5.6). text is the verbatim source —
// clause included — so serialization round-trips without re-rendering. The
// clause is the cell's identity, not its content: structural edits carry it
// with the cell, while fill, paste, and duplication copy the expression alone
// (the 023 ruling — copying a labelled thing does not copy the label).
type cell struct {
	formula tsvt.Expr
	text    string
	meta    tsvt.MetaClause
}

// isFormula reports whether the cell holds a formula.
func (c cell) isFormula() boolResult { return boolResult(c.formula != nil) }

// CellInfo describes one non-empty cell: its address, source text, and whether
// it is a formula — the projection the parse command emits.
type CellInfo struct {
	Text      string
	Address   Address
	IsFormula bool
}

// Sheet is a parsed spreadsheet grid of literal and formula cells.
type Sheet struct {
	lazy  *lazyCells // non-nil only inside a windowed evaluation (spec 016 6b)
	cells [][]cell
}

// formulaMarker is the leading byte that makes a cell a formula.
const formulaMarker = "="

// Parse reads a .tsvt grid: each TAB-separated field is a literal, or — when it
// begins with `=` — a formula compiled from the expression that follows. A
// malformed formula is a syntax error naming its cell. The bytes are indexed
// and read through the one materialization path every source shares (spec
// 016).
func Parse(src []byte) (Sheet, error) {
	source, err := newSheetSource(readerAtSize{
		at:   bytes.NewReader(src),
		size: index.SourceSize(len(src)),
	}, allCells)
	if err != nil {
		return Sheet{}, err
	}
	return materializeSheet(source)
}

// ParseWith is Parse under a residency budget (spec 018): the census scan —
// O(index) memory, no cell touched — refuses a document whose cell count
// exceeds Limits.ResidentCells (single-ceiling fallback like its siblings)
// with ErrDocTooLarge BEFORE anything materializes, so no caller pays an
// unbounded allocation it did not raise its budget to accept. An in-budget
// document parses exactly as Parse does. Parse itself stays unbounded for
// embedders that pre-vet their sources; every ecosystem load path routes
// through the bounded forms.
func ParseWith(src []byte, limits Limits) (Sheet, error) {
	if err := censusGate(src, limits); err != nil {
		return Sheet{}, err
	}
	return Parse(src)
}

// censusGate scans src's census and refuses an over-resident document. The
// scan reads bytes, not cells: an over-budget document with a malformed
// formula still refuses as too large — proof the gate precedes
// materialization, stated by
// TestParseWithRefusesOverBudgetBeforeAnyCellParses.
func censusGate(src []byte, limits Limits) error {
	ix, err := index.Scan(bytes.NewReader(src), index.SourceSize(len(src)), indexOptions())
	if err != nil {
		return readFailure(err)
	}
	budget := effectiveLimits(limits).residentBudget()
	if cells := ix.Census().Cells; cellBudget(cells) > budget {
		return constants.ErrDocTooLarge.With(nil, "cells", int64(cells), "budget", int64(budget))
	}
	return nil
}

// parseCell classifies a field as a literal or compiles its formula, naming the
// cell (in A1 notation) on a formula syntax error.
func parseCell(text textVal, row rowIndex, col colIndex) (cell, error) {
	if !strings.HasPrefix(string(text), formulaMarker) {
		return cell{text: string(text)}, nil
	}
	expr, meta, err := tsvt.ParseFormula(tsvt.FormulaText(text[1:]))
	if err != nil {
		at := Address{Row: int(row), Col: int(col)}
		return cell{}, constants.ErrSyntax.With(err, "cell", at.String())
	}
	return cell{text: string(text), formula: expr, meta: meta}, nil
}

// The storage seam (spec 016 area 3): every read of the cell grid goes through
// these accessors — height, widest, rowsView for whole-grid walks, at for one
// cell — and every edit path materializes through grid. The accessors are what
// the indexed backing replaces in area 4; the walks and grid are the surfaces
// the overlay takes over in area 5. Nothing outside this seam may touch
// s.cells.

// height is the grid's row count. The lazy backing keeps at() as its only
// read seam: whole-grid walks and the edit machinery — height's and widest's
// callers — are foreclosed on windowed documents by the load policy, so no
// lazy branch exists here to go stale.
func (s Sheet) height() int { return len(s.cells) }

// widest is the column count of the widest row.
func (s Sheet) widest() int { return widestRow(s.cells) }

// rowsView is the whole grid for a read-only walk (compute, check, source
// serialization); mutation belongs to the grid accessor below.
func (s Sheet) rowsView() [][]cell { return s.cells }

// grid is the materialized edit surface: the current cells, handed to the
// structural and fill/paste machinery that rebuilds a new grid from them.
func (s Sheet) grid() [][]cell { return s.cells }

// at returns the cell at (row, col); the boolean reports whether the position
// is within the grid.
func (s Sheet) at(row rowIndex, col colIndex) (cell, boolResult) {
	if s.lazy != nil {
		return s.lazy.at(row, col)
	}
	if row < 0 || int(row) >= len(s.cells) || col < 0 || int(col) >= len(s.cells[row]) {
		return cell{}, false
	}
	return s.cells[row][col], true
}

// Cells returns every non-empty cell of the sheet as CellInfo, in row-major
// order.
func (s Sheet) Cells() []CellInfo {
	var out []CellInfo
	for r, row := range s.rowsView() {
		for c, cl := range row {
			if cl.text == "" {
				continue
			}
			out = append(out, CellInfo{
				Address:   Address{Row: r, Col: c},
				Text:      cl.text,
				IsFormula: bool(cl.isFormula()),
			})
		}
	}
	return out
}

// Source returns the sheet's cell source texts (literals and "=formulas") as a
// grid — what an editor shows and what is saved back to the .tsvt file.
func (s Sheet) Source() Grid {
	out := make(Grid, s.height())
	for r, row := range s.rowsView() {
		out[r] = make([]string, len(row))
		for c, cl := range row {
			out[r][c] = cl.text
		}
	}
	return out
}

// Set returns a new sheet with the cell at addr replaced by text (a literal or
// a formula), growing the grid to reach an out-of-bounds position. The injected
// limits bound how far the grid may grow. A malformed formula is a syntax error
// and the sheet is unchanged (Set is immutable, so the caller simply keeps the
// old value).
func (s Sheet) Set(addr Address, text string, limits Limits) (Sheet, error) {
	if addr.Row < 0 || addr.Col < 0 {
		return Sheet{}, constants.ErrInvalidValue.With(nil, "address", addr.String())
	}
	if addr.Row >= limits.GridDim || addr.Col >= limits.GridDim {
		return Sheet{}, constants.ErrInvalidValue.With(nil, "address exceeds the grid limit", addr.String())
	}
	if WouldStartCommentLine(addr, CellText(text)) {
		return Sheet{}, constants.ErrCommentCell.With(nil, "cell", addr.String())
	}
	parsed, err := parseCell(textVal(text), rowIndex(addr.Row), colIndex(addr.Col))
	if err != nil {
		return Sheet{}, err
	}
	cells := growCells(s.grid(), addr)
	cells[addr.Row][addr.Col] = parsed
	return Sheet{cells: cells}, nil
}

// growCells deep-copies the cell grid, extending it with empty rows and — in
// the target row only — empty cells, so at is addressable.
func growCells(src [][]cell, at Address) [][]cell {
	out := make([][]cell, max(len(src), at.Row+1))
	for r := range src {
		out[r] = append([]cell(nil), src[r]...)
	}
	for len(out[at.Row]) <= at.Col {
		out[at.Row] = append(out[at.Row], cell{})
	}
	return out
}
