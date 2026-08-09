package engine

import (
	"strings"

	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// metaMarker is the byte sequence that opens a meta clause — what the index
// scan marks candidate declaration rows by.
const metaMarker = "|@"

// metaFnNamed is the one meta function (SPECIFICATION §5.6). Which functions
// are meta is semantic, like every function-name question; the identity is
// case-insensitive, like every function name. Deliberately UNTYPED: `named` is
// not a dispatchable funcName, and a typed const would join that enum group,
// obliging every exhaustive funcName switch to answer for a function no
// dispatcher can reach.
const metaFnNamed = "named"

// namedArity is `named`'s exact argument count: the name, required — never
// optional, so the bare `|@ named` a dropped-parentheses stage permits is a
// diagnostic rather than a silent binding.
const namedArity = 1

// normalName is a name's case-insensitive identity (`@Total` ≡ `@TOTAL`),
// mirroring function-name identity. The declared spelling is kept separately
// for display.
type normalName string

// normalizeName folds a name to its identity.
func normalizeName(name tsvt.NameText) normalName {
	return normalName(strings.ToLower(string(name)))
}

// isMetaFn reports whether a function name is `named` — the test that decides
// whether a clause binds, and whether a call strayed into the expression.
func isMetaFn(fn funcName) boolResult {
	return boolResult(strings.EqualFold(string(fn), metaFnNamed))
}

// binding is one name's declaration: where it is bound, and the spelling the
// author used.
type binding struct {
	spelled tsvt.NameText
	at      Address
}

// nameBindings is a sheet's collected name declarations — the statically
// knowable product of every well-formed `|@ named(X)` clause (SPECIFICATION
// §5.6: a sheet's names are knowable without evaluating anything). refused
// marks a windowed source whose declaration candidates exceeded the resident
// budget: every use then answers #LIMIT!, the budget refusal, rather than a
// wrong or partial binding.
type nameBindings struct {
	at        map[normalName]binding
	dupes     map[normalName]boolResult
	isRefused boolResult
}

// refusedBindings is the over-budget state: no bindings, every use #LIMIT!.
func refusedBindings() nameBindings { return nameBindings{isRefused: true} }

// resolve answers one name's binding address. A duplicated name refuses every
// use as #NAME? — a wholesale refusal, never a silently chosen winner — and a
// budget-refused collection refuses every use as #LIMIT!.
func (n nameBindings) resolve(name tsvt.NameText) (Address, ErrorValue) {
	if n.isRefused {
		return Address{}, ErrLimit
	}
	id := normalizeName(name)
	if n.dupes[id] {
		return Address{}, ErrName
	}
	bound, ok := n.at[id]
	if !ok {
		return Address{}, ErrName
	}
	return bound.at, ""
}

// isUnknown reports whether a use of name would compute #NAME? for the one
// reason a reader can fix by declaring it: no cell binds it. A duplicated or
// budget-refused name is NOT unknown — those defects have their own findings,
// and flagging their uses too would bury one defect under its symptoms.
func (n nameBindings) isUnknown(name tsvt.NameText) boolResult {
	id := normalizeName(name)
	_, bound := n.at[id]
	return boolResult(!bool(n.isRefused) && !bool(n.dupes[id]) && !bound)
}

// bind records one declaration, folding a repeat into the duplicate set.
func (n nameBindings) bind(name tsvt.NameText, at Address) nameBindings {
	if n.at == nil {
		n.at = map[normalName]binding{}
		n.dupes = map[normalName]boolResult{}
	}
	id := normalizeName(name)
	if _, taken := n.at[id]; taken {
		n.dupes[id] = true
		return n
	}
	n.at[id] = binding{spelled: name, at: at}
	return n
}

// bindsName reports whether a cell's meta clause is a well-formed declaration:
// the `named` function with its exact arity. A clause that is anything else —
// an unknown meta function, a missing or surplus argument — binds nothing;
// Check names the defect, and any use of the name it failed to bind is #NAME?,
// so a malformed declaration is loud twice rather than quietly half-working.
func bindsName(meta tsvt.MetaClause) boolResult {
	return boolResult(!meta.IsZero() && bool(isMetaFn(funcName(meta.Fn))) && len(meta.Args) == namedArity)
}

// windowedBindings collects a windowed source's declarations from the rows
// the index scan marked, without materializing anything else: each candidate
// row is one bounded read through the block cache, and a false candidate (the
// marker's bytes inside a string literal) parses to no clause and binds
// nothing (TestNames_WindowedMarkerInsideAStringBindsNothing). A source whose
// candidates overflowed the cap answers the refused state — every use then
// computes #LIMIT!, the budget refusal, raisable like every budget
// (TestNames_WindowedMarkOverflowRefusesAsLimit) — and a read failure fails
// the call rather than answering from partial bindings
// (TestNames_WindowedBindingReadFailureIsAnError).
func windowedBindings(src sheetSource) (nameBindings, error) {
	if src.ix.MarkedOverflowed() {
		return refusedBindings(), nil
	}
	var names nameBindings
	for _, row := range src.ix.Marked() {
		cells, err := src.reader.ReadRows(row, 1)
		if err != nil {
			return nameBindings{}, err
		}
		names = bindRow(names, cells, rowIndex(row))
	}
	return names, nil
}

// bindRow folds one fetched row's declarations into the bindings. The row is
// a census-marked grid row, so the fetch returned it and rows[0] is that row
// (the index's own ReadRows tests pin the in-grid contract).
func bindRow(names nameBindings, cells [][]cell, row rowIndex) nameBindings {
	for c, cl := range cells[0] {
		if bindsName(cl.meta) {
			names = names.bind(cl.meta.Args[0], Address{Row: int(row), Col: c})
		}
	}
	return names
}

// Names lists the sheet's declared value names in binding order (row-major),
// each in its declared spelling — what an editor's @-completion offers. A
// duplicated name appears once, first spelling; the duplicate itself is
// Check's finding, not this list's.
func (s Sheet) Names() []string {
	bindings := collectBindings(s.rowsView())
	names := make([]string, 0, len(bindings.at))
	for _, id := range sortedNames(bindings.at) {
		names = append(names, string(bindings.at[id].spelled))
	}
	return names
}

// collectBindings walks a resident sheet's cells for declarations, in
// row-major order. One pass per computer, the same order of work as the
// compute pass that follows it.
func collectBindings(cells [][]cell) nameBindings {
	var names nameBindings
	for r, row := range cells {
		for c, cl := range row {
			if bindsName(cl.meta) {
				names = names.bind(cl.meta.Args[0], Address{Row: r, Col: c})
			}
		}
	}
	return names
}
