package engine

import (
	"strings"

	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// Diagnostic is an advisory finding about a sheet: an unknown function call in
// a formula cell (which computes to #NAME?), or a view directive the parser
// rejects. A cell finding carries Cell; a directive finding carries Line, the
// 1-based physical line, because a directive occupies a line and no row.
type Diagnostic struct {
	Cell    string `json:"cell,omitempty"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
	IsFatal bool   `json:"fatal"`
}

// Check reports the static diagnostics of a parsed sheet: each unknown function
// call. Syntax errors are already rejected by Parse, and every reference the
// narrowed grammar admits is a valid A1 form, so Check never reports those.
func Check(s Sheet) []Diagnostic {
	var diags []Diagnostic
	for r, row := range s.rowsView() {
		for c, cl := range row {
			if cl.isFormula() {
				at := Address{Row: r, Col: c}
				diags = append(diags, unknownFunctions(cl.formula, at)...)
				diags = append(diags, divergenceDiagnostics(cl.formula, at)...)
			}
		}
	}
	return append(diags, nameDiagnostics(s)...)
}

// unknownFunctions flags each call to a name outside the builtin set.
func unknownFunctions(expr tsvt.Expr, at Address) []Diagnostic {
	label := at.String()
	var diags []Diagnostic
	walkCalls(expr, func(call tsvt.Call) {
		if !isKnownFunc(funcName(call.Name)) {
			diags = append(diags, Diagnostic{Cell: label, Message: "unknown function: " + call.Name})
		}
	})
	return diags
}

// isClock reports whether name is a volatile clock builtin.
func isClock(name funcName) boolResult {
	return name == fnToday || name == fnNow || name == fnIsnow
}

// lazyNamePredicates are the lazy-dispatch name predicates isKnownFunc
// consults. Every lazy dispatcher of resolver.lazyDispatchers must be
// represented here, or Check flags a name the evaluator computes.
var lazyNamePredicates = []func(funcName) boolResult{
	isConditional, isClock, isVolatileFn, isRandom, isTable, isCriteria, isArray,
	isSeries, isDigest, isText, isEmbed, isImportName,
}

// isKnownFunc reports whether name (case-insensitive) is a builtin: an eager
// registry function, a lazily-dispatched builtin, a value predicate, or the
// meta function — which is known to the language even though it is not
// callable, so a misplaced `named` gets the meta-misuse diagnostic
// (check_names.go) rather than a misleading "unknown function".
func isKnownFunc(name funcName) boolResult {
	lower := funcName(strings.ToLower(string(name)))
	if isMetaFn(lower) {
		return true
	}
	for _, isLazyName := range lazyNamePredicates {
		if isLazyName(lower) {
			return true
		}
	}
	if _, ok := inspectors[string(lower)]; ok {
		return true
	}
	_, ok := functions[string(lower)]
	return boolResult(ok)
}

// walkCalls visits every function call in an expression tree.
func walkCalls(expr tsvt.Expr, visit func(tsvt.Call)) {
	if call, ok := expr.(tsvt.Call); ok {
		visit(call)
	}
	for _, child := range children(expr) {
		walkCalls(child, visit)
	}
}

// children returns the sub-expressions of an expression (empty for a leaf).
func children(expr tsvt.Expr) []tsvt.Expr {
	switch e := expr.(type) {
	case tsvt.Unary:
		return []tsvt.Expr{e.X}
	case tsvt.Percent:
		return []tsvt.Expr{e.X}
	case tsvt.Binary:
		return []tsvt.Expr{e.Left, e.Right}
	case tsvt.Call:
		return e.Args
	default:
		return nil
	}
}
