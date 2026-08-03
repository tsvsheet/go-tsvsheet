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
