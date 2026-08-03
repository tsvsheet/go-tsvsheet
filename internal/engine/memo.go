// The compute pass's memoization: phases for cycle detection, the memo seam,
// and its two forms — dense slabs for the resident pass, a budget-bounded
// sparse map for the windowed pass (spec 016 area 6b).
package engine

// cellPhase tracks a cell's evaluation state for memoization and cycle
// detection.
type cellPhase int

const (
	phaseUnvisited cellPhase = iota
	phaseVisiting            // on the current evaluation stack → a cycle
	phaseDone
)

// memo is one compute pass's evaluation state: the dense form preallocates the
// whole grid (the resident path, unchanged behavior), the sparse form admits
// cells up to a touched-cells budget (the windowed path, spec 016 area 6b) and
// refuses the rest as #LIMIT!.
type memo interface {
	lookup(row rowIndex, col colIndex) (Value, cellPhase)
	admit(row rowIndex, col colIndex) bool // marks visiting; false → refuse as #LIMIT!
	finish(row rowIndex, col colIndex, v Value)
}

// denseMemo is the resident pass's memo: value and phase slabs sized to the
// grid, mutated in place through the shared slices.
type denseMemo struct {
	cache [][]Value
	phase [][]cellPhase
}

// newDenseMemo sizes the slabs to the sheet.
func newDenseMemo(s Sheet) denseMemo {
	cache := make([][]Value, s.height())
	phase := make([][]cellPhase, s.height())
	for r, row := range s.rowsView() {
		cache[r] = make([]Value, len(row))
		phase[r] = make([]cellPhase, len(row))
	}
	return denseMemo{cache: cache, phase: phase}
}

func (m denseMemo) lookup(row rowIndex, col colIndex) (Value, cellPhase) {
	if !m.contains(row, col) {
		return Value{}, phaseUnvisited
	}
	return m.cache[row][col], m.phase[row][col]
}

func (m denseMemo) admit(row rowIndex, col colIndex) bool {
	if !m.contains(row, col) {
		return true // out-of-grid: the read's own grid check answers #REF!; nothing to store
	}
	m.phase[row][col] = phaseVisiting
	return true
}

func (m denseMemo) finish(row rowIndex, col colIndex, v Value) {
	if !m.contains(row, col) {
		return
	}
	m.cache[row][col] = v
	m.phase[row][col] = phaseDone
}

// contains reports whether the slabs cover (row, col): read consults the memo
// before the grid, so out-of-grid coordinates reach here and must answer
// harmlessly rather than index out of range.
func (m denseMemo) contains(row rowIndex, col colIndex) bool {
	return row >= 0 && int(row) < len(m.phase) && col >= 0 && int(col) < len(m.phase[row])
}

// memoEntry is one sparse memo cell: its value and phase.
type memoEntry struct {
	value Value
	phase cellPhase
}

// sparseMemo is the windowed pass's memo: entries keyed by address, admitted
// up to the touched-cells budget — the design's cumulative bound on values,
// memory, AND (through read's admit-before-fetch order) source I/O. Refusals
// are not stored: the map only grows, so once it is full every later admit
// refuses — deterministic by monotonicity, at no memory cost (the adversary's
// M2 mutation proved a stored refusal buys nothing).
type sparseMemo struct {
	entries map[Address]memoEntry
	budget  cellBudget
}

// newSparseMemo bounds a windowed evaluation to budget distinct cells.
func newSparseMemo(budget cellBudget) sparseMemo {
	return sparseMemo{entries: make(map[Address]memoEntry), budget: budget}
}

func (m sparseMemo) lookup(row rowIndex, col colIndex) (Value, cellPhase) {
	e := m.entries[Address{Row: int(row), Col: int(col)}]
	return e.value, e.phase
}

func (m sparseMemo) admit(row rowIndex, col colIndex) bool {
	if cellBudget(len(m.entries)) >= m.budget {
		return false
	}
	m.entries[Address{Row: int(row), Col: int(col)}] = memoEntry{phase: phaseVisiting}
	return true
}

func (m sparseMemo) finish(row rowIndex, col colIndex, v Value) {
	m.entries[Address{Row: int(row), Col: int(col)}] = memoEntry{value: v, phase: phaseDone}
}
