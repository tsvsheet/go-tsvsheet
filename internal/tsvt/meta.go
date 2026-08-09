package tsvt

import (
	grammar "github.com/tsvsheet/go-tsvsheet/internal/grammar"
)

// MetaClause is a formula's trailing `|@` clause (SPECIFICATION §5.6) — an
// annotation on the CELL, not a stage in the expression: `=sum(A1:A5) |@
// named(Total)`. The grammar admits at most one, always last, and only
// identifiers as arguments, so by the time a clause exists here its shape is
// already right; which functions are meta, and their arity, are semantic and
// judged by the engine's check. The zero value is "no clause".
type MetaClause struct {
	Fn   string
	Args []NameText
	// HasParens records whether the author wrote the clause's parentheses —
	// `|@ named(Total)` against the bare `|@ named` a dropped-parentheses
	// stage permits — so a re-rendered cell keeps the author's spelling, as
	// IsPiped does for the pipe.
	HasParens bool
}

// IsZero reports whether the clause is absent. The grammar guarantees a parsed
// clause names a function, so a present clause always has a non-empty Fn.
func (m MetaClause) IsZero() bool { return m.Fn == "" }

// buildMeta converts the optional meta-clause parse node; a nil node is the
// absent (zero) clause.
func buildMeta(ctx grammar.IMetaClauseContext) MetaClause {
	if ctx == nil {
		return MetaClause{}
	}
	return MetaClause{
		Fn:        ctx.META_IDENT().GetText(),
		Args:      metaArgs(ctx.MetaArgs()),
		HasParens: ctx.META_LPAREN() != nil,
	}
}

// metaArgs collects the clause's argument identifiers; a nil list — a bare
// `|@ named` or an empty `|@ named()` — is no arguments.
func metaArgs(ctx grammar.IMetaArgsContext) []NameText {
	if ctx == nil {
		return nil
	}
	idents := ctx.AllMETA_IDENT()
	args := make([]NameText, len(idents))
	for i, ident := range idents {
		args[i] = NameText(ident.GetText())
	}
	return args
}
