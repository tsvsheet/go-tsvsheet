package engine_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/constants"
	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

// open loads src under limits via the any-size entry point.
func open(t *testing.T, src string, limits engine.Limits) (engine.Sheet, *engine.WindowedSheet) {
	t.Helper()
	sheet, windowed, err := engine.OpenSheet(
		engine.ByteSource{ReadAt: bytes.NewReader([]byte(src)), Size: int64(len(src))},
		limits,
	)
	require.NoError(t, err)
	return sheet, windowed
}

// TestOpenSheetDecidesResidencyAtExactlyTheBudget pins the census policy on
// its boundary: a document whose cell count equals the resident budget loads
// fully resident and computes exactly as Parse's sheet does; one more cell
// tips it to the windowed capability, with nothing materialized.
func TestOpenSheetDecidesResidencyAtExactlyTheBudget(t *testing.T) {
	t.Parallel()

	src := "1\t2\n3\t=A1+B1\n" // 4 cells
	atBudget := engine.Limits{ResultCells: 100, GridDim: 100, ResultBytes: 100, ResidentCells: 4}
	sheet, windowed := open(t, src, atBudget)
	require.Nil(t, windowed, "at the budget the document is resident")
	parsed, err := engine.Parse([]byte(src))
	require.NoError(t, err)
	assert.Equal(t, parsed.Compute(), sheet.Compute(), "a resident open computes as Parse does")

	oneUnder := atBudget
	oneUnder.ResidentCells = 3
	sheet, windowed = open(t, src, oneUnder)
	require.NotNil(t, windowed, "one cell over the budget is windowed")
	assert.Equal(t, engine.Sheet{}, sheet)

	census := windowed.Census()
	assert.Equal(t, 2, census.Rows)
	assert.Equal(t, 2, census.MaxWidth)
	assert.Equal(t, int64(4), census.Cells)
	assert.Equal(t, int64(1), census.Formulas)
}

// TestWindowedRowsServeSourceTextsClipped pins the viewport read: source texts
// verbatim (formulas as written, nothing computed), windows clipped to the
// grid on both ends, comments never occupying a row.
func TestWindowedRowsServeSourceTextsClipped(t *testing.T) {
	t.Parallel()

	src := "#. note\nr0\tv\n=A1\tr1\nr2\n"
	tiny := engine.Limits{ResultCells: 100, GridDim: 100, ResultBytes: 100, ResidentCells: 1}
	_, windowed := open(t, src, tiny)
	require.NotNil(t, windowed)

	got, err := windowed.Rows(1, 2)
	require.NoError(t, err)
	assert.Equal(t, engine.Grid{{"=A1", "r1"}, {"r2"}}, got, "formulas stay as written")

	got, err = windowed.Rows(-3, 99)
	require.NoError(t, err)
	assert.Equal(t, engine.Grid{{"r0", "v"}, {"=A1", "r1"}, {"r2"}}, got, "hostile windows clip")

	got, err = windowed.Rows(9, 2)
	require.NoError(t, err)
	assert.Empty(t, got)
}

// TestOpenSheetFallsBackToTheSingleCeiling pins the ResidentCells fallback: a
// pre-016 Limits shape (a --max-cells constructor filling the older fields)
// keeps one ceiling — ResultCells — deciding residency.
func TestOpenSheetFallsBackToTheSingleCeiling(t *testing.T) {
	t.Parallel()

	src := strings.Repeat("a\tb\n", 3) // 6 cells
	older := engine.Limits{ResultCells: 5, GridDim: 100, ResultBytes: 100}
	_, windowed := open(t, src, older)
	require.NotNil(t, windowed, "6 cells exceed the 5-cell single ceiling")

	roomier := engine.Limits{ResultCells: 6, GridDim: 100, ResultBytes: 100}
	sheet, windowed := open(t, src, roomier)
	require.Nil(t, windowed)
	assert.Equal(t, 3, len(sheet.Compute()))
}

// TestOpenSheetRefusesADefectiveSource pins the open path's failure contract:
// a source the scanner refuses is the documented ErrReadInput before any
// policy decision, with neither capability returned.
func TestOpenSheetRefusesADefectiveSource(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("a", (1<<20)+2) + "\n"
	sheet, windowed, err := engine.OpenSheet(
		engine.ByteSource{ReadAt: bytes.NewReader([]byte(long)), Size: int64(len(long))},
		engine.DefaultLimits(),
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, constants.ErrReadInput)
	assert.Equal(t, engine.Sheet{}, sheet)
	assert.Nil(t, windowed)
}

// flakySource serves bytes until poisoned — the probe for a source that dies
// between the open-time scan and a later viewport read.
type flakySource struct {
	data     []byte
	poisoned bool
}

func (f *flakySource) ReadAt(p []byte, off int64) (int, error) {
	if f.poisoned {
		return 0, assert.AnError
	}
	if off >= int64(len(f.data)) {
		return 0, io.EOF
	}
	return copy(p, f.data[off:]), nil
}

// TestWindowedRowsSurfaceALateSourceFailureAsErrReadInput pins the windowed
// read's failure contract: a source that scans clean at open but fails under a
// later window keeps the documented ErrReadInput, never a bare error.
func TestWindowedRowsSurfaceALateSourceFailureAsErrReadInput(t *testing.T) {
	t.Parallel()

	src := &flakySource{data: []byte("a\tb\nc\td\n")}
	tiny := engine.Limits{ResultCells: 100, GridDim: 100, ResultBytes: 100, ResidentCells: 1}
	_, windowed, err := engine.OpenSheet(engine.ByteSource{ReadAt: src, Size: int64(len(src.data))}, tiny)
	require.NoError(t, err)
	require.NotNil(t, windowed)

	src.poisoned = true
	_, err = windowed.Rows(0, 2)
	assert.ErrorIs(t, err, constants.ErrReadInput)
}

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

// regionFailSource fails reads below a byte offset once armed — the probe that
// lets a window's own block read succeed while a dependency's block, earlier
// in the file, fails.
type regionFailSource struct {
	data      []byte
	failBelow int64
}

func (r *regionFailSource) ReadAt(p []byte, off int64) (int, error) {
	if r.failBelow > 0 && off < r.failBelow {
		return 0, assert.AnError
	}
	if off >= int64(len(r.data)) {
		return 0, io.EOF
	}
	return copy(p, r.data[off:]), nil
}

// TestLazyCellsNeverServeWrongDataAfterAFailure pins the lazy latch: the window
// itself reads clean, a dependency living in an earlier block fails, and the
// whole call answers ErrReadInput — never a partial grid over unread data.
func TestLazyCellsNeverServeWrongDataAfterAFailure(t *testing.T) {
	t.Parallel()

	// A1, the dependency, in block one; 300 filler rows push the formula past
	// the first stride's block.
	doc := "7\n" + strings.Repeat("x\n", 300) + "=A1+1\n"
	src := &regionFailSource{data: []byte(doc)}

	tiny := engine.Limits{ResultCells: 1 << 20, GridDim: 1 << 20, ResultBytes: 100, ResidentCells: 1}
	_, windowed, err := engine.OpenSheet(engine.ByteSource{ReadAt: src, Size: int64(len(src.data))}, tiny)
	require.NoError(t, err)
	require.NotNil(t, windowed)

	src.failBelow = 4 // the dependency's block starts at offset 0; the window's does not
	_, err = windowed.ComputeRows(301, 1, engine.ComputeOptions{Limits: tiny})
	assert.ErrorIs(t, err, constants.ErrReadInput)
}

// TestComputeRowsAnswersARefusedCellDeterministically pins the refused phase:
// a cell past the touched budget answers #LIMIT! on its first read AND on
// every re-read — the memo remembers the refusal rather than re-deciding.
func TestComputeRowsAnswersARefusedCellDeterministically(t *testing.T) {
	t.Parallel()

	src := "1\t2\t3\n=sum(C1, C1)\tz\n" // sum collects both reads; + would short-circuit the second
	starved := engine.Limits{
		ResultCells: 100, GridDim: 100, ResultBytes: 100,
		ResidentCells: 1, TouchedCells: 1,
	}
	_, windowed := open(t, src, starved)
	require.NotNil(t, windowed)
	got, err := windowed.ComputeRows(1, 1, engine.ComputeOptions{Limits: starved})
	require.NoError(t, err)
	assert.Equal(t, string(engine.ErrLimit), got[0][0],
		"the second read of the refused C1 hits the remembered refusal")
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
		engine.ByteSource{ReadAt: failing, Size: int64(len(failing.data))}, tiny)
	require.NoError(t, err)
	require.NotNil(t, windowedFailing)
	failing.failBelow = 4
	_, err = windowedFailing.ComputeRows(301, 1, engine.ComputeOptions{Limits: tiny})
	assert.ErrorIs(t, err, constants.ErrReadInput)
}
