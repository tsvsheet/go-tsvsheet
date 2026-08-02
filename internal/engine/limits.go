package engine

import "math"

// Resource limits guard against out-of-memory from untrusted formula or edit
// input: no single cell may drive an unbounded array, string, or grid
// allocation. They are injected into every compute pass (ComputeOptions.Limits)
// and every edit (Sheet.Set), never held in a mutable global — so a frontend
// chooses its ceilings explicitly and concurrent passes never share state: the
// CLI keeps DefaultLimits (or honors --max-cells), the WASM build applies the
// smaller BrowserLimits.

// Limits bounds the sizes an untrusted sheet may drive an allocation to.
type Limits struct {
	ResultCells int // cells in one array formula result (e.g. SEQUENCE)
	GridDim     int // the highest row or column index the grid may grow to (Set)
	ResultBytes int // bytes in one string formula result (e.g. REPT)
	SpanCells   int // cells one written reference's rectangle may cover; 0 falls back to ResultCells
}

// DefaultLimits are generous for real spreadsheets while still bounding OOM.
func DefaultLimits() Limits {
	return Limits{ResultCells: 5_000_000, GridDim: 1_000_000, ResultBytes: 1 << 20, SpanCells: 5_000_000}
}

// BrowserLimits are the tighter ceilings the WASM build applies, sized for a
// browser tab rather than a workstation. SpanCells sits well above ResultCells
// deliberately: reading a large in-grid range (a whole-column SUM) is ordinary
// spreadsheet use the grid dimension already permits, while a large array
// RESULT (a SEQUENCE spill) writes that many new cells — the two budgets bound
// different costs and must not share a ceiling.
func BrowserLimits() Limits {
	return Limits{ResultCells: 100_000, GridDim: 20_000, ResultBytes: 64 << 10, SpanCells: 1_000_000}
}

// cellBudget is a ceiling on the cells a rectangle may cover, in whichever
// role (result or span) it is enforced.
type cellBudget int64

// spanBudget is the cell budget for one written reference's rectangle: the
// dedicated SpanCells when positive, else ResultCells — so a caller
// constructing Limits before SpanCells existed (a --max-cells flag filling the
// other three fields) keeps its intended single ceiling. A zero or negative
// SpanCells is "unset", never a refuse-everything budget.
func (l Limits) spanBudget() cellBudget {
	if l.SpanCells > 0 {
		return cellBudget(l.SpanCells)
	}
	return cellBudget(l.ResultCells)
}

// resultDim is a row or column count of an array formula result.
type resultDim int

// tooManyCells reports whether a rows×cols array result exceeds the cell budget.
func (l Limits) tooManyCells(rows, cols resultDim) bool {
	return overCellBudget(cellBudget(l.ResultCells), rows, cols)
}

// spanTooLarge reports whether a rows×cols reference rectangle exceeds the span
// budget.
func (l Limits) spanTooLarge(rows, cols resultDim) bool {
	return overCellBudget(l.spanBudget(), rows, cols)
}

// overCellBudget reports whether a rows×cols rectangle exceeds budget cells.
//
// Each dimension is checked against the budget before they are multiplied. The
// product alone is not enough: int64 multiplication wraps, and two dimensions
// of 9223372036854775807 multiply to 1, which is under every budget — so a
// request for 9.2 quintillion cells was authorised and the allocation panicked.
// A single dimension over the budget is already over it, whatever the other is.
func overCellBudget(budget cellBudget, rows, cols resultDim) bool {
	if cellBudget(rows) > budget || cellBudget(cols) > budget || rows < 0 || cols < 0 {
		return true
	}
	return cellBudget(rows)*cellBudget(cols) > budget
}

// effectiveLimits resolves the limits for a compute pass: the zero value (an
// unset ComputeOptions.Limits) falls back to DefaultLimits, any other value is
// honored verbatim.
func effectiveLimits(l Limits) Limits {
	if l == (Limits{}) {
		return DefaultLimits()
	}
	return l
}

// maxSafeMagnitude bounds every float a builtin converts to an integer. It sits
// far below the int64 limit deliberately: no count, position or bound this far
// out can address anything real, and keeping the accepted range small means the
// arithmetic a caller performs afterwards — adding a length, taking a
// difference, multiplying two dimensions — cannot overflow either.
const maxSafeMagnitude = 1 << 40

// boundedInt truncates a float toward zero, refusing a magnitude no integer
// arithmetic can safely carry.
//
// Go leaves an out-of-range float-to-int conversion implementation-defined, and
// on a 64-bit target it saturates to the extreme int. Callers then add to it or
// subtract from it, wrap, and slice or allocate with the wrapped result — so
// every guard downstream of the conversion passes while the value is already
// nonsense. The refusal has to happen at the conversion.
func boundedInt(num floatVal) (int64, bool) {
	truncated := math.Trunc(float64(num))
	if math.IsNaN(truncated) || truncated > maxSafeMagnitude || truncated < -maxSafeMagnitude {
		return 0, false
	}
	return int64(truncated), true
}
