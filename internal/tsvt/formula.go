package tsvt

import (
	"github.com/antlr4-go/antlr/v4"

	grammar "github.com/tsvsheet/go-tsvsheet/internal/grammar"
)

// FormulaText is the source of a single formula expression — the part of a
// spreadsheet cell after its leading `=`.
type FormulaText string

// ParseFormula parses one formula — the expression, then at most one trailing
// `|@` meta clause (SPECIFICATION §5.6) — into the typed Expr AST and the
// clause, or constants.ErrSyntax. It reuses the ANTLR-generated sublanguage
// via the grammar's `formula` entry rule, whose EOF anchor makes a trailing
// tail a syntax error, so the A1 spreadsheet model compiles each `=formula`
// cell without a hand-written parser. An absent clause is the zero MetaClause.
func ParseFormula(src FormulaText) (Expr, MetaClause, error) {
	collector := &errorCollector{sink: &errorSink{}}

	lexer := grammar.NewTsvsheetLexer(antlr.NewInputStream(string(src)))
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(collector)

	parser := grammar.NewTsvsheetParser(antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel))
	parser.RemoveErrorListeners()
	parser.AddErrorListener(collector)

	formula := parser.Formula()
	if collector.sink.err != nil {
		return nil, MetaClause{}, collector.sink.err
	}
	expr, err := buildExpr(formula.Expression())
	if err != nil {
		return nil, MetaClause{}, err
	}
	return expr, buildMeta(formula.MetaClause()), nil
}
