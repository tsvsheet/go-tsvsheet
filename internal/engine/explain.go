package engine

import (
	"strings"
	"time"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// Trace explains how one cell was produced: its value, the formula (empty for a
// literal), the resolved value of each cell the formula reads, and — when the
// formula imports — where each import went and whether it succeeded.
// Notes carries the same Excel divergences Check reports, so a reader who has
// not run the checker still learns why a cell reads the way it does.
type Trace struct {
	Cell    string        `json:"cell"`
	Value   string        `json:"value"`
	Formula string        `json:"formula,omitempty"`
	Inputs  []TraceInput  `json:"inputs,omitempty"`
	Imports []TraceImport `json:"imports,omitempty"`
	Notes   []string      `json:"notes,omitempty"`
}

// TraceInput is one reference a formula reads, with its resolved value.
type TraceInput struct {
	Ref   string `json:"ref"`
	Value string `json:"value"`
}

// TraceImport is one IMPORT* call a formula performs: the source exactly as the
// sheet wrote it, the URL that source actually resolved to, and the reason it
// failed. Every import failure is the same opaque #IMPORT! in the grid by
// design, so this is the only place an author can learn WHICH failure it was —
// a denied host and a traversal above the data base look identical in a cell.
type TraceImport struct {
	Source string `json:"source"`
	URL    string `json:"url,omitempty"`
	Error  string `json:"error,omitempty"`
}

// Explain computes the sheet and describes the cell at at: its value, and — when
// the cell is a formula — that formula and each reference it reads. Imports are
// disabled on this path (no Fetcher), so an IMPORT* cell is #IMPORT!; use
// ExplainWith to trace a sheet that imports.
func Explain(s Sheet, at Address) (Trace, error) {
	return explainWith(s, at, func() computer { return newComputer(s, time.Now()) })
}

// ExplainWith is Explain with an injected compute environment, so a sheet whose
// cells embed sub-sheets or import external data traces with those resolved
// rather than erroring.
func ExplainWith(s Sheet, at Address, opts ComputeOptions) (Trace, error) {
	return explainWith(s, at, func() computer { return passComputer(s, opts) })
}

// explainWith builds the trace against the computer newComp returns, which is
// constructed only after the cell is known to exist.
func explainWith(s Sheet, at Address, newComp func() computer) (Trace, error) {
	cl, inGrid := s.at(rowIndex(at.Row), colIndex(at.Col))
	if !inGrid {
		return Trace{}, constants.ErrNotFound.With(nil, "cell", at.String())
	}
	comp := newComp()
	trace := Trace{Cell: at.String(), Value: comp.read(rowIndex(at.Row), colIndex(at.Col)).String()}
	if cl.isFormula() {
		// The trace shows the cell's WHOLE formula, meta clause included: a
		// hover on a named cell that hid its own name would be a partial truth
		// (TestExplainShowsTheMetaClause).
		trace.Formula = renderExpr(cl.formula) + renderMeta(cl.meta)
		trace.Inputs = traceInputs(comp, cl.formula)
		trace.Imports = traceImports(comp, cl.formula)
		trace.Notes = divergenceNotes(cl.formula)
	}
	return trace, nil
}

// divergenceNotes lists the Excel divergences in one formula, in the order they
// appear. It reuses the checker so the two can never drift into telling an
// author different things about the same cell.
func divergenceNotes(expr tsvt.Expr) []string {
	diags := divergenceDiagnostics(expr, Address{})
	if len(diags) == 0 {
		return nil
	}
	notes := make([]string, len(diags))
	for i, diag := range diags {
		notes[i] = diag.Message
	}
	return notes
}

// traceImports describes every IMPORT* call in the formula. It re-asks the
// Fetcher rather than reading the compute pass, which is what keeps the trace
// scoped to THIS cell instead of every import in the sheet; a frontend's
// cross-pass cache makes the second ask free.
func traceImports(comp computer, expr tsvt.Expr) []TraceImport {
	if comp.fetcher == nil {
		return nil
	}
	res := resolver{comp: comp}
	var imports []TraceImport
	walkCalls(expr, func(call tsvt.Call) {
		media, isImport := importMedia[strings.ToLower(call.Name)]
		if isImport && len(call.Args) == 1 {
			imports = append(imports, res.traceImport(call.Args[0], media))
		}
	})
	return imports
}

// traceImport performs one import for a trace and reports where it went.
func (r resolver) traceImport(arg tsvt.Expr, media MediaType) TraceImport {
	source := r.eval(arg).String()
	res, err := r.comp.fetcher.Fetch(ImportURL(source), media)
	if err != nil {
		return TraceImport{Source: source, Error: err.Error()}
	}
	return TraceImport{Source: source, URL: string(res.URL)}
}

// traceInputs renders each input the formula reads — every A1 reference and
// every `@name` — with its computed value, in source order within each kind. A
// name IS an input edge (SPECIFICATION §7), so a trace that showed references
// and hid names would answer "why is this cell wrong" with half the inputs
// (TestExplainNamedInputs states both halves).
func traceInputs(comp computer, expr tsvt.Expr) []TraceInput {
	res := resolver{comp: comp}
	var inputs []TraceInput
	walkRefs(expr, func(ref tsvt.Reference) {
		inputs = append(inputs, TraceInput{Ref: renderReference(ref), Value: traceValue(res.resolveOperand(ref))})
	})
	walkNames(expr, func(ref tsvt.NameRef) {
		inputs = append(inputs, TraceInput{Ref: "@" + string(ref.Name), Value: res.resolveName(ref).String()})
	})
	return inputs
}

// traceValue renders a resolved reference for a trace: a single cell as its
// value, a multi-cell range as its cell values joined with ", " — so a range
// input reads informatively rather than the #VALUE! that scalar() yields for a
// range in a scalar context.
func traceValue(cs cellset) string {
	if cs.isSingle {
		return cs.scalar().String()
	}
	parts := make([]string, len(cs.values))
	for i, v := range cs.values {
		parts[i] = v.String()
	}
	return strings.Join(parts, ", ")
}

// walkRefs visits every reference operand in an expression tree.
func walkRefs(expr tsvt.Expr, visit func(tsvt.Reference)) {
	if operand, ok := expr.(tsvt.RefOperand); ok {
		visit(operand.Ref)
	}
	for _, child := range children(expr) {
		walkRefs(child, visit)
	}
}
