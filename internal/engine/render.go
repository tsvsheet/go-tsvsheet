package engine

import (
	"strconv"
	"strings"

	"github.com/uplang/go-tsvsheet/internal/tsvt"
)

// renderExpr reconstructs a readable source form of an expression, used by
// diagnostics and the explain trace.
func renderExpr(expr tsvt.Expr) string {
	switch e := expr.(type) {
	case tsvt.Number:
		return e.Text
	case tsvt.StringLit:
		return `"` + e.Value + `"`
	case tsvt.BoolLit:
		return renderBool(boolResult(e.IsTrue))
	case tsvt.ErrorLit:
		return e.Code
	case tsvt.RefOperand:
		return renderReference(e.Ref)
	case tsvt.Unary:
		return string(e.Op) + renderExpr(e.X)
	case tsvt.Percent:
		return renderExpr(e.X) + "%"
	case tsvt.Binary:
		return renderExpr(e.Left) + " " + string(e.Op) + " " + renderExpr(e.Right)
	default: // tsvt.Call
		return renderCall(expr.(tsvt.Call))
	}
}

// renderBool reconstructs a boolean literal.
func renderBool(isTrue boolResult) string {
	if isTrue {
		return "TRUE"
	}
	return "FALSE"
}

// renderCall reconstructs a function call.
func renderCall(call tsvt.Call) string {
	args := make([]string, len(call.Args))
	for i, arg := range call.Args {
		args[i] = renderExpr(arg)
	}
	return call.Name + "(" + strings.Join(args, ",") + ")"
}

// renderReference reconstructs an A1 reference: a cell or a two-cell range,
// with its `"file"!` sheet qualifier when present.
func renderReference(ref tsvt.Reference) string {
	rangeRef := ref.(tsvt.RangeRef)
	body := renderCell(rangeRef.From)
	if rangeRef.To != nil {
		body += ":" + renderCell(*rangeRef.To)
	}
	return renderQualifier(Path(rangeRef.File)) + body
}

// renderQualifier reconstructs a `"file"!` sheet qualifier, or "" for the
// current sheet.
func renderQualifier(file Path) string {
	if file == "" {
		return ""
	}
	return `"` + string(file) + `"!`
}

// renderCell reconstructs one A1 cell (`B2`).
func renderCell(cell tsvt.CellRef) string {
	return cell.Col + strconv.Itoa(cell.Row)
}
