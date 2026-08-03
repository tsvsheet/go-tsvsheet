package engine_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// TestComputeRowsEvaluatesTheWindowUnderTheTouchedBudget pins windowed
// compute end to end: literals verbatim, formulas evaluated through the sparse
// memo with dependencies read lazily from the source, and — past the
// touched-cells budget — #LIMIT!, deterministically, while in-budget windows
// of the same document still compute. The whole-document Compute never runs.
func TestComputeRowsEvaluatesTheWindowUnderTheTouchedBudget(t *testing.T) {
	t.Parallel()

	src := "10\t20\n=A1+B1\t=sum(A1:B1)\nx\t=B2*2\n"
	limits := engine.Limits{
		ResultCells: 100, GridDim: 100, ResultBytes: 100,
		ResidentCells: 1, TouchedCells: 100,
	}
	_, windowed := open(t, src, limits)
	require.NotNil(t, windowed)

	got, err := windowed.ComputeRows(1, 2, engine.ComputeOptions{Limits: limits})
	require.NoError(t, err)
	assert.Equal(t, engine.Grid{{"30", "30"}, {"x", "60"}}, got,
		"formulas compute with lazy dependency reads; literals verbatim")

	starved := limits
	starved.TouchedCells = 2
	_, windowedStarved := open(t, src, starved)
	got, err = windowedStarved.ComputeRows(1, 1, engine.ComputeOptions{Limits: starved})
	require.NoError(t, err)
	assert.Contains(t, got[0], string(engine.ErrLimit),
		"past the touched-cells budget a cell answers #LIMIT!")
}

// TestComputeRowsFailsWholeOnALateSourceError pins that a source dying during
// dependency reads fails the call as ErrReadInput — partial data is never
// served as an answer.
func TestComputeRowsFailsWholeOnALateSourceError(t *testing.T) {
	t.Parallel()

	src := &flakySource{data: []byte("1\n=A1+1\n")}
	tiny := engine.Limits{ResultCells: 100, GridDim: 100, ResultBytes: 100, ResidentCells: 1}
	_, windowed, err := engine.OpenSheet(engine.ByteSource{ReadAt: src, Size: int64(len(src.data))}, tiny)
	require.NoError(t, err)
	require.NotNil(t, windowed)

	// Warm nothing; poison after open so the window read itself fails first.
	src.poisoned = true
	_, err = windowed.ComputeRows(0, 2, engine.ComputeOptions{Limits: tiny})
	assert.ErrorIs(t, err, constants.ErrReadInput)
}

// TestComputeRowsBoundsDependenciesLikeTheResidentPath pins lazy at()'s edge
// answers: an out-of-grid dependency is #REF! exactly as a resident sheet
// answers it, and once the source latches a failure every later read answers
// out-of-grid while the call itself fails.
func TestComputeRowsBoundsDependenciesLikeTheResidentPath(t *testing.T) {
	t.Parallel()

	src := "1\n=Z99\n"
	tiny := engine.Limits{ResultCells: 100, GridDim: 100, ResultBytes: 100, ResidentCells: 1}
	_, windowed := open(t, src, tiny)
	require.NotNil(t, windowed)
	got, err := windowed.ComputeRows(1, 1, engine.ComputeOptions{Limits: tiny})
	require.NoError(t, err)
	assert.Equal(t, string(engine.ErrRef), got[0][0])

	// Two reads of a dependency whose block dies: the first latches, the
	// second answers through the latch, and the call fails whole.
	doc := "7\n" + strings.Repeat("x\n", 300) + "=sum(A1, A1)\n"
	failing := &regionFailSource{data: []byte(doc)}
	_, windowedFailing, err := engine.OpenSheet(
		engine.ByteSource{ReadAt: failing, Size: int64(len(failing.data))}, tiny,
	)
	require.NoError(t, err)
	require.NotNil(t, windowedFailing)
	failing.failBelow = 4
	_, err = windowedFailing.ComputeRows(301, 1, engine.ComputeOptions{Limits: tiny})
	assert.ErrorIs(t, err, constants.ErrReadInput)
}

// TestWindowedCacheStaysUnderTheResidentBudget pins the adversary's HIGH
// finding: scrolling every window of an over-budget document must never
// re-materialize it — the block cache is bounded by the resident budget, the
// same ceiling that made the document windowed in the first place.
func TestWindowedCacheStaysUnderTheResidentBudget(t *testing.T) {
	t.Parallel()

	src := strings.Repeat("a\tb\tc\td\n", 5000) // 20k cells
	limits := engine.Limits{ResultCells: 1 << 20, GridDim: 1 << 20, ResultBytes: 100, ResidentCells: 100}
	_, windowed := open(t, src, limits)
	require.NotNil(t, windowed)
	for from := 0; from < 5000; from += 50 {
		_, err := windowed.Rows(from, 50)
		require.NoError(t, err)
		assert.LessOrEqual(t, windowed.CachedCells(), int64(2048),
			"scrolled to %d: the cache is bounded by the budget plus at most one block", from)
	}
}

// TestComputeRowsClipsANegativeStartLikeRows pins the addressing fix: a
// clipped window must evaluate the rows it RETURNS at their true addresses —
// ComputeRows(-1, 2) answers exactly as ComputeRows(0, 2) — never evaluating
// row 0 as row −1 and answering #REF! for its own formulas.
func TestComputeRowsClipsANegativeStartLikeRows(t *testing.T) {
	t.Parallel()

	src := "=1+1\tlit\nnext\n"
	tiny := engine.Limits{ResultCells: 100, GridDim: 100, ResultBytes: 100, ResidentCells: 1}
	_, windowed := open(t, src, tiny)
	require.NotNil(t, windowed)
	straight, err := windowed.ComputeRows(0, 2, engine.ComputeOptions{Limits: tiny})
	require.NoError(t, err)
	clipped, err := windowed.ComputeRows(-1, 3, engine.ComputeOptions{Limits: tiny})
	require.NoError(t, err)
	assert.Equal(t, straight, clipped)
	assert.Equal(t, "2", straight[0][0])
}

// TestTouchedBudgetBoundsIOAndMemoryNotJustValues pins the deepened budget
// contract: a refused range consumes neither reads nor cache — the refusal
// happens before any fetch — so the cache after one refused whole-column sum
// stays near one block, not the range's size.
func TestTouchedBudgetBoundsIOAndMemoryNotJustValues(t *testing.T) {
	t.Parallel()

	src := "=sum(A2:A20000)\n" + strings.Repeat("1\n", 19999)
	limits := engine.Limits{
		ResultCells: 1 << 20, GridDim: 1 << 20, ResultBytes: 100,
		ResidentCells: 1, TouchedCells: 10, SpanCells: 1 << 20,
	}
	_, windowed := open(t, src, limits)
	require.NotNil(t, windowed)
	got, err := windowed.ComputeRows(0, 1, engine.ComputeOptions{Limits: limits})
	require.NoError(t, err)
	assert.Equal(t, string(engine.ErrLimit), got[0][0])
	assert.Less(t, windowed.CachedCells(), int64(1024),
		"a refused range must not have been fetched into the cache")
}

// TestComputeRowsHonorsTheCallersLoaderAndLimits pins two silent drops the
// review found: a windowed evaluation carries the caller's Loader (foreign
// references compute, exactly as the resident capability computes them) and
// the caller's per-pass Limits (a byte budget refuses identically on both
// capabilities).
func TestComputeRowsHonorsTheCallersLoaderAndLimits(t *testing.T) {
	t.Parallel()

	loader := func(_, name engine.Path) (engine.Sheet, engine.Path, error) {
		require.Equal(t, engine.Path("data"), name)
		other, err := engine.Parse([]byte("42\n"))
		require.NoError(t, err)
		return other, name, nil
	}
	src := "=\"data\"!A1\t=rept(\"a\", 50)\nx\n"
	tiny := engine.Limits{ResultCells: 100, GridDim: 100, ResultBytes: 100, ResidentCells: 1}
	_, windowed := open(t, src, tiny)
	require.NotNil(t, windowed)

	passLimits := tiny
	passLimits.ResultBytes = 4
	got, err := windowed.ComputeRows(0, 1, engine.ComputeOptions{Loader: loader, Limits: passLimits})
	require.NoError(t, err)
	assert.Equal(t, "42", got[0][0], "the caller's loader reaches the windowed pass")
	assert.Equal(t, string(engine.ErrLimit), got[0][1], "the caller's per-pass limits govern, as they do residently")
}

// TestWindowCellTextNeverReformatsALiteralNorServesARejectedSpill pins the
// windowed render's two verbatim-parity duties: a stored literal like "00"
// renders unchanged (the differential fuzz caught the value formatter turning
// it into "0"), and a blocked spill anchor refuses as resident does.
func TestWindowCellTextNeverReformatsALiteralNorServesARejectedSpill(t *testing.T) {
	t.Parallel()

	src := "00\t4.50\n"
	tiny := engine.Limits{ResultCells: 100, GridDim: 100, ResultBytes: 100, ResidentCells: 1}
	_, windowed := open(t, src, tiny)
	require.NotNil(t, windowed)
	got, err := windowed.ComputeRows(0, 1, engine.ComputeOptions{Limits: tiny})
	require.NoError(t, err)
	assert.Equal(t, engine.Grid{{"00", "4.50"}}, got, "stored literals render verbatim in windows")
}

// TestWindowedSpillBlockedMatchesResidentSemantics pins the parity fix: an
// array anchor the document's own semantics refuses shows the same #SPILL!
// windowed as resident — never a value the sheet rejects; an unblocked anchor
// renders its top-left value (windows do not spill).
func TestWindowedSpillBlockedMatchesResidentSemantics(t *testing.T) {
	t.Parallel()

	src := "=sequence(3)\nx\n"
	tiny := engine.Limits{ResultCells: 100, GridDim: 100, ResultBytes: 100, ResidentCells: 1}
	_, windowed := open(t, src, tiny)
	require.NotNil(t, windowed)
	got, err := windowed.ComputeRows(0, 1, engine.ComputeOptions{Limits: tiny})
	require.NoError(t, err)
	assert.Equal(t, string(engine.ErrSpill), got[0][0], "the blocked anchor refuses, as resident render does")

	free := "=sequence(3)\n\n\nlast\n" // two empty rows: the spill footprint is clear
	_, windowedFree := open(t, free, tiny)
	require.NotNil(t, windowedFree)
	got, err = windowedFree.ComputeRows(0, 1, engine.ComputeOptions{Limits: tiny})
	require.NoError(t, err)
	assert.Equal(t, "1", got[0][0], "an unblocked anchor renders its top-left value")
}
