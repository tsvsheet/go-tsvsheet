package engine

import (
	"regexp"
	"sort"

	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// The §5.6 static verdicts: everything about a sheet's names is knowable
// without evaluating anything, so everything below is decided here, at check
// time — never left to surface as an error value mid-render. The verdict table
// is work order 023's; each finding cites its cell.

// cellShapedName matches a declared spelling that reads as a canonical A1
// reference (`A1`, `RATE2`). Such a name is legal — the sigil keeps `@A1`
// unambiguous — but invites misreading, so it earns an advisory. Lowercase
// spellings (`rate2`) deliberately do not match: they collide with nothing.
var cellShapedName = regexp.MustCompile(`^[A-Z]+[0-9]+$`)

// diagText is one finding's human-readable message.
type diagText string

// fatalAt is a fatal finding at a cell; adviseAt an advisory one.
func fatalAt(at Address, message diagText) Diagnostic {
	return Diagnostic{Cell: at.String(), Message: string(message), IsFatal: true}
}

// adviseAt is an advisory finding at a cell.
func adviseAt(at Address, message diagText) Diagnostic {
	return Diagnostic{Cell: at.String(), Message: string(message)}
}

// nameDiagnostics reports every naming verdict for a sheet: the fatal
// findings (a duplicated name, a malformed or unknown meta clause, a meta
// function written as an ordinary call) and the advisories (an unbound
// reference, an unused or cell-shaped name).
func nameDiagnostics(s Sheet) []Diagnostic {
	cells := s.rowsView()
	diags := clauseDiagnostics(cells)
	diags = append(diags, duplicateDiagnostics(cells)...)
	return append(diags, referenceDiagnostics(cells)...)
}

// clauseDiagnostics judges each cell's meta clause and each meta function
// written where a value function belongs.
func clauseDiagnostics(cells [][]cell) []Diagnostic {
	var diags []Diagnostic
	for r, row := range cells {
		for c, cl := range row {
			diags = append(diags, cellClauseDiagnostics(cl, Address{Row: r, Col: c})...)
		}
	}
	return diags
}

// cellClauseDiagnostics judges one cell: its clause's function and arity, the
// advisory shape of the name it binds, and any `named` misplaced in the
// expression as an ordinary call or pipe stage.
func cellClauseDiagnostics(cl cell, at Address) []Diagnostic {
	diags := misplacedMetaDiagnostics(cl, at)
	switch {
	case cl.meta.IsZero():
	case !bool(isMetaFn(funcName(cl.meta.Fn))):
		message := diagText("unknown meta function: " + cl.meta.Fn + " (the one meta function is named)")
		diags = append(diags, fatalAt(at, message))
	case len(cl.meta.Args) != namedArity:
		diags = append(diags, fatalAt(at, "named takes exactly one argument, the name"))
	case cellShapedName.MatchString(string(cl.meta.Args[0])):
		message := diagText(
			"name " + string(cl.meta.Args[0]) + " reads as a cell reference; consider a descriptive name",
		)
		diags = append(diags, adviseAt(at, message))
	}
	return diags
}

// misplacedMetaDiagnostics flags `named` written inside the expression — an
// ordinary call or pipe stage, where no meta function is a value function —
// with the spelling that was wanted (TestCheckNames_MisplacedNamedIsFatal).
func misplacedMetaDiagnostics(cl cell, at Address) []Diagnostic {
	if !cl.isFormula() {
		return nil
	}
	var diags []Diagnostic
	walkCalls(cl.formula, func(call tsvt.Call) {
		if isMetaFn(funcName(call.Name)) {
			diags = append(diags, fatalAt(at,
				"named is a meta function: write `|@ named(…)` after the expression, not a call within it"))
		}
	})
	return diags
}

// duplicateDiagnostics flags every cell after the first that binds an
// already-bound name, each at its own cell, citing where the name was first
// bound — no second binding silently wins
// (TestCheckNames_DuplicateBindingIsFatal, and the compute half is
// TestNames_DuplicateRefusesEveryUse).
func duplicateDiagnostics(cells [][]cell) []Diagnostic {
	first := map[normalName]Address{}
	var diags []Diagnostic
	for r, row := range cells {
		for c, cl := range row {
			if d, isDuplicate := duplicateAt(first, cl, Address{Row: r, Col: c}); isDuplicate {
				diags = append(diags, d)
			}
		}
	}
	return diags
}

// duplicateAt records a cell's binding as the first of its name, or reports
// the repeat: the finding is at the repeating cell and cites the first.
func duplicateAt(first map[normalName]Address, cl cell, at Address) (Diagnostic, bool) {
	if !bindsName(cl.meta) {
		return Diagnostic{}, false
	}
	id := normalizeName(cl.meta.Args[0])
	bound, taken := first[id]
	if !taken {
		first[id] = at
		return Diagnostic{}, false
	}
	return fatalAt(at, diagText("name "+string(cl.meta.Args[0])+" is already bound at "+bound.String())), true
}

// referenceDiagnostics answers the two questions that need both sides of the
// binding relation: a `@name` no cell binds (which computes #NAME?, flagged
// like an unknown function), and a bound name no formula reads (advisory).
func referenceDiagnostics(cells [][]cell) []Diagnostic {
	names := collectBindings(cells)
	used := map[normalName]boolResult{}
	var diags []Diagnostic
	for r, row := range cells {
		for c, cl := range row {
			diags = append(diags, unknownNameDiagnostics(names, used, cl, Address{Row: r, Col: c})...)
		}
	}
	return append(diags, unusedDiagnostics(names, used)...)
}

// unknownNameDiagnostics folds one cell's name references into the used set
// and reports each reference no cell binds.
func unknownNameDiagnostics(names nameBindings, used map[normalName]boolResult, cl cell, at Address) []Diagnostic {
	var diags []Diagnostic
	for _, name := range nameRefs(cl) {
		used[normalizeName(name)] = true
		if names.isUnknown(name) {
			diags = append(diags, adviseAt(at, diagText("unknown name: @"+string(name))))
		}
	}
	return diags
}

// unusedDiagnostics flags each bound name with no reference to it, at its
// binding cell, in binding order (TestCheckNames_UnusedNameIsAdvisory).
func unusedDiagnostics(names nameBindings, used map[normalName]boolResult) []Diagnostic {
	var diags []Diagnostic
	for _, id := range sortedNames(names.at) {
		if bound := names.at[id]; !used[id] {
			diags = append(
				diags,
				adviseAt(bound.at, diagText("name "+string(bound.spelled)+" is bound but never used")),
			)
		}
	}
	return diags
}

// sortedNames orders the bound identities by their binding cell, row-major:
// map iteration order is randomized per run, so the findings are sorted
// before a reader sees them.
func sortedNames(bound map[normalName]binding) []normalName {
	ids := make([]normalName, 0, len(bound))
	for id := range bound {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		a, b := bound[ids[i]].at, bound[ids[j]].at
		if a.Row != b.Row {
			return a.Row < b.Row
		}
		return a.Col < b.Col
	})
	return ids
}

// nameRefs collects the `@name` references of one cell's formula, in source
// order.
func nameRefs(cl cell) []tsvt.NameText {
	if !cl.isFormula() {
		return nil
	}
	var names []tsvt.NameText
	walkNames(cl.formula, func(ref tsvt.NameRef) { names = append(names, ref.Name) })
	return names
}

// walkNames visits every name reference in an expression tree.
func walkNames(expr tsvt.Expr, visit func(tsvt.NameRef)) {
	if ref, ok := expr.(tsvt.NameRef); ok {
		visit(ref)
	}
	for _, child := range children(expr) {
		walkNames(child, visit)
	}
}
