// Package tsvsheet's expression facade: one compiled bare expression, detached
// from any sheet.
package tsvsheet

import "github.com/tsvsheet/go-tsvsheet/internal/engine"

// Expr is one compiled bare expression — the text that would follow `=` in a
// formula cell — detached from any sheet: compile once with CompileExpr, then
// evaluate against any number of grids, including concurrently.
type Expr = engine.Expr

// CompileExpr parses and compiles one bare expression — the text that would
// follow `=` in a formula cell. A malformed expression is ErrSyntax carrying
// line/column detail via With. The compiled Expr is an immutable value, safe
// for concurrent reuse; its Eval(g, opts) evaluates against a Grid with the
// exact semantics of a formula cell in a sheet over that grid — reference
// resolution, literal coercion, ranges, dynamic arrays, error-value
// propagation, volatile functions from opts.At, Limits enforcement, and
// Loader/Fetcher gating — returning error values, never Go errors.
func CompileExpr(src []byte) (Expr, error) { return engine.CompileExpr(src) }
