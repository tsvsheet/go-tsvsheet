package engine

import (
	"time"

	"github.com/tsvsheet/go-tsvsheet/internal/tsvt"
)

// computer memoizes cell values as they are evaluated in dependency order. Its
// cache and phase slices are allocated once and shared, so value-receiver
// methods mutate them in place (no reassignment) and every recursive read sees
// the same state. now is the wall clock sampled once for the pass (volatile
// functions).
// rng is the pass PRNG, seeded from now so rand()/randbetween()/randarray()
// re-roll each pass yet reproduce for a fixed clock (there is one engine, R1, so
// its dependency-order draw sequence is the semantics); it is a shared pointer
// because the computer is copied by value. tick is the injected recompute-pass
// ordinal read by tick()/frame().
type computer struct {
	now     time.Time
	memo    memo
	fetcher Fetcher
	rng     passRNG
	env     embedEnv
	sheet   Sheet
	limits  Limits
	tick    Tick
}

// newComputer builds a computer sized to the sheet, with the pass clock and the
// engine's generous DefaultLimits (the plain Compute/ComputeAt path); the
// embedding path (ComputeWith) overrides them with the injected limits.
func newComputer(s Sheet, now time.Time) computer {
	rng := newPassRNG(prngSeed(now.UnixNano()))
	return computer{now: now, rng: rng, sheet: s, memo: newDenseMemo(s), limits: DefaultLimits()}
}

// cellValue is a cell's evaluated Value: a literal parsed, a formula computed
// (which may be a dynamic array that later spills).
func (c computer) cellValue(row rowIndex, col colIndex, cl cell) Value {
	if !cl.isFormula() {
		return value(textVal(cl.text))
	}
	return c.read(row, col)
}

// read returns the value at (row, col), evaluating and memoizing it on first
// visit. A cell already on the evaluation stack is a circular reference; an
// out-of-grid position is #REF!.
func (c computer) read(row rowIndex, col colIndex) Value {
	cl, inGrid := c.sheet.at(row, col)
	if !inGrid {
		return errorValue(ErrRef)
	}
	cached, phase := c.memo.lookup(row, col)
	switch phase {
	case phaseDone:
		return cached
	case phaseVisiting:
		return errorValue(ErrCirc)
	case phaseRefused:
		return errorValue(ErrLimit)
	}
	if !c.memo.admit(row, col) {
		return errorValue(ErrLimit)
	}
	result := c.evalCell(cl).asCellResult()
	c.memo.finish(row, col, result)
	return result
}

// evalCell evaluates one cell: a literal parses to its value, a formula
// evaluates its expression (which reads other cells).
func (c computer) evalCell(cl cell) Value {
	if !cl.isFormula() {
		return value(textVal(cl.text))
	}
	return resolver{comp: c}.eval(cl.formula)
}

// resolver evaluates expressions against the computer, resolving A1 references.
type resolver struct {
	comp computer
}

// cellset is a resolved reference: the referenced cells' values, and whether it
// was a single cell (a range is not single).
type cellset struct {
	values   []Value
	isSingle boolResult
}

// scalar reduces a cellset to one value: a single cell is its value; a
// multi-cell range used where a scalar is required is #VALUE!.
func (c cellset) scalar() Value {
	if c.isSingle && len(c.values) == 1 {
		return c.values[0]
	}
	return errorValue(ErrValue)
}

// resolveOperand resolves a reference operand: a single A1 cell or an A1 range,
// read from another sheet when the reference carries a `"file"!` qualifier.
func (r resolver) resolveOperand(ref tsvt.Reference) cellset {
	rangeRef := ref.(tsvt.RangeRef)
	if rangeRef.File != "" {
		return r.foreignCells(rangeRef)
	}
	if rangeRef.To == nil {
		return r.resolveSingle(rangeRef.From)
	}
	return r.resolveMatrix(rangeRef.From, *rangeRef.To)
}

// resolveSingle resolves a single-cell reference; an unaddressable position
// (`A0`, an over-bound column) is a wholesale #REF! refusal, kept isSingle so
// it propagates through scalar() — and marked, so a lazy consumer reached
// through a dispatcher branch (`countif(if(…, A0, …), …)`) propagates it
// rather than stepping over an unresolved reference (pinned by the
// dispatcher-branch cases in TestCriteria_RefusedRangePropagates).
func (r resolver) resolveSingle(cell tsvt.CellRef) cellset {
	at, ok := a1Address(cell)
	if !ok {
		return cellset{values: []Value{refusalValue(ErrRef)}, isSingle: true}
	}
	return cellset{values: []Value{r.comp.read(rowIndex(at.Row), colIndex(at.Col))}, isSingle: true}
}

// resolveMatrix resolves the rectangular hull of two A1 corners (`A1:B3`). When
// both corners are the same cell (`A1:A1`, directly or after a structural edit
// collapses a span), the result is a single cell — so it reads as that cell's
// value in a scalar context rather than #VALUE!.
func (r resolver) resolveMatrix(from, to tsvt.CellRef) cellset {
	a, aok := a1Address(from)
	b, bok := a1Address(to)
	if !aok || !bok {
		return cellset{values: []Value{refusalValue(ErrRef)}}
	}
	if r.overBudget(a, b) {
		return cellset{values: []Value{refusalValue(ErrLimit)}}
	}
	return cellset{values: r.hull(a, b), isSingle: boolResult(a == b)}
}

// overBudget reports whether the inclusive rectangle spanned by a and b holds
// more cells than the pass's span budget. It is checked from the corners alone
// — before any allocation or read — so a written reference can never drive the
// materialization it names; the refusal reuses the same overflow-safe
// dimension-first bound an array result answers to.
func (r resolver) overBudget(a, b Address) boolResult {
	r0, r1 := ordered(gridPos(a.Row), gridPos(b.Row))
	c0, c1 := ordered(gridPos(a.Col), gridPos(b.Col))
	return boolResult(r.comp.limits.spanTooLarge(resultDim(r1-r0+1), resultDim(c1-c0+1)))
}

// hull reads every cell in the inclusive rectangle spanned by a and b.
func (r resolver) hull(a, b Address) []Value {
	r0, r1 := ordered(gridPos(a.Row), gridPos(b.Row))
	c0, c1 := ordered(gridPos(a.Col), gridPos(b.Col))
	values := make([]Value, 0, int(r1-r0+1)*int(c1-c0+1))
	for row := r0; row <= r1; row++ {
		for col := c0; col <= c1; col++ {
			values = append(values, r.comp.read(rowIndex(row), colIndex(col)))
		}
	}
	return values
}

// rangeCorners resolves a local range's two corner addresses; a single-cell
// reference is its own second corner. isAddressable is false when either corner is
// unaddressable (a row below 1, or a column past the bound) — #REF! at the
// caller.
func rangeCorners(rangeRef tsvt.RangeRef) (from, to Address, isAddressable boolResult) {
	from, fromOK := a1Address(rangeRef.From)
	to, toOK := from, fromOK
	if rangeRef.To != nil {
		to, toOK = a1Address(*rangeRef.To)
	}
	return from, to, fromOK && toOK
}

// a1Address converts an A1 cell to a 0-based (row, col). ok is false for a row
// below 1 (`A0`) and for a column-letter run past the addressable bound (no
// grid can contain it, so it is out of every grid — #REF! at the caller); the
// grammar admits only a column label followed by an integer row, so no other
// shape reaches here.
func a1Address(cell tsvt.CellRef) (Address, boolResult) {
	if cell.Row < 1 {
		return Address{}, false
	}
	col, ok := lettersToIndex(columnLetters(cell.Col))
	if !ok {
		return Address{}, false
	}
	return Address{Row: cell.Row - 1, Col: col}, true
}

// ordered returns its two coordinates low-first.
func ordered(x, y gridPos) (gridPos, gridPos) {
	if x <= y {
		return x, y
	}
	return y, x
}

// argMatrix resolves an argument to a 2-D block of values: a range keeps its
// rows×columns shape (for lookups), and an expression that evaluates to an
// array contributes that shape — consumed exactly like a range, so
// `index(sort(…), 2)` reads the sorted block (ADR 0004 §2 array-valued
// arguments). Any other expression is a 1×1 block. refused is the wholesale
// refusal to propagate when the range (or a nested consumer of one) could not
// be resolved at all — returned out of band precisely so no caller can mistake
// it for a block of one error cell and step over it.
func (r resolver) argMatrix(arg tsvt.Expr) (block [][]Value, refused Value) {
	if ref, isRef := arg.(tsvt.RefOperand); isRef {
		m := r.rangeMatrix(ref.Ref)
		if m[0][0].isRefusal() {
			return nil, m[0][0]
		}
		return m, Value{}
	}
	return r.evalMatrix(arg)
}

// evalMatrix shapes a non-reference argument: an array keeps its shape, a
// refusal propagated out of a nested call (sort over a refused range) stays a
// refusal, anything else is a 1×1 block.
func (r resolver) evalMatrix(arg tsvt.Expr) (block [][]Value, refused Value) {
	v := r.eval(arg)
	if v.isRefusal() {
		return nil, v
	}
	switch v.kind {
	case kindArray:
		return v.arr, Value{}
	default:
		return [][]Value{{v}}, Value{}
	}
}

// rangeMatrix resolves an A1 reference to its rows×columns of values; an
// off-grid endpoint yields a 1×1 #REF! block. A `"file"!` qualifier reads the
// block from another sheet.
func (r resolver) rangeMatrix(ref tsvt.Reference) [][]Value {
	rangeRef := ref.(tsvt.RangeRef)
	if rangeRef.File != "" {
		return r.foreignMatrix(rangeRef)
	}
	from, to, isAddressable := rangeCorners(rangeRef)
	if !isAddressable {
		return [][]Value{{refusalValue(ErrRef)}}
	}
	if r.overBudget(from, to) {
		return [][]Value{{refusalValue(ErrLimit)}}
	}
	r0, r1 := ordered(gridPos(from.Row), gridPos(to.Row))
	c0, c1 := ordered(gridPos(from.Col), gridPos(to.Col))
	rows := make([][]Value, 0, r1-r0+1)
	for row := r0; row <= r1; row++ {
		cols := make([]Value, 0, c1-c0+1)
		for col := c0; col <= c1; col++ {
			cols = append(cols, r.comp.read(rowIndex(row), colIndex(col)))
		}
		rows = append(rows, cols)
	}
	return rows
}
