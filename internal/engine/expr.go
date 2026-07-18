package engine

import "github.com/tsvsheet/go-tsvsheet/internal/tsvt"

// Expr is one compiled bare expression — the text that would follow `=` in a
// formula cell — detached from any sheet. It is an immutable value: compile
// once with CompileExpr, then evaluate against any number of grids, including
// concurrently, without re-parsing.
type Expr struct {
	expr tsvt.Expr
}

// CompileExpr parses and compiles one bare expression (no leading `=`). A
// malformed expression is constants.ErrSyntax carrying line/column detail.
// Compilation is grid-independent; the result evaluates against any Grid.
func CompileExpr(src []byte) (Expr, error) {
	expr, err := tsvt.ParseFormula(tsvt.FormulaText(src))
	if err != nil {
		return Expr{}, err
	}
	return Expr{expr: expr}, nil
}

// Eval evaluates the expression against g with the semantics of a formula cell
// in a sheet over that grid: A1 references resolve into g with the same
// literal coercion, range, dynamic-array, and error-value semantics; an
// out-of-grid reference is #REF!; volatile functions read opts.At; opts.Limits
// bounds allocations; and opts.Loader / opts.Fetcher gate SHEET and IMPORT*
// exactly as in a compute pass — Eval and ComputeWith share the pass computer,
// so they cannot diverge. Evaluation failures are error values, never Go
// errors.
func (e Expr) Eval(g Grid, opts ComputeOptions) Value {
	return resolver{comp: passComputer(literalSheet(g), opts)}.eval(e.expr)
}

// literalSheet wraps a value grid as a sheet of literal cells — the sheet a
// bare expression resolves its references into.
func literalSheet(g Grid) Sheet {
	cells := make([][]cell, len(g))
	for r, row := range g {
		cells[r] = make([]cell, len(row))
		for c, text := range row {
			cells[r][c] = cell{text: text}
		}
	}
	return Sheet{cells: cells}
}

// FormatValue renders v as its canonical computed-cell text — byte-identical
// to what WriteTSV emits for that value in a computed grid. A 2-D array value
// reduces to its scalar-context (top-left) value before formatting.
func FormatValue(v Value) string { return v.String() }
