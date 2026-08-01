// Authored parentheses.
//
// The parser records that an author wrote parentheses (tsvt.Binary.IsGrouped),
// and this is where that record is honoured. It is not a formatting preference:
// grouping is how the author said which reading was meant, and the
// Excel-divergence checker treats a grouped power as the explicit form that
// needs no diagnostic. If a rewrite dropped the parentheses, the checker would
// re-report a divergence the author had already resolved — the tool undoing the
// fix it asked for. So parentheses survive parsing, rewriting, and rendering.
package engine

import "github.com/tsvsheet/go-tsvsheet/internal/tsvt"

// isAuthorGrouped reports whether the author parenthesized this expression.
// Parentheses are re-emitted for that reason alone, even where precedence does
// not require them: `(-A1) + 1` needs none, but the author wrote them to say
// which reading was meant, and a rewrite that dropped them would undo the very
// fix the Excel-divergence checker asks for.
func isAuthorGrouped(expr tsvt.Expr) bool {
	switch e := expr.(type) {
	case tsvt.Binary:
		return e.IsGrouped
	case tsvt.Unary:
		return e.IsGrouped
	default:
		return false
	}
}
